package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gobuffalo/pop/v6"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	saml2 "github.com/russellhaering/gosaml2"
	dsig "github.com/russellhaering/goxmldsig"
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

func TestSamlHandler_getAssertionInfo_reportsIdPStatus(t *testing.T) {
	sp := &saml2.SAMLServiceProvider{IDPCertificateStore: &dsig.MemoryX509CertificateStore{}}

	errorResponse := `<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" Version="2.0">
  <saml:Issuer>https://idp.example.com</saml:Issuer>
  <samlp:Status>
    <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Responder"/>
    <samlp:StatusMessage>Could not encrypt assertion</samlp:StatusMessage>
  </samlp:Status>
</samlp:Response>`

	h := &Handler{}

	t.Run("error response reports the status of the IdP", func(t *testing.T) {
		_, err := h.getAssertionInfo(sp, base64.StdEncoding.EncodeToString([]byte(errorResponse)))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid_request")
		assert.Contains(t, err.Error(), "urn:oasis:names:tc:SAML:2.0:status:Responder")
		assert.Contains(t, err.Error(), "Could not encrypt assertion")
		assert.Contains(t, err.Error(), "missing Assertion element")
	})

	t.Run("other failures are left as they are", func(t *testing.T) {
		_, err := h.getAssertionInfo(sp, "not a saml response")

		require.Error(t, err)
		assert.NotContains(t, err.Error(), "idp returned status")
	})
}

func TestSamlHandler_getAssertionInfo_reportsInvalidRequest(t *testing.T) {
	sp := &saml2.SAMLServiceProvider{IDPCertificateStore: &dsig.MemoryX509CertificateStore{}}

	_, err := (&Handler{}).getAssertionInfo(sp, "not a saml response")

	var thirdPartyErr *thirdparty.ThirdPartyError
	require.ErrorAs(t, err, &thirdPartyErr)
	assert.Equal(t, thirdparty.ErrorCodeInvalidRequest, thirdPartyErr.Code)
}
