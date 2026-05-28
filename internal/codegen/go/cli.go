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
	case "help":
		if len(args) != 1 {
			return g.zeroValue(resultType)
		}
		return fmt.Sprintf("runeCliHelp(%s)", args[0])
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
	__args []string
	__rest []string
	__help bool
	__error any
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
		__command: command,
		__values: values,
		__flags: flags,
		__positionals: positionals,
		__args: append([]string(nil), args...),
		__rest: rest,
		__help: help,
		__error: errorValue,
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
`
	for _, line := range strings.Split(strings.Trim(src, "\n"), "\n") {
		g.line(line)
	}
}
