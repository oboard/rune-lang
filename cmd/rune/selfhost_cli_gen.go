package main

import (
	"fmt"
	"os"
	"strings"
)

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

type __RuneCliExecution struct {
	__ok     bool
	__output string
	__errors []string
}

type __RuneCliRootArgs struct {
	__globalArgs      []string
	__commandArgs     []string
	__backendExplicit bool
}

type __RuneCliArgsEntry struct {
	__backend string
	__command string
}

type __CliCommand struct {
	__name      string
	__version   any
	__about     string
	__options   []__CliOption
	__arguments []__CliArgument
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
	__command     __CliCommand
	__values      map[string]string
	__flags       map[string]bool
	__positionals map[string]string
	__args        []string
	__rest        []string
	__help        bool
	__error       any
}

func runeCliCommand(name string, about string) __CliCommand {
	return __CliCommand{__name: name, __version: any(nil), __about: about, __options: []__CliOption{}, __arguments: []__CliArgument{}}
}

func runeCliWithVersion(command __CliCommand, version string) __CliCommand {
	command.__version = version
	return command
}

func runeCliWithOption(command __CliCommand, option __CliOption) __CliCommand {
	command.__options = append(append([]__CliOption{}, command.__options...), option)
	return command
}

func runeCliWithArgument(command __CliCommand, argument __CliArgument) __CliCommand {
	command.__arguments = append(append([]__CliArgument{}, command.__arguments...), argument)
	return command
}

func runeCliFlag(name string, short string, help string) __CliOption {
	return __CliOption{__name: name, __short: short, __help: help}
}

func runeCliOption(name string, short string, valueName string, help string, required bool, defaultValue any) __CliOption {
	return __CliOption{__name: name, __short: short, __valueName: valueName, __help: help, __required: required, __defaultValue: defaultValue}
}

func runeCliArgument(name string, help string, required bool) __CliArgument {
	return __CliArgument{__name: name, __help: help, __required: required}
}

func runeCliParse(command __CliCommand) __CliParseResult {
	return runeCliParseArgs(command, append([]string(nil), os.Args[1:]...))
}

func runeCliParseArgs(command __CliCommand, args []string) __CliParseResult {
	longOptions := map[string]__CliOption{}
	shortOptions := map[string]__CliOption{}
	values := map[string]string{}
	flags := map[string]bool{}
	positionals := map[string]string{}
	positionalValues := []string{}
	rest := []string{}
	help := false
	parseError := ""

	for _, option := range command.__options {
		longOptions[option.__name] = option
		if option.__short != "" {
			shortOptions[option.__short] = option
		}
		if option.__valueName != "" && option.__defaultValue != nil {
			if defaultValue, ok := option.__defaultValue.(string); ok {
				values[option.__name] = defaultValue
			}
		}
	}

	for idx := 0; idx < len(args); idx++ {
		arg := args[idx]
		if arg == "--" {
			rest = append(rest, args[idx+1:]...)
			break
		}
		if arg == "--help" || arg == "-h" {
			help = true
			continue
		}
		if strings.HasPrefix(arg, "--") && len(arg) > 2 {
			name := arg[2:]
			value := ""
			hasValue := false
			if before, after, ok := strings.Cut(name, "="); ok {
				name = before
				value = after
				hasValue = true
			}
			if strings.HasPrefix(name, "no-") {
				if option, ok := longOptions[strings.TrimPrefix(name, "no-")]; ok && option.__valueName == "" {
					flags[option.__name] = false
					continue
				}
			}
			option, ok := longOptions[name]
			if !ok {
				parseError = "unknown option --" + name
				break
			}
			if option.__valueName != "" {
				if !hasValue {
					idx++
					if idx >= len(args) {
						parseError = "missing value for --" + name
						break
					}
					value = args[idx]
				}
				values[option.__name] = value
			} else {
				flags[option.__name] = !hasValue || value != "false"
			}
			continue
		}
		if strings.HasPrefix(arg, "-") && arg != "-" {
			cluster := []rune(arg[1:])
			for pos := 0; pos < len(cluster); pos++ {
				short := string(cluster[pos])
				if short == "h" {
					help = true
					continue
				}
				option, ok := shortOptions[short]
				if !ok {
					parseError = "unknown option -" + short
					break
				}
				if option.__valueName != "" {
					value := string(cluster[pos+1:])
					if value == "" {
						idx++
						if idx >= len(args) {
							parseError = "missing value for -" + short
							break
						}
						value = args[idx]
					}
					values[option.__name] = value
					break
				}
				flags[option.__name] = true
			}
			if parseError != "" {
				break
			}
			continue
		}
		positionalValues = append(positionalValues, arg)
	}

	if parseError == "" && !help {
		for _, option := range command.__options {
			if option.__required && option.__valueName != "" {
				if _, ok := values[option.__name]; !ok {
					parseError = "missing required option --" + option.__name
					break
				}
			}
		}
	}
	if parseError == "" && !help {
		for idx, argument := range command.__arguments {
			if idx < len(positionalValues) {
				positionals[argument.__name] = positionalValues[idx]
				continue
			}
			if argument.__required {
				parseError = "missing required argument " + argument.__name
				break
			}
		}
		if parseError == "" && len(positionalValues) > len(command.__arguments) {
			parseError = "unexpected argument " + positionalValues[len(command.__arguments)]
		}
	} else {
		for idx, argument := range command.__arguments {
			if idx < len(positionalValues) {
				positionals[argument.__name] = positionalValues[idx]
			}
		}
	}

	errorValue := any(nil)
	if parseError != "" {
		errorValue = parseError
	}
	return __CliParseResult{
		__command:     command,
		__values:      values,
		__flags:       flags,
		__positionals: positionals,
		__args:        append([]string(nil), args...),
		__rest:        rest,
		__help:        help,
		__error:       errorValue,
	}
}

