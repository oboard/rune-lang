package main

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

type CliCommand struct {
	name      string
	version   any
	about     string
	options   []CliOption
	arguments []CliArgument
	commands  []CliCommand
	aliases   []CliCommandAlias
}

type CliOption struct {
	name         string
	short        string
	valueName    string
	help         string
	required     bool
	defaultValue any
}

type CliArgument struct {
	name     string
	help     string
	required bool
}

type CliParseResult struct {
	command         CliCommand
	values          map[string]string
	flags           map[string]bool
	positionals     map[string]string
	explicitOptions []string
	args            []string
	rest            []string
	help            bool
	error_          any
}

type CliCommandAlias struct {
	from string
	to   string
}

type CliCommandParseResult struct {
	root        CliParseResult
	command     CliParseResult
	commandName string
	commandArgs []string
	error_      any
}

type CliCommandArgs struct {
	rootArgs    []string
	commandName string
	commandArgs []string
}

type CliCommandLookup struct {
	command CliCommand
	found   bool
}

type RuneCliInvocation struct {
	ok              bool
	command         string
	backend         string
	path            string
	output          string
	target          string
	pattern         string
	checkOnly       bool
	stdout          bool
	backendExplicit bool
	runArgs         []string
	errors          []string
	help            bool
	helpText        string
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

func cli_alias(from string, to string) CliCommandAlias {
	return CliCommandAlias{from: from, to: to}
}

func cli_aliasesForCommand(aliases []CliCommandAlias, commandName string, index int) []string {
	return func() []string {
		if index >= len(aliases) {
			return []string{}
		}
		return func() []string {
			if aliases[index].to == commandName {
				return func() []string {
					__rune_spread_out := []string{}
					__rune_spread_out = append(__rune_spread_out, aliases[index].from)
					__rune_spread_out = append(__rune_spread_out, cli_aliasesForCommand(aliases, commandName, index+1)...)
					return __rune_spread_out
				}()
			}
			return cli_aliasesForCommand(aliases, commandName, index+1)
		}()
	}()
}

func cli_appendRestSeparator(args []string, index int, out []string) []string {
	return func() []string {
		if args[index] == "--" {
			return cli_appendRuntimeRest(args, index, out)
		}
		return cli_appendRuntimeRest(args, index, func() []string {
			__rune_spread_out := []string{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, "--")
			return __rune_spread_out
		}())
	}()
}

func cli_appendRootOptionArgs(rootArgs []string, args []string, index int, consumesNext bool) []string {
	return func() []string {
		if consumesNext && index+1 < len(args) {
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, rootArgs...)
				__rune_spread_out = append(__rune_spread_out, args[index])
				__rune_spread_out = append(__rune_spread_out, args[index+1])
				return __rune_spread_out
			}()
		}
		return func() []string {
			__rune_spread_out := []string{}
			__rune_spread_out = append(__rune_spread_out, rootArgs...)
			__rune_spread_out = append(__rune_spread_out, args[index])
			return __rune_spread_out
		}()
	}()
}

func cli_appendRuntimeRest(args []string, index int, out []string) []string {
	for {
		if index >= len(args) {
			return out
		} else {
			args, index, out = args, index+1, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, args[index])
				return __rune_spread_out
			}()
			continue
		}
	}
}

func cli_argument(name string, help string, required bool) CliArgument {
	return CliArgument{name: name, help: help, required: required}
}

func cli_command(name string, about string) CliCommand {
	return CliCommand{name: name, version: any(nil), about: about, options: []CliOption{}, arguments: []CliArgument{}, commands: []CliCommand{}, aliases: []CliCommandAlias{}}
}

func cli_commandOptionConsumesNext(command CliCommand, arg string) bool {
	return cli_rootOptionConsumesNext(command, arg)
}

func cli_contains(values []string, value string) bool {
	return cli_containsAt(values, value, 0)
}

func cli_containsAt(values []string, value string, index int) bool {
	for {
		if index >= len(values) {
			return false
		} else {
			if values[index] == value {
				return true
			} else {
				values, value, index = values, value, index+1
				continue
			}
		}
	}
}

func cli_emptyCommand() CliCommand {
	return CliCommand{name: "", version: any(nil), about: "", options: []CliOption{}, arguments: []CliArgument{}, commands: []CliCommand{}, aliases: []CliCommandAlias{}}
}

func cli_emptyOption() CliOption {
	return CliOption{name: "", short: "", valueName: "", help: "", required: false, defaultValue: any(nil)}
}

func cli_findCommand(commands []CliCommand, name string, index int) CliCommandLookup {
	for {
		if len(name) == 0 || index >= len(commands) {
			return CliCommandLookup{command: cli_emptyCommand(), found: false}
		} else {
			if commands[index].name == name {
				return CliCommandLookup{command: commands[index], found: true}
			} else {
				commands, name, index = commands, name, index+1
				continue
			}
		}
	}
}

