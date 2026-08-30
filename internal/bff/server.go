package bff

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	orderv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/order/v1"
	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type Server struct {
	jwt     *auth.JWT
	clients Clients
}

func NewServer(jwt *auth.JWT, clients Clients) *Server { return &Server{jwt: jwt, clients: clients} }

type tokenKey struct{}

func contextWithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, tokenKey{}, token)
}
func tokenFromContext(ctx context.Context) string { v, _ := ctx.Value(tokenKey{}).(string); return v }

type ErrorResponse struct {
	Error string `json:"error"`
}
type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}
type ProductResponse struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       string `json:"price"`
	Stock       int32  `json:"stock"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}
type ProductListResponse struct {
	Products []ProductResponse `json:"products"`
	Total    int32             `json:"total"`
	Page     int32             `json:"page"`
	PageSize int32             `json:"page_size"`
}
type CreateOrderRequest struct {
	Items []CreateOrderItem `json:"items"`
}
type CreateOrderItem struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
}
type OrderItemResponse struct {
	ProductID string `json:"product_id"`
	Quantity  int32  `json:"quantity"`
	Price     string `json:"price"`
	Total     string `json:"total"`
}
type OrderResponse struct {
	ID        string              `json:"id"`
	UserID    string              `json:"user_id"`
	Items     []OrderItemResponse `json:"items"`
	Total     string              `json:"total"`
	Status    string              `json:"status"`
	CreatedAt string              `json:"created_at"`
	UpdatedAt string              `json:"updated_at"`
}
type OrderListResponse struct {
	Orders   []OrderResponse `json:"orders"`
	Total    int32           `json:"total"`
	Page     int32           `json:"page"`
	PageSize int32           `json:"page_size"`
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/users/me", s.auth(s.getMe))
	mux.HandleFunc("GET /api/v1/products", s.getProducts)
	mux.HandleFunc("GET /api/v1/products/{id}", s.getProduct)
	mux.HandleFunc("POST /api/v1/orders", s.auth(s.createOrder))
	mux.HandleFunc("GET /api/v1/orders", s.auth(s.listOrders))
	mux.HandleFunc("GET /api/v1/orders/{id}", s.auth(s.getOrder))
	return requestMiddleware(mux)
}
func (s *Server) auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		p := strings.Fields(r.Header.Get("Authorization"))
		if len(p) != 2 || !strings.EqualFold(p[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		claims, err := s.jwt.Parse(p[1])
		if err != nil {
			writeError(w, http.StatusUnauthorized, "invalid authentication token")
			return
		}
		ctx := contextWithToken(auth.WithClaims(r.Context(), claims), p[1])
		next(w, r.WithContext(ctx))
	}
}
func requestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, ErrorResponse{Error: msg})
}
func mapGRPCError(w http.ResponseWriter, err error) {
	code := status.Code(err)
	hc := http.StatusInternalServerError
	msg := "internal server error"
	switch code {
	case codes.InvalidArgument:
		hc = http.StatusBadRequest
		msg = "invalid request"
	case codes.Unauthenticated:
		hc = http.StatusUnauthorized
		msg = "authentication required"
	case codes.NotFound:
		hc = http.StatusNotFound
		msg = "resource not found"
	case codes.AlreadyExists:
		hc = http.StatusConflict
		msg = "resource already exists"
	case codes.DeadlineExceeded:
		hc = http.StatusGatewayTimeout
		msg = "service timeout"
	case codes.Unavailable:
		hc = http.StatusBadGateway
		msg = "service unavailable"
	}
	writeError(w, hc, msg)
}

func (s *Server) getMe(w http.ResponseWriter, r *http.Request) {
	c, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, 401, "authentication required")
		return
	}
	u, err := s.clients.User.Get(r.Context(), c.Subject, tokenFromContext(r.Context()))
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	if u == nil {
		writeError(w, 502, "invalid user response")
		return
	}
	writeJSON(w, 200, UserResponse{ID: u.GetId(), Name: u.GetName(), Email: u.GetEmail(), CreatedAt: u.GetCreatedAt(), UpdatedAt: u.GetUpdatedAt()})
}
func (s *Server) getProducts(w http.ResponseWriter, r *http.Request) {
	p, sz := pagination(r)
	res, err := s.clients.Product.List(r.Context(), p, sz)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	items := make([]ProductResponse, 0, len(res.GetProducts()))
	for _, v := range res.GetProducts() {
		if v != nil {
			items = append(items, toProduct(v))
		}
	}
	writeJSON(w, 200, ProductListResponse{Products: items, Total: res.GetTotal(), Page: p, PageSize: sz})
}
func (s *Server) getProduct(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, 400, "product id is required")
		return
	}
	p, err := s.clients.Product.Get(r.Context(), id)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	if p == nil {
		writeError(w, 502, "invalid product response")
		return
	}
	writeJSON(w, 200, toProduct(p))
}
func (s *Server) createOrder(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, 400, "invalid JSON body")
		return
	}
	if err := validateCreateOrder(req); err != nil {
		writeError(w, 400, err.Error())
		return
	}
	c, ok := auth.ClaimsFromContext(r.Context())
	if !ok {
		writeError(w, 401, "authentication required")
		return
	}
	items := make([]*orderv1.CreateOrderItemRequest, 0, len(req.Items))
	for _, i := range req.Items {
		items = append(items, &orderv1.CreateOrderItemRequest{ProductId: i.ProductID, Quantity: i.Quantity})
	}
	o, err := s.clients.Order.Create(r.Context(), &orderv1.CreateOrderRequest{UserId: c.Subject, Items: items})
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	if o == nil {
		writeError(w, 502, "invalid order response")
		return
	}
	writeJSON(w, 201, toOrder(o))
}
func (s *Server) listOrders(w http.ResponseWriter, r *http.Request) {
	p, sz := pagination(r)
	res, err := s.clients.Order.List(r.Context(), p, sz)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	items := make([]OrderResponse, 0, len(res.GetOrders()))
	for _, o := range res.GetOrders() {
		if o != nil {
			items = append(items, toOrder(o))
		}
	}
	writeJSON(w, 200, OrderListResponse{Orders: items, Total: res.GetTotal(), Page: p, PageSize: sz})
}
func (s *Server) getOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, 400, "order id is required")
		return
	}
	o, err := s.clients.Order.Get(r.Context(), id)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	if o == nil {
		writeError(w, 502, "invalid order response")
		return
	}
	writeJSON(w, 200, toOrder(o))
}
func pagination(r *http.Request) (int32, int32) {
	p, _ := strconv.Atoi(r.URL.Query().Get("page"))
	sz, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if p < 1 {
		p = 1
	}
	if sz < 1 {
		sz = 20
	}
	if sz > 100 {
		sz = 100
	}
	return int32(p), int32(sz)
}
func toProduct(p *productv1.Product) ProductResponse {
	return ProductResponse{ID: p.GetId(), Name: p.GetName(), Description: p.GetDescription(), Price: p.GetPrice(), Stock: p.GetStock(), CreatedAt: p.GetCreatedAt(), UpdatedAt: p.GetUpdatedAt()}
}
func toOrder(o *orderv1.Order) OrderResponse {
	items := make([]OrderItemResponse, 0, len(o.GetItems()))
	for _, i := range o.GetItems() {
		if i != nil {
			items = append(items, OrderItemResponse{ProductID: i.GetProductId(), Quantity: i.GetQuantity(), Price: i.GetPrice(), Total: i.GetTotal()})
		}
	}
	return OrderResponse{ID: o.GetId(), UserID: o.GetUserId(), Items: items, Total: o.GetTotal(), Status: o.GetStatus(), CreatedAt: o.GetCreatedAt(), UpdatedAt: o.GetUpdatedAt()}
}
func decodeJSON(r *http.Request, dst any) error {
	if r.Body == nil {
		return errors.New("body required")
	}
	defer r.Body.Close()
	dec := json.NewDecoder(io.LimitReader(r.Body, 1<<20))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	var extra any
	if err := dec.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}
func validateCreateOrder(r CreateOrderRequest) error {
	if len(r.Items) == 0 {
		return errors.New("items are required")
	}
	if len(r.Items) > 100 {
		return errors.New("too many items")
	}
	for _, i := range r.Items {
		if strings.TrimSpace(i.ProductID) == "" {
			return errors.New("product_id is required")
		}
		if i.Quantity <= 0 {
			return fmt.Errorf("quantity must be greater than zero")
		}
	}
	return nil
}

// Keep the auth token in outgoing metadata only for the User Service, which protects GetUser with its gRPC interceptor.
func withBearer(ctx context.Context, token string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", "Bearer "+token)
}
