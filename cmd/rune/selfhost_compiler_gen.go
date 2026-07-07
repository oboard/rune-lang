package main

import (
	"fmt"
	"strings"
)

type __TokenKind int

const (
	__TokenKind_EOF                __TokenKind = 0
	__TokenKind_Illegal            __TokenKind = 1
	__TokenKind_Newline            __TokenKind = 2
	__TokenKind_Ident              __TokenKind = 3
	__TokenKind_Int                __TokenKind = 4
	__TokenKind_Double             __TokenKind = 5
	__TokenKind_BigInt             __TokenKind = 6
	__TokenKind_String             __TokenKind = 7
	__TokenKind_TemplateString     __TokenKind = 8
	__TokenKind_Char               __TokenKind = 9
	__TokenKind_Regex              __TokenKind = 10
	__TokenKind_XMLText            __TokenKind = 11
	__TokenKind_Hash               __TokenKind = 12
	__TokenKind_At                 __TokenKind = 13
	__TokenKind_Dollar             __TokenKind = 14
	__TokenKind_Dot                __TokenKind = 15
	__TokenKind_DotDot             __TokenKind = 16
	__TokenKind_DotDotLess         __TokenKind = 17
	__TokenKind_DotDotEqual        __TokenKind = 18
	__TokenKind_DotDotDot          __TokenKind = 19
	__TokenKind_Comma              __TokenKind = 20
	__TokenKind_Colon              __TokenKind = 21
	__TokenKind_DoubleColon        __TokenKind = 22
	__TokenKind_LParen             __TokenKind = 23
	__TokenKind_RParen             __TokenKind = 24
	__TokenKind_LBracket           __TokenKind = 25
	__TokenKind_RBracket           __TokenKind = 26
	__TokenKind_LBrace             __TokenKind = 27
	__TokenKind_RBrace             __TokenKind = 28
	__TokenKind_Question           __TokenKind = 29
	__TokenKind_QuestionQuestion   __TokenKind = 30
	__TokenKind_Apostrophe         __TokenKind = 31
	__TokenKind_FatArrow           __TokenKind = 32
	__TokenKind_Assign             __TokenKind = 33
	__TokenKind_Declare            __TokenKind = 34
	__TokenKind_MutDeclare         __TokenKind = 35
	__TokenKind_Arrow              __TokenKind = 36
	__TokenKind_Plus               __TokenKind = 37
	__TokenKind_PlusPlus           __TokenKind = 38
	__TokenKind_Minus              __TokenKind = 39
	__TokenKind_Star               __TokenKind = 40
	__TokenKind_Slash              __TokenKind = 41
	__TokenKind_Percent            __TokenKind = 42
	__TokenKind_Bang               __TokenKind = 43
	__TokenKind_Tilde              __TokenKind = 44
	__TokenKind_BitAnd             __TokenKind = 45
	__TokenKind_BitOr              __TokenKind = 46
	__TokenKind_BitXor             __TokenKind = 47
	__TokenKind_ShiftLeft          __TokenKind = 48
	__TokenKind_ShiftRight         __TokenKind = 49
	__TokenKind_UnsignedShiftRight __TokenKind = 50
	__TokenKind_AndAnd             __TokenKind = 51
	__TokenKind_OrOr               __TokenKind = 52
	__TokenKind_EqualEqual         __TokenKind = 53
	__TokenKind_BangEqual          __TokenKind = 54
	__TokenKind_Less               __TokenKind = 55
	__TokenKind_LessEqual          __TokenKind = 56
	__TokenKind_Greater            __TokenKind = 57
	__TokenKind_GreaterEqual       __TokenKind = 58
	__TokenKind_Underscore         __TokenKind = 59
)

type __TypeRefKind int

const (
	__TypeRefKind_Unknown  __TypeRefKind = 0
	__TypeRefKind_Name     __TypeRefKind = 1
	__TypeRefKind_Group    __TypeRefKind = 2
	__TypeRefKind_Tuple    __TypeRefKind = 3
	__TypeRefKind_Function __TypeRefKind = 4
)

type __ExprKind int

const (
	__ExprKind_Unknown           __ExprKind = 0
	__ExprKind_Identifier        __ExprKind = 1
	__ExprKind_At                __ExprKind = 2
	__ExprKind_This              __ExprKind = 3
	__ExprKind_Int               __ExprKind = 4
	__ExprKind_Double            __ExprKind = 5
	__ExprKind_BigInt            __ExprKind = 6
	__ExprKind_String            __ExprKind = 7
	__ExprKind_Template          __ExprKind = 8
	__ExprKind_Char              __ExprKind = 9
	__ExprKind_Regex             __ExprKind = 10
	__ExprKind_XMLText           __ExprKind = 11
	__ExprKind_Bool              __ExprKind = 12
	__ExprKind_Null              __ExprKind = 13
	__ExprKind_Unary             __ExprKind = 14
	__ExprKind_Postfix           __ExprKind = 15
	__ExprKind_Unwrap            __ExprKind = 16
	__ExprKind_CompileTime       __ExprKind = 17
	__ExprKind_Binary            __ExprKind = 18
	__ExprKind_Ternary           __ExprKind = 19
	__ExprKind_Assign            __ExprKind = 20
	__ExprKind_Call              __ExprKind = 21
	__ExprKind_Args              __ExprKind = 22
	__ExprKind_Lambda            __ExprKind = 23
	__ExprKind_Selector          __ExprKind = 24
	__ExprKind_Index             __ExprKind = 25
	__ExprKind_Array             __ExprKind = 26
	__ExprKind_Tuple             __ExprKind = 27
	__ExprKind_Map               __ExprKind = 28
	__ExprKind_Entry             __ExprKind = 29
	__ExprKind_Spread            __ExprKind = 30
	__ExprKind_Reactive          __ExprKind = 31
	__ExprKind_Struct            __ExprKind = 32
	__ExprKind_Object            __ExprKind = 33
	__ExprKind_Field             __ExprKind = 34
	__ExprKind_PrivateField      __ExprKind = 35
	__ExprKind_Method            __ExprKind = 36
	__ExprKind_PrivateMethod     __ExprKind = 37
	__ExprKind_Block             __ExprKind = 38
	__ExprKind_PatternBlock      __ExprKind = 39
	__ExprKind_Match             __ExprKind = 40
	__ExprKind_Branch            __ExprKind = 41
	__ExprKind_Pattern           __ExprKind = 42
	__ExprKind_Let               __ExprKind = 43
	__ExprKind_ObjectDestructure __ExprKind = 44
	__ExprKind_Error             __ExprKind = 45
	__ExprKind_Watch             __ExprKind = 46
)

type __Token struct {
	__kind   __TokenKind
	__lexeme string
	__offset int
	__line   int
	__column int
}

type __LexState struct {
	__source        string
	__start         int
	__current       int
	__line          int
	__column        int
	__startLine     int
	__startColumn   int
	__canStartRegex bool
}

type __Advanced struct {
	__state __LexState
	__ch    rune
}

type __Lexed struct {
	__state __LexState
	__kind  __TokenKind
}

type __ScannedString struct {
	__state __LexState
	__ok    bool
}

type __ParsedTypeParam struct {
	__name     string
	__optional bool
	__typeRef  __ParsedTypeRef
}

type __ParsedTypeRef struct {
	__kind        __TypeRefKind
	__name        string
	__module      string
	__nullable    bool
	__args        []__ParsedTypeRef
	__params      []__ParsedTypeParam
	__returnTypes []__ParsedTypeRef
	__line        int
	__column      int
}

type __ParseError struct {
	__message string
	__line    int
	__column  int
}

type __ParsedImport struct {
	__path   string
	__go     bool
	__line   int
	__column int
}

type __ParsedAnnotation struct {
	__marker string
	__module string
	__name   string
	__args   []__ParsedExpr
	__line   int
	__column int
}

type __ParsedParam struct {
	__name    string
	__typeRef __ParsedTypeRef
	__line    int
	__column  int
}

type __ParsedExpr struct {
	__kind     __ExprKind
	__text     string
	__name     string
	__value    string
	__op       string
	__params   []__ParsedParam
	__children []__ParsedExpr
	__line     int
	__column   int
}

type __ParsedField struct {
	__name        string
	__private     bool
	__annotations []__ParsedAnnotation
	__typeRef     __ParsedTypeRef
	__line        int
	__column      int
}

type __ParsedEnumMember struct {
	__name        string
	__private     bool
	__annotations []__ParsedAnnotation
	__value       string
	__params      []__ParsedParam
	__line        int
	__column      int
}

type __ParsedFunction struct {
	__name         string
	__private      bool
	__static       bool
	__routine      bool
	__macro        bool
	__annotations  []__ParsedAnnotation
	__receiverType string
	__generics     []string
	__params       []__ParsedParam
	__returnType   __ParsedTypeRef
	__body         __ParsedExpr
	__line         int
	__column       int
}

type __ParsedType struct {
	__name        string
	__private     bool
	__enum        bool
	__annotations []__ParsedAnnotation
	__generics    []string
	__fields      []__ParsedField
	__methods     []__ParsedFunction
	__members     []__ParsedEnumMember
	__line        int
	__column      int
}

type __ParsedTest struct {
	__name   string
	__body   __ParsedExpr
	__line   int
	__column int
}

type __ParsedFile struct {
	__imports   []__ParsedImport
	__types     []__ParsedType
	__functions []__ParsedFunction
	__tests     []__ParsedTest
	__errors    []__ParseError
}

type __ParserState struct {
	__tokens  []__Token
	__current int
	__errors  []__ParseError
}

type __TokenStep struct {
	__state __ParserState
	__token __Token
}

type __BoolStep struct {
	__state __ParserState
	__ok    bool
}

type __StringStep struct {
	__state __ParserState
	__value string
}

type __TypeRefStep struct {
	__state   __ParserState
	__typeRef __ParsedTypeRef
}

type __TypeRefListStep struct {
	__state __ParserState
	__refs  []__ParsedTypeRef
}

type __TypeParamStep struct {
	__state __ParserState
	__param __ParsedTypeParam
}

type __TypeParamListStep struct {
	__state  __ParserState
	__params []__ParsedTypeParam
}

type __StringListStep struct {
	__state  __ParserState
	__values []string
}

type __AnnotationListStep struct {
	__state       __ParserState
	__annotations []__ParsedAnnotation
}

type __ParamListStep struct {
	__state  __ParserState
	__params []__ParsedParam
}

type __ExprStep struct {
	__state __ParserState
	__expr  __ParsedExpr
}

type __TemplateParse struct {
	__text     string
	__children []__ParsedExpr
}

type __FileStep struct {
	__state __ParserState
	__file  __ParsedFile
}

type __ImportStep struct {
	__state      __ParserState
	__importDecl __ParsedImport
}

type __FunctionStep struct {
	__state    __ParserState
	__function __ParsedFunction
}

type __TypeStep struct {
	__state    __ParserState
	__typeDecl __ParsedType
}

type __TestStep struct {
	__state    __ParserState
	__testDecl __ParsedTest
}

type __FieldStep struct {
	__state __ParserState
	__field __ParsedField
}

type __EnumMemberStep struct {
	__state  __ParserState
	__member __ParsedEnumMember
}

type __EnumMemberPayloadStep struct {
	__state  __ParserState
	__value  string
	__params []__ParsedParam
}

type __AnnotationStep struct {
	__state      __ParserState
	__annotation __ParsedAnnotation
}

type __IRImport struct {
	__path   string
	__go     bool
	__line   int
	__column int
}

type __IRTSImport struct {
	__path      string
	__specifier string
	__functions []__IRFunction
	__values    []__IRConst
	__line      int
	__column    int
}

type __IRParam struct {
	__name     string
	__typeName string
	__line     int
	__column   int
}

type __IRExpr struct {
	__kind     __ExprKind
	__text     string
	__name     string
	__value    string
	__op       string
	__params   []__IRParam
	__children []__IRExpr
	__line     int
	__column   int
}

type __IRField struct {
	__name     string
	__private  bool
	__typeName string
	__line     int
	__column   int
}

type __IREnumMember struct {
	__name    string
	__private bool
	__value   string
	__params  []__IRParam
	__line    int
	__column  int
}

type __IRFunction struct {
	__name         string
	__private      bool
	__routine      bool
	__macro        bool
	__receiverType string
	__generics     []string
	__params       []__IRParam
	__returnType   string
	__body         __IRExpr
	__line         int
	__column       int
}

type __IRConst struct {
	__name     string
	__private  bool
	__typeName string
	__value    __IRExpr
	__line     int
	__column   int
}

type __IRStructType struct {
	__name     string
	__private  bool
	__generics []string
	__fields   []__IRField
	__methods  []__IRFunction
	__line     int
	__column   int
}

type __IREnumType struct {
	__name     string
	__private  bool
	__generics []string
	__members  []__IREnumMember
	__methods  []__IRFunction
	__line     int
	__column   int
}

type __IRTest struct {
	__name   string
	__body   __IRExpr
	__line   int
	__column int
}

type __IRFile struct {
	__imports   []__IRImport
	__tsImports []__IRTSImport
	__structs   []__IRStructType
	__enums     []__IREnumType
	__constants []__IRConst
	__functions []__IRFunction
	__tests     []__IRTest
	__errors    []__ParseError
}

type __CompileResult struct {
	__ok     bool
	__output string
	__errors []string
}

type __SourceFile struct {
	__path   string
	__source string
}

func runeTemplateString(value any) string {
	switch v := value.(type) {
	case nil:
		return "null"
	case rune:
		return string(v)
	default:
		return fmt.Sprint(v)
	}
}

type runeError struct {
	__code    int
	__message string
	__cause   *runeError
}

func runeErrorFrom(err error) *runeError {
	if err == nil {
		return nil
	}
	return &runeError{__code: 1, __message: err.Error()}
}

func __lex(__source string) []__Token {
	return ____rune_private_2013c1e3_scan(__LexState{__source: __source, __start: 0, __current: 0, __line: 1, __column: 1, __startLine: 1, __startColumn: 1, __canStartRegex: true}, ____rune_private_2013c1e3_emptyTokens())
}

func ____rune_private_2013c1e3_emptyTokens() []__Token {
	return append([]__Token{}, []__Token{__Token{__kind: __TokenKind_EOF, __lexeme: "", __offset: 0, __line: 0, __column: 0}}[0:0]...)
}

func ____rune_private_2013c1e3_scan(__state __LexState, __tokens []__Token) []__Token {
	__skipped := ____rune_private_2013c1e3_skipIgnored(__state)
	__started := ____rune_private_2013c1e3_markStart(__skipped)
	return func() []__Token {
		if ____rune_private_2013c1e3_atEnd(__started) {
			return ____rune_private_2013c1e3_appendToken(__tokens, ____rune_private_2013c1e3_makeToken(__started, __TokenKind_EOF))
		}
		return ____rune_private_2013c1e3_scanLexed(____rune_private_2013c1e3_scanToken(____rune_private_2013c1e3_advance(__started)), __tokens)
	}()
}

func ____rune_private_2013c1e3_scanLexed(__lexed __Lexed, __tokens []__Token) []__Token {
	__nextTokens := ____rune_private_2013c1e3_appendToken(__tokens, ____rune_private_2013c1e3_makeToken(__lexed.__state, __lexed.__kind))
	return ____rune_private_2013c1e3_scan(____rune_private_2013c1e3_finishToken(__lexed.__state, __lexed.__kind), __nextTokens)
}

func ____rune_private_2013c1e3_appendToken(__tokens []__Token, __token __Token) []__Token {
	__tokens = append(__tokens, __token)
	return __tokens
}

func ____rune_private_2013c1e3_makeToken(__state __LexState, __kind __TokenKind) __Token {
	return __Token{__kind: __kind, __lexeme: func() string {
		runes := []rune(__state.__source)
		return string(runes[__state.__start:__state.__current])
	}(), __offset: __state.__start, __line: __state.__startLine, __column: __state.__startColumn}
}

func ____rune_private_2013c1e3_finishToken(__state __LexState, __kind __TokenKind) __LexState {
	return __LexState{__source: __state.__source, __start: __state.__start, __current: __state.__current, __line: __state.__line, __column: __state.__column, __startLine: __state.__startLine, __startColumn: __state.__startColumn, __canStartRegex: !____rune_private_2013c1e3_canEndExpression(__state, __kind)}
}

func ____rune_private_2013c1e3_canEndExpression(__state __LexState, __kind __TokenKind) bool {
	return ____rune_private_2013c1e3_canEndValueToken(__kind) || ____rune_private_2013c1e3_canEndXmlLess(__state, __kind)
}

func ____rune_private_2013c1e3_canEndValueToken(__kind __TokenKind) bool {
	return func() bool {
		switch {
		case (__kind == __TokenKind_Ident) || (__kind == __TokenKind_Int) || (__kind == __TokenKind_Double) || (__kind == __TokenKind_BigInt) || (__kind == __TokenKind_String) || (__kind == __TokenKind_TemplateString) || (__kind == __TokenKind_Char) || (__kind == __TokenKind_Regex) || (__kind == __TokenKind_XMLText) || (__kind == __TokenKind_RParen) || (__kind == __TokenKind_RBracket) || (__kind == __TokenKind_RBrace):
			return true
		default:
			return false
		}
	}()
}

func ____rune_private_2013c1e3_canEndXmlLess(__state __LexState, __kind __TokenKind) bool {
	return __kind == __TokenKind_Less && (____rune_private_2013c1e3_peek(__state) == '/' || ____rune_private_2013c1e3_isIdentStart(____rune_private_2013c1e3_peek(__state)))
}

func __tokenKindName(__kind __TokenKind) string {
	return func() string {
		switch {
		case __kind == __TokenKind_EOF:
			return "EOF"
		case __kind == __TokenKind_Illegal:
			return "Illegal"
		case __kind == __TokenKind_Newline:
			return "Newline"
		case __kind == __TokenKind_Ident:
			return "Ident"
		case __kind == __TokenKind_Int:
			return "Int"
		case __kind == __TokenKind_Double:
			return "Double"
		case __kind == __TokenKind_BigInt:
			return "BigInt"
		case __kind == __TokenKind_String:
			return "String"
		case __kind == __TokenKind_TemplateString:
			return "TemplateString"
		case __kind == __TokenKind_Char:
			return "Char"
		case __kind == __TokenKind_Regex:
			return "Regex"
		case __kind == __TokenKind_XMLText:
			return "XMLText"
		case __kind == __TokenKind_Hash:
			return "Hash"
		case __kind == __TokenKind_At:
			return "At"
		case __kind == __TokenKind_Dollar:
			return "Dollar"
		case __kind == __TokenKind_Dot:
			return "Dot"
		case __kind == __TokenKind_DotDot:
			return "DotDot"
		case __kind == __TokenKind_DotDotLess:
			return "DotDotLess"
		case __kind == __TokenKind_DotDotEqual:
			return "DotDotEqual"
		case __kind == __TokenKind_DotDotDot:
			return "DotDotDot"
		case __kind == __TokenKind_Comma:
			return "Comma"
		case __kind == __TokenKind_Colon:
			return "Colon"
		case __kind == __TokenKind_DoubleColon:
			return "DoubleColon"
		case __kind == __TokenKind_LParen:
			return "LParen"
		case __kind == __TokenKind_RParen:
			return "RParen"
		case __kind == __TokenKind_LBracket:
			return "LBracket"
		case __kind == __TokenKind_RBracket:
			return "RBracket"
		case __kind == __TokenKind_LBrace:
			return "LBrace"
		case __kind == __TokenKind_RBrace:
			return "RBrace"
		case __kind == __TokenKind_Question:
			return "Question"
		case __kind == __TokenKind_QuestionQuestion:
			return "QuestionQuestion"
		case __kind == __TokenKind_Apostrophe:
			return "Apostrophe"
		case __kind == __TokenKind_FatArrow:
			return "FatArrow"
		case __kind == __TokenKind_Assign:
			return "Assign"
		case __kind == __TokenKind_Declare:
			return "Declare"
		case __kind == __TokenKind_MutDeclare:
			return "MutDeclare"
		case __kind == __TokenKind_Arrow:
			return "Arrow"
		case __kind == __TokenKind_Plus:
			return "Plus"
		case __kind == __TokenKind_PlusPlus:
			return "PlusPlus"
		case __kind == __TokenKind_Minus:
			return "Minus"
		case __kind == __TokenKind_Star:
			return "Star"
		case __kind == __TokenKind_Slash:
			return "Slash"
		case __kind == __TokenKind_Percent:
			return "Percent"
		case __kind == __TokenKind_Bang:
			return "Bang"
		case __kind == __TokenKind_Tilde:
			return "Tilde"
		case __kind == __TokenKind_BitAnd:
			return "BitAnd"
		case __kind == __TokenKind_BitOr:
			return "BitOr"
		case __kind == __TokenKind_BitXor:
			return "BitXor"
		case __kind == __TokenKind_ShiftLeft:
			return "ShiftLeft"
		case __kind == __TokenKind_ShiftRight:
			return "ShiftRight"
		case __kind == __TokenKind_UnsignedShiftRight:
			return "UnsignedShiftRight"
		case __kind == __TokenKind_AndAnd:
			return "AndAnd"
		case __kind == __TokenKind_OrOr:
			return "OrOr"
		case __kind == __TokenKind_EqualEqual:
			return "EqualEqual"
		case __kind == __TokenKind_BangEqual:
			return "BangEqual"
		case __kind == __TokenKind_Less:
			return "Less"
		case __kind == __TokenKind_LessEqual:
			return "LessEqual"
		case __kind == __TokenKind_Greater:
			return "Greater"
		case __kind == __TokenKind_GreaterEqual:
			return "GreaterEqual"
		case __kind == __TokenKind_Underscore:
			return "Underscore"
		default:
			return "Unknown"
		}
	}()
}

func ____rune_private_2013c1e3_markStart(__state __LexState) __LexState {
	return __LexState{__source: __state.__source, __start: __state.__current, __current: __state.__current, __line: __state.__line, __column: __state.__column, __startLine: __state.__line, __startColumn: __state.__column, __canStartRegex: __state.__canStartRegex}
}

func ____rune_private_2013c1e3_atEnd(__state __LexState) bool {
	return __state.__current >= len([]rune(__state.__source))
}

func ____rune_private_2013c1e3_charAt(__source string, __index int) rune {
	return func() rune {
		if __index < 0 || __index >= len([]rune(__source)) {
			return ' '
		}
		return []rune(__source)[__index]
	}()
}

func ____rune_private_2013c1e3_peek(__state __LexState) rune {
	return ____rune_private_2013c1e3_charAt(__state.__source, __state.__current)
}

func ____rune_private_2013c1e3_peekNext(__state __LexState) rune {
	return ____rune_private_2013c1e3_charAt(__state.__source, __state.__current+1)
}

func ____rune_private_2013c1e3_advanceState(__state __LexState) __LexState {
	return ____rune_private_2013c1e3_advance(__state).__state
}

func ____rune_private_2013c1e3_advance(__state __LexState) __Advanced {
	return func() __Advanced {
		if ____rune_private_2013c1e3_atEnd(__state) {
			return ____rune_private_2013c1e3_advanced(__state, ' ')
		}
		return ____rune_private_2013c1e3_advanceChar(__state, []rune(__state.__source)[__state.__current])
	}()
}

func ____rune_private_2013c1e3_advanced(__state __LexState, __ch rune) __Advanced {
	return __Advanced{__state: __state, __ch: __ch}
}

func ____rune_private_2013c1e3_advanceChar(__state __LexState, __ch rune) __Advanced {
	return __Advanced{__state: __LexState{__source: __state.__source, __start: __state.__start, __current: __state.__current + 1, __line: func() int {
		if __ch == '\n' {
			return __state.__line + 1
		}
		return __state.__line
	}(), __column: func() int {
		if __ch == '\n' {
			return 1
		}
		return __state.__column + 1
	}(), __startLine: __state.__startLine, __startColumn: __state.__startColumn, __canStartRegex: __state.__canStartRegex}, __ch: __ch}
}

func ____rune_private_2013c1e3_skipIgnored(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_2013c1e3_atEnd(__state) {
			return __state
		}
		return func() __LexState {
			if ____rune_private_2013c1e3_isSpace(____rune_private_2013c1e3_peek(__state)) {
				return ____rune_private_2013c1e3_skipIgnored(____rune_private_2013c1e3_advanceState(__state))
			}
			return func() __LexState {
				if ____rune_private_2013c1e3_startsWith(__state, '/', '/') {
					return ____rune_private_2013c1e3_skipIgnored(____rune_private_2013c1e3_skipLineComment(____rune_private_2013c1e3_advanceState(____rune_private_2013c1e3_advanceState(__state))))
				}
				return func() __LexState {
					if ____rune_private_2013c1e3_startsWith(__state, '/', '*') {
						return ____rune_private_2013c1e3_skipIgnored(____rune_private_2013c1e3_skipBlockComment(____rune_private_2013c1e3_advanceState(____rune_private_2013c1e3_advanceState(__state))))
					}
					return __state
				}()
			}()
		}()
	}()
}

func ____rune_private_2013c1e3_startsWith(__state __LexState, __first rune, __second rune) bool {
	return ____rune_private_2013c1e3_peek(__state) == __first && ____rune_private_2013c1e3_peekNext(__state) == __second
}

func ____rune_private_2013c1e3_skipLineComment(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_2013c1e3_atEnd(__state) || ____rune_private_2013c1e3_peek(__state) == '\n' {
			return __state
		}
		return ____rune_private_2013c1e3_skipLineComment(____rune_private_2013c1e3_advanceState(__state))
	}()
}

func ____rune_private_2013c1e3_skipBlockComment(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_2013c1e3_atEnd(__state) {
			return __state
		}
		return func() __LexState {
			if ____rune_private_2013c1e3_startsWith(__state, '*', '/') {
				return ____rune_private_2013c1e3_advanceState(____rune_private_2013c1e3_advanceState(__state))
			}
			return ____rune_private_2013c1e3_skipBlockComment(____rune_private_2013c1e3_advanceState(__state))
		}()
	}()
}

func ____rune_private_2013c1e3_isSpace(__ch rune) bool {
	switch {
	case __ch == ' ':
		return true
	case __ch == '\t':
		return true
	case __ch == '\r':
		return true
	default:
		return false
	}
}

func ____rune_private_2013c1e3_scanToken(__step __Advanced) __Lexed {
	__destructure1 := __step
	__state := __destructure1.__state
	__ch := __destructure1.__ch
	return func() __Lexed {
		switch {
		case __ch == '\n':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Newline)
		case __ch == '#':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Hash)
		case __ch == '@':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_At)
		case __ch == '$':
			return ____rune_private_2013c1e3_lexDollar(__state)
		case __ch == '.':
			return ____rune_private_2013c1e3_lexDot(__state)
		case __ch == ',':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Comma)
		case __ch == ':':
			return ____rune_private_2013c1e3_lexColon(__state)
		case __ch == '(':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_LParen)
		case __ch == ')':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_RParen)
		case __ch == '[':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_LBracket)
		case __ch == ']':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_RBracket)
		case __ch == '{':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_LBrace)
		case __ch == '}':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_RBrace)
		case __ch == '?':
			return ____rune_private_2013c1e3_lexQuestion(__state)
		case __ch == '+':
			return ____rune_private_2013c1e3_lexPlus(__state)
		case __ch == '-':
			return ____rune_private_2013c1e3_lexMinus(__state)
		case __ch == '*':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Star)
		case __ch == '/':
			return func() __Lexed {
				if __state.__canStartRegex {
					return ____rune_private_2013c1e3_lexRegexToken(__state)
				}
				return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Slash)
			}()
		case __ch == '%':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Percent)
		case __ch == '!':
			return ____rune_private_2013c1e3_lexBang(__state)
		case __ch == '~':
			return ____rune_private_2013c1e3_lexTilde(__state)
		case __ch == '&':
			return ____rune_private_2013c1e3_lexAmp(__state)
		case __ch == '|':
			return ____rune_private_2013c1e3_lexPipe(__state)
		case __ch == '^':
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_BitXor)
		case __ch == '=':
			return ____rune_private_2013c1e3_lexEqual(__state)
		case __ch == '<':
			return ____rune_private_2013c1e3_lexLess(__state)
		case __ch == '>':
			return ____rune_private_2013c1e3_lexGreater(__state)
		case __ch == '"':
			return ____rune_private_2013c1e3_lexStringToken(__state)
		case __ch == '`':
			return ____rune_private_2013c1e3_lexTemplateStringToken(__state)
		case __ch == '\'':
			return func() __Lexed {
				if __state.__canStartRegex {
					return ____rune_private_2013c1e3_lexCharToken(__state)
				}
				return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Apostrophe)
			}()
		case __ch == '_':
			return func() __Lexed {
				if ____rune_private_2013c1e3_isIdentContinue(____rune_private_2013c1e3_peek(__state)) {
					return ____rune_private_2013c1e3_lexIdentifierToken(__state)
				}
				return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Underscore)
			}()
		default:
			return func() __Lexed {
				if ____rune_private_2013c1e3_isDigit(__ch) {
					return ____rune_private_2013c1e3_lexNumberToken(__state)
				}
				return func() __Lexed {
					if ____rune_private_2013c1e3_isIdentStart(__ch) {
						return ____rune_private_2013c1e3_lexIdentifierToken(__state)
					}
					return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Illegal)
				}()
			}()
		}
	}()
}

func ____rune_private_2013c1e3_lexed(__state __LexState, __kind __TokenKind) __Lexed {
	return __Lexed{__state: __state, __kind: __kind}
}

func ____rune_private_2013c1e3_lexDot(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '.':
			return ____rune_private_2013c1e3_lexDotDot(____rune_private_2013c1e3_advanceState(__state))
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Dot)
		}
	}()
}

func ____rune_private_2013c1e3_lexDotDot(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '.':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_DotDotDot)
		case ____rune_private_2013c1e3_peek(__state) == '<':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_DotDotLess)
		case ____rune_private_2013c1e3_peek(__state) == '=':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_DotDotEqual)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_DotDot)
		}
	}()
}

func ____rune_private_2013c1e3_lexColon(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '=':
			return func() __Lexed {
				if ____rune_private_2013c1e3_peek(____rune_private_2013c1e3_advanceState(__state)) == ':' {
					return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(____rune_private_2013c1e3_advanceState(__state)), __TokenKind_MutDeclare)
				}
				return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_Declare)
			}()
		case ____rune_private_2013c1e3_peek(__state) == ':':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_DoubleColon)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Colon)
		}
	}()
}

func ____rune_private_2013c1e3_lexQuestion(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '?':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_QuestionQuestion)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Question)
		}
	}()
}

func ____rune_private_2013c1e3_lexPlus(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '+':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_PlusPlus)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Plus)
		}
	}()
}

func ____rune_private_2013c1e3_lexMinus(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '>':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_Arrow)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Minus)
		}
	}()
}

func ____rune_private_2013c1e3_lexBang(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '=':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_BangEqual)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Bang)
		}
	}()
}

func ____rune_private_2013c1e3_lexTilde(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Tilde)
		}
	}()
}

func ____rune_private_2013c1e3_lexDollar(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Dollar)
		}
	}()
}

func ____rune_private_2013c1e3_lexAmp(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '&':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_AndAnd)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_BitAnd)
		}
	}()
}

func ____rune_private_2013c1e3_lexPipe(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '|':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_OrOr)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_BitOr)
		}
	}()
}

func ____rune_private_2013c1e3_lexEqual(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '>':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_FatArrow)
		case ____rune_private_2013c1e3_peek(__state) == '=':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_EqualEqual)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Assign)
		}
	}()
}

func ____rune_private_2013c1e3_lexLess(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '=':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_LessEqual)
		case ____rune_private_2013c1e3_peek(__state) == '<':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_ShiftLeft)
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Less)
		}
	}()
}

func ____rune_private_2013c1e3_lexGreater(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == '=':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_GreaterEqual)
		case ____rune_private_2013c1e3_peek(__state) == '>':
			return func() __Lexed {
				if ____rune_private_2013c1e3_startsWith(__state, '>', '>') {
					return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(____rune_private_2013c1e3_advanceState(__state)), __TokenKind_UnsignedShiftRight)
				}
				return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_ShiftRight)
			}()
		default:
			return ____rune_private_2013c1e3_lexed(__state, __TokenKind_Greater)
		}
	}()
}

func ____rune_private_2013c1e3_lexStringToken(__state __LexState) __Lexed {
	__scanned := ____rune_private_2013c1e3_scanString(__state, false)
	return ____rune_private_2013c1e3_lexed(__scanned.__state, func() __TokenKind {
		if __scanned.__ok {
			return __TokenKind_String
		}
		return __TokenKind_Illegal
	}())
}

func ____rune_private_2013c1e3_lexTemplateStringToken(__state __LexState) __Lexed {
	__scanned := ____rune_private_2013c1e3_scanTemplateString(__state, false)
	return ____rune_private_2013c1e3_lexed(__scanned.__state, func() __TokenKind {
		if __scanned.__ok {
			return __TokenKind_TemplateString
		}
		return __TokenKind_Illegal
	}())
}

func ____rune_private_2013c1e3_scanString(__state __LexState, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if ____rune_private_2013c1e3_atEnd(__state) {
			return ____rune_private_2013c1e3_scannedString(__state, false)
		}
		return ____rune_private_2013c1e3_scanStringStep(____rune_private_2013c1e3_advance(__state), __escaped)
	}()
}

func ____rune_private_2013c1e3_scanTemplateString(__state __LexState, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if ____rune_private_2013c1e3_atEnd(__state) {
			return ____rune_private_2013c1e3_scannedString(__state, false)
		}
		return ____rune_private_2013c1e3_scanTemplateStringStep(____rune_private_2013c1e3_advance(__state), __escaped)
	}()
}

func ____rune_private_2013c1e3_scannedString(__state __LexState, __ok bool) __ScannedString {
	return __ScannedString{__state: __state, __ok: __ok}
}

func ____rune_private_2013c1e3_scanStringStep(__step __Advanced, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if __escaped {
			return ____rune_private_2013c1e3_scanString(__step.__state, false)
		}
		return func() __ScannedString {
			switch {
			case __step.__ch == '\\':
				return ____rune_private_2013c1e3_scanString(__step.__state, true)
			case __step.__ch == '"':
				return ____rune_private_2013c1e3_scannedString(__step.__state, true)
			default:
				return ____rune_private_2013c1e3_scanString(__step.__state, false)
			}
		}()
	}()
}

func ____rune_private_2013c1e3_scanTemplateStringStep(__step __Advanced, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if __escaped {
			return ____rune_private_2013c1e3_scanTemplateString(__step.__state, false)
		}
		return func() __ScannedString {
			switch {
			case __step.__ch == '\\':
				return ____rune_private_2013c1e3_scanTemplateString(__step.__state, true)
			case __step.__ch == '`':
				return ____rune_private_2013c1e3_scannedString(__step.__state, true)
			default:
				return ____rune_private_2013c1e3_scanTemplateString(__step.__state, false)
			}
		}()
	}()
}

func ____rune_private_2013c1e3_lexCharToken(__state __LexState) __Lexed {
	__scanned := ____rune_private_2013c1e3_scanChar(__state, false)
	return ____rune_private_2013c1e3_lexed(__scanned.__state, func() __TokenKind {
		if __scanned.__ok {
			return __TokenKind_Char
		}
		return __TokenKind_Illegal
	}())
}

func ____rune_private_2013c1e3_scanChar(__state __LexState, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if ____rune_private_2013c1e3_atEnd(__state) {
			return ____rune_private_2013c1e3_scannedString(__state, false)
		}
		return ____rune_private_2013c1e3_scanCharStep(____rune_private_2013c1e3_advance(__state), __escaped)
	}()
}

func ____rune_private_2013c1e3_scanCharStep(__step __Advanced, __escaped bool) __ScannedString {
	return func() __ScannedString {
		switch {
		case __step.__ch == '\n':
			return ____rune_private_2013c1e3_scannedString(__step.__state, false)
		default:
			return func() __ScannedString {
				if __escaped {
					return ____rune_private_2013c1e3_scanChar(__step.__state, false)
				}
				return func() __ScannedString {
					switch {
					case __step.__ch == '\\':
						return ____rune_private_2013c1e3_scanChar(__step.__state, true)
					case __step.__ch == '\'':
						return ____rune_private_2013c1e3_scannedString(__step.__state, true)
					default:
						return ____rune_private_2013c1e3_scanChar(__step.__state, false)
					}
				}()
			}()
		}
	}()
}

func ____rune_private_2013c1e3_lexRegexToken(__state __LexState) __Lexed {
	__scanned := ____rune_private_2013c1e3_scanRegex(__state, false, false)
	return ____rune_private_2013c1e3_lexed(__scanned.__state, func() __TokenKind {
		if __scanned.__ok {
			return __TokenKind_Regex
		}
		return __TokenKind_Illegal
	}())
}

func ____rune_private_2013c1e3_scanRegex(__state __LexState, __escaped bool, __inClass bool) __ScannedString {
	return func() __ScannedString {
		if ____rune_private_2013c1e3_atEnd(__state) {
			return ____rune_private_2013c1e3_scannedString(__state, false)
		}
		return ____rune_private_2013c1e3_scanRegexStep(____rune_private_2013c1e3_advance(__state), __escaped, __inClass)
	}()
}

func ____rune_private_2013c1e3_scanRegexStep(__step __Advanced, __escaped bool, __inClass bool) __ScannedString {
	return func() __ScannedString {
		switch {
		case __step.__ch == '\n':
			return ____rune_private_2013c1e3_scannedString(__step.__state, false)
		default:
			return func() __ScannedString {
				if __escaped {
					return ____rune_private_2013c1e3_scanRegex(__step.__state, false, __inClass)
				}
				return func() __ScannedString {
					switch {
					case __step.__ch == '\\':
						return ____rune_private_2013c1e3_scanRegex(__step.__state, true, __inClass)
					case __step.__ch == '[':
						return ____rune_private_2013c1e3_scanRegex(__step.__state, false, true)
					case __step.__ch == ']':
						return ____rune_private_2013c1e3_scanRegex(__step.__state, false, false)
					case __step.__ch == '/':
						return func() __ScannedString {
							if __inClass {
								return ____rune_private_2013c1e3_scanRegex(__step.__state, false, __inClass)
							}
							return ____rune_private_2013c1e3_scannedString(____rune_private_2013c1e3_scanRegexFlags(__step.__state), true)
						}()
					default:
						return ____rune_private_2013c1e3_scanRegex(__step.__state, false, __inClass)
					}
				}()
			}()
		}
	}()
}

func ____rune_private_2013c1e3_scanRegexFlags(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_2013c1e3_isRegexFlag(____rune_private_2013c1e3_peek(__state)) {
			return ____rune_private_2013c1e3_scanRegexFlags(____rune_private_2013c1e3_advanceState(__state))
		}
		return __state
	}()
}

func ____rune_private_2013c1e3_lexNumberToken(__state __LexState) __Lexed {
	return ____rune_private_2013c1e3_lexNumberAfterDigits(____rune_private_2013c1e3_scanDigits(__state), false)
}

func ____rune_private_2013c1e3_scanDigits(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_2013c1e3_isDigit(____rune_private_2013c1e3_peek(__state)) {
			return ____rune_private_2013c1e3_scanDigits(____rune_private_2013c1e3_advanceState(__state))
		}
		return __state
	}()
}

func ____rune_private_2013c1e3_lexNumberAfterDigits(__state __LexState, __isDouble bool) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_2013c1e3_peek(__state) == 'n':
			return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_advanceState(__state), __TokenKind_BigInt)
		case ____rune_private_2013c1e3_peek(__state) == '.':
			return func() __Lexed {
				if ____rune_private_2013c1e3_isDigit(____rune_private_2013c1e3_peekNext(__state)) {
					return ____rune_private_2013c1e3_lexNumberAfterDot(____rune_private_2013c1e3_advanceState(__state))
				}
				return ____rune_private_2013c1e3_lexed(__state, func() __TokenKind {
					if __isDouble {
						return __TokenKind_Double
					}
					return __TokenKind_Int
				}())
			}()
		default:
			return func() __Lexed {
				if ____rune_private_2013c1e3_isExponentMarker(____rune_private_2013c1e3_peek(__state)) {
					return ____rune_private_2013c1e3_lexNumberAfterExponent(____rune_private_2013c1e3_advanceState(__state))
				}
				return ____rune_private_2013c1e3_lexed(__state, func() __TokenKind {
					if __isDouble {
						return __TokenKind_Double
					}
					return __TokenKind_Int
				}())
			}()
		}
	}()
}

func ____rune_private_2013c1e3_lexNumberAfterDot(__state __LexState) __Lexed {
	return ____rune_private_2013c1e3_lexNumberAfterDigits(____rune_private_2013c1e3_scanDigits(__state), true)
}

func ____rune_private_2013c1e3_lexNumberAfterExponent(__state __LexState) __Lexed {
	return func() __Lexed {
		if ____rune_private_2013c1e3_isExponentSign(____rune_private_2013c1e3_peek(__state)) {
			return ____rune_private_2013c1e3_lexNumberExponentDigits(____rune_private_2013c1e3_advanceState(__state))
		}
		return ____rune_private_2013c1e3_lexNumberExponentDigits(__state)
	}()
}

func ____rune_private_2013c1e3_lexNumberExponentDigits(__state __LexState) __Lexed {
	return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_scanDigits(__state), __TokenKind_Double)
}

func ____rune_private_2013c1e3_lexIdentifierToken(__state __LexState) __Lexed {
	return ____rune_private_2013c1e3_lexed(____rune_private_2013c1e3_scanIdentifier(__state), __TokenKind_Ident)
}

func ____rune_private_2013c1e3_scanIdentifier(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_2013c1e3_isIdentContinue(____rune_private_2013c1e3_peek(__state)) {
			return ____rune_private_2013c1e3_scanIdentifier(____rune_private_2013c1e3_advanceState(__state))
		}
		return __state
	}()
}

func ____rune_private_2013c1e3_isDigit(__ch rune) bool {
	switch {
	case (__ch >= '0' && __ch <= '9'):
		return true
	default:
		return false
	}
}

func ____rune_private_2013c1e3_isAlpha(__ch rune) bool {
	switch {
	case (__ch >= 'a' && __ch <= 'z'):
		return true
	case (__ch >= 'A' && __ch <= 'Z'):
		return true
	default:
		return false
	}
}

func ____rune_private_2013c1e3_isIdentStart(__ch rune) bool {
	switch {
	case __ch == '_':
		return true
	default:
		return ____rune_private_2013c1e3_isIdentText(__ch) && ____rune_private_2013c1e3_isDigit(__ch) == false
	}
}

func ____rune_private_2013c1e3_isIdentContinue(__ch rune) bool {
	return ____rune_private_2013c1e3_isIdentStart(__ch) || ____rune_private_2013c1e3_isDigit(__ch)
}

func ____rune_private_2013c1e3_isIdentText(__ch rune) bool {
	return ____rune_private_2013c1e3_isIdentBoundary(__ch) == false && ____rune_private_2013c1e3_isAsciiPunctuation(__ch) == false
}

func ____rune_private_2013c1e3_isIdentBoundary(__ch rune) bool {
	switch {
	case __ch == '\x00':
		return true
	case __ch == ' ':
		return true
	case __ch == '\t':
		return true
	case __ch == '\r':
		return true
	case __ch == '\n':
		return true
	default:
		return false
	}
}

func ____rune_private_2013c1e3_isAsciiPunctuation(__ch rune) bool {
	switch {
	case (__ch >= '!' && __ch <= '/'):
		return true
	case (__ch >= ':' && __ch <= '@'):
		return true
	case (__ch >= '[' && __ch <= '^'):
		return true
	case __ch == '`':
		return true
	case (__ch >= '{' && __ch <= '~'):
		return true
	default:
		return false
	}
}

func ____rune_private_2013c1e3_isRegexFlag(__ch rune) bool {
	return ____rune_private_2013c1e3_isAlpha(__ch)
}

func ____rune_private_2013c1e3_isExponentMarker(__ch rune) bool {
	switch {
	case __ch == 'e':
		return true
	case __ch == 'E':
		return true
	default:
		return false
	}
}

func ____rune_private_2013c1e3_isExponentSign(__ch rune) bool {
	switch {
	case __ch == '+':
		return true
	case __ch == '-':
		return true
	default:
		return false
	}
}

func __emptyParsedTypeRef() __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __TypeRefKind_Unknown, __name: "", __module: "", __nullable: false, __args: []__ParsedTypeRef{}, __params: []__ParsedTypeParam{}, __returnTypes: []__ParsedTypeRef{}, __line: 0, __column: 0}
}

func __emptyParsedTypeParam() __ParsedTypeParam {
	return __ParsedTypeParam{__name: "", __optional: false, __typeRef: __emptyParsedTypeRef()}
}

func __namedParsedTypeRef(__token __Token) __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __TypeRefKind_Name, __name: __token.__lexeme, __module: "", __nullable: false, __args: []__ParsedTypeRef{}, __params: []__ParsedTypeParam{}, __returnTypes: []__ParsedTypeRef{}, __line: __token.__line, __column: __token.__column}
}

func __qualifiedParsedTypeRef(__module __Token, __name __Token) __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __TypeRefKind_Name, __name: __name.__lexeme, __module: __module.__lexeme, __nullable: false, __args: []__ParsedTypeRef{}, __params: []__ParsedTypeParam{}, __returnTypes: []__ParsedTypeRef{}, __line: __module.__line, __column: __module.__column}
}

func __typeRefWithArgs(__typeRef __ParsedTypeRef, __args []__ParsedTypeRef) __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __typeRef.__kind, __name: __typeRef.__name, __module: __typeRef.__module, __nullable: __typeRef.__nullable, __args: __args, __params: __typeRef.__params, __returnTypes: __typeRef.__returnTypes, __line: __typeRef.__line, __column: __typeRef.__column}
}

func __nullableTypeRef(__typeRef __ParsedTypeRef) __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __typeRef.__kind, __name: __typeRef.__name, __module: __typeRef.__module, __nullable: true, __args: __typeRef.__args, __params: __typeRef.__params, __returnTypes: __typeRef.__returnTypes, __line: __typeRef.__line, __column: __typeRef.__column}
}

func __groupedTypeRef(__typeRef __ParsedTypeRef, __token __Token) __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __TypeRefKind_Group, __name: "", __module: "", __nullable: false, __args: []__ParsedTypeRef{__typeRef}, __params: []__ParsedTypeParam{}, __returnTypes: []__ParsedTypeRef{}, __line: __token.__line, __column: __token.__column}
}

func __tupleTypeRef(__params []__ParsedTypeParam, __token __Token) __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __TypeRefKind_Tuple, __name: "", __module: "", __nullable: false, __args: []__ParsedTypeRef{}, __params: __params, __returnTypes: []__ParsedTypeRef{}, __line: __token.__line, __column: __token.__column}
}

func __functionTypeRef(__params []__ParsedTypeParam, __returnType __ParsedTypeRef, __token __Token) __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __TypeRefKind_Function, __name: "", __module: "", __nullable: false, __args: []__ParsedTypeRef{}, __params: __params, __returnTypes: []__ParsedTypeRef{__returnType}, __line: __token.__line, __column: __token.__column}
}

func __typeRefToString(__typeRef __ParsedTypeRef) string {
	return func() string {
		switch {
		case __typeRef.__kind == __TypeRefKind_Name:
			return ____rune_private_61b01c1e_typeNameToString(__typeRef)
		case __typeRef.__kind == __TypeRefKind_Group:
			return func() string {
				if len(__typeRef.__args) == 0 {
					return "()"
				}
				return "(" + __typeRefToString(__typeRef.__args[0]) + ")"
			}()
		case __typeRef.__kind == __TypeRefKind_Tuple:
			return "(" + ____rune_private_61b01c1e_typeParamsToString(__typeRef.__params, 0, "") + ")"
		case __typeRef.__kind == __TypeRefKind_Function:
			return ____rune_private_61b01c1e_functionTypeToString(__typeRef)
		default:
			return ""
		}
	}()
}

func ____rune_private_61b01c1e_typeNameToString(__typeRef __ParsedTypeRef) string {
	__prefix := func() string {
		if __typeRef.__module == "" {
			return ""
		}
		return "@" + __typeRef.__module + "."
	}()
	__args := func() string {
		if len(__typeRef.__args) == 0 {
			return ""
		}
		return "[" + ____rune_private_61b01c1e_typeRefsToString(__typeRef.__args, 0, "") + "]"
	}()
	__nullable := func() string {
		if __typeRef.__nullable {
			return "?"
		}
		return ""
	}()
	return __prefix + __typeRef.__name + __args + __nullable
}

func ____rune_private_61b01c1e_functionTypeToString(__typeRef __ParsedTypeRef) string {
	__ret := func() string {
		if len(__typeRef.__returnTypes) == 0 {
			return ""
		}
		return __typeRefToString(__typeRef.__returnTypes[0])
	}()
	return "(" + ____rune_private_61b01c1e_typeParamsToString(__typeRef.__params, 0, "") + ")->" + __ret
}

func ____rune_private_61b01c1e_typeParamToString(__param __ParsedTypeParam) string {
	__prefix := func() string {
		if __param.__name == "" {
			return ""
		}
		return __param.__name + func() string {
			if __param.__optional {
				return "?:"
			}
			return ":"
		}()
	}()
	return __prefix + __typeRefToString(__param.__typeRef)
}

func ____rune_private_61b01c1e_typeRefsToString(__refs []__ParsedTypeRef, __index int, __out string) string {
	return func() string {
		if __index >= len(__refs) {
			return __out
		}
		return ____rune_private_61b01c1e_typeRefsToString(__refs, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ","
		}()+__typeRefToString(__refs[__index]))
	}()
}

func ____rune_private_61b01c1e_typeParamsToString(__params []__ParsedTypeParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_61b01c1e_typeParamsToString(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ","
		}()+____rune_private_61b01c1e_typeParamToString(__params[__index]))
	}()
}

func __exprKindName(__kind __ExprKind) string {
	return func() string {
		switch {
		case __kind == __ExprKind_Identifier:
			return "identifier"
		case __kind == __ExprKind_At:
			return "at"
		case __kind == __ExprKind_This:
			return "this"
		case __kind == __ExprKind_Int:
			return "int"
		case __kind == __ExprKind_Double:
			return "double"
		case __kind == __ExprKind_BigInt:
			return "bigint"
		case __kind == __ExprKind_String:
			return "string"
		case __kind == __ExprKind_Template:
			return "template"
		case __kind == __ExprKind_Char:
			return "char"
		case __kind == __ExprKind_Regex:
			return "regex"
		case __kind == __ExprKind_XMLText:
			return "xmlText"
		case __kind == __ExprKind_Bool:
			return "bool"
		case __kind == __ExprKind_Null:
			return "null"
		case __kind == __ExprKind_Unary:
			return "unary"
		case __kind == __ExprKind_Postfix:
			return "postfix"
		case __kind == __ExprKind_Unwrap:
			return "unwrap"
		case __kind == __ExprKind_CompileTime:
			return "compileTime"
		case __kind == __ExprKind_Binary:
			return "binary"
		case __kind == __ExprKind_Ternary:
			return "ternary"
		case __kind == __ExprKind_Assign:
			return "assign"
		case __kind == __ExprKind_Call:
			return "call"
		case __kind == __ExprKind_Args:
			return "args"
		case __kind == __ExprKind_Lambda:
			return "lambda"
		case __kind == __ExprKind_Selector:
			return "selector"
		case __kind == __ExprKind_Index:
			return "index"
		case __kind == __ExprKind_Array:
			return "array"
		case __kind == __ExprKind_Tuple:
			return "tuple"
		case __kind == __ExprKind_Map:
			return "map"
		case __kind == __ExprKind_Entry:
			return "entry"
		case __kind == __ExprKind_Spread:
			return "spread"
		case __kind == __ExprKind_Reactive:
			return "reactive"
		case __kind == __ExprKind_Struct:
			return "struct"
		case __kind == __ExprKind_Object:
			return "object"
		case __kind == __ExprKind_Field:
			return "field"
		case __kind == __ExprKind_PrivateField:
			return "privateField"
		case __kind == __ExprKind_Method:
			return "method"
		case __kind == __ExprKind_PrivateMethod:
			return "privateMethod"
		case __kind == __ExprKind_Block:
			return "block"
		case __kind == __ExprKind_PatternBlock:
			return "patternBlock"
		case __kind == __ExprKind_Match:
			return "match"
		case __kind == __ExprKind_Branch:
			return "branch"
		case __kind == __ExprKind_Pattern:
			return "pattern"
		case __kind == __ExprKind_Let:
			return "let"
		case __kind == __ExprKind_ObjectDestructure:
			return "objectDestructure"
		case __kind == __ExprKind_Error:
			return "error"
		case __kind == __ExprKind_Watch:
			return "watch"
		default:
			return "unknown"
		}
	}()
}

func __parse(__source string) __ParsedFile {
	return __parseTokens(__lex(__source))
}

func __parseTokens(__tokens []__Token) __ParsedFile {
	__errors := ____rune_private_b990f3d7_emptyParseErrors()
	return ____rune_private_b990f3d7_parseFileLoop(____rune_private_b990f3d7_parserSkipNewlines(__ParserState{__tokens: __tokens, __current: 0, __errors: __errors}), ____rune_private_b990f3d7_emptyFile(__errors)).__file
}

func ____rune_private_b990f3d7_emptyFile(__errors []__ParseError) __ParsedFile {
	return __ParsedFile{__imports: []__ParsedImport{}, __types: []__ParsedType{}, __functions: []__ParsedFunction{}, __tests: []__ParsedTest{}, __errors: __errors}
}

func ____rune_private_b990f3d7_emptyParseErrors() []__ParseError {
	return append([]__ParseError{}, []__ParseError{__ParseError{__message: "", __line: 0, __column: 0}}[0:0]...)
}

func ____rune_private_b990f3d7_emptyToken() __Token {
	return __Token{__kind: __TokenKind_EOF, __lexeme: "", __offset: 0, __line: 0, __column: 0}
}

func ____rune_private_b990f3d7_emptyExpr() __ParsedExpr {
	return __ParsedExpr{__kind: __ExprKind_Unknown, __text: "", __name: "", __value: "", __op: "", __params: []__ParsedParam{}, __children: []__ParsedExpr{}, __line: 0, __column: 0}
}

func ____rune_private_b990f3d7_emptyAnnotations() []__ParsedAnnotation {
	return append([]__ParsedAnnotation{}, []__ParsedAnnotation{__ParsedAnnotation{__marker: "", __module: "", __name: "", __args: []__ParsedExpr{}, __line: 0, __column: 0}}[0:0]...)
}

func ____rune_private_b990f3d7_emptyFunction() __ParsedFunction {
	return __ParsedFunction{__name: "", __private: true, __static: false, __routine: false, __macro: false, __annotations: ____rune_private_b990f3d7_emptyAnnotations(), __receiverType: "", __generics: []string{}, __params: []__ParsedParam{}, __returnType: __emptyParsedTypeRef(), __body: ____rune_private_b990f3d7_emptyExpr(), __line: 0, __column: 0}
}

func ____rune_private_b990f3d7_emptyType() __ParsedType {
	return __ParsedType{__name: "", __private: true, __enum: false, __annotations: ____rune_private_b990f3d7_emptyAnnotations(), __generics: []string{}, __fields: []__ParsedField{}, __methods: []__ParsedFunction{}, __members: []__ParsedEnumMember{}, __line: 0, __column: 0}
}

func ____rune_private_b990f3d7_emptyImport() __ParsedImport {
	return __ParsedImport{__path: "", __go: false, __line: 0, __column: 0}
}

func ____rune_private_b990f3d7_emptyTest() __ParsedTest {
	return __ParsedTest{__name: "", __body: ____rune_private_b990f3d7_emptyExpr(), __line: 0, __column: 0}
}

func ____rune_private_b990f3d7_emptyField() __ParsedField {
	return __ParsedField{__name: "", __private: true, __annotations: ____rune_private_b990f3d7_emptyAnnotations(), __typeRef: __emptyParsedTypeRef(), __line: 0, __column: 0}
}

func ____rune_private_b990f3d7_emptyMember() __ParsedEnumMember {
	return __ParsedEnumMember{__name: "", __private: true, __annotations: ____rune_private_b990f3d7_emptyAnnotations(), __value: "", __params: []__ParsedParam{}, __line: 0, __column: 0}
}

func ____rune_private_b990f3d7_makeExpr(__kind __ExprKind, __text string, __name string, __value string, __op string, __params []__ParsedParam, __children []__ParsedExpr, __line int, __column int) __ParsedExpr {
	return __ParsedExpr{__kind: __kind, __text: __text, __name: __name, __value: __value, __op: __op, __params: __params, __children: __children, __line: __line, __column: __column}
}

func ____rune_private_b990f3d7_node(__kind __ExprKind, __token __Token) __ParsedExpr {
	return ____rune_private_b990f3d7_makeExpr(__kind, __token.__lexeme, "", "", "", []__ParsedParam{}, []__ParsedExpr{}, __token.__line, __token.__column)
}

func ____rune_private_b990f3d7_namedNode(__kind __ExprKind, __name string, __token __Token) __ParsedExpr {
	return ____rune_private_b990f3d7_makeExpr(__kind, __token.__lexeme, __name, "", "", []__ParsedParam{}, []__ParsedExpr{}, __token.__line, __token.__column)
}

func ____rune_private_b990f3d7_valueNode(__kind __ExprKind, __value string, __token __Token) __ParsedExpr {
	return ____rune_private_b990f3d7_makeExpr(__kind, __token.__lexeme, "", __value, "", []__ParsedParam{}, []__ParsedExpr{}, __token.__line, __token.__column)
}

func ____rune_private_b990f3d7_opNode(__kind __ExprKind, __op string, __token __Token, __children []__ParsedExpr) __ParsedExpr {
	return ____rune_private_b990f3d7_makeExpr(__kind, __token.__lexeme, "", "", __op, []__ParsedParam{}, __children, __token.__line, __token.__column)
}

func ____rune_private_b990f3d7_withChildren(__expr __ParsedExpr, __children []__ParsedExpr) __ParsedExpr {
	return ____rune_private_b990f3d7_makeExpr(__expr.__kind, __expr.__text, __expr.__name, __expr.__value, __expr.__op, __expr.__params, __children, __expr.__line, __expr.__column)
}

func ____rune_private_b990f3d7_withParams(__expr __ParsedExpr, __params []__ParsedParam) __ParsedExpr {
	return ____rune_private_b990f3d7_makeExpr(__expr.__kind, __expr.__text, __expr.__name, __expr.__value, __expr.__op, __params, __expr.__children, __expr.__line, __expr.__column)
}

func ____rune_private_b990f3d7_withText(__expr __ParsedExpr, __text string) __ParsedExpr {
	return ____rune_private_b990f3d7_makeExpr(__expr.__kind, __text, __expr.__name, __expr.__value, __expr.__op, __expr.__params, __expr.__children, __expr.__line, __expr.__column)
}

func ____rune_private_b990f3d7_appendChild(__expr __ParsedExpr, __child __ParsedExpr) __ParsedExpr {
	__expr.__children = append(__expr.__children, __child)
	return __expr
}

func ____rune_private_b990f3d7_appendParam(__params []__ParsedParam, __param __ParsedParam) []__ParsedParam {
	__params = append(__params, __param)
	return __params
}

func ____rune_private_b990f3d7_appendString(__values []string, __value string) []string {
	__values = append(__values, __value)
	return __values
}

func ____rune_private_b990f3d7_parserPeek(__state __ParserState) __Token {
	return func() __Token {
		if __state.__current >= len(__state.__tokens) {
			return ____rune_private_b990f3d7_emptyToken()
		}
		return __state.__tokens[__state.__current]
	}()
}

func ____rune_private_b990f3d7_parserPrevious(__state __ParserState) __Token {
	return func() __Token {
		if __state.__current <= 0 {
			return ____rune_private_b990f3d7_emptyToken()
		}
		return __state.__tokens[__state.__current-1]
	}()
}

func ____rune_private_b990f3d7_parserTokenAt(__state __ParserState, __index int) __Token {
	return func() __Token {
		if __index >= len(__state.__tokens) {
			return ____rune_private_b990f3d7_emptyToken()
		}
		return __state.__tokens[__index]
	}()
}

func ____rune_private_b990f3d7_parserKindAt(__state __ParserState, __index int) __TokenKind {
	return ____rune_private_b990f3d7_parserTokenAt(__state, __index).__kind
}

func ____rune_private_b990f3d7_parserCheck(__state __ParserState, __kind __TokenKind) bool {
	return ____rune_private_b990f3d7_parserPeek(__state).__kind == __kind
}

func ____rune_private_b990f3d7_parserCheckNext(__state __ParserState, __kind __TokenKind) bool {
	return ____rune_private_b990f3d7_parserKindAt(__state, __state.__current+1) == __kind
}

func ____rune_private_b990f3d7_stateAt(__state __ParserState, __current int) __ParserState {
	return __ParserState{__tokens: __state.__tokens, __current: __current, __errors: __state.__errors}
}

func ____rune_private_b990f3d7_parserAdvance(__state __ParserState) __TokenStep {
	return func() __TokenStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_EOF) {
			return __TokenStep{__state: __state, __token: ____rune_private_b990f3d7_parserPeek(__state)}
		}
		return __TokenStep{__state: ____rune_private_b990f3d7_stateAt(__state, __state.__current+1), __token: ____rune_private_b990f3d7_parserPeek(__state)}
	}()
}

func ____rune_private_b990f3d7_parserMatch(__state __ParserState, __kind __TokenKind) __BoolStep {
	return func() __BoolStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __kind) {
			return __BoolStep{__state: ____rune_private_b990f3d7_parserAdvance(__state).__state, __ok: true}
		}
		return __BoolStep{__state: __state, __ok: false}
	}()
}

func ____rune_private_b990f3d7_parserConsume(__state __ParserState, __kind __TokenKind, __message string) __TokenStep {
	return func() __TokenStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __kind) {
			return ____rune_private_b990f3d7_parserAdvance(__state)
		}
		return ____rune_private_b990f3d7_parserConsumeMissing(____rune_private_b990f3d7_parserErrorAt(__state, ____rune_private_b990f3d7_parserPeek(__state), __message))
	}()
}

func ____rune_private_b990f3d7_parserConsumeMissing(__state __ParserState) __TokenStep {
	return func() __TokenStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_EOF) {
			return __TokenStep{__state: __state, __token: ____rune_private_b990f3d7_parserPeek(__state)}
		}
		return ____rune_private_b990f3d7_parserAdvance(__state)
	}()
}

func ____rune_private_b990f3d7_parserErrorAt(__state __ParserState, __token __Token, __message string) __ParserState {
	return __ParserState{__tokens: __state.__tokens, __current: __state.__current, __errors: ____rune_private_b990f3d7_appendParseError(__state.__errors, __ParseError{__message: __message, __line: __token.__line, __column: __token.__column})}
}

func ____rune_private_b990f3d7_appendParseError(__errors []__ParseError, __error __ParseError) []__ParseError {
	__out := __errors
	__out = append(__out, __error)
	return __out
}

func ____rune_private_b990f3d7_parserSkipNewlines(__state __ParserState) __ParserState {
	return func() __ParserState {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Newline) {
			return ____rune_private_b990f3d7_parserSkipNewlines(____rune_private_b990f3d7_parserAdvance(__state).__state)
		}
		return __state
	}()
}

func ____rune_private_b990f3d7_consumeStatementEnd(__state __ParserState) __ParserState {
	return func() __ParserState {
		if ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_Newline).__ok {
			return ____rune_private_b990f3d7_parserSkipNewlines(____rune_private_b990f3d7_parserAdvance(__state).__state)
		}
		return __state
	}()
}

func ____rune_private_b990f3d7_consumeFieldSeparator(__state __ParserState, __close __TokenKind, __message string) __ParserState {
	__current := ____rune_private_b990f3d7_consumeStatementEnd(__state)
	__comma := ____rune_private_b990f3d7_parserMatch(__current, __TokenKind_Comma)
	return func() __ParserState {
		if ____rune_private_b990f3d7_parserCheck(__current, __close) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return __current
		}
		return func() __ParserState {
			if __comma.__ok {
				return ____rune_private_b990f3d7_parserSkipNewlines(__comma.__state)
			}
			return ____rune_private_b990f3d7_parserErrorAt(__current, ____rune_private_b990f3d7_parserPeek(__current), __message)
		}()
	}()
}

func ____rune_private_b990f3d7_unquote(__raw string) string {
	return func() string {
		if len([]rune(__raw)) >= 2 {
			return func() string { runes := []rune(__raw); return string(runes[1 : len([]rune(__raw))-1]) }()
		}
		return __raw
	}()
}

func ____rune_private_b990f3d7_parseFileLoop(__state __ParserState, __file __ParsedFile) __FileStep {
	return func() __FileStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_EOF) {
			return __FileStep{__state: __state, __file: ____rune_private_b990f3d7_withFileErrors(__file, __state.__errors)}
		}
		return ____rune_private_b990f3d7_parseTopLevel(__state, __file)
	}()
}

func ____rune_private_b990f3d7_withFileErrors(__file __ParsedFile, __errors []__ParseError) __ParsedFile {
	return __ParsedFile{__imports: __file.__imports, __types: __file.__types, __functions: __file.__functions, __tests: __file.__tests, __errors: __errors}
}

func ____rune_private_b990f3d7_parseTopLevel(__state __ParserState, __file __ParsedFile) __FileStep {
	return func() __FileStep {
		if ____rune_private_b990f3d7_looksLikeMacroFunctionDecl(__state) {
			return ____rune_private_b990f3d7_parseTopLevelAfterResult(____rune_private_b990f3d7_parseMacroFunction(__state, __file))
		}
		return ____rune_private_b990f3d7_parseTopLevelAfterMacro(__state, __file)
	}()
}

func ____rune_private_b990f3d7_parseTopLevelAfterResult(__result __FileStep) __FileStep {
	return ____rune_private_b990f3d7_parseFileLoop(____rune_private_b990f3d7_parserSkipNewlines(__result.__state), __result.__file)
}

func ____rune_private_b990f3d7_parseTopLevelAfterMacro(__state __ParserState, __file __ParsedFile) __FileStep {
	__annotationStep := ____rune_private_b990f3d7_parseAnnotations(__state)
	__publicStep := ____rune_private_b990f3d7_parsePublicModifier(__annotationStep.__state)
	__current := __publicStep.__state
	__private := __publicStep.__ok == false
	__current = func() __ParserState {
		if __publicStep.__ok && (____rune_private_b990f3d7_parserCheck(__current, __TokenKind_At) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_Question)) {
			return ____rune_private_b990f3d7_parserErrorAt(__current, ____rune_private_b990f3d7_parserPeek(__current), "expected public declaration after '+'")
		}
		return __current
	}()
	__result := func() __FileStep {
		if !__publicStep.__ok && ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_At) {
			return ____rune_private_b990f3d7_parseTopLevelImport(__current, __file)
		}
		return func() __FileStep {
			if !__publicStep.__ok && ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_Question) {
				return ____rune_private_b990f3d7_parseTopLevelTest(__current, __file)
			}
			return func() __FileStep {
				if ____rune_private_b990f3d7_looksLikeTypeDecl(__current) {
					return ____rune_private_b990f3d7_parseTopLevelType(__current, __file, __private, __annotationStep.__annotations)
				}
				return func() __FileStep {
					if ____rune_private_b990f3d7_looksLikeFunctionDecl(__current) {
						return ____rune_private_b990f3d7_parseTopLevelFunction(__current, __file, __private, __annotationStep.__annotations)
					}
					return ____rune_private_b990f3d7_parseTopLevelError(__current, __file)
				}()
			}()
		}()
	}()
	return ____rune_private_b990f3d7_parseFileLoop(____rune_private_b990f3d7_parserSkipNewlines(__result.__state), __result.__file)
}

func ____rune_private_b990f3d7_parseMacroFunction(__state __ParserState, __file __ParsedFile) __FileStep {
	__marker := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Hash, "expected '#' before macro function")
	__step := ____rune_private_b990f3d7_parseFunctionWithReceiver(__marker.__state, "", false, false, true, ____rune_private_b990f3d7_emptyAnnotations())
	__file.__functions = append(__file.__functions, __step.__function)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_b990f3d7_parseTopLevelImport(__state __ParserState, __file __ParsedFile) __FileStep {
	__step := func() __ImportStep {
		if ____rune_private_b990f3d7_parserCheckNext(__state, __TokenKind_String) {
			return ____rune_private_b990f3d7_parseImportDecl(__state)
		}
		return ____rune_private_b990f3d7_parseGoImportDecl(__state)
	}()
	__file.__imports = append(__file.__imports, __step.__importDecl)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_b990f3d7_parseTopLevelTest(__state __ParserState, __file __ParsedFile) __FileStep {
	__step := ____rune_private_b990f3d7_parseTestDecl(__state)
	__file.__tests = append(__file.__tests, __step.__testDecl)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_b990f3d7_parseTopLevelType(__state __ParserState, __file __ParsedFile, __private bool, __annotations []__ParsedAnnotation) __FileStep {
	__step := ____rune_private_b990f3d7_parseTypeDecl(__state, __private, __annotations)
	__file.__types = append(__file.__types, __step.__typeDecl)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_b990f3d7_parseTopLevelFunction(__state __ParserState, __file __ParsedFile, __private bool, __annotations []__ParsedAnnotation) __FileStep {
	__step := ____rune_private_b990f3d7_parseFunctionWithReceiver(__state, "", __private, false, false, __annotations)
	__file.__functions = append(__file.__functions, __step.__function)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_b990f3d7_parseTopLevelError(__state __ParserState, __file __ParsedFile) __FileStep {
	return __FileStep{__state: ____rune_private_b990f3d7_parserAdvance(____rune_private_b990f3d7_parserErrorAt(__state, ____rune_private_b990f3d7_parserPeek(__state), "expected declaration")).__state, __file: __file}
}

func ____rune_private_b990f3d7_parsePublicModifier(__state __ParserState) __BoolStep {
	__step := ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_Plus)
	return __BoolStep{__state: func() __ParserState {
		if __step.__ok {
			return ____rune_private_b990f3d7_parserSkipNewlines(__step.__state)
		}
		return __state
	}(), __ok: __step.__ok}
}

func ____rune_private_b990f3d7_parseObjectPrivateModifier(__state __ParserState) __BoolStep {
	__step := ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_Minus)
	return __BoolStep{__state: func() __ParserState {
		if __step.__ok {
			return ____rune_private_b990f3d7_parserSkipNewlines(__step.__state)
		}
		return __state
	}(), __ok: __step.__ok}
}

func ____rune_private_b990f3d7_parseStaticMethodMarker(__state __ParserState) __BoolStep {
	__marker := func() __BoolStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_DoubleColon) {
			return ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_DoubleColon)
		}
		return func() __BoolStep {
			if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Ident) && ____rune_private_b990f3d7_parserPeek(__state).__lexeme == "static" {
				return __BoolStep{__state: ____rune_private_b990f3d7_parserAdvance(__state).__state, __ok: true}
			}
			return __BoolStep{__state: __state, __ok: false}
		}()
	}()
	return __BoolStep{__state: func() __ParserState {
		if __marker.__ok {
			return ____rune_private_b990f3d7_parserSkipNewlines(__marker.__state)
		}
		return __state
	}(), __ok: __marker.__ok}
}

func ____rune_private_b990f3d7_parseImportDecl(__state __ParserState) __ImportStep {
	__at := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_At, "expected '@'")
	__path := ____rune_private_b990f3d7_parserConsume(__at.__state, __TokenKind_String, "expected import path string after '@'")
	return __ImportStep{__state: __path.__state, __importDecl: __ParsedImport{__path: ____rune_private_b990f3d7_unquote(__path.__token.__lexeme), __go: false, __line: __at.__token.__line, __column: __at.__token.__column}}
}

func ____rune_private_b990f3d7_parseGoImportDecl(__state __ParserState) __ImportStep {
	__at := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_At, "expected '@'")
	__module := ____rune_private_b990f3d7_parserConsume(__at.__state, __TokenKind_Ident, "expected module name after '@'")
	__checked := func() __ParserState {
		if __module.__token.__lexeme == "go" {
			return __module.__state
		}
		return ____rune_private_b990f3d7_parserErrorAt(__module.__state, __module.__token, "only @go.import can appear at the top level")
	}()
	__dot := ____rune_private_b990f3d7_parserConsume(__checked, __TokenKind_Dot, "expected '.' after @go")
	__name := ____rune_private_b990f3d7_parserConsume(__dot.__state, __TokenKind_Ident, "expected import after @go.")
	__checkedName := func() __ParserState {
		if __name.__token.__lexeme == "import" {
			return __name.__state
		}
		return ____rune_private_b990f3d7_parserErrorAt(__name.__state, __name.__token, "only @go.import can appear at the top level")
	}()
	__open := ____rune_private_b990f3d7_parserConsume(__checkedName, __TokenKind_LParen, "expected '(' after @go.import")
	__path := ____rune_private_b990f3d7_parserConsume(__open.__state, __TokenKind_String, "expected Go import path string")
	__close := ____rune_private_b990f3d7_parserConsume(__path.__state, __TokenKind_RParen, "expected ')' after @go.import")
	return __ImportStep{__state: __close.__state, __importDecl: __ParsedImport{__path: ____rune_private_b990f3d7_unquote(__path.__token.__lexeme), __go: true, __line: __at.__token.__line, __column: __at.__token.__column}}
}

func ____rune_private_b990f3d7_parseTestDecl(__state __ParserState) __TestStep {
	__start := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Question, "expected '?'")
	__name := ____rune_private_b990f3d7_parserConsume(__start.__state, __TokenKind_String, "expected test name string after '?'")
	__bodyStart := ____rune_private_b990f3d7_parserSkipNewlines(__name.__state)
	__body := func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__bodyStart, __TokenKind_LBrace) {
			return ____rune_private_b990f3d7_parseBlock(__bodyStart)
		}
		return __ExprStep{__state: ____rune_private_b990f3d7_parserErrorAt(__bodyStart, ____rune_private_b990f3d7_parserPeek(__bodyStart), "expected test body block"), __expr: ____rune_private_b990f3d7_emptyExpr()}
	}()
	return __TestStep{__state: __body.__state, __testDecl: __ParsedTest{__name: ____rune_private_b990f3d7_unquote(__name.__token.__lexeme), __body: __body.__expr, __line: __start.__token.__line, __column: __start.__token.__column}}
}

func ____rune_private_b990f3d7_parseTypeDecl(__state __ParserState, __private bool, __annotations []__ParsedAnnotation) __TypeStep {
	__name := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected type name")
	__generics := ____rune_private_b990f3d7_parseGenericNames(__name.__state)
	__colon := ____rune_private_b990f3d7_parserConsume(__generics.__state, __TokenKind_Colon, "expected ':' after type name")
	__openStart := ____rune_private_b990f3d7_parserSkipNewlines(__colon.__state)
	__open := ____rune_private_b990f3d7_parserConsume(__openStart, __TokenKind_LBrace, "expected '{' after type declaration")
	__bodyStart := ____rune_private_b990f3d7_parserSkipNewlines(__open.__state)
	return func() __TypeStep {
		if ____rune_private_b990f3d7_looksLikeEnumMember(__bodyStart) {
			return ____rune_private_b990f3d7_parseEnumTypeBody(__bodyStart, __name.__token, __private, __annotations, __generics.__values)
		}
		return ____rune_private_b990f3d7_parseStructTypeBody(__bodyStart, __name.__token, __private, __annotations, __generics.__values)
	}()
}

func ____rune_private_b990f3d7_parseStructTypeBody(__state __ParserState, __name __Token, __private bool, __annotations []__ParsedAnnotation, __generics []string) __TypeStep {
	return ____rune_private_b990f3d7_parseStructTypeLoop(__state, __ParsedType{__name: __name.__lexeme, __private: __private, __enum: false, __annotations: __annotations, __generics: __generics, __fields: []__ParsedField{}, __methods: []__ParsedFunction{}, __members: []__ParsedEnumMember{}, __line: __name.__line, __column: __name.__column})
}

func ____rune_private_b990f3d7_parseStructTypeLoop(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __TypeStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_b990f3d7_finishType(__current, __typeDecl, "expected '}' after type declaration")
		}
		return ____rune_private_b990f3d7_parseStructTypeMember(__current, __typeDecl)
	}()
}

func ____rune_private_b990f3d7_parseStructTypeMember(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__annotationStep := ____rune_private_b990f3d7_parseAnnotations(__state)
	__current := __annotationStep.__state
	__privateStep := ____rune_private_b990f3d7_parseObjectPrivateModifier(__current)
	__memberState := __privateStep.__state
	__private := __privateStep.__ok
	__staticStep := func() __BoolStep {
		if ____rune_private_b990f3d7_looksLikeStaticFunctionDecl(__memberState) {
			return ____rune_private_b990f3d7_parseStaticMethodMarker(__memberState)
		}
		return __BoolStep{__state: __memberState, __ok: false}
	}()
	__parsed := func() __TypeStep {
		if ____rune_private_b990f3d7_looksLikeFunctionDecl(__staticStep.__state) {
			return ____rune_private_b990f3d7_parseStructMethod(__staticStep.__state, __typeDecl, __private, __staticStep.__ok, __annotationStep.__annotations)
		}
		return ____rune_private_b990f3d7_parseStructField(__staticStep.__state, __typeDecl, __private, __annotationStep.__annotations)
	}()
	__next := ____rune_private_b990f3d7_parserSkipNewlines(____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_consumeStatementEnd(__parsed.__state), __TokenKind_Comma).__state)
	return ____rune_private_b990f3d7_parseStructTypeLoop(__next, __parsed.__typeDecl)
}

func ____rune_private_b990f3d7_parseStructMethod(__state __ParserState, __typeDecl __ParsedType, __private bool, __static bool, __annotations []__ParsedAnnotation) __TypeStep {
	__step := ____rune_private_b990f3d7_parseFunctionWithReceiver(__state, __typeDecl.__name, __private, __static, false, __annotations)
	__typeDecl.__methods = append(__typeDecl.__methods, __step.__function)
	return __TypeStep{__state: __step.__state, __typeDecl: __typeDecl}
}

func ____rune_private_b990f3d7_parseStructField(__state __ParserState, __typeDecl __ParsedType, __private bool, __annotations []__ParsedAnnotation) __TypeStep {
	__field := ____rune_private_b990f3d7_parseFieldDecl(__state, __private, __annotations)
	__typeDecl.__fields = append(__typeDecl.__fields, __field.__field)
	return __TypeStep{__state: __field.__state, __typeDecl: __typeDecl}
}

func ____rune_private_b990f3d7_parseFieldDecl(__state __ParserState, __private bool, __annotations []__ParsedAnnotation) __FieldStep {
	__name := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected field name")
	__colon := ____rune_private_b990f3d7_parserConsume(__name.__state, __TokenKind_Colon, "expected ':' after field name")
	__typeRef := ____rune_private_b990f3d7_parseTypeRef(__colon.__state)
	return __FieldStep{__state: __typeRef.__state, __field: __ParsedField{__name: __name.__token.__lexeme, __private: __private, __annotations: __annotations, __typeRef: __typeRef.__typeRef, __line: __name.__token.__line, __column: __name.__token.__column}}
}

func ____rune_private_b990f3d7_finishType(__state __ParserState, __typeDecl __ParsedType, __message string) __TypeStep {
	__close := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_RBrace, __message)
	return __TypeStep{__state: __close.__state, __typeDecl: __typeDecl}
}

func ____rune_private_b990f3d7_parseEnumTypeBody(__state __ParserState, __name __Token, __private bool, __annotations []__ParsedAnnotation, __generics []string) __TypeStep {
	return ____rune_private_b990f3d7_parseEnumTypeLoop(__state, __ParsedType{__name: __name.__lexeme, __private: __private, __enum: true, __annotations: __annotations, __generics: __generics, __fields: []__ParsedField{}, __methods: []__ParsedFunction{}, __members: []__ParsedEnumMember{}, __line: __name.__line, __column: __name.__column})
}

func ____rune_private_b990f3d7_parseEnumTypeLoop(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __TypeStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_b990f3d7_finishType(__current, __typeDecl, "expected '}' after enum declaration")
		}
		return ____rune_private_b990f3d7_parseEnumTypeMember(__current, __typeDecl)
	}()
}

func ____rune_private_b990f3d7_parseEnumTypeMember(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__annotationStep := ____rune_private_b990f3d7_parseAnnotations(__state)
	__current := __annotationStep.__state
	__privateStep := ____rune_private_b990f3d7_parseObjectPrivateModifier(__current)
	__memberState := __privateStep.__state
	__staticStep := func() __BoolStep {
		if ____rune_private_b990f3d7_looksLikeStaticFunctionDecl(__memberState) {
			return ____rune_private_b990f3d7_parseStaticMethodMarker(__memberState)
		}
		return __BoolStep{__state: __memberState, __ok: false}
	}()
	__parsed := func() __TypeStep {
		if ____rune_private_b990f3d7_looksLikeFunctionDecl(__staticStep.__state) {
			return ____rune_private_b990f3d7_parseEnumMethod(__staticStep.__state, __typeDecl, __privateStep.__ok, __staticStep.__ok, __annotationStep.__annotations)
		}
		return ____rune_private_b990f3d7_parseEnumTypeMemberValue(__current, __typeDecl, __annotationStep.__annotations)
	}()
	__next := ____rune_private_b990f3d7_parserSkipNewlines(____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_consumeStatementEnd(__parsed.__state), __TokenKind_Comma).__state)
	return ____rune_private_b990f3d7_parseEnumTypeLoop(__next, __parsed.__typeDecl)
}

func ____rune_private_b990f3d7_parseEnumMethod(__state __ParserState, __typeDecl __ParsedType, __private bool, __static bool, __annotations []__ParsedAnnotation) __TypeStep {
	__step := ____rune_private_b990f3d7_parseFunctionWithReceiver(__state, __typeDecl.__name, __private, __static, false, __annotations)
	__typeDecl.__methods = append(__typeDecl.__methods, __step.__function)
	return __TypeStep{__state: __step.__state, __typeDecl: __typeDecl}
}

func ____rune_private_b990f3d7_parseEnumTypeMemberValue(__state __ParserState, __typeDecl __ParsedType, __annotations []__ParsedAnnotation) __TypeStep {
	__member := ____rune_private_b990f3d7_parseEnumMember(__state, __annotations)
	__typeDecl.__members = append(__typeDecl.__members, __member.__member)
	return __TypeStep{__state: __member.__state, __typeDecl: __typeDecl}
}

func ____rune_private_b990f3d7_parseEnumMember(__state __ParserState, __annotations []__ParsedAnnotation) __EnumMemberStep {
	__publicStep := ____rune_private_b990f3d7_parsePublicModifier(__state)
	__name := ____rune_private_b990f3d7_parserConsume(__publicStep.__state, __TokenKind_Ident, "expected enum member name")
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__name.__state)
	__parsed := ____rune_private_b990f3d7_parseEnumMemberPayload(__current)
	return __EnumMemberStep{__state: __parsed.__state, __member: __ParsedEnumMember{__name: __name.__token.__lexeme, __private: false, __annotations: __annotations, __value: __parsed.__value, __params: __parsed.__params, __line: __name.__token.__line, __column: __name.__token.__column}}
}

func ____rune_private_b990f3d7_parseEnumMemberPayload(__state __ParserState) __EnumMemberPayloadStep {
	__assign := ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_Assign)
	return func() __EnumMemberPayloadStep {
		if __assign.__ok {
			return ____rune_private_b990f3d7_parseEnumMemberValue(____rune_private_b990f3d7_parserSkipNewlines(__assign.__state))
		}
		return ____rune_private_b990f3d7_parseEnumMemberParams(__state)
	}()
}

func ____rune_private_b990f3d7_parseEnumMemberValue(__state __ParserState) __EnumMemberPayloadStep {
	__value := ____rune_private_b990f3d7_parseEnumValue(__state)
	return __EnumMemberPayloadStep{__state: __value.__state, __value: __value.__value, __params: []__ParsedParam{}}
}

func ____rune_private_b990f3d7_parseEnumMemberParams(__state __ParserState) __EnumMemberPayloadStep {
	__open := ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_LParen)
	return func() __EnumMemberPayloadStep {
		if __open.__ok {
			return ____rune_private_b990f3d7_parseEnumMemberParamList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state))
		}
		return __EnumMemberPayloadStep{__state: __state, __value: "", __params: []__ParsedParam{}}
	}()
}

func ____rune_private_b990f3d7_parseEnumMemberParamList(__state __ParserState) __EnumMemberPayloadStep {
	__params := ____rune_private_b990f3d7_parseParamList(__state)
	__close := ____rune_private_b990f3d7_parserConsume(__params.__state, __TokenKind_RParen, "expected ')' after enum constructor parameters")
	return __EnumMemberPayloadStep{__state: __close.__state, __value: "", __params: __params.__params}
}

func ____rune_private_b990f3d7_parseEnumValue(__state __ParserState) __StringStep {
	return func() __StringStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Minus) {
			return ____rune_private_b990f3d7_parseNegativeEnumValue(__state)
		}
		return ____rune_private_b990f3d7_parsePositiveEnumValue(__state)
	}()
}

func ____rune_private_b990f3d7_parseNegativeEnumValue(__state __ParserState) __StringStep {
	__minus := ____rune_private_b990f3d7_parserAdvance(__state)
	__value := ____rune_private_b990f3d7_parserConsume(__minus.__state, __TokenKind_Int, "expected integer enum value")
	return __StringStep{__state: __value.__state, __value: "-" + __value.__token.__lexeme}
}

func ____rune_private_b990f3d7_parsePositiveEnumValue(__state __ParserState) __StringStep {
	__value := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Int, "expected integer enum value")
	return __StringStep{__state: __value.__state, __value: __value.__token.__lexeme}
}

func ____rune_private_b990f3d7_parseFunctionWithReceiver(__state __ParserState, __receiverType string, __private bool, __static bool, __macro bool, __annotations []__ParsedAnnotation) __FunctionStep {
	__routineStep := ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_Tilde)
	__afterRoutine := func() __ParserState {
		if __routineStep.__ok {
			return ____rune_private_b990f3d7_parserSkipNewlines(__routineStep.__state)
		}
		return __state
	}()
	__name := ____rune_private_b990f3d7_parserConsume(__afterRoutine, __TokenKind_Ident, "expected function name")
	__generics := ____rune_private_b990f3d7_parseGenericNames(__name.__state)
	__open := ____rune_private_b990f3d7_parserConsume(__generics.__state, __TokenKind_LParen, "expected '(' after function name")
	__params := ____rune_private_b990f3d7_parseParamList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state))
	__close := ____rune_private_b990f3d7_parserConsume(__params.__state, __TokenKind_RParen, "expected ')' after parameter list")
	__afterClose := ____rune_private_b990f3d7_parserSkipNewlines(__close.__state)
	__ret := ____rune_private_b990f3d7_parserMatch(__afterClose, __TokenKind_Arrow)
	__returnType := func() __TypeRefStep {
		if __ret.__ok {
			return ____rune_private_b990f3d7_parseTypeRef(____rune_private_b990f3d7_parserSkipNewlines(__ret.__state))
		}
		return __TypeRefStep{__state: __afterClose, __typeRef: __emptyParsedTypeRef()}
	}()
	__arrow := ____rune_private_b990f3d7_parserConsume(____rune_private_b990f3d7_parserSkipNewlines(__returnType.__state), __TokenKind_FatArrow, "expected '=>' after function signature")
	__body := ____rune_private_b990f3d7_parseBody(____rune_private_b990f3d7_parserSkipNewlines(__arrow.__state))
	return __FunctionStep{__state: __body.__state, __function: __ParsedFunction{__name: __name.__token.__lexeme, __private: __private, __static: __static, __routine: __routineStep.__ok, __macro: __macro, __annotations: __annotations, __receiverType: __receiverType, __generics: __generics.__values, __params: __params.__params, __returnType: __returnType.__typeRef, __body: __body.__expr, __line: __name.__token.__line, __column: __name.__token.__column}}
}

func ____rune_private_b990f3d7_parseParamList(__state __ParserState) __ParamListStep {
	return ____rune_private_b990f3d7_parseParamListLoop(__state, []__ParsedParam{})
}

func ____rune_private_b990f3d7_parseParamListLoop(__state __ParserState, __params []__ParsedParam) __ParamListStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __ParamListStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RParen) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return __ParamListStep{__state: __current, __params: __params}
		}
		return ____rune_private_b990f3d7_parseOneParam(__current, __params)
	}()
}

func ____rune_private_b990f3d7_parseOneParam(__state __ParserState, __params []__ParsedParam) __ParamListStep {
	__name := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected parameter name")
	__optional := ____rune_private_b990f3d7_parserMatch(__name.__state, __TokenKind_Question)
	__colon := ____rune_private_b990f3d7_parserMatch(__optional.__state, __TokenKind_Colon)
	__typeRef := func() __TypeRefStep {
		if __colon.__ok {
			return ____rune_private_b990f3d7_parseTypeRef(__colon.__state)
		}
		return __TypeRefStep{__state: __optional.__state, __typeRef: __emptyParsedTypeRef()}
	}()
	__paramName := func() string {
		if __optional.__ok {
			return __name.__token.__lexeme + "?"
		}
		return __name.__token.__lexeme
	}()
	__params = append(__params, __ParsedParam{__name: __paramName, __typeRef: __typeRef.__typeRef, __line: __name.__token.__line, __column: __name.__token.__column})
	__comma := ____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_parserSkipNewlines(__typeRef.__state), __TokenKind_Comma)
	return ____rune_private_b990f3d7_parseParamListLoop(__comma.__state, __params)
}

func ____rune_private_b990f3d7_parseGenericNames(__state __ParserState) __StringListStep {
	__open := ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_LBracket)
	return func() __StringListStep {
		if __open.__ok {
			return ____rune_private_b990f3d7_parseGenericNameLoop(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), []string{})
		}
		return __StringListStep{__state: __state, __values: []string{}}
	}()
}

func ____rune_private_b990f3d7_parseGenericNameLoop(__state __ParserState, __values []string) __StringListStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __StringListStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBracket) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_b990f3d7_finishGenericNames(__current, __values)
		}
		return ____rune_private_b990f3d7_parseGenericNameValue(__current, __values)
	}()
}

func ____rune_private_b990f3d7_finishGenericNames(__state __ParserState, __values []string) __StringListStep {
	__close := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_RBracket, "expected ']' after generic parameters")
	return __StringListStep{__state: __close.__state, __values: __values}
}

func ____rune_private_b990f3d7_parseGenericNameValue(__state __ParserState, __values []string) __StringListStep {
	__nameStep := func() __StringListStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Ident) {
			return ____rune_private_b990f3d7_appendGenericName(__state, __values)
		}
		return __StringListStep{__state: ____rune_private_b990f3d7_parserAdvance(__state).__state, __values: __values}
	}()
	__colon := ____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_parserSkipNewlines(__nameStep.__state), __TokenKind_Colon)
	__current := func() __ParserState {
		if __colon.__ok {
			return ____rune_private_b990f3d7_parseTypeRef(____rune_private_b990f3d7_parserSkipNewlines(__colon.__state)).__state
		}
		return __nameStep.__state
	}()
	__comma := ____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_parserSkipNewlines(__current), __TokenKind_Comma)
	return ____rune_private_b990f3d7_parseGenericNameLoop(__comma.__state, __nameStep.__values)
}

func ____rune_private_b990f3d7_appendGenericName(__state __ParserState, __values []string) __StringListStep {
	__step := ____rune_private_b990f3d7_parserAdvance(__state)
	return __StringListStep{__state: __step.__state, __values: ____rune_private_b990f3d7_appendString(__values, __step.__token.__lexeme)}
}

func ____rune_private_b990f3d7_parseTypeRef(__state __ParserState) __TypeRefStep {
	return ____rune_private_b990f3d7_parseTypeRefPostfix(____rune_private_b990f3d7_parseTypeRefAtom(____rune_private_b990f3d7_parserSkipNewlines(__state)))
}

func ____rune_private_b990f3d7_parseTypeRefAtom(__state __ParserState) __TypeRefStep {
	return func() __TypeRefStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LParen) {
			return ____rune_private_b990f3d7_parseParenTypeRef(__state)
		}
		return func() __TypeRefStep {
			if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_At) {
				return ____rune_private_b990f3d7_parseQualifiedTypeRef(__state)
			}
			return func() __TypeRefStep {
				if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Ident) {
					return ____rune_private_b990f3d7_parseNamedTypeRef(__state)
				}
				return ____rune_private_b990f3d7_parseTypeRefError(__state)
			}()
		}()
	}()
}

func ____rune_private_b990f3d7_parseNamedTypeRef(__state __ParserState) __TypeRefStep {
	__name := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected type name")
	return __TypeRefStep{__state: __name.__state, __typeRef: __namedParsedTypeRef(__name.__token)}
}

func ____rune_private_b990f3d7_parseQualifiedTypeRef(__state __ParserState) __TypeRefStep {
	__at := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_At, "expected '@'")
	__module := ____rune_private_b990f3d7_parserConsume(__at.__state, __TokenKind_Ident, "expected module name after '@'")
	__dot := ____rune_private_b990f3d7_parserConsume(__module.__state, __TokenKind_Dot, "expected '.' after module name")
	__name := ____rune_private_b990f3d7_parserConsume(__dot.__state, __TokenKind_Ident, "expected type name after module qualifier")
	return __TypeRefStep{__state: __name.__state, __typeRef: __qualifiedParsedTypeRef(__module.__token, __name.__token)}
}

func ____rune_private_b990f3d7_parseParenTypeRef(__state __ParserState) __TypeRefStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LParen, "expected '(' before type")
	__params := ____rune_private_b990f3d7_parseTypeParamList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), []__ParsedTypeParam{})
	__close := ____rune_private_b990f3d7_parserConsume(__params.__state, __TokenKind_RParen, "expected ')' after type")
	__afterClose := ____rune_private_b990f3d7_parserSkipNewlines(__close.__state)
	__arrow := ____rune_private_b990f3d7_parserMatch(__afterClose, __TokenKind_Arrow)
	return func() __TypeRefStep {
		if __arrow.__ok {
			return ____rune_private_b990f3d7_finishFunctionTypeRef(____rune_private_b990f3d7_parserSkipNewlines(__arrow.__state), __open.__token, __params.__params)
		}
		return ____rune_private_b990f3d7_finishParenTypeRef(__close.__state, __open.__token, __params.__params)
	}()
}

func ____rune_private_b990f3d7_finishFunctionTypeRef(__state __ParserState, __token __Token, __params []__ParsedTypeParam) __TypeRefStep {
	__ret := ____rune_private_b990f3d7_parseTypeRef(__state)
	return __TypeRefStep{__state: __ret.__state, __typeRef: __functionTypeRef(__params, __ret.__typeRef, __token)}
}

func ____rune_private_b990f3d7_finishParenTypeRef(__state __ParserState, __token __Token, __params []__ParsedTypeParam) __TypeRefStep {
	return __TypeRefStep{__state: __state, __typeRef: func() __ParsedTypeRef {
		if len(__params) == 1 && __params[0].__name == "" {
			return __groupedTypeRef(__params[0].__typeRef, __token)
		}
		return __tupleTypeRef(__params, __token)
	}()}
}

func ____rune_private_b990f3d7_parseTypeRefPostfix(__step __TypeRefStep) __TypeRefStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__step.__state)
	return func() __TypeRefStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_LBracket) {
			return ____rune_private_b990f3d7_parseTypeRefPostfix(____rune_private_b990f3d7_parseTypeRefArgs(__current, __step.__typeRef))
		}
		return func() __TypeRefStep {
			if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_Question) {
				return ____rune_private_b990f3d7_parseTypeRefPostfix(____rune_private_b990f3d7_parseNullableTypeRef(__current, __step.__typeRef))
			}
			return __TypeRefStep{__state: __step.__state, __typeRef: __step.__typeRef}
		}()
	}()
}

func ____rune_private_b990f3d7_parseTypeRefArgs(__state __ParserState, __typeRef __ParsedTypeRef) __TypeRefStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBracket, "expected '[' after type name")
	__args := ____rune_private_b990f3d7_parseTypeRefList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), []__ParsedTypeRef{})
	__close := ____rune_private_b990f3d7_parserConsume(__args.__state, __TokenKind_RBracket, "expected ']' after type arguments")
	return __TypeRefStep{__state: __close.__state, __typeRef: __typeRefWithArgs(__typeRef, __args.__refs)}
}

func ____rune_private_b990f3d7_parseNullableTypeRef(__state __ParserState, __typeRef __ParsedTypeRef) __TypeRefStep {
	__question := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Question, "expected '?' after type")
	return __TypeRefStep{__state: __question.__state, __typeRef: __nullableTypeRef(__typeRef)}
}

func ____rune_private_b990f3d7_parseTypeRefList(__state __ParserState, __refs []__ParsedTypeRef) __TypeRefListStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __TypeRefListStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBracket) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return __TypeRefListStep{__state: __current, __refs: __refs}
		}
		return ____rune_private_b990f3d7_parseOneTypeRefListValue(__current, __refs)
	}()
}

func ____rune_private_b990f3d7_parseOneTypeRefListValue(__state __ParserState, __refs []__ParsedTypeRef) __TypeRefListStep {
	__typeRef := ____rune_private_b990f3d7_parseTypeRef(__state)
	__refs = append(__refs, __typeRef.__typeRef)
	__comma := ____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_parserSkipNewlines(__typeRef.__state), __TokenKind_Comma)
	return func() __TypeRefListStep {
		switch {
		case __comma.__ok == true:
			return ____rune_private_b990f3d7_parseTypeRefList(__comma.__state, __refs)
		default:
			return __TypeRefListStep{__state: __typeRef.__state, __refs: __refs}
		}
	}()
}

func ____rune_private_b990f3d7_parseTypeParamList(__state __ParserState, __params []__ParsedTypeParam) __TypeParamListStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __TypeParamListStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RParen) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return __TypeParamListStep{__state: __current, __params: __params}
		}
		return ____rune_private_b990f3d7_parseOneTypeParam(__current, __params)
	}()
}

func ____rune_private_b990f3d7_parseOneTypeParam(__state __ParserState, __params []__ParsedTypeParam) __TypeParamListStep {
	__param := ____rune_private_b990f3d7_parseTypeParam(__state)
	__params = append(__params, __param.__param)
	__comma := ____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_parserSkipNewlines(__param.__state), __TokenKind_Comma)
	return func() __TypeParamListStep {
		switch {
		case __comma.__ok == true:
			return ____rune_private_b990f3d7_parseTypeParamList(__comma.__state, __params)
		default:
			return __TypeParamListStep{__state: __param.__state, __params: __params}
		}
	}()
}

func ____rune_private_b990f3d7_parseTypeParam(__state __ParserState) __TypeParamStep {
	__named := ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Ident) && ____rune_private_b990f3d7_typeParamHasName(__state)
	return func() __TypeParamStep {
		if __named {
			return ____rune_private_b990f3d7_parseNamedTypeParam(__state)
		}
		return ____rune_private_b990f3d7_parseUnnamedTypeParam(__state)
	}()
}

func ____rune_private_b990f3d7_parseNamedTypeParam(__state __ParserState) __TypeParamStep {
	__name := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected type parameter name")
	__optional := ____rune_private_b990f3d7_parserMatch(__name.__state, __TokenKind_Question)
	__colon := ____rune_private_b990f3d7_parserConsume(__optional.__state, __TokenKind_Colon, "expected ':' after type parameter name")
	__typeRef := ____rune_private_b990f3d7_parseTypeRef(____rune_private_b990f3d7_parserSkipNewlines(__colon.__state))
	return __TypeParamStep{__state: __typeRef.__state, __param: __ParsedTypeParam{__name: __name.__token.__lexeme, __optional: __optional.__ok, __typeRef: __typeRef.__typeRef}}
}

func ____rune_private_b990f3d7_parseUnnamedTypeParam(__state __ParserState) __TypeParamStep {
	__typeRef := ____rune_private_b990f3d7_parseTypeRef(__state)
	return __TypeParamStep{__state: __typeRef.__state, __param: __ParsedTypeParam{__name: "", __optional: false, __typeRef: __typeRef.__typeRef}}
}

func ____rune_private_b990f3d7_typeParamHasName(__state __ParserState) bool {
	return ____rune_private_b990f3d7_parserKindAt(__state, __state.__current+1) == __TokenKind_Colon || ____rune_private_b990f3d7_parserKindAt(__state, __state.__current+1) == __TokenKind_Question && ____rune_private_b990f3d7_parserKindAt(__state, __state.__current+2) == __TokenKind_Colon
}

func ____rune_private_b990f3d7_parseTypeRefError(__state __ParserState) __TypeRefStep {
	return __TypeRefStep{__state: ____rune_private_b990f3d7_parserErrorAt(__state, ____rune_private_b990f3d7_parserPeek(__state), "expected type name"), __typeRef: __emptyParsedTypeRef()}
}

func ____rune_private_b990f3d7_parseBody(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LBrace) {
			return ____rune_private_b990f3d7_parseBraceBody(__state)
		}
		return ____rune_private_b990f3d7_parseExpression(__state, 1)
	}()
}

func ____rune_private_b990f3d7_parseBraceBody(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_looksLikePatternBranch(__state) == false && ____rune_private_b990f3d7_looksLikeMapLiteralBody(__state) {
			return ____rune_private_b990f3d7_parseMapLiteral(__state)
		}
		return func() __ExprStep {
			if ____rune_private_b990f3d7_looksLikePatternBranch(__state) == false && ____rune_private_b990f3d7_looksLikeObjectLiteralBody(__state) {
				return ____rune_private_b990f3d7_parseObjectLiteral(__state)
			}
			return ____rune_private_b990f3d7_parseBlock(__state)
		}()
	}()
}

func ____rune_private_b990f3d7_looksLikeObjectLiteralBody(__state __ParserState) bool {
	__first := ____rune_private_b990f3d7_skipNewlinesAt(__state, __state.__current+1)
	return ____rune_private_b990f3d7_parserKindAt(__state, __first) == __TokenKind_Ident && ____rune_private_b990f3d7_parserKindAt(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __first+1)) == __TokenKind_Colon
}

func ____rune_private_b990f3d7_parseBlock(__state __ParserState) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBrace, "expected '{'")
	__bodyStart := ____rune_private_b990f3d7_parserSkipNewlines(__open.__state)
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__bodyStart, __TokenKind_RBrace) {
			return ____rune_private_b990f3d7_finishBlock(__bodyStart, ____rune_private_b990f3d7_node(__ExprKind_Block, __open.__token))
		}
		return func() __ExprStep {
			if ____rune_private_b990f3d7_looksLikePatternBranch(__bodyStart) {
				return ____rune_private_b990f3d7_parsePatternBlock(__bodyStart, ____rune_private_b990f3d7_node(__ExprKind_PatternBlock, __open.__token))
			}
			return ____rune_private_b990f3d7_parseBlockLoop(__bodyStart, ____rune_private_b990f3d7_node(__ExprKind_Block, __open.__token))
		}()
	}()
}

func ____rune_private_b990f3d7_finishBlock(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__close := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_RBrace, "expected '}' after block")
	return __ExprStep{__state: __close.__state, __expr: __block}
}

func ____rune_private_b990f3d7_parseBlockLoop(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_b990f3d7_finishBlock(__current, __block)
		}
		return ____rune_private_b990f3d7_parseBlockStatement(__current, __block)
	}()
}

func ____rune_private_b990f3d7_parseBlockStatement(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__stmt := ____rune_private_b990f3d7_parseStatement(__state)
	__nextBlock := ____rune_private_b990f3d7_appendChild(__block, __stmt.__expr)
	return ____rune_private_b990f3d7_parseBlockLoop(____rune_private_b990f3d7_consumeStatementEnd(__stmt.__state), __nextBlock)
}

func ____rune_private_b990f3d7_parseStatement(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LBrace) && ____rune_private_b990f3d7_looksLikeObjectDestructureDecl(__state) {
			return ____rune_private_b990f3d7_parseObjectDestructureStatement(__state)
		}
		return func() __ExprStep {
			if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LBrace) {
				return ____rune_private_b990f3d7_parseBlock(__state)
			}
			return func() __ExprStep {
				if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Dollar) && ____rune_private_b990f3d7_parserCheckNext(__state, __TokenKind_Ident) {
					return ____rune_private_b990f3d7_parseDollarStatement(__state)
				}
				return func() __ExprStep {
					if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Ident) && (____rune_private_b990f3d7_parserCheckNext(__state, __TokenKind_Declare) || ____rune_private_b990f3d7_parserCheckNext(__state, __TokenKind_MutDeclare)) {
						return ____rune_private_b990f3d7_parseLetStatement(__state)
					}
					return func() __ExprStep {
						if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Ident) && ____rune_private_b990f3d7_parserCheckNext(__state, __TokenKind_Assign) {
							return ____rune_private_b990f3d7_parseAssignStatement(__state)
						}
						return ____rune_private_b990f3d7_parseExpression(__state, 1)
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_b990f3d7_parseDollarStatement(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserKindAt(__state, __state.__current+2) == __TokenKind_Declare {
			return ____rune_private_b990f3d7_parseSignalPrefixLetStatement(__state)
		}
		return ____rune_private_b990f3d7_parseExpression(__state, 1)
	}()
}

func ____rune_private_b990f3d7_parseSignalPrefixLetStatement(__state __ParserState) __ExprStep {
	__dollar := ____rune_private_b990f3d7_parserAdvance(__state)
	__name := ____rune_private_b990f3d7_parserAdvance(__dollar.__state)
	__op := ____rune_private_b990f3d7_parserAdvance(__name.__state)
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__op.__state), 1)
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Let, "$"+__name.__token.__lexeme, __name.__token.__lexeme, "", __op.__token.__lexeme, []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)}
}

func ____rune_private_b990f3d7_parseLetStatement(__state __ParserState) __ExprStep {
	__name := ____rune_private_b990f3d7_parserAdvance(__state)
	__op := ____rune_private_b990f3d7_parserAdvance(__name.__state)
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__op.__state), 1)
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Let, __name.__token.__lexeme, __name.__token.__lexeme, "", __op.__token.__lexeme, []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)}
}

func ____rune_private_b990f3d7_parseObjectDestructureStatement(__state __ParserState) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBrace, "expected '{' before object destructuring")
	__fields := ____rune_private_b990f3d7_parseObjectBindingList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state))
	__close := ____rune_private_b990f3d7_parserConsume(__fields.__state, __TokenKind_RBrace, "expected '}' after object destructuring")
	__op := ____rune_private_b990f3d7_parserAdvance(____rune_private_b990f3d7_parserSkipNewlines(__close.__state))
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__op.__state), 1)
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_ObjectDestructure, "", "", "", __op.__token.__lexeme, __fields.__params, []__ParsedExpr{__value.__expr}, __open.__token.__line, __open.__token.__column)}
}

func ____rune_private_b990f3d7_parseObjectBindingList(__state __ParserState) __ParamListStep {
	return ____rune_private_b990f3d7_parseObjectBindingListLoop(__state, []__ParsedParam{})
}

func ____rune_private_b990f3d7_parseObjectBindingListLoop(__state __ParserState, __fields []__ParsedParam) __ParamListStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __ParamListStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return __ParamListStep{__state: __current, __params: __fields}
		}
		return ____rune_private_b990f3d7_parseOneObjectBinding(__current, __fields)
	}()
}

func ____rune_private_b990f3d7_parseOneObjectBinding(__state __ParserState, __fields []__ParsedParam) __ParamListStep {
	__field := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected field name in object destructuring")
	__aliasStart := ____rune_private_b990f3d7_parserMatch(__field.__state, __TokenKind_Colon)
	__name := func() __TokenStep {
		if __aliasStart.__ok {
			return ____rune_private_b990f3d7_parserConsume(____rune_private_b990f3d7_parserSkipNewlines(__aliasStart.__state), __TokenKind_Ident, "expected binding name after ':'")
		}
		return __field
	}()
	__fields = append(__fields, __ParsedParam{__name: __name.__token.__lexeme, __typeRef: __namedParsedTypeRef(__field.__token), __line: __name.__token.__line, __column: __name.__token.__column})
	__comma := ____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_parserSkipNewlines(__name.__state), __TokenKind_Comma)
	return func() __ParamListStep {
		if __comma.__ok {
			return ____rune_private_b990f3d7_parseObjectBindingListLoop(__comma.__state, __fields)
		}
		return __ParamListStep{__state: __name.__state, __params: __fields}
	}()
}

func ____rune_private_b990f3d7_parseAssignStatement(__state __ParserState) __ExprStep {
	__name := ____rune_private_b990f3d7_parserAdvance(__state)
	__op := ____rune_private_b990f3d7_parserAdvance(__name.__state)
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__op.__state), 1)
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Assign, __name.__token.__lexeme, __name.__token.__lexeme, "", __op.__token.__lexeme, []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)}
}

func ____rune_private_b990f3d7_parseExpression(__state __ParserState, __minPrec int) __ExprStep {
	return func() __ExprStep {
		if __minPrec <= 1 && ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LParen) && ____rune_private_b990f3d7_looksLikeLambda(__state) {
			return ____rune_private_b990f3d7_parseLambda(__state)
		}
		return ____rune_private_b990f3d7_parseExpressionLoop(____rune_private_b990f3d7_parseUnary(__state), __minPrec)
	}()
}

func ____rune_private_b990f3d7_parseExpressionLoop(__left __ExprStep, __minPrec int) __ExprStep {
	__state := __left.__state
	__expr := __left.__expr
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LBrace) {
			return ____rune_private_b990f3d7_parseAfterBraceExpression(__state, __expr, __minPrec)
		}
		return func() __ExprStep {
			if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LParen) {
				return ____rune_private_b990f3d7_parseCallExpression(__state, __expr, __minPrec)
			}
			return func() __ExprStep {
				if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LBracket) {
					return ____rune_private_b990f3d7_parseIndexExpression(__state, __expr, __minPrec)
				}
				return func() __ExprStep {
					if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Dot) || ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_DoubleColon) {
						return ____rune_private_b990f3d7_parseSelectorExpression(__state, __expr, __minPrec)
					}
					return func() __ExprStep {
						if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_PlusPlus) {
							return ____rune_private_b990f3d7_parsePostfixExpression(__state, __expr, __minPrec)
						}
						return func() __ExprStep {
							if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Apostrophe) {
								return ____rune_private_b990f3d7_parseCompileTimePostfixExpression(__state, __expr, __minPrec)
							}
							return func() __ExprStep {
								if __minPrec <= 1 && ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Arrow) {
									return ____rune_private_b990f3d7_parseWatchExpression(__state, __expr, __minPrec)
								}
								return func() __ExprStep {
									if __minPrec <= 1 && ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Assign) {
										return ____rune_private_b990f3d7_parseAssignmentExpression(__state, __expr, __minPrec)
									}
									return func() __ExprStep {
										if __minPrec <= 1 && ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Tilde) {
											return ____rune_private_b990f3d7_parsePatternPredicateExpression(__state, __expr, __minPrec)
										}
										return func() __ExprStep {
											if __minPrec <= 1 && ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Question) {
												return ____rune_private_b990f3d7_parseQuestionExpression(__state, __expr, __minPrec)
											}
											return ____rune_private_b990f3d7_parseBinaryExpression(__state, __expr, __minPrec)
										}()
									}()
								}()
							}()
						}()
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_b990f3d7_parseAfterBraceExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_looksLikePatternBlockAfterSubject(__state) {
			return ____rune_private_b990f3d7_parseExpressionLoop(____rune_private_b990f3d7_parseMatchExpression(__state, __expr), __minPrec)
		}
		return func() __ExprStep {
			if __expr.__kind == __ExprKind_Identifier {
				return ____rune_private_b990f3d7_parseExpressionLoop(____rune_private_b990f3d7_parseStructLiteral(__state, __expr), __minPrec)
			}
			return __ExprStep{__state: __state, __expr: __expr}
		}()
	}()
}

func ____rune_private_b990f3d7_parseCallExpression(__state __ParserState, __callee __ParsedExpr, __minPrec int) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LParen, "expected '(' after callee")
	__args := ____rune_private_b990f3d7_parseArgumentList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), []__ParsedExpr{__callee}, __TokenKind_RParen)
	__close := ____rune_private_b990f3d7_parserConsume(__args.__state, __TokenKind_RParen, "expected ')' after arguments")
	__call := ____rune_private_b990f3d7_makeExpr(__ExprKind_Call, ____rune_private_b990f3d7_calleeText(__callee), "", "", "", []__ParsedParam{}, __args.__expr.__children, __callee.__line, __callee.__column)
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __close.__state, __expr: __call}, __minPrec)
}

func ____rune_private_b990f3d7_parseArgumentList(__state __ParserState, __holderChildren []__ParsedExpr, __endKind __TokenKind) __ExprStep {
	__holder := ____rune_private_b990f3d7_makeExpr(__ExprKind_Args, "", "", "", "", []__ParsedParam{}, __holderChildren, 0, 0)
	return ____rune_private_b990f3d7_parseArgumentListLoop(__state, __holder, __endKind)
}

func ____rune_private_b990f3d7_parseArgumentListLoop(__state __ParserState, __holder __ParsedExpr, __endKind __TokenKind) __ExprStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __endKind) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return __ExprStep{__state: __current, __expr: __holder}
		}
		return ____rune_private_b990f3d7_parseOneArgument(__current, __holder, __endKind)
	}()
}

func ____rune_private_b990f3d7_parseOneArgument(__state __ParserState, __holder __ParsedExpr, __endKind __TokenKind) __ExprStep {
	__spread := ____rune_private_b990f3d7_parserMatch(__state, __TokenKind_DotDot)
	__value := ____rune_private_b990f3d7_parseExpression(func() __ParserState {
		if __spread.__ok {
			return __spread.__state
		}
		return __state
	}(), 1)
	__arg := func() __ParsedExpr {
		if __spread.__ok {
			return ____rune_private_b990f3d7_opNode(__ExprKind_Spread, "..", ____rune_private_b990f3d7_parserPrevious(__spread.__state), []__ParsedExpr{__value.__expr})
		}
		return __value.__expr
	}()
	__nextHolder := ____rune_private_b990f3d7_appendChild(__holder, __arg)
	__comma := ____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_parserSkipNewlines(__value.__state), __TokenKind_Comma)
	return ____rune_private_b990f3d7_parseArgumentListLoop(__comma.__state, __nextHolder, __endKind)
}

func ____rune_private_b990f3d7_parseIndexExpression(__state __ParserState, __receiver __ParsedExpr, __minPrec int) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBracket, "expected '[' after receiver")
	__index := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), 1)
	__close := ____rune_private_b990f3d7_parserConsume(__index.__state, __TokenKind_RBracket, "expected ']' after index")
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __close.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Index, __receiver.__text, "", "", "", []__ParsedParam{}, []__ParsedExpr{__receiver, __index.__expr}, __receiver.__line, __receiver.__column)}, __minPrec)
}

func ____rune_private_b990f3d7_parseSelectorExpression(__state __ParserState, __receiver __ParsedExpr, __minPrec int) __ExprStep {
	__operator := func() __TokenStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_DoubleColon) {
			return ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_DoubleColon, "expected '::'")
		}
		return ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Dot, "expected '.'")
	}()
	__name := ____rune_private_b990f3d7_parserConsume(__operator.__state, __TokenKind_Ident, "expected selector name")
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __name.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Selector, __name.__token.__lexeme, __name.__token.__lexeme, "", __operator.__token.__lexeme, []__ParsedParam{}, []__ParsedExpr{__receiver}, __operator.__token.__line, __operator.__token.__column)}, __minPrec)
}

func ____rune_private_b990f3d7_parsePostfixExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	__op := ____rune_private_b990f3d7_parserAdvance(__state)
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __op.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Postfix, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__expr})}, __minPrec)
}

func ____rune_private_b990f3d7_parseCompileTimePostfixExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	__op := ____rune_private_b990f3d7_parserAdvance(__state)
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __op.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_CompileTime, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__expr})}, __minPrec)
}

func ____rune_private_b990f3d7_parseWatchExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	__arrow := ____rune_private_b990f3d7_parserAdvance(__state)
	__handler := ____rune_private_b990f3d7_parseWatchHandler(____rune_private_b990f3d7_parserSkipNewlines(__arrow.__state))
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __handler.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Watch, __arrow.__token.__lexeme, __arrow.__token, []__ParsedExpr{__expr, __handler.__expr})}, __minPrec)
}

func ____rune_private_b990f3d7_parseWatchHandler(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LParen) && ____rune_private_b990f3d7_looksLikeLambda(__state) {
			return ____rune_private_b990f3d7_parseLambda(__state)
		}
		return func() __ExprStep {
			if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LBrace) {
				return ____rune_private_b990f3d7_parseBody(__state)
			}
			return ____rune_private_b990f3d7_parseExpression(__state, 1)
		}()
	}()
}

func ____rune_private_b990f3d7_parseAssignmentExpression(__state __ParserState, __target __ParsedExpr, __minPrec int) __ExprStep {
	__op := ____rune_private_b990f3d7_parserAdvance(__state)
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__op.__state), 1)
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __value.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Assign, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__target, __value.__expr})}, __minPrec)
}

func ____rune_private_b990f3d7_parsePatternPredicateExpression(__state __ParserState, __subject __ParsedExpr, __minPrec int) __ExprStep {
	__op := ____rune_private_b990f3d7_parserAdvance(__state)
	__pattern := ____rune_private_b990f3d7_parsePredicatePatternText(____rune_private_b990f3d7_parserSkipNewlines(__op.__state))
	__trueExpr := ____rune_private_b990f3d7_valueNode(__ExprKind_Bool, "true", __op.__token)
	__falseExpr := ____rune_private_b990f3d7_valueNode(__ExprKind_Bool, "false", __op.__token)
	__patternBranch := ____rune_private_b990f3d7_makeExpr(__ExprKind_Branch, __pattern.__expr.__text, "", "", "=>", []__ParsedParam{}, []__ParsedExpr{__pattern.__expr, __trueExpr}, __pattern.__expr.__line, __pattern.__expr.__column)
	__wildcardPattern := ____rune_private_b990f3d7_makeExpr(__ExprKind_Pattern, "_", "", "", "", []__ParsedParam{}, []__ParsedExpr{}, __pattern.__expr.__line, __pattern.__expr.__column)
	__wildcardBranch := ____rune_private_b990f3d7_makeExpr(__ExprKind_Branch, "_", "", "", "=>", []__ParsedParam{}, []__ParsedExpr{__wildcardPattern, __falseExpr}, __pattern.__expr.__line, __pattern.__expr.__column)
	__matchExpr := ____rune_private_b990f3d7_appendChild(____rune_private_b990f3d7_appendChild(____rune_private_b990f3d7_appendChild(____rune_private_b990f3d7_node(__ExprKind_Match, __op.__token), __subject), __patternBranch), __wildcardBranch)
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __pattern.__state, __expr: __matchExpr}, __minPrec)
}

func ____rune_private_b990f3d7_parseQuestionExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_questionIsPostfixUnwrap(__state) {
			return ____rune_private_b990f3d7_parseExpressionLoop(____rune_private_b990f3d7_parseResultUnwrapExpression(__state, __expr), __minPrec)
		}
		return ____rune_private_b990f3d7_parseTernaryExpression(__state, __expr, __minPrec)
	}()
}

func ____rune_private_b990f3d7_parseResultUnwrapExpression(__state __ParserState, __expr __ParsedExpr) __ExprStep {
	__question := ____rune_private_b990f3d7_parserAdvance(__state)
	return __ExprStep{__state: __question.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Unwrap, "?", __question.__token, []__ParsedExpr{__expr})}
}

func ____rune_private_b990f3d7_parseTernaryExpression(__state __ParserState, __condition __ParsedExpr, __minPrec int) __ExprStep {
	__question := ____rune_private_b990f3d7_parserAdvance(__state)
	__consequence := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__question.__state), 1)
	__afterConsequence := ____rune_private_b990f3d7_parserSkipNewlines(__consequence.__state)
	__colon := ____rune_private_b990f3d7_parserMatch(__afterConsequence, __TokenKind_Colon)
	return func() __ExprStep {
		if __colon.__ok {
			return ____rune_private_b990f3d7_parseTernaryAlternative(__colon.__state, __question.__token, __condition, __consequence.__expr, __minPrec)
		}
		return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __afterConsequence, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Ternary, "?:", __question.__token, []__ParsedExpr{__condition, __consequence.__expr})}, __minPrec)
	}()
}

func ____rune_private_b990f3d7_parseTernaryAlternative(__state __ParserState, __token __Token, __condition __ParsedExpr, __consequence __ParsedExpr, __minPrec int) __ExprStep {
	__alternative := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__state), 1)
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __alternative.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Ternary, "?:", __token, []__ParsedExpr{__condition, __consequence, __alternative.__expr})}, __minPrec)
}

func ____rune_private_b990f3d7_parseBinaryExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	__prec := ____rune_private_b990f3d7_precedence(____rune_private_b990f3d7_parserPeek(__state).__kind)
	return func() __ExprStep {
		if __prec < __minPrec {
			return __ExprStep{__state: __state, __expr: __expr}
		}
		return ____rune_private_b990f3d7_parseBinaryExpressionAtPrec(__state, __expr, __minPrec, __prec)
	}()
}

func ____rune_private_b990f3d7_parseBinaryExpressionAtPrec(__state __ParserState, __expr __ParsedExpr, __minPrec int, __prec int) __ExprStep {
	__op := ____rune_private_b990f3d7_parserAdvance(__state)
	__right := ____rune_private_b990f3d7_parseExpression(__op.__state, __prec+1)
	return ____rune_private_b990f3d7_parseExpressionLoop(__ExprStep{__state: __right.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Binary, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__expr, __right.__expr})}, __minPrec)
}

func ____rune_private_b990f3d7_parseUnary(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Minus) || ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Bang) || ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Tilde) {
			return ____rune_private_b990f3d7_parseUnaryOperator(__state)
		}
		return ____rune_private_b990f3d7_parsePrimary(__state)
	}()
}

func ____rune_private_b990f3d7_parseUnaryOperator(__state __ParserState) __ExprStep {
	__op := ____rune_private_b990f3d7_parserAdvance(__state)
	__right := ____rune_private_b990f3d7_parseExpression(__op.__state, 11)
	return __ExprStep{__state: __right.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Unary, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__right.__expr})}
}

func ____rune_private_b990f3d7_parsePrimary(__state __ParserState) __ExprStep {
	__token := ____rune_private_b990f3d7_parserPeek(__state)
	return func() __ExprStep {
		switch {
		case __token.__kind == __TokenKind_Int:
			return ____rune_private_b990f3d7_parseLiteral(__state, __ExprKind_Int)
		case __token.__kind == __TokenKind_Double:
			return ____rune_private_b990f3d7_parseLiteral(__state, __ExprKind_Double)
		case __token.__kind == __TokenKind_BigInt:
			return ____rune_private_b990f3d7_parseLiteral(__state, __ExprKind_BigInt)
		case __token.__kind == __TokenKind_String:
			return ____rune_private_b990f3d7_parseLiteral(__state, __ExprKind_String)
		case __token.__kind == __TokenKind_TemplateString:
			return ____rune_private_b990f3d7_parseTemplateLiteral(__state)
		case __token.__kind == __TokenKind_Char:
			return ____rune_private_b990f3d7_parseLiteral(__state, __ExprKind_Char)
		case __token.__kind == __TokenKind_Regex:
			return ____rune_private_b990f3d7_parseLiteral(__state, __ExprKind_Regex)
		case __token.__kind == __TokenKind_XMLText:
			return ____rune_private_b990f3d7_parseLiteral(__state, __ExprKind_XMLText)
		case __token.__kind == __TokenKind_Ident:
			return ____rune_private_b990f3d7_parseIdentifierPrimary(__state)
		case __token.__kind == __TokenKind_At:
			return ____rune_private_b990f3d7_parseAtExpression(__state)
		case __token.__kind == __TokenKind_Dot:
			return ____rune_private_b990f3d7_parseThisSelector(__state)
		case __token.__kind == __TokenKind_LBracket:
			return ____rune_private_b990f3d7_parseArrayLiteral(__state)
		case __token.__kind == __TokenKind_Dollar:
			return ____rune_private_b990f3d7_parseDollarExpression(__state)
		case __token.__kind == __TokenKind_LBrace:
			return ____rune_private_b990f3d7_parseBraceLiteral(__state)
		case __token.__kind == __TokenKind_LParen:
			return ____rune_private_b990f3d7_parseParenOrTuple(__state)
		default:
			return ____rune_private_b990f3d7_parsePrimaryError(__state)
		}
	}()
}

func ____rune_private_b990f3d7_parseTemplateLiteral(__state __ParserState) __ExprStep {
	__step := ____rune_private_b990f3d7_parserAdvance(__state)
	__parsed := ____rune_private_b990f3d7_parseTemplateParts(____rune_private_b990f3d7_templateInner(__step.__token.__lexeme), 0, 0, "", []__ParsedExpr{})
	return __ExprStep{__state: __step.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Template, __step.__token.__lexeme, "", "`"+__parsed.__text+"`", "", []__ParsedParam{}, __parsed.__children, __step.__token.__line, __step.__token.__column)}
}

func ____rune_private_b990f3d7_templateInner(__raw string) string {
	return func() string {
		if len([]rune(__raw)) >= 2 {
			return func() string { runes := []rune(__raw); return string(runes[1 : len([]rune(__raw))-1]) }()
		}
		return __raw
	}()
}

func ____rune_private_b990f3d7_parseTemplateParts(__inner string, __index int, __textStart int, __out string, __children []__ParsedExpr) __TemplateParse {
	return func() __TemplateParse {
		if __index >= len([]rune(__inner)) {
			return __TemplateParse{__text: __out + func() string { runes := []rune(__inner); return string(runes[__textStart:len([]rune(__inner))]) }(), __children: __children}
		}
		return ____rune_private_b990f3d7_parseTemplatePartAt(__inner, __index, __textStart, __out, __children)
	}()
}

func ____rune_private_b990f3d7_parseTemplatePartAt(__inner string, __index int, __textStart int, __out string, __children []__ParsedExpr) __TemplateParse {
	__ch := []rune(__inner)[__index]
	return func() __TemplateParse {
		if __ch == '\\' && __index+1 < len([]rune(__inner)) && []rune(__inner)[__index+1] == '(' {
			return ____rune_private_b990f3d7_parseTemplateExprPart(__inner, __index, __textStart, __out, __children)
		}
		return ____rune_private_b990f3d7_parseTemplateParts(__inner, __index+1, __textStart, __out, __children)
	}()
}

func ____rune_private_b990f3d7_parseTemplateExprPart(__inner string, __index int, __textStart int, __out string, __children []__ParsedExpr) __TemplateParse {
	__exprStart := __index + 2
	__exprEnd := ____rune_private_b990f3d7_scanTemplateExprEnd(__inner, __exprStart, 1)
	return func() __TemplateParse {
		if __exprEnd < 0 {
			return __TemplateParse{__text: __out + func() string { runes := []rune(__inner); return string(runes[__textStart:len([]rune(__inner))]) }(), __children: __children}
		}
		return ____rune_private_b990f3d7_parseTemplateParts(__inner, __exprEnd+1, __exprEnd+1, __out+func() string { runes := []rune(__inner); return string(runes[__textStart:__index]) }()+"<<<RUNE_TEMPLATE_PART>>>", ____rune_private_b990f3d7_pushTemplateExpr(__children, func() string { runes := []rune(__inner); return string(runes[__exprStart:__exprEnd]) }()))
	}()
}

func ____rune_private_b990f3d7_scanTemplateExprEnd(__inner string, __index int, __depth int) int {
	return func() int {
		if __index >= len([]rune(__inner)) {
			return -1
		}
		return ____rune_private_b990f3d7_scanTemplateExprEndAt(__inner, __index, __depth)
	}()
}

func ____rune_private_b990f3d7_scanTemplateExprEndAt(__inner string, __index int, __depth int) int {
	__ch := []rune(__inner)[__index]
	return func() int {
		switch {
		case __ch == '"':
			return ____rune_private_b990f3d7_scanTemplateExprEnd(__inner, ____rune_private_b990f3d7_skipTemplateQuoted(__inner, __index+1, '"'), __depth)
		case __ch == '\'':
			return ____rune_private_b990f3d7_scanTemplateExprEnd(__inner, ____rune_private_b990f3d7_skipTemplateQuoted(__inner, __index+1, '\''), __depth)
		case __ch == '`':
			return ____rune_private_b990f3d7_scanTemplateExprEnd(__inner, ____rune_private_b990f3d7_skipTemplateQuoted(__inner, __index+1, '`'), __depth)
		case __ch == '(':
			return ____rune_private_b990f3d7_scanTemplateExprEnd(__inner, __index+1, __depth+1)
		case __ch == ')':
			return func() int {
				if __depth == 1 {
					return __index
				}
				return ____rune_private_b990f3d7_scanTemplateExprEnd(__inner, __index+1, __depth-1)
			}()
		default:
			return ____rune_private_b990f3d7_scanTemplateExprEnd(__inner, __index+1, __depth)
		}
	}()
}

func ____rune_private_b990f3d7_skipTemplateQuoted(__text string, __index int, __quote rune) int {
	return func() int {
		if __index >= len([]rune(__text)) {
			return __index
		}
		return ____rune_private_b990f3d7_skipTemplateQuotedAt(__text, __index, __quote)
	}()
}

func ____rune_private_b990f3d7_skipTemplateQuotedAt(__text string, __index int, __quote rune) int {
	__ch := []rune(__text)[__index]
	return func() int {
		if __ch == '\\' {
			return ____rune_private_b990f3d7_skipTemplateQuoted(__text, __index+2, __quote)
		}
		return func() int {
			if __ch == __quote {
				return __index + 1
			}
			return ____rune_private_b990f3d7_skipTemplateQuoted(__text, __index+1, __quote)
		}()
	}()
}

func ____rune_private_b990f3d7_pushTemplateExpr(__children []__ParsedExpr, __source string) []__ParsedExpr {
	__parsed := ____rune_private_b990f3d7_parseTemplateExpression(strings.TrimSpace(__source))
	return func() []__ParsedExpr {
		out := []__ParsedExpr{}
		out = append(out, __children...)
		out = append(out, __parsed)
		return out
	}()
}

func ____rune_private_b990f3d7_parseTemplateExpression(__source string) __ParsedExpr {
	return func() __ParsedExpr {
		if __source == "" {
			return ____rune_private_b990f3d7_emptyExpr()
		}
		return ____rune_private_b990f3d7_parseExpression(__ParserState{__tokens: __lex(__source), __current: 0, __errors: []__ParseError{}}, 1).__expr
	}()
}

func ____rune_private_b990f3d7_parseLiteral(__state __ParserState, __kind __ExprKind) __ExprStep {
	__step := ____rune_private_b990f3d7_parserAdvance(__state)
	return __ExprStep{__state: __step.__state, __expr: ____rune_private_b990f3d7_valueNode(__kind, __step.__token.__lexeme, __step.__token)}
}

func ____rune_private_b990f3d7_parseIdentifierPrimary(__state __ParserState) __ExprStep {
	__step := ____rune_private_b990f3d7_parserAdvance(__state)
	__kind := func() __ExprKind {
		if __step.__token.__lexeme == "true" || __step.__token.__lexeme == "false" {
			return __ExprKind_Bool
		}
		return func() __ExprKind {
			if __step.__token.__lexeme == "null" {
				return __ExprKind_Null
			}
			return __ExprKind_Identifier
		}()
	}()
	return __ExprStep{__state: __step.__state, __expr: func() __ParsedExpr {
		if __kind == __ExprKind_Identifier {
			return ____rune_private_b990f3d7_namedNode(__kind, __step.__token.__lexeme, __step.__token)
		}
		return ____rune_private_b990f3d7_valueNode(__kind, __step.__token.__lexeme, __step.__token)
	}()}
}

func ____rune_private_b990f3d7_parseAtExpression(__state __ParserState) __ExprStep {
	__at := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_At, "expected '@'")
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__at.__state, __TokenKind_String) {
			return ____rune_private_b990f3d7_parseAtImportExpression(__at)
		}
		return ____rune_private_b990f3d7_parseAtModuleExpression(__at)
	}()
}

func ____rune_private_b990f3d7_parseAtImportExpression(__at __TokenStep) __ExprStep {
	__path := ____rune_private_b990f3d7_parserAdvance(__at.__state)
	return __ExprStep{__state: __path.__state, __expr: ____rune_private_b990f3d7_valueNode(__ExprKind_At, __path.__token.__lexeme, __at.__token)}
}

func ____rune_private_b990f3d7_parseAtModuleExpression(__at __TokenStep) __ExprStep {
	__name := ____rune_private_b990f3d7_parserConsume(__at.__state, __TokenKind_Ident, "expected module name after '@'")
	return __ExprStep{__state: __name.__state, __expr: ____rune_private_b990f3d7_namedNode(__ExprKind_At, __name.__token.__lexeme, __at.__token)}
}

func ____rune_private_b990f3d7_parseThisSelector(__state __ParserState) __ExprStep {
	__dot := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Dot, "expected '.'")
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__dot.__state, __TokenKind_Ident) {
			return ____rune_private_b990f3d7_parseThisFieldSelector(__dot.__state, __dot.__token)
		}
		return __ExprStep{__state: __dot.__state, __expr: ____rune_private_b990f3d7_node(__ExprKind_This, __dot.__token)}
	}()
}

func ____rune_private_b990f3d7_parseThisFieldSelector(__state __ParserState, __dot __Token) __ExprStep {
	__name := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected field name after '.'")
	return __ExprStep{__state: __name.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Selector, __name.__token.__lexeme, __name.__token.__lexeme, "", ".", []__ParsedParam{}, []__ParsedExpr{____rune_private_b990f3d7_node(__ExprKind_This, __dot)}, __dot.__line, __dot.__column)}
}

func ____rune_private_b990f3d7_parseArrayLiteral(__state __ParserState) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBracket, "expected '['")
	__args := ____rune_private_b990f3d7_parseArgumentList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), []__ParsedExpr{}, __TokenKind_RBracket)
	__close := ____rune_private_b990f3d7_parserConsume(__args.__state, __TokenKind_RBracket, "expected ']' after array literal")
	return __ExprStep{__state: __close.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Array, "[]", "", "", "", []__ParsedParam{}, __args.__expr.__children, __open.__token.__line, __open.__token.__column)}
}

func ____rune_private_b990f3d7_parseReactiveLiteral(__state __ParserState) __ExprStep {
	__start := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Dollar, "expected '$'")
	return ____rune_private_b990f3d7_parseReactiveLiteralAfterDollar(__start)
}

func ____rune_private_b990f3d7_parseDollarExpression(__state __ParserState) __ExprStep {
	__start := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Dollar, "expected '$'")
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__start.__state, __TokenKind_Ident) {
			return ____rune_private_b990f3d7_parseSignalPrefixIdentifier(__start)
		}
		return __ExprStep{__state: ____rune_private_b990f3d7_parserErrorAt(__start.__state, ____rune_private_b990f3d7_parserPeek(__start.__state), "expected signal name after '$'"), __expr: ____rune_private_b990f3d7_emptyExpr()}
	}()
}

func ____rune_private_b990f3d7_parseSignalPrefixIdentifier(__start __TokenStep) __ExprStep {
	__name := ____rune_private_b990f3d7_parserAdvance(__start.__state)
	return __ExprStep{__state: __name.__state, __expr: ____rune_private_b990f3d7_namedNode(__ExprKind_Identifier, __name.__token.__lexeme, __name.__token)}
}

func ____rune_private_b990f3d7_parseReactiveLiteralAfterDollar(__start __TokenStep) __ExprStep {
	__value := func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__start.__state, __TokenKind_LBracket) {
			return ____rune_private_b990f3d7_parseArrayLiteral(__start.__state)
		}
		return func() __ExprStep {
			if ____rune_private_b990f3d7_parserCheck(__start.__state, __TokenKind_LBrace) {
				return ____rune_private_b990f3d7_parseBraceLiteral(__start.__state)
			}
			return __ExprStep{__state: ____rune_private_b990f3d7_parserErrorAt(__start.__state, ____rune_private_b990f3d7_parserPeek(__start.__state), "expected '[' or '{' after '$'"), __expr: ____rune_private_b990f3d7_emptyExpr()}
		}()
	}()
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_b990f3d7_opNode(__ExprKind_Reactive, "$", __start.__token, []__ParsedExpr{__value.__expr})}
}

func ____rune_private_b990f3d7_parseBraceLiteral(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_looksLikeMapLiteralBody(__state) {
			return ____rune_private_b990f3d7_parseMapLiteral(__state)
		}
		return ____rune_private_b990f3d7_parseObjectLiteral(__state)
	}()
}

func ____rune_private_b990f3d7_parseMapLiteral(__state __ParserState) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBrace, "expected '{'")
	return ____rune_private_b990f3d7_parseMapLiteralLoop(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), ____rune_private_b990f3d7_node(__ExprKind_Map, __open.__token))
}

func ____rune_private_b990f3d7_parseMapLiteralLoop(__state __ParserState, __mapExpr __ParsedExpr) __ExprStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_b990f3d7_finishMapLiteral(__current, __mapExpr)
		}
		return ____rune_private_b990f3d7_parseMapLiteralEntry(__current, __mapExpr)
	}()
}

func ____rune_private_b990f3d7_parseMapLiteralEntry(__state __ParserState, __mapExpr __ParsedExpr) __ExprStep {
	__key := ____rune_private_b990f3d7_parseExpression(__state, 1)
	__colon := ____rune_private_b990f3d7_parserConsume(__key.__state, __TokenKind_Colon, "expected ':' after map key")
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__colon.__state), 1)
	__entry := ____rune_private_b990f3d7_makeExpr(__ExprKind_Entry, "", "", "", ":", []__ParsedParam{}, []__ParsedExpr{__key.__expr, __value.__expr}, __key.__expr.__line, __key.__expr.__column)
	__nextMap := ____rune_private_b990f3d7_appendChild(__mapExpr, __entry)
	__comma := ____rune_private_b990f3d7_parserMatch(____rune_private_b990f3d7_parserSkipNewlines(____rune_private_b990f3d7_consumeStatementEnd(__value.__state)), __TokenKind_Comma)
	return ____rune_private_b990f3d7_parseMapLiteralLoop(__comma.__state, __nextMap)
}

func ____rune_private_b990f3d7_finishMapLiteral(__state __ParserState, __mapExpr __ParsedExpr) __ExprStep {
	__close := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_RBrace, "expected '}' after map literal")
	return __ExprStep{__state: __close.__state, __expr: __mapExpr}
}

func ____rune_private_b990f3d7_parseObjectLiteral(__state __ParserState) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBrace, "expected '{'")
	return ____rune_private_b990f3d7_parseObjectLiteralLoop(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), ____rune_private_b990f3d7_node(__ExprKind_Object, __open.__token))
}

func ____rune_private_b990f3d7_parseObjectLiteralLoop(__state __ParserState, __object __ParsedExpr) __ExprStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_b990f3d7_finishObjectLiteral(__current, __object)
		}
		return ____rune_private_b990f3d7_parseObjectLiteralMember(__current, __object)
	}()
}

func ____rune_private_b990f3d7_parseObjectLiteralMember(__state __ParserState, __object __ParsedExpr) __ExprStep {
	__privateStep := ____rune_private_b990f3d7_parseObjectPrivateModifier(__state)
	__memberState := __privateStep.__state
	__member := func() __ExprStep {
		if ____rune_private_b990f3d7_looksLikeFunctionDecl(__memberState) {
			return ____rune_private_b990f3d7_parseObjectMethod(__memberState, __privateStep.__ok)
		}
		return ____rune_private_b990f3d7_parseObjectField(__memberState, __privateStep.__ok)
	}()
	__nextObject := ____rune_private_b990f3d7_appendChild(__object, __member.__expr)
	return ____rune_private_b990f3d7_parseObjectLiteralLoop(____rune_private_b990f3d7_consumeFieldSeparator(__member.__state, __TokenKind_RBrace, "expected ',' between object literal fields"), __nextObject)
}

func ____rune_private_b990f3d7_parseObjectMethod(__state __ParserState, __private bool) __ExprStep {
	__fn := ____rune_private_b990f3d7_parseFunctionWithReceiver(__state, "", __private, false, false, ____rune_private_b990f3d7_emptyAnnotations())
	return __ExprStep{__state: __fn.__state, __expr: ____rune_private_b990f3d7_functionToExpr(__fn.__function)}
}

func ____rune_private_b990f3d7_parseObjectField(__state __ParserState, __private bool) __ExprStep {
	__name := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected field name")
	__colon := ____rune_private_b990f3d7_parserConsume(__name.__state, __TokenKind_Colon, "expected ':' after field name")
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__colon.__state), 1)
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_b990f3d7_makeExpr(func() __ExprKind {
		if __private {
			return __ExprKind_PrivateField
		}
		return __ExprKind_Field
	}(), __name.__token.__lexeme, __name.__token.__lexeme, "", ":", []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)}
}

func ____rune_private_b990f3d7_finishObjectLiteral(__state __ParserState, __object __ParsedExpr) __ExprStep {
	__close := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_RBrace, "expected '}' after object literal")
	return __ExprStep{__state: __close.__state, __expr: __object}
}

func ____rune_private_b990f3d7_parseStructLiteral(__state __ParserState, __typeExpr __ParsedExpr) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBrace, "expected '{' after type name")
	return ____rune_private_b990f3d7_parseStructLiteralLoop(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), ____rune_private_b990f3d7_namedNode(__ExprKind_Struct, __typeExpr.__name, __open.__token))
}

func ____rune_private_b990f3d7_parseStructLiteralLoop(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_b990f3d7_finishStructLiteral(__current, __structExpr)
		}
		return ____rune_private_b990f3d7_parseStructLiteralField(__current, __structExpr)
	}()
}

func ____rune_private_b990f3d7_parseStructLiteralField(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_DotDot) {
			return ____rune_private_b990f3d7_parseStructLiteralSpreadField(__state, __structExpr)
		}
		return ____rune_private_b990f3d7_parseStructLiteralNamedField(__state, __structExpr)
	}()
}

func ____rune_private_b990f3d7_parseStructLiteralNamedField(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	__name := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Ident, "expected field name")
	__colon := ____rune_private_b990f3d7_parserConsume(__name.__state, __TokenKind_Colon, "expected ':' after field name")
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__colon.__state), 1)
	__field := ____rune_private_b990f3d7_makeExpr(__ExprKind_Field, __name.__token.__lexeme, __name.__token.__lexeme, "", ":", []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)
	__nextStruct := ____rune_private_b990f3d7_appendChild(__structExpr, __field)
	return ____rune_private_b990f3d7_parseStructLiteralLoop(____rune_private_b990f3d7_consumeFieldSeparator(__value.__state, __TokenKind_RBrace, "expected ',' between struct literal fields"), __nextStruct)
}

func ____rune_private_b990f3d7_parseStructLiteralSpreadField(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	__spread := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_DotDot, "expected '..'")
	__value := ____rune_private_b990f3d7_parseExpression(__spread.__state, 1)
	__field := ____rune_private_b990f3d7_opNode(__ExprKind_Spread, "..", __spread.__token, []__ParsedExpr{__value.__expr})
	__nextStruct := ____rune_private_b990f3d7_appendChild(__structExpr, __field)
	return ____rune_private_b990f3d7_parseStructLiteralLoop(____rune_private_b990f3d7_consumeFieldSeparator(__value.__state, __TokenKind_RBrace, "expected ',' between struct literal fields"), __nextStruct)
}

func ____rune_private_b990f3d7_finishStructLiteral(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	__close := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_RBrace, "expected '}' after struct literal")
	return __ExprStep{__state: __close.__state, __expr: __structExpr}
}

func ____rune_private_b990f3d7_parseParenOrTuple(__state __ParserState) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LParen, "expected '('")
	__afterOpen := ____rune_private_b990f3d7_parserSkipNewlines(__open.__state)
	return ____rune_private_b990f3d7_parseParenOrTupleAfterOpen(__afterOpen, __open.__token)
}

func ____rune_private_b990f3d7_parseParenOrTupleAfterOpen(__state __ParserState, __open __Token) __ExprStep {
	__expr := ____rune_private_b990f3d7_parseExpression(__state, 1)
	__afterExpr := ____rune_private_b990f3d7_parserSkipNewlines(__expr.__state)
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__afterExpr, __TokenKind_Comma) {
			return ____rune_private_b990f3d7_parseTupleAfterFirst(__afterExpr, __open, __expr.__expr)
		}
		return ____rune_private_b990f3d7_finishParenExpression(__afterExpr, __expr.__expr)
	}()
}

func ____rune_private_b990f3d7_finishParenExpression(__state __ParserState, __expr __ParsedExpr) __ExprStep {
	__close := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_RParen, "expected ')' after expression")
	return __ExprStep{__state: __close.__state, __expr: __expr}
}

func ____rune_private_b990f3d7_parseTupleAfterFirst(__state __ParserState, __open __Token, __first __ParsedExpr) __ExprStep {
	__comma := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_Comma, "expected ','")
	__holder := ____rune_private_b990f3d7_appendChild(____rune_private_b990f3d7_node(__ExprKind_Tuple, __open), __first)
	__values := ____rune_private_b990f3d7_parseArgumentListLoop(____rune_private_b990f3d7_parserSkipNewlines(__comma.__state), __holder, __TokenKind_RParen)
	__close := ____rune_private_b990f3d7_parserConsume(__values.__state, __TokenKind_RParen, "expected ')' after tuple literal")
	return __ExprStep{__state: __close.__state, __expr: __values.__expr}
}

func ____rune_private_b990f3d7_parsePrimaryError(__state __ParserState) __ExprStep {
	__token := ____rune_private_b990f3d7_parserPeek(__state)
	__step := ____rune_private_b990f3d7_parserAdvance(____rune_private_b990f3d7_parserErrorAt(__state, __token, "expected expression, got "+__tokenKindName(__token.__kind)))
	return __ExprStep{__state: __step.__state, __expr: ____rune_private_b990f3d7_namedNode(__ExprKind_Error, "<error>", __token)}
}

func ____rune_private_b990f3d7_parseLambda(__state __ParserState) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LParen, "expected '(' before lambda parameters")
	__params := ____rune_private_b990f3d7_parseParamList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state))
	__close := ____rune_private_b990f3d7_parserConsume(__params.__state, __TokenKind_RParen, "expected ')' after lambda parameters")
	__afterParams := ____rune_private_b990f3d7_parserSkipNewlines(__close.__state)
	__ret := ____rune_private_b990f3d7_parserMatch(__afterParams, __TokenKind_Arrow)
	__returnType := func() __TypeRefStep {
		if __ret.__ok {
			return ____rune_private_b990f3d7_parseTypeRef(____rune_private_b990f3d7_parserSkipNewlines(__ret.__state))
		}
		return __TypeRefStep{__state: __afterParams, __typeRef: __emptyParsedTypeRef()}
	}()
	__arrow := ____rune_private_b990f3d7_parserConsume(____rune_private_b990f3d7_parserSkipNewlines(__returnType.__state), __TokenKind_FatArrow, "expected '=>' after lambda parameter")
	__body := ____rune_private_b990f3d7_parseBody(____rune_private_b990f3d7_parserSkipNewlines(__arrow.__state))
	return __ExprStep{__state: __body.__state, __expr: ____rune_private_b990f3d7_withChildren(____rune_private_b990f3d7_withParams(____rune_private_b990f3d7_withText(____rune_private_b990f3d7_node(__ExprKind_Lambda, __open.__token), __typeRefToString(__returnType.__typeRef)), __params.__params), []__ParsedExpr{__body.__expr})}
}

func ____rune_private_b990f3d7_parseMatchExpression(__state __ParserState, __subject __ParsedExpr) __ExprStep {
	__open := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_LBrace, "expected '{' after match subject")
	__block := ____rune_private_b990f3d7_appendChild(____rune_private_b990f3d7_node(__ExprKind_Match, __open.__token), __subject)
	return ____rune_private_b990f3d7_parsePatternBlock(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), __block)
}

func ____rune_private_b990f3d7_parsePatternBlock(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_b990f3d7_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_b990f3d7_finishPatternBlock(__current, __block)
		}
		return ____rune_private_b990f3d7_parsePatternBranch(__current, __block)
	}()
}

func ____rune_private_b990f3d7_parsePatternBranch(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__pattern := ____rune_private_b990f3d7_parsePatternText(__state)
	__arrow := ____rune_private_b990f3d7_parserConsume(__pattern.__state, __TokenKind_FatArrow, "expected '=>' after pattern")
	__value := ____rune_private_b990f3d7_parseExpression(____rune_private_b990f3d7_parserSkipNewlines(__arrow.__state), 1)
	__branch := ____rune_private_b990f3d7_makeExpr(__ExprKind_Branch, __pattern.__expr.__text, "", "", "=>", []__ParsedParam{}, []__ParsedExpr{__pattern.__expr, __value.__expr}, __pattern.__expr.__line, __pattern.__expr.__column)
	__nextBlock := ____rune_private_b990f3d7_appendChild(__block, __branch)
	return ____rune_private_b990f3d7_parsePatternBlock(____rune_private_b990f3d7_consumeStatementEnd(__value.__state), __nextBlock)
}

func ____rune_private_b990f3d7_finishPatternBlock(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__close := ____rune_private_b990f3d7_parserConsume(__state, __TokenKind_RBrace, "expected '}' after pattern block")
	return __ExprStep{__state: __close.__state, __expr: __block}
}

func ____rune_private_b990f3d7_parsePatternText(__state __ParserState) __ExprStep {
	return ____rune_private_b990f3d7_parsePatternTextLoop(__state, "", ____rune_private_b990f3d7_parserPeek(__state), 0, false)
}

func ____rune_private_b990f3d7_parsePredicatePatternText(__state __ParserState) __ExprStep {
	return ____rune_private_b990f3d7_parsePredicatePatternTextLoop(__state, "", ____rune_private_b990f3d7_parserPeek(__state), 0, false)
}

func ____rune_private_b990f3d7_parsePredicatePatternTextLoop(__state __ParserState, __text string, __start __Token, __depth int, __consumed bool) __ExprStep {
	return func() __ExprStep {
		if __depth == 0 && ____rune_private_b990f3d7_predicatePatternEnd(____rune_private_b990f3d7_parserPeek(__state).__kind) {
			return __ExprStep{__state: func() __ParserState {
				if __consumed {
					return __state
				}
				return ____rune_private_b990f3d7_parserErrorAt(__state, ____rune_private_b990f3d7_parserPeek(__state), "expected pattern")
			}(), __expr: ____rune_private_b990f3d7_withText(____rune_private_b990f3d7_node(__ExprKind_Pattern, __start), __text)}
		}
		return ____rune_private_b990f3d7_parsePredicatePatternTextToken(__state, __text, __start, __depth)
	}()
}

func ____rune_private_b990f3d7_parsePredicatePatternTextToken(__state __ParserState, __text string, __start __Token, __depth int) __ExprStep {
	__token := ____rune_private_b990f3d7_parserPeek(__state)
	__step := ____rune_private_b990f3d7_parserAdvance(__state)
	return ____rune_private_b990f3d7_parsePredicatePatternTextLoop(__step.__state, __text+__token.__lexeme, __start, ____rune_private_b990f3d7_patternNextDepth(__depth, __token.__kind), true)
}

func ____rune_private_b990f3d7_predicatePatternEnd(__kind __TokenKind) bool {
	return __kind == __TokenKind_EOF || __kind == __TokenKind_Newline || __kind == __TokenKind_Comma || __kind == __TokenKind_RParen || __kind == __TokenKind_RBracket || __kind == __TokenKind_RBrace
}

func ____rune_private_b990f3d7_parsePatternTextLoop(__state __ParserState, __text string, __start __Token, __depth int, __consumed bool) __ExprStep {
	return func() __ExprStep {
		if __depth == 0 && (____rune_private_b990f3d7_parserCheck(__state, __TokenKind_FatArrow) || ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_EOF)) {
			return __ExprStep{__state: func() __ParserState {
				if __consumed {
					return __state
				}
				return ____rune_private_b990f3d7_parserErrorAt(__state, ____rune_private_b990f3d7_parserPeek(__state), "expected pattern")
			}(), __expr: ____rune_private_b990f3d7_withText(____rune_private_b990f3d7_node(__ExprKind_Pattern, __start), __text)}
		}
		return ____rune_private_b990f3d7_parsePatternTextToken(__state, __text, __start, __depth)
	}()
}

func ____rune_private_b990f3d7_parsePatternTextToken(__state __ParserState, __text string, __start __Token, __depth int) __ExprStep {
	__token := ____rune_private_b990f3d7_parserPeek(__state)
	__step := ____rune_private_b990f3d7_parserAdvance(__state)
	return ____rune_private_b990f3d7_parsePatternTextLoop(__step.__state, __text+__token.__lexeme, __start, ____rune_private_b990f3d7_patternNextDepth(__depth, __token.__kind), true)
}

func ____rune_private_b990f3d7_patternNextDepth(__depth int, __kind __TokenKind) int {
	return func() int {
		if __kind == __TokenKind_LParen || __kind == __TokenKind_LBracket || __kind == __TokenKind_LBrace {
			return __depth + 1
		}
		return func() int {
			if (__kind == __TokenKind_RParen || __kind == __TokenKind_RBracket || __kind == __TokenKind_RBrace) && __depth > 0 {
				return __depth - 1
			}
			return __depth
		}()
	}()
}

func ____rune_private_b990f3d7_functionToExpr(__fn __ParsedFunction) __ParsedExpr {
	return ____rune_private_b990f3d7_makeExpr(func() __ExprKind {
		if __fn.__private {
			return __ExprKind_PrivateMethod
		}
		return __ExprKind_Method
	}(), __typeRefToString(__fn.__returnType), __fn.__name, "", "=>", __fn.__params, []__ParsedExpr{__fn.__body}, __fn.__line, __fn.__column)
}

func ____rune_private_b990f3d7_calleeText(__expr __ParsedExpr) string {
	return func() string {
		if __expr.__name != "" {
			return __expr.__name
		}
		return func() string {
			if __expr.__value != "" {
				return __expr.__value
			}
			return __expr.__text
		}()
	}()
}

func ____rune_private_b990f3d7_precedence(__kind __TokenKind) int {
	return func() int {
		if __kind == __TokenKind_QuestionQuestion || __kind == __TokenKind_OrOr {
			return 1
		}
		return func() int {
			if __kind == __TokenKind_AndAnd {
				return 2
			}
			return func() int {
				if __kind == __TokenKind_BitOr {
					return 3
				}
				return func() int {
					if __kind == __TokenKind_BitXor {
						return 4
					}
					return func() int {
						if __kind == __TokenKind_BitAnd {
							return 5
						}
						return func() int {
							if __kind == __TokenKind_EqualEqual || __kind == __TokenKind_BangEqual {
								return 6
							}
							return func() int {
								if __kind == __TokenKind_DotDotEqual || __kind == __TokenKind_Less || __kind == __TokenKind_LessEqual || __kind == __TokenKind_Greater || __kind == __TokenKind_GreaterEqual {
									return 7
								}
								return func() int {
									if __kind == __TokenKind_ShiftLeft || __kind == __TokenKind_ShiftRight || __kind == __TokenKind_UnsignedShiftRight {
										return 8
									}
									return func() int {
										if __kind == __TokenKind_Plus || __kind == __TokenKind_Minus {
											return 9
										}
										return func() int {
											if __kind == __TokenKind_Star || __kind == __TokenKind_Slash || __kind == __TokenKind_Percent {
												return 10
											}
											return 0
										}()
									}()
								}()
							}()
						}()
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_b990f3d7_parseAnnotations(__state __ParserState) __AnnotationListStep {
	return ____rune_private_b990f3d7_parseAnnotationsLoop(__state, ____rune_private_b990f3d7_emptyAnnotations())
}

func ____rune_private_b990f3d7_parseAnnotationsLoop(__state __ParserState, __annotations []__ParsedAnnotation) __AnnotationListStep {
	__current := ____rune_private_b990f3d7_parserSkipNewlines(__state)
	return func() __AnnotationListStep {
		if ____rune_private_b990f3d7_looksLikeAnnotation(__current) {
			return ____rune_private_b990f3d7_parseAnnotationsNext(__current, __annotations)
		}
		return __AnnotationListStep{__state: __current, __annotations: __annotations}
	}()
}

func ____rune_private_b990f3d7_parseAnnotationsNext(__state __ParserState, __annotations []__ParsedAnnotation) __AnnotationListStep {
	__step := ____rune_private_b990f3d7_parseAnnotation(__state)
	return ____rune_private_b990f3d7_parseAnnotationsLoop(__step.__state, func() []__ParsedAnnotation {
		out := []__ParsedAnnotation{}
		out = append(out, __annotations...)
		out = append(out, __step.__annotation)
		return out
	}())
}

func ____rune_private_b990f3d7_looksLikeAnnotation(__state __ParserState) bool {
	return ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Hash) || ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_At) && ____rune_private_b990f3d7_parserCheckNext(__state, __TokenKind_Ident)
}

func ____rune_private_b990f3d7_parseAnnotation(__state __ParserState) __AnnotationStep {
	__marker := ____rune_private_b990f3d7_parserAdvance(__state)
	__first := ____rune_private_b990f3d7_parserConsume(__marker.__state, __TokenKind_Ident, "expected annotation name")
	__dot := ____rune_private_b990f3d7_parserMatch(__first.__state, __TokenKind_Dot)
	__name := func() __TokenStep {
		if __dot.__ok {
			return ____rune_private_b990f3d7_parserConsume(__dot.__state, __TokenKind_Ident, "expected annotation function name after '.'")
		}
		return __first
	}()
	__open := ____rune_private_b990f3d7_parserMatch(__name.__state, __TokenKind_LParen)
	__args := func() __ExprStep {
		if __open.__ok {
			return ____rune_private_b990f3d7_parseArgumentList(____rune_private_b990f3d7_parserSkipNewlines(__open.__state), append([]__ParsedExpr{}, []__ParsedExpr{____rune_private_b990f3d7_emptyExpr()}[0:0]...), __TokenKind_RParen)
		}
		return __ExprStep{__state: __name.__state, __expr: ____rune_private_b990f3d7_makeExpr(__ExprKind_Args, "", "", "", "", []__ParsedParam{}, []__ParsedExpr{}, 0, 0)}
	}()
	__close := func() __TokenStep {
		if __open.__ok {
			return ____rune_private_b990f3d7_parserConsume(__args.__state, __TokenKind_RParen, "expected ')' after annotation arguments")
		}
		return __TokenStep{__state: __args.__state, __token: __name.__token}
	}()
	return __AnnotationStep{__state: __close.__state, __annotation: __ParsedAnnotation{__marker: __marker.__token.__lexeme, __module: func() string {
		if __dot.__ok {
			return __first.__token.__lexeme
		}
		return ""
	}(), __name: __name.__token.__lexeme, __args: __args.__expr.__children, __line: __marker.__token.__line, __column: __marker.__token.__column}}
}

func ____rune_private_b990f3d7_skipBalanced(__state __ParserState, __openKind __TokenKind, __closeKind __TokenKind, __depth int) __ParserState {
	return func() __ParserState {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_EOF) || __depth <= 0 {
			return __state
		}
		return ____rune_private_b990f3d7_skipBalancedStep(__state, __openKind, __closeKind, __depth)
	}()
}

func ____rune_private_b990f3d7_skipBalancedStep(__state __ParserState, __openKind __TokenKind, __closeKind __TokenKind, __depth int) __ParserState {
	__token := ____rune_private_b990f3d7_parserPeek(__state)
	__step := ____rune_private_b990f3d7_parserAdvance(__state)
	__nextDepth := func() int {
		if __token.__kind == __openKind {
			return __depth + 1
		}
		return func() int {
			if __token.__kind == __closeKind {
				return __depth - 1
			}
			return __depth
		}()
	}()
	return ____rune_private_b990f3d7_skipBalanced(__step.__state, __openKind, __closeKind, __nextDepth)
}

func ____rune_private_b990f3d7_questionIsPostfixUnwrap(__state __ParserState) bool {
	__next := ____rune_private_b990f3d7_parserKindAt(__state, __state.__current+1)
	return __next == __TokenKind_EOF || __next == __TokenKind_Newline || __next == __TokenKind_RParen || __next == __TokenKind_RBracket || __next == __TokenKind_RBrace || __next == __TokenKind_Comma
}

func ____rune_private_b990f3d7_looksLikeTypeDecl(__state __ParserState) bool {
	return ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Ident) && ____rune_private_b990f3d7_parserKindAt(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, ____rune_private_b990f3d7_skipGenericNamesAt(__state, __state.__current+1))) == __TokenKind_Colon
}

func ____rune_private_b990f3d7_looksLikeFunctionDecl(__state __ParserState) bool {
	__start := func() int {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Tilde) {
			return ____rune_private_b990f3d7_skipNewlinesAt(__state, __state.__current+1)
		}
		return __state.__current
	}()
	return ____rune_private_b990f3d7_parserKindAt(__state, __start) == __TokenKind_Ident && ____rune_private_b990f3d7_looksLikeFunctionAfterName(__state, ____rune_private_b990f3d7_skipGenericNamesAt(__state, __start+1))
}

func ____rune_private_b990f3d7_looksLikeMacroFunctionDecl(__state __ParserState) bool {
	return ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Hash) && ____rune_private_b990f3d7_looksLikeFunctionDecl(____rune_private_b990f3d7_stateAt(__state, __state.__current+1))
}

func ____rune_private_b990f3d7_looksLikeStaticFunctionDecl(__state __ParserState) bool {
	__markerState := func() __ParserState {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_DoubleColon) {
			return ____rune_private_b990f3d7_stateAt(__state, __state.__current+1)
		}
		return func() __ParserState {
			if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Ident) && ____rune_private_b990f3d7_parserPeek(__state).__lexeme == "static" {
				return ____rune_private_b990f3d7_stateAt(__state, __state.__current+1)
			}
			return __state
		}()
	}()
	return __markerState.__current != __state.__current && ____rune_private_b990f3d7_looksLikeFunctionDecl(____rune_private_b990f3d7_parserSkipNewlines(__markerState))
}

func ____rune_private_b990f3d7_looksLikeFunctionAfterName(__state __ParserState, __index int) bool {
	return ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_LParen && ____rune_private_b990f3d7_looksLikeFunctionAfterParams(__state, ____rune_private_b990f3d7_skipBalancedAt(__state, __index, __TokenKind_LParen, __TokenKind_RParen))
}

func ____rune_private_b990f3d7_looksLikeFunctionAfterParams(__state __ParserState, __index int) bool {
	__afterParams := ____rune_private_b990f3d7_skipNewlinesAt(__state, __index)
	__afterReturn := func() int {
		if ____rune_private_b990f3d7_parserKindAt(__state, __afterParams) == __TokenKind_Arrow {
			return ____rune_private_b990f3d7_skipTypeNameTokensAt(__state, __afterParams+1)
		}
		return __afterParams
	}()
	return ____rune_private_b990f3d7_parserKindAt(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __afterReturn)) == __TokenKind_FatArrow
}

func ____rune_private_b990f3d7_looksLikeLambda(__state __ParserState) bool {
	return ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LParen) && ____rune_private_b990f3d7_looksLikeLambdaAfterParams(__state, ____rune_private_b990f3d7_skipBalancedAt(__state, __state.__current, __TokenKind_LParen, __TokenKind_RParen))
}

func ____rune_private_b990f3d7_looksLikeLambdaAfterParams(__state __ParserState, __index int) bool {
	__afterParams := ____rune_private_b990f3d7_skipNewlinesAt(__state, __index)
	__afterReturn := func() int {
		if ____rune_private_b990f3d7_parserKindAt(__state, __afterParams) == __TokenKind_Arrow {
			return ____rune_private_b990f3d7_skipTypeNameTokensAt(__state, __afterParams+1)
		}
		return __afterParams
	}()
	return ____rune_private_b990f3d7_parserKindAt(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __afterReturn)) == __TokenKind_FatArrow
}

func ____rune_private_b990f3d7_looksLikeEnumMember(__state __ParserState) bool {
	__start := func() int {
		if ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_Plus) {
			return ____rune_private_b990f3d7_skipNewlinesAt(__state, __state.__current+1)
		}
		return __state.__current
	}()
	__token := ____rune_private_b990f3d7_parserTokenAt(__state, __start)
	__next := ____rune_private_b990f3d7_parserKindAt(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __start+1))
	return __token.__kind == __TokenKind_Ident && (__next == __TokenKind_Assign || __next == __TokenKind_LParen && ____rune_private_b990f3d7_startsWithUpper(__token.__lexeme) || __next != __TokenKind_Colon && __next != __TokenKind_LParen)
}

func ____rune_private_b990f3d7_startsWithUpper(__name string) bool {
	return func() bool {
		if len([]rune(__name)) > 0 {
			return []rune(__name)[0] >= 'A' && []rune(__name)[0] <= 'Z'
		}
		return false
	}()
}

func ____rune_private_b990f3d7_looksLikeObjectDestructureDecl(__state __ParserState) bool {
	return ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LBrace) && ____rune_private_b990f3d7_scanObjectDestructureDecl(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __state.__current+1))
}

func ____rune_private_b990f3d7_scanObjectDestructureDecl(__state __ParserState, __index int) bool {
	return func() bool {
		if ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_RBrace {
			return false
		}
		return ____rune_private_b990f3d7_scanObjectDestructureField(__state, __index)
	}()
}

func ____rune_private_b990f3d7_scanObjectDestructureField(__state __ParserState, __index int) bool {
	return func() bool {
		if ____rune_private_b990f3d7_parserKindAt(__state, __index) != __TokenKind_Ident {
			return false
		}
		return ____rune_private_b990f3d7_scanObjectDestructureAfterField(__state, __index+1)
	}()
}

func ____rune_private_b990f3d7_scanObjectDestructureAfterField(__state __ParserState, __index int) bool {
	return func() bool {
		if ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_Colon {
			return ____rune_private_b990f3d7_parserKindAt(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __index+1)) == __TokenKind_Ident && ____rune_private_b990f3d7_scanObjectDestructureAfterName(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __index+1)+1)
		}
		return ____rune_private_b990f3d7_scanObjectDestructureAfterName(__state, __index)
	}()
}

func ____rune_private_b990f3d7_scanObjectDestructureAfterName(__state __ParserState, __index int) bool {
	__current := ____rune_private_b990f3d7_skipNewlinesAt(__state, __index)
	__kind := ____rune_private_b990f3d7_parserKindAt(__state, __current)
	return func() bool {
		if __kind == __TokenKind_Comma {
			return ____rune_private_b990f3d7_scanObjectDestructureAfterComma(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __current+1))
		}
		return func() bool {
			if __kind == __TokenKind_RBrace {
				return ____rune_private_b990f3d7_scanObjectDestructureAfterClose(__state, __current+1)
			}
			return false
		}()
	}()
}

func ____rune_private_b990f3d7_scanObjectDestructureAfterComma(__state __ParserState, __index int) bool {
	return func() bool {
		if ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_RBrace {
			return ____rune_private_b990f3d7_scanObjectDestructureAfterClose(__state, __index+1)
		}
		return ____rune_private_b990f3d7_scanObjectDestructureField(__state, __index)
	}()
}

func ____rune_private_b990f3d7_scanObjectDestructureAfterClose(__state __ParserState, __index int) bool {
	__afterClose := ____rune_private_b990f3d7_skipNewlinesAt(__state, __index)
	return ____rune_private_b990f3d7_parserKindAt(__state, __afterClose) == __TokenKind_Declare || ____rune_private_b990f3d7_parserKindAt(__state, __afterClose) == __TokenKind_MutDeclare
}

func ____rune_private_b990f3d7_looksLikePatternBranch(__state __ParserState) bool {
	return ____rune_private_b990f3d7_tokensLookLikePatternBranch(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __state.__current))
}

func ____rune_private_b990f3d7_looksLikePatternBlockAfterSubject(__state __ParserState) bool {
	return ____rune_private_b990f3d7_parserCheck(__state, __TokenKind_LBrace) && ____rune_private_b990f3d7_tokensLookLikePatternBranch(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __state.__current+1))
}

func ____rune_private_b990f3d7_tokensLookLikePatternBranch(__state __ParserState, __index int) bool {
	__afterPattern := ____rune_private_b990f3d7_skipPatternLookahead(__state, __index)
	return __afterPattern >= 0 && ____rune_private_b990f3d7_parserKindAt(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __afterPattern)) == __TokenKind_FatArrow
}

func ____rune_private_b990f3d7_skipPatternLookahead(__state __ParserState, __index int) int {
	return ____rune_private_b990f3d7_skipOrPatternLookahead(__state, ____rune_private_b990f3d7_skipSinglePatternLookahead(__state, __index))
}

func ____rune_private_b990f3d7_skipOrPatternLookahead(__state __ParserState, __index int) int {
	return func() int {
		if __index >= 0 && ____rune_private_b990f3d7_parserKindAt(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __index)) == __TokenKind_BitOr {
			return ____rune_private_b990f3d7_skipOrPatternLookahead(__state, ____rune_private_b990f3d7_skipSinglePatternLookahead(__state, ____rune_private_b990f3d7_skipNewlinesAt(__state, __index)+1))
		}
		return __index
	}()
}

func ____rune_private_b990f3d7_skipSinglePatternLookahead(__state __ParserState, __index int) int {
	__kind := ____rune_private_b990f3d7_parserKindAt(__state, __index)
	__after := func() int {
		if __kind == __TokenKind_Underscore || __kind == __TokenKind_Int || __kind == __TokenKind_Double || __kind == __TokenKind_BigInt || __kind == __TokenKind_String || __kind == __TokenKind_Char {
			return __index + 1
		}
		return func() int {
			if __kind == __TokenKind_Ident {
				return ____rune_private_b990f3d7_skipIdentifierPatternLookahead(__state, __index)
			}
			return func() int {
				if __kind == __TokenKind_Less || __kind == __TokenKind_LessEqual || __kind == __TokenKind_Greater || __kind == __TokenKind_GreaterEqual {
					return ____rune_private_b990f3d7_skipComparePatternLookahead(__state, __index+1)
				}
				return func() int {
					if __kind == __TokenKind_LParen {
						return ____rune_private_b990f3d7_skipBalancedAt(__state, __index, __TokenKind_LParen, __TokenKind_RParen)
					}
					return func() int {
						if __kind == __TokenKind_LBracket {
							return ____rune_private_b990f3d7_skipBalancedAt(__state, __index, __TokenKind_LBracket, __TokenKind_RBracket)
						}
						return func() int {
							if __kind == __TokenKind_LBrace {
								return ____rune_private_b990f3d7_skipBalancedAt(__state, __index, __TokenKind_LBrace, __TokenKind_RBrace)
							}
							return -1
						}()
					}()
				}()
			}()
		}()
	}()
	return ____rune_private_b990f3d7_skipRangePatternLookahead(__state, __after)
}

func ____rune_private_b990f3d7_skipRangePatternLookahead(__state __ParserState, __index int) int {
	return func() int {
		if __index >= 0 && ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_DotDotEqual {
			return ____rune_private_b990f3d7_skipRangePatternEnd(__state, __index+1)
		}
		return func() int {
			if __index >= 0 && ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_DotDotLess {
				return ____rune_private_b990f3d7_skipRangePatternEnd(__state, __index+1)
			}
			return func() int {
				if __index >= 0 && ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_DotDot && ____rune_private_b990f3d7_parserKindAt(__state, __index+1) == __TokenKind_Less {
					return ____rune_private_b990f3d7_skipRangePatternEnd(__state, __index+2)
				}
				return __index
			}()
		}()
	}()
}

func ____rune_private_b990f3d7_skipRangePatternEnd(__state __ParserState, __index int) int {
	__kind := ____rune_private_b990f3d7_parserKindAt(__state, __index)
	return func() int {
		if __kind == __TokenKind_Underscore {
			return __index + 1
		}
		return func() int {
			if __kind == __TokenKind_Int || __kind == __TokenKind_Double || __kind == __TokenKind_BigInt || __kind == __TokenKind_String || __kind == __TokenKind_Char || __kind == __TokenKind_Ident {
				return ____rune_private_b990f3d7_skipIdentifierRangeEnd(__state, __index)
			}
			return -1
		}()
	}()
}

func ____rune_private_b990f3d7_skipIdentifierRangeEnd(__state __ParserState, __index int) int {
	return func() int {
		if ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_Ident && ____rune_private_b990f3d7_parserKindAt(__state, __index+1) == __TokenKind_Dot && ____rune_private_b990f3d7_parserKindAt(__state, __index+2) == __TokenKind_Ident {
			return __index + 3
		}
		return __index + 1
	}()
}

func ____rune_private_b990f3d7_skipIdentifierPatternLookahead(__state __ParserState, __index int) int {
	return func() int {
		if ____rune_private_b990f3d7_parserKindAt(__state, __index+1) == __TokenKind_LParen {
			return ____rune_private_b990f3d7_skipBalancedAt(__state, __index+1, __TokenKind_LParen, __TokenKind_RParen)
		}
		return func() int {
			if ____rune_private_b990f3d7_parserKindAt(__state, __index+1) == __TokenKind_Dot && ____rune_private_b990f3d7_parserKindAt(__state, __index+2) == __TokenKind_Ident {
				return __index + 3
			}
			return func() int {
				if ____rune_private_b990f3d7_isLiteralIdentifier(____rune_private_b990f3d7_parserTokenAt(__state, __index).__lexeme) {
					return __index + 1
				}
				return -1
			}()
		}()
	}()
}

func ____rune_private_b990f3d7_skipComparePatternLookahead(__state __ParserState, __index int) int {
	__kind := ____rune_private_b990f3d7_parserKindAt(__state, __index)
	return func() int {
		if __kind == __TokenKind_Int || __kind == __TokenKind_Double || __kind == __TokenKind_BigInt || __kind == __TokenKind_String || __kind == __TokenKind_Char || __kind == __TokenKind_Ident {
			return __index + 1
		}
		return -1
	}()
}

func ____rune_private_b990f3d7_looksLikeMapLiteralBody(__state __ParserState) bool {
	__start := ____rune_private_b990f3d7_skipNewlinesAt(__state, __state.__current+1)
	__first := ____rune_private_b990f3d7_parserTokenAt(__state, __start)
	return func() bool {
		if __first.__kind == __TokenKind_EOF || __first.__kind == __TokenKind_RBrace {
			return false
		}
		return func() bool {
			if __first.__kind == __TokenKind_Ident && ____rune_private_b990f3d7_isLiteralIdentifier(__first.__lexeme) == false {
				return false
			}
			return ____rune_private_b990f3d7_scanMapLiteralColon(__state, __start, 0)
		}()
	}()
}

func ____rune_private_b990f3d7_scanMapLiteralColon(__state __ParserState, __index int, __depth int) bool {
	__kind := ____rune_private_b990f3d7_parserKindAt(__state, __index)
	return func() bool {
		if __kind == __TokenKind_EOF {
			return false
		}
		return func() bool {
			if __kind == __TokenKind_LParen || __kind == __TokenKind_LBracket || __kind == __TokenKind_LBrace {
				return ____rune_private_b990f3d7_scanMapLiteralColon(__state, __index+1, __depth+1)
			}
			return func() bool {
				if __kind == __TokenKind_RParen || __kind == __TokenKind_RBracket {
					return ____rune_private_b990f3d7_scanMapLiteralColon(__state, __index+1, func() int {
						if __depth > 0 {
							return __depth - 1
						}
						return __depth
					}())
				}
				return func() bool {
					if __kind == __TokenKind_RBrace {
						return func() bool {
							if __depth == 0 {
								return false
							}
							return ____rune_private_b990f3d7_scanMapLiteralColon(__state, __index+1, __depth-1)
						}()
					}
					return func() bool {
						if (__kind == __TokenKind_Newline || __kind == __TokenKind_Question || __kind == __TokenKind_FatArrow) && __depth == 0 {
							return false
						}
						return func() bool {
							if __kind == __TokenKind_Colon && __depth == 0 {
								return true
							}
							return ____rune_private_b990f3d7_scanMapLiteralColon(__state, __index+1, __depth)
						}()
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_b990f3d7_isLiteralIdentifier(__name string) bool {
	return __name == "true" || __name == "false" || __name == "null"
}

func ____rune_private_b990f3d7_skipNewlinesAt(__state __ParserState, __index int) int {
	return func() int {
		if ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_Newline {
			return ____rune_private_b990f3d7_skipNewlinesAt(__state, __index+1)
		}
		return __index
	}()
}

func ____rune_private_b990f3d7_skipGenericNamesAt(__state __ParserState, __index int) int {
	return func() int {
		if ____rune_private_b990f3d7_parserKindAt(__state, __index) == __TokenKind_LBracket {
			return ____rune_private_b990f3d7_skipBalancedAt(__state, __index, __TokenKind_LBracket, __TokenKind_RBracket)
		}
		return __index
	}()
}

func ____rune_private_b990f3d7_skipBalancedAt(__state __ParserState, __index int, __openKind __TokenKind, __closeKind __TokenKind) int {
	return ____rune_private_b990f3d7_skipBalancedAtLoop(__state, __index, __openKind, __closeKind, 0)
}

func ____rune_private_b990f3d7_skipBalancedAtLoop(__state __ParserState, __index int, __openKind __TokenKind, __closeKind __TokenKind, __depth int) int {
	__kind := ____rune_private_b990f3d7_parserKindAt(__state, __index)
	return func() int {
		if __kind == __TokenKind_EOF {
			return __index
		}
		return func() int {
			if __kind == __openKind {
				return ____rune_private_b990f3d7_skipBalancedAtLoop(__state, __index+1, __openKind, __closeKind, __depth+1)
			}
			return func() int {
				if __kind == __closeKind {
					return func() int {
						if __depth <= 1 {
							return __index + 1
						}
						return ____rune_private_b990f3d7_skipBalancedAtLoop(__state, __index+1, __openKind, __closeKind, __depth-1)
					}()
				}
				return ____rune_private_b990f3d7_skipBalancedAtLoop(__state, __index+1, __openKind, __closeKind, __depth)
			}()
		}()
	}()
}

func ____rune_private_b990f3d7_skipTypeNameTokensAt(__state __ParserState, __index int) int {
	return ____rune_private_b990f3d7_skipTypeNameTokensAtLoop(__state, __index, 0)
}

func ____rune_private_b990f3d7_skipTypeNameTokensAtLoop(__state __ParserState, __index int, __depth int) int {
	__kind := ____rune_private_b990f3d7_parserKindAt(__state, __index)
	return func() int {
		if __kind == __TokenKind_Ident || __kind == __TokenKind_Comma || __kind == __TokenKind_Colon || __kind == __TokenKind_Question || __kind == __TokenKind_Arrow || __kind == __TokenKind_At || __kind == __TokenKind_Dot {
			return ____rune_private_b990f3d7_skipTypeNameTokensAtLoop(__state, __index+1, __depth)
		}
		return func() int {
			if __kind == __TokenKind_LBracket || __kind == __TokenKind_LParen {
				return ____rune_private_b990f3d7_skipTypeNameTokensAtLoop(__state, __index+1, __depth+1)
			}
			return func() int {
				if __kind == __TokenKind_RBracket || __kind == __TokenKind_RParen {
					return func() int {
						if __depth == 0 {
							return __index
						}
						return ____rune_private_b990f3d7_skipTypeNameTokensAtLoop(__state, __index+1, __depth-1)
					}()
				}
				return __index
			}()
		}()
	}()
}

func __lower(__source string) __IRFile {
	return __lowerParsed(__parse(__source))
}

func __lowerParsed(__file __ParsedFile) __IRFile {
	__out := __IRFile{__imports: []__IRImport{}, __tsImports: []__IRTSImport{}, __structs: []__IRStructType{}, __enums: []__IREnumType{}, __constants: []__IRConst{}, __functions: []__IRFunction{}, __tests: []__IRTest{}, __errors: __file.__errors}
	for _, __importDecl := range __file.__imports {
		_ = __importDecl
		func() int {
			__out.__imports = append(__out.__imports, ____rune_private_44103c8f_lowerImport(__importDecl))
			return len(__out.__imports)
		}()
	}
	for _, __typeDecl := range __file.__types {
		_ = __typeDecl
		__out = ____rune_private_44103c8f_lowerTypeInto(__out, __typeDecl)
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		func() int {
			__out.__functions = append(__out.__functions, ____rune_private_44103c8f_lowerFunction(__fn))
			return len(__out.__functions)
		}()
	}
	for _, __testDecl := range __file.__tests {
		_ = __testDecl
		func() int {
			__out.__tests = append(__out.__tests, ____rune_private_44103c8f_lowerTest(__testDecl))
			return len(__out.__tests)
		}()
	}
	return __out
}

func ____rune_private_44103c8f_lowerTypeInto(__file __IRFile, __typeDecl __ParsedType) __IRFile {
	return func() __IRFile {
		if __typeDecl.__enum {
			return ____rune_private_44103c8f_pushEnumType(__file, __typeDecl)
		}
		return ____rune_private_44103c8f_pushStructType(__file, __typeDecl)
	}()
}

func ____rune_private_44103c8f_pushStructType(__file __IRFile, __typeDecl __ParsedType) __IRFile {
	__file.__structs = append(__file.__structs, ____rune_private_44103c8f_lowerStructType(__typeDecl))
	return __file
}

func ____rune_private_44103c8f_pushEnumType(__file __IRFile, __typeDecl __ParsedType) __IRFile {
	__file.__enums = append(__file.__enums, ____rune_private_44103c8f_lowerEnumType(__typeDecl))
	return __file
}

func __emptyIRExpr() __IRExpr {
	return __IRExpr{__kind: __ExprKind_Unknown, __text: "", __name: "", __value: "", __op: "", __params: []__IRParam{}, __children: []__IRExpr{}, __line: 0, __column: 0}
}

func __emptyIRFunction() __IRFunction {
	return __IRFunction{__name: "", __private: false, __routine: false, __macro: false, __receiverType: "", __generics: []string{}, __params: []__IRParam{}, __returnType: "", __body: __emptyIRExpr(), __line: 0, __column: 0}
}

func ____rune_private_44103c8f_lowerImport(__importDecl __ParsedImport) __IRImport {
	return __IRImport{__path: __importDecl.__path, __go: __importDecl.__go, __line: __importDecl.__line, __column: __importDecl.__column}
}

func ____rune_private_44103c8f_lowerParam(__param __ParsedParam) __IRParam {
	return __IRParam{__name: __param.__name, __typeName: __typeRefToString(__param.__typeRef), __line: __param.__line, __column: __param.__column}
}

func ____rune_private_44103c8f_lowerField(__field __ParsedField) __IRField {
	return __IRField{__name: __field.__name, __private: __field.__private, __typeName: __typeRefToString(__field.__typeRef), __line: __field.__line, __column: __field.__column}
}

func ____rune_private_44103c8f_lowerEnumMember(__member __ParsedEnumMember) __IREnumMember {
	return __IREnumMember{__name: __member.__name, __private: __member.__private, __value: __member.__value, __params: ____rune_private_44103c8f_lowerParams(__member.__params), __line: __member.__line, __column: __member.__column}
}

func ____rune_private_44103c8f_lowerFunction(__fn __ParsedFunction) __IRFunction {
	return __IRFunction{__name: __fn.__name, __private: __fn.__private, __routine: __fn.__routine, __macro: ____rune_private_44103c8f_parsedFunctionCompileTimeOnly(__fn), __receiverType: __fn.__receiverType, __generics: __fn.__generics, __params: ____rune_private_44103c8f_lowerParams(__fn.__params), __returnType: __typeRefToString(__fn.__returnType), __body: ____rune_private_44103c8f_lowerExpr(__fn.__body), __line: __fn.__line, __column: __fn.__column}
}

func ____rune_private_44103c8f_parsedFunctionCompileTimeOnly(__fn __ParsedFunction) bool {
	return __fn.__macro || ____rune_private_44103c8f_typeRefIsSyntaxOnly(__fn.__returnType) || ____rune_private_44103c8f_paramsUseSyntaxOnly(__fn.__params, 0)
}

func ____rune_private_44103c8f_paramsUseSyntaxOnly(__params []__ParsedParam, __index int) bool {
	return func() bool {
		if __index >= len(__params) {
			return false
		}
		return ____rune_private_44103c8f_typeRefIsSyntaxOnly(__params[__index].__typeRef) || ____rune_private_44103c8f_paramsUseSyntaxOnly(__params, __index+1)
	}()
}

func ____rune_private_44103c8f_typeRefIsSyntaxOnly(__typeRef __ParsedTypeRef) bool {
	__name := __typeRefToString(__typeRef)
	return strings.HasPrefix(__name, "Syntax") || __name == "MacroContext"
}

func ____rune_private_44103c8f_lowerStructType(__typeDecl __ParsedType) __IRStructType {
	return __IRStructType{__name: __typeDecl.__name, __private: __typeDecl.__private, __generics: __typeDecl.__generics, __fields: ____rune_private_44103c8f_lowerFields(__typeDecl.__fields), __methods: ____rune_private_44103c8f_lowerFunctions(__typeDecl.__methods), __line: __typeDecl.__line, __column: __typeDecl.__column}
}

func ____rune_private_44103c8f_lowerEnumType(__typeDecl __ParsedType) __IREnumType {
	return __IREnumType{__name: __typeDecl.__name, __private: __typeDecl.__private, __generics: __typeDecl.__generics, __members: ____rune_private_44103c8f_lowerEnumMembers(__typeDecl.__members), __methods: ____rune_private_44103c8f_lowerFunctions(__typeDecl.__methods), __line: __typeDecl.__line, __column: __typeDecl.__column}
}

func ____rune_private_44103c8f_lowerTest(__testDecl __ParsedTest) __IRTest {
	return __IRTest{__name: __testDecl.__name, __body: ____rune_private_44103c8f_lowerExpr(__testDecl.__body), __line: __testDecl.__line, __column: __testDecl.__column}
}

func ____rune_private_44103c8f_lowerExpr(__expr __ParsedExpr) __IRExpr {
	__children := ____rune_private_44103c8f_lowerExprs(__expr.__children)
	return __IRExpr{__kind: __expr.__kind, __text: ____rune_private_44103c8f_inferIRExprText(__expr, __children), __name: __expr.__name, __value: __expr.__value, __op: __expr.__op, __params: ____rune_private_44103c8f_lowerParams(__expr.__params), __children: __children, __line: __expr.__line, __column: __expr.__column}
}

func ____rune_private_44103c8f_inferIRExprText(__expr __ParsedExpr, __children []__IRExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_44103c8f_inferIRBinaryText(__expr, __children)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_44103c8f_inferIRCallText(__children)
		case __expr.__kind == __ExprKind_String:
			return "String"
		case __expr.__kind == __ExprKind_Template:
			return "String"
		case __expr.__kind == __ExprKind_Int:
			return "Int"
		case __expr.__kind == __ExprKind_Double:
			return "Double"
		case __expr.__kind == __ExprKind_BigInt:
			return "BigInt"
		case __expr.__kind == __ExprKind_Char:
			return "Char"
		case __expr.__kind == __ExprKind_Bool:
			return "Bool"
		case __expr.__kind == __ExprKind_Null:
			return "Null"
		default:
			return func() string {
				if __expr.__text != "" {
					return __expr.__text
				}
				return ""
			}()
		}
	}()
}

func ____rune_private_44103c8f_inferIRBinaryText(__expr __ParsedExpr, __children []__IRExpr) string {
	return func() string {
		if __expr.__op == "??" {
			return ____rune_private_44103c8f_inferIRCoalesceText(__children)
		}
		return ""
	}()
}

func ____rune_private_44103c8f_inferIRCoalesceText(__children []__IRExpr) string {
	return func() string {
		if len(__children) < 2 {
			return ""
		}
		return func() string {
			if __children[1].__text != "" && __children[1].__text != "Null" {
				return __children[1].__text
			}
			return __children[0].__text
		}()
	}()
}

func ____rune_private_44103c8f_inferIRCallText(__children []__IRExpr) string {
	return func() string {
		if len(__children) == 0 {
			return ""
		}
		return func() string {
			if __children[0].__kind == __ExprKind_Selector {
				return ____rune_private_44103c8f_inferIRSelectorCallText(__children[0], __children)
			}
			return ""
		}()
	}()
}

func ____rune_private_44103c8f_inferIRSelectorCallText(__selector __IRExpr, __children []__IRExpr) string {
	return func() string {
		switch {
		case __selector.__name == "getOr":
			return func() string {
				if len(__children) > 2 {
					return __children[2].__text
				}
				return ""
			}()
		case __selector.__name == "isEmpty":
			return "Bool"
		case (__selector.__name == "length") || (__selector.__name == "byteLength") || (__selector.__name == "size") || (__selector.__name == "push"):
			return "Int"
		case (__selector.__name == "includes") || (__selector.__name == "contains") || (__selector.__name == "startsWith") || (__selector.__name == "endsWith"):
			return "Bool"
		case (__selector.__name == "toString") || (__selector.__name == "trim") || (__selector.__name == "toUpperCase") || (__selector.__name == "toLowerCase"):
			return "String"
		default:
			return ""
		}
	}()
}

func ____rune_private_44103c8f_lowerParams(__params []__ParsedParam) []__IRParam {
	__out := append([]__IRParam{}, []__IRParam{__IRParam{__name: "", __typeName: "", __line: 0, __column: 0}}[0:0]...)
	for _, __param := range __params {
		_ = __param
		func() int { __out = append(__out, ____rune_private_44103c8f_lowerParam(__param)); return len(__out) }()
	}
	return __out
}

func ____rune_private_44103c8f_lowerFields(__fields []__ParsedField) []__IRField {
	__out := append([]__IRField{}, []__IRField{__IRField{__name: "", __private: false, __typeName: "", __line: 0, __column: 0}}[0:0]...)
	for _, __field := range __fields {
		_ = __field
		func() int { __out = append(__out, ____rune_private_44103c8f_lowerField(__field)); return len(__out) }()
	}
	return __out
}

func ____rune_private_44103c8f_lowerEnumMembers(__members []__ParsedEnumMember) []__IREnumMember {
	__out := append([]__IREnumMember{}, []__IREnumMember{__IREnumMember{__name: "", __private: false, __value: "", __params: []__IRParam{}, __line: 0, __column: 0}}[0:0]...)
	for _, __member := range __members {
		_ = __member
		func() int {
			__out = append(__out, ____rune_private_44103c8f_lowerEnumMember(__member))
			return len(__out)
		}()
	}
	return __out
}

func ____rune_private_44103c8f_lowerFunctions(__functions []__ParsedFunction) []__IRFunction {
	__out := append([]__IRFunction{}, []__IRFunction{__emptyIRFunction()}[0:0]...)
	for _, __fn := range __functions {
		_ = __fn
		func() int { __out = append(__out, ____rune_private_44103c8f_lowerFunction(__fn)); return len(__out) }()
	}
	return __out
}

func ____rune_private_44103c8f_lowerExprs(__exprs []__ParsedExpr) []__IRExpr {
	__out := append([]__IRExpr{}, []__IRExpr{__emptyIRExpr()}[0:0]...)
	for _, __expr := range __exprs {
		_ = __expr
		func() int { __out = append(__out, ____rune_private_44103c8f_lowerExpr(__expr)); return len(__out) }()
	}
	return __out
}

func __enumValue(__member __IREnumMember, __index int) string {
	return func() string {
		if __member.__value == "" {
			return __compilerIntToString(__index)
		}
		return __member.__value
	}()
}

func __returnsValue(__typeName string) bool {
	return __typeName != "" && __typeName != "Void"
}

func __bigintLiteralDigits(__value string) string {
	return func() string {
		if strings.HasSuffix(__value, "n") {
			return func() string { runes := []rune(__value); return string(runes[0 : len([]rune(__value))-1]) }()
		}
		return __value
	}()
}

func __fileUsesPathBasenameFamily(__file __IRFile) bool {
	return __fileUsesModuleCall(__file, "path.basename") || __fileUsesModuleCall(__file, "path.extname") || __fileUsesModuleCall(__file, "path.dirname")
}

func __fileUsesPathFamily(__file __IRFile) bool {
	return __fileUsesPathBasenameFamily(__file) || __fileUsesModuleCall(__file, "path.join") || __fileUsesModuleCall(__file, "path.normalize") || __fileUsesModuleCall(__file, "path.resolve") || __fileUsesModuleCall(__file, "path.relative") || __fileUsesPathHelperFamily(__file)
}

func __fileUsesPathHelperFamily(__file __IRFile) bool {
	return __fileUsesModuleCall(__file, "path.joinParts") || __fileUsesModuleCall(__file, "path.appendPathPart") || __fileUsesModuleCall(__file, "path.normalizeParts") || __fileUsesModuleCall(__file, "path.normalizePart") || __fileUsesModuleCall(__file, "path.normalizeParent") || __fileUsesModuleCall(__file, "path.normalizePop") || __fileUsesModuleCall(__file, "path.normalizePush") || __fileUsesModuleCall(__file, "path.pathParts") || __fileUsesModuleCall(__file, "path.collectPathParts") || __fileUsesModuleCall(__file, "path.collectPathPart") || __fileUsesModuleCall(__file, "path.relativeFromParts") || __fileUsesModuleCall(__file, "path.relativeTail")
}

func __moduleCallKey(__expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_Call && len(__expr.__children) > 0 {
			return __moduleSelectorKey(__expr.__children[0])
		}
		return ""
	}()
}

func __moduleSelectorKey(__expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_Selector && len(__expr.__children) > 0 {
			return func() string {
				if __expr.__children[0].__kind == __ExprKind_At {
					return __expr.__children[0].__name + "." + __expr.__name
				}
				return ""
			}()
		}
		return ""
	}()
}

func __genericInner(__typeName string, __base string) string {
	__prefix := __base + "["
	return func() string {
		if strings.HasPrefix(__typeName, __prefix) && strings.HasSuffix(__typeName, "]") {
			return func() string {
				runes := []rune(__typeName)
				return string(runes[len([]rune(__prefix)) : len([]rune(__typeName))-1])
			}()
		}
		return ""
	}()
}

func __typeArg(__args string, __index int) string {
	__parts := func() []string { parts := strings.Split(__args, ","); return parts }()
	return func() string {
		if __index < len(__parts) {
			return strings.TrimSpace(__parts[__index])
		}
		return "Dynamic"
	}()
}

func __mangleIdent(__name string) string {
	return "__" + strings.ReplaceAll((strings.ReplaceAll((strings.ReplaceAll(__name, ".", "_")), "-", "_")), "@", "_")
}

func __indent(__level int) string {
	return func() string {
		if __level <= 0 {
			return ""
		}
		return "  " + __indent(__level-1)
	}()
}

func __line(__level int, __text string) string {
	return __indent(__level) + __text + "\n"
}

func __fileUsesUnwrap(__file __IRFile) bool {
	return ____rune_private_4226a467_functionsUseUnwrap(__file.__functions, 0) || ____rune_private_4226a467_structsUseUnwrap(__file.__structs, 0) || ____rune_private_4226a467_enumsUseUnwrap(__file.__enums, 0) || ____rune_private_4226a467_testsUseUnwrap(__file.__tests, 0)
}

func ____rune_private_4226a467_functionsUseUnwrap(__functions []__IRFunction, __index int) bool {
	return func() bool {
		if __index >= len(__functions) {
			return false
		}
		return func() bool {
			if __functions[__index].__macro {
				return ____rune_private_4226a467_functionsUseUnwrap(__functions, __index+1)
			}
			return __exprUsesUnwrap(__functions[__index].__body) || ____rune_private_4226a467_functionsUseUnwrap(__functions, __index+1)
		}()
	}()
}

func ____rune_private_4226a467_testsUseUnwrap(__tests []__IRTest, __index int) bool {
	return func() bool {
		if __index >= len(__tests) {
			return false
		}
		return __exprUsesUnwrap(__tests[__index].__body) || ____rune_private_4226a467_testsUseUnwrap(__tests, __index+1)
	}()
}

func ____rune_private_4226a467_structsUseUnwrap(__structs []__IRStructType, __index int) bool {
	return func() bool {
		if __index >= len(__structs) {
			return false
		}
		return ____rune_private_4226a467_functionsUseUnwrap(__structs[__index].__methods, 0) || ____rune_private_4226a467_structsUseUnwrap(__structs, __index+1)
	}()
}

func ____rune_private_4226a467_enumsUseUnwrap(__enums []__IREnumType, __index int) bool {
	return func() bool {
		if __index >= len(__enums) {
			return false
		}
		return ____rune_private_4226a467_functionsUseUnwrap(__enums[__index].__methods, 0) || ____rune_private_4226a467_enumsUseUnwrap(__enums, __index+1)
	}()
}

func __exprUsesUnwrap(__expr __IRExpr) bool {
	return func() bool {
		if __expr.__kind == __ExprKind_Unwrap {
			return true
		}
		return ____rune_private_4226a467_exprChildrenUseUnwrap(__expr.__children, 0)
	}()
}

func __fileUsesModuleCall(__file __IRFile, __key string) bool {
	return ____rune_private_4226a467_functionsUseModuleCall(__file.__functions, __key, 0) || ____rune_private_4226a467_structsUseModuleCall(__file.__structs, __key, 0) || ____rune_private_4226a467_enumsUseModuleCall(__file.__enums, __key, 0) || ____rune_private_4226a467_testsUseModuleCall(__file.__tests, __key, 0)
}

func ____rune_private_4226a467_functionsUseModuleCall(__functions []__IRFunction, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__functions) {
			return false
		}
		return func() bool {
			if __functions[__index].__macro {
				return ____rune_private_4226a467_functionsUseModuleCall(__functions, __key, __index+1)
			}
			return ____rune_private_4226a467_exprUsesModuleCall(__functions[__index].__body, __key) || ____rune_private_4226a467_functionsUseModuleCall(__functions, __key, __index+1)
		}()
	}()
}

func ____rune_private_4226a467_testsUseModuleCall(__tests []__IRTest, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__tests) {
			return false
		}
		return ____rune_private_4226a467_exprUsesModuleCall(__tests[__index].__body, __key) || ____rune_private_4226a467_testsUseModuleCall(__tests, __key, __index+1)
	}()
}

func ____rune_private_4226a467_structsUseModuleCall(__structs []__IRStructType, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__structs) {
			return false
		}
		return ____rune_private_4226a467_functionsUseModuleCall(__structs[__index].__methods, __key, 0) || ____rune_private_4226a467_structsUseModuleCall(__structs, __key, __index+1)
	}()
}

func ____rune_private_4226a467_enumsUseModuleCall(__enums []__IREnumType, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__enums) {
			return false
		}
		return ____rune_private_4226a467_functionsUseModuleCall(__enums[__index].__methods, __key, 0) || ____rune_private_4226a467_enumsUseModuleCall(__enums, __key, __index+1)
	}()
}

func ____rune_private_4226a467_exprUsesModuleCall(__expr __IRExpr, __key string) bool {
	return func() bool {
		if __moduleCallKey(__expr) == __key {
			return true
		}
		return ____rune_private_4226a467_exprChildrenUseModuleCall(__expr.__children, __key, 0)
	}()
}

func ____rune_private_4226a467_exprChildrenUseModuleCall(__children []__IRExpr, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__children) {
			return false
		}
		return ____rune_private_4226a467_exprUsesModuleCall(__children[__index], __key) || ____rune_private_4226a467_exprChildrenUseModuleCall(__children, __key, __index+1)
	}()
}

func ____rune_private_4226a467_exprChildrenUseUnwrap(__children []__IRExpr, __index int) bool {
	return func() bool {
		if __index >= len(__children) {
			return false
		}
		return __exprUsesUnwrap(__children[__index]) || ____rune_private_4226a467_exprChildrenUseUnwrap(__children, __index+1)
	}()
}

func __compilerIntToString(__value int) string {
	return func() string {
		if __value == 0 {
			return "0"
		}
		return func() string {
			if __value < 0 {
				return "-" + ____rune_private_4226a467_compilerUnsignedIntToString(0-__value, "")
			}
			return ____rune_private_4226a467_compilerUnsignedIntToString(__value, "")
		}()
	}()
}

func ____rune_private_4226a467_compilerUnsignedIntToString(__value int, __out string) string {
	return func() string {
		if __value <= 0 {
			return __out
		}
		return ____rune_private_4226a467_compilerUnsignedIntToString(__value/10, ____rune_private_4226a467_compilerDigitString(__value%10)+__out)
	}()
}

func ____rune_private_4226a467_compilerDigitString(__value int) string {
	return func() string {
		switch {
		case __value == 0:
			return "0"
		case __value == 1:
			return "1"
		case __value == 2:
			return "2"
		case __value == 3:
			return "3"
		case __value == 4:
			return "4"
		case __value == 5:
			return "5"
		case __value == 6:
			return "6"
		case __value == 7:
			return "7"
		case __value == 8:
			return "8"
		default:
			return "9"
		}
	}()
}

func __generateGo(__file __IRFile) string {
	__out := "package main\n\n"
	__out = __out + ____rune_private_8ddf8596_emitGoImports(__file)
	if __fileUsesUnwrap(__file) {
		__out = __out + ____rune_private_8ddf8596_emitGoUnwrapHelper()
	}
	if __fileUsesPathFamily(__file) {
		__out = __out + ____rune_private_8ddf8596_emitGoPathHelpers()
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_8ddf8596_emitGoEnum(__enumDecl) + "\n"
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_8ddf8596_emitGoEnumMethods(__file, __enumDecl)
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_8ddf8596_emitGoStruct(__typeDecl) + "\n"
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_8ddf8596_emitGoMethods(__file, __typeDecl)
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__out = func() string {
			if __fn.__macro {
				return __out
			}
			return __out + ____rune_private_8ddf8596_emitGoFunction(__file, __fn, "") + "\n"
		}()
	}
	if ____rune_private_8ddf8596_hasMain(__file) {
		__out = __out + "func main() {\n\t" + __mangleIdent("main") + "()\n}\n"
	}
	return __out
}

func ____rune_private_8ddf8596_emitGoImports(__file __IRFile) string {
	__imports := []string{}
	if ____rune_private_8ddf8596_usesPrintFile(__file) {
		func() int { __imports = append(__imports, "fmt"); return len(__imports) }()
	}
	if ____rune_private_8ddf8596_fileUsesGoStrings(__file) {
		func() int { __imports = append(__imports, "strings"); return len(__imports) }()
	}
	if __fileUsesModuleCall(__file, "process.platform") {
		func() int { __imports = append(__imports, "runtime"); return len(__imports) }()
	}
	if __fileUsesModuleCall(__file, "process.argv") || __fileUsesModuleCall(__file, "process.exit") {
		func() int { __imports = append(__imports, "os"); return len(__imports) }()
	}
	if ____rune_private_8ddf8596_fileUsesDoubleMath(__file) {
		func() int { __imports = append(__imports, "math"); return len(__imports) }()
	}
	if __fileUsesModuleCall(__file, "int.toString") || __fileUsesModuleCall(__file, "bigint.toString") {
		func() int { __imports = append(__imports, "strconv"); return len(__imports) }()
	}
	if __fileUsesUnwrap(__file) {
		func() int { __imports = append(__imports, "reflect"); return len(__imports) }()
	}
	return func() string {
		if len(__imports) == 0 {
			return ""
		}
		return "import (\n" + ____rune_private_8ddf8596_emitGoImportLines(__imports, 0, "") + ")\n\n"
	}()
}

func ____rune_private_8ddf8596_emitGoImportLines(__imports []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__imports) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoImportLines(__imports, __index+1, __out+"\t\""+__imports[__index]+"\"\n")
	}()
}

func ____rune_private_8ddf8596_emitGoUnwrapHelper() string {
	return "func __runeUnwrap(value any) any {\n\tv := reflect.ValueOf(value)\n\tif v.Kind() == reflect.Pointer {\n\t\tv = v.Elem()\n\t}\n\ttag := v.FieldByName(\"__tag\").Int()\n\tpayload := v.FieldByName(\"__payload\")\n\tif tag == 0 {\n\t\tif payload.Len() == 0 {\n\t\t\treturn nil\n\t\t}\n\t\treturn payload.Index(0).Interface()\n\t}\n\tif payload.Len() > 0 {\n\t\tpanic(payload.Index(0).Interface())\n\t}\n\tpanic(\"Result.Err\")\n}\n\n"
}

func ____rune_private_8ddf8596_emitGoPathHelpers() string {
	return "func __runePathBasename(path string) string {\n\tindex := strings.LastIndex(path, \"/\")\n\tif index < 0 {\n\t\treturn path\n\t}\n\tif index == len(path)-1 {\n\t\treturn path\n\t}\n\treturn path[index+1:]\n}\n\nfunc __runePathExtname(path string) string {\n\tbase := __runePathBasename(path)\n\tindex := strings.LastIndex(base, \".\")\n\tif index <= 0 {\n\t\treturn \"\"\n\t}\n\treturn base[index:]\n}\n\nfunc __runePathDirname(path string) string {\n\tindex := strings.LastIndex(path, \"/\")\n\tif index < 0 {\n\t\treturn \".\"\n\t}\n\tif index == 0 {\n\t\treturn \"/\"\n\t}\n\treturn path[:index]\n}\n\nfunc __runePathJoin(parts []any) string {\n\treturn __runePathNormalize(__runePathJoinParts(__runePathStringParts(parts), 0, \"\"))\n}\n\nfunc __runePathNormalize(path string) string {\n\tabsolute := strings.HasPrefix(path, \"/\")\n\tout := __runePathNormalizeParts(strings.Split(path, \"/\"), 0, absolute, []string{})\n\tjoined := __runePathJoinParts(out, 0, \"\")\n\tif absolute {\n\t\treturn \"/\" + joined\n\t}\n\tif joined == \"\" {\n\t\treturn \".\"\n\t}\n\treturn joined\n}\n\nfunc __runePathResolve(parts []any) string {\n\tif len(parts) == 0 {\n\t\treturn \".\"\n\t}\n\treturn __runePathNormalize(__runePathJoin(parts))\n}\n\nfunc __runePathRelative(from string, to string) string {\n\tfromParts := __runePathParts(__runePathResolve([]any{from}))\n\ttoParts := __runePathParts(__runePathResolve([]any{to}))\n\tindex := 0\n\tfor index < len(fromParts) && index < len(toParts) && fromParts[index] == toParts[index] {\n\t\tindex++\n\t}\n\tout := \"\"\n\tfor i := index; i < len(fromParts); i++ {\n\t\tout = __runePathAppendPart(out, \"..\")\n\t}\n\tfor i := index; i < len(toParts); i++ {\n\t\tout = __runePathAppendPart(out, toParts[i])\n\t}\n\tif out == \"\" {\n\t\treturn \".\"\n\t}\n\treturn out\n}\n\nfunc __runePathStringParts(parts []any) []string {\n\tout := make([]string, 0, len(parts))\n\tfor _, part := range parts {\n\t\tout = append(out, part.(string))\n\t}\n\treturn out\n}\n\nfunc __runePathParts(path string) []string {\n\tclean := __runePathNormalize(path)\n\tout := []string{}\n\tfor _, part := range strings.Split(clean, \"/\") {\n\t\tif part != \"\" {\n\t\t\tout = append(out, part)\n\t\t}\n\t}\n\treturn out\n}\n\nfunc __runePathJoinParts(parts []string, index int, out string) string {\n\tfor index < len(parts) {\n\t\tout = __runePathAppendPart(out, parts[index])\n\t\tindex++\n\t}\n\treturn out\n}\n\nfunc __runePathAppendPart(out string, part string) string {\n\tif out == \"\" {\n\t\treturn part\n\t}\n\tif part == \"\" {\n\t\treturn out\n\t}\n\treturn out + \"/\" + part\n}\n\nfunc __runePathNormalizeParts(parts []string, index int, absolute bool, out []string) []string {\n\tfor index < len(parts) {\n\t\tpart := parts[index]\n\t\tif part == \"\" || part == \".\" {\n\t\t\tindex++\n\t\t\tcontinue\n\t\t}\n\t\tif part == \"..\" {\n\t\t\treturn __runePathNormalizeParent(parts, index, absolute, out)\n\t\t}\n\t\treturn __runePathNormalizePush(parts, index, absolute, out, part)\n\t}\n\treturn out\n}\n\nfunc __runePathNormalizeParent(parts []string, index int, absolute bool, out []string) []string {\n\tif len(out) > 0 {\n\t\treturn __runePathNormalizePop(parts, index, absolute, out)\n\t}\n\tif absolute {\n\t\treturn __runePathNormalizeParts(parts, index+1, absolute, out)\n\t}\n\treturn __runePathNormalizePush(parts, index, absolute, out, \"..\")\n}\n\nfunc __runePathNormalizePop(parts []string, index int, absolute bool, out []string) []string {\n\treturn __runePathNormalizeParts(parts, index+1, absolute, out[:len(out)-1])\n}\n\nfunc __runePathNormalizePush(parts []string, index int, absolute bool, out []string, part string) []string {\n\treturn __runePathNormalizeParts(parts, index+1, absolute, append(out, part))\n}\n\nfunc __runePathCollectParts(parts []string, index int, out []string) []string {\n\tfor index < len(parts) {\n\t\tif parts[index] != \"\" {\n\t\t\tout = append(out, parts[index])\n\t\t}\n\t\tindex++\n\t}\n\treturn out\n}\n\nfunc __runePathCollectPart(parts []string, index int, out []string) []string {\n\tif index < len(parts) {\n\t\tout = append(out, parts[index])\n\t}\n\treturn __runePathCollectParts(parts, index+1, out)\n}\n\nfunc __runePathRelativeFromParts(fromParts []string, toParts []string, index int) string {\n\tfor index < len(fromParts) && index < len(toParts) && fromParts[index] == toParts[index] {\n\t\tindex++\n\t}\n\treturn __runePathRelativeTail(fromParts, toParts, index, index, \"\")\n}\n\nfunc __runePathRelativeTail(fromParts []string, toParts []string, fromIndex int, toIndex int, out string) string {\n\tfor fromIndex < len(fromParts) {\n\t\tout = __runePathAppendPart(out, \"..\")\n\t\tfromIndex++\n\t}\n\tfor toIndex < len(toParts) {\n\t\tout = __runePathAppendPart(out, toParts[toIndex])\n\t\ttoIndex++\n\t}\n\tif out == \"\" {\n\t\treturn \".\"\n\t}\n\treturn out\n}\n\n"
}

func ____rune_private_8ddf8596_emitGoEnum(__enumDecl __IREnumType) string {
	return func() string {
		if ____rune_private_8ddf8596_enumHasPayload(__enumDecl.__members) {
			return ____rune_private_8ddf8596_emitGoPayloadEnum(__enumDecl)
		}
		return ____rune_private_8ddf8596_emitGoSimpleEnum(__enumDecl)
	}()
}

func ____rune_private_8ddf8596_emitGoSimpleEnum(__enumDecl __IREnumType) string {
	__out := "type " + __mangleIdent(__enumDecl.__name) + " int\n\n"
	__out = __out + "const (\n"
	__out = __out + ____rune_private_8ddf8596_emitGoEnumMembers(__enumDecl.__name, __enumDecl.__members, 0, "")
	return __out + ")\n"
}

func ____rune_private_8ddf8596_emitGoPayloadEnum(__enumDecl __IREnumType) string {
	__out := "type " + __mangleIdent(__enumDecl.__name) + ____rune_private_8ddf8596_emitGoGenericsDecl(__enumDecl.__generics) + " struct {\n"
	__out = __out + "\t__tag int\n"
	__out = __out + "\t__payload []any\n"
	__out = __out + "}\n\n"
	__out = __out + "const (\n"
	__out = __out + ____rune_private_8ddf8596_emitGoPayloadEnumTags(__enumDecl.__name, __enumDecl.__members, 0, "")
	__out = __out + ")\n\n"
	return __out + ____rune_private_8ddf8596_emitGoPayloadEnumConstructors(__enumDecl.__name, __enumDecl.__generics, __enumDecl.__members, 0, "")
}

func ____rune_private_8ddf8596_emitGoPayloadEnumTags(__enumName string, __members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoPayloadEnumTags(__enumName, __members, __index+1, __out+"\t"+__mangleIdent(__enumName+"_"+__members[__index].__name+"_tag")+" = "+__enumValue(__members[__index], __index)+"\n")
	}()
}

func ____rune_private_8ddf8596_emitGoPayloadEnumConstructors(__enumName string, __generics []string, __members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoPayloadEnumConstructors(__enumName, __generics, __members, __index+1, __out+____rune_private_8ddf8596_emitGoPayloadEnumConstructor(__enumName, __generics, __members[__index]))
	}()
}

func ____rune_private_8ddf8596_emitGoPayloadEnumConstructor(__enumName string, __generics []string, __member __IREnumMember) string {
	__tagName := __mangleIdent(__enumName + "_" + __member.__name + "_tag")
	__typeName := __mangleIdent(__enumName) + ____rune_private_8ddf8596_emitGoGenericsUse(__generics)
	return func() string {
		if len(__member.__params) == 0 {
			return "var " + __mangleIdent(__enumName+"_"+__member.__name) + " = " + __typeName + "{__tag: " + __tagName + ", __payload: nil}\n"
		}
		return "func " + __mangleIdent(__member.__name) + ____rune_private_8ddf8596_emitGoGenericsDecl(__generics) + "(" + ____rune_private_8ddf8596_emitGoParams(__member.__params, 0, "") + ") " + __typeName + " {\n\treturn " + __typeName + "{__tag: " + __tagName + ", __payload: []any{" + ____rune_private_8ddf8596_emitGoParamNames(__member.__params, 0, "") + "}}\n}\n"
	}()
}

func ____rune_private_8ddf8596_emitGoParamNames(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoParamNames(__params, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name))
	}()
}

func ____rune_private_8ddf8596_enumHasPayload(__members []__IREnumMember) bool {
	return ____rune_private_8ddf8596_enumHasPayloadAt(__members, 0)
}

func ____rune_private_8ddf8596_enumHasPayloadAt(__members []__IREnumMember, __index int) bool {
	return func() bool {
		if __index >= len(__members) {
			return false
		}
		return func() bool {
			if len(__members[__index].__params) > 0 {
				return true
			}
			return ____rune_private_8ddf8596_enumHasPayloadAt(__members, __index+1)
		}()
	}()
}

func ____rune_private_8ddf8596_emitGoEnumMembers(__enumName string, __members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoEnumMembers(__enumName, __members, __index+1, __out+"\t"+__mangleIdent(__enumName+"_"+__members[__index].__name)+" "+__mangleIdent(__enumName)+" = "+__enumValue(__members[__index], __index)+"\n")
	}()
}

func ____rune_private_8ddf8596_emitGoStruct(__typeDecl __IRStructType) string {
	__out := "type " + __mangleIdent(__typeDecl.__name) + ____rune_private_8ddf8596_emitGoGenericsDecl(__typeDecl.__generics) + " struct {\n"
	for _, __field := range __typeDecl.__fields {
		_ = __field
		__out = __out + "\t" + __mangleIdent(__field.__name) + " " + ____rune_private_8ddf8596_goType(__field.__typeName) + "\n"
	}
	return __out + "}\n"
}

func ____rune_private_8ddf8596_emitGoMethods(__file __IRFile, __typeDecl __IRStructType) string {
	__out := ""
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_8ddf8596_emitGoFunction(__file, __method, __typeDecl.__name) + "\n"
	}
	return __out
}

func ____rune_private_8ddf8596_emitGoEnumMethods(__file __IRFile, __enumDecl __IREnumType) string {
	__out := ""
	for _, __method := range __enumDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_8ddf8596_emitGoFunction(__file, __method, __enumDecl.__name) + "\n"
	}
	return __out
}

func ____rune_private_8ddf8596_emitGoFunction(__file __IRFile, __fn __IRFunction, __receiverType string) string {
	__returnType := ____rune_private_8ddf8596_inferredGoReturnType(__fn)
	__params := ____rune_private_8ddf8596_emitGoFunctionParams(__fn, 0, "")
	__ret := func() string {
		if __returnsValue(__returnType) {
			return " " + ____rune_private_8ddf8596_goType(__returnType)
		}
		return ""
	}()
	__receiver := func() string {
		if __receiverType == "" {
			return ""
		}
		return "(" + __mangleIdent("this") + " " + __mangleIdent(__receiverType) + ") "
	}()
	__name := func() string {
		if __receiverType == "" {
			return __mangleIdent(__fn.__name)
		}
		return __mangleIdent(__fn.__name)
	}()
	__out := "func " + __receiver + __name + "(" + __params + ")" + __ret + " {\n"
	__out = __out + ____rune_private_8ddf8596_emitGoBody(__file, __fn.__body, __returnsValue(__returnType), __returnType, 1)
	return __out + "}\n"
}

func ____rune_private_8ddf8596_inferredGoReturnType(__fn __IRFunction) string {
	return func() string {
		if __fn.__returnType != "" {
			return __fn.__returnType
		}
		return func() string {
			if __fn.__body.__kind == __ExprKind_PatternBlock {
				return "Int"
			}
			return ""
		}()
	}()
}

func ____rune_private_8ddf8596_emitGoFunctionParams(__fn __IRFunction, __index int, __out string) string {
	return func() string {
		if __index >= len(__fn.__params) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoFunctionParams(__fn, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__fn.__params[__index].__name)+" "+____rune_private_8ddf8596_goType(____rune_private_8ddf8596_inferredGoParamType(__fn, __fn.__params[__index])))
	}()
}

func ____rune_private_8ddf8596_inferredGoParamType(__fn __IRFunction, __param __IRParam) string {
	return func() string {
		if __param.__typeName != "" {
			return __param.__typeName
		}
		return func() string {
			if __fn.__body.__kind == __ExprKind_PatternBlock && len(__fn.__params) == 1 {
				return "Int"
			}
			return ""
		}()
	}()
}

func ____rune_private_8ddf8596_emitGoBody(__file __IRFile, __expr __IRExpr, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_8ddf8596_emitGoBlock(__file, __expr.__children, 0, __returns, __returnType, __level, "")
		case __expr.__kind == __ExprKind_PatternBlock:
			return ____rune_private_8ddf8596_emitGoPatternBlock(__file, __expr, __returns, __returnType, __level)
		default:
			return __line(__level, func() string {
				if __returns {
					return "return " + ____rune_private_8ddf8596_emitGoExprExpected(__expr, __returnType)
				}
				return ____rune_private_8ddf8596_emitGoExpr(__expr)
			}())
		}
	}()
}

func ____rune_private_8ddf8596_emitGoPatternBlock(__file __IRFile, __expr __IRExpr, __returns bool, __returnType string, __level int) string {
	return ____rune_private_8ddf8596_emitGoPatternBranches(__file, __expr.__children, 0, __returns, __returnType, __level, "")
}

func ____rune_private_8ddf8596_emitGoPatternBranches(__file __IRFile, __branches []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__branches) {
			return __out + __line(__level, "}") + func() string {
				if __returns {
					return __line(__level, "return "+____rune_private_8ddf8596_goZero(__returnType))
				}
				return ""
			}()
		}
		return ____rune_private_8ddf8596_emitGoPatternBranches(__file, __branches, __index+1, __returns, __returnType, __level, __out+____rune_private_8ddf8596_emitGoPatternBranch(__file, __branches[__index], __returns, __returnType, __level, __index == 0))
	}()
}

func ____rune_private_8ddf8596_emitGoPatternBranch(__file __IRFile, __branch __IRExpr, __returns bool, __returnType string, __level int, __first bool) string {
	__pattern := __branch.__children[0]
	__value := __branch.__children[1]
	__head := func() string {
		if __first {
			return "if "
		}
		return "else if "
	}()
	return func() string {
		if __pattern.__text == "_" {
			return __line(__level, ____rune_private_8ddf8596_emitGoPatternPrefix(__first)+"else {") + ____rune_private_8ddf8596_emitGoPatternBranchBody(__file, __value, __returns, __returnType, __level+1)
		}
		return __line(__level, ____rune_private_8ddf8596_emitGoPatternPrefix(__first)+__head+____rune_private_8ddf8596_emitGoPatternCondition(__pattern)+" {") + ____rune_private_8ddf8596_emitGoPatternBranchBody(__file, __value, __returns, __returnType, __level+1)
	}()
}

func ____rune_private_8ddf8596_emitGoPatternPrefix(__first bool) string {
	return func() string {
		if __first {
			return ""
		}
		return "} "
	}()
}

func ____rune_private_8ddf8596_emitGoPatternBranchBody(__file __IRFile, __value __IRExpr, __returns bool, __returnType string, __level int) string {
	return func() string {
		if __returns {
			return __line(__level, "return "+____rune_private_8ddf8596_emitGoExprExpected(__value, __returnType))
		}
		return __line(__level, ____rune_private_8ddf8596_emitGoExpr(__value))
	}()
}

func ____rune_private_8ddf8596_emitGoPatternCondition(__pattern __IRExpr) string {
	return "__n == " + __pattern.__text
}

func ____rune_private_8ddf8596_emitGoBlock(__file __IRFile, __statements []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__statements) {
			return func() string {
				if __returns && len(__statements) == 0 {
					return __out + __line(__level, "return "+____rune_private_8ddf8596_goZero(__returnType))
				}
				return __out
			}()
		}
		return ____rune_private_8ddf8596_emitGoBlock(__file, __statements, __index+1, __returns, __returnType, __level, __out+____rune_private_8ddf8596_emitGoStatement(__file, __statements[__index], __index == len(__statements)-1, __returns, __returnType, __level))
	}()
}

func ____rune_private_8ddf8596_emitGoStatement(__file __IRFile, __expr __IRExpr, __last bool, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Let:
			return ____rune_private_8ddf8596_emitGoLet(__file, __expr, __level)
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_8ddf8596_emitGoObjectDestructure(__expr, __level)
		default:
			return func() string {
				if __last && __returns {
					return __line(__level, "return "+____rune_private_8ddf8596_emitGoExprExpected(__expr, __returnType))
				}
				return __line(__level, ____rune_private_8ddf8596_emitGoExpr(__expr))
			}()
		}
	}()
}

func ____rune_private_8ddf8596_emitGoLet(__file __IRFile, __expr __IRExpr, __level int) string {
	__value := ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0])
	__payload := ____rune_private_8ddf8596_unwrapPayloadType(__file, __expr.__children[0])
	if __payload != "" {
		__value = __value + ".(" + ____rune_private_8ddf8596_goType(__payload) + ")"
	}
	return __line(__level, __mangleIdent(__expr.__name)+" := "+__value) + __line(__level, "_ = "+__mangleIdent(__expr.__name))
}

func ____rune_private_8ddf8596_emitGoObjectDestructure(__expr __IRExpr, __level int) string {
	__source := ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0])
	__out := ""
	for _, __param := range __expr.__params {
		_ = __param
		__out = __out + __line(__level, __mangleIdent(__param.__name)+" := "+__source+"."+__mangleIdent(__param.__typeName)) + __line(__level, "_ = "+__mangleIdent(__param.__name))
	}
	return __out
}

func ____rune_private_8ddf8596_emitGoParams(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoParams(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name)+" "+____rune_private_8ddf8596_goType(__params[__index].__typeName))
	}()
}

func ____rune_private_8ddf8596_emitGoExpr(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Identifier:
			return __mangleIdent(__expr.__name)
		case __expr.__kind == __ExprKind_At:
			return __expr.__name
		case __expr.__kind == __ExprKind_This:
			return __mangleIdent("this")
		case __expr.__kind == __ExprKind_Int:
			return __expr.__value
		case __expr.__kind == __ExprKind_Double:
			return __expr.__value
		case __expr.__kind == __ExprKind_BigInt:
			return "int64(" + __bigintLiteralDigits(__expr.__value) + ")"
		case __expr.__kind == __ExprKind_String:
			return __expr.__value
		case __expr.__kind == __ExprKind_Template:
			return "\"\""
		case __expr.__kind == __ExprKind_Char:
			return __expr.__value
		case __expr.__kind == __ExprKind_Regex:
			return __expr.__value
		case __expr.__kind == __ExprKind_Bool:
			return __expr.__value
		case __expr.__kind == __ExprKind_Null:
			return "nil"
		case __expr.__kind == __ExprKind_Unary:
			return __expr.__op + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Postfix:
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + __expr.__op
		case __expr.__kind == __ExprKind_CompileTime:
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Unwrap:
			return "__runeUnwrap(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + ")"
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_8ddf8596_emitGoBinary(__expr)
		case __expr.__kind == __ExprKind_Ternary:
			return ____rune_private_8ddf8596_emitGoTernary(__expr)
		case __expr.__kind == __ExprKind_Assign:
			return ____rune_private_8ddf8596_emitGoAssign(__expr)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_8ddf8596_emitGoCall(__expr)
		case __expr.__kind == __ExprKind_Lambda:
			return ____rune_private_8ddf8596_emitGoLambda(__expr)
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_8ddf8596_emitGoSelector(__expr)
		case __expr.__kind == __ExprKind_Index:
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + "[" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "]"
		case __expr.__kind == __ExprKind_Array:
			return "[]any{" + ____rune_private_8ddf8596_emitGoExprList(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Tuple:
			return "[]any{" + ____rune_private_8ddf8596_emitGoExprList(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Map:
			return "map[any]any{" + ____rune_private_8ddf8596_emitGoMapEntries(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Spread:
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Reactive:
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Struct:
			return __mangleIdent(__expr.__name) + "{" + ____rune_private_8ddf8596_emitGoFields(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Object:
			return "struct{}{}"
		case __expr.__kind == __ExprKind_Block:
			return "func() any {\n" + ____rune_private_8ddf8596_emitGoBlockNoContext(__expr.__children, 0, true, "Dynamic", 1, "") + "}()"
		default:
			return "nil"
		}
	}()
}

func ____rune_private_8ddf8596_emitGoExprExpected(__expr __IRExpr, __expected string) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_8ddf8596_emitGoCallExpected(__expr, __expected)
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_8ddf8596_emitGoBinaryExpected(__expr, __expected)
		default:
			return ____rune_private_8ddf8596_emitGoExpr(__expr)
		}
	}()
}

func ____rune_private_8ddf8596_emitGoCallExpected(__expr __IRExpr, __expected string) string {
	__args := __genericInner(__expected, "Result")
	return func() string {
		if __args != "" && ____rune_private_8ddf8596_isResultConstructorCall(__expr) {
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + "[" + ____rune_private_8ddf8596_emitGoTypeArgs(__args) + "](" + ____rune_private_8ddf8596_emitGoExprListFrom(__expr.__children, 1, "") + ")"
		}
		return ____rune_private_8ddf8596_emitGoCall(__expr)
	}()
}

func ____rune_private_8ddf8596_isResultConstructorCall(__expr __IRExpr) bool {
	return __expr.__kind == __ExprKind_Call && len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier && (__expr.__children[0].__name == "Ok" || __expr.__children[0].__name == "Err")
}

func ____rune_private_8ddf8596_emitGoTypeArgs(__args string) string {
	return ____rune_private_8ddf8596_emitGoTypeArgList(func() []string { parts := strings.Split(__args, ","); return parts }(), 0, "")
}

func ____rune_private_8ddf8596_emitGoTypeArgList(__args []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoTypeArgList(__args, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_8ddf8596_goType(strings.TrimSpace(__args[__index])))
	}()
}

func ____rune_private_8ddf8596_emitGoBlockNoContext(__statements []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__statements) {
			return func() string {
				if __returns && len(__statements) == 0 {
					return __out + __line(__level, "return "+____rune_private_8ddf8596_goZero(__returnType))
				}
				return __out
			}()
		}
		return ____rune_private_8ddf8596_emitGoBlockNoContext(__statements, __index+1, __returns, __returnType, __level, __out+____rune_private_8ddf8596_emitGoStatementNoContext(__statements[__index], __index == len(__statements)-1, __returns, __level))
	}()
}

func ____rune_private_8ddf8596_emitGoStatementNoContext(__expr __IRExpr, __last bool, __returns bool, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Let:
			return __line(__level, __mangleIdent(__expr.__name)+" := "+____rune_private_8ddf8596_emitGoExpr(__expr.__children[0])) + __line(__level, "_ = "+__mangleIdent(__expr.__name))
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_8ddf8596_emitGoObjectDestructure(__expr, __level)
		default:
			return func() string {
				if __last && __returns {
					return __line(__level, "return "+____rune_private_8ddf8596_emitGoExpr(__expr))
				}
				return __line(__level, ____rune_private_8ddf8596_emitGoExpr(__expr))
			}()
		}
	}()
}

func ____rune_private_8ddf8596_emitGoBinary(__expr __IRExpr) string {
	return func() string {
		if __expr.__op == "??" {
			return ____rune_private_8ddf8596_emitGoNullCoalesce(__expr, __expr.__text)
		}
		return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + " " + ____rune_private_8ddf8596_goBinaryOp(__expr.__op) + " " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1])
	}()
}

func ____rune_private_8ddf8596_emitGoBinaryExpected(__expr __IRExpr, __expected string) string {
	return func() string {
		if __expr.__op == "??" {
			return ____rune_private_8ddf8596_emitGoNullCoalesce(__expr, __expected)
		}
		return ____rune_private_8ddf8596_emitGoBinary(__expr)
	}()
}

func ____rune_private_8ddf8596_emitGoNullCoalesce(__expr __IRExpr, __expected string) string {
	__resultType := ____rune_private_8ddf8596_goCoalesceResultType(__expr, __expected)
	return func() string {
		if len(__expr.__children) < 2 || __expr.__children[0].__kind == __ExprKind_Null {
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1])
		}
		return func() string {
			if __resultType == "" {
				return "func() any { __coalesce := " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + "; if __coalesce != nil { return __coalesce }; return " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + " }()"
			}
			return "func() " + ____rune_private_8ddf8596_goType(__resultType) + " { __coalesce := " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + "; if __coalesce != nil { return __coalesce.(" + ____rune_private_8ddf8596_goType(__resultType) + ") }; return " + ____rune_private_8ddf8596_emitGoExprExpected(__expr.__children[1], __resultType) + " }()"
		}()
	}()
}

func ____rune_private_8ddf8596_goCoalesceResultType(__expr __IRExpr, __expected string) string {
	__candidate := func() string {
		if __expected != "" && __expected != "Dynamic" {
			return __expected
		}
		return __expr.__text
	}()
	return func() string {
		if strings.HasSuffix(__candidate, "?") {
			return func() string { runes := []rune(__candidate); return string(runes[0 : len([]rune(__candidate))-1]) }()
		}
		return func() string {
			if __candidate == "Null" {
				return ""
			}
			return __candidate
		}()
	}()
}

func ____rune_private_8ddf8596_emitGoTernary(__expr __IRExpr) string {
	return "func() any { if " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + " { return " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + " }; return " + func() string {
		if len(__expr.__children) > 2 {
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2])
		}
		return "nil"
	}() + " }()"
}

func ____rune_private_8ddf8596_emitGoAssign(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) == 2 {
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + " = " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1])
		}
		return __mangleIdent(__expr.__name) + " = " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0])
	}()
}

func ____rune_private_8ddf8596_emitGoCall(__expr __IRExpr) string {
	return func() string {
		switch {
		case __moduleCallKey(__expr) == "io.println":
			return "fmt.Println(" + ____rune_private_8ddf8596_emitGoExprListFrom(__expr.__children, 1, "") + ")"
		case __moduleCallKey(__expr) == "map.new":
			return "map[any]any{}"
		case __moduleCallKey(__expr) == "set.new":
			return "map[any]struct{}{}"
		case __moduleCallKey(__expr) == "path.isAbsolute":
			return "strings.HasPrefix(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ", \"/\")"
		case __moduleCallKey(__expr) == "path.basename":
			return "__runePathBasename(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.extname":
			return "__runePathExtname(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.dirname":
			return "__runePathDirname(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.join":
			return "__runePathJoin(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.normalize":
			return "__runePathNormalize(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.resolve":
			return "__runePathResolve(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.relative":
			return "__runePathRelative(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.joinParts":
			return "__runePathJoinParts(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.appendPathPart":
			return "__runePathAppendPart(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.normalizeParts":
			return "__runePathNormalizeParts(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[4]) + "))"
		case __moduleCallKey(__expr) == "path.normalizePart":
			return "__runePathNormalizeParts(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[4]) + "))"
		case __moduleCallKey(__expr) == "path.normalizeParent":
			return "__runePathNormalizeParent(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[4]) + "))"
		case __moduleCallKey(__expr) == "path.normalizePop":
			return "__runePathNormalizePop(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[4]) + "))"
		case __moduleCallKey(__expr) == "path.normalizePush":
			return "__runePathNormalizePush(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[4]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[5]) + ")"
		case __moduleCallKey(__expr) == "path.pathParts":
			return "__runePathParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.collectPathParts":
			return "__runePathCollectParts(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ", __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + "))"
		case __moduleCallKey(__expr) == "path.collectPathPart":
			return "__runePathCollectPart(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + ", __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + "))"
		case __moduleCallKey(__expr) == "path.relativeFromParts":
			return "__runePathRelativeFromParts(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.relativeTail":
			return "__runePathRelativeTail(__runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "), __runePathStringParts(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + "), " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[3]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[4]) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[5]) + ")"
		case __moduleCallKey(__expr) == "process.platform":
			return "runtime.GOOS"
		case __moduleCallKey(__expr) == "process.cwd":
			return "\".\""
		case __moduleCallKey(__expr) == "process.env":
			return "(*string)(nil)"
		case __moduleCallKey(__expr) == "process.argv":
			return "append([]string(nil), os.Args...)"
		case __moduleCallKey(__expr) == "process.exit":
			return "func() struct{} { os.Exit(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "); return struct{}{} }()"
		case __moduleCallKey(__expr) == "int.toString":
			return "strconv.Itoa(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int.toDouble":
			return "float64(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int.toBigInt":
			return "int64(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int4.fromInt":
			return "func() int8 { n := (" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ") & 0xf; if n >= 8 { return int8(n - 16) }; return int8(n) }()"
		case __moduleCallKey(__expr) == "int8.fromInt":
			return "func() int8 { n := int(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "); return int8(n) }()"
		case __moduleCallKey(__expr) == "int16.fromInt":
			return "func() int16 { n := int(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "); return int16(n) }()"
		case __moduleCallKey(__expr) == "int64.fromInt":
			return "int64(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "uint.fromInt":
			return "uint(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "uint8.fromInt":
			return "func() uint8 { n := int(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "); return uint8(n) }()"
		case __moduleCallKey(__expr) == "uint16.fromInt":
			return "func() uint16 { n := int(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "); return uint16(n) }()"
		case __moduleCallKey(__expr) == "uint64.fromInt":
			return "uint64(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "float.fromDouble":
			return "float32(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case (__moduleCallKey(__expr) == "int4.toInt") || (__moduleCallKey(__expr) == "int8.toInt") || (__moduleCallKey(__expr) == "int16.toInt") || (__moduleCallKey(__expr) == "int64.toInt") || (__moduleCallKey(__expr) == "uint.toInt") || (__moduleCallKey(__expr) == "uint8.toInt") || (__moduleCallKey(__expr) == "uint16.toInt") || (__moduleCallKey(__expr) == "uint64.toInt"):
			return "int(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "float.toDouble":
			return "float64(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.fromInt":
			return "int64(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.toString":
			return "strconv.FormatInt(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ", 10)"
		case __moduleCallKey(__expr) == "bigint.toDouble":
			return "float64(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.trunc":
			return "int(math.Trunc(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "))"
		case __moduleCallKey(__expr) == "double.floor":
			return "int(math.Floor(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "))"
		case __moduleCallKey(__expr) == "double.ceil":
			return "int(math.Ceil(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "))"
		case __moduleCallKey(__expr) == "double.round":
			return "int(math.Round(" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "))"
		default:
			return ____rune_private_8ddf8596_emitGoMaybeCoreMethodCall(__expr)
		}
	}()
}

func ____rune_private_8ddf8596_fileUsesDoubleMath(__file __IRFile) bool {
	return __fileUsesModuleCall(__file, "double.trunc") || __fileUsesModuleCall(__file, "double.floor") || __fileUsesModuleCall(__file, "double.ceil") || __fileUsesModuleCall(__file, "double.round")
}

func ____rune_private_8ddf8596_fileUsesGoStrings(__file __IRFile) bool {
	return __fileUsesModuleCall(__file, "path.isAbsolute") || __fileUsesPathFamily(__file)
}

func ____rune_private_8ddf8596_emitGoMaybeCoreMethodCall(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Selector {
			return ____rune_private_8ddf8596_emitGoCoreMethodCall(__expr, __expr.__children[0])
		}
		return ____rune_private_8ddf8596_emitGoDefaultCall(__expr)
	}()
}

func ____rune_private_8ddf8596_emitGoCoreMethodCall(__expr __IRExpr, __selector __IRExpr) string {
	return func() string {
		if len(__selector.__children) > 0 && __selector.__children[0].__kind != __ExprKind_At {
			return func() string {
				switch {
				case (__selector.__name == "length") || (__selector.__name == "byteLength"):
					return ____rune_private_8ddf8596_emitGoCoreLength(__selector.__children[0])
				case __selector.__name == "isEmpty":
					return "(" + ____rune_private_8ddf8596_emitGoCoreLength(__selector.__children[0]) + ") == 0"
				case __selector.__name == "at":
					return ____rune_private_8ddf8596_emitGoCoreAt(__expr, __selector.__children[0])
				case __selector.__name == "slice":
					return ____rune_private_8ddf8596_emitGoCoreSlice(__expr, __selector.__children[0])
				case __selector.__name == "push":
					return ____rune_private_8ddf8596_emitGoCorePush(__expr, __selector.__children[0])
				case __selector.__name == "set":
					return ____rune_private_8ddf8596_emitGoCoreSet(__expr, __selector.__children[0])
				case __selector.__name == "getOr":
					return ____rune_private_8ddf8596_emitGoCoreGetOr(__expr, __selector.__children[0])
				default:
					return ____rune_private_8ddf8596_emitGoDefaultCall(__expr)
				}
			}()
		}
		return ____rune_private_8ddf8596_emitGoDefaultCall(__expr)
	}()
}

func ____rune_private_8ddf8596_emitGoCoreLength(__receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "len([]rune(" + ____rune_private_8ddf8596_emitGoExpr(__receiver) + "))"
		}
		return "len(" + ____rune_private_8ddf8596_emitGoExpr(__receiver) + ")"
	}()
}

func ____rune_private_8ddf8596_emitGoCoreAt(__expr __IRExpr, __receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "[]rune(" + ____rune_private_8ddf8596_emitGoExpr(__receiver) + ")[" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "]"
		}
		return ____rune_private_8ddf8596_emitGoExpr(__receiver) + "[" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "]"
	}()
}

func ____rune_private_8ddf8596_emitGoCoreSlice(__expr __IRExpr, __receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "func() string { runes := []rune(" + ____rune_private_8ddf8596_emitGoExpr(__receiver) + "); return string(runes[" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ":" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + "]) }()"
		}
		return ____rune_private_8ddf8596_emitGoExpr(__receiver) + "[" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + ":" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + "]"
	}()
}

func ____rune_private_8ddf8596_emitGoCorePush(__expr __IRExpr, __receiver __IRExpr) string {
	return "func() int { " + ____rune_private_8ddf8596_emitGoExpr(__receiver) + " = append(" + ____rune_private_8ddf8596_emitGoExpr(__receiver) + ", " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "); return len(" + ____rune_private_8ddf8596_emitGoExpr(__receiver) + ") }()"
}

func ____rune_private_8ddf8596_emitGoCoreSet(__expr __IRExpr, __receiver __IRExpr) string {
	return "func() any { " + ____rune_private_8ddf8596_emitGoExpr(__receiver) + "[" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "] = " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + "; return " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + " }()"
}

func ____rune_private_8ddf8596_emitGoCoreGetOr(__expr __IRExpr, __receiver __IRExpr) string {
	__resultType := ____rune_private_8ddf8596_goCoalesceResultType(__expr, __expr.__text)
	return func() string {
		if __resultType == "" {
			return "func() any { if __value, ok := " + ____rune_private_8ddf8596_emitGoExpr(__receiver) + "[" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "]; ok { return __value }; return " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[2]) + " }()"
		}
		return "func() " + ____rune_private_8ddf8596_goType(__resultType) + " { if __value, ok := " + ____rune_private_8ddf8596_emitGoExpr(__receiver) + "[" + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[1]) + "]; ok { return __value.(" + ____rune_private_8ddf8596_goType(__resultType) + ") }; return " + ____rune_private_8ddf8596_emitGoExprExpected(__expr.__children[2], __resultType) + " }()"
	}()
}

func ____rune_private_8ddf8596_emitGoDefaultCall(__expr __IRExpr) string {
	return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + "(" + ____rune_private_8ddf8596_emitGoExprListFrom(__expr.__children, 1, "") + ")"
}

func ____rune_private_8ddf8596_emitGoLambda(__expr __IRExpr) string {
	return "func(" + ____rune_private_8ddf8596_emitGoParams(__expr.__params, 0, "") + ") any { return " + ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + " }"
}

func ____rune_private_8ddf8596_emitGoSelector(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__children[0].__kind == __ExprKind_At:
			return __expr.__children[0].__name + "." + __expr.__name
		case __expr.__children[0].__kind == __ExprKind_Identifier:
			return func() string {
				if ____rune_private_8ddf8596_looksLikeTypeName(__expr.__children[0].__name) {
					return __mangleIdent(__expr.__children[0].__name + "_" + __expr.__name)
				}
				return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + "." + __mangleIdent(__expr.__name)
			}()
		default:
			return ____rune_private_8ddf8596_emitGoExpr(__expr.__children[0]) + "." + __mangleIdent(__expr.__name)
		}
	}()
}

func ____rune_private_8ddf8596_emitGoExprList(__exprs []__IRExpr, __index int, __out string) string {
	return ____rune_private_8ddf8596_emitGoExprListFrom(__exprs, __index, __out)
}

func ____rune_private_8ddf8596_emitGoExprListFrom(__exprs []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__exprs) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoExprListFrom(__exprs, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_8ddf8596_emitGoExpr(__exprs[__index]))
	}()
}

func ____rune_private_8ddf8596_emitGoMapEntries(__entries []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__entries) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoMapEntries(__entries, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_8ddf8596_emitGoExpr(__entries[__index].__children[0])+": "+____rune_private_8ddf8596_emitGoExpr(__entries[__index].__children[1]))
	}()
}

func ____rune_private_8ddf8596_emitGoFields(__fields []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__fields) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoFields(__fields, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+__mangleIdent(__fields[__index].__name)+": "+____rune_private_8ddf8596_emitGoStructFieldValue(__fields[__index].__children[0]))
	}()
}

func ____rune_private_8ddf8596_emitGoStructFieldValue(__expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_Array && len(__expr.__children) == 0 {
			return "nil"
		}
		return ____rune_private_8ddf8596_emitGoExpr(__expr)
	}()
}

func ____rune_private_8ddf8596_goBinaryOp(__op string) string {
	return __op
}

func ____rune_private_8ddf8596_hasMain(__file __IRFile) bool {
	return ____rune_private_8ddf8596_hasFunction(__file.__functions, "main", 0)
}

func ____rune_private_8ddf8596_hasFunction(__functions []__IRFunction, __name string, __index int) bool {
	return func() bool {
		if __index >= len(__functions) {
			return false
		}
		return func() bool {
			if __functions[__index].__macro == false && __functions[__index].__name == __name {
				return true
			}
			return ____rune_private_8ddf8596_hasFunction(__functions, __name, __index+1)
		}()
	}()
}

func ____rune_private_8ddf8596_usesPrintFile(__file __IRFile) bool {
	return ____rune_private_8ddf8596_functionsUsePrint(__file.__functions, 0) || ____rune_private_8ddf8596_structsUsePrint(__file.__structs, 0) || ____rune_private_8ddf8596_enumsUsePrint(__file.__enums, 0)
}

func ____rune_private_8ddf8596_functionsUsePrint(__functions []__IRFunction, __index int) bool {
	return func() bool {
		if __index >= len(__functions) {
			return false
		}
		return func() bool {
			if __functions[__index].__macro {
				return ____rune_private_8ddf8596_functionsUsePrint(__functions, __index+1)
			}
			return func() bool {
				if ____rune_private_8ddf8596_exprUsesPrint(__functions[__index].__body) {
					return true
				}
				return ____rune_private_8ddf8596_functionsUsePrint(__functions, __index+1)
			}()
		}()
	}()
}

func ____rune_private_8ddf8596_structsUsePrint(__structs []__IRStructType, __index int) bool {
	return func() bool {
		if __index >= len(__structs) {
			return false
		}
		return func() bool {
			if ____rune_private_8ddf8596_functionsUsePrint(__structs[__index].__methods, 0) {
				return true
			}
			return ____rune_private_8ddf8596_structsUsePrint(__structs, __index+1)
		}()
	}()
}

func ____rune_private_8ddf8596_enumsUsePrint(__enums []__IREnumType, __index int) bool {
	return func() bool {
		if __index >= len(__enums) {
			return false
		}
		return func() bool {
			if ____rune_private_8ddf8596_functionsUsePrint(__enums[__index].__methods, 0) {
				return true
			}
			return ____rune_private_8ddf8596_enumsUsePrint(__enums, __index+1)
		}()
	}()
}

func ____rune_private_8ddf8596_exprUsesPrint(__expr __IRExpr) bool {
	return func() bool {
		switch {
		case __moduleCallKey(__expr) == "io.println":
			return true
		default:
			return ____rune_private_8ddf8596_exprChildrenUsePrint(__expr.__children, 0)
		}
	}()
}

func ____rune_private_8ddf8596_exprChildrenUsePrint(__children []__IRExpr, __index int) bool {
	return func() bool {
		if __index >= len(__children) {
			return false
		}
		return func() bool {
			if ____rune_private_8ddf8596_exprUsesPrint(__children[__index]) {
				return true
			}
			return ____rune_private_8ddf8596_exprChildrenUsePrint(__children, __index+1)
		}()
	}()
}

func ____rune_private_8ddf8596_goType(__typeName string) string {
	return func() string {
		switch {
		case __typeName == "":
			return "any"
		case __typeName == "Void":
			return "struct{}"
		case __typeName == "Int":
			return "int"
		case (__typeName == "Int4") || (__typeName == "Int8"):
			return "int8"
		case __typeName == "Int16":
			return "int16"
		case __typeName == "Int64":
			return "int64"
		case __typeName == "BigInt":
			return "int64"
		case __typeName == "UInt":
			return "uint"
		case __typeName == "UInt8":
			return "uint8"
		case __typeName == "UInt16":
			return "uint16"
		case __typeName == "UInt64":
			return "uint64"
		case __typeName == "Double":
			return "float64"
		case __typeName == "Float":
			return "float32"
		case __typeName == "String":
			return "string"
		case __typeName == "Char":
			return "rune"
		case __typeName == "Bool":
			return "bool"
		case __typeName == "Dynamic":
			return "any"
		case (__typeName == "Data") || (__typeName == "@io.Data"):
			return "[]byte"
		default:
			return ____rune_private_8ddf8596_goTypeFallback(__typeName)
		}
	}()
}

func ____rune_private_8ddf8596_goTypeFallback(__typeName string) string {
	return func() string {
		if strings.HasSuffix(__typeName, "?") {
			return "any"
		}
		return func() string {
			if __genericInner(__typeName, "Array") != "" {
				return "[]" + ____rune_private_8ddf8596_goType(__genericInner(__typeName, "Array"))
			}
			return func() string {
				if __genericInner(__typeName, "ReadonlyArray") != "" {
					return "[]" + ____rune_private_8ddf8596_goType(__genericInner(__typeName, "ReadonlyArray"))
				}
				return func() string {
					if __genericInner(__typeName, "Map") != "" {
						return "map[" + ____rune_private_8ddf8596_goType(__typeArg(__genericInner(__typeName, "Map"), 0)) + "]" + ____rune_private_8ddf8596_goType(__typeArg(__genericInner(__typeName, "Map"), 1))
					}
					return func() string {
						if __genericInner(__typeName, "Set") != "" {
							return "map[" + ____rune_private_8ddf8596_goType(__genericInner(__typeName, "Set")) + "]struct{}"
						}
						return ____rune_private_8ddf8596_goNamedType(__typeName)
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_8ddf8596_goNamedType(__typeName string) string {
	__open := strings.Index(__typeName, "[")
	return func() string {
		if __open < 0 {
			return __mangleIdent(__typeName)
		}
		return __mangleIdent(func() string { runes := []rune(__typeName); return string(runes[0:__open]) }()) + "[" + ____rune_private_8ddf8596_emitGoTypeArgs(func() string { runes := []rune(__typeName); return string(runes[__open+1 : len([]rune(__typeName))-1]) }()) + "]"
	}()
}

func ____rune_private_8ddf8596_emitGoGenericsDecl(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "[" + ____rune_private_8ddf8596_emitGoGenericDeclItems(__generics, 0, "") + "]"
	}()
}

func ____rune_private_8ddf8596_emitGoGenericDeclItems(__generics []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__generics) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoGenericDeclItems(__generics, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__generics[__index])+" any")
	}()
}

func ____rune_private_8ddf8596_emitGoGenericsUse(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "[" + ____rune_private_8ddf8596_emitGoGenericUseItems(__generics, 0, "") + "]"
	}()
}

func ____rune_private_8ddf8596_emitGoGenericUseItems(__generics []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__generics) {
			return __out
		}
		return ____rune_private_8ddf8596_emitGoGenericUseItems(__generics, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__generics[__index]))
	}()
}

func ____rune_private_8ddf8596_unwrapPayloadType(__file __IRFile, __expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_Unwrap {
			return ____rune_private_8ddf8596_resultPayloadType(____rune_private_8ddf8596_unwrapSourceType(__file, __expr.__children[0]))
		}
		return ""
	}()
}

func ____rune_private_8ddf8596_unwrapSourceType(__file __IRFile, __expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_8ddf8596_callReturnType(__file, __expr)
		default:
			return ""
		}
	}()
}

func ____rune_private_8ddf8596_callReturnType(__file __IRFile, __expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier {
			return ____rune_private_8ddf8596_functionReturnType(__file.__functions, __expr.__children[0].__name, 0)
		}
		return ""
	}()
}

func ____rune_private_8ddf8596_functionReturnType(__functions []__IRFunction, __name string, __index int) string {
	return func() string {
		if __index >= len(__functions) {
			return ""
		}
		return func() string {
			if __functions[__index].__macro == false && __functions[__index].__name == __name {
				return __functions[__index].__returnType
			}
			return ____rune_private_8ddf8596_functionReturnType(__functions, __name, __index+1)
		}()
	}()
}

func ____rune_private_8ddf8596_resultPayloadType(__typeName string) string {
	__args := __genericInner(__typeName, "Result")
	return func() string {
		if __args == "" {
			return ""
		}
		return __typeArg(__args, 0)
	}()
}

func ____rune_private_8ddf8596_goZero(__typeName string) string {
	return func() string {
		switch {
		case (__typeName == "Int") || (__typeName == "Int4") || (__typeName == "Int8") || (__typeName == "Int16") || (__typeName == "Int64") || (__typeName == "UInt") || (__typeName == "UInt8") || (__typeName == "UInt16") || (__typeName == "UInt64") || (__typeName == "BigInt"):
			return "0"
		case (__typeName == "Double") || (__typeName == "Float"):
			return "0.0"
		case __typeName == "String":
			return "\"\""
		case __typeName == "Char":
			return "'\\x00'"
		case __typeName == "Bool":
			return "false"
		default:
			return "nil"
		}
	}()
}

func ____rune_private_8ddf8596_looksLikeTypeName(__name string) bool {
	return func() bool {
		if len([]rune(__name)) > 0 {
			return []rune(__name)[0] >= 'A' && []rune(__name)[0] <= 'Z'
		}
		return false
	}()
}

func __generateMoonBit(__file __IRFile) string {
	__out := ""
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_3050a3c7_emitMoonBitEnum(__enumDecl) + "\n"
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_3050a3c7_emitMoonBitEnumMethods(__enumDecl)
	}
	if __fileUsesUnwrap(__file) {
		__out = __out + ____rune_private_3050a3c7_emitMoonBitUnwrapHelper()
	}
	if __fileUsesPathFamily(__file) {
		__out = __out + ____rune_private_3050a3c7_emitMoonBitPathHelpers()
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_3050a3c7_emitMoonBitStruct(__typeDecl) + "\n"
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_3050a3c7_emitMoonBitMethods(__typeDecl)
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__out = func() string {
			if __fn.__macro {
				return __out
			}
			return __out + ____rune_private_3050a3c7_emitMoonBitFunction(__fn, "") + "\n"
		}()
	}
	return __out
}

func ____rune_private_3050a3c7_emitMoonBitUnwrapHelper() string {
	return "fn[T, E] rune_unwrap(value : RuneResult[T, E]) -> T {\n  match value {\n    RuneOk(value) => value\n    RuneErr(_) => abort(\"Result.Err\")\n  }\n}\n\n"
}

func ____rune_private_3050a3c7_emitMoonBitPathHelpers() string {
	return "fn __rune_path_basename(path : String) -> String {\n  let index = path.rev_find(\"/\").unwrap_or(-1)\n  if index < 0 {\n    path\n  } else if index == path.length() - 1 {\n    path\n  } else {\n    path[index + 1:].to_owned()\n  }\n}\n\nfn __rune_path_extname(path : String) -> String {\n  let base = __rune_path_basename(path)\n  let index = base.rev_find(\".\").unwrap_or(-1)\n  if index <= 0 { \"\" } else { base[index:].to_owned() }\n}\n\nfn __rune_path_dirname(path : String) -> String {\n  let index = path.rev_find(\"/\").unwrap_or(-1)\n  if index < 0 {\n    \".\"\n  } else if index == 0 {\n    \"/\"\n  } else {\n    path[:index].to_owned()\n  }\n}\n\nfn __rune_path_join(parts : Array[String]) -> String {\n  __rune_path_normalize(__rune_path_join_parts(parts, 0, \"\"))\n}\n\nfn __rune_path_normalize(path : String) -> String {\n  let absolute = path.has_prefix(\"/\")\n  let pieces = path.split(\"/\").map(fn(part) { part.to_owned() }).to_array()\n  let out = __rune_path_normalize_parts(pieces, 0, absolute, [])\n  let joined = __rune_path_join_parts(out, 0, \"\")\n  if absolute {\n    \"/\" + joined\n  } else if joined.is_empty() {\n    \".\"\n  } else {\n    joined\n  }\n}\n\nfn __rune_path_resolve(parts : Array[String]) -> String {\n  if parts.length() == 0 { \".\" } else { __rune_path_normalize(__rune_path_join(parts)) }\n}\n\nfn __rune_path_relative(from : String, to : String) -> String {\n  let from_parts = __rune_path_parts(__rune_path_resolve([from]))\n  let to_parts = __rune_path_parts(__rune_path_resolve([to]))\n  __rune_path_relative_from_parts(from_parts, to_parts, 0)\n}\n\nfn __rune_path_join_parts(parts : Array[String], index : Int, out : String) -> String {\n  if index >= parts.length() {\n    out\n  } else {\n    __rune_path_join_parts(parts, index + 1, __rune_path_append_part(out, parts[index]))\n  }\n}\n\nfn __rune_path_append_part(out : String, part : String) -> String {\n  if out.is_empty() {\n    part\n  } else if part.is_empty() {\n    out\n  } else {\n    out + \"/\" + part\n  }\n}\n\nfn __rune_path_normalize_parts(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  if index >= parts.length() {\n    out\n  } else {\n    let part = parts[index]\n    if part.is_empty() || part == \".\" {\n      __rune_path_normalize_parts(parts, index + 1, absolute, out)\n    } else if part == \"..\" {\n      __rune_path_normalize_parent(parts, index, absolute, out)\n    } else {\n      __rune_path_normalize_push(parts, index, absolute, out, part)\n    }\n  }\n}\n\nfn __rune_path_normalize_parent(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  if out.length() > 0 {\n    __rune_path_normalize_pop(parts, index, absolute, out)\n  } else if absolute {\n    __rune_path_normalize_parts(parts, index + 1, absolute, out)\n  } else {\n    __rune_path_normalize_push(parts, index, absolute, out, \"..\")\n  }\n}\n\nfn __rune_path_normalize_pop(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  __rune_path_normalize_parts(parts, index + 1, absolute, out[:out.length() - 1].to_owned())\n}\n\nfn __rune_path_normalize_push(parts : Array[String], index : Int, absolute : Bool, out : Array[String], part : String) -> Array[String] {\n  __rune_path_normalize_parts(parts, index + 1, absolute, [..out, part])\n}\n\nfn __rune_path_parts(path : String) -> Array[String] {\n  let pieces = __rune_path_normalize(path).split(\"/\").map(fn(part) { part.to_owned() }).to_array()\n  __rune_path_collect_parts(pieces, 0, [])\n}\n\nfn __rune_path_collect_parts(parts : Array[String], index : Int, out : Array[String]) -> Array[String] {\n  if index >= parts.length() {\n    out\n  } else if parts[index].is_empty() {\n    __rune_path_collect_parts(parts, index + 1, out)\n  } else {\n    __rune_path_collect_parts(parts, index + 1, [..out, parts[index]])\n  }\n}\n\nfn __rune_path_collect_part(parts : Array[String], index : Int, out : Array[String]) -> Array[String] {\n  if index < parts.length() {\n    __rune_path_collect_parts(parts, index + 1, [..out, parts[index]])\n  } else {\n    out\n  }\n}\n\nfn __rune_path_relative_from_parts(from_parts : Array[String], to_parts : Array[String], index : Int) -> String {\n  if index < from_parts.length() && index < to_parts.length() && from_parts[index] == to_parts[index] {\n    __rune_path_relative_from_parts(from_parts, to_parts, index + 1)\n  } else {\n    __rune_path_relative_tail(from_parts, to_parts, index, index, \"\")\n  }\n}\n\nfn __rune_path_relative_tail(from_parts : Array[String], to_parts : Array[String], from_index : Int, to_index : Int, out : String) -> String {\n  if from_index < from_parts.length() {\n    __rune_path_relative_tail(from_parts, to_parts, from_index + 1, to_index, __rune_path_append_part(out, \"..\"))\n  } else if to_index < to_parts.length() {\n    __rune_path_relative_tail(from_parts, to_parts, from_index, to_index + 1, __rune_path_append_part(out, to_parts[to_index]))\n  } else if out.is_empty() {\n    \".\"\n  } else {\n    out\n  }\n}\n\n"
}

func ____rune_private_3050a3c7_emitMoonBitEnum(__enumDecl __IREnumType) string {
	__out := "enum " + ____rune_private_3050a3c7_moonBitTypeIdent(__enumDecl.__name) + ____rune_private_3050a3c7_emitMoonBitGenerics(__enumDecl.__generics) + " {\n"
	__out = __out + ____rune_private_3050a3c7_emitMoonBitEnumMembers(__enumDecl.__members, 0, "")
	return __out + "} derive(Eq, Show)\n"
}

func ____rune_private_3050a3c7_emitMoonBitEnumMembers(__members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_3050a3c7_emitMoonBitEnumMembers(__members, __index+1, __out+____rune_private_3050a3c7_emitMoonBitEnumMember(__members[__index]))
	}()
}

func ____rune_private_3050a3c7_emitMoonBitEnumMember(__member __IREnumMember) string {
	return func() string {
		if len(__member.__params) == 0 {
			return __line(1, ____rune_private_3050a3c7_moonBitConstructorIdent(__member.__name))
		}
		return __line(1, ____rune_private_3050a3c7_moonBitConstructorIdent(__member.__name)+"("+____rune_private_3050a3c7_emitMoonBitParamTypes(__member.__params, 0, "")+")")
	}()
}

func ____rune_private_3050a3c7_emitMoonBitStruct(__typeDecl __IRStructType) string {
	__out := "struct " + ____rune_private_3050a3c7_moonBitTypeIdent(__typeDecl.__name) + " {\n"
	for _, __field := range __typeDecl.__fields {
		_ = __field
		__out = __out + __line(1, __mangleIdent(__field.__name)+" : "+____rune_private_3050a3c7_moonBitType(__field.__typeName))
	}
	return __out + "}\n"
}

func ____rune_private_3050a3c7_emitMoonBitMethods(__typeDecl __IRStructType) string {
	__out := ""
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_3050a3c7_emitMoonBitFunction(____rune_private_3050a3c7_methodWithMoonBitReceiver(__typeDecl.__name, __method), "") + "\n"
	}
	return __out
}

func ____rune_private_3050a3c7_emitMoonBitEnumMethods(__enumDecl __IREnumType) string {
	__out := ""
	for _, __method := range __enumDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_3050a3c7_emitMoonBitFunction(____rune_private_3050a3c7_methodWithMoonBitReceiver(__enumDecl.__name, __method), "") + "\n"
	}
	return __out
}

func ____rune_private_3050a3c7_methodWithMoonBitReceiver(__typeName string, __method __IRFunction) __IRFunction {
	return __IRFunction{__name: __typeName + "_" + __method.__name, __private: __method.__private, __routine: __method.__routine, __macro: __method.__macro, __receiverType: __method.__receiverType, __generics: __method.__generics, __params: ____rune_private_3050a3c7_prependMoonBitSelfParam(__typeName, __method.__params), __returnType: __method.__returnType, __body: __method.__body, __line: __method.__line, __column: __method.__column}
}

func ____rune_private_3050a3c7_prependMoonBitSelfParam(__typeName string, __params []__IRParam) []__IRParam {
	__out := []__IRParam{__IRParam{__name: "this", __typeName: __typeName, __line: 0, __column: 0}}
	for _, __param := range __params {
		_ = __param
		func() int { __out = append(__out, __param); return len(__out) }()
	}
	return __out
}

func ____rune_private_3050a3c7_emitMoonBitFunction(__fn __IRFunction, __receiverType string) string {
	__params := ____rune_private_3050a3c7_emitMoonBitParams(__fn.__params, 0, "")
	__ret := func() string {
		if __returnsValue(__fn.__returnType) {
			return " -> " + ____rune_private_3050a3c7_moonBitType(__fn.__returnType)
		}
		return ""
	}()
	__name := __mangleIdent(__fn.__name)
	__bodyReturns := __returnsValue(__fn.__returnType) && __fn.__name != "main"
	__head := func() string {
		if __fn.__name == "main" && __params == "" {
			return "fn main"
		}
		return "fn " + __name + ____rune_private_3050a3c7_emitMoonBitGenerics(__fn.__generics) + "(" + __params + ")" + __ret
	}()
	__out := __head + " {\n"
	__out = __out + ____rune_private_3050a3c7_emitMoonBitBody(__fn.__body, __bodyReturns, __fn.__returnType, 1)
	return __out + "}\n"
}

func ____rune_private_3050a3c7_emitMoonBitBody(__expr __IRExpr, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_3050a3c7_emitMoonBitBlock(__expr.__children, 0, __returns, __returnType, __level, "")
		default:
			return __line(__level, func() string {
				if __returns {
					return ____rune_private_3050a3c7_emitMoonBitExpr(__expr)
				}
				return ____rune_private_3050a3c7_emitMoonBitDiscard(__expr)
			}())
		}
	}()
}

func ____rune_private_3050a3c7_emitMoonBitBlock(__statements []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__statements) {
			return func() string {
				if __returns && len(__statements) == 0 {
					return __out + __line(__level, ____rune_private_3050a3c7_moonBitZero(__returnType))
				}
				return __out
			}()
		}
		return ____rune_private_3050a3c7_emitMoonBitBlock(__statements, __index+1, __returns, __returnType, __level, __out+____rune_private_3050a3c7_emitMoonBitStatement(__statements[__index], __index == len(__statements)-1, __returns, __level))
	}()
}

func ____rune_private_3050a3c7_emitMoonBitStatement(__expr __IRExpr, __last bool, __returns bool, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Let:
			return ____rune_private_3050a3c7_emitMoonBitLet(__expr, __level)
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_3050a3c7_emitMoonBitObjectDestructure(__expr, __level)
		default:
			return __line(__level, func() string {
				if __last && __returns {
					return ____rune_private_3050a3c7_emitMoonBitExpr(__expr)
				}
				return ____rune_private_3050a3c7_emitMoonBitDiscard(__expr)
			}())
		}
	}()
}

func ____rune_private_3050a3c7_emitMoonBitLet(__expr __IRExpr, __level int) string {
	return __line(__level, ____rune_private_3050a3c7_moonBitLetKeyword(__expr.__op)+__mangleIdent(__expr.__name)+" = "+____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]))
}

func ____rune_private_3050a3c7_moonBitLetKeyword(__op string) string {
	return func() string {
		switch {
		case __op == ":=:":
			return "let mut "
		default:
			return "let "
		}
	}()
}

func ____rune_private_3050a3c7_emitMoonBitObjectDestructure(__expr __IRExpr, __level int) string {
	__source := ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0])
	__tmp := __mangleIdent("__object")
	__out := __line(__level, "let "+__tmp+" = "+__source)
	for _, __param := range __expr.__params {
		_ = __param
		__out = __out + __line(__level, "let "+__mangleIdent(__param.__name)+" = "+__tmp+"."+__mangleIdent(__param.__typeName))
	}
	return __out
}

func ____rune_private_3050a3c7_emitMoonBitParams(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_3050a3c7_emitMoonBitParams(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name)+" : "+____rune_private_3050a3c7_moonBitType(__params[__index].__typeName))
	}()
}

func ____rune_private_3050a3c7_emitMoonBitParamTypes(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_3050a3c7_emitMoonBitParamTypes(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_3050a3c7_moonBitType(__params[__index].__typeName))
	}()
}

func ____rune_private_3050a3c7_emitMoonBitGenerics(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "[" + ____rune_private_3050a3c7_joinMoonBitGenericNames(__generics, 0, "") + "]"
	}()
}

func ____rune_private_3050a3c7_emitMoonBitExpr(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Identifier:
			return ____rune_private_3050a3c7_moonBitValueIdent(__expr.__name)
		case __expr.__kind == __ExprKind_At:
			return "@" + __expr.__name
		case __expr.__kind == __ExprKind_This:
			return __mangleIdent("this")
		case __expr.__kind == __ExprKind_Int:
			return __expr.__value
		case __expr.__kind == __ExprKind_Double:
			return __expr.__value
		case __expr.__kind == __ExprKind_BigInt:
			return __bigintLiteralDigits(__expr.__value)
		case __expr.__kind == __ExprKind_String:
			return __expr.__value
		case __expr.__kind == __ExprKind_Template:
			return __expr.__value
		case __expr.__kind == __ExprKind_Char:
			return __expr.__value
		case __expr.__kind == __ExprKind_Regex:
			return __expr.__value
		case __expr.__kind == __ExprKind_Bool:
			return __expr.__value
		case __expr.__kind == __ExprKind_Null:
			return "None"
		case __expr.__kind == __ExprKind_Unary:
			return ____rune_private_3050a3c7_moonBitUnaryOp(__expr.__op) + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Postfix:
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + ____rune_private_3050a3c7_moonBitPostfixOp(__expr.__op)
		case __expr.__kind == __ExprKind_CompileTime:
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Unwrap:
			return "rune_unwrap(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + ")"
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_3050a3c7_emitMoonBitBinary(__expr)
		case __expr.__kind == __ExprKind_Ternary:
			return ____rune_private_3050a3c7_emitMoonBitTernary(__expr)
		case __expr.__kind == __ExprKind_Assign:
			return ____rune_private_3050a3c7_emitMoonBitAssign(__expr)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_3050a3c7_emitMoonBitCall(__expr)
		case __expr.__kind == __ExprKind_Lambda:
			return ____rune_private_3050a3c7_emitMoonBitLambda(__expr)
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_3050a3c7_emitMoonBitSelector(__expr)
		case __expr.__kind == __ExprKind_Index:
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + "[" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + "]"
		case __expr.__kind == __ExprKind_Array:
			return "[" + ____rune_private_3050a3c7_emitMoonBitExprList(__expr.__children, 0, "") + "]"
		case __expr.__kind == __ExprKind_Tuple:
			return "(" + ____rune_private_3050a3c7_emitMoonBitExprList(__expr.__children, 0, "") + ")"
		case __expr.__kind == __ExprKind_Map:
			return "{" + ____rune_private_3050a3c7_emitMoonBitMapEntries(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Spread:
			return ".." + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Reactive:
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Struct:
			return ____rune_private_3050a3c7_moonBitTypeIdent(__expr.__name) + "::{ " + ____rune_private_3050a3c7_emitMoonBitFields(__expr.__children, 0, "") + " }"
		case __expr.__kind == __ExprKind_Object:
			return "{ " + ____rune_private_3050a3c7_emitMoonBitFields(__expr.__children, 0, "") + " }"
		case __expr.__kind == __ExprKind_Block:
			return "{\n" + ____rune_private_3050a3c7_emitMoonBitBlock(__expr.__children, 0, true, "Dynamic", 1, "") + "}"
		default:
			return "()"
		}
	}()
}

func ____rune_private_3050a3c7_emitMoonBitDiscard(__expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_Call {
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr)
		}
		return "ignore(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr) + ")"
	}()
}

func ____rune_private_3050a3c7_emitMoonBitBinary(__expr __IRExpr) string {
	return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + " " + ____rune_private_3050a3c7_moonBitBinaryOp(__expr.__op) + " " + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1])
}

func ____rune_private_3050a3c7_emitMoonBitTernary(__expr __IRExpr) string {
	return "if " + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + " { " + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + " } else { " + func() string {
		if len(__expr.__children) > 2 {
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[2])
		}
		return "()"
	}() + " }"
}

func ____rune_private_3050a3c7_emitMoonBitAssign(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) == 2 {
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + " = " + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1])
		}
		return __mangleIdent(__expr.__name) + " = " + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0])
	}()
}

func ____rune_private_3050a3c7_emitMoonBitCall(__expr __IRExpr) string {
	return func() string {
		switch {
		case __moduleCallKey(__expr) == "io.println":
			return "println(" + ____rune_private_3050a3c7_emitMoonBitPrintArgs(__expr.__children, 1, "") + ")"
		case __moduleCallKey(__expr) == "io.print":
			return "print(" + ____rune_private_3050a3c7_emitMoonBitPrintArgs(__expr.__children, 1, "") + ")"
		case __moduleCallKey(__expr) == "map.new":
			return "{}"
		case __moduleCallKey(__expr) == "set.new":
			return "Set::new()"
		case __moduleCallKey(__expr) == "path.isAbsolute":
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ".has_prefix(\"/\")"
		case __moduleCallKey(__expr) == "path.basename":
			return "__rune_path_basename(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.extname":
			return "__rune_path_extname(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.dirname":
			return "__rune_path_dirname(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.join":
			return "__rune_path_join(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.normalize":
			return "__rune_path_normalize(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.resolve":
			return "__rune_path_resolve(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.relative":
			return "__rune_path_relative(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.joinParts":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall3("__rune_path_join_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3])
		case __moduleCallKey(__expr) == "path.appendPathPart":
			return "__rune_path_append_part(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.normalizeParts":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall4("__rune_path_normalize_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4])
		case __moduleCallKey(__expr) == "path.normalizePart":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall4("__rune_path_normalize_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4])
		case __moduleCallKey(__expr) == "path.normalizeParent":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall4("__rune_path_normalize_parent", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4])
		case __moduleCallKey(__expr) == "path.normalizePop":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall4("__rune_path_normalize_pop", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4])
		case __moduleCallKey(__expr) == "path.normalizePush":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall5("__rune_path_normalize_push", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4], __expr.__children[5])
		case __moduleCallKey(__expr) == "path.pathParts":
			return "__rune_path_parts(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.collectPathParts":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall3("__rune_path_collect_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3])
		case __moduleCallKey(__expr) == "path.collectPathPart":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall3("__rune_path_collect_part", __expr.__children[1], __expr.__children[2], __expr.__children[3])
		case __moduleCallKey(__expr) == "path.relativeFromParts":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall3("__rune_path_relative_from_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3])
		case __moduleCallKey(__expr) == "path.relativeTail":
			return ____rune_private_3050a3c7_emitMoonBitRuntimeCall5("__rune_path_relative_tail", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4], __expr.__children[5])
		case __moduleCallKey(__expr) == "process.platform":
			return "\"moonbit\""
		case __moduleCallKey(__expr) == "process.cwd":
			return "\".\""
		case __moduleCallKey(__expr) == "process.env":
			return "(None : String?)"
		case __moduleCallKey(__expr) == "process.argv":
			return "([] : Array[String])"
		case __moduleCallKey(__expr) == "int.toString":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").to_string()"
		case __moduleCallKey(__expr) == "int.toDouble":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").to_double()"
		case __moduleCallKey(__expr) == "int.toBigInt":
			return "BigInt::from_int(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int4.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xf; if __n >= 8 { __n - 16 } else { __n } })(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int8.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xff; if __n >= 128 { __n - 256 } else { __n } })(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int16.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xffff; if __n >= 32768 { __n - 65536 } else { __n } })(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int64.fromInt":
			return "Int64::from_int(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case (__moduleCallKey(__expr) == "uint.fromInt") || (__moduleCallKey(__expr) == "uint8.fromInt") || (__moduleCallKey(__expr) == "uint16.fromInt"):
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").reinterpret_as_uint()"
		case __moduleCallKey(__expr) == "uint64.fromInt":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").to_uint64()"
		case (__moduleCallKey(__expr) == "int4.toInt") || (__moduleCallKey(__expr) == "int8.toInt") || (__moduleCallKey(__expr) == "int16.toInt"):
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1])
		case (__moduleCallKey(__expr) == "uint.toInt") || (__moduleCallKey(__expr) == "uint8.toInt") || (__moduleCallKey(__expr) == "uint16.toInt"):
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").reinterpret_as_int()"
		case (__moduleCallKey(__expr) == "int64.toInt") || (__moduleCallKey(__expr) == "uint64.toInt"):
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").to_int()"
		case __moduleCallKey(__expr) == "float.fromDouble":
			return "Float::from_double(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "float.toDouble":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").to_double()"
		case __moduleCallKey(__expr) == "bigint.fromInt":
			return "BigInt::from_int(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.toString":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").to_string()"
		case __moduleCallKey(__expr) == "bigint.toDouble":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").to_double()"
		case __moduleCallKey(__expr) == "double.trunc":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").trunc()"
		case __moduleCallKey(__expr) == "double.floor":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").floor()"
		case __moduleCallKey(__expr) == "double.ceil":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").ceil()"
		case __moduleCallKey(__expr) == "double.round":
			return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ").round()"
		default:
			return ____rune_private_3050a3c7_emitMoonBitMaybeCoreMethodCall(__expr)
		}
	}()
}

func ____rune_private_3050a3c7_emitMoonBitRuntimeCall3(__name string, __first __IRExpr, __second __IRExpr, __third __IRExpr) string {
	return __name + "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__first) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__second) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__third) + ")"
}

func ____rune_private_3050a3c7_emitMoonBitRuntimeCall4(__name string, __first __IRExpr, __second __IRExpr, __third __IRExpr, __fourth __IRExpr) string {
	return __name + "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__first) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__second) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__third) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__fourth) + ")"
}

func ____rune_private_3050a3c7_emitMoonBitRuntimeCall5(__name string, __first __IRExpr, __second __IRExpr, __third __IRExpr, __fourth __IRExpr, __fifth __IRExpr) string {
	return __name + "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__first) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__second) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__third) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__fourth) + ", " + ____rune_private_3050a3c7_emitMoonBitExpr(__fifth) + ")"
}

func ____rune_private_3050a3c7_emitMoonBitMaybeCoreMethodCall(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Selector {
			return ____rune_private_3050a3c7_emitMoonBitCoreMethodCall(__expr, __expr.__children[0])
		}
		return ____rune_private_3050a3c7_emitMoonBitDefaultCall(__expr)
	}()
}

func ____rune_private_3050a3c7_emitMoonBitCoreMethodCall(__expr __IRExpr, __selector __IRExpr) string {
	return func() string {
		if len(__selector.__children) > 0 && __selector.__children[0].__kind != __ExprKind_At {
			return func() string {
				switch {
				case (__selector.__name == "length") || (__selector.__name == "byteLength"):
					return ____rune_private_3050a3c7_emitMoonBitExpr(__selector.__children[0]) + ".length()"
				case __selector.__name == "isEmpty":
					return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__selector.__children[0]) + ".length() == 0)"
				case __selector.__name == "at":
					return ____rune_private_3050a3c7_emitMoonBitExpr(__selector.__children[0]) + "[" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + "]"
				case __selector.__name == "slice":
					return ____rune_private_3050a3c7_emitMoonBitExpr(__selector.__children[0]) + "[" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[1]) + ":" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[2]) + "].to_owned()"
				default:
					return ____rune_private_3050a3c7_emitMoonBitDefaultCall(__expr)
				}
			}()
		}
		return ____rune_private_3050a3c7_emitMoonBitDefaultCall(__expr)
	}()
}

func ____rune_private_3050a3c7_emitMoonBitDefaultCall(__expr __IRExpr) string {
	return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + "(" + ____rune_private_3050a3c7_emitMoonBitExprListFrom(__expr.__children, 1, "") + ")"
}

func ____rune_private_3050a3c7_emitMoonBitPrintArgs(__exprs []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__exprs) {
			return func() string {
				if __out == "" {
					return "\"\""
				}
				return __out
			}()
		}
		return ____rune_private_3050a3c7_emitMoonBitPrintArgs(__exprs, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return " + \" \" + "
		}()+____rune_private_3050a3c7_emitMoonBitShowExpr(__exprs[__index]))
	}()
}

func ____rune_private_3050a3c7_emitMoonBitShowExpr(__expr __IRExpr) string {
	return func() string {
		if __expr.__text == "String" || __expr.__kind == __ExprKind_String || __expr.__kind == __ExprKind_Template {
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr)
		}
		return "(" + ____rune_private_3050a3c7_emitMoonBitExpr(__expr) + ").to_string()"
	}()
}

func ____rune_private_3050a3c7_emitMoonBitLambda(__expr __IRExpr) string {
	return "(" + ____rune_private_3050a3c7_emitMoonBitParams(__expr.__params, 0, "") + ") => " + ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0])
}

func ____rune_private_3050a3c7_emitMoonBitSelector(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__children[0].__kind == __ExprKind_At:
			return "@" + __expr.__children[0].__name + "." + __expr.__name
		case __expr.__children[0].__kind == __ExprKind_Identifier:
			return func() string {
				if ____rune_private_3050a3c7_moonBitLooksLikeTypeName(__expr.__children[0].__name) {
					return ____rune_private_3050a3c7_moonBitTypeIdent(__expr.__children[0].__name) + "::" + ____rune_private_3050a3c7_moonBitConstructorIdent(__expr.__name)
				}
				return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + "." + __mangleIdent(__expr.__name)
			}()
		default:
			return ____rune_private_3050a3c7_emitMoonBitExpr(__expr.__children[0]) + "." + __mangleIdent(__expr.__name)
		}
	}()
}

func ____rune_private_3050a3c7_emitMoonBitExprList(__exprs []__IRExpr, __index int, __out string) string {
	return ____rune_private_3050a3c7_emitMoonBitExprListFrom(__exprs, __index, __out)
}

func ____rune_private_3050a3c7_emitMoonBitExprListFrom(__exprs []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__exprs) {
			return __out
		}
		return ____rune_private_3050a3c7_emitMoonBitExprListFrom(__exprs, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_3050a3c7_emitMoonBitExpr(__exprs[__index]))
	}()
}

func ____rune_private_3050a3c7_emitMoonBitMapEntries(__entries []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__entries) {
			return __out
		}
		return ____rune_private_3050a3c7_emitMoonBitMapEntries(__entries, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_3050a3c7_emitMoonBitExpr(__entries[__index].__children[0])+": "+____rune_private_3050a3c7_emitMoonBitExpr(__entries[__index].__children[1]))
	}()
}

func ____rune_private_3050a3c7_emitMoonBitFields(__fields []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__fields) {
			return __out
		}
		return ____rune_private_3050a3c7_emitMoonBitFields(__fields, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+__mangleIdent(__fields[__index].__name)+": "+____rune_private_3050a3c7_emitMoonBitExpr(__fields[__index].__children[0]))
	}()
}

func ____rune_private_3050a3c7_moonBitUnaryOp(__op string) string {
	return func() string {
		switch {
		case __op == "!":
			return "!"
		default:
			return __op
		}
	}()
}

func ____rune_private_3050a3c7_moonBitPostfixOp(__op string) string {
	return __op
}

func ____rune_private_3050a3c7_moonBitBinaryOp(__op string) string {
	return func() string {
		switch {
		case __op == "&&":
			return "&&"
		case __op == "||":
			return "||"
		default:
			return __op
		}
	}()
}

func ____rune_private_3050a3c7_moonBitType(__typeName string) string {
	switch {
	case (__typeName == "") || (__typeName == "Void"):
		return "Unit"
	case (__typeName == "Int") || (__typeName == "Int4") || (__typeName == "Int8") || (__typeName == "Int16"):
		return "Int"
	case (__typeName == "UInt") || (__typeName == "UInt8") || (__typeName == "UInt16"):
		return "UInt"
	case __typeName == "Int64":
		return "Int64"
	case __typeName == "UInt64":
		return "UInt64"
	case __typeName == "Double":
		return "Double"
	case __typeName == "Float":
		return "Float"
	case __typeName == "BigInt":
		return "BigInt"
	case __typeName == "String":
		return "String"
	case __typeName == "Char":
		return "Char"
	case __typeName == "Bool":
		return "Bool"
	case (__typeName == "Dynamic") || (__typeName == "Object"):
		return "Unit"
	case (__typeName == "Data") || (__typeName == "@io.Data"):
		return "Array[Int]"
	default:
		return ____rune_private_3050a3c7_moonBitTypeFallback(__typeName)
	}
}

func ____rune_private_3050a3c7_moonBitTypeFallback(__typeName string) string {
	return func() string {
		if strings.HasSuffix(__typeName, "?") {
			return ____rune_private_3050a3c7_moonBitType(func() string { runes := []rune(__typeName); return string(runes[0 : len([]rune(__typeName))-1]) }()) + "?"
		}
		return func() string {
			if __genericInner(__typeName, "Array") != "" {
				return "Array[" + ____rune_private_3050a3c7_moonBitType(__genericInner(__typeName, "Array")) + "]"
			}
			return func() string {
				if __genericInner(__typeName, "ReadonlyArray") != "" {
					return "Array[" + ____rune_private_3050a3c7_moonBitType(__genericInner(__typeName, "ReadonlyArray")) + "]"
				}
				return func() string {
					if __genericInner(__typeName, "Map") != "" {
						return "Map[" + ____rune_private_3050a3c7_moonBitType(__typeArg(__genericInner(__typeName, "Map"), 0)) + ", " + ____rune_private_3050a3c7_moonBitType(__typeArg(__genericInner(__typeName, "Map"), 1)) + "]"
					}
					return func() string {
						if __genericInner(__typeName, "Set") != "" {
							return "Set[" + ____rune_private_3050a3c7_moonBitType(__genericInner(__typeName, "Set")) + "]"
						}
						return ____rune_private_3050a3c7_moonBitNamedType(__typeName)
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_3050a3c7_moonBitNamedType(__typeName string) string {
	__open := strings.Index(__typeName, "[")
	return func() string {
		if __open < 0 {
			return ____rune_private_3050a3c7_moonBitTypeIdent(__typeName)
		}
		return ____rune_private_3050a3c7_moonBitTypeIdent(func() string { runes := []rune(__typeName); return string(runes[0:__open]) }()) + "[" + ____rune_private_3050a3c7_emitMoonBitTypeArgs(func() string { runes := []rune(__typeName); return string(runes[__open+1 : len([]rune(__typeName))-1]) }()) + "]"
	}()
}

func ____rune_private_3050a3c7_emitMoonBitTypeArgs(__args string) string {
	return ____rune_private_3050a3c7_emitMoonBitTypeArgList(func() []string { parts := strings.Split(__args, ","); return parts }(), 0, "")
}

func ____rune_private_3050a3c7_emitMoonBitTypeArgList(__args []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_3050a3c7_emitMoonBitTypeArgList(__args, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_3050a3c7_moonBitType(strings.TrimSpace(__args[__index])))
	}()
}

func ____rune_private_3050a3c7_moonBitZero(__typeName string) string {
	return func() string {
		switch {
		case (__typeName == "Int") || (__typeName == "Int4") || (__typeName == "Int8") || (__typeName == "Int16") || (__typeName == "UInt") || (__typeName == "UInt8") || (__typeName == "UInt16"):
			return "0"
		case __typeName == "Int64":
			return "0L"
		case __typeName == "UInt64":
			return "0UL"
		case (__typeName == "Double") || (__typeName == "Float"):
			return "0.0"
		case __typeName == "String":
			return "\"\""
		case __typeName == "Bool":
			return "false"
		default:
			return "()"
		}
	}()
}

func ____rune_private_3050a3c7_joinMoonBitStrings(__values []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__values) {
			return __out
		}
		return ____rune_private_3050a3c7_joinMoonBitStrings(__values, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__values[__index])
	}()
}

func ____rune_private_3050a3c7_joinMoonBitGenericNames(__values []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__values) {
			return __out
		}
		return ____rune_private_3050a3c7_joinMoonBitGenericNames(__values, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_3050a3c7_moonBitTypeParamIdent(__values[__index]))
	}()
}

func ____rune_private_3050a3c7_moonBitValueIdent(__name string) string {
	return func() string {
		if ____rune_private_3050a3c7_moonBitLooksLikeTypeName(__name) {
			return ____rune_private_3050a3c7_moonBitConstructorIdent(__name)
		}
		return __mangleIdent(__name)
	}()
}

func ____rune_private_3050a3c7_moonBitTypeIdent(__name string) string {
	return func() string {
		if ____rune_private_3050a3c7_moonBitLooksGenericParam(__name) {
			return ____rune_private_3050a3c7_moonBitTypeParamIdent(__name)
		}
		return "Rune" + ____rune_private_3050a3c7_moonBitSanitizeIdent(__name)
	}()
}

func ____rune_private_3050a3c7_moonBitConstructorIdent(__name string) string {
	return "Rune" + ____rune_private_3050a3c7_moonBitSanitizeIdent(__name)
}

func ____rune_private_3050a3c7_moonBitTypeParamIdent(__name string) string {
	return func() string {
		if ____rune_private_3050a3c7_moonBitLooksLikeTypeName(__name) {
			return __name
		}
		return "T"
	}()
}

func ____rune_private_3050a3c7_moonBitLooksGenericParam(__name string) bool {
	return len([]rune(__name)) == 1 && ____rune_private_3050a3c7_moonBitLooksLikeTypeName(__name)
}

func ____rune_private_3050a3c7_moonBitSanitizeIdent(__name string) string {
	return strings.ReplaceAll((strings.ReplaceAll((strings.ReplaceAll(__name, ".", "_")), "-", "_")), "@", "_")
}

func ____rune_private_3050a3c7_moonBitLooksLikeTypeName(__name string) bool {
	return func() bool {
		if len([]rune(__name)) > 0 {
			return ____rune_private_3050a3c7_moonBitIsUpperLetter([]rune(__name)[0])
		}
		return false
	}()
}

func ____rune_private_3050a3c7_moonBitIsUpperLetter(__ch rune) bool {
	return __ch >= 'A' && __ch <= 'Z'
}

func __generateTypeScript(__file __IRFile) string {
	__out := ____rune_private_68c6e3cf_emitTSImports(__file)
	if __fileUsesUnwrap(__file) {
		__out = __out + ____rune_private_68c6e3cf_emitTSUnwrapHelper()
	}
	if __fileUsesPathFamily(__file) {
		__out = __out + ____rune_private_68c6e3cf_emitTSPathHelpers()
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_68c6e3cf_emitTSEnum(__enumDecl) + "\n"
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_68c6e3cf_emitTSEnumMethods(__enumDecl)
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_68c6e3cf_emitTSStruct(__typeDecl) + "\n"
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_68c6e3cf_emitTSMethods(__typeDecl)
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__out = func() string {
			if __fn.__macro {
				return __out
			}
			return __out + ____rune_private_68c6e3cf_emitTSFunction(__fn) + "\n"
		}()
	}
	return __out + ____rune_private_68c6e3cf_emitTSExports(__file)
}

func ____rune_private_68c6e3cf_emitTSImports(__file __IRFile) string {
	return ____rune_private_68c6e3cf_emitTSImportList(__file.__tsImports, 0, "")
}

func ____rune_private_68c6e3cf_emitTSImportList(__imports []__IRTSImport, __index int, __out string) string {
	return func() string {
		if __index >= len(__imports) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSImportList(__imports, __index+1, __out+____rune_private_68c6e3cf_emitTSImport(__imports[__index]))
	}()
}

func ____rune_private_68c6e3cf_emitTSImport(__importDecl __IRTSImport) string {
	__names := ____rune_private_68c6e3cf_emitTSImportNames(__importDecl.__functions, __importDecl.__values, 0, 0, "")
	return func() string {
		if __names == "" {
			return ""
		}
		return "import { " + __names + " } from " + ____rune_private_68c6e3cf_tsQuoteString(____rune_private_68c6e3cf_tsRuntimeSpecifier(__importDecl.__specifier)) + ";\n"
	}()
}

func ____rune_private_68c6e3cf_emitTSImportNames(__functions []__IRFunction, __values []__IRConst, __fnIndex int, __valueIndex int, __out string) string {
	return func() string {
		if __fnIndex < len(__functions) {
			return ____rune_private_68c6e3cf_emitTSImportNames(__functions, __values, __fnIndex+1, __valueIndex, ____rune_private_68c6e3cf_appendTSImportName(__out, __functions[__fnIndex].__name))
		}
		return func() string {
			if __valueIndex < len(__values) {
				return ____rune_private_68c6e3cf_emitTSImportNames(__functions, __values, __fnIndex, __valueIndex+1, ____rune_private_68c6e3cf_appendTSImportName(__out, __values[__valueIndex].__name))
			}
			return __out
		}()
	}()
}

func ____rune_private_68c6e3cf_appendTSImportName(__out string, __name string) string {
	return __out + func() string {
		if __out == "" {
			return ""
		}
		return ", "
	}() + __name + " as " + __mangleIdent(__name)
}

func ____rune_private_68c6e3cf_tsQuoteString(__value string) string {
	return "\"" + strings.ReplaceAll((strings.ReplaceAll(__value, "\\", "\\\\")), "\"", "\\\"") + "\""
}

func ____rune_private_68c6e3cf_tsRuntimeSpecifier(__specifier string) string {
	return func() string {
		if __specifier == "" || strings.HasPrefix(__specifier, "./") || strings.HasPrefix(__specifier, "../") || strings.HasPrefix(__specifier, "/") || strings.Contains(__specifier, "://") {
			return __specifier
		}
		return "./" + __specifier
	}()
}

func ____rune_private_68c6e3cf_emitTSUnwrapHelper() string {
	return "function __runeUnwrap(value: any): any {\n  if (value && value.__tag === 0) return value.__payload?.[0];\n  if (value && value.__payload && value.__payload.length > 0) throw value.__payload[0];\n  throw new Error(\"Result.Err\");\n}\n\n"
}

func ____rune_private_68c6e3cf_emitTSPathHelpers() string {
	return "function __runePathBasename(path: string): string {\n  const index = path.lastIndexOf(\"/\");\n  if (index < 0) return path;\n  if (index === path.length - 1) return path;\n  return path.slice(index + 1);\n}\n\nfunction __runePathExtname(path: string): string {\n  const base = __runePathBasename(path);\n  const index = base.lastIndexOf(\".\");\n  if (index <= 0) return \"\";\n  return base.slice(index);\n}\n\nfunction __runePathDirname(path: string): string {\n  const index = path.lastIndexOf(\"/\");\n  if (index < 0) return \".\";\n  if (index === 0) return \"/\";\n  return path.slice(0, index);\n}\n\nfunction __runePathJoin(parts: string[]): string {\n  return __runePathNormalize(__runePathJoinParts(parts, 0, \"\"));\n}\n\nfunction __runePathNormalize(path: string): string {\n  const absolute = path.startsWith(\"/\");\n  const out = __runePathNormalizeParts(path.split(\"/\"), 0, absolute, []);\n  const joined = __runePathJoinParts(out, 0, \"\");\n  if (absolute) return \"/\" + joined;\n  return joined === \"\" ? \".\" : joined;\n}\n\nfunction __runePathResolve(parts: string[]): string {\n  if (parts.length === 0) return \".\";\n  return __runePathNormalize(__runePathJoin(parts));\n}\n\nfunction __runePathRelative(from: string, to: string): string {\n  const fromParts = __runePathParts(__runePathResolve([from]));\n  const toParts = __runePathParts(__runePathResolve([to]));\n  let index = 0;\n  while (index < fromParts.length && index < toParts.length && fromParts[index] === toParts[index]) index++;\n  let out = \"\";\n  for (let i = index; i < fromParts.length; i++) out = __runePathAppendPart(out, \"..\");\n  for (let i = index; i < toParts.length; i++) out = __runePathAppendPart(out, toParts[i]);\n  return out === \"\" ? \".\" : out;\n}\n\nfunction __runePathParts(path: string): string[] {\n  return __runePathNormalize(path).split(\"/\").filter((part) => part !== \"\");\n}\n\nfunction __runePathJoinParts(parts: string[], index: number, out: string): string {\n  for (let i = index; i < parts.length; i++) out = __runePathAppendPart(out, parts[i]);\n  return out;\n}\n\nfunction __runePathAppendPart(out: string, part: string): string {\n  if (out === \"\") return part;\n  if (part === \"\") return out;\n  return out + \"/\" + part;\n}\n\nfunction __runePathNormalizeParts(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  while (index < parts.length) {\n    const part = parts[index];\n    if (part === \"\" || part === \".\") {\n      index++;\n      continue;\n    }\n    if (part === \"..\") return __runePathNormalizeParent(parts, index, absolute, out);\n    return __runePathNormalizePush(parts, index, absolute, out, part);\n  }\n  return out;\n}\n\nfunction __runePathNormalizeParent(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  if (out.length > 0) return __runePathNormalizePop(parts, index, absolute, out);\n  if (absolute) return __runePathNormalizeParts(parts, index + 1, absolute, out);\n  return __runePathNormalizePush(parts, index, absolute, out, \"..\");\n}\n\nfunction __runePathNormalizePop(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  return __runePathNormalizeParts(parts, index + 1, absolute, out.slice(0, out.length - 1));\n}\n\nfunction __runePathNormalizePush(parts: string[], index: number, absolute: boolean, out: string[], part: string): string[] {\n  return __runePathNormalizeParts(parts, index + 1, absolute, [...out, part]);\n}\n\nfunction __runePathCollectParts(parts: string[], index: number, out: string[]): string[] {\n  for (let i = index; i < parts.length; i++) {\n    if (parts[i] !== \"\") out.push(parts[i]);\n  }\n  return out;\n}\n\nfunction __runePathCollectPart(parts: string[], index: number, out: string[]): string[] {\n  if (index < parts.length) out.push(parts[index]);\n  return __runePathCollectParts(parts, index + 1, out);\n}\n\nfunction __runePathRelativeFromParts(fromParts: string[], toParts: string[], index: number): string {\n  while (index < fromParts.length && index < toParts.length && fromParts[index] === toParts[index]) index++;\n  return __runePathRelativeTail(fromParts, toParts, index, index, \"\");\n}\n\nfunction __runePathRelativeTail(fromParts: string[], toParts: string[], fromIndex: number, toIndex: number, out: string): string {\n  for (let i = fromIndex; i < fromParts.length; i++) out = __runePathAppendPart(out, \"..\");\n  for (let i = toIndex; i < toParts.length; i++) out = __runePathAppendPart(out, toParts[i]);\n  return out === \"\" ? \".\" : out;\n}\n\n"
}

func ____rune_private_68c6e3cf_emitTSEnum(__enumDecl __IREnumType) string {
	return func() string {
		if ____rune_private_68c6e3cf_tsEnumHasPayload(__enumDecl.__members, 0) {
			return ____rune_private_68c6e3cf_emitTSPayloadEnum(__enumDecl)
		}
		return ____rune_private_68c6e3cf_emitTSSimpleEnum(__enumDecl)
	}()
}

func ____rune_private_68c6e3cf_emitTSSimpleEnum(__enumDecl __IREnumType) string {
	__out := "type " + __mangleIdent(__enumDecl.__name) + " = number;\n"
	__out = __out + "const " + __mangleIdent(__enumDecl.__name) + " = {\n"
	__out = __out + ____rune_private_68c6e3cf_emitTSEnumMembers(__enumDecl.__members, 0, "")
	return __out + "} as const;\n"
}

func ____rune_private_68c6e3cf_emitTSPayloadEnum(__enumDecl __IREnumType) string {
	__out := "type " + __mangleIdent(__enumDecl.__name) + ____rune_private_68c6e3cf_emitTSGenerics(__enumDecl.__generics) + " =\n"
	__out = __out + ____rune_private_68c6e3cf_emitTSPayloadEnumMembers(__enumDecl.__members, 0, "")
	__out = __out + ";\n\n"
	return __out + ____rune_private_68c6e3cf_emitTSPayloadEnumConstructors(__enumDecl.__name, __enumDecl.__generics, __enumDecl.__members, 0, "")
}

func ____rune_private_68c6e3cf_emitTSPayloadEnumMembers(__members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSPayloadEnumMembers(__members, __index+1, __out+____rune_private_68c6e3cf_emitTSPayloadEnumMember(__members[__index], __index, __index == 0))
	}()
}

func ____rune_private_68c6e3cf_emitTSPayloadEnumMember(__member __IREnumMember, __index int, __first bool) string {
	__prefix := func() string {
		if __first {
			return "  "
		}
		return "| "
	}()
	return __prefix + "{ __tag: " + __compilerIntToString(__index) + "; __payload: " + ____rune_private_68c6e3cf_emitTSPayloadTuple(__member.__params, 0, "") + " }\n"
}

func ____rune_private_68c6e3cf_emitTSPayloadTuple(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return "[" + __out + "]"
		}
		return ____rune_private_68c6e3cf_emitTSPayloadTuple(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_68c6e3cf_tsType(__params[__index].__typeName))
	}()
}

func ____rune_private_68c6e3cf_emitTSPayloadEnumConstructors(__enumName string, __generics []string, __members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSPayloadEnumConstructors(__enumName, __generics, __members, __index+1, __out+____rune_private_68c6e3cf_emitTSPayloadEnumConstructor(__enumName, __generics, __members[__index], __index))
	}()
}

func ____rune_private_68c6e3cf_emitTSPayloadEnumConstructor(__enumName string, __generics []string, __member __IREnumMember, __index int) string {
	__typeName := __mangleIdent(__enumName) + ____rune_private_68c6e3cf_emitTSGenericsUse(__generics)
	return func() string {
		if len(__member.__params) == 0 {
			return "const " + __mangleIdent(__enumName+"_"+__member.__name) + ": " + __typeName + " = { __tag: " + __compilerIntToString(__index) + ", __payload: [] };\n"
		}
		return "function " + __mangleIdent(__member.__name) + ____rune_private_68c6e3cf_emitTSGenerics(__generics) + "(" + ____rune_private_68c6e3cf_emitTSParams(__member.__params, 0, "") + "): " + __typeName + " {\n" + __line(1, "return { __tag: "+__compilerIntToString(__index)+", __payload: ["+____rune_private_68c6e3cf_emitTSParamNames(__member.__params, 0, "")+"] };") + "}\n"
	}()
}

func ____rune_private_68c6e3cf_emitTSParamNames(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSParamNames(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name))
	}()
}

func ____rune_private_68c6e3cf_tsEnumHasPayload(__members []__IREnumMember, __index int) bool {
	return func() bool {
		if __index >= len(__members) {
			return false
		}
		return func() bool {
			if len(__members[__index].__params) > 0 {
				return true
			}
			return ____rune_private_68c6e3cf_tsEnumHasPayload(__members, __index+1)
		}()
	}()
}

func ____rune_private_68c6e3cf_emitTSGenericsUse(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "<" + ____rune_private_68c6e3cf_emitTSGenericNames(__generics, 0, "") + ">"
	}()
}

func ____rune_private_68c6e3cf_emitTSEnumMembers(__members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSEnumMembers(__members, __index+1, __out+__indent(1)+____rune_private_68c6e3cf_tsPropertyName(__members[__index].__name)+": "+__enumValue(__members[__index], __index)+",\n")
	}()
}

func ____rune_private_68c6e3cf_emitTSStruct(__typeDecl __IRStructType) string {
	__generics := ____rune_private_68c6e3cf_emitTSGenerics(__typeDecl.__generics)
	__out := "type " + __mangleIdent(__typeDecl.__name) + __generics + " = {\n"
	for _, __field := range __typeDecl.__fields {
		_ = __field
		__out = __out + __indent(1) + ____rune_private_68c6e3cf_tsPropertyName(__field.__name) + ": " + ____rune_private_68c6e3cf_tsType(__field.__typeName) + ";\n"
	}
	return __out + "};\n"
}

func ____rune_private_68c6e3cf_emitTSMethods(__typeDecl __IRStructType) string {
	__out := ""
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_68c6e3cf_emitTSFunction(____rune_private_68c6e3cf_methodWithReceiver(__typeDecl.__name, __method)) + "\n"
	}
	return __out
}

func ____rune_private_68c6e3cf_emitTSEnumMethods(__enumDecl __IREnumType) string {
	__out := ""
	for _, __method := range __enumDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_68c6e3cf_emitTSFunction(____rune_private_68c6e3cf_methodWithReceiver(__enumDecl.__name, __method)) + "\n"
	}
	return __out
}

func ____rune_private_68c6e3cf_methodWithReceiver(__typeName string, __method __IRFunction) __IRFunction {
	return __IRFunction{__name: __typeName + "_" + __method.__name, __private: __method.__private, __routine: __method.__routine, __macro: __method.__macro, __receiverType: __method.__receiverType, __generics: __method.__generics, __params: ____rune_private_68c6e3cf_prependThisParam(__typeName, __method.__params), __returnType: __method.__returnType, __body: __method.__body, __line: __method.__line, __column: __method.__column}
}

func ____rune_private_68c6e3cf_prependThisParam(__typeName string, __params []__IRParam) []__IRParam {
	__out := []__IRParam{__IRParam{__name: "this", __typeName: __typeName, __line: 0, __column: 0}}
	for _, __param := range __params {
		_ = __param
		func() int { __out = append(__out, __param); return len(__out) }()
	}
	return __out
}

func ____rune_private_68c6e3cf_emitTSFunction(__fn __IRFunction) string {
	__ret := ____rune_private_68c6e3cf_tsType(__fn.__returnType)
	__out := "function " + __mangleIdent(__fn.__name) + ____rune_private_68c6e3cf_emitTSGenerics(__fn.__generics) + "(" + ____rune_private_68c6e3cf_emitTSParams(__fn.__params, 0, "") + "): " + __ret + " {\n"
	__out = __out + ____rune_private_68c6e3cf_emitTSBody(__fn.__body, __returnsValue(__fn.__returnType), __fn.__returnType, 1)
	return __out + "}\n"
}

func ____rune_private_68c6e3cf_emitTSBody(__expr __IRExpr, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_68c6e3cf_emitTSBlock(__expr.__children, 0, __returns, __returnType, __level, "")
		default:
			return __line(__level, func() string {
				if __returns {
					return "return " + ____rune_private_68c6e3cf_emitTSExprExpected(__expr, __returnType) + ";"
				}
				return ____rune_private_68c6e3cf_emitTSExpr(__expr) + ";"
			}())
		}
	}()
}

func ____rune_private_68c6e3cf_emitTSBlock(__statements []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__statements) {
			return func() string {
				if __returns && len(__statements) == 0 {
					return __out + __line(__level, "return undefined;")
				}
				return __out
			}()
		}
		return ____rune_private_68c6e3cf_emitTSBlock(__statements, __index+1, __returns, __returnType, __level, __out+____rune_private_68c6e3cf_emitTSStatement(__statements[__index], __index == len(__statements)-1, __returns, __returnType, __level))
	}()
}

func ____rune_private_68c6e3cf_emitTSStatement(__expr __IRExpr, __last bool, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Let:
			return ____rune_private_68c6e3cf_emitTSLet(__expr, __level)
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_68c6e3cf_emitTSObjectDestructure(__expr, __level)
		default:
			return func() string {
				if __last && __returns {
					return __line(__level, "return "+____rune_private_68c6e3cf_emitTSExprExpected(__expr, __returnType)+";")
				}
				return __line(__level, ____rune_private_68c6e3cf_emitTSExpr(__expr)+";")
			}()
		}
	}()
}

func ____rune_private_68c6e3cf_emitTSLet(__expr __IRExpr, __level int) string {
	return __line(__level, ____rune_private_68c6e3cf_tsLetKeyword(__expr.__op)+__mangleIdent(__expr.__name)+" = "+____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0])+";")
}

func ____rune_private_68c6e3cf_tsLetKeyword(__op string) string {
	return func() string {
		switch {
		case __op == ":=:":
			return "let "
		default:
			return "const "
		}
	}()
}

func ____rune_private_68c6e3cf_emitTSObjectDestructure(__expr __IRExpr, __level int) string {
	return __line(__level, "const { "+____rune_private_68c6e3cf_emitTSObjectDestructureFields(__expr.__params, 0, "")+" } = "+____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0])+";")
}

func ____rune_private_68c6e3cf_emitTSObjectDestructureFields(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSObjectDestructureFields(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_68c6e3cf_tsPropertyName(__params[__index].__typeName)+": "+__mangleIdent(__params[__index].__name))
	}()
}

func ____rune_private_68c6e3cf_emitTSParams(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSParams(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name)+": "+____rune_private_68c6e3cf_tsType(__params[__index].__typeName))
	}()
}

func ____rune_private_68c6e3cf_emitTSGenerics(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "<" + ____rune_private_68c6e3cf_emitTSGenericNames(__generics, 0, "") + ">"
	}()
}

func ____rune_private_68c6e3cf_emitTSGenericNames(__generics []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__generics) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSGenericNames(__generics, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__generics[__index]))
	}()
}

func ____rune_private_68c6e3cf_emitTSExports(__file __IRFile) string {
	__exports := ____rune_private_68c6e3cf_emitTSExportNames(__file.__functions, 0, "")
	return func() string {
		if __exports == "" {
			return ""
		}
		return "export { " + __exports + " };\n"
	}()
}

func ____rune_private_68c6e3cf_emitTSExportNames(__functions []__IRFunction, __index int, __out string) string {
	return func() string {
		if __index >= len(__functions) {
			return __out
		}
		return func() string {
			if __functions[__index].__macro || __functions[__index].__private {
				return ____rune_private_68c6e3cf_emitTSExportNames(__functions, __index+1, __out)
			}
			return ____rune_private_68c6e3cf_emitTSExportNames(__functions, __index+1, __out+func() string {
				if __out == "" {
					return ""
				}
				return ", "
			}()+__mangleIdent(__functions[__index].__name)+" as "+__functions[__index].__name)
		}()
	}()
}

func ____rune_private_68c6e3cf_emitTSExpr(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Identifier:
			return __mangleIdent(__expr.__name)
		case __expr.__kind == __ExprKind_At:
			return "@" + __expr.__name
		case __expr.__kind == __ExprKind_This:
			return __mangleIdent("this")
		case __expr.__kind == __ExprKind_Int:
			return __expr.__value
		case __expr.__kind == __ExprKind_Double:
			return __expr.__value
		case __expr.__kind == __ExprKind_BigInt:
			return __bigintLiteralDigits(__expr.__value) + "n"
		case __expr.__kind == __ExprKind_String:
			return __expr.__value
		case __expr.__kind == __ExprKind_Template:
			return __expr.__value
		case __expr.__kind == __ExprKind_Char:
			return __expr.__value
		case __expr.__kind == __ExprKind_Regex:
			return __expr.__value
		case __expr.__kind == __ExprKind_Bool:
			return __expr.__value
		case __expr.__kind == __ExprKind_Null:
			return "null"
		case __expr.__kind == __ExprKind_Unary:
			return __expr.__op + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Postfix:
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + __expr.__op
		case __expr.__kind == __ExprKind_CompileTime:
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Unwrap:
			return "__runeUnwrap(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + ")"
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_68c6e3cf_emitTSBinary(__expr)
		case __expr.__kind == __ExprKind_Ternary:
			return ____rune_private_68c6e3cf_emitTSTernary(__expr)
		case __expr.__kind == __ExprKind_Assign:
			return ____rune_private_68c6e3cf_emitTSAssign(__expr)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_68c6e3cf_emitTSCall(__expr)
		case __expr.__kind == __ExprKind_Lambda:
			return ____rune_private_68c6e3cf_emitTSLambda(__expr)
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_68c6e3cf_emitTSSelector(__expr)
		case __expr.__kind == __ExprKind_Index:
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + "[" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + "]"
		case __expr.__kind == __ExprKind_Array:
			return "[" + ____rune_private_68c6e3cf_emitTSExprList(__expr.__children, 0, "") + "]"
		case __expr.__kind == __ExprKind_Tuple:
			return "[" + ____rune_private_68c6e3cf_emitTSExprList(__expr.__children, 0, "") + "]"
		case __expr.__kind == __ExprKind_Map:
			return "new Map([" + ____rune_private_68c6e3cf_emitTSMapEntries(__expr.__children, 0, "") + "])"
		case __expr.__kind == __ExprKind_Spread:
			return "..." + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Reactive:
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Struct:
			return "{" + ____rune_private_68c6e3cf_emitTSFields(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Object:
			return "{" + ____rune_private_68c6e3cf_emitTSFields(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Block:
			return "(() => {\n" + ____rune_private_68c6e3cf_emitTSBlock(__expr.__children, 0, true, "Dynamic", 1, "") + "})()"
		default:
			return "undefined"
		}
	}()
}

func ____rune_private_68c6e3cf_emitTSExprExpected(__expr __IRExpr, __expected string) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_68c6e3cf_emitTSCallExpected(__expr, __expected)
		default:
			return ____rune_private_68c6e3cf_emitTSExpr(__expr)
		}
	}()
}

func ____rune_private_68c6e3cf_emitTSCallExpected(__expr __IRExpr, __expected string) string {
	__args := __genericInner(__expected, "Result")
	return func() string {
		if __args != "" && ____rune_private_68c6e3cf_isTSResultConstructorCall(__expr) {
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + "<" + ____rune_private_68c6e3cf_emitTSTypeArgs(__args) + ">(" + ____rune_private_68c6e3cf_emitTSExprListFrom(__expr.__children, 1, "") + ")"
		}
		return ____rune_private_68c6e3cf_emitTSCall(__expr)
	}()
}

func ____rune_private_68c6e3cf_isTSResultConstructorCall(__expr __IRExpr) bool {
	return __expr.__kind == __ExprKind_Call && len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier && (__expr.__children[0].__name == "Ok" || __expr.__children[0].__name == "Err")
}

func ____rune_private_68c6e3cf_emitTSBinary(__expr __IRExpr) string {
	return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + " " + ____rune_private_68c6e3cf_tsBinaryOp(__expr.__op) + " " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1])
}

func ____rune_private_68c6e3cf_emitTSTernary(__expr __IRExpr) string {
	return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + " ? " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + " : " + func() string {
		if len(__expr.__children) > 2 {
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2])
		}
		return "undefined"
	}()
}

func ____rune_private_68c6e3cf_emitTSAssign(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) == 2 {
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + " = " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1])
		}
		return __mangleIdent(__expr.__name) + " = " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0])
	}()
}

func ____rune_private_68c6e3cf_emitTSCall(__expr __IRExpr) string {
	return func() string {
		switch {
		case __moduleCallKey(__expr) == "io.println":
			return "console.log(" + ____rune_private_68c6e3cf_emitTSExprListFrom(__expr.__children, 1, "") + ")"
		case __moduleCallKey(__expr) == "map.new":
			return "new Map()"
		case __moduleCallKey(__expr) == "set.new":
			return "new Set()"
		case __moduleCallKey(__expr) == "path.isAbsolute":
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ".startsWith(\"/\")"
		case __moduleCallKey(__expr) == "path.basename":
			return "__runePathBasename(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.extname":
			return "__runePathExtname(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.dirname":
			return "__runePathDirname(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.join":
			return "__runePathJoin(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.normalize":
			return "__runePathNormalize(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.resolve":
			return "__runePathResolve(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.relative":
			return "__runePathRelative(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.joinParts":
			return "__runePathJoinParts(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.appendPathPart":
			return "__runePathAppendPart(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.normalizeParts":
			return "__runePathNormalizeParts(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[4]) + ")"
		case __moduleCallKey(__expr) == "path.normalizePart":
			return "__runePathNormalizeParts(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[4]) + ")"
		case __moduleCallKey(__expr) == "path.normalizeParent":
			return "__runePathNormalizeParent(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[4]) + ")"
		case __moduleCallKey(__expr) == "path.normalizePop":
			return "__runePathNormalizePop(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[4]) + ")"
		case __moduleCallKey(__expr) == "path.normalizePush":
			return "__runePathNormalizePush(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[4]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[5]) + ")"
		case __moduleCallKey(__expr) == "path.pathParts":
			return "__runePathParts(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.collectPathParts":
			return "__runePathCollectParts(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.collectPathPart":
			return "__runePathCollectPart(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.relativeFromParts":
			return "__runePathRelativeFromParts(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.relativeTail":
			return "__runePathRelativeTail(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[4]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[5]) + ")"
		case __moduleCallKey(__expr) == "process.platform":
			return "\"js\""
		case __moduleCallKey(__expr) == "process.cwd":
			return "\".\""
		case __moduleCallKey(__expr) == "process.env":
			return "null"
		case __moduleCallKey(__expr) == "process.argv":
			return "[]"
		case __moduleCallKey(__expr) == "int.toString":
			return "String(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int.toDouble":
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1])
		case __moduleCallKey(__expr) == "int.toBigInt":
			return "BigInt(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int4.fromInt":
			return "((__value: number): number => { const __n = __value & 0xf; return __n >= 8 ? __n - 16 : __n; })(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int8.fromInt":
			return "((__value: number): number => (__value << 24) >> 24)(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int16.fromInt":
			return "((__value: number): number => (__value << 16) >> 16)(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int64.fromInt":
			return "BigInt(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "uint.fromInt":
			return "(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + " >>> 0)"
		case __moduleCallKey(__expr) == "uint8.fromInt":
			return "(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + " & 0xff)"
		case __moduleCallKey(__expr) == "uint16.fromInt":
			return "(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + " & 0xffff)"
		case __moduleCallKey(__expr) == "uint64.fromInt":
			return "BigInt.asUintN(64, BigInt(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + "))"
		case __moduleCallKey(__expr) == "float.fromDouble":
			return "Math.fround(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case (__moduleCallKey(__expr) == "int4.toInt") || (__moduleCallKey(__expr) == "int8.toInt") || (__moduleCallKey(__expr) == "int16.toInt") || (__moduleCallKey(__expr) == "uint.toInt") || (__moduleCallKey(__expr) == "uint8.toInt") || (__moduleCallKey(__expr) == "uint16.toInt") || (__moduleCallKey(__expr) == "float.toDouble"):
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1])
		case (__moduleCallKey(__expr) == "int64.toInt") || (__moduleCallKey(__expr) == "uint64.toInt"):
			return "Number(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.fromInt":
			return "BigInt(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.toString":
			return "String(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.toDouble":
			return "Number(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.trunc":
			return "Math.trunc(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.floor":
			return "Math.floor(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.ceil":
			return "Math.ceil(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.round":
			return "Math.round(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ")"
		default:
			return ____rune_private_68c6e3cf_emitTSMaybeCoreMethodCall(__expr)
		}
	}()
}

func ____rune_private_68c6e3cf_emitTSMaybeCoreMethodCall(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Selector {
			return ____rune_private_68c6e3cf_emitTSCoreMethodCall(__expr, __expr.__children[0])
		}
		return ____rune_private_68c6e3cf_emitTSDefaultCall(__expr)
	}()
}

func ____rune_private_68c6e3cf_emitTSCoreMethodCall(__expr __IRExpr, __selector __IRExpr) string {
	return func() string {
		if len(__selector.__children) > 0 && __selector.__children[0].__kind != __ExprKind_At {
			return func() string {
				switch {
				case (__selector.__name == "length") || (__selector.__name == "byteLength"):
					return ____rune_private_68c6e3cf_emitTSCoreLength(__selector.__children[0])
				case __selector.__name == "isEmpty":
					return "(" + ____rune_private_68c6e3cf_emitTSCoreLength(__selector.__children[0]) + " === 0)"
				case __selector.__name == "at":
					return ____rune_private_68c6e3cf_emitTSCoreAt(__expr, __selector.__children[0])
				case __selector.__name == "slice":
					return ____rune_private_68c6e3cf_emitTSCoreSlice(__expr, __selector.__children[0])
				default:
					return ____rune_private_68c6e3cf_emitTSDefaultCall(__expr)
				}
			}()
		}
		return ____rune_private_68c6e3cf_emitTSDefaultCall(__expr)
	}()
}

func ____rune_private_68c6e3cf_emitTSCoreLength(__receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "Array.from(" + ____rune_private_68c6e3cf_emitTSExpr(__receiver) + ").length"
		}
		return ____rune_private_68c6e3cf_emitTSExpr(__receiver) + ".length"
	}()
}

func ____rune_private_68c6e3cf_emitTSCoreAt(__expr __IRExpr, __receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "(Array.from(" + ____rune_private_68c6e3cf_emitTSExpr(__receiver) + ")[" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + "] ?? \"\")"
		}
		return ____rune_private_68c6e3cf_emitTSExpr(__receiver) + "[" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + "]"
	}()
}

func ____rune_private_68c6e3cf_emitTSCoreSlice(__expr __IRExpr, __receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "Array.from(" + ____rune_private_68c6e3cf_emitTSExpr(__receiver) + ").slice(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ").join(\"\")"
		}
		return ____rune_private_68c6e3cf_emitTSExpr(__receiver) + ".slice(" + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[2]) + ")"
	}()
}

func ____rune_private_68c6e3cf_emitTSDefaultCall(__expr __IRExpr) string {
	return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + "(" + ____rune_private_68c6e3cf_emitTSExprListFrom(__expr.__children, 1, "") + ")"
}

func ____rune_private_68c6e3cf_emitTSLambda(__expr __IRExpr) string {
	return "(" + ____rune_private_68c6e3cf_emitTSParams(__expr.__params, 0, "") + ") => " + ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0])
}

func ____rune_private_68c6e3cf_emitTSSelector(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__children[0].__kind == __ExprKind_At:
			return "@" + __expr.__children[0].__name + "." + __expr.__name
		default:
			return ____rune_private_68c6e3cf_emitTSExpr(__expr.__children[0]) + "." + ____rune_private_68c6e3cf_tsPropertyName(__expr.__name)
		}
	}()
}

func ____rune_private_68c6e3cf_emitTSExprList(__exprs []__IRExpr, __index int, __out string) string {
	return ____rune_private_68c6e3cf_emitTSExprListFrom(__exprs, __index, __out)
}

func ____rune_private_68c6e3cf_emitTSExprListFrom(__exprs []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__exprs) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSExprListFrom(__exprs, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_68c6e3cf_emitTSExpr(__exprs[__index]))
	}()
}

func ____rune_private_68c6e3cf_emitTSMapEntries(__entries []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__entries) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSMapEntries(__entries, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+"["+____rune_private_68c6e3cf_emitTSExpr(__entries[__index].__children[0])+", "+____rune_private_68c6e3cf_emitTSExpr(__entries[__index].__children[1])+"]")
	}()
}

func ____rune_private_68c6e3cf_emitTSFields(__fields []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__fields) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSFields(__fields, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_68c6e3cf_tsPropertyName(__fields[__index].__name)+": "+____rune_private_68c6e3cf_emitTSExpr(__fields[__index].__children[0]))
	}()
}

func ____rune_private_68c6e3cf_tsBinaryOp(__op string) string {
	return func() string {
		switch {
		case __op == "==":
			return "==="
		case __op == "!=":
			return "!=="
		default:
			return __op
		}
	}()
}

func ____rune_private_68c6e3cf_tsType(__typeName string) string {
	switch {
	case (__typeName == "") || (__typeName == "Void"):
		return "void"
	case (__typeName == "Int") || (__typeName == "Int4") || (__typeName == "Int8") || (__typeName == "Int16") || (__typeName == "UInt") || (__typeName == "UInt8") || (__typeName == "UInt16") || (__typeName == "Double") || (__typeName == "Float"):
		return "number"
	case (__typeName == "BigInt") || (__typeName == "Int64") || (__typeName == "UInt64"):
		return "bigint"
	case (__typeName == "String") || (__typeName == "Char"):
		return "string"
	case __typeName == "Bool":
		return "boolean"
	case __typeName == "Dynamic":
		return "any"
	case (__typeName == "Data") || (__typeName == "@io.Data"):
		return "Uint8Array"
	default:
		return ____rune_private_68c6e3cf_tsTypeFallback(__typeName)
	}
}

func ____rune_private_68c6e3cf_tsTypeFallback(__typeName string) string {
	return func() string {
		if strings.HasSuffix(__typeName, "?") {
			return ____rune_private_68c6e3cf_tsType(func() string { runes := []rune(__typeName); return string(runes[0 : len([]rune(__typeName))-1]) }()) + " | null"
		}
		return func() string {
			if __genericInner(__typeName, "Array") != "" {
				return ____rune_private_68c6e3cf_tsType(__genericInner(__typeName, "Array")) + "[]"
			}
			return func() string {
				if __genericInner(__typeName, "ReadonlyArray") != "" {
					return "ReadonlyArray<" + ____rune_private_68c6e3cf_tsType(__genericInner(__typeName, "ReadonlyArray")) + ">"
				}
				return func() string {
					if __genericInner(__typeName, "Map") != "" {
						return "Map<" + ____rune_private_68c6e3cf_tsType(__typeArg(__genericInner(__typeName, "Map"), 0)) + ", " + ____rune_private_68c6e3cf_tsType(__typeArg(__genericInner(__typeName, "Map"), 1)) + ">"
					}
					return func() string {
						if __genericInner(__typeName, "Set") != "" {
							return "Set<" + ____rune_private_68c6e3cf_tsType(__genericInner(__typeName, "Set")) + ">"
						}
						return ____rune_private_68c6e3cf_tsNamedType(__typeName)
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_68c6e3cf_tsNamedType(__typeName string) string {
	__open := strings.Index(__typeName, "[")
	return func() string {
		if __open < 0 {
			return __mangleIdent(__typeName)
		}
		return __mangleIdent(func() string { runes := []rune(__typeName); return string(runes[0:__open]) }()) + "<" + ____rune_private_68c6e3cf_emitTSTypeArgs(func() string { runes := []rune(__typeName); return string(runes[__open+1 : len([]rune(__typeName))-1]) }()) + ">"
	}()
}

func ____rune_private_68c6e3cf_emitTSTypeArgs(__args string) string {
	return ____rune_private_68c6e3cf_emitTSTypeArgList(func() []string { parts := strings.Split(__args, ","); return parts }(), 0, "")
}

func ____rune_private_68c6e3cf_emitTSTypeArgList(__args []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_68c6e3cf_emitTSTypeArgList(__args, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_68c6e3cf_tsType(strings.TrimSpace(__args[__index])))
	}()
}

func ____rune_private_68c6e3cf_tsPropertyName(__name string) string {
	return __name
}

func ____rune_private_68c6e3cf_joinStrings(__values []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__values) {
			return __out
		}
		return ____rune_private_68c6e3cf_joinStrings(__values, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__values[__index])
	}()
}

func __compileTypeScript(__source string) __CompileResult {
	return __compile(__source, "ts")
}

func __compileGo(__source string) __CompileResult {
	return __compile(__source, "go")
}

func __compileMoonBit(__source string) __CompileResult {
	return __compile(__source, "mbt")
}

func __compile(__source string, __target string) __CompileResult {
	__file := __lower(__source)
	return ____rune_private_0d2ebf0f_compileFile(__file, __target)
}

func __compileTypeScriptFiles(__files []__SourceFile) __CompileResult {
	return __compileFiles(__files, "ts")
}

func __compileGoFiles(__files []__SourceFile) __CompileResult {
	return __compileFiles(__files, "go")
}

func __compileMoonBitFiles(__files []__SourceFile) __CompileResult {
	return __compileFiles(__files, "mbt")
}

func __compileFiles(__files []__SourceFile, __target string) __CompileResult {
	__file := ____rune_private_0d2ebf0f_lowerFiles(__files)
	return ____rune_private_0d2ebf0f_compileFile(__file, __target)
}

func ____rune_private_0d2ebf0f_compileFile(__file __IRFile, __target string) __CompileResult {
	return func() __CompileResult {
		if len(__file.__errors) > 0 {
			return ____rune_private_0d2ebf0f_compileResult(false, "", ____rune_private_0d2ebf0f_parseErrorMessages(__file.__errors))
		}
		return ____rune_private_0d2ebf0f_compileCheckedFile(__file, __target)
	}()
}

func ____rune_private_0d2ebf0f_compileCheckedFile(__file __IRFile, __target string) __CompileResult {
	__errors := ____rune_private_0d2ebf0f_checkFileErrors(__file)
	return func() __CompileResult {
		if len(__errors) > 0 {
			return ____rune_private_0d2ebf0f_compileResult(false, "", __errors)
		}
		return func() __CompileResult {
			switch {
			case __target == "ts":
				return ____rune_private_0d2ebf0f_compileResult(true, __generateTypeScript(__file), []string{})
			case __target == "go":
				return ____rune_private_0d2ebf0f_compileResult(true, __generateGo(__file), []string{})
			case __target == "mbt":
				return ____rune_private_0d2ebf0f_compileResult(true, __generateMoonBit(__file), []string{})
			default:
				return ____rune_private_0d2ebf0f_compileResult(false, "", ____rune_private_0d2ebf0f_unsupportedTargetErrors(__target))
			}
		}()
	}()
}

func ____rune_private_0d2ebf0f_checkFileErrors(__file __IRFile) []string {
	__names := ____rune_private_0d2ebf0f_compilerCallableNames(__file)
	__errors := append([]string{}, []string{""}[0:0]...)
	for _, __fn := range __file.__functions {
		_ = __fn
		__errors = func() []string {
			if __fn.__macro {
				return __errors
			}
			return ____rune_private_0d2ebf0f_checkExpr(__fn.__body, __names, __errors)
		}()
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		func() {
			for _, __method := range __typeDecl.__methods {
				_ = __method
				__errors = ____rune_private_0d2ebf0f_checkExpr(__method.__body, __names, __errors)
			}
		}()
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		func() {
			for _, __method := range __typeDecl.__methods {
				_ = __method
				__errors = ____rune_private_0d2ebf0f_checkExpr(__method.__body, __names, __errors)
			}
		}()
	}
	for _, __testDecl := range __file.__tests {
		_ = __testDecl
		__errors = ____rune_private_0d2ebf0f_checkExpr(__testDecl.__body, __names, __errors)
	}
	return __errors
}

func ____rune_private_0d2ebf0f_compilerCallableNames(__file __IRFile) []string {
	__names := append([]string{}, []string{""}[0:0]...)
	for _, __fn := range __file.__functions {
		_ = __fn
		func() int {
			if __fn.__macro {
				return 0
			}
			return func() int { __names = append(__names, __fn.__name); return len(__names) }()
		}()
	}
	for _, __importDecl := range __file.__tsImports {
		_ = __importDecl
		func() {
			for _, __fn := range __importDecl.__functions {
				_ = __fn
				func() int { __names = append(__names, __fn.__name); return len(__names) }()
			}
		}()
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		func() {
			for _, __member := range __typeDecl.__members {
				_ = __member
				func() int { __names = append(__names, __member.__name); return len(__names) }()
			}
		}()
	}
	return __names
}

func ____rune_private_0d2ebf0f_checkExpr(__expr __IRExpr, __names []string, __errors []string) []string {
	__next := func() []string {
		if ____rune_private_0d2ebf0f_isUnknownFunctionCall(__expr, __names) {
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "undefined function "+__expr.__children[0].__name)
				return out
			}()
		}
		return __errors
	}()
	for _, __child := range __expr.__children {
		_ = __child
		__next = ____rune_private_0d2ebf0f_checkExpr(__child, __names, __next)
	}
	return __next
}

func ____rune_private_0d2ebf0f_isUnknownFunctionCall(__expr __IRExpr, __names []string) bool {
	return func() bool {
		if __expr.__kind == __ExprKind_Call && len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier {
			return ____rune_private_0d2ebf0f_compilerContains(__names, __expr.__children[0].__name) == false
		}
		return false
	}()
}

func ____rune_private_0d2ebf0f_compilerContains(__values []string, __value string) bool {
	return ____rune_private_0d2ebf0f_compilerContainsAt(__values, __value, 0)
}

func ____rune_private_0d2ebf0f_compilerContainsAt(__values []string, __value string, __index int) bool {
	return func() bool {
		if __index >= len(__values) {
			return false
		}
		return func() bool {
			if __values[__index] == __value {
				return true
			}
			return ____rune_private_0d2ebf0f_compilerContainsAt(__values, __value, __index+1)
		}()
	}()
}

func ____rune_private_0d2ebf0f_lowerFiles(__files []__SourceFile) __IRFile {
	__out := ____rune_private_0d2ebf0f_emptyCompilerIRFile()
	for _, __file := range __files {
		_ = __file
		__out = func() __IRFile {
			if ____rune_private_0d2ebf0f_sourceFileIsTypeScript(__file) {
				return __out
			}
			return ____rune_private_0d2ebf0f_mergeCompilerIRFile(__out, __lower(__file.__source))
		}()
	}
	for _, __file := range __files {
		_ = __file
		__out = func() __IRFile {
			if ____rune_private_0d2ebf0f_sourceFileIsTypeScript(__file) {
				return ____rune_private_0d2ebf0f_mergeCompilerIRFile(__out, ____rune_private_0d2ebf0f_lowerTypeScriptSourceFile(__file, __out.__imports))
			}
			return __out
		}()
	}
	return __out
}

func ____rune_private_0d2ebf0f_sourceFileIsTypeScript(__file __SourceFile) bool {
	return strings.HasSuffix(__file.__path, ".ts")
}

func ____rune_private_0d2ebf0f_lowerTypeScriptSourceFile(__file __SourceFile, __imports []__IRImport) __IRFile {
	__out := ____rune_private_0d2ebf0f_emptyCompilerIRFile()
	__specifier := ____rune_private_0d2ebf0f_typeScriptImportSpecifier(__file.__path, __imports, 0)
	if __specifier != "" {
		func() int {
			__out.__tsImports = append(__out.__tsImports, ____rune_private_0d2ebf0f_parseTypeScriptImport(__file.__path, __specifier, __file.__source))
			return len(__out.__tsImports)
		}()
	}
	return __out
}

func ____rune_private_0d2ebf0f_typeScriptImportSpecifier(__path string, __imports []__IRImport, __index int) string {
	return func() string {
		if __index >= len(__imports) {
			return ""
		}
		return func() string {
			if strings.HasSuffix(__imports[__index].__path, ".ts") && ____rune_private_0d2ebf0f_compilerPathBasename(__path) == ____rune_private_0d2ebf0f_compilerPathBasename(__imports[__index].__path) {
				return __imports[__index].__path
			}
			return ____rune_private_0d2ebf0f_typeScriptImportSpecifier(__path, __imports, __index+1)
		}()
	}()
}

func ____rune_private_0d2ebf0f_compilerPathBasename(__path string) string {
	__slash := strings.LastIndex(__path, "/")
	return func() string {
		if __slash < 0 {
			return __path
		}
		return func() string { runes := []rune(__path); return string(runes[__slash+1 : len([]rune(__path))]) }()
	}()
}

func ____rune_private_0d2ebf0f_parseTypeScriptImport(__path string, __specifier string, __source string) __IRTSImport {
	__imports := __IRTSImport{__path: __path, __specifier: __specifier, __functions: append([]__IRFunction{}, []__IRFunction{__emptyIRFunction()}[0:0]...), __values: append([]__IRConst{}, []__IRConst{____rune_private_0d2ebf0f_emptyIRConst()}[0:0]...), __line: 0, __column: 0}
	for _, __line := range func() []string { parts := strings.Split(__source, "\n"); return parts }() {
		_ = __line
		__imports = ____rune_private_0d2ebf0f_parseTypeScriptExportLine(__imports, strings.TrimSpace(__line))
	}
	return __imports
}

func ____rune_private_0d2ebf0f_parseTypeScriptExportLine(__imports __IRTSImport, __line string) __IRTSImport {
	return func() __IRTSImport {
		if strings.HasPrefix(__line, "export async function ") {
			return ____rune_private_0d2ebf0f_pushTypeScriptFunction(__imports, func() string {
				runes := []rune(__line)
				return string(runes[len([]rune("export async function ")):len([]rune(__line))])
			}(), true)
		}
		return func() __IRTSImport {
			if strings.HasPrefix(__line, "export function ") {
				return ____rune_private_0d2ebf0f_pushTypeScriptFunction(__imports, func() string {
					runes := []rune(__line)
					return string(runes[len([]rune("export function ")):len([]rune(__line))])
				}(), false)
			}
			return func() __IRTSImport {
				if strings.HasPrefix(__line, "export const ") {
					return ____rune_private_0d2ebf0f_pushTypeScriptValue(__imports, func() string {
						runes := []rune(__line)
						return string(runes[len([]rune("export const ")):len([]rune(__line))])
					}())
				}
				return func() __IRTSImport {
					if strings.HasPrefix(__line, "export let ") {
						return ____rune_private_0d2ebf0f_pushTypeScriptValue(__imports, func() string {
							runes := []rune(__line)
							return string(runes[len([]rune("export let ")):len([]rune(__line))])
						}())
					}
					return func() __IRTSImport {
						if strings.HasPrefix(__line, "export var ") {
							return ____rune_private_0d2ebf0f_pushTypeScriptValue(__imports, func() string {
								runes := []rune(__line)
								return string(runes[len([]rune("export var ")):len([]rune(__line))])
							}())
						}
						return __imports
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_0d2ebf0f_pushTypeScriptFunction(__imports __IRTSImport, __text string, __routine bool) __IRTSImport {
	__open := strings.Index(__text, "(")
	__close := strings.Index(__text, ")")
	__name := func() string {
		if __open < 0 {
			return ""
		}
		return strings.TrimSpace((func() string { runes := []rune(__text); return string(runes[0:__open]) }()))
	}()
	__returnType := ____rune_private_0d2ebf0f_typeScriptReturnTypeName(__text)
	if __name != "" {
		func() int {
			__imports.__functions = append(__imports.__functions, __IRFunction{__name: __name, __private: false, __routine: __routine, __macro: false, __receiverType: "", __generics: []string{}, __params: func() []__IRParam {
				if __open >= 0 && __close > __open {
					return ____rune_private_0d2ebf0f_parseTypeScriptParams(func() string { runes := []rune(__text); return string(runes[__open+1 : __close]) }())
				}
				return append([]__IRParam{}, []__IRParam{____rune_private_0d2ebf0f_emptyIRParam()}[0:0]...)
			}(), __returnType: __returnType, __body: __emptyIRExpr(), __line: 0, __column: 0})
			return len(__imports.__functions)
		}()
	}
	return __imports
}

func ____rune_private_0d2ebf0f_pushTypeScriptValue(__imports __IRTSImport, __text string) __IRTSImport {
	__end := ____rune_private_0d2ebf0f_typeScriptNameEnd(__text)
	__name := strings.TrimSpace((func() string { runes := []rune(__text); return string(runes[0:__end]) }()))
	__typeName := ____rune_private_0d2ebf0f_typeScriptValueTypeName(__text)
	if __name != "" {
		func() int {
			__imports.__values = append(__imports.__values, __IRConst{__name: __name, __private: false, __typeName: __typeName, __value: __emptyIRExpr(), __line: 0, __column: 0})
			return len(__imports.__values)
		}()
	}
	return __imports
}

func ____rune_private_0d2ebf0f_typeScriptNameEnd(__text string) int {
	__colon := strings.Index(__text, ":")
	__equal := strings.Index(__text, "=")
	return func() int {
		if __colon >= 0 && (__equal < 0 || __colon < __equal) {
			return __colon
		}
		return func() int {
			if __equal >= 0 {
				return __equal
			}
			return len([]rune(__text))
		}()
	}()
}

func ____rune_private_0d2ebf0f_typeScriptReturnTypeName(__text string) string {
	__close := strings.Index(__text, ")")
	__colon := func() int {
		if __close >= 0 {
			return strings.Index((func() string { runes := []rune(__text); return string(runes[__close+1 : len([]rune(__text))]) }()), ":")
		}
		return -1
	}()
	return func() string {
		if __close >= 0 && __colon >= 0 {
			return ____rune_private_0d2ebf0f_typeScriptTextType(strings.TrimSpace((func() string {
				runes := []rune(__text)
				return string(runes[__close+1+__colon+1 : ____rune_private_0d2ebf0f_typeScriptReturnTypeEnd(__text)])
			}())))
		}
		return "Dynamic"
	}()
}

func ____rune_private_0d2ebf0f_typeScriptReturnTypeEnd(__text string) int {
	__brace := strings.Index(__text, "{")
	__semi := strings.Index(__text, ";")
	return func() int {
		if __brace >= 0 && (__semi < 0 || __brace < __semi) {
			return __brace
		}
		return func() int {
			if __semi >= 0 {
				return __semi
			}
			return len([]rune(__text))
		}()
	}()
}

func ____rune_private_0d2ebf0f_typeScriptValueTypeName(__text string) string {
	__colon := strings.Index(__text, ":")
	return func() string {
		if __colon < 0 {
			return "Dynamic"
		}
		return ____rune_private_0d2ebf0f_typeScriptTextType(strings.TrimSpace((func() string {
			runes := []rune(__text)
			return string(runes[__colon+1 : ____rune_private_0d2ebf0f_typeScriptNameEnd(func() string { runes := []rune(__text); return string(runes[__colon+1 : len([]rune(__text))]) }())+__colon+1])
		}())))
	}()
}

func ____rune_private_0d2ebf0f_parseTypeScriptParams(__text string) []__IRParam {
	__params := append([]__IRParam{}, []__IRParam{____rune_private_0d2ebf0f_emptyIRParam()}[0:0]...)
	for _, __param := range func() []string { parts := strings.Split(__text, ","); return parts }() {
		_ = __param
		func() {
			if strings.TrimSpace(__param) != "" {
				func() int {
					__params = append(__params, ____rune_private_0d2ebf0f_parseTypeScriptParam(strings.TrimSpace(__param)))
					return len(__params)
				}()
				return
			}
		}()
	}
	return __params
}

func ____rune_private_0d2ebf0f_emptyIRParam() __IRParam {
	return __IRParam{__name: "", __typeName: "Dynamic", __line: 0, __column: 0}
}

func ____rune_private_0d2ebf0f_emptyIRConst() __IRConst {
	return __IRConst{__name: "", __private: false, __typeName: "Dynamic", __value: __emptyIRExpr(), __line: 0, __column: 0}
}

func ____rune_private_0d2ebf0f_parseTypeScriptParam(__text string) __IRParam {
	__colon := strings.Index(__text, ":")
	__rawName := func() string {
		if __colon < 0 {
			return __text
		}
		return func() string { runes := []rune(__text); return string(runes[0:__colon]) }()
	}()
	__name := strings.TrimSpace((strings.ReplaceAll((strings.ReplaceAll(__rawName, "...", "")), "?", "")))
	return __IRParam{__name: __name, __typeName: func() string {
		if __colon < 0 {
			return "Dynamic"
		}
		return ____rune_private_0d2ebf0f_typeScriptTextType(strings.TrimSpace((func() string { runes := []rune(__text); return string(runes[__colon+1 : len([]rune(__text))]) }())))
	}(), __line: 0, __column: 0}
}

func ____rune_private_0d2ebf0f_typeScriptTextType(__text string) string {
	return func() string {
		switch {
		case __text == "string":
			return "String"
		case __text == "boolean":
			return "Bool"
		case __text == "bigint":
			return "BigInt"
		case __text == "number":
			return "Double"
		case (__text == "void") || (__text == "undefined"):
			return "Void"
		default:
			return "Dynamic"
		}
	}()
}

func ____rune_private_0d2ebf0f_emptyCompilerIRFile() __IRFile {
	return __IRFile{__imports: []__IRImport{}, __tsImports: []__IRTSImport{}, __structs: []__IRStructType{}, __enums: []__IREnumType{}, __constants: []__IRConst{}, __functions: []__IRFunction{}, __tests: []__IRTest{}, __errors: []__ParseError{}}
}

func ____rune_private_0d2ebf0f_mergeCompilerIRFile(__out __IRFile, __file __IRFile) __IRFile {
	for _, __importDecl := range __file.__imports {
		_ = __importDecl
		func() int { __out.__imports = append(__out.__imports, __importDecl); return len(__out.__imports) }()
	}
	for _, __importDecl := range __file.__tsImports {
		_ = __importDecl
		func() int { __out.__tsImports = append(__out.__tsImports, __importDecl); return len(__out.__tsImports) }()
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		func() int { __out.__structs = append(__out.__structs, __typeDecl); return len(__out.__structs) }()
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		func() int { __out.__enums = append(__out.__enums, __typeDecl); return len(__out.__enums) }()
	}
	for _, __constant := range __file.__constants {
		_ = __constant
		func() int { __out.__constants = append(__out.__constants, __constant); return len(__out.__constants) }()
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		func() int { __out.__functions = append(__out.__functions, __fn); return len(__out.__functions) }()
	}
	for _, __testDecl := range __file.__tests {
		_ = __testDecl
		func() int { __out.__tests = append(__out.__tests, __testDecl); return len(__out.__tests) }()
	}
	for _, __error := range __file.__errors {
		_ = __error
		func() int { __out.__errors = append(__out.__errors, __error); return len(__out.__errors) }()
	}
	return __out
}

func ____rune_private_0d2ebf0f_compileResult(__ok bool, __output string, __errors []string) __CompileResult {
	return __CompileResult{__ok: __ok, __output: __output, __errors: __errors}
}

func ____rune_private_0d2ebf0f_unsupportedTargetErrors(__target string) []string {
	__errors := []string{}
	__errors = append(__errors, "unsupported target "+__target)
	return __errors
}

func ____rune_private_0d2ebf0f_parseErrorMessages(__errors []__ParseError) []string {
	__out := []string{}
	for _, __error := range __errors {
		_ = __error
		func() int {
			__out = append(__out, "line "+__compilerIntToString(__error.__line)+":"+__compilerIntToString(__error.__column)+": "+__error.__message)
			return len(__out)
		}()
	}
	return __out
}
