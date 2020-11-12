package confinator

import (
	"encoding/binary"
	"flag"
	"fmt"
	"math"
	"reflect"
	"strings"
	"sync"
	"time"
)

type CustomTypeFunc func(fs *flag.FlagSet, varPtr interface{}, name, usage string)

type Confinator struct {
	mu    sync.RWMutex
	types map[string]CustomTypeFunc
	fs    *flag.FlagSet
}

func NewConfinator(fs *flag.FlagSet) *Confinator {
	cf := new(Confinator)
	cf.types = make(map[string]CustomTypeFunc)
	cf.fs = fs
	return cf
}

var customTypes = make(map[string]CustomTypeFunc)

func buildTypeKey(varType reflect.Type) string {
	return fmt.Sprintf("%s.%s", varType.PkgPath(), varType.Name())
}

func RegisterType(varPtr interface{}, fn CustomTypeFunc) {
	varType := reflect.TypeOf(varPtr)
	if varType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Must provided a pointer, saw %T", varPtr))
	}
	customTypes[buildTypeKey(varType)] = fn
}

// FlushRegisteredTypes cleans up the state of this package, indicating that you are done building your config and no
// longer need any state that may be contained within.
//
// This includes nil'ing out any registered custom types.  Any attempt to register a new type will result in a panic.
func FlushRegisteredTypes() {
	customTypes = nil
}

// StringDuration is a quick hack to let us use time.Duration strings as config values in hcl files
type StringDuration time.Duration

func (sd StringDuration) String() string {
	return time.Duration(sd).String()
}

func (sd *StringDuration) Set(v string) error {
	if v == "" {
		*sd = StringDuration(0)
		return nil
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		return err
	}
	*sd = StringDuration(d)
	return nil
}

func (sd StringDuration) MarshalBinary() ([]byte, error) {
	b := make([]byte, 8, 8)
	binary.LittleEndian.PutUint64(b, uint64(sd))
	return b, nil
}

func (sd *StringDuration) UnmarshalBinary(b []byte) error {
	if l := len(b); l != 8 {
		return fmt.Errorf("expected 8 bytes, saw %d", l)
	}
	uv := binary.LittleEndian.Uint64(b)
	if uv > math.MaxInt64 {
		return fmt.Errorf("int64 overflow: %d", uv)
	}
	*sd = StringDuration(uv)
	return nil
}

func (sd StringDuration) GobEncode() ([]byte, error) {
	return sd.MarshalBinary()
}

func (sd *StringDuration) GobDecode(b []byte) error {
	return sd.UnmarshalBinary(b)
}

func (sd StringDuration) MarshalText() ([]byte, error) {
	return []byte(sd.String()), nil
}

func (sd *StringDuration) UnmarshalText(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	return sd.Set(string(b))
}

func (sd StringDuration) MarshalJSON() ([]byte, error) {
	return []byte("\"" + sd.String() + "\""), nil
}

func (sd *StringDuration) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		return nil
	}
	clean := strings.Trim(string(b), "\"")
	if clean == "" {
		*sd = StringDuration(0)
		return nil
	}
	return sd.Set(clean)
}

func (sd StringDuration) Duration() time.Duration {
	return time.Duration(sd)
}

// stringMapValue is pulled from https://github.com/hashicorp/consul/blob/b5abf61963c7b0bdb674602bfb64051f8e23ddb1/agent/config/flagset.go#L118
type stringMapValue map[string]string

func newStringMapValue(p *map[string]string) *stringMapValue {
	*p = map[string]string{}
	return (*stringMapValue)(p)
}

func (s *stringMapValue) Set(val string) error {
	p := strings.SplitN(val, ":", 2)
	k, v := p[0], ""
	if len(p) == 2 {
		v = p[1]
	}
	(*s)[k] = v
	return nil
}

func (s *stringMapValue) Get() interface{} {
	return s
}

func (s *stringMapValue) String() string {
	var x []string
	for k, v := range *s {
		if v == "" {
			x = append(x, k)
		} else {
			x = append(x, k+":"+v)
		}
	}
	return strings.Join(x, " ")
}

// stringSliceValue is pulled from https://github.com/hashicorp/consul/blob/b5abf61963c7b0bdb674602bfb64051f8e23ddb1/agent/config/flagset.go#L183
type stringSliceValue []string

func newStringSliceValue(p *[]string) *stringSliceValue {
	return (*stringSliceValue)(p)
}

func (s *stringSliceValue) Set(val string) error {
	*s = append(*s, val)
	return nil
}

func (s *stringSliceValue) Get() interface{} {
	return s
}

func (s *stringSliceValue) String() string {
	return strings.Join(*s, " ")
}

