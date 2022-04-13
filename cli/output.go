// Copyright 2022 MobiledgeX, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cli

import (
	"encoding/json"
	"fmt"
	"io"

	yaml "github.com/mobiledgex/yaml/v2"
)

const (
	OutputFormatYaml        = "yaml"
	OutputFormatJson        = "json"
	OutputFormatJsonCompact = "json-compact"
	OutputFormatTable       = "table"
)

var OutputFormats = []string{
	OutputFormatYaml,
	OutputFormatJson,
	OutputFormatJsonCompact,
	OutputFormatTable,
}

// note slightly different function in cmdsup.WriteOutputGeneric,
// perhaps we can consolidate them.
func (c *Command) WriteOutput(out io.Writer, objs interface{}, format string) error {
	switch format {
	case OutputFormatYaml:
		output, err := yaml.Marshal(objs)
		if err != nil {
			return fmt.Errorf("yaml failed to marshal: %v\n", err)
		}
		fmt.Fprint(out, string(output))
	case OutputFormatJson:
		output, err := json.MarshalIndent(objs, "", "  ")
		if err != nil {
			return fmt.Errorf("json failed to marshal: %v\n", err)
		}
		fmt.Fprintln(out, string(output))
	case OutputFormatJsonCompact:
		output, err := json.Marshal(objs)
		if err != nil {
			return fmt.Errorf("json failed to marshal: %v\n", err)
		}
		fmt.Fprintln(out, string(output))
	case OutputFormatTable:
		return fmt.Errorf("table output format not supported yet")
	default:
		return fmt.Errorf("invalid output format %s", format)
	}
	return nil
}
