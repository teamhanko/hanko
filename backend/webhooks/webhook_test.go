package webhooks

import (
	"github.com/labstack/gommon/log"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestBaseWebhook_HasEvent(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   nil,
		Callback: "http://ipsum.lorem",
		Events:   events.Events{events.UserUpdate},
	}

	require.True(t, baseHook.HasEvent(events.EmailCreate))
}

func TestBaseWebhook_HasSubEvent(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   nil,
		Callback: "http://ipsum.lorem",
		Events:   events.Events{events.UserCreate},
	}

	require.True(t, baseHook.HasEvent(events.UserCreate))
}

func TestBaseWebhook_DoesNotHaveEvent(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   nil,
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
		Logger:   nil,
		Callback: server.URL,
		Events:   events.Events{events.UserCreate},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)
	require.NoError(t, err)
}

func TestBaseWebhook_TriggerWithWrongUrl(t *testing.T) {
	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: "http://broken!",
		Events:   events.Events{events.UserCreate},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)
	require.Error(t, err)
	require.Contains(t, err.Error(), "dial tcp: lookup broken!: no such host")
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
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.Error(t, err)
	require.ErrorContains(t, err, "request failed due to status code")
}

func TestBaseWebhook_TriggerWithBadServer(t *testing.T) {
	server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(600 * time.Millisecond)
	}))
	server.Config.WriteTimeout = 500 * time.Millisecond
	server.Start()
	defer server.Close()

	baseHook := BaseWebhook{
		Logger:   log.New("test"),
		Callback: server.URL,
		Events:   events.Events{events.UserCreate},
	}

	data := JobData{
		Token: "test-token",
		Event: "user",
	}

	err := baseHook.Trigger(data)

	require.Error(t, err)
	require.ErrorContains(t, err, "EOF")
}
