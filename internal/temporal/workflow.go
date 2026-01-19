package temporal

import (
	"fmt"
	"time"

	"github.com/google/uuid"
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

// Define signals.
const (
	pickOrderSignal      = "pickOrder"
	shipOrderSignal      = "shipOrder"
	orderDeliveredSignal = "markOrderAsDelivered"
	cancelOrderSignal    = "cancelOrder"
)

type Params struct {
	Order Order
}

func ProccessOrder(ctx workflow.Context, in Params) (OrderStatus, error) {
	logger := workflow.GetLogger(ctx)

	var orderStatus OrderStatus

	err := workflow.SetQueryHandler(ctx, "GetOrderStatus", func() (OrderStatus, error) {
		return orderStatus, nil
	})
	if err != nil {
		return orderStatus, fmt.Errorf("failed to setup query handler: %w", err)
	}

	// Validate order and items.
	ctx = workflow.WithActivityOptions(ctx, validateActivityOptions)

	var orderActivities *OrderActivities
	err = workflow.ExecuteActivity(ctx, orderActivities.Validate, in.Order).Get(ctx, nil)
	if err != nil {
		orderStatus = UnableToComplete
		return orderStatus, err
	}
	orderStatus = Placed

	// Wait for order picked or order cancelled signals.
	pickOrderCh := workflow.GetSignalChannel(ctx, pickOrderSignal)
	cancelOrderCh := workflow.GetSignalChannel(ctx, cancelOrderSignal)

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(pickOrderCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, nil)
		now := workflow.Now(ctx)
		logger.Info("Order picked at %s", now.Format("2006-01-02 15:04:05"))
		orderStatus = Picked
	})
	selector.AddReceive(cancelOrderCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, nil)
		orderStatus = Cancelled
	})

	// Blocks until signal is received.
	selector.Select(ctx)
	if orderStatus == Cancelled {
		workflow.GetLogger(ctx).Warn("Received cancellation signal")
		return orderStatus, nil
	}

	// Process order.
	var payID uuid.UUID
	if err = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.New()
	}).Get(&payID); err != nil {
		orderStatus = UnableToComplete
		return orderStatus, nil
	}

	future := workflow.ExecuteChildWorkflow(ctx, ProcessPayment, PaymentDetails{
		PayID:  payID,
		Amount: 100.0, // TODO: calculate amount.
	})
	err = future.Get(ctx, nil)

	// Wait for order to be shipped.
	workflow.GetSignalChannel(ctx, shipOrderSignal).Receive(ctx, nil)
	orderStatus = Shipped

	// Wait for order to be marked as delivered.
	workflow.GetSignalChannel(ctx, orderDeliveredSignal).Receive(ctx, nil)
	orderStatus = Completed

	return orderStatus, nil
}

type PaymentDetails struct {
	PayID  uuid.UUID
	Amount float64
}

func ProcessPayment(ctx workflow.Context, in PaymentDetails) error {
	// TODO: add payment processing logic.
	workflow.GetLogger(ctx).Info("Did some processing and recieved the payment")

	return nil

}
