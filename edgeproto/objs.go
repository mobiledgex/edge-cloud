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
		return a.AppInstances[i].Key.GetKeyString() < a.AppInstances[j].Key.GetKeyString()
	})
	sort.Slice(a.Applications[:], func(i, j int) bool {
		return a.Applications[i].Key.GetKeyString() < a.Applications[i].Key.GetKeyString()
	})
	sort.Slice(a.Cloudlets[:], func(i, j int) bool {
		return a.Cloudlets[i].Key.GetKeyString() < a.Cloudlets[j].Key.GetKeyString()
	})
	sort.Slice(a.Developers[:], func(i, j int) bool {
		return a.Developers[i].Key.GetKeyString() < a.Developers[j].Key.GetKeyString()
	})
	sort.Slice(a.Operators[:], func(i, j int) bool {
		return a.Operators[i].Key.GetKeyString() < a.Operators[j].Key.GetKeyString()
	})
}

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
	return nil
}

func (s *CloudletInfo) Validate(fields map[string]struct{}) error {
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
	return nil
}

func (s *AppInstInfo) Validate(fields map[string]struct{}) error {
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
