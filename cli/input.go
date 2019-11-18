package cli

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"syscall"

	"github.com/mitchellh/mapstructure"
	yaml "github.com/mobiledgex/yaml/v2"
	"golang.org/x/crypto/ssh/terminal"
)

// CLI ParseArgs and UnmarshalArgs map the arg name to the lower-case
// version of the field name. This is the default behavior of JSON and
// YAML, but may be overridden by JSON/YAML tags. Args however do not
// honor those tags, so we need to be careful when supplying raw map
// data to something that wants to unmarshal using JSON/YAML tags.

type Input struct {
	// Required argument names
	RequiredArgs []string
	// Alias argument names, format is alias=real
	AliasArgs []string
	// Special argument names, format is arg=argType
	SpecialArgs *map[string]string
	// Password arg will prompt for password if not in args list
	PasswordArg string
	// Verify password if prompting
	VerifyPassword bool
	// Mapstructure DecodeHook functions
	DecodeHook mapstructure.DecodeHookFunc
	// Allow extra args that were not mapped to target object.
	AllowUnused bool
}

// Args are format name=val, where name could be a hierarchical name
// separated by ., i.e. appdata.key.name.
// Arg names should be all lowercase, matching the struct field names.
// This returns a generic map of values set by the args, again
// based on lower case field names, ignoring any json/yaml tags.
// It also fills in obj if specified.
// NOTE: arrays and maps not supported yet.
func (s *Input) ParseArgs(args []string, obj interface{}) (map[string]interface{}, error) {
	dat := make(map[string]interface{})

	// resolve aliases first
	aliases := make(map[string]string)
	reals := make(map[string]string)
	if s.AliasArgs != nil {
		for _, alias := range s.AliasArgs {
			ar := strings.SplitN(alias, "=", 2)
			if len(ar) != 2 {
				fmt.Printf("skipping invalid alias %s\n", alias)
				continue
			}
			aliases[ar[0]] = ar[1]
			reals[ar[1]] = ar[0]
		}
	}
	required := make(map[string]struct{})
	if s.RequiredArgs != nil {
		for _, req := range s.RequiredArgs {
			req = resolveAlias(req, aliases)
			required[req] = struct{}{}
		}
	}

	// create generic data map from args
	passwordFound := false
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			return dat, fmt.Errorf("arg \"%s\" not name=val format", arg)
		}
		var argVal interface{}
		argKey, argVal := kv[0], kv[1]
		specialArgType := ""
		if s.SpecialArgs != nil {
			if argType, found := (*s.SpecialArgs)[argKey]; found {
				specialArgType = argType
				if argType == "StringToString" {
					pair := argVal.(string)
					kv := strings.SplitN(pair, "=", 2)
					if len(kv) != 2 {
						return dat, fmt.Errorf("value \"%s\" of arg \"%s\" must be formatted as key=value", pair, arg)
					}
					argVal = kv
				}
			}
		}
		key := resolveAlias(argKey, aliases)
		delete(required, key)
		setKeyVal(dat, key, argVal, specialArgType)
		if key == s.PasswordArg {
			passwordFound = true
		}
	}

	// ensure required args are present
	if len(required) != 0 {
		missing := []string{}
		for k, _ := range required {
			k = resolveAlias(k, reals)
			missing = append(missing, k)
		}
		return dat, fmt.Errorf("missing required args: %s", strings.Join(missing, " "))
	}

	// prompt for password if not in arg list
	if s.PasswordArg != "" && !passwordFound {
		pw, err := getPassword(s.VerifyPassword)
		if err != nil {
			return dat, err
		}
		setKeyVal(dat, resolveAlias(s.PasswordArg, aliases), pw, "")
	}

	// Fill in obj with values. Also checks for args that
	// don't correspond to any fields in the target object.
	if obj != nil {
		unused, err := WeakDecode(dat, obj, s.DecodeHook)
		if err != nil {
			return dat, err
		}
		if !s.AllowUnused && len(unused) > 0 {
			return dat, fmt.Errorf("Invalid args: %s",
				strings.Join(unused, " "))
		}
	}
	return dat, nil
}

// Use mapstructure to convert an args map (map[string]interface{})
// to fill in an object in output.
func WeakDecode(input, output interface{}, hook mapstructure.DecodeHookFunc) ([]string, error) {
	// use mapstructure.ComposeDecodeHookFunc if we need multiple
	// decode hook functions.
	config := &mapstructure.DecoderConfig{
		Result:           output,
		WeaklyTypedInput: true,
		DecodeHook:       hook,
		Metadata:         &mapstructure.Metadata{},
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return []string{}, err
	}
	err = decoder.Decode(input)
	return config.Metadata.Unused, err
}