func cli_findOptionByName(options []CliOption, name string, index int) CliOption {
	for {
		if index >= len(options) {
			return cli_emptyOption()
		} else {
			if options[index].name == name {
				return options[index]
			} else {
				options, name, index = options, name, index+1
				continue
			}
		}
	}
}

func cli_findOptionByShort(options []CliOption, short string, index int) CliOption {
	for {
		if index >= len(options) {
			return cli_emptyOption()
		} else {
			if options[index].short == short {
				return options[index]
			} else {
				options, short, index = options, short, index+1
				continue
			}
		}
	}
}

func cli_flag(name string, short string, help string) CliOption {
	return CliOption{name: name, short: short, valueName: "", help: help, required: false, defaultValue: any(nil)}
}

func cli_help(command CliCommand) string {
	text := "Usage: " + command.name
	func() {
		if len(command.options) == 0 {
			text = text
			return
		}
		text = text + " [options]"
	}()
	for _, argument := range command.arguments {
		_ = argument
		func() {
			text = text + " "
			text = text + func() string {
				if argument.required {
					return "<" + argument.name + ">"
				}
				return "[" + argument.name + "]"
			}()
		}()
	}
	text = text + "\n"
	func() {
		if len(command.commands) == 0 {
			text = text
			return
		}
		text = text + "\nCommands:\n"
	}()
	for _, child := range command.commands {
		_ = child
		func() {
			text = text + "  " + child.name
			text = func() string {
				if len(child.about) == 0 {
					return text
				}
				return text + "\t" + child.about
			}()
			func() {
				for _, aliasName := range cli_aliasesForCommand(command.aliases, child.name, 0) {
					_ = aliasName
					func() { text = text + " (alias: " + aliasName + ")" }()
				}
			}()
			text = text + "\n"
		}()
	}
	versionText := func() string {
		coalesce1 := command.version
		if coalesce1 != nil {
			return coalesce1.(string)
		}
		return ""
	}()
	func() {
		if len(versionText) == 0 {
			text = text
			return
		}
		text = text + "Version: " + versionText + "\n"
	}()
	func() {
		if len(command.about) == 0 {
			text = text
			return
		}
		text = text + "\n" + command.about + "\n"
	}()
	func() {
		if len(command.arguments) == 0 {
			text = text
			return
		}
		text = text + "\nArguments:\n"
	}()
	for _, argument := range command.arguments {
		_ = argument
		func() {
			text = text + "  " + argument.name
			text = func() string {
				if len(argument.help) == 0 {
					return text
				}
				return text + "\t" + argument.help
			}()
			text = text + "\n"
		}()
	}
	text = text + "\nOptions:\n"
	for _, option := range command.options {
		_ = option
		func() {
			text = text + "  "
			text = func() string {
				if len(option.short) == 0 {
					return text
				}
				return text + "-" + option.short + ", "
			}()
			text = text + "--" + option.name
			text = func() string {
				if len(option.valueName) == 0 {
					return text
				}
				return text + " <" + option.valueName + ">"
			}()
			text = func() string {
				if len(option.help) == 0 {
					return text
				}
				return text + "\t" + option.help
			}()
			text = func() string {
				if option.required {
					return text + " (required)"
				}
				return text
			}()
			defaultValue := func() string {
				coalesce2 := option.defaultValue
				if coalesce2 != nil {
					return coalesce2.(string)
				}
				return ""
			}()
			text = func() string {
				if len(defaultValue) == 0 {
					return text
				}
				return text + " (default: " + defaultValue + ")"
			}()
			text = text + "\n"
		}()
	}
	return text + "  -h, --help\tShow help\n"
}

func cli_longOptionName(arg string) string {
	raw := func() string { runes := []rune(arg); return string(runes[2:len([]rune(arg))]) }()
	equal := strings.Index(raw, "=")
	return func() string {
		if equal >= 0 {
			return func() string { runes := []rune(raw); return string(runes[0:equal]) }()
		}
		return raw
	}()
}

func cli_normalizeTrailingRestArg(command CliCommand, args []string, index int, out []string, positionalCount int, skipNext bool) []string {
	arg := args[index]
	return func() []string {
		if positionalCount >= len(command.arguments) {
			return cli_appendRestSeparator(args, index, out)
		}
		return func() []string {
			if skipNext {
				return cli_normalizeTrailingRestArgsAt(command, args, index+1, func() []string {
					__rune_spread_out := []string{}
					__rune_spread_out = append(__rune_spread_out, out...)
					__rune_spread_out = append(__rune_spread_out, arg)
					return __rune_spread_out
				}(), positionalCount, false)
			}
			return func() []string {
				if cli_commandOptionConsumesNext(command, arg) {
					return cli_normalizeTrailingRestArgsAt(command, args, index+1, func() []string {
						__rune_spread_out := []string{}
						__rune_spread_out = append(__rune_spread_out, out...)
						__rune_spread_out = append(__rune_spread_out, arg)
						return __rune_spread_out
					}(), positionalCount, true)
				}
				return cli_normalizeTrailingRestArgsAt(command, args, index+1, func() []string {
					__rune_spread_out := []string{}
					__rune_spread_out = append(__rune_spread_out, out...)
					__rune_spread_out = append(__rune_spread_out, arg)
					return __rune_spread_out
				}(), positionalCount+cli_positionalIncrement(arg), false)
			}()
		}()
	}()
}

