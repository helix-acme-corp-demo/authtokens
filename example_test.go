package authtokens_test

import (
	"fmt"
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
