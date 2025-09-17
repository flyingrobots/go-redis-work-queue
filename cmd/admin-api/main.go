package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	adminapi "github.com/flyingrobots/go-redis-work-queue/internal/admin-api"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/flyingrobots/go-redis-work-queue/internal/redisclient"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	var configPath string
	var adminConfigPath string
	var showVersion bool

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&configPath, "config", "config/config.yaml", "Path to application YAML config")
	fs.StringVar(&adminConfigPath, "admin-config", "config/admin-api.yaml", "Path to admin API YAML config")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")

	if err := fs.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "failed to parse flags: %v\n", err)
		os.Exit(2)
	}

	if showVersion {
		fmt.Println(version)
		return
	}

	appCfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	adminCfg, err := loadAdminConfig(adminConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load admin config: %v\n", err)
		os.Exit(1)
	}

	logger, err := obs.NewLogger(appCfg.Observability.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	rdb := redisclient.New(appCfg)
	defer rdb.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go handleSignals(cancel, logger)

	if err := adminapi.Run(ctx, adminCfg, appCfg, rdb, logger); err != nil {
		logger.Fatal("admin api stopped", obs.Err(err))
	}
}

func loadAdminConfig(path string) (*adminapi.Config, error) {
	cfg := adminapi.DefaultConfig()
	if path == "" {
		return cfg, nil
	}

	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	if err := v.Unmarshal(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func handleSignals(cancel context.CancelFunc, logger *zap.Logger) {
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigCh
	logger.Info("signal received, shutting down", obs.String("signal", sig.String()))
	cancel()

	select {
	case sig2 := <-sigCh:
		logger.Warn("second signal received, forcing exit", obs.String("signal", sig2.String()))
		os.Exit(1)
	case <-time.After(5 * time.Second):
	}
}
