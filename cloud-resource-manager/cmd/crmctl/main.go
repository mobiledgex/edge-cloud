package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"strings"

	"github.com/bobbae/q"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"google.golang.org/grpc"
)

type stringList []string

func (s *stringList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringList) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

var commandList = []string{
	"add-cloud-resource",
	"list-cloud-resource",
	"delete-cloud-resource",
	"deploy-application",
	"delete-application",
}

func main() {
	if len(os.Args) < 2 {
		printHelpAndExit()
	}

	switch os.Args[1] {
	case "add-cloud-resource":
		doAddCloudResource()
	case "list-cloud-resource":
		doListCloudResource()
	case "delete-cloud-resource":
		doDeleteCloudResource()
	case "deploy-application":
		doDeployApplication()
	case "delete-application":
		doDeleteApplication()
	default:
		printHelpAndExit()
	}
}

func printHelpAndExit() {
	fmt.Printf("Available commands are:\n")
	for _, cmd := range commandList {
		fmt.Printf("\t%s\n", cmd)
	}
	os.Exit(1)
}

func getAPI(address string) (edgeproto.CloudResourceManagerClient, error) {
	q.Q("grpc dial", address)

	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	q.Q("getting CRM api")
	api := edgeproto.NewCloudResourceManagerClient(conn)

	return api, nil
}

func doAddCloudResource() {
	cmd := flag.NewFlagSet("add-cloud-resource", flag.ExitOnError)

	name := cmd.String("name", "", "Name of the cloudlet (required)")
	address := cmd.String("address", "", "Address of the cloudlet (required)")
	location := cmd.String("location", "", "Location of the cloudlet (required)")
	opkey := cmd.String("opkey", "", "Operator Key for the cloudlet (required)")
	opkeyname := cmd.String("opkeyname", "", "Operator Key Name for the cloudlet (required)")

	crm := cmd.String("crm", "", "Address of Cloud Resource Manager (required)")

	cmd.Parse(os.Args[2:])

	if *name == "" || *address == "" || *location == "" || *opkey == "" || *opkeyname == "" || *crm == "" {
		cmd.PrintDefaults()
		os.Exit(1)
	}

	api, err := getAPI(*crm)
	if err != nil {
		fatalError("can't get API endpoint %v", err)
	}

	ctx := context.TODO()

	okey := edgeproto.OperatorKey{Name: *opkey}
	cloudletKey := edgeproto.CloudletKey{OperatorKey: okey, Name: *opkeyname}

	cr := edgeproto.CloudResource{CloudletKey: &cloudletKey}
	cr.Name = *name
	cr.AccessIp = net.ParseIP(*address)

	res, err := api.AddCloudResource(ctx, &cr)

	if err != nil {
		fatalError("AddCloudResource call failed %v", err)
	}

	fmt.Printf("AddCloudResource call succeeded, %v", *res)
}

func fatalError(msg string, err error) {
	fmt.Printf(msg, err)
	os.Exit(1)
}

func doListCloudResource() {
	cmd := flag.NewFlagSet("list-cloud-resource", flag.ExitOnError)

	category := cmd.Int("category", 0, "Category of the cloudlet to list")
	crm := cmd.String("crm", "", "Address of Cloud Resource Manager (required)")

	cmd.Parse(os.Args[2:])

	if *category < 0 {
		cmd.PrintDefaults()
		os.Exit(1)
	}

	if *crm == "" {
		cmd.PrintDefaults()
		os.Exit(1)
	}

	api, err := getAPI(*crm)
	if err != nil {
		fatalError("can't get API endpoint, error %v", err)
		os.Exit(1)
	}

	ctx := context.TODO()
	cloudletKey := edgeproto.CloudletKey{}
	cr := edgeproto.CloudResource{CloudletKey: &cloudletKey}
	cr.Category = edgeproto.CloudResourceCategory(*category)

	q.Q("calling ListCloudResource")

	stream, err := api.ListCloudResource(ctx, &cr)
	if err != nil {
		fatalError("ListCloudResource call failed, %v", err)
	}

	waitc := make(chan struct{})

	go func() {
		for {
			q.Q("wait for stream input")
			in, err := stream.Recv()
			if err == io.EOF {
				close(waitc)
				return
			}

			q.Q("received from stream", *in)
			if err != nil {
				fatalError("Failed to receive from stream, %v", err)
			}

			fmt.Println(*in)
		}
	}()

	<-waitc
	q.Q("stream done")
}

