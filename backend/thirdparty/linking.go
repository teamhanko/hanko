package thirdparty

import (
	"fmt"
	"strings"
	"time"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/persistence"
	"github.com/teamhanko/hanko/backend/v2/persistence/models"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
)

type AccountLinkingResult struct {
	Type         models.AuditLogType
	User         *models.User
	WebhookEvent *events.Event
	UserCreated  bool
}

func LinkAccount(tx *pop.Connection, cfg *config.Config, p persistence.Persister, userData *UserData, providerID string, isSaml bool, samlDomain *string, isFlow bool, userID *uuid.UUID) (*AccountLinkingResult, error) {
	if !isFlow {
		if cfg.Email.RequireVerification && !userData.Metadata.EmailVerified {
			return nil, ErrorUnverifiedProviderEmail("third party provider email must be verified")
		}
	}

	// Validate userData
	if userData == nil {
		return nil, ErrorInvalidRequest("user data must be set")
	}

	// Ensure the email is lowercase to avoid case sensitivity issues
	userData.Metadata.Email = strings.ToLower(userData.Metadata.Email)

	identity, err := p.GetIdentityPersister().Get(userData.Metadata.Subject, providerID)
	if err != nil {
		return nil, ErrorServer("could not get identity").WithCause(err)
	}

	if identity == nil {
		var user *models.User
		if userID != nil {
			user, err = p.GetUserPersisterWithConnection(tx).Get(*userID)
		} else {
			user, err = p.GetUserPersisterWithConnection(tx).GetByEmailAddress(userData.Metadata.Email)
		}
		if err != nil {
			return nil, ErrorServer("could not get email").WithCause(err)
		}

		if user == nil {
			return signUp(tx, cfg, p, userData, providerID, isSaml, samlDomain)
		} else {
			return link(tx, cfg, p, userData, providerID, user, isSaml, samlDomain, userID != nil)
		}
	} else {
		return signIn(tx, cfg, p, userData, identity)
	}
}

func link(tx *pop.Connection, cfg *config.Config, p persistence.Persister, userData *UserData, providerID string, user *models.User, isSaml bool, samlDomain *string, comesFromProfile bool) (*AccountLinkingResult, error) {
	if !isSaml {
		if strings.HasPrefix(providerID, "custom_") {
			provider, ok := cfg.ThirdParty.CustomProviders[strings.TrimPrefix(providerID, "custom_")]
			if !ok {
				return nil, ErrorServer(fmt.Sprintf("unknown provider: %s", providerID))
			}
			if !provider.AllowLinking {
				return nil, ErrorUserConflict("third party account linking for existing user with same email disallowed")
			}
		} else {
			provider := cfg.ThirdParty.Providers.Get(providerID)
			if provider == nil {
				return nil, fmt.Errorf("unknown provider: %s", providerID)
			}

			if !provider.AllowLinking {
				return nil, ErrorUserConflict("third party account linking for existing user with same email disallowed")
			}
		}
	}

	var emailID *uuid.UUID
	email := user.GetEmailByAddress(userData.Metadata.Email)
	if email != nil {
		emailID = &email.ID
	}
	var userID *uuid.UUID = nil
	if !comesFromProfile {
		userID = &user.ID
	}

	// TODO: when email is nil, we should create a new email and associate it with the identity

	userDataMap, err := userData.ToMap()
	if err != nil {
		return nil, ErrorServer("could not link account").WithCause(err)
	}

	identity, err := models.NewIdentity(providerID, userDataMap, emailID, userID)
	if err != nil {
		return nil, ErrorServer("could not create identity").WithCause(err)
	}

	err = p.GetIdentityPersisterWithConnection(tx).Create(*identity)
	if err != nil {
		return nil, ErrorServer("could not create identity").WithCause(err)
	}

	if isSaml && samlDomain != nil && *samlDomain != "" && email != nil {
		if existingSamlIdentity := email.GetSamlIdentityForDomain(*samlDomain); existingSamlIdentity != nil {
			identityToDeleteID := existingSamlIdentity.IdentityID
			existingSamlIdentity.IdentityID = identity.ID

			err = p.GetSamlIdentityPersisterWithConnection(tx).Update(*existingSamlIdentity)
			if err != nil {
				return nil, ErrorServer("could not update saml identity").WithCause(err)
			}

			err = p.GetIdentityPersisterWithConnection(tx).Delete(models.Identity{ID: identityToDeleteID})
			if err != nil {
				return nil, ErrorServer("could not delete identity").WithCause(err)
			}
		} else {
			samlIdentityID, _ := uuid.NewV4()
			now := time.Now().UTC()
			samlIdentity := &models.SamlIdentity{
				ID:         samlIdentityID,
				IdentityID: identity.ID,
				Domain:     *samlDomain,
				CreatedAt:  now,
				UpdatedAt:  now,
			}

			err = p.GetSamlIdentityPersisterWithConnection(tx).Create(*samlIdentity)
			if err != nil {
				return nil, ErrorServer("could not create saml identity").WithCause(err)
			}
		}
	}

	u, terr := p.GetUserPersisterWithConnection(tx).Get(user.ID)
	if terr != nil {
		return nil, ErrorServer("could not get user").WithCause(terr)
	}

	return &AccountLinkingResult{
		Type:         models.AuditLogThirdPartyLinkingSucceeded,
		User:         u,
		WebhookEvent: nil,
		UserCreated:  false,
	}, nil
}

