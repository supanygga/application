package application_test

import (
	"errors"
	"io"
	"log/slog"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/supanygga/application"
)

type closer struct{}

func (c *closer) Close() error {
	return errors.New("closer error")
}

func TestApplication(t *testing.T) {
	assert := assert.New(t)

	app := &application.Application{}
	info := "info"
	logger := slog.Default()
	onStart := func() {
		go func() {
			time.Sleep(1 * time.Second)
			if err := syscall.Kill(syscall.Getpid(), syscall.SIGTERM); err != nil {
				logger.Error("syscall kill", "error", err)
				os.Exit(1)
			}
		}()

		app.AddClosers(&closer{})
	}
	onShutdown := func() {
		app.ExecuteClosers()
	}

	app = application.New(onStart, onShutdown, info, logger)
	app.Run()

	assert.Equal(info, app.Info())
	assert.Equal(logger, app.Logger())
	assert.Equal([]io.Closer{&closer{}}, app.Closers())
}
