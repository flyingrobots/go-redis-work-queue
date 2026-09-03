// Copyright 2026 James Ross
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/flyingrobots/go-redis-work-queue/internal/config"
	"github.com/flyingrobots/go-redis-work-queue/internal/redisclient"
	"github.com/flyingrobots/go-redis-work-queue/pkg/queueclient"
)

func runEnqueue(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	fs := flag.NewFlagSet("enqueue", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var configPath string
	var payloadFile string
	var schema string
	var priority string
	var orderingKey string
	var id string
	fs.StringVar(&configPath, "config", "config/config.yaml", "path to YAML config")
	fs.StringVar(&payloadFile, "payload-file", "-", "payload file, or - for stdin")
	fs.StringVar(&schema, "schema", "", "caller-owned payload schema or version")
	fs.StringVar(&priority, "priority", "", "configured queue priority (defaults from config)")
	fs.StringVar(&orderingKey, "ordering-key", "", "per-key FIFO identity (stored now; enforced by ROADMAP Item 4)")
	fs.StringVar(&id, "id", "", "caller-owned job ID (duplicates are separate deliveries)")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if fs.NArg() != 0 {
		return fmt.Errorf("unexpected positional arguments: %v", fs.Args())
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}
	payload, err := readPayload(payloadFile, stdin)
	if err != nil {
		return err
	}

	client, err := queueclient.New(redisclient.Options(cfg), publicClientConfig(cfg))
	if err != nil {
		return fmt.Errorf("create queue client: %w", err)
	}
	defer func() { _ = client.Close() }()

	jobID, err := client.Enqueue(context.Background(), queueclient.Job{
		ID:            id,
		Payload:       payload,
		PayloadSchema: schema,
		OrderingKey:   orderingKey,
		Priority:      priority,
	})
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(stdout, jobID); err != nil {
		return fmt.Errorf("print job ID: %w", err)
	}
	return nil
}

func readPayload(path string, stdin io.Reader) ([]byte, error) {
	if path == "" || path == "-" {
		payload, err := io.ReadAll(stdin)
		if err != nil {
			return nil, fmt.Errorf("read payload from stdin: %w", err)
		}
		return payload, nil
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read payload file %q: %w", path, err)
	}
	return payload, nil
}

func publicClientConfig(cfg *config.Config) queueclient.Config {
	return queueclient.Config{
		Queues:                cfg.Worker.Queues,
		DefaultPriority:       cfg.Producer.DefaultPriority,
		ProcessingListPattern: cfg.Worker.ProcessingListPattern,
		HeartbeatKeyPattern:   cfg.Worker.HeartbeatKeyPattern,
		CompletedList:         cfg.Worker.CompletedList,
		DeadLetterList:        cfg.Worker.DeadLetterList,
		MaxPayloadSize:        cfg.Queue.MaxPayloadSize,
	}
}
