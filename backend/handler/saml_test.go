package handler

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

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

// auditError must hand the error it was given, including its cause, to the audit logger.
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

// A response that carries a status but no assertion must report that status in the error.
func TestSamlHandler_getAssertionInfo_reportsIdPStatus(t *testing.T) {
	sp := &saml2.SAMLServiceProvider{IDPCertificateStore: &dsig.MemoryX509CertificateStore{}}

	h := &Handler{}

	t.Run("error response reports the status of the IdP", func(t *testing.T) {
		_, err := h.getAssertionInfo(sp, idpErrorResponse("urn:oasis:names:tc:SAML:2.0:status:Responder", "", "Could not encrypt assertion"))

		require.Error(t, err)
		assert.Contains(t, err.Error(), "response status")
		assert.Contains(t, err.Error(), "urn:oasis:names:tc:SAML:2.0:status:Responder")
		assert.Contains(t, err.Error(), "Could not encrypt assertion")
		assert.Contains(t, err.Error(), "missing Assertion element")
	})

	t.Run("other failures are left as they are", func(t *testing.T) {
		_, err := h.getAssertionInfo(sp, "not a saml response")

		require.Error(t, err)
		assert.NotContains(t, err.Error(), "response status")
	})
}

// The status message of a response is cut to 1024 characters, counting characters and not bytes.
func TestSamlHandler_getAssertionInfo_truncatesIdPStatusMessage(t *testing.T) {
	sp := &saml2.SAMLServiceProvider{IDPCertificateStore: &dsig.MemoryX509CertificateStore{}}
	message := strings.Repeat("ü", 2000)

	_, err := (&Handler{}).getAssertionInfo(sp, idpErrorResponse("urn:oasis:names:tc:SAML:2.0:status:Responder", "", message))

	require.Error(t, err)
	assert.Contains(t, err.Error(), strings.Repeat("ü", 1024))
	assert.NotContains(t, err.Error(), strings.Repeat("ü", 1025))
}

// The nested status code of a response is reported after the top-level one.
func TestSamlHandler_getAssertionInfo_reportsNestedStatusCode(t *testing.T) {
	sp := &saml2.SAMLServiceProvider{IDPCertificateStore: &dsig.MemoryX509CertificateStore{}}

	_, err := (&Handler{}).getAssertionInfo(sp, idpErrorResponse("urn:oasis:names:tc:SAML:2.0:status:Responder", "urn:oasis:names:tc:SAML:2.0:status:AuthnFailed", "Authentication failed"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), `"urn:oasis:names:tc:SAML:2.0:status:Responder/urn:oasis:names:tc:SAML:2.0:status:AuthnFailed"`)
}

// The status of a response is also reported if the library returns the missing assertion error without wrapping it.
func TestSamlHandler_withIdPStatus_reportsStatusOfBareError(t *testing.T) {
	err := withIdPStatus(idpErrorResponse("urn:oasis:names:tc:SAML:2.0:status:Responder", "", "Could not encrypt assertion"), saml2.ErrMissingAssertion)

	require.ErrorIs(t, err, saml2.ErrMissingAssertion)
	assert.Contains(t, err.Error(), "Could not encrypt assertion")
}

// An assertion that is not valid yet or has expired must be rejected, and the error must name its validity period.
func TestSamlHandler_getAssertionInfo_rejectsAssertionOutsideValidityPeriod(t *testing.T) {
	now := time.Now()

	tests := map[string]struct {
		notBefore    time.Time
		notOnOrAfter time.Time
	}{
		"expired":       {notBefore: now.Add(-2 * time.Hour), notOnOrAfter: now.Add(-time.Hour)},
		"not valid yet": {notBefore: now.Add(time.Hour), notOnOrAfter: now.Add(2 * time.Hour)},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, err := (&Handler{}).getAssertionInfo(assertionTestServiceProvider(), assertionResponse(tt.notBefore, tt.notOnOrAfter, testAudience))

			require.Error(t, err)
			assert.Contains(t, err.Error(), "SAMLAssertion is not valid at this time")
			assert.Contains(t, err.Error(), tt.notBefore.UTC().Format(time.RFC3339))
			assert.Contains(t, err.Error(), tt.notOnOrAfter.UTC().Format(time.RFC3339))
		})
	}
}

