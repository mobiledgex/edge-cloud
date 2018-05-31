package edgeproto

import (
	"errors"

	"github.com/mobiledgex/edge-cloud/util"
)

// TypeString functions

func (key *DeveloperKey) TypeString() string { return "developer" }

func (key *OperatorKey) TypeString() string { return "operator" }

func (key *AppKey) TypeString() string { return "app" }

func (key *CloudletKey) TypeString() string { return "cloudlet" }

func (key *AppInstKey) TypeString() string { return "appinst" }

// Validate functions to validate user input

func (key *DeveloperKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid developer name")
	}
	return nil
}

func (s *Developer) Validate() error {
	return s.GetKey().Validate()
}

func (key *OperatorKey) Validate() error {
	if !util.ValidName(key.Name) {
		return errors.New("Invalid operator name")
	}
	return nil
}

func (s *Operator) Validate() error {
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

func (s *App) Validate() error {
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

func (s *Cloudlet) Validate() error {
	if err := s.GetKey().Validate(); err != nil {
		return err
	}
	if s.AccessIp != nil && !util.ValidIp(s.AccessIp) {
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

func (s *AppInst) Validate() error {
	if err := s.GetKey().Validate(); err != nil {
		return err
	}
	if s.Liveness == AppInst_UNKNOWN {
		return errors.New("Unknown liveness specified")
	}
	if !util.ValidIp(s.Ip) {
		return errors.New("Invalid IP specified")
	}
	return nil
}
