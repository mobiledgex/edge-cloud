package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/mitchellh/mapstructure"
	edgeproto "github.com/mobiledgex/edge-cloud/edgeproto"
	"github.com/mobiledgex/edge-cloud/util"
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
func (s *Input) ParseArgs(args []string, obj interface{}) (*MapData, error) {
	dat := map[string]interface{}{}

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
			return nil, fmt.Errorf("arg \"%s\" not name=val format", arg)
		}
		argKey, argVal := kv[0], kv[1]
		argKey = resolveAlias(argKey, aliases)
		specialArgType := ""
		if s.SpecialArgs != nil {
			if argType, found := (*s.SpecialArgs)[argKey]; found {
				specialArgType = argType
			}
		}
		delete(required, argKey)
		err := s.setKeyVal(dat, obj, argKey, argVal, specialArgType)
		if err != nil {
			return nil, fmt.Errorf("parsing arg \"%s\" failed: %v", arg, err)
		}
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
		return nil, fmt.Errorf("missing required args: %s", strings.Join(missing, " "))
	}

	if s.PasswordArg != "" && s.ApiKeyArg != "" {
		if apiKeyFound && passwordFound {
			return nil, fmt.Errorf("either password or apikey should passed and not both")
		}
	}

	// Do not prompt for password if API key is passed
	if s.ApiKeyArg == "" || !apiKeyFound {
		// prompt for password if not in arg list
		if s.PasswordArg != "" && !passwordFound {
			pw, err := getPassword(s.VerifyPassword)
			if err != nil {
				return nil, err
			}
			s.setKeyVal(dat, obj, resolveAlias(s.PasswordArg, aliases), pw, "")
		}
	}

	// Fill in obj with values. Also checks for args that
	// don't correspond to any fields in the target object.
	if obj != nil {
		unused, err := WeakDecode(dat, obj, s.DecodeHook)
		if err != nil {
			return nil, ConvertDecodeErr(err, reals)
		}
		if !s.AllowUnused && len(unused) > 0 {
			return nil, fmt.Errorf("invalid args: %s",
				strings.Join(unused, " "))
		}
	}
	data := MapData{
		Namespace: ArgsNamespace,
		Data:      dat,
	}
	return &data, nil
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
			if ne, ok := suberr.(*strconv.NumError); ok {
				suberr = ne.Err
			}
			help := GetParseHelpKind(e.To)
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

// FieldNamespace describes the format of field names in generic
// map[string]interface{} data, whether they correspond to the go struct's
// field names, yaml tag names, json tag names, or arg names.
// Note that arg names are just lower-cased field names.
// It does not specify what format the values are in. Values may be
// type-specific (if extracted from an object with type info), or may
// be generic values (if unmarshaling yaml/json into map[string]interface{}
// where type info is not present).
type FieldNamespace int

const (
	StructNamespace FieldNamespace = iota
	YamlNamespace
	JsonNamespace
	ArgsNamespace
)

// MapData associates generic imported mapped data with the
// namespace that the keys are in.
type MapData struct {
	Namespace FieldNamespace
	Data      map[string]interface{}
}

func (s FieldNamespace) String() string {
	switch s {
	case StructNamespace:
		return "StructNamespace"
	case YamlNamespace:
		return "YamlNamespace"
	case JsonNamespace:
		return "JsonNamespace"
	case ArgsNamespace:
		return "ArgsNamespace"
	}
	return fmt.Sprintf("UnknownNamespace(%d)", s)
}

// This converts the key namespace from whatever is specified into JSON
// key names, based on json tags on the object. The values are not changed.
// It will squash hierarchy if the "inline" tag is found.
func JsonMap(mdat *MapData, obj interface{}) (*MapData, error) {
	if mdat.Namespace == JsonNamespace {
		// already json
		return mdat, nil
	}

	js := map[string]interface{}{}
	err := MapJsonNamesT(mdat.Data, js, reflect.TypeOf(obj), mdat.Namespace)
	if err != nil {
		return nil, err
	}
	jsDat := MapData{
		Namespace: JsonNamespace,
		Data:      js,
	}
	return &jsDat, nil
}

func MapJsonNamesT(dat, js map[string]interface{}, t reflect.Type, inputNS FieldNamespace) error {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for key, val := range dat {
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
			if len(subargs) == 0 {
				// empty map/array for update
				js[jsonName] = reflect.New(sf.Type)
				continue
			}
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
		} else if list, ok := arrayOfMaps(val); ok {
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
		} else {
			// no sub-structure, so just assign value.
			js[jsonName] = val
		}
	}
	return nil
}

