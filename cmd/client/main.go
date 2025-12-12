package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"

	"github.com/google/uuid"
	config "github.com/pulinau/demo-temporal-order-processor/cmd/client/config"
	"github.com/pulinau/demo-temporal-order-processor/internal/temporal"

	"go.temporal.io/sdk/client"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))

	if len(os.Args) <= 1 {
		slog.Error("json payload required as command-line argument")
		os.Exit(1)
	}
	orderPayload := os.Args[1]

	//TODO: Read config from YAAL.
	cfg := config.Config{
		Temporal: temporal.Config{
			Host:          "localhost",
			Port:          7233,
			TaskQueueName: "order-proccesor-queue",
		},
	}

	c, err := client.Dial(client.Options{})
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
	err = json.Unmarshal([]byte(orderPayload), &order)
	if err != nil {
		slog.Error("Unable to unmarshall payload into order struct", "error", err)
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
