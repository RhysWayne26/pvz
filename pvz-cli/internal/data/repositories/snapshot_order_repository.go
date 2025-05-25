package repositories

import (
	"errors"
	"pvz-cli/internal/constants"
	"sort"
	"time"

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

func (r *snapshotOrderRepository) Load(id string) (models.Order, error) {
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

func (r *snapshotOrderRepository) Delete(id string) error {
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

func (r *snapshotOrderRepository) List(filter requests.ListOrdersFilter) ([]models.Order, int, error) {
	snap, err := r.storage.Load()
	if err != nil {
		return nil, 0, err
	}

	sort.Slice(snap.Orders, func(i, j int) bool {
		return snap.Orders[i].CreatedAt.Before(snap.Orders[j].CreatedAt)
	})

	var lastCreatedAt time.Time
	if filter.LastID != "" {
		for _, o := range snap.Orders {
			if o.OrderID == filter.LastID {
				lastCreatedAt = o.CreatedAt
				break
			}
		}
	}

	var result []models.Order
	for _, o := range snap.Orders {
		if o.UserID != filter.UserID {
			continue
		}

		if filter.LastID != "" && !o.CreatedAt.After(lastCreatedAt) {
			continue
		}

		if filter.InPvz != nil && *filter.InPvz && o.Status == models.Issued {
			continue
		}

		result = append(result, o)
	}

	total := len(result)
	if filter.Page != nil && filter.Limit != nil {
		start := (*filter.Page - 1) * (*filter.Limit)
		end := start + *filter.Limit
		if start >= len(result) {
			return []models.Order{}, total, nil
		}
		if end > len(result) {
			end = len(result)
		}
		return result[start:end], total, nil
	}

	limit := constants.DefaultLimit
	if filter.Limit != nil {
		limit = *filter.Limit
	}
	if len(result) > limit {
		result = result[:limit]
	}

	return result, total, nil
}
