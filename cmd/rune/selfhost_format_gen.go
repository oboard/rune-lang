package main

import (
	"strings"
)

type __fmt_TokenKind int

const (
	__fmt_TokenKind_EOF                __fmt_TokenKind = 0
	__fmt_TokenKind_Illegal            __fmt_TokenKind = 1
	__fmt_TokenKind_Newline            __fmt_TokenKind = 2
	__fmt_TokenKind_Ident              __fmt_TokenKind = 3
	__fmt_TokenKind_Int                __fmt_TokenKind = 4
	__fmt_TokenKind_Double             __fmt_TokenKind = 5
	__fmt_TokenKind_BigInt             __fmt_TokenKind = 6
	__fmt_TokenKind_String             __fmt_TokenKind = 7
	__fmt_TokenKind_TemplateString     __fmt_TokenKind = 8
	__fmt_TokenKind_Char               __fmt_TokenKind = 9
	__fmt_TokenKind_Regex              __fmt_TokenKind = 10
	__fmt_TokenKind_XMLText            __fmt_TokenKind = 11
	__fmt_TokenKind_Hash               __fmt_TokenKind = 12
	__fmt_TokenKind_At                 __fmt_TokenKind = 13
	__fmt_TokenKind_Dollar             __fmt_TokenKind = 14
	__fmt_TokenKind_Dot                __fmt_TokenKind = 15
	__fmt_TokenKind_DotDot             __fmt_TokenKind = 16
	__fmt_TokenKind_DotDotLess         __fmt_TokenKind = 17
	__fmt_TokenKind_DotDotEqual        __fmt_TokenKind = 18
	__fmt_TokenKind_DotDotDot          __fmt_TokenKind = 19
	__fmt_TokenKind_Comma              __fmt_TokenKind = 20
	__fmt_TokenKind_Colon              __fmt_TokenKind = 21
	__fmt_TokenKind_DoubleColon        __fmt_TokenKind = 22
	__fmt_TokenKind_LParen             __fmt_TokenKind = 23
	__fmt_TokenKind_RParen             __fmt_TokenKind = 24
	__fmt_TokenKind_LBracket           __fmt_TokenKind = 25
	__fmt_TokenKind_RBracket           __fmt_TokenKind = 26
	__fmt_TokenKind_LBrace             __fmt_TokenKind = 27
	__fmt_TokenKind_RBrace             __fmt_TokenKind = 28
	__fmt_TokenKind_Question           __fmt_TokenKind = 29
	__fmt_TokenKind_QuestionQuestion   __fmt_TokenKind = 30
	__fmt_TokenKind_Apostrophe         __fmt_TokenKind = 31
	__fmt_TokenKind_FatArrow           __fmt_TokenKind = 32
	__fmt_TokenKind_Assign             __fmt_TokenKind = 33
	__fmt_TokenKind_Declare            __fmt_TokenKind = 34
	__fmt_TokenKind_MutDeclare         __fmt_TokenKind = 35
	__fmt_TokenKind_Arrow              __fmt_TokenKind = 36
	__fmt_TokenKind_Plus               __fmt_TokenKind = 37
	__fmt_TokenKind_PlusPlus           __fmt_TokenKind = 38
	__fmt_TokenKind_Minus              __fmt_TokenKind = 39
	__fmt_TokenKind_Star               __fmt_TokenKind = 40
	__fmt_TokenKind_Slash              __fmt_TokenKind = 41
	__fmt_TokenKind_Percent            __fmt_TokenKind = 42
	__fmt_TokenKind_Bang               __fmt_TokenKind = 43
	__fmt_TokenKind_Tilde              __fmt_TokenKind = 44
	__fmt_TokenKind_BitAnd             __fmt_TokenKind = 45
	__fmt_TokenKind_BitOr              __fmt_TokenKind = 46
	__fmt_TokenKind_BitXor             __fmt_TokenKind = 47
	__fmt_TokenKind_ShiftLeft          __fmt_TokenKind = 48
	__fmt_TokenKind_ShiftRight         __fmt_TokenKind = 49
	__fmt_TokenKind_UnsignedShiftRight __fmt_TokenKind = 50
	__fmt_TokenKind_AndAnd             __fmt_TokenKind = 51
	__fmt_TokenKind_OrOr               __fmt_TokenKind = 52
	__fmt_TokenKind_EqualEqual         __fmt_TokenKind = 53
	__fmt_TokenKind_BangEqual          __fmt_TokenKind = 54
	__fmt_TokenKind_Less               __fmt_TokenKind = 55
	__fmt_TokenKind_LessEqual          __fmt_TokenKind = 56
	__fmt_TokenKind_Greater            __fmt_TokenKind = 57
	__fmt_TokenKind_GreaterEqual       __fmt_TokenKind = 58
	__fmt_TokenKind_Underscore         __fmt_TokenKind = 59
)

type __fmt_Token struct {
	__fmt_kind   __fmt_TokenKind
	__fmt_lexeme string
	__fmt_offset int
	__fmt_line   int
	__fmt_column int
}

type __fmt_LexState struct {
	__fmt_source        string
	__fmt_start         int
	__fmt_current       int
	__fmt_line          int
	__fmt_column        int
	__fmt_startLine     int
	__fmt_startColumn   int
	__fmt_canStartRegex bool
}

type __fmt_Advanced struct {
	__fmt_state __fmt_LexState
	__fmt_ch    rune
}

type __fmt_Lexed struct {
	__fmt_state __fmt_LexState
	__fmt_kind  __fmt_TokenKind
}

type __fmt_ScannedString struct {
	__fmt_state __fmt_LexState
	__fmt_ok    bool
}

type __fmt_FormatState struct {
	__fmt_text      string
	__fmt_indent    int
	__fmt_lineStart bool
	__fmt_previous  __fmt_TokenKind
}

func __fmt_lex(__fmt_source string) []__fmt_Token {
	return __fmt___fmt_rune_private_8093ad48_scan(__fmt_LexState{__fmt_source: __fmt_source, __fmt_start: 0, __fmt_current: 0, __fmt_line: 1, __fmt_column: 1, __fmt_startLine: 1, __fmt_startColumn: 1, __fmt_canStartRegex: true}, __fmt___fmt_rune_private_8093ad48_emptyTokens())
}

func __fmt___fmt_rune_private_8093ad48_emptyTokens() []__fmt_Token {
	return append([]__fmt_Token{}, []__fmt_Token{__fmt_Token{__fmt_kind: __fmt_TokenKind_EOF, __fmt_lexeme: "", __fmt_offset: 0, __fmt_line: 0, __fmt_column: 0}}[0:0]...)
}

func __fmt___fmt_rune_private_8093ad48_scan(__fmt_state __fmt_LexState, __fmt_tokens []__fmt_Token) []__fmt_Token {
	__fmt_skipped := __fmt___fmt_rune_private_8093ad48_skipIgnored(__fmt_state)
	__fmt_started := __fmt___fmt_rune_private_8093ad48_markStart(__fmt_skipped)
	return func() []__fmt_Token {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_started) {
			return __fmt___fmt_rune_private_8093ad48_appendToken(__fmt_tokens, __fmt___fmt_rune_private_8093ad48_makeToken(__fmt_started, __fmt_TokenKind_EOF))
		}
		return __fmt___fmt_rune_private_8093ad48_scanLexed(__fmt___fmt_rune_private_8093ad48_scanToken(__fmt___fmt_rune_private_8093ad48_advance(__fmt_started)), __fmt_tokens)
	}()
}

