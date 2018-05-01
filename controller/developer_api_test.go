package main

import (
	"context"
	"testing"

	"github.com/mobiledgex/edge-cloud/proto"
	"github.com/mobiledgex/edge-cloud/util"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

type Server struct {
	data map[proto.DeveloperKey]proto.Developer
	grpc.ServerStream
}

func (x *Server) Send(m *proto.Developer) error {
	x.data[*m.Key] = *m
	return nil
}

func TestDeveloperApi(t *testing.T) {
	util.SetDebugLevel(util.DebugLevelEtcd | util.DebugLevelApi)

	dummy := dummyEtcd{}
	dummy.Start()

	api := InitDeveloperApi(&dummy)
	ctx := context.TODO()

	// test create
	_, err := api.CreateDeveloper(ctx, &Dev1)
	assert.Nil(t, err, "Failed to create %s", Dev1.Key.Name)

	_, err = api.CreateDeveloper(ctx, &Dev2)
	assert.Nil(t, err, "Failed to create %s", Dev2.Key.Name)

	_, err = api.CreateDeveloper(ctx, &Dev3)
	assert.Nil(t, err, "Failed to create %s", Dev3.Key.Name)

	_, err = api.CreateDeveloper(ctx, &Dev1)
	assert.NotNil(t, err, "Failed to detect conflict")

	// check create and test show
	server := Server{}
	server.data = make(map[proto.DeveloperKey]proto.Developer)
	filterNone := proto.Developer{}

	err = api.ShowDeveloper(&filterNone, &server)
	assert.Nil(t, err, "Failed to show developers")

	_, found := server.data[*Dev1.Key]
	assert.True(t, found, "Failed to show Dev1")
	_, found = server.data[*Dev2.Key]
	assert.True(t, found, "Failed to show Dev2")
	_, found = server.data[*Dev3.Key]
	assert.True(t, found, "Failed to show Dev3")
	_, found = server.data[*Dev4.Key]
	assert.False(t, found, "Should not have found missing key")
	assert.Equal(t, 3, len(server.data), "Wrong count")

	// test show filtering
	filterDev1 := proto.Developer{
		Address: Dev1.Address,
	}
	server.data = make(map[proto.DeveloperKey]proto.Developer)
	err = api.ShowDeveloper(&filterDev1, &server)
	assert.Nil(t, err, "Failed to show developers")

	_, found = server.data[*Dev1.Key]
	assert.True(t, found, "Failed to filter for Atlantic address")
	assert.Equal(t, 1, len(server.data), "Filtering for Atlantic")

	// test delete
	_, err = api.DeleteDeveloper(ctx, &Dev2)
	assert.Nil(t, err, "Failed to delete developer")

	server.data = make(map[proto.DeveloperKey]proto.Developer)
	err = api.ShowDeveloper(&filterNone, &server)
	assert.Nil(t, err, "Failed to show developers")

	_, found = server.data[*Dev2.Key]
	assert.False(t, found, "Failed to delete Dev2")
	assert.Equal(t, 2, len(server.data), "Wrong count")

	// test update
	Dev3.Email = "new.google.com"
	_, err = api.UpdateDeveloper(ctx, &Dev3)
	assert.Nil(t, err, "Failed to update Dev3")

	server.data = make(map[proto.DeveloperKey]proto.Developer)
	err = api.ShowDeveloper(&filterNone, &server)
	assert.Nil(t, err, "Failed to show developers")

	check, found := server.data[*Dev3.Key]
	assert.True(t, found, "Did not find Dev3")
	assert.Equal(t, Dev3.Email, check.Email, "update email")

	// update just one field - note rest of in4 is empty
	in4 := proto.Developer{Key: &proto.DeveloperKey{}}
	*in4.Key = *Dev3.Key
	newEmail := "update just this"
	in4.Email = newEmail
	in4.Fields = util.GrpcFieldsNew()
	util.GrpcFieldsSet(in4.Fields, proto.DeveloperFieldEmail)
	_, err = api.UpdateDeveloper(ctx, &in4)
	assert.Nil(t, err, "Update")

	in4 = Dev3
	in4.Email = newEmail
	server.data = make(map[proto.DeveloperKey]proto.Developer)
	err = api.ShowDeveloper(&filterNone, &server)
	assert.Nil(t, err, "Failed to show developers")
	check, found = server.data[*Dev3.Key]
	assert.True(t, found, "find Dev3")
	assert.Equal(t, in4, check, "check equal")

	// test update of missing developer
	_, err = api.UpdateDeveloper(ctx, &Dev4)
	assert.NotNil(t, err, "Update missing")

	dummy.Stop()
}
