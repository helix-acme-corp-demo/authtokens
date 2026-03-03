package authtokens

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestIssueAndValidate(t *testing.T) {
	secret := []byte("test-secret-key")
	issuer := NewIssuer(WithSecret(secret))
	validator := NewValidator(WithSecret(secret))

	claims := Claims{
		Subject:   "user:123",
		Audience:  "api.example.com",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		ID:        "token-001",
	}

	token, err := issuer.Issue(claims)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if token.Raw == "" {
		t.Fatal("Issue() returned empty Raw token")
	}

	validated, err := validator.Validate(token.Raw)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.Subject != claims.Subject {
		t.Errorf("Subject = %q, want %q", validated.Subject, claims.Subject)
	}
	if validated.Audience != claims.Audience {
		t.Errorf("Audience = %q, want %q", validated.Audience, claims.Audience)
	}
	if validated.ID != claims.ID {
		t.Errorf("ID = %q, want %q", validated.ID, claims.ID)
	}
}

func TestExpiredToken(t *testing.T) {
	secret := []byte("test-secret-key")
	issuer := NewIssuer(WithSecret(secret))
	validator := NewValidator(WithSecret(secret))

	claims := Claims{
		Subject:   "user:123",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}

	token, err := issuer.Issue(claims)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	_, err = validator.Validate(token.Raw)
	if !errors.Is(err, ErrExpiredToken) {
		t.Errorf("Validate() error = %v, want %v", err, ErrExpiredToken)
	}
}

func TestInvalidSignature(t *testing.T) {
	secret := []byte("test-secret-key")
	issuer := NewIssuer(WithSecret(secret))

	otherSecret := []byte("different-secret")
	validator := NewValidator(WithSecret(otherSecret))

	claims := Claims{
		Subject:   "user:123",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}

	token, err := issuer.Issue(claims)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	_, err = validator.Validate(token.Raw)
	if !errors.Is(err, ErrInvalidSignature) {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidSignature)
	}
}

func TestInvalidToken(t *testing.T) {
	validator := NewValidator(WithSecret([]byte("secret")))

	_, err := validator.Validate("not-a-jwt")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidToken)
	}

	_, err = validator.Validate("")
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Validate() error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestAudienceValidation(t *testing.T) {
	secret := []byte("test-secret-key")
	issuer := NewIssuer(WithSecret(secret))
	validator := NewValidator(
		WithSecret(secret),
		WithAudience("api.example.com"),
	)

	// Token with correct audience should pass
	token, err := issuer.Issue(Claims{
		Subject:   "user:123",
		Audience:  "api.example.com",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	if _, err := validator.Validate(token.Raw); err != nil {
		t.Errorf("Validate() with correct audience error = %v", err)
	}

	// Token with wrong audience should fail
	token, err = issuer.Issue(Claims{
		Subject:   "user:123",
		Audience:  "other.example.com",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	_, err = validator.Validate(token.Raw)
	if !errors.Is(err, ErrInvalidToken) {
		t.Errorf("Validate() with wrong audience error = %v, want %v", err, ErrInvalidToken)
	}
}

func TestExtraClaims(t *testing.T) {
	secret := []byte("test-secret-key")
	issuer := NewIssuer(WithSecret(secret))
	validator := NewValidator(WithSecret(secret))

	extra := map[string]string{
		"role":   "admin",
		"tenant": "acme",
	}

	token, err := issuer.Issue(Claims{
		Subject:   "user:123",
		ExpiresAt: time.Now().Add(1 * time.Hour),
		Extra:     extra,
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	validated, err := validator.Validate(token.Raw)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.Extra["role"] != "admin" {
		t.Errorf("Extra[role] = %q, want %q", validated.Extra["role"], "admin")
	}
	if validated.Extra["tenant"] != "acme" {
		t.Errorf("Extra[tenant] = %q, want %q", validated.Extra["tenant"], "acme")
	}
}

func TestDefaultTTL(t *testing.T) {
	secret := []byte("test-secret-key")
	ttl := 2 * time.Hour
	issuer := NewIssuer(WithSecret(secret), WithDefaultTTL(ttl))
	validator := NewValidator(WithSecret(secret))

	before := time.Now()
	token, err := issuer.Issue(Claims{
		Subject: "user:123",
	})
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}

	// The token claims should have ExpiresAt set automatically
	if token.Claims.ExpiresAt.IsZero() {
		t.Fatal("ExpiresAt should not be zero when DefaultTTL is set")
	}

	expectedMin := before.Add(ttl)
	expectedMax := time.Now().Add(ttl)
	if token.Claims.ExpiresAt.Before(expectedMin) || token.Claims.ExpiresAt.After(expectedMax) {
		t.Errorf("ExpiresAt = %v, want between %v and %v", token.Claims.ExpiresAt, expectedMin, expectedMax)
	}

	// Validate should succeed since the token is not expired
	validated, err := validator.Validate(token.Raw)
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.Subject != "user:123" {
		t.Errorf("Subject = %q, want %q", validated.Subject, "user:123")
	}

	// Verify the token has 3 parts (proper JWT format)
	parts := strings.Split(token.Raw, ".")
	if len(parts) != 3 {
		t.Errorf("Token has %d parts, want 3", len(parts))
	}
}
