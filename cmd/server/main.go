package main

import (
	"context"
	"log"
	"time"

	"arkis_test/internal/application"
	"arkis_test/internal/config"
	"arkis_test/internal/controlplane"
	"arkis_test/internal/infrastructure/database/staticdb"

	"go.uber.org/fx"
	"go.uber.org/zap"
)

func main() {
	l, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("could not initialize logger: %v", err)
	}
	defer func(l *zap.Logger) {
		err := l.Sync()
		if err != nil {
			log.Printf("failed flushing buffered log entries - logger: %v", err)
		}
	}(l)

	undo := zap.ReplaceGlobals(l)
	defer undo()

	fxApp := fx.New(
		fx.Provide(config.New),
		fx.Provide(staticdb.New),
		fx.Provide(application.NewStreamService),
		fx.Provide(controlplane.NewServer),
		fx.Invoke(registerHooks),
	)

	if err := fxApp.Start(context.Background()); err != nil {
		zap.L().Fatal("failed to start the app %v", zap.Error(err))
	}

	<-fxApp.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := fxApp.Stop(shutdownCtx); err != nil {
		zap.L().Fatal("failed to stop the app %v", zap.Error(err))
	}
}

func registerHooks(lc fx.Lifecycle, s *controlplane.Server) error {
	lc.Append(
		fx.Hook{
			OnStart: func(_ context.Context) error {
				go s.Start()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				return s.Shutdown(ctx)
			},
		},
	)

	return nil
}
