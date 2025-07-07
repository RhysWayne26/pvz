package gateway

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	pb "pvz-cli/internal/gen/admin"
	"pvz-cli/internal/workerpool"
)

type workerStats struct {
	activeWorkers int32
	queueSize     int
	totalTasks    int64
	failedTasks   int64
	isShutdown    bool
}

// AdminGRPCRouter is a gRPC server implementation for managing worker pool settings and retrieving statistics.
type AdminGRPCRouter struct {
	pb.UnimplementedAdminServiceServer
	pool workerpool.WorkerPool
}

// NewAdminGRPCRouter creates a new instance of AdminGRPCRouter with the provided worker pool for managing worker operations.
func NewAdminGRPCRouter(pool workerpool.WorkerPool) *AdminGRPCRouter {
	return &AdminGRPCRouter{
		pool: pool,
	}
}

// SetWorkerCount adjusts the number of active workers in the worker pool and returns the old and new worker counts.
func (r *AdminGRPCRouter) SetWorkerCount(
	ctx context.Context,
	req *pb.SetWorkerCountRequest,
) (*pb.SetWorkerCountResponse, error) {
	if req.Count == 0 {
		return nil, status.Error(codes.InvalidArgument, "worker count must be greater than 0")
	}
	stats := r.pool.GetStats()
	oldCount, ok := stats["worker_count"].(int32)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get current worker count")
	}
	r.pool.SetWorkerCount(int(req.Count))

	return &pb.SetWorkerCountResponse{
		Status:   "success",
		OldCount: uint32(oldCount),
		NewCount: req.Count,
	}, nil
}

// GetWorkerStats retrieves the worker pool statistics including active workers, queued tasks, total tasks, and failure count.
func (r *AdminGRPCRouter) GetWorkerStats(
	ctx context.Context,
	req *pb.GetWorkerStatsRequest,
) (*pb.GetWorkerStatsResponse, error) {
	stats := r.pool.GetStats()
	ws, err := r.parseStats(stats)
	if err != nil {
		return nil, err
	}
	return &pb.GetWorkerStatsResponse{
		ActiveWorkers: uint32(ws.activeWorkers),
		QueuedTasks:   uint32(ws.queueSize),
		TotalTasks:    uint64(ws.totalTasks),
		FailedTasks:   uint64(ws.failedTasks),
		IsShutdown:    ws.isShutdown,
	}, nil
}

func (r *AdminGRPCRouter) parseStats(stats map[string]interface{}) (*workerStats, error) {
	activeWorkers, ok := stats["worker_count"].(int32)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get worker count")
	}
	queueSize, ok := stats["queue_size"].(int)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get queue size")
	}
	totalTasks, ok := stats["total_tasks"].(int64)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get total tasks")
	}
	failedTasks, ok := stats["failed_tasks"].(int64)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get failed tasks")
	}
	isShutdown, ok := stats["is_shutdown"].(bool)
	if !ok {
		return nil, status.Error(codes.Internal, "failed to get shutdown status")
	}
	return &workerStats{
		activeWorkers: activeWorkers,
		queueSize:     queueSize,
		totalTasks:    totalTasks,
		failedTasks:   failedTasks,
		isShutdown:    isShutdown,
	}, nil
}
