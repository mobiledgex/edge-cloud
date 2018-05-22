package crmutil

import (
	"github.com/gophercloud/gophercloud"
	"github.com/gophercloud/gophercloud/openstack"
	"github.com/gophercloud/gophercloud/openstack/compute/v2/servers"

	"github.com/bobbae/q"
)

type OpenstackServerArgs struct {
	Region, Name, Image, Flavor, Network string
}

func GetOpenstackClient(region string) (*gophercloud.ServiceClient, error) {
	authOpts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		return nil, err
	}

	provider, err := openstack.AuthenticatedClient(authOpts)
	if err != nil {
		return nil, err
	}

	client, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	return client, nil
}

// CreateOpenstackServer creates a Nova VM instance which
// is referred to as `server` in Openstack terminology.
func CreateOpenstackServer(client *gophercloud.ServiceClient,
	args *OpenstackServerArgs) error {
	actual, err := servers.Create(client, servers.CreateOpts{
		Name:      args.Name,
		ImageRef:  args.Image,
		FlavorRef: args.Flavor,
		Networks: []servers.Network{
			servers.Network{
				UUID: args.Network,
			},
		},
	}).Extract()

	if err != nil {
		return err
	}

	q.Q(*actual)

	return nil
}

// DeleteOpenstackServer deletes a server identified by `id`.
func DeleteOpenstackServer(client *gophercloud.ServiceClient, id string) error {
	res := servers.Delete(client, id)
	if res.Err != nil {
		return res.Err
	}

	return nil
}