// FieldNamespace describes the format of field names used in generic
// map[string]interface{} data, whether they correspond to the go struct's
// field names, yaml tag names, or json tag names.
type FieldNamespace int

const (
	StructNamespace FieldNamespace = iota
	YamlNamespace
	JsonNamespace
)

// JsonMap takes as input the generic args map from ParseArgs
// corresponding to obj, and uses the json tags in obj to generate
// a map with field names in the JSON namespace.
func JsonMap(args map[string]interface{}, obj interface{}, inputNS FieldNamespace) (map[string]interface{}, error) {
	if inputNS == JsonNamespace {
		// already json
		return args, nil
	}
	js := make(map[string]interface{})
	err := MapJsonNamesT(args, js, reflect.TypeOf(obj), inputNS)
	if err != nil {
		return nil, err
	}
	return js, nil
}

func MapJsonNamesT(args, js map[string]interface{}, t reflect.Type, inputNS FieldNamespace) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for key, val := range args {
		// get the StructField to get the json tag
		sf, ok := FindField(t, key, inputNS)
		if !ok {
			continue
		}
		tag := sf.Tag.Get("json")
		tagvals := strings.Split(tag, ",")
		jsonName := ""
		if len(tagvals) > 0 {
			jsonName = tagvals[0]
			tagvals = tagvals[1:]
		}
		if jsonName == "" {
			jsonName = strings.ToLower(key)
		}
		if subargs, ok := val.(map[string]interface{}); ok {
			// sub struct
			var subjson map[string]interface{}
			if hasTag("inline", tagvals) {
				subjson = js
			} else {
				subjson = getSubMap(js, jsonName)
			}
			err := MapJsonNamesT(subargs, subjson, sf.Type, inputNS)
			if err != nil {
				return err
			}
		} else if list, ok := val.([]interface{}); ok {
			if sf.Type.Kind() != reflect.Slice {
				return fmt.Errorf("key %s value is an array but type %v is not", key, sf.Type)
			}
			elemt := sf.Type.Elem()
			jslist := make([]interface{}, 0, len(list))
			for ii, _ := range list {
				if item, ok := list[ii].(map[string]interface{}); ok {
					// struct in list
					out := make(map[string]interface{})
					err := MapJsonNamesT(item, out, elemt, inputNS)
					if err != nil {
						return err
					}
					jslist = append(jslist, out)
				} else {
					jslist = append(jslist, list[ii])
				}
			}
			js[jsonName] = jslist
		} else {
			if sf.Type.Kind() == reflect.Map {
				// must be map of basic built-in types
				js[jsonName] = val
			} else {
				// allocate an object of type (gives us a pointer to it)
				v := reflect.New(sf.Type)
				// let yaml deal with converting the string to the
				// field's type. The only special case is string types
				// may need quotes around string values in case there
				// are special characters in the string.
				strval := fmt.Sprintf("%v", val)
				if v.Elem().Kind() == reflect.String {
					strval = strconv.Quote(strval)
				}
				err := yaml.Unmarshal([]byte(strval), v.Interface())
				if err != nil {
					return fmt.Errorf("unmarshal err on %s, %s, %v, %v", key, strval, v.Elem().Kind(), err)
				}
				// elem to dereference it
				js[jsonName] = v.Elem().Interface()
			}
		}
	}
	return nil
}

func FindField(t reflect.Type, name string, ns FieldNamespace) (reflect.StructField, bool) {
	if ns == StructNamespace {
		return t.FieldByNameFunc(func(n string) bool {
			return strings.ToLower(n) == strings.ToLower(name)
		})
	} else {
		for ii := 0; ii < t.NumField(); ii++ {
			sf := t.Field(ii)

			var tag string
			if ns == JsonNamespace {
				tag = sf.Tag.Get("json")
			} else if ns == YamlNamespace {
				tag = sf.Tag.Get("yaml")
			}
			tagvals := strings.Split(tag, ",")
			tagName := ""
			if len(tagvals) > 0 {
				tagName = tagvals[0]
			}
			if tagName == "" {
				tagName = strings.ToLower(sf.Name)
			}
			if tagName == name {
				return sf, true
			}
		}
		return reflect.StructField{}, false
	}
}

