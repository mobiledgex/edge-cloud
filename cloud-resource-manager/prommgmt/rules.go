package prommgmt

import (
	"github.com/prometheus/common/model"
)

// These structs are based on the ones in Prometheus here:
// https://github.com/prometheus/prometheus/blob/51c824543bca2f25baa6596827cf4fad8f18dc2e/pkg/rulefmt/rulefmt.go#L105
// We could reference the Prometheus structs directly, but I didn't see the need
// to introduce a large set of dependencies just for a few very simple structs.
// Unfortunately Prometheus uses a custom duration type which does not support
// the same string formats as the standard time.Duration, so we must use that one.

type Rule struct {
	Record      string            `yaml:"record,omitempty"`
	Alert       string            `yaml:"alert,omitempty"`
	Expr        string            `yaml:"expr"`
	For         model.Duration    `yaml:"for,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type RuleGroup struct {
	Name       string `yaml:"name"`
	Rules      []Rule `yaml:"rules"`
	org        string
	commonTags map[string]string
}

func NewRuleGroup(name, org string) *RuleGroup {
	grp := RuleGroup{}
	grp.Name = name
	grp.org = org
	return &grp
}

type GroupsData struct {
	Groups []RuleGroup `yaml:"groups"`
}
