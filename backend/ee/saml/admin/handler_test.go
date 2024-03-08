package admin

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/suite"
	"github.com/teamhanko/hanko/backend/config"
	baseDto "github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/ee/saml/dto"
	"github.com/teamhanko/hanko/backend/persistence/models"
	"github.com/teamhanko/hanko/backend/test"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSamlAdminHandler(t *testing.T) {
	t.Parallel()
	suite.Run(t, new(samlAdminHandlerSuite))
}

type samlAdminHandlerSuite struct {
	test.Suite
	server *echo.Echo
}

func (s *samlAdminHandlerSuite) setupServer() {
	e := echo.New()

	e.Validator = baseDto.NewCustomValidator()
	e.HTTPErrorHandler = baseDto.NewHTTPErrorHandler(baseDto.HTTPErrorHandlerConfig{Debug: false, Logger: e.Logger})

	cfg := config.DefaultConfig()
	handler := NewSamlAdminHandler(cfg, s.Storage)

	routingGroup := e.Group("saml")
	routingGroup.GET("", handler.List)
	routingGroup.POST("", handler.Create)

	singleProviderGroup := routingGroup.Group("/:id")
	singleProviderGroup.GET("", handler.Get)
	singleProviderGroup.PUT("", handler.Update)
	singleProviderGroup.DELETE("", handler.Delete)

	s.server = e
}

func (s *samlAdminHandlerSuite) TestSamlAdminHandler_New() {
	handler := NewSamlAdminHandler(&config.Config{}, s.Storage)
	s.Assert().NotEmpty(handler)
}

func (s *samlAdminHandlerSuite) TestSamlAdminHandler_List() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	// given
	err := s.LoadFixtures("../../../test/fixtures/saml")
	s.Require().NoError(err)

	s.setupServer()

	expectedProviders, err := s.Storage.GetSamlIdentityProviderPersister(nil).List()
	s.Require().NoError(err)

	// when
	req := httptest.NewRequest(http.MethodGet, "/saml", nil)
	rec := httptest.NewRecorder()

	s.server.ServeHTTP(rec, req)

	var providers models.SamlIdentityProviders
	err = json.Unmarshal(rec.Body.Bytes(), &providers)
	s.Require().NoError(err)

	// then
	s.Assert().Equal(http.StatusOK, rec.Code)
	s.Assert().Len(providers, 2)
	s.Assert().Equal(expectedProviders[0].ID, providers[0].ID)
	s.Assert().Equal(expectedProviders[1].ID, providers[1].ID)
}

func (s *samlAdminHandlerSuite) TestSamlAdminHandler_ListWithEmptyDb() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	// given
	s.setupServer()

	// when
	req := httptest.NewRequest(http.MethodGet, "/saml", nil)
	rec := httptest.NewRecorder()

	s.server.ServeHTTP(rec, req)

	var providers models.SamlIdentityProviders
	err := json.Unmarshal(rec.Body.Bytes(), &providers)
	s.Require().NoError(err)

	// then
	s.Assert().Equal(http.StatusOK, rec.Code)
	s.Assert().Len(providers, 0)
}