func doDeleteCloudResource() {
	cmd := flag.NewFlagSet("delete-cloud-resource", flag.ExitOnError)

	name := cmd.String("name", "", "Name of the cloudlet (required)")
	address := cmd.String("address", "", "Address of the cloudlet (required)")
	location := cmd.String("location", "", "Location of the cloudlet (required)")
	opkey := cmd.String("opkey", "", "Operator Key for the cloudlet (required)")
	opkeyname := cmd.String("opkeyname", "", "Operator Key Name for the cloudlet (required)")

	crm := cmd.String("crm", "", "Address of Cloud Resource Manager (required)")

	cmd.Parse(os.Args[2:])

	if *name == "" || *address == "" || *location == "" || *opkey == "" || *opkeyname == "" || *crm == "" {
		cmd.PrintDefaults()
		os.Exit(1)
	}

	api, err := getAPI(*crm)
	if err != nil {
		fatalError("can't get API endpoint %v", err)
	}

	ctx := context.TODO()

	okey := edgeproto.OperatorKey{Name: *opkey}
	cloudletKey := edgeproto.CloudletKey{OperatorKey: okey, Name: *opkeyname}

	cr := edgeproto.CloudResource{CloudletKey: &cloudletKey}
	cr.Name = *name
	cr.AccessIp = net.ParseIP(*address)

	res, err := api.DeleteCloudResource(ctx, &cr)

	if err != nil {
		fatalError("DeleteCloudResource call failed %v", err)
	}

	fmt.Printf("DeleteCloudResource call succeeded, %v", *res)
}

func doDeployApplication() {
	cmd := flag.NewFlagSet("deploy-application", flag.ExitOnError)

	crm := cmd.String("crm", "", "Address of Cloud Resource Manager (required)")
	name := cmd.String("name", "", "Name of the application (required)")
	kind := cmd.String("kind", "", "Type of the application, e.g. k8s-simple, k8s-manifest, (required)")
	image := cmd.String("image", "", "Image name and version (e.g. myapp:1.1.1) of the application (required)")
	namespace := cmd.String("namespace", "default", "Namespace of the application")
	replicas := cmd.Int("replicas", 1, "Number of replicas for the application")
	exposure := cmd.String("exposure", "", "Exposure specification, e.g. http,80 (required)")

	//TODO context, limitfactor, cpu, memory, repository, ...

	cmd.Parse(os.Args[2:])

	if *crm == "" || *name == "" || *kind == "" || *image == "" {
		cmd.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.TODO()

	api, err := getAPI(*crm)
	if err != nil {
		fatalError("can't get API endpoint %v", err)
	}

	edgeapplication := edgeproto.EdgeCloudApplication{}

	edgeapplication.Kind = *kind

	edgeapp := edgeproto.EdgeCloudApp{}
	edgeapp.Name = *name
	edgeapp.Image = *image
	edgeapp.Namespace = *namespace
	edgeapp.Replicas = int32(*replicas)
	edgeapp.Exposure = *exposure

	appInstKey := edgeproto.AppInstKey{}

	edgeapp.AppInstKey = &appInstKey
	edgeapplication.Apps = []*edgeproto.EdgeCloudApp{&edgeapp}

	res, err := api.DeployApplication(ctx, &edgeapplication)

	if err != nil {
		fatalError("DeployApplication call failed %v", err)
	}

	fmt.Printf("DeployApplication call succeeded, %v", *res)
}

func doDeleteApplication() {
	//TODO delete options as per k8s metav1.DeleteOptions

	cmd := flag.NewFlagSet("delete-application", flag.ExitOnError)

	name := cmd.String("name", "", "Name of the application (required)")
	crm := cmd.String("crm", "", "Address of Cloud Resource Manager (required)")
	kind := cmd.String("kind", "", "Type of the application, e.g. k8s-simple, k8s-manifest, (required)")
	namespace := cmd.String("namespace", "default", "Namespace of the application")

	cmd.Parse(os.Args[2:])

	if *crm == "" || *name == "" || *kind == "" {
		cmd.PrintDefaults()
		os.Exit(1)
	}

	ctx := context.TODO()

	api, err := getAPI(*crm)
	if err != nil {
		fatalError("can't get API endpoint %v", err)
	}

	edgeapplication := edgeproto.EdgeCloudApplication{}

	edgeapplication.Kind = *kind

	edgeapp := edgeproto.EdgeCloudApp{}
	edgeapp.Name = *name
	edgeapp.Namespace = *namespace

	appInstKey := edgeproto.AppInstKey{}

	edgeapp.AppInstKey = &appInstKey
	edgeapplication.Apps = []*edgeproto.EdgeCloudApp{&edgeapp}

	res, err := api.DeleteApplication(ctx, &edgeapplication)

	if err != nil {
		fatalError("DeleteApplication call failed %v", err)
	}

	fmt.Printf("DeleteApplication call succeeded, %v", *res)
}
