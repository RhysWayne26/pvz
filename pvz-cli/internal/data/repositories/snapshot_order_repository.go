package repositories

import (
	"errors"
	"sort"
	"time"

	"pvz-cli/internal/constants"
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
	orders, err := r.loadAndSort()
	if err != nil {
		return nil, 0, err
	}
	lastCreatedAt := findLastCreatedAt(orders, filter.LastID)
	var filters []orderFilter
	filters = append(filters, filterByUser(filter.UserID))
	if filter.LastID != "" {
		filters = append(filters, filterByLastID(lastCreatedAt))
	}
	if filter.InPvz != nil {
		filters = append(filters, filterByInPvz(filter.InPvz))
	}
	filtered := applyFilters(orders, filters...)
	total := len(filtered)
	page := constants.DefaultPage
	limit := constants.DefaultLimit
	if filter.Page != nil {
		page = *filter.Page
	}
	if filter.Limit != nil {
		limit = *filter.Limit
	}
	paged := paginate(filtered, page, limit)
	return paged, total, nil
}

func (r *snapshotOrderRepository) loadAndSort() ([]models.Order, error) {
	snap, err := r.storage.Load()
	if err != nil {
		return nil, err
	}
	sort.Slice(snap.Orders, func(i, j int) bool {
		return snap.Orders[i].CreatedAt.Before(snap.Orders[j].CreatedAt)
	})
	return snap.Orders, nil
}

func findLastCreatedAt(orders []models.Order, lastID string) time.Time {
	for _, o := range orders {
		if o.OrderID == lastID {
			return o.CreatedAt
		}
	}
	return time.Time{}
}

type orderFilter func(models.Order) bool

func filterByUser(userID string) orderFilter {
	return func(o models.Order) bool {
		return o.UserID == userID
	}
}

func filterByLastID(ts time.Time) orderFilter {
	return func(o models.Order) bool {
		return o.CreatedAt.After(ts)
	}
}

func filterByInPvz(inPvz *bool) orderFilter {
	return func(o models.Order) bool {
		if inPvz != nil && *inPvz && o.Status == models.Issued {
			return false
		}
		return true
	}
}

func applyFilters(orders []models.Order, filters ...orderFilter) []models.Order {
	var out []models.Order
	for _, o := range orders {
		keep := true
		for _, f := range filters {
			if !f(o) {
				keep = false
				break
			}
		}
		if keep {
			out = append(out, o)
		}
	}
	return out
}

func paginate(orders []models.Order, page, limit int) []models.Order {
	start := (page - 1) * limit
	if start >= len(orders) {
		return []models.Order{}
	}
	end := start + limit
	if end > len(orders) {
		end = len(orders)
	}
	return orders[start:end]
}