func (s *samlAdminHandlerSuite) TestSamlAdminHandler_Create() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../../../test/fixtures/saml")
	s.Require().NoError(err)

	s.setupServer()

	tests := []struct {
		Name          string
		Data          interface{}
		ExpectedCode  int
		ExpectedError string
		ContainsError bool
	}{
		{
			Name: "Success",
			Data: dto.SamlCreateProviderRequest{
				Enabled:               true,
				Name:                  "Testprovider",
				Domain:                "hanko2.io",
				MetadataUrl:           "https://hanko.io/metadata",
				SkipEmailVerification: false,
				AttributeMap: &dto.SamlCreateProviderAttributeMapRequest{
					Name:              "name",
					FamilyName:        "family_name",
					GivenName:         "given_name",
					MiddleName:        "middle_name",
					NickName:          "nickname",
					PreferredUsername: "preferred_username",
					Profile:           "profile",
					Picture:           "picture",
					Website:           "website",
					Gender:            "gender",
					Birthdate:         "born_at",
					ZoneInfo:          "info",
					Locale:            "locale",
					UpdatedAt:         "last_update",
					Email:             "email",
					EmailVerified:     "verified_email",
					Phone:             "phone",
					PhoneVerified:     "verified_phone",
				},
			},
			ExpectedCode:  http.StatusCreated,
			ContainsError: false,
		},
		{
			Name:          "Bind Error",
			Data:          "lorem",
			ExpectedCode:  http.StatusBadRequest,
			ExpectedError: bindRequestError,
			ContainsError: true,
		},
		{
			Name: "Validation Error",
			Data: dto.SamlCreateProviderRequest{
				Enabled:               true,
				Name:                  "Testprovider",
				Domain:                "hanko2.io",
				MetadataUrl:           "aaa",
				SkipEmailVerification: false,
			},
			ExpectedCode:  http.StatusBadRequest,
			ExpectedError: validateRequestError,
			ContainsError: true,
		},
		{
			Name: "Already Existing Domain Error",
			Data: dto.SamlCreateProviderRequest{
				Enabled:               true,
				Name:                  "Testprovider",
				Domain:                "hanko.io",
				MetadataUrl:           "https://hanko.io/metadata",
				SkipEmailVerification: false,
			},
			ExpectedCode:  http.StatusConflict,
			ExpectedError: "a provider with the domain 'hanko.io' already exists",
			ContainsError: true,
		},
	}

	for _, samlTest := range tests {
		s.T().Run(samlTest.Name, func(t *testing.T) {
			// given
			dataJson, err := json.Marshal(samlTest.Data)
			s.Require().NoError(err)

			// when
			req := httptest.NewRequest(http.MethodPost, "/saml", bytes.NewReader(dataJson))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			s.server.ServeHTTP(rec, req)

			// then
			s.Assert().Equal(samlTest.ExpectedCode, rec.Code)
			if samlTest.ContainsError {
				var hErr echo.HTTPError
				err = json.Unmarshal(rec.Body.Bytes(), &hErr)
				s.Require().NoError(err)

				s.Assert().Equal(samlTest.ExpectedError, hErr.Message)

			} else {
				data := samlTest.Data.(dto.SamlCreateProviderRequest)

				provider, err := s.Storage.GetSamlIdentityProviderPersister(nil).GetByDomain(data.Domain)
				s.Require().NoError(err)

				s.Assert().NotNil(provider.ID)
				s.Assert().Equal(data.Enabled, provider.Enabled)
				s.Assert().Equal(data.Name, provider.Name)
				s.Assert().Equal(data.Domain, provider.Domain)
				s.Assert().Equal(data.MetadataUrl, provider.MetadataUrl)
				s.Assert().Equal(data.SkipEmailVerification, provider.SkipEmailVerification)
				s.Assert().Equal(data.AttributeMap.Name, provider.AttributeMap.Name)
				s.Assert().Equal(data.AttributeMap.FamilyName, provider.AttributeMap.FamilyName)
				s.Assert().Equal(data.AttributeMap.GivenName, provider.AttributeMap.GivenName)
				s.Assert().Equal(data.AttributeMap.MiddleName, provider.AttributeMap.MiddleName)
				s.Assert().Equal(data.AttributeMap.NickName, provider.AttributeMap.NickName)
				s.Assert().Equal(data.AttributeMap.PreferredUsername, provider.AttributeMap.PreferredUsername)
				s.Assert().Equal(data.AttributeMap.Profile, provider.AttributeMap.Profile)
				s.Assert().Equal(data.AttributeMap.Picture, provider.AttributeMap.Picture)
				s.Assert().Equal(data.AttributeMap.Website, provider.AttributeMap.Website)
				s.Assert().Equal(data.AttributeMap.Gender, provider.AttributeMap.Gender)
				s.Assert().Equal(data.AttributeMap.Birthdate, provider.AttributeMap.Birthdate)
				s.Assert().Equal(data.AttributeMap.ZoneInfo, provider.AttributeMap.ZoneInfo)
				s.Assert().Equal(data.AttributeMap.Locale, provider.AttributeMap.Locale)
				s.Assert().Equal(data.AttributeMap.UpdatedAt, provider.AttributeMap.SamlUpdatedAt)
				s.Assert().Equal(data.AttributeMap.Email, provider.AttributeMap.Email)
				s.Assert().Equal(data.AttributeMap.EmailVerified, provider.AttributeMap.EmailVerified)
				s.Assert().Equal(data.AttributeMap.Phone, provider.AttributeMap.Phone)
				s.Assert().Equal(data.AttributeMap.PhoneVerified, provider.AttributeMap.PhoneVerified)
			}
		})
	}
}