func FindField(t reflect.Type, name string, ns FieldNamespace) (reflect.StructField, bool) {
	if ns == ArgsNamespace {
		// Same as StructNamespace, but lower cased
		return t.FieldByNameFunc(func(n string) bool {
			return strings.ToLower(n) == strings.ToLower(name)
		})
	} else if ns == StructNamespace {
		return t.FieldByName(name)
	} else {
		// Both JSON and YAML are case-sensitive to field names,
		// but marshaled yaml is always lower-cased, and JSON
		// unmarshal prefers case-sensitive but also matches case
		// insensitive. So here we prefer case-sensitive matching
		// but we also allow case-insensitive matching.
		var sfCaseIns reflect.StructField
		for ii := 0; ii < t.NumField(); ii++ {
			sf := t.Field(ii)

			tagName := GetFieldTaggedName(sf, ns)
			if tagName == name {
				return sf, true
			} else if strings.ToLower(tagName) == strings.ToLower(name) {
				sfCaseIns = sf
			}
		}
		if sfCaseIns.Name != "" {
			// return case insensitive match
			return sfCaseIns, true
		}
		return reflect.StructField{}, false
	}
}

// Get the field name based on the namespace
func GetFieldTaggedName(sf reflect.StructField, ns FieldNamespace) string {
	tagType := ""
	switch ns {
	case JsonNamespace:
		tagType = "json"
	case YamlNamespace:
		tagType = "yaml"
	case ArgsNamespace:
		return strings.ToLower(sf.Name)
	default:
		// default struct namespace
		return sf.Name
	}
	tag := sf.Tag.Get(tagType)
	tagvals := strings.Split(tag, ",")
	if len(tagvals) > 0 && tagvals[0] != "" {
		return tagvals[0]
	}
	return sf.Name
}

