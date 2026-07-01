package main

import (
	"fmt"
	"os"
	"strings"
)

type __RuneCliExecution struct {
	__ok     bool
	__output string
	__errors []string
}

type __RuneCliStringResult struct {
	__ok     bool
	__value  string
	__errors []string
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
	return runeCliParseArgsMode(command, args, false)
}

func runeCliParseArgsMode(command __CliCommand, args []string, trailingRest bool) __CliParseResult {
	longOptions := map[string]__CliOption{}
	shortOptions := map[string]__CliOption{}
	values := map[string]string{}
	flags := map[string]bool{}
	positionals := map[string]string{}
	positionalValues := []string{}
	explicitOptions := []string{}
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
		if trailingRest && len(positionalValues) >= len(command.__arguments) {
			if arg == "--" {
				rest = append(rest, args[idx+1:]...)
			} else {
				rest = append(rest, args[idx:]...)
			}
			break
		}
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
					explicitOptions = append(explicitOptions, option.__name)
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
				explicitOptions = append(explicitOptions, option.__name)
			} else {
				flags[option.__name] = !hasValue || value != "false"
				explicitOptions = append(explicitOptions, option.__name)
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
					explicitOptions = append(explicitOptions, option.__name)
					break
				}
				flags[option.__name] = true
				explicitOptions = append(explicitOptions, option.__name)
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
		__command:         command,
		__values:          values,
		__flags:           flags,
		__positionals:     positionals,
		__explicitOptions: explicitOptions,
		__args:            append([]string(nil), args...),
		__rest:            rest,
		__help:            help,
		__error:           errorValue,
	}
}

func runeCliParseCommandArgs(root __CliCommand, commands []__CliCommand, aliases []__CliCommandAlias, trailingRest []string, args []string) __CliCommandParseResult {
	rootArgs, commandName, commandArgs := runeCliSplitCommandArgs(root, aliases, args)
	rootResult := runeCliParseArgs(root, rootArgs)
	if commandName == "" {
		return __CliCommandParseResult{__root: rootResult, __command: rootResult, __commandName: "", __commandArgs: []string{}, __error: any(nil)}
	}
	command, ok := runeCliFindCommand(commands, commandName)
	if !ok {
		return __CliCommandParseResult{__root: rootResult, __command: rootResult, __commandName: commandName, __commandArgs: commandArgs, __error: "unknown command " + commandName}
	}
	commandResult := runeCliParseArgsMode(command, commandArgs, runeCliContains(trailingRest, commandName))
	return __CliCommandParseResult{__root: rootResult, __command: commandResult, __commandName: commandName, __commandArgs: commandArgs, __error: any(nil)}
}

func runeCliParseRune(root __CliCommand, commands []__CliCommand, aliases []__CliCommandAlias, args []string) __RuneCliInvocation {
	parsed := runeCliParseCommandArgs(root, commands, aliases, []string{"run"}, args)
	backend := runeCliMapGetString(parsed.__root.__values, "backend", "go")
	errors := runeCliErrors(parsed.__root)
	if message, ok := parsed.__error.(string); ok && message != "" {
		errors = append(errors, message)
	}
	if !runeCliContains([]string{"go", "ts", "mbt"}, backend) {
		errors = append(errors, "unsupported backend "+backend)
	}
	backendExplicit := runeCliContains(parsed.__root.__explicitOptions, "backend")
	if len(errors) > 0 {
		return runeCliInvocation("", backend, "", "", "", "", false, false, false, []string{}, errors, false, "")
	}
	if parsed.__commandName == "" {
		return runeCliInvocation("", backend, "", "", "", "", false, false, backendExplicit, []string{}, []string{}, true, runeCliHelp(root))
	}
	return runeCliInvocationFromResult(parsed.__commandName, root, commands, backend, backendExplicit, parsed.__command)
}

func runeCliInvocationFromResult(name string, root __CliCommand, commands []__CliCommand, backend string, backendExplicit bool, result __CliParseResult) __RuneCliInvocation {
	errors := runeCliErrors(result)
	target := runeCliDefaultTarget(name, backend, runeCliMapGetString(result.__values, "target", ""))
	errors = runeCliInvocationErrors(name, backend, target, errors)
	return runeCliInvocation(
		name,
		backend,
		runeCliDefaultPath(name, runeCliMapGetString(result.__positionals, "path", "")),
		runeCliMapGetString(result.__values, "output", ""),
		target,
		runeCliMapGetString(result.__positionals, "pattern", ""),
		runeCliMapGetBool(result.__flags, "check", false),
		runeCliMapGetBool(result.__flags, "stdout", false),
		backendExplicit,
		result.__rest,
		errors,
		result.__help,
		runeCliHelp(runeCliHelpCommand(root, commands, name)),
	)
}