// An assertion that is meant for another audience must be rejected.
func TestSamlHandler_getAssertionInfo_rejectsAssertionForOtherAudience(t *testing.T) {
	now := time.Now()

	_, err := (&Handler{}).getAssertionInfo(assertionTestServiceProvider(), assertionResponse(now.Add(-time.Hour), now.Add(time.Hour), "https://other.example.com"))

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not in SAML audience")
}

const (
	testAudience = "https://sp.example.com"
	testACSURL   = "https://sp.example.com/saml/callback"
)

// assertionTestServiceProvider returns a service provider that accepts an unsigned response with an assertion.
func assertionTestServiceProvider() *saml2.SAMLServiceProvider {
	return &saml2.SAMLServiceProvider{
		SkipSignatureValidation:     true,
		AllowMissingAttributes:      true,
		Clock:                       dsig.NewRealClock(),
		AudienceURI:                 testAudience,
		AssertionConsumerServiceURL: testACSURL,
	}
}

// assertionResponse returns a base64 encoded, unsigned SAML response with a successful status and one assertion
// that has the given validity period and audience.
func assertionResponse(notBefore time.Time, notOnOrAfter time.Time, audience string) string {
	response := fmt.Sprintf(`<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" Version="2.0">
  <saml:Issuer>https://idp.example.com</saml:Issuer>
  <samlp:Status>
    <samlp:StatusCode Value="urn:oasis:names:tc:SAML:2.0:status:Success"/>
  </samlp:Status>
  <saml:Assertion Version="2.0">
    <saml:Issuer>https://idp.example.com</saml:Issuer>
    <saml:Subject>
      <saml:NameID>user@example.com</saml:NameID>
      <saml:SubjectConfirmation Method="urn:oasis:names:tc:SAML:2.0:cm:bearer">
        <saml:SubjectConfirmationData Recipient="%s" NotOnOrAfter="%s"/>
      </saml:SubjectConfirmation>
    </saml:Subject>
    <saml:Conditions NotBefore="%s" NotOnOrAfter="%s">
      <saml:AudienceRestriction>
        <saml:Audience>%s</saml:Audience>
      </saml:AudienceRestriction>
    </saml:Conditions>
  </saml:Assertion>
</samlp:Response>`, testACSURL, time.Now().Add(time.Hour).UTC().Format(time.RFC3339), notBefore.UTC().Format(time.RFC3339), notOnOrAfter.UTC().Format(time.RFC3339), audience)

	return base64.StdEncoding.EncodeToString([]byte(response))
}

// idpErrorResponse returns a base64 encoded, unsigned SAML response that has a status and no assertion. The
// nested status code is left out if it is empty.
func idpErrorResponse(statusCode string, nestedStatusCode string, message string) string {
	nested := ""
	if nestedStatusCode != "" {
		nested = fmt.Sprintf(`<samlp:StatusCode Value="%s"/>`, nestedStatusCode)
	}

	response := fmt.Sprintf(`<samlp:Response xmlns:samlp="urn:oasis:names:tc:SAML:2.0:protocol" xmlns:saml="urn:oasis:names:tc:SAML:2.0:assertion" Version="2.0">
  <saml:Issuer>https://idp.example.com</saml:Issuer>
  <samlp:Status>
    <samlp:StatusCode Value="%s">%s</samlp:StatusCode>
    <samlp:StatusMessage>%s</samlp:StatusMessage>
  </samlp:Status>
</samlp:Response>`, statusCode, nested, message)

	return base64.StdEncoding.EncodeToString([]byte(response))
}
