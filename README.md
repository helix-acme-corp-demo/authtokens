# authtokens

JWT token creation, validation, and management for Go services. HS256 signing with zero external dependencies.

JSON Web Tokens (JWT) are the standard for stateless authentication in microservices. This library provides a clean interface for issuing and validating HS256-signed JWTs with support for standard claims (subject, audience, expiration) and custom claims. No external dependencies — built entirely on Go's standard library.

Designed for service-to-service authentication, API gateway token validation, user session management, and any system that needs stateless bearer tokens. The `Issuer` and `Validator` interfaces allow easy testing and future extension to other signing algorithms like RS256.

## Installation

```bash
go get github.com/helix-acme-corp-demo/authtokens
```

## Quick Start

```go
secret := []byte("your-256-bit-secret")

// Issue a token
issuer := authtokens.NewIssuer(
    authtokens.WithSecret(secret),
    authtokens.WithDefaultTTL(1*time.Hour),
)
token, err := issuer.Issue(authtokens.Claims{
    Subject: "user:42",
})

// Validate a token
validator := authtokens.NewValidator(authtokens.WithSecret(secret))
claims, err := validator.Validate(token.Raw)
fmt.Println(claims.Subject) // "user:42"
```

## Scope-Based Authorization

Tokens can carry scopes in `Extra["scopes"]` as a space-separated string. Validators can enforce that required scopes are present:

```go
token, _ := issuer.Issue(authtokens.Claims{
    Subject:   "user:42",
    ExpiresAt: time.Now().Add(1 * time.Hour),
    Extra:     map[string]string{"scopes": "read write admin"},
})

validator := authtokens.NewValidator(
    authtokens.WithSecret(secret),
    authtokens.WithRequiredScopes("read", "write"),
)
claims, err := validator.Validate(token.Raw) // passes
```

## Revocation Checking

Implement the `RevocationChecker` interface to reject revoked tokens during validation:

```go
type myChecker struct { /* ... */ }
func (c *myChecker) IsRevoked(id string) bool { /* check store */ }

validator := authtokens.NewValidator(
    authtokens.WithSecret(secret),
    authtokens.WithRevocationCheck(&myChecker{}),
)
```

## Token Refresh

Refresh an existing valid token, copying its claims with a new `IssuedAt` and `ExpiresAt`:

```go
issuer := authtokens.NewIssuer(
    authtokens.WithSecret(secret),
    authtokens.WithDefaultTTL(1*time.Hour),
)
validator := authtokens.NewValidator(authtokens.WithSecret(secret))

refreshed, err := issuer.Refresh(originalToken.Raw, validator)
```

## HTTP Middleware

Protect HTTP handlers with Bearer token validation. Claims are injected into the request context:

```go
validator := authtokens.NewValidator(authtokens.WithSecret(secret))

mux := http.NewServeMux()
mux.Handle("/api/", authtokens.Middleware(validator)(apiHandler))

// Inside your handler:
claims, ok := authtokens.ClaimsFromContext(r.Context())
```

## API Reference

### Types

- `Claims` — JWT claims: Subject, Audience, ExpiresAt, IssuedAt, ID, Extra
- `Token` — Issued token with Raw string and parsed Claims
- `Issuer` — Interface for creating and refreshing tokens
- `Validator` — Interface for validating tokens
- `RevocationChecker` — Interface for checking if a token ID has been revoked

### Constructors

- `NewIssuer(opts ...Option) Issuer` — Create an HS256 token issuer
- `NewValidator(opts ...Option) Validator` — Create an HS256 token validator

### Functions

- `Middleware(v Validator) func(http.Handler) http.Handler` — HTTP middleware for Bearer token validation
- `ClaimsFromContext(ctx context.Context) (Claims, bool)` — Retrieve claims from request context

### Options

- `WithSecret(secret []byte)` — Set the signing secret
- `WithDefaultTTL(d time.Duration)` — Set default token lifetime
- `WithAudience(aud string)` — Set expected audience for validation
- `WithRequiredScopes(scopes ...string)` — Set required scopes for validation
- `WithRevocationCheck(checker RevocationChecker)` — Set revocation checker for validation

### Errors

- `ErrInvalidToken` — Token is malformed
- `ErrExpiredToken` — Token has expired
- `ErrInvalidSignature` — Signature verification failed
- `ErrInsufficientScopes` — Token missing required scopes
- `ErrRevokedToken` — Token has been revoked
- `ErrMissingToken` — Authorization token not provided

## License

MIT