func cli_normalizeTrailingRestArgs(command CliCommand, args []string) []string {
	return cli_normalizeTrailingRestArgsAt(command, args, 0, []string{}, 0, false)
}

func cli_normalizeTrailingRestArgsAt(command CliCommand, args []string, index int, out []string, positionalCount int, skipNext bool) []string {
	return func() []string {
		if index >= len(args) {
			return out
		}
		return cli_normalizeTrailingRestArg(command, args, index, out, positionalCount, skipNext)
	}()
}

func cli_option(name string, short string, valueName string, help string, required bool, defaultValue any) CliOption {
	return CliOption{name: name, short: short, valueName: valueName, help: help, required: required, defaultValue: defaultValue}
}

func cli_optionFound(option CliOption) bool {
	return option.name != ""
}

func cli_parseArgs(command CliCommand, args []string) CliParseResult {
	values := map[string]string{}
	flags := map[string]bool{}
	positionals := map[string]string{}
	rest := []string{}
	explicitOptions := []string{}
	positionalValues := []string{}
	helpValue := false
	parseError := ""
	afterDoubleDash := false
	skipNext := false
	for _, option := range command.options {
		_ = option
		func() map[string]string {
			useDefault := !(len(option.valueName) == 0) && option.defaultValue != any(nil)
			return func() map[string]string {
				if useDefault {
					return func() map[string]string {
						values[option.name] = func() string {
							coalesce3 := option.defaultValue
							if coalesce3 != nil {
								return coalesce3.(string)
							}
							return ""
						}()
						return values
					}()
				}
				return values
			}()
		}()
	}
	for index, arg := range args {
		_ = arg
		_ = index
		func() int {
			handled := false
			func() {
				if skipNext {
					handled = true
					return
				}
				handled = handled
			}()
			func() {
				if skipNext {
					skipNext = false
					return
				}
				skipNext = skipNext
			}()
			appendRest := !(handled) && afterDoubleDash
			func() int {
				if appendRest {
					return func() int { rest = append(rest, arg); return len(rest) }()
				}
				return 0
			}()
			func() {
				if appendRest {
					handled = true
					return
				}
				handled = handled
			}()
			isSeparator := !(handled) && len(parseError) == 0 && arg == "--"
			func() {
				if isSeparator {
					afterDoubleDash = true
					return
				}
				afterDoubleDash = afterDoubleDash
			}()
			func() {
				if isSeparator {
					handled = true
					return
				}
				handled = handled
			}()
			isHelp := !(handled) && len(parseError) == 0 && (arg == "--help" || arg == "-h")
			func() {
				if isHelp {
					helpValue = true
					return
				}
				helpValue = helpValue
			}()
			func() {
				if isHelp {
					handled = true
					return
				}
				handled = handled
			}()
			isLong := !(handled) && len(parseError) == 0 && strings.HasPrefix(arg, "--") && len([]rune(arg)) > 2
			longName := func() string {
				if isLong {
					return func() string { runes := []rune(arg); return string(runes[2:len([]rune(arg))]) }()
				}
				return ""
			}()
			longValue := ""
			longHasValue := false
			eqIndex := strings.Index(longName, "=")
			hasEquals := isLong && eqIndex >= 0
			func() {
				if hasEquals {
					longValue = func() string { runes := []rune(longName); return string(runes[eqIndex+1 : len([]rune(longName))]) }()
					return
				}
				longValue = longValue
			}()
			func() {
				if hasEquals {
					longName = func() string { runes := []rune(longName); return string(runes[0:eqIndex]) }()
					return
				}
				longName = longName
			}()
			func() {
				if hasEquals {
					longHasValue = true
					return
				}
				longHasValue = longHasValue
			}()
			noName := func() string {
				if strings.HasPrefix(longName, "no-") {
					return func() string { runes := []rune(longName); return string(runes[3:len([]rune(longName))]) }()
				}
				return ""
			}()
			noFound := false
			noTakesValue := false
			func() {
				for _, option := range command.options {
					_ = option
					func() {
						match := isLong && !(handled) && !(len(noName) == 0) && option.name == noName
						func() {
							if match {
								noFound = true
								return
							}
							noFound = noFound
						}()
						func() {
							if match {
								noTakesValue = !(len(option.valueName) == 0)
								return
							}
							noTakesValue = noTakesValue
						}()
					}()
				}
			}()
			useNoFlag := isLong && !(handled) && !(len(noName) == 0) && noFound && !(noTakesValue)
			func() map[string]bool {
				if useNoFlag {
					return func() map[string]bool { flags[noName] = false; return flags }()
				}
				return flags
			}()
			func() int {
				if useNoFlag {
					return func() int { explicitOptions = append(explicitOptions, noName); return len(explicitOptions) }()
				}
				return 0
			}()
			func() {
				if useNoFlag {
					handled = true
					return
				}
				handled = handled
			}()
			longFound := false
			longTakesValue := false
			longOptionName := ""
			func() {
				for _, option := range command.options {
					_ = option
					func() {
						match := isLong && !(handled) && option.name == longName
						func() {
							if match {
								longFound = true
								return
							}
							longFound = longFound
						}()
						func() {
							if match {
								longTakesValue = !(len(option.valueName) == 0)
								return
							}
							longTakesValue = longTakesValue
						}()
						func() {
							if match {
								longOptionName = option.name
								return
							}
							longOptionName = longOptionName
						}()
					}()
				}
			}()
			unknownLong := isLong && !(handled) && !(longFound)
			func() {
				if unknownLong {
					parseError = "unknown option --" + longName
					return
				}
				parseError = parseError
			}()
			func() {
				if unknownLong {
					handled = true
					return
				}
				handled = handled
			}()
			missingLongValue := isLong && !(handled) && longFound && longTakesValue && !(longHasValue) && index+1 >= len(args)
			func() {
				if missingLongValue {
					parseError = "missing value for --" + longName
					return
				}
				parseError = parseError
			}()
			func() {
				if missingLongValue {
					handled = true
					return
				}
				handled = handled
			}()
			useLongNext := isLong && !(handled) && longFound && longTakesValue && !(longHasValue)
			func() {
				if useLongNext {
					longValue = args[index+1]
					return
				}
				longValue = longValue
			}()
			func() {
				if useLongNext {
					skipNext = true
					return
				}
				skipNext = skipNext
			}()
			func() {
				if useLongNext {
					longHasValue = true
					return
				}
				longHasValue = longHasValue
			}()
			storeLongValue := isLong && !(handled) && longFound && longTakesValue
			func() map[string]string {
				if storeLongValue {
					return func() map[string]string { values[longOptionName] = longValue; return values }()
				}
				return values
			}()
			func() int {
				if storeLongValue {
					return func() int { explicitOptions = append(explicitOptions, longOptionName); return len(explicitOptions) }()
				}
				return 0
			}()
			func() {
				if storeLongValue {
					handled = true
					return
				}
				handled = handled
			}()
			storeLongFlag := isLong && !(handled) && longFound && !(longTakesValue)
			func() map[string]bool {
				if storeLongFlag {
					return func() map[string]bool { flags[longOptionName] = !(longHasValue) || longValue != "false"; return flags }()
				}
				return flags
			}()
			func() int {
				if storeLongFlag {
					return func() int { explicitOptions = append(explicitOptions, longOptionName); return len(explicitOptions) }()
				}
				return 0
			}()
			func() {
				if storeLongFlag {
					handled = true
					return
				}
				handled = handled
			}()
			isShort := !(handled) && len(parseError) == 0 && strings.HasPrefix(arg, "-") && arg != "-"
			clusterDone := false
			func() {
				for shortIndex, short := range func() []string {
					parts := strings.Split((func() string { runes := []rune(arg); return string(runes[1:len([]rune(arg))]) }()), "")
					return parts
				}() {
					_ = short
					_ = shortIndex
					func() int {
						active := isShort && !(clusterDone) && len(parseError) == 0
						shortHelp := active && short == "h"
						func() {
							if shortHelp {
								helpValue = true
								return
							}
							helpValue = helpValue
						}()
						shortFound := false
						shortTakesValue := false
						shortOptionName := ""
						func() {
							for _, option := range command.options {
								_ = option
								func() {
									match := active && !(shortHelp) && option.short == short
									func() {
										if match {
											shortFound = true
											return
										}
										shortFound = shortFound
									}()
									func() {
										if match {
											shortTakesValue = !(len(option.valueName) == 0)
											return
										}
										shortTakesValue = shortTakesValue
									}()
									func() {
										if match {
											shortOptionName = option.name
											return
										}
										shortOptionName = shortOptionName
									}()
								}()
							}
						}()
						unknownShort := active && !(shortHelp) && !(shortFound)
						func() {
							if unknownShort {
								parseError = "unknown option -" + short
								return
							}
							parseError = parseError
						}()
						func() {
							if unknownShort {
								clusterDone = true
								return
							}
							clusterDone = clusterDone
						}()
						shortValue := func() string { runes := []rune(arg); return string(runes[shortIndex+2 : len([]rune(arg))]) }()
						missingShortValue := active && !(shortHelp) && shortFound && shortTakesValue && len(shortValue) == 0 && index+1 >= len(args)
						func() {
							if missingShortValue {
								parseError = "missing value for -" + short
								return
							}
							parseError = parseError
						}()
						func() {
							if missingShortValue {
								clusterDone = true
								return
							}
							clusterDone = clusterDone
						}()
						useShortNext := active && !(shortHelp) && shortFound && shortTakesValue && len(shortValue) == 0 && len(parseError) == 0
						func() {
							if useShortNext {
								shortValue = args[index+1]
								return
							}
							shortValue = shortValue
						}()
						func() {
							if useShortNext {
								skipNext = true
								return
							}
							skipNext = skipNext
						}()
						storeShortValue := active && !(shortHelp) && shortFound && shortTakesValue && len(parseError) == 0
						func() map[string]string {
							if storeShortValue {
								return func() map[string]string { values[shortOptionName] = shortValue; return values }()
							}
							return values
						}()
						func() int {
							if storeShortValue {
								return func() int { explicitOptions = append(explicitOptions, shortOptionName); return len(explicitOptions) }()
							}
							return 0
						}()
						func() {
							if storeShortValue {
								clusterDone = true
								return
							}
							clusterDone = clusterDone
						}()
						storeShortFlag := active && !(shortHelp) && shortFound && !(shortTakesValue) && len(parseError) == 0
						func() map[string]bool {
							if storeShortFlag {
								return func() map[string]bool { flags[shortOptionName] = true; return flags }()
							}
							return flags
						}()
						return func() int {
							if storeShortFlag {
								return func() int { explicitOptions = append(explicitOptions, shortOptionName); return len(explicitOptions) }()
							}
							return 0
						}()
					}()
				}
			}()
			func() {
				if isShort {
					handled = true
					return
				}
				handled = handled
			}()
			positional := !(handled) && len(parseError) == 0
			return func() int {
				if positional {
					return func() int { positionalValues = append(positionalValues, arg); return len(positionalValues) }()
				}
				return 0
			}()
		}()
	}
	for _, option := range command.options {
		_ = option
		func() {
			missing := len(parseError) == 0 && !(helpValue) && option.required && !(len(option.valueName) == 0) && !(func() bool { _, ok := values[option.name]; return ok }())
			func() {
				if missing {
					parseError = "missing required option --" + option.name
					return
				}
				parseError = parseError
			}()
		}()
	}
	positionIndex := 0
	for _, argument := range command.arguments {
		_ = argument
		func() {
			hasValue := positionIndex < len(positionalValues)
			func() map[string]string {
				if hasValue {
					return func() map[string]string {
						positionals[argument.name] = positionalValues[positionIndex]
						return positionals
					}()
				}
				return positionals
			}()
			positionIndex = positionIndex + 1
		}()
	}
	for _, argument := range command.arguments {
		_ = argument
		func() {
			missing := len(parseError) == 0 && !(helpValue) && argument.required && !(func() bool { _, ok := positionals[argument.name]; return ok }())
			func() {
				if missing {
					parseError = "missing required argument " + argument.name
					return
				}
				parseError = parseError
			}()
		}()
	}
	unexpected := len(parseError) == 0 && !(helpValue) && len(positionalValues) > len(command.arguments)
	func() {
		if unexpected {
			parseError = "unexpected argument " + positionalValues[len(command.arguments)]
			return
		}
		parseError = parseError
	}()
	return CliParseResult{command: command, values: values, flags: flags, positionals: positionals, explicitOptions: explicitOptions, args: args, rest: rest, help: helpValue, error_: func() any {
		if len(parseError) == 0 {
			return any(nil)
		}
		return parseError
	}()}
}

