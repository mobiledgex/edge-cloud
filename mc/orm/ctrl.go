package orm

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
)

func connectController(region string) (*grpc.ClientConn, error) {
	addr, err := getControllerAddrForRegion(region)
	if err != nil {
		return nil, err
	}
	return connectControllerAddr(addr)
}

func connectControllerAddr(addr string) (*grpc.ClientConn, error) {
	dialOption, err := tls.GetTLSClientDialOption(addr, serverConfig.TlsCertFile)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(addr, dialOption)
}

func getControllerAddrForRegion(region string) (string, error) {
	ctrl := ormapi.Controller{
		Region: region,
	}
	res := db.Where(&ctrl).First(&ctrl)
	if res.Error != nil {
		if res.RecordNotFound() {
			return "", fmt.Errorf("region \"%s\" not found", region)
		}
		return "", res.Error
	}
	return ctrl.Address, nil
}

func CreateController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctrl := ormapi.Controller{}
	if err := c.Bind(&ctrl); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid Post data"))
	}
	err = CreateControllerObj(claims, &ctrl)
	return setReply(c, err, Msg("Controller registered"))
}

func CreateControllerObj(claims *UserClaims, ctrl *ormapi.Controller) error {
	if ctrl.Region == "" {
		return fmt.Errorf("Controller Region not specified")
	}
	if ctrl.Address == "" {
		return fmt.Errorf("Controller Address not specified")
	}
	if !enforcer.Enforce(claims.Username, "", ResourceControllers, ActionManage) {
		return echo.ErrForbidden
	}
	err := db.Create(ctrl).Error
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func DeleteController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctrl := ormapi.Controller{}
	if err := c.Bind(&ctrl); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid Post data"))
	}
	err = DeleteControllerObj(claims, &ctrl)
	return setReply(c, err, Msg("Controller deregistered"))
}

func DeleteControllerObj(claims *UserClaims, ctrl *ormapi.Controller) error {
	if !enforcer.Enforce(claims.Username, "", ResourceControllers, ActionManage) {
		return echo.ErrForbidden
	}
	err := db.Delete(ctrl).Error
	if err != nil {
		return dbErr(err)
	}
	return nil
}

func ShowController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctrls, err := ShowControllerObj(claims)
	return setReply(c, err, ctrls)
}

func ShowControllerObj(claims *UserClaims) ([]ormapi.Controller, error) {
	ctrls := []ormapi.Controller{}
	if !enforcer.Enforce(claims.Username, "", ResourceControllers, ActionView) {
		return nil, echo.ErrForbidden
	}
	err := db.Find(&ctrls).Error
	if err != nil {
		return nil, dbErr(err)
	}
	return ctrls, nil
}
