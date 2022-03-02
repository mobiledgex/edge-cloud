package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/tools/go/packages"
)

// This tool starts from a (or several) main golang package directories, and traverses
// the package dependencies, gathering all direct and transient dependencies.
// The dependencies are then either
// 1) printed out with links for attestation documentation, or
// 2) license files are printed into a single output for third party compliance

// Some commands:

// shipping (from edge-cloud-infra):
// parsedeps --validatelinks ../edge-cloud/cloud-resource-manager/cmd/crmserver/ ../edge-cloud/d-match-engine/dme-server ./plugin/platform/ ./plugin/edgeevents ./shepherd ./shepherd/shepherd_platform

// not shipping (from edge-cloud-infra):
// parsedeps --validatelinks ../edge-cloud/controller ../edge-cloud/cluster-svc ../edge-cloud/edgeturn ./mc ./alertmgr-sidecar ./autoprov ../edge-cloud/notifyroot

// not shipping (test) (from edge-cloud-infra):
// parsedeps --validatelinks ../edge-cloud/setup-env/e2e-tests ../edge-cloud/setup-env/test-mex ./e2e-tests/test-mex-infra ./shepherd/e2eHttpServer ./shepherd/fake_envoy_exporter ../edge-cloud/protoc-gen-gomex ../edge-cloud/protoc-gen-test ../edge-cloud/protoc-gen-cmd ./protoc-gen-mc2 ../edge-cloud/protoc-gen-controller ../edge-cloud/protoc-gen-controller-test ../edge-cloud/protoc-gen-notify

var gennotice = flag.Bool("gennotice", false, "output third party notice file")
var validatelinks = flag.Bool("validatelinks", false, "validate license links")

func main() {
	flag.Parse()
	binDirs := flag.Args()

	pkgCfg := &packages.Config{
		Context: context.Background(),
		Mode:    packages.NeedDeps | packages.NeedImports | packages.NeedName | packages.NeedFiles,
	}

	depMap := map[string]*PkgDep{}
	licenseMatcher := NewLicenseMatcher()

	failed := false
	for _, binDir := range binDirs {
		binPkgs, err := packages.Load(pkgCfg, binDir)
		if err != nil {
			log.Fatal(err.Error())
		}

		count := 0
		packages.Visit(binPkgs, func(dep *packages.Package) bool {
			if strings.HasPrefix(dep.PkgPath, "github.com/mobiledgex/edge-cloud/") || strings.HasPrefix(dep.PkgPath, "github.com/mobiledgex/edge-cloud-infra/") {
				// self-referenced package, visit but do not process as dep
				return true
			}
			pkgFile := ""
			if len(dep.GoFiles) > 0 {
				pkgFile = dep.GoFiles[0]
			} else if len(dep.CompiledGoFiles) > 0 {
				pkgFile = dep.CompiledGoFiles[0]
			} else {
				// empty package
				return false
			}
			pkgDir := filepath.Dir(pkgFile)

			if strings.HasPrefix(pkgDir, "/usr/local/go") {
				// skip standard lib
				return false
			}

			verParts := strings.Split(pkgDir, "@")
			pkgRepo := verParts[0]

			pkgDep, ok := depMap[pkgRepo]
			if !ok {
				version := verParts[1]
				// remove trailing directory
				verParts = strings.Split(version, "/")
				version = verParts[0]

				lic := getLicensePath(pkgDir)
				if lic == "" {
					fmt.Printf("Cannot find license file in %s\n", pkgDir)
					failed = true
				}
				licType, _ := getLicType(licenseMatcher, lic)

				pkgDep = &PkgDep{
					repo:     pkgRepo,
					pkg:      dep,
					dir:      pkgDir,
					binaries: map[string]struct{}{},
					version:  version,
					license:  lic,
					licType:  licType,
				}
				depMap[pkgRepo] = pkgDep
			}
			pkgDep.binaries[path.Base(binDir)] = struct{}{}
			count++
			return true
		}, nil)
		if count == 0 {
			log.Fatal("No deps in " + binDir)
		}
	}
	if failed {
		return
	}

	deps := []*PkgDep{}
	for _, dep := range depMap {
		deps = append(deps, dep)
	}
	sort.Slice(deps, func(i, j int) bool {
		return deps[i].pkg.ID < deps[j].pkg.ID
	})

	if *gennotice {
		text := `The Edge-Cloud Platform uses third-party libraries or other resources
that may be distributed under licenses different than the one
from this software. These licenses are listed below.`
		fmt.Printf("%s\n\n", text)
	}

	errors := []error{}
	for _, dep := range deps {
		if *gennotice {
			err := genNotice(dep)
			if err != nil {
				errors = append(errors, err)
			}
		} else {
			genRef(dep)
		}
	}
	if len(errors) > 0 {
		for _, err := range errors {
			fmt.Fprintf(os.Stderr, "%s\n", err)
		}
		os.Exit(1)
	}
}