func cli_parseCommandArgs(root CliCommand, commands []CliCommand, aliases []CliCommandAlias, trailingRest []string, args []string) CliCommandParseResult {
	split := cli_splitCommandArgs(root, aliases, args, 0, []string{}, "", []string{}, false)
	rootResult := cli_parseArgs(root, split.rootArgs)
	return func() CliCommandParseResult {
		if len(split.commandName) == 0 {
			return CliCommandParseResult{root: rootResult, command: rootResult, commandName: "", commandArgs: []string{}, error_: any(nil)}
		}
		return cli_parseNamedCommandArgs(rootResult, commands, trailingRest, split)
	}()
}

func cli_parseKnownCommandArgs(rootResult CliParseResult, command CliCommand, trailingRest []string, split CliCommandArgs) CliCommandParseResult {
	args := func() []string {
		if cli_contains(trailingRest, split.commandName) {
			return cli_normalizeTrailingRestArgs(command, split.commandArgs)
		}
		return split.commandArgs
	}()
	parsed := cli_withReportedArgs(cli_parseArgs(command, args), split.commandArgs)
	return CliCommandParseResult{root: rootResult, command: parsed, commandName: split.commandName, commandArgs: split.commandArgs, error_: any(nil)}
}

func cli_parseNamedCommandArgs(rootResult CliParseResult, commands []CliCommand, trailingRest []string, split CliCommandArgs) CliCommandParseResult {
	lookup := cli_findCommand(commands, split.commandName, 0)
	return func() CliCommandParseResult {
		if lookup.found {
			return cli_parseKnownCommandArgs(rootResult, lookup.command, trailingRest, split)
		}
		return CliCommandParseResult{root: rootResult, command: rootResult, commandName: split.commandName, commandArgs: split.commandArgs, error_: "unknown command " + split.commandName}
	}()
}

