package cli

import (
	"fmt"
	"reflect"
	"regexp"
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
	// API key arg will replace password and avoid prompt for password
	ApiKeyArg string
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
	apiKeyFound := false
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		kv := strings.SplitN(arg, "=", 2)
		if len(kv) != 2 {
			return dat, fmt.Errorf("arg \"%s\" not name=val format", arg)
		}
		var argVal interface{}
		argKey, argVal := kv[0], kv[1]
		argKey = resolveAlias(argKey, aliases)
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
		delete(required, argKey)
		setKeyVal(dat, argKey, argVal, specialArgType)
		if argKey == s.PasswordArg {
			passwordFound = true
		} else if argKey == s.ApiKeyArg {
			apiKeyFound = true
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

	if s.PasswordArg != "" && s.ApiKeyArg != "" {
		if apiKeyFound && passwordFound {
			return dat, fmt.Errorf("either password or apikey should passed and not both")
		}
	}

	// Do not prompt for password if API key is passed
	if s.ApiKeyArg == "" || !apiKeyFound {
		// prompt for password if not in arg list
		if s.PasswordArg != "" && !passwordFound {
			pw, err := getPassword(s.VerifyPassword)
			if err != nil {
				return dat, err
			}
			setKeyVal(dat, resolveAlias(s.PasswordArg, aliases), pw, "")
		}
	}

	// Fill in obj with values. Also checks for args that
	// don't correspond to any fields in the target object.
	if obj != nil {
		unused, err := WeakDecode(dat, obj, s.DecodeHook)
		if err != nil {
			return dat, ConvertDecodeErr(err, reals)
		}
		if !s.AllowUnused && len(unused) > 0 {
			return dat, fmt.Errorf("invalid args: %s",
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

func ConvertDecodeErr(err error, reals map[string]string) error {
	// converts mapstructure errors into nicer errors for cli output
	wrapped, ok := err.(*mapstructure.Error)
	if !ok {
		return err
	}
	for ii, err := range wrapped.Errors {
		switch e := err.(type) {
		case *mapstructure.ParseError:
			name := getDecodeArg(e.Name, reals)
			suberr := e.Err
			help := ""
			if ne, ok := suberr.(*strconv.NumError); ok {
				suberr = ne.Err
			}
			if e.To == reflect.Bool {
				help = ", valid values are true, false"
			}
			err = fmt.Errorf(`Unable to parse "%s" value "%v" as %v: %v%s`, name, e.Val, e.To, suberr, help)
		case *mapstructure.OverflowError:
			name := getDecodeArg(e.Name, reals)
			err = fmt.Errorf(`Unable to parse "%s" value "%v" as %v: overflow error`, name, e.Val, e.To)
		}
		wrapped.Errors[ii] = err
	}
	return wrapped
}

func getDecodeArg(name string, reals map[string]string) string {
	arg := strings.ToLower(name)
	alias := resolveAlias(arg, reals)
	if alias != arg {
		return alias
	}
	// cli arg does not include parent struct name
	parts := strings.Split(arg, ".")
	if len(parts) > 1 {
		arg = strings.Join(parts[1:], ".")
	}
	return arg
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

func (s FieldNamespace) String() string {
	switch s {
	case StructNamespace:
		return "StructNamespace"
	case YamlNamespace:
		return "YamlNamespace"
	case JsonNamespace:
		return "JsonNamespace"
	}
	return fmt.Sprintf("UnknownNamespace(%d)", s)
}

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
			return fmt.Errorf("Field %s (%s) not found in struct %s", key, inputNS.String(), t.Name())
		}
		tag := sf.Tag.Get("json")
		tagvals := strings.Split(tag, ",")
		jsonName := ""
		if len(tagvals) > 0 {
			jsonName = tagvals[0]
			tagvals = tagvals[1:]
		}
		if jsonName == "" {
			jsonName = sf.Name
		}
		if subargs, ok := val.(map[string]interface{}); ok {
			// sub struct
			kind := sf.Type.Kind()
			if kind == reflect.Ptr {
				kind = sf.Type.Elem().Kind()
			}
			if kind != reflect.Struct {
				return fmt.Errorf("key %s (%s) value %v is a map (struct) but expected %v", key, inputNS.String(), val, sf.Type)
			}
			var subjson map[string]interface{}
			if hasTag("inline", tagvals) {
				subjson = js
			} else {
				subjson = getSubMap(js, jsonName, -1)
			}
			err := MapJsonNamesT(subargs, subjson, sf.Type, inputNS)
			if err != nil {
				return err
			}
		} else if list, ok := val.([]map[string]interface{}); ok {
			// arrayed struct
			if sf.Type.Kind() != reflect.Slice {
				return fmt.Errorf("key %s (%s) value %v is an array but expected %v", key, inputNS.String(), val, sf.Type)
			}
			elemt := sf.Type.Elem()
			jslist := make([]map[string]interface{}, 0, len(list))
			for ii, _ := range list {
				out := make(map[string]interface{})
				err := MapJsonNamesT(list[ii], out, elemt, inputNS)
				if err != nil {
					return err
				}
				jslist = append(jslist, out)
			}
			js[jsonName] = jslist
		} else if reflect.TypeOf(val).Kind() == reflect.Slice {
			// array of built-in types
			if sf.Type.Kind() != reflect.Slice {
				return fmt.Errorf("key %s (%s) value %v is an array but expected type %v", key, inputNS.String(), val, sf.Type)
			}
			js[jsonName] = val
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
					return fmt.Errorf("unmarshal err on %s (%s), %s, %v, %v", key, inputNS.String(), strval, v.Elem().Kind(), err)
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

func FindHierField(t reflect.Type, hierName string, ns FieldNamespace) (reflect.StructField, bool) {
	sf := reflect.StructField{}
	found := false
	for _, name := range strings.Split(hierName, ".") {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() != reflect.Struct {
			return sf, false
		}
		sf, found = FindField(t, name, ns)
		if !found {
			return sf, false
		}
		t = sf.Type
	}
	return sf, found
}

func getSubMap(cur map[string]interface{}, key string, arrIdx int) map[string]interface{} {
	if arrIdx > -1 {
		// arrayed struct
		var arr []map[string]interface{}
		val, ok := cur[key]
		if ok {
			// check that it's the right type
			arr, ok = val.([]map[string]interface{})
		}
		if !ok {
			// didn't exist, or wrong type (so overwrite)
			arr = make([]map[string]interface{}, arrIdx+1)
			cur[key] = arr
			for ii, _ := range arr {
				arr[ii] = make(map[string]interface{})
			}
		}
		// increase length if needed
		if len(arr) <= arrIdx {
			newarr := make([]map[string]interface{}, arrIdx+1)
			copy(newarr, arr)
			for ii := len(arr); ii < len(newarr); ii++ {
				newarr[ii] = make(map[string]interface{})
			}
			arr = newarr
			cur[key] = arr
		}
		return arr[arrIdx]
	}
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

var (
	reArrayNums   = regexp.MustCompile(`:\d+.`)
	reArrayPlaces = regexp.MustCompile(`:#.`)
)

func resolveAlias(name string, lookup map[string]string) string {
	// for sublist arrays, we need to convert a :1 index
	// into the generic :# used by the alias in the mapping.
	// Ex: array:1.name -> array:#.name
	matches := reArrayNums.FindAll([]byte(name), -1)
	if len(matches) == 0 {
		if mapped, ok := lookup[name]; ok {
			return mapped
		}
		return name
	}
	nameGeneric := reArrayNums.ReplaceAll([]byte(name), []byte(`:#.`))

	mappedGeneric, ok := lookup[string(nameGeneric)]
	if !ok {
		return name
	}
	// put back the array index values
	nameReplaced := reArrayPlaces.ReplaceAllFunc([]byte(mappedGeneric), func(repl []byte) []byte {
		if len(matches) == 0 {
			return repl
		}
		ret := matches[0]
		matches = matches[1:]
		return ret
	})
	return string(nameReplaced)
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
			arrIdx := -1
			// if field is repeated (arrayed) struct, it will have a :# suffix.
			colonIdx := strings.LastIndex(part, ":")
			if colonIdx != -1 && len(part) > colonIdx+1 {
				if idx, err := strconv.ParseUint(part[colonIdx+1:], 10, 32); err == nil {
					arrIdx = int(idx)
					part = part[:colonIdx]
				}
			}
			dat = getSubMap(dat, part, arrIdx)
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
// Aliases are of the form alias=hiername.
func MarshalArgs(obj interface{}, ignore []string, aliases []string) ([]string, error) {
	args := []string{}
	if obj == nil {
		return args, nil
	}

	// for updates, passed in data may already by mapped
	dat, ok := obj.(map[string]interface{})
	if !ok {
		// use mobiledgex yaml here since it always omits empty
		byt, err := yaml.Marshal(obj)
		if err != nil {
			return args, err
		}
		dat = make(map[string]interface{})
		err = yaml.Unmarshal(byt, &dat)
		if err != nil {
			return args, err
		}
	}
	// Note if generic map is passed in, it must be the StructNamespace.
	// This is because args are also nominally in the struct namespace.
	// It is difficult to accept JsonNamespace, because JSON collapses
	// embedded structs, while args/yaml/mapstructure do not.

	ignoremap := make(map[string]struct{})
	if ignore != nil {
		for _, str := range ignore {
			ignoremap[str] = struct{}{}
		}
	}
	spargs := GetSpecialArgs(obj)

	aliasm := make(map[string]string)
	for _, alias := range aliases {
		ar := strings.SplitN(alias, "=", 2)
		if len(ar) != 2 {
			continue
		}
		aliasm[ar[1]] = ar[0]
	}

	return MapToArgs([]string{}, dat, ignoremap, spargs, aliasm), nil
}

func MapToArgs(prefix []string, dat map[string]interface{}, ignore map[string]struct{}, specialArgs map[string]string, aliases map[string]string) []string {
	args := []string{}
	for kK, v := range dat {
		if v == nil {
			continue
		}
		k := strings.ToLower(kK)
		if sub, ok := v.(map[string]interface{}); ok {
			subargs := MapToArgs(append(prefix, k), sub, ignore, specialArgs, aliases)
			args = append(args, subargs...)
			continue
		}
		if sub, ok := v.([]interface{}); ok {
			for ii, subv := range sub {
				key := k
				if reflect.ValueOf(subv).Kind() == reflect.Map {
					// repeated struct
					key = fmt.Sprintf("%s:%d", k, ii)
				}
				submap := map[string]interface{}{
					key: subv,
				}
				subargs := MapToArgs(prefix, submap, ignore, specialArgs, aliases)
				args = append(args, subargs...)
			}
			continue
		}
		keys := append(prefix, k)
		name := strings.Join(keys, ".")
		if _, ok := ignore[name]; ok {
			continue
		}
		parentName := strings.Join(prefix, ".")
		sparg, _ := specialArgs[parentName]
		if sparg == "StringToString" {
			name = parentName
		}
		name = resolveAlias(name, aliases)

		val := fmt.Sprintf("%v", v)
		if strings.ContainsAny(val, " \t\r\n") || len(val) == 0 {
			val = strconv.Quote(val)
		}
		var arg string
		if sparg == "StringToString" {
			arg = fmt.Sprintf("%s=%s=%s", name, kK, val)
		} else {
			arg = fmt.Sprintf("%s=%s", name, val)
		}
		args = append(args, arg)
	}
	return args
}

func GetSpecialArgs(obj interface{}) map[string]string {
	m := make(map[string]string)
	getSpecialArgs(m, []string{}, reflect.TypeOf(obj))
	return m
}

func getSpecialArgs(special map[string]string, parents []string, t reflect.Type) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Struct {
		for ii := 0; ii < t.NumField(); ii++ {
			sf := t.Field(ii)
			getSpecialArgs(special, append(parents, sf.Name), sf.Type)
		}
	}
	if len(parents) == 0 {
		// basic type but not in a struct
		return
	}
	sptype := ""
	if t.Kind() == reflect.Map && t == reflect.MapOf(reflect.TypeOf(""), reflect.TypeOf("")) {
		sptype = "StringToString"
	}
	if t.Kind() == reflect.Slice && t == reflect.SliceOf(reflect.TypeOf("")) {
		sptype = "StringArray"
	}
	if sptype != "" {
		key := strings.Join(parents, ".")
		special[strings.ToLower(key)] = sptype
	}
}