func __fmt___fmt_rune_private_8093ad48_scanLexed(__fmt_lexed __fmt_Lexed, __fmt_tokens []__fmt_Token) []__fmt_Token {
	__fmt_nextTokens := __fmt___fmt_rune_private_8093ad48_appendToken(__fmt_tokens, __fmt___fmt_rune_private_8093ad48_makeToken(__fmt_lexed.__fmt_state, __fmt_lexed.__fmt_kind))
	return __fmt___fmt_rune_private_8093ad48_scan(__fmt___fmt_rune_private_8093ad48_finishToken(__fmt_lexed.__fmt_state, __fmt_lexed.__fmt_kind), __fmt_nextTokens)
}

func __fmt___fmt_rune_private_8093ad48_appendToken(__fmt_tokens []__fmt_Token, __fmt_token __fmt_Token) []__fmt_Token {
	__fmt_tokens = append(__fmt_tokens, __fmt_token)
	return __fmt_tokens
}

func __fmt___fmt_rune_private_8093ad48_makeToken(__fmt_state __fmt_LexState, __fmt_kind __fmt_TokenKind) __fmt_Token {
	return __fmt_Token{__fmt_kind: __fmt_kind, __fmt_lexeme: func() string {
		runes := []rune(__fmt_state.__fmt_source)
		return string(runes[__fmt_state.__fmt_start:__fmt_state.__fmt_current])
	}(), __fmt_offset: __fmt_state.__fmt_start, __fmt_line: __fmt_state.__fmt_startLine, __fmt_column: __fmt_state.__fmt_startColumn}
}

func __fmt___fmt_rune_private_8093ad48_finishToken(__fmt_state __fmt_LexState, __fmt_kind __fmt_TokenKind) __fmt_LexState {
	return __fmt_LexState{__fmt_source: __fmt_state.__fmt_source, __fmt_start: __fmt_state.__fmt_start, __fmt_current: __fmt_state.__fmt_current, __fmt_line: __fmt_state.__fmt_line, __fmt_column: __fmt_state.__fmt_column, __fmt_startLine: __fmt_state.__fmt_startLine, __fmt_startColumn: __fmt_state.__fmt_startColumn, __fmt_canStartRegex: !__fmt___fmt_rune_private_8093ad48_canEndExpression(__fmt_state, __fmt_kind)}
}

func __fmt___fmt_rune_private_8093ad48_canEndExpression(__fmt_state __fmt_LexState, __fmt_kind __fmt_TokenKind) bool {
	return __fmt___fmt_rune_private_8093ad48_canEndValueToken(__fmt_kind) || __fmt___fmt_rune_private_8093ad48_canEndXmlLess(__fmt_state, __fmt_kind)
}