func runeCliInvocationErrors(name string, backend string, target string, errors []string) []string {
	if name == "build" && !runeCliContains([]string{"go", "mbt"}, backend) {
		return append(errors, "rune build only supports --backend go or --backend mbt")
	}
	if name == "run" && target != "" && backend != "mbt" {
		return append(errors, "rune run --target is only supported with --backend mbt")
	}
	return errors
}

func runeCliDefaultTarget(name string, backend string, target string) string {
	if name == "run" && backend == "mbt" && target == "" {
		return "native"
	}
	return target
}

func runeCliDefaultPath(name string, path string) string {
	if name == "test" && path == "" {
		return "tests"
	}
	return path
}

func runeCliHelpCommand(root __CliCommand, commands []__CliCommand, name string) __CliCommand {
	for _, command := range commands {
		if command.__name == name {
			return command
		}
	}
	return root
}

func runeCliErrors(result __CliParseResult) []string {
	if message, ok := result.__error.(string); ok && message != "" {
		return []string{message}
	}
	return []string{}
}

func runeCliMapGetString(values map[string]string, key string, fallback string) string {
	if value, ok := values[key]; ok {
		return value
	}
	return fallback
}

func runeCliMapGetBool(values map[string]bool, key string, fallback bool) bool {
	if value, ok := values[key]; ok {
		return value
	}
	return fallback
}

func runeCliInvocation(command string, backend string, path string, output string, target string, pattern string, checkOnly bool, stdout bool, backendExplicit bool, runArgs []string, errors []string, help bool, helpText string) __RuneCliInvocation {
	return __RuneCliInvocation{__ok: len(errors) == 0, __command: command, __backend: backend, __path: path, __output: output, __target: target, __pattern: pattern, __checkOnly: checkOnly, __stdout: stdout, __backendExplicit: backendExplicit, __runArgs: runArgs, __errors: errors, __help: help, __helpText: helpText}
}

func runeCliSplitCommandArgs(root __CliCommand, aliases []__CliCommandAlias, args []string) ([]string, string, []string) {
	rootArgs := []string{}
	commandArgs := []string{}
	commandName := ""
	inRest := false
	skipNext := false
	for idx := 0; idx < len(args); idx++ {
		arg := args[idx]
		if skipNext {
			skipNext = false
			continue
		}
		if commandName == "" && !inRest && (arg == "--help" || arg == "-h") {
			rootArgs = append(rootArgs, arg)
			continue
		}
		if !inRest {
			matched, consumesNext := runeCliRootOptionArg(root, arg)
			if matched {
				rootArgs = append(rootArgs, arg)
				if consumesNext && idx+1 < len(args) {
					rootArgs = append(rootArgs, args[idx+1])
					skipNext = true
				}
				continue
			}
		}
		if commandName == "" {
			commandName = runeCliResolveAlias(aliases, arg)
			continue
		}
		commandArgs = append(commandArgs, arg)
		if arg == "--" {
			inRest = true
		}
	}
	return rootArgs, commandName, commandArgs
}

func runeCliRootOptionArg(root __CliCommand, arg string) (bool, bool) {
	if strings.HasPrefix(arg, "--") && len(arg) > 2 {
		name := arg[2:]
		if before, _, ok := strings.Cut(name, "="); ok {
			name = before
		}
		option, ok := runeCliFindLongOption(root, name)
		return ok, ok && option.__valueName != "" && !strings.Contains(arg, "=")
	}
	if strings.HasPrefix(arg, "-") && arg != "-" {
		short := arg[1:]
		if len(short) > 1 {
			short = short[:1]
		}
		option, ok := runeCliFindShortOption(root, short)
		return ok, ok && option.__valueName != "" && len(arg) == 2
	}
	return false, false
}

func runeCliFindLongOption(command __CliCommand, name string) (__CliOption, bool) {
	for _, option := range command.__options {
		if option.__name == name {
			return option, true
		}
	}
	return __CliOption{}, false
}

func runeCliFindShortOption(command __CliCommand, short string) (__CliOption, bool) {
	for _, option := range command.__options {
		if option.__short == short {
			return option, true
		}
	}
	return __CliOption{}, false
}

func runeCliFindCommand(commands []__CliCommand, name string) (__CliCommand, bool) {
	for _, command := range commands {
		if command.__name == name {
			return command, true
		}
	}
	return __CliCommand{}, false
}

func runeCliResolveAlias(aliases []__CliCommandAlias, name string) string {
	for _, alias := range aliases {
		if alias.__from == name {
			return alias.__to
		}
	}
	return name
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
	return runeCliParseRune(__runeCommand(), __runeCommands(), __runeAliases(), __args)
}

