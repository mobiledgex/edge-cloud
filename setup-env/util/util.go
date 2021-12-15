package util

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/cookiejar"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	influxclient "github.com/influxdata/influxdb/client/v2"
	dmeproto "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	"github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/integration/process"
	"github.com/mobiledgex/edge-cloud/testutil"
	yaml "github.com/mobiledgex/yaml/v2"
	"google.golang.org/grpc"
)

type TestSpec struct {
	Name             string            `json:"name" yaml:"name"`
	ApiType          string            `json:"api" yaml:"api"`
	ApiFile          string            `json:"apifile" yaml:"apifile"`
	ApiFileVars      map[string]string `json:"apifilevars" yaml:"apifilevars"`
	Actions          []string          `json:"actions" yaml:"actions"`
	RetryCount       int               `json:"retrycount" yaml:"retrycount"`
	RetryIntervalSec float64           `json:"retryintervalsec" yaml:"retryintervalsec"`
	CompareYaml      CompareYaml       `json:"compareyaml" yaml:"compareyaml"`
}

type CompareYaml struct {
	Yaml1    string `json:"yaml1" yaml:"yaml1"`
	Yaml2    string `json:"yaml2" yaml:"yaml2"`
	FileType string `json:"filetype" yaml:"filetype"`
}

var Deployment DeploymentData
var ApiAddrNone = "NONE"

type yamlFileType int

const (
	YamlAppdata yamlFileType = 0
	YamlOther   yamlFileType = 1
)

type SetupVariables struct {
	Vars     []map[string]string
	Includes []string
}

// replacement variables taken from the setup
var DeploymentReplacementVars map[string]string

type ReturnCodeWithText struct {
	Success bool
	Text    string
}

type GoogleCloudInfo struct {
	Cluster     string
	Zone        string
	MachineType string
}

type ClusterInfo struct {
	MexManifest string
}

type K8sPod struct {
	PodName  string
	PodCount int
	MaxWait  int
}

type K8CopyFile struct {
	PodName string
	Src     string
	Dest    string
}

type K8sDeploymentStep struct {
	File        string
	Description string
	WaitForPods []K8sPod
	CopyFiles   []K8CopyFile
}

type TLSCertInfo struct {
	CommonName string
	IPs        []string
	DNSNames   []string
}

type DeploymentData struct {
	TLSCerts       []*TLSCertInfo           `yaml:"tlscerts"`
	DockerNetworks []*process.DockerNetwork `yaml:"dockernetworks"`
	Locsims        []*process.LocApiSim     `yaml:"locsims"`
	Toksims        []*process.TokSrvSim     `yaml:"toksims"`
	Vaults         []*process.Vault         `yaml:"vaults"`
	Etcds          []*process.Etcd          `yaml:"etcds"`
	Controllers    []*process.Controller    `yaml:"controllers"`
	Dmes           []*process.Dme           `yaml:"dmes"`
	SampleApps     []*process.SampleApp     `yaml:"sampleapps"`
	Influxs        []*process.Influx        `yaml:"influxs"`
	ClusterSvcs    []*process.ClusterSvc    `yaml:"clustersvcs"`
	Crms           []*process.Crm           `yaml:"crms"`
	Jaegers        []*process.Jaeger        `yaml:"jaegers"`
	Traefiks       []*process.Traefik       `yaml:"traefiks"`
	NginxProxys    []*process.NginxProxy    `yaml:"nginxproxys"`
	NotifyRoots    []*process.NotifyRoot    `yaml:"notifyroots"`
	EdgeTurns      []*process.EdgeTurn      `yaml:"edgeturns"`
	ElasticSearchs []*process.ElasticSearch `yaml:"elasticsearchs"`
	RedisCaches    []*process.RedisCache    `yaml:"rediscaches"`
}

type errorReply struct {
	Code    int
	Message string
	Details []string
}