func __fmt___fmt_rune_private_8093ad48_canEndValueToken(__fmt_kind __fmt_TokenKind) bool {
	return func() bool {
		switch {
		case (__fmt_kind == __fmt_TokenKind_Ident) || (__fmt_kind == __fmt_TokenKind_Int) || (__fmt_kind == __fmt_TokenKind_Double) || (__fmt_kind == __fmt_TokenKind_BigInt) || (__fmt_kind == __fmt_TokenKind_String) || (__fmt_kind == __fmt_TokenKind_TemplateString) || (__fmt_kind == __fmt_TokenKind_Char) || (__fmt_kind == __fmt_TokenKind_Regex) || (__fmt_kind == __fmt_TokenKind_XMLText) || (__fmt_kind == __fmt_TokenKind_RParen) || (__fmt_kind == __fmt_TokenKind_RBracket) || (__fmt_kind == __fmt_TokenKind_RBrace):
			return true
		default:
			return false
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_canEndXmlLess(__fmt_state __fmt_LexState, __fmt_kind __fmt_TokenKind) bool {
	return __fmt_kind == __fmt_TokenKind_Less && (__fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '/' || __fmt___fmt_rune_private_8093ad48_isIdentStart(__fmt___fmt_rune_private_8093ad48_peek(__fmt_state)))
}

func __fmt_tokenKindName(__fmt_kind __fmt_TokenKind) string {
	return func() string {
		switch {
		case __fmt_kind == __fmt_TokenKind_EOF:
			return "EOF"
		case __fmt_kind == __fmt_TokenKind_Illegal:
			return "Illegal"
		case __fmt_kind == __fmt_TokenKind_Newline:
			return "Newline"
		case __fmt_kind == __fmt_TokenKind_Ident:
			return "Ident"
		case __fmt_kind == __fmt_TokenKind_Int:
			return "Int"
		case __fmt_kind == __fmt_TokenKind_Double:
			return "Double"
		case __fmt_kind == __fmt_TokenKind_BigInt:
			return "BigInt"
		case __fmt_kind == __fmt_TokenKind_String:
			return "String"
		case __fmt_kind == __fmt_TokenKind_TemplateString:
			return "TemplateString"
		case __fmt_kind == __fmt_TokenKind_Char:
			return "Char"
		case __fmt_kind == __fmt_TokenKind_Regex:
			return "Regex"
		case __fmt_kind == __fmt_TokenKind_XMLText:
			return "XMLText"
		case __fmt_kind == __fmt_TokenKind_Hash:
			return "Hash"
		case __fmt_kind == __fmt_TokenKind_At:
			return "At"
		case __fmt_kind == __fmt_TokenKind_Dollar:
			return "Dollar"
		case __fmt_kind == __fmt_TokenKind_Dot:
			return "Dot"
		case __fmt_kind == __fmt_TokenKind_DotDot:
			return "DotDot"
		case __fmt_kind == __fmt_TokenKind_DotDotLess:
			return "DotDotLess"
		case __fmt_kind == __fmt_TokenKind_DotDotEqual:
			return "DotDotEqual"
		case __fmt_kind == __fmt_TokenKind_DotDotDot:
			return "DotDotDot"
		case __fmt_kind == __fmt_TokenKind_Comma:
			return "Comma"
		case __fmt_kind == __fmt_TokenKind_Colon:
			return "Colon"
		case __fmt_kind == __fmt_TokenKind_DoubleColon:
			return "DoubleColon"
		case __fmt_kind == __fmt_TokenKind_LParen:
			return "LParen"
		case __fmt_kind == __fmt_TokenKind_RParen:
			return "RParen"
		case __fmt_kind == __fmt_TokenKind_LBracket:
			return "LBracket"
		case __fmt_kind == __fmt_TokenKind_RBracket:
			return "RBracket"
		case __fmt_kind == __fmt_TokenKind_LBrace:
			return "LBrace"
		case __fmt_kind == __fmt_TokenKind_RBrace:
			return "RBrace"
		case __fmt_kind == __fmt_TokenKind_Question:
			return "Question"
		case __fmt_kind == __fmt_TokenKind_QuestionQuestion:
			return "QuestionQuestion"
		case __fmt_kind == __fmt_TokenKind_Apostrophe:
			return "Apostrophe"
		case __fmt_kind == __fmt_TokenKind_FatArrow:
			return "FatArrow"
		case __fmt_kind == __fmt_TokenKind_Assign:
			return "Assign"
		case __fmt_kind == __fmt_TokenKind_Declare:
			return "Declare"
		case __fmt_kind == __fmt_TokenKind_MutDeclare:
			return "MutDeclare"
		case __fmt_kind == __fmt_TokenKind_Arrow:
			return "Arrow"
		case __fmt_kind == __fmt_TokenKind_Plus:
			return "Plus"
		case __fmt_kind == __fmt_TokenKind_PlusPlus:
			return "PlusPlus"
		case __fmt_kind == __fmt_TokenKind_Minus:
			return "Minus"
		case __fmt_kind == __fmt_TokenKind_Star:
			return "Star"
		case __fmt_kind == __fmt_TokenKind_Slash:
			return "Slash"
		case __fmt_kind == __fmt_TokenKind_Percent:
			return "Percent"
		case __fmt_kind == __fmt_TokenKind_Bang:
			return "Bang"
		case __fmt_kind == __fmt_TokenKind_Tilde:
			return "Tilde"
		case __fmt_kind == __fmt_TokenKind_BitAnd:
			return "BitAnd"
		case __fmt_kind == __fmt_TokenKind_BitOr:
			return "BitOr"
		case __fmt_kind == __fmt_TokenKind_BitXor:
			return "BitXor"
		case __fmt_kind == __fmt_TokenKind_ShiftLeft:
			return "ShiftLeft"
		case __fmt_kind == __fmt_TokenKind_ShiftRight:
			return "ShiftRight"
		case __fmt_kind == __fmt_TokenKind_UnsignedShiftRight:
			return "UnsignedShiftRight"
		case __fmt_kind == __fmt_TokenKind_AndAnd:
			return "AndAnd"
		case __fmt_kind == __fmt_TokenKind_OrOr:
			return "OrOr"
		case __fmt_kind == __fmt_TokenKind_EqualEqual:
			return "EqualEqual"
		case __fmt_kind == __fmt_TokenKind_BangEqual:
			return "BangEqual"
		case __fmt_kind == __fmt_TokenKind_Less:
			return "Less"
		case __fmt_kind == __fmt_TokenKind_LessEqual:
			return "LessEqual"
		case __fmt_kind == __fmt_TokenKind_Greater:
			return "Greater"
		case __fmt_kind == __fmt_TokenKind_GreaterEqual:
			return "GreaterEqual"
		case __fmt_kind == __fmt_TokenKind_Underscore:
			return "Underscore"
		default:
			return "Unknown"
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_markStart(__fmt_state __fmt_LexState) __fmt_LexState {
	return __fmt_LexState{__fmt_source: __fmt_state.__fmt_source, __fmt_start: __fmt_state.__fmt_current, __fmt_current: __fmt_state.__fmt_current, __fmt_line: __fmt_state.__fmt_line, __fmt_column: __fmt_state.__fmt_column, __fmt_startLine: __fmt_state.__fmt_line, __fmt_startColumn: __fmt_state.__fmt_column, __fmt_canStartRegex: __fmt_state.__fmt_canStartRegex}
}

func __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state __fmt_LexState) bool {
	return __fmt_state.__fmt_current >= len([]rune(__fmt_state.__fmt_source))
}

func __fmt___fmt_rune_private_8093ad48_charAt(__fmt_source string, __fmt_index int) rune {
	return func() rune {
		if __fmt_index < 0 || __fmt_index >= len([]rune(__fmt_source)) {
			return ' '
		}
		return []rune(__fmt_source)[__fmt_index]
	}()
}

func __fmt___fmt_rune_private_8093ad48_peek(__fmt_state __fmt_LexState) rune {
	return __fmt___fmt_rune_private_8093ad48_charAt(__fmt_state.__fmt_source, __fmt_state.__fmt_current)
}

func __fmt___fmt_rune_private_8093ad48_peekNext(__fmt_state __fmt_LexState) rune {
	return __fmt___fmt_rune_private_8093ad48_charAt(__fmt_state.__fmt_source, __fmt_state.__fmt_current+1)
}

func __fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state __fmt_LexState) __fmt_LexState {
	return __fmt___fmt_rune_private_8093ad48_advance(__fmt_state).__fmt_state
}

func __fmt___fmt_rune_private_8093ad48_advance(__fmt_state __fmt_LexState) __fmt_Advanced {
	return func() __fmt_Advanced {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state) {
			return __fmt___fmt_rune_private_8093ad48_advanced(__fmt_state, ' ')
		}
		return __fmt___fmt_rune_private_8093ad48_advanceChar(__fmt_state, []rune(__fmt_state.__fmt_source)[__fmt_state.__fmt_current])
	}()
}

func __fmt___fmt_rune_private_8093ad48_advanced(__fmt_state __fmt_LexState, __fmt_ch rune) __fmt_Advanced {
	return __fmt_Advanced{__fmt_state: __fmt_state, __fmt_ch: __fmt_ch}
}

func __fmt___fmt_rune_private_8093ad48_advanceChar(__fmt_state __fmt_LexState, __fmt_ch rune) __fmt_Advanced {
	return __fmt_Advanced{__fmt_state: __fmt_LexState{__fmt_source: __fmt_state.__fmt_source, __fmt_start: __fmt_state.__fmt_start, __fmt_current: __fmt_state.__fmt_current + 1, __fmt_line: func() int {
		if __fmt_ch == '\n' {
			return __fmt_state.__fmt_line + 1
		}
		return __fmt_state.__fmt_line
	}(), __fmt_column: func() int {
		if __fmt_ch == '\n' {
			return 1
		}
		return __fmt_state.__fmt_column + 1
	}(), __fmt_startLine: __fmt_state.__fmt_startLine, __fmt_startColumn: __fmt_state.__fmt_startColumn, __fmt_canStartRegex: __fmt_state.__fmt_canStartRegex}, __fmt_ch: __fmt_ch}
}

func __fmt___fmt_rune_private_8093ad48_skipIgnored(__fmt_state __fmt_LexState) __fmt_LexState {
	return func() __fmt_LexState {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state) {
			return __fmt_state
		}
		return func() __fmt_LexState {
			if __fmt___fmt_rune_private_8093ad48_isSpace(__fmt___fmt_rune_private_8093ad48_peek(__fmt_state)) {
				return __fmt___fmt_rune_private_8093ad48_skipIgnored(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
			}
			return func() __fmt_LexState {
				if __fmt___fmt_rune_private_8093ad48_startsWith(__fmt_state, '/', '/') {
					return __fmt___fmt_rune_private_8093ad48_skipIgnored(__fmt___fmt_rune_private_8093ad48_skipLineComment(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))))
				}
				return func() __fmt_LexState {
					if __fmt___fmt_rune_private_8093ad48_startsWith(__fmt_state, '/', '*') {
						return __fmt___fmt_rune_private_8093ad48_skipIgnored(__fmt___fmt_rune_private_8093ad48_skipBlockComment(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))))
					}
					return __fmt_state
				}()
			}()
		}()
	}()
}

func __fmt___fmt_rune_private_8093ad48_startsWith(__fmt_state __fmt_LexState, __fmt_first rune, __fmt_second rune) bool {
	return __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == __fmt_first && __fmt___fmt_rune_private_8093ad48_peekNext(__fmt_state) == __fmt_second
}

func __fmt___fmt_rune_private_8093ad48_skipLineComment(__fmt_state __fmt_LexState) __fmt_LexState {
	return func() __fmt_LexState {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state) || __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '\n' {
			return __fmt_state
		}
		return __fmt___fmt_rune_private_8093ad48_skipLineComment(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
	}()
}

