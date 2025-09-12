package main

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/flyingrobots/go-redis-work-queue/internal/config"
    "github.com/flyingrobots/go-redis-work-queue/internal/obs"
    "github.com/flyingrobots/go-redis-work-queue/internal/producer"
    "github.com/flyingrobots/go-redis-work-queue/internal/redisclient"
    "github.com/flyingrobots/go-redis-work-queue/internal/reaper"
    "github.com/flyingrobots/go-redis-work-queue/internal/worker"
)

func main() {
    var role string
    var configPath string
    fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
    fs.StringVar(&role, "role", "all", "Role to run: producer|worker|all")
    fs.StringVar(&configPath, "config", "config/config.yaml", "Path to YAML config")
    _ = fs.Parse(os.Args[1:])

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

    // Metrics server
    metricsSrv := obs.StartMetricsServer(cfg)
    defer func() { _ = metricsSrv.Shutdown(context.Background()) }()

    // Redis client
    rdb := redisclient.New(cfg)
    defer rdb.Close()

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
    default:
        logger.Fatal("unknown role", obs.String("role", role))
    }
}

