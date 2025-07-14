package services

import (
	"context"
	"encoding/json"
	"github.com/jackc/pgx/v5"
	"pvz-cli/internal/common/apperrors"
	"pvz-cli/internal/common/utils"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/infrastructure/db"
	"pvz-cli/internal/models"
	"pvz-cli/internal/usecases/requests"
	"pvz-cli/internal/usecases/services/validators"
	"pvz-cli/internal/workerpool"
	"pvz-cli/pkg/clock"
	"sync"
)

const SourceName = "pvz-api"

var _ OrderService = (*DefaultOrderService)(nil)

// DefaultOrderService is a default implementation of the OrderService interface
type DefaultOrderService struct {
	clk               clock.Clock
	pool              workerpool.WorkerPool
	txRunner          db.TxRunner
	orderRepo         repositories.OrderRepository
	outboxRepo        repositories.OutboxRepository
	packagePricingSvc PackagePricingService
	historySvc        HistoryService
	actorSvc          ActorService
	validator         validators.OrderValidator
}

// NewDefaultOrderService creates a new instance of DefaultOrderService
func NewDefaultOrderService(
	clk clock.Clock,
	pool workerpool.WorkerPool,
	txRunner db.TxRunner,
	orderRepo repositories.OrderRepository,
	outboxRepo repositories.OutboxRepository,
	packagePricingService PackagePricingService,
	historyService HistoryService,
	actorSvc ActorService,
	validator validators.OrderValidator) *DefaultOrderService {
	return &DefaultOrderService{
		clk:               clk,
		pool:              pool,
		txRunner:          txRunner,
		orderRepo:         orderRepo,
		outboxRepo:        outboxRepo,
		packagePricingSvc: packagePricingService,
		historySvc:        historyService,
		actorSvc:          actorSvc,
		validator:         validator,
	}
}

// AcceptOrder accepts an order with package pricing calculation and validation
func (s *DefaultOrderService) AcceptOrder(ctx context.Context, req requests.AcceptOrderRequest) (models.Order, error) {
	if ctx.Err() != nil {
		return models.Order{}, ctx.Err()
	}
	existing, err := s.orderRepo.Load(ctx, req.OrderID)
	if err != nil {
		existing = models.Order{}
	}

	if err := s.validator.ValidateAccept(existing, req); err != nil {
		return models.Order{}, err
	}

	totalPrice, err := s.packagePricingSvc.Evaluate(req.Package, req.Weight, req.Price)
	if err != nil {
		return models.Order{}, err
	}

	now := s.clk.Now()

	order := models.Order{
		OrderID:         req.OrderID,
		UserID:          req.UserID,
		CreatedAt:       now,
		UpdatedStatusAt: now,
		Status:          models.Accepted,
		ExpiresAt:       req.ExpiresAt,
		Weight:          req.Weight,
		Price:           totalPrice,
		Package:         req.Package,
	}

	actor, err := s.actorSvc.DetermineActor(ctx, models.EventAccepted, order.UserID)
	if err != nil {
		return models.Order{}, err
	}

	eventID, err := s.generateEventID(order.OrderID)
	if err != nil {
		return models.Order{}, err
	}
	event := models.KafkaEvent{
		EventID:   eventID,
		EventType: models.MapEventTypeToKafkaEvent(models.EventAccepted),
		Timestamp: now,
		Actor:     actor,
		Order:     order,
		Source:    SourceName,
	}

	payloadBytes, err := marshalEvent(event)
	if err != nil {
		return models.Order{}, err
	}
	entry := models.HistoryEntry{
		OrderID:   order.OrderID,
		Event:     models.EventAccepted,
		Timestamp: now,
	}

	err = s.txRunner.WithTx(ctx, func(tx pgx.Tx) error {
		txCtx := ctxWithTx(ctx, tx)
		if err := s.orderRepo.Save(txCtx, order); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to save order %d: %v", req.OrderID, err)
		}
		if err := s.outboxRepo.Create(txCtx, eventID, order.OrderID, payloadBytes); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to enqueue outbox event: %v", err)
		}
		if err := s.historySvc.Record(txCtx, entry); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to record history entry for order %d: %v", req.OrderID, err)
		}
		return nil
	})

	if err != nil {
		return models.Order{}, err
	}
	return order, nil
}