func cli_positionalIncrement(arg string) int {
	return func() int {
		if strings.HasPrefix(arg, "-") {
			return 0
		}
		return 1
	}()
}

func cli_resolveAlias(aliases []CliCommandAlias, name string, index int) string {
	for {
		if index >= len(aliases) {
			return name
		} else {
			if aliases[index].from == name {
				return aliases[index].to
			} else {
				aliases, name, index = aliases, name, index+1
				continue
			}
		}
	}
}

func cli_rootOption(root CliCommand, arg string) CliOption {
	return func() CliOption {
		if strings.HasPrefix(arg, "--") && len([]rune(arg)) > 2 {
			return cli_findOptionByName(root.options, cli_longOptionName(arg), 0)
		}
		return func() CliOption {
			if strings.HasPrefix(arg, "-") && arg != "-" {
				return cli_findOptionByShort(root.options, func() string { runes := []rune(arg); return string(runes[1:2]) }(), 0)
			}
			return cli_emptyOption()
		}()
	}()
}

func cli_rootOptionArg(root CliCommand, arg string) bool {
	return cli_optionFound(cli_rootOption(root, arg))
}

func cli_rootOptionConsumesNext(root CliCommand, arg string) bool {
	option := cli_rootOption(root, arg)
	hasInline := strings.Contains(arg, "=") || strings.HasPrefix(arg, "-") && !(strings.HasPrefix(arg, "--")) && len([]rune(arg)) > 2
	return cli_optionFound(option) && (option.valueName != "" && !(hasInline))
}

