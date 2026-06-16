[
  "=>"
  "->"
  ":="
  "~="
  "="
] @operator

(module_identifier) @module
(trait_definition "&" @punctuation.special)
(trait_definition name: (identifier) @type)
(trait_field name: (identifier) @property)
(trait_method name: (identifier) @function.method)
(trait_method "static" @keyword)
(method_definition "static" @keyword)
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
(type_name "&" @punctuation.special)
(type_name (identifier) @type)
(selector_expression
  operator: "." @punctuation.delimiter
  receiver: (identifier) @variable
  name: (identifier) @property)
(selector_expression
  operator: "::" @punctuation.delimiter
  receiver: (identifier) @type
  name: (identifier) @function.method)
(number) @number
(string) @string
(regex) @string.regexp
(regex "/" @punctuation.delimiter)
(regex_escape) @constant.character.escape
(regex_char_class) @constant.other.character-class.regexp
(regex_flags) @keyword.other.regex-options
(line_comment) @comment
(block_comment) @comment
