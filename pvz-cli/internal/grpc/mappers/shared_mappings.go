package mappers

import (
	pb "pvz-cli/internal/gen/orders"
	"pvz-cli/internal/models"

	"google.golang.org/protobuf/types/known/timestamppb"
)

const unknownPackage = "unknown"

func toPbOrder(o models.Order) *pb.Order {
	return &pb.Order{
		OrderId:    o.OrderID,
		UserId:     o.UserID,
		Status:     toPbOrderStatus(o.Status),
		ExpiresAt:  timestamppb.New(o.ExpiresAt),
		Weight:     o.Weight,
		TotalPrice: o.Price,
		Package:    toPbPackageTypePtr(o.Package),
	}
}

func toPbOrderStatus(s models.OrderStatus) pb.OrderStatus {
	switch s {
	case models.Accepted:
		return pb.OrderStatus_ORDER_STATUS_ACCEPTED
	case models.Returned:
		return pb.OrderStatus_ORDER_STATUS_RETURNED
	case models.Issued:
		return pb.OrderStatus_ORDER_STATUS_ISSUED
	default:
		return pb.OrderStatus_ORDER_STATUS_UNSPECIFIED
	}
}

func toPbPackageType(p models.PackageType) pb.PackageType {
	switch p {
	case models.PackageBag:
		return pb.PackageType_PACKAGE_TYPE_BAG
	case models.PackageBox:
		return pb.PackageType_PACKAGE_TYPE_BOX
	case models.PackageFilm:
		return pb.PackageType_PACKAGE_TYPE_TAPE
	case models.PackageBagFilm:
		return pb.PackageType_PACKAGE_TYPE_BAG_TAPE
	case models.PackageBoxFilm:
		return pb.PackageType_PACKAGE_TYPE_BOX_TAPE
	default:
		return pb.PackageType_PACKAGE_TYPE_UNSPECIFIED
	}
}

func toPbPackageTypePtr(p models.PackageType) *pb.PackageType {
	if p == models.PackageNone {
		return nil
	}
	val := toPbPackageType(p)
	return &val
}

func toPbEventType(e models.EventType) pb.EventType {
	switch e {
	case models.EventAccepted:
		return pb.EventType_EVENT_ACCEPTED
	case models.EventIssued:
		return pb.EventType_EVENT_ISSUED
	case models.EventReturnedFromClient:
		return pb.EventType_EVENT_RETURNED_FROM_CLIENT
	case models.EventReturnedToWarehouse:
		return pb.EventType_EVENT_RETURNED_TO_WAREHOUSE
	default:
		return pb.EventType_EVENT_UNSPECIFIED
	}
}

func fromPbPackageType(p pb.PackageType) models.PackageType {
	switch p {
	case pb.PackageType_PACKAGE_TYPE_UNSPECIFIED:
		return models.PackageNone
	case pb.PackageType_PACKAGE_TYPE_BAG:
		return models.PackageBag
	case pb.PackageType_PACKAGE_TYPE_BOX:
		return models.PackageBox
	case pb.PackageType_PACKAGE_TYPE_TAPE:
		return models.PackageFilm
	case pb.PackageType_PACKAGE_TYPE_BAG_TAPE:
		return models.PackageBagFilm
	case pb.PackageType_PACKAGE_TYPE_BOX_TAPE:
		return models.PackageBoxFilm
	default:
		return unknownPackage
	}
}
