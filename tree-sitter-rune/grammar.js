module.exports = grammar({
  name: "rune",

  extras: ($) => [
    /\s/,
    $.line_comment,
    $.block_comment
  ],

  rules: {
    source_file: ($) => repeat(choice(
      $.macro_definition,
      $.macro_attributed_declaration,
      $.trait_definition,
      $.type_definition,
      $.function_definition
    )),

    macro_definition: ($) => seq(
      "#",
      field("name", $.identifier),
      field("parameters", $.parameter_list),
      optional(field("return_type", $.return_type)),
      "=>",
      field("body", choice($.expression, $.block))
    ),

    macro_attributed_declaration: ($) => seq(
      repeat1($.macro_invocation),
      choice($.type_definition, $.function_definition)
    ),

    macro_invocation: ($) => seq(
      "#",
      optional(seq(field("module", $.identifier), ".")),
      field("name", $.identifier),
      optional(seq(
        "(",
        optional(seq($.expression, repeat(seq(",", $.expression)), optional(","))),
        ")"
      ))
    ),

    public_visibility: () => "+",
    private_visibility: () => "-",

    trait_definition: ($) => seq(
      "&",
      field("name", $.identifier),
      ":",
      "{",
      repeat(choice($.trait_field, $.trait_method)),
      "}"
    ),

    trait_field: ($) => seq(
      field("name", $.identifier),
      ":",
      field("type", $.type_name)
    ),

    trait_method: ($) => seq(
      optional("static"),
      field("name", $.identifier),
      field("parameters", $.parameter_list),
      optional(field("return_type", $.return_type))
    ),

    type_definition: ($) => seq(
      optional($.public_visibility),
      field("name", $.identifier),
      optional(field("type_parameters", $.type_parameter_list)),
      ":",
      "{",
      repeat(choice($.enum_member, $.type_field, $.method_definition)),
      "}"
    ),

    type_parameter_list: ($) => seq(
      "[",
      optional(seq($.type_parameter, repeat(seq(",", $.type_parameter)), optional(","))),
      "]"
    ),

    type_parameter: ($) => seq(
      field("name", $.identifier),
      optional(seq(":", field("constraint", $.type_name)))
    ),

    type_field: ($) => seq(
      optional($.private_visibility),
      field("name", $.identifier),
      ":",
      field("type", $.type_name)
    ),

    enum_member: ($) => seq(
      optional($.public_visibility),
      field("name", $.identifier),
      "=",
      field("value", choice($.number, seq("-", $.number)))
    ),

    function_definition: ($) => seq(
      optional($.public_visibility),
      field("name", $.identifier),
      field("parameters", $.parameter_list),
      optional(field("return_type", $.return_type)),
      "=>",
      field("body", choice($.expression, $.block))
    ),

    method_definition: ($) => seq(
      optional($.private_visibility),
      optional("static"),
      field("name", $.identifier),
      field("parameters", $.parameter_list),
      optional(field("return_type", $.return_type)),
      "=>",
      field("body", choice($.expression, $.block))
    ),

    parameter_list: ($) => seq(
      "(",
      optional(seq($.parameter, repeat(seq(",", $.parameter)), optional(","))),
      ")"
    ),

    parameter: ($) => seq(
      field("name", $.identifier),
      ":",
      field("type", $.type_name)
    ),

    return_type: ($) => seq("->", $.type_name),

    type_name: ($) => seq(
      optional("&"),
      $.identifier,
      optional($.type_argument_list)
    ),

    type_argument_list: ($) => seq(
      "[",
      optional(seq($.type_name, repeat(seq(",", $.type_name)), optional(","))),
      "]"
    ),

    block: ($) => seq(
      "{",
      repeat(choice($.pattern_branch, $.expression)),
      "}"
    ),

    pattern_branch: ($) => seq($.pattern, "=>", $.expression),

    pattern: ($) => choice(
      "_",
      $.selector_expression,
      $.number,
      $.string,
      $.char,
      seq(choice("<", "<=", ">", ">="), $.number)
    ),

    expression: ($) => choice(
      $.call_expression,
      $.selector_expression,
      $.identifier,
      $.module_identifier,
      $.number,
      $.string,
      $.char,
      $.regex,
      $.binary_expression
    ),

    call_expression: ($) => prec(3, seq(
      field("function", choice($.identifier, $.selector_expression)),
      "(",
      optional(seq($.expression, repeat(seq(",", $.expression)), optional(","))),
      ")"
    )),

    selector_expression: ($) => prec(4, seq(
      field("receiver", choice($.identifier, $.module_identifier)),
      field("operator", choice(".", "::")),
      field("name", $.identifier)
    )),

    binary_expression: ($) => choice(
      prec.left(1, seq($.expression, choice("==", "!=", "<", "<=", ">", ">="), $.expression)),
      prec.left(2, seq($.expression, choice("+", "-"), $.expression)),
      prec.left(3, seq($.expression, choice("*", "/", "%"), $.expression))
    ),

    module_identifier: ($) => seq("@", $.identifier),

    identifier: () => /[A-Za-z_][A-Za-z0-9_]*/,
    number: () => /\d+/,
    string: () => /"([^"\\]|\\.)*"/,
    char: () => /'([^'\\\n]|\\.)*'/,
    regex: ($) => seq(
      field("open", "/"),
      repeat1(choice($.regex_text, $.regex_escape, $.regex_char_class)),
      field("close", "/"),
      optional(field("flags", $.regex_flags))
    ),
    regex_text: () => token.immediate(/[^\/\\\n\[]+/),
    regex_escape: () => token.immediate(/\\./),
    regex_char_class: () => token.immediate(/\[([^\]\\\n]|\\.)*\]/),
    regex_flags: () => token.immediate(/[A-Za-z]+/),
    line_comment: () => token(seq("//", /.*/)),
    block_comment: () => token(seq("/*", /[^*]*\*+([^/*][^*]*\*+)*/, "/"))
  }
});
