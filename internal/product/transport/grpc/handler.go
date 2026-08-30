package grpc

import (
	"context"
	"errors"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/application"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/product/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	productv1.UnimplementedProductServiceServer
	app *application.ProductService
}

func NewHandler(app *application.ProductService) *Handler { return &Handler{app: app} }
func response(p *domain.Product) *productv1.Product {
	return &productv1.Product{Id: p.ID, Name: p.Name, Description: p.Description, Price: p.Price, Stock: p.Stock, CreatedAt: p.CreatedAt.UTC().Format("2006-01-02T15:04:05Z"), UpdatedAt: p.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z")}
}
func mapErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrProductNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidProduct):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
func (h *Handler) CreateProduct(ctx context.Context, in *productv1.CreateProductRequest) (*productv1.CreateProductResponse, error) {
	p, err := h.app.Create(ctx, in.Name, in.Description, in.Price, in.Stock)
	if err != nil {
		return nil, mapErr(err)
	}
	return &productv1.CreateProductResponse{Product: response(p)}, nil
}
func (h *Handler) GetProduct(ctx context.Context, in *productv1.GetProductRequest) (*productv1.GetProductResponse, error) {
	p, err := h.app.Get(ctx, in.Id)
	if err != nil {
		return nil, mapErr(err)
	}
	return &productv1.GetProductResponse{Product: response(p)}, nil
}
func (h *Handler) ListProduct(ctx context.Context, in *productv1.ListProductRequest) (*productv1.ListProductResponse, error) {
	ps, total, err := h.app.List(ctx, int(in.Page), int(in.PageSize))
	if err != nil {
		return nil, mapErr(err)
	}
	out := make([]*productv1.Product, 0, len(ps))
	for _, p := range ps {
		out = append(out, response(p))
	}
	return &productv1.ListProductResponse{Products: out, Total: int32(total)}, nil
}
func (h *Handler) UpdateProduct(ctx context.Context, in *productv1.UpdateProductRequest) (*productv1.UpdateProductResponse, error) {
	p, err := h.app.Update(ctx, in.Id, in.Name, in.Description, in.Price, in.Stock)
	if err != nil {
		return nil, mapErr(err)
	}
	return &productv1.UpdateProductResponse{Product: response(p)}, nil
}
