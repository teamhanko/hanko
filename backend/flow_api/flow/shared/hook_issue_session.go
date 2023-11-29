package shared

import (
	"errors"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/teamhanko/hanko/backend/flowpilot"
)

type IssueSession struct {
	Action
}

func (h IssueSession) Execute(c flowpilot.HookExecutionContext) error {
	deps := h.GetDeps(c)

	var userId uuid.UUID
	var err error
	if c.Stash().Get("user_id").Exists() {
		userId, err = uuid.FromString(c.Stash().Get("user_id").String())
		if err != nil {
			return fmt.Errorf("failed to parse stashed user_id into a uuid: %w", err)
		}
	} else {
		return errors.New("user_id not found in stash")
	}

	sessionToken, err := deps.SessionManager.GenerateJWT(userId)
	if err != nil {
		return fmt.Errorf("failed to generate JWT: %w", err)
	}
	cookie, err := deps.SessionManager.GenerateCookie(sessionToken)
	if err != nil {
		return fmt.Errorf("failed to generate auth cookie, %w", err)
	}

	deps.HttpContext.SetCookie(cookie)

	return nil
}
