package worker

import (
	"github.com/pulinau/demo-temporal-order-processor/internal/integrations/inventory"
	"github.com/pulinau/demo-temporal-order-processor/internal/temporal"
)

type Config struct {
	Temporal     temporal.Config
	InventoryAPI inventory.Config
}
