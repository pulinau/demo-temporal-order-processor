package demotemporalorderprocessing

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type OrderActivities struct {
}

type Order struct {
	ID        uuid.UUID
	LineItems []LineItem
}

type LineItem struct {
	ProductID    uuid.UUID
	Quantity     int32
	PricePerItem decimal.Decimal
}

func (a *OrderActivities) Validate(ctx context.Context, order Order) error {
	if (order.ID == uuid.UUID{}) {
		return fmt.Errorf("order must have a valid order ID")
	}

	if len(order.LineItems) < 1 {
		return fmt.Errorf("order must have at least one item")
	}

	return nil
}

func (a *OrderActivities) Process(ctx context.Context, order Order) (string, error) {
	return "", errors.New("not implemented")
}
