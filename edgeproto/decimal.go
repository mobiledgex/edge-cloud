package edgeproto

import (
	"encoding/json"
	fmt "fmt"
	"strconv"
	strings "strings"
)

// Decimal types allow for handling integer based
// decimal numbers without the inaccuracy associated with floats.

var DecNanos uint32 = 1
var DecMicros uint32 = 1000 * DecNanos
var DecMillis uint32 = 1000 * DecMicros
var DecWhole uint32 = 1000 * DecMillis

// number of decimal places of precision stored in Nanos
var DecPrecision = 9

func NewUdec64(whole uint64, nanos uint32) *Udec64 {
	// account for nanos greater than a whole number
	whole += uint64(nanos / DecWhole)
	nanos = nanos % DecWhole
	return &Udec64{
		Whole: whole,
		Nanos: nanos,
	}
}

// Cmp compares a and b and returns:
// -1 if a < b
//  0 if a == b
//  1 if a > b
func (a *Udec64) Cmp(b *Udec64) int {
	if a.Whole > b.Whole {
		return 1
	} else if a.Whole < b.Whole {
		return -1
	}
	// whole numbers equal
	if a.Nanos > b.Nanos {
		return 1
	} else if a.Nanos < b.Nanos {
		return -1
	}
	return 0
}

func (a *Udec64) GreaterThan(b *Udec64) bool {
	return a.Cmp(b) == 1
}

func (a *Udec64) Equal(b *Udec64) bool {
	return a.Cmp(b) == 0
}

func (a *Udec64) LessThan(b *Udec64) bool {
	return a.Cmp(b) == -1
}

func (a *Udec64) CmpUint64(val uint64) int {
	if a.Whole > val {
		return 1
	} else if a.Whole < val {
		return -1
	}
	// whole numbers equal
	if a.Nanos > 0 {
		return 1
	}
	return 0
}

func (a *Udec64) GreaterThanUint64(val uint64) bool {
	return a.CmpUint64(val) == 1
}

func (a *Udec64) EqualUint64(val uint64) bool {
	return a.CmpUint64(val) == 0
}

func (a *Udec64) LessThanUint64(val uint64) bool {
	return a.CmpUint64(val) == -1
}

func (a *Udec64) IsZero() bool {
	return a.Whole == 0 && a.Nanos == 0
}

func (a *Udec64) Set(whole uint64, nanos uint32) {
	a.Whole = whole
	a.Nanos = nanos
}

// Add b to a
func (a *Udec64) Add(b *Udec64) {
	a.Nanos += b.Nanos
	wholeExtra := uint64(a.Nanos / DecWhole)
	a.Nanos = a.Nanos % DecWhole
	a.Whole += b.Whole + wholeExtra
}

// Sub b from a. Behavior is undefined if b > a.
func (a *Udec64) Sub(b *Udec64) {
	a.Whole -= b.Whole
	if a.Nanos >= b.Nanos {
		a.Nanos = a.Nanos - b.Nanos
	} else {
		a.Nanos = DecWhole + a.Nanos - b.Nanos
		a.Whole -= 1
	}
}

func (a *Udec64) AddUint64(val uint64) {
	a.Whole += val
}

func (a *Udec64) SubUint64(val uint64) {
	a.Whole -= val
}

// Whole returns only the whole number portion of the decimal
func (a *Udec64) Uint64() uint64 {
	return a.Whole
}

func ParseUdec64(str string) (*Udec64, error) {
	vals := strings.Split(str, ".")
	if len(vals) > 2 {
		return nil, fmt.Errorf("Too many decimal points")
	}
	whole, err := strconv.ParseUint(vals[0], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("Parsing whole number %s failed, %v", vals[0], err)
	}
	nanos := uint32(0)
	if len(vals) == 2 {
		// parse decimal portion
		precision := len(vals[1])
		if precision > DecPrecision {
			return nil, fmt.Errorf("Decimal precision cannot be smaller than nanos")
		}
		nanos64, err := strconv.ParseUint(vals[1], 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Parsing decimal number %s failed, %v", vals[1], err)
		}
		nanos = uint32(nanos64)
		for p := 0; p < DecPrecision-precision; p++ {
			nanos *= 10
		}
	}
	return &Udec64{
		Whole: whole,
		Nanos: nanos,
	}, nil
}

func (u *Udec64) DecString() string {
	if u.Nanos == 0 {
		return fmt.Sprintf("%d", u.Whole)
	}
	decStr := fmt.Sprintf("%d", u.Nanos)
	// pad left zeros until nanos precision
	for i := len(decStr); i < DecPrecision; i++ {
		decStr = "0" + decStr
	}
	// drop right zeros
	decStr = strings.TrimRight(decStr, "0")
	return fmt.Sprintf("%d.%s", u.Whole, decStr)
}

func (u *Udec64) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	err := unmarshal(&str)
	if err == nil {
		udec64, err := ParseUdec64(str)
		if err != nil {
			return err
		}
		*u = *udec64
		return nil
	}
	var val uint64
	err = unmarshal(&val)
	if err == nil {
		u.Whole = val
		return nil
	}
	return fmt.Errorf("Invalid unsigned decimal 64 type")
}

func (u Udec64) MarshalYAML() (interface{}, error) {
	if u.Nanos == 0 {
		// if no nanos, marshal as integer instead of string
		return u.Whole, nil
	}
	return u.DecString(), nil
}

func (u *Udec64) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err == nil {
		udec64, err := ParseUdec64(str)
		if err != nil {
			return err
		}
		*u = *udec64
		return nil
	}
	var val uint64
	err = json.Unmarshal(b, &val)
	if err == nil {
		u.Whole = val
		return nil
	}
	return fmt.Errorf("Invalid unsigned decimal 64 type")
}

func (u Udec64) MarshalJSON() ([]byte, error) {
	if u.Nanos == 0 {
		// if no nanos, marshal as integer instead of string
		return json.Marshal(u.Whole)
	}
	return json.Marshal(u.DecString())
}
