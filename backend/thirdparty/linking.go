package thirdparty

import (
	"fmt"
	"github.com/gobuffalo/pop/v6"
	"github.com/teamhanko/hanko/backend/config"
	"github.com/teamhanko/hanko/backend/persistence"
	"github.com/teamhanko/hanko/backend/persistence/models"
)

type AccountLinkingResult struct {
	Type models.AuditLogType
	User *models.User
}

func LinkAccount(tx *pop.Connection, cfg *config.Config, p persistence.Persister, userData *UserData, providerName string) (*AccountLinkingResult, error) {
	identity, err := p.GetIdentityPersister().Get(userData.Metadata.Subject, providerName)
	if err != nil {
		return nil, ErrorServer("could not get identity").WithCause(err)
	}

	if cfg.Emails.RequireVerification && !userData.Metadata.EmailVerified {
		return nil, ErrorUnverifiedProviderEmail("third party provider email must be verified")
	}

	if identity == nil {
		return signUp(tx, p, userData, providerName)
	} else {
		return signIn(tx, cfg, p, userData, identity)
	}
}

func signIn(tx *pop.Connection, cfg *config.Config, p persistence.Persister, userData *UserData, identity *models.Identity) (*AccountLinkingResult, error) {
	var linkingResult *AccountLinkingResult

	userPersister := p.GetUserPersisterWithConnection(tx)
	emailPersister := p.GetEmailPersisterWithConnection(tx)
	identityPersister := p.GetIdentityPersisterWithConnection(tx)

	var terr error
	email := identity.Email
	if userData.Metadata.Email != identity.Email.Address {
		// The primary email address at the provider has changed, check if the new provider email already exists
		email, terr = emailPersister.FindByAddress(userData.Metadata.Email)
		if terr != nil {
			return nil, ErrorServer("could not get email").WithCause(terr)
		}

		if email != nil {
			if email.UserID == nil {
				// The email already exists but is unassigned, claim it and associate the identity with it
				email.UserID = identity.Email.UserID
				email.Verified = true

				terr = emailPersister.Update(*email)
				if terr != nil {
					return nil, ErrorServer("could not update email").WithCause(terr)
				}

				identity.EmailID = email.ID
			} else if email.UserID.String() != identity.Email.UserID.String() {
				// The email is assigned to a different user, and so the identity is linked to multiple users. There
				// is not much we can do here but return an error.
				return nil, ErrorMultipleAccounts(fmt.Sprintf("cannot identify associated user: '%s' is used by multiple accounts", email.Address))
			} else {
				// The email is assigned to the same user. This can happen if the user creates an email with an
				// address equal to the new primary provider email prior to changing the primary mail at the
				// provider and then doing a sign in with the provider. We need to update the associated email in
				// the identity.
				identity.EmailID = email.ID
			}
		} else {
			// The email does not exist. Create a new one and associate the identity with it.
			emailCount, err := emailPersister.CountByUserId(*identity.Email.UserID)
			if err != nil {
				return nil, ErrorServer("failed to count user emails").WithCause(err)
			}

			if emailCount >= cfg.Emails.MaxNumOfAddresses {
				return nil, ErrorMaxNumberOfAddresses("max number of email addresses reached")
			}

			email = models.NewEmail(identity.Email.UserID, userData.Metadata.Email)
			email.Verified = true
			terr = emailPersister.Create(*email)
			if terr != nil {
				return nil, ErrorServer("could not create email").WithCause(terr)
			}

			identity.EmailID = email.ID
		}
	}

	identity.Data = userData.ToMap()
	terr = identityPersister.Update(*identity)
	if terr != nil {
		return nil, ErrorServer("could not get identity").WithCause(terr)
	}

	user, terr := userPersister.Get(*identity.Email.UserID)
	if terr != nil {
		return nil, ErrorServer("could not get user").WithCause(terr)
	}

	linkingResult = &AccountLinkingResult{
		Type: models.AuditLogThirdPartySignInSucceeded,
		User: user,
	}

	return linkingResult, nil
}

func signUp(tx *pop.Connection, p persistence.Persister, userData *UserData, providerName string) (*AccountLinkingResult, error) {
	var linkingResult *AccountLinkingResult

	userPersister := p.GetUserPersisterWithConnection(tx)
	emailPersister := p.GetEmailPersisterWithConnection(tx)
	primaryEmailPersister := p.GetPrimaryEmailPersisterWithConnection(tx)
	identityPersister := p.GetIdentityPersisterWithConnection(tx)

	email, terr := emailPersister.FindByAddress(userData.Metadata.Email)
	if terr != nil {
		return nil, ErrorServer("could not get email").WithCause(terr)
	}

	if email != nil && email.UserID != nil {
		return nil, ErrorUserConflict("third party account linking for existing user with same email disallowed")
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
		email.Verified = true
		terr = emailPersister.Update(*email)
		if terr != nil {
			return nil, ErrorServer("could not update email").WithCause(terr)
		}
	} else {
		// No email exists, create a new one using the provider user data email
		email = models.NewEmail(&user.ID, userData.Metadata.Email)
		email.Verified = true
		terr = emailPersister.Create(*email)
		if terr != nil {
			return nil, ErrorServer("failed to store email").WithCause(terr)
		}
	}

	primaryEmail := models.NewPrimaryEmail(email.ID, *email.UserID)
	terr = primaryEmailPersister.Create(*primaryEmail)
	if terr != nil {
		return nil, ErrorServer("failed to store primary email").WithCause(terr)
	}

	identity, terr := models.NewIdentity(providerName, userData.ToMap(), email.ID)
	if terr != nil {
		return nil, ErrorServer("could not create identity").WithCause(terr)
	}

	terr = identityPersister.Create(*identity)
	if terr != nil {
		return nil, ErrorServer("could not create identity").WithCause(terr)
	}

	u, terr := userPersister.Get(*email.UserID)
	if terr != nil {
		return nil, ErrorServer("could not get user").WithCause(terr)
	}

	linkingResult = &AccountLinkingResult{
		Type: models.AuditLogThirdPartySignUpSucceeded,
		User: u,
	}

	return linkingResult, nil
}