func (s *samlAdminHandlerSuite) TestSamlAdminHandler_Get() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../../../test/fixtures/saml")
	s.Require().NoError(err)

	s.setupServer()

	tests := []struct {
		Name                 string
		ProviderId           string
		ContainsError        bool
		ExpectedCode         int
		ExpectedErrorMessage string
	}{
		{
			Name:          "Success",
			ProviderId:    "d531b0ae-4c33-48bb-ad31-e800a71a5056",
			ExpectedCode:  http.StatusOK,
			ContainsError: false,
		},
		{
			Name:                 "Validation Error",
			ProviderId:           "lorem",
			ContainsError:        true,
			ExpectedCode:         http.StatusBadRequest,
			ExpectedErrorMessage: validateRequestError,
		},
		{
			Name:                 "Not Found Error",
			ProviderId:           "00000000-4c33-48bb-ad31-e800a71a5056",
			ContainsError:        true,
			ExpectedCode:         http.StatusNotFound,
			ExpectedErrorMessage: providerNotFoundError,
		},
	}

	for _, samlTest := range tests {
		s.T().Run(samlTest.Name, func(t *testing.T) {
			// when
			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/saml/%s", samlTest.ProviderId), nil)
			rec := httptest.NewRecorder()

			s.server.ServeHTTP(rec, req)

			// then
			s.Assert().Equal(samlTest.ExpectedCode, rec.Code)

			if samlTest.ContainsError {
				var hErr echo.HTTPError
				err = json.Unmarshal(rec.Body.Bytes(), &hErr)
				s.Require().NoError(err)

				s.Assert().Equal(samlTest.ExpectedErrorMessage, hErr.Message)
			} else {
				var provider models.SamlIdentityProvider
				err := json.Unmarshal(rec.Body.Bytes(), &provider)
				s.Require().NoError(err)

				s.Assert().Equal(samlTest.ProviderId, provider.ID.String())
				s.Assert().True(provider.Enabled)
				s.Assert().False(provider.SkipEmailVerification)
				s.Assert().Equal("hanko", provider.Name)
				s.Assert().Equal("hanko.io", provider.Domain)
				s.Assert().Equal("https://localhost/metadata", provider.MetadataUrl)
			}
		})
	}
}

func (s *samlAdminHandlerSuite) TestSamlAdminHandler_Update() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../../../test/fixtures/saml")
	s.Require().NoError(err)

	s.setupServer()

	tests := []struct {
		Name                 string
		ProviderId           string
		ProviderData         interface{}
		ExpectedCode         int
		ExpectedErrorMessage string
		ContainsError        bool
	}{
		{
			Name:       "Success",
			ProviderId: "d531b0ae-4c33-48bb-ad31-e800a71a5056",
			ProviderData: dto.SamlCreateProviderRequest{
				Enabled:               true,
				Name:                  "Ipsum",
				Domain:                "test.de",
				MetadataUrl:           "http://fqdn.loc/metadata",
				SkipEmailVerification: false,
				AttributeMap: &dto.SamlCreateProviderAttributeMapRequest{
					Name: "home",
				},
			},
			ExpectedCode:  http.StatusOK,
			ContainsError: false,
		},
		{
			Name:       "No domain change success",
			ProviderId: "d531b0ae-4c33-48bb-ad31-e800a71a5056",
			ProviderData: dto.SamlCreateProviderRequest{
				Enabled:               true,
				Name:                  "Ipsum",
				Domain:                "hanko.io",
				MetadataUrl:           "http://fqdn.loc/metadata",
				SkipEmailVerification: false,
				AttributeMap: &dto.SamlCreateProviderAttributeMapRequest{
					Name: "home",
				},
			},
			ExpectedCode:  http.StatusOK,
			ContainsError: false,
		},
		{
			Name:                 "Bind request error",
			ProviderId:           "d531b0ae-4c33-48bb-ad31-e800a71a5056",
			ProviderData:         "gibberish",
			ExpectedCode:         http.StatusBadRequest,
			ExpectedErrorMessage: bindRequestError,
			ContainsError:        true,
		},
		{
			Name:       "validate request error",
			ProviderId: "d531b0ae-4c33-48bb-ad31-e800a71a5056",
			ProviderData: dto.SamlCreateProviderRequest{
				Name: "test",
			},
			ExpectedCode:         http.StatusBadRequest,
			ExpectedErrorMessage: validateRequestError,
			ContainsError:        true,
		},
		{
			Name:       "not found error",
			ProviderId: "00000000-4c33-48bb-ad31-e800a71a5056",
			ProviderData: dto.SamlCreateProviderRequest{
				Enabled:               true,
				Name:                  "Ipsum",
				Domain:                "test.de",
				MetadataUrl:           "http://fqdn.loc/metadata",
				SkipEmailVerification: false,
				AttributeMap: &dto.SamlCreateProviderAttributeMapRequest{
					Name: "home",
				},
			},
			ExpectedCode:         http.StatusNotFound,
			ExpectedErrorMessage: providerNotFoundError,
			ContainsError:        true,
		},
		{
			Name:       "conflict error",
			ProviderId: "d531b0ae-4c33-48bb-ad31-e800a71a5056",
			ProviderData: dto.SamlCreateProviderRequest{
				Enabled:               true,
				Name:                  "Ipsum",
				Domain:                "localhost",
				MetadataUrl:           "http://fqdn.loc/metadata",
				SkipEmailVerification: false,
				AttributeMap: &dto.SamlCreateProviderAttributeMapRequest{
					Name: "home",
				},
			},
			ExpectedCode:         http.StatusConflict,
			ExpectedErrorMessage: "a provider with the domain 'localhost' already exists",
			ContainsError:        true,
		},
	}

	for _, samlTest := range tests {
		s.setupServer()

		dataJson, err := json.Marshal(samlTest.ProviderData)
		s.Require().NoError(err)

		// when
		req := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/saml/%s", samlTest.ProviderId), bytes.NewReader(dataJson))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()

		s.server.ServeHTTP(rec, req)

		s.Assert().Equal(samlTest.ExpectedCode, rec.Code)

		if samlTest.ContainsError {
			var hErr echo.HTTPError
			err = json.Unmarshal(rec.Body.Bytes(), &hErr)
			s.Require().NoError(err)

			s.Assert().Equal(samlTest.ExpectedErrorMessage, hErr.Message)
		} else {
			providerUUid, err := uuid.FromString(samlTest.ProviderId)
			s.Require().NoError(err)

			provider, err := s.Storage.GetSamlIdentityProviderPersister(nil).Get(providerUUid)
			s.Require().NoError(err)

			testDto := samlTest.ProviderData.(dto.SamlCreateProviderRequest)

			s.Assert().Equal(testDto.Name, provider.Name)
			s.Assert().Equal(testDto.Domain, provider.Domain)
			s.Assert().Equal(testDto.MetadataUrl, provider.MetadataUrl)
			s.Assert().Equal(testDto.Enabled, provider.Enabled)
			s.Assert().Equal(testDto.SkipEmailVerification, provider.SkipEmailVerification)
			s.Assert().Equal(testDto.AttributeMap.Name, provider.AttributeMap.Name)
			s.Assert().Equal(testDto.AttributeMap.Profile, provider.AttributeMap.Profile)
		}
	}
}

