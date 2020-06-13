package router

import (
	"fmt"

	"github.com/traPtitech/trap-collection-server/model"
	"github.com/traPtitech/trap-collection-server/openapi"
)

// Response responseの構造体
type Response struct {
	LauncherAuthBase
	openapi.ResponseApi
}

// PostResponse POST /responses の処理部分
func (r *Response) PostResponse(response openapi.NewResponse, sess sessionMap) (openapi.NewResponse, sessionMap, error) {
	productKey, err := r.getProductKey(sess)
	if err != nil {
		return openapi.NewResponse{}, sessionMap{}, fmt.Errorf("Failed In Getting ProductKey: %w", err)
	}
	response, err = model.InsertResponses(productKey, response)
	if err != nil {
		return openapi.NewResponse{}, sessionMap{}, fmt.Errorf("Failed In Inserting Response: %w", err)
	}
	return response, sessionMap{}, nil
}
