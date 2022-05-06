package confinator

import (
	"flag"
	"fmt"
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
	cf.types[kfn(varPtr)] = fn
}

// FlagVar is a convenience method that handles a few common config struct -> flag cases
func (cf *Confinator) FlagVar(fs *flag.FlagSet, varPtr interface{}, name, usage string) {
	cf.mu.RLock()
	defer cf.mu.RUnlock()
	if varPtr == nil {
		panic("varPtr cannot be nil")
	}
	if fn, ok := cf.types[kfn(varPtr)]; ok {
		fn(fs, varPtr, name, usage)
	} else {
		panic(fmt.Sprintf("invalid type: %T", varPtr))
	}
}
