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
	flagVar     interface{}
	flagUsage   string
	flagName    string
	flagArgs    []string
	expectPanic bool
}

func TestConfinator(t *testing.T) {
	testSuites := map[string]testSuite{
		"string": {
			flagName: "steve",
			flagVar:  new(string),
			flagArgs: []string{"-steve", `"is here"`},
		},
		"string-nil": {
			flagVar:     (*string)(nil),
			expectPanic: true,
			flagName:    "steve",
			flagArgs:    []string{"-steve", `"is here"`},
		},
		"bool-true": {
			flagName: "yes",
			flagVar:  new(bool),
			flagArgs: []string{"-yes", "true"},
		},
		"bool-false": {
			flagName: "no",
			flagVar:  new(bool),
			flagArgs: []string{"-no", "false"},
		},
		"int-real": {
			flagName: "not-negative",
			flagVar:  new(int),
			flagArgs: []string{"-not-negative", "9001"},
		},
		"int-negative": {
			flagName: "negative",
			flagVar:  new(int),
			flagArgs: []string{"-negative", "-9001"},
		},
		"int64-real": {
			flagName: "bignum",
			flagVar:  new(int64),
			flagArgs: []string{"-bignum", strconv.Itoa(math.MaxInt64)},
		},
		"int64-negative": {
			flagName: "smallnum",
			flagVar:  new(int64),
			flagArgs: []string{"-smallnum", strconv.Itoa(math.MinInt64)},
		},
		"uint": {
			flagName: "uint",
			flagVar:  new(uint),
			flagArgs: []string{"-uint", strconv.FormatUint(math.MaxUint64, 10)},
		},
		"uint64": {
			flagName: "uint64",
			flagVar:  new(uint),
			flagArgs: []string{"-uint64", strconv.FormatUint(math.MaxUint64, 10)},
		},
		"[]string": {
			flagName: "strings",
			flagVar:  &([]string{}),
			flagArgs: []string{"-strings=one", "-strings", "two"},
		},
		"[]int": {
			flagName: "ints",
			flagVar:  &([]int{}),
			flagArgs: []string{"-ints=1", "-ints", "2"},
		},
		"[]uint": {
			flagName: "uints",
			flagVar:  &([]uint{}),
			flagArgs: []string{"-uints=11", "-uints", "22"},
		},
		"map[string]string": {
			flagName: "map",
			flagVar:  &(map[string]string{}),
			flagArgs: []string{"-map=key1:value1", "-map", "key2:value2"},
		},
		"map[string][]string": {
			flagName: "mapslice",
			flagVar:  &(map[string][]string{}),
			flagArgs: []string{
				"-mapslice=key1:value1",
				"-mapslice", "key1:value12",
				"-mapslice=key2:value21",
				"-mapslice", "key2:value22"},
		},
		"http.Header": {
			flagName: "header",
			flagVar:  new(http.Header),
			flagArgs: []string{
				"-header=Authorization:Basic dGhlIGNha2UgaXMgYSBsaWU=",
				"-header=Authorization:Basic ZG9scGhpbg==",
				"-header", "Content-Type:*/*",
			},
		},
		"time.Duration": {
			flagName: "td",
			flagVar:  new(time.Duration),
			flagArgs: []string{"-td", "5ns"},
		},
		"net.IP": {
			flagName: "ip",
			flagVar:  new(net.IP),
			flagArgs: []string{"-ip", "10.2.3.4"},
		},
		"url.URL": {
			flagName: "url",
			flagVar:  new(url.URL),
			flagArgs: []string{"-url", "https://google.com"},
		},
	}

	for tn, ts := range testSuites {
		t.Run(tn, func(t *testing.T) {
			ts := ts
			tn := tn
			defer func() {
				v := recover()
				if ts.expectPanic {
					if v == nil {
						t.Log("Expected panic, but got none")
						t.Fail()
					} else {
						t.Logf("Expected panic seen: %v", v)
					}
				}
			}()
			t.Parallel()
			cf := confinator.NewConfinator()
			fs := flag.NewFlagSet(fmt.Sprintf("test-fs-%s", tn), flag.ContinueOnError)
			cf.FlagVar(fs, ts.flagVar, ts.flagName, ts.flagUsage)
			if err := fs.Parse(ts.flagArgs); err != nil {
				t.Logf("Error running test %q: %v", tn, err)
				t.Fail()
			} else {
				t.Logf("%q Parsed value: %#v", tn, reflect.ValueOf(ts.flagVar).Elem())
			}
		})
	}
}
