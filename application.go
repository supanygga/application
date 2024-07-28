package application

import (
	"io"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// Application.
type Application struct {
	onStart    func()
	onShutdown func()
	info       string
	logger     *slog.Logger
	closers    []io.Closer
	exitSignal chan struct{}
}

// New.
func New(
	onStart func(),
	onShutdown func(),
	info string,
	logger *slog.Logger,
) *Application {
	app := &Application{
		onStart:    onStart,
		onShutdown: onShutdown,
		info:       info,
		logger:     logger,
	}

	if app.onStart == nil {
		app.onStart = func() {}
	}

	if app.onShutdown == nil {
		app.onShutdown = func() { app.ExecuteClosers() }
	}

	if app.logger == nil {
		app.logger = slog.Default()
	}

	return app
}

// Run.
func (a *Application) Run() {
	a.logger.Info("starting application")

	a.exitSignal = make(chan struct{})
	defer close(a.exitSignal)

	a.onStart()
	a.logger.Info("running, waiting for exit signal")

	go a.awaitTermination()
	<-a.exitSignal

	a.onShutdown()
}

// AddClosers.
func (a *Application) AddClosers(closers ...io.Closer) {
	a.closers = append(a.closers, closers...)
}

// ExecuteClosers.
func (a *Application) ExecuteClosers() {
	for i := len(a.closers) - 1; i >= 0; i-- {
		closer := a.closers[i]
		if err := closer.Close(); err != nil {
			a.logger.Error("unable to execute closer", "error", err)
		}
	}
}

// Closers.
func (a *Application) Closers() []io.Closer {
	return a.closers
}

// Info.
func (a *Application) Info() string {
	return a.info
}

// Logger.
func (a *Application) Logger() *slog.Logger {
	return a.logger
}

// awaitTermination.
func (a *Application) awaitTermination() {
	signalCh := make(chan os.Signal, 1)
	defer close(signalCh)

	signal.Notify(signalCh, syscall.SIGTERM, syscall.SIGINT)

	for s := range signalCh {
		a.logger.Info("terminating application", "signal", s)
		a.exitSignal <- struct{}{}
		return
	}
}
