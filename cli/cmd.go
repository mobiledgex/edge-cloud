package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"reflect"
	"strings"

	dme "github.com/mobiledgex/edge-cloud/d-match-engine/dme-proto"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	yaml "github.com/mobiledgex/yaml/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var Parsable bool
var Data string
var Datafile string
var Debug bool
var OutputStream bool
var SilenceUsage bool
var OutputFormat = OutputFormatYaml

func AddInputFlags(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&Data, "data", "", "json formatted input data, alternative to name=val args list")
	flagSet.StringVar(&Datafile, "datafile", "", "file containing json/yaml formatted input data, alternative to name=val args list")
}

func AddOutputFlags(flagSet *pflag.FlagSet) {
	flagSet.StringVar(&OutputFormat, "output-format", OutputFormatYaml, fmt.Sprintf("output format: %s, %s, or %s", OutputFormatYaml, OutputFormatJson, OutputFormatJsonCompact))
	flagSet.BoolVar(&Parsable, "parsable", false, "generate parsable output")
	flagSet.BoolVar(&OutputStream, "output-stream", true, "stream output incrementally if supported by command")
}

func AddDebugFlag(flagSet *pflag.FlagSet) {
	flagSet.BoolVar(&Debug, "debug", false, "debug")
	flagSet.BoolVar(&SilenceUsage, "silence-usage", false, "silence-usage")
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

type Command struct {
	Use                  string
	Short                string
	RequiredArgs         string
	OptionalArgs         string
	AliasArgs            string
	SpecialArgs          *map[string]string
	Comments             map[string]string
	ReqData              interface{}
	ReplyData            interface{}
	PasswordArg          string
	VerifyPassword       bool
	DataFlagOnly         bool
	StreamOut            bool
	StreamOutIncremental bool
	CobraCmd             *cobra.Command
	Run                  func(c *Command, args []string) error
	UsageIsHelp          bool
	Annotations          map[string]string
}

func (c *Command) GenCmd() *cobra.Command {
	short := c.Short
	cmd := &cobra.Command{
		Use:         c.Use,
		Short:       short,
		Annotations: c.Annotations,
	}
	cmd.SetUsageFunc(c.usageFunc)
	cmd.SetHelpFunc(c.helpFunc)
	c.CobraCmd = cmd

	if c.Run != nil {
		cmd.RunE = c.runE
	}
	return cmd
}

func (c *Command) helpFunc(cmd *cobra.Command, args []string) {
	c.UsageIsHelp = true
	c.usageFunc(cmd)
}

func (c *Command) usageFunc(cmd *cobra.Command) error {
	out := cmd.OutOrStderr()
	if c.UsageIsHelp {
		fmt.Fprintf(out, "%s\n\n", cmd.Short)
	}
	fmt.Fprintf(out, "Usage: %s [args]\n", cmd.UseLine())

	pad := 0
	allargs := append(strings.Split(c.RequiredArgs, " "), strings.Split(c.OptionalArgs, " ")...)
	for _, str := range allargs {
		if len(str) > pad {
			pad = len(str)
		}
	}
	pad += 2

	required := c.requiredArgsHelp(pad)
	if required != "" {
		fmt.Fprint(out, "\n", required)
	}
	optional := c.optionalArgsHelp(pad)
	if optional != "" {
		fmt.Fprint(out, "\n", optional)
	}
	if cmd.HasAvailableLocalFlags() {
		fmt.Fprint(out, "\nFlags:\n", LocalFlagsUsageNoNewline(cmd))
	}
	if c.UsageIsHelp {
		// help needs the newline, but usage does not
		fmt.Fprint(out, "\n")
	}
	return nil
}

func usageArgs(str string) []string {
	args := strings.Fields(str)
	for ii, _ := range args {
		args[ii] = args[ii] + "="
	}
	return args
}

func (c *Command) requiredArgsHelp(pad int) string {
	if strings.TrimSpace(c.RequiredArgs) == "" {
		return ""
	}
	args := strings.Split(c.RequiredArgs, " ")
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "Required Args:\n")
	fmt.Fprint(&buf, c.argsHelp(pad, args))
	return buf.String()
}

func (c *Command) optionalArgsHelp(pad int) string {
	if strings.TrimSpace(c.OptionalArgs) == "" {
		return ""
	}
	args := strings.Split(c.OptionalArgs, " ")
	buf := bytes.Buffer{}
	fmt.Fprintf(&buf, "Optional Args:\n")
	fmt.Fprint(&buf, c.argsHelp(pad, args))
	return buf.String()
}

func (c *Command) argsHelp(pad int, args []string) string {
	buf := bytes.Buffer{}
	for _, str := range args {
		comment := ""
		if c.Comments != nil {
			comment = c.Comments[str]
		}
		fmt.Fprintf(&buf, "  %-*s%s\n", pad, str, comment)
	}
	return buf.String()
}

func (c *Command) runE(cmd *cobra.Command, args []string) error {
	return c.Run(c, args)
}

