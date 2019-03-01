package orm

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/log"
	"github.com/mobiledgex/edge-cloud/mc/ormapi"
	"google.golang.org/grpc"
)

type RegionContext struct {
	region string
	claims *UserClaims
	conn   *grpc.ClientConn
}

func newResCb(c echo.Context, desc string) func(*edgeproto.Result) {
	return func(res *edgeproto.Result) {
		streamReplyMsg(c, desc, res.Message)
	}
}

func CreateData(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	data := ormapi.AllData{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	// stream back responses
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	for _, ctrl := range data.Controllers {
		desc := fmt.Sprintf("Create Controller region %s", ctrl.Region)
		err := CreateControllerObj(claims, &ctrl)
		streamReply(c, desc, err)
	}
	for _, org := range data.Orgs {
		desc := fmt.Sprintf("Create Organization %s", org.Name)
		err := CreateOrgObj(claims, &org)
		streamReply(c, desc, err)
	}
	for _, role := range data.Roles {
		desc := fmt.Sprintf("Add User Role %v", role)
		err := AddUserRoleObj(claims, &role)
		streamReply(c, desc, err)
	}
	for _, regionData := range data.RegionData {
		conn, err := connectController(regionData.Region)
		if err != nil {
			desc := fmt.Sprintf("Connect %s Controller", regionData.Region)
			streamReply(c, desc, err)
			continue
		}
		defer conn.Close()

		rc := &RegionContext{}
		rc.claims = claims
		rc.region = regionData.Region
		rc.conn = conn

		appdata := &regionData.AppData
		for _, flavor := range appdata.Flavors {
			desc := fmt.Sprintf("Create Flavor %s", flavor.Key.Name)
			err = CreateFlavorObj(rc, &flavor)
			streamReply(c, desc, err)
		}
		for _, clusterflavor := range appdata.ClusterFlavors {
			desc := fmt.Sprintf("Create ClusterFlavor %s", clusterflavor.Key.Name)
			err = CreateClusterFlavorObj(rc, &clusterflavor)
			streamReply(c, desc, err)
		}
		for _, cloudlet := range appdata.Cloudlets {
			desc := fmt.Sprintf("Create Cloudlet %v", cloudlet.Key)
			cb := newResCb(c, desc)
			err = CreateCloudletStream(rc, &cloudlet, cb)
			streamReply(c, desc, err)
		}
		for _, cinst := range appdata.ClusterInsts {
			desc := fmt.Sprintf("Create ClusterInst %v", cinst.Key)
			cb := newResCb(c, desc)
			err = CreateClusterInstStream(rc, &cinst, cb)
			streamReply(c, desc, err)
		}
		for _, app := range appdata.Applications {
			desc := fmt.Sprintf("Create App %v", app.Key)
			err = CreateAppObj(rc, &app)
			streamReply(c, desc, err)
		}
		for _, appinst := range appdata.AppInstances {
			desc := fmt.Sprintf("Create AppInst %v", appinst.Key)
			cb := newResCb(c, desc)
			err = CreateAppInstStream(rc, &appinst, cb)
			streamReply(c, desc, err)
		}
	}
	return nil
}

func DeleteData(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	data := ormapi.AllData{}
	if err := c.Bind(&data); err != nil {
		return c.JSON(http.StatusBadRequest, Msg("Invalid POST data"))
	}
	// stream back responses
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	c.Response().WriteHeader(http.StatusOK)

	for _, regionData := range data.RegionData {
		conn, err := connectController(regionData.Region)
		if err != nil {
			desc := fmt.Sprintf("Connect %s Controller", regionData.Region)
			streamReply(c, desc, err)
			continue
		}
		defer conn.Close()

		rc := &RegionContext{}
		rc.claims = claims
		rc.region = regionData.Region
		rc.conn = conn

		appdata := &regionData.AppData

		for _, appinst := range appdata.AppInstances {
			desc := fmt.Sprintf("Delete AppInst %v", appinst.Key)
			cb := newResCb(c, desc)
			err = DeleteAppInstStream(rc, &appinst, cb)
			streamReply(c, desc, err)
		}
		for _, app := range appdata.Applications {
			desc := fmt.Sprintf("Delete App %v", app.Key)
			err = DeleteAppObj(rc, &app)
			streamReply(c, desc, err)
		}
		for _, cinst := range appdata.ClusterInsts {
			desc := fmt.Sprintf("Delete ClusterInst %v", cinst.Key)
			cb := newResCb(c, desc)
			err = DeleteClusterInstStream(rc, &cinst, cb)
			streamReply(c, desc, err)
		}
		for _, cloudlet := range appdata.Cloudlets {
			desc := fmt.Sprintf("Delete Cloudlet %v", cloudlet.Key)
			cb := newResCb(c, desc)
			err = DeleteCloudletStream(rc, &cloudlet, cb)
			streamReply(c, desc, err)
		}
		for _, clusterflavor := range appdata.ClusterFlavors {
			desc := fmt.Sprintf("Delete ClusterFlavor %s", clusterflavor.Key.Name)
			err = DeleteClusterFlavorObj(rc, &clusterflavor)
			streamReply(c, desc, err)
		}
		for _, flavor := range appdata.Flavors {
			desc := fmt.Sprintf("Delete Flavor %s", flavor.Key.Name)
			err = DeleteFlavorObj(rc, &flavor)
			streamReply(c, desc, err)
		}
	}
	// roles must be deleted after orgs, otherwise we may delete the
	// role that's needed to be able to delete the org.
	for _, org := range data.Orgs {
		desc := fmt.Sprintf("Delete Organization %s", org.Name)
		err := DeleteOrgObj(claims, &org)
		streamReply(c, desc, err)
	}
	for _, role := range data.Roles {
		desc := fmt.Sprintf("Remove User Role %v", role)
		err := RemoveUserRoleObj(claims, &role)
		streamReply(c, desc, err)
	}
	for _, ctrl := range data.Controllers {
		desc := fmt.Sprintf("Delete Controller region %s", ctrl.Region)
		err := DeleteControllerObj(claims, &ctrl)
		streamReply(c, desc, err)
	}
	return c.JSON(http.StatusOK, Msg("Data delete done"))
}

func ShowData(c echo.Context) error {
	claims, err := getClaims(c)
	if err != nil {
		return err
	}
	data := ormapi.AllData{}

	ctrls, err := ShowControllerObj(claims)
	if err == nil {
		data.Controllers = ctrls
	}
	orgs, err := ShowOrgObj(claims)
	if err == nil {
		data.Orgs = orgs
	}
	roles, err := ShowUserRoleObj(claims)
	if err == nil {
		data.Roles = roles
	}

	// Iterate over all controllers. We need to look up
	// controllers this time without enforcement check.
	ctrls = []ormapi.Controller{}
	err = db.Find(&ctrls).Error
	if err != nil {
		return c.JSON(http.StatusOK, data)
	}
	for _, ctrl := range ctrls {
		conn, err := connectControllerAddr(ctrl.Address)
		if err != nil {
			log.DebugLog(log.DebugLevelApi, "ShowData connect controller", "ctrl", ctrl, "err", err)
			continue
		}
		defer conn.Close()

		rc := &RegionContext{}
		rc.claims = claims
		rc.region = ctrl.Region
		rc.conn = conn

		regionData := &ormapi.RegionData{}
		regionData.Region = ctrl.Region
		appdata := &regionData.AppData

		log.DebugLog(log.DebugLevelApi, "ShowData controller", "ctrl", ctrl)
		cloudlets, err := ShowCloudletObj(rc, &edgeproto.Cloudlet{})
		if err == nil {
			appdata.Cloudlets = cloudlets
		}
		flavors, err := ShowFlavorObj(rc, &edgeproto.Flavor{})
		if err == nil {
			appdata.Flavors = flavors
		}
		cflavors, err := ShowClusterFlavorObj(rc, &edgeproto.ClusterFlavor{})
		if err == nil {
			appdata.ClusterFlavors = cflavors
		}
		cinsts, err := ShowClusterInstObj(rc, &edgeproto.ClusterInst{})
		if err == nil {
			appdata.ClusterInsts = cinsts
		}
		apps, err := ShowAppObj(rc, &edgeproto.App{})
		if err == nil {
			appdata.Applications = apps
		}
		appinsts, err := ShowAppInstObj(rc, &edgeproto.AppInst{})
		if err == nil {
			appdata.AppInstances = appinsts
		}

		if len(flavors) > 0 || len(cflavors) > 0 ||
			len(cloudlets) > 0 || len(cinsts) > 0 ||
			len(apps) > 0 || len(appinsts) > 0 {
			data.RegionData = append(data.RegionData, *regionData)
		}
	}
	return c.JSON(http.StatusOK, data)
}