func signIn(tx *pop.Connection, cfg *config.Config, p persistence.Persister, userData *UserData, identity *models.Identity) (*AccountLinkingResult, error) {
	var linkingResult *AccountLinkingResult
	var webhookEvent events.Event

	userPersister := p.GetUserPersisterWithConnection(tx)
	emailPersister := p.GetEmailPersisterWithConnection(tx)
	identityPersister := p.GetIdentityPersisterWithConnection(tx)

	var terr error
	email := identity.Email
	if userData.Metadata.Email != "" && ((email == nil) || (email.Address != userData.Metadata.Email)) {
		// The primary email address at the provider has changed, check if the new provider email already exists
		email, terr = emailPersister.FindByAddress(userData.Metadata.Email)
		if terr != nil {
			return nil, ErrorServer("could not get email").WithCause(terr)
		}

		if email != nil {
			if email.UserID == nil {
				// The email already exists but is unassigned, claim it and associate the identity with it
				email.UserID = identity.Email.UserID
				email.Verified = userData.Metadata.EmailVerified

				terr = emailPersister.Update(*email)
				if terr != nil {
					return nil, ErrorServer("could not update email").WithCause(terr)
				}

				identity.EmailID = &email.ID
				webhookEvent = events.UserUpdate
			} else if email.UserID.String() != identity.Email.UserID.String() {
				// The email is assigned to a different user, and so the identity is linked to multiple users. There
				// is not much we can do here but return an error.
				return nil, ErrorMultipleAccounts(fmt.Sprintf("cannot identify associated user: '%s' is used by multiple accounts", email.Address))
			} else {
				// The email is assigned to the same user. This can happen if the user creates an email with an
				// address equal to the new primary provider email prior to changing the primary mail at the
				// provider and then doing a sign in with the provider. We need to update the associated email in
				// the identity.
				identity.EmailID = &email.ID
			}
		} else {
			// The email does not exist. Create a new one and associate the identity with it.
			emailCount, err := emailPersister.CountByUserId(*identity.UserID)
			if err != nil {
				return nil, ErrorServer("failed to count user emails").WithCause(err)
			}

			if emailCount >= cfg.Email.Limit {
				return nil, ErrorMaxNumberOfAddresses("max number of email addresses reached")
			}

			email = models.NewEmail(identity.UserID, userData.Metadata.Email)
			email.Verified = userData.Metadata.EmailVerified
			terr = emailPersister.Create(*email)
			if terr != nil {
				return nil, ErrorServer("could not create email").WithCause(terr)
			}

			identity.EmailID = &email.ID
			webhookEvent = events.UserEmailCreate
		}
	}

	userDataMap, err := userData.ToMap()
	if err != nil {
		return nil, ErrorServer("could not link account").WithCause(err)
	}

	identity.Data = userDataMap
	terr = identityPersister.Update(*identity)
	if terr != nil {
		return nil, ErrorServer("could not update identity").WithCause(terr)
	}

	user, terr := userPersister.Get(*identity.UserID)
	if terr != nil {
		return nil, ErrorServer("could not get user").WithCause(terr)
	}

	linkingResult = &AccountLinkingResult{
		Type:         models.AuditLogThirdPartySignInSucceeded,
		User:         user,
		WebhookEvent: &webhookEvent,
		UserCreated:  false,
	}

	return linkingResult, nil
}

