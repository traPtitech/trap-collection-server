package v2

import (
	"errors"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain/values"
	"github.com/traPtitech/trap-collection-server/src/handler/v2/openapi"
	"github.com/traPtitech/trap-collection-server/src/service"
)

type Admin struct {
	adminService service.AdminAuthV2
	session      *Session
}

func NewAdmin(adminService service.AdminAuthV2, session *Session) *Admin {
	return &Admin{
		adminService: adminService,
		session:      session,
	}
}

// adminUnimplemented
// メソッドとして実装予定だが、未実装のもの
// TODO: 実装
type adminUnimplemented interface {
	// traPの管理者一覧取得
	// (GET /admins)
	GetAdmins(ctx echo.Context) error
	// traP Collection全体の管理者追加
	// (POST /admins)
	PostAdmin(ctx echo.Context) error
	// traP Collection全体の管理者削除
	// (DELETE /admins/{userID})
	DeleteAdmin(ctx echo.Context, userID openapi.UserIDInPath) error
}

// traPの管理者一覧取得
// (GET /admins)
func (a *Admin) GetAdmins(ctx echo.Context) error {
	session, err := a.session.get(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no session")
	}

	authSession, err := a.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no auth session")
	}

	adminInfos, err := a.adminService.GetAdmins(ctx.Request().Context(), authSession)
	if err != nil {
		log.Printf("error: failed to get admins info: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to get admins info")
	}

	responseAdmins := make([]openapi.User, 0, len(adminInfos))
	for _, adminInfo := range adminInfos {
		responseAdmins = append(responseAdmins, openapi.User{
			Id:   uuid.UUID(adminInfo.GetID()),
			Name: string(adminInfo.GetName()),
		})
	}
	return ctx.JSON(http.StatusOK, responseAdmins)
}

// traP Collection全体の管理者追加
// (POST /admins)
func (a *Admin) PostAdmin(ctx echo.Context) error {
	session, err := a.session.get(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no session")
	}

	authSession, err := a.session.getAuthSession(session)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "no auth session")
	}

	req := &openapi.PostAdminJSONRequestBody{}
	err = ctx.Bind(req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bad request body")
	}
	newAdminID := values.NewTrapMemberID(req.Id)

	adminInfos, err := a.adminService.AddAdmin(ctx.Request().Context(), authSession, values.TraPMemberID(newAdminID))
	if errors.Is(err, service.ErrInvalidUserID) {
		return echo.NewHTTPError(http.StatusBadRequest, "not active user")
	}
	if errors.Is(err, service.ErrNoAdminsUpdated) {
		return echo.NewHTTPError(http.StatusBadRequest, "already admin")
	}
	if err != nil {
		log.Printf("error: failed to add admin: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to add admin")
	}

	res := make([]openapi.User, 0, len(adminInfos))
	for _, adminInfo := range adminInfos {
		res = append(res,
			openapi.User{Id: uuid.UUID(adminInfo.GetID()), Name: string(adminInfo.GetName())},
		)
	}

	return ctx.JSON(http.StatusOK, res)
}
