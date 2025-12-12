package temporal

type OrderStatus string

const (
	Placed           OrderStatus = "PLACED"
	Picked           OrderStatus = "PICKED"
	Shipped          OrderStatus = "SHIPPED"
	Completed        OrderStatus = "COMPLETED"
	Cancelled        OrderStatus = "CANCELLED"
	UnableToComplete OrderStatus = "UNABLE_TO_COMPLETE"
)

func (os OrderStatus) Valid() bool {
	switch os {
	case Placed, Picked, Shipped, Completed, Cancelled, UnableToComplete:
		return true
	default:
		return false
	}
}