// FlagVar is a convenience method that handles a few common config struct -> flag cases
func FlagVar(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
	switch x := varPtr.(type) {
	case *bool:
		var v bool
		if varPtr.(*bool) != nil {
			v = *varPtr.(*bool)
		}
		fs.BoolVar(varPtr.(*bool), name, v, usage)
	case *time.Duration:
		var v time.Duration
		if varPtr.(*time.Duration) != nil {
			v = *varPtr.(*time.Duration)
		}
		fs.DurationVar(varPtr.(*time.Duration), name, v, usage)
	case *StringDuration:
		fs.Var(varPtr.(*StringDuration), name, usage)
	case *int:
		var v int
		if varPtr.(*int) != nil {
			v = *varPtr.(*int)
		}
		fs.IntVar(varPtr.(*int), name, v, usage)
	case *uint:
		var v uint
		if varPtr.(*uint) != nil {
			v = *varPtr.(*uint)
		}
		fs.UintVar(varPtr.(*uint), name, v, usage)
	case *string:
		var v string
		if varPtr.(*string) != nil {
			v = *varPtr.(*string)
		}
		fs.StringVar(varPtr.(*string), name, v, usage)
	case *[]string:
		fs.Var(newStringSliceValue(x), name, usage)
	case *map[string]string:
		fs.Var(newStringMapValue(x), name, usage)
	default:
		if fn, ok := customTypes[buildTypeKey(reflect.TypeOf(varPtr))]; ok {
			fn(fs, varPtr, name, usage)
		} else {
			panic(fmt.Sprintf("invalid type: %T", varPtr))
		}
	}
}

type HelpTextState struct {
	FlagSet        *flag.FlagSet
	LongestName    int
	LongestDefault int
	Current        string
}

type FlagHelpHeaderFunc func(state HelpTextState) string

var DefaultFlagHelpHeaderFunc FlagHelpHeaderFunc = func(state HelpTextState) string {
	return fmt.Sprintf("%s\n", state.FlagSet.Name())
}

type FlagHelpTableHeaderFunc func(state HelpTextState) string

var DefaultFlagHelpTableHeaderFunc FlagHelpTableHeaderFunc = func(state HelpTextState) string {
	return fmt.Sprintf(
		"\t[Flag]%s[Default]%s[Usage]",
		strings.Repeat(" ", state.LongestName-1),
		strings.Repeat(" ", state.LongestDefault-5),
	)
}

type FlagHelpTableRowFunc func(flagNum int, f *flag.Flag, state HelpTextState) string

var DefaultFlagHelpTableRowFunc FlagHelpTableRowFunc = func(flagNum int, f *flag.Flag, state HelpTextState) string {
	return fmt.Sprintf(
		"\n\t-%s%s%s%s%s",
		f.Name,
		strings.Repeat(" ", state.LongestName-len(f.Name)+4),
		f.DefValue,
		strings.Repeat(" ", state.LongestDefault-len(f.DefValue)+4),
		f.Usage,
	)
}

type FlagHelpTableFooterFunc func(state HelpTextState) string

var DefaultFlagHelpTableFooterFunc FlagHelpTableFooterFunc = func(state HelpTextState) string {
	return ""
}

type FlagHelpFooterFunc func(state HelpTextState) string

var DefaultFlagHelpFooterFunc FlagHelpFooterFunc = func(state HelpTextState) string {
	return ""
}

type FlagHelpTextConf struct {
	FlagSet         *flag.FlagSet
	HeaderFunc      FlagHelpHeaderFunc
	TableHeaderFunc FlagHelpTableHeaderFunc
	TableRowFunc    FlagHelpTableRowFunc
	TableFooterFunc FlagHelpTableFooterFunc
	FooterFunc      FlagHelpFooterFunc
}

func FlagHelpText(conf FlagHelpTextConf) string {
	var (
		longestName    int
		longestDefault int
		out            string

		hf  FlagHelpHeaderFunc
		thf FlagHelpTableHeaderFunc
		trf FlagHelpTableRowFunc
		tff FlagHelpTableFooterFunc
		ff  FlagHelpFooterFunc

		fs = conf.FlagSet
	)

	if conf.HeaderFunc == nil {
		hf = DefaultFlagHelpHeaderFunc
	} else {
		hf = conf.HeaderFunc
	}
	if conf.TableHeaderFunc == nil {
		thf = DefaultFlagHelpTableHeaderFunc
	} else {
		thf = conf.TableHeaderFunc
	}
	if conf.TableRowFunc == nil {
		trf = DefaultFlagHelpTableRowFunc
	} else {
		trf = conf.TableRowFunc
	}
	if conf.TableFooterFunc == nil {
		tff = DefaultFlagHelpTableFooterFunc
	} else {
		tff = conf.TableFooterFunc
	}
	if conf.FooterFunc == nil {
		ff = DefaultFlagHelpFooterFunc
	} else {
		ff = conf.FooterFunc
	}

	fs.VisitAll(func(f *flag.Flag) {
		if l := len(f.Name); l > longestName {
			longestName = l
		}
		if l := len(f.DefValue); l > longestDefault {
			longestDefault = l
		}
	})

	makeState := func() HelpTextState {
		return HelpTextState{
			FlagSet:        fs,
			LongestName:    longestName,
			LongestDefault: longestDefault,
			Current:        out,
		}
	}

	out = hf(makeState())
	out = fmt.Sprintf("%s%s", out, thf(makeState()))
	i := 0
	fs.VisitAll(func(f *flag.Flag) {
		out = fmt.Sprintf("%s%s", out, trf(i, f, makeState()))
		i++
	})
	out = fmt.Sprintf("%s%s%s", out, tff(makeState()), ff(makeState()))

	return out
}
