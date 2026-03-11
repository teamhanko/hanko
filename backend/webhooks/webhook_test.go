package webhooks

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/v2/config"
	"github.com/teamhanko/hanko/backend/v2/webhooks/events"
)

func TestBaseWebhook_HasEvent(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: "http://ipsum.lorem",
		Events:   events.Events{events.UserUpdate},
	}

	require.True(t, baseHook.HasEvent(events.UserEmailCreate))
}

func TestWebhooks_HasEvent_WithMultipleEvents(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: "http://ipsum.lorem",
		Events:   events.Events{events.UserCreate, events.UserUpdate},
	}

	require.True(t, baseHook.HasEvent(events.UserUpdate))
}

func TestWebhooks_HasSubEvent_WithMultipleEvents(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: "http://ipsum.lorem",
		Events:   events.Events{events.UserCreate, events.UserUpdate},
	}

	require.True(t, baseHook.HasEvent(events.UserEmailCreate))
}

func TestBaseWebhook_HasSubEvent(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: "http://ipsum.lorem",
		Events:   events.Events{events.UserCreate},
	}

	require.True(t, baseHook.HasEvent(events.UserCreate))
}

func TestBaseWebhook_DoesNotHaveEvent(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: "http://ipsum.lorem",
		Events:   events.Events{events.UserCreate},
	}

	require.False(t, baseHook.HasEvent("user"))
}

func TestBaseWebhook_Trigger(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: server.URL,
		Events:   events.Events{events.UserCreate},
		Security: config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeInsecure,
			AllowedSchemes: []string{"http", "https"},
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)
	require.NoError(t, err)
}

func TestBaseWebhook_TriggerWithWrongURL(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: "http://broken!",
		Events:   events.Events{events.UserCreate},
		Security: config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeInsecure,
			AllowedSchemes: []string{"http", "https"},
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)
	require.Error(t, err)
}

func TestBaseWebhook_TriggerWithDisallowedSchemeFailsEvenInInsecureMode(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: server.URL,
		Events:   events.Events{events.UserCreate},
		Security: config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeInsecure,
			AllowedSchemes: []string{"http"},
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.Error(t, err)
	require.ErrorContains(t, err, "scheme")
}

func TestBaseWebhook_TriggerWithBadStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: server.URL,
		Events:   events.Events{events.UserCreate},
		Security: config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeInsecure,
			AllowedSchemes: []string{"http", "https"},
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.Error(t, err)
	require.ErrorContains(t, err, "status code")
}

func TestBaseWebhook_TriggerWithRedirectDisallowedFails(t *testing.T) {
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer redirectTarget.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectTarget.URL, http.StatusFound)
	}))
	defer server.Close()

	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: server.URL,
		Events:   events.Events{events.UserCreate},
		Security: config.WebhookSecurity{
			Mode:            config.WebhookSecurityModeInsecure,
			AllowedSchemes:  []string{"http", "https"},
			FollowRedirects: false,
			MaxRedirects:    0,
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.Error(t, err)
	require.ErrorContains(t, err, "status code")
}

func TestBaseWebhook_TriggerWithRedirectAllowedSucceeds(t *testing.T) {
	redirectTarget := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer redirectTarget.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectTarget.URL, http.StatusFound)
	}))
	defer server.Close()

	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: server.URL,
		Events:   events.Events{events.UserCreate},
		Security: config.WebhookSecurity{
			Mode:            config.WebhookSecurityModeInsecure,
			AllowedSchemes:  []string{"http", "https"},
			FollowRedirects: true,
			MaxRedirects:    1,
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.NoError(t, err)
}

func TestBaseWebhook_TriggerWithRedirectRejectedByPolicyFails(t *testing.T) {
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "http://forbidden.invalid/forbidden", http.StatusFound)
	}))
	defer redirectServer.Close()

	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: redirectServer.URL,
		Events:   events.Events{events.UserCreate},
		Security: config.WebhookSecurity{
			Mode:            config.WebhookSecurityModeInsecure,
			AllowedSchemes:  []string{"http", "https"},
			FollowRedirects: true,
			MaxRedirects:    1,
			BlockedHosts:    []string{"forbidden.invalid"},
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.Error(t, err)
	require.ErrorContains(t, err, "redirect target rejected by outbound policy")
}

func TestBaseWebhook_TriggerWithTooManyRedirectsFails(t *testing.T) {
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, server.URL, http.StatusFound)
	}))
	defer server.Close()

	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: server.URL,
		Events:   events.Events{events.UserCreate},
		Security: config.WebhookSecurity{
			Mode:            config.WebhookSecurityModeInsecure,
			AllowedSchemes:  []string{"http", "https"},
			FollowRedirects: true,
			MaxRedirects:    1,
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.Error(t, err)
	require.ErrorContains(t, err, "too many redirects")
}

func TestBaseWebhook_TriggerWithBadServer(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(600 * time.Millisecond)
	}))
	server.Config.WriteTimeout = 500 * time.Millisecond
	server.Start()
	defer server.Close()

	baseHook := BaseWebhook{
		Logger:         log.New("test"),
		Callback:       server.URL,
		Events:         events.Events{events.UserCreate},
		RequestTimeout: 2 * time.Second,
		Security: config.WebhookSecurity{
			Mode:           config.WebhookSecurityModeInsecure,
			AllowedSchemes: []string{"http", "https"},
		},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.Error(t, err)
}