func GetAllProcesses() []process.Process {
	all := make([]process.Process, 0)
	for _, p := range Deployment.Locsims {
		all = append(all, p)
	}
	for _, p := range Deployment.Toksims {
		all = append(all, p)
	}
	for _, p := range Deployment.Vaults {
		all = append(all, p)
	}
	for _, p := range Deployment.Etcds {
		all = append(all, p)
	}
	for _, p := range Deployment.Controllers {
		all = append(all, p)
	}
	for _, p := range Deployment.Dmes {
		all = append(all, p)
	}
	for _, p := range Deployment.SampleApps {
		all = append(all, p)
	}
	for _, p := range Deployment.Influxs {
		all = append(all, p)
	}
	for _, p := range Deployment.ClusterSvcs {
		all = append(all, p)
	}
	for _, p := range Deployment.Jaegers {
		all = append(all, p)
	}
	for _, p := range Deployment.NginxProxys {
		all = append(all, p)
	}
	for _, p := range Deployment.Traefiks {
		all = append(all, p)
	}
	for _, p := range Deployment.NotifyRoots {
		all = append(all, p)
	}
	for _, p := range Deployment.EdgeTurns {
		all = append(all, p)
	}
	for _, p := range Deployment.ElasticSearchs {
		all = append(all, p)
	}
	for _, p := range Deployment.RedisCaches {
		all = append(all, p)
	}
	return all
}

func GetProcessByName(processName string) process.Process {
	for _, p := range GetAllProcesses() {
		if processName == p.GetName() {
			return p
		}
	}
	return nil
}

//these are strings which may be present in the yaml but not in the corresponding data structures.
//These are the only allowed exceptions to the strict yaml unmarshalling
var yamlExceptions = map[string]map[string]bool{
	"setup": {
		"vars": true,
	},
	"appdata": {
		"ip_str": true, // ansible workaround
	},
}

func IsYamlOk(e error, yamltype string) bool {
	rc := true
	errstr := e.Error()
	for _, err1 := range strings.Split(errstr, "\n") {
		allowedException := false
		for ye := range yamlExceptions[yamltype] {
			if strings.Contains(err1, ye) {
				allowedException = true
			}
		}

		if allowedException || strings.Contains(err1, "yaml: unmarshal errors") {
			// ignore this summary error
		} else {
			//all other errors are unexpected and mean something is wrong in the yaml
			log.Printf("Fatal Unmarshal Error in: %v\n", err1)
			rc = false
		}
	}
	return rc
}

func ConnectController(p *process.Controller, c chan ReturnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", p.Name, p.ApiAddr)
	api, err := p.ConnectAPI(20 * time.Second)
	if err != nil {
		c <- ReturnCodeWithText{false, "Failed to connect to " + p.Name}
	} else {
		c <- ReturnCodeWithText{true, "OK connect to " + p.Name}
		api.Close()
	}
}

//default is to connect to the first controller, unless we specified otherwise
func GetController(ctrlname string) *process.Controller {
	if ctrlname == "" {
		return Deployment.Controllers[0]
	}
	for _, ctrl := range Deployment.Controllers {
		if ctrl.Name == ctrlname {
			return ctrl
		}
	}
	log.Fatalf("Error: could not find specified controller: %v\n", ctrlname)
	return nil //unreachable
}

func GetDme(dmename string) *process.Dme {
	if dmename == "" {
		return Deployment.Dmes[0]
	}
	for _, dme := range Deployment.Dmes {
		if dme.Name == dmename {
			return dme
		}
	}
	log.Fatalf("Error: could not find specified dme: %v\n", dmename)
	return nil //unreachable
}

func ConnectDme(p *process.Dme, c chan ReturnCodeWithText) {
	log.Printf("attempt to connect to process %v at %v\n", p.Name, p.ApiAddr)
	api, err := p.ConnectAPI(20 * time.Second)
	if err != nil {
		c <- ReturnCodeWithText{false, "Failed to connect to " + p.Name}
	} else {
		c <- ReturnCodeWithText{true, "OK connect to " + p.Name}
		api.Close()
	}
}

func GetInflux(name string) *process.Influx {
	if name == "" {
		return Deployment.Influxs[0]
	}
	for _, influx := range Deployment.Influxs {
		if influx.Name == name {
			return influx
		}
	}
	log.Fatalf("Error: could not find specified influx: %s\n", name)
	return nil // unreachable
}