type PkgDep struct {
	repo     string
	pkg      *packages.Package
	dir      string
	binaries map[string]struct{}
	version  string
	license  string
	licType  string
}

func getLicensePath(pkgDir string) string {
	files := []string{"LICENSE", "LICENSE.txt", "LICENSE.md", "NOTICE"}
	for {
		if len(pkgDir) == 0 || strings.HasSuffix(pkgDir, "go/pkg/mod") {
			break
		}
		for _, f := range files {
			lic := pkgDir + "/" + f
			if st, err := os.Stat(lic); err == nil && !st.IsDir() {
				return lic
			}
		}
		// go up one dir
		pkgDir = filepath.Dir(pkgDir)
	}
	return ""
}

func getSourceMirror(pkgDir string) (string, error) {
	repo := strings.TrimPrefix(pkgDir, os.Getenv("HOME")+"/go/pkg/mod/")
	repo = strings.Replace(repo, "!", "", -1)
	localDir := repo
	for to, from := range pathXlat {
		repo = strings.Replace(repo, to, from, -1)
	}
	repo = strings.Replace(repo, "/", "-", -1)
	repo = strings.Replace(repo, "@", ".", -1)
	repo += ".tar"
	tarFile := path.Base(repo)
	repo = "https://storage.googleapis.com/mobiledgex-downloads/open-source-mirror/" + repo
	// make sure mirror exists
	resp, err := http.Get(repo)
	if err != nil {
		err = fmt.Errorf("Failed to retrieve mirrored code at %s: %s", repo, err)
	}
	if err == nil && resp.StatusCode == http.StatusNotFound {
		// create a tar file to upload
		tarCmd := exec.Command("tar", "-cvf", tarFile, localDir)
		tarCmd.Dir = os.Getenv("HOME") + "/go/pkg/mod"
		out, cmdErr := tarCmd.CombinedOutput()
		if cmdErr != nil {
			log.Fatal(fmt.Sprintf(`tar command "(cd %s; %s)" failed: %s, %s`, tarCmd.Dir, tarCmd.String(), string(out), cmdErr))
		}
		err = fmt.Errorf("Please upload %s/%s to online storage at %s for mirroring", tarCmd.Dir, tarFile, repo)
	}
	if err == nil && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
		err = fmt.Errorf("Get %s failed: %d\n", repo, resp.StatusCode)
	}
	return repo, err
}

func genNotice(dep *PkgDep) error {
	var mirrorErr error
	fmt.Printf("---------------------------------------------------\n")
	fmt.Printf("License notice for %s\n", dep.pkg.PkgPath)
	if licRequiresSourceMirror(dep.licType) {
		mirrorPath, merr := getSourceMirror(dep.dir)
		fmt.Printf("A copy of the source code can be found at %s\n\n", mirrorPath)
		mirrorErr = merr
	}
	fmt.Printf("---------------------------------------------------\n\n")
	lic, err := ioutil.ReadFile(dep.license)
	if err != nil {
		log.Fatal(err.Error())
	}
	fmt.Printf("%s\n\n", string(lic))
	return mirrorErr
}

// trailing API version which is present on disk, but not in the URL
var reGithubApiVer = regexp.MustCompile("/v[0-9.]+@")

// version matching to be able to remove the version if it not a real
// tag or branch on github.
var reGithubBlobVer = regexp.MustCompile("/blob/v.+/")

// gopkg.in includes a .v# suffix which must be removed for the actual url
var rePkgInVer = regexp.MustCompile("\\.v[0-9]+@")

// These sites are all aliases for pkg.go.dev
var pkgGoDevSites = []string{
	"go.mongodb.org",
	"go.opencensus.io",
	"go.uber.org",
	"golang.org",
	"google.golang.org",
	"gortc.io",
	"cloud.google.com",
}

// Weird path translations, not sure how these are set up.
// These translate from local go/pkg/mod directory paths to
// online web repository paths.
var pathXlat = map[string]string{
	"github.com/hashicorp/vault/api": "github.com/hashicorp/vault",
	"github.com/hashicorp/vault/sdk": "github.com/hashicorp/vault",
	"gopkg.in/inf.v":                 "gopkg.in/go-inf/inf.v",
	"gopkg.in/yaml.v":                "gopkg.in/go-yaml/yaml.v",
}

