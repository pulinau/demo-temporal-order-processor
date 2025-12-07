package main

import (
	"fmt"
	"log/slog"
	"os"

	demotemporalorderprocessing "github.com/pulinau/demo-temporal-order-processor"
	"github.com/pulinau/demo-temporal-order-processor/cmd/worker/config"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	//TODO: Read config from YAAL.
	cfg := config.WorkerConfig{
		Temporal: config.TemporalConfig{
			Host:          "localhost",
			Port:          7233,
			TaskQueueName: "order-proccesor-queue",
		},
	}

	// Create the Temporal client,
	c, err := client.Dial(client.Options{
		HostPort: fmt.Sprintf("%s:%d", cfg.Temporal.Host, cfg.Temporal.Port),
		Logger:   slog.Default(),
	})
	if err != nil {
		slog.Error("Unable to create Temporal client")
		os.Exit(1)
	}
	defer c.Close()

	// Create the Temporal worker,
	w := worker.New(c, cfg.Temporal.TaskQueueName, worker.Options{})

	// inject HTTP client into the Activities Struct,
	activities := &demotemporalorderprocessing.OrderActivities{}

	// Register Workflow and Activities
	w.RegisterWorkflow(demotemporalorderprocessing.ProccessOrder)
	w.RegisterActivity(activities)

	// Start the Worker
	if err := w.Run(worker.InterruptCh()); err != nil {
		slog.Default().Error("Unable to start Temporal worker")
		os.Exit(1)
	}

	slog.Info("Shutting down...")
}
