package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	auditlog "github.com/teamhanko/hanko/backend/v3/audit_log"
	"github.com/teamhanko/hanko/backend/v3/context"
	"github.com/teamhanko/hanko/backend/v3/persistence/models"
	"github.com/teamhanko/hanko/backend/v3/thirdparty"
)

type recordedAuditEntry struct {
	logType  models.AuditLogType
	logError error
}

type recordingAuditLogger struct {
	entries []recordedAuditEntry
}

func (r *recordingAuditLogger) Create(_ echo.Context, logType models.AuditLogType, _ *models.User, logError error, _ uuid.UUID, _ ...auditlog.DetailOption) error {
	r.entries = append(r.entries, recordedAuditEntry{logType: logType, logError: logError})
	return nil
}

func (r *recordingAuditLogger) CreateWithConnection(_ *pop.Connection, c echo.Context, logType models.AuditLogType, user *models.User, logError error, tenantID uuid.UUID, opts ...auditlog.DetailOption) error {
	return r.Create(c, logType, user, logError, tenantID, opts...)
}

func TestSamlHandler_auditError_carriesOriginalError(t *testing.T) {
	logger := &recordingAuditLogger{}
	h := &Handler{auditLogger: logger}

	c := echo.New().NewContext(httptest.NewRequest(http.MethodPost, "/saml/callback", nil), httptest.NewRecorder())
	c.Set("tenant", context.Tenant{ID: uuid.Must(uuid.NewV4())})

	cause := fmt.Errorf("assertion rejected")
	err := h.auditError(c, thirdparty.ErrorInvalidRequest("could not get assertion info").WithCause(cause))

	require.NoError(t, err)
	require.Len(t, logger.entries, 1)
	assert.Equal(t, models.AuditLogThirdPartySignInSignUpFailed, logger.entries[0].logType)
	require.Error(t, logger.entries[0].logError)
	assert.Contains(t, logger.entries[0].logError.Error(), "could not get assertion info")
	assert.Contains(t, logger.entries[0].logError.Error(), cause.Error())
}