func cli_splitCommandArg(root CliCommand, aliases []CliCommandAlias, args []string, index int, rootArgs []string, commandName string, commandArgs []string, skipNext bool) CliCommandArgs {
	return func() CliCommandArgs {
		if skipNext {
			return cli_splitCommandArgs(root, aliases, args, index+1, rootArgs, commandName, commandArgs, false)
		}
		return cli_splitCommandArgValue(root, aliases, args, index, rootArgs, commandName, commandArgs)
	}()
}

func cli_splitCommandArgValue(root CliCommand, aliases []CliCommandAlias, args []string, index int, rootArgs []string, commandName string, commandArgs []string) CliCommandArgs {
	arg := args[index]
	isRootHelp := commandName == "" && (arg == "--help" || arg == "-h")
	isRootOption := cli_rootOptionArg(root, arg)
	return func() CliCommandArgs {
		if isRootHelp {
			return cli_splitCommandArgs(root, aliases, args, index+1, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, rootArgs...)
				__rune_spread_out = append(__rune_spread_out, arg)
				return __rune_spread_out
			}(), commandName, commandArgs, false)
		}
		return func() CliCommandArgs {
			if isRootOption {
				return cli_splitRootOptionArg(root, aliases, args, index, rootArgs, commandName, commandArgs, cli_rootOptionConsumesNext(root, arg))
			}
			return func() CliCommandArgs {
				if len(commandName) == 0 {
					return cli_splitCommandArgs(root, aliases, args, index+1, rootArgs, cli_resolveAlias(aliases, arg, 0), commandArgs, false)
				}
				return cli_splitCommandArgs(root, aliases, args, index+1, rootArgs, commandName, func() []string {
					__rune_spread_out := []string{}
					__rune_spread_out = append(__rune_spread_out, commandArgs...)
					__rune_spread_out = append(__rune_spread_out, arg)
					return __rune_spread_out
				}(), false)
			}()
		}()
	}()
}

func cli_splitCommandArgs(root CliCommand, aliases []CliCommandAlias, args []string, index int, rootArgs []string, commandName string, commandArgs []string, skipNext bool) CliCommandArgs {
	return func() CliCommandArgs {
		if index >= len(args) {
			return CliCommandArgs{rootArgs: rootArgs, commandName: commandName, commandArgs: commandArgs}
		}
		return cli_splitCommandArg(root, aliases, args, index, rootArgs, commandName, commandArgs, skipNext)
	}()
}

func cli_splitRootOptionArg(root CliCommand, aliases []CliCommandAlias, args []string, index int, rootArgs []string, commandName string, commandArgs []string, consumesNext bool) CliCommandArgs {
	return cli_splitCommandArgs(root, aliases, args, index+1, cli_appendRootOptionArgs(rootArgs, args, index, consumesNext), commandName, commandArgs, consumesNext)
}

func cli_withAliases(command CliCommand, aliases []CliCommandAlias) CliCommand {
	return CliCommand{name: command.name, version: command.version, about: command.about, options: command.options, arguments: command.arguments, commands: command.commands, aliases: aliases}
}

func cli_withArgument(command CliCommand, argument CliArgument) CliCommand {
	return CliCommand{name: command.name, version: command.version, about: command.about, options: command.options, commands: command.commands, aliases: command.aliases, arguments: func() []CliArgument {
		__rune_spread_out := []CliArgument{}
		__rune_spread_out = append(__rune_spread_out, command.arguments...)
		__rune_spread_out = append(__rune_spread_out, argument)
		return __rune_spread_out
	}()}
}

