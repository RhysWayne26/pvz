package app

import (
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"log/slog"
	"os"
	"pvz-cli/internal/common/config"
	"pvz-cli/internal/common/constants"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/infrastructure/brokers"
	"pvz-cli/internal/infrastructure/db"
	"pvz-cli/internal/metrics"
	"pvz-cli/internal/usecases/handlers"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/usecases/services/decorators"
	"pvz-cli/internal/usecases/services/strategies"
	"pvz-cli/internal/usecases/services/validators"
	"pvz-cli/internal/workerpool"
	"pvz-cli/internal/workers"
	"pvz-cli/pkg/cache"
	"pvz-cli/pkg/cache/policies"
	"pvz-cli/pkg/clock"
	"time"
)

// Container holds all shared business-level dependencies: configuration, repositories, services, and the facade handler.
type Container struct {
	config           *config.Config
	orderService     services.OrderService
	historyService   services.HistoryService
	facadeHandler    handlers.FacadeHandler
	outboxDispatcher *workers.DefaultOutboxDispatcher
	kafkaProducer    brokers.KafkaProducer
	responseCache    cache.Cache[string, any]
}

// NewContainer returns a new instance of an application container
func NewContainer(pool workerpool.WorkerPool) *Container {
	cfg := config.Load()
	var (
		orderRepo   repositories.OrderRepository
		historyRepo repositories.HistoryRepository
		txRunner    db.TxRunner
		outboxRepo  repositories.OutboxRepository
		producer    brokers.KafkaProducer
	)

	c := &Container{
		config: cfg,
	}

	tracer := otel.Tracer("pvz")
	switch {
	case cfg.DB != nil && cfg.DB.WriteDSN != "":
		client, err := db.NewDefaultPGXClient(cfg.DB.ReadDSN, cfg.DB.WriteDSN)
		if err != nil {
			slog.Error("failed to init DB client", "error", err)
			os.Exit(1)
		}
		client.SetConnectionSettings(20, 10, time.Hour, 30*time.Minute)
		txRunner = db.NewTracingTxRunner(client, tracer)
		orderRepo = repositories.NewPGOrderRepository(client)
		historyRepo = repositories.NewPGHistoryRepository(client)
		if cfg.Outbox != nil && cfg.Outbox.BatchSize > 0 {
			outboxRepo = repositories.NewPGOutboxRepository(client)
			producer, err = brokers.NewKafkaAsyncProducer(cfg.Kafka.Brokers)
			if err != nil {
				slog.Error("failed to init Kafka producer", "error", err)
				os.Exit(1)
			}
		} else {
			slog.Info("outbox disabled (BatchSize=0), events will be dropped")
			outboxRepo = repositories.NewNoOpOutboxRepository()
			producer = brokers.NewKafkaNoOpProducer()
		}
		dispatcher := workers.NewDefaultOutboxDispatcher(
			outboxRepo,
			producer,
			cfg.Kafka.Topic,
			cfg.Outbox.BatchSize,
			time.Duration(cfg.Outbox.RetryDelaySec)*time.Second,
			time.Duration(cfg.Outbox.PollIntervalSec)*time.Second,
		)
		c.outboxDispatcher = dispatcher
		c.kafkaProducer = producer

	case cfg.File != nil && cfg.File.Path != "":
		fileStorage := storage.NewJSONStorage(cfg.File.Path)
		orderRepo = repositories.NewSnapshotOrderRepository(fileStorage)
		historyRepo = repositories.NewSnapshotHistoryRepository(fileStorage)
		outboxRepo = repositories.NewNoOpOutboxRepository()
		txRunner = db.NewNoOpTxRunner()

	default:
		slog.Error("No valid storage config provided")
		os.Exit(1)
	}

	clk := &clock.RealClock{}

	orderValidator := validators.NewDefaultOrderValidator(clk)
	packageValidator := validators.NewDefaultPackageValidator()
	pricingStrategy := strategies.NewDefaultPricingStrategy()

	actorSvc := services.NewDefaultActorService()
	baseHistorySvc := services.NewDefaultHistoryService(historyRepo)
	historySvc := decorators.NewTracingHistoryService(baseHistorySvc, tracer)
	pricingSvc := services.NewDefaultPackagePricingService(packageValidator, pricingStrategy)
	baseOrderSvc := services.NewDefaultOrderService(clk, pool, txRunner, orderRepo, outboxRepo, pricingSvc, historySvc, actorSvc, orderValidator)
	orderSvc := decorators.NewTracingOrderService(baseOrderSvc, tracer)
	responsesCache := cache.NewInMemoryShardedCache[string, any](
		constants.CacheShardsCount,
		policies.NewTTLPolicy[string, any](),
		"pvz",
		"cache",
		prometheus.DefaultRegisterer,
	)
	handlerMetrics, err := metrics.NewDefaultHandlerMetrics(prometheus.DefaultRegisterer)
	if err != nil {
		slog.Error("failed to init handler metrics", "error", err)
		os.Exit(1)
	}

	facadeHandler := handlers.NewDefaultFacadeHandler(orderSvc, historySvc, responsesCache, handlerMetrics)

	c.orderService = orderSvc
	c.historyService = historySvc
	c.facadeHandler = facadeHandler
	c.responseCache = responsesCache
	return c
}

func (c *Container) shutdownOutbox() {
	if c.outboxDispatcher != nil {
		c.outboxDispatcher.Stop()
	}
	if c.kafkaProducer != nil {
		err := c.kafkaProducer.Close()
		if err != nil {
			slog.Error("failed to close Kafka producer", "error", err)
		}
	}
}
