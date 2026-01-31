package jobs

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/forgo/saga/api/internal/service"
)

// NudgeProcessor runs scheduled nudge processing
type NudgeProcessor struct {
	nudgeService *service.NudgeService
	interval     time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
	running      bool
	mu           sync.Mutex
}

// NewNudgeProcessor creates a new nudge processor job
func NewNudgeProcessor(nudgeService *service.NudgeService, interval time.Duration) *NudgeProcessor {
	if interval == 0 {
		interval = 15 * time.Minute // Default check every 15 minutes
	}
	return &NudgeProcessor{
		nudgeService: nudgeService,
		interval:     interval,
		stopCh:       make(chan struct{}),
	}
}

// Start begins the nudge processor job
func (p *NudgeProcessor) Start() {
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.mu.Unlock()

	p.wg.Add(1)
	go p.run()
	log.Printf("Nudge processor started (interval: %v)", p.interval)
}

// Stop gracefully stops the nudge processor job
func (p *NudgeProcessor) Stop() {
	p.mu.Lock()
	if !p.running {
		p.mu.Unlock()
		return
	}
	p.running = false
	p.mu.Unlock()

	close(p.stopCh)
	p.wg.Wait()
	log.Println("Nudge processor stopped")
}

// run is the main loop
func (p *NudgeProcessor) run() {
	defer p.wg.Done()

	// Run immediately on start (but with a short delay to let services initialize)
	time.Sleep(5 * time.Second)
	p.processNudges()

	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			p.processNudges()
		case <-p.stopCh:
			return
		}
	}
}

// processNudges processes all pending nudges
func (p *NudgeProcessor) processNudges() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	if err := p.nudgeService.ProcessPendingNudges(ctx); err != nil {
		log.Printf("Error processing nudges: %v", err)
	}
}

// RunOnce runs the nudge processing once (for testing or manual trigger)
func (p *NudgeProcessor) RunOnce(ctx context.Context) error {
	return p.nudgeService.ProcessPendingNudges(ctx)
}

// IsRunning returns whether the processor is running
func (p *NudgeProcessor) IsRunning() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.running
}
