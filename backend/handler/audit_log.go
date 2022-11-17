package handler

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/teamhanko/hanko/backend/dto"
	"github.com/teamhanko/hanko/backend/pagination"
	"github.com/teamhanko/hanko/backend/persistence"
	"net/http"
	"net/url"
	"strconv"
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

	if request.Page == 0 {
		request.Page = 1
	}

	if request.PerPage == 0 {
		request.PerPage = 20
	}

	auditLogs, err := h.persister.GetAuditLogPersister().List(request.Page, request.PerPage, request.StartTime, request.EndTime)
	if err != nil {
		return fmt.Errorf("failed to get list of audit logs: %w", err)
	}

	logCount, err := h.persister.GetAuditLogPersister().Count(request.StartTime, request.EndTime)
	if err != nil {
		return fmt.Errorf("failed to get total count of audit logs: %w", err)
	}

	u, _ := url.Parse(fmt.Sprintf("%s://%s%s", c.Scheme(), c.Request().Host, c.Request().RequestURI))

	c.Response().Header().Set("Link", pagination.CreateHeader(u, logCount, request.Page, request.PerPage))
	c.Response().Header().Set("X-Total-Count", strconv.FormatInt(int64(logCount), 10))

	return c.JSON(http.StatusOK, auditLogs)
}
