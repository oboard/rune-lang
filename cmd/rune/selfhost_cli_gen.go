package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type __CliCommand struct {
	__name      string
	__version   any
	__about     string
	__options   []__CliOption
	__arguments []__CliArgument
	__commands  []__CliCommand
	__aliases   []__CliCommandAlias
}

type __CliOption struct {
	__name         string
	__short        string
	__valueName    string
	__help         string
	__required     bool
	__defaultValue any
}

type __CliArgument struct {
	__name     string
	__help     string
	__required bool
}

type __CliParseResult struct {
	__command         __CliCommand
	__values          map[string]string
	__flags           map[string]bool
	__positionals     map[string]string
	__explicitOptions []string
	__args            []string
	__rest            []string
	__help            bool
	__error           any
}

type __CliCommandAlias struct {
	__from string
	__to   string
}

type __CliCommandParseResult struct {
	__root        __CliParseResult
	__command     __CliParseResult
	__commandName string
	__commandArgs []string
	__error       any
}

type __CliCommandArgs struct {
	__rootArgs    []string
	__commandName string
	__commandArgs []string
}

type __CliCommandLookup struct {
	__command __CliCommand
	__found   bool
}

type __RuneCliInvocation struct {
	__ok              bool
	__command         string
	__backend         string
	__path            string
	__output          string
	__target          string
	__pattern         string
	__checkOnly       bool
	__stdout          bool
	__backendExplicit bool
	__runArgs         []string
	__errors          []string
	__help            bool
	__helpText        string
}

func runeProcessArgv() []string { return append([]string(nil), os.Args...) }
func runeProcessCwd() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return cwd
}
func runeProcessEnv(name string) any {
	value, ok := os.LookupEnv(name)
	if !ok {
		return any(nil)
	}
	return value
}
func runeProcessExit(code int) struct{} { os.Exit(code); return struct{}{} }
func runeProcessPlatform() string       { return runtime.GOOS }

func __cli_alias(__from string, __to string) __CliCommandAlias {
	return __CliCommandAlias{__from: __from, __to: __to}
}

func __cli_aliasesForCommand(__aliases []__CliCommandAlias, __commandName string, __index int) []string {
	return func() []string {
		if __index >= len(__aliases) {
			return []string{}
		}
		return func() []string {
			if __aliases[__index].__to == __commandName {
				return func() []string {
					out := []string{}
					out = append(out, __aliases[__index].__from)
					out = append(out, __cli_aliasesForCommand(__aliases, __commandName, __index+1)...)
					return out
				}()
			}
			return __cli_aliasesForCommand(__aliases, __commandName, __index+1)
		}()
	}()
}

func __cli_appendRestSeparator(__args []string, __index int, __out []string) []string {
	return func() []string {
		if __args[__index] == "--" {
			return __cli_appendRuntimeRest(__args, __index, __out)
		}
		return __cli_appendRuntimeRest(__args, __index, func() []string { out := []string{}; out = append(out, __out...); out = append(out, "--"); return out }())
	}()
}

func __cli_appendRootOptionArgs(__rootArgs []string, __args []string, __index int, __consumesNext bool) []string {
	return func() []string {
		if __consumesNext && __index+1 < len(__args) {
			return func() []string {
				out := []string{}
				out = append(out, __rootArgs...)
				out = append(out, __args[__index])
				out = append(out, __args[__index+1])
				return out
			}()
		}
		return func() []string {
			out := []string{}
			out = append(out, __rootArgs...)
			out = append(out, __args[__index])
			return out
		}()
	}()
}

func __cli_appendRuntimeRest(__args []string, __index int, __out []string) []string {
	return func() []string {
		if __index >= len(__args) {
			return __out
		}
		return __cli_appendRuntimeRest(__args, __index+1, func() []string {
			out := []string{}
			out = append(out, __out...)
			out = append(out, __args[__index])
			return out
		}())
	}()
}

func __cli_argument(__name string, __help string, __required bool) __CliArgument {
	return __CliArgument{__name: __name, __help: __help, __required: __required}
}

func __cli_command(__name string, __about string) __CliCommand {
	return __CliCommand{__name: __name, __version: any(nil), __about: __about, __options: []__CliOption{}, __arguments: []__CliArgument{}, __commands: []__CliCommand{}, __aliases: []__CliCommandAlias{}}
}

func __cli_commandOptionConsumesNext(__command __CliCommand, __arg string) bool {
	return __cli_rootOptionConsumesNext(__command, __arg)
}

func __cli_contains(__values []string, __value string) bool {
	return __cli_containsAt(__values, __value, 0)
}

func __cli_containsAt(__values []string, __value string, __index int) bool {
	return func() bool {
		if __index >= len(__values) {
			return false
		}
		return func() bool {
			if __values[__index] == __value {
				return true
			}
			return __cli_containsAt(__values, __value, __index+1)
		}()
	}()
}

func __cli_emptyCommand() __CliCommand {
	return __CliCommand{__name: "", __version: any(nil), __about: "", __options: []__CliOption{}, __arguments: []__CliArgument{}, __commands: []__CliCommand{}, __aliases: []__CliCommandAlias{}}
}

func __cli_emptyOption() __CliOption {
	return __CliOption{__name: "", __short: "", __valueName: "", __help: "", __required: false, __defaultValue: any(nil)}
}

