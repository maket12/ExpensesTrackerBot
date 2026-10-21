package http

import (
	_ "backend/docs"
	"context"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

type Router struct {
	Expenses *ExpensesHandler
}

// @title Family Expenses Tracker API
// @version 1.0
// @description REST API for Family Expenses Tracker Web Panel
// @BasePath /
func NewRouter(expensesHandler *ExpensesHandler) *Router {
	return &Router{Expenses: expensesHandler}
}

func (r *Router) InitRoutes() *echo.Echo {
	router := echo.New()

	router.Use(middleware.Recover())
	router.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete, http.MethodOptions},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, "X-API-Key"},
	}))
	router.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		Skipper: func(c echo.Context) bool {
			p := c.Path()
			return p == "/health" || p == "/api/health" || c.Request().URL.Path == "/health" || c.Request().URL.Path == "/api/health"
		},
		LogLatency:       true,
		LogProtocol:      false,
		LogRemoteIP:      true,
		LogHost:          true,
		LogMethod:        true,
		LogURI:           true,
		LogRequestID:     true,
		LogUserAgent:     true,
		LogStatus:        true,
		LogError:         true,
		LogContentLength: true,
		LogResponseSize:  true,
		HandleError:      true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				slog.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.Duration("latency", v.Latency),
					slog.String("host", v.Host),
					slog.String("bytes_in", v.ContentLength),
					slog.Int64("bytes_out", v.ResponseSize),
					slog.String("user_agent", v.UserAgent),
					slog.String("remote_ip", v.RemoteIP),
					slog.String("request_id", v.RequestID),
				)
			} else {
				slog.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("method", v.Method),
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.Duration("latency", v.Latency),
					slog.String("host", v.Host),
					slog.String("bytes_in", v.ContentLength),
					slog.Int64("bytes_out", v.ResponseSize),
					slog.String("user_agent", v.UserAgent),
					slog.String("remote_ip", v.RemoteIP),
					slog.String("request_id", v.RequestID),
					slog.String("error", v.Error.Error()),
				)
			}
			return nil
		},
	}))

	router.GET("/swagger/*", echoSwagger.WrapHandler)

	api := router.Group("/api")
	{
		api.GET("/health", r.Expenses.Health)
		api.POST("/expenses", r.Expenses.Create)
		api.GET("/expenses", r.Expenses.List)
		api.GET("/expenses/:id", r.Expenses.Get)
		api.PUT("/expenses/:id", r.Expenses.Update)
		api.DELETE("/expenses/:id", r.Expenses.Delete)
		api.GET("/users", r.Expenses.GetDistinctUsers)
	}

	return router
}
