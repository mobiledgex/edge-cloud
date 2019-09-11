package gencmd

import (
	"fmt"
	"strings"
)

type StringToString map[string]string

func (i *StringToString) String() string {
	outArr := []string{}
	for k, v := range *i {
		outArr = append(outArr, fmt.Sprintf("%s='%v'", k, v))
	}
	return "[" + strings.Join(outArr, ",") + "]"
}

func (i *StringToString) Type() string {
	return "string=string"
}

func (i *StringToString) Set(value string) error {
	kv := strings.SplitN(strings.Trim(value, `"`), "=", 2)
	if len(kv) != 2 {
		return fmt.Errorf("%s must be formatted as key=value", value)
	}
	if len(*i) == 0 {
		*i = make(map[string]string)
	}
	(*i)[kv[0]] = kv[1]
	return nil
}