func (s *samlAdminHandlerSuite) TestSamlAdminHandler_Delete() {
	if testing.Short() {
		s.T().Skip("skipping test in short mode")
	}

	err := s.LoadFixtures("../../../test/fixtures/saml")
	s.Require().NoError(err)

	s.setupServer()

	tests := []struct {
		Name                 string
		ProviderId           string
		ContainsError        bool
		ExpectedCode         int
		ExpectedErrorMessage string
	}{
		{
			Name:          "Success",
			ProviderId:    "d531b0ae-4c33-48bb-ad31-e800a71a5056",
			ExpectedCode:  http.StatusNoContent,
			ContainsError: false,
		},
		{
			Name:                 "Validation Error",
			ProviderId:           "lorem",
			ContainsError:        true,
			ExpectedCode:         http.StatusBadRequest,
			ExpectedErrorMessage: validateRequestError,
		},
		{
			Name:                 "Not Found Error",
			ProviderId:           "00000000-4c33-48bb-ad31-e800a71a5056",
			ContainsError:        true,
			ExpectedCode:         http.StatusNotFound,
			ExpectedErrorMessage: providerNotFoundError,
		},
	}

	for _, samlTest := range tests {
		s.T().Run(samlTest.Name, func(t *testing.T) {
			// when
			req := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/saml/%s", samlTest.ProviderId), nil)
			rec := httptest.NewRecorder()

			s.server.ServeHTTP(rec, req)

			// then
			s.Assert().Equal(samlTest.ExpectedCode, rec.Code)

			if samlTest.ContainsError {
				var hErr echo.HTTPError
				err = json.Unmarshal(rec.Body.Bytes(), &hErr)
				s.Require().NoError(err)

				s.Assert().Equal(samlTest.ExpectedErrorMessage, hErr.Message)
			} else {
				providers, err := s.Storage.GetSamlIdentityProviderPersister(nil).List()
				s.Require().NoError(err)

				s.Assert().Len(providers, 1)
			}
		})
	}
}
