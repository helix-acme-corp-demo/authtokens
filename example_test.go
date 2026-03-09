package authtokens_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/helix-acme-corp-demo/authtokens"
)

func ExampleNewIssuer() {
	secret := []byte("my-secret-key-for-testing")
	issuer := authtokens.NewIssuer(
		authtokens.WithSecret(secret),
		authtokens.WithDefaultTTL(1*time.Hour),
	)

	token, err := issuer.Issue(authtokens.Claims{
		Subject: "user:42",
	})
	fmt.Println("error:", err)
	fmt.Println("has token:", token.Raw != "")
	// Output:
	// error: <nil>
	// has token: true
}

func ExampleNewValidator() {
	secret := []byte("my-secret-key-for-testing")
	issuer := authtokens.NewIssuer(authtokens.WithSecret(secret))
	validator := authtokens.NewValidator(authtokens.WithSecret(secret))

	token, _ := issuer.Issue(authtokens.Claims{
		Subject:   "user:42",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	claims, err := validator.Validate(token.Raw)
	fmt.Println("error:", err)
	fmt.Println("subject:", claims.Subject)
	// Output:
	// error: <nil>
	// subject: user:42
}

func ExampleWithRequiredScopes() {
	secret := []byte("my-secret-key-for-testing")
	issuer := authtokens.NewIssuer(authtokens.WithSecret(secret))

	token, _ := issuer.Issue(authtokens.Claims{
		Subject:   "user:42",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Extra:     map[string]string{"scopes": "read write"},
	})

	validator := authtokens.NewValidator(
		authtokens.WithSecret(secret),
		authtokens.WithRequiredScopes("read", "write"),
	)

	claims, err := validator.Validate(token.Raw)
	fmt.Println("error:", err)
	fmt.Println("subject:", claims.Subject)
	// Output:
	// error: <nil>
	// subject: user:42
}

func ExampleMiddleware() {
	secret := []byte("my-secret-key-for-testing")
	issuer := authtokens.NewIssuer(authtokens.WithSecret(secret))
	validator := authtokens.NewValidator(authtokens.WithSecret(secret))

	token, _ := issuer.Issue(authtokens.Claims{
		Subject:   "user:42",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})

	handler := authtokens.Middleware(validator)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _ := authtokens.ClaimsFromContext(r.Context())
		fmt.Fprintf(w, "hello %s", claims.Subject)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token.Raw)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	fmt.Println("status:", rec.Code)
	fmt.Println("body:", rec.Body.String())
	// Output:
	// status: 200
	// body: hello user:42
}

func ExampleIssuer_Refresh() {
	secret := []byte("my-secret-key-for-testing")
	ttl := 1 * time.Hour
	issuer := authtokens.NewIssuer(
		authtokens.WithSecret(secret),
		authtokens.WithDefaultTTL(ttl),
	)
	validator := authtokens.NewValidator(authtokens.WithSecret(secret))

	original, _ := issuer.Issue(authtokens.Claims{
		Subject:  "user:42",
		IssuedAt: time.Now().Add(-30 * time.Minute),
	})

	refreshed, err := issuer.Refresh(original.Raw, validator)
	fmt.Println("error:", err)
	fmt.Println("subject:", refreshed.Claims.Subject)
	fmt.Println("different token:", refreshed.Raw != original.Raw)
	// Output:
	// error: <nil>
	// subject: user:42
	// different token: true
}
