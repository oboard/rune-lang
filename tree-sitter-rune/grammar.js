module.exports = grammar({
  name: "rune",

  extras: ($) => [
    /\s/,
    $.line_comment,
    $.block_comment
  ],

  rules: {
    source_file: ($) => repeat($.function_definition),

    function_definition: ($) => seq(
      field("name", $.identifier),
      field("parameters", $.parameter_list),
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
      field("type", $.identifier)
    ),

    block: ($) => seq(
      "{",
      repeat(choice($.pattern_branch, $.expression)),
      "}"
    ),

    pattern_branch: ($) => seq($.pattern, "=>", $.expression),

    pattern: ($) => choice(
      "_",
      $.number,
      $.string,
      seq(choice("<", "<=", ">", ">="), $.number)
    ),

    expression: ($) => choice(
      $.call_expression,
      $.selector_expression,
      $.identifier,
      $.module_identifier,
      $.number,
      $.string,
      $.binary_expression
    ),

    call_expression: ($) => prec(3, seq(
      field("function", choice($.identifier, $.selector_expression)),
      "(",
      optional(seq($.expression, repeat(seq(",", $.expression)), optional(","))),
      ")"
    )),

    selector_expression: ($) => prec(4, seq(
      choice($.identifier, $.module_identifier),
      ".",
      $.identifier
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
    line_comment: () => token(seq("//", /.*/)),
    block_comment: () => token(seq("/*", /[^*]*\*+([^/*][^*]*\*+)*/, "/"))
  }
});