func runeCliHelp(command __CliCommand) string {
	var b strings.Builder
	b.WriteString("Usage: ")
	b.WriteString(command.__name)
	if len(command.__options) > 0 {
		b.WriteString(" [options]")
	}
	for _, argument := range command.__arguments {
		b.WriteByte(' ')
		if argument.__required {
			b.WriteByte('<')
			b.WriteString(argument.__name)
			b.WriteByte('>')
		} else {
			b.WriteByte('[')
			b.WriteString(argument.__name)
			b.WriteByte(']')
		}
	}
	b.WriteByte('\n')
	if version, ok := command.__version.(string); ok && version != "" {
		b.WriteString("Version: ")
		b.WriteString(version)
		b.WriteByte('\n')
	}
	if command.__about != "" {
		b.WriteByte('\n')
		b.WriteString(command.__about)
		b.WriteByte('\n')
	}
	if len(command.__arguments) > 0 {
		b.WriteString("\nArguments:\n")
		for _, argument := range command.__arguments {
			b.WriteString("  ")
			b.WriteString(argument.__name)
			if argument.__help != "" {
				b.WriteString("\t")
				b.WriteString(argument.__help)
			}
			b.WriteByte('\n')
		}
	}
	b.WriteString("\nOptions:\n")
	for _, option := range command.__options {
		b.WriteString("  ")
		if option.__short != "" {
			b.WriteByte('-')
			b.WriteString(option.__short)
			b.WriteString(", ")
		}
		b.WriteString("--")
		b.WriteString(option.__name)
		if option.__valueName != "" {
			b.WriteString(" <")
			b.WriteString(option.__valueName)
			b.WriteByte('>')
		}
		if option.__help != "" {
			b.WriteString("\t")
			b.WriteString(option.__help)
		}
		if option.__required {
			b.WriteString(" (required)")
		}
		if defaultValue, ok := option.__defaultValue.(string); ok && defaultValue != "" {
			b.WriteString(" (default: ")
			b.WriteString(defaultValue)
			b.WriteByte(')')
		}
		b.WriteByte('\n')
	}
	b.WriteString("  -h, --help\tShow help\n")
	return b.String()
}

func runeCliParseOrExit(command __CliCommand) __CliParseResult {
	result := runeCliParse(command)
	if result.__help {
		fmt.Print(runeCliHelp(command))
		os.Exit(0)
	}
	if message, ok := result.__error.(string); ok && message != "" {
		fmt.Println(message)
		fmt.Print(runeCliHelp(command))
		os.Exit(2)
	}
	return result
}

