package runner

import "context"

// Start runs a harness asynchronously. Startup failures are returned through done.
func Start(ctx context.Context, session *Session, config HarnessConfig) <-chan error {
	done := make(chan error, 1)
	go func() { done <- RunHarness(ctx, session, config); close(done) }()
	return done
}
