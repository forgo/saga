package jobs

import (
	"context"
	"log"
	"math"
	"sync"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// NexusCalculator defines the interface for awarding Nexus points
type NexusCalculator interface {
	AwardNexus(ctx context.Context, userID string, contributions []model.CircleNexusContribution) error
}

// NexusDataProvider defines the interface for getting Nexus calculation data
type NexusDataProvider interface {
	GetAllActiveUserIDs(ctx context.Context) ([]string, error)
	GetUserCirclesForNexus(ctx context.Context, userID string) ([]*model.NexusCircleData, error)
	GetCirclePairOverlap(ctx context.Context, circleID1, circleID2 string) (int, error)
}

// NexusMonthlyJob runs monthly Nexus calculation for all users
type NexusMonthlyJob struct {
	calculator   NexusCalculator
	dataProvider NexusDataProvider
	stopCh       chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.Mutex
}

// NewNexusMonthlyJob creates a new Nexus monthly job
func NewNexusMonthlyJob(calculator NexusCalculator, dataProvider NexusDataProvider) *NexusMonthlyJob {
	return &NexusMonthlyJob{
		calculator:   calculator,
		dataProvider: dataProvider,
		stopCh:       make(chan struct{}),
	}
}

// Start begins the Nexus monthly job (checks daily, runs on 1st of month)
func (j *NexusMonthlyJob) Start() {
	j.mu.Lock()
	if j.running {
		j.mu.Unlock()
		return
	}
	j.running = true
	j.mu.Unlock()

	j.wg.Add(1)
	go j.run()
	log.Println("Nexus monthly job started")
}

// Stop gracefully stops the job
func (j *NexusMonthlyJob) Stop() {
	j.mu.Lock()
	if !j.running {
		j.mu.Unlock()
		return
	}
	j.running = false
	j.mu.Unlock()

	close(j.stopCh)
	j.wg.Wait()
	log.Println("Nexus monthly job stopped")
}

// run checks daily if it's time to run monthly calculation
func (j *NexusMonthlyJob) run() {
	defer j.wg.Done()

	// Check every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Check on startup if we should run
	j.checkAndRun()

	for {
		select {
		case <-ticker.C:
			j.checkAndRun()
		case <-j.stopCh:
			return
		}
	}
}

// checkAndRun runs the calculation if it's the 1st of the month
func (j *NexusMonthlyJob) checkAndRun() {
	now := time.Now()
	if now.Day() == 1 {
		log.Println("Running monthly Nexus calculation")
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()

		if err := j.RunOnce(ctx); err != nil {
			log.Printf("Error running Nexus calculation: %v", err)
		}
	}
}

// RunOnce runs the Nexus calculation for all active users (for manual trigger or testing)
func (j *NexusMonthlyJob) RunOnce(ctx context.Context) error {
	// Get all users who have been active in the last 30 days
	userIDs, err := j.dataProvider.GetAllActiveUserIDs(ctx)
	if err != nil {
		return err
	}

	log.Printf("Calculating Nexus for %d active users", len(userIDs))

	processed := 0
	for _, userID := range userIDs {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err := j.calculateUserNexus(ctx, userID); err != nil {
			log.Printf("Error calculating Nexus for user %s: %v", userID, err)
			continue
		}
		processed++

		if processed%100 == 0 {
			log.Printf("Processed %d/%d users", processed, len(userIDs))
		}
	}

	log.Printf("Nexus calculation complete: %d/%d users processed", processed, len(userIDs))
	return nil
}

// calculateUserNexus calculates and awards Nexus points for a single user
func (j *NexusMonthlyJob) calculateUserNexus(ctx context.Context, userID string) error {
	// Get circle activity data
	circles, err := j.dataProvider.GetUserCirclesForNexus(ctx, userID)
	if err != nil {
		return err
	}

	if len(circles) == 0 {
		return nil // No circles, no Nexus
	}

	contributions := make([]model.CircleNexusContribution, 0)

	// Calculate per-circle contribution
	// Formula: round(5 × log₂(1 + ActiveMembers) × ActivityFactor × CircleActive)
	for _, circle := range circles {
		if !circle.IsActive {
			continue // Skip inactive circles
		}

		// Calculate circle Nexus: 5 × log₂(1 + activeMembers) × activityFactor
		circlePoints := 5.0 * math.Log2(1+float64(circle.ActiveMembers)) * circle.ActivityFactor
		points := int(math.Round(circlePoints))

		if points > 0 {
			contributions = append(contributions, model.CircleNexusContribution{
				CircleID:       circle.CircleID,
				CircleName:     circle.CircleName,
				Points:         points,
				ActivityFactor: circle.ActivityFactor,
				ActiveMembers:  circle.ActiveMembers,
			})
		}
	}

	// Calculate bridge bonuses for cross-circle activity
	// Formula: round(2 × log₂(1 + Overlap) × min(ActivityFactor_g, ActivityFactor_h))
	activeCircles := make([]*model.NexusCircleData, 0)
	for _, c := range circles {
		if c.IsActive && c.ActivityFactor > 0 {
			activeCircles = append(activeCircles, c)
		}
	}

	for i := 0; i < len(activeCircles); i++ {
		for k := i + 1; k < len(activeCircles); k++ {
			c1 := activeCircles[i]
			c2 := activeCircles[k]

			// Get overlap count
			overlap, err := j.dataProvider.GetCirclePairOverlap(ctx, c1.CircleID, c2.CircleID)
			if err != nil {
				log.Printf("Error getting circle overlap: %v", err)
				continue
			}

			if overlap == 0 {
				continue
			}

			// Calculate bridge bonus
			minActivity := math.Min(c1.ActivityFactor, c2.ActivityFactor)
			bridgePoints := 2.0 * math.Log2(1+float64(overlap)) * minActivity
			points := int(math.Round(bridgePoints))

			if points > 0 {
				contributions = append(contributions, model.CircleNexusContribution{
					CircleID:       c1.CircleID + "+" + c2.CircleID,
					CircleName:     c1.CircleName + " ↔ " + c2.CircleName,
					Points:         points,
					ActivityFactor: minActivity,
					ActiveMembers:  overlap,
				})
			}
		}
	}

	if len(contributions) == 0 {
		return nil
	}

	// Award points
	return j.calculator.AwardNexus(ctx, userID, contributions)
}

// IsRunning returns whether the job is running
func (j *NexusMonthlyJob) IsRunning() bool {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.running
}
