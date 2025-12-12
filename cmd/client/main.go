package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/google/uuid"
	config "github.com/pulinau/demo-temporal-order-processor/cmd/client/config"
	"github.com/pulinau/demo-temporal-order-processor/internal/temporal"

	"go.temporal.io/sdk/client"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	configPath := flag.String("config", "", "path to config file")
	orderPayload := flag.String("order", "", "json order payload")
	flag.Parse()

	if *configPath == "" {
		*configPath = "./config/client/local/config.yaml"
	}

	if *orderPayload == "" {
		slog.Error("json order payload is required")
		flag.Usage()
		os.Exit(1)
	}

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		slog.Error("Unable to load config", "error", err)
		os.Exit(1)
	}

	c, err := client.Dial(client.Options{
		HostPort: fmt.Sprintf("%s:%d", cfg.Temporal.Host, cfg.Temporal.Port),
		Logger:   slog.Default(),
	})
	if err != nil {
		slog.Error("Unable to create client", "error", err)
		os.Exit(1)
	}
	defer c.Close()

	workflowID := "order-" + uuid.New().String()

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: cfg.Temporal.TaskQueueName,
	}

	var order temporal.Order
	err = json.Unmarshal([]byte(*orderPayload), &order)
	if err != nil {
		slog.Error("Unable to unmarshall payload into order struct", "error", err)
		os.Exit(2)
	}

	we, err := c.ExecuteWorkflow(context.Background(), options, temporal.ProccessOrder, temporal.Params{
		Order: order,
	})
	if err != nil {
		slog.Error("Unable to execute workflow", "error", err)
		os.Exit(1)
	}

	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		slog.Error("Unable get workflow result", "error", err)
		os.Exit(1)
	}

}