func runeCliContains(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func __parseCli(__args []string) __RuneCliInvocation {
	__split := ____rune_private_5b8b8d7b_splitRootArgs(__args)
	__root := ____rune_private_5b8b8d7b_parseRoot(__split.__globalArgs)
	__backend := func() string {
		value, ok := __root.__values["backend"]
		if ok {
			return value
		}
		return "go"
	}()
	__errors := ____rune_private_5b8b8d7b_cliErrors(__root)
	__invalidBackend := runeCliContains([]string{"go", "ts", "mbt"}, __backend) == false
	func() int {
		if __invalidBackend {
			return func() int { __errors = append(__errors, "unsupported backend "+__backend); return len(__errors) }()
		}
		return 0
	}()
	return func() __RuneCliInvocation {
		if len(__errors) > 0 {
			return ____rune_private_5b8b8d7b_errorInvocation("", __backend, __errors)
		}
		return ____rune_private_5b8b8d7b_parseCommandAt(__split.__commandArgs, 0, __backend, __split.__backendExplicit)
	}()
}

func __runeCommand() __CliCommand {
	__command := runeCliCommand("rune", "Rune language toolchain")
	__command = runeCliWithOption(__command, runeCliOption("backend", "b", "BACKEND", "target backend", false, "go"))
	return __command
}

func ____rune_private_5b8b8d7b_parseRoot(__args []string) __CliParseResult {
	__command := __runeCommand()
	return runeCliParseArgs(__command, __args)
}

func ____rune_private_5b8b8d7b_splitRootArgs(__args []string) __RuneCliRootArgs {
	return ____rune_private_5b8b8d7b_splitRootArgsAt(__args, 0, []string{}, []string{}, false, false, false)
}

func ____rune_private_5b8b8d7b_splitRootArgsAt(__args []string, __position int, __globalArgs []string, __commandArgs []string, __backendExplicit bool, __skipNext bool, __inRest bool) __RuneCliRootArgs {
	return func() __RuneCliRootArgs {
		if __position >= len(__args) {
			return __RuneCliRootArgs{__globalArgs: __globalArgs, __commandArgs: __commandArgs, __backendExplicit: __backendExplicit}
		}
		return ____rune_private_5b8b8d7b_splitRootArgAt(__args, __position, __globalArgs, __commandArgs, __backendExplicit, __skipNext, __inRest)
	}()
}

func ____rune_private_5b8b8d7b_splitRootArgAt(__args []string, __position int, __globalArgs []string, __commandArgs []string, __backendExplicit bool, __skipNext bool, __inRest bool) __RuneCliRootArgs {
	return func() __RuneCliRootArgs {
		switch {
		case __skipNext == true:
			return ____rune_private_5b8b8d7b_splitRootArgsAt(__args, __position+1, __globalArgs, __commandArgs, __backendExplicit, false, __inRest)
		default:
			return ____rune_private_5b8b8d7b_splitRootArgValue(__args, __position, __globalArgs, __commandArgs, __backendExplicit, __inRest)
		}
	}()
}

func ____rune_private_5b8b8d7b_splitRootArgValue(__args []string, __position int, __globalArgs []string, __commandArgs []string, __backendExplicit bool, __inRest bool) __RuneCliRootArgs {
	__arg := __args[__position]
	__backendPair := ____rune_private_5b8b8d7b_isRootBackendPair(__inRest, __arg)
	__hasBackendValue := __backendPair && __position+1 < len(__args)
	__backendInline := ____rune_private_5b8b8d7b_isRootBackendInline(__inRest, __arg)
	__rootHelp := ____rune_private_5b8b8d7b_isRootHelpArg(__inRest, len(__commandArgs), __arg)
	return func() __RuneCliRootArgs {
		if __backendPair {
			return ____rune_private_5b8b8d7b_splitRootAfterBackendPair(__args, __position, __globalArgs, __commandArgs, __hasBackendValue, __inRest)
		}
		return func() __RuneCliRootArgs {
			if __backendInline {
				return ____rune_private_5b8b8d7b_splitRootAfterBackendInline(__args, __position, __globalArgs, __commandArgs, __inRest)
			}
			return func() __RuneCliRootArgs {
				if __rootHelp {
					return ____rune_private_5b8b8d7b_splitRootAfterGlobal(__args, __position, __globalArgs, __commandArgs, __backendExplicit, __inRest)
				}
				return ____rune_private_5b8b8d7b_splitRootAfterCommand(__args, __position, __globalArgs, __commandArgs, __backendExplicit, __inRest)
			}()
		}()
	}()
}

func ____rune_private_5b8b8d7b_isRootBackendPair(__inRest bool, __arg string) bool {
	return __inRest == false && runeCliContains([]string{"--backend", "-b"}, __arg)
}

func ____rune_private_5b8b8d7b_isRootBackendInline(__inRest bool, __arg string) bool {
	return __inRest == false && (strings.HasPrefix(__arg, "--backend=") || strings.HasPrefix(__arg, "-b") && len([]rune(__arg)) > 2)
}

func ____rune_private_5b8b8d7b_isRootHelpArg(__inRest bool, __commandCount int, __arg string) bool {
	return __inRest == false && __commandCount == 0 && runeCliContains([]string{"--help", "-h"}, __arg)
}

func ____rune_private_5b8b8d7b_splitRootAfterBackendPair(__args []string, __position int, __globalArgs []string, __commandArgs []string, __hasValue bool, __inRest bool) __RuneCliRootArgs {
	__globalArgs = append(__globalArgs, __args[__position])
	func() int {
		if __hasValue {
			return func() int { __globalArgs = append(__globalArgs, __args[__position+1]); return len(__globalArgs) }()
		}
		return 0
	}()
	return ____rune_private_5b8b8d7b_splitRootArgsAt(__args, __position+1, __globalArgs, __commandArgs, true, __hasValue, __inRest)
}

func ____rune_private_5b8b8d7b_splitRootAfterBackendInline(__args []string, __position int, __globalArgs []string, __commandArgs []string, __inRest bool) __RuneCliRootArgs {
	__globalArgs = append(__globalArgs, __args[__position])
	return ____rune_private_5b8b8d7b_splitRootArgsAt(__args, __position+1, __globalArgs, __commandArgs, true, false, __inRest)
}

func ____rune_private_5b8b8d7b_splitRootAfterGlobal(__args []string, __position int, __globalArgs []string, __commandArgs []string, __backendExplicit bool, __inRest bool) __RuneCliRootArgs {
	__globalArgs = append(__globalArgs, __args[__position])
	return ____rune_private_5b8b8d7b_splitRootArgsAt(__args, __position+1, __globalArgs, __commandArgs, __backendExplicit, false, __inRest)
}

func ____rune_private_5b8b8d7b_splitRootAfterCommand(__args []string, __position int, __globalArgs []string, __commandArgs []string, __backendExplicit bool, __inRest bool) __RuneCliRootArgs {
	__commandArgs = append(__commandArgs, __args[__position])
	return ____rune_private_5b8b8d7b_splitRootArgsAt(__args, __position+1, __globalArgs, __commandArgs, __backendExplicit, false, __inRest || __args[__position] == "--")
}

func ____rune_private_5b8b8d7b_parseCommandAt(__args []string, __index int, __backend string, __backendExplicit bool) __RuneCliInvocation {
	return func() __RuneCliInvocation {
		if __index >= len(__args) {
			return ____rune_private_5b8b8d7b_helpInvocation(__backend)
		}
		return ____rune_private_5b8b8d7b_dispatchCommand(__args[__index], append([]string{}, __args[__index+1:len(__args)]...), __backend, __backendExplicit)
	}()
}

func ____rune_private_5b8b8d7b_dispatchCommand(__command string, __args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	return func() __RuneCliInvocation {
		switch {
		case __command == "run":
			return ____rune_private_5b8b8d7b_parseRun(__args, __backend, __backendExplicit)
		case __command == "build":
			return ____rune_private_5b8b8d7b_parseBuild(__args, __backend, __backendExplicit)
		case __command == "check":
			return ____rune_private_5b8b8d7b_parseSinglePath(__command, __args, __backend, __backendExplicit)
		case (__command == "fmt") || (__command == "format"):
			return ____rune_private_5b8b8d7b_parseFmt(__args, __backend, __backendExplicit)
		case __command == "test":
			return ____rune_private_5b8b8d7b_parseTest(__args, __backend, __backendExplicit)
		case (__command == "ts") || (__command == "go") || (__command == "mbt"):
			return ____rune_private_5b8b8d7b_parseEmit(__command, __args, __backend, __backendExplicit)
		case __command == "repl":
			return ____rune_private_5b8b8d7b_parseNoArgs(__command, __args, __backend, __backendExplicit)
		case __command == "lsp":
			return ____rune_private_5b8b8d7b_parseLsp(__args, __backend, __backendExplicit)
		default:
			return ____rune_private_5b8b8d7b_errorInvocation(__command, __backend, []string{"unknown command " + __command})
		}
	}()
}

func ____rune_private_5b8b8d7b_parseRun(__args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	__result := runeCliParseArgs(__runCommand(), ____rune_private_5b8b8d7b_normalizeRunArgs(__args))
	return ____rune_private_5b8b8d7b_invocationFromResult("run", __backend, __backendExplicit, __result, __result.__rest, false, false)
}

func ____rune_private_5b8b8d7b_normalizeRunArgs(__args []string) []string {
	return ____rune_private_5b8b8d7b_normalizeRunArgsAt(__args, 0, []string{}, false, false)
}

func ____rune_private_5b8b8d7b_normalizeRunArgsAt(__args []string, __index int, __out []string, __seenPath bool, __skipValue bool) []string {
	return func() []string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_5b8b8d7b_normalizeRunArgAt(__args, __index, __out, __seenPath, __skipValue)
	}()
}

