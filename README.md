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

## API Reference

### Types

- `Claims` — JWT claims: Subject, Audience, ExpiresAt, IssuedAt, ID, Extra
- `Token` — Issued token with Raw string and parsed Claims
- `Issuer` — Interface for creating tokens
- `Validator` — Interface for validating tokens

### Constructors

- `NewIssuer(opts ...Option) Issuer` — Create an HS256 token issuer
- `NewValidator(opts ...Option) Validator` — Create an HS256 token validator

### Options

- `WithSecret(secret []byte)` — Set the signing secret
- `WithDefaultTTL(d time.Duration)` — Set default token lifetime
- `WithAudience(aud string)` — Set expected audience for validation

### Errors

- `ErrInvalidToken` — Token is malformed
- `ErrExpiredToken` — Token has expired
- `ErrInvalidSignature` — Signature verification failed

## License

MIT
