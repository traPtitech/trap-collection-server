/*
 * traPCollection API
 *
 * traPCollectionのAPI
 *
 * API version: 1.0.0
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"io"
	"mime/multipart"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo-contrib/session"
	echo "github.com/labstack/echo/v4"
)

type ioReadCloser = io.ReadCloser
type multipartFile = multipart.File
type sessionsSession = sessions.Session

var getSession = session.Get

// Api Apiのインターフェイス
type Api struct {
	Middleware
	GameApi
	LauncherAuthApi
	Oauth2Api
	SeatApi
	SeatVersionApi
	UserApi
	VersionApi
}

type Middleware interface {
	TrapMemberAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc
	GameMaintainerAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc
	GameOwnerAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc
	AdminAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc
	LauncherAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc
	BothAuthMiddleware(next echo.HandlerFunc) echo.HandlerFunc
}