func FindHierField(t reflect.Type, hierName string, ns FieldNamespace) (reflect.StructField, bool) {
	sf := reflect.StructField{}
	if t == nil {
		return sf, false
	}
	found := false
	for _, name := range strings.Split(hierName, ".") {
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
		if t.Kind() == reflect.Map || t.Kind() == reflect.Slice {
			t = t.Elem()
			if t.Kind() == reflect.Ptr {
				t = t.Elem()
			}
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
	// for sublists/maps, we may need to remove any :empty directive
	emptySuffix := ":" + util.EmptySet
	if strings.HasSuffix(name, emptySuffix) {
		// for lists of objects, the alias will include :empty
		if mapped, ok := lookup[name]; ok {
			return mapped
		}
		// for lists of strings, the alias will not include :empty
		trimmed := strings.TrimSuffix(name, emptySuffix)
		if mapped, ok := lookup[trimmed]; ok {
			return mapped + emptySuffix
		}
		return name
	}
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

// setKeyVal is used to build a generic map[string]interface{} set of data
// from command line arguments. The key namespace is ArgsNamespace,
// and the values are type-specific based on the object's field types.
func (s *Input) setKeyVal(dat map[string]interface{}, obj interface{}, key, val, argType string) error {
	// Lookup object field corresponding to hierarchical key name.
	// key may include an array index suffix, ignore that for field lookup.
	// key may also end in :empty.
	lookupKey := strings.TrimSuffix(key, ":"+util.EmptySet)
	lookupKey = string(reArrayNums.ReplaceAll([]byte(lookupKey), []byte(".")))
	sf, sfok := FindHierField(reflect.TypeOf(obj), lookupKey, ArgsNamespace)
	if !sfok {
		return fmt.Errorf("invalid argument: key \"%s\" not found", lookupKey)
	}

	// values passed in on the command line that
	// have spaces will be quoted, remove those now.
	valnew, err := strconv.Unquote(val)
	if err == nil {
		val = valnew
	}

	parts := strings.Split(key, ".")
	for ii, part := range parts {
		if ii == len(parts)-1 {
			// special case to specify empty list/map for update
			colonParts := strings.Split(part, ":")
			if len(colonParts) == 2 && colonParts[1] == util.EmptySet {
				b, err := strconv.ParseBool(val)
				if err != nil {
					help := GetParseHelpKind(reflect.Bool)
					return fmt.Errorf("unable to parse %q as bool%s", val, help)
				}
				if !b {
					continue
				}
				// get the hier name without the :empty
				strs := strings.Split(key, ":")
				emptyVal, ok := getFieldEmptyListMap(strs[0], sf.Type, ArgsNamespace)
				if ok {
					dat[colonParts[0]] = emptyVal
				} else {
					dat[colonParts[0]] = nil
				}
				continue
			}
			if argType == "StringArray" {
				if _, ok := dat[part]; !ok {
					dat[part] = make([]string, 0)
				}
				strarr := dat[part].([]string)
				dat[part] = append(strarr, val)
				continue
			}
			if argType == "StringToString" {
				valSlice := strings.SplitN(val, "=", 2)
				if len(valSlice) != 2 {
					return fmt.Errorf("value \"%s\" must be formatted as key=value", val)
				}
				// if value has spaces, may be quoted
				if str, err := strconv.Unquote(valSlice[1]); err == nil {
					valSlice[1] = str
				}
				if _, ok := dat[part]; !ok {
					dat[part] = make(map[string]string)
				}
				mapVal := dat[part].(map[string]string)
				mapVal[valSlice[0]] = valSlice[1]
				dat[part] = mapVal
				continue
			}
			// convert command line string val to type-specific value
			v := reflect.New(sf.Type)
			_, err := WeakDecode(val, v.Interface(), s.DecodeHook)
			if err != nil {
				asType, help, surpressErr := GetParseHelp(sf.Type)
				switch e := err.(type) {
				case *mapstructure.ParseError:
					err = e.Err
					if ne, ok := err.(*strconv.NumError); ok {
						err = ne.Err
					}
					asType = e.To.String()
				case *mapstructure.OverflowError:
					err = fmt.Errorf("overflow error")
					asType = e.To.String()
				default:
					// possible decode hook error with
					// extra junk we should remove.
					replace := "error decoding '': "
					if strings.Contains(err.Error(), replace) {
						err = errors.New(strings.Replace(err.Error(), replace, "", -1))
					}
					if surpressErr {
						err = fmt.Errorf("invalid format")
					}
				}
				return fmt.Errorf("unable to parse %q as %s: %v%s", val, asType, err, help)
			}
			// elem to dereference it
			dat[part] = v.Elem().Interface()
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
	return nil
}

// GetParseHelp gets end-user specific messages for error messages
// without any golang-specific language.
// It returns a non-golang specific type name, a help message with valid
// values, and a bool that represents whether the original error
// should be suppressed because it is too golang-specific.
func GetParseHelp(t reflect.Type) (string, string, bool) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Specially handled types are typically types with
	// custom unmarshalers and custom DecodeHooks.
	if t == reflect.TypeOf(time.Time{}) {
		return "time", fmt.Sprintf(", valid values are RFC3339 format, i.e. %q", time.RFC3339), true
	}
	if t == reflect.TypeOf(time.Duration(0)) || t == reflect.TypeOf(edgeproto.Duration(0)) {
		return "duration", fmt.Sprintf(", valid values are 300ms, 1s, 1.5h, 2h45m, etc"), true
	}
	return t.Kind().String(), GetParseHelpKind(t.Kind()), false
}

func GetParseHelpKind(k reflect.Kind) string {
	if k == reflect.Bool {
		return ", valid values are true, false"
	}
	return ""
}

// Get empty list or map value based on field type.
func getFieldEmptyListMap(hierName string, fieldType reflect.Type, ns FieldNamespace) (interface{}, bool) {
	// Note: we return generic slices and maps rather than Type
	// specific interfaces, in order to match YAML/JSON behavior,
	// which do not have type info (so in yaml, an empty map is {}
	// and an empty slice is []), so if you unmarshal those into
	// a generic map[string]interface, those become
	// map[string]interface{} and []interface{}, respectively.
	switch fieldType.Kind() {
	case reflect.Slice:
		return []interface{}{}, true
	case reflect.Map:
		return map[string]interface{}{}, true
	}
	return nil, false
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
	dat, ok := obj.(*MapData)
	if !ok {
		var err error
		dat, err = GetStructMap(obj, WithStructMapOmitEmpty())
		if err != nil {
			return args, err
		}
	}
	// If MapData passed in, it must be in either the StructNamespace
	// or ArgsNamespace, because MapToArgs does not do any key name
	// translation besides lower-casing it.
	if dat.Namespace != StructNamespace && dat.Namespace != ArgsNamespace {
		return nil, fmt.Errorf("Passed in MapData must be in the struct or args namespace, but is %s", dat.Namespace.String())
	}

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

	return MapToArgs([]string{}, dat.Data, ignoremap, spargs, aliasm), nil
}

func MapToArgs(prefix []string, dat map[string]interface{}, ignore map[string]struct{}, specialArgs map[string]string, aliases map[string]string) []string {
	args := []string{}
	for kK, v := range dat {
		if v == nil {
			continue
		}
		k := strings.ToLower(kK)
		emptySet := false

		value := reflect.ValueOf(v)
		if value.Kind() == reflect.Ptr {
			value = value.Elem()
		}
		// special case: check if the value is an empty list/map
		if value.Kind() == reflect.Slice || value.Kind() == reflect.Map {
			if value.Len() == 0 {
				emptySet = true
			}
		}

		if sub, ok := v.(map[string]interface{}); ok && len(sub) > 0 {
			subargs := MapToArgs(append(prefix, k), sub, ignore, specialArgs, aliases)
			args = append(args, subargs...)
			continue
		}
		// may be []interface{} or []map[string]interface{}
		if value.Kind() == reflect.Slice && value.Len() > 0 {
			for ii := 0; ii < value.Len(); ii++ {
				subv := value.Index(ii).Interface()
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

		// use json.Marshal in case any custom marshalers have been defined
		val := ""
		valbytes, err := json.Marshal(v)
		if err == nil {
			val = string(valbytes)
			// yaml adds quotes for strings that we want to avoid
			if str, err := strconv.Unquote(val); err == nil {
				val = str
			}
		} else {
			val = fmt.Sprintf("%v", v)
		}
		if strings.ContainsAny(val, " \t\r\n") || len(val) == 0 {
			val = strconv.Quote(val)
		}
		var arg string
		if emptySet {
			arg = fmt.Sprintf("%s:%s=true", name, util.EmptySet)
		} else if sparg == "StringToString" {
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