func signUp(tx *pop.Connection, cfg *config.Config, p persistence.Persister, userData *UserData, providerID string, isSaml bool, samlDomain *string) (*AccountLinkingResult, error) {
	if !cfg.Account.AllowSignup {
		return nil, ErrorSignUpDisabled("account signup is disabled")
	}

	var linkingResult *AccountLinkingResult

	userPersister := p.GetUserPersisterWithConnection(tx)
	emailPersister := p.GetEmailPersisterWithConnection(tx)
	primaryEmailPersister := p.GetPrimaryEmailPersisterWithConnection(tx)
	identityPersister := p.GetIdentityPersisterWithConnection(tx)

	email, terr := emailPersister.FindByAddress(userData.Metadata.Email)
	if terr != nil {
		return nil, ErrorServer("could not get email").WithCause(terr)
	}

	user := models.NewUser()
	terr = userPersister.Create(user)
	if terr != nil {
		return nil, ErrorServer("could not create user").WithCause(terr)
	}

	if email != nil && email.UserID == nil {
		// There exists an email with the same address as the primary provider address, but it is not assigned
		// to any user yet, hence we assign the new user ID to this email.
		email.UserID = &user.ID
		email.Verified = userData.Metadata.EmailVerified
		terr = emailPersister.Update(*email)
		if terr != nil {
			return nil, ErrorServer("could not update email").WithCause(terr)
		}

	} else if userData.Metadata.Email != "" {
		// No email exists, create a new one using the provider user data email
		email = models.NewEmail(&user.ID, userData.Metadata.Email)
		email.Verified = userData.Metadata.EmailVerified
		terr = emailPersister.Create(*email)
		if terr != nil {
			return nil, ErrorServer("failed to store email").WithCause(terr)
		}
	}

	var emailID *uuid.UUID = nil
	if email != nil {
		emailID = &email.ID
		primaryEmail := models.NewPrimaryEmail(email.ID, *email.UserID)
		terr = primaryEmailPersister.Create(*primaryEmail)
		if terr != nil {
			return nil, ErrorServer("failed to store primary email").WithCause(terr)
		}
	}

	userDataMap, err := userData.ToMap()
	if err != nil {
		return nil, ErrorServer("could not link account").WithCause(err)
	}

	identity, terr := models.NewIdentity(providerID, userDataMap, emailID, &user.ID)
	if terr != nil {
		return nil, ErrorServer("could not create identity").WithCause(terr)
	}

	terr = identityPersister.Create(*identity)
	if terr != nil {
		return nil, ErrorServer("could not store identity").WithCause(terr)
	}

	if isSaml && samlDomain != nil && *samlDomain != "" {
		samlIdentityID, _ := uuid.NewV4()
		now := time.Now().UTC()
		samlIdentity := &models.SamlIdentity{
			ID:         samlIdentityID,
			IdentityID: identity.ID,
			Domain:     *samlDomain,
			CreatedAt:  now,
			UpdatedAt:  now,
		}

		err = p.GetSamlIdentityPersisterWithConnection(tx).Create(*samlIdentity)
		if err != nil {
			return nil, ErrorServer("could not store saml identity").WithCause(err)
		}
	}

	u, terr := userPersister.Get(user.ID)
	if terr != nil {
		return nil, ErrorServer("could not get user").WithCause(terr)
	}

	evt := events.UserCreate
	linkingResult = &AccountLinkingResult{
		Type:         models.AuditLogThirdPartySignUpSucceeded,
		User:         u,
		WebhookEvent: &evt,
		UserCreated:  true,
	}

	return linkingResult, nil
}
