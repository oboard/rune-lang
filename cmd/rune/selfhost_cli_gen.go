package main

import (
	"fmt"
	"os"
	"runtime"
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
	__parsed := runeCliParseCommandArgs(__runeCommand(), __runeCommands(), __runeAliases(), __runeTrailingRest(), __args)
	return ____rune_private_db6df590_invocationFromParsed(__parsed, __runeCommand(), __runeCommands())
}

func ____rune_private_db6df590_invocationFromParsed(__parsed __CliCommandParseResult, __rootCommand __CliCommand, __commands []__CliCommand) __RuneCliInvocation {
	__root := __parsed.__root
	__backend := func() string {
		value, ok := __root.__values["backend"]
		if ok {
			return value
		}
		return "go"
	}()
	__errors := ____rune_private_db6df590_cliErrors(__root)
	__commandError := func() string {
		__coalesce1 := __parsed.__error
		if __coalesce1 != nil {
			return __coalesce1.(string)
		}
		return ""
	}()
	func() int {
		if len(__commandError) == 0 {
			return 0
		}
		return func() int { __errors = append(__errors, __commandError); return len(__errors) }()
	}()
	__invalidBackend := runeCliContains([]string{"go", "ts", "mbt"}, __backend) == false
	func() int {
		if __invalidBackend {
			return func() int { __errors = append(__errors, "unsupported backend "+__backend); return len(__errors) }()
		}
		return 0
	}()
	__backendExplicit := runeCliContains(__root.__explicitOptions, "backend")
	return func() __RuneCliInvocation {
		if len(__errors) > 0 {
			return ____rune_private_db6df590_errorInvocation(__backend, __errors)
		}
		return func() __RuneCliInvocation {
			if len(__parsed.__commandName) == 0 {
				return ____rune_private_db6df590_helpInvocation(__rootCommand, __backend, __backendExplicit)
			}
			return ____rune_private_db6df590_invocationFromResult(__parsed.__commandName, __rootCommand, __commands, __backend, __backendExplicit, __parsed.__command)
		}()
	}()
}

func ____rune_private_db6df590_invocationFromResult(__name string, __rootCommand __CliCommand, __commands []__CliCommand, __backend string, __backendExplicit bool, __result __CliParseResult) __RuneCliInvocation {
	__errors := ____rune_private_db6df590_cliErrors(__result)
	__target := func() string {
		value, ok := __result.__values["target"]
		if ok {
			return value
		}
		return ""
	}()
	__target = ____rune_private_db6df590_defaultTargetForCommand(__name, __backend, __target)
	__errors = ____rune_private_db6df590_invocationErrors(__name, __backend, __target, __errors)
	return __RuneCliInvocation{__ok: len(__errors) == 0, __command: __name, __backend: __backend, __path: ____rune_private_db6df590_defaultPathForCommand(__name, func() string {
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
	}(), __backendExplicit: __backendExplicit, __runArgs: __result.__rest, __errors: __errors, __help: __result.__help, __helpText: runeCliHelp(____rune_private_db6df590_invocationHelpCommand(__rootCommand, __commands, __name))}
}

func ____rune_private_db6df590_invocationErrors(__name string, __backend string, __target string, __errors []string) []string {
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

func ____rune_private_db6df590_defaultTargetForCommand(__name string, __backend string, __target string) string {
	return func() string {
		if __name == "run" && __backend == "mbt" && len(__target) == 0 {
			return "native"
		}
		return __target
	}()
}

func ____rune_private_db6df590_defaultPathForCommand(__name string, __path string) string {
	return func() string {
		if __name == "test" && len(__path) == 0 {
			return "tests"
		}
		return __path
	}()
}

func ____rune_private_db6df590_invocationHelpCommand(__rootCommand __CliCommand, __commands []__CliCommand, __name string) __CliCommand {
	return func() __CliCommand {
		__array2 := __commands
		__result3 := __rootCommand
		for _, __value5 := range __array2 {
			__result3 = func(__found __CliCommand, __command __CliCommand) __CliCommand {
				return func() __CliCommand {
					if __command.__name == __name {
						return __command
					}
					return __found
				}()
			}(__result3, __value5)
		}
		return __result3
	}()
}

func ____rune_private_db6df590_cliErrors(__result __CliParseResult) []string {
	__error := func() string {
		__coalesce6 := __result.__error
		if __coalesce6 != nil {
			return __coalesce6.(string)
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

func ____rune_private_db6df590_helpInvocation(__rootCommand __CliCommand, __backend string, __backendExplicit bool) __RuneCliInvocation {
	return __RuneCliInvocation{__ok: true, __command: "", __backend: __backend, __path: "", __output: "", __target: "", __pattern: "", __checkOnly: false, __stdout: false, __backendExplicit: __backendExplicit, __runArgs: []string{}, __errors: []string{}, __help: true, __helpText: runeCliHelp(__rootCommand)}
}

func ____rune_private_db6df590_errorInvocation(__backend string, __errors []string) __RuneCliInvocation {
	return __RuneCliInvocation{__ok: false, __command: "", __backend: __backend, __path: "", __output: "", __target: "", __pattern: "", __checkOnly: false, __stdout: false, __backendExplicit: false, __runArgs: []string{}, __errors: __errors, __help: false, __helpText: ""}
}

func ____rune_private_db6df590_emptyInvocation() __RuneCliInvocation {
	return ____rune_private_db6df590_errorInvocation("", []string{})
}

func __runeCommand() __CliCommand {
	__command := runeCliCommand("rune", "Rune language toolchain")
	__command = runeCliWithOption(__command, runeCliOption("backend", "b", "BACKEND", "target backend", false, "go"))
	return __command
}

func __runeCommands() []__CliCommand {
	return []__CliCommand{__runCommand(), __buildCommand(), __emitCommand("go"), __emitCommand("ts"), __emitCommand("dts"), __emitCommand("mbt"), __singlePathCommand("check"), __fmtCommand(), __testCommand(), runeCliCommand("repl", ""), __lspCommand()}
}

func __runeAliases() []__CliCommandAlias {
	return []__CliCommandAlias{__CliCommandAlias{__from: "format", __to: "fmt"}}
}

func __runeTrailingRest() []string {
	return []string{"run"}
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

func selfhostCliGeneratedMain() {
	__argv := runeProcessArgv()
	__invocation := __parseCli(append([]string{}, __argv[1:len(__argv)]...))
	func() {
		if __invocation.__help {
			fmt.Print(runeCliHelp(__runeCommand()))
			return
		}
		fmt.Println(__invocation.__command)
	}()
}

func selfhostCliGeneratedEntrypoint() {
	selfhostCliGeneratedMain()
}
