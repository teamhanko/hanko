package handler

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
	"time"
)

type AuditLogHandler struct {
	persister persistence.Persister
}

func NewAuditLogHandler(persister persistence.Persister) *AuditLogHandler {
	return &AuditLogHandler{
		persister: persister,
	}
}

type AuditLogListRequest struct {
	Page      int        `query:"page"`
	PerPage   int        `query:"per_page"`
	StartTime *time.Time `query:"start_time"`
	EndTime   *time.Time `query:"end_time"`
}

func (h AuditLogHandler) List(c echo.Context) error {
	var request AuditLogListRequest
	err := (&echo.DefaultBinder{}).BindQueryParams(c, &request)
	if err != nil {
		return dto.ToHttpError(err)
	}

	auditLogs, err := h.persister.GetAuditLogPersister().List(request.Page, request.PerPage, request.StartTime, request.EndTime)
	if err != nil {
		return fmt.Errorf("failed to get list of audit logs: %w", err)
	}

	return c.JSON(http.StatusOK, auditLogs)
}
