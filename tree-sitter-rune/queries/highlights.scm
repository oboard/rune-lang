[
  "=>"
  "->"
  ":="
  "~="
  "="
] @operator

(module_identifier) @module
(type_definition name: (identifier) @type)
(type_parameter_list (identifier) @type.parameter)
(type_field name: (identifier) @property)
(enum_member name: (identifier) @constant)
(function_definition name: (identifier) @function)
(macro_definition "#" @punctuation.special)
(macro_definition name: (identifier) @function.macro)
(macro_invocation "#" @punctuation.special)
(macro_invocation module: (identifier) @module)
(macro_invocation name: (identifier) @function.macro)
(call_expression function: (identifier) @function.call)
(parameter name: (identifier) @variable.parameter)
(type_name (identifier) @type)
(selector_expression "." @punctuation.delimiter)
(selector_expression receiver: (identifier) @variable)
(selector_expression name: (identifier) @property)
(number) @number
(string) @string
(regex) @string.regexp
(regex "/" @punctuation.delimiter)
(regex_escape) @constant.character.escape
(regex_char_class) @constant.other.character-class.regexp
(regex_flags) @keyword.other.regex-options
(line_comment) @comment
(block_comment) @comment