// ParseInput converts args to generic map.
// Input can come in 3 flavors, arg=value lists, yaml data, or json data.
// The output is generic map[string]interface{} holding the data,
// but normally each input format would have slightly different values
// for the map keys (field names) due to json or yaml tags being different
// from the go struct field name and each other. So we settle on the
// output map using json field names for consistency.
func (c *Command) ParseInput(args []string) (*MapData, error) {
	var in *MapData
	if Datafile != "" {
		byt, err := ioutil.ReadFile(Datafile)
		if err != nil {
			return nil, err
		}
		Data = string(byt)
	}
	if Data != "" {
		indata := make(map[string]interface{})
		err := json.Unmarshal([]byte(Data), &indata)
		if err == nil && c.ReqData != nil {
			err = json.Unmarshal([]byte(Data), c.ReqData)
			if err != nil {
				return nil, err
			}
			in = &MapData{
				Namespace: JsonNamespace,
				Data:      indata,
			}
		} else {
			// try yaml
			indata := make(map[string]interface{})
			err2 := yaml.Unmarshal([]byte(Data), &indata)
			if err2 != nil {
				return nil, fmt.Errorf("unable to unmarshal json or yaml data, %v, %v", err, err2)
			}
			in = &MapData{
				Namespace: YamlNamespace,
				Data:      indata,
			}
			// convert yaml map to json map
			in, err = JsonMap(in, c.ReqData)
			if err != nil {
				return nil, fmt.Errorf("failed to convert yaml map to json map, %v", err)
			}
			if c.ReqData != nil {
				err = yaml.Unmarshal([]byte(Data), c.ReqData)
				if err != nil {
					return nil, err
				}
			}
		}
	} else {
		if c.DataFlagOnly {
			return nil, fmt.Errorf("--data must be used to supply json/yaml-formatted input data")
		}
		input := Input{
			RequiredArgs:   strings.Fields(c.RequiredArgs),
			AliasArgs:      strings.Fields(c.AliasArgs),
			SpecialArgs:    c.SpecialArgs,
			PasswordArg:    c.PasswordArg,
			VerifyPassword: c.VerifyPassword,
			DecodeHook:     DecodeHook,
		}
		argsMap, err := input.ParseArgs(args, c.ReqData)
		if err != nil {
			return nil, err
		}
		if Debug {
			fmt.Printf("argsmap: %v\n", argsMap)
		}
		if c.ReqData != nil {
			// convert to json map
			in, err = JsonMap(argsMap, c.ReqData)
			if err != nil {
				return nil, err
			}
			if Debug {
				fmt.Printf("jsonmap: %v\n", in)
			}
		} else {
			in = argsMap
		}
	}
	return in, nil
}

func DecodeHook(from, to reflect.Type, data interface{}) (interface{}, error) {
	data, err := edgeproto.DecodeHook(from, to, data)
	if err != nil {
		return data, err
	}
	return dme.EnumDecodeHook(from, to, data)
}

func GenGroup(use, short string, cmds []*Command) *cobra.Command {
	if use == "" {
		panic("Use (command name) cannot be empty")
	}
	if short == "" {
		panic("Short description cannot be empty")
	}
	groupCmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE:  GroupRunE,
	}

	for _, c := range cmds {
		groupCmd.AddCommand(c.GenCmd())
	}
	gc := GroupCommand{}
	groupCmd.SetUsageFunc(gc.groupUsageFunc)
	groupCmd.SetHelpFunc(gc.groupHelpFunc)

	return groupCmd
}

type GroupCommand struct {
	UsageIsHelp bool
}

func (c *GroupCommand) groupHelpFunc(cmd *cobra.Command, args []string) {
	c.UsageIsHelp = true
	c.groupUsageFunc(cmd)
}

func (c *GroupCommand) groupUsageFunc(cmd *cobra.Command) error {
	out := cmd.OutOrStderr()
	if c.UsageIsHelp {
		fmt.Fprintf(out, "%s\n\n", cmd.Short)
	}
	fmt.Fprintf(out, "Usage: %s [command]\n", cmd.UseLine())
	fmt.Fprintf(out, "\nAvailable Commands:\n")
	pad := 0
	for _, sub := range cmd.Commands() {
		if pad < len(sub.Use) {
			pad = len(sub.Use)
		}
	}
	pad += 2
	if pad < 11 {
		// matches default cobra behavior of min padding of 11
		pad = 11
	}
	for _, sub := range cmd.Commands() {
		fmt.Fprintf(out, "  %-*s%s\n", pad, sub.Use, sub.Short)
	}
	if cmd.HasAvailableLocalFlags() {
		fmt.Fprint(out, "\nFlags:\n", LocalFlagsUsageNoNewline(cmd))
	}
	if c.UsageIsHelp {
		// help needs the newline, but usage does not
		fmt.Fprint(out, "\n")
	}
	return nil
}

// For group commands, if no subcommand is specified,
// we want to return an error (unless help was specified).
func GroupRunE(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("Please specify a command")
}

// For usage, the cobra code converts the entire usage to a string
// and then calls Println on it. This ends up introducing an extra
// unsightly newline in the output. Flags should be last, so this
// gets the flags output without the final new line.
func LocalFlagsUsageNoNewline(cmd *cobra.Command) string {
	str := cmd.LocalFlags().FlagUsages()
	return strings.TrimSuffix(str, "\n")
}
