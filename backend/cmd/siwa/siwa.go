package siwa

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jws"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/spf13/cobra"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

func NewSignInWithAppleCommand() *cobra.Command {
	var (
		privateKey string
		teamID     string
		servicesID string
		keyID      string
	)

	cmd := &cobra.Command{
		Use:   "siwa",
		Short: "Generate a client secret (JWT) for Sign in with Apple (SIWA)",
		Long: `Sign in with Apple requires JWTs to authorize requests. This command creates the token,
then signs it with the private key obtained from the Apple Developer console.

See: https://developer.apple.com/documentation/sign_in_with_apple/generate_and_validate_tokens#3262048`,
		Run: func(cmd *cobra.Command, args []string) {
			path := filepath.Clean(privateKey)
			bytes, err := ioutil.ReadFile(path)
			if err != nil {
				log.Fatal(err)
			}

			block, _ := pem.Decode(bytes)
			key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				log.Fatal(err)
			}

			now := time.Now().UTC()
			t := jwt.New()
			_ = t.Set(jwt.SubjectKey, servicesID)
			_ = t.Set(jwt.IssuerKey, teamID)
			_ = t.Set(jwt.IssuedAtKey, now.Unix())
			_ = t.Set(jwt.AudienceKey, "https://appleid.apple.com")
			_ = t.Set(jwt.ExpirationKey, now.Add(time.Hour*24*180).Unix())

			headers := jws.NewHeaders()
			_ = headers.Set(jws.KeyIDKey, keyID)

			signed, err := jwt.Sign(t, jwt.WithKey(jwa.ES256, key, jws.WithProtectedHeaders(headers)))
			fmt.Println(string(signed))
		},
	}

	cmd.Flags().StringVarP(&privateKey, "private_key", "p", "", "Path to the private key file")
	cmd.Flags().StringVarP(&teamID, "team_id", "t", "", "Apple team ID")
	cmd.Flags().StringVarP(&servicesID, "services_id", "s", "", "Apple services ID")
	cmd.Flags().StringVarP(&keyID, "key_id", "k", "", "Apple key ID")

	_ = cmd.MarkFlagRequired("private_key")
	_ = cmd.MarkFlagRequired("team_id")
	_ = cmd.MarkFlagRequired("services_id")
	_ = cmd.MarkFlagRequired("key_id")

	return cmd
}

func RegisterCommands(parent *cobra.Command) {
	parent.AddCommand(NewSignInWithAppleCommand())
}
