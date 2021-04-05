package confinator

import (
	"flag"
	"fmt"
	"strings"
)

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