func checkCloudletState(p *process.Crm, timeout time.Duration) error {
	filter := edgeproto.CloudletInfo{}
	err := json.Unmarshal([]byte(p.CloudletKey), &filter.Key)
	if err != nil {
		return fmt.Errorf("unable to parse CloudletKey")
	}

	conn := connectOnlineController(timeout)
	if conn == nil {
		return fmt.Errorf("unable to connect to online controller")
	}

	infoapi := edgeproto.NewCloudletInfoApiClient(conn)
	show := testutil.ShowCloudletInfo{}
	startTimeMs := time.Now().UnixNano() / int64(time.Millisecond)
	maxTimeMs := int64(timeout/time.Millisecond) + startTimeMs
	wait := 20 * time.Millisecond
	err = fmt.Errorf("unable to check CloudletInfo")
	for {
		timeout -= wait
		time.Sleep(wait)
		currTimeMs := time.Now().UnixNano() / int64(time.Millisecond)
		if currTimeMs > maxTimeMs {
			err = fmt.Errorf("timed out, last error was %s", err.Error())
			break
		}
		show.Init()
		stream, showErr := infoapi.ShowCloudletInfo(context.Background(), &filter)
		show.ReadStream(stream, showErr)
		if showErr != nil {
			err = fmt.Errorf("show CloudletInfo failed: %s", showErr.Error())
			continue
		}
		info, found := show.Data[filter.Key.GetKeyString()]
		if !found {
			err = fmt.Errorf("CloudletInfo not found")
			continue
		}
		if info.State != dmeproto.CloudletState_CLOUDLET_STATE_READY && info.State != dmeproto.CloudletState_CLOUDLET_STATE_ERRORS {
			err = fmt.Errorf("CloudletInfo bad state %s", dmeproto.CloudletState_name[int32(info.State)])
			continue
		}
		err = nil
		break
	}
	return err
}

func connectOnlineController(delay time.Duration) *grpc.ClientConn {
	for _, ctrl := range Deployment.Controllers {
		conn, err := ctrl.ConnectAPI(delay)
		if err == nil {
			return conn
		}
	}
	return nil
}

func SetLogFormat() {
	log.SetFlags(log.Flags() | log.Ltime | log.Lmicroseconds)
}

func UnsetLogFormat() {
	log.SetFlags(log.Flags() & ^log.Ltime & ^log.Lmicroseconds)
}

func PrintBlankLine() {
	UnsetLogFormat()
	log.Printf("")
	SetLogFormat()
}

func PrintStartBanner(label string) {
	PrintBlankLine()
	log.Printf("  ***  %s\n", label)
}

func PrintStepBanner(label string) {
	PrintBlankLine()
	log.Printf("  ---  %s\n", label)
}

//for specific output that we want to put in a separate file.  If no
//output dir, just  print to the stdout
func PrintToFile(fname string, outputDir string, out string, truncate bool) {
	if outputDir == "" {
		fmt.Print(out)
	} else {
		outfile := outputDir + "/" + fname
		mode := os.O_APPEND
		if truncate {
			mode = os.O_TRUNC
		}
		ofile, err := os.OpenFile(outfile, mode|os.O_CREATE|os.O_WRONLY, 0666)
		defer ofile.Close()
		if err != nil {
			log.Fatalf("unable to append output file: %s, err: %v\n", outfile, err)
		}
		log.Printf("writing file: %s\n%s\n", fname, out)
		fmt.Fprint(ofile, out)
	}
}

func PrintToYamlFile(fname, outputDir string, data interface{}, truncate bool) {
	out, err := yaml.Marshal(data)
	if err != nil {
		log.Fatalf("yaml marshal data failed, %v, %+v\n", err, data)
	}
	PrintToFile(fname, outputDir, string(out), truncate)
}

//creates an output directory with an optional timestamp.  Server log files, output from APIs, and
//output from the script itself will all go there if specified
func CreateOutputDir(useTimestamp bool, outputDir string, logFileName string) string {
	if useTimestamp {
		startTimestamp := time.Now().Format("2006-01-02T150405")
		outputDir = outputDir + "/" + startTimestamp
	}
	fmt.Printf("Creating output dir: %s\n", outputDir)
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		log.Fatalf("Error trying to create directory %v: %v\n", outputDir, err)
	}

	logName := outputDir + "/" + logFileName
	logFile, err := os.OpenFile(logName, os.O_CREATE|os.O_APPEND|os.O_RDWR, 0666)

	if err != nil {
		log.Fatalf("Error creating logfile %s\n", logName)
	}
	//log to both stdout and logfile
	mw := io.MultiWriter(os.Stdout, logFile)
	log.SetOutput(mw)
	return outputDir
}

type ReadYamlOptions struct {
	vars                 map[string]string
	validateReplacedVars bool
}

type ReadYamlOp func(opts *ReadYamlOptions)

