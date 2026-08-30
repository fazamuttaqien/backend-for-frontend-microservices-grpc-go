package bff

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	userv1 "github.com/fazamuttaqien/backend-for-frontend-microservices-grpc-go/gen/user/v1"
)

func TestCookieAuthLoginSetsHttpOnlyCookieAndHidesToken(t *testing.T) {
	s := newAuthTestServer(t, &authUserClient{login: &userv1.LoginResponse{User: &userv1.User{Id: "u1", Name: "Faza", Email: "faza@example.com"}}})
	h := CookieAuthHandler(s.Handler(), CookieConfig{Name: "access_token", Path: "/", SameSite: http.SameSiteLaxMode, MaxAge: 3600})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", strings.NewReader(`{"email":"faza@example.com","password":"password123"}`))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
	cookies := w.Result().Cookies()
	if len(cookies) != 1 { t.Fatalf("expected one cookie, got %d", len(cookies)) }
	cookie := cookies[0]
	if cookie.Name != "access_token" || !cookie.HttpOnly || cookie.Path != "/" || cookie.SameSite != http.SameSiteLaxMode { t.Fatalf("unexpected cookie: %+v", cookie) }
	if cookie.Value == "" { t.Fatal("cookie token is empty") }
	if strings.Contains(w.Body.String(), "access_token") { t.Fatalf("token leaked in response: %s", w.Body.String()) }
}

func TestCookieAuthMeUsesCookieWithoutAuthorizationHeader(t *testing.T) {
	s := newAuthTestServer(t, &authUserClient{me: &userv1.User{Id: "u1", Name: "Faza", Email: "faza@example.com"}})
	token, err := s.jwt.Issue("u1", "faza@example.com")
	if err != nil { t.Fatal(err) }
	h := CookieAuthHandler(s.Handler(), CookieConfig{Name: "access_token", Path: "/"})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusOK { t.Fatalf("status=%d body=%s", w.Code, w.Body.String()) }
}

func TestCookieAuthMeRejectsMissingAuthentication(t *testing.T) {
	s := newAuthTestServer(t, &authUserClient{})
	h := CookieAuthHandler(s.Handler(), CookieConfig{Name: "access_token", Path: "/"})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized { t.Fatalf("status=%d", w.Code) }
}

func TestCookieAuthMeRejectsInvalidCookie(t *testing.T) {
	s := newAuthTestServer(t, &authUserClient{})
	h := CookieAuthHandler(s.Handler(), CookieConfig{Name: "access_token", Path: "/"})
	r := httptest.NewRequest(http.MethodGet, "/api/v1/users/me", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: "invalid"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusUnauthorized { t.Fatalf("status=%d", w.Code) }
}

func TestCookieAuthLogoutDeletesCookie(t *testing.T) {
	s := newAuthTestServer(t, &authUserClient{})
	token, err := s.jwt.Issue("u1", "faza@example.com")
	if err != nil { t.Fatal(err) }
	h := CookieAuthHandler(s.Handler(), CookieConfig{Name: "access_token", Path: "/", SameSite: http.SameSiteLaxMode})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent { t.Fatalf("status=%d", w.Code) }
	cookies := w.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "access_token" || cookies[0].MaxAge != -1 || cookies[0].Path != "/" { t.Fatalf("unexpected deletion cookie: %+v", cookies) }
}

func TestCookieAuthRejectsCrossOriginCookieWrite(t *testing.T) {
	s := newAuthTestServer(t, &authUserClient{})
	token, err := s.jwt.Issue("u1", "faza@example.com")
	if err != nil { t.Fatal(err) }
	h := CookieAuthHandler(s.Handler(), CookieConfig{Name: "access_token", Path: "/", AllowedOrigins: []string{"http://localhost:5173"}})
	r := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.Header.Set("Origin", "https://evil.example")
	r.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden { t.Fatalf("status=%d", w.Code) }
}

func TestCORSMiddlewareAllowsConfiguredOriginWithCredentials(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := CORSMiddleware([]string{"http://localhost:5173"}, true)(next)
	r := httptest.NewRequest(http.MethodOptions, "/api/v1/users/me", nil)
	r.Header.Set("Origin", "http://localhost:5173")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusNoContent { t.Fatalf("status=%d", w.Code) }
	if w.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" { t.Fatal("origin was not allowed") }
	if w.Header().Get("Access-Control-Allow-Credentials") != "true" { t.Fatal("credentials were not enabled") }
}

func TestCORSMiddlewareRejectsUnconfiguredOrigin(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) })
	h := CORSMiddleware([]string{"http://localhost:5173"}, true)(next)
	r := httptest.NewRequest(http.MethodOptions, "/api/v1/users/me", nil)
	r.Header.Set("Origin", "https://evil.example")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != http.StatusForbidden { t.Fatalf("status=%d", w.Code) }
}
