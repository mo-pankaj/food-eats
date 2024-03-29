package custom_errors

import (
	"context"
	"errors"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"
)

var (
	ServerError = errors.New("server error")
	ClientError = errors.New("client error")
)

// ParseError error parser
func ParseError(ctx context.Context, err error, req interface{}, c echo.Context) error {
	if errors.Is(err, ClientError) {
		return parseClientError(err, c)
	}
	buffer := debug.Stack()
	trace := strings.Split(string(buffer), "\n")

	var args []interface{}
	if req != nil {
		args = append(args, "req", req)
	}
	args = append(args, "trace", trace[2:])
	slog.ErrorContext(ctx, "", "error", err.Error(), args)
	return parseServerError(err, c)
}
func parseServerError(err error, c echo.Context) error {
	return c.JSON(http.StatusInternalServerError, err.Error())
}

func parseClientError(err error, c echo.Context) error {
	return c.JSON(http.StatusBadRequest, err.Error())
}
