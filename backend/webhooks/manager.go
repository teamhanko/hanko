package webhooks

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/teamhanko/hanko/backend/config"
	hankoJwk "github.com/teamhanko/hanko/backend/crypto/jwk"
	hankoJwt "github.com/teamhanko/hanko/backend/crypto/jwt"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"time"
)

type Manager interface {
	Trigger(tx *pop.Connection, evt events.Event, data interface{})
	GenerateJWT(data interface{}, event events.Event) (string, error)
}

type manager struct {
	logger          echo.Logger
	webhooks        Webhooks
	jwtGenerator    hankoJwt.Generator
	audience        []string
	persister       persistence.Persister
	canExpireAtTime bool
}

func NewManager(cfg *config.Config, persister persistence.Persister, jwkManager hankoJwk.Manager, logger echo.Logger) (Manager, error) {
	hooks := make(Webhooks, 0)

	if cfg.Webhooks.Enabled {
		for _, cfgHook := range cfg.Webhooks.Hooks {
			hooks = append(hooks, NewConfigHook(cfgHook, logger))
		}
	}

	const generateFailureMessage = "failed to create webhook jwt generator: %w"

	signatureKey, err := jwkManager.GetSigningKey()
	if err != nil {
		errMessage := fmt.Errorf(generateFailureMessage, err)
		logger.Error(errMessage)
		return nil, errMessage
	}
	verificationKeys, err := jwkManager.GetPublicKeys()
	if err != nil {
		errMessage := fmt.Errorf(generateFailureMessage, err)
		logger.Error(errMessage)
		return nil, errMessage
	}
	g, err := hankoJwt.NewGenerator(signatureKey, verificationKeys)
	if err != nil {
		errMessage := fmt.Errorf(generateFailureMessage, err)
		logger.Error(errMessage)
		return nil, errMessage
	}

	var audience []string
	if cfg.Session.Audience != nil && len(cfg.Session.Audience) > 0 {
		audience = cfg.Session.Audience
	} else {
		audience = []string{cfg.Webauthn.RelyingParty.Id}
	}

	return &manager{
		logger:          logger,
		webhooks:        hooks,
		jwtGenerator:    g,
		audience:        audience,
		persister:       persister,
		canExpireAtTime: cfg.Webhooks.AllowTimeExpiration,
	}, nil
}

func (m *manager) Trigger(tx *pop.Connection, evt events.Event, data interface{}) {
	// add db hooks - Done here to prevent a restart in case a hook is added or removed from the database
	dbHooks, err := m.persister.GetWebhookPersister(tx).List(false)
	if err != nil {
		m.logger.Error(fmt.Errorf("unable to get database webhooks: %w", err))
		return
	}

	hooks := m.webhooks
	for _, dbHook := range dbHooks {
		hooks = append(hooks, NewDatabaseHook(dbHook, m.persister.GetWebhookPersister(nil), m.logger))
	}

	dataToken, err := m.GenerateJWT(data, evt)
	if err != nil {
		m.logger.Error(fmt.Errorf("unable to generate JWT for webhook data: %w", err))
		return
	}

	jobData := JobData{
		Token: dataToken,
		Event: evt,
	}

	hookChannel := make(chan Job, len(hooks))
	for _, hook := range hooks {
		if hook.HasEvent(evt) {
			job := Job{
				Data:            jobData,
				Hook:            hook,
				CanExpireAtTime: m.canExpireAtTime,
			}
			hookChannel <- job
		}
	}
	close(hookChannel)

	worker := NewWorker(hookChannel, m.logger)
	go worker.Run()
}

func (m *manager) GenerateJWT(data interface{}, event events.Event) (string, error) {
	issuedAt := time.Now()
	expiration := issuedAt.Add(5 * time.Minute)

	token := jwt.New()
	_ = token.Set(jwt.SubjectKey, "hanko webhooks")
	_ = token.Set(jwt.IssuedAtKey, issuedAt)
	_ = token.Set(jwt.ExpirationKey, expiration)
	_ = token.Set(jwt.AudienceKey, m.audience)
	_ = token.Set("data", data)
	_ = token.Set("evt", event)

	signed, err := m.jwtGenerator.Sign(token)
	if err != nil {
		return "", err
	}

	return string(signed), nil
}
