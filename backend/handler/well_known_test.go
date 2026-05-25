package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/v2/test"
)

func TestWellKnownSuite(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(wellKnownSuite))
}

type wellKnownSuite struct {
	test.Suite
}

func (s *wellKnownSuite) TestWellKnownHandler_GetPublicKeys() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}
	cfg := test.DefaultConfig
	err := cfg.PostProcess()
	s.Require().NoError(err)

	e := NewPublicRouter(&cfg, s.Storage, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/.well-known/jwks.json", nil)
	rec := httptest.NewRecorder()

	e.ServeHTTP(rec, req)

	s.Equal(http.StatusOK, rec.Code)
	s.Equal("max-age=600", rec.Header().Get("Cache-Control"))
}
