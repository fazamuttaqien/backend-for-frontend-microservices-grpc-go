package bff

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
	"github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/internal/auth"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authUserClient struct { registered *userv1.User; login *userv1.LoginResponse; loginErr error; me *userv1.User }
func (c *authUserClient) Register(context.Context, string, string, string) (*userv1.User, error) { return c.registered, nil }
func (c *authUserClient) Login(context.Context, string, string) (*userv1.LoginResponse, error) { return c.login, c.loginErr }
func (c *authUserClient) Get(context.Context, string, string) (*userv1.User, error) { return c.me, nil }
func newAuthTestServer(t *testing.T, client UserService) *Server { t.Helper(); jwt, err := auth.NewJWT(strings.Repeat("s", 32), "user-service", time.Hour); if err != nil { t.Fatal(err) }; return NewServer(jwt, Clients{User: client}) }
func TestRegisterSuccess(t *testing.T) { s:=newAuthTestServer(t,&authUserClient{registered:&userv1.User{Id:"u1",Name:"Faza",Email:"faza@example.com"}}); r:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/register",strings.NewReader(`{"name":"Faza","email":"faza@example.com","password":"password123"}`)); w:=httptest.NewRecorder(); s.Handler().ServeHTTP(w,r); if w.Code!=http.StatusCreated { t.Fatalf("status=%d body=%s",w.Code,w.Body.String()) } }
func TestLoginSuccessAndIssuedTokenWorks(t *testing.T) { s:=newAuthTestServer(t,&authUserClient{login:&userv1.LoginResponse{User:&userv1.User{Id:"u1",Name:"Faza",Email:"faza@example.com"}}}); r:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/login",strings.NewReader(`{"email":"faza@example.com","password":"password123"}`)); w:=httptest.NewRecorder(); s.Handler().ServeHTTP(w,r); if w.Code!=http.StatusOK { t.Fatalf("status=%d body=%s",w.Code,w.Body.String()) }; var got AuthResponse; if err:=json.Unmarshal(w.Body.Bytes(),&got);err!=nil{t.Fatal(err)}; if got.AccessToken=="" {t.Fatal("missing access token")}; claims,err:=s.jwt.Parse(got.AccessToken);if err!=nil||claims.Subject!="u1"{t.Fatalf("token parse error: %v",err)} }
func TestLoginInvalidCredentialsIsRejected(t *testing.T) { s:=newAuthTestServer(t,&authUserClient{loginErr:status.Error(codes.Unauthenticated,"invalid credentials")}); r:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/login",strings.NewReader(`{"email":"faza@example.com","password":"wrongpass"}`)); w:=httptest.NewRecorder(); s.Handler().ServeHTTP(w,r); if w.Code!=http.StatusUnauthorized{t.Fatalf("status=%d body=%s",w.Code,w.Body.String())};if strings.Contains(w.Body.String(),"invalid credentials"){t.Fatalf("internal detail exposed: %s",w.Body.String())} }
func TestLogoutRequiresAuthentication(t *testing.T) { s:=newAuthTestServer(t,&authUserClient{});r:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/logout",nil);w:=httptest.NewRecorder();s.Handler().ServeHTTP(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("status=%d",w.Code)} }
func TestLogoutIsStatelessSuccess(t *testing.T) { s:=newAuthTestServer(t,&authUserClient{});token,err:=s.jwt.Issue("u1","faza@example.com");if err!=nil{t.Fatal(err)};r:=httptest.NewRequest(http.MethodPost,"/api/v1/auth/logout",nil);r.Header.Set("Authorization","Bearer "+token);w:=httptest.NewRecorder();s.Handler().ServeHTTP(w,r);if w.Code!=http.StatusNoContent{t.Fatalf("status=%d",w.Code)} }
func TestMeWithValidToken(t *testing.T) { s:=newAuthTestServer(t,&authUserClient{me:&userv1.User{Id:"u1",Name:"Faza",Email:"faza@example.com"}});token,err:=s.jwt.Issue("u1","faza@example.com");if err!=nil{t.Fatal(err)};r:=httptest.NewRequest(http.MethodGet,"/api/v1/users/me",nil);r.Header.Set("Authorization","Bearer "+token);w:=httptest.NewRecorder();s.Handler().ServeHTTP(w,r);if w.Code!=http.StatusOK{t.Fatalf("status=%d body=%s",w.Code,w.Body.String())} }
func TestMeRejectsExpiredToken(t *testing.T) { jwt,err:=auth.NewJWT(strings.Repeat("s",32),"user-service",time.Nanosecond);if err!=nil{t.Fatal(err)};s:=NewServer(jwt,Clients{User:&authUserClient{}});token,err:=jwt.Issue("u1","faza@example.com");if err!=nil{t.Fatal(err)};time.Sleep(time.Millisecond);r:=httptest.NewRequest(http.MethodGet,"/api/v1/users/me",nil);r.Header.Set("Authorization","Bearer "+token);w:=httptest.NewRecorder();s.Handler().ServeHTTP(w,r);if w.Code!=http.StatusUnauthorized{t.Fatalf("status=%d",w.Code)} }
