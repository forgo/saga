// Package jobs implements background job processing for the Saga API.
//
// The jobs package contains scheduled and background tasks that run
// independently of HTTP request handling.
//
// # Job Types
//
// Available background jobs:
//
//   - ResonanceCalculator: Monthly social scoring updates
//   - TokenCleanup: Expired token removal
//   - NotificationProcessor: Async notification delivery
//
// # Job Runner
//
// Jobs are managed by a central runner:
//
//	runner := jobs.NewRunner(jobs.RunnerConfig{
//	    DB:     database,
//	    Logger: logger,
//	})
//	runner.Schedule("0 0 1 * *", resonanceJob) // Monthly
//	runner.Start()
//
// # Job Interface
//
// All jobs implement the Job interface:
//
//	type Job interface {
//	    Name() string
//	    Run(ctx context.Context) error
//	}
//
// # Error Handling
//
// Jobs log errors but don't crash the application.
// Failed jobs are retried according to their configuration.
package jobs
