package demotemporalorderprocessing

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type Params struct {
	Order Order
}

func ProccessOrder(ctx workflow.Context, in Params) error {

	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			MaximumInterval:    time.Minute,
			BackoffCoefficient: 2,
		},
	})

	var orderActivities *OrderActivities
	err := workflow.ExecuteActivity(ctx, orderActivities.Validate, in.Order).Get(ctx, nil)
	if err != nil {
		return err
	}

	var status string
	err = workflow.ExecuteActivity(ctx, orderActivities.Process, in.Order).Get(ctx, &status)
	if err != nil {
		return err
	}

	workflow.GetLogger(ctx).Info("Order processed", "status", status)

	return nil
}