func __fmt___fmt_rune_private_8093ad48_skipBlockComment(__fmt_state __fmt_LexState) __fmt_LexState {
	return func() __fmt_LexState {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state) {
			return __fmt_state
		}
		return func() __fmt_LexState {
			if __fmt___fmt_rune_private_8093ad48_startsWith(__fmt_state, '*', '/') {
				return __fmt___fmt_rune_private_8093ad48_advanceState(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
			}
			return __fmt___fmt_rune_private_8093ad48_skipBlockComment(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
		}()
	}()
}

func __fmt___fmt_rune_private_8093ad48_isSpace(__fmt_ch rune) bool {
	switch {
	case __fmt_ch == ' ':
		return true
	case __fmt_ch == '\t':
		return true
	case __fmt_ch == '\r':
		return true
	default:
		return false
	}
}

func __fmt___fmt_rune_private_8093ad48_scanToken(__fmt_step __fmt_Advanced) __fmt_Lexed {
	__fmt_destructure1 := __fmt_step
	__fmt_state := __fmt_destructure1.__fmt_state
	__fmt_ch := __fmt_destructure1.__fmt_ch
	return func() __fmt_Lexed {
		switch {
		case __fmt_ch == '\n':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Newline)
		case __fmt_ch == '#':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Hash)
		case __fmt_ch == '@':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_At)
		case __fmt_ch == '$':
			return __fmt___fmt_rune_private_8093ad48_lexDollar(__fmt_state)
		case __fmt_ch == '.':
			return __fmt___fmt_rune_private_8093ad48_lexDot(__fmt_state)
		case __fmt_ch == ',':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Comma)
		case __fmt_ch == ':':
			return __fmt___fmt_rune_private_8093ad48_lexColon(__fmt_state)
		case __fmt_ch == '(':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_LParen)
		case __fmt_ch == ')':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_RParen)
		case __fmt_ch == '[':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_LBracket)
		case __fmt_ch == ']':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_RBracket)
		case __fmt_ch == '{':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_LBrace)
		case __fmt_ch == '}':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_RBrace)
		case __fmt_ch == '?':
			return __fmt___fmt_rune_private_8093ad48_lexQuestion(__fmt_state)
		case __fmt_ch == '+':
			return __fmt___fmt_rune_private_8093ad48_lexPlus(__fmt_state)
		case __fmt_ch == '-':
			return __fmt___fmt_rune_private_8093ad48_lexMinus(__fmt_state)
		case __fmt_ch == '*':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Star)
		case __fmt_ch == '/':
			return func() __fmt_Lexed {
				if __fmt_state.__fmt_canStartRegex {
					return __fmt___fmt_rune_private_8093ad48_lexRegexToken(__fmt_state)
				}
				return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Slash)
			}()
		case __fmt_ch == '%':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Percent)
		case __fmt_ch == '!':
			return __fmt___fmt_rune_private_8093ad48_lexBang(__fmt_state)
		case __fmt_ch == '~':
			return __fmt___fmt_rune_private_8093ad48_lexTilde(__fmt_state)
		case __fmt_ch == '&':
			return __fmt___fmt_rune_private_8093ad48_lexAmp(__fmt_state)
		case __fmt_ch == '|':
			return __fmt___fmt_rune_private_8093ad48_lexPipe(__fmt_state)
		case __fmt_ch == '^':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_BitXor)
		case __fmt_ch == '=':
			return __fmt___fmt_rune_private_8093ad48_lexEqual(__fmt_state)
		case __fmt_ch == '<':
			return __fmt___fmt_rune_private_8093ad48_lexLess(__fmt_state)
		case __fmt_ch == '>':
			return __fmt___fmt_rune_private_8093ad48_lexGreater(__fmt_state)
		case __fmt_ch == '"':
			return __fmt___fmt_rune_private_8093ad48_lexStringToken(__fmt_state)
		case __fmt_ch == '`':
			return __fmt___fmt_rune_private_8093ad48_lexTemplateStringToken(__fmt_state)
		case __fmt_ch == '\'':
			return func() __fmt_Lexed {
				if __fmt_state.__fmt_canStartRegex {
					return __fmt___fmt_rune_private_8093ad48_lexCharToken(__fmt_state)
				}
				return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Apostrophe)
			}()
		case __fmt_ch == '_':
			return func() __fmt_Lexed {
				if __fmt___fmt_rune_private_8093ad48_isIdentContinue(__fmt___fmt_rune_private_8093ad48_peek(__fmt_state)) {
					return __fmt___fmt_rune_private_8093ad48_lexIdentifierToken(__fmt_state)
				}
				return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Underscore)
			}()
		default:
			return func() __fmt_Lexed {
				if __fmt___fmt_rune_private_8093ad48_isDigit(__fmt_ch) {
					return __fmt___fmt_rune_private_8093ad48_lexNumberToken(__fmt_state)
				}
				return func() __fmt_Lexed {
					if __fmt___fmt_rune_private_8093ad48_isIdentStart(__fmt_ch) {
						return __fmt___fmt_rune_private_8093ad48_lexIdentifierToken(__fmt_state)
					}
					return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Illegal)
				}()
			}()
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state __fmt_LexState, __fmt_kind __fmt_TokenKind) __fmt_Lexed {
	return __fmt_Lexed{__fmt_state: __fmt_state, __fmt_kind: __fmt_kind}
}

func __fmt___fmt_rune_private_8093ad48_lexDot(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '.':
			return __fmt___fmt_rune_private_8093ad48_lexDotDot(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Dot)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexDotDot(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '.':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_DotDotDot)
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '<':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_DotDotLess)
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '=':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_DotDotEqual)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_DotDot)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexColon(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '=':
			return func() __fmt_Lexed {
				if __fmt___fmt_rune_private_8093ad48_peek(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state)) == ':' {
					return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state)), __fmt_TokenKind_MutDeclare)
				}
				return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_Declare)
			}()
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == ':':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_DoubleColon)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Colon)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexQuestion(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '?':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_QuestionQuestion)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Question)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexPlus(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '+':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_PlusPlus)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Plus)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexMinus(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '>':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_Arrow)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Minus)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexBang(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '=':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_BangEqual)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Bang)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexTilde(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Tilde)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexDollar(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Dollar)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexAmp(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '&':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_AndAnd)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_BitAnd)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexPipe(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '|':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_OrOr)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_BitOr)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexEqual(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '>':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_FatArrow)
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '=':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_EqualEqual)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Assign)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexLess(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '=':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_LessEqual)
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '<':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_ShiftLeft)
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Less)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexGreater(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '=':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_GreaterEqual)
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '>':
			return func() __fmt_Lexed {
				if __fmt___fmt_rune_private_8093ad48_startsWith(__fmt_state, '>', '>') {
					return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state)), __fmt_TokenKind_UnsignedShiftRight)
				}
				return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_ShiftRight)
			}()
		default:
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, __fmt_TokenKind_Greater)
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexStringToken(__fmt_state __fmt_LexState) __fmt_Lexed {
	__fmt_scanned := __fmt___fmt_rune_private_8093ad48_scanString(__fmt_state, false)
	return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_scanned.__fmt_state, func() __fmt_TokenKind {
		if __fmt_scanned.__fmt_ok {
			return __fmt_TokenKind_String
		}
		return __fmt_TokenKind_Illegal
	}())
}

func __fmt___fmt_rune_private_8093ad48_lexTemplateStringToken(__fmt_state __fmt_LexState) __fmt_Lexed {
	__fmt_scanned := __fmt___fmt_rune_private_8093ad48_scanTemplateString(__fmt_state, false)
	return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_scanned.__fmt_state, func() __fmt_TokenKind {
		if __fmt_scanned.__fmt_ok {
			return __fmt_TokenKind_TemplateString
		}
		return __fmt_TokenKind_Illegal
	}())
}

func __fmt___fmt_rune_private_8093ad48_scanString(__fmt_state __fmt_LexState, __fmt_escaped bool) __fmt_ScannedString {
	return func() __fmt_ScannedString {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state) {
			return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_state, false)
		}
		return __fmt___fmt_rune_private_8093ad48_scanStringStep(__fmt___fmt_rune_private_8093ad48_advance(__fmt_state), __fmt_escaped)
	}()
}

func __fmt___fmt_rune_private_8093ad48_scanTemplateString(__fmt_state __fmt_LexState, __fmt_escaped bool) __fmt_ScannedString {
	return func() __fmt_ScannedString {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state) {
			return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_state, false)
		}
		return __fmt___fmt_rune_private_8093ad48_scanTemplateStringStep(__fmt___fmt_rune_private_8093ad48_advance(__fmt_state), __fmt_escaped)
	}()
}

