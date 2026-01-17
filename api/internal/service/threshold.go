package service

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/forgo/saga/api/internal/model"
)

// ThresholdChecker defines the interface for checking timers
type ThresholdChecker interface {
	GetAllActiveTimersWithActivities(ctx context.Context) ([]model.TimerWithActivity, error)
	GetCircleIDForPerson(ctx context.Context, personID string) (string, error)
}

// ThresholdMonitor monitors timers and emits threshold events
type ThresholdMonitor struct {
	checker    ThresholdChecker
	eventHub   *EventHub
	interval   time.Duration
	ticker     *time.Ticker
	done       chan struct{}
	mu         sync.Mutex
	warnSent   map[string]time.Time // timerID -> last warn time
	critSent   map[string]time.Time // timerID -> last crit time
	cooldown   time.Duration        // Cooldown between repeated threshold events
}

// ThresholdMonitorConfig holds configuration for the threshold monitor
type ThresholdMonitorConfig struct {
	Checker  ThresholdChecker
	EventHub *EventHub
	Interval time.Duration // How often to check (default 30s)
	Cooldown time.Duration // Cooldown between repeated alerts (default 5 min)
}

// NewThresholdMonitor creates a new threshold monitor
func NewThresholdMonitor(cfg ThresholdMonitorConfig) *ThresholdMonitor {
	interval := cfg.Interval
	if interval == 0 {
		interval = 30 * time.Second
	}
	cooldown := cfg.Cooldown
	if cooldown == 0 {
		cooldown = 5 * time.Minute
	}

	return &ThresholdMonitor{
		checker:  cfg.Checker,
		eventHub: cfg.EventHub,
		interval: interval,
		cooldown: cooldown,
		done:     make(chan struct{}),
		warnSent: make(map[string]time.Time),
		critSent: make(map[string]time.Time),
	}
}

// Start begins threshold monitoring
func (m *ThresholdMonitor) Start() {
	m.ticker = time.NewTicker(m.interval)
	go m.run()
	slog.Info("threshold monitor started", slog.Duration("interval", m.interval))
}

// Stop stops threshold monitoring
func (m *ThresholdMonitor) Stop() {
	close(m.done)
	if m.ticker != nil {
		m.ticker.Stop()
	}
	slog.Info("threshold monitor stopped")
}

func (m *ThresholdMonitor) run() {
	for {
		select {
		case <-m.ticker.C:
			m.check()
		case <-m.done:
			return
		}
	}
}

func (m *ThresholdMonitor) check() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	timers, err := m.checker.GetAllActiveTimersWithActivities(ctx)
	if err != nil {
		slog.Error("failed to get timers for threshold check", slog.String("error", err.Error()))
		return
	}

	now := time.Now()
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tw := range timers {
		if !tw.Timer.Enabled {
			continue
		}

		elapsed := now.Sub(tw.Timer.ResetDate).Seconds()

		// Get circle ID for this person
		circleID, err := m.checker.GetCircleIDForPerson(ctx, tw.Timer.PersonID)
		if err != nil {
			continue
		}

		// Check critical threshold first (higher priority)
		if tw.Activity.Critical > 0 && elapsed >= tw.Activity.Critical {
			if m.canSendAlert(tw.Timer.ID, "crit", now) {
				m.eventHub.Publish(&Event{
					Type:     EventTimerCritical,
					CircleID: circleID,
					Data: map[string]interface{}{
						"id":            tw.Timer.ID,
						"threshold":     tw.Activity.Critical,
						"elapsed":       elapsed,
						"person_id":     tw.Timer.PersonID,
						"activity_id":   tw.Timer.ActivityID,
						"activity_name": tw.Activity.Name,
					},
				})
				m.critSent[tw.Timer.ID] = now
			}
		} else if tw.Activity.Warn > 0 && elapsed >= tw.Activity.Warn {
			// Check warn threshold
			if m.canSendAlert(tw.Timer.ID, "warn", now) {
				m.eventHub.Publish(&Event{
					Type:     EventTimerWarn,
					CircleID: circleID,
					Data: map[string]interface{}{
						"id":            tw.Timer.ID,
						"threshold":     tw.Activity.Warn,
						"elapsed":       elapsed,
						"person_id":     tw.Timer.PersonID,
						"activity_id":   tw.Timer.ActivityID,
						"activity_name": tw.Activity.Name,
					},
				})
				m.warnSent[tw.Timer.ID] = now
			}
		}
	}

	// Cleanup old entries
	m.cleanup(now)
}

func (m *ThresholdMonitor) canSendAlert(timerID, alertType string, now time.Time) bool {
	var lastSent time.Time
	var ok bool

	if alertType == "warn" {
		lastSent, ok = m.warnSent[timerID]
	} else {
		lastSent, ok = m.critSent[timerID]
	}

	if !ok {
		return true
	}

	return now.Sub(lastSent) >= m.cooldown
}

func (m *ThresholdMonitor) cleanup(now time.Time) {
	// Remove entries older than 24 hours
	cutoff := now.Add(-24 * time.Hour)

	for id, t := range m.warnSent {
		if t.Before(cutoff) {
			delete(m.warnSent, id)
		}
	}
	for id, t := range m.critSent {
		if t.Before(cutoff) {
			delete(m.critSent, id)
		}
	}
}

// ClearTimer removes tracking for a timer (call when timer is reset)
func (m *ThresholdMonitor) ClearTimer(timerID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.warnSent, timerID)
	delete(m.critSent, timerID)
}