func ____rune_private_5b8b8d7b_normalizeRunArgAt(__args []string, __index int, __out []string, __seenPath bool, __skipValue bool) []string {
	__arg := __args[__index]
	return func() []string {
		if __seenPath && __arg != "--" {
			return ____rune_private_5b8b8d7b_appendRunRest(__args, __index, func() []string { out := []string{}; out = append(out, __out...); out = append(out, "--"); return out }())
		}
		return func() []string {
			if __arg == "--" {
				return ____rune_private_5b8b8d7b_appendRunRest(__args, __index, __out)
			}
			return ____rune_private_5b8b8d7b_normalizeRunArgBeforeRest(__args, __index, func() []string { out := []string{}; out = append(out, __out...); out = append(out, __arg); return out }(), __seenPath, __skipValue)
		}()
	}()
}

func ____rune_private_5b8b8d7b_normalizeRunArgBeforeRest(__args []string, __index int, __out []string, __seenPath bool, __skipValue bool) []string {
	return func() []string {
		if __skipValue {
			return ____rune_private_5b8b8d7b_normalizeRunArgsAt(__args, __index+1, __out, __seenPath, false)
		}
		return func() []string {
			if ____rune_private_5b8b8d7b_isRunOptionValueArg(__args[__index]) {
				return ____rune_private_5b8b8d7b_normalizeRunArgsAt(__args, __index+1, __out, __seenPath, true)
			}
			return ____rune_private_5b8b8d7b_normalizeRunArgsAt(__args, __index+1, __out, __seenPath || !strings.HasPrefix(__args[__index], "-"), false)
		}()
	}()
}

