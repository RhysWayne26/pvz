package repositories

import (
	"errors"
	"pvz-cli/internal/constants"
	"sort"

	"github.com/google/uuid"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
)

type snapshotOrderRepository struct {
	storage storage.Storage
}

func NewSnapshotOrderRepository(s storage.Storage) OrderRepository {
	return &snapshotOrderRepository{storage: s}
}

func (r *snapshotOrderRepository) Save(order models.Order) error {
	snap, err := r.storage.Load()
	if err != nil {
		return err
	}

	found := false
	for i, o := range snap.Orders {
		if o.OrderID == order.OrderID {
			snap.Orders[i] = order
			found = true
			break
		}
	}

	if !found {
		snap.Orders = append(snap.Orders, order)
	}

	return r.storage.Save(snap)
}

func (r *snapshotOrderRepository) Load(id uuid.UUID) (models.Order, error) {
	snap, err := r.storage.Load()
	if err != nil {
		return models.Order{}, err
	}

	for _, o := range snap.Orders {
		if o.OrderID == id {
			return o, nil
		}
	}
	return models.Order{}, errors.New("order not found")
}

func (r *snapshotOrderRepository) Delete(id uuid.UUID) error {
	snap, err := r.storage.Load()
	if err != nil {
		return err
	}

	filtered := make([]models.Order, 0, len(snap.Orders))
	for _, o := range snap.Orders {
		if o.OrderID != id {
			filtered = append(filtered, o)
		}
	}
	snap.Orders = filtered

	return r.storage.Save(snap)
}

func (r *snapshotOrderRepository) List(filter requests.ListOrdersFilter) ([]models.Order, error) {
	snap, err := r.storage.Load()
	if err != nil {
		return nil, err
	}

	sort.Slice(snap.Orders, func(i, j int) bool {
		return snap.Orders[i].OrderID.String() < snap.Orders[j].OrderID.String()
	})

	var result []models.Order
	for _, o := range snap.Orders {
		if o.UserID != filter.UserID {
			continue
		}

		if filter.LastID != nil && o.OrderID.String() <= filter.LastID.String() {
			continue
		}

		if filter.InPvz != nil && *filter.InPvz {
			if o.Status == models.Issued {
				continue
			}
		}

		result = append(result, o)
	}

	if filter.Page != nil && filter.Limit != nil {
		start := (*filter.Page - 1) * (*filter.Limit)
		end := start + *filter.Limit
		if start >= len(result) {
			return []models.Order{}, nil
		}

		if end > len(result) {
			end = len(result)
		}

		return result[start:end], nil
	}

	limit := constants.DefaultLimit
	if filter.Limit != nil {
		limit = *filter.Limit
	}

	if len(result) > limit {
		result = result[:limit]
	}

	return result, nil
}
