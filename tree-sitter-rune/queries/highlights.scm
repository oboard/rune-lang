[
  "=>"
  ":="
  "~="
  "="
] @operator

(module_identifier) @module
(function_definition name: (identifier) @function)
(call_expression function: (identifier) @function.call)
(parameter name: (identifier) @variable.parameter)
(parameter type: (identifier) @type)
(number) @number
(string) @string
(line_comment) @comment
(block_comment) @comment
