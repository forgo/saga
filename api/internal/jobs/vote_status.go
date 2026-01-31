package jobs

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/forgo/saga/api/internal/service"
)

// VoteStatusProcessor runs scheduled vote status transitions
// - Transitions votes from draft -> open when opens_at is reached
// - Transitions votes from open -> closed when closes_at is reached
type VoteStatusProcessor struct {
	voteService *service.VoteService
	interval    time.Duration
	stopCh      chan struct{}
	wg          sync.WaitGroup
	running     bool
	mu          sync.Mutex
}

// NewVoteStatusProcessor creates a new vote status processor job
func NewVoteStatusProcessor(voteService *service.VoteService, interval time.Duration) *VoteStatusProcessor {
	if interval == 0 {
		interval = 1 * time.Minute // Default check every minute
	}
	return &VoteStatusProcessor{
		voteService: voteService,
		interval:    interval,
		stopCh:      make(chan struct{}),
	}
}

// Start begins the vote status processor job
func (p *VoteStatusProcessor) Start() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.mu.Unlock()

	p.wg.Add(1)
	go p.run()
	log.Printf("Vote status processor started (interval: %v)", p.interval)
}

// Stop gracefully stops the vote status processor job
func (p *VoteStatusProcessor) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	close(p.stopCh)
	p.wg.Wait()
	log.Println("Vote status processor stopped")
}

// run is the main loop
func (p *VoteStatusProcessor) run() {
	defer p.wg.Done()

	// Run immediately on start (but with a short delay to let services initialize)
	time.Sleep(5 * time.Second)
	p.processVoteTransitions()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.processVoteTransitions()
		case <-p.stopCh:
			return
		}
	}
}

// processVoteTransitions processes all pending vote status transitions
func (p *VoteStatusProcessor) processVoteTransitions() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	if err := p.voteService.ProcessScheduledTransitions(ctx); err != nil {
		log.Printf("Error processing vote transitions: %v", err)
	}
}

// RunOnce runs the vote processing once (for testing or manual trigger)
func (p *VoteStatusProcessor) RunOnce(ctx context.Context) error {
	return p.voteService.ProcessScheduledTransitions(ctx)
}

// IsRunning returns whether the processor is running
func (p *VoteStatusProcessor) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
