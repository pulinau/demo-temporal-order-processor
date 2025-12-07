package demotemporalorderprocessing

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type OrderActivities struct {
}

type Order struct {
	ID uuid.UUID
}

func (a *OrderActivities) Validate(ctx context.Context, order Order) error {
	if (order.ID == uuid.UUID{}) {
		return fmt.Errorf("order must have a valid order ID")
	}

	return nil
}

func (a *OrderActivities) Process(ctx context.Context, order Order) (string, error) {
	return "", errors.New("not implemented")
}
