package v2

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/oapi-codegen/echo-middleware"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
)

func TestLatestVersionRequiresVisibilityCheck(t *testing.T) {
	spec, err := openapi.GetSpec()
	if err != nil {
		t.Fatal(err)
	}
	called := false
	e := echo.New()
	e.Use(echomiddleware.OapiRequestValidatorWithOptions(spec, &echomiddleware.Options{
		Options: openapi3filter.Options{AuthenticationFunc: func(_ context.Context, input *openapi3filter.AuthenticationInput) error {
			if input.SecuritySchemeName != "GameInfoVisibilityAuth" {
				t.Errorf("unexpected security scheme: %s", input.SecuritySchemeName)
			}
			called = true
			return echo.NewHTTPError(http.StatusUnauthorized)
		}},
	}))
	e.GET("/api/v2/games/:gameID/versions/latest", func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v2/games/00000000-0000-4000-8000-000000000001/versions/latest", nil)
	e.ServeHTTP(recorder, request)
	if !called || recorder.Code == http.StatusOK {
		t.Fatalf("visibility check called=%v, status=%d; expected rejection", called, recorder.Code)
	}
}
