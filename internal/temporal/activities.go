package temporal

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type InventoryChecker interface {
	CheckInventory(context.Context, uuid.UUID, int32) (bool, error)
}

type OrderActivities struct {
	inventoryClient InventoryChecker
}

func NewOrderActivities(inventoryClient InventoryChecker) *OrderActivities {
	return &OrderActivities{
		inventoryClient: inventoryClient,
	}
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

	// Check inventory for each line item
	for _, item := range order.LineItems {
		available, err := a.inventoryClient.CheckInventory(ctx, item.ProductID, item.Quantity)
		if err != nil {
			return fmt.Errorf("failed to check inventory for product %s: %w", item.ProductID, err)
		}
		if !available {
			return fmt.Errorf("insufficient inventory for product %s", item.ProductID)
		}
	}

	return nil
}

func (a *OrderActivities) Process(ctx context.Context, order Order) (string, error) {
	return "", errors.New("not implemented")
}
