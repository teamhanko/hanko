package jwt

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/spf13/cobra"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/crypto/jwk"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/session"
	"log"
)

func NewCreateCommand() *cobra.Command {
	var (
		configFile string
	)

	cmd := &cobra.Command{
		Use:   "create [user_id]",
		Short: "generate a JSON Web Token for a given user_id",
		Long:  ``,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("user_id required")
			}
			if _, err := uuid.FromString(args[0]); err != nil {
				return errors.New("user_id is not a uuid")
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg, err := config.Load(&configFile)
			if err != nil {
				log.Fatal(err)
			}
			persister, err := persistence.New(cfg.Database)
			if err != nil {
				log.Fatal(err)
			}
			jwkPersister := persister.GetJwkPersister()
			jwkManager, err := jwk.NewDefaultManager(cfg.Secrets.Keys, jwkPersister)
			if err != nil {
				fmt.Printf("failed to create jwk persister: %s", err)
				return
			}

			sessionManager, err := session.NewManager(jwkManager, *cfg)
			if err != nil {
				fmt.Printf("failed to create session generator: %s", err)
				return
			}

			userId := uuid.FromStringOrNil(args[0])

			emails, err := persister.GetEmailPersister().FindByUserId(userId)
			if err != nil {
				fmt.Printf("failed to get emails from db: %s", err)
				return
			}

			var emailJwt *dto.EmailJwt
			if e := emails.GetPrimary(); e != nil {
				emailJwt = dto.JwtFromEmailModel(e)
			}

			token, rawToken, err := sessionManager.GenerateJWT(userId, emailJwt)
			if err != nil {
				fmt.Printf("failed to generate token: %s", err)
				return
			}

			if cfg.Session.ServerSide.Enabled {
				sessionID, _ := rawToken.Get("session_id")

				expirationTime := rawToken.Expiration()
				sessionModel := models.Session{
					ID:        uuid.FromStringOrNil(sessionID.(string)),
					UserID:    userId,
					UserAgent: "",
					IpAddress: "",
					CreatedAt: rawToken.IssuedAt(),
					UpdatedAt: rawToken.IssuedAt(),
					ExpiresAt: &expirationTime,
					LastUsed:  rawToken.IssuedAt(),
				}

				err = persister.GetSessionPersister().Create(sessionModel)
				if err != nil {
					fmt.Printf("failed to store session: %s", err)
					return
				}
			}

			fmt.Printf("token: %s", token)
		},
	}

	cmd.Flags().StringVar(&configFile, "config", "", "config file")

	return cmd
}
