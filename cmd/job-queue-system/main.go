// Copyright 2025 James Ross
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/flyingrobots/go-redis-work-queue/internal/admin"
	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/obs"
	"github.com/flyingrobots/go-redis-work-queue/internal/producer"
	"github.com/flyingrobots/go-redis-work-queue/internal/reaper"
	"github.com/flyingrobots/go-redis-work-queue/internal/redisclient"
	"github.com/flyingrobots/go-redis-work-queue/internal/worker"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var version = "dev"

func main() {
	var role string
	var configPath string
	var adminCmd string
	var adminQueue string
	var adminN int
	var adminYes bool
	var benchCount int
	var benchRate int
	var benchPriority string
	var benchTimeout time.Duration
	var benchPayloadSize int
	var showVersion bool
	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	fs.StringVar(&role, "role", "all", "Role to run: producer|worker|all|admin")
	fs.StringVar(&configPath, "config", "config/config.yaml", "Path to YAML config")
	fs.StringVar(&adminCmd, "admin-cmd", "", "Admin command: stats|peek|purge-dlq|purge-all|bench|stats-keys")
	fs.StringVar(&adminQueue, "queue", "", "Queue alias or full key for admin peek (high|low|completed|dead_letter|jobqueue:...)")
	fs.IntVar(&adminN, "n", 10, "Number of items for admin peek")
	fs.BoolVar(&adminYes, "yes", false, "Automatic yes to prompts (dangerous operations)")
	fs.BoolVar(&showVersion, "version", false, "Print version and exit")
	fs.IntVar(&benchCount, "bench-count", 1000, "Admin bench: number of jobs")
	fs.IntVar(&benchRate, "bench-rate", 500, "Admin bench: enqueue rate jobs/sec")
	fs.StringVar(&benchPriority, "bench-priority", "low", "Admin bench: priority/queue alias")
	fs.DurationVar(&benchTimeout, "bench-timeout", 60*time.Second, "Admin bench: timeout to wait for completion")
	fs.IntVar(&benchPayloadSize, "bench-payload-size", 1024, "Admin bench: payload size in bytes")
	_ = fs.Parse(os.Args[1:])

	if showVersion {
		fmt.Println(version)
		return
	}

	// Load configuration
	cfg, err := config.Load(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}
	// Setup logging
	logger, err := obs.NewLogger(cfg.Observability.LogLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Setup tracing (optional)
	tp, err := obs.MaybeInitTracing(cfg)
	if err != nil {
		logger.Warn("tracing init failed", obs.Err(err))
	}
	if tp != nil {
		defer func() { _ = tp.Shutdown(context.Background()) }()
	}

	// Redis client
	rdb := redisclient.New(cfg)
	defer rdb.Close()

	// HTTP server: metrics, healthz, readyz (skip for admin CLI)
	if role != "admin" {
		readyCheck := func(c context.Context) error {
			_, err := rdb.Ping(c).Result()
			return err
		}
		httpSrv := obs.StartHTTPServer(cfg, readyCheck)
		defer func() { _ = httpSrv.Shutdown(context.Background()) }()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle signals for graceful shutdown
	sigCh := make(chan os.Signal, 2)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		logger.Info("signal received, shutting down", obs.String("signal", sig.String()))
		cancel()
		// If a second signal arrives, force exit
		select {
		case sig2 := <-sigCh:
			logger.Warn("second signal received, exiting immediately", obs.String("signal", sig2.String()))
			os.Exit(1)
		case <-time.After(5 * time.Second):
		}
	}()

	// Background metrics: queue lengths (skip for admin CLI)
	if role != "admin" {
		obs.StartQueueLengthUpdater(ctx, cfg, rdb, logger)
	}

	switch role {
	case "producer":
		prod := producer.New(cfg, rdb, logger)
		if err := prod.Run(ctx); err != nil {
			logger.Fatal("producer error", obs.Err(err))
		}
	case "worker":
		wrk := worker.New(cfg, rdb, logger)
		rep := reaper.New(cfg, rdb, logger)
		go rep.Run(ctx)
		if err := wrk.Run(ctx); err != nil {
			logger.Fatal("worker error", obs.Err(err))
		}
	case "all":
		prod := producer.New(cfg, rdb, logger)
		wrk := worker.New(cfg, rdb, logger)
		rep := reaper.New(cfg, rdb, logger)
		go rep.Run(ctx)
		go func() {
			if err := prod.Run(ctx); err != nil {
				logger.Error("producer error", obs.Err(err))
				cancel()
			}
		}()
		if err := wrk.Run(ctx); err != nil {
			logger.Fatal("worker error", obs.Err(err))
		}
	case "admin":
		runAdmin(ctx, cfg, rdb, logger, adminCmd, adminQueue, adminN, adminYes, benchCount, benchRate, benchPriority, benchPayloadSize, benchTimeout)
		return
	default:
		logger.Fatal("unknown role", obs.String("role", role))
	}
}

func runAdmin(ctx context.Context, cfg *config.Config, rdb *redis.Client, logger *zap.Logger, cmd, queue string, n int, yes bool, benchCount, benchRate int, benchPriority string, benchPayloadSize int, benchTimeout time.Duration) {
	switch cmd {
	case "stats":
		res, err := admin.Stats(ctx, cfg, rdb)
		if err != nil {
			logger.Fatal("admin stats error", obs.Err(err))
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
	case "peek":
		if queue == "" {
			logger.Fatal("admin peek requires --queue")
		}
		res, err := admin.Peek(ctx, cfg, rdb, queue, int64(n))
		if err != nil {
			logger.Fatal("admin peek error", obs.Err(err))
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
	case "purge-dlq":
		if !yes {
			logger.Fatal("refusing to purge without --yes")
		}
		if err := admin.PurgeDLQ(ctx, cfg, rdb); err != nil {
			logger.Fatal("admin purge-dlq error", obs.Err(err))
		}
		fmt.Println("dead letter queue purged")
	case "purge-all":
		if !yes {
			logger.Fatal("refusing to purge without --yes")
		}
		n, err := admin.PurgeAll(ctx, cfg, rdb)
		if err != nil {
			logger.Fatal("admin purge-all error", obs.Err(err))
		}
		payload, _ := json.Marshal(struct {
			Purged int `json:"purged"`
		}{Purged: n})
		fmt.Println(string(payload))
	case "bench":
		res, err := admin.Bench(ctx, cfg, rdb, benchPriority, benchCount, benchRate, benchPayloadSize, benchTimeout)
		if err != nil {
			logger.Fatal("admin bench error", obs.Err(err))
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
	case "stats-keys":
		res, err := admin.StatsKeys(ctx, cfg, rdb)
		if err != nil {
			logger.Fatal("admin stats-keys error", obs.Err(err))
		}
		b, _ := json.MarshalIndent(res, "", "  ")
		fmt.Println(string(b))
	default:
		logger.Fatal("unknown admin command", obs.String("cmd", cmd))
	}
}
