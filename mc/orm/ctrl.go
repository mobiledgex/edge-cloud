package orm

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	"github.com/mobiledgex/edge-cloud/tls"
	"google.golang.org/grpc"
)

var tlsCertFile = flag.String("tls", "", "server tls cert file")

func connectController(region string) (*grpc.ClientConn, error) {
	addr, err := getControllerAddrForRegion(region)
	if err != nil {
		return nil, err
	}
	dialOption, err := tls.GetTLSClientDialOption(addr, *tlsCertFile)
	if err != nil {
		return nil, err
	}
	return grpc.Dial(addr, dialOption)
}

func getControllerAddrForRegion(region string) (string, error) {
	ctrl := Controller{
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
	ctrl := Controller{}
	if err := c.Bind(&ctrl); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid Post data"))
	}
	if ctrl.Region == "" {
		return c.JSON(http.StatusBadRequest, Msg("Region not specified"))
	}
	if ctrl.Address == "" {
		return c.JSON(http.StatusBadRequest, Msg("Address not specified"))
	}
	if !enforcer.Enforce(id2str(claims.UserID), "", ResourceControllers, ActionManage) {
		return echo.ErrForbidden
	}
	err = db.Create(&ctrl).Error
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, Msg("Controller created"))
}

func DeleteController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctrl := Controller{}
	if err := c.Bind(&ctrl); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid Post data"))
	}
	if !enforcer.Enforce(id2str(claims.UserID), "", ResourceControllers, ActionManage) {
		return echo.ErrForbidden
	}
	err = db.Delete(&ctrl).Error
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, Msg("Controller deleted"))
}

func ShowController(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	ctrls := []Controller{}
	if !enforcer.Enforce(id2str(claims.UserID), "", ResourceControllers, ActionView) {
		return echo.ErrForbidden
	}
	err = db.Find(&ctrls).Error
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, ctrls)
}
