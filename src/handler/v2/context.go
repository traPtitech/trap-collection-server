package v2

import (
	"github.com/labstack/echo/v4"
	"github.com/traPtitech/trap-collection-server/src/domain"
)

const (
	productKeyContextKey = "productKey"
	editionContextKey    = "edition"
)

type Context struct{}

func NewContext() *Context {
	return &Context{}
}

func (context *Context) SetProductKey(c echo.Context, productKey *domain.LauncherUser) {
	c.Set(productKeyContextKey, productKey)
}

func (context *Context) GetProductKey(c echo.Context) (*domain.LauncherUser, error) {
	productKey, ok := c.Get(productKeyContextKey).(*domain.LauncherUser)
	if !ok || productKey == nil {
		return nil, ErrNoValue
	}

	return productKey, nil
}

func (context *Context) SetEdition(c echo.Context, edition *domain.LauncherVersion) {
	c.Set(editionContextKey, edition)
}

func (context *Context) GetEdition(c echo.Context) (*domain.LauncherVersion, error) {
	edition, ok := c.Get(editionContextKey).(*domain.LauncherVersion)
	if !ok || edition == nil {
		return nil, ErrNoValue
	}

	return edition, nil
}