func cli_withCommands(command CliCommand, commands []CliCommand) CliCommand {
	return CliCommand{name: command.name, version: command.version, about: command.about, options: command.options, arguments: command.arguments, aliases: command.aliases, commands: commands}
}

func cli_withOption(command CliCommand, option CliOption) CliCommand {
	return CliCommand{name: command.name, version: command.version, about: command.about, arguments: command.arguments, commands: command.commands, aliases: command.aliases, options: func() []CliOption {
		__rune_spread_out := []CliOption{}
		__rune_spread_out = append(__rune_spread_out, command.options...)
		__rune_spread_out = append(__rune_spread_out, option)
		return __rune_spread_out
	}()}
}

func cli_withReportedArgs(result CliParseResult, args []string) CliParseResult {
	return CliParseResult{command: result.command, values: result.values, flags: result.flags, positionals: result.positionals, explicitOptions: result.explicitOptions, rest: result.rest, help: result.help, error_: result.error_, args: args}
}

func parseCli(args []string) RuneCliInvocation {
	parsed := cli_parseCommandArgs(runeCommand(), runeCommands(), runeAliases(), runeTrailingRest(), args)
	return selfhost_cli_cli_invocationFromParsed(parsed, runeCommand(), runeCommands())
}

func selfhost_cli_cli_invocationFromParsed(parsed CliCommandParseResult, rootCommand CliCommand, commands []CliCommand) RuneCliInvocation {
	root := parsed.root
	backend := func() string {
		value, ok := root.values["backend"]
		if ok {
			return value
		}
		return "go"
	}()
	errors := selfhost_cli_cli_cliErrors(root)
	commandError := func() string {
		coalesce4 := parsed.error_
		if coalesce4 != nil {
			return coalesce4.(string)
		}
		return ""
	}()
	func() int {
		if len(commandError) == 0 {
			return 0
		}
		return func() int { errors = append(errors, commandError); return len(errors) }()
	}()
	invalidBackend := cli_contains([]string{"go", "ts", "mbt"}, backend) == false
	func() int {
		if invalidBackend {
			return func() int { errors = append(errors, "unsupported backend "+backend); return len(errors) }()
		}
		return 0
	}()
	backendExplicit := cli_contains(root.explicitOptions, "backend")
	return func() RuneCliInvocation {
		if len(errors) > 0 {
			return selfhost_cli_cli_errorInvocation(backend, errors)
		}
		return func() RuneCliInvocation {
			if len(parsed.commandName) == 0 {
				return selfhost_cli_cli_helpInvocation(rootCommand, backend, backendExplicit)
			}
			return selfhost_cli_cli_invocationFromResult(parsed.commandName, rootCommand, commands, backend, backendExplicit, parsed.command)
		}()
	}()
}

func selfhost_cli_cli_invocationFromResult(name string, rootCommand CliCommand, commands []CliCommand, backend string, backendExplicit bool, result CliParseResult) RuneCliInvocation {
	errors := selfhost_cli_cli_cliErrors(result)
	target := func() string {
		value, ok := result.values["target"]
		if ok {
			return value
		}
		return ""
	}()
	target = selfhost_cli_cli_defaultTargetForCommand(name, backend, target)
	errors = selfhost_cli_cli_invocationErrors(name, backend, target, errors)
	return RuneCliInvocation{ok: len(errors) == 0, command: name, backend: backend, path: selfhost_cli_cli_defaultPathForCommand(name, func() string {
		value, ok := result.positionals["path"]
		if ok {
			return value
		}
		return ""
	}()), output: func() string {
		value, ok := result.values["output"]
		if ok {
			return value
		}
		return ""
	}(), target: target, pattern: func() string {
		value, ok := result.positionals["pattern"]
		if ok {
			return value
		}
		return ""
	}(), checkOnly: func() bool {
		value, ok := result.flags["check"]
		if ok {
			return value
		}
		return false
	}(), stdout: func() bool {
		value, ok := result.flags["stdout"]
		if ok {
			return value
		}
		return false
	}(), backendExplicit: backendExplicit, runArgs: result.rest, errors: errors, help: result.help, helpText: cli_help(selfhost_cli_cli_invocationHelpCommand(rootCommand, commands, name))}
}

func selfhost_cli_cli_invocationErrors(name string, backend string, target string, errors []string) []string {
	return func() []string {
		if name == "build" && cli_contains([]string{"go", "mbt"}, backend) == false {
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "rune build only supports --backend go or --backend mbt")
				return __rune_spread_out
			}()
		}
		return func() []string {
			if name == "run" && len(target) == 0 == false && backend != "mbt" {
				return func() []string {
					__rune_spread_out := []string{}
					__rune_spread_out = append(__rune_spread_out, errors...)
					__rune_spread_out = append(__rune_spread_out, "rune run --target is only supported with --backend mbt")
					return __rune_spread_out
				}()
			}
			return errors
		}()
	}()
}

func selfhost_cli_cli_defaultTargetForCommand(name string, backend string, target string) string {
	return func() string {
		if name == "run" && backend == "mbt" && len(target) == 0 {
			return "native"
		}
		return target
	}()
}