// IssueOrders processes multiple orders for issuance to clients
func (s *DefaultOrderService) IssueOrders(
	ctx context.Context,
	req requests.IssueOrdersRequest,
) ([]models.BatchEntryProcessedResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := len(req.OrderIDs)
	results := make([]models.BatchEntryProcessedResult, n)
	var wg sync.WaitGroup
	for i, id := range req.OrderIDs {
		wg.Add(1)
		i, id := i, id
		s.pool.Submit(func() {
			defer wg.Done()
			res := models.BatchEntryProcessedResult{OrderID: id}
			order, err := s.orderRepo.Load(ctx, id)
			if err != nil {
				res.Error = apperrors.Newf(apperrors.OrderNotFound, "order %d not found", id)
				results[i] = res
				return
			}
			if err := s.validator.ValidateIssue(order, req); err != nil {
				res.Error = err
				results[i] = res
				return
			}
			eventID, err := s.generateEventID(order.OrderID)
			if err != nil {
				res.Error = err
				results[i] = res
				return
			}
			now := s.clk.Now()
			actor, err := s.actorSvc.DetermineActor(ctx, models.EventIssued, order.UserID)
			if err != nil {
				res.Error = err
				results[i] = res
				return
			}
			order.Status = models.Issued
			order.UpdatedStatusAt = now
			evt := models.KafkaEvent{
				EventID:   eventID,
				EventType: models.MapEventTypeToKafkaEvent(models.EventIssued),
				Timestamp: now,
				Actor:     actor,
				Order:     order,
				Source:    SourceName,
			}
			payloadBytes, err := marshalEvent(evt)
			if err != nil {
				res.Error = err
				results[i] = res
				return
			}
			entry := models.HistoryEntry{
				OrderID:   id,
				Event:     models.EventIssued,
				Timestamp: now,
			}
			err = s.txRunner.WithTx(ctx, func(tx pgx.Tx) error {
				txCtx := ctxWithTx(ctx, tx)
				if err := s.orderRepo.Save(txCtx, order); err != nil {
					return apperrors.Newf(apperrors.InternalError, "failed to save order %d: %v", id, err)
				}
				if err := s.outboxRepo.Create(txCtx, eventID, id, payloadBytes); err != nil {
					return apperrors.Newf(apperrors.InternalError, "failed to enqueue issue-event for order %d: %v", id, err)
				}
				if err := s.historySvc.Record(txCtx, entry); err != nil {
					return apperrors.Newf(apperrors.InternalError, "failed to record history entry for order %d: %v", id, err)
				}
				return nil
			})
			if err != nil {
				res.Error = err
				results[i] = res
				return
			}
			results[i] = res
		})
	}

	wg.Wait()
	return results, nil
}

// ListOrders retrieves filtered and paginated list of orders
func (s *DefaultOrderService) ListOrders(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, uint64, int, error) {
	if ctx.Err() != nil {
		return nil, 0, 0, ctx.Err()
	}
	result, total, err := s.orderRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, 0, apperrors.Newf(apperrors.InternalError, "failed to list orders: %v", err)
	}

	if filter.Last != nil {
		n := *filter.Last
		if len(result) > n {
			result = result[len(result)-n:]
		}
		total = len(result)
	}

	var nextLastID uint64
	if len(result) > 0 {
		nextLastID = result[len(result)-1].OrderID
	}

	return result, nextLastID, total, nil
}

// CreateClientReturns processes multiple client return requests
func (s *DefaultOrderService) CreateClientReturns(
	ctx context.Context,
	req requests.ClientReturnsRequest,
) ([]models.BatchEntryProcessedResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := len(req.OrderIDs)
	results := make([]models.BatchEntryProcessedResult, n)
	var wg sync.WaitGroup
	for i, id := range req.OrderIDs {
		wg.Add(1)
		i, id := i, id
		s.pool.Submit(func() {
			defer wg.Done()
			res := models.BatchEntryProcessedResult{OrderID: id}
			order, err := s.orderRepo.Load(ctx, id)
			if err != nil {
				res.Error = apperrors.Newf(apperrors.OrderNotFound, "order %d not found", id)
				results[i] = res
				return
			}
			if err := s.validator.ValidateClientReturn(order, req); err != nil {
				res.Error = err
				results[i] = res
				return
			}
			now := s.clk.Now()
			actor, err := s.actorSvc.DetermineActor(ctx, models.EventReturnedByClient, order.UserID)
			if err != nil {
				res.Error = err
				results[i] = res
				return
			}
			order.Status = models.Returned
			order.UpdatedStatusAt = now
			eventID, err := s.generateEventID(order.OrderID)
			if err != nil {
				res.Error = err
				results[i] = res
				return
			}
			event := models.KafkaEvent{
				EventID:   eventID,
				EventType: models.MapEventTypeToKafkaEvent(models.EventReturnedByClient),
				Timestamp: now,
				Actor:     actor,
				Order:     order,
				Source:    SourceName,
			}
			payloadBytes, err := marshalEvent(event)
			if err != nil {
				res.Error = err
				results[i] = res
				return
			}
			entry := models.HistoryEntry{
				OrderID:   id,
				Event:     models.EventReturnedByClient,
				Timestamp: now,
			}
			err = s.txRunner.WithTx(ctx, func(tx pgx.Tx) error {
				txCtx := ctxWithTx(ctx, tx)
				if err := s.orderRepo.Save(txCtx, order); err != nil {
					return apperrors.Newf(apperrors.InternalError, "failed to save returned order %d: %v", id, err)
				}
				if err := s.outboxRepo.Create(txCtx, eventID, id, payloadBytes); err != nil {
					return apperrors.Newf(apperrors.InternalError, "failed to enqueue return-event for order %d: %v", id, err)
				}
				if err := s.historySvc.Record(txCtx, entry); err != nil {
					return apperrors.Newf(apperrors.InternalError, "failed to record history entry for order %d: %v", id, err)
				}
				return nil
			})
			if err != nil {
				res.Error = err
				results[i] = res
				return
			}
			results[i] = res
		})
	}

	wg.Wait()
	return results, nil
}

