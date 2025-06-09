package gateway

import (
	"context"

	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/grpc/mappers"
	"pvz-cli/internal/usecases/handlers"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// GRPCRouter implements the gRPC OrdersService server. It maps gRPC requests to internal handlers and converts responses back to protobuf format.
type GRPCRouter struct {
	pb.UnimplementedOrdersServiceServer
	facadeHandler handlers.FacadeHandler
	facadeMapper  mappers.GRPCFacadeMapper
}

// NewGRPCRouter returns a new instance of GRPCRouter
func NewGRPCRouter(
	m mappers.GRPCFacadeMapper,
	facade handlers.FacadeHandler,
) *GRPCRouter {
	return &GRPCRouter{
		facadeMapper:  m,
		facadeHandler: facade,
	}
}

// AcceptOrder handles the AcceptOrder gRPC request and delegates to the facade handler.
func (r *GRPCRouter) AcceptOrder(
	ctx context.Context,
	req *pb.AcceptOrderRequest,
) (*pb.OrderResponse, error) {
	dto, err := r.facadeMapper.FromPbAcceptOrderRequest(req)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	res, err := r.facadeHandler.HandleAcceptOrder(ctx, dto)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return r.facadeMapper.ToPbAcceptOrderResponse(res), nil
}

// ReturnOrder handles the ReturnOrder gRPC request and delegates to the facade handler.
func (r *GRPCRouter) ReturnOrder(
	ctx context.Context,
	req *pb.OrderIdRequest,
) (*pb.OrderResponse, error) {
	dto := r.facadeMapper.FromPbReturnOrderRequest(req)

	res, err := r.facadeHandler.HandleReturnOrder(ctx, dto)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return r.facadeMapper.ToPbReturnOrderResponse(res), nil
}

// ProcessOrders handles the ProcessOrders gRPC request and delegates to the facade handler.
func (r *GRPCRouter) ProcessOrders(
	ctx context.Context,
	req *pb.ProcessOrdersRequest,
) (*pb.ProcessResult, error) {
	dto := r.facadeMapper.FromPbProcessOrdersRequest(req)

	resp, err := r.facadeHandler.HandleProcessOrders(ctx, dto)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return r.facadeMapper.ToPbProcessResult(resp), nil
}

// ListOrders handles the ListOrders gRPC request and delegates to the facade handler.
func (r *GRPCRouter) ListOrders(
	ctx context.Context,
	req *pb.ListOrdersRequest,
) (*pb.OrdersList, error) {
	dto := r.facadeMapper.FromPbListOrdersRequest(req)

	resp, err := r.facadeHandler.HandleListOrders(ctx, dto)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return r.facadeMapper.ToPbOrdersList(resp), nil
}

// ListReturns handles the ListReturns gRPC request and delegates to the facade handler.
func (r *GRPCRouter) ListReturns(
	ctx context.Context,
	req *pb.ListReturnsRequest,
) (*pb.ReturnsList, error) {
	dto := r.facadeMapper.FromPbListReturnsRequest(req)

	resp, err := r.facadeHandler.HandleListOrders(ctx, dto)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return r.facadeMapper.ToPbReturnsList(resp), nil
}

// GetHistory handles the GetHistory gRPC request and delegates to the facade handler.
func (r *GRPCRouter) GetHistory(
	ctx context.Context,
	req *pb.GetHistoryRequest,
) (*pb.OrderHistoryList, error) {
	resp, err := r.facadeHandler.HandleOrderHistory(ctx)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return r.facadeMapper.ToPbOrderHistoryList(resp), nil
}

// ImportOrders handles the ImportOrders gRPC request and delegates to the facade handler.
func (r *GRPCRouter) ImportOrders(
	ctx context.Context,
	req *pb.ImportOrdersRequest,
) (*pb.ImportResult, error) {
	dto := r.facadeMapper.FromPbImportOrdersRequest(req)

	resp, err := r.facadeHandler.HandleImportOrders(ctx, dto)
	if err != nil {
		return nil, toGRPCError(err)
	}

	return r.facadeMapper.ToPbImportResult(resp), nil
}

func toGRPCError(err error) error {
	if s, ok := status.FromError(err); ok && s.Code() != codes.Unknown {
		return err
	}
	return status.Errorf(codes.InvalidArgument, err.Error())
}