func ReadYamlFile(filename string, iface interface{}, ops ...ReadYamlOp) error {
	opts := ReadYamlOptions{}
	for _, op := range ops {
		op(&opts)
	}

	if strings.HasPrefix(filename, "~") {
		filename = strings.Replace(filename, "~", os.Getenv("HOME"), 1)
	}
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("error reading yaml file: %v err: %v\n", filename, err)
	}
	if opts.vars != nil {
		yamlStr, err := ReplaceVars(string(yamlFile), opts.vars)
		if err != nil {
			return err
		}
		yamlFile = []byte(yamlStr)
	}
	if opts.validateReplacedVars {
		//make sure there are no unreplaced variables left and inform the user if so
		re := regexp.MustCompile("{{(\\S+)}}")
		matches := re.FindAllStringSubmatch(string(yamlFile), 1)
		if len(matches) > 0 {
			return errors.New(fmt.Sprintf("Unreplaced variables in yaml: %v", matches))
		}
	}

	err = yaml.UnmarshalStrict(yamlFile, iface)
	if err != nil {
		return err
	}
	return nil
}

func ReadFile(filename string, replaceVars map[string]string) (string, error) {
	dat, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	str, err := ReplaceVars(string(dat), replaceVars)
	if err != nil {
		return "", err
	}
	return str, nil
}

// Replace variables denoted as {{variablename}}
func ReplaceVars(contents string, replaceVars map[string]string) (string, error) {
	if replaceVars == nil {
		return contents, nil
	}
	for k, v := range replaceVars {
		if strings.HasPrefix(v, "ENV=") {
			// environment variable replacement var
			envVarName := strings.Replace(v, "ENV=", "", 1)
			envVarVal := os.Getenv(envVarName)
			if envVarVal == "" {
				return "", fmt.Errorf("environment variable not set: %s", envVarName)
			}
			v = envVarVal
		}
		contents = strings.Replace(contents, "{{"+k+"}}", v, -1)
	}
	return contents, nil
}

func WithVars(vars map[string]string) ReadYamlOp {
	return func(opts *ReadYamlOptions) {
		opts.vars = vars
	}
}

func ValidateReplacedVars() ReadYamlOp {
	return func(opts *ReadYamlOptions) {
		opts.validateReplacedVars = true
	}
}

//compares two yaml files for equivalence
//TODO need to handle different types of interfaces besides appdata, currently using
//that to sort
func CompareYamlFiles(name string, actions []string, compare *CompareYaml) bool {
	PrintStepBanner("running compareYamlFiles")

	compareInfo, err := yaml.Marshal(compare)
	if err != nil {
		log.Printf("Failed to marshal compare info, %v\n", err)
		return false
	}
	log.Printf("Name: %s", name)
	log.Printf("Actions: %s\n%s", strings.Join(actions, ", "), string(compareInfo))

	// figure out which file is the expected data based on the path
	diffArgs := []string{"-au"}
	expectedFile := ""
	if strings.Contains(compare.Yaml1, "github.com/mobiledgex/") {
		diffArgs = append(diffArgs, compare.Yaml1, compare.Yaml2)
		expectedFile = compare.Yaml1
	} else {
		diffArgs = append(diffArgs, compare.Yaml2, compare.Yaml1)
		expectedFile = compare.Yaml2
	}
	runDir := filepath.Dir(expectedFile)
	diffFile := expectedFile + ".patch"

	cmd := exec.Command("diff", diffArgs...)
	cmd.Dir = runDir
	out, err := cmd.CombinedOutput()
	if err != nil {
		log.Println("Comparison fail")
		log.Print(string(out))
		err := ioutil.WriteFile(diffFile, out, 0644)
		if err != nil {
			log.Printf("Failed to write diff output to %s: %v", diffFile, err)
		}
		return false
	}
	os.Remove(diffFile)
	log.Println("Comparison success")
	return true
}

func ControllerCLI(ctrl *process.Controller, args ...string) ([]byte, error) {
	cmdargs := []string{"--addr", ctrl.ApiAddr, "controller"}
	tlsFile := ctrl.GetTlsFile()
	if tlsFile != "" {
		cmdargs = append(cmdargs, "--tls", tlsFile)
	}
	cmdargs = append(cmdargs, args...)
	log.Printf("Running: edgectl %v\n", cmdargs)
	cmd := exec.Command("edgectl", cmdargs...)
	return cmd.CombinedOutput()
}

