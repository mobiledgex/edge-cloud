package cmdsup

import (
	"encoding/json"
	"fmt"

	yaml "github.com/mobiledgex/yaml/v2"
	"github.com/spf13/pflag"
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

var OutputFormat = OutputFormatYaml

func AddOutputFormatFlag(flagSet *pflag.FlagSet) {
	formats := fmt.Sprintf("%v", OutputFormats)
	flagSet.StringVar(&OutputFormat, "output-format", OutputFormatYaml, formats)
}

// HideTags is a comma separated list of tag names that are matched
// against the protocmd.hidetag field option. Fields that match will
// be zeroed-out on output objects, effectively hiding them when outputing
// json or yaml. How this is set is up to the user, but typically a
// global flag can be used.
var HideTags string

func AddHideTagsFormatFlag(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&HideTags, "hidetags", "", "comma separated list of hide tags")
}

func WriteOutputGeneric(objs interface{}) {
	switch OutputFormat {
	case OutputFormatYaml:
		output, err := yaml.Marshal(objs)
		if err != nil {
			fmt.Printf("Yaml failed to marshal: %s\n", err)
			return
		}
		fmt.Print(string(output))
	case OutputFormatJson:
		output, err := json.MarshalIndent(objs, "", "  ")
		if err != nil {
			fmt.Printf("Json failed to marshal: %s\n", err)
			return
		}
		fmt.Println(string(output))
	case OutputFormatJsonCompact:
		output, err := json.Marshal(objs)
		if err != nil {
			fmt.Printf("Json failed to marshal: %s\n", err)
			return
		}
		fmt.Println(string(output))
	}
}