func __runCli(__args []string) __RuneCliExecution {
	__invocation := __parseCli(__args)
	return func() __RuneCliExecution {
		if __invocation.__help {
			return ____rune_private_5b8b8d7b_successExecution(__invocation.__helpText)
		}
		return func() __RuneCliExecution {
			if __invocation.__ok {
				return ____rune_private_5b8b8d7b_executeInvocation(__invocation)
			}
			return ____rune_private_5b8b8d7b_failureExecution(__invocation.__errors)
		}()
	}()
}

func ____rune_private_5b8b8d7b_executeInvocation(__invocation __RuneCliInvocation) __RuneCliExecution {
	return func() __RuneCliExecution {
		switch {
		case __invocation.__command == "run":
			return ____rune_private_5b8b8d7b_executeRun(__invocation)
		case __invocation.__command == "build":
			return ____rune_private_5b8b8d7b_executeBuild(__invocation)
		case __invocation.__command == "go":
			return ____rune_private_5b8b8d7b_hostEmitGo(__invocation.__path, __invocation.__output)
		case __invocation.__command == "ts":
			return ____rune_private_5b8b8d7b_hostEmitTypeScript(__invocation.__path, __invocation.__output)
		case __invocation.__command == "mbt":
			return ____rune_private_5b8b8d7b_hostEmitMoonBit(__invocation.__path, __invocation.__output)
		case __invocation.__command == "check":
			return ____rune_private_5b8b8d7b_hostCheck(__invocation.__path)
		case __invocation.__command == "test":
			return ____rune_private_5b8b8d7b_hostTest(__invocation.__path, __invocation.__pattern, __invocation.__backend, __invocation.__backendExplicit)
		case __invocation.__command == "fmt":
			return ____rune_private_5b8b8d7b_hostFmt(__invocation.__path, __invocation.__checkOnly, __invocation.__stdout)
		case __invocation.__command == "repl":
			return ____rune_private_5b8b8d7b_hostRepl()
		case __invocation.__command == "lsp":
			return ____rune_private_5b8b8d7b_hostLsp()
		default:
			return ____rune_private_5b8b8d7b_failureExecution([]string{"unknown command " + __invocation.__command})
		}
	}()
}

func ____rune_private_5b8b8d7b_executeRun(__invocation __RuneCliInvocation) __RuneCliExecution {
	__entry := ____rune_private_5b8b8d7b_hostResolveRunEntry(__invocation.__path)
	return func() __RuneCliExecution {
		if __entry.__ok {
			return ____rune_private_5b8b8d7b_executeResolvedRun(__invocation, __entry.__value)
		}
		return ____rune_private_5b8b8d7b_failureExecution(__entry.__errors)
	}()
}

func ____rune_private_5b8b8d7b_executeResolvedRun(__invocation __RuneCliInvocation, __entry string) __RuneCliExecution {
	__backend := ____rune_private_5b8b8d7b_hostSelectRunBackend(__entry, __invocation.__backend, __invocation.__backendExplicit)
	__validated := ____rune_private_5b8b8d7b_hostValidateBackend(__backend)
	return func() __RuneCliExecution {
		if __validated.__ok {
			return ____rune_private_5b8b8d7b_hostRunEntry(__entry, __backend, __invocation.__target, __invocation.__runArgs)
		}
		return __validated
	}()
}

func ____rune_private_5b8b8d7b_executeBuild(__invocation __RuneCliInvocation) __RuneCliExecution {
	return func() __RuneCliExecution {
		switch {
		case __invocation.__backend == "mbt":
			return ____rune_private_5b8b8d7b_hostBuildMoonBit(__invocation.__path, __invocation.__target, __invocation.__output)
		default:
			return ____rune_private_5b8b8d7b_hostBuildGo(__invocation.__path, __invocation.__target, __invocation.__output)
		}
	}()
}

func ____rune_private_5b8b8d7b_successExecution(__output string) __RuneCliExecution {
	return __RuneCliExecution{__ok: true, __output: __output, __errors: []string{}}
}

func ____rune_private_5b8b8d7b_failureExecution(__errors []string) __RuneCliExecution {
	return __RuneCliExecution{__ok: false, __output: "", __errors: __errors}
}

func ____rune_private_5b8b8d7b_hostResolveRunEntry(__path string) __RuneCliStringResult {
	return hostResolveRunEntry(__path)
}

func ____rune_private_5b8b8d7b_hostSelectRunBackend(__entry string, __backend string, __backendExplicit bool) string {
	return hostSelectRunBackend(__entry, __backend, __backendExplicit)
}

