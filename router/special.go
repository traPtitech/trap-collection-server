package router

import (
	"net/http"

	"github.com/labstack/echo"

	"github.com/traPtitech/trap-collection-server/repository"
)

//PostSpecialHandler 特例を作る
func PostSpecialHandler(c echo.Context) error {
	id := c.Param("id")
	type GameSpecial struct {
		GameName string `json:"game_name,omitempty"`
		InOut    string `json:"in_out,omitempty"`
	}
	game := GameSpecial{}
	err := c.Bind(&game)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}

	b, err := repository.IsThereSpecial(id, game.GameName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if is there the special case")
	}
	if b {
		return c.String(http.StatusInternalServerError, "there is the special case")
	}

	err = repository.InsertSpecial(id, game.GameName, game.InOut)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in inserting special case")
	}

	return c.NoContent(http.StatusOK)
}

//DeleteSpecialHandler 特例を削除
func DeleteSpecialHandler(c echo.Context) error {
	id := c.Param("id")
	type GameSpecial struct {
		GameName string `json:"game_name,omitempty"`
	}
	game := GameSpecial{}
	err := c.Bind(&game)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in binding")
	}

	b, err := repository.IsThereSpecial(id, game.GameName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in checking if is there the special case")
	}
	if !b {
		return c.String(http.StatusInternalServerError, "there is not the special case")
	}

	err = repository.DeleteSpecial(id, game.GameName)
	if err != nil {
		return c.String(http.StatusInternalServerError, "something wrong in deleting special case")
	}

	return c.NoContent(http.StatusOK)
}
