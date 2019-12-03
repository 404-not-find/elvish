// Package apptest provides utilities for testing cli.App.
package apptest

import "github.com/elves/elvish/cli"

// Fixture is a test fixture.
type Fixture struct {
	App    cli.App
	TTY    cli.TTYCtrl
	codeCh <-chan string
	errCh  <-chan error
}

// Setup sets up a test fixture. It contains an App whose ReadCode method has
// been started asynchronously.
func Setup(fns ...func(*cli.AppSpec, cli.TTYCtrl)) *Fixture {
	tty, ttyCtrl := cli.NewFakeTTY()
	spec := cli.AppSpec{TTY: tty}
	for _, fn := range fns {
		fn(&spec, ttyCtrl)
	}
	app := cli.NewApp(spec)
	codeCh, errCh := start(app)
	return &Fixture{app, ttyCtrl, codeCh, errCh}
}

// WithSpec takes a function that operates on *cli.AppSpec, and wraps it into a
// form suitable for passing to Setup.
func WithSpec(f func(*cli.AppSpec)) func(*cli.AppSpec, cli.TTYCtrl) {
	return func(spec *cli.AppSpec, _ cli.TTYCtrl) { f(spec) }
}

// WithTTY takes a function that operates on cli.TTYCtrl, and wraps it to a form
// suitable for passing to Setup.
func WithTTY(f func(cli.TTYCtrl)) func(*cli.AppSpec, cli.TTYCtrl) {
	return func(_ *cli.AppSpec, tty cli.TTYCtrl) { f(tty) }
}

func start(app cli.App) (<-chan string, <-chan error) {
	codeCh := make(chan string, 1)
	errCh := make(chan error, 1)
	go func() {
		code, err := app.ReadCode()
		codeCh <- code
		errCh <- err
		close(codeCh)
		close(errCh)
	}()
	return codeCh, errCh
}

// Wait waits for ReaCode to finish, and returns its return values.
func (f *Fixture) Wait() (string, error) {
	return <-f.codeCh, <-f.errCh
}

// Stop stops ReadCode and waits for it to finish. If ReadCode has already been
// stopped, it is a no-op.
func (f *Fixture) Stop() {
	f.App.CommitEOF()
	f.Wait()
}