func __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_state __fmt_LexState, __fmt_ok bool) __fmt_ScannedString {
	return __fmt_ScannedString{__fmt_state: __fmt_state, __fmt_ok: __fmt_ok}
}

func __fmt___fmt_rune_private_8093ad48_scanStringStep(__fmt_step __fmt_Advanced, __fmt_escaped bool) __fmt_ScannedString {
	return func() __fmt_ScannedString {
		if __fmt_escaped {
			return __fmt___fmt_rune_private_8093ad48_scanString(__fmt_step.__fmt_state, false)
		}
		return func() __fmt_ScannedString {
			switch {
			case __fmt_step.__fmt_ch == '\\':
				return __fmt___fmt_rune_private_8093ad48_scanString(__fmt_step.__fmt_state, true)
			case __fmt_step.__fmt_ch == '"':
				return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_step.__fmt_state, true)
			default:
				return __fmt___fmt_rune_private_8093ad48_scanString(__fmt_step.__fmt_state, false)
			}
		}()
	}()
}

func __fmt___fmt_rune_private_8093ad48_scanTemplateStringStep(__fmt_step __fmt_Advanced, __fmt_escaped bool) __fmt_ScannedString {
	return func() __fmt_ScannedString {
		if __fmt_escaped {
			return __fmt___fmt_rune_private_8093ad48_scanTemplateString(__fmt_step.__fmt_state, false)
		}
		return func() __fmt_ScannedString {
			switch {
			case __fmt_step.__fmt_ch == '\\':
				return __fmt___fmt_rune_private_8093ad48_scanTemplateString(__fmt_step.__fmt_state, true)
			case __fmt_step.__fmt_ch == '`':
				return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_step.__fmt_state, true)
			default:
				return __fmt___fmt_rune_private_8093ad48_scanTemplateString(__fmt_step.__fmt_state, false)
			}
		}()
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexCharToken(__fmt_state __fmt_LexState) __fmt_Lexed {
	__fmt_scanned := __fmt___fmt_rune_private_8093ad48_scanChar(__fmt_state, false)
	return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_scanned.__fmt_state, func() __fmt_TokenKind {
		if __fmt_scanned.__fmt_ok {
			return __fmt_TokenKind_Char
		}
		return __fmt_TokenKind_Illegal
	}())
}

func __fmt___fmt_rune_private_8093ad48_scanChar(__fmt_state __fmt_LexState, __fmt_escaped bool) __fmt_ScannedString {
	return func() __fmt_ScannedString {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state) {
			return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_state, false)
		}
		return __fmt___fmt_rune_private_8093ad48_scanCharStep(__fmt___fmt_rune_private_8093ad48_advance(__fmt_state), __fmt_escaped)
	}()
}

func __fmt___fmt_rune_private_8093ad48_scanCharStep(__fmt_step __fmt_Advanced, __fmt_escaped bool) __fmt_ScannedString {
	return func() __fmt_ScannedString {
		switch {
		case __fmt_step.__fmt_ch == '\n':
			return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_step.__fmt_state, false)
		default:
			return func() __fmt_ScannedString {
				if __fmt_escaped {
					return __fmt___fmt_rune_private_8093ad48_scanChar(__fmt_step.__fmt_state, false)
				}
				return func() __fmt_ScannedString {
					switch {
					case __fmt_step.__fmt_ch == '\\':
						return __fmt___fmt_rune_private_8093ad48_scanChar(__fmt_step.__fmt_state, true)
					case __fmt_step.__fmt_ch == '\'':
						return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_step.__fmt_state, true)
					default:
						return __fmt___fmt_rune_private_8093ad48_scanChar(__fmt_step.__fmt_state, false)
					}
				}()
			}()
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexRegexToken(__fmt_state __fmt_LexState) __fmt_Lexed {
	__fmt_scanned := __fmt___fmt_rune_private_8093ad48_scanRegex(__fmt_state, false, false)
	return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_scanned.__fmt_state, func() __fmt_TokenKind {
		if __fmt_scanned.__fmt_ok {
			return __fmt_TokenKind_Regex
		}
		return __fmt_TokenKind_Illegal
	}())
}

func __fmt___fmt_rune_private_8093ad48_scanRegex(__fmt_state __fmt_LexState, __fmt_escaped bool, __fmt_inClass bool) __fmt_ScannedString {
	return func() __fmt_ScannedString {
		if __fmt___fmt_rune_private_8093ad48_atEnd(__fmt_state) {
			return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_state, false)
		}
		return __fmt___fmt_rune_private_8093ad48_scanRegexStep(__fmt___fmt_rune_private_8093ad48_advance(__fmt_state), __fmt_escaped, __fmt_inClass)
	}()
}

