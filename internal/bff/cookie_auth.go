package bff

import (
	"bytes"
	"encoding/json"
	"net/http"
	"time"
)

type CookieConfig struct {
	Name           string
	Path           string
	Secure         bool
	SameSite       http.SameSite
	MaxAge         int
	AllowedOrigins []string
}

// CookieAuthHandler adapts the existing bearer-token authentication to an
// HttpOnly cookie without introducing a second JWT implementation.
// Authorization: Bearer remains supported for backward compatibility.
func CookieAuthHandler(next http.Handler, cfg CookieConfig) http.Handler {
	if cfg.Name == "" { cfg.Name = "access_token" }
	if cfg.Path == "" { cfg.Path = "/" }

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isUnsafeMethod(r.Method) && hasCookie(r, cfg.Name) && !allowedOrigin(r.Header.Get("Origin"), cfg.AllowedOrigins) {
			writeError(w, http.StatusForbidden, "forbidden origin")
			return
		}

		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/login" {
			capture := newAuthCaptureWriter()
			next.ServeHTTP(capture, r)
			if capture.statusCode == http.StatusOK {
				var payload struct { AccessToken string `json:"access_token"`; ExpiresAt int64 `json:"expires_at"` }
				if err := json.Unmarshal(capture.body.Bytes(), &payload); err == nil && payload.AccessToken != "" {
					http.SetCookie(w, &http.Cookie{Name: cfg.Name, Value: payload.AccessToken, Path: cfg.Path, HttpOnly: true, Secure: cfg.Secure, SameSite: cfg.SameSite, MaxAge: cfg.MaxAge, Expires: time.Unix(payload.ExpiresAt, 0).UTC()})
					capture.body = bytes.NewBuffer(stripAccessToken(capture.body.Bytes()))
				}
			}
			capture.writeTo(w)
			return
		}

		if r.Method == http.MethodPost && r.URL.Path == "/api/v1/auth/logout" {
			injectCookieAuthorization(r, cfg.Name)
			capture := newAuthCaptureWriter()
			next.ServeHTTP(capture, r)
			if capture.statusCode >= 200 && capture.statusCode < 300 {
				http.SetCookie(w, &http.Cookie{Name: cfg.Name, Value: "", Path: cfg.Path, HttpOnly: true, Secure: cfg.Secure, SameSite: cfg.SameSite, MaxAge: -1, Expires: time.Unix(1, 0).UTC()})
			}
			capture.writeTo(w)
			return
		}

		injectCookieAuthorization(r, cfg.Name)
		next.ServeHTTP(w, r)
	})
}

func isUnsafeMethod(method string) bool { return method != http.MethodGet && method != http.MethodHead && method != http.MethodOptions && method != http.MethodTrace }
func hasCookie(r *http.Request, name string) bool { cookie, err := r.Cookie(name); return err == nil && cookie.Value != "" }
func allowedOrigin(origin string, allowed []string) bool {
	if origin == "" { return true }
	for _, candidate := range allowed { if origin == candidate { return true } }
	return false
}

func injectCookieAuthorization(r *http.Request, cookieName string) {
	if r.Header.Get("Authorization") != "" { return }
	cookie, err := r.Cookie(cookieName)
	if err == nil && cookie.Value != "" { r.Header.Set("Authorization", "Bearer "+cookie.Value) }
}

func stripAccessToken(body []byte) []byte {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil { return body }
	delete(payload, "access_token")
	result, err := json.Marshal(payload)
	if err != nil { return body }
	return append(result, '\n')
}

type authCaptureWriter struct { header http.Header; body bytes.Buffer; statusCode int; wroteHeader bool }
func newAuthCaptureWriter() *authCaptureWriter { return &authCaptureWriter{header: make(http.Header), statusCode: http.StatusOK} }
func (w *authCaptureWriter) Header() http.Header { return w.header }
func (w *authCaptureWriter) WriteHeader(code int) { if w.wroteHeader { return }; w.statusCode = code; w.wroteHeader = true }
func (w *authCaptureWriter) Write(p []byte) (int, error) { if !w.wroteHeader { w.WriteHeader(http.StatusOK) }; return w.body.Write(p) }
func (w *authCaptureWriter) writeTo(dst http.ResponseWriter) { for key, values := range w.header { dst.Header()[key] = values }; dst.Header().Del("Content-Length"); dst.WriteHeader(w.statusCode); _, _ = dst.Write(w.body.Bytes()) }

func ParseSameSite(value string) http.SameSite {
	switch value { case "strict": return http.SameSiteStrictMode; case "none": return http.SameSiteNoneMode; default: return http.SameSiteLaxMode }
}
