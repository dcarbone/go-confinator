# go-confinator

Flag var parsing helpers

## BuildInfo

Build info is a single portable type designed to house and carry "version" information about your program.

### Example:

Code:

```go
package main

import (
	"fmt"
	"github.com/dcarbone/go-confinator"
)

var (
	BuildName   string
	BuildDate   string
	BuildBranch string
	BuildNumber string
)

func main() {
	bi := confinator.NewBuildInfo(BuildName, BuildDate, BuildBranch, BuildNumber)
	fmt.Println(bi)
}
```

Build:

```shell
BUILD_NAME="buildinfo"
BUILD_DATE="$(date -u +%Y%m%d@%H%M%S%z)"
BUILD_BRANCH="$(git branch --no-color|awk '/^\*/ {print $2}')"
BUILD_NUMBER="00000"

go build -o="${BUILD_NAME}" -ldflags "\
-X main.BuildName=${BUILD_NAME} \
-X main.BuildDate=${BUILD_DATE} \
-X main.BuildBranch=${BUILD_BRANCH} \
-X main.BuildNumber=${BUILD_NUMBER}"
```

Show:

```shell
./buildinfo
# should print out provided build ldflags.
```

## Confinator

[Confinator](./confinator.go#L10)'s main purpose is to make parsing runtime flags into "complex" config types in a
flexible and extensible manner.

As an example:

```go
package main

import (
	"fmt"
	"flag"
	"net"
	"net/http"
	"os"
	"github.com/dcarbone/go-confinator"
)

type MyConfig struct {
	IP      net.IP
	Strings []string
	Map     map[string]string
	Header  http.Header
}

func main() {
	cf := confinator.NewConfinator()
	fs := flag.NewFlagSet("confinator-test", flag.PanicOnError)
	conf := MyConfig{
		Strings: make([]string, 0),
		Map: make(map[string]string),
		Header: make(http.Header),
	}
	cf.FlagVar(fs, &conf.IP, "ip", "IP address")
	cf.FlagVar(fs, &conf.Strings, "string", "any string, may be defined multiple times")
	cf.FlagVar(fs, &conf.Map, "map", "arbitrary map value, may be defined multiple times")
	cf.FlagVar(fs, &conf.Header, "header", "HTTP header values")
	_ = fs.Parse(os.Args[1:])
	fmt.Println(conf)
}
```

Put the above in a file named `main.go` or whatever you want

```shell
go run main.go \
  -ip 10.1.2.3 \
  -string string1 -string string2 \
  -map=key1:value11 -map key1:value12 \
  -map key2: value21 -map=key2:value22 \
  -header 'Authorization:Basic dGhlIGNha2UgaXMgYSBsaWU=' \
  -header 'Authorization:Basic ZG9scGhpbg=='
```

You should see this printed to stdout:

```
{10.1.2.3 [string1 string2] map[key1:value12 key2:value22] map[Authorization:[Basic dGhlIGNha2UgaXMgYSBsaWU= Basic ZG9scGhpbg==]]}
```

## Flag Var types

This package comes with a number of built-in types, which you can view [here](./flag_types.go#L22).

### Registering new flag var types

Registering new types to use as flags is easy.  Create a suitable type that implements the 
[Getter](https://golang.org/pkg/flag/#Getter) interface, then register it to a constructed 
[Confinator](./confinator.go#L10) instance using [RegisterFlagVarType](./confinator.go#L21)

You can look at the [DefaultFlagVarTypes](./flag_types.go#L22) func to see examples of this.