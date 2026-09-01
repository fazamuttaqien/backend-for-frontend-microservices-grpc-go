package grpc

import (
	"context"
	"errors"
	orderv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/order/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/order/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	orderv1.UnimplementedOrderServiceServer
	app *application.OrderService
}

func NewHandler(app *application.OrderService) *Handler { return &Handler{app: app} }
func response(o *domain.Order) *orderv1.Order {
	items := make([]*orderv1.OrderItem, 0, len(o.Items))
	for _, item := range o.Items {
		items = append(items, &orderv1.OrderItem{ProductId: item.ProductID, Quantity: item.Quantity, Price: item.Price, Total: item.Total})
	}
	return &orderv1.Order{Id: o.ID, UserId: o.UserID, Items: items, Total: o.Total, Status: o.Status, CreatedAt: o.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"), UpdatedAt: o.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")}
}
func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrOrderNotFound), errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidOrder):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, err.Error())
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, err.Error())
	case errors.Is(err, domain.ErrDependencyUnavailable):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
func claims(ctx context.Context) (*auth.Claims, error) {
	c, ok := auth.ClaimsFromContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "authentication required")
	}
	return c, nil
}
func (h *Handler) CreateOrder(ctx context.Context, in *orderv1.CreateOrderRequest) (*orderv1.CreateOrderResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	c, err := claims(ctx)
	if err != nil {
		return nil, err
	}
	if in.UserId == "" || in.UserId != c.Subject {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	items := make([]application.CreateItem, 0, len(in.Items))
	if len(in.Items) == 0 || len(in.Items) > 100 {
		return nil, status.Error(codes.InvalidArgument, "invalid items")
	}
	for _, item := range in.Items {
		if item == nil {
			return nil, status.Error(codes.InvalidArgument, "order item is required")
		}
		items = append(items, application.CreateItem{ProductID: item.ProductId, Quantity: item.Quantity})
	}
	o, err := h.app.Create(ctx, in.UserId, items)
	if err != nil {
		return nil, mapErr(err)
	}
	return &orderv1.CreateOrderResponse{Order: response(o)}, nil
}
func (h *Handler) GetOrder(ctx context.Context, in *orderv1.GetOrderRequest) (*orderv1.GetOrderResponse, error) {
	if in == nil || in.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "order id is required")
	}
	c, err := claims(ctx)
	if err != nil {
		return nil, err
	}
	o, err := h.app.Get(ctx, in.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	if o.UserID != c.Subject {
		return nil, status.Error(codes.PermissionDenied, "forbidden")
	}
	return &orderv1.GetOrderResponse{Order: response(o)}, nil
}
func (h *Handler) ListOrders(ctx context.Context, in *orderv1.ListOrdersRequest) (*orderv1.ListOrdersResponse, error) {
	if in == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	c, err := claims(ctx)
	if err != nil {
		return nil, err
	}
	orders, _, err := h.app.List(ctx, int(in.Page), int(in.PageSize))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*orderv1.Order, 0, len(orders))
	for _, o := range orders {
		if o.UserID == c.Subject {
			out = append(out, response(o))
		}
	}
	return &orderv1.ListOrdersResponse{Orders: out, Total: int32(len(out))}, nil
}