func selfhost_cli_cli_defaultPathForCommand(name string, path string) string {
	return func() string {
		if name == "test" && len(path) == 0 {
			return "tests"
		}
		return path
	}()
}

func selfhost_cli_cli_invocationHelpCommand(rootCommand CliCommand, commands []CliCommand, name string) CliCommand {
	return func() CliCommand {
		array5 := commands
		result6 := rootCommand
		for _, value8 := range array5 {
			result6 = func(found CliCommand, command CliCommand) CliCommand {
				return func() CliCommand {
					if command.name == name {
						return command
					}
					return found
				}()
			}(result6, value8)
		}
		return result6
	}()
}

func selfhost_cli_cli_cliErrors(result CliParseResult) []string {
	error_ := func() string {
		coalesce9 := result.error_
		if coalesce9 != nil {
			return coalesce9.(string)
		}
		return ""
	}()
	errors := []string{}
	func() int {
		if len(error_) == 0 {
			return 0
		}
		return func() int { errors = append(errors, error_); return len(errors) }()
	}()
	return errors
}

func selfhost_cli_cli_helpInvocation(rootCommand CliCommand, backend string, backendExplicit bool) RuneCliInvocation {
	return RuneCliInvocation{ok: true, command: "", backend: backend, path: "", output: "", target: "", pattern: "", checkOnly: false, stdout: false, backendExplicit: backendExplicit, runArgs: []string{}, errors: []string{}, help: true, helpText: cli_help(rootCommand)}
}

func selfhost_cli_cli_errorInvocation(backend string, errors []string) RuneCliInvocation {
	return RuneCliInvocation{ok: false, command: "", backend: backend, path: "", output: "", target: "", pattern: "", checkOnly: false, stdout: false, backendExplicit: false, runArgs: []string{}, errors: errors, help: false, helpText: ""}
}

func selfhost_cli_cli_emptyInvocation() RuneCliInvocation {
	return selfhost_cli_cli_errorInvocation("", []string{})
}

func runeCommand() CliCommand {
	command := cli_command("rune", "Rune language toolchain")
	command = cli_withOption(command, cli_option("backend", "b", "BACKEND", "target backend", false, "go"))
	command = cli_withCommands(command, runeCommands())
	return cli_withAliases(command, runeAliases())
}

func runeCommands() []CliCommand {
	return []CliCommand{runCommand(), buildCommand(), emitCommand("go"), emitCommand("ts"), emitCommand("dts"), emitCommand("mbt"), singlePathCommand("check"), fmtCommand(), testCommand(), cli_command("repl", ""), lspCommand()}
}

func runeAliases() []CliCommandAlias {
	return []CliCommandAlias{cli_alias("format", "fmt")}
}

func runeTrailingRest() []string {
	return []string{"run"}
}

func runCommand() CliCommand {
	command := cli_command("run", "Compile and run a Rune program")
	command = cli_withOption(command, cli_option("target", "", "TARGET", "MoonBit run target", false, any(nil)))
	return cli_withArgument(command, cli_argument("path", "Rune source path", true))
}

func buildCommand() CliCommand {
	command := cli_command("build", "Compile a Rune program to an executable")
	command = cli_withOption(command, cli_option("output", "o", "FILE", "output executable path", false, ""))
	command = cli_withOption(command, cli_option("target", "", "TARGET", "build target", false, ""))
	return cli_withArgument(command, cli_argument("path", "Rune source path", true))
}

func emitCommand(name string) CliCommand {
	command := cli_command(name, "Compile a Rune program")
	command = cli_withOption(command, cli_option("output", "o", "FILE", "output file", false, ""))
	return cli_withArgument(command, cli_argument("path", "Rune source path", true))
}

func singlePathCommand(name string) CliCommand {
	command := cli_command(name, "Parse and type-check Rune source")
	return cli_withArgument(command, cli_argument("path", "Rune source path", true))
}

func fmtCommand() CliCommand {
	command := cli_command("fmt", "Format Rune source")
	command = cli_withOption(command, cli_flag("check", "", "fail if not formatted"))
	command = cli_withOption(command, cli_flag("stdout", "", "write formatted source to stdout"))
	return cli_withArgument(command, cli_argument("path", "Rune source path", true))
}

func testCommand() CliCommand {
	command := cli_command("test", "Run Rune tests")
	command = cli_withArgument(command, cli_argument("path", "Rune test path", false))
	return cli_withArgument(command, cli_argument("pattern", "test name pattern", false))
}

func lspCommand() CliCommand {
	command := cli_command("lsp", "Start the Rune language server")
	return cli_withOption(command, cli_flag("stdio", "", "serve LSP over stdin/stdout"))
}

func main_() {
	argv := runeProcessArgv()
	invocation := parseCli(append([]string{}, argv[1:len(argv)]...))
	func() {
		if invocation.help {
			fmt.Print(cli_help(runeCommand()))
			return
		}
		fmt.Println(invocation.command)
	}()
}
