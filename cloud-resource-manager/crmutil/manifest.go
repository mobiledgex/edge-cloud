package crmutil

import "fmt"

const APIversion = "v1"

func CheckManifest(mf *Manifest) error {
	if mf.APIVersion == "" {
		return fmt.Errorf("mf apiversion not set")
	}
	if mf.APIVersion != APIversion {
		return fmt.Errorf("invalid api version")
	}
	err := CheckManifestValues(&mf.Values)
	if err != nil {
		return err
	}
	return nil
}

func CheckManifestValues(mfv *ValueDetail) error {
	if mfv.ExternalNetwork == "" {
		return fmt.Errorf("missing external network")
	}
	return nil
}
