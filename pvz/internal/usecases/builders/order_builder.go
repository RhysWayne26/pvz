package builders

import (
	"pvz-cli/internal/common/clock"
	"pvz-cli/internal/models"
	"time"
)

type OrderBuilder struct {
	clk   clock.Clock
	order *models.Order
}

func NewOrderBuilder(clk clock.Clock) *OrderBuilder {
	now := clk.Now()
	return &OrderBuilder{
		clk: clk,
		order: &models.Order{
			Status:          models.Accepted,
			CreatedAt:       now,
			UpdatedStatusAt: now,
			Package:         models.PackageNone,
		},
	}
}

// NewOrderBuilderFrom initializes and returns a new OrderBuilder with the provided clock and base Order.
func NewOrderBuilderFrom(clk clock.Clock, o *models.Order) *OrderBuilder {
	return &OrderBuilder{
		clk:   clk,
		order: o,
	}
}

func (b *OrderBuilder) WithID(id uint64) *OrderBuilder {
	b.order.OrderID = id
	return b
}

func (b *OrderBuilder) WithUserID(userID uint64) *OrderBuilder {
	b.order.UserID = userID
	return b
}

func (b *OrderBuilder) WithStatus(status models.OrderStatus) *OrderBuilder {
	curStatus := b.order.Status
	if status == curStatus {
		return b
	}

	b.order.Status = status
	b.order.UpdatedStatusAt = b.clk.Now()
	return b
}

func (b *OrderBuilder) WithExpiresAt(t time.Time) *OrderBuilder {
	b.order.ExpiresAt = t
	return b
}

func (b *OrderBuilder) WithPackageType(packageType models.PackageType) *OrderBuilder {
	b.order.Package = packageType
	return b
}

func (b *OrderBuilder) WithWeight(weight float32) *OrderBuilder {
	b.order.Weight = weight
	return b
}

func (b *OrderBuilder) WithPrice(price float32) *OrderBuilder {
	b.order.Price = price
	return b
}

// WithUpdatedStatusAt sets UpdatedStatusAt directly. For test only.
func (b *OrderBuilder) WithUpdatedStatusAt(t time.Time) *OrderBuilder {
	b.order.UpdatedStatusAt = t
	return b
}

func (b *OrderBuilder) Build() models.Order {
	return *b.order
}