func ____rune_private_5b8b8d7b_hostValidateBackend(__backend string) __RuneCliExecution {
	return hostValidateBackend(__backend)
}

func ____rune_private_5b8b8d7b_hostRunEntry(__path string, __backend string, __target string, __args []string) __RuneCliExecution {
	return hostRunEntry(__path, __backend, __target, __args)
}

func ____rune_private_5b8b8d7b_hostBuildGo(__path string, __target string, __output string) __RuneCliExecution {
	return hostBuildGo(__path, __target, __output)
}

func ____rune_private_5b8b8d7b_hostBuildMoonBit(__path string, __target string, __output string) __RuneCliExecution {
	return hostBuildMoonBit(__path, __target, __output)
}

func ____rune_private_5b8b8d7b_hostEmitGo(__path string, __output string) __RuneCliExecution {
	return hostEmitGo(__path, __output)
}

func ____rune_private_5b8b8d7b_hostEmitTypeScript(__path string, __output string) __RuneCliExecution {
	return hostEmitTypeScript(__path, __output)
}

func ____rune_private_5b8b8d7b_hostEmitMoonBit(__path string, __output string) __RuneCliExecution {
	return hostEmitMoonBit(__path, __output)
}

func ____rune_private_5b8b8d7b_hostCheck(__path string) __RuneCliExecution {
	return hostCheck(__path)
}

func ____rune_private_5b8b8d7b_hostTest(__path string, __pattern string, __backend string, __backendExplicit bool) __RuneCliExecution {
	return hostTest(__path, __pattern, __backend, __backendExplicit)
}

func ____rune_private_5b8b8d7b_hostFmt(__path string, __checkOnly bool, __stdout bool) __RuneCliExecution {
	return hostFmt(__path, __checkOnly, __stdout)
}

func ____rune_private_5b8b8d7b_hostRepl() __RuneCliExecution {
	return hostRepl()
}

func ____rune_private_5b8b8d7b_hostLsp() __RuneCliExecution {
	return hostLsp()
}

func __runeCommand() __CliCommand {
	__command := runeCliCommand("rune", "Rune language toolchain")
	__command = runeCliWithOption(__command, runeCliOption("backend", "b", "BACKEND", "target backend", false, "go"))
	return __command
}

func __runeCommands() []__CliCommand {
	return []__CliCommand{__runCommand(), __buildCommand(), __emitCommand("go"), __emitCommand("ts"), __emitCommand("mbt"), __singlePathCommand("check"), __fmtCommand(), __testCommand(), runeCliCommand("repl", ""), __lspCommand()}
}

func __runeAliases() []__CliCommandAlias {
	return []__CliCommandAlias{__CliCommandAlias{__from: "format", __to: "fmt"}}
}

func __runCommand() __CliCommand {
	__command := runeCliCommand("run", "Compile and run a Rune program")
	__command = runeCliWithOption(__command, runeCliOption("target", "", "TARGET", "MoonBit run target", false, any(nil)))
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func __buildCommand() __CliCommand {
	__command := runeCliCommand("build", "Compile a Rune program to an executable")
	__command = runeCliWithOption(__command, runeCliOption("output", "o", "FILE", "output executable path", false, ""))
	__command = runeCliWithOption(__command, runeCliOption("target", "", "TARGET", "build target", false, ""))
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func __emitCommand(__name string) __CliCommand {
	__command := runeCliCommand(__name, "Compile a Rune program")
	__command = runeCliWithOption(__command, runeCliOption("output", "o", "FILE", "output file", false, ""))
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func __singlePathCommand(__name string) __CliCommand {
	__command := runeCliCommand(__name, "Parse and type-check Rune source")
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func __fmtCommand() __CliCommand {
	__command := runeCliCommand("fmt", "Format Rune source")
	__command = runeCliWithOption(__command, runeCliFlag("check", "", "fail if not formatted"))
	__command = runeCliWithOption(__command, runeCliFlag("stdout", "", "write formatted source to stdout"))
	return runeCliWithArgument(__command, runeCliArgument("path", "Rune source path", true))
}

func __testCommand() __CliCommand {
	__command := runeCliCommand("test", "Run Rune tests")
	__command = runeCliWithArgument(__command, runeCliArgument("path", "Rune test path", false))
	return runeCliWithArgument(__command, runeCliArgument("pattern", "test name pattern", false))
}

func __lspCommand() __CliCommand {
	__command := runeCliCommand("lsp", "Start the Rune language server")
	return runeCliWithOption(__command, runeCliFlag("stdio", "", "serve LSP over stdin/stdout"))
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
