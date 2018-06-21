package edgeproto

import (
	"errors"
	"sort"
	strings "strings"

	"github.com/mobiledgex/edge-cloud/util"
)

// contains sets of each applications for yaml marshalling
type ApplicationData struct {
	Operators    []Operator  `yaml:"operators"`
	Cloudlets    []Cloudlet  `yaml:"cloudlets"`
	Developers   []Developer `yaml:"developers"`
	Applications []App       `yaml:"apps"`
	AppInstances []AppInst   `yaml:"appinstances"`
}

// sort each slice by key
func (a *ApplicationData) Sort() {
	sort.Slice(a.AppInstances[:], func(i, j int) bool {
		return a.AppInstances[i].Key.Id < a.AppInstances[j].Key.Id
	})
	sort.Slice(a.Applications[:], func(i, j int) bool {
		return a.Applications[i].Key.Name < a.Applications[i].Key.Name
	})
	sort.Slice(a.Cloudlets[:], func(i, j int) bool {
		return a.Cloudlets[i].Key.Name < a.Cloudlets[j].Key.Name
	})
	sort.Slice(a.Developers[:], func(i, j int) bool {
		return a.Developers[i].Key.Name < a.Developers[j].Key.Name
	})
	sort.Slice(a.Operators[:], func(i, j int) bool {
		return a.Operators[i].Key.Name < a.Operators[j].Key.Name
	})
}

// TypeString functions

func DeveloperKeyTypeString() string         { return "developer" }
func (key *DeveloperKey) TypeString() string { return DeveloperKeyTypeString() }

func OperatorKeyTypeString() string         { return "operator" }
func (key *OperatorKey) TypeString() string { return OperatorKeyTypeString() }

func AppKeyTypeString() string         { return "app" }
func (key *AppKey) TypeString() string { return AppKeyTypeString() }

func CloudletKeyTypeString() string         { return "cloudlet" }
func (key *CloudletKey) TypeString() string { return CloudletKeyTypeString() }

func AppInstKeyTypeString() string         { return "appinst" }
func (key *AppInstKey) TypeString() string { return AppInstKeyTypeString() }

// Validate functions to validate user input

func (key *DeveloperKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid developer name")
	}
	return nil
}

func (s *Developer) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *OperatorKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid operator name")
	}
	return nil
}

func (s *Operator) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *AppKey) Validate() error {
	if err := key.DeveloperKey.Validate(); err != nil {
		return err
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid app name")
	}
	if !util.ValidName(key.Version) {
		return errors.New("Invalid app version string")
	}
	return nil
}

func (s *App) Validate(fields map[string]struct{}) error {
	return s.GetKey().Validate()
}

func (key *CloudletKey) Validate() error {
	if err := key.OperatorKey.Validate(); err != nil {
		return err
	}
	if !util.ValidName(key.Name) {
		return errors.New("Invalid cloudlet name")
	}
	return nil
}

func (s *Cloudlet) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().Validate(); err != nil {
		return err
	}
	if HasField(fields, CloudletFieldAccessIp) && s.AccessIp != nil &&
		!util.ValidIp(s.AccessIp) {
		return errors.New("Invalid access ip format")
	}
	return nil
}

func (key *AppInstKey) Validate() error {
	if err := key.AppKey.Validate(); err != nil {
		return err
	}
	if err := key.CloudletKey.Validate(); err != nil {
		return err
	}
	if key.Id == 0 {
		return errors.New("AppInst Id cannot be zero")
	}
	return nil
}

func (s *AppInst) Validate(fields map[string]struct{}) error {
	if err := s.GetKey().Validate(); err != nil {
		return err
	}
	if HasField(fields, AppInstFieldLiveness) && s.Liveness == AppInst_UNKNOWN {
		return errors.New("Unknown liveness specified")
	}
	if HasField(fields, AppInstFieldIp) && !util.ValidIp(s.Ip) {
		return errors.New("Invalid IP specified")
	}
	return nil
}

func MakeFieldMap(fields []string) map[string]struct{} {
	fmap := make(map[string]struct{})
	if fields == nil {
		return fmap
	}
	for _, set := range fields {
		for {
			fmap[set] = struct{}{}
			idx := strings.LastIndex(set, ".")
			if idx == -1 {
				break
			}
			set = set[:idx]
		}
	}
	return fmap
}

func HasField(fmap map[string]struct{}, field string) bool {
	_, ok := fmap[field]
	return ok
}
