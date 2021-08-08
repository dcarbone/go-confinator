package confinator_test

import (
	"flag"
	"fmt"
	"math"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/dcarbone/go-confinator"
)

type testSuite struct {
	Name     string
	Usage    string
	VarPtr   interface{}
	FlagArgs []string
}

func TestConfinator(t *testing.T) {
	testSuites := map[string]testSuite{
		"string": {
			Name:     "steve",
			VarPtr:   new(string),
			FlagArgs: []string{"-steve", "\"is here\""},
		},
		"bool-true": {
			Name:     "yes",
			VarPtr:   new(bool),
			FlagArgs: []string{"-yes", "true"},
		},
		"bool-false": {
			Name:     "no",
			VarPtr:   new(bool),
			FlagArgs: []string{"-no", "false"},
		},
		"int-real": {
			Name:     "not-negative",
			VarPtr:   new(int),
			FlagArgs: []string{"-not-negative", "9001"},
		},
		"int-negative": {
			Name:     "negative",
			VarPtr:   new(int),
			FlagArgs: []string{"-negative", "-9001"},
		},
		"int64-real": {
			Name:     "bignum",
			VarPtr:   new(int64),
			FlagArgs: []string{"-bignum", strconv.Itoa(math.MaxInt64)},
		},
		"int64-negative": {
			Name:     "smallnum",
			VarPtr:   new(int64),
			FlagArgs: []string{"-smallnum", strconv.Itoa(math.MinInt64)},
		},
		"uint": {
			Name:     "uint",
			VarPtr:   new(uint),
			FlagArgs: []string{"-uint", strconv.FormatUint(math.MaxUint64, 10)},
		},
		"uint64": {
			Name:     "uint64",
			VarPtr:   new(uint),
			FlagArgs: []string{"-uint64", strconv.FormatUint(math.MaxUint64, 10)},
		},
		"[]string": {
			Name:     "strings",
			VarPtr:   &([]string{}),
			FlagArgs: []string{"-strings=one", "-strings", "two"},
		},
		"[]int": {
			Name:     "ints",
			VarPtr:   &([]int{}),
			FlagArgs: []string{"-ints=1", "-ints", "2"},
		},
		"[]uint": {
			Name:     "uints",
			VarPtr:   &([]uint{}),
			FlagArgs: []string{"-uints=11", "-uints", "22"},
		},
		"map[string]string": {
			Name:     "map",
			VarPtr:   &(map[string]string{}),
			FlagArgs: []string{"-map=key1:value1", "-map", "key2:value2"},
		},
		"map[string][]string": {
			Name:   "mapslice",
			VarPtr: &(map[string][]string{}),
			FlagArgs: []string{
				"-mapslice=key1:value1",
				"-mapslice", "key1:value12",
				"-mapslice=key2:value21",
				"-mapslice", "key2:value22"},
		},
		"http.Header": {
			Name:   "header",
			VarPtr: new(http.Header),
			FlagArgs: []string{
				"-header=Authorization:Basic dGhlIGNha2UgaXMgYSBsaWU=",
				"-header=Authorization:Basic ZG9scGhpbg==",
				"-header", "Content-Type:*/*",
			},
		},
		"time.Duration": {
			Name:     "td",
			VarPtr:   new(time.Duration),
			FlagArgs: []string{"-td", "5ns"},
		},
		"net.IP": {
			Name:     "ip",
			VarPtr:   new(net.IP),
			FlagArgs: []string{"-ip", "10.2.3.4"},
		},
		"url.URL": {
			Name:     "url",
			VarPtr:   new(url.URL),
			FlagArgs: []string{"-url", "https://google.com"},
		},
	}

	for tn, ts := range testSuites {
		t.Run(tn, func(t *testing.T) {
			ts := ts
			tn := tn
			if reflect.ValueOf(ts.VarPtr).Kind() != reflect.Ptr {
				t.Logf("Value provided to %q is a non-pointer: %v", tn, ts.VarPtr)
				t.Fail()
				return
			}
			t.Parallel()
			cf := confinator.NewConfinator()
			fs := flag.NewFlagSet(fmt.Sprintf("test-fs-%s", tn), flag.ContinueOnError)
			cf.FlagVar(fs, ts.VarPtr, ts.Name, ts.Usage)
			if err := fs.Parse(ts.FlagArgs); err != nil {
				t.Logf("Error running test %q: %v", tn, err)
				t.Fail()
			} else {
				t.Logf("%q Parsed value: %#v", tn, reflect.ValueOf(ts.VarPtr).Elem())
			}
		})
	}
}
