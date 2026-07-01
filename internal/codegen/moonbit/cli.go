package moonbitcodegen

import (
	"fmt"

	"github.com/oboard/rune-lang/internal/checker"
	"github.com/oboard/rune-lang/internal/stdlib"
)

func (g *generator) cliModuleCall(fn *stdlib.Function, args []string, resultType checker.Type) string {
	g.useCLI = true
	switch fn.Name {
	case "command":
		if len(args) == 2 {
			return fmt.Sprintf("rune_cli_command(%s, %s)", args[0], args[1])
		}
	case "withVersion":
		if len(args) == 2 {
			return fmt.Sprintf("rune_cli_with_version(%s, %s)", args[0], args[1])
		}
	case "withOption":
		if len(args) == 2 {
			return fmt.Sprintf("rune_cli_with_option(%s, %s)", args[0], args[1])
		}
	case "withArgument":
		if len(args) == 2 {
			return fmt.Sprintf("rune_cli_with_argument(%s, %s)", args[0], args[1])
		}
	case "flag":
		if len(args) == 3 {
			return fmt.Sprintf("rune_cli_flag(%s, %s, %s)", args[0], args[1], args[2])
		}
	case "option":
		if len(args) == 6 {
			defaultValue := "Some(" + args[5] + ")"
			if args[5] == "None" {
				defaultValue = "None"
			}
			return fmt.Sprintf("rune_cli_option(%s, %s, %s, %s, %s, %s)", args[0], args[1], args[2], args[3], args[4], defaultValue)
		}
	case "argument":
		if len(args) == 3 {
			return fmt.Sprintf("rune_cli_argument(%s, %s, %s)", args[0], args[1], args[2])
		}
	case "alias":
		if len(args) == 2 {
			return fmt.Sprintf("CliCommandAlias::{ from: %s, to: %s }", args[0], args[1])
		}
	case "parse":
		if len(args) == 1 {
			return fmt.Sprintf("rune_cli_parse(%s)", args[0])
		}
	case "parseArgs":
		if len(args) == 2 {
			return fmt.Sprintf("rune_cli_parse_args(%s, %s)", args[0], args[1])
		}
	case "parseCommandArgs":
		if len(args) == 5 {
			return fmt.Sprintf("rune_cli_parse_command_args(%s, %s, %s, %s, %s)", args[0], args[1], args[2], args[3], args[4])
		}
	case "help":
		if len(args) == 1 {
			return fmt.Sprintf("rune_cli_help(%s)", args[0])
		}
	case "contains":
		if len(args) == 2 {
			return fmt.Sprintf("rune_cli_contains(%s, %s)", args[0], args[1])
		}
	case "parseOrExit":
		if len(args) == 1 {
			return fmt.Sprintf("rune_cli_parse_or_exit(%s)", args[0])
		}
	}
	g.addError(fmt.Errorf("MoonBit backend cannot lower @cli.%s with %d args", fn.Name, len(args)))
	return zeroValue(resultType)
}