func genRef(dep *PkgDep) {
	// get license link
	link := strings.TrimPrefix(dep.license, os.Getenv("HOME")+"/go/pkg/mod/")
	link = strings.Replace(link, "!", "", -1)
	// there may be a few possible links to try
	links := []string{}

	for to, from := range pathXlat {
		link = strings.Replace(link, to, from, -1)
	}

	// some links are aliases for github, translate those first
	if hasPrefix(link, "gopkg.in") {
		// ex: gopkg.in/inf.v0@v0.9.1/LICENSE
		link = strings.Replace(link, "gopkg.in", "github.com", 1)
		// ex: github.com/inf.v0@v0.9.1/LICENSE
		link = rePkgInVer.ReplaceAllString(link, "@")
		// ex: github.com/inf@v0.9.1/LICENSE
	} else if hasPrefix(link, "k8s.io") {
		link = strings.Replace(link, "k8s.io", "github.com/kubernetes", 1)
	} else if hasPrefix(link, "sigs.k8s.io") {
		link = strings.Replace(link, "k8s.io", "github.com/kubernetes-sigs", 1)
	}

	if hasPrefix(link, "github.com") {
		// ex: github.com/elastic/go-elasticsearch/v7@v7.5.0/LICENSE
		link = reGithubApiVer.ReplaceAllString(link, "@")
		// ex: github.com/elastic/go-elasticsearch@v7.5.0/LICENSE
		link = strings.Replace(link, "@", "/blob/", 1)
		// ex: github.com/elastic/go-elasticsearch/blob/v7.5.0/LICENSE
		// it's possible there are no tags/branches, and the version
		// is just a commit id. Often it won't work to refer via
		// commit id, so try pointing to master.
		link2 := reGithubBlobVer.ReplaceAllString(link, "/blob/master/")
		// ex: github.com/elastic/go-elasticsearch/blob/master/LICENSE
		links = append(links, link, link2)
	} else if hasPrefix(link, pkgGoDevSites...) {
		link = "pkg.go.dev/" + dep.pkg.PkgPath + "@" + dep.version + "?tab=licenses"
		// it's possible there is no tags in the repo, so the
		// version-based url won't work. Since no tags, use "master".
		link2 := "pkg.go.dev/" + dep.pkg.PkgPath + "?tab=licenses"
		links = append(links, link, link2)
	} else {
		links = append(links, link)
	}

	// test that link is reachable
	validLink := fmt.Sprintf("INVALID:%v", links)
	if *validatelinks {
		for _, try := range links {
			try = "https://" + try
			resp, err := httpGetRetry(try)
			if err == nil && resp.StatusCode == http.StatusOK {
				validLink = try
				break
			}
		}
	} else {
		validLink = "https://" + links[0]
	}

	bins := []string{}
	for b, _ := range dep.binaries {
		bins = append(bins, b)
	}
	sort.Strings(bins)

	cols := []string{
		strings.Join(bins, ","),  // binaries
		"Application dependency", // how relates
		dep.pkg.PkgPath,          // lib name
		dep.version,              // lib version
		"Static",                 // linking: static or dynamic
		"No",                     // modified
		validLink,                // license link
		dep.licType,              // license name
	}
	fmt.Printf("%s\n", strings.Join(cols, "\t"))
}

func hasPrefix(base string, prefixes ...string) bool {
	for _, pre := range prefixes {
		if strings.HasPrefix(base, pre) {
			return true
		}
	}
	return false
}

// Github rate limits https requests, which we'll hit.
// Be able to retry after github tells us to slow down.
func httpGetRetry(link string) (*http.Response, error) {
	resp, err := http.Get(link)
	if err == nil && resp.StatusCode == http.StatusTooManyRequests {
		if vals, ok := resp.Header["Retry-After"]; ok && len(vals) > 0 {
			sec, err := strconv.Atoi(vals[0])
			if err == nil {
				time.Sleep(time.Duration(sec) * time.Second)
				resp, err = http.Get(link)
			} else {
				fmt.Printf("Failed to parse RetryAfter val %v, %v\n", vals, err)
				log.Fatal("fail")
			}
		} else {
			fmt.Printf("RetryAfter not found in header, %+v\n", resp)
			log.Fatal("fail")
		}
	}
	return resp, err
}

func getLicType(matcher *LicenseMatcher, licFile string) (string, float32) {
	contents, err := ioutil.ReadFile(licFile)
	if err != nil {
		fmt.Printf("Failed to read %s, %s\n", licFile, err)
		log.Fatal(err.Error())
	}
	return matcher.Match(string(contents))
}

func licRequiresSourceMirror(licType string) bool {
	if licType == MozillaPublicLicense2 {
		return true
	}
	return false
}
