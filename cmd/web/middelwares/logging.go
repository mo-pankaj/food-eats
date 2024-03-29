package middleware

import (
	"context"
	"github.com/hashicorp/go-uuid"
	"github.com/labstack/echo/v4"
)

// RequestIDMiddleware generates a unique request ID for each request
func RequestIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Generate a unique request ID
		requestID, _ := uuid.GenerateUUID()

		// Set the request ID in the request context
		ctx := context.WithValue(c.Request().Context(), "correlation_id", requestID)

		request := c.Request().Clone(ctx)
		c.SetRequest(request)
		return next(c)
	}
}

// AddMetaData adding meta-information about the route. Method, Path, UserId Agent
func AddMetaData(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		path := c.Request().RequestURI
		method := c.Request().Method
		userAgent := c.Request().UserAgent()
		ctx := context.WithValue(c.Request().Context(), "request_path", path)
		ctx = context.WithValue(ctx, "request_method", method)
		ctx = context.WithValue(ctx, "request_user_agent", userAgent)
		request := c.Request().Clone(ctx)
		c.SetRequest(request)
		return next(c)
	}
}
