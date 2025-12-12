package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"

	config "github.com/pulinau/demo-temporal-order-processor/cmd/worker/config"
	"github.com/pulinau/demo-temporal-order-processor/internal/integrations/inventory"
	"github.com/pulinau/demo-temporal-order-processor/internal/temporal"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	configPath := flag.String("config", "", "path to config file")
	flag.Parse()

	if *configPath == "" {
		*configPath = "./config/worker/local/config.yaml"
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		slog.Error("Unable to load config", "error", err)
		os.Exit(1)
	}

	// Create the Temporal client,
	c, err := client.Dial(client.Options{
		HostPort: fmt.Sprintf("%s:%d", cfg.Temporal.Host, cfg.Temporal.Port),
		Logger:   slog.Default(),
	})
	if err != nil {
		slog.Error("Unable to create Temporal client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	// Create the Temporal worker,
	w := worker.New(c, cfg.Temporal.TaskQueueName, worker.Options{})

	// inject HTTP client into the Activities Struct,
	activities := temporal.NewOrderActivities(inventory.NewClient(cfg.InventoryAPI.BaseURL))

	// Register Workflow and Activities
	w.RegisterWorkflow(temporal.ProccessOrder)
	w.RegisterActivity(activities)

	// Start the Worker
	if err := w.Run(worker.InterruptCh()); err != nil {
		slog.Default().Error("Unable to start Temporal worker", "error", err)
		os.Exit(1)
	}

	slog.Info("Shutting down...")
}
