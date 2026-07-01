package tscodegen

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
type __CliCommand = {
  name: string;
  version: string | null;
  about: string;
  options: __CliOption[];
  arguments: __CliArgument[];
};

type __CliOption = {
  name: string;
  short: string;
  valueName: string;
  help: string;
  required: boolean;
  defaultValue: string | null;
};

type __CliArgument = {
  name: string;
  help: string;
  required: boolean;
};

type __CliParseResult = {
  command: __CliCommand;
  values: Map<string, string>;
  flags: Map<string, boolean>;
  positionals: Map<string, string>;
  args: string[];
  rest: string[];
  help: boolean;
  error: string | null;
};

function runeCliCommand(name: string, about: string): __CliCommand {
  return { name, version: null, about, options: [], arguments: [] };
}

function runeCliWithVersion(command: __CliCommand, version: string): __CliCommand {
  return { ...command, version };
}

function runeCliWithOption(command: __CliCommand, option: __CliOption): __CliCommand {
  return { ...command, options: [...command.options, option] };
}

function runeCliWithArgument(command: __CliCommand, argument: __CliArgument): __CliCommand {
  return { ...command, arguments: [...command.arguments, argument] };
}

function runeCliFlag(name: string, short: string, help: string): __CliOption {
  return { name, short, help, valueName: "", required: false, defaultValue: null };
}

function runeCliOption(name: string, short: string, valueName: string, help: string, required: boolean, defaultValue: string | null): __CliOption {
  return { name, short, valueName, help, required, defaultValue };
}

function runeCliArgument(name: string, help: string, required: boolean): __CliArgument {
  return { name, help, required };
}

function runeCliRuntimeArgs(): string[] {
  const proc = (globalThis as any).process;
  if (Array.isArray(proc?.argv)) return proc.argv.slice(2).map(String);
  const deno = (globalThis as any).Deno;
  if (Array.isArray(deno?.args)) return deno.args.map(String);
  return [];
}

function runeCliParse(command: __CliCommand): __CliParseResult {
  return runeCliParseArgs(command, runeCliRuntimeArgs());
}

function runeCliParseArgs(command: __CliCommand, args: string[]): __CliParseResult {
  const longOptions = new Map<string, __CliOption>();
  const shortOptions = new Map<string, __CliOption>();
  const values = new Map<string, string>();
  const flags = new Map<string, boolean>();
  const positionals = new Map<string, string>();
  const positionalValues: string[] = [];
  let rest: string[] = [];
  let help = false;
  let parseError: string | null = null;

  for (const option of command.options) {
    longOptions.set(option.name, option);
    if (option.short !== "") shortOptions.set(option.short, option);
    if (option.valueName !== "" && option.defaultValue !== null) values.set(option.name, option.defaultValue);
  }

  for (let idx = 0; idx < args.length; idx += 1) {
    const arg = args[idx]!;
    if (arg === "--") {
      rest = args.slice(idx + 1);
      break;
    }
    if (arg === "--help" || arg === "-h") {
      help = true;
      continue;
    }
    if (arg.startsWith("--") && arg.length > 2) {
      let name = arg.slice(2);
      let value = "";
      let hasValue = false;
      const eq = name.indexOf("=");
      if (eq >= 0) {
        value = name.slice(eq + 1);
        name = name.slice(0, eq);
        hasValue = true;
      }
      if (name.startsWith("no-")) {
        const option = longOptions.get(name.slice(3));
        if (option && option.valueName === "") {
          flags.set(option.name, false);
          continue;
        }
      }
      const option = longOptions.get(name);
      if (!option) {
        parseError = "unknown option --" + name;
        break;
      }
      if (option.valueName !== "") {
        if (!hasValue) {
          idx += 1;
          if (idx >= args.length) {
            parseError = "missing value for --" + name;
            break;
          }
          value = args[idx]!;
        }
        values.set(option.name, value);
      } else {
        flags.set(option.name, !hasValue || value !== "false");
      }
      continue;
    }
    if (arg.startsWith("-") && arg !== "-") {
      const cluster = Array.from(arg.slice(1));
      for (let pos = 0; pos < cluster.length; pos += 1) {
        const short = cluster[pos]!;
        if (short === "h") {
          help = true;
          continue;
        }
        const option = shortOptions.get(short);
        if (!option) {
          parseError = "unknown option -" + short;
          break;
        }
        if (option.valueName !== "") {
          let value = cluster.slice(pos + 1).join("");
          if (value === "") {
            idx += 1;
            if (idx >= args.length) {
              parseError = "missing value for -" + short;
              break;
            }
            value = args[idx]!;
          }
          values.set(option.name, value);
          break;
        }
        flags.set(option.name, true);
      }
      if (parseError !== null) break;
      continue;
    }
    positionalValues.push(arg);
  }

  if (parseError === null && !help) {
    for (const option of command.options) {
      if (option.required && option.valueName !== "" && !values.has(option.name)) {
        parseError = "missing required option --" + option.name;
        break;
      }
    }
  }
  if (parseError === null && !help) {
    for (let idx = 0; idx < command.arguments.length; idx += 1) {
      const argument = command.arguments[idx]!;
      if (idx < positionalValues.length) {
        positionals.set(argument.name, positionalValues[idx]!);
        continue;
      }
      if (argument.required) {
        parseError = "missing required argument " + argument.name;
        break;
      }
    }
    if (parseError === null && positionalValues.length > command.arguments.length) {
      parseError = "unexpected argument " + positionalValues[command.arguments.length]!;
    }
  } else {
    for (let idx = 0; idx < command.arguments.length && idx < positionalValues.length; idx += 1) {
      positionals.set(command.arguments[idx]!.name, positionalValues[idx]!);
    }
  }

  return { command, values, flags, positionals, args: [...args], rest, help, error: parseError };
}

function runeCliHelp(command: __CliCommand): string {
  const lines: string[] = [];
  let usage = "Usage: " + command.name;
  if (command.options.length > 0) usage += " [options]";
  for (const argument of command.arguments) usage += " " + (argument.required ? "<" + argument.name + ">" : "[" + argument.name + "]");
  lines.push(usage);
  if (command.version) lines.push("Version: " + command.version);
  if (command.about !== "") lines.push("", command.about);
  if (command.arguments.length > 0) {
    lines.push("", "Arguments:");
    for (const argument of command.arguments) lines.push("  " + argument.name + (argument.help ? "\t" + argument.help : ""));
  }
  lines.push("", "Options:");
  for (const option of command.options) {
    let usage = "  ";
    if (option.short !== "") usage += "-" + option.short + ", ";
    usage += "--" + option.name;
    if (option.valueName !== "") usage += " <" + option.valueName + ">";
    if (option.help !== "") usage += "\t" + option.help;
    if (option.required) usage += " (required)";
    if (option.defaultValue !== null && option.defaultValue !== "") usage += " (default: " + option.defaultValue + ")";
    lines.push(usage);
  }
  lines.push("  -h, --help\tShow help");
  return lines.join("\n") + "\n";
}

function runeCliParseOrExit(command: __CliCommand): __CliParseResult {
  const result = runeCliParse(command);
  if (result.help) {
    console.log(runeCliHelp(command).trimEnd());
    process.exit(0);
  }
  if (result.error) {
    console.error(result.error);
    console.error(runeCliHelp(command).trimEnd());
    process.exit(2);
  }
  return result;
}

function runeCliContains(values: string[], value: string): boolean {
  return values.includes(value);
}
`
	for _, line := range strings.Split(strings.Trim(src, "\n"), "\n") {
		g.line(line)
	}
}
