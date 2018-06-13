package cmdsup

import (
	"fmt"

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