// ReturnToCourier processes return of order back to courier/warehouse
func (s *DefaultOrderService) ReturnToCourier(ctx context.Context, req requests.ReturnOrderRequest) error {
	if ctx.Err() != nil {
		return ctx.Err()
	}
	orderID := req.OrderID
	o, err := s.orderRepo.Load(ctx, orderID)
	if err != nil {
		return apperrors.Newf(apperrors.OrderNotFound, "order %d not found", orderID)
	}

	if err := s.validator.ValidateReturnToCourier(o); err != nil {
		return err
	}

	now := s.clk.Now()
	actor, err := s.actorSvc.DetermineActor(ctx, models.EventReturnedToWarehouse, o.UserID)
	if err != nil {
		return err
	}
	eventID, err := s.generateEventID(o.OrderID)
	if err != nil {
		return err
	}
	event := models.KafkaEvent{
		EventID:   eventID,
		EventType: models.MapEventTypeToKafkaEvent(models.EventReturnedToWarehouse),
		Timestamp: now,
		Actor:     actor,
		Order:     o,
		Source:    SourceName,
	}
	payloadBytes, err := marshalEvent(event)
	if err != nil {
		return err
	}
	entry := models.HistoryEntry{
		OrderID:   orderID,
		Event:     models.EventReturnedToWarehouse,
		Timestamp: now,
	}

	err = s.txRunner.WithTx(ctx, func(tx pgx.Tx) error {
		txCtx := ctxWithTx(ctx, tx)
		if err := s.orderRepo.Delete(txCtx, orderID); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to delete order %d: %v", orderID, err)
		}
		if err := s.outboxRepo.Create(txCtx, eventID, orderID, payloadBytes); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to enqueue return-event: %v", err)
		}
		if err := s.historySvc.Record(txCtx, entry); err != nil {
			return apperrors.Newf(apperrors.InternalError, "failed to record history entry for order %d: %v", orderID, err)
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// ListReturns retrieves paginated list of return entries sorted by return date
func (s *DefaultOrderService) ListReturns(ctx context.Context, filter requests.OrdersFilterRequest) ([]models.Order, error) {
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	orders, _, err := s.orderRepo.List(ctx, filter)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to list returns: %v", err)
	}

	return orders, nil
}

// ImportOrders imports multiple orders concurrently, processing each status and returning a batch of results with errors, if any.
func (s *DefaultOrderService) ImportOrders(
	ctx context.Context,
	req requests.ImportOrdersRequest,
) ([]models.BatchEntryProcessedResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	n := len(req.Statuses)
	results := make([]models.BatchEntryProcessedResult, n)
	var wg sync.WaitGroup
	for i, st := range req.Statuses {
		if st.Error != nil {
			results[i] = models.BatchEntryProcessedResult{OrderID: st.Request.OrderID, Error: st.Error}
			continue
		}
		wg.Add(1)
		i, st := i, st
		s.pool.Submit(func() {
			defer wg.Done()
			order, err := s.AcceptOrder(ctx, *st.Request)
			results[i] = models.BatchEntryProcessedResult{OrderID: order.OrderID, Error: err}
		})
	}
	wg.Wait()
	return results, nil
}

func marshalEvent(e models.KafkaEvent) ([]byte, error) {
	payloadBytes, err := json.Marshal(e)
	if err != nil {
		return nil, apperrors.Newf(apperrors.InternalError, "failed to marshal event: %v", err)
	}
	return payloadBytes, nil
}

func ctxWithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return db.WithTxContext(ctx, tx)
}

func (s *DefaultOrderService) generateEventID(orderID uint64) (uint64, error) {
	eventID, err := utils.GenerateID()
	if err != nil {
		return 0, apperrors.Newf(apperrors.InternalError, "failed to save event for order %d: %v", orderID, err)
	}
	return eventID, nil
}
