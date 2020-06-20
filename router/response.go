package router

import (
	"fmt"

	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
	"github.com/traPtitech/trap-collection-server/router/base"
)

// Response responseの構造体
type Response struct {
	db model.DBMeta
	launcherAuth base.LauncherAuth
	openapi.ResponseApi
}

func newResponse(db model.DBMeta, launcherAuth base.LauncherAuth) openapi.ResponseApi {
	response := new(Response)

	response.db = db
	response.launcherAuth = launcherAuth

	return response
}

// PostResponse POST /responses の処理部分
func (r *Response) PostResponse(c echo.Context, response *openapi.NewResponse) (*openapi.NewResponse, error) {
	productKey, err := r.launcherAuth.GetProductKey(c)
	if err != nil {
		return &openapi.NewResponse{}, fmt.Errorf("Failed In Getting ProductKey: %w", err)
	}
	response, err = r.db.InsertResponses(productKey, response)
	if err != nil {
		return &openapi.NewResponse{}, fmt.Errorf("Failed In Inserting Response: %w", err)
	}
	return response, nil
}