func (g *generator) cliRuntime() {
	g.line("struct CliCommand {")
	g.indent++
	g.line("name : String")
	g.line("version : String?")
	g.line("about : String")
	g.line("options : Array[CliOption]")
	g.line("arguments : Array[CliArgument]")
	g.indent--
	g.line("}")
	g.line("")
	g.line("struct CliOption {")
	g.indent++
	g.line("name : String")
	g.line("short : String")
	g.line("value_name : String")
	g.line("help : String")
	g.line("required : Bool")
	g.line("default_value : String?")
	g.indent--
	g.line("}")
	g.line("")
	g.line("struct CliArgument {")
	g.indent++
	g.line("name : String")
	g.line("help : String")
	g.line("required : Bool")
	g.indent--
	g.line("}")
	g.line("")
	g.line("struct CliParseResult {")
	g.indent++
	g.line("command : CliCommand")
	g.line("values : Map[String, String]")
	g.line("flags : Map[String, Bool]")
	g.line("positionals : Map[String, String]")
	g.line("explicit_options : Array[String]")
	g.line("args : Array[String]")
	g.line("rest : Array[String]")
	g.line("help : Bool")
	g.line("error : String?")
	g.indent--
	g.line("}")
	g.line("")
	g.line("struct CliCommandAlias {")
	g.indent++
	g.line("from : String")
	g.line("to : String")
	g.indent--
	g.line("}")
	g.line("")
	g.line("struct CliCommandParseResult {")
	g.indent++
	g.line("root : CliParseResult")
	g.line("command : CliParseResult")
	g.line("command_name : String")
	g.line("command_args : Array[String]")
	g.line("error : String?")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_command(name : String, about : String) -> CliCommand {")
	g.indent++
	g.line("CliCommand::{ name, version: None, about, options: [], arguments: [] }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_with_version(command : CliCommand, version : String) -> CliCommand {")
	g.indent++
	g.line("CliCommand::{ name: command.name, version: Some(version), about: command.about, options: command.options, arguments: command.arguments }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_with_option(command : CliCommand, option : CliOption) -> CliCommand {")
	g.indent++
	g.line("let options = command.options.copy()")
	g.line("options.push(option)")
	g.line("CliCommand::{ name: command.name, version: command.version, about: command.about, options, arguments: command.arguments }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_with_argument(command : CliCommand, argument : CliArgument) -> CliCommand {")
	g.indent++
	g.line("let arguments = command.arguments.copy()")
	g.line("arguments.push(argument)")
	g.line("CliCommand::{ name: command.name, version: command.version, about: command.about, options: command.options, arguments }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_flag(name : String, short : String, help : String) -> CliOption {")
	g.indent++
	g.line("CliOption::{ name, short, value_name: \"\", help, required: false, default_value: None }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_option(name : String, short : String, value_name : String, help : String, required : Bool, default_value : String?) -> CliOption {")
	g.indent++
	g.line("CliOption::{ name, short, value_name, help, required, default_value }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_argument(name : String, help : String, required : Bool) -> CliArgument {")
	g.indent++
	g.line("CliArgument::{ name, help, required }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_parse(command : CliCommand) -> CliParseResult {")
	g.indent++
	g.line("let argv = @env.args()")
	g.line("let args = if argv.length() > 1 { argv[1:].to_owned() } else { [] }")
	g.line("rune_cli_parse_args(command, args)")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_parse_or_exit(command : CliCommand) -> CliParseResult {")
	g.indent++
	g.line("let result = rune_cli_parse(command)")
	g.line("match result.error {")
	g.indent++
	g.line("Some(message) => { println(message); println(rune_cli_help(command)) }")
	g.line("None => ()")
	g.indent--
	g.line("}")
	g.line("result")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_contains(values : Array[String], value : String) -> Bool {")
	g.indent++
	g.line("let mut index = 0")
	g.line("let mut found = false")
	g.line("while index < values.length() {")
	g.indent++
	g.line("if values[index] == value { found = true }")
	g.line("index = index + 1")
	g.indent--
	g.line("}")
	g.line("found")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_parse_command_args(root : CliCommand, commands : Array[CliCommand], aliases : Array[CliCommandAlias], trailing_rest : Array[String], args : Array[String]) -> CliCommandParseResult {")
	g.indent++
	g.line("let root_result = rune_cli_parse_args(root, args)")
	g.line("let mut command_name = \"\"")
	g.line("let command_args : Array[String] = []")
	g.line("let mut index = 0")
	g.line("while index < args.length() {")
	g.indent++
	g.line("let arg = args[index]")
	g.line("if command_name.is_empty() && !arg.starts_with(\"-\") { command_name = arg } else if !command_name.is_empty() { command_args.push(arg) }")
	g.line("index = index + 1")
	g.indent--
	g.line("}")
	g.line("aliases.each((alias) => { if command_name == alias.from { command_name = alias.to } })")
	g.line("let mut found = false")
	g.line("let mut command = root")
	g.line("commands.each((candidate) => { if candidate.name == command_name { command = candidate; found = true } })")
	g.line("let command_result = if found { rune_cli_parse_args(command, command_args) } else { root_result }")
	g.line("let error = if !command_name.is_empty() && !found { Some(\"unknown command \" + command_name) } else { None }")
	g.line("CliCommandParseResult::{ root: root_result, command: command_result, command_name, command_args, error }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_parse_args(command : CliCommand, args : Array[String]) -> CliParseResult {")
	g.indent++
	g.line("let values : Map[String, String] = {}")
	g.line("let flags : Map[String, Bool] = {}")
	g.line("let positionals : Map[String, String] = {}")
	g.line("let explicit_options : Array[String] = []")
	g.line("let mut rest : Array[String] = []")
	g.line("let positional_values : Array[String] = []")
	g.line("let mut help = false")
	g.line("let mut error : String? = None")
	g.line("command.options.each((option) => {")
	g.indent++
	g.line("match option.default_value {")
	g.indent++
	g.line("Some(value) => if !option.value_name.is_empty() { values[option.name] = value }")
	g.line("None => ()")
	g.indent--
	g.line("}")
	g.indent--
	g.line("})")
	g.line("let mut idx = 0")
	g.line("while idx < args.length() {")
	g.indent++
	g.line("let arg = args[idx]")
	g.line("if arg == \"--\" {")
	g.indent++
	g.line("rest = args[idx + 1:].to_owned()")
	g.line("idx = args.length()")
	g.indent--
	g.line("} else if arg == \"--help\" || arg == \"-h\" {")
	g.indent++
	g.line("help = true")
	g.line("idx += 1")
	g.indent--
	g.line("} else if arg.has_prefix(\"--\") {")
	g.indent++
	g.line("let raw_name = arg[2:].to_owned()")
	g.line("let equal = raw_name.find(\"=\").unwrap_or(-1)")
	g.line("let name = if equal >= 0 { raw_name[0:equal].to_owned() } else { raw_name }")
	g.line("let inline_value : String? = if equal >= 0 { Some(raw_name[equal + 1:raw_name.length()].to_owned()) } else { None }")
	g.line("let mut matched = false")
	g.line("command.options.each((option) => {")
	g.indent++
	g.line("if option.name == name {")
	g.indent++
	g.line("matched = true")
	g.line("if option.value_name.is_empty() {")
	g.indent++
	g.line("flags[option.name] = true")
	g.indent--
	g.line("} else if inline_value != None {")
	g.indent++
	g.line("values[option.name] = inline_value.unwrap()")
	g.indent--
	g.line("} else if idx + 1 < args.length() {")
	g.indent++
	g.line("idx += 1")
	g.line("values[option.name] = args[idx]")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("error = Some(\"missing value for --\" + name)")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.indent--
	g.line("})")
	g.line("if !matched { error = Some(\"unknown option --\" + name) }")
	g.line("idx += 1")
	g.indent--
	g.line("} else if arg.has_prefix(\"-\") && arg.length() > 1 {")
	g.indent++
	g.line("let short_arg = arg[1:].to_owned()")
	g.line("let short = short_arg[0:1].to_owned()")
	g.line("let short_value : String? = if short_arg.length() > 1 { Some(short_arg[1:short_arg.length()].to_owned()) } else { None }")
	g.line("let mut matched = false")
	g.line("command.options.each((option) => {")
	g.indent++
	g.line("if option.short == short {")
	g.indent++
	g.line("matched = true")
	g.line("if option.value_name.is_empty() {")
	g.indent++
	g.line("flags[option.name] = true")
	g.indent--
	g.line("} else if short_value != None {")
	g.indent++
	g.line("values[option.name] = short_value.unwrap()")
	g.indent--
	g.line("} else if idx + 1 < args.length() {")
	g.indent++
	g.line("idx += 1")
	g.line("values[option.name] = args[idx]")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("error = Some(\"missing value for -\" + short)")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.indent--
	g.line("})")
	g.line("if !matched { error = Some(\"unknown option -\" + short) }")
	g.line("idx += 1")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("positional_values.push(arg)")
	g.line("idx += 1")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.line("command.options.each((option) => {")
	g.indent++
	g.line("if option.required && error == None && !help {")
	g.indent++
	g.line("if option.value_name.is_empty() {")
	g.indent++
	g.line("if !flags.contains(option.name) { error = Some(\"missing required option --\" + option.name) }")
	g.indent--
	g.line("} else {")
	g.indent++
	g.line("if !values.contains(option.name) { error = Some(\"missing required option --\" + option.name) }")
	g.indent--
	g.line("}")
	g.indent--
	g.line("}")
	g.indent--
	g.line("})")
	g.line("let mut pos = 0")
	g.line("command.arguments.each((argument) => {")
	g.indent++
	g.line("if pos < positional_values.length() {")
	g.indent++
	g.line("positionals[argument.name] = positional_values[pos]")
	g.line("pos += 1")
	g.indent--
	g.line("} else if argument.required && error == None && !help {")
	g.indent++
	g.line("error = Some(\"missing argument \" + argument.name)")
	g.indent--
	g.line("}")
	g.indent--
	g.line("})")
	g.line("CliParseResult::{ command, values, flags, positionals, explicit_options, args, rest, help, error }")
	g.indent--
	g.line("}")
	g.line("")
	g.line("fn rune_cli_help(command : CliCommand) -> String {")
	g.indent++
	g.line("let mut usage = \"Usage: \" + command.name")
	g.line("if command.options.length() > 0 { usage = usage + \" [options]\" }")
	g.line("command.arguments.each((argument) => {")
	g.indent++
	g.line("usage = usage + \" \" + (if argument.required { \"<\" + argument.name + \">\" } else { \"[\" + argument.name + \"]\" })")
	g.indent--
	g.line("})")
	g.line("let mut out = usage + \"\\n\" + command.about")
	g.line("if command.options.length() > 0 {")
	g.indent++
	g.line("out = out + \"\\nOptions:\"")
	g.line("command.options.each((option) => {")
	g.indent++
	g.line("let value = (if option.value_name.is_empty() { \"\" } else { \" <\" + option.value_name + \">\" })")
	g.line("let short = (if option.short.is_empty() { \"\" } else { \"-\" + option.short + \", \" })")
	g.line("out = out + \"\\n  \" + short + \"--\" + option.name + value")
	g.indent--
	g.line("})")
	g.indent--
	g.line("}")
	g.line("out")
	g.indent--
	g.line("}")
}
