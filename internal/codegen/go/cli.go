package gocodegen

import (
	"fmt"
	"strings"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (g *generator) cliModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	name := fn.Name
	switch name {
	case "command":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliCommand(%s, %s)", args[0], args[1])
	case "withVersion":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliWithVersion(%s, %s)", args[0], args[1])
	case "withOption":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliWithOption(%s, %s)", args[0], args[1])
	case "withArgument":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliWithArgument(%s, %s)", args[0], args[1])
	case "flag":
		if len(args) != 3 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliFlag(%s, %s, %s)", args[0], args[1], args[2])
	case "option":
		if len(args) != 6 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliOption(%s, %s, %s, %s, %s, %s)", args[0], args[1], args[2], args[3], args[4], args[5])
	case "argument":
		if len(args) != 3 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliArgument(%s, %s, %s)", args[0], args[1], args[2])
	case "alias":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("__CliCommandAlias{__from: %s, __to: %s}", args[0], args[1])
	case "parse":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliParse(%s)", args[0])
	case "parseArgs":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliParseArgs(%s, %s)", args[0], args[1])
	case "parseCommandArgs":
		if len(args) != 5 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliParseCommandArgs(%s, %s, %s, %s, %s)", args[0], args[1], args[2], args[3], args[4])
	case "parseRuneCli":
		if len(args) != 4 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliParseRune(%s, %s, %s, %s)", args[0], args[1], args[2], args[3])
	case "help":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliHelp(%s)", args[0])
	case "contains":
		if len(args) != 2 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliContains(%s, %s)", args[0], args[1])
	case "parseOrExit":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliParseOrExit(%s)", args[0])
	default:
		return g.unsupportedIntrinsic(fn, resultType)
	}
}

func (g *generator) cliRuntime() {
	const src = `
type __CliCommand struct {
	__name string
	__version any
	__about string
	__options []__CliOption
	__arguments []__CliArgument
}

type __CliOption struct {
	__name string
	__short string
	__valueName string
	__help string
	__required bool
	__defaultValue any
}

type __CliArgument struct {
	__name string
	__help string
	__required bool
}

type __CliParseResult struct {
	__command __CliCommand
	__values map[string]string
	__flags map[string]bool
	__positionals map[string]string
	__explicitOptions []string
	__args []string
	__rest []string
	__help bool
	__error any
}

type __CliCommandAlias struct {
	__from string
	__to string
}

type __CliCommandParseResult struct {
	__root __CliParseResult
	__command __CliParseResult
	__commandName string
	__commandArgs []string
	__error any
}

type __RuneCliInvocation struct {
	__ok bool
	__command string
	__backend string
	__path string
	__output string
	__target string
	__pattern string
	__checkOnly bool
	__stdout bool
	__backendExplicit bool
	__runArgs []string
	__errors []string
	__help bool
	__helpText string
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
		__command: command,
		__values: values,
		__flags: flags,
		__positionals: positionals,
		__explicitOptions: explicitOptions,
		__args: append([]string(nil), args...),
		__rest: rest,
		__help: help,
		__error: errorValue,
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
`
	for _, line := range strings.Split(strings.Trim(src, "\n"), "\n") {
		g.line(line)
	}
	if g.fileHasStruct("RuneCliExecution") && g.fileHasStruct("RuneCliStringResult") {
		g.line("")
		g.runeCliHostRuntime()
	}
}

func (g *generator) fileHasStruct(name string) bool {
	for _, typ := range g.file.Types {
		if typ.Name == name {
			return true
		}
	}
	return false
}

func (g *generator) runeCliHostRuntime() {
	const src = `
func runeCliHostUnavailable() __RuneCliExecution {
	return __RuneCliExecution{__ok: false, __output: "", __errors: []string{"rune CLI host is unavailable"}}
}

func runeCliStringHostUnavailable() __RuneCliStringResult {
	return __RuneCliStringResult{__ok: false, __value: "", __errors: []string{"rune CLI host is unavailable"}}
}

var runeCliHostResolveRunEntry = func(path string) __RuneCliStringResult {
	return runeCliStringHostUnavailable()
}

var runeCliHostSelectRunBackend = func(entry string, backend string, backendExplicit bool) string {
	return backend
}

var runeCliHostValidateBackend = func(backend string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostRunEntry = func(path string, backend string, target string, args []string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostBuildGo = func(path string, target string, output string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostBuildMoonBit = func(path string, target string, output string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostEmitGo = func(path string, output string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostEmitTypeScript = func(path string, output string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostEmitMoonBit = func(path string, output string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostCheck = func(path string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostTest = func(path string, pattern string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostTestWithBackend = func(path string, pattern string, backend string) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostFmt = func(path string, checkOnly bool, stdout bool) __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostRepl = func() __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostLsp = func() __RuneCliExecution {
	return runeCliHostUnavailable()
}

var runeCliHostWriteStdout = func(text string) int {
	return 0
}

var runeCliHostWriteStderr = func(text string) int {
	return 0
}
`
	for _, line := range strings.Split(strings.Trim(src, "\n"), "\n") {
		g.line(line)
	}
}
