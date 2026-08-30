package bff

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
)

func testServer(t *testing.T)*Server{t.Helper();jwt,err:=auth.NewJWT(strings.Repeat("s",32),"user-service",time.Hour);if err!=nil{t.Fatal(err)};return NewServer(jwt,Clients{})}
func TestAuthRejectsMissingToken(t *testing.T){s:=testServer(t);r:=httptest.NewRequest(http.MethodGet,"/api/v1/users/me",nil);w:=httptest.NewRecorder();s.Handler().ServeHTTP(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("status=%d",w.Code)}}
func TestAuthRejectsInvalidToken(t *testing.T){s:=testServer(t);r:=httptest.NewRequest(http.MethodGet,"/api/v1/users/me",nil);r.Header.Set("Authorization","Bearer invalid");w:=httptest.NewRecorder();s.Handler().ServeHTTP(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("status=%d",w.Code)}}
func TestCreateOrderValidation(t *testing.T){if err:=validateCreateOrder(CreateOrderRequest{});err==nil{t.Fatal("expected validation error")};if err:=validateCreateOrder(CreateOrderRequest{Items:[]CreateOrderItem{{ProductID:"p1",Quantity:0}}});err==nil{t.Fatal("expected quantity validation error")}}
func TestContextTokenRoundTrip(t *testing.T){ctx:=contextWithToken(context.Background(),"token");if got:=tokenFromContext(ctx);got!="token"{t.Fatalf("token=%q",got)}}
