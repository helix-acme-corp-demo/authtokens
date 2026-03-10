package authtokens

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestMiddlewareValidToken(t *testing.T) {
	secret := []byte("test-secret-key")
	issuer := NewIssuer(WithSecret(secret))
	validator := NewValidator(WithSecret(secret))

	token, err := issuer.Issue(Claims{
		Subject:   "user:123",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	var captured Claims
	var found bool
	handler := Middleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured, found = ClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token.Raw)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !found {
		t.Fatal("ClaimsFromContext returned false")
	}
	if captured.Subject != "user:123" {
		t.Errorf("Subject = %q, want %q", captured.Subject, "user:123")
	}
}

func TestMiddlewareMissingHeader(t *testing.T) {
	validator := NewValidator(WithSecret([]byte("secret")))
	handler := Middleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}

	var body map[string]string
	json.NewDecoder(rec.Body).Decode(&body)
	if body["error"] != ErrMissingToken.Error() {
		t.Errorf("error = %q, want %q", body["error"], ErrMissingToken.Error())
	}
}

func TestMiddlewareInvalidToken(t *testing.T) {
	validator := NewValidator(WithSecret([]byte("secret")))
	handler := Middleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestMiddlewareExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key")
	issuer := NewIssuer(WithSecret(secret))
	validator := NewValidator(WithSecret(secret))

	token, err := issuer.Issue(Claims{
		Subject:   "user:123",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	handler := Middleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token.Raw)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