func getSubMap(cur map[string]interface{}, key string) map[string]interface{} {
	var sub map[string]interface{}
	val, ok := cur[key]
	if !ok {
		// create new one
		sub = make(map[string]interface{})
		cur[key] = sub
		return sub
	}
	// check that it's the right type
	sub, ok = val.(map[string]interface{})
	if !ok {
		// conflict, overwrite
		sub = make(map[string]interface{})
		cur[key] = sub
	}
	return sub
}

func hasTag(tag string, tags []string) bool {
	for _, t := range tags {
		if t == tag {
			return true
		}
	}
	return false
}

func resolveAlias(name string, aliases map[string]string) string {
	if real, ok := aliases[name]; ok {
		return real
	}
	return name
}

func setKeyVal(dat map[string]interface{}, key string, val interface{}, argType string) {
	parts := strings.Split(key, ".")
	for ii, part := range parts {
		if ii == len(parts)-1 {
			// values passed in on the command line that
			// have spaces will be quoted.
			if readVal, ok := val.(string); ok {
				valnew, err := strconv.Unquote(readVal)
				if err != nil {
					valnew = readVal
				}
				if argType == "StringArray" {
					if _, ok := dat[part]; !ok {
						dat[part] = make([]string, 0)
					}
					strarr := dat[part].([]string)
					dat[part] = append(strarr, valnew)
				} else {
					dat[part] = valnew
				}
			} else {
				if argType == "StringToString" {
					if _, ok := dat[part]; !ok {
						dat[part] = make(map[string]string)
					}
					valSlice := val.([]string)
					mapVal := dat[part].(map[string]string)
					mapVal[valSlice[0]] = valSlice[1]
					dat[part] = mapVal
				} else {
					dat[part] = val
				}
			}
		} else {
			dat = getSubMap(dat, part)
		}
	}
}

func getPassword(verify bool) (string, error) {
	fmt.Printf("password: ")
	pw, err := terminal.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	fmt.Println()
	if verify {
		fmt.Print("verify password: ")
		pw2, err := terminal.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return "", err
		}
		fmt.Println()
		if string(pw) != string(pw2) {
			return "", fmt.Errorf("passwords don't match")
		}
	}
	return string(pw), nil
}

func replaceMapVals(src map[string]interface{}, dst map[string]interface{}) {
	for key, dstVal := range dst {
		srcVal, found := src[key]
		if !found {
			continue
		}
		subSrc, ok := srcVal.(map[string]interface{})
		subDst, ok2 := dstVal.(map[string]interface{})
		if ok && ok2 {
			replaceMapVals(subSrc, subDst)
			continue
		}
		//fmt.Printf("replace %s %#v with %#v\n", key, dst[key], src[key])
		dst[key] = src[key]
	}
}

// MarshalArgs generates a name=val arg list from the object.
// Arg names that should be ignore can be specified. Names are the
// same format as arg names, lowercase of field names, joined by '.'
func MarshalArgs(obj interface{}, ignore []string) ([]string, error) {
	args := []string{}
	if obj == nil {
		return args, nil
	}

	// use mobiledgex yaml here since it always omits empty
	byt, err := yaml.Marshal(obj)
	if err != nil {
		return args, err
	}
	dat := make(map[string]interface{})
	err = yaml.Unmarshal(byt, &dat)

	ignoremap := make(map[string]struct{})
	if ignore != nil {
		for _, str := range ignore {
			ignoremap[str] = struct{}{}
		}
	}

	return MapToArgs([]string{}, dat, ignoremap), nil
}

func MapToArgs(prefix []string, dat map[string]interface{}, ignore map[string]struct{}) []string {
	args := []string{}
	for k, v := range dat {
		if v == nil {
			continue
		}
		if sub, ok := v.(map[string]interface{}); ok {
			subargs := MapToArgs(append(prefix, k), sub, ignore)
			args = append(args, subargs...)
			continue
		}
		keys := append(prefix, k)
		if _, ok := ignore[strings.Join(keys, ".")]; ok {
			continue
		}
		val := fmt.Sprintf("%v", v)
		if strings.ContainsAny(val, " \t\r\n") {
			val = strconv.Quote(val)
		}
		arg := fmt.Sprintf("%s=%s", strings.Join(keys, "."), val)
		args = append(args, arg)
	}
	return args
}
