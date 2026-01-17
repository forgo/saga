package jobs

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/forgo/saga/api/internal/service"
)

// PoolMatcher runs scheduled matching for pools
type PoolMatcher struct {
	poolService *service.PoolService
	interval    time.Duration
	stopCh      chan struct{}
	wg          sync.WaitGroup
	running     bool
	mu          sync.Mutex
}

// NewPoolMatcher creates a new pool matcher job
func NewPoolMatcher(poolService *service.PoolService, interval time.Duration) *PoolMatcher {
	if interval == 0 {
		interval = 1 * time.Hour // Default check every hour
	}
	return &PoolMatcher{
		poolService: poolService,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

// Start begins the pool matcher job
func (m *PoolMatcher) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.mu.Unlock()

	m.wg.Add(1)
	go m.run()
	log.Printf("Pool matcher started (interval: %v)", m.interval)
}

// Stop gracefully stops the pool matcher job
func (m *PoolMatcher) Stop() {
	m.mu.Lock()
	if !m.running {
		m.mu.Unlock()
		return
	}
	m.running = false
	m.mu.Unlock()

	close(m.stopCh)
	m.wg.Wait()
	log.Println("Pool matcher stopped")
}

// run is the main loop
func (m *PoolMatcher) run() {
	defer m.wg.Done()

	// Run immediately on start
	m.processPoolsUnsafe()

	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.processPoolsUnsafe()
		case <-m.stopCh:
			return
		}
	}
}

// processPoolsUnsafe processes all pools due for matching
func (m *PoolMatcher) processPoolsUnsafe() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	pools, err := m.poolService.GetPoolsDueForMatching(ctx)
	if err != nil {
		log.Printf("Error getting pools due for matching: %v", err)
		return
	}

	if len(pools) == 0 {
		return
	}

	log.Printf("Processing %d pools due for matching", len(pools))

	for _, pool := range pools {
		if err := m.processPool(ctx, pool.ID); err != nil {
			log.Printf("Error processing pool %s: %v", pool.ID, err)
			continue
		}
	}
}

// processPool runs matching for a single pool
func (m *PoolMatcher) processPool(ctx context.Context, poolID string) error {
	log.Printf("Running matching for pool %s", poolID)

	roundInfo, err := m.poolService.RunMatching(ctx, poolID)
	if err != nil {
		return err
	}

	log.Printf("Pool %s: created %d matches for round %s", poolID, roundInfo.MatchCount, roundInfo.Round)

	// TODO: Send notifications to matched members
	// This would integrate with a notification service

	return nil
}

// RunOnce runs the matching process once (for testing or manual trigger)
func (m *PoolMatcher) RunOnce(ctx context.Context) error {
	pools, err := m.poolService.GetPoolsDueForMatching(ctx)
	if err != nil {
		return err
	}

	for _, pool := range pools {
		if err := m.processPool(ctx, pool.ID); err != nil {
			log.Printf("Error processing pool %s: %v", pool.ID, err)
		}
	}

	return nil
}

// IsRunning returns whether the matcher is running
func (m *PoolMatcher) IsRunning() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}
