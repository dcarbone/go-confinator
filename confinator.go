package confinator

import (
	"flag"
	"fmt"
	"reflect"
	"sync"
)

type Confinator struct {
	mu    sync.RWMutex
	types map[string]FlagVarTypeHandlerFunc
}

func NewConfinator() *Confinator {
	cf := new(Confinator)
	cf.types = DefaultFlagVarTypes()
	return cf
}

func (cf *Confinator) RegisterFlagVarType(varPtr interface{}, fn FlagVarTypeHandlerFunc) {
	cf.mu.Lock()
	defer cf.mu.Unlock()
	varType := reflect.TypeOf(varPtr)
	if varType.Kind() != reflect.Ptr {
		panic(fmt.Sprintf("Must provided a pointer, saw %T", varPtr))
	}
	cf.types[buildFlagVarTypeKey(varType)] = fn
}

// FlagVar is a convenience method that handles a few common config struct -> flag cases
func (cf *Confinator) FlagVar(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	if fn, ok := cf.types[buildFlagVarTypeKey(reflect.TypeOf(varPtr))]; ok {
		fn(fs, varPtr, name, usage)
	} else {
		panic(fmt.Sprintf("invalid type: %T", varPtr))
	}
}
