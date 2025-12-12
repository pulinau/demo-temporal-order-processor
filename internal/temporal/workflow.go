package temporal

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

var (
	defaultActivityOptions = workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    4,
		},
	}

	validateActivityOptions = workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    5,
		},
	}
)

type Params struct {
	Order Order
}

func ProccessOrder(ctx workflow.Context, in Params) error {
	var orderStatus OrderStatus

	err := workflow.SetQueryHandler(ctx, GetOrderStatus, func() (OrderStatus, error) {
		return orderStatus, nil
	})
	if err != nil {
		return fmt.Errorf("failed to setup query handler: %w", err)
	}

	// Validate order and items.
	ctx = workflow.WithActivityOptions(ctx, validateActivityOptions)

	var orderActivities *OrderActivities
	err = workflow.ExecuteActivity(ctx, orderActivities.Validate, in.Order).Get(ctx, nil)
	if err != nil {
		orderStatus = UnableToComplete
		return err
	}

	orderStatus = Placed

	// Process order.
	ctx = workflow.WithActivityOptions(ctx, defaultActivityOptions)
	var status string
	err = workflow.ExecuteActivity(ctx, orderActivities.Process, in.Order).Get(ctx, &status)
	if err != nil {
		orderStatus = UnableToComplete
		return err
	}

	workflow.GetLogger(ctx).Info("Order processed", "status", status)

	orderStatus = Comopleted

	return nil
}