func __fmt___fmt_rune_private_8093ad48_scanRegexStep(__fmt_step __fmt_Advanced, __fmt_escaped bool, __fmt_inClass bool) __fmt_ScannedString {
	return func() __fmt_ScannedString {
		switch {
		case __fmt_step.__fmt_ch == '\n':
			return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt_step.__fmt_state, false)
		default:
			return func() __fmt_ScannedString {
				if __fmt_escaped {
					return __fmt___fmt_rune_private_8093ad48_scanRegex(__fmt_step.__fmt_state, false, __fmt_inClass)
				}
				return func() __fmt_ScannedString {
					switch {
					case __fmt_step.__fmt_ch == '\\':
						return __fmt___fmt_rune_private_8093ad48_scanRegex(__fmt_step.__fmt_state, true, __fmt_inClass)
					case __fmt_step.__fmt_ch == '[':
						return __fmt___fmt_rune_private_8093ad48_scanRegex(__fmt_step.__fmt_state, false, true)
					case __fmt_step.__fmt_ch == ']':
						return __fmt___fmt_rune_private_8093ad48_scanRegex(__fmt_step.__fmt_state, false, false)
					case __fmt_step.__fmt_ch == '/':
						return func() __fmt_ScannedString {
							if __fmt_inClass {
								return __fmt___fmt_rune_private_8093ad48_scanRegex(__fmt_step.__fmt_state, false, __fmt_inClass)
							}
							return __fmt___fmt_rune_private_8093ad48_scannedString(__fmt___fmt_rune_private_8093ad48_scanRegexFlags(__fmt_step.__fmt_state), true)
						}()
					default:
						return __fmt___fmt_rune_private_8093ad48_scanRegex(__fmt_step.__fmt_state, false, __fmt_inClass)
					}
				}()
			}()
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_scanRegexFlags(__fmt_state __fmt_LexState) __fmt_LexState {
	return func() __fmt_LexState {
		if __fmt___fmt_rune_private_8093ad48_isRegexFlag(__fmt___fmt_rune_private_8093ad48_peek(__fmt_state)) {
			return __fmt___fmt_rune_private_8093ad48_scanRegexFlags(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
		}
		return __fmt_state
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexNumberToken(__fmt_state __fmt_LexState) __fmt_Lexed {
	return __fmt___fmt_rune_private_8093ad48_lexNumberAfterDigits(__fmt___fmt_rune_private_8093ad48_scanDigits(__fmt_state), false)
}

func __fmt___fmt_rune_private_8093ad48_scanDigits(__fmt_state __fmt_LexState) __fmt_LexState {
	return func() __fmt_LexState {
		if __fmt___fmt_rune_private_8093ad48_isDigit(__fmt___fmt_rune_private_8093ad48_peek(__fmt_state)) {
			return __fmt___fmt_rune_private_8093ad48_scanDigits(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
		}
		return __fmt_state
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexNumberAfterDigits(__fmt_state __fmt_LexState, __fmt_isDouble bool) __fmt_Lexed {
	return func() __fmt_Lexed {
		switch {
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == 'n':
			return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state), __fmt_TokenKind_BigInt)
		case __fmt___fmt_rune_private_8093ad48_peek(__fmt_state) == '.':
			return func() __fmt_Lexed {
				if __fmt___fmt_rune_private_8093ad48_isDigit(__fmt___fmt_rune_private_8093ad48_peekNext(__fmt_state)) {
					return __fmt___fmt_rune_private_8093ad48_lexNumberAfterDot(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
				}
				return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, func() __fmt_TokenKind {
					if __fmt_isDouble {
						return __fmt_TokenKind_Double
					}
					return __fmt_TokenKind_Int
				}())
			}()
		default:
			return func() __fmt_Lexed {
				if __fmt___fmt_rune_private_8093ad48_isExponentMarker(__fmt___fmt_rune_private_8093ad48_peek(__fmt_state)) {
					return __fmt___fmt_rune_private_8093ad48_lexNumberAfterExponent(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
				}
				return __fmt___fmt_rune_private_8093ad48_lexed(__fmt_state, func() __fmt_TokenKind {
					if __fmt_isDouble {
						return __fmt_TokenKind_Double
					}
					return __fmt_TokenKind_Int
				}())
			}()
		}
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexNumberAfterDot(__fmt_state __fmt_LexState) __fmt_Lexed {
	return __fmt___fmt_rune_private_8093ad48_lexNumberAfterDigits(__fmt___fmt_rune_private_8093ad48_scanDigits(__fmt_state), true)
}

func __fmt___fmt_rune_private_8093ad48_lexNumberAfterExponent(__fmt_state __fmt_LexState) __fmt_Lexed {
	return func() __fmt_Lexed {
		if __fmt___fmt_rune_private_8093ad48_isExponentSign(__fmt___fmt_rune_private_8093ad48_peek(__fmt_state)) {
			return __fmt___fmt_rune_private_8093ad48_lexNumberExponentDigits(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
		}
		return __fmt___fmt_rune_private_8093ad48_lexNumberExponentDigits(__fmt_state)
	}()
}

func __fmt___fmt_rune_private_8093ad48_lexNumberExponentDigits(__fmt_state __fmt_LexState) __fmt_Lexed {
	return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_scanDigits(__fmt_state), __fmt_TokenKind_Double)
}

func __fmt___fmt_rune_private_8093ad48_lexIdentifierToken(__fmt_state __fmt_LexState) __fmt_Lexed {
	return __fmt___fmt_rune_private_8093ad48_lexed(__fmt___fmt_rune_private_8093ad48_scanIdentifier(__fmt_state), __fmt_TokenKind_Ident)
}

func __fmt___fmt_rune_private_8093ad48_scanIdentifier(__fmt_state __fmt_LexState) __fmt_LexState {
	return func() __fmt_LexState {
		if __fmt___fmt_rune_private_8093ad48_isIdentContinue(__fmt___fmt_rune_private_8093ad48_peek(__fmt_state)) {
			return __fmt___fmt_rune_private_8093ad48_scanIdentifier(__fmt___fmt_rune_private_8093ad48_advanceState(__fmt_state))
		}
		return __fmt_state
	}()
}

func __fmt___fmt_rune_private_8093ad48_isDigit(__fmt_ch rune) bool {
	switch {
	case (__fmt_ch >= '0' && __fmt_ch <= '9'):
		return true
	default:
		return false
	}
}

func __fmt___fmt_rune_private_8093ad48_isAlpha(__fmt_ch rune) bool {
	switch {
	case (__fmt_ch >= 'a' && __fmt_ch <= 'z'):
		return true
	case (__fmt_ch >= 'A' && __fmt_ch <= 'Z'):
		return true
	default:
		return false
	}
}

func __fmt___fmt_rune_private_8093ad48_isIdentStart(__fmt_ch rune) bool {
	switch {
	case __fmt_ch == '_':
		return true
	default:
		return __fmt___fmt_rune_private_8093ad48_isIdentText(__fmt_ch) && __fmt___fmt_rune_private_8093ad48_isDigit(__fmt_ch) == false
	}
}

func __fmt___fmt_rune_private_8093ad48_isIdentContinue(__fmt_ch rune) bool {
	return __fmt___fmt_rune_private_8093ad48_isIdentStart(__fmt_ch) || __fmt___fmt_rune_private_8093ad48_isDigit(__fmt_ch)
}

func __fmt___fmt_rune_private_8093ad48_isIdentText(__fmt_ch rune) bool {
	return __fmt___fmt_rune_private_8093ad48_isIdentBoundary(__fmt_ch) == false && __fmt___fmt_rune_private_8093ad48_isAsciiPunctuation(__fmt_ch) == false
}

func __fmt___fmt_rune_private_8093ad48_isIdentBoundary(__fmt_ch rune) bool {
	switch {
	case __fmt_ch == '\x00':
		return true
	case __fmt_ch == ' ':
		return true
	case __fmt_ch == '\t':
		return true
	case __fmt_ch == '\r':
		return true
	case __fmt_ch == '\n':
		return true
	default:
		return false
	}
}

func __fmt___fmt_rune_private_8093ad48_isAsciiPunctuation(__fmt_ch rune) bool {
	switch {
	case (__fmt_ch >= '!' && __fmt_ch <= '/'):
		return true
	case (__fmt_ch >= ':' && __fmt_ch <= '@'):
		return true
	case (__fmt_ch >= '[' && __fmt_ch <= '^'):
		return true
	case __fmt_ch == '`':
		return true
	case (__fmt_ch >= '{' && __fmt_ch <= '~'):
		return true
	default:
		return false
	}
}

func __fmt___fmt_rune_private_8093ad48_isRegexFlag(__fmt_ch rune) bool {
	return __fmt___fmt_rune_private_8093ad48_isAlpha(__fmt_ch)
}

func __fmt___fmt_rune_private_8093ad48_isExponentMarker(__fmt_ch rune) bool {
	switch {
	case __fmt_ch == 'e':
		return true
	case __fmt_ch == 'E':
		return true
	default:
		return false
	}
}

func __fmt___fmt_rune_private_8093ad48_isExponentSign(__fmt_ch rune) bool {
	switch {
	case __fmt_ch == '+':
		return true
	case __fmt_ch == '-':
		return true
	default:
		return false
	}
}

func __fmt_formatSource(__fmt_source string) string {
	__fmt_tokens := __fmt_lex(__fmt_source)
	__fmt_state := __fmt___fmt_rune_private_b2b0db24_formatTokens(__fmt_tokens, 0, __fmt___fmt_rune_private_b2b0db24_emptyFormatState())
	return __fmt___fmt_rune_private_b2b0db24_preserveComments(__fmt_source, __fmt___fmt_rune_private_b2b0db24_finishFormat(__fmt_state.__fmt_text))
}

func __fmt___fmt_rune_private_b2b0db24_emptyFormatState() __fmt_FormatState {
	return __fmt_FormatState{__fmt_text: "", __fmt_indent: 0, __fmt_lineStart: true, __fmt_previous: __fmt_TokenKind_EOF}
}

func __fmt___fmt_rune_private_b2b0db24_formatTokens(__fmt_tokens []__fmt_Token, __fmt_index int, __fmt_state __fmt_FormatState) __fmt_FormatState {
	__fmt_done := __fmt_index >= len(__fmt_tokens)
	return func() __fmt_FormatState {
		switch {
		case __fmt_done == true:
			return __fmt_state
		default:
			return __fmt___fmt_rune_private_b2b0db24_formatTokens(__fmt_tokens, __fmt_index+1, __fmt___fmt_rune_private_b2b0db24_formatToken(__fmt_state, __fmt_tokens[__fmt_index]))
		}
	}()
}

func __fmt___fmt_rune_private_b2b0db24_formatToken(__fmt_state __fmt_FormatState, __fmt_token __fmt_Token) __fmt_FormatState {
	return func() __fmt_FormatState {
		switch {
		case __fmt_token.__fmt_kind == __fmt_TokenKind_EOF:
			return __fmt_state
		case __fmt_token.__fmt_kind == __fmt_TokenKind_Newline:
			return __fmt___fmt_rune_private_b2b0db24_formatNewline(__fmt_state)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_LBrace:
			return __fmt___fmt_rune_private_b2b0db24_formatOpenBrace(__fmt_state)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_RBrace:
			return __fmt___fmt_rune_private_b2b0db24_formatCloseBrace(__fmt_state)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_Comma:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, ", ", __fmt_TokenKind_Comma)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_Colon:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, ": ", __fmt_TokenKind_Colon)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_At:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, "@", __fmt_TokenKind_At)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_Dot:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, ".", __fmt_TokenKind_Dot)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_DoubleColon:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, "::", __fmt_TokenKind_DoubleColon)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_Question:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, "?", __fmt_TokenKind_Question)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_QuestionQuestion:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, "??", __fmt_TokenKind_QuestionQuestion)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_Apostrophe:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, "'", __fmt_TokenKind_Apostrophe)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_LParen:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, "(", __fmt_TokenKind_LParen)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_RParen:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, ")", __fmt_TokenKind_RParen)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_LBracket:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, "[", __fmt_TokenKind_LBracket)
		case __fmt_token.__fmt_kind == __fmt_TokenKind_RBracket:
			return __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state, "]", __fmt_TokenKind_RBracket)
		default:
			return __fmt___fmt_rune_private_b2b0db24_formatWord(__fmt_state, __fmt_token)
		}
	}()
}