func CallRESTPost(httpAddr string, client *http.Client, pb proto.Message, reply proto.Message) error {
	str, err := new(jsonpb.Marshaler).MarshalToString(pb)
	if err != nil {
		log.Printf("Could not marshal request\n")
		return err
	}
	bytesRep := []byte(str)
	req, err := http.NewRequest("POST", httpAddr, bytes.NewBuffer(bytesRep))
	if err != nil {
		log.Printf("Failed to create a request\n")
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Failed to HTTP <%s>\n", httpAddr)
		return err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	reader := bytes.NewReader(body)
	err = jsonpb.Unmarshal(reader, reply)

	if err != nil {
		log.Printf("Failed to unmarshal reply : %s %v\n", body, err)
		//try to unmarshal it as an error yaml reply
		var ereply errorReply
		err2 := json.Unmarshal(body, &ereply)
		if err2 == nil {
			log.Printf("Reply is an error response, message: %+v\n", ereply.Message)
			return fmt.Errorf("Error reply message: %s", ereply.Message)
		}
		// not an error reply either
		log.Printf("Failed to unmarshal as an error reply : %s %v\n", body, err2)

		// return the original error
		return err
	}
	return nil
}

func FilterInfluxTime(results []influxclient.Result) {
	for ii, _ := range results {
		for jj, _ := range results[ii].Series {
			row := &results[ii].Series[jj]
			if len(row.Columns) < 1 || row.Columns[0] != "time" {
				continue
			}
			// first value in each point is time,
			// zero it out so we can ignore it
			for tt, _ := range row.Values {
				pt := row.Values[tt]
				if len(pt) > 0 {
					pt[0] = 0
				}
			}
		}
	}
}

func FilterCloudletInfoNocmp(data *edgeproto.AllData) {
	for ii, _ := range data.CloudletInfos {
		data.CloudletInfos[ii].Controller = ""
		data.CloudletInfos[ii].NotifyId = 0
	}
	for ii, _ := range data.Cloudlets {
		data.Cloudlets[ii].CrmAccessPublicKey = ""
	}
	for ii, _ := range data.Cloudlets {
		data.Cloudlets[ii].SecondaryCrmAccessPublicKey = ""
	}
}

func FilterFindCloudletReply(reply *dmeproto.FindCloudletReply) {
	reply.EdgeEventsCookie = ""
}

func FilterAppInstEdgeEventsCookies(appInstReply *dmeproto.AppInstListReply) {
	for _, cloudlet := range appInstReply.Cloudlets {
		for _, appinst := range cloudlet.Appinstances {
			appinst.EdgeEventsCookie = ""
		}
	}

}

func FilterQosPositionKpiReply(reply *dmeproto.QosPositionKpiReply) {
	for _, p := range reply.PositionResults {
		// nil the actual values as they are unpredictable.  We will just
		//compare GPS locations vs positionIds
		p.LatencyMin = 0
		p.LatencyMax = 0
		p.LatencyAvg = 0
		p.DluserthroughputMin = 0
		p.DluserthroughputMax = 0
		p.DluserthroughputAvg = 0
		p.UluserthroughputMin = 0
		p.UluserthroughputMax = 0
		p.UluserthroughputAvg = 0
	}
}

func FilterServerEdgeEvent(event *dmeproto.ServerEdgeEvent) {
	if event.Statistics != nil {
		event.Statistics.Timestamp.Seconds = 0
		event.Statistics.Timestamp.Nanos = 0
	}
	if event.NewCloudlet != nil {
		event.NewCloudlet.EdgeEventsCookie = ""
		//for ii := range event.NewCloudlet.Ports {
		//	event.NewCloudlet.Ports[ii].PublicPort = 0
		//}
	}
}

func FilterAlerts(alerts []edgeproto.Alert) {
	sort.Slice(alerts, func(i, j int) bool {
		return fmt.Sprint(alerts[i].Labels) < fmt.Sprint(alerts[j].Labels)
	})
	clearTags := map[string]struct{}{
		"nocmp":     struct{}{},
		"timestamp": struct{}{},
	}
	for ii := range alerts {
		alerts[ii].ClearTagged(clearTags)
	}
}

func FilterAppDataOutputStatus(output *testutil.AllDataOut) {
	output.Cloudlets = testutil.FilterStreamResults(output.Cloudlets)
	output.ClusterInsts = testutil.FilterStreamResults(output.ClusterInsts)
	output.AppInstances = testutil.FilterStreamResults(output.AppInstances)
}

func FilterNodeData(data *edgeproto.NodeData) {
	clearTags := map[string]struct{}{
		"nocmp":     struct{}{},
		"timestamp": struct{}{},
	}
	data.Sort()
	data.ClearTagged(clearTags)
	for ii, _ := range data.Nodes {
		// Only compare keys of properties, since Node Property
		//  values are not constant.
		ClearMapValues(data.Nodes[ii].Properties)
	}
}

// clears map values so that only key names are compared
func ClearMapValues(m map[string]string) {
	if m == nil {
		return
	}
	for k, _ := range m {
		m[k] = ""
	}
}

func ReadConsoleURL(consoleUrl string, cookies []*http.Cookie) (string, error) {
	req, err := http.NewRequest("GET", consoleUrl, nil)
	if err != nil {
		return "", err
	}

	if cookies != nil {
		for _, cookie := range cookies {
			req.AddCookie(cookie)
		}
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	options := cookiejar.Options{}

	jar, err := cookiejar.New(&options)
	if err != nil {
		return "", err
	}

	client := &http.Client{
		Transport: tr,
		Jar:       jar,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// For some reason this client is not getting 302,
	// instead it gets 502. It works fine for curl & wget
	if resp.StatusCode == http.StatusBadGateway {
		if resp.Request.URL.String() != consoleUrl {
			return ReadConsoleURL(resp.Request.URL.String(), resp.Cookies())
		}
	}
	return string(contents), nil
}

// Apply a transform to the value.
type TransformFunc func(v reflect.Value)

type Transformer struct {
	transforms []*transform
}

type transform struct {
	targetType  reflect.Type
	targetField string
	txFunc      TransformFunc
}

func NewTransformer() *Transformer {
	t := Transformer{}
	t.transforms = []*transform{}
	return &t
}

func (s *Transformer) AddTransform(target reflect.Type, field string, fn TransformFunc) {
	tx := &transform{
		targetType:  target,
		targetField: field,
		txFunc:      fn,
	}
	s.transforms = append(s.transforms, tx)
}

func (s *Transformer) AddSetZeroType(typs ...interface{}) {
	for _, typ := range typs {
		t := reflect.TypeOf(typ)
		s.AddTransform(t, "", s.SetZero)
	}
}

func (s *Transformer) AddSetZeroTypeField(typ interface{}, fields ...string) {
	t := reflect.TypeOf(typ)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for _, field := range fields {
		s.AddTransform(t, field, s.SetZero)
	}
}

func (s *Transformer) SetZero(v reflect.Value) {
	if !v.CanSet() {
		return
	}
	t := v.Type()
	z := reflect.Zero(t)
	v.Set(z)
}

func (s *Transformer) Apply(obj interface{}) {
	s.applyRecurse(reflect.ValueOf(obj))
}

func (s *Transformer) applyRecurse(v reflect.Value) {
	for _, tx := range s.transforms {
		if tx.targetType == v.Type() && tx.targetField == "" {
			tx.txFunc(v)
		}
	}
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		s.applyRecurse(v.Elem())
	} else if v.Kind() == reflect.Struct {
		for ii := 0; ii < v.NumField(); ii++ {
			sf := v.Type().Field(ii)
			// skip unexported fields, they cannot be set by reflect
			if sf.PkgPath != "" {
				continue
			}
			subv := v.Field(ii)
			// apply transform on field
			for _, tx := range s.transforms {
				if tx.targetType == v.Type() && tx.targetField == sf.Name {
					tx.txFunc(subv)
				}
			}
			s.applyRecurse(subv)

		}
	} else if v.Kind() == reflect.Slice {
		for ii := 0; ii < v.Len(); ii++ {
			subv := v.Index(ii)
			s.applyRecurse(subv)
		}
	} else if v.Kind() == reflect.Map {
		iter := v.MapRange()
		for iter.Next() {
			subv := iter.Value()
			// Maps are a special case because their values may not
			// addressable. If so they are not modifiable in-place,
			// so we need to make a copy and insert it into the map.
			newCopy := false
			if !subv.CanSet() {
				newSubv := reflect.New(subv.Type()).Elem()
				if !newSubv.CanSet() {
					fmt.Printf("subvSettable %v:%v cannot set\n", iter.Key(), subv.Type())
					continue
				}
				newSubv.Set(subv)
				subv = newSubv
				newCopy = true
			}
			s.applyRecurse(subv)
			if newCopy {
				v.SetMapIndex(iter.Key(), subv)
			}
		}
	}
}
