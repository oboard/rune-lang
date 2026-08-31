package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
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
	__module bool
	__line   int
	__column int
}

type __ParsedConst struct {
	__name    string
	__private bool
	__typeRef __ParsedTypeRef
	__value   __ParsedExpr
	__line    int
	__column  int
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
	__constants []__ParsedConst
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

type __ConstStep struct {
	__state     __ParserState
	__constDecl __ParsedConst
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
	__module bool
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
	__name       string
	__private    bool
	__typeName   string
	__jsonName   string
	__jsonIgnore bool
	__line       int
	__column     int
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
	__static       bool
	__routine      bool
	__macro        bool
	__receiverType string
	__generics     []string
	__params       []__IRParam
	__returnType   string
	__body         __IRExpr
	__sourcePath   string
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
	__name       string
	__private    bool
	__generics   []string
	__fields     []__IRField
	__methods    []__IRFunction
	__sourcePath string
	__line       int
	__column     int
}

type __IREnumType struct {
	__name       string
	__private    bool
	__generics   []string
	__members    []__IREnumMember
	__methods    []__IRFunction
	__sourcePath string
	__line       int
	__column     int
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

type __GoJSONDeclResult struct {
	__names []string
	__text  string
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

type __SelfhostSources struct {
	__files []__SourceFile
}

type __CompilerCallable struct {
	__name       string
	__arity      int
	__returnType string
	__paramTypes []string
	__private    bool
	__sourcePath string
}

type __CompilerMacroBinding struct {
	__name       string
	__macro      bool
	__paramTypes []string
}

type __CompilerTypeBinding struct {
	__name     string
	__typeName string
}

type __CompilerNamespaceAlias struct {
	__name       string
	__module     string
	__importPath string
	__go         bool
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

type runeUnit struct{}

type runeTask[T any] <-chan T

var runeTasks sync.WaitGroup

func runeGo[T any](work func() T) runeTask[T] {
	runeTasks.Add(1)
	ch := make(chan T, 1)
	go func() {
		defer runeTasks.Done()
		ch <- work()
	}()
	return ch
}

func runeWaitAll() {
	runeTasks.Wait()
}

func runeAwait[T any](task runeTask[T]) T {
	return <-task
}

type runeResult[T any, E any] struct {
	ok    bool
	value T
	err   E
}

func runeOk[T any, E any](value T) runeResult[T, E] {
	return runeResult[T, E]{ok: true, value: value}
}

func runeErr[T any, E any](err E) runeResult[T, E] {
	return runeResult[T, E]{err: err}
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

type runeFileStat struct {
	__size        int
	__isFile      bool
	__isDirectory bool
}

func runeReadFile(path string) runeTask[runeResult[[]byte, *runeError]] { return runeFsReadFile(path) }

func runeFsReadFile(path string) runeTask[runeResult[[]byte, *runeError]] {
	return runeGo(func() runeResult[[]byte, *runeError] {
		data, err := os.ReadFile(path)
		if err != nil {
			return runeErr[[]byte, *runeError](runeErrorFrom(err))
		}
		return runeOk[[]byte, *runeError](data)
	})
}

func runeFsReadFileText(path string) runeTask[runeResult[string, *runeError]] {
	return runeGo(func() runeResult[string, *runeError] {
		data, err := os.ReadFile(path)
		if err != nil {
			return runeErr[string, *runeError](runeErrorFrom(err))
		}
		return runeOk[string, *runeError](string(data))
	})
}

func runeFsWriteFile(path string, data []byte) runeTask[runeResult[struct{}, *runeError]] {
	return runeGo(func() runeResult[struct{}, *runeError] {
		if err := os.WriteFile(path, data, 0644); err != nil {
			return runeErr[struct{}, *runeError](runeErrorFrom(err))
		}
		return runeOk[struct{}, *runeError](struct{}{})
	})
}

func runeFsWriteFileText(path string, data string) runeTask[runeResult[struct{}, *runeError]] {
	return runeFsWriteFile(path, []byte(data))
}

func runeFsExists(path string) runeTask[runeResult[bool, *runeError]] {
	return runeGo(func() runeResult[bool, *runeError] {
		_, err := os.Stat(path)
		if err == nil {
			return runeOk[bool, *runeError](true)
		}
		if os.IsNotExist(err) {
			return runeOk[bool, *runeError](false)
		}
		return runeErr[bool, *runeError](runeErrorFrom(err))
	})
}

func runeFsReaddir(path string) runeTask[runeResult[[]string, *runeError]] {
	return runeGo(func() runeResult[[]string, *runeError] {
		entries, err := os.ReadDir(path)
		if err != nil {
			return runeErr[[]string, *runeError](runeErrorFrom(err))
		}
		names := make([]string, 0, len(entries))
		for _, entry := range entries {
			names = append(names, entry.Name())
		}
		return runeOk[[]string, *runeError](names)
	})
}

func runeFsMkdir(path string) runeTask[runeResult[struct{}, *runeError]] {
	return runeGo(func() runeResult[struct{}, *runeError] {
		if err := os.Mkdir(path, 0755); err != nil {
			return runeErr[struct{}, *runeError](runeErrorFrom(err))
		}
		return runeOk[struct{}, *runeError](struct{}{})
	})
}

func runeFsRemove(path string) runeTask[runeResult[struct{}, *runeError]] {
	return runeGo(func() runeResult[struct{}, *runeError] {
		if err := os.Remove(path); err != nil {
			return runeErr[struct{}, *runeError](runeErrorFrom(err))
		}
		return runeOk[struct{}, *runeError](struct{}{})
	})
}

func runeFsStat(path string) runeTask[runeResult[*runeFileStat, *runeError]] {
	return runeGo(func() runeResult[*runeFileStat, *runeError] {
		info, err := os.Stat(path)
		if err != nil {
			return runeErr[*runeFileStat, *runeError](runeErrorFrom(err))
		}
		stat := &runeFileStat{__size: int(info.Size()), __isFile: info.Mode().IsRegular(), __isDirectory: info.IsDir()}
		return runeOk[*runeFileStat, *runeError](stat)
	})
}

func __lex(__source string) []__Token {
	return ____rune_private_8093ad48_scan(__LexState{__source: __source, __start: 0, __current: 0, __line: 1, __column: 1, __startLine: 1, __startColumn: 1, __canStartRegex: true}, ____rune_private_8093ad48_emptyTokens())
}

func ____rune_private_8093ad48_emptyTokens() []__Token {
	return append([]__Token{}, []__Token{__Token{__kind: __TokenKind_EOF, __lexeme: "", __offset: 0, __line: 0, __column: 0}}[0:0]...)
}

func ____rune_private_8093ad48_scan(__state __LexState, __tokens []__Token) []__Token {
	__skipped := ____rune_private_8093ad48_skipIgnored(__state)
	__started := ____rune_private_8093ad48_markStart(__skipped)
	return func() []__Token {
		if ____rune_private_8093ad48_atEnd(__started) {
			return ____rune_private_8093ad48_appendToken(__tokens, ____rune_private_8093ad48_makeToken(__started, __TokenKind_EOF))
		}
		return ____rune_private_8093ad48_scanLexed(____rune_private_8093ad48_scanToken(____rune_private_8093ad48_advance(__started)), __tokens)
	}()
}

func ____rune_private_8093ad48_scanLexed(__lexed __Lexed, __tokens []__Token) []__Token {
	__nextTokens := ____rune_private_8093ad48_appendToken(__tokens, ____rune_private_8093ad48_makeToken(__lexed.__state, __lexed.__kind))
	return ____rune_private_8093ad48_scan(____rune_private_8093ad48_finishToken(__lexed.__state, __lexed.__kind), __nextTokens)
}

func ____rune_private_8093ad48_appendToken(__tokens []__Token, __token __Token) []__Token {
	__tokens = append(__tokens, __token)
	return __tokens
}

func ____rune_private_8093ad48_makeToken(__state __LexState, __kind __TokenKind) __Token {
	return __Token{__kind: __kind, __lexeme: func() string {
		runes := []rune(__state.__source)
		return string(runes[__state.__start:__state.__current])
	}(), __offset: __state.__start, __line: __state.__startLine, __column: __state.__startColumn}
}

func ____rune_private_8093ad48_finishToken(__state __LexState, __kind __TokenKind) __LexState {
	return __LexState{__source: __state.__source, __start: __state.__start, __current: __state.__current, __line: __state.__line, __column: __state.__column, __startLine: __state.__startLine, __startColumn: __state.__startColumn, __canStartRegex: !____rune_private_8093ad48_canEndExpression(__state, __kind)}
}

func ____rune_private_8093ad48_canEndExpression(__state __LexState, __kind __TokenKind) bool {
	return ____rune_private_8093ad48_canEndValueToken(__kind) || ____rune_private_8093ad48_canEndXmlLess(__state, __kind)
}

func ____rune_private_8093ad48_canEndValueToken(__kind __TokenKind) bool {
	return func() bool {
		switch {
		case (__kind == __TokenKind_Ident) || (__kind == __TokenKind_Int) || (__kind == __TokenKind_Double) || (__kind == __TokenKind_BigInt) || (__kind == __TokenKind_String) || (__kind == __TokenKind_TemplateString) || (__kind == __TokenKind_Char) || (__kind == __TokenKind_Regex) || (__kind == __TokenKind_XMLText) || (__kind == __TokenKind_RParen) || (__kind == __TokenKind_RBracket) || (__kind == __TokenKind_RBrace):
			return true
		default:
			return false
		}
	}()
}

func ____rune_private_8093ad48_canEndXmlLess(__state __LexState, __kind __TokenKind) bool {
	return __kind == __TokenKind_Less && (____rune_private_8093ad48_peek(__state) == '/' || ____rune_private_8093ad48_isIdentStart(____rune_private_8093ad48_peek(__state)))
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

func ____rune_private_8093ad48_markStart(__state __LexState) __LexState {
	return __LexState{__source: __state.__source, __start: __state.__current, __current: __state.__current, __line: __state.__line, __column: __state.__column, __startLine: __state.__line, __startColumn: __state.__column, __canStartRegex: __state.__canStartRegex}
}

func ____rune_private_8093ad48_atEnd(__state __LexState) bool {
	return __state.__current >= len([]rune(__state.__source))
}

func ____rune_private_8093ad48_charAt(__source string, __index int) rune {
	return func() rune {
		if __index < 0 || __index >= len([]rune(__source)) {
			return ' '
		}
		return []rune(__source)[__index]
	}()
}

func ____rune_private_8093ad48_peek(__state __LexState) rune {
	return ____rune_private_8093ad48_charAt(__state.__source, __state.__current)
}

func ____rune_private_8093ad48_peekNext(__state __LexState) rune {
	return ____rune_private_8093ad48_charAt(__state.__source, __state.__current+1)
}

func ____rune_private_8093ad48_advanceState(__state __LexState) __LexState {
	return ____rune_private_8093ad48_advance(__state).__state
}

func ____rune_private_8093ad48_advance(__state __LexState) __Advanced {
	return func() __Advanced {
		if ____rune_private_8093ad48_atEnd(__state) {
			return ____rune_private_8093ad48_advanced(__state, ' ')
		}
		return ____rune_private_8093ad48_advanceChar(__state, []rune(__state.__source)[__state.__current])
	}()
}

func ____rune_private_8093ad48_advanced(__state __LexState, __ch rune) __Advanced {
	return __Advanced{__state: __state, __ch: __ch}
}

func ____rune_private_8093ad48_advanceChar(__state __LexState, __ch rune) __Advanced {
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

func ____rune_private_8093ad48_skipIgnored(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_8093ad48_atEnd(__state) {
			return __state
		}
		return func() __LexState {
			if ____rune_private_8093ad48_isSpace(____rune_private_8093ad48_peek(__state)) {
				return ____rune_private_8093ad48_skipIgnored(____rune_private_8093ad48_advanceState(__state))
			}
			return func() __LexState {
				if ____rune_private_8093ad48_startsWith(__state, '/', '/') {
					return ____rune_private_8093ad48_skipIgnored(____rune_private_8093ad48_skipLineComment(____rune_private_8093ad48_advanceState(____rune_private_8093ad48_advanceState(__state))))
				}
				return func() __LexState {
					if ____rune_private_8093ad48_startsWith(__state, '/', '*') {
						return ____rune_private_8093ad48_skipIgnored(____rune_private_8093ad48_skipBlockComment(____rune_private_8093ad48_advanceState(____rune_private_8093ad48_advanceState(__state))))
					}
					return __state
				}()
			}()
		}()
	}()
}

func ____rune_private_8093ad48_startsWith(__state __LexState, __first rune, __second rune) bool {
	return ____rune_private_8093ad48_peek(__state) == __first && ____rune_private_8093ad48_peekNext(__state) == __second
}

func ____rune_private_8093ad48_skipLineComment(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_8093ad48_atEnd(__state) || ____rune_private_8093ad48_peek(__state) == '\n' {
			return __state
		}
		return ____rune_private_8093ad48_skipLineComment(____rune_private_8093ad48_advanceState(__state))
	}()
}

func ____rune_private_8093ad48_skipBlockComment(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_8093ad48_atEnd(__state) {
			return __state
		}
		return func() __LexState {
			if ____rune_private_8093ad48_startsWith(__state, '*', '/') {
				return ____rune_private_8093ad48_advanceState(____rune_private_8093ad48_advanceState(__state))
			}
			return ____rune_private_8093ad48_skipBlockComment(____rune_private_8093ad48_advanceState(__state))
		}()
	}()
}

func ____rune_private_8093ad48_isSpace(__ch rune) bool {
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

func ____rune_private_8093ad48_scanToken(__step __Advanced) __Lexed {
	__destructure1 := __step
	__state := __destructure1.__state
	__ch := __destructure1.__ch
	return func() __Lexed {
		switch {
		case __ch == '\n':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Newline)
		case __ch == '#':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Hash)
		case __ch == '@':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_At)
		case __ch == '$':
			return ____rune_private_8093ad48_lexDollar(__state)
		case __ch == '.':
			return ____rune_private_8093ad48_lexDot(__state)
		case __ch == ',':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Comma)
		case __ch == ':':
			return ____rune_private_8093ad48_lexColon(__state)
		case __ch == '(':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_LParen)
		case __ch == ')':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_RParen)
		case __ch == '[':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_LBracket)
		case __ch == ']':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_RBracket)
		case __ch == '{':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_LBrace)
		case __ch == '}':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_RBrace)
		case __ch == '?':
			return ____rune_private_8093ad48_lexQuestion(__state)
		case __ch == '+':
			return ____rune_private_8093ad48_lexPlus(__state)
		case __ch == '-':
			return ____rune_private_8093ad48_lexMinus(__state)
		case __ch == '*':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Star)
		case __ch == '/':
			return func() __Lexed {
				if __state.__canStartRegex {
					return ____rune_private_8093ad48_lexRegexToken(__state)
				}
				return ____rune_private_8093ad48_lexed(__state, __TokenKind_Slash)
			}()
		case __ch == '%':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Percent)
		case __ch == '!':
			return ____rune_private_8093ad48_lexBang(__state)
		case __ch == '~':
			return ____rune_private_8093ad48_lexTilde(__state)
		case __ch == '&':
			return ____rune_private_8093ad48_lexAmp(__state)
		case __ch == '|':
			return ____rune_private_8093ad48_lexPipe(__state)
		case __ch == '^':
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_BitXor)
		case __ch == '=':
			return ____rune_private_8093ad48_lexEqual(__state)
		case __ch == '<':
			return ____rune_private_8093ad48_lexLess(__state)
		case __ch == '>':
			return ____rune_private_8093ad48_lexGreater(__state)
		case __ch == '"':
			return ____rune_private_8093ad48_lexStringToken(__state)
		case __ch == '`':
			return ____rune_private_8093ad48_lexTemplateStringToken(__state)
		case __ch == '\'':
			return func() __Lexed {
				if __state.__canStartRegex {
					return ____rune_private_8093ad48_lexCharToken(__state)
				}
				return ____rune_private_8093ad48_lexed(__state, __TokenKind_Apostrophe)
			}()
		case __ch == '_':
			return func() __Lexed {
				if ____rune_private_8093ad48_isIdentContinue(____rune_private_8093ad48_peek(__state)) {
					return ____rune_private_8093ad48_lexIdentifierToken(__state)
				}
				return ____rune_private_8093ad48_lexed(__state, __TokenKind_Underscore)
			}()
		default:
			return func() __Lexed {
				if ____rune_private_8093ad48_isDigit(__ch) {
					return ____rune_private_8093ad48_lexNumberToken(__state)
				}
				return func() __Lexed {
					if ____rune_private_8093ad48_isIdentStart(__ch) {
						return ____rune_private_8093ad48_lexIdentifierToken(__state)
					}
					return ____rune_private_8093ad48_lexed(__state, __TokenKind_Illegal)
				}()
			}()
		}
	}()
}

func ____rune_private_8093ad48_lexed(__state __LexState, __kind __TokenKind) __Lexed {
	return __Lexed{__state: __state, __kind: __kind}
}

func ____rune_private_8093ad48_lexDot(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '.':
			return ____rune_private_8093ad48_lexDotDot(____rune_private_8093ad48_advanceState(__state))
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Dot)
		}
	}()
}

func ____rune_private_8093ad48_lexDotDot(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '.':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_DotDotDot)
		case ____rune_private_8093ad48_peek(__state) == '<':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_DotDotLess)
		case ____rune_private_8093ad48_peek(__state) == '=':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_DotDotEqual)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_DotDot)
		}
	}()
}

func ____rune_private_8093ad48_lexColon(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '=':
			return func() __Lexed {
				if ____rune_private_8093ad48_peek(____rune_private_8093ad48_advanceState(__state)) == ':' {
					return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(____rune_private_8093ad48_advanceState(__state)), __TokenKind_MutDeclare)
				}
				return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_Declare)
			}()
		case ____rune_private_8093ad48_peek(__state) == ':':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_DoubleColon)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Colon)
		}
	}()
}

func ____rune_private_8093ad48_lexQuestion(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '?':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_QuestionQuestion)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Question)
		}
	}()
}

func ____rune_private_8093ad48_lexPlus(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '+':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_PlusPlus)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Plus)
		}
	}()
}

func ____rune_private_8093ad48_lexMinus(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '>':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_Arrow)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Minus)
		}
	}()
}

func ____rune_private_8093ad48_lexBang(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '=':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_BangEqual)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Bang)
		}
	}()
}

func ____rune_private_8093ad48_lexTilde(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Tilde)
		}
	}()
}

func ____rune_private_8093ad48_lexDollar(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Dollar)
		}
	}()
}

func ____rune_private_8093ad48_lexAmp(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '&':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_AndAnd)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_BitAnd)
		}
	}()
}

func ____rune_private_8093ad48_lexPipe(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '|':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_OrOr)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_BitOr)
		}
	}()
}

func ____rune_private_8093ad48_lexEqual(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '>':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_FatArrow)
		case ____rune_private_8093ad48_peek(__state) == '=':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_EqualEqual)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Assign)
		}
	}()
}

func ____rune_private_8093ad48_lexLess(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '=':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_LessEqual)
		case ____rune_private_8093ad48_peek(__state) == '<':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_ShiftLeft)
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Less)
		}
	}()
}

func ____rune_private_8093ad48_lexGreater(__state __LexState) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == '=':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_GreaterEqual)
		case ____rune_private_8093ad48_peek(__state) == '>':
			return func() __Lexed {
				if ____rune_private_8093ad48_startsWith(__state, '>', '>') {
					return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(____rune_private_8093ad48_advanceState(__state)), __TokenKind_UnsignedShiftRight)
				}
				return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_ShiftRight)
			}()
		default:
			return ____rune_private_8093ad48_lexed(__state, __TokenKind_Greater)
		}
	}()
}

func ____rune_private_8093ad48_lexStringToken(__state __LexState) __Lexed {
	__scanned := ____rune_private_8093ad48_scanString(__state, false)
	return ____rune_private_8093ad48_lexed(__scanned.__state, func() __TokenKind {
		if __scanned.__ok {
			return __TokenKind_String
		}
		return __TokenKind_Illegal
	}())
}

func ____rune_private_8093ad48_lexTemplateStringToken(__state __LexState) __Lexed {
	__scanned := ____rune_private_8093ad48_scanTemplateString(__state, false)
	return ____rune_private_8093ad48_lexed(__scanned.__state, func() __TokenKind {
		if __scanned.__ok {
			return __TokenKind_TemplateString
		}
		return __TokenKind_Illegal
	}())
}

func ____rune_private_8093ad48_scanString(__state __LexState, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if ____rune_private_8093ad48_atEnd(__state) {
			return ____rune_private_8093ad48_scannedString(__state, false)
		}
		return ____rune_private_8093ad48_scanStringStep(____rune_private_8093ad48_advance(__state), __escaped)
	}()
}

func ____rune_private_8093ad48_scanTemplateString(__state __LexState, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if ____rune_private_8093ad48_atEnd(__state) {
			return ____rune_private_8093ad48_scannedString(__state, false)
		}
		return ____rune_private_8093ad48_scanTemplateStringStep(____rune_private_8093ad48_advance(__state), __escaped)
	}()
}

func ____rune_private_8093ad48_scannedString(__state __LexState, __ok bool) __ScannedString {
	return __ScannedString{__state: __state, __ok: __ok}
}

func ____rune_private_8093ad48_scanStringStep(__step __Advanced, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if __escaped {
			return ____rune_private_8093ad48_scanString(__step.__state, false)
		}
		return func() __ScannedString {
			switch {
			case __step.__ch == '\\':
				return ____rune_private_8093ad48_scanString(__step.__state, true)
			case __step.__ch == '"':
				return ____rune_private_8093ad48_scannedString(__step.__state, true)
			default:
				return ____rune_private_8093ad48_scanString(__step.__state, false)
			}
		}()
	}()
}

func ____rune_private_8093ad48_scanTemplateStringStep(__step __Advanced, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if __escaped {
			return ____rune_private_8093ad48_scanTemplateString(__step.__state, false)
		}
		return func() __ScannedString {
			switch {
			case __step.__ch == '\\':
				return ____rune_private_8093ad48_scanTemplateString(__step.__state, true)
			case __step.__ch == '`':
				return ____rune_private_8093ad48_scannedString(__step.__state, true)
			default:
				return ____rune_private_8093ad48_scanTemplateString(__step.__state, false)
			}
		}()
	}()
}

func ____rune_private_8093ad48_lexCharToken(__state __LexState) __Lexed {
	__scanned := ____rune_private_8093ad48_scanChar(__state, false)
	return ____rune_private_8093ad48_lexed(__scanned.__state, func() __TokenKind {
		if __scanned.__ok {
			return __TokenKind_Char
		}
		return __TokenKind_Illegal
	}())
}

func ____rune_private_8093ad48_scanChar(__state __LexState, __escaped bool) __ScannedString {
	return func() __ScannedString {
		if ____rune_private_8093ad48_atEnd(__state) {
			return ____rune_private_8093ad48_scannedString(__state, false)
		}
		return ____rune_private_8093ad48_scanCharStep(____rune_private_8093ad48_advance(__state), __escaped)
	}()
}

func ____rune_private_8093ad48_scanCharStep(__step __Advanced, __escaped bool) __ScannedString {
	return func() __ScannedString {
		switch {
		case __step.__ch == '\n':
			return ____rune_private_8093ad48_scannedString(__step.__state, false)
		default:
			return func() __ScannedString {
				if __escaped {
					return ____rune_private_8093ad48_scanChar(__step.__state, false)
				}
				return func() __ScannedString {
					switch {
					case __step.__ch == '\\':
						return ____rune_private_8093ad48_scanChar(__step.__state, true)
					case __step.__ch == '\'':
						return ____rune_private_8093ad48_scannedString(__step.__state, true)
					default:
						return ____rune_private_8093ad48_scanChar(__step.__state, false)
					}
				}()
			}()
		}
	}()
}

func ____rune_private_8093ad48_lexRegexToken(__state __LexState) __Lexed {
	__scanned := ____rune_private_8093ad48_scanRegex(__state, false, false)
	return ____rune_private_8093ad48_lexed(__scanned.__state, func() __TokenKind {
		if __scanned.__ok {
			return __TokenKind_Regex
		}
		return __TokenKind_Illegal
	}())
}

func ____rune_private_8093ad48_scanRegex(__state __LexState, __escaped bool, __inClass bool) __ScannedString {
	return func() __ScannedString {
		if ____rune_private_8093ad48_atEnd(__state) {
			return ____rune_private_8093ad48_scannedString(__state, false)
		}
		return ____rune_private_8093ad48_scanRegexStep(____rune_private_8093ad48_advance(__state), __escaped, __inClass)
	}()
}

func ____rune_private_8093ad48_scanRegexStep(__step __Advanced, __escaped bool, __inClass bool) __ScannedString {
	return func() __ScannedString {
		switch {
		case __step.__ch == '\n':
			return ____rune_private_8093ad48_scannedString(__step.__state, false)
		default:
			return func() __ScannedString {
				if __escaped {
					return ____rune_private_8093ad48_scanRegex(__step.__state, false, __inClass)
				}
				return func() __ScannedString {
					switch {
					case __step.__ch == '\\':
						return ____rune_private_8093ad48_scanRegex(__step.__state, true, __inClass)
					case __step.__ch == '[':
						return ____rune_private_8093ad48_scanRegex(__step.__state, false, true)
					case __step.__ch == ']':
						return ____rune_private_8093ad48_scanRegex(__step.__state, false, false)
					case __step.__ch == '/':
						return func() __ScannedString {
							if __inClass {
								return ____rune_private_8093ad48_scanRegex(__step.__state, false, __inClass)
							}
							return ____rune_private_8093ad48_scannedString(____rune_private_8093ad48_scanRegexFlags(__step.__state), true)
						}()
					default:
						return ____rune_private_8093ad48_scanRegex(__step.__state, false, __inClass)
					}
				}()
			}()
		}
	}()
}

func ____rune_private_8093ad48_scanRegexFlags(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_8093ad48_isRegexFlag(____rune_private_8093ad48_peek(__state)) {
			return ____rune_private_8093ad48_scanRegexFlags(____rune_private_8093ad48_advanceState(__state))
		}
		return __state
	}()
}

func ____rune_private_8093ad48_lexNumberToken(__state __LexState) __Lexed {
	return ____rune_private_8093ad48_lexNumberAfterDigits(____rune_private_8093ad48_scanDigits(__state), false)
}

func ____rune_private_8093ad48_scanDigits(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_8093ad48_isDigit(____rune_private_8093ad48_peek(__state)) {
			return ____rune_private_8093ad48_scanDigits(____rune_private_8093ad48_advanceState(__state))
		}
		return __state
	}()
}

func ____rune_private_8093ad48_lexNumberAfterDigits(__state __LexState, __isDouble bool) __Lexed {
	return func() __Lexed {
		switch {
		case ____rune_private_8093ad48_peek(__state) == 'n':
			return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_advanceState(__state), __TokenKind_BigInt)
		case ____rune_private_8093ad48_peek(__state) == '.':
			return func() __Lexed {
				if ____rune_private_8093ad48_isDigit(____rune_private_8093ad48_peekNext(__state)) {
					return ____rune_private_8093ad48_lexNumberAfterDot(____rune_private_8093ad48_advanceState(__state))
				}
				return ____rune_private_8093ad48_lexed(__state, func() __TokenKind {
					if __isDouble {
						return __TokenKind_Double
					}
					return __TokenKind_Int
				}())
			}()
		default:
			return func() __Lexed {
				if ____rune_private_8093ad48_isExponentMarker(____rune_private_8093ad48_peek(__state)) {
					return ____rune_private_8093ad48_lexNumberAfterExponent(____rune_private_8093ad48_advanceState(__state))
				}
				return ____rune_private_8093ad48_lexed(__state, func() __TokenKind {
					if __isDouble {
						return __TokenKind_Double
					}
					return __TokenKind_Int
				}())
			}()
		}
	}()
}

func ____rune_private_8093ad48_lexNumberAfterDot(__state __LexState) __Lexed {
	return ____rune_private_8093ad48_lexNumberAfterDigits(____rune_private_8093ad48_scanDigits(__state), true)
}

func ____rune_private_8093ad48_lexNumberAfterExponent(__state __LexState) __Lexed {
	return func() __Lexed {
		if ____rune_private_8093ad48_isExponentSign(____rune_private_8093ad48_peek(__state)) {
			return ____rune_private_8093ad48_lexNumberExponentDigits(____rune_private_8093ad48_advanceState(__state))
		}
		return ____rune_private_8093ad48_lexNumberExponentDigits(__state)
	}()
}

func ____rune_private_8093ad48_lexNumberExponentDigits(__state __LexState) __Lexed {
	return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_scanDigits(__state), __TokenKind_Double)
}

func ____rune_private_8093ad48_lexIdentifierToken(__state __LexState) __Lexed {
	return ____rune_private_8093ad48_lexed(____rune_private_8093ad48_scanIdentifier(__state), __TokenKind_Ident)
}

func ____rune_private_8093ad48_scanIdentifier(__state __LexState) __LexState {
	return func() __LexState {
		if ____rune_private_8093ad48_isIdentContinue(____rune_private_8093ad48_peek(__state)) {
			return ____rune_private_8093ad48_scanIdentifier(____rune_private_8093ad48_advanceState(__state))
		}
		return __state
	}()
}

func ____rune_private_8093ad48_isDigit(__ch rune) bool {
	switch {
	case (__ch >= '0' && __ch <= '9'):
		return true
	default:
		return false
	}
}

func ____rune_private_8093ad48_isAlpha(__ch rune) bool {
	switch {
	case (__ch >= 'a' && __ch <= 'z'):
		return true
	case (__ch >= 'A' && __ch <= 'Z'):
		return true
	default:
		return false
	}
}

func ____rune_private_8093ad48_isIdentStart(__ch rune) bool {
	switch {
	case __ch == '_':
		return true
	default:
		return ____rune_private_8093ad48_isIdentText(__ch) && ____rune_private_8093ad48_isDigit(__ch) == false
	}
}

func ____rune_private_8093ad48_isIdentContinue(__ch rune) bool {
	return ____rune_private_8093ad48_isIdentStart(__ch) || ____rune_private_8093ad48_isDigit(__ch)
}

func ____rune_private_8093ad48_isIdentText(__ch rune) bool {
	return ____rune_private_8093ad48_isIdentBoundary(__ch) == false && ____rune_private_8093ad48_isAsciiPunctuation(__ch) == false
}

func ____rune_private_8093ad48_isIdentBoundary(__ch rune) bool {
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

func ____rune_private_8093ad48_isAsciiPunctuation(__ch rune) bool {
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

func ____rune_private_8093ad48_isRegexFlag(__ch rune) bool {
	return ____rune_private_8093ad48_isAlpha(__ch)
}

func ____rune_private_8093ad48_isExponentMarker(__ch rune) bool {
	switch {
	case __ch == 'e':
		return true
	case __ch == 'E':
		return true
	default:
		return false
	}
}

func ____rune_private_8093ad48_isExponentSign(__ch rune) bool {
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
			return ____rune_private_b99e97e3_typeNameToString(__typeRef)
		case __typeRef.__kind == __TypeRefKind_Group:
			return func() string {
				if len(__typeRef.__args) == 0 {
					return "()"
				}
				return "(" + __typeRefToString(__typeRef.__args[0]) + ")"
			}()
		case __typeRef.__kind == __TypeRefKind_Tuple:
			return "(" + ____rune_private_b99e97e3_typeParamsToString(__typeRef.__params, 0, "") + ")"
		case __typeRef.__kind == __TypeRefKind_Function:
			return ____rune_private_b99e97e3_functionTypeToString(__typeRef)
		default:
			return ""
		}
	}()
}

func ____rune_private_b99e97e3_typeNameToString(__typeRef __ParsedTypeRef) string {
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
		return "[" + ____rune_private_b99e97e3_typeRefsToString(__typeRef.__args, 0, "") + "]"
	}()
	__nullable := func() string {
		if __typeRef.__nullable {
			return "?"
		}
		return ""
	}()
	return __prefix + __typeRef.__name + __args + __nullable
}

func ____rune_private_b99e97e3_functionTypeToString(__typeRef __ParsedTypeRef) string {
	__ret := func() string {
		if len(__typeRef.__returnTypes) == 0 {
			return ""
		}
		return __typeRefToString(__typeRef.__returnTypes[0])
	}()
	return "(" + ____rune_private_b99e97e3_typeParamsToString(__typeRef.__params, 0, "") + ")->" + __ret
}

func ____rune_private_b99e97e3_typeParamToString(__param __ParsedTypeParam) string {
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

func ____rune_private_b99e97e3_typeRefsToString(__refs []__ParsedTypeRef, __index int, __out string) string {
	return func() string {
		if __index >= len(__refs) {
			return __out
		}
		return ____rune_private_b99e97e3_typeRefsToString(__refs, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ","
		}()+__typeRefToString(__refs[__index]))
	}()
}

func ____rune_private_b99e97e3_typeParamsToString(__params []__ParsedTypeParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_b99e97e3_typeParamsToString(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ","
		}()+____rune_private_b99e97e3_typeParamToString(__params[__index]))
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
	__errors := ____rune_private_d1a61e34_emptyParseErrors()
	return ____rune_private_d1a61e34_parseFileLoop(____rune_private_d1a61e34_parserSkipNewlines(__ParserState{__tokens: __tokens, __current: 0, __errors: __errors}), ____rune_private_d1a61e34_emptyFile(__errors)).__file
}

func ____rune_private_d1a61e34_emptyFile(__errors []__ParseError) __ParsedFile {
	return __ParsedFile{__imports: []__ParsedImport{}, __constants: []__ParsedConst{}, __types: []__ParsedType{}, __functions: []__ParsedFunction{}, __tests: []__ParsedTest{}, __errors: __errors}
}

func ____rune_private_d1a61e34_emptyParseErrors() []__ParseError {
	return append([]__ParseError{}, []__ParseError{__ParseError{__message: "", __line: 0, __column: 0}}[0:0]...)
}

func ____rune_private_d1a61e34_emptyToken() __Token {
	return __Token{__kind: __TokenKind_EOF, __lexeme: "", __offset: 0, __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_emptyExpr() __ParsedExpr {
	return __ParsedExpr{__kind: __ExprKind_Unknown, __text: "", __name: "", __value: "", __op: "", __params: []__ParsedParam{}, __children: []__ParsedExpr{}, __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_emptyAnnotations() []__ParsedAnnotation {
	return append([]__ParsedAnnotation{}, []__ParsedAnnotation{__ParsedAnnotation{__marker: "", __module: "", __name: "", __args: []__ParsedExpr{}, __line: 0, __column: 0}}[0:0]...)
}

func ____rune_private_d1a61e34_emptyFunction() __ParsedFunction {
	return __ParsedFunction{__name: "", __private: true, __static: false, __routine: false, __macro: false, __annotations: ____rune_private_d1a61e34_emptyAnnotations(), __receiverType: "", __generics: []string{}, __params: []__ParsedParam{}, __returnType: __emptyParsedTypeRef(), __body: ____rune_private_d1a61e34_emptyExpr(), __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_emptyType() __ParsedType {
	return __ParsedType{__name: "", __private: true, __enum: false, __annotations: ____rune_private_d1a61e34_emptyAnnotations(), __generics: []string{}, __fields: []__ParsedField{}, __methods: []__ParsedFunction{}, __members: []__ParsedEnumMember{}, __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_emptyImport() __ParsedImport {
	return __ParsedImport{__path: "", __go: false, __module: false, __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_emptyConst() __ParsedConst {
	return __ParsedConst{__name: "", __private: true, __typeRef: __emptyParsedTypeRef(), __value: ____rune_private_d1a61e34_emptyExpr(), __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_emptyTest() __ParsedTest {
	return __ParsedTest{__name: "", __body: ____rune_private_d1a61e34_emptyExpr(), __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_emptyField() __ParsedField {
	return __ParsedField{__name: "", __private: true, __annotations: ____rune_private_d1a61e34_emptyAnnotations(), __typeRef: __emptyParsedTypeRef(), __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_emptyMember() __ParsedEnumMember {
	return __ParsedEnumMember{__name: "", __private: true, __annotations: ____rune_private_d1a61e34_emptyAnnotations(), __value: "", __params: []__ParsedParam{}, __line: 0, __column: 0}
}

func ____rune_private_d1a61e34_makeExpr(__kind __ExprKind, __text string, __name string, __value string, __op string, __params []__ParsedParam, __children []__ParsedExpr, __line int, __column int) __ParsedExpr {
	return __ParsedExpr{__kind: __kind, __text: __text, __name: __name, __value: __value, __op: __op, __params: __params, __children: __children, __line: __line, __column: __column}
}

func ____rune_private_d1a61e34_node(__kind __ExprKind, __token __Token) __ParsedExpr {
	return ____rune_private_d1a61e34_makeExpr(__kind, __token.__lexeme, "", "", "", []__ParsedParam{}, []__ParsedExpr{}, __token.__line, __token.__column)
}

func ____rune_private_d1a61e34_namedNode(__kind __ExprKind, __name string, __token __Token) __ParsedExpr {
	return ____rune_private_d1a61e34_makeExpr(__kind, __token.__lexeme, __name, "", "", []__ParsedParam{}, []__ParsedExpr{}, __token.__line, __token.__column)
}

func ____rune_private_d1a61e34_valueNode(__kind __ExprKind, __value string, __token __Token) __ParsedExpr {
	return ____rune_private_d1a61e34_makeExpr(__kind, __token.__lexeme, "", __value, "", []__ParsedParam{}, []__ParsedExpr{}, __token.__line, __token.__column)
}

func ____rune_private_d1a61e34_opNode(__kind __ExprKind, __op string, __token __Token, __children []__ParsedExpr) __ParsedExpr {
	return ____rune_private_d1a61e34_makeExpr(__kind, __token.__lexeme, "", "", __op, []__ParsedParam{}, __children, __token.__line, __token.__column)
}

func ____rune_private_d1a61e34_withChildren(__expr __ParsedExpr, __children []__ParsedExpr) __ParsedExpr {
	return ____rune_private_d1a61e34_makeExpr(__expr.__kind, __expr.__text, __expr.__name, __expr.__value, __expr.__op, __expr.__params, __children, __expr.__line, __expr.__column)
}

func ____rune_private_d1a61e34_withParams(__expr __ParsedExpr, __params []__ParsedParam) __ParsedExpr {
	return ____rune_private_d1a61e34_makeExpr(__expr.__kind, __expr.__text, __expr.__name, __expr.__value, __expr.__op, __params, __expr.__children, __expr.__line, __expr.__column)
}

func ____rune_private_d1a61e34_withText(__expr __ParsedExpr, __text string) __ParsedExpr {
	return ____rune_private_d1a61e34_makeExpr(__expr.__kind, __text, __expr.__name, __expr.__value, __expr.__op, __expr.__params, __expr.__children, __expr.__line, __expr.__column)
}

func ____rune_private_d1a61e34_appendChild(__expr __ParsedExpr, __child __ParsedExpr) __ParsedExpr {
	__expr.__children = append(__expr.__children, __child)
	return __expr
}

func ____rune_private_d1a61e34_appendParam(__params []__ParsedParam, __param __ParsedParam) []__ParsedParam {
	__params = append(__params, __param)
	return __params
}

func ____rune_private_d1a61e34_appendString(__values []string, __value string) []string {
	__values = append(__values, __value)
	return __values
}

func ____rune_private_d1a61e34_parserPeek(__state __ParserState) __Token {
	return func() __Token {
		if __state.__current >= len(__state.__tokens) {
			return ____rune_private_d1a61e34_emptyToken()
		}
		return __state.__tokens[__state.__current]
	}()
}

func ____rune_private_d1a61e34_parserPrevious(__state __ParserState) __Token {
	return func() __Token {
		if __state.__current <= 0 {
			return ____rune_private_d1a61e34_emptyToken()
		}
		return __state.__tokens[__state.__current-1]
	}()
}

func ____rune_private_d1a61e34_parserTokenAt(__state __ParserState, __index int) __Token {
	return func() __Token {
		if __index >= len(__state.__tokens) {
			return ____rune_private_d1a61e34_emptyToken()
		}
		return __state.__tokens[__index]
	}()
}

func ____rune_private_d1a61e34_parserKindAt(__state __ParserState, __index int) __TokenKind {
	return ____rune_private_d1a61e34_parserTokenAt(__state, __index).__kind
}

func ____rune_private_d1a61e34_parserCheck(__state __ParserState, __kind __TokenKind) bool {
	return ____rune_private_d1a61e34_parserPeek(__state).__kind == __kind
}

func ____rune_private_d1a61e34_parserCheckNext(__state __ParserState, __kind __TokenKind) bool {
	return ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+1) == __kind
}

func ____rune_private_d1a61e34_stateAt(__state __ParserState, __current int) __ParserState {
	return __ParserState{__tokens: __state.__tokens, __current: __current, __errors: __state.__errors}
}

func ____rune_private_d1a61e34_parserAdvance(__state __ParserState) __TokenStep {
	return func() __TokenStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_EOF) {
			return __TokenStep{__state: __state, __token: ____rune_private_d1a61e34_parserPeek(__state)}
		}
		return __TokenStep{__state: ____rune_private_d1a61e34_stateAt(__state, __state.__current+1), __token: ____rune_private_d1a61e34_parserPeek(__state)}
	}()
}

func ____rune_private_d1a61e34_parserMatch(__state __ParserState, __kind __TokenKind) __BoolStep {
	return func() __BoolStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __kind) {
			return __BoolStep{__state: ____rune_private_d1a61e34_parserAdvance(__state).__state, __ok: true}
		}
		return __BoolStep{__state: __state, __ok: false}
	}()
}

func ____rune_private_d1a61e34_parserConsume(__state __ParserState, __kind __TokenKind, __message string) __TokenStep {
	return func() __TokenStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __kind) {
			return ____rune_private_d1a61e34_parserAdvance(__state)
		}
		return ____rune_private_d1a61e34_parserConsumeMissing(____rune_private_d1a61e34_parserErrorAt(__state, ____rune_private_d1a61e34_parserPeek(__state), __message))
	}()
}

func ____rune_private_d1a61e34_parserConsumeMissing(__state __ParserState) __TokenStep {
	return func() __TokenStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_EOF) {
			return __TokenStep{__state: __state, __token: ____rune_private_d1a61e34_parserPeek(__state)}
		}
		return ____rune_private_d1a61e34_parserAdvance(__state)
	}()
}

func ____rune_private_d1a61e34_parserErrorAt(__state __ParserState, __token __Token, __message string) __ParserState {
	return __ParserState{__tokens: __state.__tokens, __current: __state.__current, __errors: ____rune_private_d1a61e34_appendParseError(__state.__errors, __ParseError{__message: __message, __line: __token.__line, __column: __token.__column})}
}

func ____rune_private_d1a61e34_appendParseError(__errors []__ParseError, __error __ParseError) []__ParseError {
	__out := __errors
	__out = append(__out, __error)
	return __out
}

func ____rune_private_d1a61e34_parserSkipNewlines(__state __ParserState) __ParserState {
	return func() __ParserState {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Newline) {
			return ____rune_private_d1a61e34_parserSkipNewlines(____rune_private_d1a61e34_parserAdvance(__state).__state)
		}
		return __state
	}()
}

func ____rune_private_d1a61e34_consumeStatementEnd(__state __ParserState) __ParserState {
	return func() __ParserState {
		if ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_Newline).__ok {
			return ____rune_private_d1a61e34_parserSkipNewlines(____rune_private_d1a61e34_parserAdvance(__state).__state)
		}
		return __state
	}()
}

func ____rune_private_d1a61e34_consumeFieldSeparator(__state __ParserState, __close __TokenKind, __message string) __ParserState {
	__current := ____rune_private_d1a61e34_consumeStatementEnd(__state)
	__comma := ____rune_private_d1a61e34_parserMatch(__current, __TokenKind_Comma)
	return func() __ParserState {
		if ____rune_private_d1a61e34_parserCheck(__current, __close) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return __current
		}
		return func() __ParserState {
			if __comma.__ok {
				return ____rune_private_d1a61e34_parserSkipNewlines(__comma.__state)
			}
			return ____rune_private_d1a61e34_parserErrorAt(__current, ____rune_private_d1a61e34_parserPeek(__current), __message)
		}()
	}()
}

func ____rune_private_d1a61e34_unquote(__raw string) string {
	return func() string {
		if len([]rune(__raw)) >= 2 {
			return func() string { runes := []rune(__raw); return string(runes[1 : len([]rune(__raw))-1]) }()
		}
		return __raw
	}()
}

func ____rune_private_d1a61e34_parseFileLoop(__state __ParserState, __file __ParsedFile) __FileStep {
	return func() __FileStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_EOF) {
			return __FileStep{__state: __state, __file: ____rune_private_d1a61e34_withFileErrors(__file, __state.__errors)}
		}
		return ____rune_private_d1a61e34_parseTopLevel(__state, __file)
	}()
}

func ____rune_private_d1a61e34_withFileErrors(__file __ParsedFile, __errors []__ParseError) __ParsedFile {
	return __ParsedFile{__imports: __file.__imports, __constants: __file.__constants, __types: __file.__types, __functions: __file.__functions, __tests: __file.__tests, __errors: __errors}
}

func ____rune_private_d1a61e34_parseTopLevel(__state __ParserState, __file __ParsedFile) __FileStep {
	return func() __FileStep {
		if ____rune_private_d1a61e34_looksLikeMacroFunctionDecl(__state) {
			return ____rune_private_d1a61e34_parseTopLevelAfterResult(____rune_private_d1a61e34_parseMacroFunction(__state, __file))
		}
		return ____rune_private_d1a61e34_parseTopLevelAfterMacro(__state, __file)
	}()
}

func ____rune_private_d1a61e34_parseTopLevelAfterResult(__result __FileStep) __FileStep {
	return ____rune_private_d1a61e34_parseFileLoop(____rune_private_d1a61e34_parserSkipNewlines(__result.__state), __result.__file)
}

func ____rune_private_d1a61e34_parseTopLevelAfterMacro(__state __ParserState, __file __ParsedFile) __FileStep {
	__goImport := ____rune_private_d1a61e34_looksLikeGoImportDecl(__state)
	return func() __FileStep {
		switch {
		case __goImport == true:
			return ____rune_private_d1a61e34_parseTopLevelAfterResult(____rune_private_d1a61e34_parseTopLevelImport(__state, __file))
		default:
			return func() __FileStep {
				if ____rune_private_d1a61e34_looksLikeRuneImportDecl(__state) {
					return ____rune_private_d1a61e34_parseTopLevelAfterResult(____rune_private_d1a61e34_parseTopLevelImport(__state, __file))
				}
				return ____rune_private_d1a61e34_parseTopLevelAfterAnnotations(__state, __file)
			}()
		}
	}()
}

func ____rune_private_d1a61e34_looksLikeRuneImportDecl(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_At) && (____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_String) || ____rune_private_d1a61e34_looksLikeBareModuleImportDecl(__state))
}

func ____rune_private_d1a61e34_parseTopLevelAfterAnnotations(__state __ParserState, __file __ParsedFile) __FileStep {
	__annotationStep := ____rune_private_d1a61e34_parseAnnotations(__state)
	__publicStep := ____rune_private_d1a61e34_parsePublicModifier(__annotationStep.__state)
	__current := __publicStep.__state
	__private := __publicStep.__ok == false
	__current = func() __ParserState {
		if __publicStep.__ok && (____rune_private_d1a61e34_parserCheck(__current, __TokenKind_At) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_Question)) {
			return ____rune_private_d1a61e34_parserErrorAt(__current, ____rune_private_d1a61e34_parserPeek(__current), "expected public declaration after '+'")
		}
		return __current
	}()
	__result := func() __FileStep {
		if !__publicStep.__ok && ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_At) {
			return ____rune_private_d1a61e34_parseTopLevelImport(__current, __file)
		}
		return func() __FileStep {
			if !__publicStep.__ok && ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_Question) {
				return ____rune_private_d1a61e34_parseTopLevelTest(__current, __file)
			}
			return func() __FileStep {
				if ____rune_private_d1a61e34_looksLikeConstDecl(__current) {
					return ____rune_private_d1a61e34_parseTopLevelConst(__current, __file, __private, __annotationStep.__annotations)
				}
				return func() __FileStep {
					if ____rune_private_d1a61e34_looksLikeTypeDecl(__current) {
						return ____rune_private_d1a61e34_parseTopLevelType(__current, __file, __private, __annotationStep.__annotations)
					}
					return func() __FileStep {
						if ____rune_private_d1a61e34_looksLikeFunctionDecl(__current) {
							return ____rune_private_d1a61e34_parseTopLevelFunction(__current, __file, __private, __annotationStep.__annotations)
						}
						return ____rune_private_d1a61e34_parseTopLevelError(__current, __file)
					}()
				}()
			}()
		}()
	}()
	return ____rune_private_d1a61e34_parseFileLoop(____rune_private_d1a61e34_parserSkipNewlines(__result.__state), __result.__file)
}

func ____rune_private_d1a61e34_looksLikeGoImportDecl(__state __ParserState) bool {
	__marker := ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_At)
	return func() bool {
		switch {
		case __marker == true:
			return ____rune_private_d1a61e34_looksLikeGoImportAfterMarker(__state)
		default:
			return false
		}
	}()
}

func ____rune_private_d1a61e34_looksLikeGoImportAfterMarker(__state __ParserState) bool {
	__module := ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+1) == __TokenKind_Ident && ____rune_private_d1a61e34_parserTokenAt(__state, __state.__current+1).__lexeme == "go"
	return func() bool {
		switch {
		case __module == true:
			return ____rune_private_d1a61e34_looksLikeGoImportAfterModule(__state)
		default:
			return false
		}
	}()
}

func ____rune_private_d1a61e34_looksLikeGoImportAfterModule(__state __ParserState) bool {
	__dot := ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+2) == __TokenKind_Dot
	return func() bool {
		switch {
		case __dot == true:
			return ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+3) == __TokenKind_Ident && ____rune_private_d1a61e34_parserTokenAt(__state, __state.__current+3).__lexeme == "import"
		default:
			return false
		}
	}()
}

func ____rune_private_d1a61e34_parseMacroFunction(__state __ParserState, __file __ParsedFile) __FileStep {
	__marker := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Hash, "expected '#' before macro function")
	__step := ____rune_private_d1a61e34_parseFunctionWithReceiver(__marker.__state, "", false, false, true, ____rune_private_d1a61e34_emptyAnnotations())
	__file.__functions = append(__file.__functions, __step.__function)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_d1a61e34_parseTopLevelImport(__state __ParserState, __file __ParsedFile) __FileStep {
	__step := func() __ImportStep {
		if ____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_String) {
			return ____rune_private_d1a61e34_parseImportDecl(__state)
		}
		return func() __ImportStep {
			if ____rune_private_d1a61e34_looksLikeBareModuleImportDecl(__state) {
				return ____rune_private_d1a61e34_parseModuleImportDecl(__state)
			}
			return ____rune_private_d1a61e34_parseGoImportDecl(__state)
		}()
	}()
	__file.__imports = append(__file.__imports, __step.__importDecl)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_d1a61e34_looksLikeBareModuleImportDecl(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_At) && ____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_Ident) && ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+2) != __TokenKind_Dot
}

func ____rune_private_d1a61e34_parseTopLevelTest(__state __ParserState, __file __ParsedFile) __FileStep {
	__step := ____rune_private_d1a61e34_parseTestDecl(__state)
	__file.__tests = append(__file.__tests, __step.__testDecl)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_d1a61e34_parseTopLevelConst(__state __ParserState, __file __ParsedFile, __private bool, __annotations []__ParsedAnnotation) __FileStep {
	__current := __state
	__current = ____rune_private_d1a61e34_parserRejectConstAnnotations(__current, __annotations)
	__step := ____rune_private_d1a61e34_parseConstDecl(__current, __private)
	__file.__constants = append(__file.__constants, __step.__constDecl)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_d1a61e34_parserRejectConstAnnotations(__state __ParserState, __annotations []__ParsedAnnotation) __ParserState {
	return func() __ParserState {
		switch {
		case len(__annotations) == 0:
			return __state
		default:
			return ____rune_private_d1a61e34_parserErrorAt(__state, ____rune_private_d1a61e34_parserPeek(__state), "annotations cannot be applied to constants")
		}
	}()
}

func ____rune_private_d1a61e34_looksLikeConstDecl(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) && ____rune_private_d1a61e34_parserPeek(__state).__lexeme == "const"
}

func ____rune_private_d1a61e34_parseConstDecl(__state __ParserState, __private bool) __ConstStep {
	__start := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected 'const'")
	__name := ____rune_private_d1a61e34_parserConsume(__start.__state, __TokenKind_Ident, "expected constant name")
	__typeStep := ____rune_private_d1a61e34_parseConstType(____rune_private_d1a61e34_parserSkipNewlines(__name.__state))
	__assign := ____rune_private_d1a61e34_parserConsume(____rune_private_d1a61e34_parserSkipNewlines(__typeStep.__state), __TokenKind_Assign, "expected '=' after constant name")
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__assign.__state), 1)
	return __ConstStep{__state: __value.__state, __constDecl: __ParsedConst{__name: __name.__token.__lexeme, __private: __private, __typeRef: __typeStep.__typeRef, __value: __value.__expr, __line: __start.__token.__line, __column: __start.__token.__column}}
}

func ____rune_private_d1a61e34_parseConstType(__state __ParserState) __TypeRefStep {
	__colon := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_Colon)
	return func() __TypeRefStep {
		switch {
		case __colon.__ok == true:
			return ____rune_private_d1a61e34_parseTypeRef(____rune_private_d1a61e34_parserSkipNewlines(__colon.__state))
		default:
			return __TypeRefStep{__state: __state, __typeRef: __emptyParsedTypeRef()}
		}
	}()
}

func ____rune_private_d1a61e34_parseTopLevelType(__state __ParserState, __file __ParsedFile, __private bool, __annotations []__ParsedAnnotation) __FileStep {
	__step := ____rune_private_d1a61e34_parseTypeDecl(__state, __private, __annotations)
	__file.__types = append(__file.__types, __step.__typeDecl)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_d1a61e34_parseTopLevelFunction(__state __ParserState, __file __ParsedFile, __private bool, __annotations []__ParsedAnnotation) __FileStep {
	__step := ____rune_private_d1a61e34_parseFunctionWithReceiver(__state, "", __private, false, false, __annotations)
	__file.__functions = append(__file.__functions, __step.__function)
	return __FileStep{__state: __step.__state, __file: __file}
}

func ____rune_private_d1a61e34_parseTopLevelError(__state __ParserState, __file __ParsedFile) __FileStep {
	return __FileStep{__state: ____rune_private_d1a61e34_parserAdvance(____rune_private_d1a61e34_parserErrorAt(__state, ____rune_private_d1a61e34_parserPeek(__state), "expected declaration")).__state, __file: __file}
}

func ____rune_private_d1a61e34_parsePublicModifier(__state __ParserState) __BoolStep {
	__step := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_Plus)
	return __BoolStep{__state: func() __ParserState {
		if __step.__ok {
			return ____rune_private_d1a61e34_parserSkipNewlines(__step.__state)
		}
		return __state
	}(), __ok: __step.__ok}
}

func ____rune_private_d1a61e34_parseObjectPrivateModifier(__state __ParserState) __BoolStep {
	__step := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_Minus)
	return __BoolStep{__state: func() __ParserState {
		if __step.__ok {
			return ____rune_private_d1a61e34_parserSkipNewlines(__step.__state)
		}
		return __state
	}(), __ok: __step.__ok}
}

func ____rune_private_d1a61e34_parseStaticMethodMarker(__state __ParserState) __BoolStep {
	__marker := func() __BoolStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_DoubleColon) {
			return ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_DoubleColon)
		}
		return func() __BoolStep {
			if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) && ____rune_private_d1a61e34_parserPeek(__state).__lexeme == "static" {
				return __BoolStep{__state: ____rune_private_d1a61e34_parserAdvance(__state).__state, __ok: true}
			}
			return __BoolStep{__state: __state, __ok: false}
		}()
	}()
	return __BoolStep{__state: func() __ParserState {
		if __marker.__ok {
			return ____rune_private_d1a61e34_parserSkipNewlines(__marker.__state)
		}
		return __state
	}(), __ok: __marker.__ok}
}

func ____rune_private_d1a61e34_parseImportDecl(__state __ParserState) __ImportStep {
	__at := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_At, "expected '@'")
	__path := ____rune_private_d1a61e34_parserConsume(__at.__state, __TokenKind_String, "expected import path string after '@'")
	__rawPath := ____rune_private_d1a61e34_unquote(__path.__token.__lexeme)
	__goPath := ____rune_private_d1a61e34_parserGoPackageImportPath(__rawPath)
	__isGo := __goPath != ""
	__invalidGo := strings.HasPrefix(__rawPath, "go:") && __goPath == ""
	__nextState := func() __ParserState {
		if __invalidGo {
			return ____rune_private_d1a61e34_parserErrorAt(__path.__state, __path.__token, "expected Go import path after \"go:\"")
		}
		return __path.__state
	}()
	return __ImportStep{__state: __nextState, __importDecl: __ParsedImport{__path: func() string {
		if __isGo {
			return __goPath
		}
		return __rawPath
	}(), __go: __isGo, __module: false, __line: __at.__token.__line, __column: __at.__token.__column}}
}

func ____rune_private_d1a61e34_parserGoPackageImportPath(__spec string) string {
	return func() string {
		if strings.HasPrefix(__spec, "go:") {
			return func() string { runes := []rune(__spec); return string(runes[3:len([]rune(__spec))]) }()
		}
		return ""
	}()
}

func ____rune_private_d1a61e34_parseModuleImportDecl(__state __ParserState) __ImportStep {
	__at := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_At, "expected '@'")
	__module := ____rune_private_d1a61e34_parserConsume(__at.__state, __TokenKind_Ident, "expected module name after '@'")
	return __ImportStep{__state: __module.__state, __importDecl: __ParsedImport{__path: __module.__token.__lexeme, __go: false, __module: true, __line: __at.__token.__line, __column: __at.__token.__column}}
}

func ____rune_private_d1a61e34_parseGoImportDecl(__state __ParserState) __ImportStep {
	__at := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_At, "expected '@'")
	__module := ____rune_private_d1a61e34_parserConsume(__at.__state, __TokenKind_Ident, "expected module name after '@'")
	__checked := func() __ParserState {
		if __module.__token.__lexeme == "go" {
			return __module.__state
		}
		return ____rune_private_d1a61e34_parserErrorAt(__module.__state, __module.__token, "only @go.import can appear at the top level")
	}()
	__dot := ____rune_private_d1a61e34_parserConsume(__checked, __TokenKind_Dot, "expected '.' after @go")
	__name := ____rune_private_d1a61e34_parserConsume(__dot.__state, __TokenKind_Ident, "expected import after @go.")
	__checkedName := func() __ParserState {
		if __name.__token.__lexeme == "import" {
			return __name.__state
		}
		return ____rune_private_d1a61e34_parserErrorAt(__name.__state, __name.__token, "only @go.import can appear at the top level")
	}()
	__open := ____rune_private_d1a61e34_parserConsume(__checkedName, __TokenKind_LParen, "expected '(' after @go.import")
	__path := ____rune_private_d1a61e34_parserConsume(__open.__state, __TokenKind_String, "expected Go import path string")
	__close := ____rune_private_d1a61e34_parserConsume(__path.__state, __TokenKind_RParen, "expected ')' after @go.import")
	return __ImportStep{__state: __close.__state, __importDecl: __ParsedImport{__path: ____rune_private_d1a61e34_unquote(__path.__token.__lexeme), __go: true, __module: false, __line: __at.__token.__line, __column: __at.__token.__column}}
}

func ____rune_private_d1a61e34_parseTestDecl(__state __ParserState) __TestStep {
	__start := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Question, "expected '?'")
	__name := ____rune_private_d1a61e34_parserConsume(__start.__state, __TokenKind_String, "expected test name string after '?'")
	__bodyStart := ____rune_private_d1a61e34_parserSkipNewlines(__name.__state)
	__body := func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__bodyStart, __TokenKind_LBrace) {
			return ____rune_private_d1a61e34_parseBlock(__bodyStart)
		}
		return __ExprStep{__state: ____rune_private_d1a61e34_parserErrorAt(__bodyStart, ____rune_private_d1a61e34_parserPeek(__bodyStart), "expected test body block"), __expr: ____rune_private_d1a61e34_emptyExpr()}
	}()
	return __TestStep{__state: __body.__state, __testDecl: __ParsedTest{__name: ____rune_private_d1a61e34_unquote(__name.__token.__lexeme), __body: __body.__expr, __line: __start.__token.__line, __column: __start.__token.__column}}
}

func ____rune_private_d1a61e34_parseTypeDecl(__state __ParserState, __private bool, __annotations []__ParsedAnnotation) __TypeStep {
	__name := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected type name")
	__generics := ____rune_private_d1a61e34_parseGenericNames(__name.__state)
	__colon := ____rune_private_d1a61e34_parserConsume(__generics.__state, __TokenKind_Colon, "expected ':' after type name")
	__openStart := ____rune_private_d1a61e34_parserSkipNewlines(__colon.__state)
	__open := ____rune_private_d1a61e34_parserConsume(__openStart, __TokenKind_LBrace, "expected '{' after type declaration")
	__bodyStart := ____rune_private_d1a61e34_parserSkipNewlines(__open.__state)
	return func() __TypeStep {
		if ____rune_private_d1a61e34_looksLikeEnumMember(__bodyStart) {
			return ____rune_private_d1a61e34_parseEnumTypeBody(__bodyStart, __name.__token, __private, __annotations, __generics.__values)
		}
		return ____rune_private_d1a61e34_parseStructTypeBody(__bodyStart, __name.__token, __private, __annotations, __generics.__values)
	}()
}

func ____rune_private_d1a61e34_parseStructTypeBody(__state __ParserState, __name __Token, __private bool, __annotations []__ParsedAnnotation, __generics []string) __TypeStep {
	return ____rune_private_d1a61e34_parseStructTypeLoop(__state, __ParsedType{__name: __name.__lexeme, __private: __private, __enum: false, __annotations: __annotations, __generics: __generics, __fields: []__ParsedField{}, __methods: []__ParsedFunction{}, __members: []__ParsedEnumMember{}, __line: __name.__line, __column: __name.__column})
}

func ____rune_private_d1a61e34_parseStructTypeLoop(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __TypeStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_d1a61e34_finishType(__current, __typeDecl, "expected '}' after type declaration")
		}
		return ____rune_private_d1a61e34_parseStructTypeMember(__current, __typeDecl)
	}()
}

func ____rune_private_d1a61e34_parseStructTypeMember(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	__macroMethod := ____rune_private_d1a61e34_looksLikeMacroFunctionDecl(__current)
	return func() __TypeStep {
		switch {
		case __macroMethod == true:
			return ____rune_private_d1a61e34_parseStructMacroMethod(__current, __typeDecl)
		default:
			return ____rune_private_d1a61e34_parseStructTypeMemberValue(__current, __typeDecl)
		}
	}()
}

func ____rune_private_d1a61e34_parseStructMacroMethod(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__marker := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Hash, "expected '#' before macro method")
	__step := ____rune_private_d1a61e34_parseFunctionWithReceiver(____rune_private_d1a61e34_parserSkipNewlines(__marker.__state), __typeDecl.__name, true, false, true, ____rune_private_d1a61e34_emptyAnnotations())
	__typeDecl.__methods = append(__typeDecl.__methods, __step.__function)
	__next := ____rune_private_d1a61e34_parserSkipNewlines(____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_consumeStatementEnd(__step.__state), __TokenKind_Comma).__state)
	return ____rune_private_d1a61e34_parseStructTypeLoop(__next, __typeDecl)
}

func ____rune_private_d1a61e34_parseStructTypeMemberValue(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__annotationStep := ____rune_private_d1a61e34_parseAnnotations(__state)
	__current := __annotationStep.__state
	__privateStep := ____rune_private_d1a61e34_parseObjectPrivateModifier(__current)
	__memberState := __privateStep.__state
	__private := __privateStep.__ok
	__staticStep := func() __BoolStep {
		if ____rune_private_d1a61e34_looksLikeStaticFunctionDecl(__memberState) {
			return ____rune_private_d1a61e34_parseStaticMethodMarker(__memberState)
		}
		return __BoolStep{__state: __memberState, __ok: false}
	}()
	__parsed := func() __TypeStep {
		if ____rune_private_d1a61e34_looksLikeFunctionDecl(__staticStep.__state) {
			return ____rune_private_d1a61e34_parseStructMethod(__staticStep.__state, __typeDecl, __private, __staticStep.__ok, __annotationStep.__annotations)
		}
		return ____rune_private_d1a61e34_parseStructField(__staticStep.__state, __typeDecl, __private, __annotationStep.__annotations)
	}()
	__next := ____rune_private_d1a61e34_parserSkipNewlines(____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_consumeStatementEnd(__parsed.__state), __TokenKind_Comma).__state)
	return ____rune_private_d1a61e34_parseStructTypeLoop(__next, __parsed.__typeDecl)
}

func ____rune_private_d1a61e34_parseStructMethod(__state __ParserState, __typeDecl __ParsedType, __private bool, __static bool, __annotations []__ParsedAnnotation) __TypeStep {
	__step := ____rune_private_d1a61e34_parseFunctionWithReceiver(__state, __typeDecl.__name, __private, __static, false, __annotations)
	__typeDecl.__methods = append(__typeDecl.__methods, __step.__function)
	return __TypeStep{__state: __step.__state, __typeDecl: __typeDecl}
}

func ____rune_private_d1a61e34_parseStructField(__state __ParserState, __typeDecl __ParsedType, __private bool, __annotations []__ParsedAnnotation) __TypeStep {
	__field := ____rune_private_d1a61e34_parseFieldDecl(__state, __private, __annotations)
	__typeDecl.__fields = append(__typeDecl.__fields, __field.__field)
	return __TypeStep{__state: __field.__state, __typeDecl: __typeDecl}
}

func ____rune_private_d1a61e34_parseFieldDecl(__state __ParserState, __private bool, __annotations []__ParsedAnnotation) __FieldStep {
	__name := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected field name")
	__colon := ____rune_private_d1a61e34_parserConsume(__name.__state, __TokenKind_Colon, "expected ':' after field name")
	__typeRef := ____rune_private_d1a61e34_parseTypeRef(__colon.__state)
	return __FieldStep{__state: __typeRef.__state, __field: __ParsedField{__name: __name.__token.__lexeme, __private: __private, __annotations: __annotations, __typeRef: __typeRef.__typeRef, __line: __name.__token.__line, __column: __name.__token.__column}}
}

func ____rune_private_d1a61e34_finishType(__state __ParserState, __typeDecl __ParsedType, __message string) __TypeStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RBrace, __message)
	return __TypeStep{__state: __close.__state, __typeDecl: __typeDecl}
}

func ____rune_private_d1a61e34_parseEnumTypeBody(__state __ParserState, __name __Token, __private bool, __annotations []__ParsedAnnotation, __generics []string) __TypeStep {
	return ____rune_private_d1a61e34_parseEnumTypeLoop(__state, __ParsedType{__name: __name.__lexeme, __private: __private, __enum: true, __annotations: __annotations, __generics: __generics, __fields: []__ParsedField{}, __methods: []__ParsedFunction{}, __members: []__ParsedEnumMember{}, __line: __name.__line, __column: __name.__column})
}

func ____rune_private_d1a61e34_parseEnumTypeLoop(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __TypeStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_d1a61e34_finishType(__current, __typeDecl, "expected '}' after enum declaration")
		}
		return ____rune_private_d1a61e34_parseEnumTypeMember(__current, __typeDecl)
	}()
}

func ____rune_private_d1a61e34_parseEnumTypeMember(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	__macroMethod := ____rune_private_d1a61e34_looksLikeMacroFunctionDecl(__current)
	return func() __TypeStep {
		switch {
		case __macroMethod == true:
			return ____rune_private_d1a61e34_parseEnumMacroMethod(__current, __typeDecl)
		default:
			return ____rune_private_d1a61e34_parseEnumTypeMemberValueOrMethod(__current, __typeDecl)
		}
	}()
}

func ____rune_private_d1a61e34_parseEnumMacroMethod(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__marker := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Hash, "expected '#' before macro method")
	__step := ____rune_private_d1a61e34_parseFunctionWithReceiver(____rune_private_d1a61e34_parserSkipNewlines(__marker.__state), __typeDecl.__name, true, false, true, ____rune_private_d1a61e34_emptyAnnotations())
	__typeDecl.__methods = append(__typeDecl.__methods, __step.__function)
	__next := ____rune_private_d1a61e34_parserSkipNewlines(____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_consumeStatementEnd(__step.__state), __TokenKind_Comma).__state)
	return ____rune_private_d1a61e34_parseEnumTypeLoop(__next, __typeDecl)
}

func ____rune_private_d1a61e34_parseEnumTypeMemberValueOrMethod(__state __ParserState, __typeDecl __ParsedType) __TypeStep {
	__annotationStep := ____rune_private_d1a61e34_parseAnnotations(__state)
	__current := __annotationStep.__state
	__privateStep := ____rune_private_d1a61e34_parseObjectPrivateModifier(__current)
	__memberState := __privateStep.__state
	__staticStep := func() __BoolStep {
		if ____rune_private_d1a61e34_looksLikeStaticFunctionDecl(__memberState) {
			return ____rune_private_d1a61e34_parseStaticMethodMarker(__memberState)
		}
		return __BoolStep{__state: __memberState, __ok: false}
	}()
	__parsed := func() __TypeStep {
		if ____rune_private_d1a61e34_looksLikeFunctionDecl(__staticStep.__state) {
			return ____rune_private_d1a61e34_parseEnumMethod(__staticStep.__state, __typeDecl, __privateStep.__ok, __staticStep.__ok, __annotationStep.__annotations)
		}
		return ____rune_private_d1a61e34_parseEnumTypeMemberValue(__current, __typeDecl, __annotationStep.__annotations)
	}()
	__next := ____rune_private_d1a61e34_parserSkipNewlines(____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_consumeStatementEnd(__parsed.__state), __TokenKind_Comma).__state)
	return ____rune_private_d1a61e34_parseEnumTypeLoop(__next, __parsed.__typeDecl)
}

func ____rune_private_d1a61e34_parseEnumMethod(__state __ParserState, __typeDecl __ParsedType, __private bool, __static bool, __annotations []__ParsedAnnotation) __TypeStep {
	__step := ____rune_private_d1a61e34_parseFunctionWithReceiver(__state, __typeDecl.__name, __private, __static, false, __annotations)
	__typeDecl.__methods = append(__typeDecl.__methods, __step.__function)
	return __TypeStep{__state: __step.__state, __typeDecl: __typeDecl}
}

func ____rune_private_d1a61e34_parseEnumTypeMemberValue(__state __ParserState, __typeDecl __ParsedType, __annotations []__ParsedAnnotation) __TypeStep {
	__member := ____rune_private_d1a61e34_parseEnumMember(__state, __annotations)
	__typeDecl.__members = append(__typeDecl.__members, __member.__member)
	return __TypeStep{__state: __member.__state, __typeDecl: __typeDecl}
}

func ____rune_private_d1a61e34_parseEnumMember(__state __ParserState, __annotations []__ParsedAnnotation) __EnumMemberStep {
	__publicStep := ____rune_private_d1a61e34_parsePublicModifier(__state)
	__name := ____rune_private_d1a61e34_parserConsume(__publicStep.__state, __TokenKind_Ident, "expected enum member name")
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__name.__state)
	__parsed := ____rune_private_d1a61e34_parseEnumMemberPayload(__current)
	return __EnumMemberStep{__state: __parsed.__state, __member: __ParsedEnumMember{__name: __name.__token.__lexeme, __private: false, __annotations: __annotations, __value: __parsed.__value, __params: __parsed.__params, __line: __name.__token.__line, __column: __name.__token.__column}}
}

func ____rune_private_d1a61e34_parseEnumMemberPayload(__state __ParserState) __EnumMemberPayloadStep {
	__assign := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_Assign)
	return func() __EnumMemberPayloadStep {
		if __assign.__ok {
			return ____rune_private_d1a61e34_parseEnumMemberValue(____rune_private_d1a61e34_parserSkipNewlines(__assign.__state))
		}
		return ____rune_private_d1a61e34_parseEnumMemberParams(__state)
	}()
}

func ____rune_private_d1a61e34_parseEnumMemberValue(__state __ParserState) __EnumMemberPayloadStep {
	__value := ____rune_private_d1a61e34_parseEnumValue(__state)
	return __EnumMemberPayloadStep{__state: __value.__state, __value: __value.__value, __params: []__ParsedParam{}}
}

func ____rune_private_d1a61e34_parseEnumMemberParams(__state __ParserState) __EnumMemberPayloadStep {
	__open := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_LParen)
	return func() __EnumMemberPayloadStep {
		if __open.__ok {
			return ____rune_private_d1a61e34_parseEnumMemberParamList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state))
		}
		return __EnumMemberPayloadStep{__state: __state, __value: "", __params: []__ParsedParam{}}
	}()
}

func ____rune_private_d1a61e34_parseEnumMemberParamList(__state __ParserState) __EnumMemberPayloadStep {
	__params := ____rune_private_d1a61e34_parseParamList(__state)
	__close := ____rune_private_d1a61e34_parserConsume(__params.__state, __TokenKind_RParen, "expected ')' after enum constructor parameters")
	return __EnumMemberPayloadStep{__state: __close.__state, __value: "", __params: __params.__params}
}

func ____rune_private_d1a61e34_parseEnumValue(__state __ParserState) __StringStep {
	return func() __StringStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Minus) {
			return ____rune_private_d1a61e34_parseNegativeEnumValue(__state)
		}
		return ____rune_private_d1a61e34_parsePositiveEnumValue(__state)
	}()
}

func ____rune_private_d1a61e34_parseNegativeEnumValue(__state __ParserState) __StringStep {
	__minus := ____rune_private_d1a61e34_parserAdvance(__state)
	__value := ____rune_private_d1a61e34_parserConsume(__minus.__state, __TokenKind_Int, "expected integer enum value")
	return __StringStep{__state: __value.__state, __value: "-" + __value.__token.__lexeme}
}

func ____rune_private_d1a61e34_parsePositiveEnumValue(__state __ParserState) __StringStep {
	__value := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Int, "expected integer enum value")
	return __StringStep{__state: __value.__state, __value: __value.__token.__lexeme}
}

func ____rune_private_d1a61e34_parseFunctionWithReceiver(__state __ParserState, __receiverType string, __private bool, __static bool, __macro bool, __annotations []__ParsedAnnotation) __FunctionStep {
	__routineStep := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_Tilde)
	__afterRoutine := func() __ParserState {
		if __routineStep.__ok {
			return ____rune_private_d1a61e34_parserSkipNewlines(__routineStep.__state)
		}
		return __state
	}()
	__name := ____rune_private_d1a61e34_parserConsume(__afterRoutine, __TokenKind_Ident, "expected function name")
	__generics := ____rune_private_d1a61e34_parseGenericNames(__name.__state)
	__open := ____rune_private_d1a61e34_parserConsume(__generics.__state, __TokenKind_LParen, "expected '(' after function name")
	__params := ____rune_private_d1a61e34_parseParamList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state))
	__close := ____rune_private_d1a61e34_parserConsume(__params.__state, __TokenKind_RParen, "expected ')' after parameter list")
	__afterClose := ____rune_private_d1a61e34_parserSkipNewlines(__close.__state)
	__ret := ____rune_private_d1a61e34_parserMatch(__afterClose, __TokenKind_Arrow)
	__returnType := func() __TypeRefStep {
		if __ret.__ok {
			return ____rune_private_d1a61e34_parseTypeRef(____rune_private_d1a61e34_parserSkipNewlines(__ret.__state))
		}
		return __TypeRefStep{__state: __afterClose, __typeRef: __emptyParsedTypeRef()}
	}()
	__arrow := ____rune_private_d1a61e34_parserConsume(____rune_private_d1a61e34_parserSkipNewlines(__returnType.__state), __TokenKind_FatArrow, "expected '=>' after function signature")
	__body := ____rune_private_d1a61e34_parseBody(____rune_private_d1a61e34_parserSkipNewlines(__arrow.__state))
	return __FunctionStep{__state: __body.__state, __function: __ParsedFunction{__name: __name.__token.__lexeme, __private: __private, __static: __static, __routine: __routineStep.__ok, __macro: __macro, __annotations: __annotations, __receiverType: __receiverType, __generics: __generics.__values, __params: __params.__params, __returnType: __returnType.__typeRef, __body: __body.__expr, __line: __name.__token.__line, __column: __name.__token.__column}}
}

func ____rune_private_d1a61e34_parseParamList(__state __ParserState) __ParamListStep {
	return ____rune_private_d1a61e34_parseParamListLoop(__state, []__ParsedParam{})
}

func ____rune_private_d1a61e34_parseParamListLoop(__state __ParserState, __params []__ParsedParam) __ParamListStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ParamListStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RParen) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return __ParamListStep{__state: __current, __params: __params}
		}
		return ____rune_private_d1a61e34_parseOneParam(__current, __params)
	}()
}

func ____rune_private_d1a61e34_parseOneParam(__state __ParserState, __params []__ParsedParam) __ParamListStep {
	__name := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected parameter name")
	__optional := ____rune_private_d1a61e34_parserMatch(__name.__state, __TokenKind_Question)
	__colon := ____rune_private_d1a61e34_parserMatch(__optional.__state, __TokenKind_Colon)
	__typeRef := func() __TypeRefStep {
		if __colon.__ok {
			return ____rune_private_d1a61e34_parseTypeRef(__colon.__state)
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
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__typeRef.__state), __TokenKind_Comma)
	return ____rune_private_d1a61e34_parseParamListLoop(__comma.__state, __params)
}

func ____rune_private_d1a61e34_parseGenericNames(__state __ParserState) __StringListStep {
	__open := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_LBracket)
	return func() __StringListStep {
		if __open.__ok {
			return ____rune_private_d1a61e34_parseGenericNameLoop(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), []string{})
		}
		return __StringListStep{__state: __state, __values: []string{}}
	}()
}

func ____rune_private_d1a61e34_parseGenericNameLoop(__state __ParserState, __values []string) __StringListStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __StringListStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBracket) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_d1a61e34_finishGenericNames(__current, __values)
		}
		return ____rune_private_d1a61e34_parseGenericNameValue(__current, __values)
	}()
}

func ____rune_private_d1a61e34_finishGenericNames(__state __ParserState, __values []string) __StringListStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RBracket, "expected ']' after generic parameters")
	return __StringListStep{__state: __close.__state, __values: __values}
}

func ____rune_private_d1a61e34_parseGenericNameValue(__state __ParserState, __values []string) __StringListStep {
	__nameStep := func() __StringListStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) {
			return ____rune_private_d1a61e34_appendGenericName(__state, __values)
		}
		return __StringListStep{__state: ____rune_private_d1a61e34_parserAdvance(__state).__state, __values: __values}
	}()
	__colon := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__nameStep.__state), __TokenKind_Colon)
	__current := func() __ParserState {
		if __colon.__ok {
			return ____rune_private_d1a61e34_parseTypeRef(____rune_private_d1a61e34_parserSkipNewlines(__colon.__state)).__state
		}
		return __nameStep.__state
	}()
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__current), __TokenKind_Comma)
	return ____rune_private_d1a61e34_parseGenericNameLoop(__comma.__state, __nameStep.__values)
}

func ____rune_private_d1a61e34_appendGenericName(__state __ParserState, __values []string) __StringListStep {
	__step := ____rune_private_d1a61e34_parserAdvance(__state)
	return __StringListStep{__state: __step.__state, __values: ____rune_private_d1a61e34_appendString(__values, __step.__token.__lexeme)}
}

func ____rune_private_d1a61e34_parseTypeRef(__state __ParserState) __TypeRefStep {
	return ____rune_private_d1a61e34_parseTypeRefPostfix(____rune_private_d1a61e34_parseTypeRefAtom(____rune_private_d1a61e34_parserSkipNewlines(__state)))
}

func ____rune_private_d1a61e34_parseTypeRefAtom(__state __ParserState) __TypeRefStep {
	return func() __TypeRefStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LParen) {
			return ____rune_private_d1a61e34_parseParenTypeRef(__state)
		}
		return func() __TypeRefStep {
			if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_At) {
				return ____rune_private_d1a61e34_parseQualifiedTypeRef(__state)
			}
			return func() __TypeRefStep {
				if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) {
					return ____rune_private_d1a61e34_parseNamedTypeRef(__state)
				}
				return ____rune_private_d1a61e34_parseTypeRefError(__state)
			}()
		}()
	}()
}

func ____rune_private_d1a61e34_parseNamedTypeRef(__state __ParserState) __TypeRefStep {
	__name := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected type name")
	return __TypeRefStep{__state: __name.__state, __typeRef: __namedParsedTypeRef(__name.__token)}
}

func ____rune_private_d1a61e34_parseQualifiedTypeRef(__state __ParserState) __TypeRefStep {
	__at := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_At, "expected '@'")
	__module := ____rune_private_d1a61e34_parserConsume(__at.__state, __TokenKind_Ident, "expected module name after '@'")
	__dot := ____rune_private_d1a61e34_parserConsume(__module.__state, __TokenKind_Dot, "expected '.' after module name")
	__name := ____rune_private_d1a61e34_parserConsume(__dot.__state, __TokenKind_Ident, "expected type name after module qualifier")
	return __TypeRefStep{__state: __name.__state, __typeRef: __qualifiedParsedTypeRef(__module.__token, __name.__token)}
}

func ____rune_private_d1a61e34_parseParenTypeRef(__state __ParserState) __TypeRefStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LParen, "expected '(' before type")
	__params := ____rune_private_d1a61e34_parseTypeParamList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), []__ParsedTypeParam{})
	__close := ____rune_private_d1a61e34_parserConsume(__params.__state, __TokenKind_RParen, "expected ')' after type")
	__afterClose := ____rune_private_d1a61e34_parserSkipNewlines(__close.__state)
	__arrow := ____rune_private_d1a61e34_parserMatch(__afterClose, __TokenKind_Arrow)
	return func() __TypeRefStep {
		if __arrow.__ok {
			return ____rune_private_d1a61e34_finishFunctionTypeRef(____rune_private_d1a61e34_parserSkipNewlines(__arrow.__state), __open.__token, __params.__params)
		}
		return ____rune_private_d1a61e34_finishParenTypeRef(__close.__state, __open.__token, __params.__params)
	}()
}

func ____rune_private_d1a61e34_finishFunctionTypeRef(__state __ParserState, __token __Token, __params []__ParsedTypeParam) __TypeRefStep {
	__ret := ____rune_private_d1a61e34_parseTypeRef(__state)
	return __TypeRefStep{__state: __ret.__state, __typeRef: __functionTypeRef(__params, __ret.__typeRef, __token)}
}

func ____rune_private_d1a61e34_finishParenTypeRef(__state __ParserState, __token __Token, __params []__ParsedTypeParam) __TypeRefStep {
	return __TypeRefStep{__state: __state, __typeRef: func() __ParsedTypeRef {
		if len(__params) == 1 && __params[0].__name == "" {
			return __groupedTypeRef(__params[0].__typeRef, __token)
		}
		return __tupleTypeRef(__params, __token)
	}()}
}

func ____rune_private_d1a61e34_parseTypeRefPostfix(__step __TypeRefStep) __TypeRefStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__step.__state)
	return func() __TypeRefStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_LBracket) {
			return ____rune_private_d1a61e34_parseTypeRefPostfix(____rune_private_d1a61e34_parseTypeRefArgs(__current, __step.__typeRef))
		}
		return func() __TypeRefStep {
			if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_Question) {
				return ____rune_private_d1a61e34_parseTypeRefPostfix(____rune_private_d1a61e34_parseNullableTypeRef(__current, __step.__typeRef))
			}
			return __TypeRefStep{__state: __step.__state, __typeRef: __step.__typeRef}
		}()
	}()
}

func ____rune_private_d1a61e34_parseTypeRefArgs(__state __ParserState, __typeRef __ParsedTypeRef) __TypeRefStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBracket, "expected '[' after type name")
	__args := ____rune_private_d1a61e34_parseTypeRefList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), []__ParsedTypeRef{})
	__close := ____rune_private_d1a61e34_parserConsume(__args.__state, __TokenKind_RBracket, "expected ']' after type arguments")
	return __TypeRefStep{__state: __close.__state, __typeRef: __typeRefWithArgs(__typeRef, __args.__refs)}
}

func ____rune_private_d1a61e34_parseNullableTypeRef(__state __ParserState, __typeRef __ParsedTypeRef) __TypeRefStep {
	__question := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Question, "expected '?' after type")
	return __TypeRefStep{__state: __question.__state, __typeRef: __nullableTypeRef(__typeRef)}
}

func ____rune_private_d1a61e34_parseTypeRefList(__state __ParserState, __refs []__ParsedTypeRef) __TypeRefListStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __TypeRefListStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBracket) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return __TypeRefListStep{__state: __current, __refs: __refs}
		}
		return ____rune_private_d1a61e34_parseOneTypeRefListValue(__current, __refs)
	}()
}

func ____rune_private_d1a61e34_parseOneTypeRefListValue(__state __ParserState, __refs []__ParsedTypeRef) __TypeRefListStep {
	__typeRef := ____rune_private_d1a61e34_parseTypeRef(__state)
	__refs = append(__refs, __typeRef.__typeRef)
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__typeRef.__state), __TokenKind_Comma)
	return func() __TypeRefListStep {
		switch {
		case __comma.__ok == true:
			return ____rune_private_d1a61e34_parseTypeRefList(__comma.__state, __refs)
		default:
			return __TypeRefListStep{__state: __typeRef.__state, __refs: __refs}
		}
	}()
}

func ____rune_private_d1a61e34_parseTypeParamList(__state __ParserState, __params []__ParsedTypeParam) __TypeParamListStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __TypeParamListStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RParen) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return __TypeParamListStep{__state: __current, __params: __params}
		}
		return ____rune_private_d1a61e34_parseOneTypeParam(__current, __params)
	}()
}

func ____rune_private_d1a61e34_parseOneTypeParam(__state __ParserState, __params []__ParsedTypeParam) __TypeParamListStep {
	__param := ____rune_private_d1a61e34_parseTypeParam(__state)
	__params = append(__params, __param.__param)
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__param.__state), __TokenKind_Comma)
	return func() __TypeParamListStep {
		switch {
		case __comma.__ok == true:
			return ____rune_private_d1a61e34_parseTypeParamList(__comma.__state, __params)
		default:
			return __TypeParamListStep{__state: __param.__state, __params: __params}
		}
	}()
}

func ____rune_private_d1a61e34_parseTypeParam(__state __ParserState) __TypeParamStep {
	__named := ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) && ____rune_private_d1a61e34_typeParamHasName(__state)
	return func() __TypeParamStep {
		if __named {
			return ____rune_private_d1a61e34_parseNamedTypeParam(__state)
		}
		return ____rune_private_d1a61e34_parseUnnamedTypeParam(__state)
	}()
}

func ____rune_private_d1a61e34_parseNamedTypeParam(__state __ParserState) __TypeParamStep {
	__name := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected type parameter name")
	__optional := ____rune_private_d1a61e34_parserMatch(__name.__state, __TokenKind_Question)
	__colon := ____rune_private_d1a61e34_parserConsume(__optional.__state, __TokenKind_Colon, "expected ':' after type parameter name")
	__typeRef := ____rune_private_d1a61e34_parseTypeRef(____rune_private_d1a61e34_parserSkipNewlines(__colon.__state))
	return __TypeParamStep{__state: __typeRef.__state, __param: __ParsedTypeParam{__name: __name.__token.__lexeme, __optional: __optional.__ok, __typeRef: __typeRef.__typeRef}}
}

func ____rune_private_d1a61e34_parseUnnamedTypeParam(__state __ParserState) __TypeParamStep {
	__typeRef := ____rune_private_d1a61e34_parseTypeRef(__state)
	return __TypeParamStep{__state: __typeRef.__state, __param: __ParsedTypeParam{__name: "", __optional: false, __typeRef: __typeRef.__typeRef}}
}

func ____rune_private_d1a61e34_typeParamHasName(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+1) == __TokenKind_Colon || ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+1) == __TokenKind_Question && ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+2) == __TokenKind_Colon
}

func ____rune_private_d1a61e34_parseTypeRefError(__state __ParserState) __TypeRefStep {
	return __TypeRefStep{__state: ____rune_private_d1a61e34_parserErrorAt(__state, ____rune_private_d1a61e34_parserPeek(__state), "expected type name"), __typeRef: __emptyParsedTypeRef()}
}

func ____rune_private_d1a61e34_parseBody(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LBrace) {
			return ____rune_private_d1a61e34_parseBraceBody(__state)
		}
		return ____rune_private_d1a61e34_parseExpression(__state, 1)
	}()
}

func ____rune_private_d1a61e34_parseBraceBody(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_looksLikePatternBranch(__state) == false && ____rune_private_d1a61e34_looksLikeMapLiteralBody(__state) {
			return ____rune_private_d1a61e34_parseMapLiteral(__state)
		}
		return func() __ExprStep {
			if ____rune_private_d1a61e34_looksLikePatternBranch(__state) == false && ____rune_private_d1a61e34_looksLikeObjectLiteralBody(__state) {
				return ____rune_private_d1a61e34_parseObjectLiteral(__state)
			}
			return ____rune_private_d1a61e34_parseBlock(__state)
		}()
	}()
}

func ____rune_private_d1a61e34_looksLikeObjectLiteralBody(__state __ParserState) bool {
	__first := ____rune_private_d1a61e34_skipNewlinesAt(__state, __state.__current+1)
	return ____rune_private_d1a61e34_parserKindAt(__state, __first) == __TokenKind_Ident && ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __first+1)) == __TokenKind_Colon
}

func ____rune_private_d1a61e34_parseBlock(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBrace, "expected '{'")
	__bodyStart := ____rune_private_d1a61e34_parserSkipNewlines(__open.__state)
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__bodyStart, __TokenKind_RBrace) {
			return ____rune_private_d1a61e34_finishBlock(__bodyStart, ____rune_private_d1a61e34_node(__ExprKind_Block, __open.__token))
		}
		return func() __ExprStep {
			if ____rune_private_d1a61e34_looksLikePatternBranch(__bodyStart) {
				return ____rune_private_d1a61e34_parsePatternBlock(__bodyStart, ____rune_private_d1a61e34_node(__ExprKind_PatternBlock, __open.__token))
			}
			return ____rune_private_d1a61e34_parseBlockLoop(__bodyStart, ____rune_private_d1a61e34_node(__ExprKind_Block, __open.__token))
		}()
	}()
}

func ____rune_private_d1a61e34_finishBlock(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RBrace, "expected '}' after block")
	return __ExprStep{__state: __close.__state, __expr: __block}
}

func ____rune_private_d1a61e34_parseBlockLoop(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_d1a61e34_finishBlock(__current, __block)
		}
		return ____rune_private_d1a61e34_parseBlockStatement(__current, __block)
	}()
}

func ____rune_private_d1a61e34_parseBlockStatement(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__stmt := ____rune_private_d1a61e34_parseStatement(__state)
	__nextBlock := ____rune_private_d1a61e34_appendChild(__block, __stmt.__expr)
	return ____rune_private_d1a61e34_parseBlockLoop(____rune_private_d1a61e34_consumeStatementEnd(__stmt.__state), __nextBlock)
}

func ____rune_private_d1a61e34_parseStatement(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LBrace) && ____rune_private_d1a61e34_looksLikeObjectDestructureDecl(__state) {
			return ____rune_private_d1a61e34_parseObjectDestructureStatement(__state)
		}
		return func() __ExprStep {
			if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LBrace) {
				return ____rune_private_d1a61e34_parseBlock(__state)
			}
			return func() __ExprStep {
				if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Dollar) && ____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_Ident) {
					return ____rune_private_d1a61e34_parseDollarStatement(__state)
				}
				return func() __ExprStep {
					if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) && (____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_Declare) || ____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_MutDeclare)) {
						return ____rune_private_d1a61e34_parseLetStatement(__state)
					}
					return func() __ExprStep {
						if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) && ____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_Assign) {
							return ____rune_private_d1a61e34_parseAssignStatement(__state)
						}
						return ____rune_private_d1a61e34_parseExpression(__state, 1)
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_d1a61e34_parseDollarStatement(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+2) == __TokenKind_Declare {
			return ____rune_private_d1a61e34_parseSignalPrefixLetStatement(__state)
		}
		return ____rune_private_d1a61e34_parseExpression(__state, 1)
	}()
}

func ____rune_private_d1a61e34_parseSignalPrefixLetStatement(__state __ParserState) __ExprStep {
	__dollar := ____rune_private_d1a61e34_parserAdvance(__state)
	__name := ____rune_private_d1a61e34_parserAdvance(__dollar.__state)
	__op := ____rune_private_d1a61e34_parserAdvance(__name.__state)
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__op.__state), 1)
	__typeName := ____rune_private_d1a61e34_parseLetTypeAnnotation(__value.__state)
	return __ExprStep{__state: __typeName.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Let, "$"+__name.__token.__lexeme, __name.__token.__lexeme, __typeName.__value, __op.__token.__lexeme, []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)}
}

func ____rune_private_d1a61e34_parseLetStatement(__state __ParserState) __ExprStep {
	__name := ____rune_private_d1a61e34_parserAdvance(__state)
	__op := ____rune_private_d1a61e34_parserAdvance(__name.__state)
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__op.__state), 1)
	__typeName := ____rune_private_d1a61e34_parseLetTypeAnnotation(__value.__state)
	return __ExprStep{__state: __typeName.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Let, __name.__token.__lexeme, __name.__token.__lexeme, __typeName.__value, __op.__token.__lexeme, []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)}
}

func ____rune_private_d1a61e34_parseLetTypeAnnotation(__state __ParserState) __StringStep {
	__colon := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__state), __TokenKind_Colon)
	return func() __StringStep {
		switch {
		case __colon.__ok == true:
			return ____rune_private_d1a61e34_parseLetTypeAnnotationRef(__colon.__state)
		default:
			return __StringStep{__state: __state, __value: ""}
		}
	}()
}

func ____rune_private_d1a61e34_parseLetTypeAnnotationRef(__state __ParserState) __StringStep {
	__typeRef := ____rune_private_d1a61e34_parseTypeRef(____rune_private_d1a61e34_parserSkipNewlines(__state))
	return __StringStep{__state: __typeRef.__state, __value: __typeRefToString(__typeRef.__typeRef)}
}

func ____rune_private_d1a61e34_parseObjectDestructureStatement(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBrace, "expected '{' before object destructuring")
	__fields := ____rune_private_d1a61e34_parseObjectBindingList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state))
	__close := ____rune_private_d1a61e34_parserConsume(__fields.__state, __TokenKind_RBrace, "expected '}' after object destructuring")
	__op := ____rune_private_d1a61e34_parserAdvance(____rune_private_d1a61e34_parserSkipNewlines(__close.__state))
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__op.__state), 1)
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_ObjectDestructure, "", "", "", __op.__token.__lexeme, __fields.__params, []__ParsedExpr{__value.__expr}, __open.__token.__line, __open.__token.__column)}
}

func ____rune_private_d1a61e34_parseObjectBindingList(__state __ParserState) __ParamListStep {
	return ____rune_private_d1a61e34_parseObjectBindingListLoop(__state, []__ParsedParam{})
}

func ____rune_private_d1a61e34_parseObjectBindingListLoop(__state __ParserState, __fields []__ParsedParam) __ParamListStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ParamListStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return __ParamListStep{__state: __current, __params: __fields}
		}
		return ____rune_private_d1a61e34_parseOneObjectBinding(__current, __fields)
	}()
}

func ____rune_private_d1a61e34_parseOneObjectBinding(__state __ParserState, __fields []__ParsedParam) __ParamListStep {
	__field := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected field name in object destructuring")
	__aliasStart := ____rune_private_d1a61e34_parserMatch(__field.__state, __TokenKind_Colon)
	__name := func() __TokenStep {
		if __aliasStart.__ok {
			return ____rune_private_d1a61e34_parserConsume(____rune_private_d1a61e34_parserSkipNewlines(__aliasStart.__state), __TokenKind_Ident, "expected binding name after ':'")
		}
		return __field
	}()
	__fields = append(__fields, __ParsedParam{__name: __name.__token.__lexeme, __typeRef: __namedParsedTypeRef(__field.__token), __line: __name.__token.__line, __column: __name.__token.__column})
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__name.__state), __TokenKind_Comma)
	return func() __ParamListStep {
		if __comma.__ok {
			return ____rune_private_d1a61e34_parseObjectBindingListLoop(__comma.__state, __fields)
		}
		return __ParamListStep{__state: __name.__state, __params: __fields}
	}()
}

func ____rune_private_d1a61e34_parseAssignStatement(__state __ParserState) __ExprStep {
	__name := ____rune_private_d1a61e34_parserAdvance(__state)
	__op := ____rune_private_d1a61e34_parserAdvance(__name.__state)
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__op.__state), 1)
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Assign, __name.__token.__lexeme, __name.__token.__lexeme, "", __op.__token.__lexeme, []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)}
}

func ____rune_private_d1a61e34_parseExpression(__state __ParserState, __minPrec int) __ExprStep {
	return func() __ExprStep {
		if __minPrec <= 1 && ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LParen) && ____rune_private_d1a61e34_looksLikeLambda(__state) {
			return ____rune_private_d1a61e34_parseLambda(__state)
		}
		return ____rune_private_d1a61e34_parseExpressionLoop(____rune_private_d1a61e34_parseUnary(__state), __minPrec)
	}()
}

func ____rune_private_d1a61e34_parseExpressionLoop(__left __ExprStep, __minPrec int) __ExprStep {
	__state := __left.__state
	__expr := __left.__expr
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LBrace) {
			return ____rune_private_d1a61e34_parseAfterBraceExpression(__state, __expr, __minPrec)
		}
		return func() __ExprStep {
			if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LParen) {
				return ____rune_private_d1a61e34_parseCallExpression(__state, __expr, __minPrec)
			}
			return func() __ExprStep {
				if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LBracket) {
					return ____rune_private_d1a61e34_parseIndexExpression(__state, __expr, __minPrec)
				}
				return func() __ExprStep {
					if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Dot) || ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_DoubleColon) {
						return ____rune_private_d1a61e34_parseSelectorExpression(__state, __expr, __minPrec)
					}
					return func() __ExprStep {
						if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_PlusPlus) {
							return ____rune_private_d1a61e34_parsePostfixExpression(__state, __expr, __minPrec)
						}
						return func() __ExprStep {
							if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Apostrophe) {
								return ____rune_private_d1a61e34_parseCompileTimePostfixExpression(__state, __expr, __minPrec)
							}
							return func() __ExprStep {
								if __minPrec <= 1 && ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Arrow) {
									return ____rune_private_d1a61e34_parseWatchExpression(__state, __expr, __minPrec)
								}
								return func() __ExprStep {
									if __minPrec <= 1 && ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Assign) {
										return ____rune_private_d1a61e34_parseAssignmentExpression(__state, __expr, __minPrec)
									}
									return func() __ExprStep {
										if __minPrec <= 1 && ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Tilde) {
											return ____rune_private_d1a61e34_parsePatternPredicateExpression(__state, __expr, __minPrec)
										}
										return func() __ExprStep {
											if __minPrec <= 1 && ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Question) {
												return ____rune_private_d1a61e34_parseQuestionExpression(__state, __expr, __minPrec)
											}
											return func() __ExprStep {
												if __minPrec <= 1 && ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_QuestionQuestion) {
													return ____rune_private_d1a61e34_parseQuestionQuestionExpression(__state, __expr, __minPrec)
												}
												return ____rune_private_d1a61e34_parseBinaryExpression(__state, __expr, __minPrec)
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
	}()
}

func ____rune_private_d1a61e34_parseAfterBraceExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_looksLikePatternBlockAfterSubject(__state) {
			return ____rune_private_d1a61e34_parseExpressionLoop(____rune_private_d1a61e34_parseMatchExpression(__state, __expr), __minPrec)
		}
		return func() __ExprStep {
			if __expr.__kind == __ExprKind_Identifier {
				return ____rune_private_d1a61e34_parseExpressionLoop(____rune_private_d1a61e34_parseStructLiteral(__state, __expr), __minPrec)
			}
			return __ExprStep{__state: __state, __expr: __expr}
		}()
	}()
}

func ____rune_private_d1a61e34_parseCallExpression(__state __ParserState, __callee __ParsedExpr, __minPrec int) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LParen, "expected '(' after callee")
	__args := ____rune_private_d1a61e34_parseArgumentList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), []__ParsedExpr{__callee}, __TokenKind_RParen)
	__close := ____rune_private_d1a61e34_parserConsume(__args.__state, __TokenKind_RParen, "expected ')' after arguments")
	__call := ____rune_private_d1a61e34_makeExpr(__ExprKind_Call, ____rune_private_d1a61e34_calleeText(__callee), "", "", "", []__ParsedParam{}, __args.__expr.__children, __callee.__line, __callee.__column)
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __close.__state, __expr: __call}, __minPrec)
}

func ____rune_private_d1a61e34_parseArgumentList(__state __ParserState, __holderChildren []__ParsedExpr, __endKind __TokenKind) __ExprStep {
	__holder := ____rune_private_d1a61e34_makeExpr(__ExprKind_Args, "", "", "", "", []__ParsedParam{}, __holderChildren, 0, 0)
	return ____rune_private_d1a61e34_parseArgumentListLoop(__state, __holder, __endKind)
}

func ____rune_private_d1a61e34_parseArgumentListLoop(__state __ParserState, __holder __ParsedExpr, __endKind __TokenKind) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __endKind) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return __ExprStep{__state: __current, __expr: __holder}
		}
		return ____rune_private_d1a61e34_parseOneArgument(__current, __holder, __endKind)
	}()
}

func ____rune_private_d1a61e34_parseOneArgument(__state __ParserState, __holder __ParsedExpr, __endKind __TokenKind) __ExprStep {
	__spread := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_DotDot)
	__value := ____rune_private_d1a61e34_parseExpression(func() __ParserState {
		if __spread.__ok {
			return __spread.__state
		}
		return __state
	}(), 1)
	__arg := func() __ParsedExpr {
		if __spread.__ok {
			return ____rune_private_d1a61e34_opNode(__ExprKind_Spread, "..", ____rune_private_d1a61e34_parserPrevious(__spread.__state), []__ParsedExpr{__value.__expr})
		}
		return __value.__expr
	}()
	__nextHolder := ____rune_private_d1a61e34_appendChild(__holder, __arg)
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__value.__state), __TokenKind_Comma)
	return ____rune_private_d1a61e34_parseArgumentListLoop(__comma.__state, __nextHolder, __endKind)
}

func ____rune_private_d1a61e34_parseIndexExpression(__state __ParserState, __receiver __ParsedExpr, __minPrec int) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBracket, "expected '[' after receiver")
	__index := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), 1)
	__close := ____rune_private_d1a61e34_parserConsume(__index.__state, __TokenKind_RBracket, "expected ']' after index")
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __close.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Index, __receiver.__text, "", "", "", []__ParsedParam{}, []__ParsedExpr{__receiver, __index.__expr}, __receiver.__line, __receiver.__column)}, __minPrec)
}

func ____rune_private_d1a61e34_parseSelectorExpression(__state __ParserState, __receiver __ParsedExpr, __minPrec int) __ExprStep {
	__operator := func() __TokenStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_DoubleColon) {
			return ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_DoubleColon, "expected '::'")
		}
		return ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Dot, "expected '.'")
	}()
	__name := ____rune_private_d1a61e34_parserConsume(__operator.__state, __TokenKind_Ident, "expected selector name")
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __name.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Selector, __name.__token.__lexeme, __name.__token.__lexeme, "", __operator.__token.__lexeme, []__ParsedParam{}, []__ParsedExpr{__receiver}, __operator.__token.__line, __operator.__token.__column)}, __minPrec)
}

func ____rune_private_d1a61e34_parsePostfixExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	__op := ____rune_private_d1a61e34_parserAdvance(__state)
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __op.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Postfix, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__expr})}, __minPrec)
}

func ____rune_private_d1a61e34_parseCompileTimePostfixExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	__op := ____rune_private_d1a61e34_parserAdvance(__state)
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __op.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_CompileTime, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__expr})}, __minPrec)
}

func ____rune_private_d1a61e34_parseWatchExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	__arrow := ____rune_private_d1a61e34_parserAdvance(__state)
	__handler := ____rune_private_d1a61e34_parseWatchHandler(____rune_private_d1a61e34_parserSkipNewlines(__arrow.__state))
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __handler.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Watch, __arrow.__token.__lexeme, __arrow.__token, []__ParsedExpr{__expr, __handler.__expr})}, __minPrec)
}

func ____rune_private_d1a61e34_parseWatchHandler(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LParen) && ____rune_private_d1a61e34_looksLikeLambda(__state) {
			return ____rune_private_d1a61e34_parseLambda(__state)
		}
		return func() __ExprStep {
			if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LBrace) {
				return ____rune_private_d1a61e34_parseBody(__state)
			}
			return ____rune_private_d1a61e34_parseExpression(__state, 1)
		}()
	}()
}

func ____rune_private_d1a61e34_parseAssignmentExpression(__state __ParserState, __target __ParsedExpr, __minPrec int) __ExprStep {
	__op := ____rune_private_d1a61e34_parserAdvance(__state)
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__op.__state), 1)
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __value.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Assign, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__target, __value.__expr})}, __minPrec)
}

func ____rune_private_d1a61e34_parsePatternPredicateExpression(__state __ParserState, __subject __ParsedExpr, __minPrec int) __ExprStep {
	__op := ____rune_private_d1a61e34_parserAdvance(__state)
	__pattern := ____rune_private_d1a61e34_parsePredicatePatternText(____rune_private_d1a61e34_parserSkipNewlines(__op.__state))
	__trueExpr := ____rune_private_d1a61e34_valueNode(__ExprKind_Bool, "true", __op.__token)
	__falseExpr := ____rune_private_d1a61e34_valueNode(__ExprKind_Bool, "false", __op.__token)
	__patternBranch := ____rune_private_d1a61e34_makeExpr(__ExprKind_Branch, __pattern.__expr.__text, "", "", "=>", []__ParsedParam{}, []__ParsedExpr{__pattern.__expr, __trueExpr}, __pattern.__expr.__line, __pattern.__expr.__column)
	__wildcardPattern := ____rune_private_d1a61e34_makeExpr(__ExprKind_Pattern, "_", "", "", "", []__ParsedParam{}, []__ParsedExpr{}, __pattern.__expr.__line, __pattern.__expr.__column)
	__wildcardBranch := ____rune_private_d1a61e34_makeExpr(__ExprKind_Branch, "_", "", "", "=>", []__ParsedParam{}, []__ParsedExpr{__wildcardPattern, __falseExpr}, __pattern.__expr.__line, __pattern.__expr.__column)
	__matchExpr := ____rune_private_d1a61e34_appendChild(____rune_private_d1a61e34_appendChild(____rune_private_d1a61e34_appendChild(____rune_private_d1a61e34_node(__ExprKind_Match, __op.__token), __subject), __patternBranch), __wildcardBranch)
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __pattern.__state, __expr: __matchExpr}, __minPrec)
}

func ____rune_private_d1a61e34_parseQuestionExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_questionIsPostfixUnwrap(__state) {
			return ____rune_private_d1a61e34_parseExpressionLoop(____rune_private_d1a61e34_parseResultUnwrapExpression(__state, __expr), __minPrec)
		}
		return ____rune_private_d1a61e34_parseTernaryExpression(__state, __expr, __minPrec)
	}()
}

func ____rune_private_d1a61e34_parseQuestionQuestionExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_questionQuestionIsPostfixUnwrap(__state) {
			return ____rune_private_d1a61e34_parseExpressionLoop(____rune_private_d1a61e34_parseResultUnwrapExpressionQQ(__state, __expr), __minPrec)
		}
		return ____rune_private_d1a61e34_parseBinaryExpression(__state, __expr, __minPrec)
	}()
}

func ____rune_private_d1a61e34_parseResultUnwrapExpression(__state __ParserState, __expr __ParsedExpr) __ExprStep {
	__question := ____rune_private_d1a61e34_parserAdvance(__state)
	return __ExprStep{__state: __question.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Unwrap, "?", __question.__token, []__ParsedExpr{__expr})}
}

func ____rune_private_d1a61e34_parseResultUnwrapExpressionQQ(__state __ParserState, __expr __ParsedExpr) __ExprStep {
	__token := ____rune_private_d1a61e34_parserAdvance(__state)
	return __ExprStep{__state: __token.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Unwrap, "??", __token.__token, []__ParsedExpr{__expr})}
}

func ____rune_private_d1a61e34_parseTernaryExpression(__state __ParserState, __condition __ParsedExpr, __minPrec int) __ExprStep {
	__question := ____rune_private_d1a61e34_parserAdvance(__state)
	__consequence := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__question.__state), 1)
	__afterConsequence := ____rune_private_d1a61e34_parserSkipNewlines(__consequence.__state)
	__colon := ____rune_private_d1a61e34_parserMatch(__afterConsequence, __TokenKind_Colon)
	return func() __ExprStep {
		if __colon.__ok {
			return ____rune_private_d1a61e34_parseTernaryAlternative(__colon.__state, __question.__token, __condition, __consequence.__expr, __minPrec)
		}
		return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __afterConsequence, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Ternary, "?:", __question.__token, []__ParsedExpr{__condition, __consequence.__expr})}, __minPrec)
	}()
}

func ____rune_private_d1a61e34_parseTernaryAlternative(__state __ParserState, __token __Token, __condition __ParsedExpr, __consequence __ParsedExpr, __minPrec int) __ExprStep {
	__alternative := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__state), 1)
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __alternative.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Ternary, "?:", __token, []__ParsedExpr{__condition, __consequence, __alternative.__expr})}, __minPrec)
}

func ____rune_private_d1a61e34_parseBinaryExpression(__state __ParserState, __expr __ParsedExpr, __minPrec int) __ExprStep {
	__prec := ____rune_private_d1a61e34_precedence(____rune_private_d1a61e34_parserPeek(__state).__kind)
	return func() __ExprStep {
		if __prec < __minPrec {
			return __ExprStep{__state: __state, __expr: __expr}
		}
		return ____rune_private_d1a61e34_parseBinaryExpressionAtPrec(__state, __expr, __minPrec, __prec)
	}()
}

func ____rune_private_d1a61e34_parseBinaryExpressionAtPrec(__state __ParserState, __expr __ParsedExpr, __minPrec int, __prec int) __ExprStep {
	__op := ____rune_private_d1a61e34_parserAdvance(__state)
	__right := ____rune_private_d1a61e34_parseExpression(__op.__state, __prec+1)
	return ____rune_private_d1a61e34_parseExpressionLoop(__ExprStep{__state: __right.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Binary, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__expr, __right.__expr})}, __minPrec)
}

func ____rune_private_d1a61e34_parseUnary(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Minus) || ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Bang) || ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Tilde) {
			return ____rune_private_d1a61e34_parseUnaryOperator(__state)
		}
		return ____rune_private_d1a61e34_parsePrimary(__state)
	}()
}

func ____rune_private_d1a61e34_parseUnaryOperator(__state __ParserState) __ExprStep {
	__op := ____rune_private_d1a61e34_parserAdvance(__state)
	__right := ____rune_private_d1a61e34_parseExpression(__op.__state, 11)
	return __ExprStep{__state: __right.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Unary, __op.__token.__lexeme, __op.__token, []__ParsedExpr{__right.__expr})}
}

func ____rune_private_d1a61e34_parsePrimary(__state __ParserState) __ExprStep {
	__token := ____rune_private_d1a61e34_parserPeek(__state)
	return func() __ExprStep {
		switch {
		case __token.__kind == __TokenKind_Int:
			return ____rune_private_d1a61e34_parseLiteral(__state, __ExprKind_Int)
		case __token.__kind == __TokenKind_Double:
			return ____rune_private_d1a61e34_parseLiteral(__state, __ExprKind_Double)
		case __token.__kind == __TokenKind_BigInt:
			return ____rune_private_d1a61e34_parseLiteral(__state, __ExprKind_BigInt)
		case __token.__kind == __TokenKind_String:
			return ____rune_private_d1a61e34_parseLiteral(__state, __ExprKind_String)
		case __token.__kind == __TokenKind_TemplateString:
			return ____rune_private_d1a61e34_parseTemplateLiteral(__state)
		case __token.__kind == __TokenKind_Char:
			return ____rune_private_d1a61e34_parseLiteral(__state, __ExprKind_Char)
		case __token.__kind == __TokenKind_Regex:
			return ____rune_private_d1a61e34_parseLiteral(__state, __ExprKind_Regex)
		case __token.__kind == __TokenKind_XMLText:
			return ____rune_private_d1a61e34_parseLiteral(__state, __ExprKind_XMLText)
		case __token.__kind == __TokenKind_Ident:
			return ____rune_private_d1a61e34_parseIdentifierPrimary(__state)
		case __token.__kind == __TokenKind_At:
			return ____rune_private_d1a61e34_parseAtExpression(__state)
		case __token.__kind == __TokenKind_Dot:
			return ____rune_private_d1a61e34_parseThisSelector(__state)
		case __token.__kind == __TokenKind_LBracket:
			return ____rune_private_d1a61e34_parseArrayLiteral(__state)
		case __token.__kind == __TokenKind_Dollar:
			return ____rune_private_d1a61e34_parseDollarExpression(__state)
		case __token.__kind == __TokenKind_LBrace:
			return ____rune_private_d1a61e34_parseBraceLiteral(__state)
		case __token.__kind == __TokenKind_LParen:
			return ____rune_private_d1a61e34_parseParenOrTuple(__state)
		default:
			return ____rune_private_d1a61e34_parsePrimaryError(__state)
		}
	}()
}

func ____rune_private_d1a61e34_parseTemplateLiteral(__state __ParserState) __ExprStep {
	__step := ____rune_private_d1a61e34_parserAdvance(__state)
	__parsed := ____rune_private_d1a61e34_parseTemplateParts(____rune_private_d1a61e34_templateInner(__step.__token.__lexeme), 0, 0, "", []__ParsedExpr{})
	return __ExprStep{__state: __step.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Template, __step.__token.__lexeme, "", "`"+__parsed.__text+"`", "", []__ParsedParam{}, __parsed.__children, __step.__token.__line, __step.__token.__column)}
}

func ____rune_private_d1a61e34_templateInner(__raw string) string {
	return func() string {
		if len([]rune(__raw)) >= 2 {
			return func() string { runes := []rune(__raw); return string(runes[1 : len([]rune(__raw))-1]) }()
		}
		return __raw
	}()
}

func ____rune_private_d1a61e34_parseTemplateParts(__inner string, __index int, __textStart int, __out string, __children []__ParsedExpr) __TemplateParse {
	return func() __TemplateParse {
		if __index >= len([]rune(__inner)) {
			return __TemplateParse{__text: __out + func() string { runes := []rune(__inner); return string(runes[__textStart:len([]rune(__inner))]) }(), __children: __children}
		}
		return ____rune_private_d1a61e34_parseTemplatePartAt(__inner, __index, __textStart, __out, __children)
	}()
}

func ____rune_private_d1a61e34_parseTemplatePartAt(__inner string, __index int, __textStart int, __out string, __children []__ParsedExpr) __TemplateParse {
	__ch := []rune(__inner)[__index]
	return func() __TemplateParse {
		if __ch == '\\' && __index+1 < len([]rune(__inner)) && []rune(__inner)[__index+1] == '(' {
			return ____rune_private_d1a61e34_parseTemplateExprPart(__inner, __index, __textStart, __out, __children)
		}
		return ____rune_private_d1a61e34_parseTemplateParts(__inner, __index+1, __textStart, __out, __children)
	}()
}

func ____rune_private_d1a61e34_parseTemplateExprPart(__inner string, __index int, __textStart int, __out string, __children []__ParsedExpr) __TemplateParse {
	__exprStart := __index + 2
	__exprEnd := ____rune_private_d1a61e34_scanTemplateExprEnd(__inner, __exprStart, 1)
	return func() __TemplateParse {
		if __exprEnd < 0 {
			return __TemplateParse{__text: __out + func() string { runes := []rune(__inner); return string(runes[__textStart:len([]rune(__inner))]) }(), __children: __children}
		}
		return ____rune_private_d1a61e34_parseTemplateParts(__inner, __exprEnd+1, __exprEnd+1, __out+func() string { runes := []rune(__inner); return string(runes[__textStart:__index]) }()+"<<<RUNE_TEMPLATE_PART>>>", ____rune_private_d1a61e34_pushTemplateExpr(__children, func() string { runes := []rune(__inner); return string(runes[__exprStart:__exprEnd]) }()))
	}()
}

func ____rune_private_d1a61e34_scanTemplateExprEnd(__inner string, __index int, __depth int) int {
	return func() int {
		if __index >= len([]rune(__inner)) {
			return -1
		}
		return ____rune_private_d1a61e34_scanTemplateExprEndAt(__inner, __index, __depth)
	}()
}

func ____rune_private_d1a61e34_scanTemplateExprEndAt(__inner string, __index int, __depth int) int {
	__ch := []rune(__inner)[__index]
	return func() int {
		switch {
		case __ch == '"':
			return ____rune_private_d1a61e34_scanTemplateExprEnd(__inner, ____rune_private_d1a61e34_skipTemplateQuoted(__inner, __index+1, '"'), __depth)
		case __ch == '\'':
			return ____rune_private_d1a61e34_scanTemplateExprEnd(__inner, ____rune_private_d1a61e34_skipTemplateQuoted(__inner, __index+1, '\''), __depth)
		case __ch == '`':
			return ____rune_private_d1a61e34_scanTemplateExprEnd(__inner, ____rune_private_d1a61e34_skipTemplateQuoted(__inner, __index+1, '`'), __depth)
		case __ch == '(':
			return ____rune_private_d1a61e34_scanTemplateExprEnd(__inner, __index+1, __depth+1)
		case __ch == ')':
			return func() int {
				if __depth == 1 {
					return __index
				}
				return ____rune_private_d1a61e34_scanTemplateExprEnd(__inner, __index+1, __depth-1)
			}()
		default:
			return ____rune_private_d1a61e34_scanTemplateExprEnd(__inner, __index+1, __depth)
		}
	}()
}

func ____rune_private_d1a61e34_skipTemplateQuoted(__text string, __index int, __quote rune) int {
	return func() int {
		if __index >= len([]rune(__text)) {
			return __index
		}
		return ____rune_private_d1a61e34_skipTemplateQuotedAt(__text, __index, __quote)
	}()
}

func ____rune_private_d1a61e34_skipTemplateQuotedAt(__text string, __index int, __quote rune) int {
	__ch := []rune(__text)[__index]
	return func() int {
		if __ch == '\\' {
			return ____rune_private_d1a61e34_skipTemplateQuoted(__text, __index+2, __quote)
		}
		return func() int {
			if __ch == __quote {
				return __index + 1
			}
			return ____rune_private_d1a61e34_skipTemplateQuoted(__text, __index+1, __quote)
		}()
	}()
}

func ____rune_private_d1a61e34_pushTemplateExpr(__children []__ParsedExpr, __source string) []__ParsedExpr {
	__parsed := ____rune_private_d1a61e34_parseTemplateExpression(strings.TrimSpace(__source))
	return func() []__ParsedExpr {
		out := []__ParsedExpr{}
		out = append(out, __children...)
		out = append(out, __parsed)
		return out
	}()
}

func ____rune_private_d1a61e34_parseTemplateExpression(__source string) __ParsedExpr {
	return func() __ParsedExpr {
		if __source == "" {
			return ____rune_private_d1a61e34_emptyExpr()
		}
		return ____rune_private_d1a61e34_parseExpression(__ParserState{__tokens: __lex(__source), __current: 0, __errors: []__ParseError{}}, 1).__expr
	}()
}

func ____rune_private_d1a61e34_parseLiteral(__state __ParserState, __kind __ExprKind) __ExprStep {
	__step := ____rune_private_d1a61e34_parserAdvance(__state)
	return __ExprStep{__state: __step.__state, __expr: ____rune_private_d1a61e34_valueNode(__kind, __step.__token.__lexeme, __step.__token)}
}

func ____rune_private_d1a61e34_parseIdentifierPrimary(__state __ParserState) __ExprStep {
	__step := ____rune_private_d1a61e34_parserAdvance(__state)
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
			return ____rune_private_d1a61e34_namedNode(__kind, __step.__token.__lexeme, __step.__token)
		}
		return ____rune_private_d1a61e34_valueNode(__kind, __step.__token.__lexeme, __step.__token)
	}()}
}

func ____rune_private_d1a61e34_parseAtExpression(__state __ParserState) __ExprStep {
	__at := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_At, "expected '@'")
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__at.__state, __TokenKind_String) {
			return ____rune_private_d1a61e34_parseAtImportExpression(__at)
		}
		return ____rune_private_d1a61e34_parseAtModuleExpression(__at)
	}()
}

func ____rune_private_d1a61e34_parseAtImportExpression(__at __TokenStep) __ExprStep {
	__path := ____rune_private_d1a61e34_parserAdvance(__at.__state)
	return __ExprStep{__state: __path.__state, __expr: ____rune_private_d1a61e34_valueNode(__ExprKind_At, __path.__token.__lexeme, __at.__token)}
}

func ____rune_private_d1a61e34_parseAtModuleExpression(__at __TokenStep) __ExprStep {
	__name := ____rune_private_d1a61e34_parserConsume(__at.__state, __TokenKind_Ident, "expected module name after '@'")
	return __ExprStep{__state: __name.__state, __expr: ____rune_private_d1a61e34_namedNode(__ExprKind_At, __name.__token.__lexeme, __at.__token)}
}

func ____rune_private_d1a61e34_parseThisSelector(__state __ParserState) __ExprStep {
	__dot := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Dot, "expected '.'")
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__dot.__state, __TokenKind_Ident) {
			return ____rune_private_d1a61e34_parseThisFieldSelector(__dot.__state, __dot.__token)
		}
		return __ExprStep{__state: __dot.__state, __expr: ____rune_private_d1a61e34_node(__ExprKind_This, __dot.__token)}
	}()
}

func ____rune_private_d1a61e34_parseThisFieldSelector(__state __ParserState, __dot __Token) __ExprStep {
	__name := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected field name after '.'")
	return __ExprStep{__state: __name.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Selector, __name.__token.__lexeme, __name.__token.__lexeme, "", ".", []__ParsedParam{}, []__ParsedExpr{____rune_private_d1a61e34_node(__ExprKind_This, __dot)}, __dot.__line, __dot.__column)}
}

func ____rune_private_d1a61e34_parseArrayLiteral(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBracket, "expected '['")
	__args := ____rune_private_d1a61e34_parseArgumentList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), []__ParsedExpr{}, __TokenKind_RBracket)
	__close := ____rune_private_d1a61e34_parserConsume(__args.__state, __TokenKind_RBracket, "expected ']' after array literal")
	return __ExprStep{__state: __close.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Array, "[]", "", "", "", []__ParsedParam{}, __args.__expr.__children, __open.__token.__line, __open.__token.__column)}
}

func ____rune_private_d1a61e34_parseReactiveLiteral(__state __ParserState) __ExprStep {
	__start := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Dollar, "expected '$'")
	return ____rune_private_d1a61e34_parseReactiveLiteralAfterDollar(__start)
}

func ____rune_private_d1a61e34_parseDollarExpression(__state __ParserState) __ExprStep {
	__start := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Dollar, "expected '$'")
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__start.__state, __TokenKind_Ident) {
			return ____rune_private_d1a61e34_parseSignalPrefixIdentifier(__start)
		}
		return __ExprStep{__state: ____rune_private_d1a61e34_parserErrorAt(__start.__state, ____rune_private_d1a61e34_parserPeek(__start.__state), "expected signal name after '$'"), __expr: ____rune_private_d1a61e34_emptyExpr()}
	}()
}

func ____rune_private_d1a61e34_parseSignalPrefixIdentifier(__start __TokenStep) __ExprStep {
	__name := ____rune_private_d1a61e34_parserAdvance(__start.__state)
	return __ExprStep{__state: __name.__state, __expr: ____rune_private_d1a61e34_namedNode(__ExprKind_Identifier, __name.__token.__lexeme, __name.__token)}
}

func ____rune_private_d1a61e34_parseReactiveLiteralAfterDollar(__start __TokenStep) __ExprStep {
	__value := func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__start.__state, __TokenKind_LBracket) {
			return ____rune_private_d1a61e34_parseArrayLiteral(__start.__state)
		}
		return func() __ExprStep {
			if ____rune_private_d1a61e34_parserCheck(__start.__state, __TokenKind_LBrace) {
				return ____rune_private_d1a61e34_parseBraceLiteral(__start.__state)
			}
			return __ExprStep{__state: ____rune_private_d1a61e34_parserErrorAt(__start.__state, ____rune_private_d1a61e34_parserPeek(__start.__state), "expected '[' or '{' after '$'"), __expr: ____rune_private_d1a61e34_emptyExpr()}
		}()
	}()
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_d1a61e34_opNode(__ExprKind_Reactive, "$", __start.__token, []__ParsedExpr{__value.__expr})}
}

func ____rune_private_d1a61e34_parseBraceLiteral(__state __ParserState) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_looksLikeMapLiteralBody(__state) {
			return ____rune_private_d1a61e34_parseMapLiteral(__state)
		}
		return ____rune_private_d1a61e34_parseObjectLiteral(__state)
	}()
}

func ____rune_private_d1a61e34_parseMapLiteral(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBrace, "expected '{'")
	return ____rune_private_d1a61e34_parseMapLiteralLoop(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), ____rune_private_d1a61e34_node(__ExprKind_Map, __open.__token))
}

func ____rune_private_d1a61e34_parseMapLiteralLoop(__state __ParserState, __mapExpr __ParsedExpr) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_d1a61e34_finishMapLiteral(__current, __mapExpr)
		}
		return ____rune_private_d1a61e34_parseMapLiteralEntry(__current, __mapExpr)
	}()
}

func ____rune_private_d1a61e34_parseMapLiteralEntry(__state __ParserState, __mapExpr __ParsedExpr) __ExprStep {
	__key := ____rune_private_d1a61e34_parseExpression(__state, 1)
	__colon := ____rune_private_d1a61e34_parserConsume(__key.__state, __TokenKind_Colon, "expected ':' after map key")
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__colon.__state), 1)
	__entry := ____rune_private_d1a61e34_makeExpr(__ExprKind_Entry, "", "", "", ":", []__ParsedParam{}, []__ParsedExpr{__key.__expr, __value.__expr}, __key.__expr.__line, __key.__expr.__column)
	__nextMap := ____rune_private_d1a61e34_appendChild(__mapExpr, __entry)
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(____rune_private_d1a61e34_consumeStatementEnd(__value.__state)), __TokenKind_Comma)
	return ____rune_private_d1a61e34_parseMapLiteralLoop(__comma.__state, __nextMap)
}

func ____rune_private_d1a61e34_finishMapLiteral(__state __ParserState, __mapExpr __ParsedExpr) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RBrace, "expected '}' after map literal")
	return __ExprStep{__state: __close.__state, __expr: __mapExpr}
}

func ____rune_private_d1a61e34_parseObjectLiteral(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBrace, "expected '{'")
	return ____rune_private_d1a61e34_parseObjectLiteralLoop(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), ____rune_private_d1a61e34_node(__ExprKind_Object, __open.__token))
}

func ____rune_private_d1a61e34_parseObjectLiteralLoop(__state __ParserState, __object __ParsedExpr) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_d1a61e34_finishObjectLiteral(__current, __object)
		}
		return ____rune_private_d1a61e34_parseObjectLiteralMember(__current, __object)
	}()
}

func ____rune_private_d1a61e34_parseObjectLiteralMember(__state __ParserState, __object __ParsedExpr) __ExprStep {
	__privateStep := ____rune_private_d1a61e34_parseObjectPrivateModifier(__state)
	__memberState := __privateStep.__state
	__member := func() __ExprStep {
		if ____rune_private_d1a61e34_looksLikeFunctionDecl(__memberState) {
			return ____rune_private_d1a61e34_parseObjectMethod(__memberState, __privateStep.__ok)
		}
		return ____rune_private_d1a61e34_parseObjectField(__memberState, __privateStep.__ok)
	}()
	__nextObject := ____rune_private_d1a61e34_appendChild(__object, __member.__expr)
	return ____rune_private_d1a61e34_parseObjectLiteralLoop(____rune_private_d1a61e34_consumeFieldSeparator(__member.__state, __TokenKind_RBrace, "expected ',' between object literal fields"), __nextObject)
}

func ____rune_private_d1a61e34_parseObjectMethod(__state __ParserState, __private bool) __ExprStep {
	__fn := ____rune_private_d1a61e34_parseFunctionWithReceiver(__state, "", __private, false, false, ____rune_private_d1a61e34_emptyAnnotations())
	return __ExprStep{__state: __fn.__state, __expr: ____rune_private_d1a61e34_functionToExpr(__fn.__function)}
}

func ____rune_private_d1a61e34_parseObjectField(__state __ParserState, __private bool) __ExprStep {
	__name := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected field name")
	__colon := ____rune_private_d1a61e34_parserConsume(__name.__state, __TokenKind_Colon, "expected ':' after field name")
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__colon.__state), 1)
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_d1a61e34_makeExpr(func() __ExprKind {
		if __private {
			return __ExprKind_PrivateField
		}
		return __ExprKind_Field
	}(), __name.__token.__lexeme, __name.__token.__lexeme, "", ":", []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)}
}

func ____rune_private_d1a61e34_finishObjectLiteral(__state __ParserState, __object __ParsedExpr) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RBrace, "expected '}' after object literal")
	return __ExprStep{__state: __close.__state, __expr: __object}
}

func ____rune_private_d1a61e34_parseStructLiteral(__state __ParserState, __typeExpr __ParsedExpr) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBrace, "expected '{' after type name")
	return ____rune_private_d1a61e34_parseStructLiteralLoop(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), ____rune_private_d1a61e34_namedNode(__ExprKind_Struct, __typeExpr.__name, __open.__token))
}

func ____rune_private_d1a61e34_parseStructLiteralLoop(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_d1a61e34_finishStructLiteral(__current, __structExpr)
		}
		return ____rune_private_d1a61e34_parseStructLiteralField(__current, __structExpr)
	}()
}

func ____rune_private_d1a61e34_parseStructLiteralField(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_DotDot) {
			return ____rune_private_d1a61e34_parseStructLiteralSpreadField(__state, __structExpr)
		}
		return ____rune_private_d1a61e34_parseStructLiteralNamedField(__state, __structExpr)
	}()
}

func ____rune_private_d1a61e34_parseStructLiteralNamedField(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	__name := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected field name")
	__colon := ____rune_private_d1a61e34_parserConsume(__name.__state, __TokenKind_Colon, "expected ':' after field name")
	__value := ____rune_private_d1a61e34_parseExpression(____rune_private_d1a61e34_parserSkipNewlines(__colon.__state), 1)
	__field := ____rune_private_d1a61e34_makeExpr(__ExprKind_Field, __name.__token.__lexeme, __name.__token.__lexeme, "", ":", []__ParsedParam{}, []__ParsedExpr{__value.__expr}, __name.__token.__line, __name.__token.__column)
	__nextStruct := ____rune_private_d1a61e34_appendChild(__structExpr, __field)
	return ____rune_private_d1a61e34_parseStructLiteralLoop(____rune_private_d1a61e34_consumeFieldSeparator(__value.__state, __TokenKind_RBrace, "expected ',' between struct literal fields"), __nextStruct)
}

func ____rune_private_d1a61e34_parseStructLiteralSpreadField(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	__spread := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_DotDot, "expected '..'")
	__value := ____rune_private_d1a61e34_parseExpression(__spread.__state, 1)
	__field := ____rune_private_d1a61e34_opNode(__ExprKind_Spread, "..", __spread.__token, []__ParsedExpr{__value.__expr})
	__nextStruct := ____rune_private_d1a61e34_appendChild(__structExpr, __field)
	return ____rune_private_d1a61e34_parseStructLiteralLoop(____rune_private_d1a61e34_consumeFieldSeparator(__value.__state, __TokenKind_RBrace, "expected ',' between struct literal fields"), __nextStruct)
}

func ____rune_private_d1a61e34_finishStructLiteral(__state __ParserState, __structExpr __ParsedExpr) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RBrace, "expected '}' after struct literal")
	return __ExprStep{__state: __close.__state, __expr: __structExpr}
}

func ____rune_private_d1a61e34_parseParenOrTuple(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LParen, "expected '('")
	__afterOpen := ____rune_private_d1a61e34_parserSkipNewlines(__open.__state)
	return ____rune_private_d1a61e34_parseParenOrTupleAfterOpen(__afterOpen, __open.__token)
}

func ____rune_private_d1a61e34_parseParenOrTupleAfterOpen(__state __ParserState, __open __Token) __ExprStep {
	__expr := ____rune_private_d1a61e34_parseExpression(__state, 1)
	__afterExpr := ____rune_private_d1a61e34_parserSkipNewlines(__expr.__state)
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__afterExpr, __TokenKind_Comma) {
			return ____rune_private_d1a61e34_parseTupleAfterFirst(__afterExpr, __open, __expr.__expr)
		}
		return ____rune_private_d1a61e34_finishParenExpression(__afterExpr, __expr.__expr)
	}()
}

func ____rune_private_d1a61e34_finishParenExpression(__state __ParserState, __expr __ParsedExpr) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RParen, "expected ')' after expression")
	return __ExprStep{__state: __close.__state, __expr: __expr}
}

func ____rune_private_d1a61e34_parseTupleAfterFirst(__state __ParserState, __open __Token, __first __ParsedExpr) __ExprStep {
	__comma := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Comma, "expected ','")
	__holder := ____rune_private_d1a61e34_appendChild(____rune_private_d1a61e34_node(__ExprKind_Tuple, __open), __first)
	__values := ____rune_private_d1a61e34_parseArgumentListLoop(____rune_private_d1a61e34_parserSkipNewlines(__comma.__state), __holder, __TokenKind_RParen)
	__close := ____rune_private_d1a61e34_parserConsume(__values.__state, __TokenKind_RParen, "expected ')' after tuple literal")
	return __ExprStep{__state: __close.__state, __expr: __values.__expr}
}

func ____rune_private_d1a61e34_parsePrimaryError(__state __ParserState) __ExprStep {
	__token := ____rune_private_d1a61e34_parserPeek(__state)
	__step := ____rune_private_d1a61e34_parserAdvance(____rune_private_d1a61e34_parserErrorAt(__state, __token, "expected expression, got "+__tokenKindName(__token.__kind)))
	return __ExprStep{__state: __step.__state, __expr: ____rune_private_d1a61e34_namedNode(__ExprKind_Error, "<error>", __token)}
}

func ____rune_private_d1a61e34_parseLambda(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LParen, "expected '(' before lambda parameters")
	__params := ____rune_private_d1a61e34_parseParamList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state))
	__close := ____rune_private_d1a61e34_parserConsume(__params.__state, __TokenKind_RParen, "expected ')' after lambda parameters")
	__afterParams := ____rune_private_d1a61e34_parserSkipNewlines(__close.__state)
	__ret := ____rune_private_d1a61e34_parserMatch(__afterParams, __TokenKind_Arrow)
	__returnType := func() __TypeRefStep {
		if __ret.__ok {
			return ____rune_private_d1a61e34_parseTypeRef(____rune_private_d1a61e34_parserSkipNewlines(__ret.__state))
		}
		return __TypeRefStep{__state: __afterParams, __typeRef: __emptyParsedTypeRef()}
	}()
	__arrow := ____rune_private_d1a61e34_parserConsume(____rune_private_d1a61e34_parserSkipNewlines(__returnType.__state), __TokenKind_FatArrow, "expected '=>' after lambda parameter")
	__body := ____rune_private_d1a61e34_parseBody(____rune_private_d1a61e34_parserSkipNewlines(__arrow.__state))
	return __ExprStep{__state: __body.__state, __expr: ____rune_private_d1a61e34_withChildren(____rune_private_d1a61e34_withParams(____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Lambda, __open.__token), __typeRefToString(__returnType.__typeRef)), __params.__params), []__ParsedExpr{__body.__expr})}
}

func ____rune_private_d1a61e34_parseMatchExpression(__state __ParserState, __subject __ParsedExpr) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBrace, "expected '{' after match subject")
	__block := ____rune_private_d1a61e34_appendChild(____rune_private_d1a61e34_node(__ExprKind_Match, __open.__token), __subject)
	return ____rune_private_d1a61e34_parsePatternBlock(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), __block)
}

func ____rune_private_d1a61e34_parsePatternBlock(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ExprStep {
		if ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF) {
			return ____rune_private_d1a61e34_finishPatternBlock(__current, __block)
		}
		return ____rune_private_d1a61e34_parsePatternBranch(__current, __block)
	}()
}

func ____rune_private_d1a61e34_parsePatternBranch(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__pattern := ____rune_private_d1a61e34_parsePatternText(__state)
	__arrow := ____rune_private_d1a61e34_parserConsume(__pattern.__state, __TokenKind_FatArrow, "expected '=>' after pattern")
	__value := ____rune_private_d1a61e34_parseBody(____rune_private_d1a61e34_parserSkipNewlines(__arrow.__state))
	__branch := ____rune_private_d1a61e34_makeExpr(__ExprKind_Branch, __pattern.__expr.__text, "", "", "=>", []__ParsedParam{}, []__ParsedExpr{__pattern.__expr, __value.__expr}, __pattern.__expr.__line, __pattern.__expr.__column)
	__nextBlock := ____rune_private_d1a61e34_appendChild(__block, __branch)
	return ____rune_private_d1a61e34_parsePatternBlock(____rune_private_d1a61e34_consumeStatementEnd(__value.__state), __nextBlock)
}

func ____rune_private_d1a61e34_finishPatternBlock(__state __ParserState, __block __ParsedExpr) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RBrace, "expected '}' after pattern block")
	return __ExprStep{__state: __close.__state, __expr: __block}
}

func ____rune_private_d1a61e34_parsePatternText(__state __ParserState) __ExprStep {
	return ____rune_private_d1a61e34_parsePattern(__state)
}

func ____rune_private_d1a61e34_parsePredicatePatternText(__state __ParserState) __ExprStep {
	return ____rune_private_d1a61e34_parsePattern(__state)
}

func ____rune_private_d1a61e34_parsePattern(__state __ParserState) __ExprStep {
	return ____rune_private_d1a61e34_parseOrPatternRest(____rune_private_d1a61e34_parseAliasPattern(__state))
}

func ____rune_private_d1a61e34_parseOrPatternRest(__left __ExprStep) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__left.__state)
	__op := ____rune_private_d1a61e34_parserMatch(__current, __TokenKind_BitOr)
	return func() __ExprStep {
		switch {
		case __op.__ok == true:
			return ____rune_private_d1a61e34_parseOrPatternRest(____rune_private_d1a61e34_parseOrPatternRight(__left, __op.__state))
		default:
			return __ExprStep{__state: __current, __expr: __left.__expr}
		}
	}()
}

func ____rune_private_d1a61e34_parseOrPatternRight(__left __ExprStep, __state __ParserState) __ExprStep {
	__right := ____rune_private_d1a61e34_parseAliasPattern(____rune_private_d1a61e34_parserSkipNewlines(__state))
	return __ExprStep{__state: __right.__state, __expr: ____rune_private_d1a61e34_withText(__left.__expr, __left.__expr.__text+"|"+__right.__expr.__text)}
}

func ____rune_private_d1a61e34_parseAliasPattern(__state __ParserState) __ExprStep {
	__pattern := ____rune_private_d1a61e34_parseRangePattern(__state)
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__pattern.__state)
	__alias := ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_At)
	return func() __ExprStep {
		switch {
		case __alias == true:
			return ____rune_private_d1a61e34_parseAliasPatternName(__pattern, ____rune_private_d1a61e34_parserAdvance(__current).__state)
		default:
			return __ExprStep{__state: __current, __expr: __pattern.__expr}
		}
	}()
}

func ____rune_private_d1a61e34_parseAliasPatternName(__pattern __ExprStep, __state __ParserState) __ExprStep {
	__name := ____rune_private_d1a61e34_parserConsume(____rune_private_d1a61e34_parserSkipNewlines(__state), __TokenKind_Ident, "expected binding name after '@'")
	return __ExprStep{__state: __name.__state, __expr: ____rune_private_d1a61e34_withText(__pattern.__expr, __pattern.__expr.__text+"@"+__name.__token.__lexeme)}
}

func ____rune_private_d1a61e34_parseRangePattern(__state __ParserState) __ExprStep {
	__pattern := ____rune_private_d1a61e34_parsePatternAtom(__state)
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__pattern.__state)
	return func() __ExprStep {
		switch {
		case ____rune_private_d1a61e34_parserPeek(__current).__kind == __TokenKind_DotDotEqual:
			return ____rune_private_d1a61e34_parseRangePatternEnd(__pattern, ____rune_private_d1a61e34_parserAdvance(__current), "..=")
		case ____rune_private_d1a61e34_parserPeek(__current).__kind == __TokenKind_DotDotLess:
			return ____rune_private_d1a61e34_parseRangePatternEnd(__pattern, ____rune_private_d1a61e34_parserAdvance(__current), "..<")
		case ____rune_private_d1a61e34_parserPeek(__current).__kind == __TokenKind_DotDot:
			return ____rune_private_d1a61e34_parseDotDotRangePattern(__pattern, ____rune_private_d1a61e34_parserAdvance(__current))
		default:
			return __ExprStep{__state: __current, __expr: __pattern.__expr}
		}
	}()
}

func ____rune_private_d1a61e34_parseDotDotRangePattern(__pattern __ExprStep, __op __TokenStep) __ExprStep {
	__less := ____rune_private_d1a61e34_parserConsume(__op.__state, __TokenKind_Less, "expected '<' after '..' in range pattern")
	return ____rune_private_d1a61e34_parseRangePatternEnd(__pattern, __less, "..<")
}

func ____rune_private_d1a61e34_parseRangePatternEnd(__pattern __ExprStep, __op __TokenStep, __text string) __ExprStep {
	__bound := ____rune_private_d1a61e34_parseRangeBoundPattern(____rune_private_d1a61e34_parserSkipNewlines(__op.__state))
	return __ExprStep{__state: __bound.__state, __expr: ____rune_private_d1a61e34_withText(__pattern.__expr, __pattern.__expr.__text+__text+__bound.__expr.__text)}
}

func ____rune_private_d1a61e34_parseRangeBoundPattern(__state __ParserState) __ExprStep {
	__token := ____rune_private_d1a61e34_parserPeek(__state)
	return func() __ExprStep {
		switch {
		case __token.__kind == __TokenKind_Underscore:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_Int:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_Double:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_BigInt:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_String:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_Char:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_Ident:
			return ____rune_private_d1a61e34_parseIdentifierRangeBoundPattern(__state)
		default:
			return ____rune_private_d1a61e34_parsePatternError(__state)
		}
	}()
}

func ____rune_private_d1a61e34_parseIdentifierRangeBoundPattern(__state __ParserState) __ExprStep {
	__name := ____rune_private_d1a61e34_parserAdvance(__state)
	__dotted := ____rune_private_d1a61e34_parserMatch(__name.__state, __TokenKind_Dot)
	return func() __ExprStep {
		switch {
		case __dotted.__ok == true:
			return ____rune_private_d1a61e34_parseDottedPatternName(__name.__token, __dotted.__state)
		default:
			return __ExprStep{__state: __name.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __name.__token), __name.__token.__lexeme)}
		}
	}()
}

func ____rune_private_d1a61e34_parsePatternAtom(__state __ParserState) __ExprStep {
	__token := ____rune_private_d1a61e34_parserPeek(__state)
	return func() __ExprStep {
		switch {
		case __token.__kind == __TokenKind_Underscore:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_Int:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_Double:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_BigInt:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_String:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_Char:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		case __token.__kind == __TokenKind_Ident:
			return ____rune_private_d1a61e34_parseIdentifierPattern(__state)
		case __token.__kind == __TokenKind_Less:
			return ____rune_private_d1a61e34_parseComparePattern(__state)
		case __token.__kind == __TokenKind_LessEqual:
			return ____rune_private_d1a61e34_parseComparePattern(__state)
		case __token.__kind == __TokenKind_Greater:
			return ____rune_private_d1a61e34_parseComparePattern(__state)
		case __token.__kind == __TokenKind_GreaterEqual:
			return ____rune_private_d1a61e34_parseComparePattern(__state)
		case __token.__kind == __TokenKind_LBrace:
			return ____rune_private_d1a61e34_parseMapOrObjectPattern(__state)
		case __token.__kind == __TokenKind_LBracket:
			return ____rune_private_d1a61e34_parseArrayPattern(__state)
		case __token.__kind == __TokenKind_LParen:
			return ____rune_private_d1a61e34_parseTupleOrGroupedPattern(__state)
		default:
			return ____rune_private_d1a61e34_parsePatternError(__state)
		}
	}()
}

func ____rune_private_d1a61e34_parsePatternToken(__state __ParserState) __ExprStep {
	__step := ____rune_private_d1a61e34_parserAdvance(__state)
	return __ExprStep{__state: __step.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __step.__token), __step.__token.__lexeme)}
}

func ____rune_private_d1a61e34_parseIdentifierPattern(__state __ParserState) __ExprStep {
	__constructor := ____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_LParen)
	return func() __ExprStep {
		switch {
		case __constructor == true:
			return ____rune_private_d1a61e34_parseConstructorPattern(__state)
		default:
			return ____rune_private_d1a61e34_parseIdentifierPatternNonConstructor(__state)
		}
	}()
}

func ____rune_private_d1a61e34_parseIdentifierPatternNonConstructor(__state __ParserState) __ExprStep {
	__qualified := ____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_Dot)
	return func() __ExprStep {
		switch {
		case __qualified == true:
			return ____rune_private_d1a61e34_parseQualifiedIdentifierPattern(__state)
		default:
			return ____rune_private_d1a61e34_parsePatternToken(__state)
		}
	}()
}

func ____rune_private_d1a61e34_parseQualifiedIdentifierPattern(__state __ParserState) __ExprStep {
	__name := ____rune_private_d1a61e34_parserAdvance(__state)
	__dot := ____rune_private_d1a61e34_parserConsume(__name.__state, __TokenKind_Dot, "expected '.' after pattern qualifier")
	return ____rune_private_d1a61e34_parseDottedPatternName(__name.__token, __dot.__state)
}

func ____rune_private_d1a61e34_parseDottedPatternName(__first __Token, __state __ParserState) __ExprStep {
	__second := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected name after '.'")
	return __ExprStep{__state: __second.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __first), __first.__lexeme+"."+__second.__token.__lexeme)}
}

func ____rune_private_d1a61e34_parseComparePattern(__state __ParserState) __ExprStep {
	__op := ____rune_private_d1a61e34_parserAdvance(__state)
	__value := ____rune_private_d1a61e34_parseRangeBoundPattern(____rune_private_d1a61e34_parserSkipNewlines(__op.__state))
	return __ExprStep{__state: __value.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __op.__token), __op.__token.__lexeme+__value.__expr.__text)}
}

func ____rune_private_d1a61e34_parseConstructorPattern(__state __ParserState) __ExprStep {
	__name := ____rune_private_d1a61e34_parserAdvance(__state)
	__open := ____rune_private_d1a61e34_parserConsume(__name.__state, __TokenKind_LParen, "expected '(' after pattern constructor")
	__args := ____rune_private_d1a61e34_parseConstructorPatternArgs(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), "")
	__close := ____rune_private_d1a61e34_parserConsume(__args.__state, __TokenKind_RParen, "expected ')' after constructor pattern")
	return __ExprStep{__state: __close.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __name.__token), __name.__token.__lexeme+"("+__args.__expr.__text+")")}
}

func ____rune_private_d1a61e34_parseConstructorPatternArgs(__state __ParserState, __out string) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	__done := ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RParen) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF)
	return func() __ExprStep {
		switch {
		case __done == true:
			return __ExprStep{__state: __current, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, ____rune_private_d1a61e34_parserPeek(__current)), __out)}
		default:
			return ____rune_private_d1a61e34_parseConstructorPatternArg(__current, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseConstructorPatternArg(__state __ParserState, __out string) __ExprStep {
	__rest := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_DotDot)
	return func() __ExprStep {
		switch {
		case __rest.__ok == true:
			return ____rune_private_d1a61e34_parseConstructorPatternAfterArg(__rest.__state, ____rune_private_d1a61e34_appendPatternPart(__out, ".."), true)
		default:
			return ____rune_private_d1a61e34_parseConstructorPatternValue(__state, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseConstructorPatternValue(__state __ParserState, __out string) __ExprStep {
	__arg := ____rune_private_d1a61e34_parsePattern(__state)
	return ____rune_private_d1a61e34_parseConstructorPatternAfterArg(__arg.__state, ____rune_private_d1a61e34_appendPatternPart(__out, __arg.__expr.__text), false)
}

func ____rune_private_d1a61e34_parseConstructorPatternAfterArg(__state __ParserState, __out string, __rest bool) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	__comma := ____rune_private_d1a61e34_parserMatch(__current, __TokenKind_Comma)
	__done := __rest || __comma.__ok == false
	return func() __ExprStep {
		switch {
		case __done == true:
			return __ExprStep{__state: __comma.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, ____rune_private_d1a61e34_parserPeek(__comma.__state)), __out)}
		default:
			return ____rune_private_d1a61e34_parseConstructorPatternArgs(__comma.__state, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseArrayPattern(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBracket, "expected '[' before array pattern")
	__parts := ____rune_private_d1a61e34_parseArrayPatternParts(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), "")
	__close := ____rune_private_d1a61e34_parserConsume(__parts.__state, __TokenKind_RBracket, "expected ']' after array pattern")
	return __ExprStep{__state: __close.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __open.__token), "["+__parts.__expr.__text+"]")}
}

func ____rune_private_d1a61e34_parseArrayPatternParts(__state __ParserState, __out string) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	__done := ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBracket) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF)
	return func() __ExprStep {
		switch {
		case __done == true:
			return __ExprStep{__state: __current, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, ____rune_private_d1a61e34_parserPeek(__current)), __out)}
		default:
			return ____rune_private_d1a61e34_parseArrayPatternPart(__current, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseArrayPatternPart(__state __ParserState, __out string) __ExprStep {
	__spread := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_DotDot)
	return func() __ExprStep {
		switch {
		case __spread.__ok == true:
			return ____rune_private_d1a61e34_parseArraySpreadPattern(__spread.__state, __out)
		default:
			return ____rune_private_d1a61e34_parseArrayValuePattern(__state, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseArraySpreadPattern(__state __ParserState, __out string) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __ExprStep {
		switch {
		case ____rune_private_d1a61e34_parserPeek(__current).__kind == __TokenKind_String:
			return ____rune_private_d1a61e34_parseArraySpreadValue(__current, __out)
		case ____rune_private_d1a61e34_parserPeek(__current).__kind == __TokenKind_Ident:
			return ____rune_private_d1a61e34_parseArraySpreadIdentifier(__current, __out)
		default:
			return ____rune_private_d1a61e34_parseArrayPatternAfterPart(__current, ____rune_private_d1a61e34_appendPatternPart(__out, ".."))
		}
	}()
}

func ____rune_private_d1a61e34_parseArraySpreadIdentifier(__state __ParserState, __out string) __ExprStep {
	__identifier := ____rune_private_d1a61e34_parserAdvance(__state)
	__constantSpread := ____rune_private_d1a61e34_isPatternSpreadIdentifier(__identifier.__token.__lexeme)
	__text := func() string {
		switch {
		case __constantSpread == true:
			return ".. " + __identifier.__token.__lexeme
		default:
			return ".." + __identifier.__token.__lexeme
		}
	}()
	return ____rune_private_d1a61e34_parseArrayPatternAfterPart(__identifier.__state, ____rune_private_d1a61e34_appendPatternPart(__out, __text))
}

func ____rune_private_d1a61e34_parseArraySpreadValue(__state __ParserState, __out string) __ExprStep {
	__value := ____rune_private_d1a61e34_parserAdvance(__state)
	return ____rune_private_d1a61e34_parseArrayPatternAfterPart(__value.__state, ____rune_private_d1a61e34_appendPatternPart(__out, ".."+__value.__token.__lexeme))
}

func ____rune_private_d1a61e34_parseArrayValuePattern(__state __ParserState, __out string) __ExprStep {
	__value := ____rune_private_d1a61e34_parsePattern(__state)
	return ____rune_private_d1a61e34_parseArrayPatternAfterPart(__value.__state, ____rune_private_d1a61e34_appendPatternPart(__out, __value.__expr.__text))
}

func ____rune_private_d1a61e34_parseArrayPatternAfterPart(__state __ParserState, __out string) __ExprStep {
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__state), __TokenKind_Comma)
	return func() __ExprStep {
		switch {
		case __comma.__ok == true:
			return ____rune_private_d1a61e34_parseArrayPatternParts(__comma.__state, __out)
		default:
			return __ExprStep{__state: __comma.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, ____rune_private_d1a61e34_parserPeek(__comma.__state)), __out)}
		}
	}()
}

func ____rune_private_d1a61e34_parseMapOrObjectPattern(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LBrace, "expected '{' before pattern")
	__parts := ____rune_private_d1a61e34_parseMapOrObjectPatternParts(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), "")
	__close := ____rune_private_d1a61e34_parserConsume(__parts.__state, __TokenKind_RBrace, "expected '}' after pattern")
	return __ExprStep{__state: __close.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __open.__token), "{"+__parts.__expr.__text+"}")}
}

func ____rune_private_d1a61e34_parseMapOrObjectPatternParts(__state __ParserState, __out string) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	__done := ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RBrace) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF)
	return func() __ExprStep {
		switch {
		case __done == true:
			return __ExprStep{__state: __current, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, ____rune_private_d1a61e34_parserPeek(__current)), __out)}
		default:
			return ____rune_private_d1a61e34_parseMapOrObjectPatternPart(__current, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseMapOrObjectPatternPart(__state __ParserState, __out string) __ExprStep {
	__rest := ____rune_private_d1a61e34_parserMatch(__state, __TokenKind_DotDot)
	return func() __ExprStep {
		switch {
		case __rest.__ok == true:
			return ____rune_private_d1a61e34_parseMapOrObjectPatternAfterPart(__rest.__state, ____rune_private_d1a61e34_appendPatternPart(__out, ".."))
		default:
			return ____rune_private_d1a61e34_parseMapOrObjectEntryPattern(__state, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseMapOrObjectEntryPattern(__state __ParserState, __out string) __ExprStep {
	__objectField := ____rune_private_d1a61e34_parserPatternLooksLikeObjectField(__state)
	return func() __ExprStep {
		switch {
		case __objectField == true:
			return ____rune_private_d1a61e34_parseObjectFieldPattern(__state, __out)
		default:
			return ____rune_private_d1a61e34_parseMapEntryPattern(__state, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseObjectFieldPattern(__state __ParserState, __out string) __ExprStep {
	__field := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_Ident, "expected object pattern field")
	__optional := ____rune_private_d1a61e34_parserMatch(__field.__state, __TokenKind_Question)
	__colon := ____rune_private_d1a61e34_parserMatch(__optional.__state, __TokenKind_Colon)
	return func() __ExprStep {
		switch {
		case __colon.__ok == true:
			return ____rune_private_d1a61e34_parseObjectFieldValuePattern(__field.__token, __optional.__ok, __colon.__state, __out)
		default:
			return ____rune_private_d1a61e34_parseMapOrObjectPatternAfterPart(__optional.__state, ____rune_private_d1a61e34_appendPatternPart(__out, ____rune_private_d1a61e34_objectFieldPatternText(__field.__token.__lexeme, __optional.__ok, __field.__token.__lexeme)))
		}
	}()
}

func ____rune_private_d1a61e34_parseObjectFieldValuePattern(__field __Token, __optional bool, __state __ParserState, __out string) __ExprStep {
	__value := ____rune_private_d1a61e34_parsePattern(____rune_private_d1a61e34_parserSkipNewlines(__state))
	return ____rune_private_d1a61e34_parseMapOrObjectPatternAfterPart(__value.__state, ____rune_private_d1a61e34_appendPatternPart(__out, ____rune_private_d1a61e34_objectFieldPatternText(__field.__lexeme, __optional, __value.__expr.__text)))
}

func ____rune_private_d1a61e34_objectFieldPatternText(__name string, __optional bool, __value string) string {
	return func() string {
		switch {
		case __optional == true:
			return __name + "?:" + __value
		default:
			return __name + ":" + __value
		}
	}()
}

func ____rune_private_d1a61e34_parseMapEntryPattern(__state __ParserState, __out string) __ExprStep {
	__key := ____rune_private_d1a61e34_parseMapPatternKey(__state)
	__optional := ____rune_private_d1a61e34_parserMatch(__key.__state, __TokenKind_Question)
	__colon := ____rune_private_d1a61e34_parserConsume(__optional.__state, __TokenKind_Colon, "expected ':' after map pattern key")
	__value := ____rune_private_d1a61e34_parsePattern(____rune_private_d1a61e34_parserSkipNewlines(__colon.__state))
	__entry := func() string {
		switch {
		case __optional.__ok == true:
			return __key.__expr.__text + "?:" + __value.__expr.__text
		default:
			return __key.__expr.__text + ":" + __value.__expr.__text
		}
	}()
	return ____rune_private_d1a61e34_parseMapOrObjectPatternAfterPart(__value.__state, ____rune_private_d1a61e34_appendPatternPart(__out, __entry))
}

func ____rune_private_d1a61e34_parseMapPatternKey(__state __ParserState) __ExprStep {
	return ____rune_private_d1a61e34_parseRangeBoundPattern(__state)
}

func ____rune_private_d1a61e34_parseMapOrObjectPatternAfterPart(__state __ParserState, __out string) __ExprStep {
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__state), __TokenKind_Comma)
	return func() __ExprStep {
		switch {
		case __comma.__ok == true:
			return ____rune_private_d1a61e34_parseMapOrObjectPatternParts(__comma.__state, __out)
		default:
			return __ExprStep{__state: __comma.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, ____rune_private_d1a61e34_parserPeek(__comma.__state)), __out)}
		}
	}()
}

func ____rune_private_d1a61e34_parseTupleOrGroupedPattern(__state __ParserState) __ExprStep {
	__open := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_LParen, "expected '(' before pattern")
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__open.__state)
	__empty := ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RParen)
	return func() __ExprStep {
		switch {
		case __empty == true:
			return ____rune_private_d1a61e34_finishEmptyTuplePattern(__open.__token, __current)
		default:
			return ____rune_private_d1a61e34_parseTupleOrGroupedPatternFirst(__open.__token, __current)
		}
	}()
}

func ____rune_private_d1a61e34_finishEmptyTuplePattern(__open __Token, __state __ParserState) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(__state, __TokenKind_RParen, "expected ')' after tuple pattern")
	return __ExprStep{__state: __close.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __open), "()")}
}

func ____rune_private_d1a61e34_parseTupleOrGroupedPatternFirst(__open __Token, __state __ParserState) __ExprStep {
	__first := ____rune_private_d1a61e34_parsePattern(__state)
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__first.__state), __TokenKind_Comma)
	return func() __ExprStep {
		switch {
		case __comma.__ok == true:
			return ____rune_private_d1a61e34_parseTuplePatternRest(__open, __comma.__state, __first.__expr.__text)
		default:
			return ____rune_private_d1a61e34_finishGroupedPattern(__first)
		}
	}()
}

func ____rune_private_d1a61e34_finishGroupedPattern(__first __ExprStep) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(____rune_private_d1a61e34_parserSkipNewlines(__first.__state), __TokenKind_RParen, "expected ')' after pattern")
	return __ExprStep{__state: __close.__state, __expr: __first.__expr}
}

func ____rune_private_d1a61e34_parseTuplePatternRest(__open __Token, __state __ParserState, __out string) __ExprStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	__done := ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_RParen) || ____rune_private_d1a61e34_parserCheck(__current, __TokenKind_EOF)
	return func() __ExprStep {
		switch {
		case __done == true:
			return ____rune_private_d1a61e34_finishTuplePattern(__open, __current, __out)
		default:
			return ____rune_private_d1a61e34_parseTuplePatternPart(__open, __current, __out)
		}
	}()
}

func ____rune_private_d1a61e34_parseTuplePatternPart(__open __Token, __state __ParserState, __out string) __ExprStep {
	__value := ____rune_private_d1a61e34_parsePattern(__state)
	__next := ____rune_private_d1a61e34_appendPatternPart(__out, __value.__expr.__text)
	__comma := ____rune_private_d1a61e34_parserMatch(____rune_private_d1a61e34_parserSkipNewlines(__value.__state), __TokenKind_Comma)
	return func() __ExprStep {
		switch {
		case __comma.__ok == true:
			return ____rune_private_d1a61e34_parseTuplePatternRest(__open, __comma.__state, __next)
		default:
			return ____rune_private_d1a61e34_finishTuplePattern(__open, __comma.__state, __next)
		}
	}()
}

func ____rune_private_d1a61e34_finishTuplePattern(__open __Token, __state __ParserState, __out string) __ExprStep {
	__close := ____rune_private_d1a61e34_parserConsume(____rune_private_d1a61e34_parserSkipNewlines(__state), __TokenKind_RParen, "expected ')' after tuple pattern")
	return __ExprStep{__state: __close.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __open), "("+__out+")")}
}

func ____rune_private_d1a61e34_appendPatternPart(__out string, __part string) string {
	return func() string {
		if __out == "" {
			return __part
		}
		return __out + "," + __part
	}()
}

func ____rune_private_d1a61e34_parsePatternError(__state __ParserState) __ExprStep {
	__token := ____rune_private_d1a61e34_parserPeek(__state)
	__step := ____rune_private_d1a61e34_parserAdvance(____rune_private_d1a61e34_parserErrorAt(__state, __token, "expected pattern"))
	return __ExprStep{__state: __step.__state, __expr: ____rune_private_d1a61e34_withText(____rune_private_d1a61e34_node(__ExprKind_Pattern, __token), "_")}
}

func ____rune_private_d1a61e34_functionToExpr(__fn __ParsedFunction) __ParsedExpr {
	return ____rune_private_d1a61e34_makeExpr(func() __ExprKind {
		if __fn.__private {
			return __ExprKind_PrivateMethod
		}
		return __ExprKind_Method
	}(), __typeRefToString(__fn.__returnType), __fn.__name, "", "=>", __fn.__params, []__ParsedExpr{__fn.__body}, __fn.__line, __fn.__column)
}

func ____rune_private_d1a61e34_calleeText(__expr __ParsedExpr) string {
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

func ____rune_private_d1a61e34_precedence(__kind __TokenKind) int {
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

func ____rune_private_d1a61e34_parseAnnotations(__state __ParserState) __AnnotationListStep {
	return ____rune_private_d1a61e34_parseAnnotationsLoop(__state, ____rune_private_d1a61e34_emptyAnnotations())
}

func ____rune_private_d1a61e34_parseAnnotationsLoop(__state __ParserState, __annotations []__ParsedAnnotation) __AnnotationListStep {
	__current := ____rune_private_d1a61e34_parserSkipNewlines(__state)
	return func() __AnnotationListStep {
		if ____rune_private_d1a61e34_looksLikeAnnotation(__current) {
			return ____rune_private_d1a61e34_parseAnnotationsNext(__current, __annotations)
		}
		return __AnnotationListStep{__state: __current, __annotations: __annotations}
	}()
}

func ____rune_private_d1a61e34_parseAnnotationsNext(__state __ParserState, __annotations []__ParsedAnnotation) __AnnotationListStep {
	__step := ____rune_private_d1a61e34_parseAnnotation(__state)
	return ____rune_private_d1a61e34_parseAnnotationsLoop(__step.__state, func() []__ParsedAnnotation {
		out := []__ParsedAnnotation{}
		out = append(out, __annotations...)
		out = append(out, __step.__annotation)
		return out
	}())
}

func ____rune_private_d1a61e34_looksLikeAnnotation(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Hash) || ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_At) && ____rune_private_d1a61e34_parserCheckNext(__state, __TokenKind_Ident)
}

func ____rune_private_d1a61e34_parseAnnotation(__state __ParserState) __AnnotationStep {
	__marker := ____rune_private_d1a61e34_parserAdvance(__state)
	__first := ____rune_private_d1a61e34_parserConsume(__marker.__state, __TokenKind_Ident, "expected annotation name")
	__dot := ____rune_private_d1a61e34_parserMatch(__first.__state, __TokenKind_Dot)
	__name := func() __TokenStep {
		if __dot.__ok {
			return ____rune_private_d1a61e34_parserConsume(__dot.__state, __TokenKind_Ident, "expected annotation function name after '.'")
		}
		return __first
	}()
	__open := ____rune_private_d1a61e34_parserMatch(__name.__state, __TokenKind_LParen)
	__args := func() __ExprStep {
		if __open.__ok {
			return ____rune_private_d1a61e34_parseArgumentList(____rune_private_d1a61e34_parserSkipNewlines(__open.__state), append([]__ParsedExpr{}, []__ParsedExpr{____rune_private_d1a61e34_emptyExpr()}[0:0]...), __TokenKind_RParen)
		}
		return __ExprStep{__state: __name.__state, __expr: ____rune_private_d1a61e34_makeExpr(__ExprKind_Args, "", "", "", "", []__ParsedParam{}, []__ParsedExpr{}, 0, 0)}
	}()
	__close := func() __TokenStep {
		if __open.__ok {
			return ____rune_private_d1a61e34_parserConsume(__args.__state, __TokenKind_RParen, "expected ')' after annotation arguments")
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

func ____rune_private_d1a61e34_skipBalanced(__state __ParserState, __openKind __TokenKind, __closeKind __TokenKind, __depth int) __ParserState {
	return func() __ParserState {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_EOF) || __depth <= 0 {
			return __state
		}
		return ____rune_private_d1a61e34_skipBalancedStep(__state, __openKind, __closeKind, __depth)
	}()
}

func ____rune_private_d1a61e34_skipBalancedStep(__state __ParserState, __openKind __TokenKind, __closeKind __TokenKind, __depth int) __ParserState {
	__token := ____rune_private_d1a61e34_parserPeek(__state)
	__step := ____rune_private_d1a61e34_parserAdvance(__state)
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
	return ____rune_private_d1a61e34_skipBalanced(__step.__state, __openKind, __closeKind, __nextDepth)
}

func ____rune_private_d1a61e34_questionIsPostfixUnwrap(__state __ParserState) bool {
	__next := ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+1)
	return __next == __TokenKind_EOF || __next == __TokenKind_Newline || __next == __TokenKind_RParen || __next == __TokenKind_RBracket || __next == __TokenKind_RBrace || __next == __TokenKind_Comma
}

func ____rune_private_d1a61e34_questionQuestionIsPostfixUnwrap(__state __ParserState) bool {
	__next := ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+1)
	return __next == __TokenKind_EOF || __next == __TokenKind_Newline || __next == __TokenKind_RParen || __next == __TokenKind_RBracket || __next == __TokenKind_RBrace || __next == __TokenKind_Comma
}

func ____rune_private_d1a61e34_looksLikeTypeDecl(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) && ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, ____rune_private_d1a61e34_skipGenericNamesAt(__state, __state.__current+1))) == __TokenKind_Colon
}

func ____rune_private_d1a61e34_looksLikeFunctionDecl(__state __ParserState) bool {
	__start := func() int {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Tilde) {
			return ____rune_private_d1a61e34_skipNewlinesAt(__state, __state.__current+1)
		}
		return __state.__current
	}()
	return ____rune_private_d1a61e34_parserKindAt(__state, __start) == __TokenKind_Ident && ____rune_private_d1a61e34_looksLikeFunctionAfterName(__state, ____rune_private_d1a61e34_skipGenericNamesAt(__state, __start+1))
}

func ____rune_private_d1a61e34_looksLikeMacroFunctionDecl(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Hash) && ____rune_private_d1a61e34_looksLikeFunctionDecl(____rune_private_d1a61e34_stateAt(__state, __state.__current+1))
}

func ____rune_private_d1a61e34_looksLikeStaticFunctionDecl(__state __ParserState) bool {
	__markerState := func() __ParserState {
		if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_DoubleColon) {
			return ____rune_private_d1a61e34_stateAt(__state, __state.__current+1)
		}
		return func() __ParserState {
			if ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_Ident) && ____rune_private_d1a61e34_parserPeek(__state).__lexeme == "static" {
				return ____rune_private_d1a61e34_stateAt(__state, __state.__current+1)
			}
			return __state
		}()
	}()
	return __markerState.__current != __state.__current && ____rune_private_d1a61e34_looksLikeFunctionDecl(____rune_private_d1a61e34_parserSkipNewlines(__markerState))
}

func ____rune_private_d1a61e34_looksLikeFunctionAfterName(__state __ParserState, __index int) bool {
	return ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_LParen && ____rune_private_d1a61e34_looksLikeFunctionAfterParams(__state, ____rune_private_d1a61e34_skipBalancedAt(__state, __index, __TokenKind_LParen, __TokenKind_RParen))
}

func ____rune_private_d1a61e34_looksLikeFunctionAfterParams(__state __ParserState, __index int) bool {
	__afterParams := ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)
	__afterReturn := func() int {
		if ____rune_private_d1a61e34_parserKindAt(__state, __afterParams) == __TokenKind_Arrow {
			return ____rune_private_d1a61e34_skipTypeNameTokensAt(__state, __afterParams+1)
		}
		return __afterParams
	}()
	return ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __afterReturn)) == __TokenKind_FatArrow
}

func ____rune_private_d1a61e34_looksLikeLambda(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LParen) && ____rune_private_d1a61e34_looksLikeLambdaAfterParams(__state, ____rune_private_d1a61e34_skipBalancedAt(__state, __state.__current, __TokenKind_LParen, __TokenKind_RParen))
}

func ____rune_private_d1a61e34_looksLikeLambdaAfterParams(__state __ParserState, __index int) bool {
	__afterParams := ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)
	__afterReturn := func() int {
		if ____rune_private_d1a61e34_parserKindAt(__state, __afterParams) == __TokenKind_Arrow {
			return ____rune_private_d1a61e34_skipTypeNameTokensAt(__state, __afterParams+1)
		}
		return __afterParams
	}()
	return ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __afterReturn)) == __TokenKind_FatArrow
}

func ____rune_private_d1a61e34_skipAnnotationsAt(__state __ParserState, __index int) int {
	__current := ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)
	__annotation := ____rune_private_d1a61e34_looksLikeAnnotationAt(__state, __current)
	return func() int {
		switch {
		case __annotation == true:
			return ____rune_private_d1a61e34_skipAnnotationsAt(__state, ____rune_private_d1a61e34_skipAnnotationAt(__state, __current))
		default:
			return __current
		}
	}()
}

func ____rune_private_d1a61e34_looksLikeAnnotationAt(__state __ParserState, __index int) bool {
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __index)
	return func() bool {
		switch {
		case __kind == __TokenKind_Hash:
			return ____rune_private_d1a61e34_parserKindAt(__state, __index+1) == __TokenKind_Ident
		case __kind == __TokenKind_At:
			return ____rune_private_d1a61e34_parserKindAt(__state, __index+1) == __TokenKind_Ident
		default:
			return false
		}
	}()
}

func ____rune_private_d1a61e34_skipAnnotationAt(__state __ParserState, __index int) int {
	__afterName := ____rune_private_d1a61e34_skipAnnotationNameAt(__state, __index+2)
	__afterNewlines := ____rune_private_d1a61e34_skipNewlinesAt(__state, __afterName)
	__hasArgs := ____rune_private_d1a61e34_parserKindAt(__state, __afterNewlines) == __TokenKind_LParen
	return func() int {
		switch {
		case __hasArgs == true:
			return ____rune_private_d1a61e34_skipNewlinesAt(__state, ____rune_private_d1a61e34_skipBalancedAt(__state, __afterNewlines, __TokenKind_LParen, __TokenKind_RParen))
		default:
			return __afterNewlines
		}
	}()
}

func ____rune_private_d1a61e34_skipAnnotationNameAt(__state __ParserState, __index int) int {
	__qualified := ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_Dot && ____rune_private_d1a61e34_parserKindAt(__state, __index+1) == __TokenKind_Ident
	return func() int {
		switch {
		case __qualified == true:
			return __index + 2
		default:
			return __index
		}
	}()
}

func ____rune_private_d1a61e34_looksLikeEnumMember(__state __ParserState) bool {
	__afterAnnotations := ____rune_private_d1a61e34_skipAnnotationsAt(__state, __state.__current)
	__plus := ____rune_private_d1a61e34_parserKindAt(__state, __afterAnnotations) == __TokenKind_Plus
	__start := func() int {
		switch {
		case __plus == true:
			return ____rune_private_d1a61e34_skipNewlinesAt(__state, __afterAnnotations+1)
		default:
			return __afterAnnotations
		}
	}()
	__token := ____rune_private_d1a61e34_parserTokenAt(__state, __start)
	__next := ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __start+1))
	return __token.__kind == __TokenKind_Ident && (__next == __TokenKind_Assign || __next == __TokenKind_LParen && ____rune_private_d1a61e34_startsWithUpper(__token.__lexeme) || __next != __TokenKind_Colon && __next != __TokenKind_LParen)
}

func ____rune_private_d1a61e34_startsWithUpper(__name string) bool {
	return func() bool {
		if len([]rune(__name)) > 0 {
			return []rune(__name)[0] >= 'A' && []rune(__name)[0] <= 'Z'
		}
		return false
	}()
}

func ____rune_private_d1a61e34_looksLikeObjectDestructureDecl(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LBrace) && ____rune_private_d1a61e34_scanObjectDestructureDecl(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __state.__current+1))
}

func ____rune_private_d1a61e34_scanObjectDestructureDecl(__state __ParserState, __index int) bool {
	return func() bool {
		if ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_RBrace {
			return false
		}
		return ____rune_private_d1a61e34_scanObjectDestructureField(__state, __index)
	}()
}

func ____rune_private_d1a61e34_scanObjectDestructureField(__state __ParserState, __index int) bool {
	return func() bool {
		if ____rune_private_d1a61e34_parserKindAt(__state, __index) != __TokenKind_Ident {
			return false
		}
		return ____rune_private_d1a61e34_scanObjectDestructureAfterField(__state, __index+1)
	}()
}

func ____rune_private_d1a61e34_scanObjectDestructureAfterField(__state __ParserState, __index int) bool {
	return func() bool {
		if ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_Colon {
			return ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __index+1)) == __TokenKind_Ident && ____rune_private_d1a61e34_scanObjectDestructureAfterName(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __index+1)+1)
		}
		return ____rune_private_d1a61e34_scanObjectDestructureAfterName(__state, __index)
	}()
}

func ____rune_private_d1a61e34_scanObjectDestructureAfterName(__state __ParserState, __index int) bool {
	__current := ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __current)
	return func() bool {
		if __kind == __TokenKind_Comma {
			return ____rune_private_d1a61e34_scanObjectDestructureAfterComma(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __current+1))
		}
		return func() bool {
			if __kind == __TokenKind_RBrace {
				return ____rune_private_d1a61e34_scanObjectDestructureAfterClose(__state, __current+1)
			}
			return false
		}()
	}()
}

func ____rune_private_d1a61e34_scanObjectDestructureAfterComma(__state __ParserState, __index int) bool {
	return func() bool {
		if ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_RBrace {
			return ____rune_private_d1a61e34_scanObjectDestructureAfterClose(__state, __index+1)
		}
		return ____rune_private_d1a61e34_scanObjectDestructureField(__state, __index)
	}()
}

func ____rune_private_d1a61e34_scanObjectDestructureAfterClose(__state __ParserState, __index int) bool {
	__afterClose := ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)
	return ____rune_private_d1a61e34_parserKindAt(__state, __afterClose) == __TokenKind_Declare || ____rune_private_d1a61e34_parserKindAt(__state, __afterClose) == __TokenKind_MutDeclare
}

func ____rune_private_d1a61e34_looksLikePatternBranch(__state __ParserState) bool {
	return ____rune_private_d1a61e34_tokensLookLikePatternBranch(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __state.__current))
}

func ____rune_private_d1a61e34_looksLikePatternBlockAfterSubject(__state __ParserState) bool {
	return ____rune_private_d1a61e34_parserCheck(__state, __TokenKind_LBrace) && ____rune_private_d1a61e34_tokensLookLikePatternBranch(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __state.__current+1))
}

func ____rune_private_d1a61e34_tokensLookLikePatternBranch(__state __ParserState, __index int) bool {
	__afterPattern := ____rune_private_d1a61e34_skipPatternLookahead(__state, __index)
	return __afterPattern >= 0 && ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __afterPattern)) == __TokenKind_FatArrow
}

func ____rune_private_d1a61e34_skipPatternLookahead(__state __ParserState, __index int) int {
	return ____rune_private_d1a61e34_skipOrPatternLookahead(__state, ____rune_private_d1a61e34_skipSinglePatternLookahead(__state, __index))
}

func ____rune_private_d1a61e34_skipOrPatternLookahead(__state __ParserState, __index int) int {
	return func() int {
		if __index >= 0 && ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)) == __TokenKind_BitOr {
			return ____rune_private_d1a61e34_skipOrPatternLookahead(__state, ____rune_private_d1a61e34_skipSinglePatternLookahead(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)+1))
		}
		return __index
	}()
}

func ____rune_private_d1a61e34_skipAliasPatternLookahead(__state __ParserState, __index int) int {
	return func() int {
		if __index >= 0 && ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)) == __TokenKind_At && ____rune_private_d1a61e34_parserKindAt(__state, ____rune_private_d1a61e34_skipNewlinesAt(__state, __index)+1) == __TokenKind_Ident {
			return ____rune_private_d1a61e34_skipNewlinesAt(__state, __index) + 2
		}
		return __index
	}()
}

func ____rune_private_d1a61e34_skipSinglePatternLookahead(__state __ParserState, __index int) int {
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __index)
	__after := func() int {
		if __kind == __TokenKind_Underscore || __kind == __TokenKind_Int || __kind == __TokenKind_Double || __kind == __TokenKind_BigInt || __kind == __TokenKind_String || __kind == __TokenKind_Char {
			return __index + 1
		}
		return func() int {
			if __kind == __TokenKind_Ident {
				return ____rune_private_d1a61e34_skipIdentifierPatternLookahead(__state, __index)
			}
			return func() int {
				if __kind == __TokenKind_Less || __kind == __TokenKind_LessEqual || __kind == __TokenKind_Greater || __kind == __TokenKind_GreaterEqual {
					return ____rune_private_d1a61e34_skipComparePatternLookahead(__state, __index+1)
				}
				return func() int {
					if __kind == __TokenKind_LParen {
						return ____rune_private_d1a61e34_skipBalancedAt(__state, __index, __TokenKind_LParen, __TokenKind_RParen)
					}
					return func() int {
						if __kind == __TokenKind_LBracket {
							return ____rune_private_d1a61e34_skipBalancedAt(__state, __index, __TokenKind_LBracket, __TokenKind_RBracket)
						}
						return func() int {
							if __kind == __TokenKind_LBrace {
								return ____rune_private_d1a61e34_skipBalancedAt(__state, __index, __TokenKind_LBrace, __TokenKind_RBrace)
							}
							return -1
						}()
					}()
				}()
			}()
		}()
	}()
	return ____rune_private_d1a61e34_skipAliasPatternLookahead(__state, ____rune_private_d1a61e34_skipRangePatternLookahead(__state, __after))
}

func ____rune_private_d1a61e34_skipRangePatternLookahead(__state __ParserState, __index int) int {
	return func() int {
		if __index >= 0 && ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_DotDotEqual {
			return ____rune_private_d1a61e34_skipRangePatternEnd(__state, __index+1)
		}
		return func() int {
			if __index >= 0 && ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_DotDotLess {
				return ____rune_private_d1a61e34_skipRangePatternEnd(__state, __index+1)
			}
			return func() int {
				if __index >= 0 && ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_DotDot && ____rune_private_d1a61e34_parserKindAt(__state, __index+1) == __TokenKind_Less {
					return ____rune_private_d1a61e34_skipRangePatternEnd(__state, __index+2)
				}
				return __index
			}()
		}()
	}()
}

func ____rune_private_d1a61e34_skipRangePatternEnd(__state __ParserState, __index int) int {
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __index)
	return func() int {
		if __kind == __TokenKind_Underscore {
			return __index + 1
		}
		return func() int {
			if __kind == __TokenKind_Int || __kind == __TokenKind_Double || __kind == __TokenKind_BigInt || __kind == __TokenKind_String || __kind == __TokenKind_Char || __kind == __TokenKind_Ident {
				return ____rune_private_d1a61e34_skipIdentifierRangeEnd(__state, __index)
			}
			return -1
		}()
	}()
}

func ____rune_private_d1a61e34_skipIdentifierRangeEnd(__state __ParserState, __index int) int {
	return func() int {
		if ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_Ident && ____rune_private_d1a61e34_parserKindAt(__state, __index+1) == __TokenKind_Dot && ____rune_private_d1a61e34_parserKindAt(__state, __index+2) == __TokenKind_Ident {
			return __index + 3
		}
		return __index + 1
	}()
}

func ____rune_private_d1a61e34_skipIdentifierPatternLookahead(__state __ParserState, __index int) int {
	return func() int {
		if ____rune_private_d1a61e34_parserKindAt(__state, __index+1) == __TokenKind_LParen {
			return ____rune_private_d1a61e34_skipBalancedAt(__state, __index+1, __TokenKind_LParen, __TokenKind_RParen)
		}
		return func() int {
			if ____rune_private_d1a61e34_parserKindAt(__state, __index+1) == __TokenKind_Dot && ____rune_private_d1a61e34_parserKindAt(__state, __index+2) == __TokenKind_Ident {
				return __index + 3
			}
			return __index + 1
		}()
	}()
}

func ____rune_private_d1a61e34_skipComparePatternLookahead(__state __ParserState, __index int) int {
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __index)
	return func() int {
		if __kind == __TokenKind_Int || __kind == __TokenKind_Double || __kind == __TokenKind_BigInt || __kind == __TokenKind_String || __kind == __TokenKind_Char || __kind == __TokenKind_Ident {
			return __index + 1
		}
		return -1
	}()
}

func ____rune_private_d1a61e34_looksLikeMapLiteralBody(__state __ParserState) bool {
	__start := ____rune_private_d1a61e34_skipNewlinesAt(__state, __state.__current+1)
	__first := ____rune_private_d1a61e34_parserTokenAt(__state, __start)
	return func() bool {
		if __first.__kind == __TokenKind_EOF || __first.__kind == __TokenKind_RBrace {
			return false
		}
		return func() bool {
			if __first.__kind == __TokenKind_Ident && ____rune_private_d1a61e34_isLiteralIdentifier(__first.__lexeme) == false {
				return false
			}
			return ____rune_private_d1a61e34_scanMapLiteralColon(__state, __start, 0)
		}()
	}()
}

func ____rune_private_d1a61e34_scanMapLiteralColon(__state __ParserState, __index int, __depth int) bool {
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __index)
	return func() bool {
		if __kind == __TokenKind_EOF {
			return false
		}
		return func() bool {
			if __kind == __TokenKind_LParen || __kind == __TokenKind_LBracket || __kind == __TokenKind_LBrace {
				return ____rune_private_d1a61e34_scanMapLiteralColon(__state, __index+1, __depth+1)
			}
			return func() bool {
				if __kind == __TokenKind_RParen || __kind == __TokenKind_RBracket {
					return ____rune_private_d1a61e34_scanMapLiteralColon(__state, __index+1, func() int {
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
							return ____rune_private_d1a61e34_scanMapLiteralColon(__state, __index+1, __depth-1)
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
							return ____rune_private_d1a61e34_scanMapLiteralColon(__state, __index+1, __depth)
						}()
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_d1a61e34_isLiteralIdentifier(__name string) bool {
	return __name == "true" || __name == "false" || __name == "null"
}

func ____rune_private_d1a61e34_parserPatternLooksLikeObjectField(__state __ParserState) bool {
	__token := ____rune_private_d1a61e34_parserPeek(__state)
	return func() bool {
		switch {
		case __token.__kind == __TokenKind_Ident:
			return ____rune_private_d1a61e34_parserPatternIdentLooksLikeObjectField(__state, __token.__lexeme)
		default:
			return false
		}
	}()
}

func ____rune_private_d1a61e34_parserPatternIdentLooksLikeObjectField(__state __ParserState, __name string) bool {
	__spread := ____rune_private_d1a61e34_isPatternSpreadIdentifier(__name)
	return func() bool {
		switch {
		case __spread == true:
			return false
		default:
			return ____rune_private_d1a61e34_parserPatternObjectFieldTail(__state)
		}
	}()
}

func ____rune_private_d1a61e34_parserPatternObjectFieldTail(__state __ParserState) bool {
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __state.__current+1)
	return func() bool {
		switch {
		case __kind == __TokenKind_Colon:
			return true
		case __kind == __TokenKind_Question:
			return true
		case __kind == __TokenKind_Comma:
			return true
		case __kind == __TokenKind_RBrace:
			return true
		default:
			return false
		}
	}()
}

func ____rune_private_d1a61e34_isPatternSpreadIdentifier(__name string) bool {
	__empty := len([]rune(__name)) == 0
	return func() bool {
		switch {
		case __empty == true:
			return false
		default:
			return ____rune_private_d1a61e34_isUpperAsciiLetter([]rune(__name)[0])
		}
	}()
}

func ____rune_private_d1a61e34_isUpperAsciiLetter(__ch rune) bool {
	return __ch >= 'A' && __ch <= 'Z'
}

func ____rune_private_d1a61e34_skipNewlinesAt(__state __ParserState, __index int) int {
	return func() int {
		if ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_Newline {
			return ____rune_private_d1a61e34_skipNewlinesAt(__state, __index+1)
		}
		return __index
	}()
}

func ____rune_private_d1a61e34_skipGenericNamesAt(__state __ParserState, __index int) int {
	return func() int {
		if ____rune_private_d1a61e34_parserKindAt(__state, __index) == __TokenKind_LBracket {
			return ____rune_private_d1a61e34_skipBalancedAt(__state, __index, __TokenKind_LBracket, __TokenKind_RBracket)
		}
		return __index
	}()
}

func ____rune_private_d1a61e34_skipBalancedAt(__state __ParserState, __index int, __openKind __TokenKind, __closeKind __TokenKind) int {
	return ____rune_private_d1a61e34_skipBalancedAtLoop(__state, __index, __openKind, __closeKind, 0)
}

func ____rune_private_d1a61e34_skipBalancedAtLoop(__state __ParserState, __index int, __openKind __TokenKind, __closeKind __TokenKind, __depth int) int {
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __index)
	return func() int {
		if __kind == __TokenKind_EOF {
			return __index
		}
		return func() int {
			if __kind == __openKind {
				return ____rune_private_d1a61e34_skipBalancedAtLoop(__state, __index+1, __openKind, __closeKind, __depth+1)
			}
			return func() int {
				if __kind == __closeKind {
					return func() int {
						if __depth <= 1 {
							return __index + 1
						}
						return ____rune_private_d1a61e34_skipBalancedAtLoop(__state, __index+1, __openKind, __closeKind, __depth-1)
					}()
				}
				return ____rune_private_d1a61e34_skipBalancedAtLoop(__state, __index+1, __openKind, __closeKind, __depth)
			}()
		}()
	}()
}

func ____rune_private_d1a61e34_skipTypeNameTokensAt(__state __ParserState, __index int) int {
	return ____rune_private_d1a61e34_skipTypeNameTokensAtLoop(__state, __index, 0)
}

func ____rune_private_d1a61e34_skipTypeNameTokensAtLoop(__state __ParserState, __index int, __depth int) int {
	__kind := ____rune_private_d1a61e34_parserKindAt(__state, __index)
	return func() int {
		if __kind == __TokenKind_Ident || __kind == __TokenKind_Comma || __kind == __TokenKind_Colon || __kind == __TokenKind_Question || __kind == __TokenKind_Arrow || __kind == __TokenKind_At || __kind == __TokenKind_Dot {
			return ____rune_private_d1a61e34_skipTypeNameTokensAtLoop(__state, __index+1, __depth)
		}
		return func() int {
			if __kind == __TokenKind_LBracket || __kind == __TokenKind_LParen {
				return ____rune_private_d1a61e34_skipTypeNameTokensAtLoop(__state, __index+1, __depth+1)
			}
			return func() int {
				if __kind == __TokenKind_RBracket || __kind == __TokenKind_RParen {
					return func() int {
						if __depth == 0 {
							return __index
						}
						return ____rune_private_d1a61e34_skipTypeNameTokensAtLoop(__state, __index+1, __depth-1)
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

func __lowerWithSourcePath(__source string, __sourcePath string) __IRFile {
	return __withIRFileSourcePath(__lower(__source), __sourcePath)
}

func __lowerParsed(__file __ParsedFile) __IRFile {
	__out := __IRFile{__imports: []__IRImport{}, __tsImports: []__IRTSImport{}, __structs: []__IRStructType{}, __enums: []__IREnumType{}, __constants: []__IRConst{}, __functions: []__IRFunction{}, __tests: []__IRTest{}, __errors: __file.__errors}
	for _, __importDecl := range __file.__imports {
		_ = __importDecl
		func() int {
			__out.__imports = append(__out.__imports, ____rune_private_0495825c_lowerImport(__importDecl))
			return len(__out.__imports)
		}()
	}
	for _, __constant := range __file.__constants {
		_ = __constant
		func() int {
			__out.__constants = append(__out.__constants, ____rune_private_0495825c_lowerConst(__constant))
			return len(__out.__constants)
		}()
	}
	for _, __typeDecl := range __file.__types {
		_ = __typeDecl
		__out = ____rune_private_0495825c_lowerTypeInto(__out, __typeDecl)
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		func() int {
			__out.__functions = append(__out.__functions, ____rune_private_0495825c_lowerFunction(__fn))
			return len(__out.__functions)
		}()
	}
	for _, __testDecl := range __file.__tests {
		_ = __testDecl
		func() int {
			__out.__tests = append(__out.__tests, ____rune_private_0495825c_lowerTest(__testDecl))
			return len(__out.__tests)
		}()
	}
	return __out
}

func ____rune_private_0495825c_lowerTypeInto(__file __IRFile, __typeDecl __ParsedType) __IRFile {
	return func() __IRFile {
		if __typeDecl.__enum {
			return ____rune_private_0495825c_pushEnumType(__file, __typeDecl)
		}
		return ____rune_private_0495825c_pushStructType(__file, __typeDecl)
	}()
}

func ____rune_private_0495825c_pushStructType(__file __IRFile, __typeDecl __ParsedType) __IRFile {
	__file.__structs = append(__file.__structs, ____rune_private_0495825c_lowerStructType(__typeDecl))
	return __file
}

func ____rune_private_0495825c_pushEnumType(__file __IRFile, __typeDecl __ParsedType) __IRFile {
	__file.__enums = append(__file.__enums, ____rune_private_0495825c_lowerEnumType(__typeDecl))
	return __file
}

func __emptyIRExpr() __IRExpr {
	return __IRExpr{__kind: __ExprKind_Unknown, __text: "", __name: "", __value: "", __op: "", __params: []__IRParam{}, __children: []__IRExpr{}, __line: 0, __column: 0}
}

func __emptyIRFunction() __IRFunction {
	return __IRFunction{__name: "", __private: false, __static: false, __routine: false, __macro: false, __receiverType: "", __generics: []string{}, __params: []__IRParam{}, __returnType: "", __body: __emptyIRExpr(), __sourcePath: "", __line: 0, __column: 0}
}

func ____rune_private_0495825c_lowerImport(__importDecl __ParsedImport) __IRImport {
	return __IRImport{__path: __importDecl.__path, __go: __importDecl.__go, __module: __importDecl.__module, __line: __importDecl.__line, __column: __importDecl.__column}
}

func ____rune_private_0495825c_lowerConst(__constDecl __ParsedConst) __IRConst {
	return __IRConst{__name: __constDecl.__name, __private: __constDecl.__private, __typeName: __typeRefToString(__constDecl.__typeRef), __value: ____rune_private_0495825c_lowerExpr(__constDecl.__value), __line: __constDecl.__line, __column: __constDecl.__column}
}

func ____rune_private_0495825c_lowerParam(__param __ParsedParam) __IRParam {
	return __IRParam{__name: __param.__name, __typeName: __typeRefToString(__param.__typeRef), __line: __param.__line, __column: __param.__column}
}

func ____rune_private_0495825c_lowerField(__field __ParsedField) __IRField {
	return __IRField{__name: __field.__name, __private: __field.__private, __typeName: __typeRefToString(__field.__typeRef), __jsonName: ____rune_private_0495825c_lowerJsonFieldName(__field), __jsonIgnore: ____rune_private_0495825c_lowerHasAnnotation(__field.__annotations, "#", "json", "ignore", 0), __line: __field.__line, __column: __field.__column}
}

func ____rune_private_0495825c_lowerJsonFieldName(__field __ParsedField) string {
	return ____rune_private_0495825c_lowerJsonFieldNameAt(__field.__annotations, __field.__name, 0)
}

func ____rune_private_0495825c_lowerJsonFieldNameAt(__annotations []__ParsedAnnotation, __fallback string, __index int) string {
	__done := __index >= len(__annotations)
	return func() string {
		switch {
		case __done == true:
			return __fallback
		default:
			return ____rune_private_0495825c_lowerJsonFieldNameStep(__annotations, __fallback, __index)
		}
	}()
}

func ____rune_private_0495825c_lowerJsonFieldNameStep(__annotations []__ParsedAnnotation, __fallback string, __index int) string {
	__annotation := __annotations[__index]
	__matched := __annotation.__marker == "#" && __annotation.__module == "json" && __annotation.__name == "name" && len(__annotation.__args) > 0
	return func() string {
		switch {
		case __matched == true:
			return ____rune_private_0495825c_lowerAnnotationStringArg(__annotation, 0, __fallback)
		default:
			return ____rune_private_0495825c_lowerJsonFieldNameAt(__annotations, __fallback, __index+1)
		}
	}()
}

func ____rune_private_0495825c_lowerHasAnnotation(__annotations []__ParsedAnnotation, __marker string, __module string, __name string, __index int) bool {
	__done := __index >= len(__annotations)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_0495825c_lowerHasAnnotationStep(__annotations, __marker, __module, __name, __index)
		}
	}()
}

func ____rune_private_0495825c_lowerHasAnnotationStep(__annotations []__ParsedAnnotation, __marker string, __module string, __name string, __index int) bool {
	__annotation := __annotations[__index]
	__matched := __annotation.__marker == __marker && __annotation.__module == __module && __annotation.__name == __name
	return func() bool {
		switch {
		case __matched == true:
			return true
		default:
			return ____rune_private_0495825c_lowerHasAnnotation(__annotations, __marker, __module, __name, __index+1)
		}
	}()
}

func ____rune_private_0495825c_lowerAnnotationStringArg(__annotation __ParsedAnnotation, __index int, __fallback string) string {
	__valid := __index < len(__annotation.__args) && __annotation.__args[__index].__kind == __ExprKind_String
	return func() string {
		switch {
		case __valid == true:
			return ____rune_private_0495825c_lowerUnquoteString(__annotation.__args[__index].__value)
		default:
			return __fallback
		}
	}()
}

func ____rune_private_0495825c_lowerUnquoteString(__raw string) string {
	__quoted := len([]rune(__raw)) >= 2
	return func() string {
		switch {
		case __quoted == true:
			return func() string { runes := []rune(__raw); return string(runes[1 : len([]rune(__raw))-1]) }()
		default:
			return __raw
		}
	}()
}

func ____rune_private_0495825c_lowerEnumMember(__member __ParsedEnumMember) __IREnumMember {
	return __IREnumMember{__name: __member.__name, __private: __member.__private, __value: __member.__value, __params: ____rune_private_0495825c_lowerParams(__member.__params), __line: __member.__line, __column: __member.__column}
}

func ____rune_private_0495825c_lowerFunction(__fn __ParsedFunction) __IRFunction {
	return __IRFunction{__name: __fn.__name, __private: __fn.__private, __static: __fn.__static, __routine: __fn.__routine, __macro: ____rune_private_0495825c_parsedFunctionCompileTimeOnly(__fn), __receiverType: __fn.__receiverType, __generics: __fn.__generics, __params: ____rune_private_0495825c_lowerParams(__fn.__params), __returnType: __typeRefToString(__fn.__returnType), __body: ____rune_private_0495825c_lowerExpr(__fn.__body), __sourcePath: "", __line: __fn.__line, __column: __fn.__column}
}

func __withIRFileSourcePath(__file __IRFile, __sourcePath string) __IRFile {
	return __IRFile{__imports: __file.__imports, __tsImports: __file.__tsImports, __structs: ____rune_private_0495825c_withIRStructSourcePaths(__file.__structs, __sourcePath), __enums: ____rune_private_0495825c_withIREnumSourcePaths(__file.__enums, __sourcePath), __constants: __file.__constants, __functions: ____rune_private_0495825c_withIRFunctionSourcePaths(__file.__functions, __sourcePath), __tests: __file.__tests, __errors: __file.__errors}
}

func ____rune_private_0495825c_withIRFunctionSourcePath(__fn __IRFunction, __sourcePath string) __IRFunction {
	return __IRFunction{__name: __fn.__name, __private: __fn.__private, __static: __fn.__static, __routine: __fn.__routine, __macro: __fn.__macro, __receiverType: __fn.__receiverType, __generics: __fn.__generics, __params: __fn.__params, __returnType: __fn.__returnType, __body: __fn.__body, __sourcePath: __sourcePath, __line: __fn.__line, __column: __fn.__column}
}

func ____rune_private_0495825c_withIRFunctionSourcePaths(__functions []__IRFunction, __sourcePath string) []__IRFunction {
	__out := append([]__IRFunction{}, __functions[0:0]...)
	for _, __fn := range __functions {
		_ = __fn
		func() int {
			__out = append(__out, ____rune_private_0495825c_withIRFunctionSourcePath(__fn, __sourcePath))
			return len(__out)
		}()
	}
	return __out
}

func ____rune_private_0495825c_withIRStructSourcePath(__typeDecl __IRStructType, __sourcePath string) __IRStructType {
	return __IRStructType{__name: __typeDecl.__name, __private: __typeDecl.__private, __generics: __typeDecl.__generics, __fields: __typeDecl.__fields, __methods: ____rune_private_0495825c_withIRFunctionSourcePaths(__typeDecl.__methods, __sourcePath), __sourcePath: __sourcePath, __line: __typeDecl.__line, __column: __typeDecl.__column}
}

func ____rune_private_0495825c_withIRStructSourcePaths(__structs []__IRStructType, __sourcePath string) []__IRStructType {
	__out := append([]__IRStructType{}, __structs[0:0]...)
	for _, __typeDecl := range __structs {
		_ = __typeDecl
		func() int {
			__out = append(__out, ____rune_private_0495825c_withIRStructSourcePath(__typeDecl, __sourcePath))
			return len(__out)
		}()
	}
	return __out
}

func ____rune_private_0495825c_withIREnumSourcePath(__typeDecl __IREnumType, __sourcePath string) __IREnumType {
	return __IREnumType{__name: __typeDecl.__name, __private: __typeDecl.__private, __generics: __typeDecl.__generics, __members: __typeDecl.__members, __methods: ____rune_private_0495825c_withIRFunctionSourcePaths(__typeDecl.__methods, __sourcePath), __sourcePath: __sourcePath, __line: __typeDecl.__line, __column: __typeDecl.__column}
}

func ____rune_private_0495825c_withIREnumSourcePaths(__enums []__IREnumType, __sourcePath string) []__IREnumType {
	__out := append([]__IREnumType{}, __enums[0:0]...)
	for _, __typeDecl := range __enums {
		_ = __typeDecl
		func() int {
			__out = append(__out, ____rune_private_0495825c_withIREnumSourcePath(__typeDecl, __sourcePath))
			return len(__out)
		}()
	}
	return __out
}

func ____rune_private_0495825c_parsedFunctionCompileTimeOnly(__fn __ParsedFunction) bool {
	return __fn.__macro || ____rune_private_0495825c_typeRefIsSyntaxOnly(__fn.__returnType) || ____rune_private_0495825c_paramsUseSyntaxOnly(__fn.__params, 0)
}

func ____rune_private_0495825c_paramsUseSyntaxOnly(__params []__ParsedParam, __index int) bool {
	return func() bool {
		if __index >= len(__params) {
			return false
		}
		return ____rune_private_0495825c_typeRefIsSyntaxOnly(__params[__index].__typeRef) || ____rune_private_0495825c_paramsUseSyntaxOnly(__params, __index+1)
	}()
}

func ____rune_private_0495825c_typeRefIsSyntaxOnly(__typeRef __ParsedTypeRef) bool {
	__name := __typeRefToString(__typeRef)
	return strings.HasPrefix(__name, "Syntax") || __name == "MacroContext"
}

func ____rune_private_0495825c_lowerStructType(__typeDecl __ParsedType) __IRStructType {
	return __IRStructType{__name: __typeDecl.__name, __private: __typeDecl.__private, __generics: __typeDecl.__generics, __fields: ____rune_private_0495825c_lowerFields(__typeDecl.__fields), __methods: ____rune_private_0495825c_lowerFunctions(__typeDecl.__methods), __sourcePath: "", __line: __typeDecl.__line, __column: __typeDecl.__column}
}

func ____rune_private_0495825c_lowerEnumType(__typeDecl __ParsedType) __IREnumType {
	return __IREnumType{__name: __typeDecl.__name, __private: __typeDecl.__private, __generics: __typeDecl.__generics, __members: ____rune_private_0495825c_lowerEnumMembers(__typeDecl.__members), __methods: ____rune_private_0495825c_lowerFunctions(__typeDecl.__methods), __sourcePath: "", __line: __typeDecl.__line, __column: __typeDecl.__column}
}

func ____rune_private_0495825c_lowerTest(__testDecl __ParsedTest) __IRTest {
	return __IRTest{__name: __testDecl.__name, __body: ____rune_private_0495825c_lowerExpr(__testDecl.__body), __line: __testDecl.__line, __column: __testDecl.__column}
}

func ____rune_private_0495825c_lowerExpr(__expr __ParsedExpr) __IRExpr {
	__children := ____rune_private_0495825c_lowerExprs(__expr.__children)
	return __IRExpr{__kind: __expr.__kind, __text: ____rune_private_0495825c_inferIRExprText(__expr, __children), __name: __expr.__name, __value: __expr.__value, __op: __expr.__op, __params: ____rune_private_0495825c_lowerParams(__expr.__params), __children: __children, __line: __expr.__line, __column: __expr.__column}
}

func ____rune_private_0495825c_inferIRExprText(__expr __ParsedExpr, __children []__IRExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_0495825c_inferIRBinaryText(__expr, __children)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_0495825c_inferIRCallText(__children)
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

func ____rune_private_0495825c_inferIRBinaryText(__expr __ParsedExpr, __children []__IRExpr) string {
	return func() string {
		if __expr.__op == "??" {
			return ____rune_private_0495825c_inferIRCoalesceText(__children)
		}
		return ""
	}()
}

func ____rune_private_0495825c_inferIRCoalesceText(__children []__IRExpr) string {
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

func ____rune_private_0495825c_inferIRCallText(__children []__IRExpr) string {
	return func() string {
		if len(__children) == 0 {
			return ""
		}
		return func() string {
			if __children[0].__kind == __ExprKind_Selector {
				return ____rune_private_0495825c_inferIRSelectorCallText(__children[0], __children)
			}
			return ""
		}()
	}()
}

func ____rune_private_0495825c_inferIRSelectorCallText(__selector __IRExpr, __children []__IRExpr) string {
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

func ____rune_private_0495825c_lowerParams(__params []__ParsedParam) []__IRParam {
	__out := append([]__IRParam{}, []__IRParam{__IRParam{__name: "", __typeName: "", __line: 0, __column: 0}}[0:0]...)
	for _, __param := range __params {
		_ = __param
		func() int { __out = append(__out, ____rune_private_0495825c_lowerParam(__param)); return len(__out) }()
	}
	return __out
}

func ____rune_private_0495825c_lowerFields(__fields []__ParsedField) []__IRField {
	__out := append([]__IRField{}, []__IRField{__IRField{__name: "", __private: false, __typeName: "", __jsonName: "", __jsonIgnore: false, __line: 0, __column: 0}}[0:0]...)
	for _, __field := range __fields {
		_ = __field
		func() int { __out = append(__out, ____rune_private_0495825c_lowerField(__field)); return len(__out) }()
	}
	return __out
}

func ____rune_private_0495825c_lowerEnumMembers(__members []__ParsedEnumMember) []__IREnumMember {
	__out := append([]__IREnumMember{}, []__IREnumMember{__IREnumMember{__name: "", __private: false, __value: "", __params: []__IRParam{}, __line: 0, __column: 0}}[0:0]...)
	for _, __member := range __members {
		_ = __member
		func() int {
			__out = append(__out, ____rune_private_0495825c_lowerEnumMember(__member))
			return len(__out)
		}()
	}
	return __out
}

func ____rune_private_0495825c_lowerFunctions(__functions []__ParsedFunction) []__IRFunction {
	__out := append([]__IRFunction{}, []__IRFunction{__emptyIRFunction()}[0:0]...)
	for _, __fn := range __functions {
		_ = __fn
		func() int { __out = append(__out, ____rune_private_0495825c_lowerFunction(__fn)); return len(__out) }()
	}
	return __out
}

func ____rune_private_0495825c_lowerExprs(__exprs []__ParsedExpr) []__IRExpr {
	__out := append([]__IRExpr{}, []__IRExpr{__emptyIRExpr()}[0:0]...)
	for _, __expr := range __exprs {
		_ = __expr
		func() int { __out = append(__out, ____rune_private_0495825c_lowerExpr(__expr)); return len(__out) }()
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

func __compilerGoPackageImportPath(__spec string) string {
	return func() string {
		if strings.HasPrefix(__spec, "go:") {
			return func() string { runes := []rune(__spec); return string(runes[3:len([]rune(__spec))]) }()
		}
		return ""
	}()
}

func __compilerGoPackageName(__path string) string {
	__slash := strings.LastIndex(__path, "/")
	return func() string {
		if __slash >= 0 {
			return func() string { runes := []rune(__path); return string(runes[__slash+1 : len([]rune(__path))]) }()
		}
		return __path
	}()
}

func __compilerIRAtImportPath(__expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_At && __expr.__value != "" {
			return func() string {
				runes := []rune(__expr.__value)
				return string(runes[1 : len([]rune(__expr.__value))-1])
			}()
		}
		return ""
	}()
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
	return ____rune_private_6ab383c8_functionsUseUnwrap(__file.__functions, 0) || ____rune_private_6ab383c8_structsUseUnwrap(__file.__structs, 0) || ____rune_private_6ab383c8_enumsUseUnwrap(__file.__enums, 0) || ____rune_private_6ab383c8_testsUseUnwrap(__file.__tests, 0)
}

func ____rune_private_6ab383c8_functionsUseUnwrap(__functions []__IRFunction, __index int) bool {
	return func() bool {
		if __index >= len(__functions) {
			return false
		}
		return func() bool {
			if __functions[__index].__macro {
				return ____rune_private_6ab383c8_functionsUseUnwrap(__functions, __index+1)
			}
			return __exprUsesUnwrap(__functions[__index].__body) || ____rune_private_6ab383c8_functionsUseUnwrap(__functions, __index+1)
		}()
	}()
}

func ____rune_private_6ab383c8_testsUseUnwrap(__tests []__IRTest, __index int) bool {
	return func() bool {
		if __index >= len(__tests) {
			return false
		}
		return __exprUsesUnwrap(__tests[__index].__body) || ____rune_private_6ab383c8_testsUseUnwrap(__tests, __index+1)
	}()
}

func ____rune_private_6ab383c8_structsUseUnwrap(__structs []__IRStructType, __index int) bool {
	return func() bool {
		if __index >= len(__structs) {
			return false
		}
		return ____rune_private_6ab383c8_functionsUseUnwrap(__structs[__index].__methods, 0) || ____rune_private_6ab383c8_structsUseUnwrap(__structs, __index+1)
	}()
}

func ____rune_private_6ab383c8_enumsUseUnwrap(__enums []__IREnumType, __index int) bool {
	return func() bool {
		if __index >= len(__enums) {
			return false
		}
		return ____rune_private_6ab383c8_functionsUseUnwrap(__enums[__index].__methods, 0) || ____rune_private_6ab383c8_enumsUseUnwrap(__enums, __index+1)
	}()
}

func __exprUsesUnwrap(__expr __IRExpr) bool {
	return func() bool {
		if __expr.__kind == __ExprKind_Unwrap {
			return true
		}
		return ____rune_private_6ab383c8_exprChildrenUseUnwrap(__expr.__children, 0)
	}()
}

func __fileUsesModuleCall(__file __IRFile, __key string) bool {
	return ____rune_private_6ab383c8_functionsUseModuleCall(__file.__functions, __key, 0) || ____rune_private_6ab383c8_structsUseModuleCall(__file.__structs, __key, 0) || ____rune_private_6ab383c8_enumsUseModuleCall(__file.__enums, __key, 0) || ____rune_private_6ab383c8_testsUseModuleCall(__file.__tests, __key, 0)
}

func ____rune_private_6ab383c8_functionsUseModuleCall(__functions []__IRFunction, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__functions) {
			return false
		}
		return func() bool {
			if __functions[__index].__macro {
				return ____rune_private_6ab383c8_functionsUseModuleCall(__functions, __key, __index+1)
			}
			return ____rune_private_6ab383c8_exprUsesModuleCall(__functions[__index].__body, __key) || ____rune_private_6ab383c8_functionsUseModuleCall(__functions, __key, __index+1)
		}()
	}()
}

func ____rune_private_6ab383c8_testsUseModuleCall(__tests []__IRTest, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__tests) {
			return false
		}
		return ____rune_private_6ab383c8_exprUsesModuleCall(__tests[__index].__body, __key) || ____rune_private_6ab383c8_testsUseModuleCall(__tests, __key, __index+1)
	}()
}

func ____rune_private_6ab383c8_structsUseModuleCall(__structs []__IRStructType, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__structs) {
			return false
		}
		return ____rune_private_6ab383c8_functionsUseModuleCall(__structs[__index].__methods, __key, 0) || ____rune_private_6ab383c8_structsUseModuleCall(__structs, __key, __index+1)
	}()
}

func ____rune_private_6ab383c8_enumsUseModuleCall(__enums []__IREnumType, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__enums) {
			return false
		}
		return ____rune_private_6ab383c8_functionsUseModuleCall(__enums[__index].__methods, __key, 0) || ____rune_private_6ab383c8_enumsUseModuleCall(__enums, __key, __index+1)
	}()
}

func ____rune_private_6ab383c8_exprUsesModuleCall(__expr __IRExpr, __key string) bool {
	return func() bool {
		if __moduleCallKey(__expr) == __key {
			return true
		}
		return ____rune_private_6ab383c8_exprChildrenUseModuleCall(__expr.__children, __key, 0)
	}()
}

func ____rune_private_6ab383c8_exprChildrenUseModuleCall(__children []__IRExpr, __key string, __index int) bool {
	return func() bool {
		if __index >= len(__children) {
			return false
		}
		return ____rune_private_6ab383c8_exprUsesModuleCall(__children[__index], __key) || ____rune_private_6ab383c8_exprChildrenUseModuleCall(__children, __key, __index+1)
	}()
}

func ____rune_private_6ab383c8_exprChildrenUseUnwrap(__children []__IRExpr, __index int) bool {
	return func() bool {
		if __index >= len(__children) {
			return false
		}
		return __exprUsesUnwrap(__children[__index]) || ____rune_private_6ab383c8_exprChildrenUseUnwrap(__children, __index+1)
	}()
}

func __compilerIntToString(__value int) string {
	return func() string {
		if __value == 0 {
			return "0"
		}
		return func() string {
			if __value < 0 {
				return "-" + ____rune_private_6ab383c8_compilerUnsignedIntToString(0-__value, "")
			}
			return ____rune_private_6ab383c8_compilerUnsignedIntToString(__value, "")
		}()
	}()
}

func ____rune_private_6ab383c8_compilerUnsignedIntToString(__value int, __out string) string {
	return func() string {
		if __value <= 0 {
			return __out
		}
		return ____rune_private_6ab383c8_compilerUnsignedIntToString(__value/10, ____rune_private_6ab383c8_compilerDigitString(__value%10)+__out)
	}()
}

func ____rune_private_6ab383c8_compilerDigitString(__value int) string {
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
	__out = __out + ____rune_private_b4d281ed_emitGoImports(__file)
	if __fileUsesUnwrap(__file) {
		__out = __out + ____rune_private_b4d281ed_emitGoUnwrapHelper()
	}
	if __fileUsesPathFamily(__file) {
		__out = __out + ____rune_private_b4d281ed_emitGoPathHelpers()
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_b4d281ed_emitGoEnum(__enumDecl) + "\n"
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_b4d281ed_emitGoEnumMethods(__file, __enumDecl)
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_b4d281ed_emitGoStruct(__typeDecl) + "\n"
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_b4d281ed_emitGoMethods(__file, __typeDecl)
	}
	for _, __constant := range __file.__constants {
		_ = __constant
		__out = __out + ____rune_private_b4d281ed_emitGoConst(__file, __constant) + "\n"
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__out = func() string {
			if __fn.__macro {
				return __out
			}
			return __out + ____rune_private_b4d281ed_emitGoFunction(__file, __fn, "") + "\n"
		}()
	}
	if ____rune_private_b4d281ed_hasMain(__file) {
		__out = __out + "func main() {\n\t" + __mangleIdent("main") + "()\n}\n"
	}
	return __out
}

func ____rune_private_b4d281ed_emitGoConst(__file __IRFile, __constant __IRConst) string {
	return "var " + __mangleIdent(__constant.__name) + " " + ____rune_private_b4d281ed_goType(__constant.__typeName) + " = " + ____rune_private_b4d281ed_emitGoExprExpectedForFile(__file, __constant.__value, __constant.__typeName)
}

func ____rune_private_b4d281ed_emitGoImports(__file __IRFile) string {
	__imports := ____rune_private_b4d281ed_appendGoImportDecls([]string{}, __file.__imports, 0)
	__imports = ____rune_private_b4d281ed_appendGoImportIf(__imports, ____rune_private_b4d281ed_usesPrintFile(__file), "fmt")
	__imports = ____rune_private_b4d281ed_appendGoImportIf(__imports, ____rune_private_b4d281ed_fileUsesGoStrings(__file), "strings")
	__imports = ____rune_private_b4d281ed_appendGoImportIf(__imports, __fileUsesModuleCall(__file, "process.platform"), "runtime")
	__imports = ____rune_private_b4d281ed_appendGoImportIf(__imports, __fileUsesModuleCall(__file, "process.argv") || __fileUsesModuleCall(__file, "process.exit"), "os")
	__imports = ____rune_private_b4d281ed_appendGoImportIf(__imports, ____rune_private_b4d281ed_fileUsesDoubleMath(__file), "math")
	__imports = ____rune_private_b4d281ed_appendGoImportIf(__imports, __fileUsesModuleCall(__file, "int.toString") || __fileUsesModuleCall(__file, "bigint.toString"), "strconv")
	__imports = ____rune_private_b4d281ed_appendGoImportIf(__imports, __fileUsesUnwrap(__file), "reflect")
	__imports = ____rune_private_b4d281ed_appendGoImportIf(__imports, __fileUsesModuleCall(__file, "json.parse") || __fileUsesModuleCall(__file, "json.stringify"), "encoding/json")
	__empty := len(__imports) == 0
	return func() string {
		switch {
		case __empty == true:
			return ""
		default:
			return "import (\n" + ____rune_private_b4d281ed_emitGoImportLines(__imports, 0, "") + ")\n\n"
		}
	}()
}

func ____rune_private_b4d281ed_appendGoImportDecls(__imports []string, __importDecls []__IRImport, __index int) []string {
	__done := __index >= len(__importDecls)
	return func() []string {
		switch {
		case __done == true:
			return __imports
		default:
			return func() []string {
				__importDecl := __importDecls[__index]
				__next := func() []string {
					switch {
					case __importDecl.__go == true:
						return ____rune_private_b4d281ed_appendGoImport(__imports, __importDecl.__path)
					default:
						return __imports
					}
				}()
				return ____rune_private_b4d281ed_appendGoImportDecls(__next, __importDecls, __index+1)
			}()
		}
	}()
}

func ____rune_private_b4d281ed_appendGoImport(__imports []string, __path string) []string {
	__found := ____rune_private_b4d281ed_goImportContains(__imports, __path, 0)
	return func() []string {
		switch {
		case __found == true:
			return __imports
		default:
			return func() []string {
				out := []string{}
				out = append(out, __imports...)
				out = append(out, __path)
				return out
			}()
		}
	}()
}

func ____rune_private_b4d281ed_appendGoImportIf(__imports []string, __condition bool, __path string) []string {
	return func() []string {
		switch {
		case __condition == true:
			return ____rune_private_b4d281ed_appendGoImport(__imports, __path)
		default:
			return __imports
		}
	}()
}

func ____rune_private_b4d281ed_goImportContains(__imports []string, __path string, __index int) bool {
	__done := __index >= len(__imports)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return func() bool {
				__matched := __imports[__index] == __path
				return func() bool {
					switch {
					case __matched == true:
						return true
					default:
						return ____rune_private_b4d281ed_goImportContains(__imports, __path, __index+1)
					}
				}()
			}()
		}
	}()
}

func ____rune_private_b4d281ed_emitGoImportLines(__imports []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__imports) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoImportLines(__imports, __index+1, __out+"\t\""+__imports[__index]+"\"\n")
	}()
}

func ____rune_private_b4d281ed_emitGoUnwrapHelper() string {
	return "func __runeUnwrap(value any) any {\n\tv := reflect.ValueOf(value)\n\tif v.Kind() == reflect.Pointer {\n\t\tv = v.Elem()\n\t}\n\ttag := v.FieldByName(\"__tag\").Int()\n\tpayload := v.FieldByName(\"__payload\")\n\tif tag == 0 {\n\t\tif payload.Len() == 0 {\n\t\t\treturn nil\n\t\t}\n\t\treturn payload.Index(0).Interface()\n\t}\n\tif payload.Len() > 0 {\n\t\tpanic(payload.Index(0).Interface())\n\t}\n\tpanic(\"Result.Err\")\n}\n\n"
}

func ____rune_private_b4d281ed_emitGoTaskUnwrap(__expr __IRExpr) string {
	return "__runeUnwrap(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + ")"
}

func ____rune_private_b4d281ed_emitGoPathHelpers() string {
	return "func __runePathBasename(path string) string {\n\tindex := strings.LastIndex(path, \"/\")\n\tif index < 0 {\n\t\treturn path\n\t}\n\tif index == len(path)-1 {\n\t\treturn path\n\t}\n\treturn path[index+1:]\n}\n\nfunc __runePathExtname(path string) string {\n\tbase := __runePathBasename(path)\n\tindex := strings.LastIndex(base, \".\")\n\tif index <= 0 {\n\t\treturn \"\"\n\t}\n\treturn base[index:]\n}\n\nfunc __runePathDirname(path string) string {\n\tindex := strings.LastIndex(path, \"/\")\n\tif index < 0 {\n\t\treturn \".\"\n\t}\n\tif index == 0 {\n\t\treturn \"/\"\n\t}\n\treturn path[:index]\n}\n\nfunc __runePathJoin(parts []any) string {\n\treturn __runePathNormalize(__runePathJoinParts(__runePathStringParts(parts), 0, \"\"))\n}\n\nfunc __runePathNormalize(path string) string {\n\tabsolute := strings.HasPrefix(path, \"/\")\n\tout := __runePathNormalizeParts(strings.Split(path, \"/\"), 0, absolute, []string{})\n\tjoined := __runePathJoinParts(out, 0, \"\")\n\tif absolute {\n\t\treturn \"/\" + joined\n\t}\n\tif joined == \"\" {\n\t\treturn \".\"\n\t}\n\treturn joined\n}\n\nfunc __runePathResolve(parts []any) string {\n\tif len(parts) == 0 {\n\t\treturn \".\"\n\t}\n\treturn __runePathNormalize(__runePathJoin(parts))\n}\n\nfunc __runePathRelative(from string, to string) string {\n\tfromParts := __runePathParts(__runePathResolve([]any{from}))\n\ttoParts := __runePathParts(__runePathResolve([]any{to}))\n\tindex := 0\n\tfor index < len(fromParts) && index < len(toParts) && fromParts[index] == toParts[index] {\n\t\tindex++\n\t}\n\tout := \"\"\n\tfor i := index; i < len(fromParts); i++ {\n\t\tout = __runePathAppendPart(out, \"..\")\n\t}\n\tfor i := index; i < len(toParts); i++ {\n\t\tout = __runePathAppendPart(out, toParts[i])\n\t}\n\tif out == \"\" {\n\t\treturn \".\"\n\t}\n\treturn out\n}\n\nfunc __runePathStringParts(parts []any) []string {\n\tout := make([]string, 0, len(parts))\n\tfor _, part := range parts {\n\t\tout = append(out, part.(string))\n\t}\n\treturn out\n}\n\nfunc __runePathParts(path string) []string {\n\tclean := __runePathNormalize(path)\n\tout := []string{}\n\tfor _, part := range strings.Split(clean, \"/\") {\n\t\tif part != \"\" {\n\t\t\tout = append(out, part)\n\t\t}\n\t}\n\treturn out\n}\n\nfunc __runePathJoinParts(parts []string, index int, out string) string {\n\tfor index < len(parts) {\n\t\tout = __runePathAppendPart(out, parts[index])\n\t\tindex++\n\t}\n\treturn out\n}\n\nfunc __runePathAppendPart(out string, part string) string {\n\tif out == \"\" {\n\t\treturn part\n\t}\n\tif part == \"\" {\n\t\treturn out\n\t}\n\treturn out + \"/\" + part\n}\n\nfunc __runePathNormalizeParts(parts []string, index int, absolute bool, out []string) []string {\n\tfor index < len(parts) {\n\t\tpart := parts[index]\n\t\tif part == \"\" || part == \".\" {\n\t\t\tindex++\n\t\t\tcontinue\n\t\t}\n\t\tif part == \"..\" {\n\t\t\treturn __runePathNormalizeParent(parts, index, absolute, out)\n\t\t}\n\t\treturn __runePathNormalizePush(parts, index, absolute, out, part)\n\t}\n\treturn out\n}\n\nfunc __runePathNormalizeParent(parts []string, index int, absolute bool, out []string) []string {\n\tif len(out) > 0 {\n\t\treturn __runePathNormalizePop(parts, index, absolute, out)\n\t}\n\tif absolute {\n\t\treturn __runePathNormalizeParts(parts, index+1, absolute, out)\n\t}\n\treturn __runePathNormalizePush(parts, index, absolute, out, \"..\")\n}\n\nfunc __runePathNormalizePop(parts []string, index int, absolute bool, out []string) []string {\n\treturn __runePathNormalizeParts(parts, index+1, absolute, out[:len(out)-1])\n}\n\nfunc __runePathNormalizePush(parts []string, index int, absolute bool, out []string, part string) []string {\n\treturn __runePathNormalizeParts(parts, index+1, absolute, append(out, part))\n}\n\nfunc __runePathCollectParts(parts []string, index int, out []string) []string {\n\tfor index < len(parts) {\n\t\tif parts[index] != \"\" {\n\t\t\tout = append(out, parts[index])\n\t\t}\n\t\tindex++\n\t}\n\treturn out\n}\n\nfunc __runePathCollectPart(parts []string, index int, out []string) []string {\n\tif index < len(parts) {\n\t\tout = append(out, parts[index])\n\t}\n\treturn __runePathCollectParts(parts, index+1, out)\n}\n\nfunc __runePathRelativeFromParts(fromParts []string, toParts []string, index int) string {\n\tfor index < len(fromParts) && index < len(toParts) && fromParts[index] == toParts[index] {\n\t\tindex++\n\t}\n\treturn __runePathRelativeTail(fromParts, toParts, index, index, \"\")\n}\n\nfunc __runePathRelativeTail(fromParts []string, toParts []string, fromIndex int, toIndex int, out string) string {\n\tfor fromIndex < len(fromParts) {\n\t\tout = __runePathAppendPart(out, \"..\")\n\t\tfromIndex++\n\t}\n\tfor toIndex < len(toParts) {\n\t\tout = __runePathAppendPart(out, toParts[toIndex])\n\t\ttoIndex++\n\t}\n\tif out == \"\" {\n\t\treturn \".\"\n\t}\n\treturn out\n}\n\n"
}

func ____rune_private_b4d281ed_emitGoEnum(__enumDecl __IREnumType) string {
	return func() string {
		if ____rune_private_b4d281ed_enumHasPayload(__enumDecl.__members) {
			return ____rune_private_b4d281ed_emitGoPayloadEnum(__enumDecl)
		}
		return ____rune_private_b4d281ed_emitGoSimpleEnum(__enumDecl)
	}()
}

func ____rune_private_b4d281ed_emitGoSimpleEnum(__enumDecl __IREnumType) string {
	__out := "type " + __mangleIdent(__enumDecl.__name) + " int\n\n"
	__out = __out + "const (\n"
	__out = __out + ____rune_private_b4d281ed_emitGoEnumMembers(__enumDecl.__name, __enumDecl.__members, 0, "")
	return __out + ")\n"
}

func ____rune_private_b4d281ed_emitGoPayloadEnum(__enumDecl __IREnumType) string {
	__out := "type " + __mangleIdent(__enumDecl.__name) + ____rune_private_b4d281ed_emitGoGenericsDecl(__enumDecl.__generics) + " struct {\n"
	__out = __out + "\t__tag int\n"
	__out = __out + "\t__payload []any\n"
	__out = __out + "}\n\n"
	__out = __out + "const (\n"
	__out = __out + ____rune_private_b4d281ed_emitGoPayloadEnumTags(__enumDecl.__name, __enumDecl.__members, 0, "")
	__out = __out + ")\n\n"
	return __out + ____rune_private_b4d281ed_emitGoPayloadEnumConstructors(__enumDecl.__name, __enumDecl.__generics, __enumDecl.__members, 0, "")
}

func ____rune_private_b4d281ed_emitGoPayloadEnumTags(__enumName string, __members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoPayloadEnumTags(__enumName, __members, __index+1, __out+"\t"+__mangleIdent(__enumName+"_"+__members[__index].__name+"_tag")+" = "+__enumValue(__members[__index], __index)+"\n")
	}()
}

func ____rune_private_b4d281ed_emitGoPayloadEnumConstructors(__enumName string, __generics []string, __members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoPayloadEnumConstructors(__enumName, __generics, __members, __index+1, __out+____rune_private_b4d281ed_emitGoPayloadEnumConstructor(__enumName, __generics, __members[__index]))
	}()
}

func ____rune_private_b4d281ed_emitGoPayloadEnumConstructor(__enumName string, __generics []string, __member __IREnumMember) string {
	__tagName := __mangleIdent(__enumName + "_" + __member.__name + "_tag")
	__typeName := __mangleIdent(__enumName) + ____rune_private_b4d281ed_emitGoGenericsUse(__generics)
	return func() string {
		if len(__member.__params) == 0 {
			return "var " + __mangleIdent(__enumName+"_"+__member.__name) + " = " + __typeName + "{__tag: " + __tagName + ", __payload: nil}\n"
		}
		return "func " + __mangleIdent(__member.__name) + ____rune_private_b4d281ed_emitGoGenericsDecl(__generics) + "(" + ____rune_private_b4d281ed_emitGoParams(__member.__params, 0, "") + ") " + __typeName + " {\n\treturn " + __typeName + "{__tag: " + __tagName + ", __payload: []any{" + ____rune_private_b4d281ed_emitGoParamNames(__member.__params, 0, "") + "}}\n}\n"
	}()
}

func ____rune_private_b4d281ed_emitGoParamNames(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoParamNames(__params, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name))
	}()
}

func ____rune_private_b4d281ed_enumHasPayload(__members []__IREnumMember) bool {
	return ____rune_private_b4d281ed_enumHasPayloadAt(__members, 0)
}

func ____rune_private_b4d281ed_enumHasPayloadAt(__members []__IREnumMember, __index int) bool {
	return func() bool {
		if __index >= len(__members) {
			return false
		}
		return func() bool {
			if len(__members[__index].__params) > 0 {
				return true
			}
			return ____rune_private_b4d281ed_enumHasPayloadAt(__members, __index+1)
		}()
	}()
}

func ____rune_private_b4d281ed_emitGoEnumMembers(__enumName string, __members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoEnumMembers(__enumName, __members, __index+1, __out+"\t"+__mangleIdent(__enumName+"_"+__members[__index].__name)+" "+__mangleIdent(__enumName)+" = "+__enumValue(__members[__index], __index)+"\n")
	}()
}

func ____rune_private_b4d281ed_emitGoStruct(__typeDecl __IRStructType) string {
	__out := "type " + __mangleIdent(__typeDecl.__name) + ____rune_private_b4d281ed_emitGoGenericsDecl(__typeDecl.__generics) + " struct {\n"
	for _, __field := range __typeDecl.__fields {
		_ = __field
		__out = __out + "\t" + __mangleIdent(__field.__name) + " " + ____rune_private_b4d281ed_goType(__field.__typeName) + "\n"
	}
	return __out + "}\n"
}

func ____rune_private_b4d281ed_emitGoMethods(__file __IRFile, __typeDecl __IRStructType) string {
	__out := ""
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_b4d281ed_emitGoFunction(__file, __method, __typeDecl.__name) + "\n"
	}
	return __out
}

func ____rune_private_b4d281ed_emitGoEnumMethods(__file __IRFile, __enumDecl __IREnumType) string {
	__out := ""
	for _, __method := range __enumDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_b4d281ed_emitGoFunction(__file, __method, __enumDecl.__name) + "\n"
	}
	return __out
}

func ____rune_private_b4d281ed_emitGoFunction(__file __IRFile, __fn __IRFunction, __receiverType string) string {
	__returnType := ____rune_private_b4d281ed_inferredGoReturnType(__fn)
	__params := ____rune_private_b4d281ed_emitGoFunctionParams(__fn, 0, "")
	__ret := func() string {
		if __returnsValue(__returnType) {
			return " " + ____rune_private_b4d281ed_goType(__returnType)
		}
		return ""
	}()
	__effectiveReceiverType := func() string {
		switch {
		case __fn.__static == true:
			return ""
		case __fn.__static == false:
			return __receiverType
		}
		return ""
	}()
	__receiver := func() string {
		if __effectiveReceiverType == "" {
			return ""
		}
		return "(" + __mangleIdent("this") + " " + __mangleIdent(__effectiveReceiverType) + ") "
	}()
	__name := func() string {
		if __receiverType == "" {
			return __mangleIdent(__fn.__name)
		}
		return func() string {
			switch {
			case __fn.__static == true:
				return __mangleIdent(__receiverType + "_" + __fn.__name)
			case __fn.__static == false:
				return __mangleIdent(__fn.__name)
			}
			return ""
		}()
	}()
	__out := "func " + __receiver + __name + "(" + __params + ")" + __ret + " {\n"
	__out = __out + ____rune_private_b4d281ed_emitGoBody(__file, __fn.__body, __returnsValue(__returnType), __returnType, 1)
	return __out + "}\n"
}

func ____rune_private_b4d281ed_inferredGoReturnType(__fn __IRFunction) string {
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

func ____rune_private_b4d281ed_emitGoFunctionParams(__fn __IRFunction, __index int, __out string) string {
	return func() string {
		if __index >= len(__fn.__params) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoFunctionParams(__fn, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__fn.__params[__index].__name)+" "+____rune_private_b4d281ed_goType(____rune_private_b4d281ed_inferredGoParamType(__fn, __fn.__params[__index])))
	}()
}

func ____rune_private_b4d281ed_inferredGoParamType(__fn __IRFunction, __param __IRParam) string {
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

func ____rune_private_b4d281ed_emitGoBody(__file __IRFile, __expr __IRExpr, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_b4d281ed_emitGoBlock(__file, __expr.__children, 0, __returns, __returnType, __level, "")
		case __expr.__kind == __ExprKind_PatternBlock:
			return ____rune_private_b4d281ed_emitGoPatternBlock(__file, __expr, __returns, __returnType, __level)
		default:
			return __line(__level, func() string {
				if __returns {
					return "return " + ____rune_private_b4d281ed_emitGoExprExpectedForFile(__file, __expr, __returnType)
				}
				return ____rune_private_b4d281ed_emitGoExpr(__expr)
			}())
		}
	}()
}

func ____rune_private_b4d281ed_emitGoPatternBlock(__file __IRFile, __expr __IRExpr, __returns bool, __returnType string, __level int) string {
	return ____rune_private_b4d281ed_emitGoPatternBranches(__file, __expr.__children, 0, __returns, __returnType, __level, "")
}

func ____rune_private_b4d281ed_emitGoPatternBranches(__file __IRFile, __branches []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__branches) {
			return __out + __line(__level, "}") + func() string {
				if __returns {
					return __line(__level, "return "+____rune_private_b4d281ed_goZero(__returnType))
				}
				return ""
			}()
		}
		return ____rune_private_b4d281ed_emitGoPatternBranches(__file, __branches, __index+1, __returns, __returnType, __level, __out+____rune_private_b4d281ed_emitGoPatternBranch(__file, __branches[__index], __returns, __returnType, __level, __index == 0))
	}()
}

func ____rune_private_b4d281ed_emitGoPatternBranch(__file __IRFile, __branch __IRExpr, __returns bool, __returnType string, __level int, __first bool) string {
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
			return __line(__level, ____rune_private_b4d281ed_emitGoPatternPrefix(__first)+"else {") + ____rune_private_b4d281ed_emitGoPatternBranchBody(__file, __value, __returns, __returnType, __level+1)
		}
		return __line(__level, ____rune_private_b4d281ed_emitGoPatternPrefix(__first)+__head+____rune_private_b4d281ed_emitGoPatternCondition(__pattern)+" {") + ____rune_private_b4d281ed_emitGoPatternBranchBody(__file, __value, __returns, __returnType, __level+1)
	}()
}

func ____rune_private_b4d281ed_emitGoPatternPrefix(__first bool) string {
	return func() string {
		if __first {
			return ""
		}
		return "} "
	}()
}

func ____rune_private_b4d281ed_emitGoPatternBranchBody(__file __IRFile, __value __IRExpr, __returns bool, __returnType string, __level int) string {
	return func() string {
		if __returns {
			return __line(__level, "return "+____rune_private_b4d281ed_emitGoExprExpectedForFile(__file, __value, __returnType))
		}
		return __line(__level, ____rune_private_b4d281ed_emitGoExpr(__value))
	}()
}

func ____rune_private_b4d281ed_emitGoPatternCondition(__pattern __IRExpr) string {
	return "__n == " + __pattern.__text
}

func ____rune_private_b4d281ed_emitGoBlock(__file __IRFile, __statements []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__statements) {
			return func() string {
				if __returns && len(__statements) == 0 {
					return __out + __line(__level, "return "+____rune_private_b4d281ed_goZero(__returnType))
				}
				return __out
			}()
		}
		return ____rune_private_b4d281ed_emitGoBlock(__file, __statements, __index+1, __returns, __returnType, __level, __out+____rune_private_b4d281ed_emitGoStatement(__file, __statements[__index], __index == len(__statements)-1, __returns, __returnType, __level))
	}()
}

func ____rune_private_b4d281ed_emitGoStatement(__file __IRFile, __expr __IRExpr, __last bool, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Let:
			return ____rune_private_b4d281ed_emitGoLet(__file, __expr, __level)
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_b4d281ed_emitGoObjectDestructure(__expr, __level)
		default:
			return func() string {
				if __last && __returns {
					return __line(__level, "return "+____rune_private_b4d281ed_emitGoExprExpectedForFile(__file, __expr, __returnType))
				}
				return __line(__level, ____rune_private_b4d281ed_emitGoExpr(__expr))
			}()
		}
	}()
}

func ____rune_private_b4d281ed_emitGoLet(__file __IRFile, __expr __IRExpr, __level int) string {
	__value := ____rune_private_b4d281ed_emitGoLetValue(__file, __expr)
	__payload := ____rune_private_b4d281ed_unwrapPayloadType(__file, __expr.__children[0])
	if __payload != "" {
		__value = __value + ".(" + ____rune_private_b4d281ed_goType(__payload) + ")"
	}
	return __line(__level, __mangleIdent(__expr.__name)+" := "+__value) + __line(__level, "_ = "+__mangleIdent(__expr.__name))
}

func ____rune_private_b4d281ed_emitGoLetValue(__file __IRFile, __expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__value == "":
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0])
		default:
			return ____rune_private_b4d281ed_emitGoExprExpectedForFile(__file, __expr.__children[0], __expr.__value)
		}
	}()
}

func ____rune_private_b4d281ed_emitGoObjectDestructure(__expr __IRExpr, __level int) string {
	__source := ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0])
	__out := ""
	for _, __param := range __expr.__params {
		_ = __param
		__out = __out + __line(__level, __mangleIdent(__param.__name)+" := "+__source+"."+__mangleIdent(__param.__typeName)) + __line(__level, "_ = "+__mangleIdent(__param.__name))
	}
	return __out
}

func ____rune_private_b4d281ed_emitGoParams(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoParams(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name)+" "+____rune_private_b4d281ed_goType(__params[__index].__typeName))
	}()
}

func ____rune_private_b4d281ed_emitGoExpr(__expr __IRExpr) string {
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
			return __expr.__op + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Postfix:
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + __expr.__op
		case __expr.__kind == __ExprKind_CompileTime:
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Unwrap:
			return "__runeUnwrap(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + ")"
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_b4d281ed_emitGoBinary(__expr)
		case __expr.__kind == __ExprKind_Ternary:
			return ____rune_private_b4d281ed_emitGoTernary(__expr)
		case __expr.__kind == __ExprKind_Assign:
			return ____rune_private_b4d281ed_emitGoAssign(__expr)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_b4d281ed_emitGoCall(__expr)
		case __expr.__kind == __ExprKind_Lambda:
			return ____rune_private_b4d281ed_emitGoLambda(__expr)
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_b4d281ed_emitGoSelector(__expr)
		case __expr.__kind == __ExprKind_Index:
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "[" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "]"
		case __expr.__kind == __ExprKind_Array:
			return "[]any{" + ____rune_private_b4d281ed_emitGoExprList(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Tuple:
			return "[]any{" + ____rune_private_b4d281ed_emitGoExprList(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Map:
			return "map[any]any{" + ____rune_private_b4d281ed_emitGoMapEntries(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Spread:
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Reactive:
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Struct:
			return __mangleIdent(__expr.__name) + "{" + ____rune_private_b4d281ed_emitGoFields(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Object:
			return "struct{}{}"
		case __expr.__kind == __ExprKind_Block:
			return "func() any {\n" + ____rune_private_b4d281ed_emitGoBlockNoContext(__expr.__children, 0, true, "Dynamic", 1, "") + "}()"
		default:
			return "nil"
		}
	}()
}

func ____rune_private_b4d281ed_emitGoExprExpected(__expr __IRExpr, __expected string) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_b4d281ed_emitGoCallExpected(__expr, __expected)
		case __expr.__kind == __ExprKind_Object:
			return ____rune_private_b4d281ed_emitGoObjectExpected(__expr, __expected)
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_b4d281ed_emitGoBinaryExpected(__expr, __expected)
		default:
			return ____rune_private_b4d281ed_emitGoExpr(__expr)
		}
	}()
}

func ____rune_private_b4d281ed_emitGoExprExpectedForFile(__file __IRFile, __expr __IRExpr, __expected string) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_b4d281ed_emitGoCallExpectedForFile(__file, __expr, __expected)
		case __expr.__kind == __ExprKind_Object:
			return ____rune_private_b4d281ed_emitGoObjectExpected(__expr, __expected)
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_b4d281ed_emitGoBinaryExpectedForFile(__file, __expr, __expected)
		default:
			return ____rune_private_b4d281ed_emitGoExpr(__expr)
		}
	}()
}

func ____rune_private_b4d281ed_emitGoObjectExpected(__expr __IRExpr, __expected string) string {
	return func() string {
		switch {
		case __expected == "":
			return ____rune_private_b4d281ed_emitGoExpr(__expr)
		case __expected == "Dynamic":
			return ____rune_private_b4d281ed_emitGoExpr(__expr)
		default:
			return ____rune_private_b4d281ed_goType(__expected) + "{" + ____rune_private_b4d281ed_emitGoFields(__expr.__children, 0, "") + "}"
		}
	}()
}

func ____rune_private_b4d281ed_emitGoCallExpected(__expr __IRExpr, __expected string) string {
	__args := __genericInner(__expected, "Result")
	return func() string {
		if __args != "" && ____rune_private_b4d281ed_isResultConstructorCall(__expr) {
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "[" + ____rune_private_b4d281ed_emitGoTypeArgs(__args) + "](" + ____rune_private_b4d281ed_emitGoExprListFrom(__expr.__children, 1, "") + ")"
		}
		return ____rune_private_b4d281ed_emitGoCall(__expr)
	}()
}

func ____rune_private_b4d281ed_emitGoCallExpectedForFile(__file __IRFile, __expr __IRExpr, __expected string) string {
	return func() string {
		switch {
		case __moduleCallKey(__expr) == "json.parse":
			return ____rune_private_b4d281ed_emitGoJSONParse(__file, __expr, __expected)
		default:
			return ____rune_private_b4d281ed_emitGoCallExpected(__expr, __expected)
		}
	}()
}

func ____rune_private_b4d281ed_isResultConstructorCall(__expr __IRExpr) bool {
	return __expr.__kind == __ExprKind_Call && len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier && (__expr.__children[0].__name == "Ok" || __expr.__children[0].__name == "Err")
}

func ____rune_private_b4d281ed_emitGoTypeArgs(__args string) string {
	return ____rune_private_b4d281ed_emitGoTypeArgList(func() []string { parts := strings.Split(__args, ","); return parts }(), 0, "")
}

func ____rune_private_b4d281ed_emitGoTypeArgList(__args []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoTypeArgList(__args, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_b4d281ed_goType(strings.TrimSpace(__args[__index])))
	}()
}

func ____rune_private_b4d281ed_emitGoBlockNoContext(__statements []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__statements) {
			return func() string {
				if __returns && len(__statements) == 0 {
					return __out + __line(__level, "return "+____rune_private_b4d281ed_goZero(__returnType))
				}
				return __out
			}()
		}
		return ____rune_private_b4d281ed_emitGoBlockNoContext(__statements, __index+1, __returns, __returnType, __level, __out+____rune_private_b4d281ed_emitGoStatementNoContext(__statements[__index], __index == len(__statements)-1, __returns, __level))
	}()
}

func ____rune_private_b4d281ed_emitGoStatementNoContext(__expr __IRExpr, __last bool, __returns bool, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Let:
			return __line(__level, __mangleIdent(__expr.__name)+" := "+____rune_private_b4d281ed_emitGoExpr(__expr.__children[0])) + __line(__level, "_ = "+__mangleIdent(__expr.__name))
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_b4d281ed_emitGoObjectDestructure(__expr, __level)
		default:
			return func() string {
				if __last && __returns {
					return __line(__level, "return "+____rune_private_b4d281ed_emitGoExpr(__expr))
				}
				return __line(__level, ____rune_private_b4d281ed_emitGoExpr(__expr))
			}()
		}
	}()
}

func ____rune_private_b4d281ed_emitGoBinary(__expr __IRExpr) string {
	return func() string {
		if __expr.__op == "??" {
			return ____rune_private_b4d281ed_emitGoNullCoalesce(__expr, __expr.__text)
		}
		return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + " " + ____rune_private_b4d281ed_goBinaryOp(__expr.__op) + " " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1])
	}()
}

func ____rune_private_b4d281ed_emitGoBinaryExpected(__expr __IRExpr, __expected string) string {
	return func() string {
		if __expr.__op == "??" {
			return ____rune_private_b4d281ed_emitGoNullCoalesce(__expr, __expected)
		}
		return ____rune_private_b4d281ed_emitGoBinary(__expr)
	}()
}

func ____rune_private_b4d281ed_emitGoBinaryExpectedForFile(__file __IRFile, __expr __IRExpr, __expected string) string {
	return func() string {
		if __expr.__op == "??" {
			return ____rune_private_b4d281ed_emitGoNullCoalesceForFile(__file, __expr, __expected)
		}
		return ____rune_private_b4d281ed_emitGoBinary(__expr)
	}()
}

func ____rune_private_b4d281ed_emitGoNullCoalesce(__expr __IRExpr, __expected string) string {
	__resultType := ____rune_private_b4d281ed_goCoalesceResultType(__expr, __expected)
	return func() string {
		if len(__expr.__children) < 2 || __expr.__children[0].__kind == __ExprKind_Null {
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1])
		}
		return func() string {
			if __resultType == "" {
				return "func() any { __coalesce := " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "; if __coalesce != nil { return __coalesce }; return " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + " }()"
			}
			return "func() " + ____rune_private_b4d281ed_goType(__resultType) + " { __coalesce := " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "; if __coalesce != nil { return __coalesce.(" + ____rune_private_b4d281ed_goType(__resultType) + ") }; return " + ____rune_private_b4d281ed_emitGoExprExpected(__expr.__children[1], __resultType) + " }()"
		}()
	}()
}

func ____rune_private_b4d281ed_emitGoNullCoalesceForFile(__file __IRFile, __expr __IRExpr, __expected string) string {
	__resultType := ____rune_private_b4d281ed_goCoalesceResultType(__expr, __expected)
	return func() string {
		if len(__expr.__children) < 2 || __expr.__children[0].__kind == __ExprKind_Null {
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1])
		}
		return func() string {
			if __resultType == "" {
				return "func() any { __coalesce := " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "; if __coalesce != nil { return __coalesce }; return " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + " }()"
			}
			return "func() " + ____rune_private_b4d281ed_goType(__resultType) + " { __coalesce := " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "; if __coalesce != nil { return __coalesce.(" + ____rune_private_b4d281ed_goType(__resultType) + ") }; return " + ____rune_private_b4d281ed_emitGoExprExpectedForFile(__file, __expr.__children[1], __resultType) + " }()"
		}()
	}()
}

func ____rune_private_b4d281ed_goCoalesceResultType(__expr __IRExpr, __expected string) string {
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

func ____rune_private_b4d281ed_emitGoTernary(__expr __IRExpr) string {
	return "func() any { if " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + " { return " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + " }; return " + func() string {
		if len(__expr.__children) > 2 {
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2])
		}
		return "nil"
	}() + " }()"
}

func ____rune_private_b4d281ed_emitGoAssign(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) == 2 {
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + " = " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1])
		}
		return __mangleIdent(__expr.__name) + " = " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0])
	}()
}

func ____rune_private_b4d281ed_emitGoCall(__expr __IRExpr) string {
	return func() string {
		switch {
		case __moduleCallKey(__expr) == "io.println":
			return "fmt.Println(" + ____rune_private_b4d281ed_emitGoExprListFrom(__expr.__children, 1, "") + ")"
		case __moduleCallKey(__expr) == "json.stringify":
			return ____rune_private_b4d281ed_emitGoJSONStringify(__expr)
		case __moduleCallKey(__expr) == "json.parse":
			return ____rune_private_b4d281ed_emitGoJSONParseDynamic(__expr)
		case __moduleCallKey(__expr) == "map.new":
			return "map[any]any{}"
		case __moduleCallKey(__expr) == "set.new":
			return "map[any]struct{}{}"
		case __moduleCallKey(__expr) == "path.isAbsolute":
			return "strings.HasPrefix(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ", \"/\")"
		case __moduleCallKey(__expr) == "path.basename":
			return "__runePathBasename(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.extname":
			return "__runePathExtname(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.dirname":
			return "__runePathDirname(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.join":
			return "__runePathJoin(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.normalize":
			return "__runePathNormalize(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.resolve":
			return "__runePathResolve(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.relative":
			return "__runePathRelative(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.joinParts":
			return "__runePathJoinParts(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.appendPathPart":
			return "__runePathAppendPart(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.normalizeParts":
			return "__runePathNormalizeParts(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[4]) + "))"
		case __moduleCallKey(__expr) == "path.normalizePart":
			return "__runePathNormalizeParts(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[4]) + "))"
		case __moduleCallKey(__expr) == "path.normalizeParent":
			return "__runePathNormalizeParent(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[4]) + "))"
		case __moduleCallKey(__expr) == "path.normalizePop":
			return "__runePathNormalizePop(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[4]) + "))"
		case __moduleCallKey(__expr) == "path.normalizePush":
			return "__runePathNormalizePush(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + ", __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[4]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[5]) + ")"
		case __moduleCallKey(__expr) == "path.pathParts":
			return "__runePathParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.collectPathParts":
			return "__runePathCollectParts(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ", __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + "))"
		case __moduleCallKey(__expr) == "path.collectPathPart":
			return "__runePathCollectPart(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + ", __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + "))"
		case __moduleCallKey(__expr) == "path.relativeFromParts":
			return "__runePathRelativeFromParts(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.relativeTail":
			return "__runePathRelativeTail(__runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), __runePathStringParts(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + "), " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[3]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[4]) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[5]) + ")"
		case __moduleCallKey(__expr) == "process.platform":
			return "runtime.GOOS"
		case __moduleCallKey(__expr) == "process.cwd":
			return "\".\""
		case __moduleCallKey(__expr) == "process.env":
			return "(*string)(nil)"
		case __moduleCallKey(__expr) == "process.argv":
			return "append([]string(nil), os.Args[1:]...)"
		case __moduleCallKey(__expr) == "process.exit":
			return "func() struct{} { os.Exit(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "); return struct{}{} }()"
		case __moduleCallKey(__expr) == "int.toString":
			return "strconv.Itoa(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int.toDouble":
			return "float64(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int.toBigInt":
			return "int64(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int4.fromInt":
			return "func() int8 { n := (" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ") & 0xf; if n >= 8 { return int8(n - 16) }; return int8(n) }()"
		case __moduleCallKey(__expr) == "int8.fromInt":
			return "func() int8 { n := int(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "); return int8(n) }()"
		case __moduleCallKey(__expr) == "int16.fromInt":
			return "func() int16 { n := int(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "); return int16(n) }()"
		case __moduleCallKey(__expr) == "int64.fromInt":
			return "int64(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "uint.fromInt":
			return "uint(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "uint8.fromInt":
			return "func() uint8 { n := int(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "); return uint8(n) }()"
		case __moduleCallKey(__expr) == "uint16.fromInt":
			return "func() uint16 { n := int(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "); return uint16(n) }()"
		case __moduleCallKey(__expr) == "uint64.fromInt":
			return "uint64(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "float.fromDouble":
			return "float32(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case (__moduleCallKey(__expr) == "int4.toInt") || (__moduleCallKey(__expr) == "int8.toInt") || (__moduleCallKey(__expr) == "int16.toInt") || (__moduleCallKey(__expr) == "int64.toInt") || (__moduleCallKey(__expr) == "uint.toInt") || (__moduleCallKey(__expr) == "uint8.toInt") || (__moduleCallKey(__expr) == "uint16.toInt") || (__moduleCallKey(__expr) == "uint64.toInt"):
			return "int(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "float.toDouble":
			return "float64(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.fromInt":
			return "int64(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.toString":
			return "strconv.FormatInt(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ", 10)"
		case __moduleCallKey(__expr) == "bigint.toDouble":
			return "float64(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.trunc":
			return "int(math.Trunc(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "))"
		case __moduleCallKey(__expr) == "double.floor":
			return "int(math.Floor(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "))"
		case __moduleCallKey(__expr) == "double.ceil":
			return "int(math.Ceil(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "))"
		case __moduleCallKey(__expr) == "double.round":
			return "int(math.Round(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "))"
		default:
			return ____rune_private_b4d281ed_emitGoMaybeCoreMethodCall(__expr)
		}
	}()
}

func ____rune_private_b4d281ed_fileUsesDoubleMath(__file __IRFile) bool {
	return __fileUsesModuleCall(__file, "double.trunc") || __fileUsesModuleCall(__file, "double.floor") || __fileUsesModuleCall(__file, "double.ceil") || __fileUsesModuleCall(__file, "double.round")
}

func ____rune_private_b4d281ed_fileUsesGoStrings(__file __IRFile) bool {
	return __fileUsesModuleCall(__file, "path.isAbsolute") || __fileUsesPathFamily(__file)
}

func ____rune_private_b4d281ed_emitGoMaybeCoreMethodCall(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Selector {
			return ____rune_private_b4d281ed_emitGoCoreMethodCall(__expr, __expr.__children[0])
		}
		return ____rune_private_b4d281ed_emitGoDefaultCall(__expr)
	}()
}

func ____rune_private_b4d281ed_emitGoCoreMethodCall(__expr __IRExpr, __selector __IRExpr) string {
	return func() string {
		if len(__selector.__children) > 0 && __selector.__children[0].__kind != __ExprKind_At {
			return func() string {
				switch {
				case (__selector.__name == "length") || (__selector.__name == "byteLength"):
					return ____rune_private_b4d281ed_emitGoCoreLength(__selector.__children[0])
				case __selector.__name == "isEmpty":
					return "(" + ____rune_private_b4d281ed_emitGoCoreLength(__selector.__children[0]) + ") == 0"
				case __selector.__name == "at":
					return ____rune_private_b4d281ed_emitGoCoreAt(__expr, __selector.__children[0])
				case __selector.__name == "slice":
					return ____rune_private_b4d281ed_emitGoCoreSlice(__expr, __selector.__children[0])
				case __selector.__name == "push":
					return ____rune_private_b4d281ed_emitGoCorePush(__expr, __selector.__children[0])
				case __selector.__name == "set":
					return ____rune_private_b4d281ed_emitGoCoreSet(__expr, __selector.__children[0])
				case __selector.__name == "getOr":
					return ____rune_private_b4d281ed_emitGoCoreGetOr(__expr, __selector.__children[0])
				case (__selector.__name == "toDouble") || (__selector.__name == "toInt"):
					return ____rune_private_b4d281ed_emitGoCoreNumericConversion(__selector.__children[0], __selector.__name)
				default:
					return ____rune_private_b4d281ed_emitGoDefaultCall(__expr)
				}
			}()
		}
		return ____rune_private_b4d281ed_emitGoDefaultCall(__expr)
	}()
}

func ____rune_private_b4d281ed_emitGoCoreLength(__receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "len([]rune(" + ____rune_private_b4d281ed_emitGoExpr(__receiver) + "))"
		}
		return "len(" + ____rune_private_b4d281ed_emitGoExpr(__receiver) + ")"
	}()
}

func ____rune_private_b4d281ed_emitGoCoreAt(__expr __IRExpr, __receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "[]rune(" + ____rune_private_b4d281ed_emitGoExpr(__receiver) + ")[" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "]"
		}
		return ____rune_private_b4d281ed_emitGoExpr(__receiver) + "[" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "]"
	}()
}

func ____rune_private_b4d281ed_emitGoCoreSlice(__expr __IRExpr, __receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "func() string { runes := []rune(" + ____rune_private_b4d281ed_emitGoExpr(__receiver) + "); return string(runes[" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ":" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + "]) }()"
		}
		return ____rune_private_b4d281ed_emitGoExpr(__receiver) + "[" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + ":" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + "]"
	}()
}

func ____rune_private_b4d281ed_emitGoCorePush(__expr __IRExpr, __receiver __IRExpr) string {
	return "func() int { " + ____rune_private_b4d281ed_emitGoExpr(__receiver) + " = append(" + ____rune_private_b4d281ed_emitGoExpr(__receiver) + ", " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "); return len(" + ____rune_private_b4d281ed_emitGoExpr(__receiver) + ") }()"
}

func ____rune_private_b4d281ed_emitGoCoreSet(__expr __IRExpr, __receiver __IRExpr) string {
	return "func() any { " + ____rune_private_b4d281ed_emitGoExpr(__receiver) + "[" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "] = " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + "; return " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + " }()"
}

func ____rune_private_b4d281ed_emitGoCoreGetOr(__expr __IRExpr, __receiver __IRExpr) string {
	__resultType := ____rune_private_b4d281ed_goCoalesceResultType(__expr, __expr.__text)
	return func() string {
		if __resultType == "" {
			return "func() any { if __value, ok := " + ____rune_private_b4d281ed_emitGoExpr(__receiver) + "[" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "]; ok { return __value }; return " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[2]) + " }()"
		}
		return "func() " + ____rune_private_b4d281ed_goType(__resultType) + " { if __value, ok := " + ____rune_private_b4d281ed_emitGoExpr(__receiver) + "[" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "]; ok { return __value.(" + ____rune_private_b4d281ed_goType(__resultType) + ") }; return " + ____rune_private_b4d281ed_emitGoExprExpected(__expr.__children[2], __resultType) + " }()"
	}()
}

func ____rune_private_b4d281ed_emitGoDefaultCall(__expr __IRExpr) string {
	return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "(" + ____rune_private_b4d281ed_emitGoExprListFrom(__expr.__children, 1, "") + ")"
}

func ____rune_private_b4d281ed_emitGoCoreNumericConversion(__receiver __IRExpr, __method string) string {
	return func() string {
		if __method == "toDouble" {
			return "float64(" + ____rune_private_b4d281ed_emitGoExpr(__receiver) + ")"
		}
		return func() string {
			if __method == "toInt" {
				return "int(" + ____rune_private_b4d281ed_emitGoExpr(__receiver) + ")"
			}
			return ____rune_private_b4d281ed_emitGoExpr(__receiver)
		}()
	}()
}

func ____rune_private_b4d281ed_emitGoLambda(__expr __IRExpr) string {
	return "func(" + ____rune_private_b4d281ed_emitGoParams(__expr.__params, 0, "") + ") any { return " + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + " }"
}

func ____rune_private_b4d281ed_emitGoSelector(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__children[0].__kind == __ExprKind_At:
			return ____rune_private_b4d281ed_emitGoAtSelector(__expr)
		case __expr.__children[0].__kind == __ExprKind_Identifier:
			return func() string {
				if ____rune_private_b4d281ed_looksLikeTypeName(__expr.__children[0].__name) {
					return __mangleIdent(__expr.__children[0].__name + "_" + __expr.__name)
				}
				return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "." + __mangleIdent(__expr.__name)
			}()
		default:
			return ____rune_private_b4d281ed_emitGoExpr(__expr.__children[0]) + "." + __mangleIdent(__expr.__name)
		}
	}()
}

func ____rune_private_b4d281ed_emitGoAtSelector(__expr __IRExpr) string {
	__importPath := __compilerIRAtImportPath(__expr.__children[0])
	__goPath := __compilerGoPackageImportPath(__importPath)
	return func() string {
		switch {
		case __goPath == "":
			return func() string {
				__imported := __importPath != ""
				return func() string {
					switch {
					case __imported == true:
						return __mangleIdent(__expr.__name)
					default:
						return __expr.__children[0].__name + "." + __expr.__name
					}
				}()
			}()
		default:
			return __compilerGoPackageName(__goPath) + "." + __expr.__name
		}
	}()
}

func ____rune_private_b4d281ed_emitGoExprList(__exprs []__IRExpr, __index int, __out string) string {
	return ____rune_private_b4d281ed_emitGoExprListFrom(__exprs, __index, __out)
}

func ____rune_private_b4d281ed_emitGoExprListFrom(__exprs []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__exprs) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoExprListFrom(__exprs, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_b4d281ed_emitGoExpr(__exprs[__index]))
	}()
}

func ____rune_private_b4d281ed_emitGoMapEntries(__entries []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__entries) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoMapEntries(__entries, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_b4d281ed_emitGoExpr(__entries[__index].__children[0])+": "+____rune_private_b4d281ed_emitGoExpr(__entries[__index].__children[1]))
	}()
}

func ____rune_private_b4d281ed_emitGoFields(__fields []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__fields) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoFields(__fields, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+__mangleIdent(__fields[__index].__name)+": "+____rune_private_b4d281ed_emitGoStructFieldValue(__fields[__index].__children[0]))
	}()
}

func ____rune_private_b4d281ed_emitGoStructFieldValue(__expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_Array && len(__expr.__children) == 0 {
			return "nil"
		}
		return ____rune_private_b4d281ed_emitGoExpr(__expr)
	}()
}

func ____rune_private_b4d281ed_goBinaryOp(__op string) string {
	return __op
}

func ____rune_private_b4d281ed_hasMain(__file __IRFile) bool {
	return ____rune_private_b4d281ed_hasFunction(__file.__functions, "main", 0)
}

func ____rune_private_b4d281ed_hasFunction(__functions []__IRFunction, __name string, __index int) bool {
	return func() bool {
		if __index >= len(__functions) {
			return false
		}
		return func() bool {
			if __functions[__index].__macro == false && __functions[__index].__name == __name {
				return true
			}
			return ____rune_private_b4d281ed_hasFunction(__functions, __name, __index+1)
		}()
	}()
}

func ____rune_private_b4d281ed_usesPrintFile(__file __IRFile) bool {
	return ____rune_private_b4d281ed_functionsUsePrint(__file.__functions, 0) || ____rune_private_b4d281ed_structsUsePrint(__file.__structs, 0) || ____rune_private_b4d281ed_enumsUsePrint(__file.__enums, 0)
}

func ____rune_private_b4d281ed_functionsUsePrint(__functions []__IRFunction, __index int) bool {
	return func() bool {
		if __index >= len(__functions) {
			return false
		}
		return func() bool {
			if __functions[__index].__macro {
				return ____rune_private_b4d281ed_functionsUsePrint(__functions, __index+1)
			}
			return func() bool {
				if ____rune_private_b4d281ed_exprUsesPrint(__functions[__index].__body) {
					return true
				}
				return ____rune_private_b4d281ed_functionsUsePrint(__functions, __index+1)
			}()
		}()
	}()
}

func ____rune_private_b4d281ed_structsUsePrint(__structs []__IRStructType, __index int) bool {
	return func() bool {
		if __index >= len(__structs) {
			return false
		}
		return func() bool {
			if ____rune_private_b4d281ed_functionsUsePrint(__structs[__index].__methods, 0) {
				return true
			}
			return ____rune_private_b4d281ed_structsUsePrint(__structs, __index+1)
		}()
	}()
}

func ____rune_private_b4d281ed_enumsUsePrint(__enums []__IREnumType, __index int) bool {
	return func() bool {
		if __index >= len(__enums) {
			return false
		}
		return func() bool {
			if ____rune_private_b4d281ed_functionsUsePrint(__enums[__index].__methods, 0) {
				return true
			}
			return ____rune_private_b4d281ed_enumsUsePrint(__enums, __index+1)
		}()
	}()
}

func ____rune_private_b4d281ed_exprUsesPrint(__expr __IRExpr) bool {
	return func() bool {
		switch {
		case __moduleCallKey(__expr) == "io.println":
			return true
		default:
			return ____rune_private_b4d281ed_exprChildrenUsePrint(__expr.__children, 0)
		}
	}()
}

func ____rune_private_b4d281ed_exprChildrenUsePrint(__children []__IRExpr, __index int) bool {
	return func() bool {
		if __index >= len(__children) {
			return false
		}
		return func() bool {
			if ____rune_private_b4d281ed_exprUsesPrint(__children[__index]) {
				return true
			}
			return ____rune_private_b4d281ed_exprChildrenUsePrint(__children, __index+1)
		}()
	}()
}

func ____rune_private_b4d281ed_emitGoJSONStringify(__expr __IRExpr) string {
	return "func() string { __bytes, _ := json.Marshal(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "); return string(__bytes) }()"
}

func ____rune_private_b4d281ed_emitGoJSONParseDynamic(__expr __IRExpr) string {
	return "func() any { var __raw any; if err := json.Unmarshal([]byte(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), &__raw); err != nil { panic(err) }; return __raw }()"
}

func ____rune_private_b4d281ed_emitGoJSONParse(__file __IRFile, __expr __IRExpr, __expected string) string {
	__target := __expected
	func() {
		switch {
		case __target == "" == true:
			__target = "Dynamic"
			return
		default:
			__target = __target
			return
		}
	}()
	__dynamic := __target == "Dynamic"
	return func() string {
		switch {
		case __dynamic == true:
			return ____rune_private_b4d281ed_emitGoJSONParseDynamic(__expr)
		default:
			return ____rune_private_b4d281ed_emitGoJSONParseTyped(__file, __expr, __target)
		}
	}()
}

func ____rune_private_b4d281ed_emitGoJSONParseTyped(__file __IRFile, __expr __IRExpr, __target string) string {
	__decls := ____rune_private_b4d281ed_goJSONDeclsForType(__file, __target, []string{})
	__wireType := ____rune_private_b4d281ed_goJSONDecodeType(__file, __target)
	__value := ____rune_private_b4d281ed_goJSONDecodedValue(__file, "__raw", __target)
	return "func() " + ____rune_private_b4d281ed_goType(__target) + " { " + __decls.__text + "var __raw " + __wireType + "; if err := json.Unmarshal([]byte(" + ____rune_private_b4d281ed_emitGoExpr(__expr.__children[1]) + "), &__raw); err != nil { panic(err) }; return " + __value + " }()"
}

func ____rune_private_b4d281ed_emptyGoJSONDeclResult(__names []string, __text string) __GoJSONDeclResult {
	return __GoJSONDeclResult{__names: __names, __text: __text}
}

func ____rune_private_b4d281ed_goJSONDecodeType(__file __IRFile, __typeName string) string {
	__arrayInner := __genericInner(__typeName, "Array")
	return func() string {
		switch {
		case __arrayInner != "" == true:
			return "[]" + ____rune_private_b4d281ed_goJSONDecodeType(__file, __arrayInner)
		default:
			return ____rune_private_b4d281ed_goJSONDecodeNonArrayType(__file, __typeName)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDecodeNonArrayType(__file __IRFile, __typeName string) string {
	__structIndex := ____rune_private_b4d281ed_goJSONStructIndex(__file.__structs, __typeName, 0)
	return func() string {
		switch {
		case __structIndex >= 0 == true:
			return ____rune_private_b4d281ed_goJSONStructWireType(__file.__structs[__structIndex])
		default:
			return ____rune_private_b4d281ed_goJSONScalarDecodeType(__typeName)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONScalarDecodeType(__typeName string) string {
	return func() string {
		switch {
		case __typeName == "Char":
			return "string"
		default:
			return ____rune_private_b4d281ed_goType(__typeName)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONStructWireType(__typeDecl __IRStructType) string {
	return __mangleIdent("rune_json_" + __typeDecl.__name)
}

func ____rune_private_b4d281ed_goJSONDeclsForType(__file __IRFile, __typeName string, __names []string) __GoJSONDeclResult {
	__arrayInner := __genericInner(__typeName, "Array")
	return func() __GoJSONDeclResult {
		switch {
		case __arrayInner != "" == true:
			return ____rune_private_b4d281ed_goJSONDeclsForType(__file, __arrayInner, __names)
		default:
			return ____rune_private_b4d281ed_goJSONDeclsForNonArrayType(__file, __typeName, __names)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDeclsForNonArrayType(__file __IRFile, __typeName string, __names []string) __GoJSONDeclResult {
	__structIndex := ____rune_private_b4d281ed_goJSONStructIndex(__file.__structs, __typeName, 0)
	return func() __GoJSONDeclResult {
		switch {
		case __structIndex >= 0 == true:
			return ____rune_private_b4d281ed_goJSONDeclsForStruct(__file, __file.__structs[__structIndex], __names)
		default:
			return ____rune_private_b4d281ed_emptyGoJSONDeclResult(__names, "")
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDeclsForStruct(__file __IRFile, __typeDecl __IRStructType, __names []string) __GoJSONDeclResult {
	__name := ____rune_private_b4d281ed_goJSONStructWireType(__typeDecl)
	__seen := ____rune_private_b4d281ed_goJSONDeclSeen(__names, __name, 0)
	return func() __GoJSONDeclResult {
		switch {
		case __seen == true:
			return ____rune_private_b4d281ed_emptyGoJSONDeclResult(__names, "")
		default:
			return ____rune_private_b4d281ed_goJSONDeclsForNewStruct(__file, __typeDecl, func() []string {
				out := []string{}
				out = append(out, __names...)
				out = append(out, __name)
				return out
			}())
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDeclsForNewStruct(__file __IRFile, __typeDecl __IRStructType, __names []string) __GoJSONDeclResult {
	__nested := ____rune_private_b4d281ed_goJSONDeclsForFields(__file, __typeDecl.__fields, 0, __names, "")
	__fields := ____rune_private_b4d281ed_goJSONStructWireFields(__file, __typeDecl.__fields, 0, "")
	__decl := "type " + ____rune_private_b4d281ed_goJSONStructWireType(__typeDecl) + " struct{" + __fields + "}; "
	return ____rune_private_b4d281ed_emptyGoJSONDeclResult(__nested.__names, __nested.__text+__decl)
}

func ____rune_private_b4d281ed_goJSONDeclsForFields(__file __IRFile, __fields []__IRField, __index int, __names []string, __text string) __GoJSONDeclResult {
	__done := __index >= len(__fields)
	return func() __GoJSONDeclResult {
		switch {
		case __done == true:
			return ____rune_private_b4d281ed_emptyGoJSONDeclResult(__names, __text)
		default:
			return ____rune_private_b4d281ed_goJSONDeclsForField(__file, __fields, __index, __names, __text)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDeclsForField(__file __IRFile, __fields []__IRField, __index int, __names []string, __text string) __GoJSONDeclResult {
	__field := __fields[__index]
	__include := ____rune_private_b4d281ed_goJSONIncludeField(__field)
	return func() __GoJSONDeclResult {
		switch {
		case __include == true:
			return func() __GoJSONDeclResult {
				__nested := ____rune_private_b4d281ed_goJSONDeclsForType(__file, __field.__typeName, __names)
				return ____rune_private_b4d281ed_goJSONDeclsForFields(__file, __fields, __index+1, __nested.__names, __text+__nested.__text)
			}()
		default:
			return ____rune_private_b4d281ed_goJSONDeclsForFields(__file, __fields, __index+1, __names, __text)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONStructWireFields(__file __IRFile, __fields []__IRField, __index int, __out string) string {
	__done := __index >= len(__fields)
	return func() string {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_b4d281ed_goJSONStructWireField(__file, __fields, __index, __out)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONStructWireField(__file __IRFile, __fields []__IRField, __index int, __out string) string {
	__field := __fields[__index]
	__include := ____rune_private_b4d281ed_goJSONIncludeField(__field)
	return func() string {
		switch {
		case __include == true:
			return func() string {
				__prefix := func() string {
					switch {
					case __out == "" == true:
						return ""
					default:
						return "; "
					}
				}()
				__part := "F" + __compilerIntToString(__index) + " " + ____rune_private_b4d281ed_goJSONDecodeType(__file, __field.__typeName) + " " + ____rune_private_b4d281ed_goJSONTag(__field.__jsonName)
				return ____rune_private_b4d281ed_goJSONStructWireFields(__file, __fields, __index+1, __out+__prefix+__part)
			}()
		default:
			return ____rune_private_b4d281ed_goJSONStructWireFields(__file, __fields, __index+1, __out)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONTag(__name string) string {
	return "`json:\"" + strings.ReplaceAll((strings.ReplaceAll(__name, "\\", "\\\\")), "\"", "\\\"") + "\"`"
}

func ____rune_private_b4d281ed_goJSONIncludeField(__field __IRField) bool {
	__omit := __field.__jsonIgnore || ____rune_private_b4d281ed_goJSONOmitType(__field.__typeName)
	return func() bool {
		switch {
		case __omit == true:
			return false
		default:
			return true
		}
	}()
}

func ____rune_private_b4d281ed_goJSONOmitType(__typeName string) bool {
	return func() bool {
		switch {
		case (__typeName == "") || (__typeName == "Void") || (__typeName == "Symbol"):
			return true
		default:
			return false
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDeclSeen(__names []string, __name string, __index int) bool {
	__done := __index >= len(__names)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return func() bool {
				__matched := __names[__index] == __name
				return func() bool {
					switch {
					case __matched == true:
						return true
					default:
						return ____rune_private_b4d281ed_goJSONDeclSeen(__names, __name, __index+1)
					}
				}()
			}()
		}
	}()
}

func ____rune_private_b4d281ed_goJSONStructIndex(__structs []__IRStructType, __typeName string, __index int) int {
	__done := __index >= len(__structs)
	return func() int {
		switch {
		case __done == true:
			return -1
		default:
			return func() int {
				__matched := __structs[__index].__name == __typeName
				return func() int {
					switch {
					case __matched == true:
						return __index
					default:
						return ____rune_private_b4d281ed_goJSONStructIndex(__structs, __typeName, __index+1)
					}
				}()
			}()
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDecodedValue(__file __IRFile, __source string, __typeName string) string {
	__arrayInner := __genericInner(__typeName, "Array")
	return func() string {
		switch {
		case __arrayInner != "" == true:
			return ____rune_private_b4d281ed_goJSONDecodedArray(__file, __source, __typeName, __arrayInner)
		default:
			return ____rune_private_b4d281ed_goJSONDecodedNonArrayValue(__file, __source, __typeName)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDecodedNonArrayValue(__file __IRFile, __source string, __typeName string) string {
	__structIndex := ____rune_private_b4d281ed_goJSONStructIndex(__file.__structs, __typeName, 0)
	return func() string {
		switch {
		case __structIndex >= 0 == true:
			return ____rune_private_b4d281ed_goJSONDecodedStruct(__file, __source, __file.__structs[__structIndex])
		default:
			return ____rune_private_b4d281ed_goJSONDecodedScalar(__source, __typeName)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDecodedScalar(__source string, __typeName string) string {
	return func() string {
		switch {
		case __typeName == "Char":
			return "[]rune(" + __source + ")[0]"
		default:
			return __source
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDecodedArray(__file __IRFile, __source string, __typeName string, __inner string) string {
	return "func() " + ____rune_private_b4d281ed_goType(__typeName) + " { __out := make(" + ____rune_private_b4d281ed_goType(__typeName) + ", len(" + __source + "); for __idx, __item := range " + __source + " { __out[__idx] = " + ____rune_private_b4d281ed_goJSONDecodedValue(__file, "__item", __inner) + " }; return __out }()"
}

func ____rune_private_b4d281ed_goJSONDecodedStruct(__file __IRFile, __source string, __typeDecl __IRStructType) string {
	return "func() " + ____rune_private_b4d281ed_goType(__typeDecl.__name) + " { __out := " + ____rune_private_b4d281ed_goType(__typeDecl.__name) + "{}; " + ____rune_private_b4d281ed_goJSONDecodedFieldAssignments(__file, __source, __typeDecl.__fields, 0, "") + "return __out }()"
}

func ____rune_private_b4d281ed_goJSONDecodedFieldAssignments(__file __IRFile, __source string, __fields []__IRField, __index int, __out string) string {
	__done := __index >= len(__fields)
	return func() string {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_b4d281ed_goJSONDecodedFieldAssignment(__file, __source, __fields, __index, __out)
		}
	}()
}

func ____rune_private_b4d281ed_goJSONDecodedFieldAssignment(__file __IRFile, __source string, __fields []__IRField, __index int, __out string) string {
	__field := __fields[__index]
	__include := ____rune_private_b4d281ed_goJSONIncludeField(__field)
	return func() string {
		switch {
		case __include == true:
			return func() string {
				__sourceField := __source + ".F" + __compilerIntToString(__index)
				__value := ____rune_private_b4d281ed_goJSONDecodedValue(__file, __sourceField, __field.__typeName)
				__assignment := "__out." + __mangleIdent(__field.__name) + " = " + __value + "; "
				return ____rune_private_b4d281ed_goJSONDecodedFieldAssignments(__file, __source, __fields, __index+1, __out+__assignment)
			}()
		default:
			return ____rune_private_b4d281ed_goJSONDecodedFieldAssignments(__file, __source, __fields, __index+1, __out)
		}
	}()
}

func ____rune_private_b4d281ed_goType(__typeName string) string {
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
			return ____rune_private_b4d281ed_goTypeFallback(__typeName)
		}
	}()
}

func ____rune_private_b4d281ed_goTypeFallback(__typeName string) string {
	return func() string {
		if strings.HasSuffix(__typeName, "?") {
			return "any"
		}
		return func() string {
			if __genericInner(__typeName, "Array") != "" {
				return "[]" + ____rune_private_b4d281ed_goType(__genericInner(__typeName, "Array"))
			}
			return func() string {
				if __genericInner(__typeName, "ReadonlyArray") != "" {
					return "[]" + ____rune_private_b4d281ed_goType(__genericInner(__typeName, "ReadonlyArray"))
				}
				return func() string {
					if __genericInner(__typeName, "Map") != "" {
						return "map[" + ____rune_private_b4d281ed_goType(__typeArg(__genericInner(__typeName, "Map"), 0)) + "]" + ____rune_private_b4d281ed_goType(__typeArg(__genericInner(__typeName, "Map"), 1))
					}
					return func() string {
						if __genericInner(__typeName, "Set") != "" {
							return "map[" + ____rune_private_b4d281ed_goType(__genericInner(__typeName, "Set")) + "]struct{}"
						}
						return ____rune_private_b4d281ed_goNamedType(__typeName)
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_b4d281ed_goNamedType(__typeName string) string {
	__open := strings.Index(__typeName, "[")
	return func() string {
		if __open < 0 {
			return __mangleIdent(__typeName)
		}
		return __mangleIdent(func() string { runes := []rune(__typeName); return string(runes[0:__open]) }()) + "[" + ____rune_private_b4d281ed_emitGoTypeArgs(func() string { runes := []rune(__typeName); return string(runes[__open+1 : len([]rune(__typeName))-1]) }()) + "]"
	}()
}

func ____rune_private_b4d281ed_emitGoGenericsDecl(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "[" + ____rune_private_b4d281ed_emitGoGenericDeclItems(__generics, 0, "") + "]"
	}()
}

func ____rune_private_b4d281ed_emitGoGenericDeclItems(__generics []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__generics) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoGenericDeclItems(__generics, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__generics[__index])+" any")
	}()
}

func ____rune_private_b4d281ed_emitGoGenericsUse(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "[" + ____rune_private_b4d281ed_emitGoGenericUseItems(__generics, 0, "") + "]"
	}()
}

func ____rune_private_b4d281ed_emitGoGenericUseItems(__generics []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__generics) {
			return __out
		}
		return ____rune_private_b4d281ed_emitGoGenericUseItems(__generics, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__generics[__index]))
	}()
}

func ____rune_private_b4d281ed_unwrapPayloadType(__file __IRFile, __expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_Unwrap {
			return ____rune_private_b4d281ed_resultPayloadType(____rune_private_b4d281ed_unwrapSourceType(__file, __expr.__children[0]))
		}
		return ""
	}()
}

func ____rune_private_b4d281ed_unwrapSourceType(__file __IRFile, __expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_b4d281ed_callReturnType(__file, __expr)
		default:
			return ""
		}
	}()
}

func ____rune_private_b4d281ed_callReturnType(__file __IRFile, __expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier {
			return ____rune_private_b4d281ed_functionReturnType(__file.__functions, __expr.__children[0].__name, 0)
		}
		return ""
	}()
}

func ____rune_private_b4d281ed_functionReturnType(__functions []__IRFunction, __name string, __index int) string {
	return func() string {
		if __index >= len(__functions) {
			return ""
		}
		return func() string {
			if __functions[__index].__macro == false && __functions[__index].__name == __name {
				return __functions[__index].__returnType
			}
			return ____rune_private_b4d281ed_functionReturnType(__functions, __name, __index+1)
		}()
	}()
}

func ____rune_private_b4d281ed_resultPayloadType(__typeName string) string {
	__args := __genericInner(__typeName, "Result")
	return func() string {
		if __args == "" {
			return ""
		}
		return __typeArg(__args, 0)
	}()
}

func ____rune_private_b4d281ed_goZero(__typeName string) string {
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

func ____rune_private_b4d281ed_looksLikeTypeName(__name string) bool {
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
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitEnum(__enumDecl) + "\n"
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitEnumMethods(__enumDecl)
	}
	if __fileUsesUnwrap(__file) {
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitUnwrapHelper()
	}
	if __fileUsesPathFamily(__file) {
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitPathHelpers()
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitStruct(__typeDecl) + "\n"
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitMethods(__typeDecl)
	}
	for _, __constant := range __file.__constants {
		_ = __constant
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitConst(__constant) + "\n"
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__out = func() string {
			if __fn.__macro {
				return __out
			}
			return __out + ____rune_private_c9e1c9a2_emitMoonBitFunction(__fn, "") + "\n"
		}()
	}
	return __out
}

func ____rune_private_c9e1c9a2_emitMoonBitConst(__constant __IRConst) string {
	return "let " + ____rune_private_c9e1c9a2_moonBitValueIdent(__constant.__name) + " : " + ____rune_private_c9e1c9a2_moonBitType(__constant.__typeName) + " = " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__constant.__value)
}

func ____rune_private_c9e1c9a2_emitMoonBitUnwrapHelper() string {
	return "fn[T, E] rune_unwrap(value : RuneResult[T, E]) -> T {\n  match value {\n    RuneOk(value) => value\n    RuneErr(_) => abort(\"Result.Err\")\n  }\n}\n\n"
}

func ____rune_private_c9e1c9a2_emitMoonBitPathHelpers() string {
	return "fn __rune_path_basename(path : String) -> String {\n  let index = path.rev_find(\"/\").unwrap_or(-1)\n  if index < 0 {\n    path\n  } else if index == path.length() - 1 {\n    path\n  } else {\n    path[index + 1:].to_owned()\n  }\n}\n\nfn __rune_path_extname(path : String) -> String {\n  let base = __rune_path_basename(path)\n  let index = base.rev_find(\".\").unwrap_or(-1)\n  if index <= 0 { \"\" } else { base[index:].to_owned() }\n}\n\nfn __rune_path_dirname(path : String) -> String {\n  let index = path.rev_find(\"/\").unwrap_or(-1)\n  if index < 0 {\n    \".\"\n  } else if index == 0 {\n    \"/\"\n  } else {\n    path[:index].to_owned()\n  }\n}\n\nfn __rune_path_join(parts : Array[String]) -> String {\n  __rune_path_normalize(__rune_path_join_parts(parts, 0, \"\"))\n}\n\nfn __rune_path_normalize(path : String) -> String {\n  let absolute = path.has_prefix(\"/\")\n  let pieces = path.split(\"/\").map(fn(part) { part.to_owned() }).to_array()\n  let out = __rune_path_normalize_parts(pieces, 0, absolute, [])\n  let joined = __rune_path_join_parts(out, 0, \"\")\n  if absolute {\n    \"/\" + joined\n  } else if joined.is_empty() {\n    \".\"\n  } else {\n    joined\n  }\n}\n\nfn __rune_path_resolve(parts : Array[String]) -> String {\n  if parts.length() == 0 { \".\" } else { __rune_path_normalize(__rune_path_join(parts)) }\n}\n\nfn __rune_path_relative(from : String, to : String) -> String {\n  let from_parts = __rune_path_parts(__rune_path_resolve([from]))\n  let to_parts = __rune_path_parts(__rune_path_resolve([to]))\n  __rune_path_relative_from_parts(from_parts, to_parts, 0)\n}\n\nfn __rune_path_join_parts(parts : Array[String], index : Int, out : String) -> String {\n  if index >= parts.length() {\n    out\n  } else {\n    __rune_path_join_parts(parts, index + 1, __rune_path_append_part(out, parts[index]))\n  }\n}\n\nfn __rune_path_append_part(out : String, part : String) -> String {\n  if out.is_empty() {\n    part\n  } else if part.is_empty() {\n    out\n  } else {\n    out + \"/\" + part\n  }\n}\n\nfn __rune_path_normalize_parts(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  if index >= parts.length() {\n    out\n  } else {\n    let part = parts[index]\n    if part.is_empty() || part == \".\" {\n      __rune_path_normalize_parts(parts, index + 1, absolute, out)\n    } else if part == \"..\" {\n      __rune_path_normalize_parent(parts, index, absolute, out)\n    } else {\n      __rune_path_normalize_push(parts, index, absolute, out, part)\n    }\n  }\n}\n\nfn __rune_path_normalize_parent(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  if out.length() > 0 {\n    __rune_path_normalize_pop(parts, index, absolute, out)\n  } else if absolute {\n    __rune_path_normalize_parts(parts, index + 1, absolute, out)\n  } else {\n    __rune_path_normalize_push(parts, index, absolute, out, \"..\")\n  }\n}\n\nfn __rune_path_normalize_pop(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  __rune_path_normalize_parts(parts, index + 1, absolute, out[:out.length() - 1].to_owned())\n}\n\nfn __rune_path_normalize_push(parts : Array[String], index : Int, absolute : Bool, out : Array[String], part : String) -> Array[String] {\n  __rune_path_normalize_parts(parts, index + 1, absolute, [..out, part])\n}\n\nfn __rune_path_parts(path : String) -> Array[String] {\n  let pieces = __rune_path_normalize(path).split(\"/\").map(fn(part) { part.to_owned() }).to_array()\n  __rune_path_collect_parts(pieces, 0, [])\n}\n\nfn __rune_path_collect_parts(parts : Array[String], index : Int, out : Array[String]) -> Array[String] {\n  if index >= parts.length() {\n    out\n  } else if parts[index].is_empty() {\n    __rune_path_collect_parts(parts, index + 1, out)\n  } else {\n    __rune_path_collect_parts(parts, index + 1, [..out, parts[index]])\n  }\n}\n\nfn __rune_path_collect_part(parts : Array[String], index : Int, out : Array[String]) -> Array[String] {\n  if index < parts.length() {\n    __rune_path_collect_parts(parts, index + 1, [..out, parts[index]])\n  } else {\n    out\n  }\n}\n\nfn __rune_path_relative_from_parts(from_parts : Array[String], to_parts : Array[String], index : Int) -> String {\n  if index < from_parts.length() && index < to_parts.length() && from_parts[index] == to_parts[index] {\n    __rune_path_relative_from_parts(from_parts, to_parts, index + 1)\n  } else {\n    __rune_path_relative_tail(from_parts, to_parts, index, index, \"\")\n  }\n}\n\nfn __rune_path_relative_tail(from_parts : Array[String], to_parts : Array[String], from_index : Int, to_index : Int, out : String) -> String {\n  if from_index < from_parts.length() {\n    __rune_path_relative_tail(from_parts, to_parts, from_index + 1, to_index, __rune_path_append_part(out, \"..\"))\n  } else if to_index < to_parts.length() {\n    __rune_path_relative_tail(from_parts, to_parts, from_index, to_index + 1, __rune_path_append_part(out, to_parts[to_index]))\n  } else if out.is_empty() {\n    \".\"\n  } else {\n    out\n  }\n}\n\n"
}

func ____rune_private_c9e1c9a2_emitMoonBitEnum(__enumDecl __IREnumType) string {
	__out := "enum " + ____rune_private_c9e1c9a2_moonBitTypeIdent(__enumDecl.__name) + ____rune_private_c9e1c9a2_emitMoonBitGenerics(__enumDecl.__generics) + " {\n"
	__out = __out + ____rune_private_c9e1c9a2_emitMoonBitEnumMembers(__enumDecl.__members, 0, "")
	return __out + "} derive(Eq, Show)\n"
}

func ____rune_private_c9e1c9a2_emitMoonBitEnumMembers(__members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_c9e1c9a2_emitMoonBitEnumMembers(__members, __index+1, __out+____rune_private_c9e1c9a2_emitMoonBitEnumMember(__members[__index]))
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitEnumMember(__member __IREnumMember) string {
	return func() string {
		if len(__member.__params) == 0 {
			return __line(1, ____rune_private_c9e1c9a2_moonBitConstructorIdent(__member.__name))
		}
		return __line(1, ____rune_private_c9e1c9a2_moonBitConstructorIdent(__member.__name)+"("+____rune_private_c9e1c9a2_emitMoonBitParamTypes(__member.__params, 0, "")+")")
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitStruct(__typeDecl __IRStructType) string {
	__out := "struct " + ____rune_private_c9e1c9a2_moonBitTypeIdent(__typeDecl.__name) + " {\n"
	for _, __field := range __typeDecl.__fields {
		_ = __field
		__out = __out + __line(1, __mangleIdent(__field.__name)+" : "+____rune_private_c9e1c9a2_moonBitType(__field.__typeName))
	}
	return __out + "}\n"
}

func ____rune_private_c9e1c9a2_emitMoonBitMethods(__typeDecl __IRStructType) string {
	__out := ""
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitFunction(____rune_private_c9e1c9a2_methodWithMoonBitReceiver(__typeDecl.__name, __method), "") + "\n"
	}
	return __out
}

func ____rune_private_c9e1c9a2_emitMoonBitEnumMethods(__enumDecl __IREnumType) string {
	__out := ""
	for _, __method := range __enumDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_c9e1c9a2_emitMoonBitFunction(____rune_private_c9e1c9a2_methodWithMoonBitReceiver(__enumDecl.__name, __method), "") + "\n"
	}
	return __out
}

func ____rune_private_c9e1c9a2_methodWithMoonBitReceiver(__typeName string, __method __IRFunction) __IRFunction {
	return __IRFunction{__name: __typeName + "_" + __method.__name, __private: __method.__private, __static: __method.__static, __routine: __method.__routine, __macro: __method.__macro, __receiverType: __method.__receiverType, __generics: __method.__generics, __params: func() []__IRParam {
		switch {
		case __method.__static == true:
			return __method.__params
		case __method.__static == false:
			return ____rune_private_c9e1c9a2_prependMoonBitSelfParam(__typeName, __method.__params)
		}
		return nil
	}(), __returnType: __method.__returnType, __body: __method.__body, __sourcePath: __method.__sourcePath, __line: __method.__line, __column: __method.__column}
}

func ____rune_private_c9e1c9a2_prependMoonBitSelfParam(__typeName string, __params []__IRParam) []__IRParam {
	__out := []__IRParam{__IRParam{__name: "this", __typeName: __typeName, __line: 0, __column: 0}}
	for _, __param := range __params {
		_ = __param
		func() int { __out = append(__out, __param); return len(__out) }()
	}
	return __out
}

func ____rune_private_c9e1c9a2_emitMoonBitFunction(__fn __IRFunction, __receiverType string) string {
	__params := ____rune_private_c9e1c9a2_emitMoonBitParams(__fn.__params, 0, "")
	__ret := func() string {
		if __returnsValue(__fn.__returnType) {
			return " -> " + ____rune_private_c9e1c9a2_moonBitType(__fn.__returnType)
		}
		return ""
	}()
	__name := __mangleIdent(__fn.__name)
	__bodyReturns := __returnsValue(__fn.__returnType) && __fn.__name != "main"
	__head := func() string {
		if __fn.__name == "main" && __params == "" {
			return "fn main"
		}
		return "fn " + __name + ____rune_private_c9e1c9a2_emitMoonBitGenerics(__fn.__generics) + "(" + __params + ")" + __ret
	}()
	__out := __head + " {\n"
	__out = __out + ____rune_private_c9e1c9a2_emitMoonBitBody(__fn.__body, __bodyReturns, __fn.__returnType, 1)
	return __out + "}\n"
}

func ____rune_private_c9e1c9a2_emitMoonBitBody(__expr __IRExpr, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_c9e1c9a2_emitMoonBitBlock(__expr.__children, 0, __returns, __returnType, __level, "")
		default:
			return __line(__level, func() string {
				if __returns {
					return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr)
				}
				return ____rune_private_c9e1c9a2_emitMoonBitDiscard(__expr)
			}())
		}
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitBlock(__statements []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__statements) {
			return func() string {
				if __returns && len(__statements) == 0 {
					return __out + __line(__level, ____rune_private_c9e1c9a2_moonBitZero(__returnType))
				}
				return __out
			}()
		}
		return ____rune_private_c9e1c9a2_emitMoonBitBlock(__statements, __index+1, __returns, __returnType, __level, __out+____rune_private_c9e1c9a2_emitMoonBitStatement(__statements[__index], __index == len(__statements)-1, __returns, __level))
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitStatement(__expr __IRExpr, __last bool, __returns bool, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Let:
			return ____rune_private_c9e1c9a2_emitMoonBitLet(__expr, __level)
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_c9e1c9a2_emitMoonBitObjectDestructure(__expr, __level)
		default:
			return __line(__level, func() string {
				if __last && __returns {
					return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr)
				}
				return ____rune_private_c9e1c9a2_emitMoonBitDiscard(__expr)
			}())
		}
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitLet(__expr __IRExpr, __level int) string {
	return __line(__level, ____rune_private_c9e1c9a2_moonBitLetKeyword(__expr.__op)+__mangleIdent(__expr.__name)+" = "+____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]))
}

func ____rune_private_c9e1c9a2_moonBitLetKeyword(__op string) string {
	return func() string {
		switch {
		case __op == ":=:":
			return "let mut "
		default:
			return "let "
		}
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitObjectDestructure(__expr __IRExpr, __level int) string {
	__source := ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0])
	__tmp := __mangleIdent("__object")
	__out := __line(__level, "let "+__tmp+" = "+__source)
	for _, __param := range __expr.__params {
		_ = __param
		__out = __out + __line(__level, "let "+__mangleIdent(__param.__name)+" = "+__tmp+"."+__mangleIdent(__param.__typeName))
	}
	return __out
}

func ____rune_private_c9e1c9a2_emitMoonBitParams(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_c9e1c9a2_emitMoonBitParams(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name)+" : "+____rune_private_c9e1c9a2_moonBitType(__params[__index].__typeName))
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitParamTypes(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_c9e1c9a2_emitMoonBitParamTypes(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_c9e1c9a2_moonBitType(__params[__index].__typeName))
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitGenerics(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "[" + ____rune_private_c9e1c9a2_joinMoonBitGenericNames(__generics, 0, "") + "]"
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Identifier:
			return ____rune_private_c9e1c9a2_moonBitValueIdent(__expr.__name)
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
			return ____rune_private_c9e1c9a2_moonBitUnaryOp(__expr.__op) + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Postfix:
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + ____rune_private_c9e1c9a2_moonBitPostfixOp(__expr.__op)
		case __expr.__kind == __ExprKind_CompileTime:
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Unwrap:
			return "rune_unwrap(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + ")"
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_c9e1c9a2_emitMoonBitBinary(__expr)
		case __expr.__kind == __ExprKind_Ternary:
			return ____rune_private_c9e1c9a2_emitMoonBitTernary(__expr)
		case __expr.__kind == __ExprKind_Assign:
			return ____rune_private_c9e1c9a2_emitMoonBitAssign(__expr)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_c9e1c9a2_emitMoonBitCall(__expr)
		case __expr.__kind == __ExprKind_Lambda:
			return ____rune_private_c9e1c9a2_emitMoonBitLambda(__expr)
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_c9e1c9a2_emitMoonBitSelector(__expr)
		case __expr.__kind == __ExprKind_Index:
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + "[" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + "]"
		case __expr.__kind == __ExprKind_Array:
			return "[" + ____rune_private_c9e1c9a2_emitMoonBitExprList(__expr.__children, 0, "") + "]"
		case __expr.__kind == __ExprKind_Tuple:
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExprList(__expr.__children, 0, "") + ")"
		case __expr.__kind == __ExprKind_Map:
			return "{" + ____rune_private_c9e1c9a2_emitMoonBitMapEntries(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Spread:
			return ".." + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Reactive:
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Struct:
			return ____rune_private_c9e1c9a2_moonBitTypeIdent(__expr.__name) + "::{ " + ____rune_private_c9e1c9a2_emitMoonBitFields(__expr.__children, 0, "") + " }"
		case __expr.__kind == __ExprKind_Object:
			return "{ " + ____rune_private_c9e1c9a2_emitMoonBitFields(__expr.__children, 0, "") + " }"
		case __expr.__kind == __ExprKind_Block:
			return "{\n" + ____rune_private_c9e1c9a2_emitMoonBitBlock(__expr.__children, 0, true, "Dynamic", 1, "") + "}"
		default:
			return "()"
		}
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitDiscard(__expr __IRExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_Call {
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr)
		}
		return "ignore(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr) + ")"
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitBinary(__expr __IRExpr) string {
	return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + " " + ____rune_private_c9e1c9a2_moonBitBinaryOp(__expr.__op) + " " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1])
}

func ____rune_private_c9e1c9a2_emitMoonBitTernary(__expr __IRExpr) string {
	return "if " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + " { " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + " } else { " + func() string {
		if len(__expr.__children) > 2 {
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[2])
		}
		return "()"
	}() + " }"
}

func ____rune_private_c9e1c9a2_emitMoonBitAssign(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) == 2 {
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + " = " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1])
		}
		return __mangleIdent(__expr.__name) + " = " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0])
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitCall(__expr __IRExpr) string {
	return func() string {
		switch {
		case __moduleCallKey(__expr) == "io.println":
			return "println(" + ____rune_private_c9e1c9a2_emitMoonBitPrintArgs(__expr.__children, 1, "") + ")"
		case __moduleCallKey(__expr) == "io.print":
			return "print(" + ____rune_private_c9e1c9a2_emitMoonBitPrintArgs(__expr.__children, 1, "") + ")"
		case __moduleCallKey(__expr) == "map.new":
			return "{}"
		case __moduleCallKey(__expr) == "set.new":
			return "Set::new()"
		case __moduleCallKey(__expr) == "path.isAbsolute":
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ".has_prefix(\"/\")"
		case __moduleCallKey(__expr) == "path.basename":
			return "__rune_path_basename(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.extname":
			return "__rune_path_extname(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.dirname":
			return "__rune_path_dirname(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.join":
			return "__rune_path_join(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.normalize":
			return "__rune_path_normalize(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.resolve":
			return "__rune_path_resolve(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.relative":
			return "__rune_path_relative(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.joinParts":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall3("__rune_path_join_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3])
		case __moduleCallKey(__expr) == "path.appendPathPart":
			return "__rune_path_append_part(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.normalizeParts":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall4("__rune_path_normalize_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4])
		case __moduleCallKey(__expr) == "path.normalizePart":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall4("__rune_path_normalize_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4])
		case __moduleCallKey(__expr) == "path.normalizeParent":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall4("__rune_path_normalize_parent", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4])
		case __moduleCallKey(__expr) == "path.normalizePop":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall4("__rune_path_normalize_pop", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4])
		case __moduleCallKey(__expr) == "path.normalizePush":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall5("__rune_path_normalize_push", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4], __expr.__children[5])
		case __moduleCallKey(__expr) == "path.pathParts":
			return "__rune_path_parts(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.collectPathParts":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall3("__rune_path_collect_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3])
		case __moduleCallKey(__expr) == "path.collectPathPart":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall3("__rune_path_collect_part", __expr.__children[1], __expr.__children[2], __expr.__children[3])
		case __moduleCallKey(__expr) == "path.relativeFromParts":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall3("__rune_path_relative_from_parts", __expr.__children[1], __expr.__children[2], __expr.__children[3])
		case __moduleCallKey(__expr) == "path.relativeTail":
			return ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall5("__rune_path_relative_tail", __expr.__children[1], __expr.__children[2], __expr.__children[3], __expr.__children[4], __expr.__children[5])
		case __moduleCallKey(__expr) == "process.platform":
			return "\"moonbit\""
		case __moduleCallKey(__expr) == "process.cwd":
			return "\".\""
		case __moduleCallKey(__expr) == "process.env":
			return "(None : String?)"
		case __moduleCallKey(__expr) == "process.argv":
			return "([] : Array[String])"
		case __moduleCallKey(__expr) == "int.toString":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").to_string()"
		case __moduleCallKey(__expr) == "int.toDouble":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").to_double()"
		case __moduleCallKey(__expr) == "int.toBigInt":
			return "BigInt::from_int(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int4.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xf; if __n >= 8 { __n - 16 } else { __n } })(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int8.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xff; if __n >= 128 { __n - 256 } else { __n } })(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int16.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xffff; if __n >= 32768 { __n - 65536 } else { __n } })(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int64.fromInt":
			return "Int64::from_int(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case (__moduleCallKey(__expr) == "uint.fromInt") || (__moduleCallKey(__expr) == "uint8.fromInt") || (__moduleCallKey(__expr) == "uint16.fromInt"):
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").reinterpret_as_uint()"
		case __moduleCallKey(__expr) == "uint64.fromInt":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").to_uint64()"
		case (__moduleCallKey(__expr) == "int4.toInt") || (__moduleCallKey(__expr) == "int8.toInt") || (__moduleCallKey(__expr) == "int16.toInt"):
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1])
		case (__moduleCallKey(__expr) == "uint.toInt") || (__moduleCallKey(__expr) == "uint8.toInt") || (__moduleCallKey(__expr) == "uint16.toInt"):
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").reinterpret_as_int()"
		case (__moduleCallKey(__expr) == "int64.toInt") || (__moduleCallKey(__expr) == "uint64.toInt"):
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").to_int()"
		case __moduleCallKey(__expr) == "float.fromDouble":
			return "Float::from_double(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "float.toDouble":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").to_double()"
		case __moduleCallKey(__expr) == "bigint.fromInt":
			return "BigInt::from_int(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.toString":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").to_string()"
		case __moduleCallKey(__expr) == "bigint.toDouble":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").to_double()"
		case __moduleCallKey(__expr) == "double.trunc":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").trunc()"
		case __moduleCallKey(__expr) == "double.floor":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").floor()"
		case __moduleCallKey(__expr) == "double.ceil":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").ceil()"
		case __moduleCallKey(__expr) == "double.round":
			return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ").round()"
		default:
			return ____rune_private_c9e1c9a2_emitMoonBitMaybeCoreMethodCall(__expr)
		}
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall3(__name string, __first __IRExpr, __second __IRExpr, __third __IRExpr) string {
	return __name + "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__first) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__second) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__third) + ")"
}

func ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall4(__name string, __first __IRExpr, __second __IRExpr, __third __IRExpr, __fourth __IRExpr) string {
	return __name + "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__first) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__second) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__third) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__fourth) + ")"
}

func ____rune_private_c9e1c9a2_emitMoonBitRuntimeCall5(__name string, __first __IRExpr, __second __IRExpr, __third __IRExpr, __fourth __IRExpr, __fifth __IRExpr) string {
	return __name + "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__first) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__second) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__third) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__fourth) + ", " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__fifth) + ")"
}

func ____rune_private_c9e1c9a2_emitMoonBitMaybeCoreMethodCall(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Selector {
			return ____rune_private_c9e1c9a2_emitMoonBitCoreMethodCall(__expr, __expr.__children[0])
		}
		return ____rune_private_c9e1c9a2_emitMoonBitDefaultCall(__expr)
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitCoreMethodCall(__expr __IRExpr, __selector __IRExpr) string {
	return func() string {
		if len(__selector.__children) > 0 && __selector.__children[0].__kind != __ExprKind_At {
			return func() string {
				switch {
				case (__selector.__name == "length") || (__selector.__name == "byteLength"):
					return ____rune_private_c9e1c9a2_emitMoonBitExpr(__selector.__children[0]) + ".length()"
				case __selector.__name == "isEmpty":
					return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__selector.__children[0]) + ".length() == 0)"
				case __selector.__name == "at":
					return ____rune_private_c9e1c9a2_emitMoonBitExpr(__selector.__children[0]) + "[" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + "]"
				case __selector.__name == "slice":
					return ____rune_private_c9e1c9a2_emitMoonBitExpr(__selector.__children[0]) + "[" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[1]) + ":" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[2]) + "].to_owned()"
				default:
					return ____rune_private_c9e1c9a2_emitMoonBitDefaultCall(__expr)
				}
			}()
		}
		return ____rune_private_c9e1c9a2_emitMoonBitDefaultCall(__expr)
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitDefaultCall(__expr __IRExpr) string {
	return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + "(" + ____rune_private_c9e1c9a2_emitMoonBitExprListFrom(__expr.__children, 1, "") + ")"
}

func ____rune_private_c9e1c9a2_emitMoonBitPrintArgs(__exprs []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__exprs) {
			return func() string {
				if __out == "" {
					return "\"\""
				}
				return __out
			}()
		}
		return ____rune_private_c9e1c9a2_emitMoonBitPrintArgs(__exprs, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return " + \" \" + "
		}()+____rune_private_c9e1c9a2_emitMoonBitShowExpr(__exprs[__index]))
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitShowExpr(__expr __IRExpr) string {
	return func() string {
		if __expr.__text == "String" || __expr.__kind == __ExprKind_String || __expr.__kind == __ExprKind_Template {
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr)
		}
		return "(" + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr) + ").to_string()"
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitLambda(__expr __IRExpr) string {
	return "(" + ____rune_private_c9e1c9a2_emitMoonBitParams(__expr.__params, 0, "") + ") => " + ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0])
}

func ____rune_private_c9e1c9a2_emitMoonBitSelector(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__children[0].__kind == __ExprKind_At:
			return ____rune_private_c9e1c9a2_emitMoonBitAtSelector(__expr)
		case __expr.__children[0].__kind == __ExprKind_Identifier:
			return func() string {
				switch {
				case __expr.__op == "::":
					return __mangleIdent(__expr.__children[0].__name + "_" + __expr.__name)
				default:
					return func() string {
						if ____rune_private_c9e1c9a2_moonBitLooksLikeTypeName(__expr.__children[0].__name) {
							return ____rune_private_c9e1c9a2_moonBitTypeIdent(__expr.__children[0].__name) + "::" + ____rune_private_c9e1c9a2_moonBitConstructorIdent(__expr.__name)
						}
						return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + "." + __mangleIdent(__expr.__name)
					}()
				}
			}()
		default:
			return ____rune_private_c9e1c9a2_emitMoonBitExpr(__expr.__children[0]) + "." + __mangleIdent(__expr.__name)
		}
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitAtSelector(__expr __IRExpr) string {
	__imported := __expr.__children[0].__value != ""
	return func() string {
		switch {
		case __imported == true:
			return __mangleIdent(__expr.__name)
		default:
			return "@" + __expr.__children[0].__name + "." + __expr.__name
		}
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitExprList(__exprs []__IRExpr, __index int, __out string) string {
	return ____rune_private_c9e1c9a2_emitMoonBitExprListFrom(__exprs, __index, __out)
}

func ____rune_private_c9e1c9a2_emitMoonBitExprListFrom(__exprs []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__exprs) {
			return __out
		}
		return ____rune_private_c9e1c9a2_emitMoonBitExprListFrom(__exprs, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_c9e1c9a2_emitMoonBitExpr(__exprs[__index]))
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitMapEntries(__entries []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__entries) {
			return __out
		}
		return ____rune_private_c9e1c9a2_emitMoonBitMapEntries(__entries, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_c9e1c9a2_emitMoonBitExpr(__entries[__index].__children[0])+": "+____rune_private_c9e1c9a2_emitMoonBitExpr(__entries[__index].__children[1]))
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitFields(__fields []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__fields) {
			return __out
		}
		return ____rune_private_c9e1c9a2_emitMoonBitFields(__fields, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+__mangleIdent(__fields[__index].__name)+": "+____rune_private_c9e1c9a2_emitMoonBitExpr(__fields[__index].__children[0]))
	}()
}

func ____rune_private_c9e1c9a2_moonBitUnaryOp(__op string) string {
	return func() string {
		switch {
		case __op == "!":
			return "!"
		default:
			return __op
		}
	}()
}

func ____rune_private_c9e1c9a2_moonBitPostfixOp(__op string) string {
	return __op
}

func ____rune_private_c9e1c9a2_moonBitBinaryOp(__op string) string {
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

func ____rune_private_c9e1c9a2_moonBitType(__typeName string) string {
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
		return ____rune_private_c9e1c9a2_moonBitTypeFallback(__typeName)
	}
}

func ____rune_private_c9e1c9a2_moonBitTypeFallback(__typeName string) string {
	return func() string {
		if strings.HasSuffix(__typeName, "?") {
			return ____rune_private_c9e1c9a2_moonBitType(func() string { runes := []rune(__typeName); return string(runes[0 : len([]rune(__typeName))-1]) }()) + "?"
		}
		return func() string {
			if __genericInner(__typeName, "Array") != "" {
				return "Array[" + ____rune_private_c9e1c9a2_moonBitType(__genericInner(__typeName, "Array")) + "]"
			}
			return func() string {
				if __genericInner(__typeName, "ReadonlyArray") != "" {
					return "Array[" + ____rune_private_c9e1c9a2_moonBitType(__genericInner(__typeName, "ReadonlyArray")) + "]"
				}
				return func() string {
					if __genericInner(__typeName, "Map") != "" {
						return "Map[" + ____rune_private_c9e1c9a2_moonBitType(__typeArg(__genericInner(__typeName, "Map"), 0)) + ", " + ____rune_private_c9e1c9a2_moonBitType(__typeArg(__genericInner(__typeName, "Map"), 1)) + "]"
					}
					return func() string {
						if __genericInner(__typeName, "Set") != "" {
							return "Set[" + ____rune_private_c9e1c9a2_moonBitType(__genericInner(__typeName, "Set")) + "]"
						}
						return ____rune_private_c9e1c9a2_moonBitNamedType(__typeName)
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_c9e1c9a2_moonBitNamedType(__typeName string) string {
	__open := strings.Index(__typeName, "[")
	return func() string {
		if __open < 0 {
			return ____rune_private_c9e1c9a2_moonBitTypeIdent(__typeName)
		}
		return ____rune_private_c9e1c9a2_moonBitTypeIdent(func() string { runes := []rune(__typeName); return string(runes[0:__open]) }()) + "[" + ____rune_private_c9e1c9a2_emitMoonBitTypeArgs(func() string { runes := []rune(__typeName); return string(runes[__open+1 : len([]rune(__typeName))-1]) }()) + "]"
	}()
}

func ____rune_private_c9e1c9a2_emitMoonBitTypeArgs(__args string) string {
	return ____rune_private_c9e1c9a2_emitMoonBitTypeArgList(func() []string { parts := strings.Split(__args, ","); return parts }(), 0, "")
}

func ____rune_private_c9e1c9a2_emitMoonBitTypeArgList(__args []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_c9e1c9a2_emitMoonBitTypeArgList(__args, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_c9e1c9a2_moonBitType(strings.TrimSpace(__args[__index])))
	}()
}

func ____rune_private_c9e1c9a2_moonBitZero(__typeName string) string {
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

func ____rune_private_c9e1c9a2_joinMoonBitStrings(__values []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__values) {
			return __out
		}
		return ____rune_private_c9e1c9a2_joinMoonBitStrings(__values, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__values[__index])
	}()
}

func ____rune_private_c9e1c9a2_joinMoonBitGenericNames(__values []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__values) {
			return __out
		}
		return ____rune_private_c9e1c9a2_joinMoonBitGenericNames(__values, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_c9e1c9a2_moonBitTypeParamIdent(__values[__index]))
	}()
}

func ____rune_private_c9e1c9a2_moonBitValueIdent(__name string) string {
	return func() string {
		if ____rune_private_c9e1c9a2_moonBitLooksLikeTypeName(__name) {
			return ____rune_private_c9e1c9a2_moonBitConstructorIdent(__name)
		}
		return __mangleIdent(__name)
	}()
}

func ____rune_private_c9e1c9a2_moonBitTypeIdent(__name string) string {
	return func() string {
		if ____rune_private_c9e1c9a2_moonBitLooksGenericParam(__name) {
			return ____rune_private_c9e1c9a2_moonBitTypeParamIdent(__name)
		}
		return "Rune" + ____rune_private_c9e1c9a2_moonBitSanitizeIdent(__name)
	}()
}

func ____rune_private_c9e1c9a2_moonBitConstructorIdent(__name string) string {
	return "Rune" + ____rune_private_c9e1c9a2_moonBitSanitizeIdent(__name)
}

func ____rune_private_c9e1c9a2_moonBitTypeParamIdent(__name string) string {
	return func() string {
		if ____rune_private_c9e1c9a2_moonBitLooksLikeTypeName(__name) {
			return __name
		}
		return "T"
	}()
}

func ____rune_private_c9e1c9a2_moonBitLooksGenericParam(__name string) bool {
	return len([]rune(__name)) == 1 && ____rune_private_c9e1c9a2_moonBitLooksLikeTypeName(__name)
}

func ____rune_private_c9e1c9a2_moonBitSanitizeIdent(__name string) string {
	return strings.ReplaceAll((strings.ReplaceAll((strings.ReplaceAll(__name, ".", "_")), "-", "_")), "@", "_")
}

func ____rune_private_c9e1c9a2_moonBitLooksLikeTypeName(__name string) bool {
	return func() bool {
		if len([]rune(__name)) > 0 {
			return ____rune_private_c9e1c9a2_moonBitIsUpperLetter([]rune(__name)[0])
		}
		return false
	}()
}

func ____rune_private_c9e1c9a2_moonBitIsUpperLetter(__ch rune) bool {
	return __ch >= 'A' && __ch <= 'Z'
}

func __generateTypeScript(__file __IRFile) string {
	__out := ____rune_private_6bd43c78_emitTSImports(__file)
	if __fileUsesUnwrap(__file) {
		__out = __out + ____rune_private_6bd43c78_emitTSUnwrapHelper()
	}
	if __fileUsesPathFamily(__file) {
		__out = __out + ____rune_private_6bd43c78_emitTSPathHelpers()
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_6bd43c78_emitTSEnum(__enumDecl) + "\n"
	}
	for _, __enumDecl := range __file.__enums {
		_ = __enumDecl
		__out = __out + ____rune_private_6bd43c78_emitTSEnumMethods(__enumDecl)
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_6bd43c78_emitTSStruct(__typeDecl) + "\n"
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__out = __out + ____rune_private_6bd43c78_emitTSMethods(__typeDecl)
	}
	for _, __constant := range __file.__constants {
		_ = __constant
		__out = __out + ____rune_private_6bd43c78_emitTSConst(__constant) + "\n"
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__out = func() string {
			if __fn.__macro {
				return __out
			}
			return __out + ____rune_private_6bd43c78_emitTSFunction(__fn) + "\n"
		}()
	}
	return __out + ____rune_private_6bd43c78_emitTSExports(__file)
}

func ____rune_private_6bd43c78_emitTSImports(__file __IRFile) string {
	return ____rune_private_6bd43c78_emitTSImportList(__file.__tsImports, 0, "")
}

func ____rune_private_6bd43c78_emitTSConst(__constant __IRConst) string {
	return "const " + __mangleIdent(__constant.__name) + ": " + ____rune_private_6bd43c78_tsType(__constant.__typeName) + " = " + ____rune_private_6bd43c78_emitTSExprExpected(__constant.__value, __constant.__typeName) + ";"
}

func ____rune_private_6bd43c78_emitTSImportList(__imports []__IRTSImport, __index int, __out string) string {
	return func() string {
		if __index >= len(__imports) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSImportList(__imports, __index+1, __out+____rune_private_6bd43c78_emitTSImport(__imports[__index]))
	}()
}

func ____rune_private_6bd43c78_emitTSImport(__importDecl __IRTSImport) string {
	__names := ____rune_private_6bd43c78_emitTSImportNames(__importDecl.__functions, __importDecl.__values, 0, 0, "")
	return func() string {
		if __names == "" {
			return ""
		}
		return "import { " + __names + " } from " + ____rune_private_6bd43c78_tsQuoteString(____rune_private_6bd43c78_tsRuntimeSpecifier(__importDecl.__specifier)) + ";\n"
	}()
}

func ____rune_private_6bd43c78_emitTSImportNames(__functions []__IRFunction, __values []__IRConst, __fnIndex int, __valueIndex int, __out string) string {
	return func() string {
		if __fnIndex < len(__functions) {
			return ____rune_private_6bd43c78_emitTSImportNames(__functions, __values, __fnIndex+1, __valueIndex, ____rune_private_6bd43c78_appendTSImportName(__out, __functions[__fnIndex].__name))
		}
		return func() string {
			if __valueIndex < len(__values) {
				return ____rune_private_6bd43c78_emitTSImportNames(__functions, __values, __fnIndex, __valueIndex+1, ____rune_private_6bd43c78_appendTSImportName(__out, __values[__valueIndex].__name))
			}
			return __out
		}()
	}()
}

func ____rune_private_6bd43c78_appendTSImportName(__out string, __name string) string {
	return __out + func() string {
		if __out == "" {
			return ""
		}
		return ", "
	}() + __name + " as " + __mangleIdent(__name)
}

func ____rune_private_6bd43c78_tsQuoteString(__value string) string {
	return "\"" + strings.ReplaceAll((strings.ReplaceAll(__value, "\\", "\\\\")), "\"", "\\\"") + "\""
}

func ____rune_private_6bd43c78_tsRuntimeSpecifier(__specifier string) string {
	return func() string {
		if __specifier == "" || strings.HasPrefix(__specifier, "./") || strings.HasPrefix(__specifier, "../") || strings.HasPrefix(__specifier, "/") || strings.Contains(__specifier, "://") {
			return __specifier
		}
		return "./" + __specifier
	}()
}

func ____rune_private_6bd43c78_emitTSUnwrapHelper() string {
	return "function __runeUnwrap(value: any): any {\n  if (value && value.__tag === 0) return value.__payload?.[0];\n  if (value && value.__payload && value.__payload.length > 0) throw value.__payload[0];\n  throw new Error(\"Result.Err\");\n}\n\n"
}

func ____rune_private_6bd43c78_emitTSPathHelpers() string {
	return "function __runePathBasename(path: string): string {\n  const index = path.lastIndexOf(\"/\");\n  if (index < 0) return path;\n  if (index === path.length - 1) return path;\n  return path.slice(index + 1);\n}\n\nfunction __runePathExtname(path: string): string {\n  const base = __runePathBasename(path);\n  const index = base.lastIndexOf(\".\");\n  if (index <= 0) return \"\";\n  return base.slice(index);\n}\n\nfunction __runePathDirname(path: string): string {\n  const index = path.lastIndexOf(\"/\");\n  if (index < 0) return \".\";\n  if (index === 0) return \"/\";\n  return path.slice(0, index);\n}\n\nfunction __runePathJoin(parts: string[]): string {\n  return __runePathNormalize(__runePathJoinParts(parts, 0, \"\"));\n}\n\nfunction __runePathNormalize(path: string): string {\n  const absolute = path.startsWith(\"/\");\n  const out = __runePathNormalizeParts(path.split(\"/\"), 0, absolute, []);\n  const joined = __runePathJoinParts(out, 0, \"\");\n  if (absolute) return \"/\" + joined;\n  return joined === \"\" ? \".\" : joined;\n}\n\nfunction __runePathResolve(parts: string[]): string {\n  if (parts.length === 0) return \".\";\n  return __runePathNormalize(__runePathJoin(parts));\n}\n\nfunction __runePathRelative(from: string, to: string): string {\n  const fromParts = __runePathParts(__runePathResolve([from]));\n  const toParts = __runePathParts(__runePathResolve([to]));\n  let index = 0;\n  while (index < fromParts.length && index < toParts.length && fromParts[index] === toParts[index]) index++;\n  let out = \"\";\n  for (let i = index; i < fromParts.length; i++) out = __runePathAppendPart(out, \"..\");\n  for (let i = index; i < toParts.length; i++) out = __runePathAppendPart(out, toParts[i]);\n  return out === \"\" ? \".\" : out;\n}\n\nfunction __runePathParts(path: string): string[] {\n  return __runePathNormalize(path).split(\"/\").filter((part) => part !== \"\");\n}\n\nfunction __runePathJoinParts(parts: string[], index: number, out: string): string {\n  for (let i = index; i < parts.length; i++) out = __runePathAppendPart(out, parts[i]);\n  return out;\n}\n\nfunction __runePathAppendPart(out: string, part: string): string {\n  if (out === \"\") return part;\n  if (part === \"\") return out;\n  return out + \"/\" + part;\n}\n\nfunction __runePathNormalizeParts(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  while (index < parts.length) {\n    const part = parts[index];\n    if (part === \"\" || part === \".\") {\n      index++;\n      continue;\n    }\n    if (part === \"..\") return __runePathNormalizeParent(parts, index, absolute, out);\n    return __runePathNormalizePush(parts, index, absolute, out, part);\n  }\n  return out;\n}\n\nfunction __runePathNormalizeParent(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  if (out.length > 0) return __runePathNormalizePop(parts, index, absolute, out);\n  if (absolute) return __runePathNormalizeParts(parts, index + 1, absolute, out);\n  return __runePathNormalizePush(parts, index, absolute, out, \"..\");\n}\n\nfunction __runePathNormalizePop(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  return __runePathNormalizeParts(parts, index + 1, absolute, out.slice(0, out.length - 1));\n}\n\nfunction __runePathNormalizePush(parts: string[], index: number, absolute: boolean, out: string[], part: string): string[] {\n  return __runePathNormalizeParts(parts, index + 1, absolute, [...out, part]);\n}\n\nfunction __runePathCollectParts(parts: string[], index: number, out: string[]): string[] {\n  for (let i = index; i < parts.length; i++) {\n    if (parts[i] !== \"\") out.push(parts[i]);\n  }\n  return out;\n}\n\nfunction __runePathCollectPart(parts: string[], index: number, out: string[]): string[] {\n  if (index < parts.length) out.push(parts[index]);\n  return __runePathCollectParts(parts, index + 1, out);\n}\n\nfunction __runePathRelativeFromParts(fromParts: string[], toParts: string[], index: number): string {\n  while (index < fromParts.length && index < toParts.length && fromParts[index] === toParts[index]) index++;\n  return __runePathRelativeTail(fromParts, toParts, index, index, \"\");\n}\n\nfunction __runePathRelativeTail(fromParts: string[], toParts: string[], fromIndex: number, toIndex: number, out: string): string {\n  for (let i = fromIndex; i < fromParts.length; i++) out = __runePathAppendPart(out, \"..\");\n  for (let i = toIndex; i < toParts.length; i++) out = __runePathAppendPart(out, toParts[i]);\n  return out === \"\" ? \".\" : out;\n}\n\n"
}

func ____rune_private_6bd43c78_emitTSEnum(__enumDecl __IREnumType) string {
	return func() string {
		if ____rune_private_6bd43c78_tsEnumHasPayload(__enumDecl.__members, 0) {
			return ____rune_private_6bd43c78_emitTSPayloadEnum(__enumDecl)
		}
		return ____rune_private_6bd43c78_emitTSSimpleEnum(__enumDecl)
	}()
}

func ____rune_private_6bd43c78_emitTSSimpleEnum(__enumDecl __IREnumType) string {
	__out := "type " + __mangleIdent(__enumDecl.__name) + " = number;\n"
	__out = __out + "const " + __mangleIdent(__enumDecl.__name) + " = {\n"
	__out = __out + ____rune_private_6bd43c78_emitTSEnumMembers(__enumDecl.__members, 0, "")
	return __out + "} as const;\n"
}

func ____rune_private_6bd43c78_emitTSPayloadEnum(__enumDecl __IREnumType) string {
	__out := "type " + __mangleIdent(__enumDecl.__name) + ____rune_private_6bd43c78_emitTSGenerics(__enumDecl.__generics) + " =\n"
	__out = __out + ____rune_private_6bd43c78_emitTSPayloadEnumMembers(__enumDecl.__members, 0, "")
	__out = __out + ";\n\n"
	return __out + ____rune_private_6bd43c78_emitTSPayloadEnumConstructors(__enumDecl.__name, __enumDecl.__generics, __enumDecl.__members, 0, "")
}

func ____rune_private_6bd43c78_emitTSPayloadEnumMembers(__members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSPayloadEnumMembers(__members, __index+1, __out+____rune_private_6bd43c78_emitTSPayloadEnumMember(__members[__index], __index, __index == 0))
	}()
}

func ____rune_private_6bd43c78_emitTSPayloadEnumMember(__member __IREnumMember, __index int, __first bool) string {
	__prefix := func() string {
		if __first {
			return "  "
		}
		return "| "
	}()
	return __prefix + "{ __tag: " + __compilerIntToString(__index) + "; __payload: " + ____rune_private_6bd43c78_emitTSPayloadTuple(__member.__params, 0, "") + " }\n"
}

func ____rune_private_6bd43c78_emitTSPayloadTuple(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return "[" + __out + "]"
		}
		return ____rune_private_6bd43c78_emitTSPayloadTuple(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_6bd43c78_tsType(__params[__index].__typeName))
	}()
}

func ____rune_private_6bd43c78_emitTSPayloadEnumConstructors(__enumName string, __generics []string, __members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSPayloadEnumConstructors(__enumName, __generics, __members, __index+1, __out+____rune_private_6bd43c78_emitTSPayloadEnumConstructor(__enumName, __generics, __members[__index], __index))
	}()
}

func ____rune_private_6bd43c78_emitTSPayloadEnumConstructor(__enumName string, __generics []string, __member __IREnumMember, __index int) string {
	__typeName := __mangleIdent(__enumName) + ____rune_private_6bd43c78_emitTSGenericsUse(__generics)
	return func() string {
		if len(__member.__params) == 0 {
			return "const " + __mangleIdent(__enumName+"_"+__member.__name) + ": " + __typeName + " = { __tag: " + __compilerIntToString(__index) + ", __payload: [] };\n"
		}
		return "function " + __mangleIdent(__member.__name) + ____rune_private_6bd43c78_emitTSGenerics(__generics) + "(" + ____rune_private_6bd43c78_emitTSParams(__member.__params, 0, "") + "): " + __typeName + " {\n" + __line(1, "return { __tag: "+__compilerIntToString(__index)+", __payload: ["+____rune_private_6bd43c78_emitTSParamNames(__member.__params, 0, "")+"] };") + "}\n"
	}()
}

func ____rune_private_6bd43c78_emitTSParamNames(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSParamNames(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name))
	}()
}

func ____rune_private_6bd43c78_tsEnumHasPayload(__members []__IREnumMember, __index int) bool {
	return func() bool {
		if __index >= len(__members) {
			return false
		}
		return func() bool {
			if len(__members[__index].__params) > 0 {
				return true
			}
			return ____rune_private_6bd43c78_tsEnumHasPayload(__members, __index+1)
		}()
	}()
}

func ____rune_private_6bd43c78_emitTSGenericsUse(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "<" + ____rune_private_6bd43c78_emitTSGenericNames(__generics, 0, "") + ">"
	}()
}

func ____rune_private_6bd43c78_emitTSEnumMembers(__members []__IREnumMember, __index int, __out string) string {
	return func() string {
		if __index >= len(__members) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSEnumMembers(__members, __index+1, __out+__indent(1)+____rune_private_6bd43c78_tsPropertyName(__members[__index].__name)+": "+__enumValue(__members[__index], __index)+",\n")
	}()
}

func ____rune_private_6bd43c78_emitTSStruct(__typeDecl __IRStructType) string {
	__generics := ____rune_private_6bd43c78_emitTSGenerics(__typeDecl.__generics)
	__out := "type " + __mangleIdent(__typeDecl.__name) + __generics + " = {\n"
	for _, __field := range __typeDecl.__fields {
		_ = __field
		__out = __out + __indent(1) + ____rune_private_6bd43c78_tsPropertyName(__field.__name) + ": " + ____rune_private_6bd43c78_tsType(__field.__typeName) + ";\n"
	}
	return __out + "};\n"
}

func ____rune_private_6bd43c78_emitTSMethods(__typeDecl __IRStructType) string {
	__out := ""
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_6bd43c78_emitTSFunction(____rune_private_6bd43c78_methodWithTSReceiver(__typeDecl.__name, __method)) + "\n"
	}
	return __out
}

func ____rune_private_6bd43c78_emitTSEnumMethods(__enumDecl __IREnumType) string {
	__out := ""
	for _, __method := range __enumDecl.__methods {
		_ = __method
		__out = __out + ____rune_private_6bd43c78_emitTSFunction(____rune_private_6bd43c78_methodWithTSReceiver(__enumDecl.__name, __method)) + "\n"
	}
	return __out
}

func ____rune_private_6bd43c78_methodWithTSReceiver(__typeName string, __method __IRFunction) __IRFunction {
	return __IRFunction{__name: __typeName + "_" + __method.__name, __private: __method.__private, __static: __method.__static, __routine: __method.__routine, __macro: __method.__macro, __receiverType: __method.__receiverType, __generics: __method.__generics, __params: func() []__IRParam {
		switch {
		case __method.__static == true:
			return __method.__params
		case __method.__static == false:
			return ____rune_private_6bd43c78_prependThisParam(__typeName, __method.__params)
		}
		return nil
	}(), __returnType: __method.__returnType, __body: __method.__body, __sourcePath: __method.__sourcePath, __line: __method.__line, __column: __method.__column}
}

func ____rune_private_6bd43c78_prependThisParam(__typeName string, __params []__IRParam) []__IRParam {
	__out := []__IRParam{__IRParam{__name: "this", __typeName: __typeName, __line: 0, __column: 0}}
	for _, __param := range __params {
		_ = __param
		func() int { __out = append(__out, __param); return len(__out) }()
	}
	return __out
}

func ____rune_private_6bd43c78_emitTSFunction(__fn __IRFunction) string {
	__ret := ____rune_private_6bd43c78_tsType(__fn.__returnType)
	__out := "function " + __mangleIdent(__fn.__name) + ____rune_private_6bd43c78_emitTSGenerics(__fn.__generics) + "(" + ____rune_private_6bd43c78_emitTSParams(__fn.__params, 0, "") + "): " + __ret + " {\n"
	__out = __out + ____rune_private_6bd43c78_emitTSBody(__fn.__body, __returnsValue(__fn.__returnType), __fn.__returnType, 1)
	return __out + "}\n"
}

func ____rune_private_6bd43c78_emitTSBody(__expr __IRExpr, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_6bd43c78_emitTSBlock(__expr.__children, 0, __returns, __returnType, __level, "")
		default:
			return __line(__level, func() string {
				if __returns {
					return "return " + ____rune_private_6bd43c78_emitTSExprExpected(__expr, __returnType) + ";"
				}
				return ____rune_private_6bd43c78_emitTSExpr(__expr) + ";"
			}())
		}
	}()
}

func ____rune_private_6bd43c78_emitTSBlock(__statements []__IRExpr, __index int, __returns bool, __returnType string, __level int, __out string) string {
	return func() string {
		if __index >= len(__statements) {
			return func() string {
				if __returns && len(__statements) == 0 {
					return __out + __line(__level, "return undefined;")
				}
				return __out
			}()
		}
		return ____rune_private_6bd43c78_emitTSBlock(__statements, __index+1, __returns, __returnType, __level, __out+____rune_private_6bd43c78_emitTSStatement(__statements[__index], __index == len(__statements)-1, __returns, __returnType, __level))
	}()
}

func ____rune_private_6bd43c78_emitTSStatement(__expr __IRExpr, __last bool, __returns bool, __returnType string, __level int) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Let:
			return ____rune_private_6bd43c78_emitTSLet(__expr, __level)
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_6bd43c78_emitTSObjectDestructure(__expr, __level)
		default:
			return func() string {
				if __last && __returns {
					return __line(__level, "return "+____rune_private_6bd43c78_emitTSExprExpected(__expr, __returnType)+";")
				}
				return __line(__level, ____rune_private_6bd43c78_emitTSExpr(__expr)+";")
			}()
		}
	}()
}

func ____rune_private_6bd43c78_emitTSLet(__expr __IRExpr, __level int) string {
	return __line(__level, ____rune_private_6bd43c78_tsLetKeyword(__expr.__op)+__mangleIdent(__expr.__name)+" = "+____rune_private_6bd43c78_emitTSExpr(__expr.__children[0])+";")
}

func ____rune_private_6bd43c78_tsLetKeyword(__op string) string {
	return func() string {
		switch {
		case __op == ":=:":
			return "let "
		default:
			return "const "
		}
	}()
}

func ____rune_private_6bd43c78_emitTSObjectDestructure(__expr __IRExpr, __level int) string {
	return __line(__level, "const { "+____rune_private_6bd43c78_emitTSObjectDestructureFields(__expr.__params, 0, "")+" } = "+____rune_private_6bd43c78_emitTSExpr(__expr.__children[0])+";")
}

func ____rune_private_6bd43c78_emitTSObjectDestructureFields(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSObjectDestructureFields(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_6bd43c78_tsPropertyName(__params[__index].__typeName)+": "+__mangleIdent(__params[__index].__name))
	}()
}

func ____rune_private_6bd43c78_emitTSParams(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSParams(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name)+": "+____rune_private_6bd43c78_tsType(__params[__index].__typeName))
	}()
}

func ____rune_private_6bd43c78_emitTSGenerics(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "<" + ____rune_private_6bd43c78_emitTSGenericNames(__generics, 0, "") + ">"
	}()
}

func ____rune_private_6bd43c78_emitTSGenericNames(__generics []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__generics) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSGenericNames(__generics, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__generics[__index]))
	}()
}

func ____rune_private_6bd43c78_emitTSExports(__file __IRFile) string {
	__exports := ____rune_private_6bd43c78_emitTSConstExportNames(__file.__constants, 0, ____rune_private_6bd43c78_emitTSExportNames(__file.__functions, 0, ""))
	return func() string {
		if __exports == "" {
			return ""
		}
		return "export { " + __exports + " };\n"
	}()
}

func ____rune_private_6bd43c78_emitTSConstExportNames(__constants []__IRConst, __index int, __out string) string {
	return func() string {
		if __index >= len(__constants) {
			return __out
		}
		return func() string {
			if __constants[__index].__private {
				return ____rune_private_6bd43c78_emitTSConstExportNames(__constants, __index+1, __out)
			}
			return ____rune_private_6bd43c78_emitTSConstExportNames(__constants, __index+1, ____rune_private_6bd43c78_appendTSExportName(__out, __constants[__index].__name))
		}()
	}()
}

func ____rune_private_6bd43c78_emitTSExportNames(__functions []__IRFunction, __index int, __out string) string {
	return func() string {
		if __index >= len(__functions) {
			return __out
		}
		return func() string {
			if __functions[__index].__macro || __functions[__index].__private {
				return ____rune_private_6bd43c78_emitTSExportNames(__functions, __index+1, __out)
			}
			return ____rune_private_6bd43c78_emitTSExportNames(__functions, __index+1, ____rune_private_6bd43c78_appendTSExportName(__out, __functions[__index].__name))
		}()
	}()
}

func ____rune_private_6bd43c78_appendTSExportName(__out string, __name string) string {
	return __out + func() string {
		if __out == "" {
			return ""
		}
		return ", "
	}() + __mangleIdent(__name) + " as " + __name
}

func ____rune_private_6bd43c78_emitTSExpr(__expr __IRExpr) string {
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
			return __expr.__op + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Postfix:
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + __expr.__op
		case __expr.__kind == __ExprKind_CompileTime:
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Unwrap:
			return "__runeUnwrap(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + ")"
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_6bd43c78_emitTSBinary(__expr)
		case __expr.__kind == __ExprKind_Ternary:
			return ____rune_private_6bd43c78_emitTSTernary(__expr)
		case __expr.__kind == __ExprKind_Assign:
			return ____rune_private_6bd43c78_emitTSAssign(__expr)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_6bd43c78_emitTSCall(__expr)
		case __expr.__kind == __ExprKind_Lambda:
			return ____rune_private_6bd43c78_emitTSLambda(__expr)
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_6bd43c78_emitTSSelector(__expr)
		case __expr.__kind == __ExprKind_Index:
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + "[" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + "]"
		case __expr.__kind == __ExprKind_Array:
			return "[" + ____rune_private_6bd43c78_emitTSExprList(__expr.__children, 0, "") + "]"
		case __expr.__kind == __ExprKind_Tuple:
			return "[" + ____rune_private_6bd43c78_emitTSExprList(__expr.__children, 0, "") + "]"
		case __expr.__kind == __ExprKind_Map:
			return "new Map([" + ____rune_private_6bd43c78_emitTSMapEntries(__expr.__children, 0, "") + "])"
		case __expr.__kind == __ExprKind_Spread:
			return "..." + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Reactive:
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0])
		case __expr.__kind == __ExprKind_Struct:
			return "{" + ____rune_private_6bd43c78_emitTSFields(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Object:
			return "{" + ____rune_private_6bd43c78_emitTSFields(__expr.__children, 0, "") + "}"
		case __expr.__kind == __ExprKind_Block:
			return "(() => {\n" + ____rune_private_6bd43c78_emitTSBlock(__expr.__children, 0, true, "Dynamic", 1, "") + "})()"
		default:
			return "undefined"
		}
	}()
}

func ____rune_private_6bd43c78_emitTSExprExpected(__expr __IRExpr, __expected string) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_6bd43c78_emitTSCallExpected(__expr, __expected)
		default:
			return ____rune_private_6bd43c78_emitTSExpr(__expr)
		}
	}()
}

func ____rune_private_6bd43c78_emitTSCallExpected(__expr __IRExpr, __expected string) string {
	__args := __genericInner(__expected, "Result")
	return func() string {
		if __args != "" && ____rune_private_6bd43c78_isTSResultConstructorCall(__expr) {
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + "<" + ____rune_private_6bd43c78_emitTSTypeArgs(__args) + ">(" + ____rune_private_6bd43c78_emitTSExprListFrom(__expr.__children, 1, "") + ")"
		}
		return ____rune_private_6bd43c78_emitTSCall(__expr)
	}()
}

func ____rune_private_6bd43c78_isTSResultConstructorCall(__expr __IRExpr) bool {
	return __expr.__kind == __ExprKind_Call && len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier && (__expr.__children[0].__name == "Ok" || __expr.__children[0].__name == "Err")
}

func ____rune_private_6bd43c78_emitTSBinary(__expr __IRExpr) string {
	return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + " " + ____rune_private_6bd43c78_tsBinaryOp(__expr.__op) + " " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1])
}

func ____rune_private_6bd43c78_emitTSTernary(__expr __IRExpr) string {
	return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + " ? " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + " : " + func() string {
		if len(__expr.__children) > 2 {
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2])
		}
		return "undefined"
	}()
}

func ____rune_private_6bd43c78_emitTSAssign(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) == 2 {
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + " = " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1])
		}
		return __mangleIdent(__expr.__name) + " = " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0])
	}()
}

func ____rune_private_6bd43c78_emitTSCall(__expr __IRExpr) string {
	return func() string {
		switch {
		case __moduleCallKey(__expr) == "io.println":
			return "console.log(" + ____rune_private_6bd43c78_emitTSExprListFrom(__expr.__children, 1, "") + ")"
		case __moduleCallKey(__expr) == "json.stringify":
			return "JSON.stringify(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "json.parse":
			return "JSON.parse(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "map.new":
			return "new Map()"
		case __moduleCallKey(__expr) == "set.new":
			return "new Set()"
		case __moduleCallKey(__expr) == "path.isAbsolute":
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ".startsWith(\"/\")"
		case __moduleCallKey(__expr) == "path.basename":
			return "__runePathBasename(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.extname":
			return "__runePathExtname(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.dirname":
			return "__runePathDirname(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.join":
			return "__runePathJoin(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.normalize":
			return "__runePathNormalize(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.resolve":
			return "__runePathResolve(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.relative":
			return "__runePathRelative(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.joinParts":
			return "__runePathJoinParts(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.appendPathPart":
			return "__runePathAppendPart(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ")"
		case __moduleCallKey(__expr) == "path.normalizeParts":
			return "__runePathNormalizeParts(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[4]) + ")"
		case __moduleCallKey(__expr) == "path.normalizePart":
			return "__runePathNormalizeParts(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[4]) + ")"
		case __moduleCallKey(__expr) == "path.normalizeParent":
			return "__runePathNormalizeParent(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[4]) + ")"
		case __moduleCallKey(__expr) == "path.normalizePop":
			return "__runePathNormalizePop(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[4]) + ")"
		case __moduleCallKey(__expr) == "path.normalizePush":
			return "__runePathNormalizePush(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[4]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[5]) + ")"
		case __moduleCallKey(__expr) == "path.pathParts":
			return "__runePathParts(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "path.collectPathParts":
			return "__runePathCollectParts(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.collectPathPart":
			return "__runePathCollectPart(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.relativeFromParts":
			return "__runePathRelativeFromParts(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ")"
		case __moduleCallKey(__expr) == "path.relativeTail":
			return "__runePathRelativeTail(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[3]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[4]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[5]) + ")"
		case __moduleCallKey(__expr) == "process.platform":
			return "\"js\""
		case __moduleCallKey(__expr) == "process.cwd":
			return "\".\""
		case __moduleCallKey(__expr) == "process.env":
			return "null"
		case __moduleCallKey(__expr) == "process.argv":
			return "[]"
		case __moduleCallKey(__expr) == "int.toString":
			return "String(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int.toDouble":
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1])
		case __moduleCallKey(__expr) == "int.toBigInt":
			return "BigInt(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int4.fromInt":
			return "((__value: number): number => { const __n = __value & 0xf; return __n >= 8 ? __n - 16 : __n; })(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int8.fromInt":
			return "((__value: number): number => (__value << 24) >> 24)(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int16.fromInt":
			return "((__value: number): number => (__value << 16) >> 16)(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "int64.fromInt":
			return "BigInt(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "uint.fromInt":
			return "(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + " >>> 0)"
		case __moduleCallKey(__expr) == "uint8.fromInt":
			return "(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + " & 0xff)"
		case __moduleCallKey(__expr) == "uint16.fromInt":
			return "(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + " & 0xffff)"
		case __moduleCallKey(__expr) == "uint64.fromInt":
			return "BigInt.asUintN(64, BigInt(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + "))"
		case __moduleCallKey(__expr) == "float.fromDouble":
			return "Math.fround(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case (__moduleCallKey(__expr) == "int4.toInt") || (__moduleCallKey(__expr) == "int8.toInt") || (__moduleCallKey(__expr) == "int16.toInt") || (__moduleCallKey(__expr) == "uint.toInt") || (__moduleCallKey(__expr) == "uint8.toInt") || (__moduleCallKey(__expr) == "uint16.toInt") || (__moduleCallKey(__expr) == "float.toDouble"):
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1])
		case (__moduleCallKey(__expr) == "int64.toInt") || (__moduleCallKey(__expr) == "uint64.toInt"):
			return "Number(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.fromInt":
			return "BigInt(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.toString":
			return "String(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "bigint.toDouble":
			return "Number(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.trunc":
			return "Math.trunc(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.floor":
			return "Math.floor(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.ceil":
			return "Math.ceil(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		case __moduleCallKey(__expr) == "double.round":
			return "Math.round(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ")"
		default:
			return ____rune_private_6bd43c78_emitTSMaybeCoreMethodCall(__expr)
		}
	}()
}

func ____rune_private_6bd43c78_emitTSMaybeCoreMethodCall(__expr __IRExpr) string {
	return func() string {
		if len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Selector {
			return ____rune_private_6bd43c78_emitTSCoreMethodCall(__expr, __expr.__children[0])
		}
		return ____rune_private_6bd43c78_emitTSDefaultCall(__expr)
	}()
}

func ____rune_private_6bd43c78_emitTSCoreMethodCall(__expr __IRExpr, __selector __IRExpr) string {
	return func() string {
		if len(__selector.__children) > 0 && __selector.__children[0].__kind != __ExprKind_At {
			return func() string {
				switch {
				case (__selector.__name == "length") || (__selector.__name == "byteLength"):
					return ____rune_private_6bd43c78_emitTSCoreLength(__selector.__children[0])
				case __selector.__name == "isEmpty":
					return "(" + ____rune_private_6bd43c78_emitTSCoreLength(__selector.__children[0]) + " === 0)"
				case __selector.__name == "at":
					return ____rune_private_6bd43c78_emitTSCoreAt(__expr, __selector.__children[0])
				case __selector.__name == "slice":
					return ____rune_private_6bd43c78_emitTSCoreSlice(__expr, __selector.__children[0])
				default:
					return ____rune_private_6bd43c78_emitTSDefaultCall(__expr)
				}
			}()
		}
		return ____rune_private_6bd43c78_emitTSDefaultCall(__expr)
	}()
}

func ____rune_private_6bd43c78_emitTSCoreLength(__receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "Array.from(" + ____rune_private_6bd43c78_emitTSExpr(__receiver) + ").length"
		}
		return ____rune_private_6bd43c78_emitTSExpr(__receiver) + ".length"
	}()
}

func ____rune_private_6bd43c78_emitTSCoreAt(__expr __IRExpr, __receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "(Array.from(" + ____rune_private_6bd43c78_emitTSExpr(__receiver) + ")[" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + "] ?? \"\")"
		}
		return ____rune_private_6bd43c78_emitTSExpr(__receiver) + "[" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + "]"
	}()
}

func ____rune_private_6bd43c78_emitTSCoreSlice(__expr __IRExpr, __receiver __IRExpr) string {
	return func() string {
		if __receiver.__text == "String" {
			return "Array.from(" + ____rune_private_6bd43c78_emitTSExpr(__receiver) + ").slice(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ").join(\"\")"
		}
		return ____rune_private_6bd43c78_emitTSExpr(__receiver) + ".slice(" + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[1]) + ", " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[2]) + ")"
	}()
}

func ____rune_private_6bd43c78_emitTSDefaultCall(__expr __IRExpr) string {
	return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + "(" + ____rune_private_6bd43c78_emitTSExprListFrom(__expr.__children, 1, "") + ")"
}

func ____rune_private_6bd43c78_emitTSLambda(__expr __IRExpr) string {
	return "(" + ____rune_private_6bd43c78_emitTSParams(__expr.__params, 0, "") + ") => " + ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0])
}

func ____rune_private_6bd43c78_emitTSSelector(__expr __IRExpr) string {
	return func() string {
		switch {
		case __expr.__children[0].__kind == __ExprKind_At:
			return ____rune_private_6bd43c78_emitTSAtSelector(__expr)
		case __expr.__children[0].__kind == __ExprKind_Identifier:
			return func() string {
				switch {
				case __expr.__op == "::":
					return __mangleIdent(__expr.__children[0].__name + "_" + __expr.__name)
				default:
					return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + "." + ____rune_private_6bd43c78_tsPropertyName(__expr.__name)
				}
			}()
		default:
			return ____rune_private_6bd43c78_emitTSExpr(__expr.__children[0]) + "." + ____rune_private_6bd43c78_tsPropertyName(__expr.__name)
		}
	}()
}

func ____rune_private_6bd43c78_emitTSAtSelector(__expr __IRExpr) string {
	__imported := __expr.__children[0].__value != ""
	return func() string {
		switch {
		case __imported == true:
			return __mangleIdent(__expr.__name)
		default:
			return "@" + __expr.__children[0].__name + "." + __expr.__name
		}
	}()
}

func ____rune_private_6bd43c78_emitTSExprList(__exprs []__IRExpr, __index int, __out string) string {
	return ____rune_private_6bd43c78_emitTSExprListFrom(__exprs, __index, __out)
}

func ____rune_private_6bd43c78_emitTSExprListFrom(__exprs []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__exprs) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSExprListFrom(__exprs, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_6bd43c78_emitTSExpr(__exprs[__index]))
	}()
}

func ____rune_private_6bd43c78_emitTSMapEntries(__entries []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__entries) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSMapEntries(__entries, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+"["+____rune_private_6bd43c78_emitTSExpr(__entries[__index].__children[0])+", "+____rune_private_6bd43c78_emitTSExpr(__entries[__index].__children[1])+"]")
	}()
}

func ____rune_private_6bd43c78_emitTSFields(__fields []__IRExpr, __index int, __out string) string {
	return func() string {
		if __index >= len(__fields) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSFields(__fields, __index+1, __out+func() string {
			if __out == "" {
				return ""
			}
			return ", "
		}()+____rune_private_6bd43c78_tsPropertyName(__fields[__index].__name)+": "+____rune_private_6bd43c78_emitTSExpr(__fields[__index].__children[0]))
	}()
}

func ____rune_private_6bd43c78_tsBinaryOp(__op string) string {
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

func ____rune_private_6bd43c78_tsType(__typeName string) string {
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
		return ____rune_private_6bd43c78_tsTypeFallback(__typeName)
	}
}

func ____rune_private_6bd43c78_tsTypeFallback(__typeName string) string {
	return func() string {
		if strings.HasSuffix(__typeName, "?") {
			return ____rune_private_6bd43c78_tsType(func() string { runes := []rune(__typeName); return string(runes[0 : len([]rune(__typeName))-1]) }()) + " | null"
		}
		return func() string {
			if __genericInner(__typeName, "Array") != "" {
				return ____rune_private_6bd43c78_tsType(__genericInner(__typeName, "Array")) + "[]"
			}
			return func() string {
				if __genericInner(__typeName, "ReadonlyArray") != "" {
					return "ReadonlyArray<" + ____rune_private_6bd43c78_tsType(__genericInner(__typeName, "ReadonlyArray")) + ">"
				}
				return func() string {
					if __genericInner(__typeName, "Map") != "" {
						return "Map<" + ____rune_private_6bd43c78_tsType(__typeArg(__genericInner(__typeName, "Map"), 0)) + ", " + ____rune_private_6bd43c78_tsType(__typeArg(__genericInner(__typeName, "Map"), 1)) + ">"
					}
					return func() string {
						if __genericInner(__typeName, "Set") != "" {
							return "Set<" + ____rune_private_6bd43c78_tsType(__genericInner(__typeName, "Set")) + ">"
						}
						return ____rune_private_6bd43c78_tsNamedType(__typeName)
					}()
				}()
			}()
		}()
	}()
}

func ____rune_private_6bd43c78_tsNamedType(__typeName string) string {
	__open := strings.Index(__typeName, "[")
	return func() string {
		if __open < 0 {
			return __mangleIdent(__typeName)
		}
		return __mangleIdent(func() string { runes := []rune(__typeName); return string(runes[0:__open]) }()) + "<" + ____rune_private_6bd43c78_emitTSTypeArgs(func() string { runes := []rune(__typeName); return string(runes[__open+1 : len([]rune(__typeName))-1]) }()) + ">"
	}()
}

func ____rune_private_6bd43c78_emitTSTypeArgs(__args string) string {
	return ____rune_private_6bd43c78_emitTSTypeArgList(func() []string { parts := strings.Split(__args, ","); return parts }(), 0, "")
}

func ____rune_private_6bd43c78_emitTSTypeArgList(__args []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_6bd43c78_emitTSTypeArgList(__args, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_6bd43c78_tsType(strings.TrimSpace(__args[__index])))
	}()
}

func ____rune_private_6bd43c78_tsPropertyName(__name string) string {
	return __name
}

func ____rune_private_6bd43c78_joinStrings(__values []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__values) {
			return __out
		}
		return ____rune_private_6bd43c78_joinStrings(__values, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__values[__index])
	}()
}

func __generateDeclarations(__file __IRFile) string {
	__out := ____rune_private_01b5d206_dtsPreamble()
	__out = __out + ____rune_private_01b5d206_dtsStructs(__file.__structs, 0)
	__out = __out + ____rune_private_01b5d206_dtsEnums(__file.__enums, 0)
	__out = __out + ____rune_private_01b5d206_dtsConsts(__file.__constants, 0)
	__out = __out + ____rune_private_01b5d206_dtsFunctions(__file.__functions, 0)
	return __out + ____rune_private_01b5d206_dtsExports(__file)
}

func ____rune_private_01b5d206_dtsPreamble() string {
	return "type RuneResult<T, E> = { ok: true; value: T } | { ok: false; error: E };\ntype RuneError = { code: number; message: string; cause: RuneError | null };\ntype RuneIter<T> = { next: () => [T, boolean] };\ntype RuneFileStat = { size: number; isFile: boolean; isDirectory: boolean };\ntype RuneTCPConnection = { socket: unknown };\ntype RuneTCPListener = { server: unknown; address: string };\ndeclare class RuneStringBuffer {}\ndeclare class RuneBuffer {}\ndeclare class RuneReader {}\ndeclare class RuneWriter {}\n\n"
}

func ____rune_private_01b5d206_dtsStructs(__structs []__IRStructType, __index int) string {
	return func() string {
		if __index >= len(__structs) {
			return ""
		}
		return func() string {
			if __structs[__index].__private {
				return ____rune_private_01b5d206_dtsStructs(__structs, __index+1)
			}
			return ____rune_private_01b5d206_dtsStruct(__structs[__index]) + ____rune_private_01b5d206_dtsStructs(__structs, __index+1)
		}()
	}()
}

func ____rune_private_01b5d206_dtsStruct(__typeDecl __IRStructType) string {
	__generics := ____rune_private_01b5d206_dtsGenerics(__typeDecl.__generics)
	__out := "type " + __mangleIdent(__typeDecl.__name) + __generics + " = {\n"
	__out = __out + ____rune_private_01b5d206_dtsStructFields(__typeDecl.__fields, 0)
	return __out + "};\n" + "\n"
}

func ____rune_private_01b5d206_dtsStructFields(__fields []__IRField, __index int) string {
	return func() string {
		if __index >= len(__fields) {
			return ""
		}
		return func() string {
			if __fields[__index].__private {
				return ____rune_private_01b5d206_dtsStructFields(__fields, __index+1)
			}
			return __indent(1) + ____rune_private_01b5d206_dtsPropertyName(__fields[__index].__name) + ": " + ____rune_private_01b5d206_dtsType(__fields[__index].__typeName) + ";\n" + ____rune_private_01b5d206_dtsStructFields(__fields, __index+1)
		}()
	}()
}

func ____rune_private_01b5d206_dtsEnums(__enums []__IREnumType, __index int) string {
	return func() string {
		if __index >= len(__enums) {
			return ""
		}
		return func() string {
			if __enums[__index].__private {
				return ____rune_private_01b5d206_dtsEnums(__enums, __index+1)
			}
			return ____rune_private_01b5d206_dtsEnum(__enums[__index]) + ____rune_private_01b5d206_dtsEnums(__enums, __index+1)
		}()
	}()
}

func ____rune_private_01b5d206_dtsEnum(__enumDecl __IREnumType) string {
	__name := __mangleIdent(__enumDecl.__name)
	__generics := ____rune_private_01b5d206_dtsGenerics(__enumDecl.__generics)
	__shape := func() string {
		switch {
		case ____rune_private_01b5d206_dtsEnumHasPayload(__enumDecl.__members, 0) == true:
			return "{ tag: number; payload: any[] }"
		case ____rune_private_01b5d206_dtsEnumHasPayload(__enumDecl.__members, 0) == false:
			return "number"
		}
		return ""
	}()
	__out := "type " + __name + __generics + " = " + __shape + ";\n"
	__out = __out + "declare const " + __name + ": {\n"
	__out = __out + ____rune_private_01b5d206_dtsEnumMembers(__enumDecl.__members, 0)
	return __out + "};\n" + "\n"
}

func ____rune_private_01b5d206_dtsEnumHasPayload(__members []__IREnumMember, __index int) bool {
	return func() bool {
		if __index >= len(__members) {
			return false
		}
		return func() bool {
			if len(__members[__index].__params) > 0 {
				return true
			}
			return ____rune_private_01b5d206_dtsEnumHasPayload(__members, __index+1)
		}()
	}()
}

func ____rune_private_01b5d206_dtsEnumMembers(__members []__IREnumMember, __index int) string {
	return func() string {
		if __index >= len(__members) {
			return ""
		}
		return ____rune_private_01b5d206_dtsEnumMembersAt(__members, __index)
	}()
}

func ____rune_private_01b5d206_dtsEnumMembersAt(__members []__IREnumMember, __index int) string {
	return func() string {
		if __members[__index].__private {
			return ____rune_private_01b5d206_dtsEnumMembers(__members, __index+1)
		}
		return __indent(1) + "readonly " + ____rune_private_01b5d206_dtsPropertyName(__members[__index].__name) + ": " + __enumValue(__members[__index], __index) + ";\n" + ____rune_private_01b5d206_dtsEnumMembers(__members, __index+1)
	}()
}

func ____rune_private_01b5d206_dtsConsts(__constants []__IRConst, __index int) string {
	return func() string {
		if __index >= len(__constants) {
			return ""
		}
		return ____rune_private_01b5d206_dtsConstsAt(__constants, __index)
	}()
}

func ____rune_private_01b5d206_dtsConstsAt(__constants []__IRConst, __index int) string {
	return func() string {
		if __constants[__index].__private {
			return ____rune_private_01b5d206_dtsConsts(__constants, __index+1)
		}
		return "declare const " + __mangleIdent(__constants[__index].__name) + ": " + ____rune_private_01b5d206_dtsType(__constants[__index].__typeName) + ";\n" + ____rune_private_01b5d206_dtsConsts(__constants, __index+1)
	}()
}

func ____rune_private_01b5d206_dtsFunctions(__functions []__IRFunction, __index int) string {
	return func() string {
		if __index >= len(__functions) {
			return ""
		}
		return ____rune_private_01b5d206_dtsFunctionAt(__functions, __index) + ____rune_private_01b5d206_dtsFunctions(__functions, __index+1)
	}()
}

func ____rune_private_01b5d206_dtsFunctionAt(__functions []__IRFunction, __index int) string {
	return func() string {
		if __functions[__index].__macro || __functions[__index].__private {
			return ""
		}
		return ____rune_private_01b5d206_dtsFunction(__functions[__index])
	}()
}

func ____rune_private_01b5d206_dtsFunction(__fn __IRFunction) string {
	__returnType := ____rune_private_01b5d206_dtsType(__fn.__returnType)
	__wrapped := func() string {
		switch {
		case __fn.__routine == true:
			return "Promise<" + __returnType + ">"
		case __fn.__routine == false:
			return __returnType
		}
		return ""
	}()
	return "declare function " + __mangleIdent(__fn.__name) + ____rune_private_01b5d206_dtsGenerics(__fn.__generics) + "(" + ____rune_private_01b5d206_dtsParams(__fn.__params, 0, "") + "): " + __wrapped + ";\n"
}

func ____rune_private_01b5d206_dtsGenerics(__generics []string) string {
	return func() string {
		if len(__generics) == 0 {
			return ""
		}
		return "<" + ____rune_private_01b5d206_dtsGenericNames(__generics, 0, "") + ">"
	}()
}

func ____rune_private_01b5d206_dtsGenericNames(__generics []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__generics) {
			return __out
		}
		return ____rune_private_01b5d206_dtsGenericNames(__generics, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__generics[__index])+" extends unknown")
	}()
}

func ____rune_private_01b5d206_dtsParams(__params []__IRParam, __index int, __out string) string {
	return func() string {
		if __index >= len(__params) {
			return __out
		}
		return ____rune_private_01b5d206_dtsParams(__params, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+__mangleIdent(__params[__index].__name)+": "+____rune_private_01b5d206_dtsType(__params[__index].__typeName))
	}()
}

func ____rune_private_01b5d206_dtsType(__typeName string) string {
	switch {
	case (__typeName == "") || (__typeName == "Void"):
		return "void"
	case (__typeName == "Int") || (__typeName == "Int4") || (__typeName == "Int8") || (__typeName == "Int16") || (__typeName == "UInt") || (__typeName == "UInt8") || (__typeName == "UInt16") || (__typeName == "Byte") || (__typeName == "Double") || (__typeName == "Float"):
		return "number"
	case (__typeName == "BigInt") || (__typeName == "Int64") || (__typeName == "UInt64"):
		return "bigint"
	case (__typeName == "String") || (__typeName == "Char"):
		return "string"
	case __typeName == "Bool":
		return "boolean"
	case __typeName == "Null":
		return "null"
	case __typeName == "Object":
		return "object"
	case __typeName == "Bytes":
		return "DataView"
	case __typeName == "Buffer":
		return "RuneBuffer"
	case __typeName == "Reader":
		return "RuneReader"
	case __typeName == "Writer":
		return "RuneWriter"
	case __typeName == "StringBuffer":
		return "RuneStringBuffer"
	case __typeName == "FileStat":
		return "RuneFileStat"
	case __typeName == "TCPConnection":
		return "RuneTCPConnection"
	case __typeName == "TCPListener":
		return "RuneTCPListener"
	case (__typeName == "Data") || (__typeName == "@io.Data"):
		return "Uint8Array"
	case __typeName == "Error":
		return "RuneError"
	case __typeName == "Never":
		return "never"
	case __typeName == "Symbol":
		return "symbol"
	case __typeName == "Regex":
		return "RegExp"
	case __typeName == "HTMLElement":
		return "HTMLElement"
	case __typeName == "WebComponent":
		return "CustomElementConstructor"
	case (__typeName == "Dynamic") || (__typeName == "Unknown"):
		return "any"
	default:
		return ____rune_private_01b5d206_dtsTypeFallback(__typeName)
	}
}

func ____rune_private_01b5d206_dtsTypeFallback(__typeName string) string {
	return func() string {
		if strings.HasPrefix(__typeName, "&") {
			return "any"
		}
		return func() string {
			if strings.HasSuffix(__typeName, "?") {
				return ____rune_private_01b5d206_dtsType(func() string { runes := []rune(__typeName); return string(runes[0 : len([]rune(__typeName))-1]) }()) + " | null"
			}
			return func() string {
				if __genericInner(__typeName, "Array") != "" {
					return ____rune_private_01b5d206_dtsType(__genericInner(__typeName, "Array")) + "[]"
				}
				return func() string {
					if __genericInner(__typeName, "Result") != "" {
						return ____rune_private_01b5d206_dtsGenericType("RuneResult", __genericInner(__typeName, "Result"))
					}
					return func() string {
						if __genericInner(__typeName, "Task") != "" {
							return ____rune_private_01b5d206_dtsGenericType("Promise", __genericInner(__typeName, "Task"))
						}
						return func() string {
							if __genericInner(__typeName, "Iter") != "" {
								return ____rune_private_01b5d206_dtsGenericType("RuneIter", __genericInner(__typeName, "Iter"))
							}
							return func() string {
								if __genericInner(__typeName, "ReadonlyArray") != "" {
									return ____rune_private_01b5d206_dtsGenericType("ReadonlyArray", __genericInner(__typeName, "ReadonlyArray"))
								}
								return func() string {
									if __genericInner(__typeName, "Tuple") != "" {
										return "[" + ____rune_private_01b5d206_dtsTypeArgs(__genericInner(__typeName, "Tuple")) + "]"
									}
									return func() string {
										if __genericInner(__typeName, "ReadonlyTuple") != "" {
											return "readonly [" + ____rune_private_01b5d206_dtsTypeArgs(__genericInner(__typeName, "ReadonlyTuple")) + "]"
										}
										return func() string {
											if __genericInner(__typeName, "Map") != "" {
												return ____rune_private_01b5d206_dtsGenericType("Map", __genericInner(__typeName, "Map"))
											}
											return func() string {
												if __genericInner(__typeName, "Set") != "" {
													return ____rune_private_01b5d206_dtsGenericType("Set", __genericInner(__typeName, "Set"))
												}
												return func() string {
													if __genericInner(__typeName, "WeakMap") != "" {
														return ____rune_private_01b5d206_dtsGenericType("WeakMap", __genericInner(__typeName, "WeakMap"))
													}
													return func() string {
														if __genericInner(__typeName, "WeakSet") != "" {
															return ____rune_private_01b5d206_dtsGenericType("WeakSet", __genericInner(__typeName, "WeakSet"))
														}
														return func() string {
															if __genericInner(__typeName, "Record") != "" {
																return ____rune_private_01b5d206_dtsGenericType("Record", __genericInner(__typeName, "Record"))
															}
															return ____rune_private_01b5d206_dtsNamedType(__typeName)
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
				}()
			}()
		}()
	}()
}

func ____rune_private_01b5d206_dtsGenericType(__name string, __args string) string {
	return __name + "<" + ____rune_private_01b5d206_dtsTypeArgs(__args) + ">"
}

func ____rune_private_01b5d206_dtsNamedType(__typeName string) string {
	__open := strings.Index(__typeName, "[")
	return func() string {
		if __open < 0 {
			return __mangleIdent(__typeName)
		}
		return __mangleIdent(func() string { runes := []rune(__typeName); return string(runes[0:__open]) }()) + "<" + ____rune_private_01b5d206_dtsTypeArgs(func() string { runes := []rune(__typeName); return string(runes[__open+1 : len([]rune(__typeName))-1]) }()) + ">"
	}()
}

func ____rune_private_01b5d206_dtsTypeArgs(__args string) string {
	return ____rune_private_01b5d206_dtsTypeArgList(func() []string { parts := strings.Split(__args, ","); return parts }(), 0, "")
}

func ____rune_private_01b5d206_dtsTypeArgList(__args []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__args) {
			return __out
		}
		return ____rune_private_01b5d206_dtsTypeArgList(__args, __index+1, __out+func() string {
			if __index == 0 {
				return ""
			}
			return ", "
		}()+____rune_private_01b5d206_dtsType(strings.TrimSpace(__args[__index])))
	}()
}

func ____rune_private_01b5d206_dtsExports(__file __IRFile) string {
	return func() string {
		if len(__file.__structs)+len(__file.__enums)+len(__file.__constants)+len(__file.__functions) == 0 {
			return ""
		}
		return "\n" + ____rune_private_01b5d206_dtsStructExports(__file.__structs, 0) + ____rune_private_01b5d206_dtsEnumExports(__file.__enums, 0) + ____rune_private_01b5d206_dtsConstExports(__file.__constants, 0) + ____rune_private_01b5d206_dtsFunctionExports(__file.__functions, 0)
	}()
}

func ____rune_private_01b5d206_dtsStructExports(__structs []__IRStructType, __index int) string {
	return func() string {
		if __index >= len(__structs) {
			return ""
		}
		return func() string {
			if __structs[__index].__private {
				return ____rune_private_01b5d206_dtsStructExports(__structs, __index+1)
			}
			return ____rune_private_01b5d206_dtsExportTypeAlias(__structs[__index].__name) + ____rune_private_01b5d206_dtsStructExports(__structs, __index+1)
		}()
	}()
}

func ____rune_private_01b5d206_dtsEnumExports(__enums []__IREnumType, __index int) string {
	return func() string {
		if __index >= len(__enums) {
			return ""
		}
		return func() string {
			if __enums[__index].__private {
				return ____rune_private_01b5d206_dtsEnumExports(__enums, __index+1)
			}
			return ____rune_private_01b5d206_dtsExportTypeAlias(__enums[__index].__name) + ____rune_private_01b5d206_dtsExportValueAlias(__enums[__index].__name) + ____rune_private_01b5d206_dtsEnumExports(__enums, __index+1)
		}()
	}()
}

func ____rune_private_01b5d206_dtsConstExports(__constants []__IRConst, __index int) string {
	return func() string {
		if __index >= len(__constants) {
			return ""
		}
		return func() string {
			if __constants[__index].__private {
				return ____rune_private_01b5d206_dtsConstExports(__constants, __index+1)
			}
			return ____rune_private_01b5d206_dtsExportValueAlias(__constants[__index].__name) + ____rune_private_01b5d206_dtsConstExports(__constants, __index+1)
		}()
	}()
}

func ____rune_private_01b5d206_dtsFunctionExports(__functions []__IRFunction, __index int) string {
	return func() string {
		if __index >= len(__functions) {
			return ""
		}
		return func() string {
			if __functions[__index].__macro || __functions[__index].__private {
				return ____rune_private_01b5d206_dtsFunctionExports(__functions, __index+1)
			}
			return ____rune_private_01b5d206_dtsExportValueAlias(__functions[__index].__name) + ____rune_private_01b5d206_dtsFunctionExports(__functions, __index+1)
		}()
	}()
}

func ____rune_private_01b5d206_dtsExportTypeAlias(__name string) string {
	return func() string {
		if ____rune_private_01b5d206_dtsCanUseBareProperty(__name) {
			return "export type " + __name + " = " + __mangleIdent(__name) + ";\n"
		}
		return "export type { " + __mangleIdent(__name) + " as " + ____rune_private_01b5d206_dtsExportName(__name) + " };\n"
	}()
}

func ____rune_private_01b5d206_dtsExportValueAlias(__name string) string {
	return func() string {
		if ____rune_private_01b5d206_dtsCanUseBareProperty(__name) {
			return "export declare const " + __name + ": typeof " + __mangleIdent(__name) + ";\n"
		}
		return "export { " + __mangleIdent(__name) + " as " + ____rune_private_01b5d206_dtsExportName(__name) + " };\n"
	}()
}

func ____rune_private_01b5d206_dtsPropertyName(__name string) string {
	return func() string {
		if ____rune_private_01b5d206_dtsCanUseBareProperty(__name) {
			return __name
		}
		return ____rune_private_01b5d206_dtsQuoteName(__name)
	}()
}

func ____rune_private_01b5d206_dtsExportName(__name string) string {
	return ____rune_private_01b5d206_dtsPropertyName(__name)
}

func ____rune_private_01b5d206_dtsQuoteName(__name string) string {
	return "\"" + ____rune_private_01b5d206_dtsEscapeName(__name, 0, "") + "\""
}

func ____rune_private_01b5d206_dtsEscapeName(__name string, __index int, __out string) string {
	return func() string {
		if __index >= len([]rune(__name)) {
			return __out
		}
		return ____rune_private_01b5d206_dtsEscapeName(__name, __index+1, __out+____rune_private_01b5d206_dtsEscapeNameChar([]rune(__name)[__index]))
	}()
}

func ____rune_private_01b5d206_dtsEscapeNameChar(__ch rune) string {
	return func() string {
		switch {
		case __ch == '\\':
			return "\\\\"
		case __ch == '"':
			return "\\\""
		case __ch == '\n':
			return "\\n"
		case __ch == '\r':
			return "\\r"
		case __ch == '\t':
			return "\\t"
		default:
			return string(__ch)
		}
	}()
}

func ____rune_private_01b5d206_dtsCanUseBareProperty(__name string) bool {
	return __name != "" && ____rune_private_01b5d206_dtsReservedName(__name) == false && ____rune_private_01b5d206_dtsSafeIdent(__name, 0, true)
}

func ____rune_private_01b5d206_dtsSafeIdent(__name string, __index int, __first bool) bool {
	return func() bool {
		if __index >= len([]rune(__name)) {
			return true
		}
		return func() bool {
			if ____rune_private_01b5d206_dtsSafeIdentChar([]rune(__name)[__index], __first) {
				return ____rune_private_01b5d206_dtsSafeIdent(__name, __index+1, false)
			}
			return false
		}()
	}()
}

func ____rune_private_01b5d206_dtsSafeIdentChar(__ch rune, __first bool) bool {
	return func() bool {
		switch {
		case (__ch == '_') || (__ch == '$') || (__ch >= 'a' && __ch <= 'z') || (__ch >= 'A' && __ch <= 'Z'):
			return true
		case (__ch >= '0' && __ch <= '9'):
			return __first == false
		default:
			return false
		}
	}()
}

func ____rune_private_01b5d206_dtsReservedName(__name string) bool {
	return func() bool {
		switch {
		case (__name == "await") || (__name == "break") || (__name == "case") || (__name == "catch") || (__name == "class") || (__name == "const") || (__name == "continue") || (__name == "debugger") || (__name == "default") || (__name == "delete") || (__name == "do") || (__name == "else") || (__name == "enum") || (__name == "export") || (__name == "extends") || (__name == "false") || (__name == "finally") || (__name == "for") || (__name == "function") || (__name == "if") || (__name == "implements") || (__name == "import") || (__name == "in") || (__name == "instanceof") || (__name == "interface") || (__name == "let") || (__name == "new") || (__name == "null") || (__name == "package") || (__name == "private") || (__name == "protected") || (__name == "public") || (__name == "return") || (__name == "static") || (__name == "super") || (__name == "switch") || (__name == "this") || (__name == "throw") || (__name == "true") || (__name == "try") || (__name == "typeof") || (__name == "var") || (__name == "void") || (__name == "while") || (__name == "with") || (__name == "yield"):
			return true
		default:
			return false
		}
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

func __compileDeclarations(__source string) __CompileResult {
	return __compile(__source, "dts")
}

func __checkSource(__source string) __CompileResult {
	return ____rune_private_1ed26dbc_checkFile(____rune_private_1ed26dbc_lowerCompilerSource(__source))
}

func __compile(__source string, __target string) __CompileResult {
	__file := ____rune_private_1ed26dbc_lowerCompilerSource(__source)
	return ____rune_private_1ed26dbc_compileFile(__file, __target)
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

func __compileDeclarationsFiles(__files []__SourceFile) __CompileResult {
	return __compileFiles(__files, "dts")
}

func __compileFiles(__files []__SourceFile, __target string) __CompileResult {
	__file := ____rune_private_1ed26dbc_lowerFiles(__files)
	return ____rune_private_1ed26dbc_compileFile(__file, __target)
}

func __discoverSources(__root string) runeTask[runeResult[[]__SourceFile, *runeError]] {
	return runeGo(func() runeResult[[]__SourceFile, *runeError] {
		__result2 := runeAwait(runeFsReadFileText(__root))
		if !__result2.ok {
			return runeErr[[]__SourceFile, *runeError](__result2.err)
		}
		__source := __result2.value
		return runeOk[[]__SourceFile, *runeError](____rune_private_1ed26dbc_discoverSourceGraph([]__SourceFile{__SourceFile{__path: __root, __source: __source}}))
	})
}

func __hostBridgeSources(__root string, __files []__SourceFile) []__SourceFile {
	return ____rune_private_1ed26dbc_discoverSourceGraph(__files)
}

func ____rune_private_1ed26dbc_discoverSourceGraph(__files []__SourceFile) []__SourceFile {
	return ____rune_private_1ed26dbc_discoverMergeSourceFiles(append([]__SourceFile{}, []__SourceFile{____rune_private_1ed26dbc_emptySourceFile()}[0:0]...), __files, 0)
}

func __getSelfhostSources(__root string) runeTask[runeResult[__SelfhostSources, *runeError]] {
	return runeGo(func() runeResult[__SelfhostSources, *runeError] {
		__result3 := runeAwait(__discoverSources(__root))
		if !__result3.ok {
			return runeErr[__SelfhostSources, *runeError](__result3.err)
		}
		__files := __result3.value
		return runeOk[__SelfhostSources, *runeError](__SelfhostSources{__files: __files})
	})
}

func ____discoverSourcesPath(__root string) runeTask[runeResult[[]__SourceFile, *runeError]] {
	return runeGo(func() runeResult[[]__SourceFile, *runeError] {
		__result4 := runeAwait(__discoverSources(__root))
		if !__result4.ok {
			return runeErr[[]__SourceFile, *runeError](__result4.err)
		}
		__files := __result4.value
		return runeOk[[]__SourceFile, *runeError](__files)
	})
}

func ____rune_private_1ed26dbc_discoverSourceImportPaths(__file __SourceFile) []string {
	__parsed := ____rune_private_1ed26dbc_expandCompilerMacros(__parse(__file.__source))
	return ____rune_private_1ed26dbc_appendDiscoverImportPaths([]string{}, __file.__path, __parsed.__imports, 0)
}

func ____rune_private_1ed26dbc_appendDiscoverImportPaths(__pending []string, __basePath string, __imports []__ParsedImport, __index int) []string {
	return func() []string {
		if __index >= len(__imports) {
			return __pending
		}
		return ____rune_private_1ed26dbc_appendDiscoverImportPath(__pending, __basePath, __imports, __index)
	}()
}

func ____rune_private_1ed26dbc_appendDiscoverImportPath(__pending []string, __basePath string, __imports []__ParsedImport, __index int) []string {
	__importDecl := __imports[__index]
	__skip := __importDecl.__go || __importDecl.__module
	return func() []string {
		switch {
		case __skip == true:
			return ____rune_private_1ed26dbc_appendDiscoverImportPaths(__pending, __basePath, __imports, __index+1)
		default:
			return ____rune_private_1ed26dbc_appendDiscoverImportPaths(func() []string {
				out := []string{}
				out = append(out, __pending...)
				out = append(out, ____rune_private_1ed26dbc_resolveCompilerImportPath(__basePath, __importDecl.__path))
				return out
			}(), __basePath, __imports, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_discoverMergeSourceFiles(__left []__SourceFile, __right []__SourceFile, __index int) []__SourceFile {
	return func() []__SourceFile {
		if __index >= len(__right) {
			return __left
		}
		return ____rune_private_1ed26dbc_discoverMergeSourceFiles(func() []__SourceFile {
			if ____rune_private_1ed26dbc_compilerContainsSourceFile(__left, __right[__index].__path, 0) {
				return __left
			}
			return func() []__SourceFile {
				out := []__SourceFile{}
				out = append(out, __left...)
				out = append(out, __right[__index])
				return out
			}()
		}(), __right, __index+1)
	}()
}

func ____rune_private_1ed26dbc_compilerContainsSourceFile(__files []__SourceFile, __path string, __index int) bool {
	return func() bool {
		if __index >= len(__files) {
			return false
		}
		return ____rune_private_1ed26dbc_compilerPathNormalize(__files[__index].__path) == ____rune_private_1ed26dbc_compilerPathNormalize(__path) || ____rune_private_1ed26dbc_compilerContainsSourceFile(__files, __path, __index+1)
	}()
}

func ____rune_private_1ed26dbc_compileFile(__file __IRFile, __target string) __CompileResult {
	return func() __CompileResult {
		if len(__file.__errors) > 0 {
			return ____rune_private_1ed26dbc_compileResult(false, "", ____rune_private_1ed26dbc_parseErrorMessages(__file.__errors))
		}
		return ____rune_private_1ed26dbc_compileCheckedFile(__file, __target)
	}()
}

func ____rune_private_1ed26dbc_checkFile(__file __IRFile) __CompileResult {
	return func() __CompileResult {
		if len(__file.__errors) > 0 {
			return ____rune_private_1ed26dbc_compileResult(false, "", ____rune_private_1ed26dbc_parseErrorMessages(__file.__errors))
		}
		return ____rune_private_1ed26dbc_checkCheckedFile(__file)
	}()
}

func ____rune_private_1ed26dbc_checkCheckedFile(__file __IRFile) __CompileResult {
	__errors := ____rune_private_1ed26dbc_checkFileErrors(__file)
	return func() __CompileResult {
		if len(__errors) > 0 {
			return ____rune_private_1ed26dbc_compileResult(false, "", __errors)
		}
		return ____rune_private_1ed26dbc_compileResult(true, "", []string{})
	}()
}

func ____rune_private_1ed26dbc_compileCheckedFile(__file __IRFile, __target string) __CompileResult {
	__errors := ____rune_private_1ed26dbc_checkFileErrors(__file)
	__targetErrors := ____rune_private_1ed26dbc_checkTargetFileErrors(__file, __target)
	__allErrors := func() []string {
		out := []string{}
		out = append(out, __errors...)
		out = append(out, __targetErrors...)
		return out
	}()
	return func() __CompileResult {
		if len(__allErrors) > 0 {
			return ____rune_private_1ed26dbc_compileResult(false, "", __allErrors)
		}
		return func() __CompileResult {
			switch {
			case __target == "ts":
				return ____rune_private_1ed26dbc_compileResult(true, __generateTypeScript(__file), []string{})
			case __target == "go":
				return ____rune_private_1ed26dbc_compileResult(true, __generateGo(__file), []string{})
			case __target == "mbt":
				return ____rune_private_1ed26dbc_compileResult(true, __generateMoonBit(__file), []string{})
			case __target == "dts":
				return ____rune_private_1ed26dbc_compileResult(true, __generateDeclarations(__file), []string{})
			default:
				return ____rune_private_1ed26dbc_compileResult(false, "", ____rune_private_1ed26dbc_unsupportedTargetErrors(__target))
			}
		}()
	}()
}

func ____rune_private_1ed26dbc_checkTargetFileErrors(__file __IRFile, __target string) []string {
	return func() []string {
		switch {
		case __target == "ts":
			return ____rune_private_1ed26dbc_checkTypeScriptTargetFileErrors(__file)
		case __target == "go":
			return ____rune_private_1ed26dbc_checkGoTargetFileErrors(__file)
		case __target == "mbt":
			return ____rune_private_1ed26dbc_checkMoonBitTargetFileErrors(__file)
		default:
			return []string{}
		}
	}()
}

func ____rune_private_1ed26dbc_checkTypeScriptTargetFileErrors(__file __IRFile) []string {
	__hasGoImports := ____rune_private_1ed26dbc_fileHasGoImports(__file)
	return func() []string {
		switch {
		case __hasGoImports == true:
			return []string{"TypeScript backend does not support Go package imports"}
		default:
			return func() []string {
				switch {
				case ____rune_private_1ed26dbc_fileUsesGoFFI(__file) == true:
					return []string{"TypeScript backend does not support @go FFI"}
				default:
					return []string{}
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkGoTargetFileErrors(__file __IRFile) []string {
	__hasTypeScriptImports := len(__file.__tsImports) > 0
	return func() []string {
		switch {
		case __hasTypeScriptImports == true:
			return []string{"Go backend does not support TypeScript imports"}
		default:
			return []string{}
		}
	}()
}

func ____rune_private_1ed26dbc_checkMoonBitTargetFileErrors(__file __IRFile) []string {
	__errors := append([]string{}, []string{""}[0:0]...)
	__hasTypeScriptImports := len(__file.__tsImports) > 0
	__errors = ____rune_private_1ed26dbc_compilerAppendErrorIf(__errors, __hasTypeScriptImports, "MoonBit backend does not support TypeScript imports")
	__hasGoImports := ____rune_private_1ed26dbc_fileHasGoImports(__file)
	__errors = ____rune_private_1ed26dbc_compilerAppendErrorIf(__errors, __hasGoImports, "MoonBit backend does not support Go package imports")
	__hasGoFFI := __hasGoImports == false && ____rune_private_1ed26dbc_fileUsesGoFFI(__file)
	__errors = ____rune_private_1ed26dbc_compilerAppendErrorIf(__errors, __hasGoFFI, "MoonBit backend does not support @go FFI")
	return __errors
}

func ____rune_private_1ed26dbc_fileUsesGoFFI(__file __IRFile) bool {
	return __fileUsesModuleCall(__file, "go.stmt") || __fileUsesModuleCall(__file, "go.expr")
}

func ____rune_private_1ed26dbc_compilerAppendErrorIf(__errors []string, __condition bool, __message string) []string {
	return func() []string {
		switch {
		case __condition == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, __message)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_fileHasGoImports(__file __IRFile) bool {
	return ____rune_private_1ed26dbc_fileHasGoImportsAt(__file.__imports, 0)
}

func ____rune_private_1ed26dbc_fileHasGoImportsAt(__imports []__IRImport, __index int) bool {
	__done := __index >= len(__imports)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_1ed26dbc_fileHasGoImportAt(__imports, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_fileHasGoImportAt(__imports []__IRImport, __index int) bool {
	return func() bool {
		switch {
		case __imports[__index].__go == true:
			return true
		default:
			return ____rune_private_1ed26dbc_fileHasGoImportsAt(__imports, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_checkFileErrors(__file __IRFile) []string {
	__callables := ____rune_private_1ed26dbc_compilerCallables(__file)
	__bindings := ____rune_private_1ed26dbc_compilerInitialBindings(__file, __callables)
	__knownTypes := ____rune_private_1ed26dbc_compilerKnownTypes(__file)
	__errors := ____rune_private_1ed26dbc_checkDuplicateDeclarations(__file, append([]string{}, []string{""}[0:0]...))
	__errors = ____rune_private_1ed26dbc_checkDeclarationTypes(__file, __knownTypes, __errors)
	for _, __constant := range __file.__constants {
		_ = __constant
		__errors = ____rune_private_1ed26dbc_checkConstantErrors(__constant, __file.__structs, __callables, __errors, __bindings)
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__errors = ____rune_private_1ed26dbc_checkTopLevelFunctionErrors(__fn, __file.__structs, __callables, __errors, __bindings)
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		func() {
			for _, __method := range __typeDecl.__methods {
				_ = __method
				__errors = ____rune_private_1ed26dbc_checkMethodErrors(__typeDecl.__name, __method, __file.__structs, __callables, __errors, __bindings)
			}
		}()
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		func() {
			for _, __method := range __typeDecl.__methods {
				_ = __method
				__errors = ____rune_private_1ed26dbc_checkMethodErrors(__typeDecl.__name, __method, __file.__structs, __callables, __errors, __bindings)
			}
		}()
	}
	for _, __testDecl := range __file.__tests {
		_ = __testDecl
		__errors = ____rune_private_1ed26dbc_checkExpr(__testDecl.__body, __file.__structs, __callables, __errors, __bindings)
	}
	return __errors
}

func ____rune_private_1ed26dbc_checkConstantErrors(__constant __IRConst, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExpr(__constant.__value, __structs, __callables, __errors, __bindings)
	__actual := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__constant.__value, __structs, __callables, __bindings)
	__mismatch := ____rune_private_1ed26dbc_compilerShouldCheckArgType(__constant.__typeName, __actual) && ____rune_private_1ed26dbc_compilerTypesCompatible(__constant.__typeName, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __checked...)
				out = append(out, "constant "+__constant.__name+" has type "+__actual+", expected "+__constant.__typeName)
				return out
			}()
		default:
			return __checked
		}
	}()
}

func ____rune_private_1ed26dbc_checkTopLevelFunctionErrors(__fn __IRFunction, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case __fn.__macro == true:
			return ____rune_private_1ed26dbc_checkMacroFunctionReturn(__fn, __structs, __callables, __errors, ____rune_private_1ed26dbc_compilerSourcePathBindings(__fn.__sourcePath, __bindings))
		default:
			return ____rune_private_1ed26dbc_checkFunctionErrors(__fn, __structs, __callables, __errors, ____rune_private_1ed26dbc_compilerSourcePathBindings(__fn.__sourcePath, __bindings))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerSourcePathBindings(__sourcePath string, __bindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	return func() []__CompilerTypeBinding {
		switch {
		case __sourcePath == "":
			return __bindings
		default:
			return ____rune_private_1ed26dbc_addCompilerValueBinding(__bindings, "__sourcePath", __sourcePath)
		}
	}()
}

func ____rune_private_1ed26dbc_checkMacroFunctionReturn(__fn __IRFunction, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__expected := __fn.__returnType
	__actual := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__fn.__body, __structs, __callables, ____rune_private_1ed26dbc_compilerFunctionBindings(__fn.__params, __bindings))
	__shouldCheck := __expected != "" && __expected != "Dynamic" && __actual != ""
	return func() []string {
		switch {
		case __shouldCheck == true:
			return ____rune_private_1ed26dbc_checkFunctionReturnType(__fn.__name, __expected, __actual, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_compilerKnownTypes(__file __IRFile) []string {
	__names := ____rune_private_1ed26dbc_compilerBuiltinTypes()
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		func() int { __names = append(__names, __typeDecl.__name); return len(__names) }()
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		func() int { __names = append(__names, __typeDecl.__name); return len(__names) }()
	}
	return __names
}

func ____rune_private_1ed26dbc_compilerBuiltinTypes() []string {
	return []string{"Int", "Int4", "Int8", "Int16", "Int64", "UInt", "UInt8", "UInt16", "UInt64", "Double", "Float", "Bool", "String", "Char", "BigInt", "Byte", "Bytes", "Object", "Dynamic", "Void", "Null", "Error", "Regex", "Symbol", "MacroContext", "Array", "ReadonlyArray", "Tuple", "ReadonlyTuple", "Map", "Set", "Result"}
}

func ____rune_private_1ed26dbc_checkDuplicateDeclarations(__file __IRFile, __errors []string) []string {
	__next := ____rune_private_1ed26dbc_checkDuplicateStructTypeNames(__file.__structs, 0, __errors)
	__next = ____rune_private_1ed26dbc_checkDuplicateEnumTypeNames(__file.__enums, __file.__structs, 0, __next)
	__next = ____rune_private_1ed26dbc_checkDuplicateConstantNames(__file.__constants, 0, __next)
	__next = ____rune_private_1ed26dbc_checkDuplicateFunctionNames(__file.__functions, 0, __next)
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__next = ____rune_private_1ed26dbc_checkDuplicateStructMembers(__typeDecl, __next)
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		__next = ____rune_private_1ed26dbc_checkDuplicateEnumMembers(__typeDecl, __next)
	}
	return __next
}

func ____rune_private_1ed26dbc_checkDuplicateStructTypeNames(__structs []__IRStructType, __index int, __errors []string) []string {
	__done := __index >= len(__structs)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateStructTypeName(__structs, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateStructTypeName(__structs []__IRStructType, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerStructNameAppearsAfter(__structs, __structs[__index].__name, __index+1)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate type \""+__structs[__index].__name+"\"")
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateStructTypeNames(__structs, __index+1, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateEnumTypeNames(__enums []__IREnumType, __structs []__IRStructType, __index int, __errors []string) []string {
	__done := __index >= len(__enums)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateEnumTypeName(__enums, __structs, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateEnumTypeName(__enums []__IREnumType, __structs []__IRStructType, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerEnumNameAppearsAfter(__enums, __enums[__index].__name, __index+1) || ____rune_private_1ed26dbc_compilerStructNameAppearsAfter(__structs, __enums[__index].__name, 0)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate type \""+__enums[__index].__name+"\"")
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateEnumTypeNames(__enums, __structs, __index+1, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateFunctionNames(__functions []__IRFunction, __index int, __errors []string) []string {
	__done := __index >= len(__functions)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateFunctionName(__functions, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateFunctionName(__functions []__IRFunction, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerFunctionNameAppearsBefore(__functions, __functions[__index].__name, __functions[__index].__macro, __index-1)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate function \""+__functions[__index].__name+"\"")
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateFunctionNames(__functions, __index+1, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateConstantNames(__constants []__IRConst, __index int, __errors []string) []string {
	__done := __index >= len(__constants)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateConstantName(__constants, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateConstantName(__constants []__IRConst, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerConstNameAppearsBefore(__constants, __constants[__index].__name, __index-1)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate declaration \""+__constants[__index].__name+"\"")
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateConstantNames(__constants, __index+1, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateStructMembers(__typeDecl __IRStructType, __errors []string) []string {
	__next := ____rune_private_1ed26dbc_checkDuplicateStructFields(__typeDecl.__fields, 0, __errors)
	return ____rune_private_1ed26dbc_checkDuplicateStructMethods(__typeDecl.__name, __typeDecl.__methods, 0, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateStructFields(__fields []__IRField, __index int, __errors []string) []string {
	__done := __index >= len(__fields)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateStructField(__fields, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateStructField(__fields []__IRField, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerFieldNameAppearsBefore(__fields, __fields[__index].__name, __index-1)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate field \""+__fields[__index].__name+"\"")
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateStructFields(__fields, __index+1, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateStructMethods(__typeName string, __methods []__IRFunction, __index int, __errors []string) []string {
	__done := __index >= len(__methods)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateStructMethod(__typeName, __methods, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateStructMethod(__typeName string, __methods []__IRFunction, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerMethodNameAppearsBefore(__methods, __methods[__index].__name, __index-1)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate method "+__typeName+"."+__methods[__index].__name)
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateStructMethods(__typeName, __methods, __index+1, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateEnumMembers(__typeDecl __IREnumType, __errors []string) []string {
	__next := ____rune_private_1ed26dbc_checkDuplicateEnumConstructors(__typeDecl.__name, __typeDecl.__members, 0, __errors)
	return ____rune_private_1ed26dbc_checkDuplicateEnumMethods(__typeDecl.__name, __typeDecl.__methods, 0, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateEnumConstructors(__enumName string, __members []__IREnumMember, __index int, __errors []string) []string {
	__done := __index >= len(__members)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateEnumConstructor(__enumName, __members, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateEnumConstructor(__enumName string, __members []__IREnumMember, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerEnumMemberNameAppearsBefore(__members, __members[__index].__name, __index-1)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate enum member "+__enumName+"."+__members[__index].__name)
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateEnumConstructors(__enumName, __members, __index+1, __next)
}

func ____rune_private_1ed26dbc_checkDuplicateEnumMethods(__enumName string, __methods []__IRFunction, __index int, __errors []string) []string {
	__done := __index >= len(__methods)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateEnumMethod(__enumName, __methods, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateEnumMethod(__enumName string, __methods []__IRFunction, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerMethodNameAppearsBefore(__methods, __methods[__index].__name, __index-1)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate method "+__enumName+"."+__methods[__index].__name)
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateEnumMethods(__enumName, __methods, __index+1, __next)
}

func ____rune_private_1ed26dbc_checkDeclarationTypes(__file __IRFile, __knownTypes []string, __errors []string) []string {
	__next := __errors
	for _, __constant := range __file.__constants {
		_ = __constant
		__next = ____rune_private_1ed26dbc_checkConstDeclarationTypes(__constant, __knownTypes, __next)
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		__next = ____rune_private_1ed26dbc_checkStructDeclarationTypes(__typeDecl, __knownTypes, __next)
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		__next = ____rune_private_1ed26dbc_checkEnumDeclarationTypes(__typeDecl, __knownTypes, __next)
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__next = ____rune_private_1ed26dbc_checkFunctionDeclarationTypes(__fn, __knownTypes, __next)
	}
	return __next
}

func ____rune_private_1ed26dbc_checkConstDeclarationTypes(__constant __IRConst, __knownTypes []string, __errors []string) []string {
	return ____rune_private_1ed26dbc_checkCompilerTypeName(__constant.__typeName, __knownTypes, []string{}, __errors)
}

func ____rune_private_1ed26dbc_checkStructDeclarationTypes(__typeDecl __IRStructType, __knownTypes []string, __errors []string) []string {
	__next := __errors
	for _, __field := range __typeDecl.__fields {
		_ = __field
		__next = ____rune_private_1ed26dbc_checkCompilerTypeName(__field.__typeName, __knownTypes, __typeDecl.__generics, __next)
	}
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__next = ____rune_private_1ed26dbc_checkFunctionDeclarationTypesWithGenerics(__method, __knownTypes, __typeDecl.__generics, __next)
	}
	return __next
}

func ____rune_private_1ed26dbc_checkEnumDeclarationTypes(__typeDecl __IREnumType, __knownTypes []string, __errors []string) []string {
	__next := __errors
	for _, __member := range __typeDecl.__members {
		_ = __member
		func() {
			for _, __param := range __member.__params {
				_ = __param
				__next = ____rune_private_1ed26dbc_checkCompilerTypeName(__param.__typeName, __knownTypes, __typeDecl.__generics, __next)
			}
		}()
	}
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__next = ____rune_private_1ed26dbc_checkFunctionDeclarationTypesWithGenerics(__method, __knownTypes, __typeDecl.__generics, __next)
	}
	return __next
}

func ____rune_private_1ed26dbc_checkFunctionDeclarationTypes(__fn __IRFunction, __knownTypes []string, __errors []string) []string {
	return ____rune_private_1ed26dbc_checkFunctionDeclarationTypesWithGenerics(__fn, __knownTypes, []string{}, __errors)
}

func ____rune_private_1ed26dbc_checkFunctionDeclarationTypesWithGenerics(__fn __IRFunction, __knownTypes []string, __parentGenerics []string, __errors []string) []string {
	__generics := ____rune_private_1ed26dbc_compilerMergeGenerics(__parentGenerics, __fn.__generics)
	__next := ____rune_private_1ed26dbc_checkDuplicateParams(__fn.__params, 0, __errors)
	__next = ____rune_private_1ed26dbc_checkCompilerTypeName(__fn.__returnType, __knownTypes, __generics, __next)
	for _, __param := range __fn.__params {
		_ = __param
		__next = ____rune_private_1ed26dbc_checkCompilerTypeName(__param.__typeName, __knownTypes, __generics, __next)
	}
	return __next
}

func ____rune_private_1ed26dbc_checkDuplicateParams(__params []__IRParam, __index int, __errors []string) []string {
	__done := __index >= len(__params)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkDuplicateParam(__params, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkDuplicateParam(__params []__IRParam, __index int, __errors []string) []string {
	__duplicate := ____rune_private_1ed26dbc_compilerParamNameAppearsBefore(__params, __params[__index].__name, __index-1)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate parameter \""+__params[__index].__name+"\"")
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkDuplicateParams(__params, __index+1, __next)
}

func ____rune_private_1ed26dbc_compilerMergeGenerics(__parentGenerics []string, __functionGenerics []string) []string {
	__out := append([]string{}, __parentGenerics[0:len(__parentGenerics)]...)
	for _, __name := range __functionGenerics {
		_ = __name
		func() int { __out = append(__out, __name); return len(__out) }()
	}
	return __out
}

func ____rune_private_1ed26dbc_checkCompilerTypeName(__typeName string, __knownTypes []string, __generics []string, __errors []string) []string {
	__normalized := ____rune_private_1ed26dbc_compilerNormalizeTypeName(__typeName)
	__shouldSkip := ____rune_private_1ed26dbc_compilerShouldSkipTypeName(__normalized)
	return func() []string {
		switch {
		case __shouldSkip == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkCompilerNamedType(__normalized, __knownTypes, __generics, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerNormalizeTypeName(__typeName string) string {
	__nullable := strings.HasSuffix(__typeName, "?")
	return func() string {
		switch {
		case __nullable == true:
			return ____rune_private_1ed26dbc_compilerNormalizeTypeName(func() string { runes := []rune(__typeName); return string(runes[0 : len([]rune(__typeName))-1]) }())
		default:
			return __typeName
		}
	}()
}

func ____rune_private_1ed26dbc_compilerShouldSkipTypeName(__typeName string) bool {
	return __typeName == "" || (strings.HasPrefix(__typeName, "@") || strings.HasPrefix(__typeName, "(") || strings.HasPrefix(__typeName, "Syntax"))
}

func ____rune_private_1ed26dbc_checkCompilerNamedType(__typeName string, __knownTypes []string, __generics []string, __errors []string) []string {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	__known := ____rune_private_1ed26dbc_compilerContains(__knownTypes, __base) || ____rune_private_1ed26dbc_compilerContains(__generics, __base)
	return func() []string {
		switch {
		case __known == true:
			return ____rune_private_1ed26dbc_checkCompilerTypeArgs(__typeName, __knownTypes, __generics, __errors)
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "unknown type \""+__base+"\"")
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkCompilerTypeArgs(__typeName string, __knownTypes []string, __generics []string, __errors []string) []string {
	__inner := ____rune_private_1ed26dbc_compilerGenericInner(__typeName)
	__empty := __inner == ""
	return func() []string {
		switch {
		case __empty == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkCompilerTypeArgList(func() []string { parts := strings.Split(__inner, ","); return parts }(), __knownTypes, __generics, __errors, 0)
		}
	}()
}

func ____rune_private_1ed26dbc_checkCompilerTypeArgList(__args []string, __knownTypes []string, __generics []string, __errors []string, __index int) []string {
	__done := __index >= len(__args)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkCompilerTypeArgList(__args, __knownTypes, __generics, ____rune_private_1ed26dbc_checkCompilerTypeName(strings.TrimSpace(__args[__index]), __knownTypes, __generics, __errors), __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerGenericInner(__typeName string) string {
	__open := strings.Index(__typeName, "[")
	__complete := __open >= 0 && strings.HasSuffix(__typeName, "]")
	return func() string {
		switch {
		case __complete == true:
			return func() string { runes := []rune(__typeName); return string(runes[__open+1 : len([]rune(__typeName))-1]) }()
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerCallables(__file __IRFile) []__CompilerCallable {
	__callables := append([]__CompilerCallable{}, []__CompilerCallable{____rune_private_1ed26dbc_emptyCompilerCallable()}[0:0]...)
	for _, __fn := range __file.__functions {
		_ = __fn
		func() int {
			if __fn.__macro {
				return 0
			}
			return func() int {
				__callables = append(__callables, ____rune_private_1ed26dbc_compilerFunctionCallable(__fn))
				return len(__callables)
			}()
		}()
	}
	for _, __importDecl := range __file.__tsImports {
		_ = __importDecl
		func() {
			for _, __fn := range __importDecl.__functions {
				_ = __fn
				func() int {
					__callables = append(__callables, ____rune_private_1ed26dbc_compilerCallable(__fn.__name, len(__fn.__params), __fn.__returnType, ____rune_private_1ed26dbc_compilerParamTypeNames(__fn.__params), false, ""))
					return len(__callables)
				}()
			}
		}()
	}
	for _, __typeDecl := range __file.__structs {
		_ = __typeDecl
		func() {
			for _, __method := range __typeDecl.__methods {
				_ = __method
				func() int {
					__callables = append(__callables, ____rune_private_1ed26dbc_compilerMethodCallable(__typeDecl.__name, __method))
					return len(__callables)
				}()
			}
		}()
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		func() {
			for _, __method := range __typeDecl.__methods {
				_ = __method
				func() int {
					__callables = append(__callables, ____rune_private_1ed26dbc_compilerMethodCallable(__typeDecl.__name, __method))
					return len(__callables)
				}()
			}
		}()
	}
	for _, __typeDecl := range __file.__enums {
		_ = __typeDecl
		func() {
			for _, __member := range __typeDecl.__members {
				_ = __member
				func() int {
					__callables = append(__callables, ____rune_private_1ed26dbc_compilerCallable(__member.__name, len(__member.__params), __typeDecl.__name, ____rune_private_1ed26dbc_compilerParamTypeNames(__member.__params), false, ""))
					return len(__callables)
				}()
			}
		}()
	}
	return __callables
}

func ____rune_private_1ed26dbc_compilerFunctionCallable(__fn __IRFunction) __CompilerCallable {
	return ____rune_private_1ed26dbc_compilerCallable(__fn.__name, len(__fn.__params), __fn.__returnType, ____rune_private_1ed26dbc_compilerParamTypeNames(__fn.__params), __fn.__private, __fn.__sourcePath)
}

func ____rune_private_1ed26dbc_compilerMethodCallable(__typeName string, __method __IRFunction) __CompilerCallable {
	return ____rune_private_1ed26dbc_compilerCallable(____rune_private_1ed26dbc_compilerMethodCallableName(__typeName, __method), len(__method.__params), __method.__returnType, ____rune_private_1ed26dbc_compilerParamTypeNames(__method.__params), __method.__private, __method.__sourcePath)
}

func ____rune_private_1ed26dbc_compilerMethodCallableName(__typeName string, __method __IRFunction) string {
	return func() string {
		switch {
		case __method.__static == true:
			return ____rune_private_1ed26dbc_compilerStaticMethodName(__typeName, __method.__name)
		default:
			return ____rune_private_1ed26dbc_compilerInstanceMethodName(__typeName, __method.__name)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerStaticMethodName(__typeName string, __methodName string) string {
	return __typeName + "::" + __methodName
}

func ____rune_private_1ed26dbc_compilerInstanceMethodName(__typeName string, __methodName string) string {
	return __typeName + "." + __methodName
}

func ____rune_private_1ed26dbc_checkFunctionErrors(__fn __IRFunction, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __baseBindings []__CompilerTypeBinding) []string {
	__bindings := ____rune_private_1ed26dbc_compilerFunctionBindings(__fn.__params, __baseBindings)
	return ____rune_private_1ed26dbc_checkFunctionReturn(__fn, __structs, __callables, ____rune_private_1ed26dbc_checkExpr(__fn.__body, __structs, __callables, __errors, __bindings), __bindings)
}

func ____rune_private_1ed26dbc_checkMethodErrors(__typeName string, __method __IRFunction, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __baseBindings []__CompilerTypeBinding) []string {
	return ____rune_private_1ed26dbc_checkFunctionErrors(__method, __structs, __callables, __errors, ____rune_private_1ed26dbc_compilerMethodBaseBindings(__typeName, __method, __baseBindings))
}

func ____rune_private_1ed26dbc_compilerMethodBaseBindings(__typeName string, __method __IRFunction, __baseBindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	return func() []__CompilerTypeBinding {
		switch {
		case __method.__static == true:
			return __baseBindings
		default:
			return ____rune_private_1ed26dbc_addCompilerTypeBinding(__baseBindings, "this", __typeName)
		}
	}()
}

func ____rune_private_1ed26dbc_checkExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_1ed26dbc_checkBlockExpr(__expr.__children, 0, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Let:
			return ____rune_private_1ed26dbc_checkLetExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Struct:
			return ____rune_private_1ed26dbc_checkStructExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Ternary:
			return ____rune_private_1ed26dbc_checkTernaryExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Unary:
			return ____rune_private_1ed26dbc_checkUnaryExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Postfix:
			return ____rune_private_1ed26dbc_checkPostfixExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_1ed26dbc_checkBinaryExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Array:
			return ____rune_private_1ed26dbc_checkArrayExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Map:
			return ____rune_private_1ed26dbc_checkMapExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Index:
			return ____rune_private_1ed26dbc_checkIndexExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_1ed26dbc_checkCallExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_1ed26dbc_checkSelectorValueExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Lambda:
			return ____rune_private_1ed26dbc_checkLambdaExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Assign:
			return ____rune_private_1ed26dbc_checkAssignExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_PatternBlock:
			return ____rune_private_1ed26dbc_checkPatternBlockExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Match:
			return ____rune_private_1ed26dbc_checkMatchExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_1ed26dbc_checkObjectDestructureExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Unwrap:
			return ____rune_private_1ed26dbc_checkUnwrapExpr(__expr, __structs, __callables, __errors, __bindings)
		case __expr.__kind == __ExprKind_Identifier:
			return ____rune_private_1ed26dbc_checkIdentifierExpr(__expr, __callables, __bindings, __errors)
		case __expr.__kind == __ExprKind_This:
			return ____rune_private_1ed26dbc_checkThisExpr(__bindings, __errors)
		default:
			return ____rune_private_1ed26dbc_checkExprDefault(__expr, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkIdentifierExpr(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case ____rune_private_1ed26dbc_compilerIdentifierDefined(__expr.__name, __callables, __bindings) == true:
			return __errors
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "undefined name \""+__expr.__name+"\"")
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerIdentifierDefined(__name string, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) bool {
	return func() bool {
		switch {
		case ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, __name, 0).__name == "":
			return ____rune_private_1ed26dbc_findCompilerCallable(__callables, __name, 0).__name != ""
		default:
			return true
		}
	}()
}

func ____rune_private_1ed26dbc_checkThisExpr(__bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, "this", 0).__typeName == "":
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "implicit this selector can only be used inside a method")
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkCallExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	return ____rune_private_1ed26dbc_checkCallArgExprs(__expr.__children, 1, __structs, __callables, ____rune_private_1ed26dbc_checkExprCall(__expr, __structs, __callables, __errors, __bindings), __bindings)
}

func ____rune_private_1ed26dbc_checkCallArgExprs(__args []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__done := __index >= len(__args)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkCallArgExprs(__args, __index+1, __structs, __callables, ____rune_private_1ed26dbc_checkExpr(__args[__index], __structs, __callables, __errors, __bindings), __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkSelectorValueExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	return ____rune_private_1ed26dbc_checkSelectorReceiverValueExpr(__expr, __structs, __callables, __bindings, ____rune_private_1ed26dbc_checkSelectorExpr(__expr, __structs, __callables, __bindings, __errors))
}

func ____rune_private_1ed26dbc_checkSelectorReceiverValueExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__hasReceiver := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __hasReceiver == true:
			return ____rune_private_1ed26dbc_checkSelectorReceiverValue(__expr, __expr.__children[0], __structs, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkSelectorReceiverValue(__expr __IRExpr, __receiver __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__staticTypeReceiver := ____rune_private_1ed26dbc_compilerStaticSelectorReceiver(__expr, __receiver)
	__skipReceiver := __staticTypeReceiver || __receiver.__kind == __ExprKind_At
	return func() []string {
		switch {
		case __skipReceiver == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkExpr(__receiver, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerStaticSelectorReceiver(__expr __IRExpr, __receiver __IRExpr) bool {
	return func() bool {
		switch {
		case __receiver.__kind == __ExprKind_Identifier:
			return __expr.__op == "::" || ____rune_private_1ed26dbc_compilerLooksLikeTypeName(__receiver.__name)
		default:
			return false
		}
	}()
}

func ____rune_private_1ed26dbc_compilerLooksLikeTypeName(__name string) bool {
	__empty := len([]rune(__name)) == 0
	return func() bool {
		switch {
		case __empty == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerUppercaseChar([]rune(__name)[0])
		}
	}()
}

func ____rune_private_1ed26dbc_compilerUppercaseChar(__ch rune) bool {
	return __ch >= 'A' && __ch <= 'Z'
}

func ____rune_private_1ed26dbc_checkLambdaExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__hasBody := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __hasBody == true:
			return ____rune_private_1ed26dbc_checkExpr(__expr.__children[0], __structs, __callables, __errors, ____rune_private_1ed26dbc_compilerFunctionBindings(__expr.__params, __bindings))
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkAssignExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__named := __expr.__name != ""
	return func() []string {
		switch {
		case __named == true:
			return ____rune_private_1ed26dbc_checkNamedAssignExpr(__expr, __structs, __callables, __errors, __bindings)
		default:
			return ____rune_private_1ed26dbc_checkAssignExpressionExpr(__expr, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkNamedAssignExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__complete := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkIdentifierAssignExpr(__expr, __expr.__children[0], __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkAssignExpressionExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__complete := len(__expr.__children) >= 2
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkAssignTargetExpr(__expr.__children[0], __expr.__children[1], __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkAssignTargetExpr(__target __IRExpr, __value __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case __target.__kind == __ExprKind_Identifier:
			return ____rune_private_1ed26dbc_checkIdentifierAssignExpr(__target, __value, __structs, __callables, __errors, __bindings)
		case __target.__kind == __ExprKind_Index:
			return ____rune_private_1ed26dbc_checkIndexAssignExpr(__target, __value, __structs, __callables, __errors, __bindings)
		default:
			return ____rune_private_1ed26dbc_checkTargetAssignExpr(__target, __value, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkIdentifierAssignExpr(__target __IRExpr, __value __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__binding := ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, __target.__name, 0)
	__checked := ____rune_private_1ed26dbc_checkExpr(__value, __structs, __callables, __errors, __bindings)
	return func() []string {
		switch {
		case __binding.__name == "":
			return func() []string {
				out := []string{}
				out = append(out, __checked...)
				out = append(out, "cannot assign undefined name \""+__target.__name+"\"")
				return out
			}()
		default:
			return ____rune_private_1ed26dbc_checkAssignmentType(__binding.__typeName, ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__value, __structs, __callables, __bindings), __checked)
		}
	}()
}

func ____rune_private_1ed26dbc_checkIndexAssignExpr(__target __IRExpr, __value __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkIndexExpr(__target, __structs, __callables, ____rune_private_1ed26dbc_checkExpr(__value, __structs, __callables, __errors, __bindings), __bindings)
	return ____rune_private_1ed26dbc_checkAssignmentType(____rune_private_1ed26dbc_inferCompilerIndexType(__target, __callables, __bindings), ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__value, __structs, __callables, __bindings), __checked)
}

func ____rune_private_1ed26dbc_checkTargetAssignExpr(__target __IRExpr, __value __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExpr(__value, __structs, __callables, ____rune_private_1ed26dbc_checkExpr(__target, __structs, __callables, __errors, __bindings), __bindings)
	return ____rune_private_1ed26dbc_checkAssignmentType(____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__target, __structs, __callables, __bindings), ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__value, __structs, __callables, __bindings), __checked)
}

func ____rune_private_1ed26dbc_checkAssignmentType(__expected string, __actual string, __errors []string) []string {
	__mismatch := ____rune_private_1ed26dbc_compilerShouldCheckArgType(__expected, __actual) && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "assignment has type "+__actual+", expected "+__expected)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkPatternBlockExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	return ____rune_private_1ed26dbc_checkPatternBlockBranches(__expr.__children, 0, __structs, __callables, __errors, __bindings, "")
}

func ____rune_private_1ed26dbc_checkMatchExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__hasSubject := len(__expr.__children) > 0
	__checked := func() []string {
		switch {
		case __hasSubject == true:
			return ____rune_private_1ed26dbc_checkExpr(__expr.__children[0], __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkPatternBlockBranches(__expr.__children, 1, __structs, __callables, __checked, __bindings, "")
}

func ____rune_private_1ed26dbc_checkPatternBlockBranches(__branches []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding, __expected string) []string {
	__done := __index >= len(__branches)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkPatternBlockBranch(__branches, __index, __structs, __callables, __errors, __bindings, __expected)
		}
	}()
}

func ____rune_private_1ed26dbc_checkPatternBlockBranch(__branches []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding, __expected string) []string {
	__branch := __branches[__index]
	__complete := len(__branch.__children) >= 2
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkPatternBlockBranchValue(__branches, __index, __branch.__children[1], __structs, __callables, __errors, __bindings, __expected)
		default:
			return ____rune_private_1ed26dbc_checkPatternBlockBranches(__branches, __index+1, __structs, __callables, __errors, __bindings, __expected)
		}
	}()
}

func ____rune_private_1ed26dbc_checkPatternBlockBranchValue(__branches []__IRExpr, __index int, __value __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding, __expected string) []string {
	__checked := ____rune_private_1ed26dbc_checkExpr(__value, __structs, __callables, __errors, __bindings)
	__actual := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__value, __structs, __callables, __bindings)
	__nextExpected := ____rune_private_1ed26dbc_compilerNextPatternBranchType(__expected, __actual)
	__nextErrors := ____rune_private_1ed26dbc_checkPatternBranchTypeError(__expected, __actual, __checked)
	return ____rune_private_1ed26dbc_checkPatternBlockBranches(__branches, __index+1, __structs, __callables, __nextErrors, __bindings, __nextExpected)
}

func ____rune_private_1ed26dbc_compilerNextPatternBranchType(__expected string, __actual string) string {
	return func() string {
		switch {
		case __expected == "":
			return __actual
		default:
			return __expected
		}
	}()
}

func ____rune_private_1ed26dbc_checkPatternBranchTypeError(__expected string, __actual string, __errors []string) []string {
	__mismatch := ____rune_private_1ed26dbc_compilerShouldCheckArgType(__expected, __actual) && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "pattern branch returns "+__actual+", expected "+__expected)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkObjectDestructureExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__hasValue := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __hasValue == true:
			return ____rune_private_1ed26dbc_checkObjectDestructureValue(__expr, __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkObjectDestructureValue(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__value := __expr.__children[0]
	__checked := ____rune_private_1ed26dbc_checkExpr(__value, __structs, __callables, __errors, __bindings)
	__sourceType := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__value, __structs, __callables, __bindings)
	return ____rune_private_1ed26dbc_checkObjectDestructureSourceType(__expr.__params, __sourceType, __structs, __checked)
}

func ____rune_private_1ed26dbc_checkObjectDestructureSourceType(__params []__IRParam, __sourceType string, __structs []__IRStructType, __errors []string) []string {
	return func() []string {
		switch {
		case __sourceType == "":
			return __errors
		default:
			return func() []string {
				__typeDecl := ____rune_private_1ed26dbc_findCompilerStruct(__structs, ____rune_private_1ed26dbc_compilerTypeBase(__sourceType), 0)
				__found := __typeDecl.__name != ""
				return func() []string {
					switch {
					case __found == true:
						return ____rune_private_1ed26dbc_checkObjectDestructureFields(__params, 0, __sourceType, __typeDecl.__fields, __errors)
					default:
						return func() []string {
							out := []string{}
							out = append(out, __errors...)
							out = append(out, "type "+__sourceType+" has no fields")
							return out
						}()
					}
				}()
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkObjectDestructureFields(__params []__IRParam, __index int, __sourceType string, __fields []__IRField, __errors []string) []string {
	__done := __index >= len(__params)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkObjectDestructureField(__params, __index, __sourceType, __fields, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkObjectDestructureField(__params []__IRParam, __index int, __sourceType string, __fields []__IRField, __errors []string) []string {
	__fieldName := __params[__index].__typeName
	__duplicate := ____rune_private_1ed26dbc_compilerDestructureFieldAppearsBefore(__params, __fieldName, __index-1)
	__field := ____rune_private_1ed26dbc_findCompilerStructField(__fields, __fieldName, 0)
	__next := func() []string {
		switch {
		case __duplicate == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "duplicate destructured field \""+__fieldName+"\"")
				return out
			}()
		default:
			return func() []string {
				switch {
				case __field.__name == "":
					return func() []string {
						out := []string{}
						out = append(out, __errors...)
						out = append(out, "type "+__sourceType+" has no field \""+__fieldName+"\"")
						return out
					}()
				default:
					return __errors
				}
			}()
		}
	}()
	return ____rune_private_1ed26dbc_checkObjectDestructureFields(__params, __index+1, __sourceType, __fields, __next)
}

func ____rune_private_1ed26dbc_compilerDestructureFieldAppearsBefore(__params []__IRParam, __name string, __index int) bool {
	__done := __index < 0
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerDestructureFieldAppearsBeforeAt(__params, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerDestructureFieldAppearsBeforeAt(__params []__IRParam, __name string, __index int) bool {
	__matched := __params[__index].__typeName == __name
	return func() bool {
		switch {
		case __matched == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerDestructureFieldAppearsBefore(__params, __name, __index-1)
		}
	}()
}

func ____rune_private_1ed26dbc_checkUnwrapExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__complete := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkUnwrapSourceExpr(__expr.__children[0], __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkUnwrapSourceExpr(__source __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExpr(__source, __structs, __callables, __errors, __bindings)
	__sourceType := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__source, __structs, __callables, __bindings)
	return ____rune_private_1ed26dbc_checkUnwrapSourceType(__sourceType, __checked)
}

func ____rune_private_1ed26dbc_checkUnwrapSourceType(__sourceType string, __errors []string) []string {
	return func() []string {
		switch {
		case __sourceType == "":
			return __errors
		default:
			return func() []string {
				switch {
				case ____rune_private_1ed26dbc_compilerResultOkType(__sourceType) == "":
					return func() []string {
						out := []string{}
						out = append(out, __errors...)
						out = append(out, "operator '?' expects Result, got "+__sourceType)
						return out
					}()
				default:
					return __errors
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkExprDefault(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__next := ____rune_private_1ed26dbc_checkExprCall(__expr, __structs, __callables, __errors, __bindings)
	for _, __child := range __expr.__children {
		_ = __child
		__next = ____rune_private_1ed26dbc_checkExpr(__child, __structs, __callables, __next, __bindings)
	}
	return __next
}

func ____rune_private_1ed26dbc_checkTernaryExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExprDefault(__expr, __structs, __callables, __errors, __bindings)
	__withCondition := ____rune_private_1ed26dbc_checkTernaryCondition(__expr, __callables, __bindings, __checked)
	return ____rune_private_1ed26dbc_checkTernaryBranches(__expr, __callables, __bindings, __withCondition)
}

func ____rune_private_1ed26dbc_checkTernaryCondition(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__complete := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkTernaryConditionType(__expr.__children[0], __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkTernaryConditionType(__condition __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprType(__condition, __callables, __bindings)
	__mismatch := __actual != "" && __actual != "Bool"
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "ternary condition expects Bool, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkTernaryBranches(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__complete := len(__expr.__children) >= 3
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkTernaryBranchTypes(__expr.__children[1], __expr.__children[2], __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkTernaryBranchTypes(__consequence __IRExpr, __alternative __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__left := ____rune_private_1ed26dbc_inferCompilerExprType(__consequence, __callables, __bindings)
	__right := ____rune_private_1ed26dbc_inferCompilerExprType(__alternative, __callables, __bindings)
	__shouldCheck := __left != "" && __right != ""
	__mismatch := __shouldCheck && ____rune_private_1ed26dbc_inferCompilerCommonType(__left, __right) == ""
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "ternary branches return "+__left+" and "+__right)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkUnaryExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExprDefault(__expr, __structs, __callables, __errors, __bindings)
	return func() []string {
		switch {
		case __expr.__op == "!":
			return ____rune_private_1ed26dbc_checkBoolOperand("!", __expr.__children[0], __callables, __bindings, __checked)
		case __expr.__op == "-":
			return ____rune_private_1ed26dbc_checkNumericUnaryOperand(__expr, __callables, __bindings, __checked)
		case __expr.__op == "~":
			return ____rune_private_1ed26dbc_checkBitwiseUnaryOperand(__expr, __callables, __bindings, __checked)
		default:
			return __checked
		}
	}()
}

func ____rune_private_1ed26dbc_checkPostfixExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExprDefault(__expr, __structs, __callables, __errors, __bindings)
	return func() []string {
		switch {
		case __expr.__op == "++":
			return ____rune_private_1ed26dbc_checkPostfixNumericOperand(__expr, __callables, __bindings, __checked)
		default:
			return __checked
		}
	}()
}

func ____rune_private_1ed26dbc_checkBinaryExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExprDefault(__expr, __structs, __callables, __errors, __bindings)
	__boolOp := __expr.__op == "&&" || __expr.__op == "||"
	return func() []string {
		switch {
		case __boolOp == true:
			return ____rune_private_1ed26dbc_checkBinaryBoolOperands(__expr, __callables, __bindings, __checked)
		default:
			return ____rune_private_1ed26dbc_checkTypedBinaryExpr(__expr, __callables, __bindings, __checked)
		}
	}()
}

func ____rune_private_1ed26dbc_checkTypedBinaryExpr(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__checked := ____rune_private_1ed26dbc_checkEqualityComparisonExpr(__expr, __callables, __bindings, __errors)
	__checked = ____rune_private_1ed26dbc_checkOrderedComparisonExpr(__expr, __callables, __bindings, __checked)
	__checked = ____rune_private_1ed26dbc_checkNumericBinaryExpr(__expr, __callables, __bindings, __checked)
	return ____rune_private_1ed26dbc_checkBitwiseBinaryExpr(__expr, __callables, __bindings, __checked)
}

func ____rune_private_1ed26dbc_checkBinaryBoolOperands(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__complete := len(__expr.__children) >= 2
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkBoolOperand(__expr.__op, __expr.__children[1], __callables, __bindings, ____rune_private_1ed26dbc_checkBoolOperand(__expr.__op, __expr.__children[0], __callables, __bindings, __errors))
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkBoolOperand(__op string, __operand __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprType(__operand, __callables, __bindings)
	__mismatch := __actual != "" && __actual != "Bool"
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "operator '"+__op+"' expects Bool, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkArrayExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExprDefault(__expr, __structs, __callables, __errors, __bindings)
	return ____rune_private_1ed26dbc_checkArrayElementTypes(__expr.__children, __callables, __bindings, __checked, 0, "")
}

func ____rune_private_1ed26dbc_checkArrayElementTypes(__elements []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string, __index int, __expected string) []string {
	__done := __index >= len(__elements)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkArrayElementType(__elements, __callables, __bindings, __errors, __index, __expected)
		}
	}()
}

func ____rune_private_1ed26dbc_checkArrayElementType(__elements []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string, __index int, __expected string) []string {
	__elem := __elements[__index]
	__actual := ____rune_private_1ed26dbc_inferCompilerArrayLiteralElementType(__elem, __callables, __bindings)
	__nextExpected := ____rune_private_1ed26dbc_compilerNextArrayElementType(__expected, __actual)
	__nextErrors := ____rune_private_1ed26dbc_checkArrayElementTypeError(__elem, __expected, __actual, __callables, __bindings, __errors)
	return ____rune_private_1ed26dbc_checkArrayElementTypes(__elements, __callables, __bindings, __nextErrors, __index+1, __nextExpected)
}

func ____rune_private_1ed26dbc_compilerNextArrayElementType(__expected string, __actual string) string {
	return func() string {
		switch {
		case __expected == "":
			return __actual
		default:
			return __expected
		}
	}()
}

func ____rune_private_1ed26dbc_checkArrayElementTypeError(__elem __IRExpr, __expected string, __actual string, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__spreadErrors := ____rune_private_1ed26dbc_checkArraySpreadElementType(__elem, __callables, __bindings, __errors)
	__mismatch := __expected != "" && __actual != "" && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __spreadErrors...)
				out = append(out, "array element has type "+__actual+", expected "+__expected)
				return out
			}()
		default:
			return __spreadErrors
		}
	}()
}

func ____rune_private_1ed26dbc_checkArraySpreadElementType(__elem __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case __elem.__kind == __ExprKind_Spread:
			return ____rune_private_1ed26dbc_checkArraySpreadReceiverType(__elem, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkArraySpreadReceiverType(__elem __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprType(__elem.__children[0], __callables, __bindings)
	__mismatch := __actual != "" && ____rune_private_1ed26dbc_compilerArrayElementType(__actual) == ""
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "spread expects Array, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkMapExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExprDefault(__expr, __structs, __callables, __errors, __bindings)
	return ____rune_private_1ed26dbc_checkMapEntryTypes(__expr.__children, __callables, __bindings, __checked, 0, "", "")
}

func ____rune_private_1ed26dbc_checkMapEntryTypes(__entries []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string, __index int, __expectedKey string, __expectedValue string) []string {
	__done := __index >= len(__entries)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkMapEntryType(__entries, __callables, __bindings, __errors, __index, __expectedKey, __expectedValue)
		}
	}()
}

func ____rune_private_1ed26dbc_checkMapEntryType(__entries []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string, __index int, __expectedKey string, __expectedValue string) []string {
	__entry := __entries[__index]
	__key := ____rune_private_1ed26dbc_inferCompilerMapEntryKeyType(__entry, __callables, __bindings)
	__value := ____rune_private_1ed26dbc_inferCompilerMapEntryValueType(__entry, __callables, __bindings)
	__nextKey := ____rune_private_1ed26dbc_compilerNextMapEntryType(__expectedKey, __key)
	__nextValue := ____rune_private_1ed26dbc_compilerNextMapEntryType(__expectedValue, __value)
	__checked := ____rune_private_1ed26dbc_checkMapEntryTypeErrors(__entry, __expectedKey, __key, __expectedValue, __value, __errors)
	return ____rune_private_1ed26dbc_checkMapEntryTypes(__entries, __callables, __bindings, __checked, __index+1, __nextKey, __nextValue)
}

func ____rune_private_1ed26dbc_compilerNextMapEntryType(__expected string, __actual string) string {
	return func() string {
		switch {
		case __expected == "":
			return __actual
		default:
			return __expected
		}
	}()
}

func ____rune_private_1ed26dbc_checkMapEntryTypeErrors(__entry __IRExpr, __expectedKey string, __key string, __expectedValue string, __value string, __errors []string) []string {
	__checked := ____rune_private_1ed26dbc_checkMapEntryKeyTypeError(__expectedKey, __key, __errors)
	return ____rune_private_1ed26dbc_checkMapEntryValueTypeError(__expectedValue, __value, __checked)
}

func ____rune_private_1ed26dbc_checkMapEntryKeyTypeError(__expected string, __actual string, __errors []string) []string {
	__checked := ____rune_private_1ed26dbc_checkMapKeySupported(__actual, __errors)
	__mismatch := __expected != "" && __actual != "" && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __checked...)
				out = append(out, "map key has type "+__actual+", expected "+__expected)
				return out
			}()
		default:
			return __checked
		}
	}()
}

func ____rune_private_1ed26dbc_checkMapKeySupported(__actual string, __errors []string) []string {
	__invalid := __actual != "" && ____rune_private_1ed26dbc_compilerSupportedMapKeyType(__actual) == false
	return func() []string {
		switch {
		case __invalid == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "map key type "+__actual+" is not supported")
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkMapEntryValueTypeError(__expected string, __actual string, __errors []string) []string {
	__mismatch := __expected != "" && __actual != "" && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "map value has type "+__actual+", expected "+__expected)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkIndexExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExprDefault(__expr, __structs, __callables, __errors, __bindings)
	__complete := len(__expr.__children) >= 2
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkIndexExprTypes(__expr, __callables, __bindings, __checked)
		default:
			return __checked
		}
	}()
}

func ____rune_private_1ed26dbc_checkIndexExprTypes(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__receiver := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
	__index := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[1], __callables, __bindings)
	return ____rune_private_1ed26dbc_checkIndexReceiverType(__receiver, __index, __expr.__children[1], __errors)
}

func ____rune_private_1ed26dbc_checkIndexReceiverType(__receiver string, __index string, __indexExpr __IRExpr, __errors []string) []string {
	__arrayElem := ____rune_private_1ed26dbc_compilerArrayElementType(__receiver)
	__mapKey := ____rune_private_1ed26dbc_compilerMapKeyType(__receiver)
	__tuple := ____rune_private_1ed26dbc_compilerTupleElementTypes(__receiver)
	return func() []string {
		switch {
		case __arrayElem == "":
			return ____rune_private_1ed26dbc_checkNonArrayIndexReceiver(__receiver, __mapKey, __tuple, __index, __indexExpr, __errors)
		default:
			return ____rune_private_1ed26dbc_checkArrayIndexType(__index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkNonArrayIndexReceiver(__receiver string, __mapKey string, __tuple []string, __index string, __indexExpr __IRExpr, __errors []string) []string {
	return func() []string {
		switch {
		case __mapKey == "":
			return ____rune_private_1ed26dbc_checkTupleOrUnknownIndexReceiver(__receiver, __tuple, __index, __indexExpr, __errors)
		default:
			return ____rune_private_1ed26dbc_checkMapIndexType(__mapKey, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkTupleOrUnknownIndexReceiver(__receiver string, __tuple []string, __index string, __indexExpr __IRExpr, __errors []string) []string {
	return func() []string {
		switch {
		case len(__tuple) > 0 == true:
			return ____rune_private_1ed26dbc_checkTupleIndexType(__tuple, __index, __indexExpr, __errors)
		default:
			return func() []string {
				switch {
				case __receiver == "":
					return __errors
				default:
					return func() []string {
						out := []string{}
						out = append(out, __errors...)
						out = append(out, "type "+__receiver+" is not indexable")
						return out
					}()
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkArrayIndexType(__index string, __errors []string) []string {
	__mismatch := __index != "" && __index != "Int"
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "array index expects Int, got "+__index)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkMapIndexType(__expected string, __actual string, __errors []string) []string {
	__mismatch := __expected != "" && __actual != "" && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "map index has type "+__actual+", expected "+__expected)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkTupleIndexType(__tuple []string, __index string, __indexExpr __IRExpr, __errors []string) []string {
	__mismatch := __index != "" && __index != "Int"
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "tuple index expects an integer literal")
				return out
			}()
		default:
			return ____rune_private_1ed26dbc_checkTupleIndexLiteral(__tuple, __indexExpr, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkTupleIndexLiteral(__tuple []string, __indexExpr __IRExpr, __errors []string) []string {
	return func() []string {
		switch {
		case __indexExpr.__kind == __ExprKind_Int:
			return ____rune_private_1ed26dbc_checkTupleIndexRange(__tuple, ____rune_private_1ed26dbc_compilerParseIntText(__indexExpr.__value), __errors)
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "tuple index expects an integer literal")
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkTupleIndexRange(__tuple []string, __index int, __errors []string) []string {
	__outOfRange := __index < 0 || __index >= len(__tuple)
	return func() []string {
		switch {
		case __outOfRange == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "tuple index "+__compilerIntToString(__index)+" out of range")
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkEqualityComparisonExpr(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case ____rune_private_1ed26dbc_compilerEqualityComparisonOp(__expr.__op) == true:
			return ____rune_private_1ed26dbc_checkEqualityComparisonOperands(__expr, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkEqualityComparisonOperands(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__complete := len(__expr.__children) >= 2
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkEqualityComparisonTypes(__expr, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkEqualityComparisonTypes(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__left := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
	__right := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[1], __callables, __bindings)
	__mismatch := __left != "" && __right != "" && ____rune_private_1ed26dbc_compilerTypesComparable(__left, __right) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "cannot compare "+__left+" and "+__right)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkOrderedComparisonExpr(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case ____rune_private_1ed26dbc_compilerOrderedComparisonOp(__expr.__op) == true:
			return ____rune_private_1ed26dbc_checkOrderedComparisonOperands(__expr, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkOrderedComparisonOperands(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__complete := len(__expr.__children) >= 2
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkOrderedComparisonMatch(__expr, __callables, __bindings, ____rune_private_1ed26dbc_checkOrderedComparisonOperand(__expr.__children[1], __callables, __bindings, ____rune_private_1ed26dbc_checkOrderedComparisonOperand(__expr.__children[0], __callables, __bindings, __errors)))
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkOrderedComparisonOperand(__operand __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprType(__operand, __callables, __bindings)
	__mismatch := __actual != "" && ____rune_private_1ed26dbc_compilerOrderedComparisonType(__actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "ordered comparison expects a numeric type or String, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkOrderedComparisonMatch(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__left := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
	__right := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[1], __callables, __bindings)
	__mismatch := __left != "" && __right != "" && __left != __right
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "ordered comparison requires matching types, got "+__left+" and "+__right)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_compilerOrderedComparisonOp(__op string) bool {
	return __op == "<" || __op == "<=" || __op == ">" || __op == ">="
}

func ____rune_private_1ed26dbc_compilerEqualityComparisonOp(__op string) bool {
	return __op == "==" || __op == "!="
}

func ____rune_private_1ed26dbc_compilerOrderedComparisonType(__typeName string) bool {
	__numeric := ____rune_private_1ed26dbc_compilerNumericType(__typeName)
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	return __numeric || (__base == "String" || __base == "Char")
}

func ____rune_private_1ed26dbc_checkNumericUnaryOperand(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
	__mismatch := __actual != "" && ____rune_private_1ed26dbc_compilerNumericType(__actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "operator '-' expects a numeric type, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkPostfixNumericOperand(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__complete := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkPostfixNumericOperandType(__expr.__op, ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings), __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkPostfixNumericOperandType(__op string, __actual string, __errors []string) []string {
	__mismatch := __actual != "" && ____rune_private_1ed26dbc_compilerNumericType(__actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "operator '"+__op+"' expects a numeric type, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkBitwiseUnaryOperand(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
	__mismatch := __actual != "" && ____rune_private_1ed26dbc_compilerBitwiseType(__actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "operator '~' expects an integer type, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkNumericBinaryExpr(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case ____rune_private_1ed26dbc_compilerArithmeticOp(__expr.__op) == true:
			return ____rune_private_1ed26dbc_checkArithmeticOperands(__expr, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkBitwiseBinaryExpr(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case ____rune_private_1ed26dbc_compilerBitwiseOp(__expr.__op) == true:
			return ____rune_private_1ed26dbc_checkBitwiseOperands(__expr, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkArithmeticOperands(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__complete := len(__expr.__children) >= 2
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkArithmeticOperandTypes(__expr, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkBitwiseOperands(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__complete := len(__expr.__children) >= 2
	return func() []string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_checkBitwiseOperandTypes(__expr, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkArithmeticOperandTypes(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__left := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
	__right := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[1], __callables, __bindings)
	return func() []string {
		switch {
		case __expr.__op == "+":
			return ____rune_private_1ed26dbc_checkPlusOperandTypes(__left, __right, __errors)
		default:
			return ____rune_private_1ed26dbc_checkNumericOperandTypes(__expr.__op, __left, __right, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkBitwiseOperandTypes(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__left := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
	__right := ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[1], __callables, __bindings)
	__checked := ____rune_private_1ed26dbc_checkBitwiseOperand(__right, ____rune_private_1ed26dbc_checkBitwiseOperand(__left, __errors))
	__withUnsigned := ____rune_private_1ed26dbc_checkUnsignedShiftLeftOperand(__expr.__op, __left, __checked)
	return ____rune_private_1ed26dbc_checkBitwiseOperandMatch(__left, __right, __withUnsigned)
}

func ____rune_private_1ed26dbc_checkPlusOperandTypes(__left string, __right string, __errors []string) []string {
	__stringConcat := __left == "String" || __right == "String"
	return func() []string {
		switch {
		case __stringConcat == true:
			return ____rune_private_1ed26dbc_checkStringConcatOperandTypes(__left, __right, __errors)
		default:
			return ____rune_private_1ed26dbc_checkNumericOperandTypes("+", __left, __right, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkStringConcatOperandTypes(__left string, __right string, __errors []string) []string {
	__next := ____rune_private_1ed26dbc_checkStringConcatOperand(__left, __errors)
	return ____rune_private_1ed26dbc_checkStringConcatOperand(__right, __next)
}

func ____rune_private_1ed26dbc_checkStringConcatOperand(__actual string, __errors []string) []string {
	__mismatch := __actual != "" && __actual != "String"
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "string concatenation expects String, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkBitwiseOperand(__actual string, __errors []string) []string {
	__mismatch := __actual != "" && ____rune_private_1ed26dbc_compilerBitwiseType(__actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "bitwise operator expects integer operands, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkUnsignedShiftLeftOperand(__op string, __left string, __errors []string) []string {
	__invalid := __op == ">>>" && __left != "" && ____rune_private_1ed26dbc_compilerUnsignedIntegerType(__left) == false
	return func() []string {
		switch {
		case __invalid == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "operator '>>>' expects an unsigned integer left operand, got "+__left)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkBitwiseOperandMatch(__left string, __right string, __errors []string) []string {
	__shouldCheck := __left != "" && __right != "" && (____rune_private_1ed26dbc_compilerBitwiseType(__left) && ____rune_private_1ed26dbc_compilerBitwiseType(__right))
	__mismatch := __shouldCheck && __left != __right
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "bitwise operator requires matching integer types, got "+__left+" and "+__right)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkNumericOperandTypes(__op string, __left string, __right string, __errors []string) []string {
	__checked := ____rune_private_1ed26dbc_checkNumericOperand(__right, ____rune_private_1ed26dbc_checkNumericOperand(__left, __errors))
	return ____rune_private_1ed26dbc_checkNumericOperandMatch(__op, __left, __right, __checked)
}

func ____rune_private_1ed26dbc_checkNumericOperand(__actual string, __errors []string) []string {
	__mismatch := __actual != "" && ____rune_private_1ed26dbc_compilerNumericType(__actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "arithmetic expects numeric operands, got "+__actual)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkNumericOperandMatch(__op string, __left string, __right string, __errors []string) []string {
	__shouldCheck := __left != "" && __right != "" && (____rune_private_1ed26dbc_compilerNumericType(__left) && ____rune_private_1ed26dbc_compilerNumericType(__right))
	__mismatch := __shouldCheck && __left != __right
	__withMatch := func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "arithmetic requires matching numeric types, got "+__left+" and "+__right)
				return out
			}()
		default:
			return __errors
		}
	}()
	return ____rune_private_1ed26dbc_checkModuloOperand(__op, __left, __right, __withMatch)
}

func ____rune_private_1ed26dbc_checkModuloOperand(__op string, __left string, __right string, __errors []string) []string {
	__invalid := __op == "%" && (__left == "Double" || __left == "Float" || (__right == "Double" || __right == "Float"))
	return func() []string {
		switch {
		case __invalid == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "operator '%' expects integer operands, got "+__left)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_compilerArithmeticOp(__op string) bool {
	return __op == "+" || __op == "-" || __op == "*" || (__op == "/" || __op == "%")
}

func ____rune_private_1ed26dbc_compilerBitwiseOp(__op string) bool {
	return __op == "&" || __op == "|" || (__op == "^" || __op == "<<") || (__op == ">>" || __op == ">>>")
}

func ____rune_private_1ed26dbc_compilerNumericType(__typeName string) bool {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	return ____rune_private_1ed26dbc_compilerIntegerType(__base) || __base == "BigInt" || (__base == "Double" || __base == "Float")
}

func ____rune_private_1ed26dbc_compilerBitwiseType(__typeName string) bool {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	return ____rune_private_1ed26dbc_compilerIntegerType(__base) || __base == "BigInt"
}

func ____rune_private_1ed26dbc_compilerIntegerType(__base string) bool {
	return ____rune_private_1ed26dbc_compilerSignedIntegerType(__base) || ____rune_private_1ed26dbc_compilerUnsignedIntegerType(__base)
}

func ____rune_private_1ed26dbc_compilerSignedIntegerType(__base string) bool {
	return __base == "Int" || __base == "Int4" || (__base == "Int8" || __base == "Int16") || __base == "Int64"
}

func ____rune_private_1ed26dbc_compilerUnsignedIntegerType(__base string) bool {
	return __base == "UInt" || __base == "UInt8" || __base == "UInt16" || __base == "UInt64"
}

func ____rune_private_1ed26dbc_compilerSupportedMapKeyType(__typeName string) bool {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	return __base == "String" || __base == "Char" || (__base == "Bool" || ____rune_private_1ed26dbc_compilerIntegerType(__base)) || (__base == "Double" || __base == "Float")
}

func ____rune_private_1ed26dbc_compilerParseIntText(__text string) int {
	return func() int {
		switch {
		case strings.HasPrefix(__text, "-") == true:
			return 0 - ____rune_private_1ed26dbc_compilerParseUnsignedIntText(__text, 1, 0)
		default:
			return ____rune_private_1ed26dbc_compilerParseUnsignedIntText(__text, 0, 0)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParseUnsignedIntText(__text string, __index int, __out int) int {
	__done := __index >= len([]rune(__text))
	return func() int {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_compilerParseUnsignedIntText(__text, __index+1, __out*10+____rune_private_1ed26dbc_compilerDigitValue([]rune(__text)[__index]))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerDigitValue(__ch rune) int {
	return func() int {
		switch {
		case __ch == '0':
			return 0
		case __ch == '1':
			return 1
		case __ch == '2':
			return 2
		case __ch == '3':
			return 3
		case __ch == '4':
			return 4
		case __ch == '5':
			return 5
		case __ch == '6':
			return 6
		case __ch == '7':
			return 7
		case __ch == '8':
			return 8
		case __ch == '9':
			return 9
		default:
			return 0
		}
	}()
}

func ____rune_private_1ed26dbc_checkLetExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__hasValue := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __hasValue == true:
			return ____rune_private_1ed26dbc_checkLetValueExpr(__expr, __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkLetValueExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__value := __expr.__children[0]
	__checked := ____rune_private_1ed26dbc_checkExprExpected(__value, __expr.__value, __structs, __callables, __errors, __bindings)
	return ____rune_private_1ed26dbc_checkLetDeclaredType(__expr.__name, __value, __expr.__value, __structs, __callables, __bindings, __checked)
}

func ____rune_private_1ed26dbc_checkExprExpected(__expr __IRExpr, __expected string, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkExpr(__expr, __structs, __callables, __errors, __bindings)
	return ____rune_private_1ed26dbc_checkExpectedExprType(__expr, __expected, __structs, __callables, __bindings, __checked)
}

func ____rune_private_1ed26dbc_checkExpectedExprType(__expr __IRExpr, __expected string, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__checked := func() []string {
		switch {
		case __moduleCallKey(__expr) == "json.parse":
			return ____rune_private_1ed26dbc_checkJsonParseTarget(__expected, __structs, __errors)
		default:
			return __errors
		}
	}()
	return __checked
}

func ____rune_private_1ed26dbc_checkExpectedSelectorExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_1ed26dbc_checkSelectorExpr(__expr, __structs, __callables, __bindings, __errors)
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_1ed26dbc_checkExpectedBlockSelectorExpr(__expr.__children, __structs, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkExpectedBlockSelectorExpr(__statements []__IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__empty := len(__statements) == 0
	return func() []string {
		switch {
		case __empty == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkExpectedSelectorExpr(__statements[len(__statements)-1], __structs, __callables, __bindings, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkSelectorExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__hasReceiver := len(__expr.__children) > 0
	return func() []string {
		switch {
		case __hasReceiver == true:
			return ____rune_private_1ed26dbc_checkSelectorReceiverExpr(__expr, __expr.__children[0], __structs, __callables, __bindings, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkSelectorReceiverExpr(__expr __IRExpr, __receiver __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case __receiver.__kind == __ExprKind_At:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkStructSelectorExpr(__expr, ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__receiver, __structs, __callables, __bindings), __structs, __bindings, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkStructSelectorExpr(__expr __IRExpr, __receiverType string, __structs []__IRStructType, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case __receiverType == "":
			return __errors
		default:
			return func() []string {
				__typeDecl := ____rune_private_1ed26dbc_findCompilerStruct(__structs, ____rune_private_1ed26dbc_compilerTypeBase(__receiverType), 0)
				__found := __typeDecl.__name != ""
				return func() []string {
					switch {
					case __found == true:
						return func() []string {
							switch {
							case ____rune_private_1ed26dbc_compilerCanAccessPrivate(__typeDecl.__private, __typeDecl.__sourcePath, __bindings) == true:
								return ____rune_private_1ed26dbc_checkStructSelectorField(__expr, __receiverType, __typeDecl, __bindings, __errors)
							default:
								return func() []string {
									out := []string{}
									out = append(out, __errors...)
									out = append(out, "type \""+__typeDecl.__name+"\" is private")
									return out
								}()
							}
						}()
					default:
						return func() []string {
							out := []string{}
							out = append(out, __errors...)
							out = append(out, "type "+__receiverType+" has no fields")
							return out
						}()
					}
				}()
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkStructSelectorField(__expr __IRExpr, __receiverType string, __typeDecl __IRStructType, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__field := ____rune_private_1ed26dbc_findCompilerStructField(__typeDecl.__fields, __expr.__name, 0)
	__found := __field.__name != ""
	return func() []string {
		switch {
		case __found == true:
			return ____rune_private_1ed26dbc_checkCompilerPrivateAccess("field", __typeDecl.__name+"."+__field.__name, __field.__private, __typeDecl.__sourcePath, __bindings, __errors)
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "type "+__receiverType+" has no field \""+__expr.__name+"\"")
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkLetDeclaredType(__name string, __value __IRExpr, __expected string, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__value, __structs, __callables, __bindings)
	__mismatch := ____rune_private_1ed26dbc_compilerShouldCheckArgType(__expected, __actual) && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "binding \""+__name+"\" has type "+__actual+", expected "+__expected)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkJsonParseTarget(__expected string, __structs []__IRStructType, __errors []string) []string {
	return func() []string {
		switch {
		case __expected == "":
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "@json.parse target type cannot be inferred; add ': Type' to the binding")
				return out
			}()
		case __expected == "Dynamic":
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "@json.parse target type cannot be inferred; add ': Type' to the binding")
				return out
			}()
		case __expected == "Object":
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkJsonParseFromJsonTarget(__expected, __structs, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkJsonParseFromJsonTarget(__expected string, __structs []__IRStructType, __errors []string) []string {
	__typeDecl := ____rune_private_1ed26dbc_findCompilerStruct(__structs, ____rune_private_1ed26dbc_compilerTypeBase(__expected), 0)
	__ok := __typeDecl.__name != "" && ____rune_private_1ed26dbc_compilerStructHasFromJson(__typeDecl, __expected, 0)
	return func() []string {
		switch {
		case __ok == true:
			return __errors
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "@json.parse target type "+__expected+" does not implement &FromJson")
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerStructHasFromJson(__typeDecl __IRStructType, __expected string, __index int) bool {
	__done := __index >= len(__typeDecl.__methods)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerStructHasFromJsonAt(__typeDecl, __expected, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerStructHasFromJsonAt(__typeDecl __IRStructType, __expected string, __index int) bool {
	__method := __typeDecl.__methods[__index]
	__matched := __method.__name == "fromJson" && __method.__static && ____rune_private_1ed26dbc_compilerFromJsonSignatureMatches(__method, __expected)
	return func() bool {
		switch {
		case __matched == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerStructHasFromJson(__typeDecl, __expected, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerFromJsonSignatureMatches(__method __IRFunction, __expected string) bool {
	__arityOk := len(__method.__params) == 1
	return func() bool {
		switch {
		case __arityOk == true:
			return __method.__params[0].__typeName == "String" && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __method.__returnType)
		default:
			return false
		}
	}()
}

func ____rune_private_1ed26dbc_checkBlockExpr(__statements []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__done := __index >= len(__statements)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkBlockExprStep(__statements, __index, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkBlockExprStep(__statements []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__statement := __statements[__index]
	__nextErrors := ____rune_private_1ed26dbc_checkExpr(__statement, __structs, __callables, __errors, __bindings)
	__nextBindings := ____rune_private_1ed26dbc_blockBindingsAfterStatement(__statement, __structs, __callables, __bindings)
	return ____rune_private_1ed26dbc_checkBlockExpr(__statements, __index+1, __structs, __callables, __nextErrors, __nextBindings)
}

func ____rune_private_1ed26dbc_checkStructExpr(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__typeDecl := ____rune_private_1ed26dbc_findCompilerStruct(__structs, __expr.__name, 0)
	__found := __typeDecl.__name != ""
	return func() []string {
		switch {
		case __found == true:
			return func() []string {
				switch {
				case ____rune_private_1ed26dbc_compilerCanAccessPrivate(__typeDecl.__private, __typeDecl.__sourcePath, __bindings) == true:
					return ____rune_private_1ed26dbc_checkStructExprFields(__expr, __typeDecl, __structs, __callables, __errors, __bindings)
				default:
					return func() []string {
						out := []string{}
						out = append(out, __errors...)
						out = append(out, "type \""+__expr.__name+"\" is private")
						return out
					}()
				}
			}()
		default:
			return ____rune_private_1ed26dbc_checkStructExprUnknownType(__expr, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkStructExprUnknownType(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__next := func() []string {
		out := []string{}
		out = append(out, __errors...)
		out = append(out, "unknown type \""+__expr.__name+"\"")
		return out
	}()
	for _, __field := range __expr.__children {
		_ = __field
		__next = ____rune_private_1ed26dbc_checkExpr(__field.__children[0], __structs, __callables, __next, __bindings)
	}
	return __next
}

func ____rune_private_1ed26dbc_checkStructExprFields(__expr __IRExpr, __typeDecl __IRStructType, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__checked := ____rune_private_1ed26dbc_checkStructExprFieldValues(__expr.__name, __expr.__children, __typeDecl, 0, __structs, __callables, __errors, __bindings)
	return ____rune_private_1ed26dbc_checkStructMissingFields(__expr.__name, __typeDecl.__fields, __expr.__children, 0, __checked)
}

func ____rune_private_1ed26dbc_checkStructExprFieldValues(__typeName string, __fields []__IRExpr, __typeDecl __IRStructType, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__done := __index >= len(__fields)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkStructExprFieldValue(__typeName, __fields, __typeDecl, __index, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkStructExprFieldValue(__typeName string, __fields []__IRExpr, __typeDecl __IRStructType, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__field := __fields[__index]
	__value := __field.__children[0]
	__expected := ____rune_private_1ed26dbc_findCompilerStructField(__typeDecl.__fields, __field.__name, 0)
	__found := __expected.__name != ""
	__next := func() []string {
		switch {
		case __found == true:
			return func() []string {
				__checkedField := ____rune_private_1ed26dbc_checkCompilerPrivateAccess("field", __typeName+"."+__field.__name, __expected.__private, __typeDecl.__sourcePath, __bindings, __errors)
				__checkedValue := ____rune_private_1ed26dbc_checkExprExpected(__value, __expected.__typeName, __structs, __callables, __checkedField, __bindings)
				return ____rune_private_1ed26dbc_checkStructFieldType(__typeName, __field.__name, __value, __expected.__typeName, __structs, __callables, __bindings, __checkedValue)
			}()
		default:
			return func() []string {
				out := []string{}
				out = append(out, ____rune_private_1ed26dbc_checkExpr(__value, __structs, __callables, __errors, __bindings)...)
				out = append(out, "type "+__typeName+" has no field \""+__field.__name+"\"")
				return out
			}()
		}
	}()
	return ____rune_private_1ed26dbc_checkStructExprFieldValues(__typeName, __fields, __typeDecl, __index+1, __structs, __callables, __next, __bindings)
}

func ____rune_private_1ed26dbc_checkStructFieldType(__typeName string, __fieldName string, __value __IRExpr, __expected string, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__value, __structs, __callables, __bindings)
	__mismatch := ____rune_private_1ed26dbc_compilerShouldCheckArgType(__expected, __actual) && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "field "+__typeName+"."+__fieldName+" has type "+__actual+", expected "+__expected)
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkStructMissingFields(__typeName string, __expectedFields []__IRField, __fields []__IRExpr, __index int, __errors []string) []string {
	__done := __index >= len(__expectedFields)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkStructMissingField(__typeName, __expectedFields, __fields, __index, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkStructMissingField(__typeName string, __expectedFields []__IRField, __fields []__IRExpr, __index int, __errors []string) []string {
	__field := __expectedFields[__index]
	__present := ____rune_private_1ed26dbc_compilerExprFieldContains(__fields, __field.__name, 0)
	return func() []string {
		switch {
		case __present == true:
			return ____rune_private_1ed26dbc_checkStructMissingFields(__typeName, __expectedFields, __fields, __index+1, __errors)
		default:
			return ____rune_private_1ed26dbc_checkStructMissingFields(__typeName, __expectedFields, __fields, __index+1, func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "missing field "+__typeName+"."+__field.__name)
				return out
			}())
		}
	}()
}

func ____rune_private_1ed26dbc_compilerExprFieldContains(__fields []__IRExpr, __name string, __index int) bool {
	__done := __index >= len(__fields)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerExprFieldContainsAt(__fields, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerExprFieldContainsAt(__fields []__IRExpr, __name string, __index int) bool {
	__matched := __fields[__index].__name == __name
	return func() bool {
		switch {
		case __matched == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerExprFieldContains(__fields, __name, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_blockBindingsAfterStatement(__statement __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	return func() []__CompilerTypeBinding {
		switch {
		case __statement.__kind == __ExprKind_Let:
			return ____rune_private_1ed26dbc_bindCompilerLet(__statement, __structs, __callables, __bindings)
		case __statement.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_1ed26dbc_bindCompilerObjectDestructure(__statement, __structs, __callables, __bindings)
		default:
			return __bindings
		}
	}()
}

func ____rune_private_1ed26dbc_bindCompilerLet(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	__hasValue := len(__expr.__children) > 0
	return func() []__CompilerTypeBinding {
		switch {
		case __hasValue == true:
			return ____rune_private_1ed26dbc_addCompilerValueBinding(__bindings, __expr.__name, ____rune_private_1ed26dbc_compilerLetBindingType(__expr, __structs, __callables, __bindings))
		default:
			return __bindings
		}
	}()
}

func ____rune_private_1ed26dbc_compilerLetBindingType(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __expr.__value == "":
			return ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__expr.__children[0], __structs, __callables, __bindings)
		default:
			return __expr.__value
		}
	}()
}

func ____rune_private_1ed26dbc_addCompilerTypeBinding(__bindings []__CompilerTypeBinding, __name string, __typeName string) []__CompilerTypeBinding {
	__known := __typeName != ""
	return func() []__CompilerTypeBinding {
		switch {
		case __known == true:
			return func() []__CompilerTypeBinding {
				out := []__CompilerTypeBinding{}
				out = append(out, ____rune_private_1ed26dbc_dropCompilerTypeBinding(__bindings, __name, 0, append([]__CompilerTypeBinding{}, []__CompilerTypeBinding{____rune_private_1ed26dbc_emptyCompilerTypeBinding()}[0:0]...))...)
				out = append(out, ____rune_private_1ed26dbc_compilerTypeBinding(__name, __typeName))
				return out
			}()
		default:
			return __bindings
		}
	}()
}

func ____rune_private_1ed26dbc_addCompilerValueBinding(__bindings []__CompilerTypeBinding, __name string, __typeName string) []__CompilerTypeBinding {
	return func() []__CompilerTypeBinding {
		out := []__CompilerTypeBinding{}
		out = append(out, ____rune_private_1ed26dbc_dropCompilerTypeBinding(__bindings, __name, 0, append([]__CompilerTypeBinding{}, []__CompilerTypeBinding{____rune_private_1ed26dbc_emptyCompilerTypeBinding()}[0:0]...))...)
		out = append(out, ____rune_private_1ed26dbc_compilerTypeBinding(__name, __typeName))
		return out
	}()
}

func ____rune_private_1ed26dbc_bindCompilerObjectDestructure(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	__hasValue := len(__expr.__children) > 0
	return func() []__CompilerTypeBinding {
		switch {
		case __hasValue == true:
			return ____rune_private_1ed26dbc_bindCompilerObjectDestructureParams(__expr.__params, 0, ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__expr.__children[0], __structs, __callables, __bindings), __structs, __bindings)
		default:
			return __bindings
		}
	}()
}

func ____rune_private_1ed26dbc_bindCompilerObjectDestructureParams(__params []__IRParam, __index int, __sourceType string, __structs []__IRStructType, __bindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	__done := __index >= len(__params)
	return func() []__CompilerTypeBinding {
		switch {
		case __done == true:
			return __bindings
		default:
			return ____rune_private_1ed26dbc_bindCompilerObjectDestructureParams(__params, __index+1, __sourceType, __structs, ____rune_private_1ed26dbc_addCompilerValueBinding(__bindings, __params[__index].__name, ____rune_private_1ed26dbc_compilerObjectDestructureFieldType(__params[__index], __sourceType, __structs)))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerObjectDestructureFieldType(__param __IRParam, __sourceType string, __structs []__IRStructType) string {
	return func() string {
		switch {
		case __sourceType == "":
			return ""
		default:
			return func() string {
				__typeDecl := ____rune_private_1ed26dbc_findCompilerStruct(__structs, ____rune_private_1ed26dbc_compilerTypeBase(__sourceType), 0)
				__field := ____rune_private_1ed26dbc_findCompilerStructField(__typeDecl.__fields, __param.__typeName, 0)
				return __field.__typeName
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_dropCompilerTypeBinding(__bindings []__CompilerTypeBinding, __name string, __index int, __out []__CompilerTypeBinding) []__CompilerTypeBinding {
	__done := __index >= len(__bindings)
	return func() []__CompilerTypeBinding {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_dropCompilerTypeBindingStep(__bindings, __name, __index, __out)
		}
	}()
}

func ____rune_private_1ed26dbc_dropCompilerTypeBindingStep(__bindings []__CompilerTypeBinding, __name string, __index int, __out []__CompilerTypeBinding) []__CompilerTypeBinding {
	__matched := __bindings[__index].__name == __name
	return func() []__CompilerTypeBinding {
		switch {
		case __matched == true:
			return ____rune_private_1ed26dbc_dropCompilerTypeBinding(__bindings, __name, __index+1, __out)
		default:
			return ____rune_private_1ed26dbc_dropCompilerTypeBinding(__bindings, __name, __index+1, func() []__CompilerTypeBinding {
				out := []__CompilerTypeBinding{}
				out = append(out, __out...)
				out = append(out, __bindings[__index])
				return out
			}())
		}
	}()
}

func ____rune_private_1ed26dbc_checkExprCall(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__identifierCall := __expr.__kind == __ExprKind_Call && len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier
	return func() []string {
		switch {
		case __identifierCall == true:
			return ____rune_private_1ed26dbc_checkIdentifierCall(__expr, __expr.__children[0].__name, len(__expr.__children)-1, __structs, __callables, __errors, __bindings)
		default:
			return ____rune_private_1ed26dbc_checkSelectorCall(__expr, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkSelectorCall(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__selectorCall := __expr.__kind == __ExprKind_Call && len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Selector
	return func() []string {
		switch {
		case __selectorCall == true:
			return ____rune_private_1ed26dbc_checkSelectorCallExpr(__expr, __expr.__children[0], __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkSelectorCallExpr(__expr __IRExpr, __selector __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__hasReceiver := len(__selector.__children) > 0
	return func() []string {
		switch {
		case __hasReceiver == true:
			return ____rune_private_1ed26dbc_checkSelectorCallReceiver(__expr, __selector, __selector.__children[0], __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkSelectorCallReceiver(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case __receiver.__kind == __ExprKind_At:
			return ____rune_private_1ed26dbc_checkAtSelectorCall(__expr, __selector, __receiver, __structs, __callables, __errors, __bindings)
		case __receiver.__kind == __ExprKind_Identifier:
			return ____rune_private_1ed26dbc_checkIdentifierSelectorCall(__expr, __selector, __receiver, __structs, __callables, __errors, __bindings)
		default:
			return ____rune_private_1ed26dbc_checkInstanceSelectorCall(__expr, __selector, __receiver, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkAtSelectorCall(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__importPath := __compilerIRAtImportPath(__receiver)
	return func() []string {
		switch {
		case __importPath == "":
			return ____rune_private_1ed26dbc_checkModuleSelectorCall(__expr, __selector, __receiver, __errors)
		default:
			return func() []string {
				if __compilerGoPackageImportPath(__importPath) != "" {
					return __errors
				}
				return ____rune_private_1ed26dbc_checkIdentifierCall(__expr, __selector.__name, len(__expr.__children)-1, __structs, __callables, __errors, __bindings)
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkModuleSelectorCall(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __errors []string) []string {
	return func() []string {
		switch {
		case __receiver.__name == "go":
			return ____rune_private_1ed26dbc_checkGoModuleSelectorCall(__expr, __selector.__name, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkGoModuleSelectorCall(__expr __IRExpr, __name string, __errors []string) []string {
	return func() []string {
		switch {
		case __name == "expr":
			return ____rune_private_1ed26dbc_checkGoExprCall(__expr, __errors)
		case __name == "stmt":
			return ____rune_private_1ed26dbc_checkGoStmtCall(__expr, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkGoExprCall(__expr __IRExpr, __errors []string) []string {
	return func() []string {
		__match5 := ____rune_private_1ed26dbc_compilerCallArgCount(__expr)
		switch {
		case __match5 == 1:
			return ____rune_private_1ed26dbc_checkGoExprStringLiteral(__expr, __errors)
		case true:
			__count := __match5
			_ = __count
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "@go.expr expects 1 args, got "+strconv.Itoa(__count))
				return out
			}()
		}
		return nil
	}()
}

func ____rune_private_1ed26dbc_checkGoStmtCall(__expr __IRExpr, __errors []string) []string {
	return func() []string {
		__match6 := ____rune_private_1ed26dbc_compilerCallArgCount(__expr)
		switch {
		case __match6 == 1:
			return ____rune_private_1ed26dbc_checkGoStmtStringLiteral(__expr, __errors)
		case true:
			__count := __match6
			_ = __count
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "@go.stmt expects 1 args, got "+strconv.Itoa(__count))
				return out
			}()
		}
		return nil
	}()
}

func ____rune_private_1ed26dbc_checkGoExprStringLiteral(__expr __IRExpr, __errors []string) []string {
	return func() []string {
		switch {
		case __expr.__children[1].__kind == __ExprKind_String:
			return __errors
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "@go.expr body must be a string literal")
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkGoStmtStringLiteral(__expr __IRExpr, __errors []string) []string {
	return func() []string {
		switch {
		case __expr.__children[1].__kind == __ExprKind_String:
			return __errors
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "@go.stmt argument must be a string literal")
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerCallArgCount(__expr __IRExpr) int {
	return len(__expr.__children) - 1
}

func ____rune_private_1ed26dbc_checkIdentifierSelectorCall(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case __selector.__op == "::":
			return ____rune_private_1ed26dbc_checkStaticSelectorCall(__expr, __selector, __receiver.__name, __structs, __callables, __errors, __bindings)
		default:
			return ____rune_private_1ed26dbc_checkInstanceSelectorCall(__expr, __selector, __receiver, __structs, __callables, __errors, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_checkStaticSelectorCall(__expr __IRExpr, __selector __IRExpr, __receiverType string, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__name := ____rune_private_1ed26dbc_compilerStaticMethodName(__receiverType, __selector.__name)
	__callable := ____rune_private_1ed26dbc_findCompilerCallable(__callables, __name, 0)
	__found := __callable.__name != ""
	return func() []string {
		switch {
		case __found == true:
			return func() []string {
				switch {
				case ____rune_private_1ed26dbc_checkCallableVisibility(__callable, __bindings) == true:
					return ____rune_private_1ed26dbc_checkCallableCall(__callable, __expr, len(__expr.__children)-1, __structs, __callables, __bindings, __errors)
				default:
					return func() []string {
						out := []string{}
						out = append(out, __errors...)
						out = append(out, "static method \""+__name+"\" is private")
						return out
					}()
				}
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkInstanceSelectorCall(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__receiverType := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__receiver, __structs, __callables, __bindings)
	__known := __receiverType != ""
	return func() []string {
		switch {
		case __known == true:
			return ____rune_private_1ed26dbc_checkKnownInstanceSelectorCall(__expr, __selector, ____rune_private_1ed26dbc_compilerTypeBase(__receiverType), __structs, __callables, __errors, __bindings)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkKnownInstanceSelectorCall(__expr __IRExpr, __selector __IRExpr, __receiverType string, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__name := ____rune_private_1ed26dbc_compilerInstanceMethodName(__receiverType, __selector.__name)
	__callable := ____rune_private_1ed26dbc_findCompilerCallable(__callables, __name, 0)
	__found := __callable.__name != ""
	return func() []string {
		switch {
		case __found == true:
			return func() []string {
				switch {
				case ____rune_private_1ed26dbc_checkCallableVisibility(__callable, __bindings) == true:
					return ____rune_private_1ed26dbc_checkCallableCall(__callable, __expr, len(__expr.__children)-1, __structs, __callables, __bindings, __errors)
				default:
					return func() []string {
						out := []string{}
						out = append(out, __errors...)
						out = append(out, "method \""+__name+"\" is private")
						return out
					}()
				}
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_checkIdentifierCall(__expr __IRExpr, __name string, __arity int, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__callable := ____rune_private_1ed26dbc_findCompilerCallable(__callables, __name, 0)
	__found := __callable.__name != ""
	return func() []string {
		switch {
		case __found == true:
			return func() []string {
				switch {
				case ____rune_private_1ed26dbc_checkCallableVisibility(__callable, __bindings) == true:
					return ____rune_private_1ed26dbc_checkCallableCall(__callable, __expr, __arity, __structs, __callables, __bindings, __errors)
				default:
					return func() []string {
						out := []string{}
						out = append(out, __errors...)
						out = append(out, "function \""+__name+"\" is private")
						return out
					}()
				}
			}()
		default:
			return ____rune_private_1ed26dbc_checkUndefinedIdentifierCall(__name, __bindings, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkUndefinedIdentifierCall(__name string, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, __name, 0).__typeName == "MacroFunction":
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, __name+" is a macro and can only be used with '#'")
				return out
			}()
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "undefined function "+__name)
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkCallableVisibility(__callable __CompilerCallable, __bindings []__CompilerTypeBinding) bool {
	return ____rune_private_1ed26dbc_compilerCanAccessPrivate(__callable.__private, __callable.__sourcePath, __bindings)
}

func ____rune_private_1ed26dbc_checkCompilerPrivateAccess(__kind string, __name string, __private bool, __sourcePath string, __bindings []__CompilerTypeBinding, __errors []string) []string {
	return func() []string {
		switch {
		case ____rune_private_1ed26dbc_compilerCanAccessPrivate(__private, __sourcePath, __bindings) == true:
			return __errors
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, __kind+" \""+__name+"\" is private")
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerCanAccessPrivate(__private bool, __sourcePath string, __bindings []__CompilerTypeBinding) bool {
	return func() bool {
		switch {
		case __private == false:
			return true
		default:
			return func() bool {
				__current := ____rune_private_1ed26dbc_compilerCurrentSourcePath(__bindings)
				return __sourcePath == "" || (__current == "" || __current == __sourcePath)
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerCurrentSourcePath(__bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, "__sourcePath", 0).__typeName
}

func ____rune_private_1ed26dbc_checkCallableCall(__callable __CompilerCallable, __expr __IRExpr, __arity int, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string) []string {
	__arityOk := __callable.__arity == __arity
	return func() []string {
		switch {
		case __arityOk == true:
			return ____rune_private_1ed26dbc_checkCallableArgTypes(__callable, __expr, __structs, __callables, __bindings, __errors, 0)
		default:
			return ____rune_private_1ed26dbc_checkCallableArity(__callable, __arity, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_checkCallableArity(__callable __CompilerCallable, __arity int, __errors []string) []string {
	__ok := __callable.__arity == __arity
	return func() []string {
		switch {
		case __ok == true:
			return __errors
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "function \""+__callable.__name+"\" expects "+__compilerIntToString(__callable.__arity)+" args, got "+__compilerIntToString(__arity))
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_checkCallableArgTypes(__callable __CompilerCallable, __expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string, __index int) []string {
	__done := __index >= len(__callable.__paramTypes)
	return func() []string {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_checkCallableArgType(__callable, __expr, __structs, __callables, __bindings, __errors, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_checkCallableArgType(__callable __CompilerCallable, __expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __errors []string, __index int) []string {
	__expected := __callable.__paramTypes[__index]
	__arg := __expr.__children[__index+1]
	__checked := ____rune_private_1ed26dbc_checkExpectedExprType(__arg, __expected, __structs, __callables, __bindings, __errors)
	__actual := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__arg, __structs, __callables, __bindings)
	__mismatch := ____rune_private_1ed26dbc_compilerShouldCheckArgType(__expected, __actual) && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	__next := func() []string {
		switch {
		case __mismatch == true:
			return func() []string {
				out := []string{}
				out = append(out, __checked...)
				out = append(out, ____rune_private_1ed26dbc_compilerArgumentTypeError(__callable.__name, __index+1, __actual, __expected))
				return out
			}()
		default:
			return __checked
		}
	}()
	return ____rune_private_1ed26dbc_checkCallableArgTypes(__callable, __expr, __structs, __callables, __bindings, __next, __index+1)
}

func ____rune_private_1ed26dbc_compilerArgumentTypeError(__name string, __index int, __actual string, __expected string) string {
	return "argument " + __compilerIntToString(__index) + " to \"" + __name + "\" has type " + __actual + ", expected " + __expected
}

func ____rune_private_1ed26dbc_compilerShouldCheckArgType(__expected string, __actual string) bool {
	return __expected != "" && (__expected != "Dynamic" && __actual != "") && ____rune_private_1ed26dbc_compilerTypeIsGenericPlaceholder(__expected) == false
}

func ____rune_private_1ed26dbc_compilerTypeIsGenericPlaceholder(__typeName string) bool {
	return len([]rune(__typeName)) == 1 && ([]rune(__typeName)[0] >= 'A' && []rune(__typeName)[0] <= 'Z')
}

func ____rune_private_1ed26dbc_checkFunctionReturn(__fn __IRFunction, __structs []__IRStructType, __callables []__CompilerCallable, __errors []string, __bindings []__CompilerTypeBinding) []string {
	__expected := __fn.__returnType
	__checked := ____rune_private_1ed26dbc_checkExpectedExprType(__fn.__body, __expected, __structs, __callables, __bindings, __errors)
	__actual := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__fn.__body, __structs, __callables, __bindings)
	__shouldCheck := __expected != "" && __expected != "Dynamic" && __actual != ""
	return func() []string {
		switch {
		case __shouldCheck == true:
			return ____rune_private_1ed26dbc_checkFunctionReturnType(__fn.__name, __expected, __actual, __checked)
		default:
			return __checked
		}
	}()
}

func ____rune_private_1ed26dbc_checkFunctionReturnType(__name string, __expected string, __actual string, __errors []string) []string {
	__ok := ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual)
	return func() []string {
		switch {
		case __ok == true:
			return __errors
		default:
			return func() []string {
				out := []string{}
				out = append(out, __errors...)
				out = append(out, "function \""+__name+"\" returns "+__actual+", expected "+__expected)
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerTypesCompatible(__expected string, __actual string) bool {
	__nullable := strings.HasSuffix(__expected, "?")
	return func() bool {
		switch {
		case __nullable == true:
			return __actual == "Null" || ____rune_private_1ed26dbc_compilerTypesCompatible(func() string { runes := []rune(__expected); return string(runes[0 : len([]rune(__expected))-1]) }(), __actual)
		default:
			return ____rune_private_1ed26dbc_compilerNonNullableTypesCompatible(__expected, __actual)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerNonNullableTypesCompatible(__expected string, __actual string) bool {
	__same := __expected == __actual
	return func() bool {
		switch {
		case __same == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerGenericTypesCompatible(__expected, __actual)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerGenericTypesCompatible(__expected string, __actual string) bool {
	__expectedBase := ____rune_private_1ed26dbc_compilerTypeBase(__expected)
	__actualBase := ____rune_private_1ed26dbc_compilerTypeBase(__actual)
	__sameBase := __expectedBase == __actualBase
	return func() bool {
		switch {
		case __sameBase == true:
			return ____rune_private_1ed26dbc_compilerGenericTypeArgsCompatible(__expected, __actual)
		default:
			return __expectedBase == __actual
		}
	}()
}

func ____rune_private_1ed26dbc_compilerGenericTypeArgsCompatible(__expected string, __actual string) bool {
	__expectedArgs := ____rune_private_1ed26dbc_compilerGenericArgs(__expected)
	__actualArgs := ____rune_private_1ed26dbc_compilerGenericArgs(__actual)
	__hasExpected := len(__expectedArgs) > 0
	__hasActual := len(__actualArgs) > 0
	return func() bool {
		switch {
		case __hasExpected == true:
			return func() bool {
				switch {
				case __hasActual == true:
					return ____rune_private_1ed26dbc_compilerGenericArgListsCompatible(__expectedArgs, __actualArgs, 0)
				default:
					return true
				}
			}()
		default:
			return true
		}
	}()
}

func ____rune_private_1ed26dbc_compilerGenericArgListsCompatible(__expected []string, __actual []string, __index int) bool {
	__lengthMismatch := len(__expected) != len(__actual)
	return func() bool {
		switch {
		case __lengthMismatch == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerGenericArgListsCompatibleAt(__expected, __actual, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerGenericArgListsCompatibleAt(__expected []string, __actual []string, __index int) bool {
	__done := __index >= len(__expected)
	return func() bool {
		switch {
		case __done == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerTypesCompatible(strings.TrimSpace(__expected[__index]), strings.TrimSpace(__actual[__index])) && ____rune_private_1ed26dbc_compilerGenericArgListsCompatibleAt(__expected, __actual, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerTypesComparable(__left string, __right string) bool {
	__unknown := __left == "" || __right == ""
	return func() bool {
		switch {
		case __unknown == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerKnownTypesComparable(__left, __right)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerKnownTypesComparable(__left string, __right string) bool {
	__same := __left == __right
	return func() bool {
		switch {
		case __same == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerNullableTypesComparable(__left, __right)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerNullableTypesComparable(__left string, __right string) bool {
	__leftNullable := strings.HasSuffix(__left, "?")
	return func() bool {
		switch {
		case __leftNullable == true:
			return ____rune_private_1ed26dbc_compilerLeftNullableComparable(__left, __right)
		default:
			return ____rune_private_1ed26dbc_compilerRightNullableComparable(__left, __right)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerLeftNullableComparable(__left string, __right string) bool {
	__rightNull := __right == "Null"
	return func() bool {
		switch {
		case __rightNull == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerTypesComparable(func() string { runes := []rune(__left); return string(runes[0 : len([]rune(__left))-1]) }(), __right)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerRightNullableComparable(__left string, __right string) bool {
	__rightNullable := strings.HasSuffix(__right, "?")
	return func() bool {
		switch {
		case __rightNullable == true:
			return ____rune_private_1ed26dbc_compilerRightNullableInnerComparable(__left, __right)
		default:
			return false
		}
	}()
}

func ____rune_private_1ed26dbc_compilerRightNullableInnerComparable(__left string, __right string) bool {
	__leftNull := __left == "Null"
	return func() bool {
		switch {
		case __leftNull == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerTypesComparable(__left, func() string { runes := []rune(__right); return string(runes[0 : len([]rune(__right))-1]) }())
		}
	}()
}

func ____rune_private_1ed26dbc_compilerTypeBase(__typeName string) string {
	__generic := strings.Index(__typeName, "[")
	return func() string {
		if __generic < 0 {
			return __typeName
		}
		return func() string { runes := []rune(__typeName); return string(runes[0:__generic]) }()
	}()
}

func ____rune_private_1ed26dbc_compilerGenericArgs(__typeName string) []string {
	__inner := ____rune_private_1ed26dbc_compilerGenericInner(__typeName)
	return func() []string {
		switch {
		case __inner == "":
			return []string{}
		default:
			return func() []string { parts := strings.Split(__inner, ","); return parts }()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerArrayType(__elem string) string {
	return func() string {
		switch {
		case __elem == "":
			return "Array"
		default:
			return "Array[" + __elem + "]"
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMapType(__key string, __value string) string {
	__complete := __key != "" && __value != ""
	return func() string {
		switch {
		case __complete == true:
			return "Map[" + __key + ", " + __value + "]"
		default:
			return "Map"
		}
	}()
}

func ____rune_private_1ed26dbc_compilerTupleLiteralType(__elements []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return "Tuple[" + ____rune_private_1ed26dbc_compilerTupleLiteralTypes(__elements, __callables, __bindings, 0, "") + "]"
}

func ____rune_private_1ed26dbc_compilerTupleLiteralTypes(__elements []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __index int, __out string) string {
	__done := __index >= len(__elements)
	return func() string {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_compilerTupleLiteralTypes(__elements, __callables, __bindings, __index+1, __out+____rune_private_1ed26dbc_compilerTupleTypeSeparator(__index)+____rune_private_1ed26dbc_inferCompilerExprType(__elements[__index], __callables, __bindings))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerTupleTypeSeparator(__index int) string {
	return func() string {
		switch {
		case __index == 0:
			return ""
		default:
			return ","
		}
	}()
}

func ____rune_private_1ed26dbc_compilerNullableType(__inner string) string {
	return func() string {
		switch {
		case __inner == "":
			return ""
		case __inner == "Null":
			return "Null"
		default:
			return func() string {
				switch {
				case strings.HasSuffix(__inner, "?") == true:
					return __inner
				default:
					return __inner + "?"
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerArrayElementType(__typeName string) string {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	__args := ____rune_private_1ed26dbc_compilerGenericArgs(__typeName)
	__valid := __base == "Array" && len(__args) == 1
	return func() string {
		switch {
		case __valid == true:
			return strings.TrimSpace(__args[0])
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMapKeyType(__typeName string) string {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	__args := ____rune_private_1ed26dbc_compilerGenericArgs(__typeName)
	__valid := __base == "Map" && len(__args) == 2
	return func() string {
		switch {
		case __valid == true:
			return strings.TrimSpace(__args[0])
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMapValueType(__typeName string) string {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	__args := ____rune_private_1ed26dbc_compilerGenericArgs(__typeName)
	__valid := __base == "Map" && len(__args) == 2
	return func() string {
		switch {
		case __valid == true:
			return strings.TrimSpace(__args[1])
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerTupleElementTypes(__typeName string) []string {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__typeName)
	__args := ____rune_private_1ed26dbc_compilerGenericArgs(__typeName)
	__valid := (__base == "Tuple" || __base == "ReadonlyTuple") && len(__args) > 0
	return func() []string {
		switch {
		case __valid == true:
			return ____rune_private_1ed26dbc_compilerTrimTypes(__args, 0, []string{})
		default:
			return []string{}
		}
	}()
}

func ____rune_private_1ed26dbc_compilerTrimTypes(__types []string, __index int, __out []string) []string {
	__done := __index >= len(__types)
	return func() []string {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_compilerTrimTypes(__types, __index+1, func() []string {
				out := []string{}
				out = append(out, __out...)
				out = append(out, strings.TrimSpace(__types[__index]))
				return out
			}())
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerExprType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Identifier:
			return ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, __expr.__name, 0).__typeName
		case __expr.__kind == __ExprKind_This:
			return ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, "this", 0).__typeName
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_1ed26dbc_inferCompilerSelectorType(__expr, __callables, __bindings)
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
		case __expr.__kind == __ExprKind_Struct:
			return __expr.__name
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_1ed26dbc_inferCompilerCallType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_1ed26dbc_inferCompilerBlockType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Ternary:
			return ____rune_private_1ed26dbc_inferCompilerTernaryType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Unary:
			return ____rune_private_1ed26dbc_inferCompilerUnaryType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Postfix:
			return ____rune_private_1ed26dbc_inferCompilerPostfixType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Binary:
			return ____rune_private_1ed26dbc_inferCompilerBinaryType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Array:
			return ____rune_private_1ed26dbc_inferCompilerArrayType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Tuple:
			return ____rune_private_1ed26dbc_inferCompilerTupleType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Map:
			return ____rune_private_1ed26dbc_inferCompilerMapType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Spread:
			return ____rune_private_1ed26dbc_inferCompilerSpreadType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Index:
			return ____rune_private_1ed26dbc_inferCompilerIndexType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_PatternBlock:
			return ____rune_private_1ed26dbc_inferCompilerPatternBlockType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Match:
			return ____rune_private_1ed26dbc_inferCompilerMatchType(__expr, __callables, __bindings)
		case __expr.__kind == __ExprKind_Unwrap:
			return ____rune_private_1ed26dbc_inferCompilerUnwrapType(__expr, __callables, __bindings)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_1ed26dbc_inferCompilerSelectorTypeWithStructs(__expr, __structs, __callables, __bindings)
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_1ed26dbc_inferCompilerBlockTypeWithStructs(__expr, __structs, __callables, __bindings)
		case __expr.__kind == __ExprKind_PatternBlock:
			return ____rune_private_1ed26dbc_inferCompilerPatternBlockTypeWithStructs(__expr, __structs, __callables, __bindings)
		case __expr.__kind == __ExprKind_Match:
			return ____rune_private_1ed26dbc_inferCompilerMatchTypeWithStructs(__expr, __structs, __callables, __bindings)
		case __expr.__kind == __ExprKind_Unwrap:
			return ____rune_private_1ed26dbc_inferCompilerUnwrapTypeWithStructs(__expr, __structs, __callables, __bindings)
		default:
			return ____rune_private_1ed26dbc_inferCompilerExprType(__expr, __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerSelectorTypeWithStructs(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__hasReceiver := len(__expr.__children) > 0
	return func() string {
		switch {
		case __hasReceiver == true:
			return ____rune_private_1ed26dbc_inferCompilerSelectorTypeFromReceiverWithStructs(__expr, __expr.__children[0], __structs, __callables, __bindings)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerSelectorTypeFromReceiverWithStructs(__expr __IRExpr, __receiver __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __receiver.__kind == __ExprKind_At:
			return ____rune_private_1ed26dbc_inferCompilerSelectorTypeFromReceiver(__expr, __receiver, __callables, __bindings)
		default:
			return ____rune_private_1ed26dbc_inferCompilerStructSelectorType(__expr, ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__receiver, __structs, __callables, __bindings), __structs)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerStructSelectorType(__expr __IRExpr, __receiverType string, __structs []__IRStructType) string {
	return func() string {
		switch {
		case __receiverType == "":
			return ""
		default:
			return func() string {
				__typeDecl := ____rune_private_1ed26dbc_findCompilerStruct(__structs, ____rune_private_1ed26dbc_compilerTypeBase(__receiverType), 0)
				__field := ____rune_private_1ed26dbc_findCompilerStructField(__typeDecl.__fields, __expr.__name, 0)
				return __field.__typeName
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerArrayType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_compilerArrayType(____rune_private_1ed26dbc_inferCompilerArrayElementType(__expr.__children, __callables, __bindings, 0, ""))
}

func ____rune_private_1ed26dbc_inferCompilerArrayElementType(__elements []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __index int, __expected string) string {
	__done := __index >= len(__elements)
	return func() string {
		switch {
		case __done == true:
			return __expected
		default:
			return ____rune_private_1ed26dbc_inferCompilerArrayElementTypeAt(__elements, __callables, __bindings, __index, __expected)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerArrayElementTypeAt(__elements []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __index int, __expected string) string {
	__actual := ____rune_private_1ed26dbc_inferCompilerArrayLiteralElementType(__elements[__index], __callables, __bindings)
	__next := ____rune_private_1ed26dbc_compilerNextArrayElementType(__expected, __actual)
	return ____rune_private_1ed26dbc_inferCompilerArrayElementType(__elements, __callables, __bindings, __index+1, __next)
}

func ____rune_private_1ed26dbc_inferCompilerArrayLiteralElementType(__elem __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __elem.__kind == __ExprKind_Spread:
			return ____rune_private_1ed26dbc_compilerArrayElementType(____rune_private_1ed26dbc_inferCompilerExprType(__elem.__children[0], __callables, __bindings))
		default:
			return ____rune_private_1ed26dbc_inferCompilerExprType(__elem, __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerTupleType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_compilerTupleLiteralType(__expr.__children, __callables, __bindings)
}

func ____rune_private_1ed26dbc_inferCompilerPostfixType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__complete := len(__expr.__children) > 0
	return func() string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerMapType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_compilerMapType(____rune_private_1ed26dbc_inferCompilerMapKeyType(__expr.__children, __callables, __bindings, 0, ""), ____rune_private_1ed26dbc_inferCompilerMapValueType(__expr.__children, __callables, __bindings, 0, ""))
}

func ____rune_private_1ed26dbc_inferCompilerMapKeyType(__entries []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __index int, __expected string) string {
	__done := __index >= len(__entries)
	return func() string {
		switch {
		case __done == true:
			return __expected
		default:
			return ____rune_private_1ed26dbc_inferCompilerMapKeyType(__entries, __callables, __bindings, __index+1, ____rune_private_1ed26dbc_compilerNextMapEntryType(__expected, ____rune_private_1ed26dbc_inferCompilerMapEntryKeyType(__entries[__index], __callables, __bindings)))
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerMapValueType(__entries []__IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __index int, __expected string) string {
	__done := __index >= len(__entries)
	return func() string {
		switch {
		case __done == true:
			return __expected
		default:
			return ____rune_private_1ed26dbc_inferCompilerMapValueType(__entries, __callables, __bindings, __index+1, ____rune_private_1ed26dbc_compilerNextMapEntryType(__expected, ____rune_private_1ed26dbc_inferCompilerMapEntryValueType(__entries[__index], __callables, __bindings)))
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerMapEntryKeyType(__entry __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__complete := len(__entry.__children) > 0
	return func() string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_inferCompilerExprType(__entry.__children[0], __callables, __bindings)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerMapEntryValueType(__entry __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__complete := len(__entry.__children) > 1
	return func() string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_inferCompilerExprType(__entry.__children[1], __callables, __bindings)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerSpreadType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__complete := len(__expr.__children) > 0
	return func() string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerIndexType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__complete := len(__expr.__children) >= 2
	return func() string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_inferCompilerIndexReceiverType(____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings), __expr.__children[1])
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerPatternBlockType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_inferCompilerPatternBlockTypeWithStructs(__expr, append([]__IRStructType{}, []__IRStructType{____rune_private_1ed26dbc_emptyCompilerStruct()}[0:0]...), __callables, __bindings)
}

func ____rune_private_1ed26dbc_inferCompilerMatchType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_inferCompilerMatchTypeWithStructs(__expr, append([]__IRStructType{}, []__IRStructType{____rune_private_1ed26dbc_emptyCompilerStruct()}[0:0]...), __callables, __bindings)
}

func ____rune_private_1ed26dbc_inferCompilerMatchTypeWithStructs(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_inferCompilerPatternBranchTypesWithStructs(__expr.__children, 1, __structs, __callables, __bindings, "")
}

func ____rune_private_1ed26dbc_inferCompilerUnwrapType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_inferCompilerUnwrapTypeWithStructs(__expr, append([]__IRStructType{}, []__IRStructType{____rune_private_1ed26dbc_emptyCompilerStruct()}[0:0]...), __callables, __bindings)
}

func ____rune_private_1ed26dbc_inferCompilerUnwrapTypeWithStructs(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__complete := len(__expr.__children) > 0
	return func() string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_compilerResultOkType(____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__expr.__children[0], __structs, __callables, __bindings))
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerResultOkType(__sourceType string) string {
	__base := ____rune_private_1ed26dbc_compilerTypeBase(__sourceType)
	__args := ____rune_private_1ed26dbc_compilerGenericArgs(__sourceType)
	__valid := __base == "Result" && len(__args) == 2
	return func() string {
		switch {
		case __valid == true:
			return strings.TrimSpace(__args[0])
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerPatternBlockTypeWithStructs(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_inferCompilerPatternBranchTypesWithStructs(__expr.__children, 0, __structs, __callables, __bindings, "")
}

func ____rune_private_1ed26dbc_inferCompilerPatternBranchTypesWithStructs(__branches []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __expected string) string {
	__done := __index >= len(__branches)
	return func() string {
		switch {
		case __done == true:
			return __expected
		default:
			return ____rune_private_1ed26dbc_inferCompilerPatternBranchTypeWithStructs(__branches, __index, __structs, __callables, __bindings, __expected)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerPatternBranchTypeWithStructs(__branches []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __expected string) string {
	__branch := __branches[__index]
	__complete := len(__branch.__children) >= 2
	return func() string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_inferCompilerPatternBranchValueTypeWithStructs(__branches, __index, __branch.__children[1], __structs, __callables, __bindings, __expected)
		default:
			return ____rune_private_1ed26dbc_inferCompilerPatternBranchTypesWithStructs(__branches, __index+1, __structs, __callables, __bindings, __expected)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerPatternBranchValueTypeWithStructs(__branches []__IRExpr, __index int, __value __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __expected string) string {
	__actual := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__value, __structs, __callables, __bindings)
	__next := ____rune_private_1ed26dbc_inferCompilerPatternCommonType(__expected, __actual)
	return ____rune_private_1ed26dbc_inferCompilerPatternBranchTypesWithStructs(__branches, __index+1, __structs, __callables, __bindings, __next)
}

func ____rune_private_1ed26dbc_inferCompilerPatternCommonType(__expected string, __actual string) string {
	return func() string {
		switch {
		case __expected == "":
			return __actual
		default:
			return func() string {
				switch {
				case __actual == "":
					return __expected
				default:
					return ____rune_private_1ed26dbc_inferCompilerCommonType(__expected, __actual)
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerIndexReceiverType(__receiverType string, __indexExpr __IRExpr) string {
	__arrayElem := ____rune_private_1ed26dbc_compilerArrayElementType(__receiverType)
	__mapValue := ____rune_private_1ed26dbc_compilerMapValueType(__receiverType)
	__tuple := ____rune_private_1ed26dbc_compilerTupleElementTypes(__receiverType)
	return func() string {
		switch {
		case __arrayElem == "":
			return ____rune_private_1ed26dbc_inferCompilerNonArrayIndexReceiverType(__mapValue, __tuple, __indexExpr)
		default:
			return __arrayElem
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerNonArrayIndexReceiverType(__mapValue string, __tuple []string, __indexExpr __IRExpr) string {
	return func() string {
		switch {
		case __mapValue == "":
			return ____rune_private_1ed26dbc_inferCompilerTupleIndexReceiverType(__tuple, __indexExpr)
		default:
			return ____rune_private_1ed26dbc_compilerNullableType(__mapValue)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerTupleIndexReceiverType(__tuple []string, __indexExpr __IRExpr) string {
	return func() string {
		switch {
		case len(__tuple) > 0 == true:
			return ____rune_private_1ed26dbc_inferCompilerTupleIndexLiteralType(__tuple, __indexExpr)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerTupleIndexLiteralType(__tuple []string, __indexExpr __IRExpr) string {
	return func() string {
		switch {
		case __indexExpr.__kind == __ExprKind_Int:
			return ____rune_private_1ed26dbc_inferCompilerTupleIndexElementType(__tuple, ____rune_private_1ed26dbc_compilerParseIntText(__indexExpr.__value))
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerTupleIndexElementType(__tuple []string, __index int) string {
	__outOfRange := __index < 0 || __index >= len(__tuple)
	return func() string {
		switch {
		case __outOfRange == true:
			return ""
		default:
			return __tuple[__index]
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerCallType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__identifierCall := len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Identifier
	return func() string {
		switch {
		case __identifierCall == true:
			return ____rune_private_1ed26dbc_compilerCallableReturnOrText(____rune_private_1ed26dbc_findCompilerCallable(__callables, __expr.__children[0].__name, 0), __expr.__text)
		default:
			return ____rune_private_1ed26dbc_inferCompilerSelectorCallType(__expr, __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerCallableReturnOrText(__callable __CompilerCallable, __fallback string) string {
	__found := __callable.__name != ""
	return func() string {
		switch {
		case __found == true:
			return __callable.__returnType
		default:
			return __fallback
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerSelectorCallType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__selectorCall := len(__expr.__children) > 0 && __expr.__children[0].__kind == __ExprKind_Selector
	return func() string {
		switch {
		case __selectorCall == true:
			return ____rune_private_1ed26dbc_inferCompilerSelectorCallTypeFromSelector(__expr, __expr.__children[0], __callables, __bindings)
		default:
			return __expr.__text
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerSelectorCallTypeFromSelector(__expr __IRExpr, __selector __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__hasReceiver := len(__selector.__children) > 0
	return func() string {
		switch {
		case __hasReceiver == true:
			return ____rune_private_1ed26dbc_inferCompilerSelectorCallTypeFromReceiver(__expr, __selector, __selector.__children[0], __callables, __bindings)
		default:
			return __expr.__text
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerSelectorCallTypeFromReceiver(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __receiver.__kind == __ExprKind_At:
			return ____rune_private_1ed26dbc_inferCompilerAtSelectorCallType(__expr, __selector, __receiver, __callables)
		case __receiver.__kind == __ExprKind_Identifier:
			return ____rune_private_1ed26dbc_inferCompilerIdentifierSelectorCallType(__expr, __selector, __receiver, __callables, __bindings)
		default:
			return ____rune_private_1ed26dbc_inferCompilerInstanceSelectorCallType(__expr, __selector, __receiver, __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerAtSelectorCallType(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __callables []__CompilerCallable) string {
	__importPath := __compilerIRAtImportPath(__receiver)
	return func() string {
		switch {
		case __importPath == "" == true:
			return __expr.__text
		default:
			return func() string {
				if __compilerGoPackageImportPath(__importPath) != "" {
					return __expr.__text
				}
				return ____rune_private_1ed26dbc_compilerCallableReturnOrText(____rune_private_1ed26dbc_findCompilerCallable(__callables, __selector.__name, 0), __expr.__text)
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerSelectorType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__hasReceiver := len(__expr.__children) > 0
	return func() string {
		switch {
		case __hasReceiver == true:
			return ____rune_private_1ed26dbc_inferCompilerSelectorTypeFromReceiver(__expr, __expr.__children[0], __callables, __bindings)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerSelectorTypeFromReceiver(__expr __IRExpr, __receiver __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __receiver.__kind == __ExprKind_At:
			return func() string {
				switch {
				case __compilerIRAtImportPath(__receiver) == "":
					return ""
				default:
					return ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, __expr.__name, 0).__typeName
				}
			}()
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerIdentifierSelectorCallType(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __selector.__op == "::":
			return ____rune_private_1ed26dbc_compilerCallableReturnOrText(____rune_private_1ed26dbc_findCompilerCallable(__callables, ____rune_private_1ed26dbc_compilerStaticMethodName(__receiver.__name, __selector.__name), 0), __expr.__text)
		default:
			return ____rune_private_1ed26dbc_inferCompilerInstanceSelectorCallType(__expr, __selector, __receiver, __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerInstanceSelectorCallType(__expr __IRExpr, __selector __IRExpr, __receiver __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__receiverType := ____rune_private_1ed26dbc_inferCompilerExprType(__receiver, __callables, __bindings)
	__known := __receiverType != ""
	return func() string {
		switch {
		case __known == true:
			return ____rune_private_1ed26dbc_compilerCallableReturnOrText(____rune_private_1ed26dbc_findCompilerCallable(__callables, ____rune_private_1ed26dbc_compilerInstanceMethodName(____rune_private_1ed26dbc_compilerTypeBase(__receiverType), __selector.__name), 0), __expr.__text)
		default:
			return __expr.__text
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerBlockType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_inferCompilerBlockTypeWithStructs(__expr, append([]__IRStructType{}, []__IRStructType{____rune_private_1ed26dbc_emptyCompilerStruct()}[0:0]...), __callables, __bindings)
}

func ____rune_private_1ed26dbc_inferCompilerBlockTypeWithStructs(__expr __IRExpr, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_inferCompilerBlockTypeAtWithStructs(__expr.__children, 0, __structs, __callables, __bindings, "Void")
}

func ____rune_private_1ed26dbc_inferCompilerBlockTypeAtWithStructs(__statements []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __lastType string) string {
	__done := __index >= len(__statements)
	return func() string {
		switch {
		case __done == true:
			return __lastType
		default:
			return ____rune_private_1ed26dbc_inferCompilerBlockTypeStepWithStructs(__statements, __index, __structs, __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerBlockTypeStepWithStructs(__statements []__IRExpr, __index int, __structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__statement := __statements[__index]
	__statementType := ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__statement, __structs, __callables, __bindings)
	__nextBindings := ____rune_private_1ed26dbc_blockBindingsAfterStatement(__statement, __structs, __callables, __bindings)
	return ____rune_private_1ed26dbc_inferCompilerBlockTypeAtWithStructs(__statements, __index+1, __structs, __callables, __nextBindings, __statementType)
}

func ____rune_private_1ed26dbc_inferCompilerTernaryType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__complete := len(__expr.__children) >= 3
	return func() string {
		switch {
		case __complete == true:
			return ____rune_private_1ed26dbc_inferCompilerCommonType(____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[1], __callables, __bindings), ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[2], __callables, __bindings))
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerUnaryType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case __expr.__op == "!":
			return "Bool"
		default:
			return ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerBinaryType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	__comparison := __expr.__op == "==" || __expr.__op == "!=" || (__expr.__op == "<" || __expr.__op == "<=") || (__expr.__op == ">" || __expr.__op == ">=")
	__boolOp := __expr.__op == "&&" || __expr.__op == "||"
	__knownBool := __comparison || __boolOp
	return func() string {
		switch {
		case __knownBool == true:
			return "Bool"
		default:
			return ____rune_private_1ed26dbc_inferCompilerBinaryValueType(__expr, __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerBinaryValueType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case ____rune_private_1ed26dbc_compilerArithmeticOp(__expr.__op) == true:
			return ____rune_private_1ed26dbc_inferCompilerNumericBinaryType(__expr, __callables, __bindings)
		default:
			return ____rune_private_1ed26dbc_inferCompilerBitwiseBinaryType(__expr, __callables, __bindings)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerNumericBinaryType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return ____rune_private_1ed26dbc_inferCompilerNumericPairType(____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings), ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[1], __callables, __bindings))
}

func ____rune_private_1ed26dbc_inferCompilerBitwiseBinaryType(__expr __IRExpr, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding) string {
	return func() string {
		switch {
		case ____rune_private_1ed26dbc_compilerBitwiseOp(__expr.__op) == true:
			return ____rune_private_1ed26dbc_inferCompilerBitwisePairType(____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[0], __callables, __bindings), ____rune_private_1ed26dbc_inferCompilerExprType(__expr.__children[1], __callables, __bindings))
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerBitwisePairType(__left string, __right string) string {
	__missing := __left == "" || __right == ""
	return func() string {
		switch {
		case __missing == true:
			return ""
		default:
			return ____rune_private_1ed26dbc_inferCompilerMatchingPairType(__left, __right)
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerMatchingPairType(__left string, __right string) string {
	__matched := __left == __right
	return func() string {
		switch {
		case __matched == true:
			return __left
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_inferCompilerNumericPairType(__left string, __right string) string {
	return func() string {
		if __left == "" || __right == "" {
			return ""
		}
		return func() string {
			if __left == "Double" || __right == "Double" {
				return "Double"
			}
			return ____rune_private_1ed26dbc_inferCompilerCommonType(__left, __right)
		}()
	}()
}

func ____rune_private_1ed26dbc_inferCompilerCommonType(__left string, __right string) string {
	return func() string {
		if __left == __right {
			return __left
		}
		return ""
	}()
}

func ____rune_private_1ed26dbc_findCompilerCallable(__callables []__CompilerCallable, __name string, __index int) __CompilerCallable {
	__done := __index >= len(__callables)
	return func() __CompilerCallable {
		switch {
		case __done == true:
			return ____rune_private_1ed26dbc_emptyCompilerCallable()
		default:
			return ____rune_private_1ed26dbc_findCompilerCallableAt(__callables, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_findCompilerCallableAt(__callables []__CompilerCallable, __name string, __index int) __CompilerCallable {
	__matched := __callables[__index].__name == __name
	return func() __CompilerCallable {
		switch {
		case __matched == true:
			return __callables[__index]
		default:
			return ____rune_private_1ed26dbc_findCompilerCallable(__callables, __name, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerCallable(__name string, __arity int, __returnType string, __paramTypes []string, __private bool, __sourcePath string) __CompilerCallable {
	return __CompilerCallable{__name: __name, __arity: __arity, __returnType: __returnType, __paramTypes: __paramTypes, __private: __private, __sourcePath: __sourcePath}
}

func ____rune_private_1ed26dbc_emptyCompilerCallable() __CompilerCallable {
	return ____rune_private_1ed26dbc_compilerCallable("", 0, "", []string{}, false, "")
}

func ____rune_private_1ed26dbc_compilerParamTypeNames(__params []__IRParam) []string {
	__names := append([]string{}, []string{""}[0:0]...)
	for _, __param := range __params {
		_ = __param
		func() int { __names = append(__names, __param.__typeName); return len(__names) }()
	}
	return __names
}

func ____rune_private_1ed26dbc_compilerParamBindings(__params []__IRParam) []__CompilerTypeBinding {
	return ____rune_private_1ed26dbc_compilerFunctionBindings(__params, append([]__CompilerTypeBinding{}, []__CompilerTypeBinding{____rune_private_1ed26dbc_emptyCompilerTypeBinding()}[0:0]...))
}

func ____rune_private_1ed26dbc_compilerInitialBindings(__file __IRFile, __callables []__CompilerCallable) []__CompilerTypeBinding {
	__bindings := append([]__CompilerTypeBinding{}, []__CompilerTypeBinding{____rune_private_1ed26dbc_emptyCompilerTypeBinding()}[0:0]...)
	__bindings = ____rune_private_1ed26dbc_compilerMacroFunctionBindings(__file.__functions, __bindings)
	for _, __constant := range __file.__constants {
		_ = __constant
		__bindings = ____rune_private_1ed26dbc_compilerConstBinding(__file.__structs, __callables, __bindings, __constant)
	}
	for _, __importDecl := range __file.__tsImports {
		_ = __importDecl
		__bindings = ____rune_private_1ed26dbc_compilerImportValueBindings(__importDecl.__values, __bindings)
	}
	return __bindings
}

func ____rune_private_1ed26dbc_compilerMacroFunctionBindings(__functions []__IRFunction, __bindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	__out := __bindings
	for _, __fn := range __functions {
		_ = __fn
		__out = ____rune_private_1ed26dbc_compilerMacroFunctionBinding(__fn, __out)
	}
	return __out
}

func ____rune_private_1ed26dbc_compilerMacroFunctionBinding(__fn __IRFunction, __bindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	return func() []__CompilerTypeBinding {
		switch {
		case __fn.__macro == true:
			return ____rune_private_1ed26dbc_addCompilerTypeBinding(__bindings, __fn.__name, "MacroFunction")
		default:
			return __bindings
		}
	}()
}

func ____rune_private_1ed26dbc_compilerConstBinding(__structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __constant __IRConst) []__CompilerTypeBinding {
	return ____rune_private_1ed26dbc_addCompilerTypeBinding(__bindings, __constant.__name, ____rune_private_1ed26dbc_compilerConstBindingType(__structs, __callables, __bindings, __constant))
}

func ____rune_private_1ed26dbc_compilerConstBindingType(__structs []__IRStructType, __callables []__CompilerCallable, __bindings []__CompilerTypeBinding, __constant __IRConst) string {
	return func() string {
		switch {
		case __constant.__typeName == "":
			return ____rune_private_1ed26dbc_inferCompilerExprTypeWithStructs(__constant.__value, __structs, __callables, __bindings)
		default:
			return __constant.__typeName
		}
	}()
}

func ____rune_private_1ed26dbc_compilerImportValueBindings(__values []__IRConst, __bindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	__out := __bindings
	for _, __value := range __values {
		_ = __value
		__out = ____rune_private_1ed26dbc_addCompilerTypeBinding(__out, __value.__name, __value.__typeName)
	}
	return __out
}

func ____rune_private_1ed26dbc_compilerFunctionBindings(__params []__IRParam, __baseBindings []__CompilerTypeBinding) []__CompilerTypeBinding {
	__bindings := __baseBindings
	for _, __param := range __params {
		_ = __param
		__bindings = ____rune_private_1ed26dbc_addCompilerValueBinding(__bindings, __param.__name, __param.__typeName)
	}
	return __bindings
}

func ____rune_private_1ed26dbc_compilerTypeBinding(__name string, __typeName string) __CompilerTypeBinding {
	return __CompilerTypeBinding{__name: __name, __typeName: __typeName}
}

func ____rune_private_1ed26dbc_emptyCompilerTypeBinding() __CompilerTypeBinding {
	return ____rune_private_1ed26dbc_compilerTypeBinding("", "")
}

func ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings []__CompilerTypeBinding, __name string, __index int) __CompilerTypeBinding {
	__done := __index >= len(__bindings)
	return func() __CompilerTypeBinding {
		switch {
		case __done == true:
			return ____rune_private_1ed26dbc_emptyCompilerTypeBinding()
		default:
			return ____rune_private_1ed26dbc_findCompilerTypeBindingAt(__bindings, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_findCompilerTypeBindingAt(__bindings []__CompilerTypeBinding, __name string, __index int) __CompilerTypeBinding {
	__matched := __bindings[__index].__name == __name
	return func() __CompilerTypeBinding {
		switch {
		case __matched == true:
			return __bindings[__index]
		default:
			return ____rune_private_1ed26dbc_findCompilerTypeBinding(__bindings, __name, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_findCompilerStruct(__structs []__IRStructType, __name string, __index int) __IRStructType {
	__done := __index >= len(__structs)
	return func() __IRStructType {
		switch {
		case __done == true:
			return ____rune_private_1ed26dbc_emptyCompilerStruct()
		default:
			return ____rune_private_1ed26dbc_findCompilerStructAt(__structs, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_findCompilerStructAt(__structs []__IRStructType, __name string, __index int) __IRStructType {
	__matched := __structs[__index].__name == __name
	return func() __IRStructType {
		switch {
		case __matched == true:
			return __structs[__index]
		default:
			return ____rune_private_1ed26dbc_findCompilerStruct(__structs, __name, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_emptyCompilerStruct() __IRStructType {
	return __IRStructType{__name: "", __private: false, __generics: []string{}, __fields: []__IRField{}, __methods: []__IRFunction{}, __sourcePath: "", __line: 0, __column: 0}
}

func ____rune_private_1ed26dbc_findCompilerStructField(__fields []__IRField, __name string, __index int) __IRField {
	__done := __index >= len(__fields)
	return func() __IRField {
		switch {
		case __done == true:
			return ____rune_private_1ed26dbc_emptyCompilerField()
		default:
			return ____rune_private_1ed26dbc_findCompilerStructFieldAt(__fields, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_findCompilerStructFieldAt(__fields []__IRField, __name string, __index int) __IRField {
	__matched := __fields[__index].__name == __name
	return func() __IRField {
		switch {
		case __matched == true:
			return __fields[__index]
		default:
			return ____rune_private_1ed26dbc_findCompilerStructField(__fields, __name, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_emptyCompilerField() __IRField {
	return __IRField{__name: "", __private: false, __typeName: "", __jsonName: "", __jsonIgnore: false, __line: 0, __column: 0}
}

func ____rune_private_1ed26dbc_compilerContains(__values []string, __value string) bool {
	return ____rune_private_1ed26dbc_compilerContainsAt(__values, __value, 0)
}

func ____rune_private_1ed26dbc_compilerContainsAt(__values []string, __value string, __index int) bool {
	return func() bool {
		if __index >= len(__values) {
			return false
		}
		return func() bool {
			if __values[__index] == __value {
				return true
			}
			return ____rune_private_1ed26dbc_compilerContainsAt(__values, __value, __index+1)
		}()
	}()
}

func ____rune_private_1ed26dbc_compilerStructNameAppearsAfter(__structs []__IRStructType, __name string, __index int) bool {
	__done := __index >= len(__structs)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return func() bool {
				switch {
				case __structs[__index].__name == __name == true:
					return true
				default:
					return ____rune_private_1ed26dbc_compilerStructNameAppearsAfter(__structs, __name, __index+1)
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerEnumNameAppearsAfter(__enums []__IREnumType, __name string, __index int) bool {
	__done := __index >= len(__enums)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return func() bool {
				switch {
				case __enums[__index].__name == __name == true:
					return true
				default:
					return ____rune_private_1ed26dbc_compilerEnumNameAppearsAfter(__enums, __name, __index+1)
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerFunctionNameAppearsBefore(__functions []__IRFunction, __name string, __macro bool, __index int) bool {
	__done := __index < 0
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerFunctionNameAppearsBeforeAt(__functions, __name, __macro, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerFunctionNameAppearsBeforeAt(__functions []__IRFunction, __name string, __macro bool, __index int) bool {
	__matched := __functions[__index].__name == __name && __functions[__index].__macro == __macro
	return func() bool {
		switch {
		case __matched == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerFunctionNameAppearsBefore(__functions, __name, __macro, __index-1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerConstNameAppearsBefore(__constants []__IRConst, __name string, __index int) bool {
	__done := __index < 0
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerConstNameAppearsBeforeAt(__constants, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerConstNameAppearsBeforeAt(__constants []__IRConst, __name string, __index int) bool {
	__matched := __constants[__index].__name == __name
	return func() bool {
		switch {
		case __matched == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerConstNameAppearsBefore(__constants, __name, __index-1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerFieldNameAppearsBefore(__fields []__IRField, __name string, __index int) bool {
	__done := __index < 0
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return func() bool {
				switch {
				case __fields[__index].__name == __name == true:
					return true
				default:
					return ____rune_private_1ed26dbc_compilerFieldNameAppearsBefore(__fields, __name, __index-1)
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMethodNameAppearsBefore(__methods []__IRFunction, __name string, __index int) bool {
	__done := __index < 0
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return func() bool {
				switch {
				case __methods[__index].__name == __name == true:
					return true
				default:
					return ____rune_private_1ed26dbc_compilerMethodNameAppearsBefore(__methods, __name, __index-1)
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerEnumMemberNameAppearsBefore(__members []__IREnumMember, __name string, __index int) bool {
	__done := __index < 0
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return func() bool {
				switch {
				case __members[__index].__name == __name == true:
					return true
				default:
					return ____rune_private_1ed26dbc_compilerEnumMemberNameAppearsBefore(__members, __name, __index-1)
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParamNameAppearsBefore(__params []__IRParam, __name string, __index int) bool {
	__done := __index < 0
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return func() bool {
				switch {
				case __params[__index].__name == __name == true:
					return true
				default:
					return ____rune_private_1ed26dbc_compilerParamNameAppearsBefore(__params, __name, __index-1)
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_lowerFiles(__files []__SourceFile) __IRFile {
	return func() __IRFile {
		if len(__files) == 0 {
			return ____rune_private_1ed26dbc_emptyCompilerIRFile()
		}
		return ____rune_private_1ed26dbc_lowerReachableRuneFiles(__files, []string{__files[0].__path}, []string{}, ____rune_private_1ed26dbc_emptyCompilerIRFile())
	}()
}

func ____rune_private_1ed26dbc_lowerCompilerSource(__source string) __IRFile {
	return __lowerParsed(____rune_private_1ed26dbc_expandCompilerMacros(__parse(__source)))
}

func ____rune_private_1ed26dbc_lowerCompilerSourceWithPath(__source string, __sourcePath string) __IRFile {
	return __withIRFileSourcePath(____rune_private_1ed26dbc_lowerCompilerSource(__source), __sourcePath)
}

func ____rune_private_1ed26dbc_expandCompilerMacros(__file __ParsedFile) __ParsedFile {
	__imports := ____rune_private_1ed26dbc_compilerMergeParsedImports(__file.__imports, ____rune_private_1ed26dbc_compilerImportExpressions(__file), 0)
	__errors := ____rune_private_1ed26dbc_compilerMacroErrors(__file, __file.__errors)
	__out := __ParsedFile{__imports: __imports, __constants: __file.__constants, __types: []__ParsedType{}, __functions: []__ParsedFunction{}, __tests: []__ParsedTest{}, __errors: __errors}
	for _, __typeDecl := range __file.__types {
		_ = __typeDecl
		func() int {
			__out.__types = append(__out.__types, ____rune_private_1ed26dbc_expandCompilerTypeMacros(__typeDecl))
			return len(__out.__types)
		}()
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		func() int {
			__out.__functions = append(__out.__functions, ____rune_private_1ed26dbc_expandCompilerFunctionMacros(__fn))
			return len(__out.__functions)
		}()
	}
	for _, __testDecl := range __file.__tests {
		_ = __testDecl
		func() int {
			__out.__tests = append(__out.__tests, ____rune_private_1ed26dbc_expandCompilerTestMacros(__testDecl))
			return len(__out.__tests)
		}()
	}
	return __out
}

func ____rune_private_1ed26dbc_expandCompilerTypeMacros(__typeDecl __ParsedType) __ParsedType {
	__typeName := ____rune_private_1ed26dbc_compilerRenameDeclarationName(__typeDecl.__annotations, __typeDecl.__name)
	__fields := append([]__ParsedField{}, __typeDecl.__fields[0:0]...)
	__methods := append([]__ParsedFunction{}, __typeDecl.__methods[0:0]...)
	__members := append([]__ParsedEnumMember{}, __typeDecl.__members[0:0]...)
	for _, __field := range __typeDecl.__fields {
		_ = __field
		func() int {
			__fields = append(__fields, ____rune_private_1ed26dbc_expandCompilerFieldMacros(__field))
			return len(__fields)
		}()
	}
	for _, __method := range __typeDecl.__methods {
		_ = __method
		func() int {
			__methods = append(__methods, ____rune_private_1ed26dbc_expandCompilerFunctionMacros(__method))
			return len(__methods)
		}()
	}
	for _, __member := range __typeDecl.__members {
		_ = __member
		func() int {
			__members = append(__members, ____rune_private_1ed26dbc_expandCompilerEnumMemberMacros(__member))
			return len(__members)
		}()
	}
	return __ParsedType{__name: __typeName, __private: __typeDecl.__private, __enum: __typeDecl.__enum, __annotations: __typeDecl.__annotations, __generics: __typeDecl.__generics, __fields: __fields, __methods: ____rune_private_1ed26dbc_expandCompilerTypeMacroMethods(__typeDecl, __typeName, __methods), __members: __members, __line: __typeDecl.__line, __column: __typeDecl.__column}
}

func ____rune_private_1ed26dbc_expandCompilerTypeMacroMethods(__typeDecl __ParsedType, __typeName string, __methods []__ParsedFunction) []__ParsedFunction {
	return func() []__ParsedFunction {
		switch {
		case __typeDecl.__enum == true:
			return __methods
		default:
			return ____rune_private_1ed26dbc_expandCompilerStructMacroMethods(__typeDecl, __typeName, __methods)
		}
	}()
}

func ____rune_private_1ed26dbc_expandCompilerStructMacroMethods(__typeDecl __ParsedType, __typeName string, __methods []__ParsedFunction) []__ParsedFunction {
	__shouldAddFromJson := ____rune_private_1ed26dbc_compilerHasAnnotation(__typeDecl.__annotations, "#", "json", "object") && ____rune_private_1ed26dbc_compilerHasFromJsonMethod(__methods, 0) == false
	return func() []__ParsedFunction {
		switch {
		case __shouldAddFromJson == true:
			return func() []__ParsedFunction {
				out := []__ParsedFunction{}
				out = append(out, __methods...)
				out = append(out, ____rune_private_1ed26dbc_compilerJsonFromJsonMethod(__typeDecl, __typeName))
				return out
			}()
		default:
			return __methods
		}
	}()
}

func ____rune_private_1ed26dbc_compilerJsonFromJsonMethod(__typeDecl __ParsedType, __typeName string) __ParsedFunction {
	return __ParsedFunction{__name: "fromJson", __private: false, __static: true, __routine: false, __macro: false, __annotations: ____rune_private_1ed26dbc_compilerEmptyAnnotations(), __receiverType: __typeName, __generics: []string{}, __params: []__ParsedParam{__ParsedParam{__name: "text", __typeRef: ____rune_private_1ed26dbc_compilerTypeRef("String", __typeDecl.__line, __typeDecl.__column), __line: __typeDecl.__line, __column: __typeDecl.__column}}, __returnType: ____rune_private_1ed26dbc_compilerGenericTypeRef(__typeName, __typeDecl.__generics, __typeDecl.__line, __typeDecl.__column), __body: ____rune_private_1ed26dbc_compilerJsonParseExpr(__typeDecl.__line, __typeDecl.__column), __line: __typeDecl.__line, __column: __typeDecl.__column}
}

func ____rune_private_1ed26dbc_compilerJsonParseExpr(__line int, __column int) __ParsedExpr {
	__jsonModule := ____rune_private_1ed26dbc_compilerParsedExpr(__ExprKind_At, "@", "json", "", "", []__ParsedParam{}, []__ParsedExpr{}, __line, __column)
	__parseSelector := ____rune_private_1ed26dbc_compilerParsedExpr(__ExprKind_Selector, "parse", "parse", "", ".", []__ParsedParam{}, []__ParsedExpr{__jsonModule}, __line, __column)
	__textArg := ____rune_private_1ed26dbc_compilerParsedExpr(__ExprKind_Identifier, "text", "text", "", "", []__ParsedParam{}, []__ParsedExpr{}, __line, __column)
	return ____rune_private_1ed26dbc_compilerParsedExpr(__ExprKind_Call, "parse", "", "", "", []__ParsedParam{}, []__ParsedExpr{__parseSelector, __textArg}, __line, __column)
}

func ____rune_private_1ed26dbc_compilerEmptyAnnotations() []__ParsedAnnotation {
	return append([]__ParsedAnnotation{}, []__ParsedAnnotation{__ParsedAnnotation{__marker: "", __module: "", __name: "", __args: []__ParsedExpr{}, __line: 0, __column: 0}}[0:0]...)
}

func ____rune_private_1ed26dbc_compilerParsedExpr(__kind __ExprKind, __text string, __name string, __value string, __op string, __params []__ParsedParam, __children []__ParsedExpr, __line int, __column int) __ParsedExpr {
	return __ParsedExpr{__kind: __kind, __text: __text, __name: __name, __value: __value, __op: __op, __params: __params, __children: __children, __line: __line, __column: __column}
}

func ____rune_private_1ed26dbc_compilerGenericTypeRef(__name string, __generics []string, __line int, __column int) __ParsedTypeRef {
	__args := ____rune_private_1ed26dbc_compilerGenericTypeRefArgs(__generics, 0, __line, __column, append([]__ParsedTypeRef{}, []__ParsedTypeRef{__emptyParsedTypeRef()}[0:0]...))
	return ____rune_private_1ed26dbc_compilerTypeRefWithArgs(__name, __args, __line, __column)
}

func ____rune_private_1ed26dbc_compilerGenericTypeRefArgs(__generics []string, __index int, __line int, __column int, __out []__ParsedTypeRef) []__ParsedTypeRef {
	__done := __index >= len(__generics)
	return func() []__ParsedTypeRef {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_compilerGenericTypeRefArgsStep(__generics, __index, __line, __column, __out)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerGenericTypeRefArgsStep(__generics []string, __index int, __line int, __column int, __out []__ParsedTypeRef) []__ParsedTypeRef {
	__out = append(__out, ____rune_private_1ed26dbc_compilerTypeRef(__generics[__index], __line, __column))
	return ____rune_private_1ed26dbc_compilerGenericTypeRefArgs(__generics, __index+1, __line, __column, __out)
}

func ____rune_private_1ed26dbc_compilerTypeRef(__name string, __line int, __column int) __ParsedTypeRef {
	return ____rune_private_1ed26dbc_compilerTypeRefWithArgs(__name, append([]__ParsedTypeRef{}, []__ParsedTypeRef{__emptyParsedTypeRef()}[0:0]...), __line, __column)
}

func ____rune_private_1ed26dbc_compilerTypeRefWithArgs(__name string, __args []__ParsedTypeRef, __line int, __column int) __ParsedTypeRef {
	return __ParsedTypeRef{__kind: __TypeRefKind_Name, __name: __name, __module: "", __nullable: false, __args: __args, __params: []__ParsedTypeParam{}, __returnTypes: []__ParsedTypeRef{}, __line: __line, __column: __column}
}

func ____rune_private_1ed26dbc_compilerHasFromJsonMethod(__methods []__ParsedFunction, __index int) bool {
	__done := __index >= len(__methods)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerHasFromJsonMethodAt(__methods, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerHasFromJsonMethodAt(__methods []__ParsedFunction, __index int) bool {
	__matched := __methods[__index].__name == "fromJson"
	return func() bool {
		switch {
		case __matched == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerHasFromJsonMethod(__methods, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerHasAnnotation(__annotations []__ParsedAnnotation, __marker string, __module string, __name string) bool {
	return ____rune_private_1ed26dbc_compilerHasAnnotationAt(__annotations, __marker, __module, __name, 0)
}

func ____rune_private_1ed26dbc_compilerHasAnnotationAt(__annotations []__ParsedAnnotation, __marker string, __module string, __name string, __index int) bool {
	__done := __index >= len(__annotations)
	return func() bool {
		switch {
		case __done == true:
			return false
		default:
			return ____rune_private_1ed26dbc_compilerHasAnnotationStep(__annotations, __marker, __module, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerHasAnnotationStep(__annotations []__ParsedAnnotation, __marker string, __module string, __name string, __index int) bool {
	__annotation := __annotations[__index]
	__matched := __annotation.__marker == __marker && __annotation.__module == __module && __annotation.__name == __name
	return func() bool {
		switch {
		case __matched == true:
			return true
		default:
			return ____rune_private_1ed26dbc_compilerHasAnnotationAt(__annotations, __marker, __module, __name, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_expandCompilerFieldMacros(__field __ParsedField) __ParsedField {
	return __ParsedField{__name: ____rune_private_1ed26dbc_compilerRenameDeclarationName(__field.__annotations, __field.__name), __private: __field.__private, __annotations: __field.__annotations, __typeRef: __field.__typeRef, __line: __field.__line, __column: __field.__column}
}

func ____rune_private_1ed26dbc_expandCompilerEnumMemberMacros(__member __ParsedEnumMember) __ParsedEnumMember {
	return __ParsedEnumMember{__name: ____rune_private_1ed26dbc_compilerRenameDeclarationName(__member.__annotations, __member.__name), __private: __member.__private, __annotations: __member.__annotations, __value: __member.__value, __params: __member.__params, __line: __member.__line, __column: __member.__column}
}

func ____rune_private_1ed26dbc_expandCompilerFunctionMacros(__fn __ParsedFunction) __ParsedFunction {
	return __ParsedFunction{__name: ____rune_private_1ed26dbc_compilerRenameDeclarationName(__fn.__annotations, __fn.__name), __private: __fn.__private, __static: __fn.__static, __routine: __fn.__routine, __macro: __fn.__macro, __annotations: __fn.__annotations, __receiverType: __fn.__receiverType, __generics: __fn.__generics, __params: __fn.__params, __returnType: __fn.__returnType, __body: ____rune_private_1ed26dbc_expandCompilerNamespaceAliases(__fn.__body, append([]__CompilerNamespaceAlias{}, []__CompilerNamespaceAlias{____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()}[0:0]...)), __line: __fn.__line, __column: __fn.__column}
}

func ____rune_private_1ed26dbc_compilerMacroErrors(__file __ParsedFile, __errors []__ParseError) []__ParseError {
	__functionErrors := ____rune_private_1ed26dbc_compilerMacroFunctionErrors(__file.__functions, 0, __errors)
	__methodErrors := ____rune_private_1ed26dbc_compilerMacroMethodErrors(__file.__types, 0, __functionErrors)
	return ____rune_private_1ed26dbc_compilerAnnotationErrors(__file, __methodErrors)
}

func ____rune_private_1ed26dbc_compilerMacroMethodErrors(__types []__ParsedType, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__types)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerMacroMethodErrors(__types, __index+1, ____rune_private_1ed26dbc_compilerMacroTypeMethodErrors(__types[__index].__methods, 0, __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMacroTypeMethodErrors(__methods []__ParsedFunction, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__methods)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerMacroTypeMethodErrors(__methods, __index+1, ____rune_private_1ed26dbc_compilerMacroTypeMethodError(__methods[__index], __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMacroTypeMethodError(__method __ParsedFunction, __errors []__ParseError) []__ParseError {
	return func() []__ParseError {
		switch {
		case __method.__macro == true:
			return func() []__ParseError {
				out := []__ParseError{}
				out = append(out, __errors...)
				out = append(out, ____rune_private_1ed26dbc_compilerParseError("macro declarations must be top-level functions", __method.__line, __method.__column))
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMacroFunctionErrors(__functions []__ParsedFunction, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__functions)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerMacroFunctionErrors(__functions, __index+1, ____rune_private_1ed26dbc_compilerMacroFunctionError(__functions[__index], __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMacroFunctionError(__fn __ParsedFunction, __errors []__ParseError) []__ParseError {
	return func() []__ParseError {
		switch {
		case __fn.__macro == true:
			return func() []__ParseError {
				__next := func() []__ParseError {
					switch {
					case ____rune_private_1ed26dbc_compilerSyntaxMacroSignatureOk(__fn) == true:
						return __errors
					default:
						return func() []__ParseError {
							out := []__ParseError{}
							out = append(out, __errors...)
							out = append(out, ____rune_private_1ed26dbc_compilerParseError("macro "+__fn.__name+" must accept SyntaxFile and MacroContext first and return SyntaxFile", __fn.__line, __fn.__column))
							return out
						}()
					}
				}()
				return ____rune_private_1ed26dbc_compilerMacroPurityError(__fn, __next)
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMacroPurityError(__fn __ParsedFunction, __errors []__ParseError) []__ParseError {
	return func() []__ParseError {
		__match7 := ____rune_private_1ed26dbc_compilerParsedMacroPurityMessage(__fn.__body)
		switch {
		case __match7 == "":
			return __errors
		case true:
			__message := __match7
			_ = __message
			return func() []__ParseError {
				out := []__ParseError{}
				out = append(out, __errors...)
				out = append(out, ____rune_private_1ed26dbc_compilerParseError("macro "+__fn.__name+" is not pure: "+__message, __fn.__line, __fn.__column))
				return out
			}()
		}
		return nil
	}()
}

func ____rune_private_1ed26dbc_compilerParsedMacroPurityMessage(__expr __ParsedExpr) string {
	__current := ____rune_private_1ed26dbc_compilerParsedMacroCallPurityMessage(__expr)
	return func() string {
		switch {
		case __current == "":
			return ____rune_private_1ed26dbc_compilerParsedMacroChildPurityMessage(__expr.__children, 0)
		default:
			return __current
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedMacroChildPurityMessage(__children []__ParsedExpr, __index int) string {
	__done := __index >= len(__children)
	return func() string {
		switch {
		case __done == true:
			return ""
		default:
			return ____rune_private_1ed26dbc_compilerParsedMacroChildPurityMessageAt(__children, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedMacroChildPurityMessageAt(__children []__ParsedExpr, __index int) string {
	__message := ____rune_private_1ed26dbc_compilerParsedMacroPurityMessage(__children[__index])
	return func() string {
		switch {
		case __message == "":
			return ____rune_private_1ed26dbc_compilerParsedMacroChildPurityMessage(__children, __index+1)
		default:
			return __message
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedMacroCallPurityMessage(__expr __ParsedExpr) string {
	return func() string {
		switch {
		case __expr.__kind == __ExprKind_Call:
			return ____rune_private_1ed26dbc_compilerParsedCallPurityMessage(__expr)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedCallPurityMessage(__expr __ParsedExpr) string {
	return func() string {
		switch {
		case len(__expr.__children) == 0:
			return "cannot prove dynamic call is pure"
		default:
			return ____rune_private_1ed26dbc_compilerParsedCalleePurityMessage(__expr.__children[0])
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedCalleePurityMessage(__callee __ParsedExpr) string {
	return func() string {
		switch {
		case __callee.__kind == __ExprKind_Selector:
			return ____rune_private_1ed26dbc_compilerParsedSelectorPurityMessage(__callee)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedSelectorPurityMessage(__selector __ParsedExpr) string {
	return func() string {
		switch {
		case len(__selector.__children) == 0:
			return ""
		default:
			return ____rune_private_1ed26dbc_compilerParsedAtSelectorPurityMessage(__selector, __selector.__children[0])
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedAtSelectorPurityMessage(__selector __ParsedExpr, __receiver __ParsedExpr) string {
	return func() string {
		switch {
		case __receiver.__kind == __ExprKind_At:
			return ____rune_private_1ed26dbc_compilerParsedModuleSelectorPurityMessage(__receiver.__name, __selector.__name)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedModuleSelectorPurityMessage(__module string, __name string) string {
	return func() string {
		switch {
		case __module == "io":
			return ____rune_private_1ed26dbc_compilerParsedIOSelectorPurityMessage(__name)
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedIOSelectorPurityMessage(__name string) string {
	return func() string {
		switch {
		case __name == "println":
			return "calls impure function @io.println"
		default:
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_compilerSyntaxMacroSignatureOk(__fn __ParsedFunction) bool {
	__returnOk := __typeRefToString(__fn.__returnType) == "SyntaxFile"
	__paramsOk := len(__fn.__params) >= 2 && (__typeRefToString(__fn.__params[0].__typeRef) == "SyntaxFile" && __typeRefToString(__fn.__params[1].__typeRef) == "MacroContext")
	return __returnOk && __paramsOk
}

func ____rune_private_1ed26dbc_compilerAnnotationErrors(__file __ParsedFile, __errors []__ParseError) []__ParseError {
	__next := ____rune_private_1ed26dbc_compilerTypeAnnotationErrors(__file.__types, __file.__functions, 0, __errors)
	return ____rune_private_1ed26dbc_compilerFunctionAnnotationErrors(__file.__functions, __file.__functions, 0, __next)
}

func ____rune_private_1ed26dbc_compilerTypeAnnotationErrors(__types []__ParsedType, __functions []__ParsedFunction, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__types)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerTypeAnnotationErrors(__types, __functions, __index+1, ____rune_private_1ed26dbc_compilerTypeAnnotationError(__types[__index], __functions, __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerTypeAnnotationError(__typeDecl __ParsedType, __functions []__ParsedFunction, __errors []__ParseError) []__ParseError {
	__next := ____rune_private_1ed26dbc_compilerAnnotationListErrors(__typeDecl.__annotations, __functions, 0, __errors)
	__next = ____rune_private_1ed26dbc_compilerFieldAnnotationErrors(__typeDecl.__fields, __functions, 0, __next)
	__next = ____rune_private_1ed26dbc_compilerEnumMemberAnnotationErrors(__typeDecl.__members, __functions, 0, __next)
	return ____rune_private_1ed26dbc_compilerFunctionAnnotationErrors(__typeDecl.__methods, __functions, 0, __next)
}

func ____rune_private_1ed26dbc_compilerFieldAnnotationErrors(__fields []__ParsedField, __functions []__ParsedFunction, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__fields)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerFieldAnnotationErrors(__fields, __functions, __index+1, ____rune_private_1ed26dbc_compilerAnnotationListErrors(__fields[__index].__annotations, __functions, 0, __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerEnumMemberAnnotationErrors(__members []__ParsedEnumMember, __functions []__ParsedFunction, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__members)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerEnumMemberAnnotationErrors(__members, __functions, __index+1, ____rune_private_1ed26dbc_compilerAnnotationListErrors(__members[__index].__annotations, __functions, 0, __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerFunctionAnnotationErrors(__functions []__ParsedFunction, __topLevelFunctions []__ParsedFunction, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__functions)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerFunctionAnnotationErrors(__functions, __topLevelFunctions, __index+1, ____rune_private_1ed26dbc_compilerAnnotationListErrors(__functions[__index].__annotations, __topLevelFunctions, 0, __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerAnnotationListErrors(__annotations []__ParsedAnnotation, __functions []__ParsedFunction, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__annotations)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerAnnotationListErrors(__annotations, __functions, __index+1, ____rune_private_1ed26dbc_compilerAnnotationError(__annotations[__index], __functions, __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerAnnotationError(__annotation __ParsedAnnotation, __functions []__ParsedFunction, __errors []__ParseError) []__ParseError {
	return func() []__ParseError {
		switch {
		case __annotation.__marker == "#":
			return ____rune_private_1ed26dbc_compilerHashAnnotationError(__annotation, __functions, __errors)
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_compilerHashAnnotationError(__annotation __ParsedAnnotation, __functions []__ParsedFunction, __errors []__ParseError) []__ParseError {
	return func() []__ParseError {
		switch {
		case __annotation.__module == "":
			return ____rune_private_1ed26dbc_compilerLocalAnnotationError(__annotation, __functions, __errors)
		default:
			return ____rune_private_1ed26dbc_compilerModuleAnnotationError(__annotation, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerLocalAnnotationError(__annotation __ParsedAnnotation, __functions []__ParsedFunction, __errors []__ParseError) []__ParseError {
	return func() []__ParseError {
		switch {
		case __annotation.__name == "":
			return __errors
		case __annotation.__name == "alias":
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerResolvedLocalAnnotationError(__annotation, ____rune_private_1ed26dbc_findCompilerMacroBinding(__functions, __annotation.__name, 0), __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerResolvedLocalAnnotationError(__annotation __ParsedAnnotation, __binding __CompilerMacroBinding, __errors []__ParseError) []__ParseError {
	return func() []__ParseError {
		switch {
		case __binding.__name == "":
			return func() []__ParseError {
				out := []__ParseError{}
				out = append(out, __errors...)
				out = append(out, ____rune_private_1ed26dbc_compilerParseError("unknown macro #"+__annotation.__name, __annotation.__line, __annotation.__column))
				return out
			}()
		default:
			return func() []__ParseError {
				switch {
				case __binding.__macro == true:
					return ____rune_private_1ed26dbc_compilerAnnotationArgErrors(__annotation, __binding, __errors)
				default:
					return func() []__ParseError {
						out := []__ParseError{}
						out = append(out, __errors...)
						out = append(out, ____rune_private_1ed26dbc_compilerParseError("#"+__annotation.__name+" refers to a function that is not a macro", __annotation.__line, __annotation.__column))
						return out
					}()
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerModuleAnnotationError(__annotation __ParsedAnnotation, __errors []__ParseError) []__ParseError {
	__binding := ____rune_private_1ed26dbc_compilerBuiltinMacroBinding(__annotation.__module, __annotation.__name)
	return func() []__ParseError {
		switch {
		case __binding.__name == "":
			return ____rune_private_1ed26dbc_compilerUnknownModuleMacroError(__annotation, __errors)
		default:
			return ____rune_private_1ed26dbc_compilerAnnotationArgErrors(__annotation, __binding, __errors)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerUnknownModuleMacroError(__annotation __ParsedAnnotation, __errors []__ParseError) []__ParseError {
	return func() []__ParseError {
		switch {
		case ____rune_private_1ed26dbc_compilerKnownOrdinaryAnnotationFunction(__annotation.__module, __annotation.__name) == true:
			return func() []__ParseError {
				out := []__ParseError{}
				out = append(out, __errors...)
				out = append(out, ____rune_private_1ed26dbc_compilerParseError("#"+__annotation.__module+"."+__annotation.__name+" refers to a function that is not a macro", __annotation.__line, __annotation.__column))
				return out
			}()
		default:
			return func() []__ParseError {
				out := []__ParseError{}
				out = append(out, __errors...)
				out = append(out, ____rune_private_1ed26dbc_compilerParseError("unknown macro #"+__annotation.__module+"."+__annotation.__name, __annotation.__line, __annotation.__column))
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerAnnotationArgErrors(__annotation __ParsedAnnotation, __binding __CompilerMacroBinding, __errors []__ParseError) []__ParseError {
	__arityOk := len(__annotation.__args) == len(__binding.__paramTypes)
	return func() []__ParseError {
		switch {
		case __arityOk == true:
			return ____rune_private_1ed26dbc_compilerAnnotationArgTypeErrors(__annotation, __binding, 0, __errors)
		default:
			return func() []__ParseError {
				out := []__ParseError{}
				out = append(out, __errors...)
				out = append(out, ____rune_private_1ed26dbc_compilerParseError("function \""+__binding.__name+"\" expects "+__compilerIntToString(len(__binding.__paramTypes))+" args, got "+__compilerIntToString(len(__annotation.__args)), __annotation.__line, __annotation.__column))
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerAnnotationArgTypeErrors(__annotation __ParsedAnnotation, __binding __CompilerMacroBinding, __index int, __errors []__ParseError) []__ParseError {
	__done := __index >= len(__binding.__paramTypes)
	return func() []__ParseError {
		switch {
		case __done == true:
			return __errors
		default:
			return ____rune_private_1ed26dbc_compilerAnnotationArgTypeErrors(__annotation, __binding, __index+1, ____rune_private_1ed26dbc_compilerAnnotationArgTypeError(__annotation, __binding, __index, __errors))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerAnnotationArgTypeError(__annotation __ParsedAnnotation, __binding __CompilerMacroBinding, __index int, __errors []__ParseError) []__ParseError {
	__expected := __binding.__paramTypes[__index]
	__actual := ____rune_private_1ed26dbc_compilerParsedAnnotationArgType(__annotation.__args[__index])
	__shouldCheck := ____rune_private_1ed26dbc_compilerShouldCheckArgType(__expected, __actual)
	__mismatch := __shouldCheck && ____rune_private_1ed26dbc_compilerTypesCompatible(__expected, __actual) == false
	return func() []__ParseError {
		switch {
		case __mismatch == true:
			return func() []__ParseError {
				out := []__ParseError{}
				out = append(out, __errors...)
				out = append(out, ____rune_private_1ed26dbc_compilerParseError(____rune_private_1ed26dbc_compilerArgumentTypeError(__binding.__name, __index+1, __actual, __expected), __annotation.__line, __annotation.__column))
				return out
			}()
		default:
			return __errors
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedAnnotationArgType(__expr __ParsedExpr) string {
	return func() string {
		switch {
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
			return ""
		}
	}()
}

func ____rune_private_1ed26dbc_findCompilerMacroBinding(__functions []__ParsedFunction, __name string, __index int) __CompilerMacroBinding {
	__done := __index >= len(__functions)
	return func() __CompilerMacroBinding {
		switch {
		case __done == true:
			return ____rune_private_1ed26dbc_emptyCompilerMacroBinding()
		default:
			return ____rune_private_1ed26dbc_findCompilerMacroBindingAt(__functions, __name, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_findCompilerMacroBindingAt(__functions []__ParsedFunction, __name string, __index int) __CompilerMacroBinding {
	__matched := __functions[__index].__name == __name
	return func() __CompilerMacroBinding {
		switch {
		case __matched == true:
			return ____rune_private_1ed26dbc_compilerMacroBindingFromFunction(__functions[__index])
		default:
			return ____rune_private_1ed26dbc_findCompilerMacroBinding(__functions, __name, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMacroBindingFromFunction(__fn __ParsedFunction) __CompilerMacroBinding {
	return __CompilerMacroBinding{__name: __fn.__name, __macro: __fn.__macro, __paramTypes: func() []string {
		switch {
		case __fn.__macro == true:
			return ____rune_private_1ed26dbc_compilerVisibleMacroParamTypes(__fn.__params, 2, append([]string{}, []string{""}[0:0]...))
		default:
			return ____rune_private_1ed26dbc_compilerParsedParamTypeNames(__fn.__params, 0, append([]string{}, []string{""}[0:0]...))
		}
	}()}
}

func ____rune_private_1ed26dbc_compilerVisibleMacroParamTypes(__params []__ParsedParam, __index int, __out []string) []string {
	__done := __index >= len(__params)
	return func() []string {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_compilerVisibleMacroParamTypes(__params, __index+1, func() []string {
				out := []string{}
				out = append(out, __out...)
				out = append(out, __typeRefToString(__params[__index].__typeRef))
				return out
			}())
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedParamTypeNames(__params []__ParsedParam, __index int, __out []string) []string {
	__done := __index >= len(__params)
	return func() []string {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_compilerParsedParamTypeNames(__params, __index+1, func() []string {
				out := []string{}
				out = append(out, __out...)
				out = append(out, __typeRefToString(__params[__index].__typeRef))
				return out
			}())
		}
	}()
}

func ____rune_private_1ed26dbc_compilerBuiltinMacroBinding(__module string, __name string) __CompilerMacroBinding {
	return func() __CompilerMacroBinding {
		switch {
		case __module == "macro":
			return ____rune_private_1ed26dbc_compilerMacroModuleBinding(__name)
		case __module == "json":
			return ____rune_private_1ed26dbc_compilerJsonModuleMacroBinding(__name)
		case __module == "cli":
			return ____rune_private_1ed26dbc_compilerCliModuleMacroBinding(__name)
		default:
			return ____rune_private_1ed26dbc_emptyCompilerMacroBinding()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMacroModuleBinding(__name string) __CompilerMacroBinding {
	return func() __CompilerMacroBinding {
		switch {
		case __name == "renameDeclaration":
			return ____rune_private_1ed26dbc_compilerMacroBinding("renameDeclaration", true, []string{"String"})
		default:
			return ____rune_private_1ed26dbc_emptyCompilerMacroBinding()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerJsonModuleMacroBinding(__name string) __CompilerMacroBinding {
	return func() __CompilerMacroBinding {
		switch {
		case __name == "object":
			return ____rune_private_1ed26dbc_compilerMacroBinding("object", true, []string{})
		case __name == "name":
			return ____rune_private_1ed26dbc_compilerMacroBinding("name", true, []string{"String"})
		case __name == "ignore":
			return ____rune_private_1ed26dbc_compilerMacroBinding("ignore", true, []string{})
		default:
			return ____rune_private_1ed26dbc_emptyCompilerMacroBinding()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerCliModuleMacroBinding(__name string) __CompilerMacroBinding {
	return func() __CompilerMacroBinding {
		switch {
		case __name == "command":
			return ____rune_private_1ed26dbc_compilerMacroBinding("command", true, []string{"String", "String", "String"})
		case __name == "flag":
			return ____rune_private_1ed26dbc_compilerMacroBinding("flag", true, []string{"String", "String"})
		case __name == "option":
			return ____rune_private_1ed26dbc_compilerMacroBinding("option", true, []string{"String", "String", "String", "String"})
		case __name == "arg":
			return ____rune_private_1ed26dbc_compilerMacroBinding("arg", true, []string{"String"})
		case __name == "parser":
			return ____rune_private_1ed26dbc_compilerMacroBinding("parser", true, []string{"String", "String", "String", "String", "String"})
		case __name == "main":
			return ____rune_private_1ed26dbc_compilerMacroBinding("main", true, []string{})
		default:
			return ____rune_private_1ed26dbc_emptyCompilerMacroBinding()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerKnownOrdinaryAnnotationFunction(__module string, __name string) bool {
	return func() bool {
		switch {
		case __module == "json":
			return __name == "parse" || __name == "stringify"
		case __module == "go":
			return __name == "import" || __name == "stmt" || __name == "expr"
		default:
			return false
		}
	}()
}

func ____rune_private_1ed26dbc_compilerMacroBinding(__name string, __macro bool, __paramTypes []string) __CompilerMacroBinding {
	return __CompilerMacroBinding{__name: __name, __macro: __macro, __paramTypes: __paramTypes}
}

func ____rune_private_1ed26dbc_emptyCompilerMacroBinding() __CompilerMacroBinding {
	return ____rune_private_1ed26dbc_compilerMacroBinding("", false, []string{})
}

func ____rune_private_1ed26dbc_compilerParseError(__message string, __line int, __column int) __ParseError {
	return __ParseError{__message: __message, __line: __line, __column: __column}
}

func ____rune_private_1ed26dbc_expandCompilerTestMacros(__testDecl __ParsedTest) __ParsedTest {
	return __ParsedTest{__name: __testDecl.__name, __body: ____rune_private_1ed26dbc_expandCompilerNamespaceAliases(__testDecl.__body, append([]__CompilerNamespaceAlias{}, []__CompilerNamespaceAlias{____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()}[0:0]...)), __line: __testDecl.__line, __column: __testDecl.__column}
}

func ____rune_private_1ed26dbc_compilerImportExpressions(__file __ParsedFile) []__ParsedImport {
	__imports := append([]__ParsedImport{}, []__ParsedImport{____rune_private_1ed26dbc_compilerEmptyParsedImport()}[0:0]...)
	for _, __typeDecl := range __file.__types {
		_ = __typeDecl
		__imports = ____rune_private_1ed26dbc_compilerCollectTypeImportExprs(__typeDecl, __imports)
	}
	for _, __fn := range __file.__functions {
		_ = __fn
		__imports = ____rune_private_1ed26dbc_compilerCollectFunctionImportExprs(__fn, __imports)
	}
	for _, __testDecl := range __file.__tests {
		_ = __testDecl
		__imports = ____rune_private_1ed26dbc_compilerCollectImportExprs(__testDecl.__body, __imports)
	}
	return __imports
}

func ____rune_private_1ed26dbc_compilerCollectTypeImportExprs(__typeDecl __ParsedType, __imports []__ParsedImport) []__ParsedImport {
	__out := __imports
	for _, __method := range __typeDecl.__methods {
		_ = __method
		__out = ____rune_private_1ed26dbc_compilerCollectFunctionImportExprs(__method, __out)
	}
	return __out
}

func ____rune_private_1ed26dbc_compilerCollectFunctionImportExprs(__fn __ParsedFunction, __imports []__ParsedImport) []__ParsedImport {
	return ____rune_private_1ed26dbc_compilerCollectImportExprs(__fn.__body, __imports)
}

func ____rune_private_1ed26dbc_compilerCollectImportExprs(__expr __ParsedExpr, __imports []__ParsedImport) []__ParsedImport {
	__next := func() []__ParsedImport {
		switch {
		case __expr.__kind == __ExprKind_At:
			return ____rune_private_1ed26dbc_compilerAppendImportExpr(__imports, __expr)
		default:
			return __imports
		}
	}()
	for _, __child := range __expr.__children {
		_ = __child
		__next = ____rune_private_1ed26dbc_compilerCollectImportExprs(__child, __next)
	}
	return __next
}

func ____rune_private_1ed26dbc_compilerAppendImportExpr(__imports []__ParsedImport, __expr __ParsedExpr) []__ParsedImport {
	__path := ____rune_private_1ed26dbc_compilerAtImportPath(__expr)
	__goPath := __compilerGoPackageImportPath(__path)
	return func() []__ParsedImport {
		if __path == "" {
			return __imports
		}
		return ____rune_private_1ed26dbc_compilerAppendParsedImportIfMissing(__imports, __ParsedImport{__path: func() string {
			if __goPath != "" {
				return __goPath
			}
			return __path
		}(), __go: __goPath != "", __module: false, __line: __expr.__line, __column: __expr.__column})
	}()
}

func ____rune_private_1ed26dbc_compilerMergeParsedImports(__imports []__ParsedImport, __extra []__ParsedImport, __index int) []__ParsedImport {
	return func() []__ParsedImport {
		if __index >= len(__extra) {
			return __imports
		}
		return ____rune_private_1ed26dbc_compilerMergeParsedImports(____rune_private_1ed26dbc_compilerAppendParsedImportIfMissing(__imports, __extra[__index]), __extra, __index+1)
	}()
}

func ____rune_private_1ed26dbc_compilerAppendParsedImportIfMissing(__imports []__ParsedImport, __importDecl __ParsedImport) []__ParsedImport {
	return func() []__ParsedImport {
		switch {
		case ____rune_private_1ed26dbc_compilerParsedImportContains(__imports, __importDecl.__path, __importDecl.__go, __importDecl.__module, 0) == true:
			return __imports
		default:
			return func() []__ParsedImport {
				out := []__ParsedImport{}
				out = append(out, __imports...)
				out = append(out, __importDecl)
				return out
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerParsedImportContains(__imports []__ParsedImport, __path string, __go bool, __module bool, __index int) bool {
	return func() bool {
		if __index >= len(__imports) {
			return false
		}
		return func() bool {
			if __imports[__index].__path == __path && __imports[__index].__go == __go && __imports[__index].__module == __module {
				return true
			}
			return ____rune_private_1ed26dbc_compilerParsedImportContains(__imports, __path, __go, __module, __index+1)
		}()
	}()
}

func ____rune_private_1ed26dbc_compilerEmptyParsedImport() __ParsedImport {
	return __ParsedImport{__path: "", __go: false, __module: false, __line: 0, __column: 0}
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliases(__expr __ParsedExpr, __aliases []__CompilerNamespaceAlias) __ParsedExpr {
	return func() __ParsedExpr {
		switch {
		case __expr.__kind == __ExprKind_Block:
			return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasBlock(__expr, __aliases)
		case __expr.__kind == __ExprKind_Lambda:
			return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasLambda(__expr, __aliases)
		case __expr.__kind == __ExprKind_Selector:
			return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasSelector(__expr, __aliases)
		default:
			return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasChildren(__expr, __aliases)
		}
	}()
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasBlock(__expr __ParsedExpr, __aliases []__CompilerNamespaceAlias) __ParsedExpr {
	return ____rune_private_1ed26dbc_compilerWithChildren(__expr, ____rune_private_1ed26dbc_expandCompilerNamespaceAliasBlockChildren(__expr.__children, 0, __aliases, append([]__ParsedExpr{}, __expr.__children[0:0]...)))
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasBlockChildren(__statements []__ParsedExpr, __index int, __aliases []__CompilerNamespaceAlias, __out []__ParsedExpr) []__ParsedExpr {
	__done := __index >= len(__statements)
	return func() []__ParsedExpr {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasBlockStep(__statements, __index, __aliases, __out)
		}
	}()
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasBlockStep(__statements []__ParsedExpr, __index int, __aliases []__CompilerNamespaceAlias, __out []__ParsedExpr) []__ParsedExpr {
	__statement := __statements[__index]
	__alias := ____rune_private_1ed26dbc_compilerNamespaceAliasFromLet(__statement)
	__isAlias := __alias.__name != ""
	return func() []__ParsedExpr {
		switch {
		case __isAlias == true:
			return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasBlockChildren(__statements, __index+1, ____rune_private_1ed26dbc_addCompilerNamespaceAlias(__aliases, __alias), __out)
		default:
			return func() []__ParsedExpr {
				__expanded := ____rune_private_1ed26dbc_expandCompilerNamespaceAliases(__statement, __aliases)
				__nextAliases := ____rune_private_1ed26dbc_compilerNamespaceAliasesAfterBinding(__expanded, __aliases)
				return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasBlockChildren(__statements, __index+1, __nextAliases, func() []__ParsedExpr {
					out := []__ParsedExpr{}
					out = append(out, __out...)
					out = append(out, __expanded)
					return out
				}())
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerNamespaceAliasesAfterBinding(__statement __ParsedExpr, __aliases []__CompilerNamespaceAlias) []__CompilerNamespaceAlias {
	return func() []__CompilerNamespaceAlias {
		switch {
		case __statement.__kind == __ExprKind_Let:
			return ____rune_private_1ed26dbc_dropCompilerNamespaceAlias(__aliases, __statement.__name, 0, append([]__CompilerNamespaceAlias{}, []__CompilerNamespaceAlias{____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()}[0:0]...))
		case __statement.__kind == __ExprKind_ObjectDestructure:
			return ____rune_private_1ed26dbc_dropCompilerNamespaceAliasParams(__aliases, __statement.__params, 0)
		default:
			return __aliases
		}
	}()
}

func ____rune_private_1ed26dbc_dropCompilerNamespaceAliasParams(__aliases []__CompilerNamespaceAlias, __params []__ParsedParam, __index int) []__CompilerNamespaceAlias {
	return func() []__CompilerNamespaceAlias {
		if __index >= len(__params) {
			return __aliases
		}
		return ____rune_private_1ed26dbc_dropCompilerNamespaceAliasParams(____rune_private_1ed26dbc_dropCompilerNamespaceAlias(__aliases, __params[__index].__name, 0, append([]__CompilerNamespaceAlias{}, []__CompilerNamespaceAlias{____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()}[0:0]...)), __params, __index+1)
	}()
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasLambda(__expr __ParsedExpr, __aliases []__CompilerNamespaceAlias) __ParsedExpr {
	return ____rune_private_1ed26dbc_compilerWithChildren(__expr, ____rune_private_1ed26dbc_expandCompilerNamespaceAliasChildrenList(__expr.__children, 0, ____rune_private_1ed26dbc_dropCompilerNamespaceAliasParams(__aliases, __expr.__params, 0), append([]__ParsedExpr{}, __expr.__children[0:0]...)))
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasSelector(__expr __ParsedExpr, __aliases []__CompilerNamespaceAlias) __ParsedExpr {
	__noReceiver := len(__expr.__children) == 0
	return func() __ParsedExpr {
		switch {
		case __noReceiver == true:
			return __expr
		default:
			return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasSelectorReceiver(__expr, ____rune_private_1ed26dbc_expandCompilerNamespaceAliases(__expr.__children[0], __aliases), __aliases)
		}
	}()
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasSelectorReceiver(__expr __ParsedExpr, __receiver __ParsedExpr, __aliases []__CompilerNamespaceAlias) __ParsedExpr {
	__canAlias := __receiver.__kind == __ExprKind_Identifier
	__alias := func() __CompilerNamespaceAlias {
		if __canAlias {
			return ____rune_private_1ed26dbc_findCompilerNamespaceAlias(__aliases, __receiver.__name, 0)
		}
		return ____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()
	}()
	__found := __alias.__name != ""
	return func() __ParsedExpr {
		switch {
		case __found == true:
			return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasSelectorFound(__expr, __receiver, __alias)
		default:
			return ____rune_private_1ed26dbc_compilerWithChildren(__expr, []__ParsedExpr{__receiver})
		}
	}()
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasSelectorFound(__expr __ParsedExpr, __receiver __ParsedExpr, __alias __CompilerNamespaceAlias) __ParsedExpr {
	__moduleAlias := __alias.__module != ""
	return func() __ParsedExpr {
		switch {
		case __alias.__go == true:
			return ____rune_private_1ed26dbc_compilerWithChildren(__expr, []__ParsedExpr{____rune_private_1ed26dbc_compilerImportAtExpr("go:"+__alias.__importPath, __receiver.__line, __receiver.__column)})
		default:
			return func() __ParsedExpr {
				switch {
				case __moduleAlias == true:
					return ____rune_private_1ed26dbc_compilerWithChildren(__expr, []__ParsedExpr{____rune_private_1ed26dbc_compilerModuleAtExpr(__alias.__module, __receiver.__line, __receiver.__column)})
				default:
					return ____rune_private_1ed26dbc_compilerParsedExpr(__ExprKind_Identifier, __expr.__name, __expr.__name, "", "", []__ParsedParam{}, []__ParsedExpr{}, __expr.__line, __expr.__column)
				}
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasChildren(__expr __ParsedExpr, __aliases []__CompilerNamespaceAlias) __ParsedExpr {
	return ____rune_private_1ed26dbc_compilerWithChildren(__expr, ____rune_private_1ed26dbc_expandCompilerNamespaceAliasChildrenList(__expr.__children, 0, __aliases, append([]__ParsedExpr{}, __expr.__children[0:0]...)))
}

func ____rune_private_1ed26dbc_expandCompilerNamespaceAliasChildrenList(__children []__ParsedExpr, __index int, __aliases []__CompilerNamespaceAlias, __out []__ParsedExpr) []__ParsedExpr {
	return func() []__ParsedExpr {
		if __index >= len(__children) {
			return __out
		}
		return ____rune_private_1ed26dbc_expandCompilerNamespaceAliasChildrenList(__children, __index+1, __aliases, func() []__ParsedExpr {
			out := []__ParsedExpr{}
			out = append(out, __out...)
			out = append(out, ____rune_private_1ed26dbc_expandCompilerNamespaceAliases(__children[__index], __aliases))
			return out
		}())
	}()
}

func ____rune_private_1ed26dbc_compilerNamespaceAliasFromLet(__expr __ParsedExpr) __CompilerNamespaceAlias {
	__valid := __expr.__kind == __ExprKind_Let && len(__expr.__children) > 0
	return func() __CompilerNamespaceAlias {
		switch {
		case __valid == true:
			return ____rune_private_1ed26dbc_compilerNamespaceAliasFromLetValue(__expr.__name, __expr.__children[0])
		default:
			return ____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerNamespaceAliasFromLetValue(__name string, __value __ParsedExpr) __CompilerNamespaceAlias {
	__valid := __value.__kind == __ExprKind_At
	return func() __CompilerNamespaceAlias {
		switch {
		case __valid == true:
			return ____rune_private_1ed26dbc_compilerNamespaceAliasFromAt(__name, __value)
		default:
			return ____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerNamespaceAliasFromAt(__name string, __expr __ParsedExpr) __CompilerNamespaceAlias {
	__importPath := ____rune_private_1ed26dbc_compilerAtImportPath(__expr)
	__goPath := __compilerGoPackageImportPath(__importPath)
	return func() __CompilerNamespaceAlias {
		if __importPath != "" {
			return func() __CompilerNamespaceAlias {
				if __goPath != "" {
					return ____rune_private_1ed26dbc_compilerGoNamespaceAlias(__name, __goPath)
				}
				return ____rune_private_1ed26dbc_compilerNamespaceAlias(__name, "")
			}()
		}
		return func() __CompilerNamespaceAlias {
			if __expr.__name != "" {
				return ____rune_private_1ed26dbc_compilerNamespaceAlias(__name, __expr.__name)
			}
			return ____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()
		}()
	}()
}

func ____rune_private_1ed26dbc_compilerAtImportPath(__expr __ParsedExpr) string {
	return func() string {
		if __expr.__kind == __ExprKind_At && __expr.__value != "" {
			return ____rune_private_1ed26dbc_compilerUnquoteString(__expr.__value)
		}
		return ""
	}()
}

func ____rune_private_1ed26dbc_compilerImportAtExpr(__path string, __line int, __column int) __ParsedExpr {
	return ____rune_private_1ed26dbc_compilerParsedExpr(__ExprKind_At, "@", "", "\""+__path+"\"", "", []__ParsedParam{}, []__ParsedExpr{}, __line, __column)
}

func ____rune_private_1ed26dbc_compilerModuleAtExpr(__module string, __line int, __column int) __ParsedExpr {
	return ____rune_private_1ed26dbc_compilerParsedExpr(__ExprKind_At, "@", __module, "", "", []__ParsedParam{}, []__ParsedExpr{}, __line, __column)
}

func ____rune_private_1ed26dbc_compilerWithChildren(__expr __ParsedExpr, __children []__ParsedExpr) __ParsedExpr {
	return ____rune_private_1ed26dbc_compilerParsedExpr(__expr.__kind, __expr.__text, __expr.__name, __expr.__value, __expr.__op, __expr.__params, __children, __expr.__line, __expr.__column)
}

func ____rune_private_1ed26dbc_compilerNamespaceAlias(__name string, __module string) __CompilerNamespaceAlias {
	return __CompilerNamespaceAlias{__name: __name, __module: __module, __importPath: "", __go: false}
}

func ____rune_private_1ed26dbc_compilerGoNamespaceAlias(__name string, __importPath string) __CompilerNamespaceAlias {
	return __CompilerNamespaceAlias{__name: __name, __module: "", __importPath: __importPath, __go: true}
}

func ____rune_private_1ed26dbc_emptyCompilerNamespaceAlias() __CompilerNamespaceAlias {
	return ____rune_private_1ed26dbc_compilerNamespaceAlias("", "")
}

func ____rune_private_1ed26dbc_addCompilerNamespaceAlias(__aliases []__CompilerNamespaceAlias, __alias __CompilerNamespaceAlias) []__CompilerNamespaceAlias {
	return func() []__CompilerNamespaceAlias {
		out := []__CompilerNamespaceAlias{}
		out = append(out, ____rune_private_1ed26dbc_dropCompilerNamespaceAlias(__aliases, __alias.__name, 0, append([]__CompilerNamespaceAlias{}, []__CompilerNamespaceAlias{____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()}[0:0]...))...)
		out = append(out, __alias)
		return out
	}()
}

func ____rune_private_1ed26dbc_dropCompilerNamespaceAlias(__aliases []__CompilerNamespaceAlias, __name string, __index int, __out []__CompilerNamespaceAlias) []__CompilerNamespaceAlias {
	return func() []__CompilerNamespaceAlias {
		if __index >= len(__aliases) {
			return __out
		}
		return func() []__CompilerNamespaceAlias {
			if __aliases[__index].__name == __name {
				return ____rune_private_1ed26dbc_dropCompilerNamespaceAlias(__aliases, __name, __index+1, __out)
			}
			return ____rune_private_1ed26dbc_dropCompilerNamespaceAlias(__aliases, __name, __index+1, func() []__CompilerNamespaceAlias {
				out := []__CompilerNamespaceAlias{}
				out = append(out, __out...)
				out = append(out, __aliases[__index])
				return out
			}())
		}()
	}()
}

func ____rune_private_1ed26dbc_findCompilerNamespaceAlias(__aliases []__CompilerNamespaceAlias, __name string, __index int) __CompilerNamespaceAlias {
	return func() __CompilerNamespaceAlias {
		if __index >= len(__aliases) {
			return ____rune_private_1ed26dbc_emptyCompilerNamespaceAlias()
		}
		return func() __CompilerNamespaceAlias {
			if __aliases[__index].__name == __name {
				return __aliases[__index]
			}
			return ____rune_private_1ed26dbc_findCompilerNamespaceAlias(__aliases, __name, __index+1)
		}()
	}()
}

func ____rune_private_1ed26dbc_compilerRenameDeclarationName(__annotations []__ParsedAnnotation, __fallback string) string {
	return ____rune_private_1ed26dbc_compilerRenameDeclarationNameAt(__annotations, __fallback, 0)
}

func ____rune_private_1ed26dbc_compilerRenameDeclarationNameAt(__annotations []__ParsedAnnotation, __fallback string, __index int) string {
	__done := __index >= len(__annotations)
	return func() string {
		switch {
		case __done == true:
			return __fallback
		default:
			return ____rune_private_1ed26dbc_compilerRenameDeclarationNameStep(__annotations, __fallback, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerRenameDeclarationNameStep(__annotations []__ParsedAnnotation, __fallback string, __index int) string {
	__annotation := __annotations[__index]
	__matched := __annotation.__marker == "#" && __annotation.__module == "macro" && __annotation.__name == "renameDeclaration" && len(__annotation.__args) > 0
	return func() string {
		switch {
		case __matched == true:
			return ____rune_private_1ed26dbc_compilerAnnotationStringArg(__annotation, 0, __fallback)
		default:
			return ____rune_private_1ed26dbc_compilerRenameDeclarationNameAt(__annotations, __fallback, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_compilerAnnotationStringArg(__annotation __ParsedAnnotation, __index int, __fallback string) string {
	__valid := __index < len(__annotation.__args) && __annotation.__args[__index].__kind == __ExprKind_String
	return func() string {
		switch {
		case __valid == true:
			return ____rune_private_1ed26dbc_compilerUnquoteString(__annotation.__args[__index].__value)
		default:
			return __fallback
		}
	}()
}

func ____rune_private_1ed26dbc_compilerUnquoteString(__raw string) string {
	return func() string {
		if len([]rune(__raw)) >= 2 {
			return func() string { runes := []rune(__raw); return string(runes[1 : len([]rune(__raw))-1]) }()
		}
		return __raw
	}()
}

func ____rune_private_1ed26dbc_lowerReachableRuneFiles(__files []__SourceFile, __pending []string, __seen []string, __out __IRFile) __IRFile {
	__empty := len(__pending) == 0
	return func() __IRFile {
		switch {
		case __empty == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_lowerReachableRuneFile(__files, __pending[0], append([]string{}, __pending[1:len(__pending)]...), __seen, __out)
		}
	}()
}

func ____rune_private_1ed26dbc_lowerReachableRuneFile(__files []__SourceFile, __path string, __rest []string, __seen []string, __out __IRFile) __IRFile {
	__alreadySeen := ____rune_private_1ed26dbc_compilerContains(__seen, __path)
	return func() __IRFile {
		switch {
		case __alreadySeen == true:
			return ____rune_private_1ed26dbc_lowerReachableRuneFiles(__files, __rest, __seen, __out)
		default:
			return ____rune_private_1ed26dbc_lowerUnseenRuneFile(__files, __path, __rest, func() []string {
				out := []string{}
				out = append(out, __seen...)
				out = append(out, __path)
				return out
			}(), __out)
		}
	}()
}

func ____rune_private_1ed26dbc_lowerUnseenRuneFile(__files []__SourceFile, __path string, __rest []string, __seen []string, __out __IRFile) __IRFile {
	__file := ____rune_private_1ed26dbc_findSourceFile(__files, __path, 0)
	__missing := len(__file.__path) == 0
	return func() __IRFile {
		switch {
		case __missing == true:
			return ____rune_private_1ed26dbc_lowerReachableRuneFiles(__files, __rest, __seen, ____rune_private_1ed26dbc_mergeCompilerIRFile(__out, ____rune_private_1ed26dbc_missingImportFile(__path)))
		default:
			return ____rune_private_1ed26dbc_lowerFoundSourceFile(__files, __file, __rest, __seen, __out)
		}
	}()
}

func ____rune_private_1ed26dbc_lowerFoundSourceFile(__files []__SourceFile, __file __SourceFile, __rest []string, __seen []string, __out __IRFile) __IRFile {
	__typeScript := ____rune_private_1ed26dbc_sourceFileIsTypeScript(__file)
	return func() __IRFile {
		switch {
		case __typeScript == true:
			return ____rune_private_1ed26dbc_lowerReachableRuneFiles(__files, __rest, __seen, __out)
		default:
			return ____rune_private_1ed26dbc_lowerFoundRuneSourceFile(__files, __file, __rest, __seen, __out)
		}
	}()
}

func ____rune_private_1ed26dbc_lowerFoundRuneSourceFile(__files []__SourceFile, __file __SourceFile, __rest []string, __seen []string, __out __IRFile) __IRFile {
	__lowered := ____rune_private_1ed26dbc_lowerCompilerSourceWithPath(__file.__source, __file.__path)
	__merged := ____rune_private_1ed26dbc_mergeCompilerIRFile(__out, __lowered)
	__withTypeScript := ____rune_private_1ed26dbc_lowerTypeScriptImportsForRuneFile(__files, __file.__path, __lowered.__imports, 0, __merged)
	return ____rune_private_1ed26dbc_lowerReachableRuneFiles(__files, ____rune_private_1ed26dbc_appendImportPaths(__rest, __file.__path, __lowered.__imports, 0), __seen, __withTypeScript)
}

func ____rune_private_1ed26dbc_lowerTypeScriptImportsForRuneFile(__files []__SourceFile, __basePath string, __imports []__IRImport, __index int, __out __IRFile) __IRFile {
	__done := __index >= len(__imports)
	return func() __IRFile {
		switch {
		case __done == true:
			return __out
		default:
			return ____rune_private_1ed26dbc_lowerTypeScriptImportForRuneFile(__files, __basePath, __imports, __index, __out)
		}
	}()
}

func ____rune_private_1ed26dbc_lowerTypeScriptImportForRuneFile(__files []__SourceFile, __basePath string, __imports []__IRImport, __index int, __out __IRFile) __IRFile {
	__importDecl := __imports[__index]
	__shouldLoad := __importDecl.__go == false && __importDecl.__module == false && strings.HasSuffix(__importDecl.__path, ".ts")
	return func() __IRFile {
		switch {
		case __shouldLoad == true:
			return ____rune_private_1ed26dbc_lowerTypeScriptImportPathForRuneFile(__files, __basePath, __imports, __index, __out)
		default:
			return ____rune_private_1ed26dbc_lowerTypeScriptImportsForRuneFile(__files, __basePath, __imports, __index+1, __out)
		}
	}()
}

func ____rune_private_1ed26dbc_lowerTypeScriptImportPathForRuneFile(__files []__SourceFile, __basePath string, __imports []__IRImport, __index int, __out __IRFile) __IRFile {
	__importDecl := __imports[__index]
	__resolved := ____rune_private_1ed26dbc_resolveCompilerImportPath(__basePath, __importDecl.__path)
	__file := ____rune_private_1ed26dbc_findSourceFile(__files, __resolved, 0)
	__missing := len(__file.__path) == 0
	__next := func() __IRFile {
		switch {
		case __missing == true:
			return ____rune_private_1ed26dbc_mergeCompilerIRFile(__out, ____rune_private_1ed26dbc_missingImportFile(__resolved))
		default:
			return ____rune_private_1ed26dbc_mergeCompilerIRFile(__out, ____rune_private_1ed26dbc_lowerTypeScriptSourceFile(__file, __importDecl.__path))
		}
	}()
	return ____rune_private_1ed26dbc_lowerTypeScriptImportsForRuneFile(__files, __basePath, __imports, __index+1, __next)
}

func ____rune_private_1ed26dbc_appendImportPaths(__pending []string, __basePath string, __imports []__IRImport, __index int) []string {
	__done := __index >= len(__imports)
	return func() []string {
		switch {
		case __done == true:
			return __pending
		default:
			return ____rune_private_1ed26dbc_appendImportPath(__pending, __basePath, __imports, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_appendImportPath(__pending []string, __basePath string, __imports []__IRImport, __index int) []string {
	__skip := __imports[__index].__go || __imports[__index].__module
	return func() []string {
		switch {
		case __skip == true:
			return ____rune_private_1ed26dbc_appendImportPaths(__pending, __basePath, __imports, __index+1)
		default:
			return ____rune_private_1ed26dbc_appendImportPathIfRune(__pending, __basePath, __imports, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_appendImportPathIfRune(__pending []string, __basePath string, __imports []__IRImport, __index int) []string {
	__typeScript := __imports[__index].__module || strings.HasSuffix(__imports[__index].__path, ".ts")
	return func() []string {
		switch {
		case __typeScript == true:
			return ____rune_private_1ed26dbc_appendImportPaths(__pending, __basePath, __imports, __index+1)
		default:
			return ____rune_private_1ed26dbc_appendImportPaths(func() []string {
				out := []string{}
				out = append(out, __pending...)
				out = append(out, ____rune_private_1ed26dbc_resolveCompilerImportPath(__basePath, __imports[__index].__path))
				return out
			}(), __basePath, __imports, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_findSourceFile(__files []__SourceFile, __path string, __index int) __SourceFile {
	__exact := ____rune_private_1ed26dbc_findSourceFileExact(__files, ____rune_private_1ed26dbc_compilerPathNormalize(__path), __index)
	__found := len(__exact.__path) == 0 == false
	return func() __SourceFile {
		switch {
		case __found == true:
			return __exact
		default:
			return ____rune_private_1ed26dbc_findSourceFileBasename(__files, __path, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_findSourceFileExact(__files []__SourceFile, __path string, __index int) __SourceFile {
	__done := __index >= len(__files)
	return func() __SourceFile {
		switch {
		case __done == true:
			return ____rune_private_1ed26dbc_emptySourceFile()
		default:
			return ____rune_private_1ed26dbc_findSourceFileExactAt(__files, __path, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_findSourceFileExactAt(__files []__SourceFile, __path string, __index int) __SourceFile {
	__matched := ____rune_private_1ed26dbc_compilerPathNormalize(__files[__index].__path) == __path
	return func() __SourceFile {
		switch {
		case __matched == true:
			return __files[__index]
		default:
			return ____rune_private_1ed26dbc_findSourceFileExact(__files, __path, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_findSourceFileBasename(__files []__SourceFile, __path string, __index int) __SourceFile {
	__done := __index >= len(__files)
	return func() __SourceFile {
		switch {
		case __done == true:
			return ____rune_private_1ed26dbc_emptySourceFile()
		default:
			return ____rune_private_1ed26dbc_findSourceFileBasenameAt(__files, __path, __index)
		}
	}()
}

func ____rune_private_1ed26dbc_findSourceFileBasenameAt(__files []__SourceFile, __path string, __index int) __SourceFile {
	__matched := ____rune_private_1ed26dbc_compilerPathBasename(__files[__index].__path) == ____rune_private_1ed26dbc_compilerPathBasename(__path)
	return func() __SourceFile {
		switch {
		case __matched == true:
			return __files[__index]
		default:
			return ____rune_private_1ed26dbc_findSourceFileBasename(__files, __path, __index+1)
		}
	}()
}

func ____rune_private_1ed26dbc_emptySourceFile() __SourceFile {
	return __SourceFile{__path: "", __source: ""}
}

func ____rune_private_1ed26dbc_missingImportFile(__path string) __IRFile {
	__out := ____rune_private_1ed26dbc_emptyCompilerIRFile()
	__out.__errors = append(__out.__errors, __ParseError{__message: "missing import " + __path, __line: 0, __column: 0})
	return __out
}

func ____rune_private_1ed26dbc_sourceFileIsTypeScript(__file __SourceFile) bool {
	return strings.HasSuffix(__file.__path, ".ts")
}

func ____rune_private_1ed26dbc_resolveCompilerImportPath(__basePath string, __importPath string) string {
	__absolute := strings.HasPrefix(__importPath, "/")
	return func() string {
		switch {
		case __absolute == true:
			return ____rune_private_1ed26dbc_compilerPathNormalize(__importPath)
		default:
			return ____rune_private_1ed26dbc_compilerPathNormalize(____rune_private_1ed26dbc_compilerPathJoin(____rune_private_1ed26dbc_compilerPathDirname(__basePath), __importPath))
		}
	}()
}

func ____rune_private_1ed26dbc_compilerPathDirname(__path string) string {
	__slash := strings.LastIndex(__path, "/")
	return func() string {
		if __slash < 0 {
			return ""
		}
		return func() string {
			if __slash == 0 {
				return "/"
			}
			return func() string { runes := []rune(__path); return string(runes[0:__slash]) }()
		}()
	}()
}

func ____rune_private_1ed26dbc_compilerPathJoin(__base string, __child string) string {
	return func() string {
		if len(__base) == 0 {
			return __child
		}
		return func() string {
			if __base == "/" {
				return "/" + __child
			}
			return __base + "/" + __child
		}()
	}()
}

func ____rune_private_1ed26dbc_compilerPathNormalize(__path string) string {
	__absolute := strings.HasPrefix(__path, "/")
	__parts := func() []string { parts := strings.Split(__path, "/"); return parts }()
	__normalized := ____rune_private_1ed26dbc_compilerPathNormalizeParts(__parts, 0, __absolute, append([]string{}, __parts[0:0]...))
	__joined := ____rune_private_1ed26dbc_compilerPathJoinParts(__normalized, 0, "")
	return func() string {
		switch {
		case __absolute == true:
			return "/" + __joined
		default:
			return func() string {
				if len(__joined) == 0 {
					return "."
				}
				return __joined
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerPathNormalizeParts(__parts []string, __index int, __absolute bool, __out []string) []string {
	return func() []string {
		if __index >= len(__parts) {
			return __out
		}
		return ____rune_private_1ed26dbc_compilerPathNormalizePart(__parts, __index, __absolute, __out)
	}()
}

func ____rune_private_1ed26dbc_compilerPathNormalizePart(__parts []string, __index int, __absolute bool, __out []string) []string {
	__part := __parts[__index]
	return func() []string {
		if len(__part) == 0 || __part == "." {
			return ____rune_private_1ed26dbc_compilerPathNormalizeParts(__parts, __index+1, __absolute, __out)
		}
		return func() []string {
			if __part == ".." {
				return ____rune_private_1ed26dbc_compilerPathNormalizeParent(__parts, __index, __absolute, __out)
			}
			return ____rune_private_1ed26dbc_compilerPathNormalizeParts(__parts, __index+1, __absolute, func() []string { out := []string{}; out = append(out, __out...); out = append(out, __part); return out }())
		}()
	}()
}

func ____rune_private_1ed26dbc_compilerPathNormalizeParent(__parts []string, __index int, __absolute bool, __out []string) []string {
	__canPop := len(__out) > 0 && __out[len(__out)-1] != ".."
	return func() []string {
		switch {
		case __canPop == true:
			return ____rune_private_1ed26dbc_compilerPathNormalizeParts(__parts, __index+1, __absolute, append([]string{}, __out[0:len(__out)-1]...))
		default:
			return func() []string {
				if __absolute {
					return ____rune_private_1ed26dbc_compilerPathNormalizeParts(__parts, __index+1, __absolute, __out)
				}
				return ____rune_private_1ed26dbc_compilerPathNormalizeParts(__parts, __index+1, __absolute, func() []string { out := []string{}; out = append(out, __out...); out = append(out, ".."); return out }())
			}()
		}
	}()
}

func ____rune_private_1ed26dbc_compilerPathJoinParts(__parts []string, __index int, __out string) string {
	return func() string {
		if __index >= len(__parts) {
			return __out
		}
		return ____rune_private_1ed26dbc_compilerPathJoinParts(__parts, __index+1, ____rune_private_1ed26dbc_compilerPathJoin(__out, __parts[__index]))
	}()
}

func ____rune_private_1ed26dbc_lowerTypeScriptSourceFile(__file __SourceFile, __specifier string) __IRFile {
	__out := ____rune_private_1ed26dbc_emptyCompilerIRFile()
	if __specifier != "" {
		func() int {
			__out.__tsImports = append(__out.__tsImports, ____rune_private_1ed26dbc_parseTypeScriptImport(__file.__path, __specifier, __file.__source))
			return len(__out.__tsImports)
		}()
	}
	return __out
}

func ____rune_private_1ed26dbc_typeScriptImportSpecifier(__path string, __imports []__IRImport, __index int) string {
	return func() string {
		if __index >= len(__imports) {
			return ""
		}
		return func() string {
			if strings.HasSuffix(__imports[__index].__path, ".ts") && ____rune_private_1ed26dbc_compilerPathBasename(__path) == ____rune_private_1ed26dbc_compilerPathBasename(__imports[__index].__path) {
				return __imports[__index].__path
			}
			return ____rune_private_1ed26dbc_typeScriptImportSpecifier(__path, __imports, __index+1)
		}()
	}()
}

func ____rune_private_1ed26dbc_compilerPathBasename(__path string) string {
	__slash := strings.LastIndex(__path, "/")
	return func() string {
		if __slash < 0 {
			return __path
		}
		return func() string { runes := []rune(__path); return string(runes[__slash+1 : len([]rune(__path))]) }()
	}()
}

func ____rune_private_1ed26dbc_parseTypeScriptImport(__path string, __specifier string, __source string) __IRTSImport {
	__imports := __IRTSImport{__path: __path, __specifier: __specifier, __functions: append([]__IRFunction{}, []__IRFunction{__emptyIRFunction()}[0:0]...), __values: append([]__IRConst{}, []__IRConst{____rune_private_1ed26dbc_emptyIRConst()}[0:0]...), __line: 0, __column: 0}
	for _, __line := range func() []string { parts := strings.Split(__source, "\n"); return parts }() {
		_ = __line
		__imports = ____rune_private_1ed26dbc_parseTypeScriptExportLine(__imports, strings.TrimSpace(__line))
	}
	return __imports
}

func ____rune_private_1ed26dbc_parseTypeScriptExportLine(__imports __IRTSImport, __line string) __IRTSImport {
	return func() __IRTSImport {
		if strings.HasPrefix(__line, "export async function ") {
			return ____rune_private_1ed26dbc_pushTypeScriptFunction(__imports, func() string {
				runes := []rune(__line)
				return string(runes[len([]rune("export async function ")):len([]rune(__line))])
			}(), true)
		}
		return func() __IRTSImport {
			if strings.HasPrefix(__line, "export function ") {
				return ____rune_private_1ed26dbc_pushTypeScriptFunction(__imports, func() string {
					runes := []rune(__line)
					return string(runes[len([]rune("export function ")):len([]rune(__line))])
				}(), false)
			}
			return func() __IRTSImport {
				if strings.HasPrefix(__line, "export const ") {
					return ____rune_private_1ed26dbc_pushTypeScriptValue(__imports, func() string {
						runes := []rune(__line)
						return string(runes[len([]rune("export const ")):len([]rune(__line))])
					}())
				}
				return func() __IRTSImport {
					if strings.HasPrefix(__line, "export let ") {
						return ____rune_private_1ed26dbc_pushTypeScriptValue(__imports, func() string {
							runes := []rune(__line)
							return string(runes[len([]rune("export let ")):len([]rune(__line))])
						}())
					}
					return func() __IRTSImport {
						if strings.HasPrefix(__line, "export var ") {
							return ____rune_private_1ed26dbc_pushTypeScriptValue(__imports, func() string {
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

func ____rune_private_1ed26dbc_pushTypeScriptFunction(__imports __IRTSImport, __text string, __routine bool) __IRTSImport {
	__open := strings.Index(__text, "(")
	__close := strings.Index(__text, ")")
	__name := func() string {
		if __open < 0 {
			return ""
		}
		return strings.TrimSpace((func() string { runes := []rune(__text); return string(runes[0:__open]) }()))
	}()
	__returnType := ____rune_private_1ed26dbc_typeScriptReturnTypeName(__text)
	if __name != "" {
		func() int {
			__imports.__functions = append(__imports.__functions, __IRFunction{__name: __name, __private: false, __static: false, __routine: __routine, __macro: false, __receiverType: "", __generics: []string{}, __params: func() []__IRParam {
				if __open >= 0 && __close > __open {
					return ____rune_private_1ed26dbc_parseTypeScriptParams(func() string { runes := []rune(__text); return string(runes[__open+1 : __close]) }())
				}
				return append([]__IRParam{}, []__IRParam{____rune_private_1ed26dbc_emptyIRParam()}[0:0]...)
			}(), __returnType: __returnType, __body: __emptyIRExpr(), __sourcePath: "", __line: 0, __column: 0})
			return len(__imports.__functions)
		}()
	}
	return __imports
}

func ____rune_private_1ed26dbc_pushTypeScriptValue(__imports __IRTSImport, __text string) __IRTSImport {
	__end := ____rune_private_1ed26dbc_typeScriptNameEnd(__text)
	__name := strings.TrimSpace((func() string { runes := []rune(__text); return string(runes[0:__end]) }()))
	__typeName := ____rune_private_1ed26dbc_typeScriptValueTypeName(__text)
	if __name != "" {
		func() int {
			__imports.__values = append(__imports.__values, __IRConst{__name: __name, __private: false, __typeName: __typeName, __value: __emptyIRExpr(), __line: 0, __column: 0})
			return len(__imports.__values)
		}()
	}
	return __imports
}

func ____rune_private_1ed26dbc_typeScriptNameEnd(__text string) int {
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

func ____rune_private_1ed26dbc_typeScriptReturnTypeName(__text string) string {
	__close := strings.Index(__text, ")")
	__colon := func() int {
		if __close >= 0 {
			return strings.Index((func() string { runes := []rune(__text); return string(runes[__close+1 : len([]rune(__text))]) }()), ":")
		}
		return -1
	}()
	return func() string {
		if __close >= 0 && __colon >= 0 {
			return ____rune_private_1ed26dbc_typeScriptTextType(strings.TrimSpace((func() string {
				runes := []rune(__text)
				return string(runes[__close+1+__colon+1 : ____rune_private_1ed26dbc_typeScriptReturnTypeEnd(__text)])
			}())))
		}
		return "Dynamic"
	}()
}

func ____rune_private_1ed26dbc_typeScriptReturnTypeEnd(__text string) int {
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

func ____rune_private_1ed26dbc_typeScriptValueTypeName(__text string) string {
	__colon := strings.Index(__text, ":")
	return func() string {
		if __colon < 0 {
			return "Dynamic"
		}
		return ____rune_private_1ed26dbc_typeScriptTextType(strings.TrimSpace((func() string {
			runes := []rune(__text)
			return string(runes[__colon+1 : ____rune_private_1ed26dbc_typeScriptNameEnd(func() string { runes := []rune(__text); return string(runes[__colon+1 : len([]rune(__text))]) }())+__colon+1])
		}())))
	}()
}

func ____rune_private_1ed26dbc_parseTypeScriptParams(__text string) []__IRParam {
	__params := append([]__IRParam{}, []__IRParam{____rune_private_1ed26dbc_emptyIRParam()}[0:0]...)
	for _, __param := range func() []string { parts := strings.Split(__text, ","); return parts }() {
		_ = __param
		func() {
			if strings.TrimSpace(__param) != "" {
				func() int {
					__params = append(__params, ____rune_private_1ed26dbc_parseTypeScriptParam(strings.TrimSpace(__param)))
					return len(__params)
				}()
				return
			}
		}()
	}
	return __params
}

func ____rune_private_1ed26dbc_emptyIRParam() __IRParam {
	return __IRParam{__name: "", __typeName: "Dynamic", __line: 0, __column: 0}
}

func ____rune_private_1ed26dbc_emptyIRConst() __IRConst {
	return __IRConst{__name: "", __private: false, __typeName: "Dynamic", __value: __emptyIRExpr(), __line: 0, __column: 0}
}

func ____rune_private_1ed26dbc_parseTypeScriptParam(__text string) __IRParam {
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
		return ____rune_private_1ed26dbc_typeScriptTextType(strings.TrimSpace((func() string { runes := []rune(__text); return string(runes[__colon+1 : len([]rune(__text))]) }())))
	}(), __line: 0, __column: 0}
}

func ____rune_private_1ed26dbc_typeScriptTextType(__text string) string {
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

func ____rune_private_1ed26dbc_emptyCompilerIRFile() __IRFile {
	return __IRFile{__imports: []__IRImport{}, __tsImports: []__IRTSImport{}, __structs: []__IRStructType{}, __enums: []__IREnumType{}, __constants: []__IRConst{}, __functions: []__IRFunction{}, __tests: []__IRTest{}, __errors: []__ParseError{}}
}

func ____rune_private_1ed26dbc_mergeCompilerIRFile(__out __IRFile, __file __IRFile) __IRFile {
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

func ____rune_private_1ed26dbc_compileResult(__ok bool, __output string, __errors []string) __CompileResult {
	return __CompileResult{__ok: __ok, __output: __output, __errors: __errors}
}

func ____rune_private_1ed26dbc_unsupportedTargetErrors(__target string) []string {
	__errors := []string{}
	__errors = append(__errors, "unsupported target "+__target)
	return __errors
}

func ____rune_private_1ed26dbc_parseErrorMessages(__errors []__ParseError) []string {
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