func __fmt___fmt_rune_private_b2b0db24_formatNewline(__fmt_state __fmt_FormatState) __fmt_FormatState {
	return func() __fmt_FormatState {
		if __fmt_state.__fmt_lineStart {
			return __fmt_state
		}
		return __fmt_FormatState{__fmt_text: __fmt___fmt_rune_private_b2b0db24_trimTrailingSpace(__fmt_state.__fmt_text) + "\n", __fmt_indent: __fmt_state.__fmt_indent, __fmt_lineStart: true, __fmt_previous: __fmt_state.__fmt_previous}
	}()
}

func __fmt___fmt_rune_private_b2b0db24_formatOpenBrace(__fmt_state __fmt_FormatState) __fmt_FormatState {
	__fmt_text := __fmt___fmt_rune_private_b2b0db24_ensureSpace(__fmt_state.__fmt_text, __fmt_state.__fmt_lineStart)
	return __fmt_FormatState{__fmt_text: __fmt_text + "{\n", __fmt_indent: __fmt_state.__fmt_indent + 1, __fmt_lineStart: true, __fmt_previous: __fmt_TokenKind_LBrace}
}

func __fmt___fmt_rune_private_b2b0db24_formatCloseBrace(__fmt_state __fmt_FormatState) __fmt_FormatState {
	__fmt_indent := func() int {
		if __fmt_state.__fmt_indent > 0 {
			return __fmt_state.__fmt_indent - 1
		}
		return 0
	}()
	__fmt_text := __fmt___fmt_rune_private_b2b0db24_trimTrailingSpace(__fmt_state.__fmt_text)
	__fmt_text = func() string {
		if strings.HasSuffix(__fmt_text, "\n") {
			return __fmt_text
		}
		return __fmt_text + "\n"
	}()
	return __fmt_FormatState{__fmt_text: __fmt_text + __fmt___fmt_rune_private_b2b0db24_indentText(__fmt_indent) + "}", __fmt_indent: __fmt_indent, __fmt_lineStart: false, __fmt_previous: __fmt_TokenKind_RBrace}
}

func __fmt___fmt_rune_private_b2b0db24_formatPunctuation(__fmt_state __fmt_FormatState, __fmt_text string, __fmt_kind __fmt_TokenKind) __fmt_FormatState {
	return __fmt_FormatState{__fmt_text: __fmt___fmt_rune_private_b2b0db24_prefixIndent(__fmt_state.__fmt_text, __fmt_state.__fmt_indent, __fmt_state.__fmt_lineStart) + __fmt_text, __fmt_indent: __fmt_state.__fmt_indent, __fmt_lineStart: false, __fmt_previous: __fmt_kind}
}

func __fmt___fmt_rune_private_b2b0db24_formatWord(__fmt_state __fmt_FormatState, __fmt_token __fmt_Token) __fmt_FormatState {
	__fmt_prefix := __fmt___fmt_rune_private_b2b0db24_prefixIndent(__fmt_state.__fmt_text, __fmt_state.__fmt_indent, __fmt_state.__fmt_lineStart)
	__fmt_separated := __fmt___fmt_rune_private_b2b0db24_needsSpace(__fmt_state.__fmt_previous, __fmt_token.__fmt_kind) && !__fmt_state.__fmt_lineStart
	return __fmt_FormatState{__fmt_text: func() string {
		if __fmt_separated {
			return __fmt___fmt_rune_private_b2b0db24_ensureSpace(__fmt_prefix, false)
		}
		return __fmt_prefix
	}() + __fmt_token.__fmt_lexeme, __fmt_indent: __fmt_state.__fmt_indent, __fmt_lineStart: false, __fmt_previous: __fmt_token.__fmt_kind}
}

func __fmt___fmt_rune_private_b2b0db24_needsSpace(__fmt_previous __fmt_TokenKind, __fmt_current __fmt_TokenKind) bool {
	return __fmt___fmt_rune_private_b2b0db24_isBinaryOperator(__fmt_previous) || __fmt___fmt_rune_private_b2b0db24_isBinaryOperator(__fmt_current) || __fmt_previous != __fmt_TokenKind_At && __fmt_previous != __fmt_TokenKind_Dot && __fmt_previous != __fmt_TokenKind_DoubleColon && (__fmt___fmt_rune_private_b2b0db24_endsWord(__fmt_previous) && __fmt___fmt_rune_private_b2b0db24_startsWord(__fmt_current))
}

func __fmt___fmt_rune_private_b2b0db24_endsWord(__fmt_kind __fmt_TokenKind) bool {
	return __fmt___fmt_rune_private_b2b0db24_isTokenIn(__fmt_kind, []__fmt_TokenKind{__fmt_TokenKind_Ident, __fmt_TokenKind_Int, __fmt_TokenKind_Double, __fmt_TokenKind_BigInt, __fmt_TokenKind_String, __fmt_TokenKind_Char, __fmt_TokenKind_Regex, __fmt_TokenKind_TemplateString, __fmt_TokenKind_RParen, __fmt_TokenKind_RBracket})
}

func __fmt___fmt_rune_private_b2b0db24_startsWord(__fmt_kind __fmt_TokenKind) bool {
	return __fmt___fmt_rune_private_b2b0db24_isTokenIn(__fmt_kind, []__fmt_TokenKind{__fmt_TokenKind_Ident, __fmt_TokenKind_Int, __fmt_TokenKind_Double, __fmt_TokenKind_BigInt, __fmt_TokenKind_String, __fmt_TokenKind_Char, __fmt_TokenKind_Regex, __fmt_TokenKind_TemplateString, __fmt_TokenKind_At, __fmt_TokenKind_Dollar})
}

