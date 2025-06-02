package main

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/younsl/eip-rotation-handler/pkg/configs"
	"github.com/younsl/eip-rotation-handler/pkg/health"
	"github.com/younsl/eip-rotation-handler/pkg/logger"
	"github.com/younsl/eip-rotation-handler/pkg/rotation"
)

var (
	// Version is set during build time
	Version   = "dev"
	Commit    = "unknown"
	BuildTime = "unknown"
)

// Application manages the entire application lifecycle
type Application struct {
	cfg           *configs.Config
	log           *logrus.Logger
	healthServer  *health.Server
	eipHandler    *rotation.Handler
	shutdownOnce  sync.Once
	shutdownFuncs []func() error
}

// newApplication creates a new application instance
func newApplication() (*Application, error) {
	// Load configuration
	cfg, err := configs.Load()
	if err != nil {
		return nil, err
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)

	log.WithFields(logrus.Fields{
		"version":  Version,
		"interval": cfg.RotationInterval,
	}).Info("Initializing EIP Rotation Handler")

	// Create EIP handler
	eipHandler, err := rotation.New(cfg, log)
	if err != nil {
		return nil, err
	}

	// Create health server
	healthServer := health.New(log)

	app := &Application{
		cfg:          cfg,
		log:          log,
		healthServer: healthServer,
		eipHandler:   eipHandler,
	}

	return app, nil
}

// start starts all application services
func (a *Application) start(ctx context.Context) error {
	a.log.Info("Starting application services...")

	// Start health check server
	a.healthServer.Start(ctx)
	a.addShutdownFunc(func() error {
		a.healthServer.Shutdown()
		return nil
	})

	// Validate AWS credentials before starting main service
	if err := a.eipHandler.ValidateAWSCredentials(ctx); err != nil {
		return err
	}

	// Start EIP rotation handler
	go a.eipHandler.Start(ctx)

	a.log.Info("All services started successfully")
	return nil
}

// run starts the application and waits for shutdown signals
func (a *Application) run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start all services
	if err := a.start(ctx); err != nil {
		return err
	}

	// Wait for shutdown signal
	a.waitForShutdown(cancel)

	// Graceful shutdown
	a.shutdown()

	return nil
}

// waitForShutdown waits for shutdown signals
func (a *Application) waitForShutdown(cancel context.CancelFunc) {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	a.log.WithField("signal", sig).Info("Received shutdown signal")
	cancel()
}

// shutdown performs graceful shutdown
func (a *Application) shutdown() {
	a.shutdownOnce.Do(func() {
		a.log.Info("Shutting down application...")

		// Execute all shutdown functions in reverse order
		for i := len(a.shutdownFuncs) - 1; i >= 0; i-- {
			if err := a.shutdownFuncs[i](); err != nil {
				a.log.WithError(err).Error("Error during shutdown")
			}
		}

		a.log.Info("Application shutdown complete")
	})
}

// addShutdownFunc adds a function to be called during shutdown
func (a *Application) addShutdownFunc(fn func() error) {
	a.shutdownFuncs = append(a.shutdownFuncs, fn)
}

func main() {
	// Create and start application
	app, err := newApplication()
	if err != nil {
		logrus.WithError(err).Fatal("Failed to initialize application")
	}

	// Run application until shutdown
	if err := app.run(); err != nil {
		logrus.WithError(err).Fatal("Application failed")
	}
}
