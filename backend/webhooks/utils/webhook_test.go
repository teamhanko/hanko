package utils

import (
	"github.com/gobuffalo/pop/v6"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/require"
	"github.com/teamhanko/hanko/backend/webhooks/events"
	"net/http"
	"net/http/httptest"
	"testing"
)

type testManager struct {
	TestFunc func()
}

func (tm *testManager) Trigger(tx *pop.Connection, evt events.Event, data interface{}) {
	tm.TestFunc()
}

func (tm *testManager) GenerateJWT(data interface{}, event events.Event) (string, error) {
	return "", nil
}

func TestWebhook_TriggerWithoutManager(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	rec := httptest.NewRecorder()

	ctx := e.NewContext(req, rec)

	err := TriggerWebhooks(ctx, nil, "user", "lorem")
	require.Error(t, err)

	err = e.Close()
	require.NoError(t, err)
}

func TestWebhook_Trigger(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/path", nil)
	rec := httptest.NewRecorder()

	tm := &testManager{TestFunc: func() {
		require.True(t, true)
	}}

	ctx := e.NewContext(req, rec)
	ctx.Set("webhook_manager", tm)

	err := TriggerWebhooks(ctx, nil, "user", "lorem")
	require.NoError(t, err)

	err = e.Close()
	require.NoError(t, err)
}