func __fmt___fmt_rune_private_b2b0db24_isBinaryOperator(__fmt_kind __fmt_TokenKind) bool {
	return __fmt___fmt_rune_private_b2b0db24_isTokenIn(__fmt_kind, []__fmt_TokenKind{__fmt_TokenKind_FatArrow, __fmt_TokenKind_Arrow, __fmt_TokenKind_Assign, __fmt_TokenKind_Declare, __fmt_TokenKind_MutDeclare, __fmt_TokenKind_Plus, __fmt_TokenKind_Minus, __fmt_TokenKind_Star, __fmt_TokenKind_Slash, __fmt_TokenKind_Percent, __fmt_TokenKind_EqualEqual, __fmt_TokenKind_BangEqual, __fmt_TokenKind_Less, __fmt_TokenKind_LessEqual, __fmt_TokenKind_Greater, __fmt_TokenKind_GreaterEqual, __fmt_TokenKind_AndAnd, __fmt_TokenKind_OrOr, __fmt_TokenKind_QuestionQuestion, __fmt_TokenKind_BitAnd, __fmt_TokenKind_BitOr, __fmt_TokenKind_BitXor, __fmt_TokenKind_ShiftLeft, __fmt_TokenKind_ShiftRight, __fmt_TokenKind_UnsignedShiftRight})
}

func __fmt___fmt_rune_private_b2b0db24_isTokenIn(__fmt_kind __fmt_TokenKind, __fmt_kinds []__fmt_TokenKind) bool {
	return __fmt___fmt_rune_private_b2b0db24_isTokenInAt(__fmt_kind, __fmt_kinds, 0)
}

func __fmt___fmt_rune_private_b2b0db24_isTokenInAt(__fmt_kind __fmt_TokenKind, __fmt_kinds []__fmt_TokenKind, __fmt_index int) bool {
	return func() bool {
		if __fmt_index >= len(__fmt_kinds) {
			return false
		}
		return func() bool {
			if __fmt_kind == __fmt_kinds[__fmt_index] {
				return true
			}
			return __fmt___fmt_rune_private_b2b0db24_isTokenInAt(__fmt_kind, __fmt_kinds, __fmt_index+1)
		}()
	}()
}

func __fmt___fmt_rune_private_b2b0db24_prefixIndent(__fmt_text string, __fmt_indent int, __fmt_lineStart bool) string {
	return func() string {
		if __fmt_lineStart {
			return __fmt_text + __fmt___fmt_rune_private_b2b0db24_indentText(__fmt_indent)
		}
		return __fmt_text
	}()
}

func __fmt___fmt_rune_private_b2b0db24_ensureSpace(__fmt_text string, __fmt_lineStart bool) string {
	return func() string {
		if __fmt_lineStart || len(__fmt_text) == 0 || strings.HasSuffix(__fmt_text, " ") || strings.HasSuffix(__fmt_text, "\n") {
			return __fmt_text
		}
		return __fmt_text + " "
	}()
}

func __fmt___fmt_rune_private_b2b0db24_trimTrailingSpace(__fmt_text string) string {
	return func() string {
		if strings.HasSuffix(__fmt_text, " ") {
			return __fmt___fmt_rune_private_b2b0db24_trimTrailingSpace(func() string { runes := []rune(__fmt_text); return string(runes[0 : len([]rune(__fmt_text))-1]) }())
		}
		return __fmt_text
	}()
}

func __fmt___fmt_rune_private_b2b0db24_indentText(__fmt_indent int) string {
	return func() string {
		if __fmt_indent <= 0 {
			return ""
		}
		return "  " + __fmt___fmt_rune_private_b2b0db24_indentText(__fmt_indent-1)
	}()
}

func __fmt___fmt_rune_private_b2b0db24_preserveComments(__fmt_source string, __fmt_formatted string) string {
	return __fmt___fmt_rune_private_b2b0db24_preserveCommentLines(func() []string { parts := strings.Split(__fmt_formatted, "\n"); return parts }(), func() []string { parts := strings.Split(__fmt_source, "\n"); return parts }(), 0, "")
}

func __fmt___fmt_rune_private_b2b0db24_preserveCommentLines(__fmt_formatted []string, __fmt_source []string, __fmt_index int, __fmt_out string) string {
	return func() string {
		if __fmt_index >= len(__fmt_formatted) {
			return __fmt_out
		}
		return __fmt___fmt_rune_private_b2b0db24_preserveCommentLines(__fmt_formatted, __fmt_source, __fmt_index+1, __fmt_out+__fmt___fmt_rune_private_b2b0db24_preserveCommentLine(__fmt_formatted[__fmt_index], __fmt_source, 0)+func() string {
			if __fmt_index+1 < len(__fmt_formatted) {
				return "\n"
			}
			return ""
		}())
	}()
}

func __fmt___fmt_rune_private_b2b0db24_preserveCommentLine(__fmt_formatted string, __fmt_source []string, __fmt_index int) string {
	__fmt_comment := __fmt___fmt_rune_private_b2b0db24_findInlineComment(__fmt_source, __fmt___fmt_rune_private_b2b0db24_canonicalLine(__fmt_formatted), __fmt_index)
	return func() string {
		if __fmt_comment == "" {
			return __fmt_formatted
		}
		return __fmt_formatted + " " + __fmt_comment
	}()
}

func __fmt___fmt_rune_private_b2b0db24_findInlineComment(__fmt_source []string, __fmt_formatted string, __fmt_index int) string {
	return func() string {
		if __fmt_index >= len(__fmt_source) {
			return ""
		}
		return __fmt___fmt_rune_private_b2b0db24_inlineCommentForLine(__fmt_source[__fmt_index], __fmt_formatted, __fmt_source, __fmt_index)
	}()
}

func __fmt___fmt_rune_private_b2b0db24_inlineCommentForLine(__fmt_line string, __fmt_formatted string, __fmt_source []string, __fmt_index int) string {
	__fmt_commentStart := __fmt___fmt_rune_private_b2b0db24_firstCommentStart(__fmt_line)
	return func() string {
		if __fmt_commentStart < 0 {
			return __fmt___fmt_rune_private_b2b0db24_findInlineComment(__fmt_source, __fmt_formatted, __fmt_index+1)
		}
		return func() string {
			if __fmt___fmt_rune_private_b2b0db24_canonicalLine(func() string { runes := []rune(__fmt_line); return string(runes[0:__fmt_commentStart]) }()) == __fmt_formatted {
				return strings.TrimSpace((func() string { runes := []rune(__fmt_line); return string(runes[__fmt_commentStart:len([]rune(__fmt_line))]) }()))
			}
			return __fmt___fmt_rune_private_b2b0db24_findInlineComment(__fmt_source, __fmt_formatted, __fmt_index+1)
		}()
	}()
}

func __fmt___fmt_rune_private_b2b0db24_firstCommentStart(__fmt_line string) int {
	return strings.Index(__fmt_line, "//")
}

func __fmt___fmt_rune_private_b2b0db24_canonicalLine(__fmt_line string) string {
	return strings.ReplaceAll((strings.ReplaceAll((strings.ReplaceAll((strings.TrimSpace(__fmt_line)), " ", "")), "\t", "")), "\r", "")
}

func __fmt___fmt_rune_private_b2b0db24_finishFormat(__fmt_text string) string {
	__fmt_trimmed := __fmt___fmt_rune_private_b2b0db24_trimTrailingSpace(__fmt_text)
	return func() string {
		if strings.HasSuffix(__fmt_trimmed, "\n") {
			return __fmt_trimmed
		}
		return __fmt_trimmed + "\n"
	}()
}
