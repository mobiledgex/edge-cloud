package orm

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
)

type M map[string]interface{}

func Msg(msg string) *ormapi.Result {
	return &ormapi.Result{Message: msg}
}

func MsgErr(err error) *ormapi.Result {
	return &ormapi.Result{Message: err.Error()}
}

func MsgID(msg string, id int64) *ormapi.ResultID {
	return &ormapi.ResultID{
		Message: msg,
		ID:      id,
	}
}

func MsgName(msg, name string) *ormapi.ResultName {
	return &ormapi.ResultName{
		Message: msg,
		Name:    name,
	}
}

func ctrlErr(c echo.Context, err error) error {
	msg := "controller connect error, " + err.Error()
	return c.JSON(http.StatusBadRequest, Msg(msg))
}

func dbErr(err error) error {
	return fmt.Errorf("database error, %s", err.Error())
}

func setReply(c echo.Context, err error, successReply interface{}) error {
	if err == echo.ErrForbidden {
		return err
	}
	if err != nil {
		return c.JSON(http.StatusBadRequest, MsgErr(err))
	}
	return c.JSON(http.StatusOK, successReply)
}

func streamReply(c echo.Context, desc string, err error) {
	res := "ok"
	if err == echo.ErrForbidden {
		res = "forbidden"
	} else if err != nil {
		res = err.Error()
	}
	streamReplyMsg(c, desc, res)
}

func streamReplyMsg(c echo.Context, desc, res string) {
	msg := Msg(fmt.Sprintf("%s: %s", desc, res))
	json.NewEncoder(c.Response()).Encode(msg)
	c.Response().Flush()
}