func __cli_findCommand(__commands []__CliCommand, __name string, __index int) __CliCommandLookup {
	return func() __CliCommandLookup {
		if len(__name) == 0 || __index >= len(__commands) {
			return __CliCommandLookup{__command: __cli_emptyCommand(), __found: false}
		}
		return func() __CliCommandLookup {
			if __commands[__index].__name == __name {
				return __CliCommandLookup{__command: __commands[__index], __found: true}
			}
			return __cli_findCommand(__commands, __name, __index+1)
		}()
	}()
}

func __cli_findOptionByName(__options []__CliOption, __name string, __index int) __CliOption {
	return func() __CliOption {
		if __index >= len(__options) {
			return __cli_emptyOption()
		}
		return func() __CliOption {
			if __options[__index].__name == __name {
				return __options[__index]
			}
			return __cli_findOptionByName(__options, __name, __index+1)
		}()
	}()
}

func __cli_findOptionByShort(__options []__CliOption, __short string, __index int) __CliOption {
	return func() __CliOption {
		if __index >= len(__options) {
			return __cli_emptyOption()
		}
		return func() __CliOption {
			if __options[__index].__short == __short {
				return __options[__index]
			}
			return __cli_findOptionByShort(__options, __short, __index+1)
		}()
	}()
}

func __cli_flag(__name string, __short string, __help string) __CliOption {
	return __CliOption{__name: __name, __short: __short, __valueName: "", __help: __help, __required: false, __defaultValue: any(nil)}
}

func __cli_help(__command __CliCommand) string {
	__text := "Usage: " + __command.__name
	func() {
		if len(__command.__options) == 0 {
			__text = __text
			return
		}
		__text = __text + " [options]"
	}()
	for _, __argument := range __command.__arguments {
		_ = __argument
		func() {
			__text = __text + " "
			__text = __text + func() string {
				if __argument.__required {
					return "<" + __argument.__name + ">"
				}
				return "[" + __argument.__name + "]"
			}()
		}()
	}
	__text = __text + "\n"
	func() {
		if len(__command.__commands) == 0 {
			__text = __text
			return
		}
		__text = __text + "\nCommands:\n"
	}()
	for _, __child := range __command.__commands {
		_ = __child
		func() {
			__text = __text + "  " + __child.__name
			__text = func() string {
				if len(__child.__about) == 0 {
					return __text
				}
				return __text + "\t" + __child.__about
			}()
			func() {
				for _, __aliasName := range __cli_aliasesForCommand(__command.__aliases, __child.__name, 0) {
					_ = __aliasName
					func() { __text = __text + " (alias: " + __aliasName + ")" }()
				}
			}()
			__text = __text + "\n"
		}()
	}
	__versionText := func() string {
		__coalesce1 := __command.__version
		if __coalesce1 != nil {
			return __coalesce1.(string)
		}
		return ""
	}()
	func() {
		if len(__versionText) == 0 {
			__text = __text
			return
		}
		__text = __text + "Version: " + __versionText + "\n"
	}()
	func() {
		if len(__command.__about) == 0 {
			__text = __text
			return
		}
		__text = __text + "\n" + __command.__about + "\n"
	}()
	func() {
		if len(__command.__arguments) == 0 {
			__text = __text
			return
		}
		__text = __text + "\nArguments:\n"
	}()
	for _, __argument := range __command.__arguments {
		_ = __argument
		func() {
			__text = __text + "  " + __argument.__name
			__text = func() string {
				if len(__argument.__help) == 0 {
					return __text
				}
				return __text + "\t" + __argument.__help
			}()
			__text = __text + "\n"
		}()
	}
	__text = __text + "\nOptions:\n"
	for _, __option := range __command.__options {
		_ = __option
		func() {
			__text = __text + "  "
			__text = func() string {
				if len(__option.__short) == 0 {
					return __text
				}
				return __text + "-" + __option.__short + ", "
			}()
			__text = __text + "--" + __option.__name
			__text = func() string {
				if len(__option.__valueName) == 0 {
					return __text
				}
				return __text + " <" + __option.__valueName + ">"
			}()
			__text = func() string {
				if len(__option.__help) == 0 {
					return __text
				}
				return __text + "\t" + __option.__help
			}()
			__text = func() string {
				if __option.__required {
					return __text + " (required)"
				}
				return __text
			}()
			__defaultValue := func() string {
				__coalesce2 := __option.__defaultValue
				if __coalesce2 != nil {
					return __coalesce2.(string)
				}
				return ""
			}()
			__text = func() string {
				if len(__defaultValue) == 0 {
					return __text
				}
				return __text + " (default: " + __defaultValue + ")"
			}()
			__text = __text + "\n"
		}()
	}
	return __text + "  -h, --help\tShow help\n"
}

func __cli_longOptionName(__arg string) string {
	__raw := func() string { runes := []rune(__arg); return string(runes[2:len([]rune(__arg))]) }()
	__equal := strings.Index(__raw, "=")
	return func() string {
		if __equal >= 0 {
			return func() string { runes := []rune(__raw); return string(runes[0:__equal]) }()
		}
		return __raw
	}()
}

func __cli_normalizeTrailingRestArg(__command __CliCommand, __args []string, __index int, __out []string, __positionalCount int, __skipNext bool) []string {
	__arg := __args[__index]
	return func() []string {
		if __positionalCount >= len(__command.__arguments) {
			return __cli_appendRestSeparator(__args, __index, __out)
		}
		return func() []string {
			if __skipNext {
				return __cli_normalizeTrailingRestArgsAt(__command, __args, __index+1, func() []string { out := []string{}; out = append(out, __out...); out = append(out, __arg); return out }(), __positionalCount, false)
			}
			return func() []string {
				if __cli_commandOptionConsumesNext(__command, __arg) {
					return __cli_normalizeTrailingRestArgsAt(__command, __args, __index+1, func() []string { out := []string{}; out = append(out, __out...); out = append(out, __arg); return out }(), __positionalCount, true)
				}
				return __cli_normalizeTrailingRestArgsAt(__command, __args, __index+1, func() []string { out := []string{}; out = append(out, __out...); out = append(out, __arg); return out }(), __positionalCount+__cli_positionalIncrement(__arg), false)
			}()
		}()
	}()
}

