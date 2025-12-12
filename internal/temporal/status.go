package temporal

type OrderStatus string

const (
	Placed           OrderStatus = "PLACED"
	Picked           OrderStatus = "PICKED"
	Shipped          OrderStatus = "SHIPPED"
	Comopleted       OrderStatus = "COMPLETED"
	UnableToComplete OrderStatus = "UNABLE_TO_COMPLETE"
)

const GetOrderStatus = "GetOrderStatus"
