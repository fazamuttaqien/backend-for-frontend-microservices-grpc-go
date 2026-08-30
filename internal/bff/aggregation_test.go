package bff

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	orderv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/order/v1"
	productv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/product/v1"
	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
)

type aggregationUserClient struct {
	user *userv1.User
	err  error
}

func (c *aggregationUserClient) Get(context.Context, string, string) (*userv1.User, error) {
	return c.user, c.err
}

type aggregationProductClient struct {
	mu       sync.Mutex
	products map[string]*productv1.Product
	errors   map[string]error
	calls    int
}

func (c *aggregationProductClient) List(context.Context, int32, int32) (*productv1.ListProductResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *aggregationProductClient) Get(_ context.Context, id string) (*productv1.Product, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls++
	if err := c.errors[id]; err != nil {
		return nil, err
	}
	return c.products[id], nil
}

type aggregationOrderClient struct {
	order *orderv1.Order
	err   error
}

func (c *aggregationOrderClient) Create(context.Context, *orderv1.CreateOrderRequest) (*orderv1.Order, error) {
	return nil, errors.New("not implemented")
}

func (c *aggregationOrderClient) List(context.Context, int32, int32) (*orderv1.ListOrdersResponse, error) {
	return nil, errors.New("not implemented")
}

func (c *aggregationOrderClient) Get(context.Context, string) (*orderv1.Order, error) {
	return c.order, c.err
}

func aggregationTestToken(t *testing.T) string {
	t.Helper()
	jwt, err := auth.NewJWT(strings.Repeat("s", 32), "user-service", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	token, err := jwt.Issue("user-1", "user@example.com")
	if err != nil {
		t.Fatal(err)
	}
	return token
}

func TestGetOrderAggregatesOrderCustomerAndProducts(t *testing.T) {
	products := &aggregationProductClient{
		products: map[string]*productv1.Product{
			"product-1": {Id: "product-1", Name: "Keyboard", Price: "100"},
			"product-2": {Id: "product-2", Name: "Mouse", Price: "50"},
		},
		errors: map[string]error{},
	}
	server := NewServer(nil, Clients{
		Order: &aggregationOrderClient{order: &orderv1.Order{
			Id: "order-1", UserId: "user-1", Status: "pending",
			Items: []*orderv1.OrderItem{{ProductId: "product-1", Quantity: 2}, {ProductId: "product-2", Quantity: 1}},
		}},
		User:    &aggregationUserClient{user: &userv1.User{Id: "user-1", Name: "Faza", Email: "user@example.com"}},
		Product: products,
	})

	jwt, err := auth.NewJWT(strings.Repeat("s", 32), "user-service", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	server.jwt = jwt
	r := httptest.NewRequest(http.MethodGet, "/api/v1/orders/order-1", nil)
	r.Header.Set("Authorization", "Bearer "+aggregationTestToken(t))
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	body := w.Body.String()
	for _, want := range []string{"order-1", "user@example.com", "Keyboard", "Mouse"} {
		if !strings.Contains(body, want) {
			t.Fatalf("response missing %q: %s", want, body)
		}
	}
	if products.calls != 2 {
		t.Fatalf("product calls=%d, want 2", products.calls)
	}
}

func TestGetOrderReportsPartialProductFailure(t *testing.T) {
	products := &aggregationProductClient{
		products: map[string]*productv1.Product{
			"product-1": {Id: "product-1", Name: "Keyboard"},
		},
		errors: map[string]error{"product-2": errors.New("product service unavailable")},
	}
	jwt, err := auth.NewJWT(strings.Repeat("s", 32), "user-service", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	server := NewServer(jwt, Clients{
		Order:   &aggregationOrderClient{order: &orderv1.Order{Id: "order-1", UserId: "user-1", Items: []*orderv1.OrderItem{{ProductId: "product-1"}, {ProductId: "product-2"}}}},
		User:    &aggregationUserClient{user: &userv1.User{Id: "user-1", Name: "Faza"}},
		Product: products,
	})

	token, _ := jwt.Issue("user-1", "user@example.com")
	r := httptest.NewRequest(http.MethodGet, "/api/v1/orders/order-1", nil)
	r.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	server.Handler().ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "product(s) unavailable") {
		t.Fatalf("partial failure not reported: %s", w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "Keyboard") {
		t.Fatalf("successful product missing: %s", w.Body.String())
	}
}