func __cli_normalizeTrailingRestArgs(__command __CliCommand, __args []string) []string {
	return __cli_normalizeTrailingRestArgsAt(__command, __args, 0, []string{}, 0, false)
}

func __cli_normalizeTrailingRestArgsAt(__command __CliCommand, __args []string, __index int, __out []string, __positionalCount int, __skipNext bool) []string {
	return func() []string {
		if __index >= len(__args) {
			return __out
		}
		return __cli_normalizeTrailingRestArg(__command, __args, __index, __out, __positionalCount, __skipNext)
	}()
}

func __cli_option(__name string, __short string, __valueName string, __help string, __required bool, __defaultValue any) __CliOption {
	return __CliOption{__name: __name, __short: __short, __valueName: __valueName, __help: __help, __required: __required, __defaultValue: __defaultValue}
}

func __cli_optionFound(__option __CliOption) bool {
	return __option.__name != ""
}

func __cli_parseArgs(__command __CliCommand, __args []string) __CliParseResult {
	__values := map[string]string{}
	__flags := map[string]bool{}
	__positionals := map[string]string{}
	__rest := append([]string{}, __args[0:0]...)
	__explicitOptions := append([]string{}, __args[0:0]...)
	__positionalValues := append([]string{}, __args[0:0]...)
	__helpValue := false
	__parseError := ""
	__afterDoubleDash := false
	__skipNext := false
	for _, __option := range __command.__options {
		_ = __option
		func() map[string]string {
			__useDefault := !(len(__option.__valueName) == 0) && __option.__defaultValue != any(nil)
			return func() map[string]string {
				if __useDefault {
					return func() map[string]string {
						__values[__option.__name] = func() string {
							__coalesce3 := __option.__defaultValue
							if __coalesce3 != nil {
								return __coalesce3.(string)
							}
							return ""
						}()
						return __values
					}()
				}
				return __values
			}()
		}()
	}
	for __index, __arg := range __args {
		_ = __arg
		_ = __index
		func() int {
			__handled := false
			func() {
				if __skipNext {
					__handled = true
					return
				}
				__handled = __handled
			}()
			func() {
				if __skipNext {
					__skipNext = false
					return
				}
				__skipNext = __skipNext
			}()
			__appendRest := !(__handled) && __afterDoubleDash
			func() int {
				if __appendRest {
					return func() int { __rest = append(__rest, __arg); return len(__rest) }()
				}
				return 0
			}()
			func() {
				if __appendRest {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__isSeparator := !(__handled) && len(__parseError) == 0 && __arg == "--"
			func() {
				if __isSeparator {
					__afterDoubleDash = true
					return
				}
				__afterDoubleDash = __afterDoubleDash
			}()
			func() {
				if __isSeparator {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__isHelp := !(__handled) && len(__parseError) == 0 && (__arg == "--help" || __arg == "-h")
			func() {
				if __isHelp {
					__helpValue = true
					return
				}
				__helpValue = __helpValue
			}()
			func() {
				if __isHelp {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__isLong := !(__handled) && len(__parseError) == 0 && strings.HasPrefix(__arg, "--") && len([]rune(__arg)) > 2
			__longName := func() string {
				if __isLong {
					return func() string { runes := []rune(__arg); return string(runes[2:len([]rune(__arg))]) }()
				}
				return ""
			}()
			__longValue := ""
			__longHasValue := false
			__eqIndex := strings.Index(__longName, "=")
			__hasEquals := __isLong && __eqIndex >= 0
			func() {
				if __hasEquals {
					__longValue = func() string {
						runes := []rune(__longName)
						return string(runes[__eqIndex+1 : len([]rune(__longName))])
					}()
					return
				}
				__longValue = __longValue
			}()
			func() {
				if __hasEquals {
					__longName = func() string { runes := []rune(__longName); return string(runes[0:__eqIndex]) }()
					return
				}
				__longName = __longName
			}()
			func() {
				if __hasEquals {
					__longHasValue = true
					return
				}
				__longHasValue = __longHasValue
			}()
			__noName := func() string {
				if strings.HasPrefix(__longName, "no-") {
					return func() string { runes := []rune(__longName); return string(runes[3:len([]rune(__longName))]) }()
				}
				return ""
			}()
			__noFound := false
			__noTakesValue := false
			func() {
				for _, __option := range __command.__options {
					_ = __option
					func() {
						__match := __isLong && !(__handled) && !(len(__noName) == 0) && __option.__name == __noName
						func() {
							if __match {
								__noFound = true
								return
							}
							__noFound = __noFound
						}()
						func() {
							if __match {
								__noTakesValue = !(len(__option.__valueName) == 0)
								return
							}
							__noTakesValue = __noTakesValue
						}()
					}()
				}
			}()
			__useNoFlag := __isLong && !(__handled) && !(len(__noName) == 0) && __noFound && !(__noTakesValue)
			func() map[string]bool {
				if __useNoFlag {
					return func() map[string]bool { __flags[__noName] = false; return __flags }()
				}
				return __flags
			}()
			func() int {
				if __useNoFlag {
					return func() int { __explicitOptions = append(__explicitOptions, __noName); return len(__explicitOptions) }()
				}
				return 0
			}()
			func() {
				if __useNoFlag {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__longFound := false
			__longTakesValue := false
			__longOptionName := ""
			func() {
				for _, __option := range __command.__options {
					_ = __option
					func() {
						__match := __isLong && !(__handled) && __option.__name == __longName
						func() {
							if __match {
								__longFound = true
								return
							}
							__longFound = __longFound
						}()
						func() {
							if __match {
								__longTakesValue = !(len(__option.__valueName) == 0)
								return
							}
							__longTakesValue = __longTakesValue
						}()
						func() {
							if __match {
								__longOptionName = __option.__name
								return
							}
							__longOptionName = __longOptionName
						}()
					}()
				}
			}()
			__unknownLong := __isLong && !(__handled) && !(__longFound)
			func() {
				if __unknownLong {
					__parseError = "unknown option --" + __longName
					return
				}
				__parseError = __parseError
			}()
			func() {
				if __unknownLong {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__missingLongValue := __isLong && !(__handled) && __longFound && __longTakesValue && !(__longHasValue) && __index+1 >= len(__args)
			func() {
				if __missingLongValue {
					__parseError = "missing value for --" + __longName
					return
				}
				__parseError = __parseError
			}()
			func() {
				if __missingLongValue {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__useLongNext := __isLong && !(__handled) && __longFound && __longTakesValue && !(__longHasValue)
			func() {
				if __useLongNext {
					__longValue = __args[__index+1]
					return
				}
				__longValue = __longValue
			}()
			func() {
				if __useLongNext {
					__skipNext = true
					return
				}
				__skipNext = __skipNext
			}()
			func() {
				if __useLongNext {
					__longHasValue = true
					return
				}
				__longHasValue = __longHasValue
			}()
			__storeLongValue := __isLong && !(__handled) && __longFound && __longTakesValue
			func() map[string]string {
				if __storeLongValue {
					return func() map[string]string { __values[__longOptionName] = __longValue; return __values }()
				}
				return __values
			}()
			func() int {
				if __storeLongValue {
					return func() int {
						__explicitOptions = append(__explicitOptions, __longOptionName)
						return len(__explicitOptions)
					}()
				}
				return 0
			}()
			func() {
				if __storeLongValue {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__storeLongFlag := __isLong && !(__handled) && __longFound && !(__longTakesValue)
			func() map[string]bool {
				if __storeLongFlag {
					return func() map[string]bool {
						__flags[__longOptionName] = !(__longHasValue) || __longValue != "false"
						return __flags
					}()
				}
				return __flags
			}()
			func() int {
				if __storeLongFlag {
					return func() int {
						__explicitOptions = append(__explicitOptions, __longOptionName)
						return len(__explicitOptions)
					}()
				}
				return 0
			}()
			func() {
				if __storeLongFlag {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__isShort := !(__handled) && len(__parseError) == 0 && strings.HasPrefix(__arg, "-") && __arg != "-"
			__clusterDone := false
			func() {
				for __shortIndex, __short := range func() []string {
					parts := strings.Split((func() string { runes := []rune(__arg); return string(runes[1:len([]rune(__arg))]) }()), "")
					return parts
				}() {
					_ = __short
					_ = __shortIndex
					func() int {
						__active := __isShort && !(__clusterDone) && len(__parseError) == 0
						__shortHelp := __active && __short == "h"
						func() {
							if __shortHelp {
								__helpValue = true
								return
							}
							__helpValue = __helpValue
						}()
						__shortFound := false
						__shortTakesValue := false
						__shortOptionName := ""
						func() {
							for _, __option := range __command.__options {
								_ = __option
								func() {
									__match := __active && !(__shortHelp) && __option.__short == __short
									func() {
										if __match {
											__shortFound = true
											return
										}
										__shortFound = __shortFound
									}()
									func() {
										if __match {
											__shortTakesValue = !(len(__option.__valueName) == 0)
											return
										}
										__shortTakesValue = __shortTakesValue
									}()
									func() {
										if __match {
											__shortOptionName = __option.__name
											return
										}
										__shortOptionName = __shortOptionName
									}()
								}()
							}
						}()
						__unknownShort := __active && !(__shortHelp) && !(__shortFound)
						func() {
							if __unknownShort {
								__parseError = "unknown option -" + __short
								return
							}
							__parseError = __parseError
						}()
						func() {
							if __unknownShort {
								__clusterDone = true
								return
							}
							__clusterDone = __clusterDone
						}()
						__shortValue := func() string { runes := []rune(__arg); return string(runes[__shortIndex+2 : len([]rune(__arg))]) }()
						__missingShortValue := __active && !(__shortHelp) && __shortFound && __shortTakesValue && len(__shortValue) == 0 && __index+1 >= len(__args)
						func() {
							if __missingShortValue {
								__parseError = "missing value for -" + __short
								return
							}
							__parseError = __parseError
						}()
						func() {
							if __missingShortValue {
								__clusterDone = true
								return
							}
							__clusterDone = __clusterDone
						}()
						__useShortNext := __active && !(__shortHelp) && __shortFound && __shortTakesValue && len(__shortValue) == 0 && len(__parseError) == 0
						func() {
							if __useShortNext {
								__shortValue = __args[__index+1]
								return
							}
							__shortValue = __shortValue
						}()
						func() {
							if __useShortNext {
								__skipNext = true
								return
							}
							__skipNext = __skipNext
						}()
						__storeShortValue := __active && !(__shortHelp) && __shortFound && __shortTakesValue && len(__parseError) == 0
						func() map[string]string {
							if __storeShortValue {
								return func() map[string]string { __values[__shortOptionName] = __shortValue; return __values }()
							}
							return __values
						}()
						func() int {
							if __storeShortValue {
								return func() int {
									__explicitOptions = append(__explicitOptions, __shortOptionName)
									return len(__explicitOptions)
								}()
							}
							return 0
						}()
						func() {
							if __storeShortValue {
								__clusterDone = true
								return
							}
							__clusterDone = __clusterDone
						}()
						__storeShortFlag := __active && !(__shortHelp) && __shortFound && !(__shortTakesValue) && len(__parseError) == 0
						func() map[string]bool {
							if __storeShortFlag {
								return func() map[string]bool { __flags[__shortOptionName] = true; return __flags }()
							}
							return __flags
						}()
						return func() int {
							if __storeShortFlag {
								return func() int {
									__explicitOptions = append(__explicitOptions, __shortOptionName)
									return len(__explicitOptions)
								}()
							}
							return 0
						}()
					}()
				}
			}()
			func() {
				if __isShort {
					__handled = true
					return
				}
				__handled = __handled
			}()
			__positional := !(__handled) && len(__parseError) == 0
			return func() int {
				if __positional {
					return func() int { __positionalValues = append(__positionalValues, __arg); return len(__positionalValues) }()
				}
				return 0
			}()
		}()
	}
	for _, __option := range __command.__options {
		_ = __option
		func() {
			__missing := len(__parseError) == 0 && !(__helpValue) && __option.__required && !(len(__option.__valueName) == 0) && !(func() bool { _, ok := __values[__option.__name]; return ok }())
			func() {
				if __missing {
					__parseError = "missing required option --" + __option.__name
					return
				}
				__parseError = __parseError
			}()
		}()
	}
	__positionIndex := 0
	for _, __argument := range __command.__arguments {
		_ = __argument
		func() {
			__hasValue := __positionIndex < len(__positionalValues)
			func() map[string]string {
				if __hasValue {
					return func() map[string]string {
						__positionals[__argument.__name] = __positionalValues[__positionIndex]
						return __positionals
					}()
				}
				return __positionals
			}()
			__positionIndex = __positionIndex + 1
		}()
	}
	for _, __argument := range __command.__arguments {
		_ = __argument
		func() {
			__missing := len(__parseError) == 0 && !(__helpValue) && __argument.__required && !(func() bool { _, ok := __positionals[__argument.__name]; return ok }())
			func() {
				if __missing {
					__parseError = "missing required argument " + __argument.__name
					return
				}
				__parseError = __parseError
			}()
		}()
	}
	__unexpected := len(__parseError) == 0 && !(__helpValue) && len(__positionalValues) > len(__command.__arguments)
	func() {
		if __unexpected {
			__parseError = "unexpected argument " + __positionalValues[len(__command.__arguments)]
			return
		}
		__parseError = __parseError
	}()
	return __CliParseResult{__command: __command, __values: __values, __flags: __flags, __positionals: __positionals, __explicitOptions: __explicitOptions, __args: __args, __rest: __rest, __help: __helpValue, __error: func() any {
		if len(__parseError) == 0 {
			return any(nil)
		}
		return __parseError
	}()}
}

func __cli_parseCommandArgs(__root __CliCommand, __commands []__CliCommand, __aliases []__CliCommandAlias, __trailingRest []string, __args []string) __CliCommandParseResult {
	__split := __cli_splitCommandArgs(__root, __aliases, __args, 0, []string{}, "", []string{}, false)
	__rootResult := __cli_parseArgs(__root, __split.__rootArgs)
	return func() __CliCommandParseResult {
		if len(__split.__commandName) == 0 {
			return __CliCommandParseResult{__root: __rootResult, __command: __rootResult, __commandName: "", __commandArgs: []string{}, __error: any(nil)}
		}
		return __cli_parseNamedCommandArgs(__rootResult, __commands, __trailingRest, __split)
	}()
}

func __cli_parseKnownCommandArgs(__rootResult __CliParseResult, __command __CliCommand, __trailingRest []string, __split __CliCommandArgs) __CliCommandParseResult {
	__args := func() []string {
		if __cli_contains(__trailingRest, __split.__commandName) {
			return __cli_normalizeTrailingRestArgs(__command, __split.__commandArgs)
		}
		return __split.__commandArgs
	}()
	__parsed := __cli_withReportedArgs(__cli_parseArgs(__command, __args), __split.__commandArgs)
	return __CliCommandParseResult{__root: __rootResult, __command: __parsed, __commandName: __split.__commandName, __commandArgs: __split.__commandArgs, __error: any(nil)}
}

func __cli_parseNamedCommandArgs(__rootResult __CliParseResult, __commands []__CliCommand, __trailingRest []string, __split __CliCommandArgs) __CliCommandParseResult {
	__lookup := __cli_findCommand(__commands, __split.__commandName, 0)
	return func() __CliCommandParseResult {
		if __lookup.__found {
			return __cli_parseKnownCommandArgs(__rootResult, __lookup.__command, __trailingRest, __split)
		}
		return __CliCommandParseResult{__root: __rootResult, __command: __rootResult, __commandName: __split.__commandName, __commandArgs: __split.__commandArgs, __error: "unknown command " + __split.__commandName}
	}()
}

func __cli_positionalIncrement(__arg string) int {
	return func() int {
		if strings.HasPrefix(__arg, "-") {
			return 0
		}
		return 1
	}()
}

func __cli_resolveAlias(__aliases []__CliCommandAlias, __name string, __index int) string {
	return func() string {
		if __index >= len(__aliases) {
			return __name
		}
		return func() string {
			if __aliases[__index].__from == __name {
				return __aliases[__index].__to
			}
			return __cli_resolveAlias(__aliases, __name, __index+1)
		}()
	}()
}

func __cli_rootOption(__root __CliCommand, __arg string) __CliOption {
	return func() __CliOption {
		if strings.HasPrefix(__arg, "--") && len([]rune(__arg)) > 2 {
			return __cli_findOptionByName(__root.__options, __cli_longOptionName(__arg), 0)
		}
		return func() __CliOption {
			if strings.HasPrefix(__arg, "-") && __arg != "-" {
				return __cli_findOptionByShort(__root.__options, func() string { runes := []rune(__arg); return string(runes[1:2]) }(), 0)
			}
			return __cli_emptyOption()
		}()
	}()
}

func __cli_rootOptionArg(__root __CliCommand, __arg string) bool {
	return __cli_optionFound(__cli_rootOption(__root, __arg))
}

func __cli_rootOptionConsumesNext(__root __CliCommand, __arg string) bool {
	__option := __cli_rootOption(__root, __arg)
	__hasInline := strings.Contains(__arg, "=") || strings.HasPrefix(__arg, "-") && !(strings.HasPrefix(__arg, "--")) && len([]rune(__arg)) > 2
	return __cli_optionFound(__option) && (__option.__valueName != "" && !(__hasInline))
}

func __cli_splitCommandArg(__root __CliCommand, __aliases []__CliCommandAlias, __args []string, __index int, __rootArgs []string, __commandName string, __commandArgs []string, __skipNext bool) __CliCommandArgs {
	return func() __CliCommandArgs {
		if __skipNext {
			return __cli_splitCommandArgs(__root, __aliases, __args, __index+1, __rootArgs, __commandName, __commandArgs, false)
		}
		return __cli_splitCommandArgValue(__root, __aliases, __args, __index, __rootArgs, __commandName, __commandArgs)
	}()
}

func __cli_splitCommandArgValue(__root __CliCommand, __aliases []__CliCommandAlias, __args []string, __index int, __rootArgs []string, __commandName string, __commandArgs []string) __CliCommandArgs {
	__arg := __args[__index]
	__isRootHelp := __commandName == "" && (__arg == "--help" || __arg == "-h")
	__isRootOption := __cli_rootOptionArg(__root, __arg)
	return func() __CliCommandArgs {
		if __isRootHelp {
			return __cli_splitCommandArgs(__root, __aliases, __args, __index+1, func() []string {
				out := []string{}
				out = append(out, __rootArgs...)
				out = append(out, __arg)
				return out
			}(), __commandName, __commandArgs, false)
		}
		return func() __CliCommandArgs {
			if __isRootOption {
				return __cli_splitRootOptionArg(__root, __aliases, __args, __index, __rootArgs, __commandName, __commandArgs, __cli_rootOptionConsumesNext(__root, __arg))
			}
			return func() __CliCommandArgs {
				if len(__commandName) == 0 {
					return __cli_splitCommandArgs(__root, __aliases, __args, __index+1, __rootArgs, __cli_resolveAlias(__aliases, __arg, 0), __commandArgs, false)
				}
				return __cli_splitCommandArgs(__root, __aliases, __args, __index+1, __rootArgs, __commandName, func() []string {
					out := []string{}
					out = append(out, __commandArgs...)
					out = append(out, __arg)
					return out
				}(), false)
			}()
		}()
	}()
}

func __cli_splitCommandArgs(__root __CliCommand, __aliases []__CliCommandAlias, __args []string, __index int, __rootArgs []string, __commandName string, __commandArgs []string, __skipNext bool) __CliCommandArgs {
	return func() __CliCommandArgs {
		if __index >= len(__args) {
			return __CliCommandArgs{__rootArgs: __rootArgs, __commandName: __commandName, __commandArgs: __commandArgs}
		}
		return __cli_splitCommandArg(__root, __aliases, __args, __index, __rootArgs, __commandName, __commandArgs, __skipNext)
	}()
}

func __cli_splitRootOptionArg(__root __CliCommand, __aliases []__CliCommandAlias, __args []string, __index int, __rootArgs []string, __commandName string, __commandArgs []string, __consumesNext bool) __CliCommandArgs {
	return __cli_splitCommandArgs(__root, __aliases, __args, __index+1, __cli_appendRootOptionArgs(__rootArgs, __args, __index, __consumesNext), __commandName, __commandArgs, __consumesNext)
}

func __cli_withAliases(__command __CliCommand, __aliases []__CliCommandAlias) __CliCommand {
	return __CliCommand{__name: __command.__name, __version: __command.__version, __about: __command.__about, __options: __command.__options, __arguments: __command.__arguments, __commands: __command.__commands, __aliases: __aliases}
}

func __cli_withArgument(__command __CliCommand, __argument __CliArgument) __CliCommand {
	return __CliCommand{__name: __command.__name, __version: __command.__version, __about: __command.__about, __options: __command.__options, __commands: __command.__commands, __aliases: __command.__aliases, __arguments: func() []__CliArgument {
		out := []__CliArgument{}
		out = append(out, __command.__arguments...)
		out = append(out, __argument)
		return out
	}()}
}

func __cli_withCommands(__command __CliCommand, __commands []__CliCommand) __CliCommand {
	return __CliCommand{__name: __command.__name, __version: __command.__version, __about: __command.__about, __options: __command.__options, __arguments: __command.__arguments, __aliases: __command.__aliases, __commands: __commands}
}

func __cli_withOption(__command __CliCommand, __option __CliOption) __CliCommand {
	return __CliCommand{__name: __command.__name, __version: __command.__version, __about: __command.__about, __arguments: __command.__arguments, __commands: __command.__commands, __aliases: __command.__aliases, __options: func() []__CliOption {
		out := []__CliOption{}
		out = append(out, __command.__options...)
		out = append(out, __option)
		return out
	}()}
}

func __cli_withReportedArgs(__result __CliParseResult, __args []string) __CliParseResult {
	return __CliParseResult{__command: __result.__command, __values: __result.__values, __flags: __result.__flags, __positionals: __result.__positionals, __explicitOptions: __result.__explicitOptions, __rest: __result.__rest, __help: __result.__help, __error: __result.__error, __args: __args}
}

func __parseCli(__args []string) __RuneCliInvocation {
	__parsed := __cli_parseCommandArgs(__runeCommand(), __runeCommands(), __runeAliases(), __runeTrailingRest(), __args)
	return __selfhost_cli_cli_invocationFromParsed(__parsed, __runeCommand(), __runeCommands())
}

func __selfhost_cli_cli_invocationFromParsed(__parsed __CliCommandParseResult, __rootCommand __CliCommand, __commands []__CliCommand) __RuneCliInvocation {
	__root := __parsed.__root
	__backend := func() string {
		value, ok := __root.__values["backend"]
		if ok {
			return value
		}
		return "go"
	}()
	__errors := __selfhost_cli_cli_cliErrors(__root)
	__commandError := func() string {
		__coalesce4 := __parsed.__error
		if __coalesce4 != nil {
			return __coalesce4.(string)
		}
		return ""
	}()
	func() int {
		if len(__commandError) == 0 {
			return 0
		}
		return func() int { __errors = append(__errors, __commandError); return len(__errors) }()
	}()
	__invalidBackend := __cli_contains([]string{"go", "ts", "mbt"}, __backend) == false
	func() int {
		if __invalidBackend {
			return func() int { __errors = append(__errors, "unsupported backend "+__backend); return len(__errors) }()
		}
		return 0
	}()
	__backendExplicit := __cli_contains(__root.__explicitOptions, "backend")
	return func() __RuneCliInvocation {
		if len(__errors) > 0 {
			return __selfhost_cli_cli_errorInvocation(__backend, __errors)
		}
		return func() __RuneCliInvocation {
			if len(__parsed.__commandName) == 0 {
				return __selfhost_cli_cli_helpInvocation(__rootCommand, __backend, __backendExplicit)
			}
			return __selfhost_cli_cli_invocationFromResult(__parsed.__commandName, __rootCommand, __commands, __backend, __backendExplicit, __parsed.__command)
		}()
	}()
}

func __selfhost_cli_cli_invocationFromResult(__name string, __rootCommand __CliCommand, __commands []__CliCommand, __backend string, __backendExplicit bool, __result __CliParseResult) __RuneCliInvocation {
	__errors := __selfhost_cli_cli_cliErrors(__result)
	__target := func() string {
		value, ok := __result.__values["target"]
		if ok {
			return value
		}
		return ""
	}()
	__target = __selfhost_cli_cli_defaultTargetForCommand(__name, __backend, __target)
	__errors = __selfhost_cli_cli_invocationErrors(__name, __backend, __target, __errors)
	return __RuneCliInvocation{__ok: len(__errors) == 0, __command: __name, __backend: __backend, __path: __selfhost_cli_cli_defaultPathForCommand(__name, func() string {
		value, ok := __result.__positionals["path"]
		if ok {
			return value
		}
		return ""
	}()), __output: func() string {
		value, ok := __result.__values["output"]
		if ok {
			return value
		}
		return ""
	}(), __target: __target, __pattern: func() string {
		value, ok := __result.__positionals["pattern"]
		if ok {
			return value
		}
		return ""
	}(), __checkOnly: func() bool {
		value, ok := __result.__flags["check"]
		if ok {
			return value
		}
		return false
	}(), __stdout: func() bool {
		value, ok := __result.__flags["stdout"]
		if ok {
			return value
		}
		return false
	}(), __backendExplicit: __backendExplicit, __runArgs: __result.__rest, __errors: __errors, __help: __result.__help, __helpText: __cli_help(__selfhost_cli_cli_invocationHelpCommand(__rootCommand, __commands, __name))}
}

func __selfhost_cli_cli_invocationErrors(__name string, __backend string, __target string, __errors []string) []string {
	return func() []string {
		if __name == "build" && __cli_contains([]string{"go", "mbt"}, __backend) == false {
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "rune build only supports --backend go or --backend mbt")
				return out
			}()
		}
		return func() []string {
			if __name == "run" && len(__target) == 0 == false && __backend != "mbt" {
				return func() []string {
					out := []string{}
					out = append(out, __errors...)
					out = append(out, "rune run --target is only supported with --backend mbt")
					return out
				}()
			}
			return __errors
		}()
	}()
}

func __selfhost_cli_cli_defaultTargetForCommand(__name string, __backend string, __target string) string {
	return func() string {
		if __name == "run" && __backend == "mbt" && len(__target) == 0 {
			return "native"
		}
		return __target
	}()
}

func __selfhost_cli_cli_defaultPathForCommand(__name string, __path string) string {
	return func() string {
		if __name == "test" && len(__path) == 0 {
			return "tests"
		}
		return __path
	}()
}

func __selfhost_cli_cli_invocationHelpCommand(__rootCommand __CliCommand, __commands []__CliCommand, __name string) __CliCommand {
	return func() __CliCommand {
		__array5 := __commands
		__result6 := __rootCommand
		for _, __value8 := range __array5 {
			__result6 = func(__found __CliCommand, __command __CliCommand) __CliCommand {
				return func() __CliCommand {
					if __command.__name == __name {
						return __command
					}
					return __found
				}()
			}(__result6, __value8)
		}
		return __result6
	}()
}

func __selfhost_cli_cli_cliErrors(__result __CliParseResult) []string {
	__error := func() string {
		__coalesce9 := __result.__error
		if __coalesce9 != nil {
			return __coalesce9.(string)
		}
		return ""
	}()
	__errors := []string{}
	func() int {
		if len(__error) == 0 {
			return 0
		}
		return func() int { __errors = append(__errors, __error); return len(__errors) }()
	}()
	return __errors
}

func __selfhost_cli_cli_helpInvocation(__rootCommand __CliCommand, __backend string, __backendExplicit bool) __RuneCliInvocation {
	return __RuneCliInvocation{__ok: true, __command: "", __backend: __backend, __path: "", __output: "", __target: "", __pattern: "", __checkOnly: false, __stdout: false, __backendExplicit: __backendExplicit, __runArgs: []string{}, __errors: []string{}, __help: true, __helpText: __cli_help(__rootCommand)}
}

func __selfhost_cli_cli_errorInvocation(__backend string, __errors []string) __RuneCliInvocation {
	return __RuneCliInvocation{__ok: false, __command: "", __backend: __backend, __path: "", __output: "", __target: "", __pattern: "", __checkOnly: false, __stdout: false, __backendExplicit: false, __runArgs: []string{}, __errors: __errors, __help: false, __helpText: ""}
}

func __selfhost_cli_cli_emptyInvocation() __RuneCliInvocation {
	return __selfhost_cli_cli_errorInvocation("", []string{})
}

func __runeCommand() __CliCommand {
	__command := __cli_command("rune", "Rune language toolchain")
	__command = __cli_withOption(__command, __cli_option("backend", "b", "BACKEND", "target backend", false, "go"))
	__command = __cli_withCommands(__command, __runeCommands())
	return __cli_withAliases(__command, __runeAliases())
}

func __runeCommands() []__CliCommand {
	return []__CliCommand{__runCommand(), __buildCommand(), __emitCommand("go"), __emitCommand("ts"), __emitCommand("dts"), __emitCommand("mbt"), __singlePathCommand("check"), __fmtCommand(), __testCommand(), __cli_command("repl", ""), __lspCommand()}
}

func __runeAliases() []__CliCommandAlias {
	return []__CliCommandAlias{__cli_alias("format", "fmt")}
}

func __runeTrailingRest() []string {
	return []string{"run"}
}

func __runCommand() __CliCommand {
	__command := __cli_command("run", "Compile and run a Rune program")
	__command = __cli_withOption(__command, __cli_option("target", "", "TARGET", "MoonBit run target", false, any(nil)))
	return __cli_withArgument(__command, __cli_argument("path", "Rune source path", true))
}

func __buildCommand() __CliCommand {
	__command := __cli_command("build", "Compile a Rune program to an executable")
	__command = __cli_withOption(__command, __cli_option("output", "o", "FILE", "output executable path", false, ""))
	__command = __cli_withOption(__command, __cli_option("target", "", "TARGET", "build target", false, ""))
	return __cli_withArgument(__command, __cli_argument("path", "Rune source path", true))
}

func __emitCommand(__name string) __CliCommand {
	__command := __cli_command(__name, "Compile a Rune program")
	__command = __cli_withOption(__command, __cli_option("output", "o", "FILE", "output file", false, ""))
	return __cli_withArgument(__command, __cli_argument("path", "Rune source path", true))
}

func __singlePathCommand(__name string) __CliCommand {
	__command := __cli_command(__name, "Parse and type-check Rune source")
	return __cli_withArgument(__command, __cli_argument("path", "Rune source path", true))
}

func __fmtCommand() __CliCommand {
	__command := __cli_command("fmt", "Format Rune source")
	__command = __cli_withOption(__command, __cli_flag("check", "", "fail if not formatted"))
	__command = __cli_withOption(__command, __cli_flag("stdout", "", "write formatted source to stdout"))
	return __cli_withArgument(__command, __cli_argument("path", "Rune source path", true))
}

func __testCommand() __CliCommand {
	__command := __cli_command("test", "Run Rune tests")
	__command = __cli_withArgument(__command, __cli_argument("path", "Rune test path", false))
	return __cli_withArgument(__command, __cli_argument("pattern", "test name pattern", false))
}

func __lspCommand() __CliCommand {
	__command := __cli_command("lsp", "Start the Rune language server")
	return __cli_withOption(__command, __cli_flag("stdio", "", "serve LSP over stdin/stdout"))
}

func __main() {
	__argv := runeProcessArgv()
	__invocation := __parseCli(append([]string{}, __argv[1:len(__argv)]...))
	func() {
		if __invocation.__help {
			fmt.Print(__cli_help(__runeCommand()))
			return
		}
		fmt.Println(__invocation.__command)
	}()
}
