package app

import (
	"log/slog"
	"os"
	"pvz-cli/infrastructure/brokers"
	"pvz-cli/infrastructure/db"
	"pvz-cli/internal/common/config"
	"pvz-cli/internal/data/repositories"
	"pvz-cli/internal/data/storage"
	"pvz-cli/internal/usecases/handlers"
	"pvz-cli/internal/usecases/services"
	"pvz-cli/internal/usecases/services/strategies"
	"pvz-cli/internal/usecases/services/validators"
	"pvz-cli/internal/workerpool"
	"pvz-cli/internal/workers"
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
}

// NewContainer returns a new instance of an application container
func NewContainer(pool workerpool.WorkerPool) *Container {
	cfg := config.Load()
	var (
		orderRepo   repositories.OrderRepository
		historyRepo repositories.HistoryRepository
		txRunner    db.TxRunner
		outboxRepo  repositories.OutboxRepository
	)

	c := &Container{
		config: cfg,
	}

	switch {
	case cfg.DB != nil && cfg.DB.WriteDSN != "":
		client, err := db.NewDefaultPGXClient(cfg.DB.ReadDSN, cfg.DB.WriteDSN)
		if err != nil {
			slog.Error("failed to init DB client", "error", err)
			os.Exit(1)
		}
		client.SetConnectionSettings(20, 10, time.Hour, 30*time.Minute)
		txRunner = client
		orderRepo = repositories.NewPGOrderRepository(client)
		historyRepo = repositories.NewPGHistoryRepository(client)
		outboxRepo = repositories.NewPGOutboxRepository(client)
		producer := brokers.NewDefaultProducer(cfg.Kafka.Brokers)
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
	historySvc := services.NewDefaultHistoryService(historyRepo)
	pricingSvc := services.NewDefaultPackagePricingService(packageValidator, pricingStrategy)
	orderSvc := services.NewDefaultOrderService(clk, pool, txRunner, orderRepo, outboxRepo, pricingSvc, historySvc, actorSvc, orderValidator)

	facadeHandler := handlers.NewDefaultFacadeHandler(orderSvc, historySvc)

	c.orderService = orderSvc
	c.historyService = historySvc
	c.facadeHandler = facadeHandler
	return c
}

func (c *Container) shutdownOutbox() {
	if c.outboxDispatcher != nil {
		c.outboxDispatcher.Stop()
	}
	if c.kafkaProducer != nil {
		if err := c.kafkaProducer.Close(); err != nil {
			slog.Error("failed to close Kafka producer", "error", err)
		}
	}
}
