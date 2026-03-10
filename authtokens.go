package authtokens

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

var (
	ErrInvalidToken       = errors.New("token is malformed")
	ErrExpiredToken       = errors.New("token has expired")
	ErrInvalidSignature   = errors.New("signature verification failed")
	ErrInsufficientScopes = errors.New("token missing required scopes")
	ErrRevokedToken       = errors.New("token has been revoked")
)

// Claims represents the payload of a JWT.
type Claims struct {
	Subject   string
	Audience  string
	ExpiresAt time.Time
	IssuedAt  time.Time
	ID        string
	Extra     map[string]string
}

// Token represents an issued JWT with its raw encoded form and parsed claims.
type Token struct {
	Raw    string
	Claims Claims
}

// Issuer creates signed JWT tokens.
type Issuer interface {
	Issue(claims Claims) (Token, error)
	Refresh(raw string, v Validator) (Token, error)
}

// RevocationChecker determines whether a token has been revoked by its ID.
type RevocationChecker interface {
	IsRevoked(id string) bool
}

// Validator verifies and parses JWT tokens.
type Validator interface {
	Validate(raw string) (Claims, error)
}

// Option configures an Issuer or Validator.
type Option func(*config)

type config struct {
	secret         []byte
	defaultTTL     time.Duration
	audience       string
	requiredScopes []string
	revocation     RevocationChecker
}

// WithSecret sets the HMAC signing secret.
func WithSecret(secret []byte) Option {
	return func(c *config) {
		c.secret = secret
	}
}

// WithDefaultTTL sets the default token lifetime applied when Claims.ExpiresAt is zero.
func WithDefaultTTL(d time.Duration) Option {
	return func(c *config) {
		c.defaultTTL = d
	}
}

// WithAudience sets the expected audience for validation.
func WithAudience(aud string) Option {
	return func(c *config) {
		c.audience = aud
	}
}

// WithRequiredScopes sets scopes that must be present in the token's Extra["scopes"] field.
// Scopes in the token are expected to be space-separated.
func WithRequiredScopes(scopes ...string) Option {
	return func(c *config) {
		c.requiredScopes = scopes
	}
}

// WithRevocationCheck sets a RevocationChecker used during validation
// to reject tokens that have been revoked.
func WithRevocationCheck(checker RevocationChecker) Option {
	return func(c *config) {
		c.revocation = checker
	}
}

// NewIssuer returns an Issuer that creates HS256-signed JWTs.
func NewIssuer(opts ...Option) Issuer {
	cfg := &config{}
	for _, o := range opts {
		o(cfg)
	}
	return &hs256Issuer{cfg: cfg}
}

// NewValidator returns a Validator that verifies HS256-signed JWTs.
func NewValidator(opts ...Option) Validator {
	cfg := &config{}
	for _, o := range opts {
		o(cfg)
	}
	return &hs256Validator{cfg: cfg}
}

// hs256Issuer implements Issuer using HMAC-SHA256.
type hs256Issuer struct {
	cfg *config
}

type jwtHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type jwtPayload struct {
	Sub   string            `json:"sub,omitempty"`
	Aud   string            `json:"aud,omitempty"`
	Exp   int64             `json:"exp,omitempty"`
	Iat   int64             `json:"iat,omitempty"`
	Jti   string            `json:"jti,omitempty"`
	Extra map[string]string `json:"extra,omitempty"`
}

func (i *hs256Issuer) Issue(claims Claims) (Token, error) {
	now := time.Now()
	if claims.IssuedAt.IsZero() {
		claims.IssuedAt = now
	}
	if claims.ExpiresAt.IsZero() && i.cfg.defaultTTL > 0 {
		claims.ExpiresAt = now.Add(i.cfg.defaultTTL)
	}

	header := jwtHeader{Alg: "HS256", Typ: "JWT"}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return Token{}, err
	}

	payload := jwtPayload{
		Sub:   claims.Subject,
		Aud:   claims.Audience,
		Iat:   claims.IssuedAt.Unix(),
		Jti:   claims.ID,
		Extra: claims.Extra,
	}
	if !claims.ExpiresAt.IsZero() {
		payload.Exp = claims.ExpiresAt.Unix()
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return Token{}, err
	}

	headerEncoded := base64URLEncode(headerBytes)
	payloadEncoded := base64URLEncode(payloadBytes)
	signingInput := headerEncoded + "." + payloadEncoded

	signature := sign([]byte(signingInput), i.cfg.secret)
	signatureEncoded := base64URLEncode(signature)

	raw := signingInput + "." + signatureEncoded

	return Token{Raw: raw, Claims: claims}, nil
}

func (i *hs256Issuer) Refresh(raw string, v Validator) (Token, error) {
	old, err := v.Validate(raw)
	if err != nil {
		return Token{}, err
	}

	now := time.Now()
	refreshed := Claims{
		Subject:  old.Subject,
		Audience: old.Audience,
		ID:       old.ID,
		Extra:    old.Extra,
		IssuedAt: now,
	}
	if i.cfg.defaultTTL > 0 {
		refreshed.ExpiresAt = now.Add(i.cfg.defaultTTL)
	}

	return i.Issue(refreshed)
}

// hs256Validator implements Validator using HMAC-SHA256.
type hs256Validator struct {
	cfg *config
}

func (v *hs256Validator) Validate(raw string) (Claims, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalidToken
	}

	signingInput := parts[0] + "." + parts[1]
	signatureBytes, err := base64URLDecode(parts[2])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	expected := sign([]byte(signingInput), v.cfg.secret)
	if !hmac.Equal(signatureBytes, expected) {
		return Claims{}, ErrInvalidSignature
	}

	payloadBytes, err := base64URLDecode(parts[1])
	if err != nil {
		return Claims{}, ErrInvalidToken
	}

	var payload jwtPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return Claims{}, ErrInvalidToken
	}

	claims := Claims{
		Subject:  payload.Sub,
		Audience: payload.Aud,
		ID:       payload.Jti,
		Extra:    payload.Extra,
	}
	if payload.Iat != 0 {
		claims.IssuedAt = time.Unix(payload.Iat, 0)
	}
	if payload.Exp != 0 {
		claims.ExpiresAt = time.Unix(payload.Exp, 0)
		if time.Now().After(claims.ExpiresAt) {
			return Claims{}, ErrExpiredToken
		}
	}

	if v.cfg.revocation != nil && claims.ID != "" {
		if v.cfg.revocation.IsRevoked(claims.ID) {
			return Claims{}, ErrRevokedToken
		}
	}

	if v.cfg.audience != "" && claims.Audience != v.cfg.audience {
		return Claims{}, ErrInvalidToken
	}

	if len(v.cfg.requiredScopes) > 0 {
		granted := make(map[string]bool)
		if claims.Extra != nil {
			for _, s := range strings.Split(claims.Extra["scopes"], " ") {
				if s != "" {
					granted[s] = true
				}
			}
		}
		for _, required := range v.cfg.requiredScopes {
			if !granted[required] {
				return Claims{}, ErrInsufficientScopes
			}
		}
	}

	return claims, nil
}

func sign(data, secret []byte) []byte {
	mac := hmac.New(sha256.New, secret)
	mac.Write(data)
	return mac.Sum(nil)
}

func base64URLEncode(data []byte) string {
	return base64.RawURLEncoding.EncodeToString(data)
}

func base64URLDecode(s string) ([]byte, error) {
	return base64.RawURLEncoding.DecodeString(s)
}
