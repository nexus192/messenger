package main

import (
	"log/slog"
	"messenger/src/config"
	"messenger/src/internal/adapter"
	"messenger/src/internal/data/repository"
	"messenger/src/internal/domain/service"
	"messenger/src/internal/server"
	"os"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	cfg := config.MustLoad()

	log := SetupLogger(cfg.Env)
	log.Info("Start", slog.String("env", cfg.Env))
	log.Debug("debag enable")

	repo, err := repository.NewPGRepository(cfg.DatabaseMess, log)
	if err != nil {
		log.Error("cannot connect to db", slog.Any("err", err))
		return
	}

	h := service.NewHub()

	adaRepo := &adapter.PostgresAdapter{Repo: repo}

	auth := service.NewAuthService(adaRepo)

	chat := service.NewChatService(adaRepo, h)

	s := server.NewServer(h, auth, chat, log)

	go h.Run()
	s.Start()
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}
