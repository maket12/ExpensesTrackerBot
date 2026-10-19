package http

import (
	_ "backend/docs"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Family Expenses Tracker API
// @version 1.0
// @description REST API for Family Expenses Tracker Web Panel
// @BasePath /
func NewRouter(h *Handler) *echo.Echo {
	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"*"},
	}))

	e.GET("/swagger/*", echoSwagger.WrapHandler)

	api := e.Group("/api")
	api.GET("/health", h.Health)
	api.POST("/expenses", h.CreateExpense)
	api.GET("/expenses", h.ListExpenses)
	api.GET("/expenses/:id", h.GetExpense)
	api.PUT("/expenses/:id", h.UpdateExpense)
	api.DELETE("/expenses/:id", h.DeleteExpense)
	api.GET("/users", h.GetDistinctUsers)

	return e
}
