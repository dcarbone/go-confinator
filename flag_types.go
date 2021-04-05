package confinator

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func buildFlagVarTypeKey(varType reflect.Type) string {
	const kfmt = "%s%s%s"
	return fmt.Sprintf(kfmt, varType.PkgPath(), varType.Name(), varType.String())
}

type FlagVarTypeHandlerFunc func(fs *flag.FlagSet, varPtr interface{}, name, usage string)

// DefaultFlagVarTypes returns a list of
func DefaultFlagVarTypes() map[string]FlagVarTypeHandlerFunc {
	kfn := func(ptr interface{}) string {
		return buildFlagVarTypeKey(reflect.TypeOf(ptr))
	}
	return map[string]FlagVarTypeHandlerFunc{
		// *string
		kfn(new(string)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			var v string
			if varPtr.(*string) != nil {
				v = *varPtr.(*string)
			}
			fs.StringVar(varPtr.(*string), name, v, usage)
		},
		// *bool
		kfn(new(bool)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			var v bool
			if varPtr.(*bool) != nil {
				v = *varPtr.(*bool)
			}
			fs.BoolVar(varPtr.(*bool), name, v, usage)
		},
		// *int
		kfn(new(int)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			var v int
			if varPtr.(*int) != nil {
				v = *varPtr.(*int)
			}
			fs.IntVar(varPtr.(*int), name, v, usage)
		},
		// *int64
		kfn(new(int64)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			var v int64
			if varPtr.(*int64) != nil {
				v = *varPtr.(*int64)
			}
			fs.Int64Var(varPtr.(*int64), name, v, usage)
		},
		// *uint
		kfn(new(uint)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			var v uint
			if varPtr.(*uint) != nil {
				v = *varPtr.(*uint)
			}
			fs.UintVar(varPtr.(*uint), name, v, usage)
		},
		// *uint64
		kfn(new(uint64)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			var v uint64
			if varPtr.(*uint64) != nil {
				v = *varPtr.(*uint64)
			}
			fs.Uint64Var(varPtr.(*uint64), name, v, usage)
		},
		// *[]string
		kfn(new([]string)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			fs.Var(newStringSliceValue(varPtr.(*[]string)), name, usage)
		},
		// *[]int
		kfn(new([]int)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			fs.Var(newIntSliceValue(varPtr.(*[]int)), name, usage)
		},
		// *[]uint
		kfn(new([]uint)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			fs.Var(newUintSliceValue(varPtr.(*[]uint)), name, usage)
		},
		// *map[string]string
		kfn(new(map[string]string)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			fs.Var(newStringMapValue(varPtr.(*map[string]string)), name, usage)
		},
		// *map[string][]string
		kfn(new(map[string][]string)): func(fs *flag.FlagSet, vartPtr interface{}, name, usage string) {
			fs.Var(newStringSliceMapValue(vartPtr.(*map[string][]string)), name, usage)
		},
		// *http.Header
		kfn(new(http.Header)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			fs.Var(newStringSliceMapFromHTTPHeaderValue(varPtr.(*http.Header)), name, usage)
		},
		// *time.Duration
		kfn(new(time.Duration)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			var v time.Duration
			if varPtr.(*time.Duration) != nil {
				v = *varPtr.(*time.Duration)
			}
			fs.DurationVar(varPtr.(*time.Duration), name, v, usage)
		},
		// *net.IP
		kfn(new(net.IP)): func(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
			fs.Var(newIPStrValue(varPtr.(*net.IP)), name, usage)
		},
	}
}

// stringMapValue is pulled from https://github.com/hashicorp/consul/blob/b5abf61963c7b0bdb674602bfb64051f8e23ddb1/agent/config/flagset.go#L118
type stringMapValue map[string]string

func newStringMapValue(p *map[string]string) *stringMapValue {
	*p = make(map[string]string)
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

type stringSliceMapValue map[string][]string

func newStringSliceMapValue(p *map[string][]string) *stringSliceMapValue {
	*p = make(map[string][]string)
	return (*stringSliceMapValue)(p)
}

func newStringSliceMapFromHTTPHeaderValue(p *http.Header) *stringSliceMapValue {
	*p = make(http.Header)
	return (*stringSliceMapValue)(p)
}

func (s *stringSliceMapValue) Set(val string) error {
	p := strings.SplitN(val, ":", 2)
	k, v := p[0], ""
	if len(p) == 2 {
		v = p[1]
	}
	if c, ok := (*s)[k]; ok {
		(*s)[k] = append(c, v)
	} else {
		(*s)[k] = make([]string, 1)
		(*s)[k][0] = v
	}
	return nil
}

func (s *stringSliceMapValue) Get() interface{} {
	return s
}

func (s *stringSliceMapValue) String() string {
	out := ""
	for k, vs := range *s {
		out = fmt.Sprintf("%s;%s:", out, k)
		for i, v := range vs {
			out = fmt.Sprintf("%s%d=%q,", out, i, v)
		}
	}
	return out
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

type intSliceValue []int

func newIntSliceValue(p *[]int) *intSliceValue {
	return (*intSliceValue)(p)
}

func (s *intSliceValue) Set(val string) error {
	if i, err := strconv.Atoi(val); err != nil {
		return err
	} else {
		*s = append(*s, i)
		return nil
	}
}

func (s *intSliceValue) Get() interface{} {
	return s
}

func (s *intSliceValue) String() string {
	l := len(*s)
	tmp := make([]string, l, l)
	for i, v := range *s {
		tmp[i] = strconv.Itoa(v)
	}
	return strings.Join(tmp, " ")
}

type uintSliceValue []uint

func newUintSliceValue(p *[]uint) *uintSliceValue {
	return (*uintSliceValue)(p)
}

func (s *uintSliceValue) Set(val string) error {
	if i, err := strconv.ParseUint(val, 10, 64); err != nil {
		return err
	} else {
		*s = append(*s, uint(i))
		return nil
	}
}

func (s *uintSliceValue) Get() interface{} {
	return s
}

func (s *uintSliceValue) String() string {
	l := len(*s)
	tmp := make([]string, l, l)
	for i, v := range *s {
		tmp[i] = strconv.FormatUint(uint64(v), 10)
	}
	return strings.Join(tmp, " ")
}

type ipStrValue net.IP

func newIPStrValue(p *net.IP) *ipStrValue {
	return (*ipStrValue)(p)
}

func (ip *ipStrValue) Set(val string) error {
	*ip = ipStrValue(net.ParseIP(val))
	return nil
}

func (ip *ipStrValue) Get() interface{} {
	return ip
}

func (ip *ipStrValue) String() string {
	return net.IP(*ip).String()
}