func ____rune_private_5b8b8d7b_isRunOptionValueArg(__arg string) bool {
	return __arg == "--target"
}

func ____rune_private_5b8b8d7b_appendRunRest(__args []string, __index int, __out []string) []string {
	return func() []string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_5b8b8d7b_appendRunRest(__args, __index+1, func() []string {
			out := []string{}
			out = append(out, __out...)
			out = append(out, __args[__index])
			return out
		}())
	}()
}

func __runCommand() __CliCommand {
	__command := runeCliCommand("run", "Compile and run a Rune program")
	__command = runeCliWithOption(__command, runeCliOption("target", "", "TARGET", "MoonBit run target", false, any(nil)))
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func ____rune_private_5b8b8d7b_parseBuild(__args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	__result := runeCliParseArgs(__buildCommand(), __args)
	return ____rune_private_5b8b8d7b_invocationFromResult("build", __backend, __backendExplicit, __result, []string{}, false, false)
}

func __buildCommand() __CliCommand {
	__command := runeCliCommand("build", "Compile a Rune program to an executable")
	__command = runeCliWithOption(__command, runeCliOption("output", "o", "FILE", "output executable path", false, ""))
	__command = runeCliWithOption(__command, runeCliOption("target", "", "TARGET", "build target", false, ""))
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func ____rune_private_5b8b8d7b_parseEmit(__name string, __args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	__result := runeCliParseArgs(__emitCommand(__name), __args)
	return ____rune_private_5b8b8d7b_invocationFromResult(__name, __backend, __backendExplicit, __result, []string{}, false, false)
}

func __emitCommand(__name string) __CliCommand {
	__command := runeCliCommand(__name, "Compile a Rune program")
	__command = runeCliWithOption(__command, runeCliOption("output", "o", "FILE", "output file", false, ""))
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func ____rune_private_5b8b8d7b_parseSinglePath(__name string, __args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	__result := runeCliParseArgs(__singlePathCommand(__name), __args)
	return ____rune_private_5b8b8d7b_invocationFromResult(__name, __backend, __backendExplicit, __result, []string{}, false, false)
}

func __singlePathCommand(__name string) __CliCommand {
	__command := runeCliCommand(__name, "Parse and type-check Rune source")
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func ____rune_private_5b8b8d7b_parseFmt(__args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	__result := runeCliParseArgs(__fmtCommand(), __args)
	return ____rune_private_5b8b8d7b_invocationFromResult("fmt", __backend, __backendExplicit, __result, []string{}, func() bool {
		value, ok := __result.__flags["check"]
		if ok {
			return value
		}
		return false
	}(), func() bool {
		value, ok := __result.__flags["stdout"]
		if ok {
			return value
		}
		return false
	}())
}

func __fmtCommand() __CliCommand {
	__command := runeCliCommand("fmt", "Format Rune source")
	__command = runeCliWithOption(__command, runeCliFlag("check", "", "fail if not formatted"))
	__command = runeCliWithOption(__command, runeCliFlag("stdout", "", "write formatted source to stdout"))
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func ____rune_private_5b8b8d7b_parseTest(__args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	__result := runeCliParseArgs(__testCommand(), __args)
	return ____rune_private_5b8b8d7b_invocationFromResult("test", __backend, __backendExplicit, __result, []string{}, false, false)
}

func __testCommand() __CliCommand {
	__command := runeCliCommand("test", "Run Rune tests")
	__command = runeCliWithArgument(__command, runeCliArgument("path", "Rune test path", false))
	return runeCliWithArgument(__command, runeCliArgument("pattern", "test name pattern", false))
}

func ____rune_private_5b8b8d7b_parseNoArgs(__name string, __args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	__result := runeCliParseArgs(runeCliCommand(__name, ""), __args)
	return ____rune_private_5b8b8d7b_invocationFromResult(__name, __backend, __backendExplicit, __result, []string{}, false, false)
}

func ____rune_private_5b8b8d7b_parseLsp(__args []string, __backend string, __backendExplicit bool) __RuneCliInvocation {
	__result := runeCliParseArgs(__lspCommand(), __args)
	return ____rune_private_5b8b8d7b_invocationFromResult("lsp", __backend, __backendExplicit, __result, []string{}, false, false)
}

func __lspCommand() __CliCommand {
	__command := runeCliCommand("lsp", "Start the Rune language server")
	return runeCliWithOption(__command, runeCliFlag("stdio", "", "serve LSP over stdin/stdout"))
}

func ____rune_private_5b8b8d7b_invocationFromResult(__name string, __backend string, __backendExplicit bool, __result __CliParseResult, __runArgs []string, __checkOnly bool, __stdout bool) __RuneCliInvocation {
	__errors := ____rune_private_5b8b8d7b_cliErrors(__result)
	__target := func() string {
		value, ok := __result.__values["target"]
		if ok {
			return value
		}
		return ""
	}()
	__errors = ____rune_private_5b8b8d7b_invocationErrors(__name, __backend, __target, __errors)
	return __RuneCliInvocation{__ok: len(__errors) == 0, __command: __name, __backend: __backend, __path: ____rune_private_5b8b8d7b_defaultPathForCommand(__name, func() string {
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
	}(), __checkOnly: __checkOnly, __stdout: __stdout, __backendExplicit: __backendExplicit, __runArgs: __runArgs, __errors: __errors, __help: __result.__help, __helpText: ____rune_private_5b8b8d7b_invocationHelpText(__name)}
}

func ____rune_private_5b8b8d7b_invocationErrors(__name string, __backend string, __target string, __errors []string) []string {
	return func() []string {
		if __name == "build" && runeCliContains([]string{"go", "mbt"}, __backend) == false {
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

func ____rune_private_5b8b8d7b_defaultPathForCommand(__name string, __path string) string {
	return func() string {
		if __name == "test" && len(__path) == 0 {
			return "tests"
		}
		return __path
	}()
}

func ____rune_private_5b8b8d7b_invocationHelpText(__name string) string {
	return func() string {
		switch {
		case __name == "run":
			return runeCliHelp(__runCommand())
		case __name == "build":
			return runeCliHelp(__buildCommand())
		case __name == "go":
			return runeCliHelp(__emitCommand("go"))
		case __name == "ts":
			return runeCliHelp(__emitCommand("ts"))
		case __name == "mbt":
			return runeCliHelp(__emitCommand("mbt"))
		case __name == "check":
			return runeCliHelp(__singlePathCommand("check"))
		case __name == "fmt":
			return runeCliHelp(__fmtCommand())
		case __name == "test":
			return runeCliHelp(__testCommand())
		case __name == "repl":
			return runeCliHelp(runeCliCommand("repl", ""))
		case __name == "lsp":
			return runeCliHelp(__lspCommand())
		default:
			return runeCliHelp(__runeCommand())
		}
	}()
}

func ____rune_private_5b8b8d7b_cliErrors(__result __CliParseResult) []string {
	__error := func() string {
		__coalesce1 := __result.__error
		if __coalesce1 != nil {
			return __coalesce1.(string)
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

func ____rune_private_5b8b8d7b_helpInvocation(__backend string) __RuneCliInvocation {
	return __RuneCliInvocation{__ok: true, __command: "", __backend: __backend, __path: "", __output: "", __target: "", __pattern: "", __checkOnly: false, __stdout: false, __backendExplicit: false, __runArgs: []string{}, __errors: []string{}, __help: true, __helpText: runeCliHelp(__runeCommand())}
}

func ____rune_private_5b8b8d7b_errorInvocation(__command string, __backend string, __errors []string) __RuneCliInvocation {
	return __RuneCliInvocation{__ok: false, __command: __command, __backend: __backend, __path: "", __output: "", __target: "", __pattern: "", __checkOnly: false, __stdout: false, __backendExplicit: false, __runArgs: []string{}, __errors: __errors, __help: false, __helpText: ""}
}

func ____rune_private_5b8b8d7b___cliMain(__args __RuneCliArgsEntry) {
	__invocation := __parseCli([]string{"--backend=" + __args.__backend, __args.__command})
	func() {
		if __invocation.__help {
			fmt.Print(runeCliHelp(__runeCommand()))
			return
		}
		fmt.Println(__invocation.__command)
	}()
}

func selfhostCliGeneratedMain() {
	__result := runeCliParseOrExit(runeCliWithArgument(runeCliWithOption(runeCliWithVersion(runeCliCommand("rune-selfhost", "Self-hosted Rune CLI entry point"), "0.0.0"), runeCliOption("backend", "b", "BACKEND", "target backend", false, "go")), runeCliArgument("command", "rune command", true)))
	____rune_private_5b8b8d7b___cliMain(__RuneCliArgsEntry{__backend: func() string {
		value, ok := __result.__values["backend"]
		if ok {
			return value
		}
		return "go"
	}(), __command: func() string {
		value, ok := __result.__positionals["command"]
		if ok {
			return value
		}
		return ""
	}()})
}

func selfhostCliGeneratedEntrypoint() {
	selfhostCliGeneratedMain()
}
