package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
)

type TokenKind int

const (
	TokenKind_EOF                TokenKind = 0
	TokenKind_Illegal            TokenKind = 1
	TokenKind_Newline            TokenKind = 2
	TokenKind_Ident              TokenKind = 3
	TokenKind_Int                TokenKind = 4
	TokenKind_Double             TokenKind = 5
	TokenKind_BigInt             TokenKind = 6
	TokenKind_String             TokenKind = 7
	TokenKind_TemplateString     TokenKind = 8
	TokenKind_Char               TokenKind = 9
	TokenKind_Regex              TokenKind = 10
	TokenKind_XMLText            TokenKind = 11
	TokenKind_Hash               TokenKind = 12
	TokenKind_At                 TokenKind = 13
	TokenKind_Dollar             TokenKind = 14
	TokenKind_Dot                TokenKind = 15
	TokenKind_DotDot             TokenKind = 16
	TokenKind_DotDotLess         TokenKind = 17
	TokenKind_DotDotEqual        TokenKind = 18
	TokenKind_DotDotDot          TokenKind = 19
	TokenKind_Comma              TokenKind = 20
	TokenKind_Colon              TokenKind = 21
	TokenKind_DoubleColon        TokenKind = 22
	TokenKind_LParen             TokenKind = 23
	TokenKind_RParen             TokenKind = 24
	TokenKind_LBracket           TokenKind = 25
	TokenKind_RBracket           TokenKind = 26
	TokenKind_LBrace             TokenKind = 27
	TokenKind_RBrace             TokenKind = 28
	TokenKind_Question           TokenKind = 29
	TokenKind_QuestionQuestion   TokenKind = 30
	TokenKind_Apostrophe         TokenKind = 31
	TokenKind_FatArrow           TokenKind = 32
	TokenKind_Assign             TokenKind = 33
	TokenKind_Declare            TokenKind = 34
	TokenKind_MutDeclare         TokenKind = 35
	TokenKind_Arrow              TokenKind = 36
	TokenKind_Plus               TokenKind = 37
	TokenKind_PlusPlus           TokenKind = 38
	TokenKind_Minus              TokenKind = 39
	TokenKind_Star               TokenKind = 40
	TokenKind_Slash              TokenKind = 41
	TokenKind_Percent            TokenKind = 42
	TokenKind_Bang               TokenKind = 43
	TokenKind_Tilde              TokenKind = 44
	TokenKind_BitAnd             TokenKind = 45
	TokenKind_BitOr              TokenKind = 46
	TokenKind_BitXor             TokenKind = 47
	TokenKind_ShiftLeft          TokenKind = 48
	TokenKind_ShiftRight         TokenKind = 49
	TokenKind_UnsignedShiftRight TokenKind = 50
	TokenKind_AndAnd             TokenKind = 51
	TokenKind_OrOr               TokenKind = 52
	TokenKind_EqualEqual         TokenKind = 53
	TokenKind_BangEqual          TokenKind = 54
	TokenKind_Less               TokenKind = 55
	TokenKind_LessEqual          TokenKind = 56
	TokenKind_Greater            TokenKind = 57
	TokenKind_GreaterEqual       TokenKind = 58
	TokenKind_Underscore         TokenKind = 59
)

type TypeRefKind int

const (
	TypeRefKind_Unknown  TypeRefKind = 0
	TypeRefKind_Name     TypeRefKind = 1
	TypeRefKind_Group    TypeRefKind = 2
	TypeRefKind_Tuple    TypeRefKind = 3
	TypeRefKind_Function TypeRefKind = 4
)

type ExprKind int

const (
	ExprKind_Unknown           ExprKind = 0
	ExprKind_Identifier        ExprKind = 1
	ExprKind_At                ExprKind = 2
	ExprKind_This              ExprKind = 3
	ExprKind_Int               ExprKind = 4
	ExprKind_Double            ExprKind = 5
	ExprKind_BigInt            ExprKind = 6
	ExprKind_String            ExprKind = 7
	ExprKind_Template          ExprKind = 8
	ExprKind_Char              ExprKind = 9
	ExprKind_Regex             ExprKind = 10
	ExprKind_XMLText           ExprKind = 11
	ExprKind_Bool              ExprKind = 12
	ExprKind_Null              ExprKind = 13
	ExprKind_Unary             ExprKind = 14
	ExprKind_Postfix           ExprKind = 15
	ExprKind_Unwrap            ExprKind = 16
	ExprKind_CompileTime       ExprKind = 17
	ExprKind_Binary            ExprKind = 18
	ExprKind_Ternary           ExprKind = 19
	ExprKind_Assign            ExprKind = 20
	ExprKind_Call              ExprKind = 21
	ExprKind_Args              ExprKind = 22
	ExprKind_Lambda            ExprKind = 23
	ExprKind_Selector          ExprKind = 24
	ExprKind_Index             ExprKind = 25
	ExprKind_Array             ExprKind = 26
	ExprKind_Tuple             ExprKind = 27
	ExprKind_Map               ExprKind = 28
	ExprKind_Entry             ExprKind = 29
	ExprKind_Spread            ExprKind = 30
	ExprKind_Reactive          ExprKind = 31
	ExprKind_Struct            ExprKind = 32
	ExprKind_Object            ExprKind = 33
	ExprKind_Field             ExprKind = 34
	ExprKind_PrivateField      ExprKind = 35
	ExprKind_Method            ExprKind = 36
	ExprKind_PrivateMethod     ExprKind = 37
	ExprKind_Block             ExprKind = 38
	ExprKind_PatternBlock      ExprKind = 39
	ExprKind_Match             ExprKind = 40
	ExprKind_Branch            ExprKind = 41
	ExprKind_Pattern           ExprKind = 42
	ExprKind_Let               ExprKind = 43
	ExprKind_ObjectDestructure ExprKind = 44
	ExprKind_Error             ExprKind = 45
	ExprKind_Watch             ExprKind = 46
	ExprKind_XMLElement        ExprKind = 47
)

type Token struct {
	kind   TokenKind
	lexeme string
	offset int
	line   int
	column int
}

type LexState struct {
	source        string
	start         int
	current       int
	line          int
	column        int
	startLine     int
	startColumn   int
	canStartRegex bool
	mode          int
	xmlDepth      int
	xmlClosing    bool
	xmlSelfClosed bool
	xmlExprMode   int
	xmlExprDepth  int
	xmlAfterMode  int
	xmlAfterDepth int
}

type Advanced struct {
	state LexState
	ch    rune
}

type Lexed struct {
	state LexState
	kind  TokenKind
}

type ScannedString struct {
	state LexState
	ok    bool
}

type ParsedTypeParam struct {
	name     string
	optional bool
	typeRef  ParsedTypeRef
}

type ParsedTypeRef struct {
	kind        TypeRefKind
	name        string
	module      string
	nullable    bool
	args        []ParsedTypeRef
	params      []ParsedTypeParam
	returnTypes []ParsedTypeRef
	line        int
	column      int
}

type ParseError struct {
	message string
	line    int
	column  int
}

type ParsedImport struct {
	path   string
	go_    bool
	module bool
	line   int
	column int
}

type ParsedConst struct {
	name    string
	private bool
	typeRef ParsedTypeRef
	value   ParsedExpr
	line    int
	column  int
}

type ParsedAnnotation struct {
	marker string
	module string
	name   string
	args   []ParsedExpr
	line   int
	column int
}

type ParsedParam struct {
	name    string
	typeRef ParsedTypeRef
	line    int
	column  int
}

type ParsedExpr struct {
	kind     ExprKind
	text     string
	name     string
	value    string
	op       string
	params   []ParsedParam
	children []ParsedExpr
	line     int
	column   int
}

type ParsedField struct {
	name        string
	private     bool
	annotations []ParsedAnnotation
	typeRef     ParsedTypeRef
	line        int
	column      int
}

type ParsedEnumMember struct {
	name        string
	private     bool
	annotations []ParsedAnnotation
	value       string
	params      []ParsedParam
	line        int
	column      int
}

type ParsedFunction struct {
	name         string
	private      bool
	static       bool
	routine      bool
	macro        bool
	annotations  []ParsedAnnotation
	receiverType string
	generics     []string
	params       []ParsedParam
	returnType   ParsedTypeRef
	body         ParsedExpr
	line         int
	column       int
}

type ParsedType struct {
	name        string
	private     bool
	enum        bool
	annotations []ParsedAnnotation
	generics    []string
	fields      []ParsedField
	methods     []ParsedFunction
	members     []ParsedEnumMember
	line        int
	column      int
}

type ParsedTest struct {
	name   string
	body   ParsedExpr
	line   int
	column int
}

type ParsedFile struct {
	imports   []ParsedImport
	constants []ParsedConst
	types     []ParsedType
	functions []ParsedFunction
	tests     []ParsedTest
	errors    []ParseError
}

type ParserState struct {
	tokens  []Token
	current int
	errors  []ParseError
}

type TokenStep struct {
	state ParserState
	token Token
}

type BoolStep struct {
	state ParserState
	ok    bool
}

type StringStep struct {
	state ParserState
	value string
}

type TypeRefStep struct {
	state   ParserState
	typeRef ParsedTypeRef
}

type TypeRefListStep struct {
	state ParserState
	refs  []ParsedTypeRef
}

type TypeParamStep struct {
	state ParserState
	param ParsedTypeParam
}

type TypeParamListStep struct {
	state  ParserState
	params []ParsedTypeParam
}

type StringListStep struct {
	state  ParserState
	values []string
}

type AnnotationListStep struct {
	state       ParserState
	annotations []ParsedAnnotation
}

type ParamListStep struct {
	state  ParserState
	params []ParsedParam
}

type ExprStep struct {
	state ParserState
	expr  ParsedExpr
}

type TemplateParse struct {
	text     string
	children []ParsedExpr
}

type XMLAttrStep struct {
	state       ParserState
	element     ParsedExpr
	selfClosing bool
}

type FileStep struct {
	state ParserState
	file  ParsedFile
}

type ImportStep struct {
	state      ParserState
	importDecl ParsedImport
}

type ConstStep struct {
	state     ParserState
	constDecl ParsedConst
}

type FunctionStep struct {
	state    ParserState
	function ParsedFunction
}

type TypeStep struct {
	state    ParserState
	typeDecl ParsedType
}

type TestStep struct {
	state    ParserState
	testDecl ParsedTest
}

type FieldStep struct {
	state ParserState
	field ParsedField
}

type EnumMemberStep struct {
	state  ParserState
	member ParsedEnumMember
}

type EnumMemberPayloadStep struct {
	state  ParserState
	value  string
	params []ParsedParam
}

type AnnotationStep struct {
	state      ParserState
	annotation ParsedAnnotation
}

type IRImport struct {
	path   string
	go_    bool
	module bool
	line   int
	column int
}

type IRTSImport struct {
	path      string
	specifier string
	functions []IRFunction
	values    []IRConst
	line      int
	column    int
}

type IRParam struct {
	name     string
	typeName string
	line     int
	column   int
}

type IRExpr struct {
	kind     ExprKind
	text     string
	name     string
	value    string
	op       string
	params   []IRParam
	children []IRExpr
	line     int
	column   int
}

type IRField struct {
	name       string
	private    bool
	typeName   string
	jsonName   string
	jsonIgnore bool
	line       int
	column     int
}

type IREnumMember struct {
	name    string
	private bool
	value   string
	params  []IRParam
	line    int
	column  int
}

type IRFunction struct {
	name         string
	private      bool
	static       bool
	routine      bool
	macro        bool
	receiverType string
	generics     []string
	params       []IRParam
	returnType   string
	body         IRExpr
	sourcePath   string
	line         int
	column       int
}

type IRConst struct {
	name     string
	private  bool
	typeName string
	value    IRExpr
	line     int
	column   int
}

type IRStructType struct {
	name       string
	private    bool
	generics   []string
	fields     []IRField
	methods    []IRFunction
	sourcePath string
	line       int
	column     int
}

type IREnumType struct {
	name       string
	private    bool
	generics   []string
	members    []IREnumMember
	methods    []IRFunction
	sourcePath string
	line       int
	column     int
}

type IRTest struct {
	name   string
	body   IRExpr
	line   int
	column int
}

type IRFile struct {
	imports   []IRImport
	tsImports []IRTSImport
	structs   []IRStructType
	enums     []IREnumType
	constants []IRConst
	functions []IRFunction
	tests     []IRTest
	errors    []ParseError
}

type CompilerCallable struct {
	name       string
	arity      int
	returnType string
	paramTypes []string
	private    bool
	sourcePath string
}

type CompilerTypeBinding struct {
	name     string
	typeName string
}

type InferBlockOut struct {
	statements []IRExpr
	bindings   []CompilerTypeBinding
}

type GoJSONDeclResult struct {
	names []string
	text  string
}

type CompileResult struct {
	ok     bool
	output string
	errors []string
}

type SourceFile struct {
	path   string
	source string
}

type SelfhostSources struct {
	files []SourceFile
}

type CompilerMacroBinding struct {
	name       string
	macro      bool
	paramTypes []string
}

type CompilerNamespaceAlias struct {
	name       string
	module     string
	importPath string
	go_        bool
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

func selfhost_lexer_lexer_xmlModeCode() int {
	return 0
}

func selfhost_lexer_lexer_xmlModeTag() int {
	return 1
}

func selfhost_lexer_lexer_xmlModeText() int {
	return 2
}

func selfhost_lexer_lexer_xmlModeExpr() int {
	return 3
}

func lex(source string) []Token {
	return selfhost_lexer_lexer_scan(LexState{source: source, start: 0, current: 0, line: 1, column: 1, startLine: 1, startColumn: 1, canStartRegex: true, mode: selfhost_lexer_lexer_xmlModeCode(), xmlDepth: 0, xmlClosing: false, xmlSelfClosed: false, xmlExprMode: selfhost_lexer_lexer_xmlModeCode(), xmlExprDepth: 0, xmlAfterMode: selfhost_lexer_lexer_xmlModeCode(), xmlAfterDepth: 0}, selfhost_lexer_lexer_emptyTokens())
}

func selfhost_lexer_lexer_emptyTokens() []Token {
	return []Token{}
}

func selfhost_lexer_lexer_scan(state LexState, tokens []Token) []Token {
	prepared := selfhost_lexer_lexer_prepareLexState(state)
	started := selfhost_lexer_lexer_markStart(prepared)
	return func() []Token {
		if selfhost_lexer_lexer_atEnd(started) {
			return selfhost_lexer_lexer_appendToken(tokens, selfhost_lexer_lexer_makeToken(started, TokenKind_EOF))
		}
		return selfhost_lexer_lexer_scanLexed(selfhost_lexer_lexer_scanModeToken(started), tokens)
	}()
}

func selfhost_lexer_lexer_prepareLexState(state LexState) LexState {
	return func() LexState {
		switch {
		case state.mode == selfhost_lexer_lexer_xmlModeCode() || state.mode == selfhost_lexer_lexer_xmlModeExpr() == true:
			return selfhost_lexer_lexer_skipIgnored(state)
		default:
			return selfhost_lexer_lexer_skipXMLSpaces(state)
		}
	}()
}

func selfhost_lexer_lexer_scanModeToken(state LexState) Lexed {
	return func() Lexed {
		switch {
		case state.mode == 1:
			return selfhost_lexer_lexer_scanXMLTagToken(selfhost_lexer_lexer_advance(state))
		case state.mode == 2:
			return selfhost_lexer_lexer_scanXMLTextToken(state)
		case state.mode == 3:
			return selfhost_lexer_lexer_scanXMLExprToken(selfhost_lexer_lexer_advance(state))
		default:
			return selfhost_lexer_lexer_scanToken(selfhost_lexer_lexer_advance(state))
		}
	}()
}

func selfhost_lexer_lexer_scanLexed(lexed Lexed, tokens []Token) []Token {
	nextTokens := selfhost_lexer_lexer_appendToken(tokens, selfhost_lexer_lexer_makeToken(lexed.state, lexed.kind))
	return selfhost_lexer_lexer_scan(selfhost_lexer_lexer_finishToken(lexed.state, lexed.kind), nextTokens)
}

func selfhost_lexer_lexer_appendToken(tokens []Token, token Token) []Token {
	tokens = append(tokens, token)
	return tokens
}

func selfhost_lexer_lexer_makeToken(state LexState, kind TokenKind) Token {
	return Token{kind: kind, lexeme: func() string { runes := []rune(state.source); return string(runes[state.start:state.current]) }(), offset: state.start, line: state.startLine, column: state.startColumn}
}

func selfhost_lexer_lexer_finishToken(state LexState, kind TokenKind) LexState {
	return LexState{source: state.source, start: state.start, current: state.current, line: state.line, column: state.column, startLine: state.startLine, startColumn: state.startColumn, mode: state.mode, xmlDepth: state.xmlDepth, xmlClosing: state.xmlClosing, xmlSelfClosed: state.xmlSelfClosed, xmlExprMode: state.xmlExprMode, xmlExprDepth: state.xmlExprDepth, xmlAfterMode: state.xmlAfterMode, xmlAfterDepth: state.xmlAfterDepth, canStartRegex: !(selfhost_lexer_lexer_canEndExpression(state, kind))}
}

func selfhost_lexer_lexer_lexStateXML(state LexState, mode int, depth int, closing bool, selfClosed bool, exprMode int, exprDepth int) LexState {
	return LexState{source: state.source, start: state.start, current: state.current, line: state.line, column: state.column, startLine: state.startLine, startColumn: state.startColumn, canStartRegex: state.canStartRegex, xmlAfterMode: state.xmlAfterMode, xmlAfterDepth: state.xmlAfterDepth, mode: mode, xmlDepth: depth, xmlClosing: closing, xmlSelfClosed: selfClosed, xmlExprMode: exprMode, xmlExprDepth: exprDepth}
}

func selfhost_lexer_lexer_enterXMLTagState(state LexState) LexState {
	depth := func() int {
		if state.mode == selfhost_lexer_lexer_xmlModeExpr() {
			return 0
		}
		return state.xmlDepth
	}()
	return selfhost_lexer_lexer_lexStateXMLAfterMode(selfhost_lexer_lexer_lexStateXML(state, selfhost_lexer_lexer_xmlModeTag(), depth, false, false, state.xmlExprMode, state.xmlExprDepth), state.mode, state.xmlDepth)
}

func selfhost_lexer_lexer_lexStateXMLAfterMode(state LexState, afterMode int, afterDepth int) LexState {
	return LexState{source: state.source, start: state.start, current: state.current, line: state.line, column: state.column, startLine: state.startLine, startColumn: state.startColumn, canStartRegex: state.canStartRegex, mode: state.mode, xmlDepth: state.xmlDepth, xmlClosing: state.xmlClosing, xmlSelfClosed: state.xmlSelfClosed, xmlExprMode: state.xmlExprMode, xmlExprDepth: state.xmlExprDepth, xmlAfterMode: afterMode, xmlAfterDepth: afterDepth}
}

func selfhost_lexer_lexer_canEndExpression(state LexState, kind TokenKind) bool {
	return selfhost_lexer_lexer_canEndValueToken(kind) || selfhost_lexer_lexer_canEndXmlLess(state, kind)
}

func selfhost_lexer_lexer_canEndValueToken(kind TokenKind) bool {
	return func() bool {
		switch {
		case (kind == TokenKind_Ident) || (kind == TokenKind_Int) || (kind == TokenKind_Double) || (kind == TokenKind_BigInt) || (kind == TokenKind_String) || (kind == TokenKind_TemplateString) || (kind == TokenKind_Char) || (kind == TokenKind_Regex) || (kind == TokenKind_XMLText) || (kind == TokenKind_RParen) || (kind == TokenKind_RBracket) || (kind == TokenKind_RBrace):
			return true
		default:
			return false
		}
	}()
}

func selfhost_lexer_lexer_canEndXmlLess(state LexState, kind TokenKind) bool {
	return kind == TokenKind_Less && (selfhost_lexer_lexer_peek(state) == '/' || selfhost_lexer_lexer_isIdentStart(selfhost_lexer_lexer_peek(state)))
}

func tokenKindName(kind TokenKind) string {
	return func() string {
		switch {
		case kind == TokenKind_EOF:
			return "EOF"
		case kind == TokenKind_Illegal:
			return "Illegal"
		case kind == TokenKind_Newline:
			return "Newline"
		case kind == TokenKind_Ident:
			return "Ident"
		case kind == TokenKind_Int:
			return "Int"
		case kind == TokenKind_Double:
			return "Double"
		case kind == TokenKind_BigInt:
			return "BigInt"
		case kind == TokenKind_String:
			return "String"
		case kind == TokenKind_TemplateString:
			return "TemplateString"
		case kind == TokenKind_Char:
			return "Char"
		case kind == TokenKind_Regex:
			return "Regex"
		case kind == TokenKind_XMLText:
			return "XMLText"
		case kind == TokenKind_Hash:
			return "Hash"
		case kind == TokenKind_At:
			return "At"
		case kind == TokenKind_Dollar:
			return "Dollar"
		case kind == TokenKind_Dot:
			return "Dot"
		case kind == TokenKind_DotDot:
			return "DotDot"
		case kind == TokenKind_DotDotLess:
			return "DotDotLess"
		case kind == TokenKind_DotDotEqual:
			return "DotDotEqual"
		case kind == TokenKind_DotDotDot:
			return "DotDotDot"
		case kind == TokenKind_Comma:
			return "Comma"
		case kind == TokenKind_Colon:
			return "Colon"
		case kind == TokenKind_DoubleColon:
			return "DoubleColon"
		case kind == TokenKind_LParen:
			return "LParen"
		case kind == TokenKind_RParen:
			return "RParen"
		case kind == TokenKind_LBracket:
			return "LBracket"
		case kind == TokenKind_RBracket:
			return "RBracket"
		case kind == TokenKind_LBrace:
			return "LBrace"
		case kind == TokenKind_RBrace:
			return "RBrace"
		case kind == TokenKind_Question:
			return "Question"
		case kind == TokenKind_QuestionQuestion:
			return "QuestionQuestion"
		case kind == TokenKind_Apostrophe:
			return "Apostrophe"
		case kind == TokenKind_FatArrow:
			return "FatArrow"
		case kind == TokenKind_Assign:
			return "Assign"
		case kind == TokenKind_Declare:
			return "Declare"
		case kind == TokenKind_MutDeclare:
			return "MutDeclare"
		case kind == TokenKind_Arrow:
			return "Arrow"
		case kind == TokenKind_Plus:
			return "Plus"
		case kind == TokenKind_PlusPlus:
			return "PlusPlus"
		case kind == TokenKind_Minus:
			return "Minus"
		case kind == TokenKind_Star:
			return "Star"
		case kind == TokenKind_Slash:
			return "Slash"
		case kind == TokenKind_Percent:
			return "Percent"
		case kind == TokenKind_Bang:
			return "Bang"
		case kind == TokenKind_Tilde:
			return "Tilde"
		case kind == TokenKind_BitAnd:
			return "BitAnd"
		case kind == TokenKind_BitOr:
			return "BitOr"
		case kind == TokenKind_BitXor:
			return "BitXor"
		case kind == TokenKind_ShiftLeft:
			return "ShiftLeft"
		case kind == TokenKind_ShiftRight:
			return "ShiftRight"
		case kind == TokenKind_UnsignedShiftRight:
			return "UnsignedShiftRight"
		case kind == TokenKind_AndAnd:
			return "AndAnd"
		case kind == TokenKind_OrOr:
			return "OrOr"
		case kind == TokenKind_EqualEqual:
			return "EqualEqual"
		case kind == TokenKind_BangEqual:
			return "BangEqual"
		case kind == TokenKind_Less:
			return "Less"
		case kind == TokenKind_LessEqual:
			return "LessEqual"
		case kind == TokenKind_Greater:
			return "Greater"
		case kind == TokenKind_GreaterEqual:
			return "GreaterEqual"
		case kind == TokenKind_Underscore:
			return "Underscore"
		default:
			return "Unknown"
		}
	}()
}

func selfhost_lexer_lexer_markStart(state LexState) LexState {
	return LexState{source: state.source, current: state.current, line: state.line, column: state.column, canStartRegex: state.canStartRegex, mode: state.mode, xmlDepth: state.xmlDepth, xmlClosing: state.xmlClosing, xmlSelfClosed: state.xmlSelfClosed, xmlExprMode: state.xmlExprMode, xmlExprDepth: state.xmlExprDepth, xmlAfterMode: state.xmlAfterMode, xmlAfterDepth: state.xmlAfterDepth, start: state.current, startLine: state.line, startColumn: state.column}
}

func selfhost_lexer_lexer_atEnd(state LexState) bool {
	return state.current >= len([]rune(state.source))
}

func selfhost_lexer_lexer_charAt(source string, index int) rune {
	return func() rune {
		if index < 0 || index >= len([]rune(source)) {
			return ' '
		}
		return []rune(source)[index]
	}()
}

func selfhost_lexer_lexer_peek(state LexState) rune {
	return selfhost_lexer_lexer_charAt(state.source, state.current)
}

func selfhost_lexer_lexer_peekNext(state LexState) rune {
	return selfhost_lexer_lexer_charAt(state.source, state.current+1)
}

func selfhost_lexer_lexer_advanceState(state LexState) LexState {
	return selfhost_lexer_lexer_advance(state).state
}

func selfhost_lexer_lexer_advance(state LexState) Advanced {
	return func() Advanced {
		if selfhost_lexer_lexer_atEnd(state) {
			return selfhost_lexer_lexer_advanced(state, ' ')
		}
		return selfhost_lexer_lexer_advanceChar(state, []rune(state.source)[state.current])
	}()
}

func selfhost_lexer_lexer_advanced(state LexState, ch rune) Advanced {
	return Advanced{state: state, ch: ch}
}

func selfhost_lexer_lexer_advanceChar(state LexState, ch rune) Advanced {
	return Advanced{state: LexState{source: state.source, start: state.start, current: state.current + 1, line: func() int {
		if ch == '\n' {
			return state.line + 1
		}
		return state.line
	}(), column: func() int {
		if ch == '\n' {
			return 1
		}
		return state.column + 1
	}(), startLine: state.startLine, startColumn: state.startColumn, canStartRegex: state.canStartRegex, mode: state.mode, xmlDepth: state.xmlDepth, xmlClosing: state.xmlClosing, xmlSelfClosed: state.xmlSelfClosed, xmlExprMode: state.xmlExprMode, xmlExprDepth: state.xmlExprDepth, xmlAfterMode: state.xmlAfterMode, xmlAfterDepth: state.xmlAfterDepth}, ch: ch}
}

func selfhost_lexer_lexer_skipIgnored(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_atEnd(state) {
			return state
		}
		return func() LexState {
			if selfhost_lexer_lexer_isSpace(selfhost_lexer_lexer_peek(state)) {
				return selfhost_lexer_lexer_skipIgnored(selfhost_lexer_lexer_advanceState(state))
			}
			return func() LexState {
				if selfhost_lexer_lexer_startsWith(state, '/', '/') {
					return selfhost_lexer_lexer_skipIgnored(selfhost_lexer_lexer_skipLineComment(selfhost_lexer_lexer_advanceState(selfhost_lexer_lexer_advanceState(state))))
				}
				return func() LexState {
					if selfhost_lexer_lexer_startsWith(state, '/', '*') {
						return selfhost_lexer_lexer_skipIgnored(selfhost_lexer_lexer_skipBlockComment(selfhost_lexer_lexer_advanceState(selfhost_lexer_lexer_advanceState(state))))
					}
					return state
				}()
			}()
		}()
	}()
}

func selfhost_lexer_lexer_startsWith(state LexState, first rune, second rune) bool {
	return selfhost_lexer_lexer_peek(state) == first && selfhost_lexer_lexer_peekNext(state) == second
}

func selfhost_lexer_lexer_skipLineComment(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_atEnd(state) || selfhost_lexer_lexer_peek(state) == '\n' {
			return state
		}
		return selfhost_lexer_lexer_skipLineComment(selfhost_lexer_lexer_advanceState(state))
	}()
}

func selfhost_lexer_lexer_skipBlockComment(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_atEnd(state) {
			return state
		}
		return func() LexState {
			if selfhost_lexer_lexer_startsWith(state, '*', '/') {
				return selfhost_lexer_lexer_advanceState(selfhost_lexer_lexer_advanceState(state))
			}
			return selfhost_lexer_lexer_skipBlockComment(selfhost_lexer_lexer_advanceState(state))
		}()
	}()
}

func selfhost_lexer_lexer_isSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\r'
}

func selfhost_lexer_lexer_isXMLSpace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

func selfhost_lexer_lexer_skipXMLSpaces(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_atEnd(state) {
			return state
		}
		return func() LexState {
			if selfhost_lexer_lexer_isXMLSpace(selfhost_lexer_lexer_peek(state)) {
				return selfhost_lexer_lexer_skipXMLSpaces(selfhost_lexer_lexer_advanceState(state))
			}
			return state
		}()
	}()
}

func selfhost_lexer_lexer_scanToken(step Advanced) Lexed {
	destructure1 := step
	state := destructure1.state
	ch := destructure1.ch
	return func() Lexed {
		switch {
		case ch == '\n':
			return selfhost_lexer_lexer_lexed(state, TokenKind_Newline)
		case ch == '#':
			return selfhost_lexer_lexer_lexed(state, TokenKind_Hash)
		case ch == '@':
			return selfhost_lexer_lexer_lexed(state, TokenKind_At)
		case ch == '$':
			return selfhost_lexer_lexer_lexDollar(state)
		case ch == '.':
			return selfhost_lexer_lexer_lexDot(state)
		case ch == ',':
			return selfhost_lexer_lexer_lexed(state, TokenKind_Comma)
		case ch == ':':
			return selfhost_lexer_lexer_lexColon(state)
		case ch == '(':
			return selfhost_lexer_lexer_lexed(state, TokenKind_LParen)
		case ch == ')':
			return selfhost_lexer_lexer_lexed(state, TokenKind_RParen)
		case ch == '[':
			return selfhost_lexer_lexer_lexed(state, TokenKind_LBracket)
		case ch == ']':
			return selfhost_lexer_lexer_lexed(state, TokenKind_RBracket)
		case ch == '{':
			return selfhost_lexer_lexer_lexed(state, TokenKind_LBrace)
		case ch == '}':
			return selfhost_lexer_lexer_lexed(state, TokenKind_RBrace)
		case ch == '?':
			return selfhost_lexer_lexer_lexQuestion(state)
		case ch == '+':
			return selfhost_lexer_lexer_lexPlus(state)
		case ch == '-':
			return selfhost_lexer_lexer_lexMinus(state)
		case ch == '*':
			return selfhost_lexer_lexer_lexed(state, TokenKind_Star)
		case ch == '/':
			return func() Lexed {
				if state.canStartRegex {
					return selfhost_lexer_lexer_lexRegexToken(state)
				}
				return selfhost_lexer_lexer_lexed(state, TokenKind_Slash)
			}()
		case ch == '%':
			return selfhost_lexer_lexer_lexed(state, TokenKind_Percent)
		case ch == '!':
			return selfhost_lexer_lexer_lexBang(state)
		case ch == '~':
			return selfhost_lexer_lexer_lexTilde(state)
		case ch == '&':
			return selfhost_lexer_lexer_lexAmp(state)
		case ch == '|':
			return selfhost_lexer_lexer_lexPipe(state)
		case ch == '^':
			return selfhost_lexer_lexer_lexed(state, TokenKind_BitXor)
		case ch == '=':
			return selfhost_lexer_lexer_lexEqual(state)
		case ch == '<':
			return selfhost_lexer_lexer_lexLess(state)
		case ch == '>':
			return selfhost_lexer_lexer_lexGreater(state)
		case ch == '"':
			return selfhost_lexer_lexer_lexStringToken(state)
		case ch == '`':
			return selfhost_lexer_lexer_lexTemplateStringToken(state)
		case ch == '\'':
			return func() Lexed {
				if state.canStartRegex {
					return selfhost_lexer_lexer_lexCharToken(state)
				}
				return selfhost_lexer_lexer_lexed(state, TokenKind_Apostrophe)
			}()
		case ch == '_':
			return func() Lexed {
				if selfhost_lexer_lexer_isIdentContinue(selfhost_lexer_lexer_peek(state)) {
					return selfhost_lexer_lexer_lexIdentifierToken(state)
				}
				return selfhost_lexer_lexer_lexed(state, TokenKind_Underscore)
			}()
		default:
			return func() Lexed {
				if selfhost_lexer_lexer_isDigit(ch) {
					return selfhost_lexer_lexer_lexNumberToken(state)
				}
				return func() Lexed {
					if selfhost_lexer_lexer_isIdentStart(ch) {
						return selfhost_lexer_lexer_lexIdentifierToken(state)
					}
					return selfhost_lexer_lexer_lexed(state, TokenKind_Illegal)
				}()
			}()
		}
	}()
}

func selfhost_lexer_lexer_scanXMLTagToken(step Advanced) Lexed {
	destructure2 := step
	state := destructure2.state
	ch := destructure2.ch
	return func() Lexed {
		switch {
		case ch == '@':
			return selfhost_lexer_lexer_lexed(state, TokenKind_At)
		case ch == '=':
			return selfhost_lexer_lexer_lexed(state, TokenKind_Assign)
		case ch == '{':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_lexStateXML(state, selfhost_lexer_lexer_xmlModeExpr(), state.xmlDepth, state.xmlClosing, state.xmlSelfClosed, selfhost_lexer_lexer_xmlModeTag(), 1), TokenKind_LBrace)
		case ch == '/':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_lexStateXML(state, state.mode, state.xmlDepth, selfhost_lexer_lexer_peek(state) != '>', selfhost_lexer_lexer_peek(state) == '>', state.xmlExprMode, state.xmlExprDepth), TokenKind_Slash)
		case ch == '>':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_xmlStateAfterTagEnd(state), TokenKind_Greater)
		case ch == '"':
			return selfhost_lexer_lexer_lexStringToken(state)
		default:
			return func() Lexed {
				if selfhost_lexer_lexer_isIdentStart(ch) {
					return selfhost_lexer_lexer_lexXMLIdentifierToken(state)
				}
				return selfhost_lexer_lexer_lexed(state, TokenKind_Illegal)
			}()
		}
	}()
}

func selfhost_lexer_lexer_xmlStateAfterTagEnd(state LexState) LexState {
	nextDepth := func() int {
		if state.xmlClosing {
			return state.xmlDepth - 1
		}
		return func() int {
			if state.xmlSelfClosed {
				return state.xmlDepth
			}
			return state.xmlDepth + 1
		}()
	}()
	nextMode := func() int {
		if nextDepth > 0 {
			return selfhost_lexer_lexer_xmlModeText()
		}
		return state.xmlAfterMode
	}()
	restoredDepth := func() int {
		if nextDepth > 0 {
			return nextDepth
		}
		return state.xmlAfterDepth
	}()
	return selfhost_lexer_lexer_lexStateXML(state, nextMode, restoredDepth, false, false, state.xmlExprMode, state.xmlExprDepth)
}

func selfhost_lexer_lexer_scanXMLTextToken(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '<':
			return selfhost_lexer_lexer_scanXMLTextLess(selfhost_lexer_lexer_advanceState(state))
		case selfhost_lexer_lexer_peek(state) == '{':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_lexStateXML(selfhost_lexer_lexer_advanceState(state), selfhost_lexer_lexer_xmlModeExpr(), state.xmlDepth, state.xmlClosing, state.xmlSelfClosed, selfhost_lexer_lexer_xmlModeText(), 1), TokenKind_LBrace)
		default:
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_scanXMLTextContent(state), TokenKind_XMLText)
		}
	}()
}

func selfhost_lexer_lexer_scanXMLTextLess(state LexState) Lexed {
	return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_lexStateXML(state, selfhost_lexer_lexer_xmlModeTag(), state.xmlDepth, selfhost_lexer_lexer_peek(state) == '/', false, state.xmlExprMode, state.xmlExprDepth), TokenKind_Less)
}

func selfhost_lexer_lexer_scanXMLTextContent(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_atEnd(state) || selfhost_lexer_lexer_peek(state) == '<' || selfhost_lexer_lexer_peek(state) == '{' {
			return state
		}
		return selfhost_lexer_lexer_scanXMLTextContent(selfhost_lexer_lexer_advanceState(state))
	}()
}

func selfhost_lexer_lexer_scanXMLExprToken(step Advanced) Lexed {
	destructure3 := step
	state := destructure3.state
	ch := destructure3.ch
	return func() Lexed {
		switch {
		case ch == '{':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_lexStateXML(state, selfhost_lexer_lexer_xmlModeExpr(), state.xmlDepth, state.xmlClosing, state.xmlSelfClosed, state.xmlExprMode, state.xmlExprDepth+1), TokenKind_LBrace)
		case ch == '}':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_xmlStateAfterExprBrace(state), TokenKind_RBrace)
		default:
			return selfhost_lexer_lexer_scanToken(step)
		}
	}()
}

func selfhost_lexer_lexer_xmlStateAfterExprBrace(state LexState) LexState {
	nextDepth := state.xmlExprDepth - 1
	nextMode := func() int {
		if nextDepth <= 0 {
			return state.xmlExprMode
		}
		return selfhost_lexer_lexer_xmlModeExpr()
	}()
	return selfhost_lexer_lexer_lexStateXMLAfterMode(selfhost_lexer_lexer_lexStateXML(state, nextMode, state.xmlDepth, state.xmlClosing, state.xmlSelfClosed, selfhost_lexer_lexer_xmlModeCode(), nextDepth), selfhost_lexer_lexer_xmlModeCode(), 0)
}

func selfhost_lexer_lexer_lexed(state LexState, kind TokenKind) Lexed {
	return Lexed{state: state, kind: kind}
}

func selfhost_lexer_lexer_lexDot(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '.':
			return selfhost_lexer_lexer_lexDotDot(selfhost_lexer_lexer_advanceState(state))
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Dot)
		}
	}()
}

func selfhost_lexer_lexer_lexDotDot(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '.':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_DotDotDot)
		case selfhost_lexer_lexer_peek(state) == '<':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_DotDotLess)
		case selfhost_lexer_lexer_peek(state) == '=':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_DotDotEqual)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_DotDot)
		}
	}()
}

func selfhost_lexer_lexer_lexColon(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '=':
			return func() Lexed {
				if selfhost_lexer_lexer_peek(selfhost_lexer_lexer_advanceState(state)) == ':' {
					return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(selfhost_lexer_lexer_advanceState(state)), TokenKind_MutDeclare)
				}
				return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_Declare)
			}()
		case selfhost_lexer_lexer_peek(state) == ':':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_DoubleColon)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Colon)
		}
	}()
}

func selfhost_lexer_lexer_lexQuestion(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '?':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_QuestionQuestion)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Question)
		}
	}()
}

func selfhost_lexer_lexer_lexPlus(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '+':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_PlusPlus)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Plus)
		}
	}()
}

func selfhost_lexer_lexer_lexMinus(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '>':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_Arrow)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Minus)
		}
	}()
}

func selfhost_lexer_lexer_lexBang(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '=':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_BangEqual)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Bang)
		}
	}()
}

func selfhost_lexer_lexer_lexTilde(state LexState) Lexed {
	return func() Lexed {
		switch {
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Tilde)
		}
	}()
}

func selfhost_lexer_lexer_lexDollar(state LexState) Lexed {
	return func() Lexed {
		switch {
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Dollar)
		}
	}()
}

func selfhost_lexer_lexer_lexAmp(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '&':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_AndAnd)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_BitAnd)
		}
	}()
}

func selfhost_lexer_lexer_lexPipe(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '|':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_OrOr)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_BitOr)
		}
	}()
}

func selfhost_lexer_lexer_lexEqual(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '>':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_FatArrow)
		case selfhost_lexer_lexer_peek(state) == '=':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_EqualEqual)
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Assign)
		}
	}()
}

func selfhost_lexer_lexer_lexLess(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '=':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_LessEqual)
		case selfhost_lexer_lexer_peek(state) == '<':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_ShiftLeft)
		case selfhost_lexer_lexer_peek(state) == '/':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_enterXMLTagState(state), TokenKind_Less)
		default:
			return func() Lexed {
				if selfhost_lexer_lexer_isIdentStart(selfhost_lexer_lexer_peek(state)) {
					return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_enterXMLTagState(state), TokenKind_Less)
				}
				return selfhost_lexer_lexer_lexed(state, TokenKind_Less)
			}()
		}
	}()
}

func selfhost_lexer_lexer_lexGreater(state LexState) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == '=':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_GreaterEqual)
		case selfhost_lexer_lexer_peek(state) == '>':
			return func() Lexed {
				if selfhost_lexer_lexer_startsWith(state, '>', '>') {
					return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(selfhost_lexer_lexer_advanceState(state)), TokenKind_UnsignedShiftRight)
				}
				return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_ShiftRight)
			}()
		default:
			return selfhost_lexer_lexer_lexed(state, TokenKind_Greater)
		}
	}()
}

func selfhost_lexer_lexer_lexStringToken(state LexState) Lexed {
	scanned := selfhost_lexer_lexer_scanString(state, false)
	return selfhost_lexer_lexer_lexed(scanned.state, func() TokenKind {
		if scanned.ok {
			return TokenKind_String
		}
		return TokenKind_Illegal
	}())
}

func selfhost_lexer_lexer_lexTemplateStringToken(state LexState) Lexed {
	scanned := selfhost_lexer_lexer_scanTemplateString(state, false)
	return selfhost_lexer_lexer_lexed(scanned.state, func() TokenKind {
		if scanned.ok {
			return TokenKind_TemplateString
		}
		return TokenKind_Illegal
	}())
}

func selfhost_lexer_lexer_scanString(state LexState, escaped bool) ScannedString {
	return func() ScannedString {
		if selfhost_lexer_lexer_atEnd(state) {
			return selfhost_lexer_lexer_scannedString(state, false)
		}
		return selfhost_lexer_lexer_scanStringStep(selfhost_lexer_lexer_advance(state), escaped)
	}()
}

func selfhost_lexer_lexer_scanTemplateString(state LexState, escaped bool) ScannedString {
	return func() ScannedString {
		if selfhost_lexer_lexer_atEnd(state) {
			return selfhost_lexer_lexer_scannedString(state, false)
		}
		return selfhost_lexer_lexer_scanTemplateStringStep(selfhost_lexer_lexer_advance(state), escaped)
	}()
}

func selfhost_lexer_lexer_scannedString(state LexState, ok bool) ScannedString {
	return ScannedString{state: state, ok: ok}
}

func selfhost_lexer_lexer_scanStringStep(step Advanced, escaped bool) ScannedString {
	return func() ScannedString {
		if escaped {
			return selfhost_lexer_lexer_scanString(step.state, false)
		}
		return func() ScannedString {
			switch {
			case step.ch == '\\':
				return selfhost_lexer_lexer_scanString(step.state, true)
			case step.ch == '"':
				return selfhost_lexer_lexer_scannedString(step.state, true)
			default:
				return selfhost_lexer_lexer_scanString(step.state, false)
			}
		}()
	}()
}

func selfhost_lexer_lexer_scanTemplateStringStep(step Advanced, escaped bool) ScannedString {
	return func() ScannedString {
		if escaped {
			return selfhost_lexer_lexer_scanTemplateString(step.state, false)
		}
		return func() ScannedString {
			switch {
			case step.ch == '\\':
				return selfhost_lexer_lexer_scanTemplateString(step.state, true)
			case step.ch == '`':
				return selfhost_lexer_lexer_scannedString(step.state, true)
			default:
				return selfhost_lexer_lexer_scanTemplateString(step.state, false)
			}
		}()
	}()
}

func selfhost_lexer_lexer_lexCharToken(state LexState) Lexed {
	scanned := selfhost_lexer_lexer_scanChar(state, false)
	return selfhost_lexer_lexer_lexed(scanned.state, func() TokenKind {
		if scanned.ok {
			return TokenKind_Char
		}
		return TokenKind_Illegal
	}())
}

func selfhost_lexer_lexer_scanChar(state LexState, escaped bool) ScannedString {
	return func() ScannedString {
		if selfhost_lexer_lexer_atEnd(state) {
			return selfhost_lexer_lexer_scannedString(state, false)
		}
		return selfhost_lexer_lexer_scanCharStep(selfhost_lexer_lexer_advance(state), escaped)
	}()
}

func selfhost_lexer_lexer_scanCharStep(step Advanced, escaped bool) ScannedString {
	return func() ScannedString {
		switch {
		case step.ch == '\n':
			return selfhost_lexer_lexer_scannedString(step.state, false)
		default:
			return func() ScannedString {
				if escaped {
					return selfhost_lexer_lexer_scanChar(step.state, false)
				}
				return func() ScannedString {
					switch {
					case step.ch == '\\':
						return selfhost_lexer_lexer_scanChar(step.state, true)
					case step.ch == '\'':
						return selfhost_lexer_lexer_scannedString(step.state, true)
					default:
						return selfhost_lexer_lexer_scanChar(step.state, false)
					}
				}()
			}()
		}
	}()
}

func selfhost_lexer_lexer_lexRegexToken(state LexState) Lexed {
	scanned := selfhost_lexer_lexer_scanRegex(state, false, false)
	return selfhost_lexer_lexer_lexed(scanned.state, func() TokenKind {
		if scanned.ok {
			return TokenKind_Regex
		}
		return TokenKind_Illegal
	}())
}

func selfhost_lexer_lexer_scanRegex(state LexState, escaped bool, inClass bool) ScannedString {
	return func() ScannedString {
		if selfhost_lexer_lexer_atEnd(state) {
			return selfhost_lexer_lexer_scannedString(state, false)
		}
		return selfhost_lexer_lexer_scanRegexStep(selfhost_lexer_lexer_advance(state), escaped, inClass)
	}()
}

func selfhost_lexer_lexer_scanRegexStep(step Advanced, escaped bool, inClass bool) ScannedString {
	return func() ScannedString {
		switch {
		case step.ch == '\n':
			return selfhost_lexer_lexer_scannedString(step.state, false)
		default:
			return func() ScannedString {
				if escaped {
					return selfhost_lexer_lexer_scanRegex(step.state, false, inClass)
				}
				return func() ScannedString {
					switch {
					case step.ch == '\\':
						return selfhost_lexer_lexer_scanRegex(step.state, true, inClass)
					case step.ch == '[':
						return selfhost_lexer_lexer_scanRegex(step.state, false, true)
					case step.ch == ']':
						return selfhost_lexer_lexer_scanRegex(step.state, false, false)
					case step.ch == '/':
						return func() ScannedString {
							if inClass {
								return selfhost_lexer_lexer_scanRegex(step.state, false, inClass)
							}
							return selfhost_lexer_lexer_scannedString(selfhost_lexer_lexer_scanRegexFlags(step.state), true)
						}()
					default:
						return selfhost_lexer_lexer_scanRegex(step.state, false, inClass)
					}
				}()
			}()
		}
	}()
}

func selfhost_lexer_lexer_scanRegexFlags(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_isRegexFlag(selfhost_lexer_lexer_peek(state)) {
			return selfhost_lexer_lexer_scanRegexFlags(selfhost_lexer_lexer_advanceState(state))
		}
		return state
	}()
}

func selfhost_lexer_lexer_lexNumberToken(state LexState) Lexed {
	return selfhost_lexer_lexer_lexNumberAfterDigits(selfhost_lexer_lexer_scanDigits(state), false)
}

func selfhost_lexer_lexer_scanDigits(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_isDigit(selfhost_lexer_lexer_peek(state)) {
			return selfhost_lexer_lexer_scanDigits(selfhost_lexer_lexer_advanceState(state))
		}
		return state
	}()
}

func selfhost_lexer_lexer_lexNumberAfterDigits(state LexState, isDouble bool) Lexed {
	return func() Lexed {
		switch {
		case selfhost_lexer_lexer_peek(state) == 'n':
			return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_advanceState(state), TokenKind_BigInt)
		case selfhost_lexer_lexer_peek(state) == '.':
			return func() Lexed {
				if selfhost_lexer_lexer_isDigit(selfhost_lexer_lexer_peekNext(state)) {
					return selfhost_lexer_lexer_lexNumberAfterDot(selfhost_lexer_lexer_advanceState(state))
				}
				return selfhost_lexer_lexer_lexed(state, func() TokenKind {
					if isDouble {
						return TokenKind_Double
					}
					return TokenKind_Int
				}())
			}()
		default:
			return func() Lexed {
				if selfhost_lexer_lexer_isExponentMarker(selfhost_lexer_lexer_peek(state)) {
					return selfhost_lexer_lexer_lexNumberAfterExponent(selfhost_lexer_lexer_advanceState(state))
				}
				return selfhost_lexer_lexer_lexed(state, func() TokenKind {
					if isDouble {
						return TokenKind_Double
					}
					return TokenKind_Int
				}())
			}()
		}
	}()
}

func selfhost_lexer_lexer_lexNumberAfterDot(state LexState) Lexed {
	return selfhost_lexer_lexer_lexNumberAfterDigits(selfhost_lexer_lexer_scanDigits(state), true)
}

func selfhost_lexer_lexer_lexNumberAfterExponent(state LexState) Lexed {
	return func() Lexed {
		if selfhost_lexer_lexer_isExponentSign(selfhost_lexer_lexer_peek(state)) {
			return selfhost_lexer_lexer_lexNumberExponentDigits(selfhost_lexer_lexer_advanceState(state))
		}
		return selfhost_lexer_lexer_lexNumberExponentDigits(state)
	}()
}

func selfhost_lexer_lexer_lexNumberExponentDigits(state LexState) Lexed {
	return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_scanDigits(state), TokenKind_Double)
}

func selfhost_lexer_lexer_lexIdentifierToken(state LexState) Lexed {
	return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_scanIdentifier(state), TokenKind_Ident)
}

func selfhost_lexer_lexer_lexXMLIdentifierToken(state LexState) Lexed {
	return selfhost_lexer_lexer_lexed(selfhost_lexer_lexer_scanXMLIdentifier(state), TokenKind_Ident)
}

func selfhost_lexer_lexer_scanIdentifier(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_isIdentContinue(selfhost_lexer_lexer_peek(state)) {
			return selfhost_lexer_lexer_scanIdentifier(selfhost_lexer_lexer_advanceState(state))
		}
		return state
	}()
}

func selfhost_lexer_lexer_scanXMLIdentifier(state LexState) LexState {
	return func() LexState {
		if selfhost_lexer_lexer_isIdentContinue(selfhost_lexer_lexer_peek(state)) || selfhost_lexer_lexer_peek(state) == '-' || selfhost_lexer_lexer_peek(state) == ':' {
			return selfhost_lexer_lexer_scanXMLIdentifier(selfhost_lexer_lexer_advanceState(state))
		}
		return state
	}()
}

func selfhost_lexer_lexer_isDigit(ch rune) bool {
	return ch >= '0' && ch <= '9'
}

func selfhost_lexer_lexer_isAlpha(ch rune) bool {
	return ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z'
}

func selfhost_lexer_lexer_isIdentStart(ch rune) bool {
	return ch == '_' || selfhost_lexer_lexer_isIdentText(ch) && selfhost_lexer_lexer_isDigit(ch) == false
}

func selfhost_lexer_lexer_isIdentContinue(ch rune) bool {
	return selfhost_lexer_lexer_isIdentStart(ch) || selfhost_lexer_lexer_isDigit(ch)
}

func selfhost_lexer_lexer_isIdentText(ch rune) bool {
	return selfhost_lexer_lexer_isIdentBoundary(ch) == false && selfhost_lexer_lexer_isAsciiPunctuation(ch) == false
}

func selfhost_lexer_lexer_isIdentBoundary(ch rune) bool {
	return ch == '\x00' || ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

func selfhost_lexer_lexer_isAsciiPunctuation(ch rune) bool {
	return ch >= '!' && ch <= '/' || ch >= ':' && ch <= '@' || ch >= '[' && ch <= '^' || ch == '`' || ch >= '{' && ch <= '~'
}

func selfhost_lexer_lexer_isRegexFlag(ch rune) bool {
	return selfhost_lexer_lexer_isAlpha(ch)
}

func selfhost_lexer_lexer_isExponentMarker(ch rune) bool {
	return ch == 'e' || ch == 'E'
}

func selfhost_lexer_lexer_isExponentSign(ch rune) bool {
	return ch == '+' || ch == '-'
}

func emptyParsedTypeRef() ParsedTypeRef {
	return ParsedTypeRef{kind: TypeRefKind_Unknown, name: "", module: "", nullable: false, args: []ParsedTypeRef{}, params: []ParsedTypeParam{}, returnTypes: []ParsedTypeRef{}, line: 0, column: 0}
}

func emptyParsedTypeParam() ParsedTypeParam {
	return ParsedTypeParam{name: "", optional: false, typeRef: emptyParsedTypeRef()}
}

func namedParsedTypeRef(token Token) ParsedTypeRef {
	return ParsedTypeRef{kind: TypeRefKind_Name, name: token.lexeme, module: "", nullable: false, args: []ParsedTypeRef{}, params: []ParsedTypeParam{}, returnTypes: []ParsedTypeRef{}, line: token.line, column: token.column}
}

func qualifiedParsedTypeRef(module Token, name Token) ParsedTypeRef {
	return ParsedTypeRef{kind: TypeRefKind_Name, name: name.lexeme, module: module.lexeme, nullable: false, args: []ParsedTypeRef{}, params: []ParsedTypeParam{}, returnTypes: []ParsedTypeRef{}, line: module.line, column: module.column}
}

func typeRefWithArgs(typeRef ParsedTypeRef, args []ParsedTypeRef) ParsedTypeRef {
	return ParsedTypeRef{kind: typeRef.kind, name: typeRef.name, module: typeRef.module, nullable: typeRef.nullable, args: args, params: typeRef.params, returnTypes: typeRef.returnTypes, line: typeRef.line, column: typeRef.column}
}

func nullableTypeRef(typeRef ParsedTypeRef) ParsedTypeRef {
	return ParsedTypeRef{kind: typeRef.kind, name: typeRef.name, module: typeRef.module, nullable: true, args: typeRef.args, params: typeRef.params, returnTypes: typeRef.returnTypes, line: typeRef.line, column: typeRef.column}
}

func groupedTypeRef(typeRef ParsedTypeRef, token Token) ParsedTypeRef {
	return ParsedTypeRef{kind: TypeRefKind_Group, name: "", module: "", nullable: false, args: []ParsedTypeRef{typeRef}, params: []ParsedTypeParam{}, returnTypes: []ParsedTypeRef{}, line: token.line, column: token.column}
}

func tupleTypeRef(params []ParsedTypeParam, token Token) ParsedTypeRef {
	return ParsedTypeRef{kind: TypeRefKind_Tuple, name: "", module: "", nullable: false, args: []ParsedTypeRef{}, params: params, returnTypes: []ParsedTypeRef{}, line: token.line, column: token.column}
}

func functionTypeRef(params []ParsedTypeParam, returnType ParsedTypeRef, token Token) ParsedTypeRef {
	return ParsedTypeRef{kind: TypeRefKind_Function, name: "", module: "", nullable: false, args: []ParsedTypeRef{}, params: params, returnTypes: []ParsedTypeRef{returnType}, line: token.line, column: token.column}
}

func typeRefToString(typeRef ParsedTypeRef) string {
	return func() string {
		switch {
		case typeRef.kind == TypeRefKind_Name:
			return selfhost_parser_ast_typeNameToString(typeRef)
		case typeRef.kind == TypeRefKind_Group:
			return func() string {
				if len(typeRef.args) == 0 {
					return "()"
				}
				return "(" + typeRefToString(typeRef.args[0]) + ")"
			}()
		case typeRef.kind == TypeRefKind_Tuple:
			return "(" + selfhost_parser_ast_typeParamsToString(typeRef.params, 0, "") + ")"
		case typeRef.kind == TypeRefKind_Function:
			return selfhost_parser_ast_functionTypeToString(typeRef)
		default:
			return ""
		}
	}()
}

func selfhost_parser_ast_typeNameToString(typeRef ParsedTypeRef) string {
	prefix := func() string {
		if typeRef.module == "" {
			return ""
		}
		return "@" + typeRef.module + "."
	}()
	args := func() string {
		if len(typeRef.args) == 0 {
			return ""
		}
		return "[" + selfhost_parser_ast_typeRefsToString(typeRef.args, 0, "") + "]"
	}()
	nullable := func() string {
		if typeRef.nullable {
			return "?"
		}
		return ""
	}()
	return prefix + typeRef.name + args + nullable
}

func selfhost_parser_ast_functionTypeToString(typeRef ParsedTypeRef) string {
	ret := func() string {
		if len(typeRef.returnTypes) == 0 {
			return ""
		}
		return typeRefToString(typeRef.returnTypes[0])
	}()
	return "(" + selfhost_parser_ast_typeParamsToString(typeRef.params, 0, "") + ")->" + ret
}

func selfhost_parser_ast_typeParamToString(param ParsedTypeParam) string {
	prefix := func() string {
		if param.name == "" {
			return ""
		}
		return param.name + func() string {
			if param.optional {
				return "?:"
			}
			return ":"
		}()
	}()
	return prefix + typeRefToString(param.typeRef)
}

func selfhost_parser_ast_typeRefsToString(refs []ParsedTypeRef, index int, out string) string {
	return func() string {
		if index >= len(refs) {
			return out
		}
		return selfhost_parser_ast_typeRefsToString(refs, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ","
		}()+typeRefToString(refs[index]))
	}()
}

func selfhost_parser_ast_typeParamsToString(params []ParsedTypeParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_parser_ast_typeParamsToString(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ","
		}()+selfhost_parser_ast_typeParamToString(params[index]))
	}()
}

func exprKindName(kind ExprKind) string {
	return func() string {
		switch {
		case kind == ExprKind_Identifier:
			return "identifier"
		case kind == ExprKind_At:
			return "at"
		case kind == ExprKind_This:
			return "this"
		case kind == ExprKind_Int:
			return "int"
		case kind == ExprKind_Double:
			return "double"
		case kind == ExprKind_BigInt:
			return "bigint"
		case kind == ExprKind_String:
			return "string"
		case kind == ExprKind_Template:
			return "template"
		case kind == ExprKind_Char:
			return "char"
		case kind == ExprKind_Regex:
			return "regex"
		case kind == ExprKind_XMLText:
			return "xmlText"
		case kind == ExprKind_Bool:
			return "bool"
		case kind == ExprKind_Null:
			return "null"
		case kind == ExprKind_Unary:
			return "unary"
		case kind == ExprKind_Postfix:
			return "postfix"
		case kind == ExprKind_Unwrap:
			return "unwrap"
		case kind == ExprKind_CompileTime:
			return "compileTime"
		case kind == ExprKind_Binary:
			return "binary"
		case kind == ExprKind_Ternary:
			return "ternary"
		case kind == ExprKind_Assign:
			return "assign"
		case kind == ExprKind_Call:
			return "call"
		case kind == ExprKind_Args:
			return "args"
		case kind == ExprKind_Lambda:
			return "lambda"
		case kind == ExprKind_Selector:
			return "selector"
		case kind == ExprKind_Index:
			return "index"
		case kind == ExprKind_Array:
			return "array"
		case kind == ExprKind_Tuple:
			return "tuple"
		case kind == ExprKind_Map:
			return "map"
		case kind == ExprKind_Entry:
			return "entry"
		case kind == ExprKind_Spread:
			return "spread"
		case kind == ExprKind_Reactive:
			return "reactive"
		case kind == ExprKind_Struct:
			return "struct"
		case kind == ExprKind_Object:
			return "object"
		case kind == ExprKind_Field:
			return "field"
		case kind == ExprKind_PrivateField:
			return "privateField"
		case kind == ExprKind_Method:
			return "method"
		case kind == ExprKind_PrivateMethod:
			return "privateMethod"
		case kind == ExprKind_Block:
			return "block"
		case kind == ExprKind_PatternBlock:
			return "patternBlock"
		case kind == ExprKind_Match:
			return "match"
		case kind == ExprKind_Branch:
			return "branch"
		case kind == ExprKind_Pattern:
			return "pattern"
		case kind == ExprKind_Let:
			return "let"
		case kind == ExprKind_ObjectDestructure:
			return "objectDestructure"
		case kind == ExprKind_Error:
			return "error"
		case kind == ExprKind_Watch:
			return "watch"
		case kind == ExprKind_XMLElement:
			return "xmlElement"
		default:
			return "unknown"
		}
	}()
}

func parse(source string) ParsedFile {
	return parseTokens(lex(source))
}

func parseTokens(tokens []Token) ParsedFile {
	errors := selfhost_parser_parser_emptyParseErrors()
	return selfhost_parser_parser_parseFileLoop(selfhost_parser_parser_parserSkipNewlines(ParserState{tokens: tokens, current: 0, errors: errors}), selfhost_parser_parser_emptyFile(errors)).file
}

func selfhost_parser_parser_emptyFile(errors []ParseError) ParsedFile {
	return ParsedFile{imports: []ParsedImport{}, constants: []ParsedConst{}, types: []ParsedType{}, functions: []ParsedFunction{}, tests: []ParsedTest{}, errors: errors}
}

func selfhost_parser_parser_emptyParseErrors() []ParseError {
	return []ParseError{}
}

func selfhost_parser_parser_emptyToken() Token {
	return Token{kind: TokenKind_EOF, lexeme: "", offset: 0, line: 0, column: 0}
}

func selfhost_parser_parser_emptyExpr() ParsedExpr {
	return ParsedExpr{kind: ExprKind_Unknown, text: "", name: "", value: "", op: "", params: []ParsedParam{}, children: []ParsedExpr{}, line: 0, column: 0}
}

func selfhost_parser_parser_emptyAnnotations() []ParsedAnnotation {
	return []ParsedAnnotation{}
}

func selfhost_parser_parser_makeExpr(kind ExprKind, text string, name string, value string, op string, params []ParsedParam, children []ParsedExpr, line int, column int) ParsedExpr {
	return ParsedExpr{kind: kind, text: text, name: name, value: value, op: op, params: params, children: children, line: line, column: column}
}

func selfhost_parser_parser_node(kind ExprKind, token Token) ParsedExpr {
	return selfhost_parser_parser_makeExpr(kind, token.lexeme, "", "", "", []ParsedParam{}, []ParsedExpr{}, token.line, token.column)
}

func selfhost_parser_parser_namedNode(kind ExprKind, name string, token Token) ParsedExpr {
	return selfhost_parser_parser_makeExpr(kind, token.lexeme, name, "", "", []ParsedParam{}, []ParsedExpr{}, token.line, token.column)
}

func selfhost_parser_parser_valueNode(kind ExprKind, value string, token Token) ParsedExpr {
	return selfhost_parser_parser_makeExpr(kind, token.lexeme, "", value, "", []ParsedParam{}, []ParsedExpr{}, token.line, token.column)
}

func selfhost_parser_parser_opNode(kind ExprKind, op string, token Token, children []ParsedExpr) ParsedExpr {
	return selfhost_parser_parser_makeExpr(kind, token.lexeme, "", "", op, []ParsedParam{}, children, token.line, token.column)
}

func selfhost_parser_parser_withChildren(expr ParsedExpr, children []ParsedExpr) ParsedExpr {
	return selfhost_parser_parser_makeExpr(expr.kind, expr.text, expr.name, expr.value, expr.op, expr.params, children, expr.line, expr.column)
}

func selfhost_parser_parser_withParams(expr ParsedExpr, params []ParsedParam) ParsedExpr {
	return selfhost_parser_parser_makeExpr(expr.kind, expr.text, expr.name, expr.value, expr.op, params, expr.children, expr.line, expr.column)
}

func selfhost_parser_parser_withText(expr ParsedExpr, text string) ParsedExpr {
	return selfhost_parser_parser_makeExpr(expr.kind, text, expr.name, expr.value, expr.op, expr.params, expr.children, expr.line, expr.column)
}

func selfhost_parser_parser_appendChild(expr ParsedExpr, child ParsedExpr) ParsedExpr {
	expr.children = append(expr.children, child)
	return expr
}

func selfhost_parser_parser_appendString(values []string, value string) []string {
	values = append(values, value)
	return values
}

func selfhost_parser_parser_parserPeek(state ParserState) Token {
	return func() Token {
		if state.current >= len(state.tokens) {
			return selfhost_parser_parser_emptyToken()
		}
		return state.tokens[state.current]
	}()
}

func selfhost_parser_parser_parserPrevious(state ParserState) Token {
	return func() Token {
		if state.current <= 0 {
			return selfhost_parser_parser_emptyToken()
		}
		return state.tokens[state.current-1]
	}()
}

func selfhost_parser_parser_parserTokenAt(state ParserState, index int) Token {
	return func() Token {
		if index >= len(state.tokens) {
			return selfhost_parser_parser_emptyToken()
		}
		return state.tokens[index]
	}()
}

func selfhost_parser_parser_parserKindAt(state ParserState, index int) TokenKind {
	return selfhost_parser_parser_parserTokenAt(state, index).kind
}

func selfhost_parser_parser_parserCheck(state ParserState, kind TokenKind) bool {
	return selfhost_parser_parser_parserPeek(state).kind == kind
}

func selfhost_parser_parser_parserCheckNext(state ParserState, kind TokenKind) bool {
	return selfhost_parser_parser_parserKindAt(state, state.current+1) == kind
}

func selfhost_parser_parser_stateAt(state ParserState, current int) ParserState {
	return ParserState{tokens: state.tokens, errors: state.errors, current: current}
}

func selfhost_parser_parser_parserAdvance(state ParserState) TokenStep {
	return func() TokenStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_EOF) {
			return TokenStep{state: state, token: selfhost_parser_parser_parserPeek(state)}
		}
		return TokenStep{state: selfhost_parser_parser_stateAt(state, state.current+1), token: selfhost_parser_parser_parserPeek(state)}
	}()
}

func selfhost_parser_parser_parserMatch(state ParserState, kind TokenKind) BoolStep {
	return func() BoolStep {
		if selfhost_parser_parser_parserCheck(state, kind) {
			return BoolStep{state: selfhost_parser_parser_parserAdvance(state).state, ok: true}
		}
		return BoolStep{state: state, ok: false}
	}()
}

func selfhost_parser_parser_parserConsume(state ParserState, kind TokenKind, message string) TokenStep {
	return func() TokenStep {
		if selfhost_parser_parser_parserCheck(state, kind) {
			return selfhost_parser_parser_parserAdvance(state)
		}
		return selfhost_parser_parser_parserConsumeMissing(selfhost_parser_parser_parserErrorAt(state, selfhost_parser_parser_parserPeek(state), message))
	}()
}

func selfhost_parser_parser_parserConsumeMissing(state ParserState) TokenStep {
	return func() TokenStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_EOF) {
			return TokenStep{state: state, token: selfhost_parser_parser_parserPeek(state)}
		}
		return selfhost_parser_parser_parserAdvance(state)
	}()
}

func selfhost_parser_parser_parserErrorAt(state ParserState, token Token, message string) ParserState {
	return ParserState{tokens: state.tokens, current: state.current, errors: selfhost_parser_parser_appendParseError(state.errors, ParseError{message: message, line: token.line, column: token.column})}
}

func selfhost_parser_parser_appendParseError(errors []ParseError, error_ ParseError) []ParseError {
	out := errors
	out = append(out, error_)
	return out
}

func selfhost_parser_parser_parserSkipNewlines(state ParserState) ParserState {
	return func() ParserState {
		if selfhost_parser_parser_parserCheck(state, TokenKind_Newline) {
			return selfhost_parser_parser_parserSkipNewlines(selfhost_parser_parser_parserAdvance(state).state)
		}
		return state
	}()
}

func selfhost_parser_parser_consumeStatementEnd(state ParserState) ParserState {
	return func() ParserState {
		if selfhost_parser_parser_parserMatch(state, TokenKind_Newline).ok {
			return selfhost_parser_parser_parserSkipNewlines(selfhost_parser_parser_parserAdvance(state).state)
		}
		return state
	}()
}

func selfhost_parser_parser_consumeFieldSeparator(state ParserState, close TokenKind, message string) ParserState {
	current := selfhost_parser_parser_consumeStatementEnd(state)
	comma := selfhost_parser_parser_parserMatch(current, TokenKind_Comma)
	return func() ParserState {
		if selfhost_parser_parser_parserCheck(current, close) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return current
		}
		return func() ParserState {
			if comma.ok {
				return selfhost_parser_parser_parserSkipNewlines(comma.state)
			}
			return selfhost_parser_parser_parserErrorAt(current, selfhost_parser_parser_parserPeek(current), message)
		}()
	}()
}

func selfhost_parser_parser_unquote(raw string) string {
	return func() string {
		if len([]rune(raw)) >= 2 {
			return func() string { runes := []rune(raw); return string(runes[1 : len([]rune(raw))-1]) }()
		}
		return raw
	}()
}

func selfhost_parser_parser_parseFileLoop(state ParserState, file ParsedFile) FileStep {
	return func() FileStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_EOF) {
			return FileStep{state: state, file: selfhost_parser_parser_withFileErrors(file, state.errors)}
		}
		return selfhost_parser_parser_parseTopLevel(state, file)
	}()
}

func selfhost_parser_parser_withFileErrors(file ParsedFile, errors []ParseError) ParsedFile {
	return ParsedFile{imports: file.imports, constants: file.constants, types: file.types, functions: file.functions, tests: file.tests, errors: errors}
}

func selfhost_parser_parser_parseTopLevel(state ParserState, file ParsedFile) FileStep {
	return func() FileStep {
		if selfhost_parser_parser_looksLikeMacroFunctionDecl(state) {
			return selfhost_parser_parser_parseTopLevelAfterResult(selfhost_parser_parser_parseMacroFunction(state, file))
		}
		return selfhost_parser_parser_parseTopLevelAfterMacro(state, file)
	}()
}

func selfhost_parser_parser_parseTopLevelAfterResult(result FileStep) FileStep {
	return selfhost_parser_parser_parseFileLoop(selfhost_parser_parser_parserSkipNewlines(result.state), result.file)
}

func selfhost_parser_parser_parseTopLevelAfterMacro(state ParserState, file ParsedFile) FileStep {
	return func() FileStep {
		if selfhost_parser_parser_looksLikeGoImportDecl(state) {
			return selfhost_parser_parser_parseTopLevelAfterResult(selfhost_parser_parser_parseTopLevelImport(state, file))
		}
		return func() FileStep {
			if selfhost_parser_parser_looksLikeRuneImportDecl(state) {
				return selfhost_parser_parser_parseTopLevelAfterResult(selfhost_parser_parser_parseTopLevelImport(state, file))
			}
			return selfhost_parser_parser_parseTopLevelAfterAnnotations(state, file)
		}()
	}()
}

func selfhost_parser_parser_looksLikeRuneImportDecl(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_At) && (selfhost_parser_parser_parserCheckNext(state, TokenKind_String) || selfhost_parser_parser_looksLikeBareModuleImportDecl(state))
}

func selfhost_parser_parser_parseTopLevelAfterAnnotations(state ParserState, file ParsedFile) FileStep {
	annotationStep := selfhost_parser_parser_parseAnnotations(state)
	publicStep := selfhost_parser_parser_parsePublicModifier(annotationStep.state)
	current := publicStep.state
	private := publicStep.ok == false
	current = func() ParserState {
		if publicStep.ok && (selfhost_parser_parser_parserCheck(current, TokenKind_At) || selfhost_parser_parser_parserCheck(current, TokenKind_Question)) {
			return selfhost_parser_parser_parserErrorAt(current, selfhost_parser_parser_parserPeek(current), "expected public declaration after '+'")
		}
		return current
	}()
	result := func() FileStep {
		if !(publicStep.ok) && selfhost_parser_parser_parserCheck(current, TokenKind_At) {
			return selfhost_parser_parser_parseTopLevelImport(current, file)
		}
		return func() FileStep {
			if !(publicStep.ok) && selfhost_parser_parser_parserCheck(current, TokenKind_Question) {
				return selfhost_parser_parser_parseTopLevelTest(current, file)
			}
			return func() FileStep {
				if selfhost_parser_parser_looksLikeConstDecl(current) {
					return selfhost_parser_parser_parseTopLevelConst(current, file, private, annotationStep.annotations)
				}
				return func() FileStep {
					if selfhost_parser_parser_looksLikeTypeDecl(current) {
						return selfhost_parser_parser_parseTopLevelType(current, file, private, annotationStep.annotations)
					}
					return func() FileStep {
						if selfhost_parser_parser_looksLikeFunctionDecl(current) {
							return selfhost_parser_parser_parseTopLevelFunction(current, file, private, annotationStep.annotations)
						}
						return selfhost_parser_parser_parseTopLevelError(current, file)
					}()
				}()
			}()
		}()
	}()
	return selfhost_parser_parser_parseFileLoop(selfhost_parser_parser_parserSkipNewlines(result.state), result.file)
}

func selfhost_parser_parser_looksLikeGoImportDecl(state ParserState) bool {
	marker := selfhost_parser_parser_parserCheck(state, TokenKind_At)
	return func() bool {
		if marker {
			return selfhost_parser_parser_looksLikeGoImportAfterMarker(state)
		}
		return false
	}()
}

func selfhost_parser_parser_looksLikeGoImportAfterMarker(state ParserState) bool {
	module := selfhost_parser_parser_parserKindAt(state, state.current+1) == TokenKind_Ident && selfhost_parser_parser_parserTokenAt(state, state.current+1).lexeme == "go"
	return func() bool {
		if module {
			return selfhost_parser_parser_looksLikeGoImportAfterModule(state)
		}
		return false
	}()
}

func selfhost_parser_parser_looksLikeGoImportAfterModule(state ParserState) bool {
	dot := selfhost_parser_parser_parserKindAt(state, state.current+2) == TokenKind_Dot
	return func() bool {
		if dot {
			return selfhost_parser_parser_parserKindAt(state, state.current+3) == TokenKind_Ident && selfhost_parser_parser_parserTokenAt(state, state.current+3).lexeme == "import"
		}
		return false
	}()
}

func selfhost_parser_parser_parseMacroFunction(state ParserState, file ParsedFile) FileStep {
	marker := selfhost_parser_parser_parserConsume(state, TokenKind_Hash, "expected '#' before macro function")
	step := selfhost_parser_parser_parseFunctionWithReceiver(marker.state, "", false, false, true, selfhost_parser_parser_emptyAnnotations())
	file.functions = append(file.functions, step.function)
	return FileStep{state: step.state, file: file}
}

func selfhost_parser_parser_parseTopLevelImport(state ParserState, file ParsedFile) FileStep {
	step := func() ImportStep {
		if selfhost_parser_parser_parserCheckNext(state, TokenKind_String) {
			return selfhost_parser_parser_parseImportDecl(state)
		}
		return func() ImportStep {
			if selfhost_parser_parser_looksLikeBareModuleImportDecl(state) {
				return selfhost_parser_parser_parseModuleImportDecl(state)
			}
			return selfhost_parser_parser_parseGoImportDecl(state)
		}()
	}()
	file.imports = append(file.imports, step.importDecl)
	return FileStep{state: step.state, file: file}
}

func selfhost_parser_parser_looksLikeBareModuleImportDecl(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_At) && selfhost_parser_parser_parserCheckNext(state, TokenKind_Ident) && selfhost_parser_parser_parserKindAt(state, state.current+2) != TokenKind_Dot
}

func selfhost_parser_parser_parseTopLevelTest(state ParserState, file ParsedFile) FileStep {
	step := selfhost_parser_parser_parseTestDecl(state)
	file.tests = append(file.tests, step.testDecl)
	return FileStep{state: step.state, file: file}
}

func selfhost_parser_parser_parseTopLevelConst(state ParserState, file ParsedFile, private bool, annotations []ParsedAnnotation) FileStep {
	current := state
	current = selfhost_parser_parser_parserRejectConstAnnotations(current, annotations)
	step := selfhost_parser_parser_parseConstDecl(current, private)
	file.constants = append(file.constants, step.constDecl)
	return FileStep{state: step.state, file: file}
}

func selfhost_parser_parser_parserRejectConstAnnotations(state ParserState, annotations []ParsedAnnotation) ParserState {
	return func() ParserState {
		switch {
		case len(annotations) == 0:
			return state
		default:
			return selfhost_parser_parser_parserErrorAt(state, selfhost_parser_parser_parserPeek(state), "annotations cannot be applied to constants")
		}
	}()
}

func selfhost_parser_parser_looksLikeConstDecl(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_Ident) && selfhost_parser_parser_parserCheckNext(state, TokenKind_Declare)
}

func selfhost_parser_parser_parseConstDecl(state ParserState, private bool) ConstStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected constant name")
	declare := selfhost_parser_parser_parserConsume(selfhost_parser_parser_parserSkipNewlines(name.state), TokenKind_Declare, "expected ':=' after constant name")
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(declare.state), 1)
	return ConstStep{state: value.state, constDecl: ParsedConst{name: name.token.lexeme, private: private, typeRef: emptyParsedTypeRef(), value: value.expr, line: name.token.line, column: name.token.column}}
}

func selfhost_parser_parser_parseTopLevelType(state ParserState, file ParsedFile, private bool, annotations []ParsedAnnotation) FileStep {
	step := selfhost_parser_parser_parseTypeDecl(state, private, annotations)
	file.types = append(file.types, step.typeDecl)
	return FileStep{state: step.state, file: file}
}

func selfhost_parser_parser_parseTopLevelFunction(state ParserState, file ParsedFile, private bool, annotations []ParsedAnnotation) FileStep {
	step := selfhost_parser_parser_parseFunctionWithReceiver(state, "", private, false, false, annotations)
	file.functions = append(file.functions, step.function)
	return FileStep{state: step.state, file: file}
}

func selfhost_parser_parser_parseTopLevelError(state ParserState, file ParsedFile) FileStep {
	return FileStep{state: selfhost_parser_parser_parserAdvance(selfhost_parser_parser_parserErrorAt(state, selfhost_parser_parser_parserPeek(state), "expected declaration")).state, file: file}
}

func selfhost_parser_parser_parsePublicModifier(state ParserState) BoolStep {
	step := selfhost_parser_parser_parserMatch(state, TokenKind_Plus)
	return BoolStep{ok: step.ok, state: func() ParserState {
		if step.ok {
			return selfhost_parser_parser_parserSkipNewlines(step.state)
		}
		return state
	}()}
}

func selfhost_parser_parser_parseObjectPrivateModifier(state ParserState) BoolStep {
	step := selfhost_parser_parser_parserMatch(state, TokenKind_Minus)
	return BoolStep{ok: step.ok, state: func() ParserState {
		if step.ok {
			return selfhost_parser_parser_parserSkipNewlines(step.state)
		}
		return state
	}()}
}

func selfhost_parser_parser_parseStaticMethodMarker(state ParserState) BoolStep {
	marker := func() BoolStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_DoubleColon) {
			return selfhost_parser_parser_parserMatch(state, TokenKind_DoubleColon)
		}
		return func() BoolStep {
			if selfhost_parser_parser_parserCheck(state, TokenKind_Ident) && selfhost_parser_parser_parserPeek(state).lexeme == "static" {
				return BoolStep{state: selfhost_parser_parser_parserAdvance(state).state, ok: true}
			}
			return BoolStep{state: state, ok: false}
		}()
	}()
	return BoolStep{ok: marker.ok, state: func() ParserState {
		if marker.ok {
			return selfhost_parser_parser_parserSkipNewlines(marker.state)
		}
		return state
	}()}
}

func selfhost_parser_parser_parseImportDecl(state ParserState) ImportStep {
	at := selfhost_parser_parser_parserConsume(state, TokenKind_At, "expected '@'")
	path := selfhost_parser_parser_parserConsume(at.state, TokenKind_String, "expected import path string after '@'")
	rawPath := selfhost_parser_parser_unquote(path.token.lexeme)
	goPath := selfhost_parser_parser_parserGoPackageImportPath(rawPath)
	isGo := goPath != ""
	invalidGo := strings.HasPrefix(rawPath, "go:") && goPath == ""
	nextState := func() ParserState {
		if invalidGo {
			return selfhost_parser_parser_parserErrorAt(path.state, path.token, "expected Go import path after \"go:\"")
		}
		return path.state
	}()
	return ImportStep{state: nextState, importDecl: ParsedImport{path: func() string {
		if isGo {
			return goPath
		}
		return rawPath
	}(), go_: isGo, module: false, line: at.token.line, column: at.token.column}}
}

func selfhost_parser_parser_parserGoPackageImportPath(spec string) string {
	return func() string {
		if strings.HasPrefix(spec, "go:") {
			return func() string { runes := []rune(spec); return string(runes[3:len([]rune(spec))]) }()
		}
		return ""
	}()
}

func selfhost_parser_parser_parseModuleImportDecl(state ParserState) ImportStep {
	at := selfhost_parser_parser_parserConsume(state, TokenKind_At, "expected '@'")
	module := selfhost_parser_parser_parserConsume(at.state, TokenKind_Ident, "expected module name after '@'")
	return ImportStep{state: module.state, importDecl: ParsedImport{path: module.token.lexeme, go_: false, module: true, line: at.token.line, column: at.token.column}}
}

func selfhost_parser_parser_parseGoImportDecl(state ParserState) ImportStep {
	at := selfhost_parser_parser_parserConsume(state, TokenKind_At, "expected '@'")
	module := selfhost_parser_parser_parserConsume(at.state, TokenKind_Ident, "expected module name after '@'")
	checked := func() ParserState {
		if module.token.lexeme == "go" {
			return module.state
		}
		return selfhost_parser_parser_parserErrorAt(module.state, module.token, "only @go.import can appear at the top level")
	}()
	dot := selfhost_parser_parser_parserConsume(checked, TokenKind_Dot, "expected '.' after @go")
	name := selfhost_parser_parser_parserConsume(dot.state, TokenKind_Ident, "expected import after @go.")
	checkedName := func() ParserState {
		if name.token.lexeme == "import" {
			return name.state
		}
		return selfhost_parser_parser_parserErrorAt(name.state, name.token, "only @go.import can appear at the top level")
	}()
	open := selfhost_parser_parser_parserConsume(checkedName, TokenKind_LParen, "expected '(' after @go.import")
	path := selfhost_parser_parser_parserConsume(open.state, TokenKind_String, "expected Go import path string")
	close := selfhost_parser_parser_parserConsume(path.state, TokenKind_RParen, "expected ')' after @go.import")
	return ImportStep{state: close.state, importDecl: ParsedImport{path: selfhost_parser_parser_unquote(path.token.lexeme), go_: true, module: false, line: at.token.line, column: at.token.column}}
}

func selfhost_parser_parser_parseTestDecl(state ParserState) TestStep {
	start := selfhost_parser_parser_parserConsume(state, TokenKind_Question, "expected '?'")
	name := selfhost_parser_parser_parserConsume(start.state, TokenKind_String, "expected test name string after '?'")
	bodyStart := selfhost_parser_parser_parserSkipNewlines(name.state)
	body := func() ExprStep {
		if selfhost_parser_parser_parserCheck(bodyStart, TokenKind_LBrace) {
			return selfhost_parser_parser_parseBlock(bodyStart)
		}
		return ExprStep{state: selfhost_parser_parser_parserErrorAt(bodyStart, selfhost_parser_parser_parserPeek(bodyStart), "expected test body block"), expr: selfhost_parser_parser_emptyExpr()}
	}()
	return TestStep{state: body.state, testDecl: ParsedTest{name: selfhost_parser_parser_unquote(name.token.lexeme), body: body.expr, line: start.token.line, column: start.token.column}}
}

func selfhost_parser_parser_parseTypeDecl(state ParserState, private bool, annotations []ParsedAnnotation) TypeStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected type name")
	generics := selfhost_parser_parser_parseGenericNames(name.state)
	colon := selfhost_parser_parser_parserConsume(generics.state, TokenKind_Colon, "expected ':' after type name")
	openStart := selfhost_parser_parser_parserSkipNewlines(colon.state)
	open := selfhost_parser_parser_parserConsume(openStart, TokenKind_LBrace, "expected '{' after type declaration")
	bodyStart := selfhost_parser_parser_parserSkipNewlines(open.state)
	return func() TypeStep {
		if selfhost_parser_parser_looksLikeEnumMember(bodyStart) {
			return selfhost_parser_parser_parseEnumTypeBody(bodyStart, name.token, private, annotations, generics.values)
		}
		return selfhost_parser_parser_parseStructTypeBody(bodyStart, name.token, private, annotations, generics.values)
	}()
}

func selfhost_parser_parser_parseStructTypeBody(state ParserState, name Token, private bool, annotations []ParsedAnnotation, generics []string) TypeStep {
	return selfhost_parser_parser_parseStructTypeLoop(state, ParsedType{name: name.lexeme, private: private, enum: false, annotations: annotations, generics: generics, fields: []ParsedField{}, methods: []ParsedFunction{}, members: []ParsedEnumMember{}, line: name.line, column: name.column})
}

func selfhost_parser_parser_parseStructTypeLoop(state ParserState, typeDecl ParsedType) TypeStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() TypeStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return selfhost_parser_parser_finishType(current, typeDecl, "expected '}' after type declaration")
		}
		return selfhost_parser_parser_parseStructTypeMember(current, typeDecl)
	}()
}

func selfhost_parser_parser_parseStructTypeMember(state ParserState, typeDecl ParsedType) TypeStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	macroMethod := selfhost_parser_parser_looksLikeMacroFunctionDecl(current)
	return func() TypeStep {
		if macroMethod {
			return selfhost_parser_parser_parseStructMacroMethod(current, typeDecl)
		}
		return selfhost_parser_parser_parseStructTypeMemberValue(current, typeDecl)
	}()
}

func selfhost_parser_parser_parseStructMacroMethod(state ParserState, typeDecl ParsedType) TypeStep {
	marker := selfhost_parser_parser_parserConsume(state, TokenKind_Hash, "expected '#' before macro method")
	step := selfhost_parser_parser_parseFunctionWithReceiver(selfhost_parser_parser_parserSkipNewlines(marker.state), typeDecl.name, true, false, true, selfhost_parser_parser_emptyAnnotations())
	typeDecl.methods = append(typeDecl.methods, step.function)
	next := selfhost_parser_parser_parserSkipNewlines(selfhost_parser_parser_parserMatch(selfhost_parser_parser_consumeStatementEnd(step.state), TokenKind_Comma).state)
	return selfhost_parser_parser_parseStructTypeLoop(next, typeDecl)
}

func selfhost_parser_parser_parseStructTypeMemberValue(state ParserState, typeDecl ParsedType) TypeStep {
	annotationStep := selfhost_parser_parser_parseAnnotations(state)
	current := annotationStep.state
	privateStep := selfhost_parser_parser_parseObjectPrivateModifier(current)
	memberState := privateStep.state
	private := privateStep.ok
	staticStep := func() BoolStep {
		if selfhost_parser_parser_looksLikeStaticFunctionDecl(memberState) {
			return selfhost_parser_parser_parseStaticMethodMarker(memberState)
		}
		return BoolStep{state: memberState, ok: false}
	}()
	parsed := func() TypeStep {
		if selfhost_parser_parser_looksLikeFunctionDecl(staticStep.state) {
			return selfhost_parser_parser_parseStructMethod(staticStep.state, typeDecl, private, staticStep.ok, annotationStep.annotations)
		}
		return selfhost_parser_parser_parseStructField(staticStep.state, typeDecl, private, annotationStep.annotations)
	}()
	next := selfhost_parser_parser_parserSkipNewlines(selfhost_parser_parser_parserMatch(selfhost_parser_parser_consumeStatementEnd(parsed.state), TokenKind_Comma).state)
	return selfhost_parser_parser_parseStructTypeLoop(next, parsed.typeDecl)
}

func selfhost_parser_parser_parseStructMethod(state ParserState, typeDecl ParsedType, private bool, static bool, annotations []ParsedAnnotation) TypeStep {
	step := selfhost_parser_parser_parseFunctionWithReceiver(state, typeDecl.name, private, static, false, annotations)
	typeDecl.methods = append(typeDecl.methods, step.function)
	return TypeStep{state: step.state, typeDecl: typeDecl}
}

func selfhost_parser_parser_parseStructField(state ParserState, typeDecl ParsedType, private bool, annotations []ParsedAnnotation) TypeStep {
	field := selfhost_parser_parser_parseFieldDecl(state, private, annotations)
	typeDecl.fields = append(typeDecl.fields, field.field)
	return TypeStep{state: field.state, typeDecl: typeDecl}
}

func selfhost_parser_parser_parseFieldDecl(state ParserState, private bool, annotations []ParsedAnnotation) FieldStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected field name")
	colon := selfhost_parser_parser_parserConsume(name.state, TokenKind_Colon, "expected ':' after field name")
	typeRef := selfhost_parser_parser_parseTypeRef(colon.state)
	return FieldStep{state: typeRef.state, field: ParsedField{name: name.token.lexeme, private: private, annotations: annotations, typeRef: typeRef.typeRef, line: name.token.line, column: name.token.column}}
}

func selfhost_parser_parser_finishType(state ParserState, typeDecl ParsedType, message string) TypeStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RBrace, message)
	return TypeStep{state: close.state, typeDecl: typeDecl}
}

func selfhost_parser_parser_parseEnumTypeBody(state ParserState, name Token, private bool, annotations []ParsedAnnotation, generics []string) TypeStep {
	return selfhost_parser_parser_parseEnumTypeLoop(state, ParsedType{name: name.lexeme, private: private, enum: true, annotations: annotations, generics: generics, fields: []ParsedField{}, methods: []ParsedFunction{}, members: []ParsedEnumMember{}, line: name.line, column: name.column})
}

func selfhost_parser_parser_parseEnumTypeLoop(state ParserState, typeDecl ParsedType) TypeStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() TypeStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return selfhost_parser_parser_finishType(current, typeDecl, "expected '}' after enum declaration")
		}
		return selfhost_parser_parser_parseEnumTypeMember(current, typeDecl)
	}()
}

func selfhost_parser_parser_parseEnumTypeMember(state ParserState, typeDecl ParsedType) TypeStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	macroMethod := selfhost_parser_parser_looksLikeMacroFunctionDecl(current)
	return func() TypeStep {
		if macroMethod {
			return selfhost_parser_parser_parseEnumMacroMethod(current, typeDecl)
		}
		return selfhost_parser_parser_parseEnumTypeMemberValueOrMethod(current, typeDecl)
	}()
}

func selfhost_parser_parser_parseEnumMacroMethod(state ParserState, typeDecl ParsedType) TypeStep {
	marker := selfhost_parser_parser_parserConsume(state, TokenKind_Hash, "expected '#' before macro method")
	step := selfhost_parser_parser_parseFunctionWithReceiver(selfhost_parser_parser_parserSkipNewlines(marker.state), typeDecl.name, true, false, true, selfhost_parser_parser_emptyAnnotations())
	typeDecl.methods = append(typeDecl.methods, step.function)
	next := selfhost_parser_parser_parserSkipNewlines(selfhost_parser_parser_parserMatch(selfhost_parser_parser_consumeStatementEnd(step.state), TokenKind_Comma).state)
	return selfhost_parser_parser_parseEnumTypeLoop(next, typeDecl)
}

func selfhost_parser_parser_parseEnumTypeMemberValueOrMethod(state ParserState, typeDecl ParsedType) TypeStep {
	annotationStep := selfhost_parser_parser_parseAnnotations(state)
	current := annotationStep.state
	privateStep := selfhost_parser_parser_parseObjectPrivateModifier(current)
	memberState := privateStep.state
	staticStep := func() BoolStep {
		if selfhost_parser_parser_looksLikeStaticFunctionDecl(memberState) {
			return selfhost_parser_parser_parseStaticMethodMarker(memberState)
		}
		return BoolStep{state: memberState, ok: false}
	}()
	parsed := func() TypeStep {
		if selfhost_parser_parser_looksLikeFunctionDecl(staticStep.state) {
			return selfhost_parser_parser_parseEnumMethod(staticStep.state, typeDecl, privateStep.ok, staticStep.ok, annotationStep.annotations)
		}
		return selfhost_parser_parser_parseEnumTypeMemberValue(current, typeDecl, annotationStep.annotations)
	}()
	next := selfhost_parser_parser_parserSkipNewlines(selfhost_parser_parser_parserMatch(selfhost_parser_parser_consumeStatementEnd(parsed.state), TokenKind_Comma).state)
	return selfhost_parser_parser_parseEnumTypeLoop(next, parsed.typeDecl)
}

func selfhost_parser_parser_parseEnumMethod(state ParserState, typeDecl ParsedType, private bool, static bool, annotations []ParsedAnnotation) TypeStep {
	step := selfhost_parser_parser_parseFunctionWithReceiver(state, typeDecl.name, private, static, false, annotations)
	typeDecl.methods = append(typeDecl.methods, step.function)
	return TypeStep{state: step.state, typeDecl: typeDecl}
}

func selfhost_parser_parser_parseEnumTypeMemberValue(state ParserState, typeDecl ParsedType, annotations []ParsedAnnotation) TypeStep {
	member := selfhost_parser_parser_parseEnumMember(state, annotations)
	typeDecl.members = append(typeDecl.members, member.member)
	return TypeStep{state: member.state, typeDecl: typeDecl}
}

func selfhost_parser_parser_parseEnumMember(state ParserState, annotations []ParsedAnnotation) EnumMemberStep {
	publicStep := selfhost_parser_parser_parsePublicModifier(state)
	name := selfhost_parser_parser_parserConsume(publicStep.state, TokenKind_Ident, "expected enum member name")
	current := selfhost_parser_parser_parserSkipNewlines(name.state)
	parsed := selfhost_parser_parser_parseEnumMemberPayload(current)
	return EnumMemberStep{state: parsed.state, member: ParsedEnumMember{name: name.token.lexeme, private: false, annotations: annotations, value: parsed.value, params: parsed.params, line: name.token.line, column: name.token.column}}
}

func selfhost_parser_parser_parseEnumMemberPayload(state ParserState) EnumMemberPayloadStep {
	assign := selfhost_parser_parser_parserMatch(state, TokenKind_Assign)
	return func() EnumMemberPayloadStep {
		if assign.ok {
			return selfhost_parser_parser_parseEnumMemberValue(selfhost_parser_parser_parserSkipNewlines(assign.state))
		}
		return selfhost_parser_parser_parseEnumMemberParams(state)
	}()
}

func selfhost_parser_parser_parseEnumMemberValue(state ParserState) EnumMemberPayloadStep {
	value := selfhost_parser_parser_parseEnumValue(state)
	return EnumMemberPayloadStep{state: value.state, value: value.value, params: []ParsedParam{}}
}

func selfhost_parser_parser_parseEnumMemberParams(state ParserState) EnumMemberPayloadStep {
	open := selfhost_parser_parser_parserMatch(state, TokenKind_LParen)
	return func() EnumMemberPayloadStep {
		if open.ok {
			return selfhost_parser_parser_parseEnumMemberParamList(selfhost_parser_parser_parserSkipNewlines(open.state))
		}
		return EnumMemberPayloadStep{state: state, value: "", params: []ParsedParam{}}
	}()
}

func selfhost_parser_parser_parseEnumMemberParamList(state ParserState) EnumMemberPayloadStep {
	params := selfhost_parser_parser_parseParamList(state)
	close := selfhost_parser_parser_parserConsume(params.state, TokenKind_RParen, "expected ')' after enum constructor parameters")
	return EnumMemberPayloadStep{state: close.state, value: "", params: params.params}
}

func selfhost_parser_parser_parseEnumValue(state ParserState) StringStep {
	return func() StringStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_Minus) {
			return selfhost_parser_parser_parseNegativeEnumValue(state)
		}
		return selfhost_parser_parser_parsePositiveEnumValue(state)
	}()
}

func selfhost_parser_parser_parseNegativeEnumValue(state ParserState) StringStep {
	minus := selfhost_parser_parser_parserAdvance(state)
	value := selfhost_parser_parser_parserConsume(minus.state, TokenKind_Int, "expected integer enum value")
	return StringStep{state: value.state, value: "-" + value.token.lexeme}
}

func selfhost_parser_parser_parsePositiveEnumValue(state ParserState) StringStep {
	value := selfhost_parser_parser_parserConsume(state, TokenKind_Int, "expected integer enum value")
	return StringStep{state: value.state, value: value.token.lexeme}
}

func selfhost_parser_parser_parseFunctionWithReceiver(state ParserState, receiverType string, private bool, static bool, macro bool, annotations []ParsedAnnotation) FunctionStep {
	routineStep := selfhost_parser_parser_parserMatch(state, TokenKind_Tilde)
	afterRoutine := func() ParserState {
		if routineStep.ok {
			return selfhost_parser_parser_parserSkipNewlines(routineStep.state)
		}
		return state
	}()
	name := selfhost_parser_parser_parserConsume(afterRoutine, TokenKind_Ident, "expected function name")
	generics := selfhost_parser_parser_parseGenericNames(name.state)
	open := selfhost_parser_parser_parserConsume(generics.state, TokenKind_LParen, "expected '(' after function name")
	params := selfhost_parser_parser_parseParamList(selfhost_parser_parser_parserSkipNewlines(open.state))
	close := selfhost_parser_parser_parserConsume(params.state, TokenKind_RParen, "expected ')' after parameter list")
	afterClose := selfhost_parser_parser_parserSkipNewlines(close.state)
	ret := selfhost_parser_parser_parserMatch(afterClose, TokenKind_Arrow)
	returnType := func() TypeRefStep {
		if ret.ok {
			return selfhost_parser_parser_parseTypeRef(selfhost_parser_parser_parserSkipNewlines(ret.state))
		}
		return TypeRefStep{state: afterClose, typeRef: emptyParsedTypeRef()}
	}()
	arrow := selfhost_parser_parser_parserConsume(selfhost_parser_parser_parserSkipNewlines(returnType.state), TokenKind_FatArrow, "expected '=>' after function signature")
	body := selfhost_parser_parser_parseBody(selfhost_parser_parser_parserSkipNewlines(arrow.state))
	return FunctionStep{state: body.state, function: ParsedFunction{name: name.token.lexeme, private: private, static: static, routine: routineStep.ok, macro: macro, annotations: annotations, receiverType: receiverType, generics: generics.values, params: params.params, returnType: returnType.typeRef, body: body.expr, line: name.token.line, column: name.token.column}}
}

func selfhost_parser_parser_parseParamList(state ParserState) ParamListStep {
	return selfhost_parser_parser_parseParamListLoop(state, []ParsedParam{})
}

func selfhost_parser_parser_parseParamListLoop(state ParserState, params []ParsedParam) ParamListStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ParamListStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RParen) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return ParamListStep{state: current, params: params}
		}
		return selfhost_parser_parser_parseOneParam(current, params)
	}()
}

func selfhost_parser_parser_parseOneParam(state ParserState, params []ParsedParam) ParamListStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected parameter name")
	optional := selfhost_parser_parser_parserMatch(name.state, TokenKind_Question)
	colon := selfhost_parser_parser_parserMatch(optional.state, TokenKind_Colon)
	typeRef := func() TypeRefStep {
		if colon.ok {
			return selfhost_parser_parser_parseTypeRef(colon.state)
		}
		return TypeRefStep{state: optional.state, typeRef: emptyParsedTypeRef()}
	}()
	paramName := func() string {
		if optional.ok {
			return name.token.lexeme + "?"
		}
		return name.token.lexeme
	}()
	params = append(params, ParsedParam{name: paramName, typeRef: typeRef.typeRef, line: name.token.line, column: name.token.column})
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(typeRef.state), TokenKind_Comma)
	return selfhost_parser_parser_parseParamListLoop(comma.state, params)
}

func selfhost_parser_parser_parseGenericNames(state ParserState) StringListStep {
	open := selfhost_parser_parser_parserMatch(state, TokenKind_LBracket)
	return func() StringListStep {
		if open.ok {
			return selfhost_parser_parser_parseGenericNameLoop(selfhost_parser_parser_parserSkipNewlines(open.state), []string{})
		}
		return StringListStep{state: state, values: []string{}}
	}()
}

func selfhost_parser_parser_parseGenericNameLoop(state ParserState, values []string) StringListStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() StringListStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBracket) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return selfhost_parser_parser_finishGenericNames(current, values)
		}
		return selfhost_parser_parser_parseGenericNameValue(current, values)
	}()
}

func selfhost_parser_parser_finishGenericNames(state ParserState, values []string) StringListStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RBracket, "expected ']' after generic parameters")
	return StringListStep{state: close.state, values: values}
}

func selfhost_parser_parser_parseGenericNameValue(state ParserState, values []string) StringListStep {
	nameStep := func() StringListStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_Ident) {
			return selfhost_parser_parser_appendGenericName(state, values)
		}
		return StringListStep{state: selfhost_parser_parser_parserAdvance(state).state, values: values}
	}()
	colon := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(nameStep.state), TokenKind_Colon)
	current := func() ParserState {
		if colon.ok {
			return selfhost_parser_parser_parseTypeRef(selfhost_parser_parser_parserSkipNewlines(colon.state)).state
		}
		return nameStep.state
	}()
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(current), TokenKind_Comma)
	return selfhost_parser_parser_parseGenericNameLoop(comma.state, nameStep.values)
}

func selfhost_parser_parser_appendGenericName(state ParserState, values []string) StringListStep {
	step := selfhost_parser_parser_parserAdvance(state)
	return StringListStep{state: step.state, values: selfhost_parser_parser_appendString(values, step.token.lexeme)}
}

func selfhost_parser_parser_parseTypeRef(state ParserState) TypeRefStep {
	return selfhost_parser_parser_parseTypeRefPostfix(selfhost_parser_parser_parseTypeRefAtom(selfhost_parser_parser_parserSkipNewlines(state)))
}

func selfhost_parser_parser_parseTypeRefAtom(state ParserState) TypeRefStep {
	return func() TypeRefStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_LParen) {
			return selfhost_parser_parser_parseParenTypeRef(state)
		}
		return func() TypeRefStep {
			if selfhost_parser_parser_parserCheck(state, TokenKind_At) {
				return selfhost_parser_parser_parseQualifiedTypeRef(state)
			}
			return func() TypeRefStep {
				if selfhost_parser_parser_parserCheck(state, TokenKind_Ident) {
					return selfhost_parser_parser_parseNamedTypeRef(state)
				}
				return selfhost_parser_parser_parseTypeRefError(state)
			}()
		}()
	}()
}

func selfhost_parser_parser_parseNamedTypeRef(state ParserState) TypeRefStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected type name")
	return TypeRefStep{state: name.state, typeRef: namedParsedTypeRef(name.token)}
}

func selfhost_parser_parser_parseQualifiedTypeRef(state ParserState) TypeRefStep {
	at := selfhost_parser_parser_parserConsume(state, TokenKind_At, "expected '@'")
	module := selfhost_parser_parser_parserConsume(at.state, TokenKind_Ident, "expected module name after '@'")
	dot := selfhost_parser_parser_parserConsume(module.state, TokenKind_Dot, "expected '.' after module name")
	name := selfhost_parser_parser_parserConsume(dot.state, TokenKind_Ident, "expected type name after module qualifier")
	return TypeRefStep{state: name.state, typeRef: qualifiedParsedTypeRef(module.token, name.token)}
}

func selfhost_parser_parser_parseParenTypeRef(state ParserState) TypeRefStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LParen, "expected '(' before type")
	params := selfhost_parser_parser_parseTypeParamList(selfhost_parser_parser_parserSkipNewlines(open.state), []ParsedTypeParam{})
	close := selfhost_parser_parser_parserConsume(params.state, TokenKind_RParen, "expected ')' after type")
	afterClose := selfhost_parser_parser_parserSkipNewlines(close.state)
	arrow := selfhost_parser_parser_parserMatch(afterClose, TokenKind_Arrow)
	return func() TypeRefStep {
		if arrow.ok {
			return selfhost_parser_parser_finishFunctionTypeRef(selfhost_parser_parser_parserSkipNewlines(arrow.state), open.token, params.params)
		}
		return selfhost_parser_parser_finishParenTypeRef(close.state, open.token, params.params)
	}()
}

func selfhost_parser_parser_finishFunctionTypeRef(state ParserState, token Token, params []ParsedTypeParam) TypeRefStep {
	ret := selfhost_parser_parser_parseTypeRef(state)
	return TypeRefStep{state: ret.state, typeRef: functionTypeRef(params, ret.typeRef, token)}
}

func selfhost_parser_parser_finishParenTypeRef(state ParserState, token Token, params []ParsedTypeParam) TypeRefStep {
	return TypeRefStep{state: state, typeRef: func() ParsedTypeRef {
		if len(params) == 1 && params[0].name == "" {
			return groupedTypeRef(params[0].typeRef, token)
		}
		return tupleTypeRef(params, token)
	}()}
}

func selfhost_parser_parser_parseTypeRefPostfix(step TypeRefStep) TypeRefStep {
	current := selfhost_parser_parser_parserSkipNewlines(step.state)
	return func() TypeRefStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_LBracket) {
			return selfhost_parser_parser_parseTypeRefPostfix(selfhost_parser_parser_parseTypeRefArgs(current, step.typeRef))
		}
		return func() TypeRefStep {
			if selfhost_parser_parser_parserCheck(current, TokenKind_Question) {
				return selfhost_parser_parser_parseTypeRefPostfix(selfhost_parser_parser_parseNullableTypeRef(current, step.typeRef))
			}
			return TypeRefStep{state: step.state, typeRef: step.typeRef}
		}()
	}()
}

func selfhost_parser_parser_parseTypeRefArgs(state ParserState, typeRef ParsedTypeRef) TypeRefStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBracket, "expected '[' after type name")
	args := selfhost_parser_parser_parseTypeRefList(selfhost_parser_parser_parserSkipNewlines(open.state), []ParsedTypeRef{})
	close := selfhost_parser_parser_parserConsume(args.state, TokenKind_RBracket, "expected ']' after type arguments")
	return TypeRefStep{state: close.state, typeRef: typeRefWithArgs(typeRef, args.refs)}
}

func selfhost_parser_parser_parseNullableTypeRef(state ParserState, typeRef ParsedTypeRef) TypeRefStep {
	question := selfhost_parser_parser_parserConsume(state, TokenKind_Question, "expected '?' after type")
	return TypeRefStep{state: question.state, typeRef: nullableTypeRef(typeRef)}
}

func selfhost_parser_parser_parseTypeRefList(state ParserState, refs []ParsedTypeRef) TypeRefListStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() TypeRefListStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBracket) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return TypeRefListStep{state: current, refs: refs}
		}
		return selfhost_parser_parser_parseOneTypeRefListValue(current, refs)
	}()
}

func selfhost_parser_parser_parseOneTypeRefListValue(state ParserState, refs []ParsedTypeRef) TypeRefListStep {
	typeRef := selfhost_parser_parser_parseTypeRef(state)
	refs = append(refs, typeRef.typeRef)
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(typeRef.state), TokenKind_Comma)
	return func() TypeRefListStep {
		if comma.ok {
			return selfhost_parser_parser_parseTypeRefList(comma.state, refs)
		}
		return TypeRefListStep{state: typeRef.state, refs: refs}
	}()
}

func selfhost_parser_parser_parseTypeParamList(state ParserState, params []ParsedTypeParam) TypeParamListStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() TypeParamListStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RParen) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return TypeParamListStep{state: current, params: params}
		}
		return selfhost_parser_parser_parseOneTypeParam(current, params)
	}()
}

func selfhost_parser_parser_parseOneTypeParam(state ParserState, params []ParsedTypeParam) TypeParamListStep {
	param := selfhost_parser_parser_parseTypeParam(state)
	params = append(params, param.param)
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(param.state), TokenKind_Comma)
	return func() TypeParamListStep {
		if comma.ok {
			return selfhost_parser_parser_parseTypeParamList(comma.state, params)
		}
		return TypeParamListStep{state: param.state, params: params}
	}()
}

func selfhost_parser_parser_parseTypeParam(state ParserState) TypeParamStep {
	named := selfhost_parser_parser_parserCheck(state, TokenKind_Ident) && selfhost_parser_parser_typeParamHasName(state)
	return func() TypeParamStep {
		if named {
			return selfhost_parser_parser_parseNamedTypeParam(state)
		}
		return selfhost_parser_parser_parseUnnamedTypeParam(state)
	}()
}

func selfhost_parser_parser_parseNamedTypeParam(state ParserState) TypeParamStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected type parameter name")
	optional := selfhost_parser_parser_parserMatch(name.state, TokenKind_Question)
	colon := selfhost_parser_parser_parserConsume(optional.state, TokenKind_Colon, "expected ':' after type parameter name")
	typeRef := selfhost_parser_parser_parseTypeRef(selfhost_parser_parser_parserSkipNewlines(colon.state))
	return TypeParamStep{state: typeRef.state, param: ParsedTypeParam{name: name.token.lexeme, optional: optional.ok, typeRef: typeRef.typeRef}}
}

func selfhost_parser_parser_parseUnnamedTypeParam(state ParserState) TypeParamStep {
	typeRef := selfhost_parser_parser_parseTypeRef(state)
	return TypeParamStep{state: typeRef.state, param: ParsedTypeParam{name: "", optional: false, typeRef: typeRef.typeRef}}
}

func selfhost_parser_parser_typeParamHasName(state ParserState) bool {
	return selfhost_parser_parser_parserKindAt(state, state.current+1) == TokenKind_Colon || selfhost_parser_parser_parserKindAt(state, state.current+1) == TokenKind_Question && selfhost_parser_parser_parserKindAt(state, state.current+2) == TokenKind_Colon
}

func selfhost_parser_parser_parseTypeRefError(state ParserState) TypeRefStep {
	return TypeRefStep{state: selfhost_parser_parser_parserErrorAt(state, selfhost_parser_parser_parserPeek(state), "expected type name"), typeRef: emptyParsedTypeRef()}
}

func selfhost_parser_parser_parseBody(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_LBrace) {
			return selfhost_parser_parser_parseBraceBody(state)
		}
		return selfhost_parser_parser_parseExpression(state, 1)
	}()
}

func selfhost_parser_parser_parseBraceBody(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_looksLikePatternBranch(state) == false && selfhost_parser_parser_looksLikeMapLiteralBody(state) {
			return selfhost_parser_parser_parseMapLiteral(state)
		}
		return func() ExprStep {
			if selfhost_parser_parser_looksLikePatternBranch(state) == false && selfhost_parser_parser_looksLikeObjectLiteralBody(state) {
				return selfhost_parser_parser_parseObjectLiteral(state)
			}
			return selfhost_parser_parser_parseBlock(state)
		}()
	}()
}

func selfhost_parser_parser_looksLikeObjectLiteralBody(state ParserState) bool {
	first := selfhost_parser_parser_skipNewlinesAt(state, state.current+1)
	return selfhost_parser_parser_parserKindAt(state, first) == TokenKind_DotDot || selfhost_parser_parser_parserKindAt(state, first) == TokenKind_Ident && selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, first+1)) == TokenKind_Colon
}

func selfhost_parser_parser_parseBlock(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBrace, "expected '{'")
	bodyStart := selfhost_parser_parser_parserSkipNewlines(open.state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(bodyStart, TokenKind_RBrace) {
			return selfhost_parser_parser_finishBlock(bodyStart, selfhost_parser_parser_node(ExprKind_Block, open.token))
		}
		return func() ExprStep {
			if selfhost_parser_parser_looksLikePatternBranch(bodyStart) {
				return selfhost_parser_parser_parsePatternBlock(bodyStart, selfhost_parser_parser_node(ExprKind_PatternBlock, open.token))
			}
			return selfhost_parser_parser_parseBlockLoop(bodyStart, selfhost_parser_parser_node(ExprKind_Block, open.token))
		}()
	}()
}

func selfhost_parser_parser_finishBlock(state ParserState, block ParsedExpr) ExprStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RBrace, "expected '}' after block")
	return ExprStep{state: close.state, expr: block}
}

func selfhost_parser_parser_parseBlockLoop(state ParserState, block ParsedExpr) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return selfhost_parser_parser_finishBlock(current, block)
		}
		return selfhost_parser_parser_parseBlockStatement(current, block)
	}()
}

func selfhost_parser_parser_parseBlockStatement(state ParserState, block ParsedExpr) ExprStep {
	stmt := selfhost_parser_parser_parseStatement(state)
	nextBlock := selfhost_parser_parser_appendChild(block, stmt.expr)
	return selfhost_parser_parser_parseBlockLoop(selfhost_parser_parser_consumeStatementEnd(stmt.state), nextBlock)
}

func selfhost_parser_parser_parseStatement(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_LBrace) && selfhost_parser_parser_looksLikeObjectDestructureDecl(state) {
			return selfhost_parser_parser_parseObjectDestructureStatement(state)
		}
		return func() ExprStep {
			if selfhost_parser_parser_parserCheck(state, TokenKind_LBrace) {
				return selfhost_parser_parser_parseBlock(state)
			}
			return func() ExprStep {
				if selfhost_parser_parser_parserCheck(state, TokenKind_Dollar) && selfhost_parser_parser_parserCheckNext(state, TokenKind_Ident) {
					return selfhost_parser_parser_parseDollarStatement(state)
				}
				return func() ExprStep {
					if selfhost_parser_parser_parserCheck(state, TokenKind_Ident) && (selfhost_parser_parser_parserCheckNext(state, TokenKind_Declare) || selfhost_parser_parser_parserCheckNext(state, TokenKind_MutDeclare)) {
						return selfhost_parser_parser_parseLetStatement(state)
					}
					return func() ExprStep {
						if selfhost_parser_parser_parserCheck(state, TokenKind_Ident) && selfhost_parser_parser_parserCheckNext(state, TokenKind_Assign) {
							return selfhost_parser_parser_parseAssignStatement(state)
						}
						return selfhost_parser_parser_parseExpression(state, 1)
					}()
				}()
			}()
		}()
	}()
}

func selfhost_parser_parser_parseDollarStatement(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserKindAt(state, state.current+2) == TokenKind_Declare {
			return selfhost_parser_parser_parseSignalPrefixLetStatement(state)
		}
		return selfhost_parser_parser_parseExpression(state, 1)
	}()
}

func selfhost_parser_parser_parseSignalPrefixLetStatement(state ParserState) ExprStep {
	dollar := selfhost_parser_parser_parserAdvance(state)
	name := selfhost_parser_parser_parserAdvance(dollar.state)
	op := selfhost_parser_parser_parserAdvance(name.state)
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(op.state), 1)
	typeName := selfhost_parser_parser_parseLetTypeAnnotation(value.state)
	return ExprStep{state: typeName.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Let, "$"+name.token.lexeme, name.token.lexeme, typeName.value, op.token.lexeme, []ParsedParam{}, []ParsedExpr{value.expr}, name.token.line, name.token.column)}
}

func selfhost_parser_parser_parseLetStatement(state ParserState) ExprStep {
	name := selfhost_parser_parser_parserAdvance(state)
	op := selfhost_parser_parser_parserAdvance(name.state)
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(op.state), 1)
	typeName := selfhost_parser_parser_parseLetTypeAnnotation(value.state)
	return ExprStep{state: typeName.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Let, name.token.lexeme, name.token.lexeme, typeName.value, op.token.lexeme, []ParsedParam{}, []ParsedExpr{value.expr}, name.token.line, name.token.column)}
}

func selfhost_parser_parser_parseLetTypeAnnotation(state ParserState) StringStep {
	colon := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(state), TokenKind_Colon)
	return func() StringStep {
		if colon.ok {
			return selfhost_parser_parser_parseLetTypeAnnotationRef(colon.state)
		}
		return StringStep{state: state, value: ""}
	}()
}

func selfhost_parser_parser_parseLetTypeAnnotationRef(state ParserState) StringStep {
	typeRef := selfhost_parser_parser_parseTypeRef(selfhost_parser_parser_parserSkipNewlines(state))
	return StringStep{state: typeRef.state, value: typeRefToString(typeRef.typeRef)}
}

func selfhost_parser_parser_parseObjectDestructureStatement(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBrace, "expected '{' before object destructuring")
	fields := selfhost_parser_parser_parseObjectBindingList(selfhost_parser_parser_parserSkipNewlines(open.state))
	close := selfhost_parser_parser_parserConsume(fields.state, TokenKind_RBrace, "expected '}' after object destructuring")
	op := selfhost_parser_parser_parserAdvance(selfhost_parser_parser_parserSkipNewlines(close.state))
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(op.state), 1)
	return ExprStep{state: value.state, expr: selfhost_parser_parser_makeExpr(ExprKind_ObjectDestructure, "", "", "", op.token.lexeme, fields.params, []ParsedExpr{value.expr}, open.token.line, open.token.column)}
}

func selfhost_parser_parser_parseObjectBindingList(state ParserState) ParamListStep {
	return selfhost_parser_parser_parseObjectBindingListLoop(state, []ParsedParam{})
}

func selfhost_parser_parser_parseObjectBindingListLoop(state ParserState, fields []ParsedParam) ParamListStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ParamListStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return ParamListStep{state: current, params: fields}
		}
		return selfhost_parser_parser_parseOneObjectBinding(current, fields)
	}()
}

func selfhost_parser_parser_parseOneObjectBinding(state ParserState, fields []ParsedParam) ParamListStep {
	field := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected field name in object destructuring")
	aliasStart := selfhost_parser_parser_parserMatch(field.state, TokenKind_Colon)
	name := func() TokenStep {
		if aliasStart.ok {
			return selfhost_parser_parser_parserConsume(selfhost_parser_parser_parserSkipNewlines(aliasStart.state), TokenKind_Ident, "expected binding name after ':'")
		}
		return field
	}()
	fields = append(fields, ParsedParam{name: name.token.lexeme, typeRef: namedParsedTypeRef(field.token), line: name.token.line, column: name.token.column})
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(name.state), TokenKind_Comma)
	return func() ParamListStep {
		if comma.ok {
			return selfhost_parser_parser_parseObjectBindingListLoop(comma.state, fields)
		}
		return ParamListStep{state: name.state, params: fields}
	}()
}

func selfhost_parser_parser_parseAssignStatement(state ParserState) ExprStep {
	name := selfhost_parser_parser_parserAdvance(state)
	op := selfhost_parser_parser_parserAdvance(name.state)
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(op.state), 1)
	return ExprStep{state: value.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Assign, name.token.lexeme, name.token.lexeme, "", op.token.lexeme, []ParsedParam{}, []ParsedExpr{value.expr}, name.token.line, name.token.column)}
}

func selfhost_parser_parser_parseExpression(state ParserState, minPrec int) ExprStep {
	return func() ExprStep {
		if minPrec <= 1 && selfhost_parser_parser_parserCheck(state, TokenKind_LParen) && selfhost_parser_parser_looksLikeLambda(state) {
			return selfhost_parser_parser_parseLambda(state)
		}
		return selfhost_parser_parser_parseExpressionLoop(selfhost_parser_parser_parseUnary(state), minPrec)
	}()
}

func selfhost_parser_parser_parseExpressionLoop(left ExprStep, minPrec int) ExprStep {
	state := left.state
	expr := left.expr
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_LBrace) {
			return selfhost_parser_parser_parseAfterBraceExpression(state, expr, minPrec)
		}
		return func() ExprStep {
			if selfhost_parser_parser_parserCheck(state, TokenKind_LParen) {
				return selfhost_parser_parser_parseCallExpression(state, expr, minPrec)
			}
			return func() ExprStep {
				if selfhost_parser_parser_parserCheck(state, TokenKind_LBracket) {
					return selfhost_parser_parser_parseIndexExpression(state, expr, minPrec)
				}
				return func() ExprStep {
					if selfhost_parser_parser_parserCheck(state, TokenKind_Dot) || selfhost_parser_parser_parserCheck(state, TokenKind_DoubleColon) {
						return selfhost_parser_parser_parseSelectorExpression(state, expr, minPrec)
					}
					return func() ExprStep {
						if selfhost_parser_parser_parserCheck(state, TokenKind_PlusPlus) {
							return selfhost_parser_parser_parsePostfixExpression(state, expr, minPrec)
						}
						return func() ExprStep {
							if selfhost_parser_parser_parserCheck(state, TokenKind_Apostrophe) {
								return selfhost_parser_parser_parseCompileTimePostfixExpression(state, expr, minPrec)
							}
							return func() ExprStep {
								if minPrec <= 1 && selfhost_parser_parser_parserCheck(state, TokenKind_Arrow) {
									return selfhost_parser_parser_parseWatchExpression(state, expr, minPrec)
								}
								return func() ExprStep {
									if minPrec <= 1 && selfhost_parser_parser_parserCheck(state, TokenKind_Assign) {
										return selfhost_parser_parser_parseAssignmentExpression(state, expr, minPrec)
									}
									return func() ExprStep {
										if minPrec <= 1 && selfhost_parser_parser_parserCheck(state, TokenKind_Tilde) {
											return selfhost_parser_parser_parsePatternPredicateExpression(state, expr, minPrec)
										}
										return func() ExprStep {
											if minPrec <= 1 && selfhost_parser_parser_parserCheck(state, TokenKind_Question) {
												return selfhost_parser_parser_parseQuestionExpression(state, expr, minPrec)
											}
											return func() ExprStep {
												if minPrec <= 1 && selfhost_parser_parser_parserCheck(state, TokenKind_QuestionQuestion) {
													return selfhost_parser_parser_parseQuestionQuestionExpression(state, expr, minPrec)
												}
												return selfhost_parser_parser_parseBinaryExpression(state, expr, minPrec)
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

func selfhost_parser_parser_parseAfterBraceExpression(state ParserState, expr ParsedExpr, minPrec int) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_looksLikePatternBlockAfterSubject(state) {
			return selfhost_parser_parser_parseExpressionLoop(selfhost_parser_parser_parseMatchExpression(state, expr), minPrec)
		}
		return func() ExprStep {
			if expr.kind == ExprKind_Identifier {
				return selfhost_parser_parser_parseExpressionLoop(selfhost_parser_parser_parseStructLiteral(state, expr), minPrec)
			}
			return ExprStep{state: state, expr: expr}
		}()
	}()
}

func selfhost_parser_parser_parseCallExpression(state ParserState, callee ParsedExpr, minPrec int) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LParen, "expected '(' after callee")
	args := selfhost_parser_parser_parseArgumentList(selfhost_parser_parser_parserSkipNewlines(open.state), []ParsedExpr{callee}, TokenKind_RParen)
	close := selfhost_parser_parser_parserConsume(args.state, TokenKind_RParen, "expected ')' after arguments")
	call := selfhost_parser_parser_makeExpr(ExprKind_Call, selfhost_parser_parser_calleeText(callee), "", "", "", []ParsedParam{}, args.expr.children, callee.line, callee.column)
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: close.state, expr: call}, minPrec)
}

func selfhost_parser_parser_parseArgumentList(state ParserState, holderChildren []ParsedExpr, endKind TokenKind) ExprStep {
	holder := selfhost_parser_parser_makeExpr(ExprKind_Args, "", "", "", "", []ParsedParam{}, holderChildren, 0, 0)
	return selfhost_parser_parser_parseArgumentListLoop(state, holder, endKind)
}

func selfhost_parser_parser_parseArgumentListLoop(state ParserState, holder ParsedExpr, endKind TokenKind) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(current, endKind) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return ExprStep{state: current, expr: holder}
		}
		return selfhost_parser_parser_parseOneArgument(current, holder, endKind)
	}()
}

func selfhost_parser_parser_parseOneArgument(state ParserState, holder ParsedExpr, endKind TokenKind) ExprStep {
	spread := selfhost_parser_parser_parserMatch(state, TokenKind_DotDot)
	value := selfhost_parser_parser_parseExpression(func() ParserState {
		if spread.ok {
			return spread.state
		}
		return state
	}(), 1)
	arg := func() ParsedExpr {
		if spread.ok {
			return selfhost_parser_parser_opNode(ExprKind_Spread, "..", selfhost_parser_parser_parserPrevious(spread.state), []ParsedExpr{value.expr})
		}
		return value.expr
	}()
	nextHolder := selfhost_parser_parser_appendChild(holder, arg)
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(value.state), TokenKind_Comma)
	return selfhost_parser_parser_parseArgumentListLoop(comma.state, nextHolder, endKind)
}

func selfhost_parser_parser_parseIndexExpression(state ParserState, receiver ParsedExpr, minPrec int) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBracket, "expected '[' after receiver")
	index := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(open.state), 1)
	close := selfhost_parser_parser_parserConsume(index.state, TokenKind_RBracket, "expected ']' after index")
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: close.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Index, receiver.text, "", "", "", []ParsedParam{}, []ParsedExpr{receiver, index.expr}, receiver.line, receiver.column)}, minPrec)
}

func selfhost_parser_parser_parseSelectorExpression(state ParserState, receiver ParsedExpr, minPrec int) ExprStep {
	operator := func() TokenStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_DoubleColon) {
			return selfhost_parser_parser_parserConsume(state, TokenKind_DoubleColon, "expected '::'")
		}
		return selfhost_parser_parser_parserConsume(state, TokenKind_Dot, "expected '.'")
	}()
	name := selfhost_parser_parser_parserConsume(operator.state, TokenKind_Ident, "expected selector name")
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: name.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Selector, name.token.lexeme, name.token.lexeme, "", operator.token.lexeme, []ParsedParam{}, []ParsedExpr{receiver}, operator.token.line, operator.token.column)}, minPrec)
}

func selfhost_parser_parser_parsePostfixExpression(state ParserState, expr ParsedExpr, minPrec int) ExprStep {
	op := selfhost_parser_parser_parserAdvance(state)
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: op.state, expr: selfhost_parser_parser_opNode(ExprKind_Postfix, op.token.lexeme, op.token, []ParsedExpr{expr})}, minPrec)
}

func selfhost_parser_parser_parseCompileTimePostfixExpression(state ParserState, expr ParsedExpr, minPrec int) ExprStep {
	op := selfhost_parser_parser_parserAdvance(state)
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: op.state, expr: selfhost_parser_parser_opNode(ExprKind_CompileTime, op.token.lexeme, op.token, []ParsedExpr{expr})}, minPrec)
}

func selfhost_parser_parser_parseWatchExpression(state ParserState, expr ParsedExpr, minPrec int) ExprStep {
	arrow := selfhost_parser_parser_parserAdvance(state)
	handler := selfhost_parser_parser_parseWatchHandler(selfhost_parser_parser_parserSkipNewlines(arrow.state))
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: handler.state, expr: selfhost_parser_parser_opNode(ExprKind_Watch, arrow.token.lexeme, arrow.token, []ParsedExpr{expr, handler.expr})}, minPrec)
}

func selfhost_parser_parser_parseWatchHandler(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_LParen) && selfhost_parser_parser_looksLikeLambda(state) {
			return selfhost_parser_parser_parseLambda(state)
		}
		return func() ExprStep {
			if selfhost_parser_parser_parserCheck(state, TokenKind_LBrace) {
				return selfhost_parser_parser_parseBody(state)
			}
			return selfhost_parser_parser_parseExpression(state, 1)
		}()
	}()
}

func selfhost_parser_parser_parseAssignmentExpression(state ParserState, target ParsedExpr, minPrec int) ExprStep {
	op := selfhost_parser_parser_parserAdvance(state)
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(op.state), 1)
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: value.state, expr: selfhost_parser_parser_opNode(ExprKind_Assign, op.token.lexeme, op.token, []ParsedExpr{target, value.expr})}, minPrec)
}

func selfhost_parser_parser_parsePatternPredicateExpression(state ParserState, subject ParsedExpr, minPrec int) ExprStep {
	op := selfhost_parser_parser_parserAdvance(state)
	pattern := selfhost_parser_parser_parsePredicatePatternText(selfhost_parser_parser_parserSkipNewlines(op.state))
	trueExpr := selfhost_parser_parser_valueNode(ExprKind_Bool, "true", op.token)
	falseExpr := selfhost_parser_parser_valueNode(ExprKind_Bool, "false", op.token)
	patternBranch := selfhost_parser_parser_makeExpr(ExprKind_Branch, pattern.expr.text, "", "", "=>", []ParsedParam{}, []ParsedExpr{pattern.expr, trueExpr}, pattern.expr.line, pattern.expr.column)
	wildcardPattern := selfhost_parser_parser_makeExpr(ExprKind_Pattern, "_", "", "", "", []ParsedParam{}, []ParsedExpr{}, pattern.expr.line, pattern.expr.column)
	wildcardBranch := selfhost_parser_parser_makeExpr(ExprKind_Branch, "_", "", "", "=>", []ParsedParam{}, []ParsedExpr{wildcardPattern, falseExpr}, pattern.expr.line, pattern.expr.column)
	matchExpr := selfhost_parser_parser_appendChild(selfhost_parser_parser_appendChild(selfhost_parser_parser_appendChild(selfhost_parser_parser_node(ExprKind_Match, op.token), subject), patternBranch), wildcardBranch)
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: pattern.state, expr: matchExpr}, minPrec)
}

func selfhost_parser_parser_parseQuestionExpression(state ParserState, expr ParsedExpr, minPrec int) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_questionIsPostfixUnwrap(state) {
			return selfhost_parser_parser_parseExpressionLoop(selfhost_parser_parser_parseResultUnwrapExpression(state, expr), minPrec)
		}
		return selfhost_parser_parser_parseTernaryExpression(state, expr, minPrec)
	}()
}

func selfhost_parser_parser_parseQuestionQuestionExpression(state ParserState, expr ParsedExpr, minPrec int) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_questionQuestionIsPostfixUnwrap(state) {
			return selfhost_parser_parser_parseExpressionLoop(selfhost_parser_parser_parseResultUnwrapExpressionQQ(state, expr), minPrec)
		}
		return selfhost_parser_parser_parseBinaryExpression(state, expr, minPrec)
	}()
}

func selfhost_parser_parser_parseResultUnwrapExpression(state ParserState, expr ParsedExpr) ExprStep {
	question := selfhost_parser_parser_parserAdvance(state)
	return ExprStep{state: question.state, expr: selfhost_parser_parser_opNode(ExprKind_Unwrap, "?", question.token, []ParsedExpr{expr})}
}

func selfhost_parser_parser_parseResultUnwrapExpressionQQ(state ParserState, expr ParsedExpr) ExprStep {
	token := selfhost_parser_parser_parserAdvance(state)
	return ExprStep{state: token.state, expr: selfhost_parser_parser_opNode(ExprKind_Unwrap, "??", token.token, []ParsedExpr{expr})}
}

func selfhost_parser_parser_parseTernaryExpression(state ParserState, condition ParsedExpr, minPrec int) ExprStep {
	question := selfhost_parser_parser_parserAdvance(state)
	consequence := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(question.state), 1)
	afterConsequence := selfhost_parser_parser_parserSkipNewlines(consequence.state)
	colon := selfhost_parser_parser_parserMatch(afterConsequence, TokenKind_Colon)
	return func() ExprStep {
		if colon.ok {
			return selfhost_parser_parser_parseTernaryAlternative(colon.state, question.token, condition, consequence.expr, minPrec)
		}
		return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: afterConsequence, expr: selfhost_parser_parser_opNode(ExprKind_Ternary, "?:", question.token, []ParsedExpr{condition, consequence.expr})}, minPrec)
	}()
}

func selfhost_parser_parser_parseTernaryAlternative(state ParserState, token Token, condition ParsedExpr, consequence ParsedExpr, minPrec int) ExprStep {
	alternative := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(state), 1)
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: alternative.state, expr: selfhost_parser_parser_opNode(ExprKind_Ternary, "?:", token, []ParsedExpr{condition, consequence, alternative.expr})}, minPrec)
}

func selfhost_parser_parser_parseBinaryExpression(state ParserState, expr ParsedExpr, minPrec int) ExprStep {
	prec := selfhost_parser_parser_precedence(selfhost_parser_parser_parserPeek(state).kind)
	return func() ExprStep {
		if prec < minPrec {
			return ExprStep{state: state, expr: expr}
		}
		return selfhost_parser_parser_parseBinaryExpressionAtPrec(state, expr, minPrec, prec)
	}()
}

func selfhost_parser_parser_parseBinaryExpressionAtPrec(state ParserState, expr ParsedExpr, minPrec int, prec int) ExprStep {
	op := selfhost_parser_parser_parserAdvance(state)
	right := selfhost_parser_parser_parseExpression(op.state, prec+1)
	return selfhost_parser_parser_parseExpressionLoop(ExprStep{state: right.state, expr: selfhost_parser_parser_opNode(ExprKind_Binary, op.token.lexeme, op.token, []ParsedExpr{expr, right.expr})}, minPrec)
}

func selfhost_parser_parser_parseUnary(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_Minus) || selfhost_parser_parser_parserCheck(state, TokenKind_Bang) || selfhost_parser_parser_parserCheck(state, TokenKind_Tilde) {
			return selfhost_parser_parser_parseUnaryOperator(state)
		}
		return selfhost_parser_parser_parsePrimary(state)
	}()
}

func selfhost_parser_parser_parseUnaryOperator(state ParserState) ExprStep {
	op := selfhost_parser_parser_parserAdvance(state)
	right := selfhost_parser_parser_parseExpression(op.state, 11)
	return ExprStep{state: right.state, expr: selfhost_parser_parser_opNode(ExprKind_Unary, op.token.lexeme, op.token, []ParsedExpr{right.expr})}
}

func selfhost_parser_parser_parsePrimary(state ParserState) ExprStep {
	token := selfhost_parser_parser_parserPeek(state)
	return func() ExprStep {
		switch {
		case token.kind == TokenKind_Int:
			return selfhost_parser_parser_parseLiteral(state, ExprKind_Int)
		case token.kind == TokenKind_Double:
			return selfhost_parser_parser_parseLiteral(state, ExprKind_Double)
		case token.kind == TokenKind_BigInt:
			return selfhost_parser_parser_parseLiteral(state, ExprKind_BigInt)
		case token.kind == TokenKind_String:
			return selfhost_parser_parser_parseLiteral(state, ExprKind_String)
		case token.kind == TokenKind_TemplateString:
			return selfhost_parser_parser_parseTemplateLiteral(state)
		case token.kind == TokenKind_Char:
			return selfhost_parser_parser_parseLiteral(state, ExprKind_Char)
		case token.kind == TokenKind_Regex:
			return selfhost_parser_parser_parseLiteral(state, ExprKind_Regex)
		case token.kind == TokenKind_XMLText:
			return selfhost_parser_parser_parseLiteral(state, ExprKind_XMLText)
		case token.kind == TokenKind_Ident:
			return selfhost_parser_parser_parseIdentifierPrimary(state)
		case token.kind == TokenKind_At:
			return selfhost_parser_parser_parseAtExpression(state)
		case token.kind == TokenKind_Dot:
			return selfhost_parser_parser_parseThisSelector(state)
		case token.kind == TokenKind_LBracket:
			return selfhost_parser_parser_parseArrayLiteral(state)
		case token.kind == TokenKind_Dollar:
			return selfhost_parser_parser_parseDollarExpression(state)
		case token.kind == TokenKind_LBrace:
			return selfhost_parser_parser_parseBraceLiteral(state)
		case token.kind == TokenKind_LParen:
			return selfhost_parser_parser_parseParenOrTuple(state)
		case token.kind == TokenKind_Less:
			return selfhost_parser_parser_parseXMLElement(state)
		default:
			return selfhost_parser_parser_parsePrimaryError(state)
		}
	}()
}

func selfhost_parser_parser_parseXMLElement(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_Less, "expected '<'")
	name := selfhost_parser_parser_parserConsume(open.state, TokenKind_Ident, "expected XML tag name")
	element := selfhost_parser_parser_namedNode(ExprKind_XMLElement, name.token.lexeme, open.token)
	attrs := selfhost_parser_parser_parseXMLAttributes(name.state, element)
	greater := selfhost_parser_parser_parserConsume(attrs.state, TokenKind_Greater, "expected '>' after XML tag")
	return func() ExprStep {
		if attrs.selfClosing {
			return ExprStep{state: greater.state, expr: attrs.element}
		}
		return selfhost_parser_parser_parseXMLChildren(greater.state, attrs.element)
	}()
}

func selfhost_parser_parser_parseXMLAttributes(state ParserState, element ParsedExpr) XMLAttrStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() XMLAttrStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_Slash) {
			return XMLAttrStep{state: selfhost_parser_parser_parserAdvance(current).state, element: element, selfClosing: true}
		}
		return func() XMLAttrStep {
			if selfhost_parser_parser_parserCheck(current, TokenKind_Greater) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
				return XMLAttrStep{state: current, element: element, selfClosing: false}
			}
			return selfhost_parser_parser_parseXMLAttribute(current, element)
		}()
	}()
}

func selfhost_parser_parser_parseXMLAttribute(state ParserState, element ParsedExpr) XMLAttrStep {
	eventStep := selfhost_parser_parser_parserMatch(state, TokenKind_At)
	name := selfhost_parser_parser_parserConsume(eventStep.state, TokenKind_Ident, "expected XML attribute name")
	valueStep := func() ExprStep {
		if selfhost_parser_parser_parserCheck(name.state, TokenKind_Assign) {
			return selfhost_parser_parser_parseXMLAttributeValue(selfhost_parser_parser_parserAdvance(name.state).state)
		}
		return ExprStep{state: name.state, expr: selfhost_parser_parser_emptyExpr()}
	}()
	op := func() string {
		if eventStep.ok {
			return "xmlEvent"
		}
		return func() string {
			if valueStep.expr.kind == ExprKind_Unknown {
				return "xmlBareAttr"
			}
			return "xmlAttr"
		}()
	}()
	attr := selfhost_parser_parser_makeExpr(ExprKind_Field, name.token.lexeme, name.token.lexeme, "", op, []ParsedParam{}, func() []ParsedExpr {
		if valueStep.expr.kind == ExprKind_Unknown {
			return []ParsedExpr{}
		}
		return []ParsedExpr{valueStep.expr}
	}(), name.token.line, name.token.column)
	return selfhost_parser_parser_parseXMLAttributes(valueStep.state, selfhost_parser_parser_appendChild(element, attr))
}

func selfhost_parser_parser_parseXMLAttributeValue(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_LBrace) {
			return selfhost_parser_parser_parseXMLBracedExpression(state)
		}
		return selfhost_parser_parser_parseLiteral(state, ExprKind_String)
	}()
}

func selfhost_parser_parser_parseXMLChildren(state ParserState, element ParsedExpr) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return ExprStep{state: current, expr: element}
		}
		return func() ExprStep {
			if selfhost_parser_parser_parserCheck(current, TokenKind_Less) && selfhost_parser_parser_parserCheckNext(current, TokenKind_Slash) {
				return ExprStep{state: selfhost_parser_parser_parseXMLClose(current, element.name), expr: element}
			}
			return selfhost_parser_parser_parseXMLChild(current, element)
		}()
	}()
}

func selfhost_parser_parser_parseXMLChild(state ParserState, element ParsedExpr) ExprStep {
	return func() ExprStep {
		switch {
		case selfhost_parser_parser_parserPeek(state).kind == TokenKind_XMLText:
			return selfhost_parser_parser_parseXMLTextChild(state, element)
		case selfhost_parser_parser_parserPeek(state).kind == TokenKind_Less:
			return selfhost_parser_parser_parseXMLNestedElementChild(state, element)
		case selfhost_parser_parser_parserPeek(state).kind == TokenKind_LBrace:
			return selfhost_parser_parser_parseXMLExpressionChild(state, element)
		default:
			return ExprStep{state: selfhost_parser_parser_parserAdvance(selfhost_parser_parser_parserErrorAt(state, selfhost_parser_parser_parserPeek(state), "expected XML child")).state, expr: element}
		}
	}()
}

func selfhost_parser_parser_parseXMLTextChild(state ParserState, element ParsedExpr) ExprStep {
	text := selfhost_parser_parser_parserAdvance(state)
	normalized := selfhost_parser_parser_normalizeXMLText(text.token.lexeme)
	nextElement := func() ParsedExpr {
		if normalized == "" {
			return element
		}
		return selfhost_parser_parser_appendChild(element, selfhost_parser_parser_valueNode(ExprKind_XMLText, normalized, text.token))
	}()
	return selfhost_parser_parser_parseXMLChildren(text.state, nextElement)
}

func selfhost_parser_parser_parseXMLNestedElementChild(state ParserState, element ParsedExpr) ExprStep {
	child := selfhost_parser_parser_parseXMLElement(state)
	return selfhost_parser_parser_parseXMLChildren(child.state, selfhost_parser_parser_appendChild(element, child.expr))
}

func selfhost_parser_parser_parseXMLExpressionChild(state ParserState, element ParsedExpr) ExprStep {
	child := selfhost_parser_parser_parseXMLBracedExpression(state)
	return selfhost_parser_parser_parseXMLChildren(child.state, selfhost_parser_parser_appendChild(element, child.expr))
}

func selfhost_parser_parser_parseXMLBracedExpression(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBrace, "expected '{'")
	expr := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(open.state), 1)
	close := selfhost_parser_parser_parserConsume(selfhost_parser_parser_parserSkipNewlines(expr.state), TokenKind_RBrace, "expected '}' after XML expression")
	return ExprStep{state: close.state, expr: expr.expr}
}

func selfhost_parser_parser_parseXMLClose(state ParserState, tag string) ParserState {
	less := selfhost_parser_parser_parserConsume(state, TokenKind_Less, "expected XML closing tag")
	slash := selfhost_parser_parser_parserConsume(less.state, TokenKind_Slash, "expected '/' in XML closing tag")
	name := selfhost_parser_parser_parserConsume(slash.state, TokenKind_Ident, "expected XML closing tag name")
	checked := func() ParserState {
		if name.token.lexeme == tag {
			return name.state
		}
		return selfhost_parser_parser_parserErrorAt(name.state, name.token, "mismatched XML closing tag </"+name.token.lexeme+">, expected </"+tag+">")
	}()
	return selfhost_parser_parser_parserConsume(checked, TokenKind_Greater, "expected '>' after XML closing tag").state
}

func selfhost_parser_parser_normalizeXMLText(text string) string {
	return selfhost_parser_parser_normalizeXMLTextLoop(strings.TrimSpace(text), 0, false, "")
}

func selfhost_parser_parser_normalizeXMLTextLoop(text string, index int, spacing bool, out string) string {
	return func() string {
		if index >= len([]rune(text)) {
			return out
		}
		return selfhost_parser_parser_normalizeXMLTextChar(text, index, spacing, out)
	}()
}

func selfhost_parser_parser_normalizeXMLTextChar(text string, index int, spacing bool, out string) string {
	ch := []rune(text)[index]
	isSpace := ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
	return func() string {
		if isSpace {
			return selfhost_parser_parser_normalizeXMLTextLoop(text, index+1, true, out)
		}
		return selfhost_parser_parser_normalizeXMLTextLoop(text, index+1, false, out+func() string {
			if spacing && out != "" {
				return " "
			}
			return ""
		}()+string(ch))
	}()
}

func selfhost_parser_parser_parseTemplateLiteral(state ParserState) ExprStep {
	step := selfhost_parser_parser_parserAdvance(state)
	parsed := selfhost_parser_parser_parseTemplateParts(selfhost_parser_parser_templateInner(step.token.lexeme), 0, 0, "", []ParsedExpr{})
	return ExprStep{state: step.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Template, step.token.lexeme, "", "`"+parsed.text+"`", "", []ParsedParam{}, parsed.children, step.token.line, step.token.column)}
}

func selfhost_parser_parser_templateInner(raw string) string {
	return func() string {
		if len([]rune(raw)) >= 2 {
			return func() string { runes := []rune(raw); return string(runes[1 : len([]rune(raw))-1]) }()
		}
		return raw
	}()
}

func selfhost_parser_parser_parseTemplateParts(inner string, index int, textStart int, out string, children []ParsedExpr) TemplateParse {
	return func() TemplateParse {
		if index >= len([]rune(inner)) {
			return TemplateParse{text: out + func() string { runes := []rune(inner); return string(runes[textStart:len([]rune(inner))]) }(), children: children}
		}
		return selfhost_parser_parser_parseTemplatePartAt(inner, index, textStart, out, children)
	}()
}

func selfhost_parser_parser_parseTemplatePartAt(inner string, index int, textStart int, out string, children []ParsedExpr) TemplateParse {
	ch := []rune(inner)[index]
	return func() TemplateParse {
		if ch == '\\' && index+1 < len([]rune(inner)) && []rune(inner)[index+1] == '(' {
			return selfhost_parser_parser_parseTemplateExprPart(inner, index, textStart, out, children)
		}
		return selfhost_parser_parser_parseTemplateParts(inner, index+1, textStart, out, children)
	}()
}

func selfhost_parser_parser_parseTemplateExprPart(inner string, index int, textStart int, out string, children []ParsedExpr) TemplateParse {
	exprStart := index + 2
	exprEnd := selfhost_parser_parser_scanTemplateExprEnd(inner, exprStart, 1)
	return func() TemplateParse {
		if exprEnd < 0 {
			return TemplateParse{text: out + func() string { runes := []rune(inner); return string(runes[textStart:len([]rune(inner))]) }(), children: children}
		}
		return selfhost_parser_parser_parseTemplateParts(inner, exprEnd+1, exprEnd+1, out+func() string { runes := []rune(inner); return string(runes[textStart:index]) }()+"<<<RUNE_TEMPLATE_PART>>>", selfhost_parser_parser_pushTemplateExpr(children, func() string { runes := []rune(inner); return string(runes[exprStart:exprEnd]) }()))
	}()
}

func selfhost_parser_parser_scanTemplateExprEnd(inner string, index int, depth int) int {
	return func() int {
		if index >= len([]rune(inner)) {
			return -1
		}
		return selfhost_parser_parser_scanTemplateExprEndAt(inner, index, depth)
	}()
}

func selfhost_parser_parser_scanTemplateExprEndAt(inner string, index int, depth int) int {
	ch := []rune(inner)[index]
	return func() int {
		switch {
		case ch == '"':
			return selfhost_parser_parser_scanTemplateExprEnd(inner, selfhost_parser_parser_skipTemplateQuoted(inner, index+1, '"'), depth)
		case ch == '\'':
			return selfhost_parser_parser_scanTemplateExprEnd(inner, selfhost_parser_parser_skipTemplateQuoted(inner, index+1, '\''), depth)
		case ch == '`':
			return selfhost_parser_parser_scanTemplateExprEnd(inner, selfhost_parser_parser_skipTemplateQuoted(inner, index+1, '`'), depth)
		case ch == '(':
			return selfhost_parser_parser_scanTemplateExprEnd(inner, index+1, depth+1)
		case ch == ')':
			return func() int {
				if depth == 1 {
					return index
				}
				return selfhost_parser_parser_scanTemplateExprEnd(inner, index+1, depth-1)
			}()
		default:
			return selfhost_parser_parser_scanTemplateExprEnd(inner, index+1, depth)
		}
	}()
}

func selfhost_parser_parser_skipTemplateQuoted(text string, index int, quote rune) int {
	return func() int {
		if index >= len([]rune(text)) {
			return index
		}
		return selfhost_parser_parser_skipTemplateQuotedAt(text, index, quote)
	}()
}

func selfhost_parser_parser_skipTemplateQuotedAt(text string, index int, quote rune) int {
	ch := []rune(text)[index]
	return func() int {
		if ch == '\\' {
			return selfhost_parser_parser_skipTemplateQuoted(text, index+2, quote)
		}
		return func() int {
			if ch == quote {
				return index + 1
			}
			return selfhost_parser_parser_skipTemplateQuoted(text, index+1, quote)
		}()
	}()
}

func selfhost_parser_parser_pushTemplateExpr(children []ParsedExpr, source string) []ParsedExpr {
	parsed := selfhost_parser_parser_parseTemplateExpression(strings.TrimSpace(source))
	return func() []ParsedExpr {
		__rune_spread_out := []ParsedExpr{}
		__rune_spread_out = append(__rune_spread_out, children...)
		__rune_spread_out = append(__rune_spread_out, parsed)
		return __rune_spread_out
	}()
}

func selfhost_parser_parser_parseTemplateExpression(source string) ParsedExpr {
	return func() ParsedExpr {
		if source == "" {
			return selfhost_parser_parser_emptyExpr()
		}
		return selfhost_parser_parser_parseExpression(ParserState{tokens: lex(source), current: 0, errors: []ParseError{}}, 1).expr
	}()
}

func selfhost_parser_parser_parseLiteral(state ParserState, kind ExprKind) ExprStep {
	step := selfhost_parser_parser_parserAdvance(state)
	return ExprStep{state: step.state, expr: selfhost_parser_parser_valueNode(kind, step.token.lexeme, step.token)}
}

func selfhost_parser_parser_parseIdentifierPrimary(state ParserState) ExprStep {
	step := selfhost_parser_parser_parserAdvance(state)
	kind := func() ExprKind {
		if step.token.lexeme == "true" || step.token.lexeme == "false" {
			return ExprKind_Bool
		}
		return func() ExprKind {
			if step.token.lexeme == "null" {
				return ExprKind_Null
			}
			return ExprKind_Identifier
		}()
	}()
	return ExprStep{state: step.state, expr: func() ParsedExpr {
		if kind == ExprKind_Identifier {
			return selfhost_parser_parser_namedNode(kind, step.token.lexeme, step.token)
		}
		return selfhost_parser_parser_valueNode(kind, step.token.lexeme, step.token)
	}()}
}

func selfhost_parser_parser_parseAtExpression(state ParserState) ExprStep {
	at := selfhost_parser_parser_parserConsume(state, TokenKind_At, "expected '@'")
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(at.state, TokenKind_String) {
			return selfhost_parser_parser_parseAtImportExpression(at)
		}
		return selfhost_parser_parser_parseAtModuleExpression(at)
	}()
}

func selfhost_parser_parser_parseAtImportExpression(at TokenStep) ExprStep {
	path := selfhost_parser_parser_parserAdvance(at.state)
	return ExprStep{state: path.state, expr: selfhost_parser_parser_valueNode(ExprKind_At, path.token.lexeme, at.token)}
}

func selfhost_parser_parser_parseAtModuleExpression(at TokenStep) ExprStep {
	name := selfhost_parser_parser_parserConsume(at.state, TokenKind_Ident, "expected module name after '@'")
	return ExprStep{state: name.state, expr: selfhost_parser_parser_namedNode(ExprKind_At, name.token.lexeme, at.token)}
}

func selfhost_parser_parser_parseThisSelector(state ParserState) ExprStep {
	dot := selfhost_parser_parser_parserConsume(state, TokenKind_Dot, "expected '.'")
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(dot.state, TokenKind_Ident) {
			return selfhost_parser_parser_parseThisFieldSelector(dot.state, dot.token)
		}
		return ExprStep{state: dot.state, expr: selfhost_parser_parser_node(ExprKind_This, dot.token)}
	}()
}

func selfhost_parser_parser_parseThisFieldSelector(state ParserState, dot Token) ExprStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected field name after '.'")
	return ExprStep{state: name.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Selector, name.token.lexeme, name.token.lexeme, "", ".", []ParsedParam{}, []ParsedExpr{selfhost_parser_parser_node(ExprKind_This, dot)}, dot.line, dot.column)}
}

func selfhost_parser_parser_parseArrayLiteral(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBracket, "expected '['")
	args := selfhost_parser_parser_parseArgumentList(selfhost_parser_parser_parserSkipNewlines(open.state), []ParsedExpr{}, TokenKind_RBracket)
	close := selfhost_parser_parser_parserConsume(args.state, TokenKind_RBracket, "expected ']' after array literal")
	return ExprStep{state: close.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Array, "[]", "", "", "", []ParsedParam{}, args.expr.children, open.token.line, open.token.column)}
}

func selfhost_parser_parser_parseDollarExpression(state ParserState) ExprStep {
	start := selfhost_parser_parser_parserConsume(state, TokenKind_Dollar, "expected '$'")
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(start.state, TokenKind_Ident) {
			return selfhost_parser_parser_parseSignalPrefixIdentifier(start)
		}
		return ExprStep{state: selfhost_parser_parser_parserErrorAt(start.state, selfhost_parser_parser_parserPeek(start.state), "expected signal name after '$'"), expr: selfhost_parser_parser_emptyExpr()}
	}()
}

func selfhost_parser_parser_parseSignalPrefixIdentifier(start TokenStep) ExprStep {
	name := selfhost_parser_parser_parserAdvance(start.state)
	return ExprStep{state: name.state, expr: selfhost_parser_parser_namedNode(ExprKind_Identifier, name.token.lexeme, name.token)}
}

func selfhost_parser_parser_parseBraceLiteral(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_looksLikeMapLiteralBody(state) {
			return selfhost_parser_parser_parseMapLiteral(state)
		}
		return selfhost_parser_parser_parseObjectLiteral(state)
	}()
}

func selfhost_parser_parser_parseMapLiteral(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBrace, "expected '{'")
	return selfhost_parser_parser_parseMapLiteralLoop(selfhost_parser_parser_parserSkipNewlines(open.state), selfhost_parser_parser_node(ExprKind_Map, open.token))
}

func selfhost_parser_parser_parseMapLiteralLoop(state ParserState, mapExpr ParsedExpr) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return selfhost_parser_parser_finishMapLiteral(current, mapExpr)
		}
		return selfhost_parser_parser_parseMapLiteralEntry(current, mapExpr)
	}()
}

func selfhost_parser_parser_parseMapLiteralEntry(state ParserState, mapExpr ParsedExpr) ExprStep {
	key := selfhost_parser_parser_parseExpression(state, 1)
	colon := selfhost_parser_parser_parserConsume(key.state, TokenKind_Colon, "expected ':' after map key")
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(colon.state), 1)
	entry := selfhost_parser_parser_makeExpr(ExprKind_Entry, "", "", "", ":", []ParsedParam{}, []ParsedExpr{key.expr, value.expr}, key.expr.line, key.expr.column)
	nextMap := selfhost_parser_parser_appendChild(mapExpr, entry)
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(selfhost_parser_parser_consumeStatementEnd(value.state)), TokenKind_Comma)
	return selfhost_parser_parser_parseMapLiteralLoop(comma.state, nextMap)
}

func selfhost_parser_parser_finishMapLiteral(state ParserState, mapExpr ParsedExpr) ExprStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RBrace, "expected '}' after map literal")
	return ExprStep{state: close.state, expr: mapExpr}
}

func selfhost_parser_parser_parseObjectLiteral(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBrace, "expected '{'")
	return selfhost_parser_parser_parseObjectLiteralLoop(selfhost_parser_parser_parserSkipNewlines(open.state), selfhost_parser_parser_node(ExprKind_Object, open.token))
}

func selfhost_parser_parser_parseObjectLiteralLoop(state ParserState, object ParsedExpr) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return selfhost_parser_parser_finishObjectLiteral(current, object)
		}
		return selfhost_parser_parser_parseObjectLiteralMember(current, object)
	}()
}

func selfhost_parser_parser_parseObjectLiteralMember(state ParserState, object ParsedExpr) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_DotDot) {
			return selfhost_parser_parser_parseObjectLiteralSpreadMember(state, object)
		}
		return selfhost_parser_parser_parseObjectLiteralNamedMember(state, object)
	}()
}

func selfhost_parser_parser_parseObjectLiteralNamedMember(state ParserState, object ParsedExpr) ExprStep {
	privateStep := selfhost_parser_parser_parseObjectPrivateModifier(state)
	memberState := privateStep.state
	member := func() ExprStep {
		if selfhost_parser_parser_looksLikeFunctionDecl(memberState) {
			return selfhost_parser_parser_parseObjectMethod(memberState, privateStep.ok)
		}
		return selfhost_parser_parser_parseObjectField(memberState, privateStep.ok)
	}()
	nextObject := selfhost_parser_parser_appendChild(object, member.expr)
	return selfhost_parser_parser_parseObjectLiteralLoop(selfhost_parser_parser_consumeFieldSeparator(member.state, TokenKind_RBrace, "expected ',' between object literal fields"), nextObject)
}

func selfhost_parser_parser_parseObjectLiteralSpreadMember(state ParserState, object ParsedExpr) ExprStep {
	spread := selfhost_parser_parser_parserConsume(state, TokenKind_DotDot, "expected '..'")
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(spread.state), 1)
	member := selfhost_parser_parser_opNode(ExprKind_Spread, "..", spread.token, []ParsedExpr{value.expr})
	nextObject := selfhost_parser_parser_appendChild(object, member)
	return selfhost_parser_parser_parseObjectLiteralLoop(selfhost_parser_parser_consumeFieldSeparator(value.state, TokenKind_RBrace, "expected ',' between object literal fields"), nextObject)
}

func selfhost_parser_parser_parseObjectMethod(state ParserState, private bool) ExprStep {
	fn := selfhost_parser_parser_parseFunctionWithReceiver(state, "", private, false, false, selfhost_parser_parser_emptyAnnotations())
	return ExprStep{state: fn.state, expr: selfhost_parser_parser_functionToExpr(fn.function)}
}

func selfhost_parser_parser_parseObjectField(state ParserState, private bool) ExprStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected field name")
	colon := selfhost_parser_parser_parserConsume(name.state, TokenKind_Colon, "expected ':' after field name")
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(colon.state), 1)
	return ExprStep{state: value.state, expr: selfhost_parser_parser_makeExpr(func() ExprKind {
		if private {
			return ExprKind_PrivateField
		}
		return ExprKind_Field
	}(), name.token.lexeme, name.token.lexeme, "", ":", []ParsedParam{}, []ParsedExpr{value.expr}, name.token.line, name.token.column)}
}

func selfhost_parser_parser_finishObjectLiteral(state ParserState, object ParsedExpr) ExprStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RBrace, "expected '}' after object literal")
	return ExprStep{state: close.state, expr: object}
}

func selfhost_parser_parser_parseStructLiteral(state ParserState, typeExpr ParsedExpr) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBrace, "expected '{' after type name")
	return selfhost_parser_parser_parseStructLiteralLoop(selfhost_parser_parser_parserSkipNewlines(open.state), selfhost_parser_parser_namedNode(ExprKind_Struct, typeExpr.name, open.token))
}

func selfhost_parser_parser_parseStructLiteralLoop(state ParserState, structExpr ParsedExpr) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return selfhost_parser_parser_finishStructLiteral(current, structExpr)
		}
		return selfhost_parser_parser_parseStructLiteralField(current, structExpr)
	}()
}

func selfhost_parser_parser_parseStructLiteralField(state ParserState, structExpr ParsedExpr) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_DotDot) {
			return selfhost_parser_parser_parseStructLiteralSpreadField(state, structExpr)
		}
		return selfhost_parser_parser_parseStructLiteralNamedField(state, structExpr)
	}()
}

func selfhost_parser_parser_parseStructLiteralNamedField(state ParserState, structExpr ParsedExpr) ExprStep {
	name := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected field name")
	colon := selfhost_parser_parser_parserConsume(name.state, TokenKind_Colon, "expected ':' after field name")
	value := selfhost_parser_parser_parseExpression(selfhost_parser_parser_parserSkipNewlines(colon.state), 1)
	field := selfhost_parser_parser_makeExpr(ExprKind_Field, name.token.lexeme, name.token.lexeme, "", ":", []ParsedParam{}, []ParsedExpr{value.expr}, name.token.line, name.token.column)
	nextStruct := selfhost_parser_parser_appendChild(structExpr, field)
	return selfhost_parser_parser_parseStructLiteralLoop(selfhost_parser_parser_consumeFieldSeparator(value.state, TokenKind_RBrace, "expected ',' between struct literal fields"), nextStruct)
}

func selfhost_parser_parser_parseStructLiteralSpreadField(state ParserState, structExpr ParsedExpr) ExprStep {
	spread := selfhost_parser_parser_parserConsume(state, TokenKind_DotDot, "expected '..'")
	value := selfhost_parser_parser_parseExpression(spread.state, 1)
	field := selfhost_parser_parser_opNode(ExprKind_Spread, "..", spread.token, []ParsedExpr{value.expr})
	nextStruct := selfhost_parser_parser_appendChild(structExpr, field)
	return selfhost_parser_parser_parseStructLiteralLoop(selfhost_parser_parser_consumeFieldSeparator(value.state, TokenKind_RBrace, "expected ',' between struct literal fields"), nextStruct)
}

func selfhost_parser_parser_finishStructLiteral(state ParserState, structExpr ParsedExpr) ExprStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RBrace, "expected '}' after struct literal")
	return ExprStep{state: close.state, expr: structExpr}
}

func selfhost_parser_parser_parseParenOrTuple(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LParen, "expected '('")
	afterOpen := selfhost_parser_parser_parserSkipNewlines(open.state)
	return selfhost_parser_parser_parseParenOrTupleAfterOpen(afterOpen, open.token)
}

func selfhost_parser_parser_parseParenOrTupleAfterOpen(state ParserState, open Token) ExprStep {
	expr := selfhost_parser_parser_parseExpression(state, 1)
	afterExpr := selfhost_parser_parser_parserSkipNewlines(expr.state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(afterExpr, TokenKind_Comma) {
			return selfhost_parser_parser_parseTupleAfterFirst(afterExpr, open, expr.expr)
		}
		return selfhost_parser_parser_finishParenExpression(afterExpr, expr.expr)
	}()
}

func selfhost_parser_parser_finishParenExpression(state ParserState, expr ParsedExpr) ExprStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RParen, "expected ')' after expression")
	return ExprStep{state: close.state, expr: expr}
}

func selfhost_parser_parser_parseTupleAfterFirst(state ParserState, open Token, first ParsedExpr) ExprStep {
	comma := selfhost_parser_parser_parserConsume(state, TokenKind_Comma, "expected ','")
	holder := selfhost_parser_parser_appendChild(selfhost_parser_parser_node(ExprKind_Tuple, open), first)
	values := selfhost_parser_parser_parseArgumentListLoop(selfhost_parser_parser_parserSkipNewlines(comma.state), holder, TokenKind_RParen)
	close := selfhost_parser_parser_parserConsume(values.state, TokenKind_RParen, "expected ')' after tuple literal")
	return ExprStep{expr: values.expr, state: close.state}
}

func selfhost_parser_parser_parsePrimaryError(state ParserState) ExprStep {
	token := selfhost_parser_parser_parserPeek(state)
	step := selfhost_parser_parser_parserAdvance(selfhost_parser_parser_parserErrorAt(state, token, "expected expression, got "+tokenKindName(token.kind)))
	return ExprStep{state: step.state, expr: selfhost_parser_parser_namedNode(ExprKind_Error, "<error>", token)}
}

func selfhost_parser_parser_parseLambda(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LParen, "expected '(' before lambda parameters")
	params := selfhost_parser_parser_parseParamList(selfhost_parser_parser_parserSkipNewlines(open.state))
	close := selfhost_parser_parser_parserConsume(params.state, TokenKind_RParen, "expected ')' after lambda parameters")
	afterParams := selfhost_parser_parser_parserSkipNewlines(close.state)
	ret := selfhost_parser_parser_parserMatch(afterParams, TokenKind_Arrow)
	returnType := func() TypeRefStep {
		if ret.ok {
			return selfhost_parser_parser_parseTypeRef(selfhost_parser_parser_parserSkipNewlines(ret.state))
		}
		return TypeRefStep{state: afterParams, typeRef: emptyParsedTypeRef()}
	}()
	arrow := selfhost_parser_parser_parserConsume(selfhost_parser_parser_parserSkipNewlines(returnType.state), TokenKind_FatArrow, "expected '=>' after lambda parameter")
	body := selfhost_parser_parser_parseBody(selfhost_parser_parser_parserSkipNewlines(arrow.state))
	return ExprStep{state: body.state, expr: selfhost_parser_parser_withChildren(selfhost_parser_parser_withParams(selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Lambda, open.token), typeRefToString(returnType.typeRef)), params.params), []ParsedExpr{body.expr})}
}

func selfhost_parser_parser_parseMatchExpression(state ParserState, subject ParsedExpr) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBrace, "expected '{' after match subject")
	block := selfhost_parser_parser_appendChild(selfhost_parser_parser_node(ExprKind_Match, open.token), subject)
	return selfhost_parser_parser_parsePatternBlock(selfhost_parser_parser_parserSkipNewlines(open.state), block)
}

func selfhost_parser_parser_parsePatternBlock(state ParserState, block ParsedExpr) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF) {
			return selfhost_parser_parser_finishPatternBlock(current, block)
		}
		return selfhost_parser_parser_parsePatternBranch(current, block)
	}()
}

func selfhost_parser_parser_parsePatternBranch(state ParserState, block ParsedExpr) ExprStep {
	pattern := selfhost_parser_parser_parsePatternText(state)
	arrow := selfhost_parser_parser_parserConsume(pattern.state, TokenKind_FatArrow, "expected '=>' after pattern")
	value := selfhost_parser_parser_parseBody(selfhost_parser_parser_parserSkipNewlines(arrow.state))
	branch := selfhost_parser_parser_makeExpr(ExprKind_Branch, pattern.expr.text, "", "", "=>", []ParsedParam{}, []ParsedExpr{pattern.expr, value.expr}, pattern.expr.line, pattern.expr.column)
	nextBlock := selfhost_parser_parser_appendChild(block, branch)
	return selfhost_parser_parser_parsePatternBlock(selfhost_parser_parser_consumeStatementEnd(value.state), nextBlock)
}

func selfhost_parser_parser_finishPatternBlock(state ParserState, block ParsedExpr) ExprStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RBrace, "expected '}' after pattern block")
	return ExprStep{state: close.state, expr: block}
}

func selfhost_parser_parser_parsePatternText(state ParserState) ExprStep {
	return selfhost_parser_parser_parsePattern(state)
}

func selfhost_parser_parser_parsePredicatePatternText(state ParserState) ExprStep {
	return selfhost_parser_parser_parsePattern(state)
}

func selfhost_parser_parser_parsePattern(state ParserState) ExprStep {
	return selfhost_parser_parser_parseOrPatternRest(selfhost_parser_parser_parseAliasPattern(state))
}

func selfhost_parser_parser_parseOrPatternRest(left ExprStep) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(left.state)
	op := selfhost_parser_parser_parserMatch(current, TokenKind_BitOr)
	return func() ExprStep {
		if op.ok {
			return selfhost_parser_parser_parseOrPatternRest(selfhost_parser_parser_parseOrPatternRight(left, op.state))
		}
		return ExprStep{expr: left.expr, state: current}
	}()
}

func selfhost_parser_parser_parseOrPatternRight(left ExprStep, state ParserState) ExprStep {
	right := selfhost_parser_parser_parseAliasPattern(selfhost_parser_parser_parserSkipNewlines(state))
	return ExprStep{state: right.state, expr: selfhost_parser_parser_withText(left.expr, left.expr.text+"|"+right.expr.text)}
}

func selfhost_parser_parser_parseAliasPattern(state ParserState) ExprStep {
	pattern := selfhost_parser_parser_parseRangePattern(state)
	current := selfhost_parser_parser_parserSkipNewlines(pattern.state)
	alias := selfhost_parser_parser_parserCheck(current, TokenKind_At)
	return func() ExprStep {
		if alias {
			return selfhost_parser_parser_parseAliasPatternName(pattern, selfhost_parser_parser_parserAdvance(current).state)
		}
		return ExprStep{expr: pattern.expr, state: current}
	}()
}

func selfhost_parser_parser_parseAliasPatternName(pattern ExprStep, state ParserState) ExprStep {
	name := selfhost_parser_parser_parserConsume(selfhost_parser_parser_parserSkipNewlines(state), TokenKind_Ident, "expected binding name after '@'")
	return ExprStep{state: name.state, expr: selfhost_parser_parser_withText(pattern.expr, pattern.expr.text+"@"+name.token.lexeme)}
}

func selfhost_parser_parser_parseRangePattern(state ParserState) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(state, TokenKind_DotDotEqual) {
			return selfhost_parser_parser_parseOpenRangePattern(state)
		}
		return selfhost_parser_parser_parseRangePatternAfterAtom(state)
	}()
}

func selfhost_parser_parser_parseOpenRangePattern(state ParserState) ExprStep {
	start := selfhost_parser_parser_parserAdvance(state)
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(start.state, TokenKind_FatArrow) {
			return ExprStep{state: start.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, start.token), start.token.lexeme)}
		}
		return selfhost_parser_parser_parseOpenRangePatternBound(start)
	}()
}

func selfhost_parser_parser_parseOpenRangePatternBound(start TokenStep) ExprStep {
	bound := selfhost_parser_parser_parseRangeBoundPattern(selfhost_parser_parser_parserSkipNewlines(start.state))
	return ExprStep{state: bound.state, expr: selfhost_parser_parser_withText(bound.expr, "..="+bound.expr.text)}
}

func selfhost_parser_parser_parseRangePatternAfterAtom(state ParserState) ExprStep {
	pattern := selfhost_parser_parser_parsePatternAtom(state)
	current := selfhost_parser_parser_parserSkipNewlines(pattern.state)
	return func() ExprStep {
		switch {
		case selfhost_parser_parser_parserPeek(current).kind == TokenKind_DotDotEqual:
			return selfhost_parser_parser_parseRangePatternEnd(pattern, selfhost_parser_parser_parserAdvance(current), "..=")
		case selfhost_parser_parser_parserPeek(current).kind == TokenKind_DotDotLess:
			return selfhost_parser_parser_parseRangePatternEnd(pattern, selfhost_parser_parser_parserAdvance(current), "..<")
		case selfhost_parser_parser_parserPeek(current).kind == TokenKind_DotDot:
			return selfhost_parser_parser_parseDotDotRangePattern(pattern, selfhost_parser_parser_parserAdvance(current))
		default:
			return ExprStep{expr: pattern.expr, state: current}
		}
	}()
}

func selfhost_parser_parser_parseDotDotRangePattern(pattern ExprStep, op TokenStep) ExprStep {
	less := selfhost_parser_parser_parserConsume(op.state, TokenKind_Less, "expected '<' after '..' in range pattern")
	return selfhost_parser_parser_parseRangePatternEnd(pattern, less, "..<")
}

func selfhost_parser_parser_parseRangePatternEnd(pattern ExprStep, op TokenStep, text string) ExprStep {
	return func() ExprStep {
		if selfhost_parser_parser_parserCheck(op.state, TokenKind_FatArrow) {
			return ExprStep{state: op.state, expr: selfhost_parser_parser_withText(pattern.expr, pattern.expr.text+text)}
		}
		return selfhost_parser_parser_parseRangePatternEndBound(pattern, op, text)
	}()
}

func selfhost_parser_parser_parseRangePatternEndBound(pattern ExprStep, op TokenStep, text string) ExprStep {
	bound := selfhost_parser_parser_parseRangeBoundPattern(selfhost_parser_parser_parserSkipNewlines(op.state))
	return ExprStep{state: bound.state, expr: selfhost_parser_parser_withText(pattern.expr, pattern.expr.text+text+bound.expr.text)}
}

func selfhost_parser_parser_parseRangeBoundPattern(state ParserState) ExprStep {
	token := selfhost_parser_parser_parserPeek(state)
	return func() ExprStep {
		switch {
		case token.kind == TokenKind_Underscore:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_Int:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_Double:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_BigInt:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_String:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_Char:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_Ident:
			return selfhost_parser_parser_parseIdentifierRangeBoundPattern(state)
		default:
			return selfhost_parser_parser_parsePatternError(state)
		}
	}()
}

func selfhost_parser_parser_parseIdentifierRangeBoundPattern(state ParserState) ExprStep {
	name := selfhost_parser_parser_parserAdvance(state)
	dotted := selfhost_parser_parser_parserMatch(name.state, TokenKind_Dot)
	return func() ExprStep {
		if dotted.ok {
			return selfhost_parser_parser_parseDottedPatternName(name.token, dotted.state)
		}
		return ExprStep{state: name.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, name.token), name.token.lexeme)}
	}()
}

func selfhost_parser_parser_parsePatternAtom(state ParserState) ExprStep {
	token := selfhost_parser_parser_parserPeek(state)
	return func() ExprStep {
		switch {
		case token.kind == TokenKind_Underscore:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_Int:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_Double:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_BigInt:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_String:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_Char:
			return selfhost_parser_parser_parsePatternToken(state)
		case token.kind == TokenKind_Ident:
			return selfhost_parser_parser_parseIdentifierPattern(state)
		case token.kind == TokenKind_Less:
			return selfhost_parser_parser_parseComparePattern(state)
		case token.kind == TokenKind_LessEqual:
			return selfhost_parser_parser_parseComparePattern(state)
		case token.kind == TokenKind_Greater:
			return selfhost_parser_parser_parseComparePattern(state)
		case token.kind == TokenKind_GreaterEqual:
			return selfhost_parser_parser_parseComparePattern(state)
		case token.kind == TokenKind_LBrace:
			return selfhost_parser_parser_parseMapOrObjectPattern(state)
		case token.kind == TokenKind_LBracket:
			return selfhost_parser_parser_parseArrayPattern(state)
		case token.kind == TokenKind_LParen:
			return selfhost_parser_parser_parseTupleOrGroupedPattern(state)
		case token.kind == TokenKind_DotDotEqual:
			return selfhost_parser_parser_parsePatternToken(state)
		default:
			return selfhost_parser_parser_parsePatternError(state)
		}
	}()
}

func selfhost_parser_parser_parsePatternToken(state ParserState) ExprStep {
	step := selfhost_parser_parser_parserAdvance(state)
	return ExprStep{state: step.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, step.token), step.token.lexeme)}
}

func selfhost_parser_parser_parseIdentifierPattern(state ParserState) ExprStep {
	constructor := selfhost_parser_parser_parserCheckNext(state, TokenKind_LParen)
	return func() ExprStep {
		if constructor {
			return selfhost_parser_parser_parseConstructorPattern(state)
		}
		return selfhost_parser_parser_parseIdentifierPatternNonConstructor(state)
	}()
}

func selfhost_parser_parser_parseIdentifierPatternNonConstructor(state ParserState) ExprStep {
	qualified := selfhost_parser_parser_parserCheckNext(state, TokenKind_Dot)
	return func() ExprStep {
		if qualified {
			return selfhost_parser_parser_parseQualifiedIdentifierPattern(state)
		}
		return selfhost_parser_parser_parsePatternToken(state)
	}()
}

func selfhost_parser_parser_parseQualifiedIdentifierPattern(state ParserState) ExprStep {
	name := selfhost_parser_parser_parserAdvance(state)
	dot := selfhost_parser_parser_parserConsume(name.state, TokenKind_Dot, "expected '.' after pattern qualifier")
	return selfhost_parser_parser_parseDottedPatternName(name.token, dot.state)
}

func selfhost_parser_parser_parseDottedPatternName(first Token, state ParserState) ExprStep {
	second := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected name after '.'")
	return ExprStep{state: second.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, first), first.lexeme+"."+second.token.lexeme)}
}

func selfhost_parser_parser_parseComparePattern(state ParserState) ExprStep {
	op := selfhost_parser_parser_parserAdvance(state)
	value := selfhost_parser_parser_parseRangeBoundPattern(selfhost_parser_parser_parserSkipNewlines(op.state))
	return ExprStep{state: value.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, op.token), op.token.lexeme+value.expr.text)}
}

func selfhost_parser_parser_parseConstructorPattern(state ParserState) ExprStep {
	name := selfhost_parser_parser_parserAdvance(state)
	open := selfhost_parser_parser_parserConsume(name.state, TokenKind_LParen, "expected '(' after pattern constructor")
	args := selfhost_parser_parser_parseConstructorPatternArgs(selfhost_parser_parser_parserSkipNewlines(open.state), "")
	close := selfhost_parser_parser_parserConsume(args.state, TokenKind_RParen, "expected ')' after constructor pattern")
	return ExprStep{state: close.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, name.token), name.token.lexeme+"("+args.expr.text+")")}
}

func selfhost_parser_parser_parseConstructorPatternArgs(state ParserState, out string) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	done := selfhost_parser_parser_parserCheck(current, TokenKind_RParen) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF)
	return func() ExprStep {
		if done {
			return ExprStep{state: current, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, selfhost_parser_parser_parserPeek(current)), out)}
		}
		return selfhost_parser_parser_parseConstructorPatternArg(current, out)
	}()
}

func selfhost_parser_parser_parseConstructorPatternArg(state ParserState, out string) ExprStep {
	rest := selfhost_parser_parser_parserMatch(state, TokenKind_DotDot)
	return func() ExprStep {
		if rest.ok {
			return selfhost_parser_parser_parseConstructorPatternAfterArg(rest.state, selfhost_parser_parser_appendPatternPart(out, ".."), true)
		}
		return selfhost_parser_parser_parseConstructorPatternValue(state, out)
	}()
}

func selfhost_parser_parser_parseConstructorPatternValue(state ParserState, out string) ExprStep {
	arg := selfhost_parser_parser_parsePattern(state)
	return selfhost_parser_parser_parseConstructorPatternAfterArg(arg.state, selfhost_parser_parser_appendPatternPart(out, arg.expr.text), false)
}

func selfhost_parser_parser_parseConstructorPatternAfterArg(state ParserState, out string, rest bool) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	comma := selfhost_parser_parser_parserMatch(current, TokenKind_Comma)
	done := rest || comma.ok == false
	return func() ExprStep {
		if done {
			return ExprStep{state: comma.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, selfhost_parser_parser_parserPeek(comma.state)), out)}
		}
		return selfhost_parser_parser_parseConstructorPatternArgs(comma.state, out)
	}()
}

func selfhost_parser_parser_parseArrayPattern(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBracket, "expected '[' before array pattern")
	parts := selfhost_parser_parser_parseArrayPatternParts(selfhost_parser_parser_parserSkipNewlines(open.state), "")
	close := selfhost_parser_parser_parserConsume(parts.state, TokenKind_RBracket, "expected ']' after array pattern")
	return ExprStep{state: close.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, open.token), "["+parts.expr.text+"]")}
}

func selfhost_parser_parser_parseArrayPatternParts(state ParserState, out string) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	done := selfhost_parser_parser_parserCheck(current, TokenKind_RBracket) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF)
	return func() ExprStep {
		if done {
			return ExprStep{state: current, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, selfhost_parser_parser_parserPeek(current)), out)}
		}
		return selfhost_parser_parser_parseArrayPatternPart(current, out)
	}()
}

func selfhost_parser_parser_parseArrayPatternPart(state ParserState, out string) ExprStep {
	spread := selfhost_parser_parser_parserMatch(state, TokenKind_DotDot)
	return func() ExprStep {
		if spread.ok {
			return selfhost_parser_parser_parseArraySpreadPattern(spread.state, out)
		}
		return selfhost_parser_parser_parseArrayValuePattern(state, out)
	}()
}

func selfhost_parser_parser_parseArraySpreadPattern(state ParserState, out string) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() ExprStep {
		switch {
		case selfhost_parser_parser_parserPeek(current).kind == TokenKind_String:
			return selfhost_parser_parser_parseArraySpreadValue(current, out)
		case selfhost_parser_parser_parserPeek(current).kind == TokenKind_Ident:
			return selfhost_parser_parser_parseArraySpreadIdentifier(current, out)
		default:
			return selfhost_parser_parser_parseArrayPatternAfterPart(current, selfhost_parser_parser_appendPatternPart(out, ".."))
		}
	}()
}

func selfhost_parser_parser_parseArraySpreadIdentifier(state ParserState, out string) ExprStep {
	identifier := selfhost_parser_parser_parserAdvance(state)
	constantSpread := selfhost_parser_parser_isPatternSpreadIdentifier(identifier.token.lexeme)
	text := func() string {
		if constantSpread {
			return ".. " + identifier.token.lexeme
		}
		return ".." + identifier.token.lexeme
	}()
	return selfhost_parser_parser_parseArrayPatternAfterPart(identifier.state, selfhost_parser_parser_appendPatternPart(out, text))
}

func selfhost_parser_parser_parseArraySpreadValue(state ParserState, out string) ExprStep {
	value := selfhost_parser_parser_parserAdvance(state)
	return selfhost_parser_parser_parseArrayPatternAfterPart(value.state, selfhost_parser_parser_appendPatternPart(out, ".."+value.token.lexeme))
}

func selfhost_parser_parser_parseArrayValuePattern(state ParserState, out string) ExprStep {
	value := selfhost_parser_parser_parsePattern(state)
	return selfhost_parser_parser_parseArrayPatternAfterPart(value.state, selfhost_parser_parser_appendPatternPart(out, value.expr.text))
}

func selfhost_parser_parser_parseArrayPatternAfterPart(state ParserState, out string) ExprStep {
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(state), TokenKind_Comma)
	return func() ExprStep {
		if comma.ok {
			return selfhost_parser_parser_parseArrayPatternParts(comma.state, out)
		}
		return ExprStep{state: comma.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, selfhost_parser_parser_parserPeek(comma.state)), out)}
	}()
}

func selfhost_parser_parser_parseMapOrObjectPattern(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LBrace, "expected '{' before pattern")
	parts := selfhost_parser_parser_parseMapOrObjectPatternParts(selfhost_parser_parser_parserSkipNewlines(open.state), "")
	close := selfhost_parser_parser_parserConsume(parts.state, TokenKind_RBrace, "expected '}' after pattern")
	return ExprStep{state: close.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, open.token), "{"+parts.expr.text+"}")}
}

func selfhost_parser_parser_parseMapOrObjectPatternParts(state ParserState, out string) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	done := selfhost_parser_parser_parserCheck(current, TokenKind_RBrace) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF)
	return func() ExprStep {
		if done {
			return ExprStep{state: current, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, selfhost_parser_parser_parserPeek(current)), out)}
		}
		return selfhost_parser_parser_parseMapOrObjectPatternPart(current, out)
	}()
}

func selfhost_parser_parser_parseMapOrObjectPatternPart(state ParserState, out string) ExprStep {
	rest := selfhost_parser_parser_parserMatch(state, TokenKind_DotDot)
	return func() ExprStep {
		if rest.ok {
			return selfhost_parser_parser_parseMapOrObjectPatternAfterPart(rest.state, selfhost_parser_parser_appendPatternPart(out, ".."))
		}
		return selfhost_parser_parser_parseMapOrObjectEntryPattern(state, out)
	}()
}

func selfhost_parser_parser_parseMapOrObjectEntryPattern(state ParserState, out string) ExprStep {
	objectField := selfhost_parser_parser_parserPatternLooksLikeObjectField(state)
	return func() ExprStep {
		if objectField {
			return selfhost_parser_parser_parseObjectFieldPattern(state, out)
		}
		return selfhost_parser_parser_parseMapEntryPattern(state, out)
	}()
}

func selfhost_parser_parser_parseObjectFieldPattern(state ParserState, out string) ExprStep {
	field := selfhost_parser_parser_parserConsume(state, TokenKind_Ident, "expected object pattern field")
	optional := selfhost_parser_parser_parserMatch(field.state, TokenKind_Question)
	colon := selfhost_parser_parser_parserMatch(optional.state, TokenKind_Colon)
	return func() ExprStep {
		if colon.ok {
			return selfhost_parser_parser_parseObjectFieldValuePattern(field.token, optional.ok, colon.state, out)
		}
		return selfhost_parser_parser_parseMapOrObjectPatternAfterPart(optional.state, selfhost_parser_parser_appendPatternPart(out, selfhost_parser_parser_objectFieldPatternText(field.token.lexeme, optional.ok, field.token.lexeme)))
	}()
}

func selfhost_parser_parser_parseObjectFieldValuePattern(field Token, optional bool, state ParserState, out string) ExprStep {
	value := selfhost_parser_parser_parsePattern(selfhost_parser_parser_parserSkipNewlines(state))
	return selfhost_parser_parser_parseMapOrObjectPatternAfterPart(value.state, selfhost_parser_parser_appendPatternPart(out, selfhost_parser_parser_objectFieldPatternText(field.lexeme, optional, value.expr.text)))
}

func selfhost_parser_parser_objectFieldPatternText(name string, optional bool, value string) string {
	return func() string {
		if optional {
			return name + "?:" + value
		}
		return name + ":" + value
	}()
}

func selfhost_parser_parser_parseMapEntryPattern(state ParserState, out string) ExprStep {
	key := selfhost_parser_parser_parseMapPatternKey(state)
	optional := selfhost_parser_parser_parserMatch(key.state, TokenKind_Question)
	colon := selfhost_parser_parser_parserConsume(optional.state, TokenKind_Colon, "expected ':' after map pattern key")
	value := selfhost_parser_parser_parsePattern(selfhost_parser_parser_parserSkipNewlines(colon.state))
	entry := func() string {
		if optional.ok {
			return key.expr.text + "?:" + value.expr.text
		}
		return key.expr.text + ":" + value.expr.text
	}()
	return selfhost_parser_parser_parseMapOrObjectPatternAfterPart(value.state, selfhost_parser_parser_appendPatternPart(out, entry))
}

func selfhost_parser_parser_parseMapPatternKey(state ParserState) ExprStep {
	return selfhost_parser_parser_parseRangeBoundPattern(state)
}

func selfhost_parser_parser_parseMapOrObjectPatternAfterPart(state ParserState, out string) ExprStep {
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(state), TokenKind_Comma)
	return func() ExprStep {
		if comma.ok {
			return selfhost_parser_parser_parseMapOrObjectPatternParts(comma.state, out)
		}
		return ExprStep{state: comma.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, selfhost_parser_parser_parserPeek(comma.state)), out)}
	}()
}

func selfhost_parser_parser_parseTupleOrGroupedPattern(state ParserState) ExprStep {
	open := selfhost_parser_parser_parserConsume(state, TokenKind_LParen, "expected '(' before pattern")
	current := selfhost_parser_parser_parserSkipNewlines(open.state)
	empty := selfhost_parser_parser_parserCheck(current, TokenKind_RParen)
	return func() ExprStep {
		if empty {
			return selfhost_parser_parser_finishEmptyTuplePattern(open.token, current)
		}
		return selfhost_parser_parser_parseTupleOrGroupedPatternFirst(open.token, current)
	}()
}

func selfhost_parser_parser_finishEmptyTuplePattern(open Token, state ParserState) ExprStep {
	close := selfhost_parser_parser_parserConsume(state, TokenKind_RParen, "expected ')' after tuple pattern")
	return ExprStep{state: close.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, open), "()")}
}

func selfhost_parser_parser_parseTupleOrGroupedPatternFirst(open Token, state ParserState) ExprStep {
	first := selfhost_parser_parser_parsePattern(state)
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(first.state), TokenKind_Comma)
	return func() ExprStep {
		if comma.ok {
			return selfhost_parser_parser_parseTuplePatternRest(open, comma.state, first.expr.text)
		}
		return selfhost_parser_parser_finishGroupedPattern(first)
	}()
}

func selfhost_parser_parser_finishGroupedPattern(first ExprStep) ExprStep {
	close := selfhost_parser_parser_parserConsume(selfhost_parser_parser_parserSkipNewlines(first.state), TokenKind_RParen, "expected ')' after pattern")
	return ExprStep{state: close.state, expr: first.expr}
}

func selfhost_parser_parser_parseTuplePatternRest(open Token, state ParserState, out string) ExprStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	done := selfhost_parser_parser_parserCheck(current, TokenKind_RParen) || selfhost_parser_parser_parserCheck(current, TokenKind_EOF)
	return func() ExprStep {
		if done {
			return selfhost_parser_parser_finishTuplePattern(open, current, out)
		}
		return selfhost_parser_parser_parseTuplePatternPart(open, current, out)
	}()
}

func selfhost_parser_parser_parseTuplePatternPart(open Token, state ParserState, out string) ExprStep {
	value := selfhost_parser_parser_parsePattern(state)
	next := selfhost_parser_parser_appendPatternPart(out, value.expr.text)
	comma := selfhost_parser_parser_parserMatch(selfhost_parser_parser_parserSkipNewlines(value.state), TokenKind_Comma)
	return func() ExprStep {
		if comma.ok {
			return selfhost_parser_parser_parseTuplePatternRest(open, comma.state, next)
		}
		return selfhost_parser_parser_finishTuplePattern(open, comma.state, next)
	}()
}

func selfhost_parser_parser_finishTuplePattern(open Token, state ParserState, out string) ExprStep {
	close := selfhost_parser_parser_parserConsume(selfhost_parser_parser_parserSkipNewlines(state), TokenKind_RParen, "expected ')' after tuple pattern")
	return ExprStep{state: close.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, open), "("+out+")")}
}

func selfhost_parser_parser_appendPatternPart(out string, part string) string {
	return func() string {
		if out == "" {
			return part
		}
		return out + "," + part
	}()
}

func selfhost_parser_parser_parsePatternError(state ParserState) ExprStep {
	token := selfhost_parser_parser_parserPeek(state)
	step := selfhost_parser_parser_parserAdvance(selfhost_parser_parser_parserErrorAt(state, token, "expected pattern"))
	return ExprStep{state: step.state, expr: selfhost_parser_parser_withText(selfhost_parser_parser_node(ExprKind_Pattern, token), "_")}
}

func selfhost_parser_parser_functionToExpr(fn ParsedFunction) ParsedExpr {
	return selfhost_parser_parser_makeExpr(func() ExprKind {
		if fn.private {
			return ExprKind_PrivateMethod
		}
		return ExprKind_Method
	}(), typeRefToString(fn.returnType), fn.name, "", "=>", fn.params, []ParsedExpr{fn.body}, fn.line, fn.column)
}

func selfhost_parser_parser_calleeText(expr ParsedExpr) string {
	return func() string {
		if expr.name != "" {
			return expr.name
		}
		return func() string {
			if expr.value != "" {
				return expr.value
			}
			return expr.text
		}()
	}()
}

func selfhost_parser_parser_precedence(kind TokenKind) int {
	return func() int {
		if kind == TokenKind_QuestionQuestion || kind == TokenKind_OrOr {
			return 1
		}
		return func() int {
			if kind == TokenKind_AndAnd {
				return 2
			}
			return func() int {
				if kind == TokenKind_BitOr {
					return 3
				}
				return func() int {
					if kind == TokenKind_BitXor {
						return 4
					}
					return func() int {
						if kind == TokenKind_BitAnd {
							return 5
						}
						return func() int {
							if kind == TokenKind_EqualEqual || kind == TokenKind_BangEqual {
								return 6
							}
							return func() int {
								if kind == TokenKind_DotDotEqual || kind == TokenKind_Less || kind == TokenKind_LessEqual || kind == TokenKind_Greater || kind == TokenKind_GreaterEqual {
									return 7
								}
								return func() int {
									if kind == TokenKind_ShiftLeft || kind == TokenKind_ShiftRight || kind == TokenKind_UnsignedShiftRight {
										return 8
									}
									return func() int {
										if kind == TokenKind_Plus || kind == TokenKind_Minus {
											return 9
										}
										return func() int {
											if kind == TokenKind_Star || kind == TokenKind_Slash || kind == TokenKind_Percent {
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

func selfhost_parser_parser_parseAnnotations(state ParserState) AnnotationListStep {
	return selfhost_parser_parser_parseAnnotationsLoop(state, selfhost_parser_parser_emptyAnnotations())
}

func selfhost_parser_parser_parseAnnotationsLoop(state ParserState, annotations []ParsedAnnotation) AnnotationListStep {
	current := selfhost_parser_parser_parserSkipNewlines(state)
	return func() AnnotationListStep {
		if selfhost_parser_parser_looksLikeAnnotation(current) {
			return selfhost_parser_parser_parseAnnotationsNext(current, annotations)
		}
		return AnnotationListStep{state: current, annotations: annotations}
	}()
}

func selfhost_parser_parser_parseAnnotationsNext(state ParserState, annotations []ParsedAnnotation) AnnotationListStep {
	step := selfhost_parser_parser_parseAnnotation(state)
	return selfhost_parser_parser_parseAnnotationsLoop(step.state, func() []ParsedAnnotation {
		__rune_spread_out := []ParsedAnnotation{}
		__rune_spread_out = append(__rune_spread_out, annotations...)
		__rune_spread_out = append(__rune_spread_out, step.annotation)
		return __rune_spread_out
	}())
}

func selfhost_parser_parser_looksLikeAnnotation(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_Hash) || selfhost_parser_parser_parserCheck(state, TokenKind_At) && selfhost_parser_parser_parserCheckNext(state, TokenKind_Ident)
}

func selfhost_parser_parser_parseAnnotation(state ParserState) AnnotationStep {
	marker := selfhost_parser_parser_parserAdvance(state)
	first := selfhost_parser_parser_parserConsume(marker.state, TokenKind_Ident, "expected annotation name")
	dot := selfhost_parser_parser_parserMatch(first.state, TokenKind_Dot)
	name := func() TokenStep {
		if dot.ok {
			return selfhost_parser_parser_parserConsume(dot.state, TokenKind_Ident, "expected annotation function name after '.'")
		}
		return first
	}()
	open := selfhost_parser_parser_parserMatch(name.state, TokenKind_LParen)
	args := func() ExprStep {
		if open.ok {
			return selfhost_parser_parser_parseArgumentList(selfhost_parser_parser_parserSkipNewlines(open.state), []ParsedExpr{}, TokenKind_RParen)
		}
		return ExprStep{state: name.state, expr: selfhost_parser_parser_makeExpr(ExprKind_Args, "", "", "", "", []ParsedParam{}, []ParsedExpr{}, 0, 0)}
	}()
	close := func() TokenStep {
		if open.ok {
			return selfhost_parser_parser_parserConsume(args.state, TokenKind_RParen, "expected ')' after annotation arguments")
		}
		return TokenStep{state: args.state, token: name.token}
	}()
	return AnnotationStep{state: close.state, annotation: ParsedAnnotation{marker: marker.token.lexeme, module: func() string {
		if dot.ok {
			return first.token.lexeme
		}
		return ""
	}(), name: name.token.lexeme, args: args.expr.children, line: marker.token.line, column: marker.token.column}}
}

func selfhost_parser_parser_skipBalanced(state ParserState, openKind TokenKind, closeKind TokenKind, depth int) ParserState {
	return func() ParserState {
		if selfhost_parser_parser_parserCheck(state, TokenKind_EOF) || depth <= 0 {
			return state
		}
		return selfhost_parser_parser_skipBalancedStep(state, openKind, closeKind, depth)
	}()
}

func selfhost_parser_parser_skipBalancedStep(state ParserState, openKind TokenKind, closeKind TokenKind, depth int) ParserState {
	token := selfhost_parser_parser_parserPeek(state)
	step := selfhost_parser_parser_parserAdvance(state)
	nextDepth := func() int {
		if token.kind == openKind {
			return depth + 1
		}
		return func() int {
			if token.kind == closeKind {
				return depth - 1
			}
			return depth
		}()
	}()
	return selfhost_parser_parser_skipBalanced(step.state, openKind, closeKind, nextDepth)
}

func selfhost_parser_parser_questionIsPostfixUnwrap(state ParserState) bool {
	next := selfhost_parser_parser_parserKindAt(state, state.current+1)
	return next == TokenKind_EOF || next == TokenKind_Newline || next == TokenKind_RParen || next == TokenKind_RBracket || next == TokenKind_RBrace || next == TokenKind_Comma
}

func selfhost_parser_parser_questionQuestionIsPostfixUnwrap(state ParserState) bool {
	next := selfhost_parser_parser_parserKindAt(state, state.current+1)
	return next == TokenKind_EOF || next == TokenKind_Newline || next == TokenKind_RParen || next == TokenKind_RBracket || next == TokenKind_RBrace || next == TokenKind_Comma
}

func selfhost_parser_parser_looksLikeTypeDecl(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_Ident) && selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, selfhost_parser_parser_skipGenericNamesAt(state, state.current+1))) == TokenKind_Colon
}

func selfhost_parser_parser_looksLikeFunctionDecl(state ParserState) bool {
	start := func() int {
		if selfhost_parser_parser_parserCheck(state, TokenKind_Tilde) {
			return selfhost_parser_parser_skipNewlinesAt(state, state.current+1)
		}
		return state.current
	}()
	return selfhost_parser_parser_parserKindAt(state, start) == TokenKind_Ident && selfhost_parser_parser_looksLikeFunctionAfterName(state, selfhost_parser_parser_skipGenericNamesAt(state, start+1))
}

func selfhost_parser_parser_looksLikeMacroFunctionDecl(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_Hash) && selfhost_parser_parser_looksLikeFunctionDecl(selfhost_parser_parser_stateAt(state, state.current+1))
}

func selfhost_parser_parser_looksLikeStaticFunctionDecl(state ParserState) bool {
	markerState := func() ParserState {
		if selfhost_parser_parser_parserCheck(state, TokenKind_DoubleColon) {
			return selfhost_parser_parser_stateAt(state, state.current+1)
		}
		return func() ParserState {
			if selfhost_parser_parser_parserCheck(state, TokenKind_Ident) && selfhost_parser_parser_parserPeek(state).lexeme == "static" {
				return selfhost_parser_parser_stateAt(state, state.current+1)
			}
			return state
		}()
	}()
	return markerState.current != state.current && selfhost_parser_parser_looksLikeFunctionDecl(selfhost_parser_parser_parserSkipNewlines(markerState))
}

func selfhost_parser_parser_looksLikeFunctionAfterName(state ParserState, index int) bool {
	return selfhost_parser_parser_parserKindAt(state, index) == TokenKind_LParen && selfhost_parser_parser_looksLikeFunctionAfterParams(state, selfhost_parser_parser_skipBalancedAt(state, index, TokenKind_LParen, TokenKind_RParen))
}

func selfhost_parser_parser_looksLikeFunctionAfterParams(state ParserState, index int) bool {
	afterParams := selfhost_parser_parser_skipNewlinesAt(state, index)
	afterReturn := func() int {
		if selfhost_parser_parser_parserKindAt(state, afterParams) == TokenKind_Arrow {
			return selfhost_parser_parser_skipTypeNameTokensAt(state, afterParams+1)
		}
		return afterParams
	}()
	return selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, afterReturn)) == TokenKind_FatArrow
}

func selfhost_parser_parser_looksLikeLambda(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_LParen) && selfhost_parser_parser_looksLikeLambdaAfterParams(state, selfhost_parser_parser_skipBalancedAt(state, state.current, TokenKind_LParen, TokenKind_RParen))
}

func selfhost_parser_parser_looksLikeLambdaAfterParams(state ParserState, index int) bool {
	afterParams := selfhost_parser_parser_skipNewlinesAt(state, index)
	afterReturn := func() int {
		if selfhost_parser_parser_parserKindAt(state, afterParams) == TokenKind_Arrow {
			return selfhost_parser_parser_skipTypeNameTokensAt(state, afterParams+1)
		}
		return afterParams
	}()
	return selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, afterReturn)) == TokenKind_FatArrow
}

func selfhost_parser_parser_skipAnnotationsAt(state ParserState, index int) int {
	current := selfhost_parser_parser_skipNewlinesAt(state, index)
	annotation := selfhost_parser_parser_looksLikeAnnotationAt(state, current)
	return func() int {
		if annotation {
			return selfhost_parser_parser_skipAnnotationsAt(state, selfhost_parser_parser_skipAnnotationAt(state, current))
		}
		return current
	}()
}

func selfhost_parser_parser_looksLikeAnnotationAt(state ParserState, index int) bool {
	kind := selfhost_parser_parser_parserKindAt(state, index)
	return func() bool {
		switch {
		case kind == TokenKind_Hash:
			return selfhost_parser_parser_parserKindAt(state, index+1) == TokenKind_Ident
		case kind == TokenKind_At:
			return selfhost_parser_parser_parserKindAt(state, index+1) == TokenKind_Ident
		default:
			return false
		}
	}()
}

func selfhost_parser_parser_skipAnnotationAt(state ParserState, index int) int {
	afterName := selfhost_parser_parser_skipAnnotationNameAt(state, index+2)
	afterNewlines := selfhost_parser_parser_skipNewlinesAt(state, afterName)
	hasArgs := selfhost_parser_parser_parserKindAt(state, afterNewlines) == TokenKind_LParen
	return func() int {
		if hasArgs {
			return selfhost_parser_parser_skipNewlinesAt(state, selfhost_parser_parser_skipBalancedAt(state, afterNewlines, TokenKind_LParen, TokenKind_RParen))
		}
		return afterNewlines
	}()
}

func selfhost_parser_parser_skipAnnotationNameAt(state ParserState, index int) int {
	qualified := selfhost_parser_parser_parserKindAt(state, index) == TokenKind_Dot && selfhost_parser_parser_parserKindAt(state, index+1) == TokenKind_Ident
	return func() int {
		if qualified {
			return index + 2
		}
		return index
	}()
}

func selfhost_parser_parser_looksLikeEnumMember(state ParserState) bool {
	afterAnnotations := selfhost_parser_parser_skipAnnotationsAt(state, state.current)
	plus := selfhost_parser_parser_parserKindAt(state, afterAnnotations) == TokenKind_Plus
	start := func() int {
		if plus {
			return selfhost_parser_parser_skipNewlinesAt(state, afterAnnotations+1)
		}
		return afterAnnotations
	}()
	token := selfhost_parser_parser_parserTokenAt(state, start)
	next := selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, start+1))
	return token.kind == TokenKind_Ident && (next == TokenKind_Assign || next == TokenKind_LParen && selfhost_parser_parser_startsWithUpper(token.lexeme) || next != TokenKind_Colon && next != TokenKind_LParen)
}

func selfhost_parser_parser_startsWithUpper(name string) bool {
	return func() bool {
		if len([]rune(name)) > 0 {
			return []rune(name)[0] >= 'A' && []rune(name)[0] <= 'Z'
		}
		return false
	}()
}

func selfhost_parser_parser_looksLikeObjectDestructureDecl(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_LBrace) && selfhost_parser_parser_scanObjectDestructureDecl(state, selfhost_parser_parser_skipNewlinesAt(state, state.current+1))
}

func selfhost_parser_parser_scanObjectDestructureDecl(state ParserState, index int) bool {
	return func() bool {
		if selfhost_parser_parser_parserKindAt(state, index) == TokenKind_RBrace {
			return false
		}
		return selfhost_parser_parser_scanObjectDestructureField(state, index)
	}()
}

func selfhost_parser_parser_scanObjectDestructureField(state ParserState, index int) bool {
	return func() bool {
		if selfhost_parser_parser_parserKindAt(state, index) != TokenKind_Ident {
			return false
		}
		return selfhost_parser_parser_scanObjectDestructureAfterField(state, index+1)
	}()
}

func selfhost_parser_parser_scanObjectDestructureAfterField(state ParserState, index int) bool {
	return func() bool {
		if selfhost_parser_parser_parserKindAt(state, index) == TokenKind_Colon {
			return selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, index+1)) == TokenKind_Ident && selfhost_parser_parser_scanObjectDestructureAfterName(state, selfhost_parser_parser_skipNewlinesAt(state, index+1)+1)
		}
		return selfhost_parser_parser_scanObjectDestructureAfterName(state, index)
	}()
}

func selfhost_parser_parser_scanObjectDestructureAfterName(state ParserState, index int) bool {
	current := selfhost_parser_parser_skipNewlinesAt(state, index)
	kind := selfhost_parser_parser_parserKindAt(state, current)
	return func() bool {
		if kind == TokenKind_Comma {
			return selfhost_parser_parser_scanObjectDestructureAfterComma(state, selfhost_parser_parser_skipNewlinesAt(state, current+1))
		}
		return func() bool {
			if kind == TokenKind_RBrace {
				return selfhost_parser_parser_scanObjectDestructureAfterClose(state, current+1)
			}
			return false
		}()
	}()
}

func selfhost_parser_parser_scanObjectDestructureAfterComma(state ParserState, index int) bool {
	return func() bool {
		if selfhost_parser_parser_parserKindAt(state, index) == TokenKind_RBrace {
			return selfhost_parser_parser_scanObjectDestructureAfterClose(state, index+1)
		}
		return selfhost_parser_parser_scanObjectDestructureField(state, index)
	}()
}

func selfhost_parser_parser_scanObjectDestructureAfterClose(state ParserState, index int) bool {
	afterClose := selfhost_parser_parser_skipNewlinesAt(state, index)
	return selfhost_parser_parser_parserKindAt(state, afterClose) == TokenKind_Declare || selfhost_parser_parser_parserKindAt(state, afterClose) == TokenKind_MutDeclare
}

func selfhost_parser_parser_looksLikePatternBranch(state ParserState) bool {
	return selfhost_parser_parser_tokensLookLikePatternBranch(state, selfhost_parser_parser_skipNewlinesAt(state, state.current))
}

func selfhost_parser_parser_looksLikePatternBlockAfterSubject(state ParserState) bool {
	return selfhost_parser_parser_parserCheck(state, TokenKind_LBrace) && selfhost_parser_parser_tokensLookLikePatternBranch(state, selfhost_parser_parser_skipNewlinesAt(state, state.current+1))
}

func selfhost_parser_parser_tokensLookLikePatternBranch(state ParserState, index int) bool {
	afterPattern := selfhost_parser_parser_skipPatternLookahead(state, index)
	return afterPattern >= 0 && selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, afterPattern)) == TokenKind_FatArrow
}

func selfhost_parser_parser_skipPatternLookahead(state ParserState, index int) int {
	return selfhost_parser_parser_skipOrPatternLookahead(state, selfhost_parser_parser_skipSinglePatternLookahead(state, index))
}

func selfhost_parser_parser_skipOrPatternLookahead(state ParserState, index int) int {
	return func() int {
		if index >= 0 && selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, index)) == TokenKind_BitOr {
			return selfhost_parser_parser_skipOrPatternLookahead(state, selfhost_parser_parser_skipSinglePatternLookahead(state, selfhost_parser_parser_skipNewlinesAt(state, index)+1))
		}
		return index
	}()
}

func selfhost_parser_parser_skipAliasPatternLookahead(state ParserState, index int) int {
	return func() int {
		if index >= 0 && selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, index)) == TokenKind_At && selfhost_parser_parser_parserKindAt(state, selfhost_parser_parser_skipNewlinesAt(state, index)+1) == TokenKind_Ident {
			return selfhost_parser_parser_skipNewlinesAt(state, index) + 2
		}
		return index
	}()
}

func selfhost_parser_parser_skipSinglePatternLookahead(state ParserState, index int) int {
	kind := selfhost_parser_parser_parserKindAt(state, index)
	after := func() int {
		if kind == TokenKind_Underscore || kind == TokenKind_Int || kind == TokenKind_Double || kind == TokenKind_BigInt || kind == TokenKind_String || kind == TokenKind_Char {
			return index + 1
		}
		return func() int {
			if kind == TokenKind_Ident {
				return selfhost_parser_parser_skipIdentifierPatternLookahead(state, index)
			}
			return func() int {
				if kind == TokenKind_Less || kind == TokenKind_LessEqual || kind == TokenKind_Greater || kind == TokenKind_GreaterEqual {
					return selfhost_parser_parser_skipComparePatternLookahead(state, index+1)
				}
				return func() int {
					if kind == TokenKind_LParen {
						return selfhost_parser_parser_skipBalancedAt(state, index, TokenKind_LParen, TokenKind_RParen)
					}
					return func() int {
						if kind == TokenKind_LBracket {
							return selfhost_parser_parser_skipBalancedAt(state, index, TokenKind_LBracket, TokenKind_RBracket)
						}
						return func() int {
							if kind == TokenKind_LBrace {
								return selfhost_parser_parser_skipBalancedAt(state, index, TokenKind_LBrace, TokenKind_RBrace)
							}
							return -1
						}()
					}()
				}()
			}()
		}()
	}()
	return selfhost_parser_parser_skipAliasPatternLookahead(state, selfhost_parser_parser_skipRangePatternLookahead(state, after))
}

func selfhost_parser_parser_skipRangePatternLookahead(state ParserState, index int) int {
	return func() int {
		if index >= 0 && selfhost_parser_parser_parserKindAt(state, index) == TokenKind_DotDotEqual {
			return selfhost_parser_parser_skipRangePatternEnd(state, index+1)
		}
		return func() int {
			if index >= 0 && selfhost_parser_parser_parserKindAt(state, index) == TokenKind_DotDotLess {
				return selfhost_parser_parser_skipRangePatternEnd(state, index+1)
			}
			return func() int {
				if index >= 0 && selfhost_parser_parser_parserKindAt(state, index) == TokenKind_DotDot && selfhost_parser_parser_parserKindAt(state, index+1) == TokenKind_Less {
					return selfhost_parser_parser_skipRangePatternEnd(state, index+2)
				}
				return index
			}()
		}()
	}()
}

func selfhost_parser_parser_skipRangePatternEnd(state ParserState, index int) int {
	kind := selfhost_parser_parser_parserKindAt(state, index)
	return func() int {
		if kind == TokenKind_Underscore {
			return index + 1
		}
		return func() int {
			if kind == TokenKind_Int || kind == TokenKind_Double || kind == TokenKind_BigInt || kind == TokenKind_String || kind == TokenKind_Char || kind == TokenKind_Ident {
				return selfhost_parser_parser_skipIdentifierRangeEnd(state, index)
			}
			return -1
		}()
	}()
}

func selfhost_parser_parser_skipIdentifierRangeEnd(state ParserState, index int) int {
	return func() int {
		if selfhost_parser_parser_parserKindAt(state, index) == TokenKind_Ident && selfhost_parser_parser_parserKindAt(state, index+1) == TokenKind_Dot && selfhost_parser_parser_parserKindAt(state, index+2) == TokenKind_Ident {
			return index + 3
		}
		return index + 1
	}()
}

func selfhost_parser_parser_skipIdentifierPatternLookahead(state ParserState, index int) int {
	return func() int {
		if selfhost_parser_parser_parserKindAt(state, index+1) == TokenKind_LParen {
			return selfhost_parser_parser_skipBalancedAt(state, index+1, TokenKind_LParen, TokenKind_RParen)
		}
		return func() int {
			if selfhost_parser_parser_parserKindAt(state, index+1) == TokenKind_Dot && selfhost_parser_parser_parserKindAt(state, index+2) == TokenKind_Ident {
				return index + 3
			}
			return index + 1
		}()
	}()
}

func selfhost_parser_parser_skipComparePatternLookahead(state ParserState, index int) int {
	kind := selfhost_parser_parser_parserKindAt(state, index)
	return func() int {
		if kind == TokenKind_Int || kind == TokenKind_Double || kind == TokenKind_BigInt || kind == TokenKind_String || kind == TokenKind_Char || kind == TokenKind_Ident {
			return index + 1
		}
		return -1
	}()
}

func selfhost_parser_parser_looksLikeMapLiteralBody(state ParserState) bool {
	start := selfhost_parser_parser_skipNewlinesAt(state, state.current+1)
	first := selfhost_parser_parser_parserTokenAt(state, start)
	return func() bool {
		if first.kind == TokenKind_EOF || first.kind == TokenKind_RBrace {
			return false
		}
		return func() bool {
			if first.kind == TokenKind_Ident && selfhost_parser_parser_isLiteralIdentifier(first.lexeme) == false {
				return false
			}
			return selfhost_parser_parser_scanMapLiteralColon(state, start, 0)
		}()
	}()
}

func selfhost_parser_parser_scanMapLiteralColon(state ParserState, index int, depth int) bool {
	kind := selfhost_parser_parser_parserKindAt(state, index)
	return func() bool {
		if kind == TokenKind_EOF {
			return false
		}
		return func() bool {
			if kind == TokenKind_LParen || kind == TokenKind_LBracket || kind == TokenKind_LBrace {
				return selfhost_parser_parser_scanMapLiteralColon(state, index+1, depth+1)
			}
			return func() bool {
				if kind == TokenKind_RParen || kind == TokenKind_RBracket {
					return selfhost_parser_parser_scanMapLiteralColon(state, index+1, func() int {
						if depth > 0 {
							return depth - 1
						}
						return depth
					}())
				}
				return func() bool {
					if kind == TokenKind_RBrace {
						return func() bool {
							if depth == 0 {
								return false
							}
							return selfhost_parser_parser_scanMapLiteralColon(state, index+1, depth-1)
						}()
					}
					return func() bool {
						if (kind == TokenKind_Newline || kind == TokenKind_Question || kind == TokenKind_FatArrow) && depth == 0 {
							return false
						}
						return func() bool {
							if kind == TokenKind_Colon && depth == 0 {
								return true
							}
							return selfhost_parser_parser_scanMapLiteralColon(state, index+1, depth)
						}()
					}()
				}()
			}()
		}()
	}()
}

func selfhost_parser_parser_isLiteralIdentifier(name string) bool {
	return name == "true" || name == "false" || name == "null"
}

func selfhost_parser_parser_parserPatternLooksLikeObjectField(state ParserState) bool {
	token := selfhost_parser_parser_parserPeek(state)
	return func() bool {
		switch {
		case token.kind == TokenKind_Ident:
			return selfhost_parser_parser_parserPatternIdentLooksLikeObjectField(state, token.lexeme)
		default:
			return false
		}
	}()
}

func selfhost_parser_parser_parserPatternIdentLooksLikeObjectField(state ParserState, name string) bool {
	spread := selfhost_parser_parser_isPatternSpreadIdentifier(name)
	return func() bool {
		if spread {
			return false
		}
		return selfhost_parser_parser_parserPatternObjectFieldTail(state)
	}()
}

func selfhost_parser_parser_parserPatternObjectFieldTail(state ParserState) bool {
	kind := selfhost_parser_parser_parserKindAt(state, state.current+1)
	return func() bool {
		switch {
		case kind == TokenKind_Colon:
			return true
		case kind == TokenKind_Question:
			return true
		case kind == TokenKind_Comma:
			return true
		case kind == TokenKind_RBrace:
			return true
		default:
			return false
		}
	}()
}

func selfhost_parser_parser_isPatternSpreadIdentifier(name string) bool {
	empty := len([]rune(name)) == 0
	return func() bool {
		if empty {
			return false
		}
		return selfhost_parser_parser_isUpperAsciiLetter([]rune(name)[0])
	}()
}

func selfhost_parser_parser_isUpperAsciiLetter(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}

func selfhost_parser_parser_skipNewlinesAt(state ParserState, index int) int {
	return func() int {
		if selfhost_parser_parser_parserKindAt(state, index) == TokenKind_Newline {
			return selfhost_parser_parser_skipNewlinesAt(state, index+1)
		}
		return index
	}()
}

func selfhost_parser_parser_skipGenericNamesAt(state ParserState, index int) int {
	return func() int {
		if selfhost_parser_parser_parserKindAt(state, index) == TokenKind_LBracket {
			return selfhost_parser_parser_skipBalancedAt(state, index, TokenKind_LBracket, TokenKind_RBracket)
		}
		return index
	}()
}

func selfhost_parser_parser_skipBalancedAt(state ParserState, index int, openKind TokenKind, closeKind TokenKind) int {
	return selfhost_parser_parser_skipBalancedAtLoop(state, index, openKind, closeKind, 0)
}

func selfhost_parser_parser_skipBalancedAtLoop(state ParserState, index int, openKind TokenKind, closeKind TokenKind, depth int) int {
	kind := selfhost_parser_parser_parserKindAt(state, index)
	return func() int {
		if kind == TokenKind_EOF {
			return index
		}
		return func() int {
			if kind == openKind {
				return selfhost_parser_parser_skipBalancedAtLoop(state, index+1, openKind, closeKind, depth+1)
			}
			return func() int {
				if kind == closeKind {
					return func() int {
						if depth <= 1 {
							return index + 1
						}
						return selfhost_parser_parser_skipBalancedAtLoop(state, index+1, openKind, closeKind, depth-1)
					}()
				}
				return selfhost_parser_parser_skipBalancedAtLoop(state, index+1, openKind, closeKind, depth)
			}()
		}()
	}()
}

func selfhost_parser_parser_skipTypeNameTokensAt(state ParserState, index int) int {
	return selfhost_parser_parser_skipTypeNameTokensAtLoop(state, index, 0)
}

func selfhost_parser_parser_skipTypeNameTokensAtLoop(state ParserState, index int, depth int) int {
	kind := selfhost_parser_parser_parserKindAt(state, index)
	return func() int {
		if kind == TokenKind_Ident || kind == TokenKind_Comma || kind == TokenKind_Colon || kind == TokenKind_Question || kind == TokenKind_Arrow || kind == TokenKind_At || kind == TokenKind_Dot {
			return selfhost_parser_parser_skipTypeNameTokensAtLoop(state, index+1, depth)
		}
		return func() int {
			if kind == TokenKind_LBracket || kind == TokenKind_LParen {
				return selfhost_parser_parser_skipTypeNameTokensAtLoop(state, index+1, depth+1)
			}
			return func() int {
				if kind == TokenKind_RBracket || kind == TokenKind_RParen {
					return func() int {
						if depth == 0 {
							return index
						}
						return selfhost_parser_parser_skipTypeNameTokensAtLoop(state, index+1, depth-1)
					}()
				}
				return index
			}()
		}()
	}()
}

func lower(source string) IRFile {
	return lowerParsed(parse(source))
}

func lowerWithSourcePath(source string, sourcePath string) IRFile {
	return withIRFileSourcePath(lower(source), sourcePath)
}

func lowerParsed(file ParsedFile) IRFile {
	out := IRFile{imports: []IRImport{}, tsImports: []IRTSImport{}, structs: []IRStructType{}, enums: []IREnumType{}, constants: []IRConst{}, functions: []IRFunction{}, tests: []IRTest{}, errors: file.errors}
	for _, importDecl := range file.imports {
		_ = importDecl
		func() int {
			out.imports = append(out.imports, selfhost_ir_ir_lowerImport(importDecl))
			return len(out.imports)
		}()
	}
	for _, constant := range file.constants {
		_ = constant
		func() int {
			out.constants = append(out.constants, selfhost_ir_ir_lowerConst(constant))
			return len(out.constants)
		}()
	}
	for _, typeDecl := range file.types {
		_ = typeDecl
		out = selfhost_ir_ir_lowerTypeInto(out, typeDecl)
	}
	for _, fn := range file.functions {
		_ = fn
		func() int {
			out.functions = append(out.functions, selfhost_ir_ir_lowerFunction(fn))
			return len(out.functions)
		}()
	}
	for _, testDecl := range file.tests {
		_ = testDecl
		func() int { out.tests = append(out.tests, selfhost_ir_ir_lowerTest(testDecl)); return len(out.tests) }()
	}
	return out
}

func selfhost_ir_ir_lowerTypeInto(file IRFile, typeDecl ParsedType) IRFile {
	return func() IRFile {
		if typeDecl.enum {
			return selfhost_ir_ir_pushEnumType(file, typeDecl)
		}
		return selfhost_ir_ir_pushStructType(file, typeDecl)
	}()
}

func selfhost_ir_ir_pushStructType(file IRFile, typeDecl ParsedType) IRFile {
	file.structs = append(file.structs, selfhost_ir_ir_lowerStructType(typeDecl))
	return file
}

func selfhost_ir_ir_pushEnumType(file IRFile, typeDecl ParsedType) IRFile {
	file.enums = append(file.enums, selfhost_ir_ir_lowerEnumType(typeDecl))
	return file
}

func emptyIRExpr() IRExpr {
	return IRExpr{kind: ExprKind_Unknown, text: "", name: "", value: "", op: "", params: []IRParam{}, children: []IRExpr{}, line: 0, column: 0}
}

func emptyIRFunction() IRFunction {
	return IRFunction{name: "", private: false, static: false, routine: false, macro: false, receiverType: "", generics: []string{}, params: []IRParam{}, returnType: "", body: emptyIRExpr(), sourcePath: "", line: 0, column: 0}
}

func selfhost_ir_ir_lowerImport(importDecl ParsedImport) IRImport {
	return IRImport{path: importDecl.path, go_: importDecl.go_, module: importDecl.module, line: importDecl.line, column: importDecl.column}
}

func selfhost_ir_ir_lowerConst(constDecl ParsedConst) IRConst {
	return IRConst{name: constDecl.name, private: constDecl.private, typeName: typeRefToString(constDecl.typeRef), value: selfhost_ir_ir_lowerExpr(constDecl.value), line: constDecl.line, column: constDecl.column}
}

func selfhost_ir_ir_lowerParam(param ParsedParam) IRParam {
	return IRParam{name: param.name, typeName: typeRefToString(param.typeRef), line: param.line, column: param.column}
}

func selfhost_ir_ir_lowerField(field ParsedField) IRField {
	return IRField{name: field.name, private: field.private, typeName: typeRefToString(field.typeRef), jsonName: selfhost_ir_ir_lowerJsonFieldName(field), jsonIgnore: selfhost_ir_ir_lowerHasAnnotation(field.annotations, "#", "json", "ignore", 0), line: field.line, column: field.column}
}

func selfhost_ir_ir_lowerJsonFieldName(field ParsedField) string {
	return selfhost_ir_ir_lowerJsonFieldNameAt(field.annotations, field.name, 0)
}

func selfhost_ir_ir_lowerJsonFieldNameAt(annotations []ParsedAnnotation, fallback string, index int) string {
	done := index >= len(annotations)
	return func() string {
		if done {
			return fallback
		}
		return selfhost_ir_ir_lowerJsonFieldNameStep(annotations, fallback, index)
	}()
}

func selfhost_ir_ir_lowerJsonFieldNameStep(annotations []ParsedAnnotation, fallback string, index int) string {
	annotation := annotations[index]
	matched := annotation.marker == "#" && annotation.module == "json" && annotation.name == "name" && len(annotation.args) > 0
	return func() string {
		if matched {
			return selfhost_ir_ir_lowerAnnotationStringArg(annotation, 0, fallback)
		}
		return selfhost_ir_ir_lowerJsonFieldNameAt(annotations, fallback, index+1)
	}()
}

func selfhost_ir_ir_lowerHasAnnotation(annotations []ParsedAnnotation, marker string, module string, name string, index int) bool {
	done := index >= len(annotations)
	return func() bool {
		if done {
			return false
		}
		return selfhost_ir_ir_lowerHasAnnotationStep(annotations, marker, module, name, index)
	}()
}

func selfhost_ir_ir_lowerHasAnnotationStep(annotations []ParsedAnnotation, marker string, module string, name string, index int) bool {
	annotation := annotations[index]
	matched := annotation.marker == marker && annotation.module == module && annotation.name == name
	return func() bool {
		if matched {
			return true
		}
		return selfhost_ir_ir_lowerHasAnnotation(annotations, marker, module, name, index+1)
	}()
}

func selfhost_ir_ir_lowerAnnotationStringArg(annotation ParsedAnnotation, index int, fallback string) string {
	valid := index < len(annotation.args) && annotation.args[index].kind == ExprKind_String
	return func() string {
		if valid {
			return selfhost_ir_ir_lowerUnquoteString(annotation.args[index].value)
		}
		return fallback
	}()
}

func selfhost_ir_ir_lowerUnquoteString(raw string) string {
	quoted := len([]rune(raw)) >= 2
	return func() string {
		if quoted {
			return func() string { runes := []rune(raw); return string(runes[1 : len([]rune(raw))-1]) }()
		}
		return raw
	}()
}

func selfhost_ir_ir_lowerEnumMember(member ParsedEnumMember) IREnumMember {
	return IREnumMember{name: member.name, private: member.private, value: member.value, params: selfhost_ir_ir_lowerParams(member.params), line: member.line, column: member.column}
}

func selfhost_ir_ir_lowerFunction(fn ParsedFunction) IRFunction {
	return IRFunction{name: fn.name, private: fn.private, static: fn.static, routine: fn.routine, macro: selfhost_ir_ir_parsedFunctionCompileTimeOnly(fn), receiverType: fn.receiverType, generics: fn.generics, params: selfhost_ir_ir_lowerParams(fn.params), returnType: typeRefToString(fn.returnType), body: selfhost_ir_ir_lowerExpr(fn.body), sourcePath: "", line: fn.line, column: fn.column}
}

func withIRFileSourcePath(file IRFile, sourcePath string) IRFile {
	return IRFile{imports: file.imports, tsImports: file.tsImports, constants: file.constants, tests: file.tests, errors: file.errors, structs: selfhost_ir_ir_withIRStructSourcePaths(file.structs, sourcePath), enums: selfhost_ir_ir_withIREnumSourcePaths(file.enums, sourcePath), functions: selfhost_ir_ir_withIRFunctionSourcePaths(file.functions, sourcePath)}
}

func selfhost_ir_ir_withIRFunctionSourcePath(fn IRFunction, sourcePath string) IRFunction {
	return IRFunction{name: fn.name, private: fn.private, static: fn.static, routine: fn.routine, macro: fn.macro, receiverType: fn.receiverType, generics: fn.generics, params: fn.params, returnType: fn.returnType, body: fn.body, line: fn.line, column: fn.column, sourcePath: sourcePath}
}

func selfhost_ir_ir_withIRFunctionSourcePaths(functions []IRFunction, sourcePath string) []IRFunction {
	out := []IRFunction{}
	for _, fn := range functions {
		_ = fn
		func() int {
			out = append(out, selfhost_ir_ir_withIRFunctionSourcePath(fn, sourcePath))
			return len(out)
		}()
	}
	return out
}

func selfhost_ir_ir_withIRStructSourcePath(typeDecl IRStructType, sourcePath string) IRStructType {
	return IRStructType{name: typeDecl.name, private: typeDecl.private, generics: typeDecl.generics, fields: typeDecl.fields, line: typeDecl.line, column: typeDecl.column, methods: selfhost_ir_ir_withIRFunctionSourcePaths(typeDecl.methods, sourcePath), sourcePath: sourcePath}
}

func selfhost_ir_ir_withIRStructSourcePaths(structs []IRStructType, sourcePath string) []IRStructType {
	out := []IRStructType{}
	for _, typeDecl := range structs {
		_ = typeDecl
		func() int {
			out = append(out, selfhost_ir_ir_withIRStructSourcePath(typeDecl, sourcePath))
			return len(out)
		}()
	}
	return out
}

func selfhost_ir_ir_withIREnumSourcePath(typeDecl IREnumType, sourcePath string) IREnumType {
	return IREnumType{name: typeDecl.name, private: typeDecl.private, generics: typeDecl.generics, members: typeDecl.members, line: typeDecl.line, column: typeDecl.column, methods: selfhost_ir_ir_withIRFunctionSourcePaths(typeDecl.methods, sourcePath), sourcePath: sourcePath}
}

func selfhost_ir_ir_withIREnumSourcePaths(enums []IREnumType, sourcePath string) []IREnumType {
	out := []IREnumType{}
	for _, typeDecl := range enums {
		_ = typeDecl
		func() int {
			out = append(out, selfhost_ir_ir_withIREnumSourcePath(typeDecl, sourcePath))
			return len(out)
		}()
	}
	return out
}

func selfhost_ir_ir_parsedFunctionCompileTimeOnly(fn ParsedFunction) bool {
	return fn.macro || selfhost_ir_ir_typeRefIsSyntaxOnly(fn.returnType) || selfhost_ir_ir_paramsUseSyntaxOnly(fn.params, 0)
}

func selfhost_ir_ir_paramsUseSyntaxOnly(params []ParsedParam, index int) bool {
	return func() bool {
		if index >= len(params) {
			return false
		}
		return selfhost_ir_ir_typeRefIsSyntaxOnly(params[index].typeRef) || selfhost_ir_ir_paramsUseSyntaxOnly(params, index+1)
	}()
}

func selfhost_ir_ir_typeRefIsSyntaxOnly(typeRef ParsedTypeRef) bool {
	name := typeRefToString(typeRef)
	return strings.HasPrefix(name, "Syntax") || name == "MacroContext"
}

func selfhost_ir_ir_lowerStructType(typeDecl ParsedType) IRStructType {
	return IRStructType{name: typeDecl.name, private: typeDecl.private, generics: typeDecl.generics, fields: selfhost_ir_ir_lowerFields(typeDecl.fields), methods: selfhost_ir_ir_lowerFunctions(typeDecl.methods), sourcePath: "", line: typeDecl.line, column: typeDecl.column}
}

func selfhost_ir_ir_lowerEnumType(typeDecl ParsedType) IREnumType {
	return IREnumType{name: typeDecl.name, private: typeDecl.private, generics: typeDecl.generics, members: selfhost_ir_ir_lowerEnumMembers(typeDecl.members), methods: selfhost_ir_ir_lowerFunctions(typeDecl.methods), sourcePath: "", line: typeDecl.line, column: typeDecl.column}
}

func selfhost_ir_ir_lowerTest(testDecl ParsedTest) IRTest {
	return IRTest{name: testDecl.name, body: selfhost_ir_ir_lowerExpr(testDecl.body), line: testDecl.line, column: testDecl.column}
}

func selfhost_ir_ir_lowerExpr(expr ParsedExpr) IRExpr {
	children := selfhost_ir_ir_lowerExprs(expr.children)
	return IRExpr{kind: expr.kind, text: selfhost_ir_ir_inferIRExprText(expr, children), name: expr.name, value: expr.value, op: expr.op, params: selfhost_ir_ir_lowerParams(expr.params), children: children, line: expr.line, column: expr.column}
}

func selfhost_ir_ir_inferIRExprText(expr ParsedExpr, children []IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Binary:
			return selfhost_ir_ir_inferIRBinaryText(expr, children)
		case expr.kind == ExprKind_Call:
			return selfhost_ir_ir_inferIRCallText(children)
		case expr.kind == ExprKind_String:
			return "String"
		case expr.kind == ExprKind_Template:
			return "String"
		case expr.kind == ExprKind_XMLText:
			return "String"
		case expr.kind == ExprKind_XMLElement:
			return "HTMLElement"
		case expr.kind == ExprKind_Int:
			return "Int"
		case expr.kind == ExprKind_Double:
			return "Double"
		case expr.kind == ExprKind_BigInt:
			return "BigInt"
		case expr.kind == ExprKind_Char:
			return "Char"
		case expr.kind == ExprKind_Bool:
			return "Bool"
		case expr.kind == ExprKind_Null:
			return "Null"
		default:
			return func() string {
				if expr.text != "" {
					return expr.text
				}
				return ""
			}()
		}
	}()
}

func selfhost_ir_ir_inferIRBinaryText(expr ParsedExpr, children []IRExpr) string {
	return func() string {
		if expr.op == "??" {
			return selfhost_ir_ir_inferIRCoalesceText(children)
		}
		return ""
	}()
}

func selfhost_ir_ir_inferIRCoalesceText(children []IRExpr) string {
	return func() string {
		if len(children) < 2 {
			return ""
		}
		return func() string {
			if children[1].text != "" && children[1].text != "Null" {
				return children[1].text
			}
			return children[0].text
		}()
	}()
}

func selfhost_ir_ir_inferIRCallText(children []IRExpr) string {
	return func() string {
		if len(children) == 0 {
			return ""
		}
		return func() string {
			if children[0].kind == ExprKind_Selector {
				return selfhost_ir_ir_inferIRSelectorCallText(children[0], children)
			}
			return ""
		}()
	}()
}

func selfhost_ir_ir_inferIRSelectorCallText(selector IRExpr, children []IRExpr) string {
	return func() string {
		switch {
		case selector.name == "getOr":
			return func() string {
				if len(children) > 2 {
					return children[2].text
				}
				return ""
			}()
		case selector.name == "isEmpty":
			return "Bool"
		case (selector.name == "length") || (selector.name == "byteLength") || (selector.name == "size") || (selector.name == "push"):
			return "Int"
		case (selector.name == "includes") || (selector.name == "contains") || (selector.name == "startsWith") || (selector.name == "endsWith"):
			return "Bool"
		case (selector.name == "toString") || (selector.name == "trim") || (selector.name == "toUpperCase") || (selector.name == "toLowerCase"):
			return "String"
		default:
			return ""
		}
	}()
}

func selfhost_ir_ir_lowerParams(params []ParsedParam) []IRParam {
	out := []IRParam{}
	for _, param := range params {
		_ = param
		func() int { out = append(out, selfhost_ir_ir_lowerParam(param)); return len(out) }()
	}
	return out
}

func selfhost_ir_ir_lowerFields(fields []ParsedField) []IRField {
	out := []IRField{}
	for _, field := range fields {
		_ = field
		func() int { out = append(out, selfhost_ir_ir_lowerField(field)); return len(out) }()
	}
	return out
}

func selfhost_ir_ir_lowerEnumMembers(members []ParsedEnumMember) []IREnumMember {
	out := []IREnumMember{}
	for _, member := range members {
		_ = member
		func() int { out = append(out, selfhost_ir_ir_lowerEnumMember(member)); return len(out) }()
	}
	return out
}

func selfhost_ir_ir_lowerFunctions(functions []ParsedFunction) []IRFunction {
	out := []IRFunction{}
	for _, fn := range functions {
		_ = fn
		func() int { out = append(out, selfhost_ir_ir_lowerFunction(fn)); return len(out) }()
	}
	return out
}

func selfhost_ir_ir_lowerExprs(exprs []ParsedExpr) []IRExpr {
	out := []IRExpr{}
	for _, expr := range exprs {
		_ = expr
		func() int { out = append(out, selfhost_ir_ir_lowerExpr(expr)); return len(out) }()
	}
	return out
}

func enumValue(member IREnumMember, index int) string {
	return func() string {
		if member.value == "" {
			return compilerIntToString(index)
		}
		return member.value
	}()
}

func returnsValue(typeName string) bool {
	return typeName != "" && typeName != "Void"
}

func bigintLiteralDigits(value string) string {
	return func() string {
		if strings.HasSuffix(value, "n") {
			return func() string { runes := []rune(value); return string(runes[0 : len([]rune(value))-1]) }()
		}
		return value
	}()
}

func fileUsesPathBasenameFamily(file IRFile) bool {
	return fileUsesModuleCall(file, "path.basename") || fileUsesModuleCall(file, "path.extname") || fileUsesModuleCall(file, "path.dirname")
}

func fileUsesPathFamily(file IRFile) bool {
	return fileUsesPathBasenameFamily(file) || fileUsesModuleCall(file, "path.join") || fileUsesModuleCall(file, "path.normalize") || fileUsesModuleCall(file, "path.resolve") || fileUsesModuleCall(file, "path.relative") || fileUsesPathHelperFamily(file)
}

func fileUsesPathHelperFamily(file IRFile) bool {
	return fileUsesModuleCall(file, "path.joinParts") || fileUsesModuleCall(file, "path.appendPathPart") || fileUsesModuleCall(file, "path.normalizeParts") || fileUsesModuleCall(file, "path.normalizePart") || fileUsesModuleCall(file, "path.normalizeParent") || fileUsesModuleCall(file, "path.normalizePop") || fileUsesModuleCall(file, "path.normalizePush") || fileUsesModuleCall(file, "path.pathParts") || fileUsesModuleCall(file, "path.collectPathParts") || fileUsesModuleCall(file, "path.collectPathPart") || fileUsesModuleCall(file, "path.relativeFromParts") || fileUsesModuleCall(file, "path.relativeTail")
}

func moduleCallKey(expr IRExpr) string {
	return func() string {
		if expr.kind == ExprKind_Call && len(expr.children) > 0 {
			return moduleSelectorKey(expr.children[0])
		}
		return ""
	}()
}

func moduleSelectorKey(expr IRExpr) string {
	return func() string {
		if expr.kind == ExprKind_Selector && len(expr.children) > 0 {
			return func() string {
				if expr.children[0].kind == ExprKind_At {
					return expr.children[0].name + "." + expr.name
				}
				return ""
			}()
		}
		return ""
	}()
}

func genericInner(typeName string, base string) string {
	prefix := base + "["
	return func() string {
		if strings.HasPrefix(typeName, prefix) && strings.HasSuffix(typeName, "]") {
			return func() string {
				runes := []rune(typeName)
				return string(runes[len([]rune(prefix)) : len([]rune(typeName))-1])
			}()
		}
		return ""
	}()
}

func typeArg(args string, index int) string {
	parts := func() []string { parts := strings.Split(args, ","); return parts }()
	return func() string {
		if index < len(parts) {
			return strings.TrimSpace(parts[index])
		}
		return "Dynamic"
	}()
}

func mangleIdent(name string) string {
	mangled := strings.ReplaceAll((strings.ReplaceAll((strings.ReplaceAll(name, ".", "_")), "-", "_")), "@", "_")
	return func() string {
		if targetKeyword(mangled) {
			return mangled + "_"
		}
		return mangled
	}()
}

func targetKeyword(name string) bool {
	return func() bool {
		switch {
		case (name == "abstract") || (name == "alias") || (name == "and") || (name == "anyframe") || (name == "anytype") || (name == "as") || (name == "asm") || (name == "assert") || (name == "assume") || (name == "async") || (name == "atomic") || (name == "await") || (name == "break") || (name == "case") || (name == "catch") || (name == "chan") || (name == "class") || (name == "comptime") || (name == "const") || (name == "constructor") || (name == "continue") || (name == "declare") || (name == "default") || (name == "defer") || (name == "define") || (name == "delete") || (name == "derive") || (name == "do") || (name == "downcast") || (name == "dyn") || (name == "dynclass") || (name == "dynobj") || (name == "dynrec") || (name == "else") || (name == "enum") || (name == "enumview") || (name == "errdefer") || (name == "export") || (name == "extends") || (name == "extern") || (name == "extenum") || (name == "false") || (name == "fallthrough") || (name == "final") || (name == "finally") || (name == "fn") || (name == "fnalias") || (name == "for") || (name == "func") || (name == "function") || (name == "go") || (name == "goto") || (name == "guard") || (name == "if") || (name == "implements") || (name == "import") || (name == "in") || (name == "any") || (name == "bool") || (name == "byte") || (name == "comparable") || (name == "complex64") || (name == "complex128") || (name == "error") || (name == "float32") || (name == "float64") || (name == "int") || (name == "int8") || (name == "int16") || (name == "int32") || (name == "int64") || (name == "rune") || (name == "string") || (name == "uint") || (name == "uint8") || (name == "uint16") || (name == "uint32") || (name == "uint64") || (name == "uintptr") || (name == "include") || (name == "inherit") || (name == "instanceof") || (name == "interface") || (name == "is") || (name == "isnot") || (name == "lazy") || (name == "let") || (name == "letrec") || (name == "lexmatch") || (name == "local") || (name == "loop") || (name == "macro") || (name == "main") || (name == "map") || (name == "match") || (name == "member") || (name == "method") || (name == "mixin") || (name == "module") || (name == "move") || (name == "mut") || (name == "namespace") || (name == "new") || (name == "noasync") || (name == "nobreak") || (name == "noraise") || (name == "null") || (name == "opaque") || (name == "orelse") || (name == "override") || (name == "package") || (name == "priv") || (name == "private") || (name == "proof_assert") || (name == "proof_let") || (name == "protected") || (name == "pub") || (name == "public") || (name == "raise") || (name == "range") || (name == "readonly") || (name == "recur") || (name == "ref") || (name == "resume") || (name == "return") || (name == "sealed") || (name == "select") || (name == "static") || (name == "struct") || (name == "suberror") || (name == "super") || (name == "switch") || (name == "test") || (name == "this") || (name == "threadlocal") || (name == "throw") || (name == "trait") || (name == "traitalias") || (name == "true") || (name == "try") || (name == "type") || (name == "typealias") || (name == "typeof") || (name == "unsafe") || (name == "unreachable") || (name == "upcast") || (name == "use") || (name == "using") || (name == "var") || (name == "virtual") || (name == "void") || (name == "volatile") || (name == "where") || (name == "while") || (name == "with") || (name == "yield"):
			return true
		default:
			return false
		}
	}()
}

func compilerGoPackageImportPath(spec string) string {
	return func() string {
		if strings.HasPrefix(spec, "go:") {
			return func() string { runes := []rune(spec); return string(runes[3:len([]rune(spec))]) }()
		}
		return ""
	}()
}

func compilerGoPackageName(path string) string {
	slash := strings.LastIndex(path, "/")
	return func() string {
		if slash >= 0 {
			return func() string { runes := []rune(path); return string(runes[slash+1 : len([]rune(path))]) }()
		}
		return path
	}()
}

func compilerIRAtImportPath(expr IRExpr) string {
	return func() string {
		if expr.kind == ExprKind_At && expr.value != "" {
			return func() string { runes := []rune(expr.value); return string(runes[1 : len([]rune(expr.value))-1]) }()
		}
		return ""
	}()
}

func indent(level int) string {
	return func() string {
		if level <= 0 {
			return ""
		}
		return "  " + indent(level-1)
	}()
}

func line(level int, text string) string {
	return indent(level) + text + "\n"
}

func fileUsesUnwrap(file IRFile) bool {
	return selfhost_compiler_common_functionsUseUnwrap(file.functions, 0) || selfhost_compiler_common_structsUseUnwrap(file.structs, 0) || selfhost_compiler_common_enumsUseUnwrap(file.enums, 0) || selfhost_compiler_common_testsUseUnwrap(file.tests, 0)
}

func selfhost_compiler_common_functionsUseUnwrap(functions []IRFunction, index int) bool {
	return func() bool {
		if index >= len(functions) {
			return false
		}
		return func() bool {
			if functions[index].macro {
				return selfhost_compiler_common_functionsUseUnwrap(functions, index+1)
			}
			return exprUsesUnwrap(functions[index].body) || selfhost_compiler_common_functionsUseUnwrap(functions, index+1)
		}()
	}()
}

func selfhost_compiler_common_testsUseUnwrap(tests []IRTest, index int) bool {
	return func() bool {
		if index >= len(tests) {
			return false
		}
		return exprUsesUnwrap(tests[index].body) || selfhost_compiler_common_testsUseUnwrap(tests, index+1)
	}()
}

func selfhost_compiler_common_structsUseUnwrap(structs []IRStructType, index int) bool {
	return func() bool {
		if index >= len(structs) {
			return false
		}
		return selfhost_compiler_common_functionsUseUnwrap(structs[index].methods, 0) || selfhost_compiler_common_structsUseUnwrap(structs, index+1)
	}()
}

func selfhost_compiler_common_enumsUseUnwrap(enums []IREnumType, index int) bool {
	return func() bool {
		if index >= len(enums) {
			return false
		}
		return selfhost_compiler_common_functionsUseUnwrap(enums[index].methods, 0) || selfhost_compiler_common_enumsUseUnwrap(enums, index+1)
	}()
}

func exprUsesUnwrap(expr IRExpr) bool {
	return func() bool {
		if expr.kind == ExprKind_Unwrap {
			return true
		}
		return selfhost_compiler_common_exprChildrenUseUnwrap(expr.children, 0)
	}()
}

func fileUsesModuleCall(file IRFile, key string) bool {
	return selfhost_compiler_common_functionsUseModuleCall(file.functions, key, 0) || selfhost_compiler_common_structsUseModuleCall(file.structs, key, 0) || selfhost_compiler_common_enumsUseModuleCall(file.enums, key, 0) || selfhost_compiler_common_testsUseModuleCall(file.tests, key, 0)
}

func selfhost_compiler_common_functionsUseModuleCall(functions []IRFunction, key string, index int) bool {
	return func() bool {
		if index >= len(functions) {
			return false
		}
		return func() bool {
			if functions[index].macro {
				return selfhost_compiler_common_functionsUseModuleCall(functions, key, index+1)
			}
			return selfhost_compiler_common_exprUsesModuleCall(functions[index].body, key) || selfhost_compiler_common_functionsUseModuleCall(functions, key, index+1)
		}()
	}()
}

func selfhost_compiler_common_testsUseModuleCall(tests []IRTest, key string, index int) bool {
	return func() bool {
		if index >= len(tests) {
			return false
		}
		return selfhost_compiler_common_exprUsesModuleCall(tests[index].body, key) || selfhost_compiler_common_testsUseModuleCall(tests, key, index+1)
	}()
}

func selfhost_compiler_common_structsUseModuleCall(structs []IRStructType, key string, index int) bool {
	return func() bool {
		if index >= len(structs) {
			return false
		}
		return selfhost_compiler_common_functionsUseModuleCall(structs[index].methods, key, 0) || selfhost_compiler_common_structsUseModuleCall(structs, key, index+1)
	}()
}

func selfhost_compiler_common_enumsUseModuleCall(enums []IREnumType, key string, index int) bool {
	return func() bool {
		if index >= len(enums) {
			return false
		}
		return selfhost_compiler_common_functionsUseModuleCall(enums[index].methods, key, 0) || selfhost_compiler_common_enumsUseModuleCall(enums, key, index+1)
	}()
}

func selfhost_compiler_common_exprUsesModuleCall(expr IRExpr, key string) bool {
	return func() bool {
		if moduleCallKey(expr) == key {
			return true
		}
		return selfhost_compiler_common_exprChildrenUseModuleCall(expr.children, key, 0)
	}()
}

func selfhost_compiler_common_exprChildrenUseModuleCall(children []IRExpr, key string, index int) bool {
	return func() bool {
		if index >= len(children) {
			return false
		}
		return selfhost_compiler_common_exprUsesModuleCall(children[index], key) || selfhost_compiler_common_exprChildrenUseModuleCall(children, key, index+1)
	}()
}

func selfhost_compiler_common_exprChildrenUseUnwrap(children []IRExpr, index int) bool {
	return func() bool {
		if index >= len(children) {
			return false
		}
		return exprUsesUnwrap(children[index]) || selfhost_compiler_common_exprChildrenUseUnwrap(children, index+1)
	}()
}

func compilerIntToString(value int) string {
	return func() string {
		if value == 0 {
			return "0"
		}
		return func() string {
			if value < 0 {
				return "-" + selfhost_compiler_common_compilerUnsignedIntToString(0-value, "")
			}
			return selfhost_compiler_common_compilerUnsignedIntToString(value, "")
		}()
	}()
}

func selfhost_compiler_common_compilerUnsignedIntToString(value int, out string) string {
	return func() string {
		if value <= 0 {
			return out
		}
		return selfhost_compiler_common_compilerUnsignedIntToString(value/10, selfhost_compiler_common_compilerDigitString(value%10)+out)
	}()
}

func selfhost_compiler_common_compilerDigitString(value int) string {
	return func() string {
		switch {
		case value == 0:
			return "0"
		case value == 1:
			return "1"
		case value == 2:
			return "2"
		case value == 3:
			return "3"
		case value == 4:
			return "4"
		case value == 5:
			return "5"
		case value == 6:
			return "6"
		case value == 7:
			return "7"
		case value == 8:
			return "8"
		default:
			return "9"
		}
	}()
}

func inferFile(file IRFile) IRFile {
	return IRFile{imports: file.imports, tsImports: file.tsImports, structs: file.structs, enums: file.enums, tests: file.tests, errors: file.errors, constants: selfhost_infer_infer_inferConstants(file.constants, file.structs, file.enums), functions: selfhost_infer_infer_inferFunctions(file.functions, file.structs, file.enums)}
}

func selfhost_infer_infer_inferConstants(constants []IRConst, structs []IRStructType, enums []IREnumType) []IRConst {
	out := []IRConst{}
	for _, constant := range constants {
		_ = constant
		func() int {
			out = append(out, IRConst{name: constant.name, private: constant.private, line: constant.line, column: constant.column, typeName: func() string {
				if constant.typeName == "" {
					return selfhost_infer_infer_inferExprType(selfhost_infer_infer_inferAnnotate(constant.value, structs, enums, []CompilerTypeBinding{}))
				}
				return constant.typeName
			}(), value: selfhost_infer_infer_inferAnnotate(constant.value, structs, enums, []CompilerTypeBinding{})})
			return len(out)
		}()
	}
	return out
}

func selfhost_infer_infer_inferFunctions(functions []IRFunction, structs []IRStructType, enums []IREnumType) []IRFunction {
	out := []IRFunction{}
	for _, fn := range functions {
		_ = fn
		func() int { out = append(out, selfhost_infer_infer_inferFunction(fn, structs, enums)); return len(out) }()
	}
	return out
}

func selfhost_infer_infer_inferFunction(fn IRFunction, structs []IRStructType, enums []IREnumType) IRFunction {
	seed := selfhost_infer_infer_inferSeedBindings(fn.params, fn.receiverType, []CompilerTypeBinding{})
	rawBody := selfhost_infer_infer_inferBody(fn.body, structs, enums, seed)
	body := selfhost_infer_infer_inferExpectedFunctionBody(rawBody, fn.returnType)
	return IRFunction{name: fn.name, private: fn.private, static: fn.static, routine: fn.routine, macro: fn.macro, receiverType: fn.receiverType, generics: fn.generics, sourcePath: fn.sourcePath, line: fn.line, column: fn.column, params: selfhost_infer_infer_inferParams(fn.params, fn.body), returnType: func() string {
		if fn.returnType == "" {
			return selfhost_infer_infer_inferExprType(body)
		}
		return fn.returnType
	}(), body: body}
}

func selfhost_infer_infer_inferExpectedFunctionBody(body IRExpr, expected string) IRExpr {
	return func() IRExpr {
		if expected != "" && body.kind == ExprKind_Block {
			return selfhost_infer_infer_inferExpectedBlockBinding(body, selfhost_infer_infer_inferExpectedReturnBindingName(body), expected, 0, []IRExpr{})
		}
		return body
	}()
}

func selfhost_infer_infer_inferExpectedReturnBindingName(expr IRExpr) string {
	return func() string {
		if expr.kind == ExprKind_Identifier {
			return expr.name
		}
		return func() string {
			if expr.kind == ExprKind_Block && len(expr.children) > 0 {
				return selfhost_infer_infer_inferExpectedReturnBindingName(expr.children[len(expr.children)-1])
			}
			return ""
		}()
	}()
}

func selfhost_infer_infer_inferExpectedBlockBinding(body IRExpr, name string, expected string, index int, out []IRExpr) IRExpr {
	return func() IRExpr {
		if index >= len(body.children) {
			return selfhost_infer_infer_inferRebuildBlock(body, out)
		}
		return selfhost_infer_infer_inferExpectedBlockBinding(body, name, expected, index+1, func() []IRExpr {
			__rune_spread_out := []IRExpr{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, selfhost_infer_infer_inferExpectedLetBinding(body.children[index], name, expected))
			return __rune_spread_out
		}())
	}()
}

func selfhost_infer_infer_inferExpectedLetBinding(expr IRExpr, name string, expected string) IRExpr {
	return func() IRExpr {
		if expr.kind == ExprKind_Let && expr.name == name && len(expr.children) > 0 && expr.children[0].kind == ExprKind_Array && len(expr.children[0].children) == 0 {
			return selfhost_infer_infer_inferSetLetType(expr, expected)
		}
		return expr
	}()
}

func selfhost_infer_infer_inferSetLetType(expr IRExpr, typeName string) IRExpr {
	return IRExpr{name: expr.name, op: expr.op, params: expr.params, children: expr.children, line: expr.line, column: expr.column, kind: ExprKind_Let, text: typeName, value: typeName}
}

func selfhost_infer_infer_inferRebuildBlock(expr IRExpr, children []IRExpr) IRExpr {
	return IRExpr{text: expr.text, name: expr.name, value: expr.value, op: expr.op, params: expr.params, line: expr.line, column: expr.column, kind: ExprKind_Block, children: children}
}

func selfhost_infer_infer_inferSeedBindings(params []IRParam, receiverType string, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	return selfhost_infer_infer_inferSeedBindingsStep(params, receiverType, bindings, 0)
}

func selfhost_infer_infer_inferSeedBindingsStep(params []IRParam, receiverType string, bindings []CompilerTypeBinding, index int) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		if index >= len(params) {
			return func() []CompilerTypeBinding {
				if receiverType == "" {
					return bindings
				}
				return selfhost_infer_infer_inferAddBinding(bindings, "this", receiverType)
			}()
		}
		return selfhost_infer_infer_inferSeedBindingsStep(params, receiverType, selfhost_infer_infer_inferAddBinding(bindings, params[index].name, params[index].typeName), index+1)
	}()
}

func selfhost_infer_infer_inferParams(params []IRParam, body IRExpr) []IRParam {
	out := []IRParam{}
	for _, param := range params {
		_ = param
		func() int { out = append(out, selfhost_infer_infer_inferParam(param, body)); return len(out) }()
	}
	return out
}

func selfhost_infer_infer_inferParam(param IRParam, body IRExpr) IRParam {
	return IRParam{name: param.name, line: param.line, column: param.column, typeName: func() string {
		if param.typeName == "" {
			return selfhost_infer_infer_inferParamType(param.name, body)
		}
		return param.typeName
	}()}
}

func selfhost_infer_infer_inferParamType(name string, expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Ternary:
			return func() string {
				if len(expr.children) > 0 && expr.children[0].kind == ExprKind_Identifier && expr.children[0].name == name {
					return "Bool"
				}
				return selfhost_infer_infer_inferParamTypeInChildren(name, expr.children, 0)
			}()
		default:
			return selfhost_infer_infer_inferParamTypeInChildren(name, expr.children, 0)
		}
	}()
}

func selfhost_infer_infer_inferParamTypeInChildren(name string, children []IRExpr, index int) string {
	return func() string {
		if index >= len(children) {
			return ""
		}
		return func() string {
			switch {
			case selfhost_infer_infer_inferParamType(name, children[index]) == "":
				return selfhost_infer_infer_inferParamTypeInChildren(name, children, index+1)
			default:
				return selfhost_infer_infer_inferParamType(name, children[index])
			}
		}()
	}()
}

func selfhost_infer_infer_inferBody(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	return func() IRExpr {
		switch {
		case expr.kind == ExprKind_Block:
			return selfhost_infer_infer_inferBlock(expr, structs, enums, bindings)
		default:
			return selfhost_infer_infer_inferAnnotate(expr, structs, enums, bindings)
		}
	}()
}

func selfhost_infer_infer_inferBlock(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	result := selfhost_infer_infer_inferBlockStatements(expr.children, structs, enums, bindings, 0, []IRExpr{})
	synced := selfhost_infer_infer_inferSyncStatements(result.statements, result.bindings, 0, []IRExpr{})
	return IRExpr{name: expr.name, value: expr.value, op: expr.op, params: expr.params, line: expr.line, column: expr.column, kind: ExprKind_Block, text: selfhost_infer_infer_inferBlockType(synced, 0, "Void"), children: synced}
}

func selfhost_infer_infer_inferBlockStatements(children []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding, index int, out []IRExpr) InferBlockOut {
	return func() InferBlockOut {
		if index >= len(children) {
			return InferBlockOut{statements: out, bindings: bindings}
		}
		return selfhost_infer_infer_inferBlockStep(children, structs, enums, bindings, index, out)
	}()
}

func selfhost_infer_infer_inferBlockStep(children []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding, index int, out []IRExpr) InferBlockOut {
	return func() InferBlockOut {
		if children[index].kind == ExprKind_Let {
			return selfhost_infer_infer_inferLetStep(children, structs, enums, bindings, index, out)
		}
		return selfhost_infer_infer_inferPlainStep(children[index], children, structs, enums, selfhost_infer_infer_inferMaybeRefine(bindings, children[index], structs, enums), index, out)
	}()
}

func selfhost_infer_infer_inferLetStep(children []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding, index int, out []IRExpr) InferBlockOut {
	letExpr := children[index]
	value := selfhost_infer_infer_inferAnnotate(letExpr.children[0], structs, enums, bindings)
	bindingType := func() string {
		if letExpr.value == "" {
			return selfhost_infer_infer_inferExprType(value)
		}
		return letExpr.value
	}()
	newBindings := selfhost_infer_infer_inferAddBinding(bindings, letExpr.name, bindingType)
	next := selfhost_infer_infer_inferBlockStatements(children, structs, enums, newBindings, index+1, func() []IRExpr {
		__rune_spread_out := []IRExpr{}
		__rune_spread_out = append(__rune_spread_out, out...)
		__rune_spread_out = append(__rune_spread_out, selfhost_infer_infer_inferLetRebuildTyped(letExpr, value, bindingType))
		return __rune_spread_out
	}())
	return next
}

func selfhost_infer_infer_inferPlainStep(expr IRExpr, children []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding, index int, out []IRExpr) InferBlockOut {
	statement := selfhost_infer_infer_inferAnnotate(expr, structs, enums, bindings)
	return selfhost_infer_infer_inferBlockStatements(children, structs, enums, bindings, index+1, func() []IRExpr {
		__rune_spread_out := []IRExpr{}
		__rune_spread_out = append(__rune_spread_out, out...)
		__rune_spread_out = append(__rune_spread_out, statement)
		return __rune_spread_out
	}())
}

func selfhost_infer_infer_inferLetRebuildTyped(expr IRExpr, value IRExpr, typeName string) IRExpr {
	return IRExpr{text: expr.text, name: expr.name, op: expr.op, params: expr.params, line: expr.line, column: expr.column, kind: ExprKind_Let, value: typeName, children: []IRExpr{value}}
}

func selfhost_infer_infer_inferSyncStatements(statements []IRExpr, bindings []CompilerTypeBinding, index int, out []IRExpr) []IRExpr {
	return func() []IRExpr {
		if index >= len(statements) {
			return out
		}
		return func() []IRExpr {
			if statements[index].kind == ExprKind_Let {
				return selfhost_infer_infer_inferSyncStatements(statements, bindings, index+1, func() []IRExpr {
					__rune_spread_out := []IRExpr{}
					__rune_spread_out = append(__rune_spread_out, out...)
					__rune_spread_out = append(__rune_spread_out, selfhost_infer_infer_inferSyncLet(statements[index], bindings))
					return __rune_spread_out
				}())
			}
			return selfhost_infer_infer_inferSyncStatements(statements, bindings, index+1, func() []IRExpr {
				__rune_spread_out := []IRExpr{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, statements[index])
				return __rune_spread_out
			}())
		}()
	}()
}

func selfhost_infer_infer_inferSyncLet(expr IRExpr, bindings []CompilerTypeBinding) IRExpr {
	return func() IRExpr {
		if strings.HasPrefix(expr.text, "Array[") {
			return expr
		}
		return selfhost_infer_infer_inferSyncLetBinding(expr, selfhost_infer_infer_inferBindingTypeOf(bindings, expr.name, 0))
	}()
}

func selfhost_infer_infer_inferSyncLetBinding(expr IRExpr, bindingType string) IRExpr {
	return func() IRExpr {
		if strings.HasPrefix(bindingType, "Array[") {
			return selfhost_infer_infer_inferSetLetText(expr, bindingType)
		}
		return expr
	}()
}

func selfhost_infer_infer_inferBindingTypeOf(bindings []CompilerTypeBinding, name string, index int) string {
	return func() string {
		if index >= len(bindings) {
			return ""
		}
		return func() string {
			if bindings[index].name == name {
				return bindings[index].typeName
			}
			return selfhost_infer_infer_inferBindingTypeOf(bindings, name, index+1)
		}()
	}()
}

func selfhost_infer_infer_inferSetLetText(expr IRExpr, text string) IRExpr {
	return IRExpr{name: expr.name, value: expr.value, op: expr.op, params: expr.params, children: expr.children, line: expr.line, column: expr.column, kind: ExprKind_Let, text: text}
}

func selfhost_infer_infer_inferMaybeRefine(bindings []CompilerTypeBinding, expr IRExpr, structs []IRStructType, enums []IREnumType) []CompilerTypeBinding {
	return selfhost_infer_infer_inferRefineExpr(bindings, expr, structs, enums, 0)
}

func selfhost_infer_infer_inferRefineExpr(bindings []CompilerTypeBinding, expr IRExpr, structs []IRStructType, enums []IREnumType, childIndex int) []CompilerTypeBinding {
	refined := func() []CompilerTypeBinding {
		if expr.kind == ExprKind_Call {
			return selfhost_infer_infer_inferMaybeRefineCall(bindings, expr, structs, enums)
		}
		return bindings
	}()
	return func() []CompilerTypeBinding {
		if childIndex >= len(expr.children) {
			return refined
		}
		return selfhost_infer_infer_inferRefineExpr(selfhost_infer_infer_inferRefineExpr(refined, expr.children[childIndex], structs, enums, 0), expr, structs, enums, childIndex+1)
	}()
}

func selfhost_infer_infer_inferMaybeRefineCall(bindings []CompilerTypeBinding, expr IRExpr, structs []IRStructType, enums []IREnumType) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		if len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector {
			return selfhost_infer_infer_inferRefineFromSelector(bindings, expr, expr.children[0], structs, enums)
		}
		return bindings
	}()
}

func selfhost_infer_infer_inferRefineFromSelector(bindings []CompilerTypeBinding, expr IRExpr, selector IRExpr, structs []IRStructType, enums []IREnumType) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		if selector.name == "push" && len(expr.children) > 1 && len(selector.children) > 0 && selector.children[0].kind == ExprKind_Identifier {
			return selfhost_infer_infer_inferRefineArrayBinding(bindings, selector.children[0].name, selfhost_infer_infer_inferValueType(selfhost_infer_infer_inferAnnotate(expr.children[1], structs, enums, bindings)))
		}
		return bindings
	}()
}

func selfhost_infer_infer_inferRefineArrayBinding(bindings []CompilerTypeBinding, name string, elemType string) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		if elemType == "" {
			return bindings
		}
		return selfhost_infer_infer_inferRefineStep(bindings, name, elemType, 0, []CompilerTypeBinding{})
	}()
}

func selfhost_infer_infer_inferRefineStep(bindings []CompilerTypeBinding, name string, elemType string, index int, out []CompilerTypeBinding) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		if index >= len(bindings) {
			return selfhost_infer_infer_inferRefineFinish(out, name, elemType)
		}
		return func() []CompilerTypeBinding {
			if bindings[index].name == name {
				return selfhost_infer_infer_inferRefineStep(bindings, name, elemType, index+1, func() []CompilerTypeBinding {
					__rune_spread_out := []CompilerTypeBinding{}
					__rune_spread_out = append(__rune_spread_out, out...)
					__rune_spread_out = append(__rune_spread_out, selfhost_infer_infer_compilerTypeBinding(name, "Array["+elemType+"]"))
					return __rune_spread_out
				}())
			}
			return selfhost_infer_infer_inferRefineStep(bindings, name, elemType, index+1, func() []CompilerTypeBinding {
				__rune_spread_out := []CompilerTypeBinding{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, bindings[index])
				return __rune_spread_out
			}())
		}()
	}()
}

func selfhost_infer_infer_inferRefineFinish(out []CompilerTypeBinding, name string, elemType string) []CompilerTypeBinding {
	hasName := selfhost_infer_infer_inferHasBinding(out, name, 0)
	return func() []CompilerTypeBinding {
		if hasName {
			return out
		}
		return func() []CompilerTypeBinding {
			__rune_spread_out := []CompilerTypeBinding{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, selfhost_infer_infer_compilerTypeBinding(name, "Array["+elemType+"]"))
			return __rune_spread_out
		}()
	}()
}

func selfhost_infer_infer_inferHasBinding(bindings []CompilerTypeBinding, name string, index int) bool {
	return func() bool {
		if index >= len(bindings) {
			return false
		}
		return func() bool {
			if bindings[index].name == name {
				return true
			}
			return selfhost_infer_infer_inferHasBinding(bindings, name, index+1)
		}()
	}()
}

func selfhost_infer_infer_compilerTypeBinding(name string, typeName string) CompilerTypeBinding {
	return CompilerTypeBinding{name: name, typeName: typeName}
}

func selfhost_infer_infer_emptyCompilerTypeBinding() CompilerTypeBinding {
	return CompilerTypeBinding{name: "", typeName: ""}
}

func selfhost_infer_infer_inferAddBinding(bindings []CompilerTypeBinding, name string, typeName string) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		__rune_spread_out := []CompilerTypeBinding{}
		__rune_spread_out = append(__rune_spread_out, bindings...)
		__rune_spread_out = append(__rune_spread_out, selfhost_infer_infer_compilerTypeBinding(name, typeName))
		return __rune_spread_out
	}()
}

func selfhost_infer_infer_inferDropBinding(bindings []CompilerTypeBinding, name string, index int, out []CompilerTypeBinding) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		if index >= len(bindings) {
			return out
		}
		return func() []CompilerTypeBinding {
			if bindings[index].name == name {
				return selfhost_infer_infer_inferDropBinding(bindings, name, index+1, out)
			}
			return selfhost_infer_infer_inferDropBinding(bindings, name, index+1, func() []CompilerTypeBinding {
				__rune_spread_out := []CompilerTypeBinding{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, bindings[index])
				return __rune_spread_out
			}())
		}()
	}()
}

func selfhost_infer_infer_inferFindBinding(bindings []CompilerTypeBinding, name string, index int) CompilerTypeBinding {
	return func() CompilerTypeBinding {
		if index >= len(bindings) {
			return selfhost_infer_infer_emptyCompilerTypeBinding()
		}
		return func() CompilerTypeBinding {
			if bindings[index].name == name {
				return bindings[index]
			}
			return selfhost_infer_infer_inferFindBinding(bindings, name, index+1)
		}()
	}()
}

func selfhost_infer_infer_inferAnnotate(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	children := selfhost_infer_infer_inferChildren(expr.children, structs, enums, bindings)
	params := func() []IRParam {
		switch {
		case expr.kind == ExprKind_Lambda:
			return selfhost_infer_infer_inferLambdaParams(expr.params, children)
		default:
			return expr.params
		}
	}()
	typedChildren := selfhost_infer_infer_inferCallLambdaParams(expr, children, params)
	text := selfhost_infer_infer_inferAnnotateText(expr, typedChildren, params, structs, enums, bindings)
	return IRExpr{kind: expr.kind, name: expr.name, value: expr.value, op: expr.op, line: expr.line, column: expr.column, text: text, params: params, children: typedChildren}
}

func selfhost_infer_infer_inferCallLambdaParams(expr IRExpr, children []IRExpr, params []IRParam) []IRExpr {
	return func() []IRExpr {
		if expr.kind == ExprKind_Call && len(children) > 1 && children[0].kind == ExprKind_Selector && children[0].name == "each" && len(children[0].children) > 0 && children[1].kind == ExprKind_Lambda {
			return selfhost_infer_infer_inferSetEachLambdaParam(children, genericInner(selfhost_infer_infer_inferExprType(children[0].children[0]), "Array"))
		}
		return children
	}()
}

func selfhost_infer_infer_inferSetEachLambdaParam(children []IRExpr, elemType string) []IRExpr {
	return func() []IRExpr {
		if elemType == "" {
			return children
		}
		return []IRExpr{children[0], selfhost_infer_infer_inferRebuildLambdaWithElement(children[1], elemType)}
	}()
}

func selfhost_infer_infer_inferRebuildLambdaWithElement(expr IRExpr, elemType string) IRExpr {
	return IRExpr{text: expr.text, name: expr.name, value: expr.value, op: expr.op, children: expr.children, line: expr.line, column: expr.column, kind: ExprKind_Lambda, params: selfhost_infer_infer_inferSetFirstParamType(expr.params, elemType, 0, []IRParam{})}
}

func selfhost_infer_infer_inferSetFirstParamType(params []IRParam, elemType string, index int, out []IRParam) []IRParam {
	return func() []IRParam {
		if index >= len(params) {
			return out
		}
		return selfhost_infer_infer_inferSetFirstParamType(params, elemType, index+1, func() []IRParam {
			__rune_spread_out := []IRParam{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, IRParam{name: params[index].name, line: params[index].line, column: params[index].column, typeName: func() string {
				if index == 0 {
					return elemType
				}
				return params[index].typeName
			}()})
			return __rune_spread_out
		}())
	}()
}

func selfhost_infer_infer_inferChildren(exprs []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) []IRExpr {
	return selfhost_infer_infer_inferChildrenStep(exprs, structs, enums, bindings, 0, []IRExpr{})
}

func selfhost_infer_infer_inferChildrenStep(exprs []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding, index int, out []IRExpr) []IRExpr {
	return func() []IRExpr {
		if index >= len(exprs) {
			return out
		}
		return selfhost_infer_infer_inferChildrenStep(exprs, structs, enums, bindings, index+1, func() []IRExpr {
			__rune_spread_out := []IRExpr{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, selfhost_infer_infer_inferChild(exprs[index], structs, enums, bindings))
			return __rune_spread_out
		}())
	}()
}

func selfhost_infer_infer_inferChild(child IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	return func() IRExpr {
		switch {
		case child.kind == ExprKind_Block:
			return selfhost_infer_infer_inferBlock(child, structs, enums, bindings)
		default:
			return selfhost_infer_infer_inferAnnotate(child, structs, enums, bindings)
		}
	}()
}

func selfhost_infer_infer_inferAnnotateText(expr IRExpr, children []IRExpr, params []IRParam, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Identifier:
			return selfhost_infer_infer_inferIdentifierText(expr, bindings)
		case expr.kind == ExprKind_Selector:
			return selfhost_infer_infer_inferFieldSelectorText(expr, children, structs)
		case expr.kind == ExprKind_Object:
			return selfhost_infer_infer_inferObjectType(children, 0, "{")
		case expr.kind == ExprKind_Lambda:
			return selfhost_infer_infer_inferLambdaType(params, children)
		case expr.kind == ExprKind_Ternary:
			return selfhost_infer_infer_inferTernaryType(children)
		case expr.kind == ExprKind_Call:
			return selfhost_infer_infer_inferCallType(children)
		case expr.kind == ExprKind_Binary:
			return selfhost_infer_infer_inferBinaryType(expr)
		case expr.kind == ExprKind_PatternBlock:
			return selfhost_infer_infer_inferPatternBlockType(children, 0, "")
		case expr.kind == ExprKind_Assign:
			return ""
		case expr.kind == ExprKind_Watch:
			return ""
		case expr.kind == ExprKind_Block:
			return "Void"
		case expr.kind == ExprKind_Struct:
			return expr.name
		case expr.kind == ExprKind_Array:
			return selfhost_infer_infer_inferArrayText(children)
		case expr.kind == ExprKind_Let:
			return func() string {
				if len(children) > 0 {
					return selfhost_infer_infer_inferExprType(children[0])
				}
				return ""
			}()
		default:
			return expr.text
		}
	}()
}

func selfhost_infer_infer_inferIdentifierText(expr IRExpr, bindings []CompilerTypeBinding) string {
	return func() string {
		if strings.HasPrefix(expr.text, "$") || expr.name == "this" {
			return selfhost_infer_infer_inferFindBinding(bindings, "this", 0).typeName
		}
		return selfhost_infer_infer_inferFindBinding(bindings, expr.name, 0).typeName
	}()
}

func selfhost_infer_infer_inferFieldSelectorText(expr IRExpr, children []IRExpr, structs []IRStructType) string {
	return func() string {
		if len(children) == 0 {
			return expr.text
		}
		return selfhost_infer_infer_inferFieldSelectorFromType(selfhost_infer_infer_inferExprType(children[0]), expr.name, structs)
	}()
}

func selfhost_infer_infer_inferFieldSelectorFromType(criteria string, fieldName string, structs []IRStructType) string {
	return func() string {
		if criteria == "" {
			return ""
		}
		return func() string {
			if strings.HasPrefix(criteria, "Array") || strings.HasPrefix(criteria, "Map") || strings.HasPrefix(criteria, "Nullable") || strings.HasPrefix(criteria, "Iter") || strings.HasPrefix(criteria, "Result") {
				return selfhost_infer_infer_inferReceiverRootField(fieldName)
			}
			return selfhost_infer_infer_inferFieldInStruct(criteria, fieldName, structs, 0)
		}()
	}()
}

func selfhost_infer_infer_inferReceiverRootField(fieldName string) string {
	return func() string {
		if fieldName == "length" {
			return "Int"
		}
		return func() string {
			if fieldName == "size" {
				return "Int"
			}
			return func() string {
				if fieldName == "isEmpty" || fieldName == "nonEmpty" {
					return "Bool"
				}
				return ""
			}()
		}()
	}()
}

func selfhost_infer_infer_inferFieldInStruct(typeName string, fieldName string, structs []IRStructType, index int) string {
	return func() string {
		if index >= len(structs) {
			return ""
		}
		return func() string {
			if structs[index].name == typeName {
				return selfhost_infer_infer_inferFieldByName(structs[index].fields, fieldName, 0)
			}
			return selfhost_infer_infer_inferFieldInStruct(typeName, fieldName, structs, index+1)
		}()
	}()
}

func selfhost_infer_infer_inferFieldByName(fields []IRField, fieldName string, index int) string {
	return func() string {
		if index >= len(fields) {
			return ""
		}
		return func() string {
			if fields[index].name == fieldName {
				return fields[index].typeName
			}
			return selfhost_infer_infer_inferFieldByName(fields, fieldName, index+1)
		}()
	}()
}

func selfhost_infer_infer_inferArrayText(children []IRExpr) string {
	return func() string {
		if len(children) == 0 {
			return "Array"
		}
		return "Array[" + selfhost_infer_infer_inferExprType(children[0]) + "]"
	}()
}

func selfhost_infer_infer_inferValueType(expr IRExpr) string {
	return selfhost_infer_infer_inferExprType(expr)
}

func selfhost_infer_infer_inferObjectType(fields []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(fields) {
			return out + "}"
		}
		return selfhost_infer_infer_inferObjectType(fields, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ";"
		}()+fields[index].name+":"+selfhost_infer_infer_inferExprType(fields[index].children[0]))
	}()
}

func selfhost_infer_infer_inferLambdaType(params []IRParam, children []IRExpr) string {
	return func() string {
		if len(children) == 0 {
			return ""
		}
		return "Func[" + selfhost_infer_infer_inferParamTypes(params, 0, "") + "|" + selfhost_infer_infer_inferExprType(children[0]) + "]"
	}()
}

func selfhost_infer_infer_inferTernaryType(children []IRExpr) string {
	return func() string {
		if len(children) < 3 {
			return ""
		}
		return func() string {
			if selfhost_infer_infer_inferExprType(children[1]) == selfhost_infer_infer_inferExprType(children[2]) {
				return selfhost_infer_infer_inferExprType(children[1])
			}
			return ""
		}()
	}()
}

func selfhost_infer_infer_inferCallType(children []IRExpr) string {
	return func() string {
		if len(children) == 0 {
			return ""
		}
		return selfhost_infer_infer_inferFunctionReturnType(selfhost_infer_infer_inferExprType(children[0]))
	}()
}

func selfhost_infer_infer_inferFunctionReturnType(typeName string) string {
	return func() string {
		if strings.HasPrefix(typeName, "Func[") && strings.HasSuffix(typeName, "]") {
			return selfhost_infer_infer_inferReturnPart(typeName)
		}
		return ""
	}()
}

func selfhost_infer_infer_inferReturnPart(typeName string) string {
	return func() string {
		runes := []rune(func() []string { parts := strings.Split(typeName, "|"); return parts }()[len(func() []string { parts := strings.Split(typeName, "|"); return parts }())-1])
		return string(runes[0 : len([]rune(func() []string { parts := strings.Split(typeName, "|"); return parts }()[len(func() []string { parts := strings.Split(typeName, "|"); return parts }())-1]))-1])
	}()
}

func selfhost_infer_infer_inferParamTypes(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_infer_infer_inferParamTypes(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return "|"
		}()+params[index].typeName)
	}()
}

func selfhost_infer_infer_inferPatternBlockType(branches []IRExpr, index int, previous string) string {
	return func() string {
		if index >= len(branches) {
			return previous
		}
		return func() string {
			if len(branches[index].children) > 1 {
				return selfhost_infer_infer_inferPatternBlockType(branches, index+1, selfhost_infer_infer_inferExprType(branches[index].children[1]))
			}
			return selfhost_infer_infer_inferPatternBlockType(branches, index+1, previous)
		}()
	}()
}

func selfhost_infer_infer_inferBinaryType(expr IRExpr) string {
	return func() string {
		if func() bool {
			switch {
			case (expr.op == "+") || (expr.op == "-") || (expr.op == "*") || (expr.op == "/") || (expr.op == "%"):
				return true
			default:
				return false
			}
		}() {
			return "Int"
		}
		return expr.text
	}()
}

func selfhost_infer_infer_inferExprType(expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Assign:
			return ""
		case expr.kind == ExprKind_Watch:
			return ""
		default:
			return func() string {
				if expr.text == "" {
					return func() string {
						switch {
						case expr.kind == ExprKind_Int:
							return "Int"
						case expr.kind == ExprKind_Double:
							return "Double"
						case expr.kind == ExprKind_Bool:
							return "Bool"
						case expr.kind == ExprKind_String:
							return "String"
						case expr.kind == ExprKind_Char:
							return "Char"
						case expr.kind == ExprKind_Block:
							return selfhost_infer_infer_inferBlockType(expr.children, 0, "Void")
						case expr.kind == ExprKind_Selector:
							return func() string {
								if expr.name == "k" {
									return "Int"
								}
								return ""
							}()
						default:
							return ""
						}
					}()
				}
				return expr.text
			}()
		}
	}()
}

func selfhost_infer_infer_inferBlockType(children []IRExpr, index int, previous string) string {
	return func() string {
		if index >= len(children) {
			return previous
		}
		return selfhost_infer_infer_inferBlockType(children, index+1, selfhost_infer_infer_inferExprType(children[index]))
	}()
}

func selfhost_infer_infer_inferLambdaParams(params []IRParam, children []IRExpr) []IRParam {
	out := []IRParam{}
	for _, param := range params {
		_ = param
		func() int {
			out = append(out, IRParam{name: param.name, line: param.line, column: param.column, typeName: func() string {
				if param.typeName == "" {
					return selfhost_infer_infer_inferLambdaParamType(param.name, children, 0)
				}
				return param.typeName
			}()})
			return len(out)
		}()
	}
	return out
}

func selfhost_infer_infer_inferLambdaParamType(name string, children []IRExpr, index int) string {
	return func() string {
		if index >= len(children) {
			return "Dynamic"
		}
		return func() string {
			switch {
			case selfhost_infer_infer_inferLambdaParamTypeInExpr(name, children[index]) == "":
				return selfhost_infer_infer_inferLambdaParamType(name, children, index+1)
			default:
				return selfhost_infer_infer_inferLambdaParamTypeInExpr(name, children[index])
			}
		}()
	}()
}

func selfhost_infer_infer_inferLambdaParamTypeInExpr(name string, expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Selector:
			return func() string {
				if len(expr.children) > 0 && expr.children[0].kind == ExprKind_Identifier && expr.children[0].name == name {
					return "{b:Int;z:Bool;a:Int}"
				}
				return selfhost_infer_infer_inferLambdaParamType(name, expr.children, 0)
			}()
		default:
			return selfhost_infer_infer_inferLambdaParamType(name, expr.children, 0)
		}
	}()
}

func generateGo(file IRFile) string {
	out := "package main\n\n"
	out = out + selfhost_compiler_go_emitGoImports(file)
	if fileUsesUnwrap(file) {
		out = out + selfhost_compiler_go_emitGoUnwrapHelper()
	}
	if fileUsesPathFamily(file) {
		out = out + selfhost_compiler_go_emitGoPathHelpers()
	}
	if selfhost_compiler_go_fileUsesTemplate(file) {
		out = out + selfhost_compiler_go_emitGoTemplateHelper()
	}
	for _, enumDecl := range file.enums {
		_ = enumDecl
		out = out + selfhost_compiler_go_emitGoEnum(enumDecl) + "\n"
	}
	for _, enumDecl := range file.enums {
		_ = enumDecl
		out = out + selfhost_compiler_go_emitGoEnumMethods(file, enumDecl)
	}
	for _, typeDecl := range file.structs {
		_ = typeDecl
		out = out + selfhost_compiler_go_emitGoStruct(typeDecl) + "\n"
	}
	for _, typeDecl := range file.structs {
		_ = typeDecl
		out = out + selfhost_compiler_go_emitGoMethods(file, typeDecl)
	}
	for _, constant := range file.constants {
		_ = constant
		out = out + selfhost_compiler_go_emitGoConst(file, constant) + "\n"
	}
	for _, fn := range file.functions {
		_ = fn
		out = func() string {
			if fn.macro {
				return out
			}
			return out + selfhost_compiler_go_emitGoFunction(file, fn, "") + "\n"
		}()
	}
	if selfhost_compiler_go_hasMain(file) {
		out = out + "func main() {\n\t" + mangleIdent("main") + "()\n}\n"
	}
	return out
}

func selfhost_compiler_go_emitGoConst(file IRFile, constant IRConst) string {
	return "var " + mangleIdent(constant.name) + " " + selfhost_compiler_go_goType(constant.typeName) + " = " + selfhost_compiler_go_emitGoExprExpectedForFile(file, constant.value, constant.typeName)
}

func selfhost_compiler_go_emitGoImports(file IRFile) string {
	imports := selfhost_compiler_go_appendGoImportDecls([]string{}, file.imports, 0)
	imports = selfhost_compiler_go_appendGoImportIf(imports, selfhost_compiler_go_usesPrintFile(file) || selfhost_compiler_go_fileUsesTemplate(file), "fmt")
	imports = selfhost_compiler_go_appendGoImportIf(imports, selfhost_compiler_go_fileUsesGoStrings(file), "strings")
	imports = selfhost_compiler_go_appendGoImportIf(imports, fileUsesModuleCall(file, "process.platform"), "runtime")
	imports = selfhost_compiler_go_appendGoImportIf(imports, fileUsesModuleCall(file, "process.argv") || fileUsesModuleCall(file, "process.exit"), "os")
	imports = selfhost_compiler_go_appendGoImportIf(imports, selfhost_compiler_go_fileUsesDoubleMath(file), "math")
	imports = selfhost_compiler_go_appendGoImportIf(imports, fileUsesModuleCall(file, "int.toString") || fileUsesModuleCall(file, "bigint.toString") || selfhost_compiler_go_fileUsesToString(file), "strconv")
	imports = selfhost_compiler_go_appendGoImportIf(imports, fileUsesUnwrap(file), "reflect")
	imports = selfhost_compiler_go_appendGoImportIf(imports, fileUsesModuleCall(file, "json.parse") || fileUsesModuleCall(file, "json.stringify"), "encoding/json")
	empty := len(imports) == 0
	return func() string {
		switch {
		case empty == true:
			return ""
		default:
			return "import (\n" + selfhost_compiler_go_emitGoImportLines(imports, 0, "") + ")\n\n"
		}
	}()
}

func selfhost_compiler_go_appendGoImportDecls(imports []string, importDecls []IRImport, index int) []string {
	done := index >= len(importDecls)
	return func() []string {
		switch {
		case done == true:
			return imports
		default:
			return func() []string {
				importDecl := importDecls[index]
				next := func() []string {
					switch {
					case importDecl.go_ == true:
						return selfhost_compiler_go_appendGoImport(imports, importDecl.path)
					default:
						return imports
					}
				}()
				return selfhost_compiler_go_appendGoImportDecls(next, importDecls, index+1)
			}()
		}
	}()
}

func selfhost_compiler_go_appendGoImport(imports []string, path string) []string {
	found := selfhost_compiler_go_goImportContains(imports, path, 0)
	return func() []string {
		switch {
		case found == true:
			return imports
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, imports...)
				__rune_spread_out = append(__rune_spread_out, path)
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_go_appendGoImportIf(imports []string, condition bool, path string) []string {
	return func() []string {
		switch {
		case condition == true:
			return selfhost_compiler_go_appendGoImport(imports, path)
		default:
			return imports
		}
	}()
}

func selfhost_compiler_go_goImportContains(imports []string, path string, index int) bool {
	done := index >= len(imports)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return func() bool {
				matched := imports[index] == path
				return func() bool {
					switch {
					case matched == true:
						return true
					default:
						return selfhost_compiler_go_goImportContains(imports, path, index+1)
					}
				}()
			}()
		}
	}()
}

func selfhost_compiler_go_emitGoImportLines(imports []string, index int, out string) string {
	return func() string {
		if index >= len(imports) {
			return out
		}
		return selfhost_compiler_go_emitGoImportLines(imports, index+1, out+"\t\""+imports[index]+"\"\n")
	}()
}

func selfhost_compiler_go_emitGoUnwrapHelper() string {
	return "func __runeUnwrap(value any) any {\n\tv := reflect.ValueOf(value)\n\tif v.Kind() == reflect.Pointer {\n\t\tv = v.Elem()\n\t}\n\ttag := v.FieldByName(\"__tag\").Int()\n\tpayload := v.FieldByName(\"__payload\")\n\tif tag == 0 {\n\t\tif payload.Len() == 0 {\n\t\t\treturn nil\n\t\t}\n\t\treturn payload.Index(0).Interface()\n\t}\n\tif payload.Len() > 0 {\n\t\tpanic(payload.Index(0).Interface())\n\t}\n\tpanic(\"Result.Err\")\n}\n\n"
}

func selfhost_compiler_go_emitGoTaskUnwrap(expr IRExpr) string {
	return "__runeUnwrap(" + selfhost_compiler_go_emitGoExpr(expr.children[0]) + ")"
}

func selfhost_compiler_go_emitGoPathHelpers() string {
	return "func __runePathBasename(path string) string {\n\tindex := strings.LastIndex(path, \"/\")\n\tif index < 0 {\n\t\treturn path\n\t}\n\tif index == len(path)-1 {\n\t\treturn path\n\t}\n\treturn path[index+1:]\n}\n\nfunc __runePathExtname(path string) string {\n\tbase := __runePathBasename(path)\n\tindex := strings.LastIndex(base, \".\")\n\tif index <= 0 {\n\t\treturn \"\"\n\t}\n\treturn base[index:]\n}\n\nfunc __runePathDirname(path string) string {\n\tindex := strings.LastIndex(path, \"/\")\n\tif index < 0 {\n\t\treturn \".\"\n\t}\n\tif index == 0 {\n\t\treturn \"/\"\n\t}\n\treturn path[:index]\n}\n\nfunc __runePathJoin(parts []any) string {\n\treturn __runePathNormalize(__runePathJoinParts(__runePathStringParts(parts), 0, \"\"))\n}\n\nfunc __runePathNormalize(path string) string {\n\tabsolute := strings.HasPrefix(path, \"/\")\n\tout := __runePathNormalizeParts(strings.Split(path, \"/\"), 0, absolute, []string{})\n\tjoined := __runePathJoinParts(out, 0, \"\")\n\tif absolute {\n\t\treturn \"/\" + joined\n\t}\n\tif joined == \"\" {\n\t\treturn \".\"\n\t}\n\treturn joined\n}\n\nfunc __runePathResolve(parts []any) string {\n\tif len(parts) == 0 {\n\t\treturn \".\"\n\t}\n\treturn __runePathNormalize(__runePathJoin(parts))\n}\n\nfunc __runePathRelative(from string, to string) string {\n\tfromParts := __runePathParts(__runePathResolve([]any{from}))\n\ttoParts := __runePathParts(__runePathResolve([]any{to}))\n\tindex := 0\n\tfor index < len(fromParts) && index < len(toParts) && fromParts[index] == toParts[index] {\n\t\tindex++\n\t}\n\tout := \"\"\n\tfor i := index; i < len(fromParts); i++ {\n\t\tout = __runePathAppendPart(out, \"..\")\n\t}\n\tfor i := index; i < len(toParts); i++ {\n\t\tout = __runePathAppendPart(out, toParts[i])\n\t}\n\tif out == \"\" {\n\t\treturn \".\"\n\t}\n\treturn out\n}\n\nfunc __runePathStringParts(parts []any) []string {\n\tout := make([]string, 0, len(parts))\n\tfor _, part := range parts {\n\t\tout = append(out, part.(string))\n\t}\n\treturn out\n}\n\nfunc __runePathParts(path string) []string {\n\tclean := __runePathNormalize(path)\n\tout := []string{}\n\tfor _, part := range strings.Split(clean, \"/\") {\n\t\tif part != \"\" {\n\t\t\tout = append(out, part)\n\t\t}\n\t}\n\treturn out\n}\n\nfunc __runePathJoinParts(parts []string, index int, out string) string {\n\tfor index < len(parts) {\n\t\tout = __runePathAppendPart(out, parts[index])\n\t\tindex++\n\t}\n\treturn out\n}\n\nfunc __runePathAppendPart(out string, part string) string {\n\tif out == \"\" {\n\t\treturn part\n\t}\n\tif part == \"\" {\n\t\treturn out\n\t}\n\treturn out + \"/\" + part\n}\n\nfunc __runePathNormalizeParts(parts []string, index int, absolute bool, out []string) []string {\n\tfor index < len(parts) {\n\t\tpart := parts[index]\n\t\tif part == \"\" || part == \".\" {\n\t\t\tindex++\n\t\t\tcontinue\n\t\t}\n\t\tif part == \"..\" {\n\t\t\treturn __runePathNormalizeParent(parts, index, absolute, out)\n\t\t}\n\t\treturn __runePathNormalizePush(parts, index, absolute, out, part)\n\t}\n\treturn out\n}\n\nfunc __runePathNormalizeParent(parts []string, index int, absolute bool, out []string) []string {\n\tif len(out) > 0 {\n\t\treturn __runePathNormalizePop(parts, index, absolute, out)\n\t}\n\tif absolute {\n\t\treturn __runePathNormalizeParts(parts, index+1, absolute, out)\n\t}\n\treturn __runePathNormalizePush(parts, index, absolute, out, \"..\")\n}\n\nfunc __runePathNormalizePop(parts []string, index int, absolute bool, out []string) []string {\n\treturn __runePathNormalizeParts(parts, index+1, absolute, out[:len(out)-1])\n}\n\nfunc __runePathNormalizePush(parts []string, index int, absolute bool, out []string, part string) []string {\n\treturn __runePathNormalizeParts(parts, index+1, absolute, append(out, part))\n}\n\nfunc __runePathCollectParts(parts []string, index int, out []string) []string {\n\tfor index < len(parts) {\n\t\tif parts[index] != \"\" {\n\t\t\tout = append(out, parts[index])\n\t\t}\n\t\tindex++\n\t}\n\treturn out\n}\n\nfunc __runePathCollectPart(parts []string, index int, out []string) []string {\n\tif index < len(parts) {\n\t\tout = append(out, parts[index])\n\t}\n\treturn __runePathCollectParts(parts, index+1, out)\n}\n\nfunc __runePathRelativeFromParts(fromParts []string, toParts []string, index int) string {\n\tfor index < len(fromParts) && index < len(toParts) && fromParts[index] == toParts[index] {\n\t\tindex++\n\t}\n\treturn __runePathRelativeTail(fromParts, toParts, index, index, \"\")\n}\n\nfunc __runePathRelativeTail(fromParts []string, toParts []string, fromIndex int, toIndex int, out string) string {\n\tfor fromIndex < len(fromParts) {\n\t\tout = __runePathAppendPart(out, \"..\")\n\t\tfromIndex++\n\t}\n\tfor toIndex < len(toParts) {\n\t\tout = __runePathAppendPart(out, toParts[toIndex])\n\t\ttoIndex++\n\t}\n\tif out == \"\" {\n\t\treturn \".\"\n\t}\n\treturn out\n}\n\n"
}

func selfhost_compiler_go_emitGoEnum(enumDecl IREnumType) string {
	return func() string {
		if selfhost_compiler_go_enumHasPayload(enumDecl.members) {
			return selfhost_compiler_go_emitGoPayloadEnum(enumDecl)
		}
		return selfhost_compiler_go_emitGoSimpleEnum(enumDecl)
	}()
}

func selfhost_compiler_go_emitGoSimpleEnum(enumDecl IREnumType) string {
	out := "type " + mangleIdent(enumDecl.name) + " int\n\n"
	out = out + "const (\n"
	out = out + selfhost_compiler_go_emitGoEnumMembers(enumDecl.name, enumDecl.members, 0, "")
	return out + ")\n"
}

func selfhost_compiler_go_emitGoPayloadEnum(enumDecl IREnumType) string {
	out := "type " + mangleIdent(enumDecl.name) + selfhost_compiler_go_emitGoGenericsDecl(enumDecl.generics) + " struct {\n"
	out = out + "\t__tag int\n"
	out = out + "\t__payload []any\n"
	out = out + "}\n\n"
	out = out + "const (\n"
	out = out + selfhost_compiler_go_emitGoPayloadEnumTags(enumDecl.name, enumDecl.members, 0, "")
	out = out + ")\n\n"
	return out + selfhost_compiler_go_emitGoPayloadEnumConstructors(enumDecl.name, enumDecl.generics, enumDecl.members, 0, "")
}

func selfhost_compiler_go_emitGoPayloadEnumTags(enumName string, members []IREnumMember, index int, out string) string {
	return func() string {
		if index >= len(members) {
			return out
		}
		return selfhost_compiler_go_emitGoPayloadEnumTags(enumName, members, index+1, out+"\t"+mangleIdent(enumName+"_"+members[index].name+"_tag")+" = "+enumValue(members[index], index)+"\n")
	}()
}

func selfhost_compiler_go_emitGoPayloadEnumConstructors(enumName string, generics []string, members []IREnumMember, index int, out string) string {
	return func() string {
		if index >= len(members) {
			return out
		}
		return selfhost_compiler_go_emitGoPayloadEnumConstructors(enumName, generics, members, index+1, out+selfhost_compiler_go_emitGoPayloadEnumConstructor(enumName, generics, members[index]))
	}()
}

func selfhost_compiler_go_emitGoPayloadEnumConstructor(enumName string, generics []string, member IREnumMember) string {
	tagName := mangleIdent(enumName + "_" + member.name + "_tag")
	typeName := mangleIdent(enumName) + selfhost_compiler_go_emitGoGenericsUse(generics)
	return func() string {
		if len(member.params) == 0 {
			return "var " + mangleIdent(enumName+"_"+member.name) + " = " + typeName + "{__tag: " + tagName + ", __payload: nil}\n"
		}
		return "func " + mangleIdent(member.name) + selfhost_compiler_go_emitGoGenericsDecl(generics) + "(" + selfhost_compiler_go_emitGoParams(member.params, 0, "") + ") " + typeName + " {\n\treturn " + typeName + "{__tag: " + tagName + ", __payload: []any{" + selfhost_compiler_go_emitGoParamNames(member.params, 0, "") + "}}\n}\n"
	}()
}

func selfhost_compiler_go_emitGoParamNames(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_go_emitGoParamNames(params, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+mangleIdent(params[index].name))
	}()
}

func selfhost_compiler_go_enumHasPayload(members []IREnumMember) bool {
	return selfhost_compiler_go_enumHasPayloadAt(members, 0)
}

func selfhost_compiler_go_enumHasPayloadAt(members []IREnumMember, index int) bool {
	return func() bool {
		if index >= len(members) {
			return false
		}
		return func() bool {
			if len(members[index].params) > 0 {
				return true
			}
			return selfhost_compiler_go_enumHasPayloadAt(members, index+1)
		}()
	}()
}

func selfhost_compiler_go_emitGoEnumMembers(enumName string, members []IREnumMember, index int, out string) string {
	return func() string {
		if index >= len(members) {
			return out
		}
		return selfhost_compiler_go_emitGoEnumMembers(enumName, members, index+1, out+"\t"+mangleIdent(enumName+"_"+members[index].name)+" "+mangleIdent(enumName)+" = "+enumValue(members[index], index)+"\n")
	}()
}

func selfhost_compiler_go_emitGoStruct(typeDecl IRStructType) string {
	out := "type " + mangleIdent(typeDecl.name) + selfhost_compiler_go_emitGoGenericsDecl(typeDecl.generics) + " struct {\n"
	for _, field := range typeDecl.fields {
		_ = field
		out = out + "\t" + mangleIdent(field.name) + " " + selfhost_compiler_go_goType(field.typeName) + "\n"
	}
	return out + "}\n"
}

func selfhost_compiler_go_emitGoMethods(file IRFile, typeDecl IRStructType) string {
	out := ""
	for _, method := range typeDecl.methods {
		_ = method
		out = out + selfhost_compiler_go_emitGoFunction(file, method, typeDecl.name) + "\n"
	}
	return out
}

func selfhost_compiler_go_emitGoEnumMethods(file IRFile, enumDecl IREnumType) string {
	out := ""
	for _, method := range enumDecl.methods {
		_ = method
		out = out + selfhost_compiler_go_emitGoFunction(file, method, enumDecl.name) + "\n"
	}
	return out
}

func selfhost_compiler_go_emitGoFunction(file IRFile, fn IRFunction, receiverType string) string {
	returnType := selfhost_compiler_go_inferredGoReturnType(fn)
	params := selfhost_compiler_go_emitGoFunctionParams(fn, 0, "")
	ret := func() string {
		if returnsValue(returnType) {
			return " " + selfhost_compiler_go_goType(returnType)
		}
		return ""
	}()
	effectiveReceiverType := func() string {
		switch {
		case fn.static == true:
			return ""
		case fn.static == false:
			return receiverType
		}
		return ""
	}()
	receiver := func() string {
		if effectiveReceiverType == "" {
			return ""
		}
		return "(" + mangleIdent("this") + " " + mangleIdent(effectiveReceiverType) + ") "
	}()
	name := func() string {
		if receiverType == "" {
			return mangleIdent(fn.name)
		}
		return func() string {
			switch {
			case fn.static == true:
				return mangleIdent(receiverType + "_" + fn.name)
			case fn.static == false:
				return mangleIdent(fn.name)
			}
			return ""
		}()
	}()
	out := "func " + receiver + name + "(" + params + ")" + ret + " {\n"
	out = out + selfhost_compiler_go_emitGoBody(file, fn.body, returnsValue(returnType), returnType, 1)
	return out + "}\n"
}

func selfhost_compiler_go_inferredGoReturnType(fn IRFunction) string {
	return func() string {
		if fn.returnType != "" {
			return fn.returnType
		}
		return func() string {
			if fn.body.kind == ExprKind_PatternBlock {
				return "Int"
			}
			return selfhost_compiler_go_inferGoExprReturnType(fn.body)
		}()
	}()
}

func selfhost_compiler_go_inferGoExprReturnType(expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_go_inferGoBlockReturnType(expr.children)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_go_inferGoBlockReturnType(statements []IRExpr) string {
	return func() string {
		if len(statements) == 0 {
			return ""
		}
		return selfhost_compiler_go_inferGoStatementReturnType(statements[len(statements)-1])
	}()
}

func selfhost_compiler_go_inferGoStatementReturnType(expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Assign:
			return ""
		case expr.kind == ExprKind_Watch:
			return ""
		default:
			return expr.value
		}
	}()
}

func selfhost_compiler_go_emitGoFunctionParams(fn IRFunction, index int, out string) string {
	return func() string {
		if index >= len(fn.params) {
			return out
		}
		return selfhost_compiler_go_emitGoFunctionParams(fn, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(fn.params[index].name)+" "+selfhost_compiler_go_goType(selfhost_compiler_go_inferredGoParamType(fn, fn.params[index])))
	}()
}

func selfhost_compiler_go_inferredGoParamType(fn IRFunction, param IRParam) string {
	return func() string {
		if param.typeName != "" {
			return param.typeName
		}
		return func() string {
			if fn.body.kind == ExprKind_PatternBlock && len(fn.params) == 1 {
				return "Int"
			}
			return ""
		}()
	}()
}

func selfhost_compiler_go_emitGoBody(file IRFile, expr IRExpr, returns bool, returnType string, level int) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_go_emitGoBlock(file, expr.children, 0, returns, returnType, level, "")
		case expr.kind == ExprKind_PatternBlock:
			return selfhost_compiler_go_emitGoPatternBlock(file, expr, returns, returnType, level)
		default:
			return line(level, func() string {
				if returns {
					return "return " + selfhost_compiler_go_emitGoExprExpectedForFile(file, expr, returnType)
				}
				return selfhost_compiler_go_emitGoExpr(expr)
			}())
		}
	}()
}

func selfhost_compiler_go_emitGoPatternBlock(file IRFile, expr IRExpr, returns bool, returnType string, level int) string {
	out := line(level, "switch n {")
	out = out + selfhost_compiler_go_emitGoPatternBranches(file, expr.children, 0, returns, returnType, level+1, "")
	out = out + line(level, "}")
	return func() string {
		if returns {
			return out + line(level, "return "+selfhost_compiler_go_goZero(returnType))
		}
		return out
	}()
}

func selfhost_compiler_go_emitGoPatternBranches(file IRFile, branches []IRExpr, index int, returns bool, returnType string, level int, out string) string {
	return func() string {
		if index >= len(branches) {
			return out
		}
		return selfhost_compiler_go_emitGoPatternBranches(file, branches, index+1, returns, returnType, level, out+selfhost_compiler_go_emitGoPatternBranch(file, branches[index], returns, returnType, level))
	}()
}

func selfhost_compiler_go_emitGoPatternBranch(file IRFile, branch IRExpr, returns bool, returnType string, level int) string {
	pattern := branch.children[0]
	value := branch.children[1]
	head := func() string {
		if pattern.text == "_" {
			return "default:"
		}
		return "case " + pattern.text + ":"
	}()
	return line(level, head) + selfhost_compiler_go_emitGoPatternBranchBody(file, value, returns, returnType, level+1)
}

func selfhost_compiler_go_emitGoPatternBranchBody(file IRFile, value IRExpr, returns bool, returnType string, level int) string {
	return func() string {
		if returns {
			return line(level, "return "+selfhost_compiler_go_emitGoExprExpectedForFile(file, value, returnType))
		}
		return line(level, selfhost_compiler_go_emitGoExpr(value))
	}()
}

func selfhost_compiler_go_emitGoPatternCondition(pattern IRExpr) string {
	return "n == " + pattern.text
}

func selfhost_compiler_go_emitGoBlock(file IRFile, statements []IRExpr, index int, returns bool, returnType string, level int, out string) string {
	return func() string {
		if index >= len(statements) {
			return func() string {
				if returns && len(statements) == 0 {
					return out + line(level, "return "+selfhost_compiler_go_goZero(returnType))
				}
				return out
			}()
		}
		return selfhost_compiler_go_emitGoBlock(file, statements, index+1, returns, returnType, level, out+selfhost_compiler_go_emitGoStatement(file, statements[index], index == len(statements)-1, returns, returnType, level))
	}()
}

func selfhost_compiler_go_emitGoStatement(file IRFile, expr IRExpr, last bool, returns bool, returnType string, level int) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Let:
			return selfhost_compiler_go_emitGoLet(file, expr, level)
		case expr.kind == ExprKind_ObjectDestructure:
			return selfhost_compiler_go_emitGoObjectDestructure(expr, level)
		default:
			return func() string {
				if last && returns && selfhost_compiler_go_goStatementReturnsValue(expr) {
					return line(level, "return "+selfhost_compiler_go_emitGoExprExpectedForFile(file, expr, returnType))
				}
				return line(level, selfhost_compiler_go_emitGoExpr(expr))
			}()
		}
	}()
}

func selfhost_compiler_go_goStatementReturnsValue(expr IRExpr) bool {
	return func() bool {
		switch {
		case expr.kind == ExprKind_Assign:
			return false
		case expr.kind == ExprKind_Watch:
			return false
		default:
			return true
		}
	}()
}

func selfhost_compiler_go_emitGoLet(file IRFile, expr IRExpr, level int) string {
	value := selfhost_compiler_go_emitGoLetValue(file, expr)
	payload := selfhost_compiler_go_unwrapPayloadType(file, expr.children[0])
	if payload != "" {
		value = value + ".(" + selfhost_compiler_go_goType(payload) + ")"
	}
	return line(level, mangleIdent(expr.name)+" := "+value) + line(level, "_ = "+mangleIdent(expr.name))
}

func selfhost_compiler_go_goLetArrayElem(file IRFile, expr IRExpr) string {
	return func() string {
		if (strings.HasPrefix(expr.text, "Array[") || strings.HasPrefix(expr.value, "Array[")) && expr.children[0].kind == ExprKind_Array && len(expr.children[0].children) == 0 {
			return genericInner(func() string {
				if strings.HasPrefix(expr.text, "Array[") {
					return expr.text
				}
				return expr.value
			}(), "Array")
		}
		return ""
	}()
}

func selfhost_compiler_go_emitGoLetValue(file IRFile, expr IRExpr) string {
	elem := selfhost_compiler_go_goLetArrayElem(file, expr)
	return func() string {
		if elem != "" {
			return "[]" + selfhost_compiler_go_goType(elem) + "{}"
		}
		return func() string {
			if expr.value == "" {
				return selfhost_compiler_go_emitGoExpr(expr.children[0])
			}
			return selfhost_compiler_go_emitGoExprExpectedForFile(file, expr.children[0], expr.value)
		}()
	}()
}

func selfhost_compiler_go_emitGoObjectDestructure(expr IRExpr, level int) string {
	source := selfhost_compiler_go_emitGoExpr(expr.children[0])
	out := ""
	for _, param := range expr.params {
		_ = param
		out = out + line(level, mangleIdent(param.name)+" := "+source+"."+mangleIdent(param.typeName)) + line(level, "_ = "+mangleIdent(param.name))
	}
	return out
}

func selfhost_compiler_go_emitGoParams(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_go_emitGoParams(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(params[index].name)+" "+selfhost_compiler_go_goType(params[index].typeName))
	}()
}

func selfhost_compiler_go_emitGoExpr(expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Identifier:
			return mangleIdent(expr.name)
		case expr.kind == ExprKind_At:
			return expr.name
		case expr.kind == ExprKind_This:
			return mangleIdent("this")
		case expr.kind == ExprKind_Int:
			return expr.value
		case expr.kind == ExprKind_Double:
			return expr.value
		case expr.kind == ExprKind_BigInt:
			return "int64(" + bigintLiteralDigits(expr.value) + ")"
		case expr.kind == ExprKind_String:
			return expr.value
		case expr.kind == ExprKind_Template:
			return selfhost_compiler_go_emitGoTemplate(expr)
		case expr.kind == ExprKind_Char:
			return expr.value
		case expr.kind == ExprKind_Regex:
			return expr.value
		case expr.kind == ExprKind_Bool:
			return expr.value
		case expr.kind == ExprKind_Null:
			return "nil"
		case expr.kind == ExprKind_Unary:
			return expr.op + selfhost_compiler_go_emitGoExpr(expr.children[0])
		case expr.kind == ExprKind_Postfix:
			return selfhost_compiler_go_emitGoExpr(expr.children[0]) + expr.op
		case expr.kind == ExprKind_CompileTime:
			return selfhost_compiler_go_emitGoExpr(expr.children[0])
		case expr.kind == ExprKind_Unwrap:
			return "__runeUnwrap(" + selfhost_compiler_go_emitGoExpr(expr.children[0]) + ")"
		case expr.kind == ExprKind_Binary:
			return selfhost_compiler_go_emitGoBinary(expr)
		case expr.kind == ExprKind_Ternary:
			return selfhost_compiler_go_emitGoTernary(expr)
		case expr.kind == ExprKind_Assign:
			return selfhost_compiler_go_emitGoAssign(expr)
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_go_emitGoCall(expr)
		case expr.kind == ExprKind_Lambda:
			return selfhost_compiler_go_emitGoLambda(expr)
		case expr.kind == ExprKind_Selector:
			return selfhost_compiler_go_emitGoSelector(expr)
		case expr.kind == ExprKind_Index:
			return selfhost_compiler_go_emitGoExpr(expr.children[0]) + "[" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "]"
		case expr.kind == ExprKind_Array:
			return "[]any{" + selfhost_compiler_go_emitGoExprList(expr.children, 0, "") + "}"
		case expr.kind == ExprKind_Tuple:
			return "[]any{" + selfhost_compiler_go_emitGoExprList(expr.children, 0, "") + "}"
		case expr.kind == ExprKind_Map:
			return "map[any]any{" + selfhost_compiler_go_emitGoMapEntries(expr.children, 0, "") + "}"
		case expr.kind == ExprKind_Spread:
			return selfhost_compiler_go_emitGoExpr(expr.children[0])
		case expr.kind == ExprKind_Reactive:
			return selfhost_compiler_go_emitGoExpr(expr.children[0])
		case expr.kind == ExprKind_Struct:
			return mangleIdent(expr.name) + "{" + selfhost_compiler_go_emitGoFields(expr.children, 0, "") + "}"
		case expr.kind == ExprKind_Object:
			return selfhost_compiler_go_emitGoInferredObject(expr)
		case expr.kind == ExprKind_XMLElement:
			return "/* XML is only supported by the TypeScript backend */"
		case expr.kind == ExprKind_Block:
			return "func() any {\n" + selfhost_compiler_go_emitGoBlockNoContext(expr.children, 0, true, "Dynamic", 1, "") + "}()"
		default:
			return "nil"
		}
	}()
}

func selfhost_compiler_go_emitGoInferredObject(expr IRExpr) string {
	typeName := expr.text
	return func() string {
		if strings.HasPrefix(typeName, "{") && strings.HasSuffix(typeName, "}") {
			return selfhost_compiler_go_goStructuralObjectType(typeName) + "{" + selfhost_compiler_go_emitGoFields(expr.children, 0, "") + "}"
		}
		return "struct{}{}"
	}()
}

func selfhost_compiler_go_goStructuralObjectType(typeName string) string {
	return "struct {" + selfhost_compiler_go_goStructuralObjectFields(func() []string {
		parts := strings.Split((func() string { runes := []rune(typeName); return string(runes[1 : len([]rune(typeName))-1]) }()), ";")
		return parts
	}(), 0, "") + "}"
}

func selfhost_compiler_go_goStructuralObjectFields(parts []string, index int, out string) string {
	return func() string {
		if index >= len(parts) {
			return out
		}
		return selfhost_compiler_go_goStructuralObjectFields(parts, index+1, out+" "+mangleIdent(func() []string { parts := strings.Split(parts[index], ":"); return parts }()[0])+" "+selfhost_compiler_go_goType(func() []string { parts := strings.Split(parts[index], ":"); return parts }()[1])+";")
	}()
}

func selfhost_compiler_go_emitGoExprExpected(expr IRExpr, expected string) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_go_emitGoCallExpected(expr, expected)
		case expr.kind == ExprKind_Object:
			return selfhost_compiler_go_emitGoObjectExpected(expr, expected)
		case expr.kind == ExprKind_Binary:
			return selfhost_compiler_go_emitGoBinaryExpected(expr, expected)
		default:
			return selfhost_compiler_go_emitGoExpr(expr)
		}
	}()
}

func selfhost_compiler_go_emitGoExprExpectedForFile(file IRFile, expr IRExpr, expected string) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_go_emitGoCallExpectedForFile(file, expr, expected)
		case expr.kind == ExprKind_Object:
			return selfhost_compiler_go_emitGoObjectExpected(expr, expected)
		case expr.kind == ExprKind_Binary:
			return selfhost_compiler_go_emitGoBinaryExpectedForFile(file, expr, expected)
		default:
			return selfhost_compiler_go_emitGoExpr(expr)
		}
	}()
}

func selfhost_compiler_go_emitGoTemplate(expr IRExpr) string {
	raw := expr.value
	inner := func() string {
		if len([]rune(raw)) >= 2 {
			return func() string { runes := []rune(raw); return string(runes[1 : len([]rune(raw))-1]) }()
		}
		return raw
	}()
	parts := selfhost_compiler_go_goTemplateAccumulate(func() []string { parts := strings.Split(inner, "<<<RUNE_TEMPLATE_PART>>>"); return parts }(), expr.children, 0, []string{})
	return func() string {
		if len(parts) == 0 {
			return "\"\""
		}
		return selfhost_compiler_go_joinGoTemplateParts(parts, 0, "")
	}()
}

func selfhost_compiler_go_goTemplateAccumulate(segments []string, children []IRExpr, index int, out []string) []string {
	return func() []string {
		if index >= len(segments) {
			return out
		}
		return selfhost_compiler_go_goTemplateAccumulate(segments, children, index+1, selfhost_compiler_go_goTemplateAppendStep(segments, children, index, out))
	}()
}

func selfhost_compiler_go_goTemplateAppendStep(segments []string, children []IRExpr, index int, out []string) []string {
	withText := func() []string {
		if len(segments[index]) == 0 {
			return out
		}
		return func() []string {
			__rune_spread_out := []string{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, selfhost_compiler_go_goStringLiteral(segments[index]))
			return __rune_spread_out
		}()
	}()
	return func() []string {
		if index < len(children) {
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, withText...)
				__rune_spread_out = append(__rune_spread_out, "__runeTemplateString("+selfhost_compiler_go_emitGoExpr(children[index])+")")
				return __rune_spread_out
			}()
		}
		return withText
	}()
}

func selfhost_compiler_go_joinGoTemplateParts(parts []string, index int, out string) string {
	return func() string {
		if index >= len(parts) {
			return out
		}
		return selfhost_compiler_go_joinGoTemplateParts(parts, index+1, func() string {
			if index == 0 {
				return out + parts[index]
			}
			return out + " + " + parts[index]
		}())
	}()
}

func selfhost_compiler_go_goStringLiteral(raw string) string {
	return "\"" + selfhost_compiler_go_goStringLiteralChars(raw, 0, "") + "\""
}

func selfhost_compiler_go_goStringLiteralChars(raw string, index int, out string) string {
	return func() string {
		if index >= len([]rune(raw)) {
			return out
		}
		return selfhost_compiler_go_goStringLiteralChar(raw, index, out)
	}()
}

func selfhost_compiler_go_goStringLiteralChar(raw string, index int, out string) string {
	return func() string {
		switch {
		case []rune(raw)[index] == '\\' && index+1 < len([]rune(raw)) == true:
			return selfhost_compiler_go_goStringLiteralNext(raw, index+2, out+selfhost_compiler_go_goStringLiteralEscaped([]rune(raw)[index+1]))
		default:
			return selfhost_compiler_go_goStringLiteralNext(raw, index+1, out+selfhost_compiler_go_goStringLiteralPlain([]rune(raw)[index]))
		}
	}()
}

func selfhost_compiler_go_goStringLiteralEscaped(ch rune) string {
	return func() string {
		if ch == 'n' {
			return "\\n"
		}
		return func() string {
			if ch == 't' {
				return "\\t"
			}
			return func() string {
				if ch == 'r' {
					return "\\r"
				}
				return func() string {
					if ch == '\\' {
						return "\\\\"
					}
					return func() string {
						if ch == '"' {
							return "\\\""
						}
						return selfhost_compiler_go_goStringLiteralPlain(ch)
					}()
				}()
			}()
		}()
	}()
}

func selfhost_compiler_go_goStringLiteralPlain(ch rune) string {
	return func() string {
		if ch == '"' {
			return "\\\""
		}
		return func() string {
			if ch == '\\' {
				return "\\\\"
			}
			return func() string {
				if ch == '\n' {
					return "\\n"
				}
				return func() string {
					if ch == '\t' {
						return "\\t"
					}
					return func() string {
						if ch == '\r' {
							return "\\r"
						}
						return string(ch)
					}()
				}()
			}()
		}()
	}()
}

func selfhost_compiler_go_goStringLiteralNext(raw string, index int, out string) string {
	return selfhost_compiler_go_goStringLiteralChars(raw, index, out)
}

func selfhost_compiler_go_emitGoTemplateHelper() string {
	return "func __runeTemplateString(value any) string {\n\tswitch v := value.(type) {\n\tcase nil:\n\t\treturn \"null\"\n\tcase rune:\n\t\treturn string(v)\n\tdefault:\n\t\treturn fmt.Sprint(v)\n\t}\n}\n\n"
}

func selfhost_compiler_go_emitGoObjectExpected(expr IRExpr, expected string) string {
	return func() string {
		switch {
		case expected == "":
			return selfhost_compiler_go_emitGoExpr(expr)
		case expected == "Dynamic":
			return selfhost_compiler_go_emitGoExpr(expr)
		default:
			return selfhost_compiler_go_goType(expected) + "{" + selfhost_compiler_go_emitGoFields(expr.children, 0, "") + "}"
		}
	}()
}

func selfhost_compiler_go_emitGoCallExpected(expr IRExpr, expected string) string {
	args := genericInner(expected, "Result")
	return func() string {
		if args != "" && selfhost_compiler_go_isResultConstructorCall(expr) {
			return selfhost_compiler_go_emitGoExpr(expr.children[0]) + "[" + selfhost_compiler_go_emitGoTypeArgs(args) + "](" + selfhost_compiler_go_emitGoExprListFrom(expr.children, 1, "") + ")"
		}
		return selfhost_compiler_go_emitGoCall(expr)
	}()
}

func selfhost_compiler_go_emitGoCallExpectedForFile(file IRFile, expr IRExpr, expected string) string {
	return func() string {
		switch {
		case moduleCallKey(expr) == "json.parse":
			return selfhost_compiler_go_emitGoJSONParse(file, expr, expected)
		default:
			return selfhost_compiler_go_emitGoCallExpected(expr, expected)
		}
	}()
}

func selfhost_compiler_go_isResultConstructorCall(expr IRExpr) bool {
	return expr.kind == ExprKind_Call && len(expr.children) > 0 && expr.children[0].kind == ExprKind_Identifier && (expr.children[0].name == "Ok" || expr.children[0].name == "Err")
}

func selfhost_compiler_go_emitGoTypeArgs(args string) string {
	return selfhost_compiler_go_emitGoTypeArgList(func() []string { parts := strings.Split(args, ","); return parts }(), 0, "")
}

func selfhost_compiler_go_emitGoTypeArgList(args []string, index int, out string) string {
	return func() string {
		if index >= len(args) {
			return out
		}
		return selfhost_compiler_go_emitGoTypeArgList(args, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_go_goType(strings.TrimSpace(args[index])))
	}()
}

func selfhost_compiler_go_emitGoBlockNoContext(statements []IRExpr, index int, returns bool, returnType string, level int, out string) string {
	return func() string {
		if index >= len(statements) {
			return func() string {
				if returns && len(statements) == 0 {
					return out + line(level, "return "+selfhost_compiler_go_goZero(returnType))
				}
				return out
			}()
		}
		return selfhost_compiler_go_emitGoBlockNoContext(statements, index+1, returns, returnType, level, out+selfhost_compiler_go_emitGoStatementNoContext(statements[index], index == len(statements)-1, returns, level))
	}()
}

func selfhost_compiler_go_emitGoStatementNoContext(expr IRExpr, last bool, returns bool, level int) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Let:
			return line(level, mangleIdent(expr.name)+" := "+selfhost_compiler_go_emitGoExpr(expr.children[0])) + line(level, "_ = "+mangleIdent(expr.name))
		case expr.kind == ExprKind_ObjectDestructure:
			return selfhost_compiler_go_emitGoObjectDestructure(expr, level)
		default:
			return func() string {
				if last && returns {
					return line(level, "return "+selfhost_compiler_go_emitGoExpr(expr))
				}
				return line(level, selfhost_compiler_go_emitGoExpr(expr))
			}()
		}
	}()
}

func selfhost_compiler_go_emitGoBinary(expr IRExpr) string {
	return func() string {
		if expr.op == "??" {
			return selfhost_compiler_go_emitGoNullCoalesce(expr, expr.text)
		}
		return selfhost_compiler_go_emitGoExpr(expr.children[0]) + " " + selfhost_compiler_go_goBinaryOp(expr.op) + " " + selfhost_compiler_go_emitGoExpr(expr.children[1])
	}()
}

func selfhost_compiler_go_emitGoBinaryExpected(expr IRExpr, expected string) string {
	return func() string {
		if expr.op == "??" {
			return selfhost_compiler_go_emitGoNullCoalesce(expr, expected)
		}
		return selfhost_compiler_go_emitGoBinary(expr)
	}()
}

func selfhost_compiler_go_emitGoBinaryExpectedForFile(file IRFile, expr IRExpr, expected string) string {
	return func() string {
		if expr.op == "??" {
			return selfhost_compiler_go_emitGoNullCoalesceForFile(file, expr, expected)
		}
		return selfhost_compiler_go_emitGoBinary(expr)
	}()
}

func selfhost_compiler_go_emitGoNullCoalesce(expr IRExpr, expected string) string {
	resultType := selfhost_compiler_go_goCoalesceResultType(expr, expected)
	return func() string {
		if len(expr.children) < 2 || expr.children[0].kind == ExprKind_Null {
			return selfhost_compiler_go_emitGoExpr(expr.children[1])
		}
		return func() string {
			if resultType == "" {
				return "func() any { __coalesce := " + selfhost_compiler_go_emitGoExpr(expr.children[0]) + "; if __coalesce != nil { return __coalesce }; return " + selfhost_compiler_go_emitGoExpr(expr.children[1]) + " }()"
			}
			return "func() " + selfhost_compiler_go_goType(resultType) + " { __coalesce := " + selfhost_compiler_go_emitGoExpr(expr.children[0]) + "; if __coalesce != nil { return __coalesce.(" + selfhost_compiler_go_goType(resultType) + ") }; return " + selfhost_compiler_go_emitGoExprExpected(expr.children[1], resultType) + " }()"
		}()
	}()
}

func selfhost_compiler_go_emitGoNullCoalesceForFile(file IRFile, expr IRExpr, expected string) string {
	resultType := selfhost_compiler_go_goCoalesceResultType(expr, expected)
	return func() string {
		if len(expr.children) < 2 || expr.children[0].kind == ExprKind_Null {
			return selfhost_compiler_go_emitGoExpr(expr.children[1])
		}
		return func() string {
			if resultType == "" {
				return "func() any { __coalesce := " + selfhost_compiler_go_emitGoExpr(expr.children[0]) + "; if __coalesce != nil { return __coalesce }; return " + selfhost_compiler_go_emitGoExpr(expr.children[1]) + " }()"
			}
			return "func() " + selfhost_compiler_go_goType(resultType) + " { __coalesce := " + selfhost_compiler_go_emitGoExpr(expr.children[0]) + "; if __coalesce != nil { return __coalesce.(" + selfhost_compiler_go_goType(resultType) + ") }; return " + selfhost_compiler_go_emitGoExprExpectedForFile(file, expr.children[1], resultType) + " }()"
		}()
	}()
}

func selfhost_compiler_go_goCoalesceResultType(expr IRExpr, expected string) string {
	candidate := func() string {
		if expected != "" && expected != "Dynamic" {
			return expected
		}
		return expr.text
	}()
	return func() string {
		if strings.HasSuffix(candidate, "?") {
			return func() string { runes := []rune(candidate); return string(runes[0 : len([]rune(candidate))-1]) }()
		}
		return func() string {
			if candidate == "Null" {
				return ""
			}
			return candidate
		}()
	}()
}

func selfhost_compiler_go_emitGoTernary(expr IRExpr) string {
	resultType := expr.text
	return func() string {
		if strings.HasPrefix(resultType, "Func[") {
			return "func() " + selfhost_compiler_go_goFunctionType(resultType) + " { if " + selfhost_compiler_go_emitGoExpr(expr.children[0]) + " { return " + selfhost_compiler_go_emitGoExpr(expr.children[1]) + " }; return " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + " }()"
		}
		return "func() any { if " + selfhost_compiler_go_emitGoExpr(expr.children[0]) + " { return " + selfhost_compiler_go_emitGoExpr(expr.children[1]) + " }; return " + func() string {
			if len(expr.children) > 2 {
				return selfhost_compiler_go_emitGoExpr(expr.children[2])
			}
			return "nil"
		}() + " }()"
	}()
}

func selfhost_compiler_go_goFunctionType(typeName string) string {
	parts := func() []string {
		parts := strings.Split((func() string { runes := []rune(typeName); return string(runes[5 : len([]rune(typeName))-1]) }()), "|")
		return parts
	}()
	last := len(parts) - 1
	return "func(" + selfhost_compiler_go_goFunctionParamTypes(parts, 0, last, "") + ") " + selfhost_compiler_go_goType(parts[last])
}

func selfhost_compiler_go_goFunctionParamTypes(parts []string, index int, last int, out string) string {
	return func() string {
		if index >= last {
			return out
		}
		return selfhost_compiler_go_goFunctionParamTypes(parts, index+1, last, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_go_goType(parts[index]))
	}()
}

func selfhost_compiler_go_emitGoAssign(expr IRExpr) string {
	return func() string {
		if len(expr.children) == 2 {
			return selfhost_compiler_go_emitGoExpr(expr.children[0]) + " = " + selfhost_compiler_go_emitGoExpr(expr.children[1])
		}
		return mangleIdent(expr.name) + " = " + selfhost_compiler_go_emitGoExpr(expr.children[0])
	}()
}

func selfhost_compiler_go_emitGoCall(expr IRExpr) string {
	return func() string {
		switch {
		case moduleCallKey(expr) == "io.println":
			return "fmt.Println(" + selfhost_compiler_go_emitGoExprListFrom(expr.children, 1, "") + ")"
		case moduleCallKey(expr) == "json.stringify":
			return selfhost_compiler_go_emitGoJSONStringify(expr)
		case moduleCallKey(expr) == "json.parse":
			return selfhost_compiler_go_emitGoJSONParseDynamic(expr)
		case moduleCallKey(expr) == "map.new":
			return "map[any]any{}"
		case moduleCallKey(expr) == "set.new":
			return "map[any]struct{}{}"
		case moduleCallKey(expr) == "path.isAbsolute":
			return "strings.HasPrefix(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ", \"/\")"
		case moduleCallKey(expr) == "path.basename":
			return "__runePathBasename(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.extname":
			return "__runePathExtname(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.dirname":
			return "__runePathDirname(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.join":
			return "__runePathJoin(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.normalize":
			return "__runePathNormalize(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.resolve":
			return "__runePathResolve(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.relative":
			return "__runePathRelative(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ")"
		case moduleCallKey(expr) == "path.joinParts":
			return "__runePathJoinParts(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[3]) + ")"
		case moduleCallKey(expr) == "path.appendPathPart":
			return "__runePathAppendPart(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ")"
		case moduleCallKey(expr) == "path.normalizeParts":
			return "__runePathNormalizeParts(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[3]) + ", __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[4]) + "))"
		case moduleCallKey(expr) == "path.normalizePart":
			return "__runePathNormalizeParts(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[3]) + ", __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[4]) + "))"
		case moduleCallKey(expr) == "path.normalizeParent":
			return "__runePathNormalizeParent(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[3]) + ", __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[4]) + "))"
		case moduleCallKey(expr) == "path.normalizePop":
			return "__runePathNormalizePop(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[3]) + ", __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[4]) + "))"
		case moduleCallKey(expr) == "path.normalizePush":
			return "__runePathNormalizePush(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[3]) + ", __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[4]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[5]) + ")"
		case moduleCallKey(expr) == "path.pathParts":
			return "__runePathParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.collectPathParts":
			return "__runePathCollectParts(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ", __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[3]) + "))"
		case moduleCallKey(expr) == "path.collectPathPart":
			return "__runePathCollectPart(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + ", __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[3]) + "))"
		case moduleCallKey(expr) == "path.relativeFromParts":
			return "__runePathRelativeFromParts(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[2]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[3]) + ")"
		case moduleCallKey(expr) == "path.relativeTail":
			return "__runePathRelativeTail(__runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), __runePathStringParts(" + selfhost_compiler_go_emitGoExpr(expr.children[2]) + "), " + selfhost_compiler_go_emitGoExpr(expr.children[3]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[4]) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[5]) + ")"
		case moduleCallKey(expr) == "process.platform":
			return "runtime.GOOS"
		case moduleCallKey(expr) == "process.cwd":
			return "\".\""
		case moduleCallKey(expr) == "process.env":
			return "(*string)(nil)"
		case moduleCallKey(expr) == "process.argv":
			return "append([]string(nil), os.Args[1:]...)"
		case moduleCallKey(expr) == "process.exit":
			return "func() struct{} { os.Exit(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "); return struct{}{} }()"
		case moduleCallKey(expr) == "int.toString":
			return "strconv.Itoa(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int.toDouble":
			return "float64(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int.toBigInt":
			return "int64(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int4.fromInt":
			return "func() int8 { n := (" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ") & 0xf; if n >= 8 { return int8(n - 16) }; return int8(n) }()"
		case moduleCallKey(expr) == "int8.fromInt":
			return "func() int8 { n := int(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "); return int8(n) }()"
		case moduleCallKey(expr) == "int16.fromInt":
			return "func() int16 { n := int(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "); return int16(n) }()"
		case moduleCallKey(expr) == "int64.fromInt":
			return "int64(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "uint.fromInt":
			return "uint(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "uint8.fromInt":
			return "func() uint8 { n := int(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "); return uint8(n) }()"
		case moduleCallKey(expr) == "uint16.fromInt":
			return "func() uint16 { n := int(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "); return uint16(n) }()"
		case moduleCallKey(expr) == "uint64.fromInt":
			return "uint64(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "float.fromDouble":
			return "float32(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case (moduleCallKey(expr) == "int4.toInt") || (moduleCallKey(expr) == "int8.toInt") || (moduleCallKey(expr) == "int16.toInt") || (moduleCallKey(expr) == "int64.toInt") || (moduleCallKey(expr) == "uint.toInt") || (moduleCallKey(expr) == "uint8.toInt") || (moduleCallKey(expr) == "uint16.toInt") || (moduleCallKey(expr) == "uint64.toInt"):
			return "int(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "float.toDouble":
			return "float64(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "bigint.fromInt":
			return "int64(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "bigint.toString":
			return "strconv.FormatInt(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ", 10)"
		case moduleCallKey(expr) == "bigint.toDouble":
			return "float64(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "double.trunc":
			return "int(math.Trunc(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "))"
		case moduleCallKey(expr) == "double.floor":
			return "int(math.Floor(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "))"
		case moduleCallKey(expr) == "double.ceil":
			return "int(math.Ceil(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "))"
		case moduleCallKey(expr) == "double.round":
			return "int(math.Round(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "))"
		default:
			return selfhost_compiler_go_emitGoMaybeCoreMethodCall(expr)
		}
	}()
}

func selfhost_compiler_go_fileUsesDoubleMath(file IRFile) bool {
	return fileUsesModuleCall(file, "double.trunc") || fileUsesModuleCall(file, "double.floor") || fileUsesModuleCall(file, "double.ceil") || fileUsesModuleCall(file, "double.round")
}

func selfhost_compiler_go_fileUsesGoStrings(file IRFile) bool {
	return fileUsesModuleCall(file, "path.isAbsolute") || fileUsesPathFamily(file)
}

func selfhost_compiler_go_emitGoMaybeCoreMethodCall(expr IRExpr) string {
	return func() string {
		if len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector {
			return selfhost_compiler_go_emitGoCoreMethodCall(expr, expr.children[0])
		}
		return selfhost_compiler_go_emitGoDefaultCall(expr)
	}()
}

func selfhost_compiler_go_emitGoCoreMethodCall(expr IRExpr, selector IRExpr) string {
	return func() string {
		if len(selector.children) > 0 && selector.children[0].kind != ExprKind_At {
			return func() string {
				switch {
				case (selector.name == "length") || (selector.name == "byteLength"):
					return selfhost_compiler_go_emitGoCoreLength(selector.children[0])
				case selector.name == "isEmpty":
					return "(" + selfhost_compiler_go_emitGoCoreLength(selector.children[0]) + ") == 0"
				case selector.name == "at":
					return selfhost_compiler_go_emitGoCoreAt(expr, selector.children[0])
				case selector.name == "slice":
					return selfhost_compiler_go_emitGoCoreSlice(expr, selector.children[0])
				case selector.name == "push":
					return selfhost_compiler_go_emitGoCorePush(expr, selector.children[0])
				case selector.name == "set":
					return selfhost_compiler_go_emitGoCoreSet(expr, selector.children[0])
				case selector.name == "getOr":
					return selfhost_compiler_go_emitGoCoreGetOr(expr, selector.children[0])
				case (selector.name == "toDouble") || (selector.name == "toInt"):
					return selfhost_compiler_go_emitGoCoreNumericConversion(selector.children[0], selector.name)
				case selector.name == "toString":
					return selfhost_compiler_go_emitGoCoreToString(selector.children[0])
				default:
					return selfhost_compiler_go_emitGoDefaultCall(expr)
				}
			}()
		}
		return selfhost_compiler_go_emitGoDefaultCall(expr)
	}()
}

func selfhost_compiler_go_emitGoCoreLength(receiver IRExpr) string {
	return func() string {
		if receiver.text == "String" {
			return "len([]rune(" + selfhost_compiler_go_emitGoExpr(receiver) + "))"
		}
		return "len(" + selfhost_compiler_go_emitGoExpr(receiver) + ")"
	}()
}

func selfhost_compiler_go_emitGoCoreAt(expr IRExpr, receiver IRExpr) string {
	return func() string {
		if receiver.text == "String" {
			return "[]rune(" + selfhost_compiler_go_emitGoExpr(receiver) + ")[" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "]"
		}
		return selfhost_compiler_go_emitGoExpr(receiver) + "[" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "]"
	}()
}

func selfhost_compiler_go_emitGoCoreSlice(expr IRExpr, receiver IRExpr) string {
	return func() string {
		if receiver.text == "String" {
			return "func() string { runes := []rune(" + selfhost_compiler_go_emitGoExpr(receiver) + "); return string(runes[" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ":" + selfhost_compiler_go_emitGoExpr(expr.children[2]) + "]) }()"
		}
		return selfhost_compiler_go_emitGoExpr(receiver) + "[" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + ":" + selfhost_compiler_go_emitGoExpr(expr.children[2]) + "]"
	}()
}

func selfhost_compiler_go_emitGoCorePush(expr IRExpr, receiver IRExpr) string {
	return "func() int { " + selfhost_compiler_go_emitGoExpr(receiver) + " = append(" + selfhost_compiler_go_emitGoExpr(receiver) + ", " + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "); return len(" + selfhost_compiler_go_emitGoExpr(receiver) + ") }()"
}

func selfhost_compiler_go_emitGoCoreSet(expr IRExpr, receiver IRExpr) string {
	return "func() any { " + selfhost_compiler_go_emitGoExpr(receiver) + "[" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "] = " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + "; return " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + " }()"
}

func selfhost_compiler_go_emitGoCoreGetOr(expr IRExpr, receiver IRExpr) string {
	resultType := selfhost_compiler_go_goCoalesceResultType(expr, expr.text)
	return func() string {
		if resultType == "" {
			return "func() any { if __value, ok := " + selfhost_compiler_go_emitGoExpr(receiver) + "[" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "]; ok { return __value }; return " + selfhost_compiler_go_emitGoExpr(expr.children[2]) + " }()"
		}
		return "func() " + selfhost_compiler_go_goType(resultType) + " { if __value, ok := " + selfhost_compiler_go_emitGoExpr(receiver) + "[" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "]; ok { return __value.(" + selfhost_compiler_go_goType(resultType) + ") }; return " + selfhost_compiler_go_emitGoExprExpected(expr.children[2], resultType) + " }()"
	}()
}

func selfhost_compiler_go_emitGoDefaultCall(expr IRExpr) string {
	return selfhost_compiler_go_emitGoExpr(expr.children[0]) + "(" + selfhost_compiler_go_emitGoExprListFrom(expr.children, 1, "") + ")"
}

func selfhost_compiler_go_emitGoCoreNumericConversion(receiver IRExpr, method string) string {
	return func() string {
		if method == "toDouble" {
			return "float64(" + selfhost_compiler_go_emitGoExpr(receiver) + ")"
		}
		return func() string {
			if method == "toInt" {
				return "int(" + selfhost_compiler_go_emitGoExpr(receiver) + ")"
			}
			return selfhost_compiler_go_emitGoExpr(receiver)
		}()
	}()
}

func selfhost_compiler_go_emitGoCoreToString(receiver IRExpr) string {
	return func() string {
		switch {
		case receiver.text == "Bool":
			return "strconv.FormatBool(" + selfhost_compiler_go_emitGoExpr(receiver) + ")"
		case (receiver.text == "Int") || (receiver.text == "int4") || (receiver.text == "int8") || (receiver.text == "int16") || (receiver.text == "int64") || (receiver.text == "uint") || (receiver.text == "uint8") || (receiver.text == "uint16") || (receiver.text == "uint64"):
			return "strconv.Itoa(int(" + selfhost_compiler_go_emitGoExpr(receiver) + "))"
		case receiver.text == "Double":
			return "strconv.FormatFloat(" + selfhost_compiler_go_emitGoExpr(receiver) + ", 'f', -1, 64)"
		case receiver.text == "Float":
			return "strconv.FormatFloat(float64(" + selfhost_compiler_go_emitGoExpr(receiver) + "), 'f', -1, 32)"
		case receiver.text == "Char":
			return "string(" + selfhost_compiler_go_emitGoExpr(receiver) + ")"
		case receiver.text == "BigInt":
			return "strconv.FormatInt(int64(" + selfhost_compiler_go_emitGoExpr(receiver) + "), 10)"
		default:
			return selfhost_compiler_go_emitGoExpr(receiver)
		}
	}()
}

func selfhost_compiler_go_emitGoLambda(expr IRExpr) string {
	returnType := func() string {
		if strings.HasPrefix(expr.text, "Func[") {
			return selfhost_compiler_go_goFunctionReturnType(expr.text)
		}
		return "any"
	}()
	return "func(" + selfhost_compiler_go_emitGoParams(expr.params, 0, "") + ") " + selfhost_compiler_go_goType(returnType) + " { return " + selfhost_compiler_go_emitGoExpr(expr.children[0]) + " }"
}

func selfhost_compiler_go_goFunctionReturnType(typeName string) string {
	parts := func() []string {
		parts := strings.Split((func() string { runes := []rune(typeName); return string(runes[5 : len([]rune(typeName))-1]) }()), "|")
		return parts
	}()
	return parts[len(parts)-1]
}

func selfhost_compiler_go_emitGoSelector(expr IRExpr) string {
	return func() string {
		switch {
		case expr.children[0].kind == ExprKind_At:
			return selfhost_compiler_go_emitGoAtSelector(expr)
		case expr.children[0].kind == ExprKind_Identifier:
			return func() string {
				if selfhost_compiler_go_looksLikeTypeName(expr.children[0].name) {
					return mangleIdent(expr.children[0].name + "_" + expr.name)
				}
				return selfhost_compiler_go_emitGoExpr(expr.children[0]) + "." + mangleIdent(expr.name)
			}()
		default:
			return selfhost_compiler_go_emitGoExpr(expr.children[0]) + "." + mangleIdent(expr.name)
		}
	}()
}

func selfhost_compiler_go_emitGoAtSelector(expr IRExpr) string {
	importPath := compilerIRAtImportPath(expr.children[0])
	goPath := compilerGoPackageImportPath(importPath)
	return func() string {
		switch {
		case goPath == "":
			return func() string {
				imported := importPath != ""
				return func() string {
					switch {
					case imported == true:
						return mangleIdent(expr.name)
					default:
						return expr.children[0].name + "." + expr.name
					}
				}()
			}()
		default:
			return compilerGoPackageName(goPath) + "." + expr.name
		}
	}()
}

func selfhost_compiler_go_emitGoExprList(exprs []IRExpr, index int, out string) string {
	return selfhost_compiler_go_emitGoExprListFrom(exprs, index, out)
}

func selfhost_compiler_go_emitGoExprListFrom(exprs []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(exprs) {
			return out
		}
		return selfhost_compiler_go_emitGoExprListFrom(exprs, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+selfhost_compiler_go_emitGoExpr(exprs[index]))
	}()
}

func selfhost_compiler_go_emitGoMapEntries(entries []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(entries) {
			return out
		}
		return selfhost_compiler_go_emitGoMapEntries(entries, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+selfhost_compiler_go_emitGoExpr(entries[index].children[0])+": "+selfhost_compiler_go_emitGoExpr(entries[index].children[1]))
	}()
}

func selfhost_compiler_go_emitGoFields(fields []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(fields) {
			return out
		}
		return selfhost_compiler_go_emitGoFields(fields, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+mangleIdent(fields[index].name)+": "+selfhost_compiler_go_emitGoStructFieldValue(fields[index].children[0]))
	}()
}

func selfhost_compiler_go_emitGoStructFieldValue(expr IRExpr) string {
	return func() string {
		if expr.kind == ExprKind_Array && len(expr.children) == 0 {
			return "nil"
		}
		return selfhost_compiler_go_emitGoExpr(expr)
	}()
}

func selfhost_compiler_go_goBinaryOp(op string) string {
	return op
}

func selfhost_compiler_go_hasMain(file IRFile) bool {
	return selfhost_compiler_go_hasFunction(file.functions, "main", 0)
}

func selfhost_compiler_go_hasFunction(functions []IRFunction, name string, index int) bool {
	return func() bool {
		if index >= len(functions) {
			return false
		}
		return func() bool {
			if functions[index].macro == false && functions[index].name == name {
				return true
			}
			return selfhost_compiler_go_hasFunction(functions, name, index+1)
		}()
	}()
}

func selfhost_compiler_go_usesPrintFile(file IRFile) bool {
	return selfhost_compiler_go_functionsUsePrint(file.functions, 0) || selfhost_compiler_go_structsUsePrint(file.structs, 0) || selfhost_compiler_go_enumsUsePrint(file.enums, 0)
}

func selfhost_compiler_go_fileUsesTemplate(file IRFile) bool {
	return selfhost_compiler_go_functionsUseTemplate(file.functions, 0) || selfhost_compiler_go_structsUseTemplate(file.structs, 0) || selfhost_compiler_go_enumsUseTemplate(file.enums, 0)
}

func selfhost_compiler_go_functionsUseTemplate(functions []IRFunction, index int) bool {
	return func() bool {
		if index >= len(functions) {
			return false
		}
		return func() bool {
			if functions[index].macro {
				return selfhost_compiler_go_functionsUseTemplate(functions, index+1)
			}
			return func() bool {
				if selfhost_compiler_go_exprUsesTemplate(functions[index].body) {
					return true
				}
				return selfhost_compiler_go_functionsUseTemplate(functions, index+1)
			}()
		}()
	}()
}

func selfhost_compiler_go_structsUseTemplate(structs []IRStructType, index int) bool {
	return func() bool {
		if index >= len(structs) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_functionsUseTemplate(structs[index].methods, 0) {
				return true
			}
			return selfhost_compiler_go_structsUseTemplate(structs, index+1)
		}()
	}()
}

func selfhost_compiler_go_enumsUseTemplate(enums []IREnumType, index int) bool {
	return func() bool {
		if index >= len(enums) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_functionsUseTemplate(enums[index].methods, 0) {
				return true
			}
			return selfhost_compiler_go_enumsUseTemplate(enums, index+1)
		}()
	}()
}

func selfhost_compiler_go_exprUsesTemplate(expr IRExpr) bool {
	return expr.kind == ExprKind_Template || selfhost_compiler_go_exprChildrenUseTemplate(expr.children, 0)
}

func selfhost_compiler_go_exprChildrenUseTemplate(children []IRExpr, index int) bool {
	return func() bool {
		if index >= len(children) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_exprUsesTemplate(children[index]) {
				return true
			}
			return selfhost_compiler_go_exprChildrenUseTemplate(children, index+1)
		}()
	}()
}

func selfhost_compiler_go_fileUsesToString(file IRFile) bool {
	return selfhost_compiler_go_functionsUseToString(file.functions, 0) || selfhost_compiler_go_structsUseToString(file.structs, 0) || selfhost_compiler_go_enumsUseToString(file.enums, 0)
}

func selfhost_compiler_go_functionsUseToString(functions []IRFunction, index int) bool {
	return func() bool {
		if index >= len(functions) {
			return false
		}
		return func() bool {
			if functions[index].macro {
				return selfhost_compiler_go_functionsUseToString(functions, index+1)
			}
			return func() bool {
				if selfhost_compiler_go_exprUsesToString(functions[index].body) {
					return true
				}
				return selfhost_compiler_go_functionsUseToString(functions, index+1)
			}()
		}()
	}()
}

func selfhost_compiler_go_structsUseToString(structs []IRStructType, index int) bool {
	return func() bool {
		if index >= len(structs) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_functionsUseToString(structs[index].methods, 0) {
				return true
			}
			return selfhost_compiler_go_structsUseToString(structs, index+1)
		}()
	}()
}

func selfhost_compiler_go_enumsUseToString(enums []IREnumType, index int) bool {
	return func() bool {
		if index >= len(enums) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_functionsUseToString(enums[index].methods, 0) {
				return true
			}
			return selfhost_compiler_go_enumsUseToString(enums, index+1)
		}()
	}()
}

func selfhost_compiler_go_exprUsesToString(expr IRExpr) bool {
	return selfhost_compiler_go_exprIsScalarToStringCall(expr) || selfhost_compiler_go_exprChildrenUseToString(expr.children, 0)
}

func selfhost_compiler_go_exprIsScalarToStringCall(expr IRExpr) bool {
	return func() bool {
		if expr.kind == ExprKind_Call && len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector && expr.children[0].name == "toString" && len(expr.children[0].children) > 0 {
			return selfhost_compiler_go_goScalarToStringReceiver(expr.children[0].children[0].text)
		}
		return false
	}()
}

func selfhost_compiler_go_goScalarToStringReceiver(typeName string) bool {
	return typeName == "Bool" || typeName == "Int" || typeName == "int4" || typeName == "int8" || typeName == "int16" || typeName == "int64" || typeName == "uint" || typeName == "uint8" || typeName == "uint16" || typeName == "uint64" || typeName == "Double" || typeName == "Float" || typeName == "Char" || typeName == "BigInt"
}

func selfhost_compiler_go_exprChildrenUseToString(children []IRExpr, index int) bool {
	return func() bool {
		if index >= len(children) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_exprUsesToString(children[index]) {
				return true
			}
			return selfhost_compiler_go_exprChildrenUseToString(children, index+1)
		}()
	}()
}

func selfhost_compiler_go_functionsUsePrint(functions []IRFunction, index int) bool {
	return func() bool {
		if index >= len(functions) {
			return false
		}
		return func() bool {
			if functions[index].macro {
				return selfhost_compiler_go_functionsUsePrint(functions, index+1)
			}
			return func() bool {
				if selfhost_compiler_go_exprUsesPrint(functions[index].body) {
					return true
				}
				return selfhost_compiler_go_functionsUsePrint(functions, index+1)
			}()
		}()
	}()
}

func selfhost_compiler_go_structsUsePrint(structs []IRStructType, index int) bool {
	return func() bool {
		if index >= len(structs) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_functionsUsePrint(structs[index].methods, 0) {
				return true
			}
			return selfhost_compiler_go_structsUsePrint(structs, index+1)
		}()
	}()
}

func selfhost_compiler_go_enumsUsePrint(enums []IREnumType, index int) bool {
	return func() bool {
		if index >= len(enums) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_functionsUsePrint(enums[index].methods, 0) {
				return true
			}
			return selfhost_compiler_go_enumsUsePrint(enums, index+1)
		}()
	}()
}

func selfhost_compiler_go_exprUsesPrint(expr IRExpr) bool {
	return func() bool {
		switch {
		case moduleCallKey(expr) == "io.println":
			return true
		default:
			return selfhost_compiler_go_exprChildrenUsePrint(expr.children, 0)
		}
	}()
}

func selfhost_compiler_go_exprChildrenUsePrint(children []IRExpr, index int) bool {
	return func() bool {
		if index >= len(children) {
			return false
		}
		return func() bool {
			if selfhost_compiler_go_exprUsesPrint(children[index]) {
				return true
			}
			return selfhost_compiler_go_exprChildrenUsePrint(children, index+1)
		}()
	}()
}

func selfhost_compiler_go_emitGoJSONStringify(expr IRExpr) string {
	return "func() string { __bytes, _ := json.Marshal(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "); return string(__bytes) }()"
}

func selfhost_compiler_go_emitGoJSONParseDynamic(expr IRExpr) string {
	return "func() any { var __raw any; if err := json.Unmarshal([]byte(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), &__raw); err != nil { panic(err) }; return __raw }()"
}

func selfhost_compiler_go_emitGoJSONParse(file IRFile, expr IRExpr, expected string) string {
	target := expected
	func() {
		switch {
		case target == "" == true:
			target = "Dynamic"
			return
		default:
			target = target
			return
		}
	}()
	dynamic := target == "Dynamic"
	return func() string {
		switch {
		case dynamic == true:
			return selfhost_compiler_go_emitGoJSONParseDynamic(expr)
		default:
			return selfhost_compiler_go_emitGoJSONParseTyped(file, expr, target)
		}
	}()
}

func selfhost_compiler_go_emitGoJSONParseTyped(file IRFile, expr IRExpr, target string) string {
	decls := selfhost_compiler_go_goJSONDeclsForType(file, target, []string{})
	wireType := selfhost_compiler_go_goJSONDecodeType(file, target)
	value := selfhost_compiler_go_goJSONDecodedValue(file, "__raw", target)
	return "func() " + selfhost_compiler_go_goType(target) + " { " + decls.text + "var __raw " + wireType + "; if err := json.Unmarshal([]byte(" + selfhost_compiler_go_emitGoExpr(expr.children[1]) + "), &__raw); err != nil { panic(err) }; return " + value + " }()"
}

func selfhost_compiler_go_emptyGoJSONDeclResult(names []string, text string) GoJSONDeclResult {
	return GoJSONDeclResult{names: names, text: text}
}

func selfhost_compiler_go_goJSONDecodeType(file IRFile, typeName string) string {
	arrayInner := genericInner(typeName, "Array")
	return func() string {
		switch {
		case arrayInner != "" == true:
			return "[]" + selfhost_compiler_go_goJSONDecodeType(file, arrayInner)
		default:
			return selfhost_compiler_go_goJSONDecodeNonArrayType(file, typeName)
		}
	}()
}

func selfhost_compiler_go_goJSONDecodeNonArrayType(file IRFile, typeName string) string {
	structIndex := selfhost_compiler_go_goJSONStructIndex(file.structs, typeName, 0)
	return func() string {
		switch {
		case structIndex >= 0 == true:
			return selfhost_compiler_go_goJSONStructWireType(file.structs[structIndex])
		default:
			return selfhost_compiler_go_goJSONScalarDecodeType(typeName)
		}
	}()
}

func selfhost_compiler_go_goJSONScalarDecodeType(typeName string) string {
	return func() string {
		switch {
		case typeName == "Char":
			return "string"
		default:
			return selfhost_compiler_go_goType(typeName)
		}
	}()
}

func selfhost_compiler_go_goJSONStructWireType(typeDecl IRStructType) string {
	return mangleIdent("rune_json_" + typeDecl.name)
}

func selfhost_compiler_go_goJSONDeclsForType(file IRFile, typeName string, names []string) GoJSONDeclResult {
	arrayInner := genericInner(typeName, "Array")
	return func() GoJSONDeclResult {
		switch {
		case arrayInner != "" == true:
			return selfhost_compiler_go_goJSONDeclsForType(file, arrayInner, names)
		default:
			return selfhost_compiler_go_goJSONDeclsForNonArrayType(file, typeName, names)
		}
	}()
}

func selfhost_compiler_go_goJSONDeclsForNonArrayType(file IRFile, typeName string, names []string) GoJSONDeclResult {
	structIndex := selfhost_compiler_go_goJSONStructIndex(file.structs, typeName, 0)
	return func() GoJSONDeclResult {
		switch {
		case structIndex >= 0 == true:
			return selfhost_compiler_go_goJSONDeclsForStruct(file, file.structs[structIndex], names)
		default:
			return selfhost_compiler_go_emptyGoJSONDeclResult(names, "")
		}
	}()
}

func selfhost_compiler_go_goJSONDeclsForStruct(file IRFile, typeDecl IRStructType, names []string) GoJSONDeclResult {
	name := selfhost_compiler_go_goJSONStructWireType(typeDecl)
	seen := selfhost_compiler_go_goJSONDeclSeen(names, name, 0)
	return func() GoJSONDeclResult {
		switch {
		case seen == true:
			return selfhost_compiler_go_emptyGoJSONDeclResult(names, "")
		default:
			return selfhost_compiler_go_goJSONDeclsForNewStruct(file, typeDecl, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, names...)
				__rune_spread_out = append(__rune_spread_out, name)
				return __rune_spread_out
			}())
		}
	}()
}

func selfhost_compiler_go_goJSONDeclsForNewStruct(file IRFile, typeDecl IRStructType, names []string) GoJSONDeclResult {
	nested := selfhost_compiler_go_goJSONDeclsForFields(file, typeDecl.fields, 0, names, "")
	fields := selfhost_compiler_go_goJSONStructWireFields(file, typeDecl.fields, 0, "")
	decl := "type " + selfhost_compiler_go_goJSONStructWireType(typeDecl) + " struct{" + fields + "}; "
	return selfhost_compiler_go_emptyGoJSONDeclResult(nested.names, nested.text+decl)
}

func selfhost_compiler_go_goJSONDeclsForFields(file IRFile, fields []IRField, index int, names []string, text string) GoJSONDeclResult {
	done := index >= len(fields)
	return func() GoJSONDeclResult {
		switch {
		case done == true:
			return selfhost_compiler_go_emptyGoJSONDeclResult(names, text)
		default:
			return selfhost_compiler_go_goJSONDeclsForField(file, fields, index, names, text)
		}
	}()
}

func selfhost_compiler_go_goJSONDeclsForField(file IRFile, fields []IRField, index int, names []string, text string) GoJSONDeclResult {
	field := fields[index]
	include := selfhost_compiler_go_goJSONIncludeField(field)
	return func() GoJSONDeclResult {
		switch {
		case include == true:
			return func() GoJSONDeclResult {
				nested := selfhost_compiler_go_goJSONDeclsForType(file, field.typeName, names)
				return selfhost_compiler_go_goJSONDeclsForFields(file, fields, index+1, nested.names, text+nested.text)
			}()
		default:
			return selfhost_compiler_go_goJSONDeclsForFields(file, fields, index+1, names, text)
		}
	}()
}

func selfhost_compiler_go_goJSONStructWireFields(file IRFile, fields []IRField, index int, out string) string {
	done := index >= len(fields)
	return func() string {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_go_goJSONStructWireField(file, fields, index, out)
		}
	}()
}

func selfhost_compiler_go_goJSONStructWireField(file IRFile, fields []IRField, index int, out string) string {
	field := fields[index]
	include := selfhost_compiler_go_goJSONIncludeField(field)
	return func() string {
		switch {
		case include == true:
			return func() string {
				prefix := func() string {
					switch {
					case out == "" == true:
						return ""
					default:
						return "; "
					}
				}()
				part := "F" + compilerIntToString(index) + " " + selfhost_compiler_go_goJSONDecodeType(file, field.typeName) + " " + selfhost_compiler_go_goJSONTag(field.jsonName)
				return selfhost_compiler_go_goJSONStructWireFields(file, fields, index+1, out+prefix+part)
			}()
		default:
			return selfhost_compiler_go_goJSONStructWireFields(file, fields, index+1, out)
		}
	}()
}

func selfhost_compiler_go_goJSONTag(name string) string {
	return "`json:\"" + strings.ReplaceAll((strings.ReplaceAll(name, "\\", "\\\\")), "\"", "\\\"") + "\"`"
}

func selfhost_compiler_go_goJSONIncludeField(field IRField) bool {
	omit := field.jsonIgnore || selfhost_compiler_go_goJSONOmitType(field.typeName)
	return func() bool {
		switch {
		case omit == true:
			return false
		default:
			return true
		}
	}()
}

func selfhost_compiler_go_goJSONOmitType(typeName string) bool {
	return func() bool {
		switch {
		case (typeName == "") || (typeName == "Void") || (typeName == "Symbol"):
			return true
		default:
			return false
		}
	}()
}

func selfhost_compiler_go_goJSONDeclSeen(names []string, name string, index int) bool {
	done := index >= len(names)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return func() bool {
				matched := names[index] == name
				return func() bool {
					switch {
					case matched == true:
						return true
					default:
						return selfhost_compiler_go_goJSONDeclSeen(names, name, index+1)
					}
				}()
			}()
		}
	}()
}

func selfhost_compiler_go_goJSONStructIndex(structs []IRStructType, typeName string, index int) int {
	done := index >= len(structs)
	return func() int {
		switch {
		case done == true:
			return -1
		default:
			return func() int {
				matched := structs[index].name == typeName
				return func() int {
					switch {
					case matched == true:
						return index
					default:
						return selfhost_compiler_go_goJSONStructIndex(structs, typeName, index+1)
					}
				}()
			}()
		}
	}()
}

func selfhost_compiler_go_goJSONDecodedValue(file IRFile, source string, typeName string) string {
	arrayInner := genericInner(typeName, "Array")
	return func() string {
		switch {
		case arrayInner != "" == true:
			return selfhost_compiler_go_goJSONDecodedArray(file, source, typeName, arrayInner)
		default:
			return selfhost_compiler_go_goJSONDecodedNonArrayValue(file, source, typeName)
		}
	}()
}

func selfhost_compiler_go_goJSONDecodedNonArrayValue(file IRFile, source string, typeName string) string {
	structIndex := selfhost_compiler_go_goJSONStructIndex(file.structs, typeName, 0)
	return func() string {
		switch {
		case structIndex >= 0 == true:
			return selfhost_compiler_go_goJSONDecodedStruct(file, source, file.structs[structIndex])
		default:
			return selfhost_compiler_go_goJSONDecodedScalar(source, typeName)
		}
	}()
}

func selfhost_compiler_go_goJSONDecodedScalar(source string, typeName string) string {
	return func() string {
		switch {
		case typeName == "Char":
			return "[]rune(" + source + ")[0]"
		default:
			return source
		}
	}()
}

func selfhost_compiler_go_goJSONDecodedArray(file IRFile, source string, typeName string, inner string) string {
	return "func() " + selfhost_compiler_go_goType(typeName) + " { __out := make(" + selfhost_compiler_go_goType(typeName) + ", len(" + source + "); for __idx, __item := range " + source + " { __out[__idx] = " + selfhost_compiler_go_goJSONDecodedValue(file, "__item", inner) + " }; return __out }()"
}

func selfhost_compiler_go_goJSONDecodedStruct(file IRFile, source string, typeDecl IRStructType) string {
	return "func() " + selfhost_compiler_go_goType(typeDecl.name) + " { __out := " + selfhost_compiler_go_goType(typeDecl.name) + "{}; " + selfhost_compiler_go_goJSONDecodedFieldAssignments(file, source, typeDecl.fields, 0, "") + "return __out }()"
}

func selfhost_compiler_go_goJSONDecodedFieldAssignments(file IRFile, source string, fields []IRField, index int, out string) string {
	done := index >= len(fields)
	return func() string {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_go_goJSONDecodedFieldAssignment(file, source, fields, index, out)
		}
	}()
}

func selfhost_compiler_go_goJSONDecodedFieldAssignment(file IRFile, source string, fields []IRField, index int, out string) string {
	field := fields[index]
	include := selfhost_compiler_go_goJSONIncludeField(field)
	return func() string {
		switch {
		case include == true:
			return func() string {
				sourceField := source + ".F" + compilerIntToString(index)
				value := selfhost_compiler_go_goJSONDecodedValue(file, sourceField, field.typeName)
				assignment := "__out." + mangleIdent(field.name) + " = " + value + "; "
				return selfhost_compiler_go_goJSONDecodedFieldAssignments(file, source, fields, index+1, out+assignment)
			}()
		default:
			return selfhost_compiler_go_goJSONDecodedFieldAssignments(file, source, fields, index+1, out)
		}
	}()
}

func selfhost_compiler_go_goType(typeName string) string {
	return func() string {
		switch {
		case typeName == "":
			return "any"
		case typeName == "Void":
			return "struct{}"
		case typeName == "Int":
			return "int"
		case (typeName == "Int4") || (typeName == "Int8"):
			return "int8"
		case typeName == "Int16":
			return "int16"
		case typeName == "Int64":
			return "int64"
		case typeName == "BigInt":
			return "int64"
		case typeName == "UInt":
			return "uint"
		case typeName == "UInt8":
			return "uint8"
		case typeName == "UInt16":
			return "uint16"
		case typeName == "UInt64":
			return "uint64"
		case typeName == "Double":
			return "float64"
		case typeName == "Float":
			return "float32"
		case typeName == "String":
			return "string"
		case typeName == "Char":
			return "rune"
		case typeName == "Bool":
			return "bool"
		case typeName == "Dynamic":
			return "any"
		case (typeName == "Data") || (typeName == "@io.Data"):
			return "[]byte"
		default:
			return selfhost_compiler_go_goTypeFallback(typeName)
		}
	}()
}

func selfhost_compiler_go_goTypeFallback(typeName string) string {
	return func() string {
		if strings.HasPrefix(typeName, "{") && strings.HasSuffix(typeName, "}") {
			return selfhost_compiler_go_goStructuralObjectType(typeName)
		}
		return func() string {
			if strings.HasSuffix(typeName, "?") {
				return "any"
			}
			return func() string {
				if genericInner(typeName, "Array") != "" {
					return "[]" + selfhost_compiler_go_goType(genericInner(typeName, "Array"))
				}
				return func() string {
					if genericInner(typeName, "ReadonlyArray") != "" {
						return "[]" + selfhost_compiler_go_goType(genericInner(typeName, "ReadonlyArray"))
					}
					return func() string {
						if genericInner(typeName, "Map") != "" {
							return "map[" + selfhost_compiler_go_goType(typeArg(genericInner(typeName, "Map"), 0)) + "]" + selfhost_compiler_go_goType(typeArg(genericInner(typeName, "Map"), 1))
						}
						return func() string {
							if genericInner(typeName, "Set") != "" {
								return "map[" + selfhost_compiler_go_goType(genericInner(typeName, "Set")) + "]struct{}"
							}
							return selfhost_compiler_go_goNamedType(typeName)
						}()
					}()
				}()
			}()
		}()
	}()
}

func selfhost_compiler_go_goNamedType(typeName string) string {
	open := strings.Index(typeName, "[")
	return func() string {
		if open < 0 {
			return mangleIdent(typeName)
		}
		return mangleIdent(func() string { runes := []rune(typeName); return string(runes[0:open]) }()) + "[" + selfhost_compiler_go_emitGoTypeArgs(func() string { runes := []rune(typeName); return string(runes[open+1 : len([]rune(typeName))-1]) }()) + "]"
	}()
}

func selfhost_compiler_go_emitGoGenericsDecl(generics []string) string {
	return func() string {
		if len(generics) == 0 {
			return ""
		}
		return "[" + selfhost_compiler_go_emitGoGenericDeclItems(generics, 0, "") + "]"
	}()
}

func selfhost_compiler_go_emitGoGenericDeclItems(generics []string, index int, out string) string {
	return func() string {
		if index >= len(generics) {
			return out
		}
		return selfhost_compiler_go_emitGoGenericDeclItems(generics, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(generics[index])+" any")
	}()
}

func selfhost_compiler_go_emitGoGenericsUse(generics []string) string {
	return func() string {
		if len(generics) == 0 {
			return ""
		}
		return "[" + selfhost_compiler_go_emitGoGenericUseItems(generics, 0, "") + "]"
	}()
}

func selfhost_compiler_go_emitGoGenericUseItems(generics []string, index int, out string) string {
	return func() string {
		if index >= len(generics) {
			return out
		}
		return selfhost_compiler_go_emitGoGenericUseItems(generics, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(generics[index]))
	}()
}

func selfhost_compiler_go_unwrapPayloadType(file IRFile, expr IRExpr) string {
	return func() string {
		if expr.kind == ExprKind_Unwrap {
			return selfhost_compiler_go_resultPayloadType(selfhost_compiler_go_unwrapSourceType(file, expr.children[0]))
		}
		return ""
	}()
}

func selfhost_compiler_go_unwrapSourceType(file IRFile, expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_go_callReturnType(file, expr)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_go_callReturnType(file IRFile, expr IRExpr) string {
	return func() string {
		if len(expr.children) > 0 && expr.children[0].kind == ExprKind_Identifier {
			return selfhost_compiler_go_functionReturnType(file.functions, expr.children[0].name, 0)
		}
		return ""
	}()
}

func selfhost_compiler_go_functionReturnType(functions []IRFunction, name string, index int) string {
	return func() string {
		if index >= len(functions) {
			return ""
		}
		return func() string {
			if functions[index].macro == false && functions[index].name == name {
				return functions[index].returnType
			}
			return selfhost_compiler_go_functionReturnType(functions, name, index+1)
		}()
	}()
}

func selfhost_compiler_go_resultPayloadType(typeName string) string {
	args := genericInner(typeName, "Result")
	return func() string {
		if args == "" {
			return ""
		}
		return typeArg(args, 0)
	}()
}

func selfhost_compiler_go_goZero(typeName string) string {
	return func() string {
		switch {
		case (typeName == "Int") || (typeName == "Int4") || (typeName == "Int8") || (typeName == "Int16") || (typeName == "Int64") || (typeName == "UInt") || (typeName == "UInt8") || (typeName == "UInt16") || (typeName == "UInt64") || (typeName == "BigInt"):
			return "0"
		case (typeName == "Double") || (typeName == "Float"):
			return "0.0"
		case typeName == "String":
			return "\"\""
		case typeName == "Char":
			return "'\\x00'"
		case typeName == "Bool":
			return "false"
		default:
			return "nil"
		}
	}()
}

func selfhost_compiler_go_looksLikeTypeName(name string) bool {
	return func() bool {
		if len([]rune(name)) > 0 {
			return []rune(name)[0] >= 'A' && []rune(name)[0] <= 'Z'
		}
		return false
	}()
}

func generateMoonBit(file IRFile) string {
	out := ""
	for _, enumDecl := range file.enums {
		_ = enumDecl
		out = out + selfhost_compiler_mbt_emitMoonBitEnum(enumDecl) + "\n"
	}
	for _, enumDecl := range file.enums {
		_ = enumDecl
		out = out + selfhost_compiler_mbt_emitMoonBitEnumMethods(enumDecl)
	}
	if fileUsesUnwrap(file) {
		out = out + selfhost_compiler_mbt_emitMoonBitUnwrapHelper()
	}
	if fileUsesPathFamily(file) {
		out = out + selfhost_compiler_mbt_emitMoonBitPathHelpers()
	}
	for _, typeDecl := range file.structs {
		_ = typeDecl
		out = out + selfhost_compiler_mbt_emitMoonBitStruct(typeDecl) + "\n"
	}
	for _, typeDecl := range file.structs {
		_ = typeDecl
		out = out + selfhost_compiler_mbt_emitMoonBitMethods(typeDecl)
	}
	for _, constant := range file.constants {
		_ = constant
		out = out + selfhost_compiler_mbt_emitMoonBitConst(constant) + "\n"
	}
	for _, fn := range file.functions {
		_ = fn
		out = func() string {
			if fn.macro {
				return out
			}
			return out + selfhost_compiler_mbt_emitMoonBitFunction(fn, "") + "\n"
		}()
	}
	return out
}

func selfhost_compiler_mbt_emitMoonBitConst(constant IRConst) string {
	return "let " + selfhost_compiler_mbt_moonBitValueIdent(constant.name) + " : " + selfhost_compiler_mbt_moonBitType(constant.typeName) + " = " + selfhost_compiler_mbt_emitMoonBitExpr(constant.value)
}

func selfhost_compiler_mbt_emitMoonBitUnwrapHelper() string {
	return "fn[T, E] rune_unwrap(value : RuneResult[T, E]) -> T {\n  match value {\n    RuneOk(value) => value\n    RuneErr(_) => abort(\"Result.Err\")\n  }\n}\n\n"
}

func selfhost_compiler_mbt_emitMoonBitPathHelpers() string {
	return "fn __rune_path_basename(path : String) -> String {\n  let index = path.rev_find(\"/\").unwrap_or(-1)\n  if index < 0 {\n    path\n  } else if index == path.length() - 1 {\n    path\n  } else {\n    path[index + 1:].to_owned()\n  }\n}\n\nfn __rune_path_extname(path : String) -> String {\n  let base = __rune_path_basename(path)\n  let index = base.rev_find(\".\").unwrap_or(-1)\n  if index <= 0 { \"\" } else { base[index:].to_owned() }\n}\n\nfn __rune_path_dirname(path : String) -> String {\n  let index = path.rev_find(\"/\").unwrap_or(-1)\n  if index < 0 {\n    \".\"\n  } else if index == 0 {\n    \"/\"\n  } else {\n    path[:index].to_owned()\n  }\n}\n\nfn __rune_path_join(parts : Array[String]) -> String {\n  __rune_path_normalize(__rune_path_join_parts(parts, 0, \"\"))\n}\n\nfn __rune_path_normalize(path : String) -> String {\n  let absolute = path.has_prefix(\"/\")\n  let pieces = path.split(\"/\").map(fn(part) { part.to_owned() }).to_array()\n  let out = __rune_path_normalize_parts(pieces, 0, absolute, [])\n  let joined = __rune_path_join_parts(out, 0, \"\")\n  if absolute {\n    \"/\" + joined\n  } else if joined.is_empty() {\n    \".\"\n  } else {\n    joined\n  }\n}\n\nfn __rune_path_resolve(parts : Array[String]) -> String {\n  if parts.length() == 0 { \".\" } else { __rune_path_normalize(__rune_path_join(parts)) }\n}\n\nfn __rune_path_relative(from : String, to : String) -> String {\n  let from_parts = __rune_path_parts(__rune_path_resolve([from]))\n  let to_parts = __rune_path_parts(__rune_path_resolve([to]))\n  __rune_path_relative_from_parts(from_parts, to_parts, 0)\n}\n\nfn __rune_path_join_parts(parts : Array[String], index : Int, out : String) -> String {\n  if index >= parts.length() {\n    out\n  } else {\n    __rune_path_join_parts(parts, index + 1, __rune_path_append_part(out, parts[index]))\n  }\n}\n\nfn __rune_path_append_part(out : String, part : String) -> String {\n  if out.is_empty() {\n    part\n  } else if part.is_empty() {\n    out\n  } else {\n    out + \"/\" + part\n  }\n}\n\nfn __rune_path_normalize_parts(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  if index >= parts.length() {\n    out\n  } else {\n    let part = parts[index]\n    if part.is_empty() || part == \".\" {\n      __rune_path_normalize_parts(parts, index + 1, absolute, out)\n    } else if part == \"..\" {\n      __rune_path_normalize_parent(parts, index, absolute, out)\n    } else {\n      __rune_path_normalize_push(parts, index, absolute, out, part)\n    }\n  }\n}\n\nfn __rune_path_normalize_parent(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  if out.length() > 0 {\n    __rune_path_normalize_pop(parts, index, absolute, out)\n  } else if absolute {\n    __rune_path_normalize_parts(parts, index + 1, absolute, out)\n  } else {\n    __rune_path_normalize_push(parts, index, absolute, out, \"..\")\n  }\n}\n\nfn __rune_path_normalize_pop(parts : Array[String], index : Int, absolute : Bool, out : Array[String]) -> Array[String] {\n  __rune_path_normalize_parts(parts, index + 1, absolute, out[:out.length() - 1].to_owned())\n}\n\nfn __rune_path_normalize_push(parts : Array[String], index : Int, absolute : Bool, out : Array[String], part : String) -> Array[String] {\n  __rune_path_normalize_parts(parts, index + 1, absolute, [..out, part])\n}\n\nfn __rune_path_parts(path : String) -> Array[String] {\n  let pieces = __rune_path_normalize(path).split(\"/\").map(fn(part) { part.to_owned() }).to_array()\n  __rune_path_collect_parts(pieces, 0, [])\n}\n\nfn __rune_path_collect_parts(parts : Array[String], index : Int, out : Array[String]) -> Array[String] {\n  if index >= parts.length() {\n    out\n  } else if parts[index].is_empty() {\n    __rune_path_collect_parts(parts, index + 1, out)\n  } else {\n    __rune_path_collect_parts(parts, index + 1, [..out, parts[index]])\n  }\n}\n\nfn __rune_path_collect_part(parts : Array[String], index : Int, out : Array[String]) -> Array[String] {\n  if index < parts.length() {\n    __rune_path_collect_parts(parts, index + 1, [..out, parts[index]])\n  } else {\n    out\n  }\n}\n\nfn __rune_path_relative_from_parts(from_parts : Array[String], to_parts : Array[String], index : Int) -> String {\n  if index < from_parts.length() && index < to_parts.length() && from_parts[index] == to_parts[index] {\n    __rune_path_relative_from_parts(from_parts, to_parts, index + 1)\n  } else {\n    __rune_path_relative_tail(from_parts, to_parts, index, index, \"\")\n  }\n}\n\nfn __rune_path_relative_tail(from_parts : Array[String], to_parts : Array[String], from_index : Int, to_index : Int, out : String) -> String {\n  if from_index < from_parts.length() {\n    __rune_path_relative_tail(from_parts, to_parts, from_index + 1, to_index, __rune_path_append_part(out, \"..\"))\n  } else if to_index < to_parts.length() {\n    __rune_path_relative_tail(from_parts, to_parts, from_index, to_index + 1, __rune_path_append_part(out, to_parts[to_index]))\n  } else if out.is_empty() {\n    \".\"\n  } else {\n    out\n  }\n}\n\n"
}

func selfhost_compiler_mbt_emitMoonBitEnum(enumDecl IREnumType) string {
	out := "enum " + selfhost_compiler_mbt_moonBitTypeIdent(enumDecl.name) + selfhost_compiler_mbt_emitMoonBitGenerics(enumDecl.generics) + " {\n"
	out = out + selfhost_compiler_mbt_emitMoonBitEnumMembers(enumDecl.members, 0, "")
	return out + "} derive(Eq, Show)\n"
}

func selfhost_compiler_mbt_emitMoonBitEnumMembers(members []IREnumMember, index int, out string) string {
	return func() string {
		if index >= len(members) {
			return out
		}
		return selfhost_compiler_mbt_emitMoonBitEnumMembers(members, index+1, out+selfhost_compiler_mbt_emitMoonBitEnumMember(members[index]))
	}()
}

func selfhost_compiler_mbt_emitMoonBitEnumMember(member IREnumMember) string {
	return func() string {
		if len(member.params) == 0 {
			return line(1, selfhost_compiler_mbt_moonBitConstructorIdent(member.name))
		}
		return line(1, selfhost_compiler_mbt_moonBitConstructorIdent(member.name)+"("+selfhost_compiler_mbt_emitMoonBitParamTypes(member.params, 0, "")+")")
	}()
}

func selfhost_compiler_mbt_emitMoonBitStruct(typeDecl IRStructType) string {
	out := "struct " + selfhost_compiler_mbt_moonBitTypeIdent(typeDecl.name) + " {\n"
	for _, field := range typeDecl.fields {
		_ = field
		out = out + line(1, mangleIdent(field.name)+" : "+selfhost_compiler_mbt_moonBitType(field.typeName))
	}
	return out + "}\n"
}

func selfhost_compiler_mbt_emitMoonBitMethods(typeDecl IRStructType) string {
	out := ""
	for _, method := range typeDecl.methods {
		_ = method
		out = out + selfhost_compiler_mbt_emitMoonBitFunction(selfhost_compiler_mbt_methodWithMoonBitReceiver(typeDecl.name, method), "") + "\n"
	}
	return out
}

func selfhost_compiler_mbt_emitMoonBitEnumMethods(enumDecl IREnumType) string {
	out := ""
	for _, method := range enumDecl.methods {
		_ = method
		out = out + selfhost_compiler_mbt_emitMoonBitFunction(selfhost_compiler_mbt_methodWithMoonBitReceiver(enumDecl.name, method), "") + "\n"
	}
	return out
}

func selfhost_compiler_mbt_methodWithMoonBitReceiver(typeName string, method IRFunction) IRFunction {
	return IRFunction{name: typeName + "_" + method.name, private: method.private, static: method.static, routine: method.routine, macro: method.macro, receiverType: method.receiverType, generics: method.generics, params: func() []IRParam {
		switch {
		case method.static == true:
			return method.params
		case method.static == false:
			return selfhost_compiler_mbt_prependMoonBitSelfParam(typeName, method.params)
		}
		return nil
	}(), returnType: method.returnType, body: method.body, sourcePath: method.sourcePath, line: method.line, column: method.column}
}

func selfhost_compiler_mbt_prependMoonBitSelfParam(typeName string, params []IRParam) []IRParam {
	out := []IRParam{IRParam{name: "this", typeName: typeName, line: 0, column: 0}}
	for _, param := range params {
		_ = param
		func() int { out = append(out, param); return len(out) }()
	}
	return out
}

func selfhost_compiler_mbt_emitMoonBitFunction(fn IRFunction, receiverType string) string {
	params := selfhost_compiler_mbt_emitMoonBitParams(fn.params, 0, "")
	ret := func() string {
		if returnsValue(fn.returnType) {
			return " -> " + selfhost_compiler_mbt_moonBitType(fn.returnType)
		}
		return ""
	}()
	name := mangleIdent(fn.name)
	bodyReturns := returnsValue(fn.returnType) && fn.name != "main"
	head := func() string {
		if fn.name == "main" && params == "" {
			return "fn main"
		}
		return "fn " + name + selfhost_compiler_mbt_emitMoonBitGenerics(fn.generics) + "(" + params + ")" + ret
	}()
	out := head + " {\n"
	out = out + selfhost_compiler_mbt_emitMoonBitBody(fn.body, bodyReturns, fn.returnType, 1)
	return out + "}\n"
}

func selfhost_compiler_mbt_emitMoonBitBody(expr IRExpr, returns bool, returnType string, level int) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_mbt_emitMoonBitBlock(expr.children, 0, returns, returnType, level, "")
		default:
			return line(level, func() string {
				if returns {
					return selfhost_compiler_mbt_emitMoonBitExpr(expr)
				}
				return selfhost_compiler_mbt_emitMoonBitDiscard(expr)
			}())
		}
	}()
}

func selfhost_compiler_mbt_emitMoonBitBlock(statements []IRExpr, index int, returns bool, returnType string, level int, out string) string {
	return func() string {
		if index >= len(statements) {
			return func() string {
				if returns && len(statements) == 0 {
					return out + line(level, selfhost_compiler_mbt_moonBitZero(returnType))
				}
				return out
			}()
		}
		return selfhost_compiler_mbt_emitMoonBitBlock(statements, index+1, returns, returnType, level, out+selfhost_compiler_mbt_emitMoonBitStatement(statements[index], index == len(statements)-1, returns, level))
	}()
}

func selfhost_compiler_mbt_emitMoonBitStatement(expr IRExpr, last bool, returns bool, level int) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Let:
			return selfhost_compiler_mbt_emitMoonBitLet(expr, level)
		case expr.kind == ExprKind_ObjectDestructure:
			return selfhost_compiler_mbt_emitMoonBitObjectDestructure(expr, level)
		default:
			return line(level, func() string {
				if last && returns {
					return selfhost_compiler_mbt_emitMoonBitExpr(expr)
				}
				return selfhost_compiler_mbt_emitMoonBitDiscard(expr)
			}())
		}
	}()
}

func selfhost_compiler_mbt_emitMoonBitLet(expr IRExpr, level int) string {
	return line(level, selfhost_compiler_mbt_moonBitLetKeyword(expr.op)+mangleIdent(expr.name)+" = "+selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]))
}

func selfhost_compiler_mbt_moonBitLetKeyword(op string) string {
	return func() string {
		switch {
		case op == ":=:":
			return "let mut "
		default:
			return "let "
		}
	}()
}

func selfhost_compiler_mbt_emitMoonBitObjectDestructure(expr IRExpr, level int) string {
	source := selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0])
	tmp := mangleIdent("__object")
	out := line(level, "let "+tmp+" = "+source)
	for _, param := range expr.params {
		_ = param
		out = out + line(level, "let "+mangleIdent(param.name)+" = "+tmp+"."+mangleIdent(param.typeName))
	}
	return out
}

func selfhost_compiler_mbt_emitMoonBitParams(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_mbt_emitMoonBitParams(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(params[index].name)+" : "+selfhost_compiler_mbt_moonBitType(params[index].typeName))
	}()
}

func selfhost_compiler_mbt_emitMoonBitParamTypes(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_mbt_emitMoonBitParamTypes(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_mbt_moonBitType(params[index].typeName))
	}()
}

func selfhost_compiler_mbt_emitMoonBitGenerics(generics []string) string {
	return func() string {
		if len(generics) == 0 {
			return ""
		}
		return "[" + selfhost_compiler_mbt_joinMoonBitGenericNames(generics, 0, "") + "]"
	}()
}

func selfhost_compiler_mbt_emitMoonBitExpr(expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Identifier:
			return selfhost_compiler_mbt_moonBitValueIdent(expr.name)
		case expr.kind == ExprKind_At:
			return "@" + expr.name
		case expr.kind == ExprKind_This:
			return mangleIdent("this")
		case expr.kind == ExprKind_Int:
			return expr.value
		case expr.kind == ExprKind_Double:
			return expr.value
		case expr.kind == ExprKind_BigInt:
			return bigintLiteralDigits(expr.value)
		case expr.kind == ExprKind_String:
			return expr.value
		case expr.kind == ExprKind_Template:
			return selfhost_compiler_mbt_emitMoonBitTemplate(expr)
		case expr.kind == ExprKind_Char:
			return expr.value
		case expr.kind == ExprKind_Regex:
			return expr.value
		case expr.kind == ExprKind_Bool:
			return expr.value
		case expr.kind == ExprKind_Null:
			return "None"
		case expr.kind == ExprKind_Unary:
			return selfhost_compiler_mbt_moonBitUnaryOp(expr.op) + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0])
		case expr.kind == ExprKind_Postfix:
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + selfhost_compiler_mbt_moonBitPostfixOp(expr.op)
		case expr.kind == ExprKind_CompileTime:
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0])
		case expr.kind == ExprKind_Unwrap:
			return "rune_unwrap(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + ")"
		case expr.kind == ExprKind_Binary:
			return selfhost_compiler_mbt_emitMoonBitBinary(expr)
		case expr.kind == ExprKind_Ternary:
			return selfhost_compiler_mbt_emitMoonBitTernary(expr)
		case expr.kind == ExprKind_Assign:
			return selfhost_compiler_mbt_emitMoonBitAssign(expr)
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_mbt_emitMoonBitCall(expr)
		case expr.kind == ExprKind_Lambda:
			return selfhost_compiler_mbt_emitMoonBitLambda(expr)
		case expr.kind == ExprKind_Selector:
			return selfhost_compiler_mbt_emitMoonBitSelector(expr)
		case expr.kind == ExprKind_Index:
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + "[" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + "]"
		case expr.kind == ExprKind_Array:
			return "[" + selfhost_compiler_mbt_emitMoonBitExprList(expr.children, 0, "") + "]"
		case expr.kind == ExprKind_Tuple:
			return "(" + selfhost_compiler_mbt_emitMoonBitExprList(expr.children, 0, "") + ")"
		case expr.kind == ExprKind_Map:
			return "{" + selfhost_compiler_mbt_emitMoonBitMapEntries(expr.children, 0, "") + "}"
		case expr.kind == ExprKind_Spread:
			return ".." + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0])
		case expr.kind == ExprKind_Reactive:
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0])
		case expr.kind == ExprKind_Struct:
			return selfhost_compiler_mbt_moonBitTypeIdent(expr.name) + "::{ " + selfhost_compiler_mbt_emitMoonBitFields(expr.children, 0, "") + " }"
		case expr.kind == ExprKind_Object:
			return "{ " + selfhost_compiler_mbt_emitMoonBitFields(expr.children, 0, "") + " }"
		case expr.kind == ExprKind_XMLElement:
			return "()"
		case expr.kind == ExprKind_Block:
			return "{\n" + selfhost_compiler_mbt_emitMoonBitBlock(expr.children, 0, true, "Dynamic", 1, "") + "}"
		default:
			return "()"
		}
	}()
}

func selfhost_compiler_mbt_emitMoonBitTemplate(expr IRExpr) string {
	raw := expr.value
	inner := func() string {
		if len([]rune(raw)) >= 2 {
			return func() string { runes := []rune(raw); return string(runes[1 : len([]rune(raw))-1]) }()
		}
		return raw
	}()
	parts := selfhost_compiler_mbt_mbtTemplateAccumulate(func() []string { parts := strings.Split(inner, "<<<RUNE_TEMPLATE_PART>>>"); return parts }(), expr.children, 0, []string{})
	return func() string {
		if len(parts) == 0 {
			return "\"\""
		}
		return selfhost_compiler_mbt_joinMoonBitTemplateParts(parts, 0, "")
	}()
}

func selfhost_compiler_mbt_mbtTemplateAccumulate(segments []string, children []IRExpr, index int, out []string) []string {
	return func() []string {
		if index >= len(segments) {
			return out
		}
		return selfhost_compiler_mbt_mbtTemplateAccumulate(segments, children, index+1, selfhost_compiler_mbt_mbtTemplateAppendStep(segments, children, index, out))
	}()
}

func selfhost_compiler_mbt_mbtTemplateAppendStep(segments []string, children []IRExpr, index int, out []string) []string {
	withText := func() []string {
		if len(segments[index]) == 0 {
			return out
		}
		return func() []string {
			__rune_spread_out := []string{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, selfhost_compiler_mbt_mbtStringLiteral(segments[index]))
			return __rune_spread_out
		}()
	}()
	return func() []string {
		if index < len(children) {
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, withText...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_mbt_mbtTemplateInterpolation(children[index]))
				return __rune_spread_out
			}()
		}
		return withText
	}()
}

func selfhost_compiler_mbt_mbtTemplateInterpolation(child IRExpr) string {
	return func() string {
		if child.text == "String" {
			return selfhost_compiler_mbt_emitMoonBitExpr(child)
		}
		return selfhost_compiler_mbt_emitMoonBitExpr(child) + ".to_string()"
	}()
}

func selfhost_compiler_mbt_joinMoonBitTemplateParts(parts []string, index int, out string) string {
	return func() string {
		if index >= len(parts) {
			return out
		}
		return selfhost_compiler_mbt_joinMoonBitTemplateParts(parts, index+1, func() string {
			if index == 0 {
				return out + parts[index]
			}
			return out + " + " + parts[index]
		}())
	}()
}

func selfhost_compiler_mbt_mbtStringLiteral(raw string) string {
	return "\"" + selfhost_compiler_mbt_mbtStringLiteralChars(raw, 0, "") + "\""
}

func selfhost_compiler_mbt_mbtStringLiteralChars(raw string, index int, out string) string {
	return func() string {
		if index >= len([]rune(raw)) {
			return out
		}
		return selfhost_compiler_mbt_mbtStringLiteralChar(raw, index, out)
	}()
}

func selfhost_compiler_mbt_mbtStringLiteralChar(raw string, index int, out string) string {
	return func() string {
		switch {
		case []rune(raw)[index] == '\\' && index+1 < len([]rune(raw)) == true:
			return selfhost_compiler_mbt_mbtStringLiteralNext(raw, index+2, out+selfhost_compiler_mbt_mbtStringLiteralEscaped([]rune(raw)[index+1]))
		default:
			return selfhost_compiler_mbt_mbtStringLiteralNext(raw, index+1, out+selfhost_compiler_mbt_mbtStringLiteralPlain([]rune(raw)[index]))
		}
	}()
}

func selfhost_compiler_mbt_mbtStringLiteralEscaped(ch rune) string {
	return func() string {
		if ch == 'n' {
			return "\\n"
		}
		return func() string {
			if ch == 't' {
				return "\\t"
			}
			return func() string {
				if ch == 'r' {
					return "\\r"
				}
				return func() string {
					if ch == '\\' {
						return "\\\\"
					}
					return func() string {
						if ch == '"' {
							return "\\\""
						}
						return selfhost_compiler_mbt_mbtStringLiteralPlain(ch)
					}()
				}()
			}()
		}()
	}()
}

func selfhost_compiler_mbt_mbtStringLiteralPlain(ch rune) string {
	return func() string {
		if ch == '"' {
			return "\\\""
		}
		return func() string {
			if ch == '\\' {
				return "\\\\"
			}
			return func() string {
				if ch == '\n' {
					return "\\n"
				}
				return func() string {
					if ch == '\t' {
						return "\\t"
					}
					return func() string {
						if ch == '\r' {
							return "\\r"
						}
						return string(ch)
					}()
				}()
			}()
		}()
	}()
}

func selfhost_compiler_mbt_mbtStringLiteralNext(raw string, index int, out string) string {
	return selfhost_compiler_mbt_mbtStringLiteralChars(raw, index, out)
}

func selfhost_compiler_mbt_emitMoonBitDiscard(expr IRExpr) string {
	return func() string {
		if expr.kind == ExprKind_Call && selfhost_compiler_mbt_moonBitUnitCallResult(expr) {
			return selfhost_compiler_mbt_emitMoonBitExpr(expr)
		}
		return "ignore(" + selfhost_compiler_mbt_emitMoonBitExpr(expr) + ")"
	}()
}

func selfhost_compiler_mbt_moonBitUnitCallResult(expr IRExpr) bool {
	return func() bool {
		if expr.text != "Unit" && expr.text != "()" && expr.text != "" {
			return false
		}
		return !(selfhost_compiler_mbt_moonBitValueCoreMethodCall(expr))
	}()
}

func selfhost_compiler_mbt_moonBitValueCoreMethodCall(expr IRExpr) bool {
	return func() bool {
		if len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector {
			return selfhost_compiler_mbt_moonBitValueCoreMethodName(expr.children[0].name)
		}
		return false
	}()
}

func selfhost_compiler_mbt_moonBitValueCoreMethodName(name string) bool {
	return func() bool {
		switch {
		case (name == "length") || (name == "byteLength") || (name == "isEmpty") || (name == "at") || (name == "slice") || (name == "toString"):
			return true
		case (name == "push") || (name == "set") || (name == "pop") || (name == "first") || (name == "last") || (name == "clone") || (name == "reverse") || (name == "contains"):
			return true
		default:
			return false
		}
	}()
}

func selfhost_compiler_mbt_emitMoonBitBinary(expr IRExpr) string {
	return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + " " + selfhost_compiler_mbt_moonBitBinaryOp(expr.op) + " " + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1])
}

func selfhost_compiler_mbt_emitMoonBitTernary(expr IRExpr) string {
	return "if " + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + " { " + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + " } else { " + func() string {
		if len(expr.children) > 2 {
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[2])
		}
		return "()"
	}() + " }"
}

func selfhost_compiler_mbt_emitMoonBitAssign(expr IRExpr) string {
	return func() string {
		if len(expr.children) == 2 {
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + " = " + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1])
		}
		return mangleIdent(expr.name) + " = " + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0])
	}()
}

func selfhost_compiler_mbt_emitMoonBitCall(expr IRExpr) string {
	return func() string {
		switch {
		case moduleCallKey(expr) == "io.println":
			return "println(" + selfhost_compiler_mbt_emitMoonBitPrintArgs(expr.children, 1, "") + ")"
		case moduleCallKey(expr) == "io.print":
			return "print(" + selfhost_compiler_mbt_emitMoonBitPrintArgs(expr.children, 1, "") + ")"
		case moduleCallKey(expr) == "map.new":
			return "{}"
		case moduleCallKey(expr) == "set.new":
			return "Set::new()"
		case moduleCallKey(expr) == "path.isAbsolute":
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ".has_prefix(\"/\")"
		case moduleCallKey(expr) == "path.basename":
			return "__rune_path_basename(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.extname":
			return "__rune_path_extname(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.dirname":
			return "__rune_path_dirname(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.join":
			return "__rune_path_join(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.normalize":
			return "__rune_path_normalize(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.resolve":
			return "__rune_path_resolve(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.relative":
			return "__rune_path_relative(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[2]) + ")"
		case moduleCallKey(expr) == "path.joinParts":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall3("__rune_path_join_parts", expr.children[1], expr.children[2], expr.children[3])
		case moduleCallKey(expr) == "path.appendPathPart":
			return "__rune_path_append_part(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[2]) + ")"
		case moduleCallKey(expr) == "path.normalizeParts":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall4("__rune_path_normalize_parts", expr.children[1], expr.children[2], expr.children[3], expr.children[4])
		case moduleCallKey(expr) == "path.normalizePart":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall4("__rune_path_normalize_parts", expr.children[1], expr.children[2], expr.children[3], expr.children[4])
		case moduleCallKey(expr) == "path.normalizeParent":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall4("__rune_path_normalize_parent", expr.children[1], expr.children[2], expr.children[3], expr.children[4])
		case moduleCallKey(expr) == "path.normalizePop":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall4("__rune_path_normalize_pop", expr.children[1], expr.children[2], expr.children[3], expr.children[4])
		case moduleCallKey(expr) == "path.normalizePush":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall5("__rune_path_normalize_push", expr.children[1], expr.children[2], expr.children[3], expr.children[4], expr.children[5])
		case moduleCallKey(expr) == "path.pathParts":
			return "__rune_path_parts(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.collectPathParts":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall3("__rune_path_collect_parts", expr.children[1], expr.children[2], expr.children[3])
		case moduleCallKey(expr) == "path.collectPathPart":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall3("__rune_path_collect_part", expr.children[1], expr.children[2], expr.children[3])
		case moduleCallKey(expr) == "path.relativeFromParts":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall3("__rune_path_relative_from_parts", expr.children[1], expr.children[2], expr.children[3])
		case moduleCallKey(expr) == "path.relativeTail":
			return selfhost_compiler_mbt_emitMoonBitRuntimeCall5("__rune_path_relative_tail", expr.children[1], expr.children[2], expr.children[3], expr.children[4], expr.children[5])
		case moduleCallKey(expr) == "process.platform":
			return "\"moonbit\""
		case moduleCallKey(expr) == "process.cwd":
			return "\".\""
		case moduleCallKey(expr) == "process.env":
			return "(None : String?)"
		case moduleCallKey(expr) == "process.argv":
			return "([] : Array[String])"
		case moduleCallKey(expr) == "int.toString":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").to_string()"
		case moduleCallKey(expr) == "int.toDouble":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").to_double()"
		case moduleCallKey(expr) == "int.toBigInt":
			return "BigInt::from_int(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int4.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xf; if __n >= 8 { __n - 16 } else { __n } })(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int8.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xff; if __n >= 128 { __n - 256 } else { __n } })(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int16.fromInt":
			return "(fn(__value : Int) -> Int { let __n = __value & 0xffff; if __n >= 32768 { __n - 65536 } else { __n } })(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int64.fromInt":
			return "Int64::from_int(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case (moduleCallKey(expr) == "uint.fromInt") || (moduleCallKey(expr) == "uint8.fromInt") || (moduleCallKey(expr) == "uint16.fromInt"):
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").reinterpret_as_uint()"
		case moduleCallKey(expr) == "uint64.fromInt":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").to_uint64()"
		case (moduleCallKey(expr) == "int4.toInt") || (moduleCallKey(expr) == "int8.toInt") || (moduleCallKey(expr) == "int16.toInt"):
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1])
		case (moduleCallKey(expr) == "uint.toInt") || (moduleCallKey(expr) == "uint8.toInt") || (moduleCallKey(expr) == "uint16.toInt"):
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").reinterpret_as_int()"
		case (moduleCallKey(expr) == "int64.toInt") || (moduleCallKey(expr) == "uint64.toInt"):
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").to_int()"
		case moduleCallKey(expr) == "float.fromDouble":
			return "Float::from_double(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "float.toDouble":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").to_double()"
		case moduleCallKey(expr) == "bigint.fromInt":
			return "BigInt::from_int(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "bigint.toString":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").to_string()"
		case moduleCallKey(expr) == "bigint.toDouble":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").to_double()"
		case moduleCallKey(expr) == "double.trunc":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").trunc()"
		case moduleCallKey(expr) == "double.floor":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").floor()"
		case moduleCallKey(expr) == "double.ceil":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").ceil()"
		case moduleCallKey(expr) == "double.round":
			return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ").round()"
		default:
			return selfhost_compiler_mbt_emitMoonBitMaybeCoreMethodCall(expr)
		}
	}()
}

func selfhost_compiler_mbt_emitMoonBitRuntimeCall3(name string, first IRExpr, second IRExpr, third IRExpr) string {
	return name + "(" + selfhost_compiler_mbt_emitMoonBitExpr(first) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(second) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(third) + ")"
}

func selfhost_compiler_mbt_emitMoonBitRuntimeCall4(name string, first IRExpr, second IRExpr, third IRExpr, fourth IRExpr) string {
	return name + "(" + selfhost_compiler_mbt_emitMoonBitExpr(first) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(second) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(third) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(fourth) + ")"
}

func selfhost_compiler_mbt_emitMoonBitRuntimeCall5(name string, first IRExpr, second IRExpr, third IRExpr, fourth IRExpr, fifth IRExpr) string {
	return name + "(" + selfhost_compiler_mbt_emitMoonBitExpr(first) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(second) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(third) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(fourth) + ", " + selfhost_compiler_mbt_emitMoonBitExpr(fifth) + ")"
}

func selfhost_compiler_mbt_emitMoonBitMaybeCoreMethodCall(expr IRExpr) string {
	return func() string {
		if len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector {
			return selfhost_compiler_mbt_emitMoonBitCoreMethodCall(expr, expr.children[0])
		}
		return selfhost_compiler_mbt_emitMoonBitDefaultCall(expr)
	}()
}

func selfhost_compiler_mbt_emitMoonBitCoreMethodCall(expr IRExpr, selector IRExpr) string {
	return func() string {
		if len(selector.children) > 0 && selector.children[0].kind != ExprKind_At {
			return func() string {
				switch {
				case (selector.name == "length") || (selector.name == "byteLength"):
					return selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + ".length()"
				case selector.name == "isEmpty":
					return "(" + selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + ".length() == 0)"
				case selector.name == "at":
					return selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + "[" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + "]"
				case selector.name == "slice":
					return selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + "[" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ":" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[2]) + "].to_owned()"
				case selector.name == "toString":
					return selfhost_compiler_mbt_emitMoonBitCoreToString(expr, selector.children[0])
				case selector.name == "push":
					return selfhost_compiler_mbt_emitMoonBitArrayPush(selector.children[0], expr.children[1])
				case selector.name == "set":
					return selfhost_compiler_mbt_emitMoonBitArraySet(selector.children[0], expr.children[1], expr.children[2])
				case selector.name == "pop":
					return selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + ".pop().unwrap()"
				case selector.name == "first":
					return selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + "[0]"
				case selector.name == "last":
					return selfhost_compiler_mbt_emitMoonBitArrayLast(selector.children[0])
				case selector.name == "clone":
					return selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + ".copy()"
				case selector.name == "reverse":
					return selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + ".rev()"
				case selector.name == "contains":
					return selfhost_compiler_mbt_emitMoonBitExpr(selector.children[0]) + ".contains(" + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[1]) + ")"
				default:
					return selfhost_compiler_mbt_emitMoonBitDefaultCall(expr)
				}
			}()
		}
		return selfhost_compiler_mbt_emitMoonBitDefaultCall(expr)
	}()
}

func selfhost_compiler_mbt_emitMoonBitArrayPush(receiver IRExpr, value IRExpr) string {
	out := "{ let __array = " + selfhost_compiler_mbt_emitMoonBitExpr(receiver)
	out = out + "; __array.push(" + selfhost_compiler_mbt_emitMoonBitExpr(value)
	return out + "); __array.length() }"
}

func selfhost_compiler_mbt_emitMoonBitArraySet(receiver IRExpr, index IRExpr, value IRExpr) string {
	out := "{ let __array = " + selfhost_compiler_mbt_emitMoonBitExpr(receiver)
	out = out + "; let __index = " + selfhost_compiler_mbt_emitMoonBitExpr(index)
	out = out + "; let __value = " + selfhost_compiler_mbt_emitMoonBitExpr(value)
	return out + "; __array[__index] = __value; __value }"
}

func selfhost_compiler_mbt_emitMoonBitArrayLast(receiver IRExpr) string {
	return selfhost_compiler_mbt_emitMoonBitExpr(receiver) + "[" + selfhost_compiler_mbt_emitMoonBitExpr(receiver) + ".length() - 1]"
}

func selfhost_compiler_mbt_emitMoonBitCoreToString(expr IRExpr, receiver IRExpr) string {
	return func() string {
		if receiver.text == "String" || receiver.kind == ExprKind_String {
			return selfhost_compiler_mbt_emitMoonBitExpr(receiver)
		}
		return "(" + selfhost_compiler_mbt_emitMoonBitExpr(receiver) + ").to_string()"
	}()
}

func selfhost_compiler_mbt_emitMoonBitDefaultCall(expr IRExpr) string {
	return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + "(" + selfhost_compiler_mbt_emitMoonBitExprListFrom(expr.children, 1, "") + ")"
}

func selfhost_compiler_mbt_emitMoonBitPrintArgs(exprs []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(exprs) {
			return func() string {
				if out == "" {
					return "\"\""
				}
				return out
			}()
		}
		return selfhost_compiler_mbt_emitMoonBitPrintArgs(exprs, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return " + \" \" + "
		}()+selfhost_compiler_mbt_emitMoonBitShowExpr(exprs[index]))
	}()
}

func selfhost_compiler_mbt_emitMoonBitShowExpr(expr IRExpr) string {
	return func() string {
		if expr.text == "String" || expr.kind == ExprKind_String || expr.kind == ExprKind_Template {
			return selfhost_compiler_mbt_emitMoonBitExpr(expr)
		}
		return "(" + selfhost_compiler_mbt_emitMoonBitExpr(expr) + ").to_string()"
	}()
}

func selfhost_compiler_mbt_emitMoonBitLambda(expr IRExpr) string {
	return "(" + selfhost_compiler_mbt_emitMoonBitParams(expr.params, 0, "") + ") => " + selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0])
}

func selfhost_compiler_mbt_emitMoonBitSelector(expr IRExpr) string {
	return func() string {
		switch {
		case expr.children[0].kind == ExprKind_At:
			return selfhost_compiler_mbt_emitMoonBitAtSelector(expr)
		case expr.children[0].kind == ExprKind_Identifier:
			return func() string {
				switch {
				case expr.op == "::":
					return mangleIdent(expr.children[0].name + "_" + expr.name)
				default:
					return func() string {
						if selfhost_compiler_mbt_moonBitLooksLikeTypeName(expr.children[0].name) {
							return selfhost_compiler_mbt_moonBitTypeIdent(expr.children[0].name) + "::" + selfhost_compiler_mbt_moonBitConstructorIdent(expr.name)
						}
						return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + "." + mangleIdent(expr.name)
					}()
				}
			}()
		default:
			return selfhost_compiler_mbt_emitMoonBitExpr(expr.children[0]) + "." + mangleIdent(expr.name)
		}
	}()
}

func selfhost_compiler_mbt_emitMoonBitAtSelector(expr IRExpr) string {
	imported := expr.children[0].value != ""
	return func() string {
		switch {
		case imported == true:
			return mangleIdent(expr.name)
		default:
			return "@" + expr.children[0].name + "." + expr.name
		}
	}()
}

func selfhost_compiler_mbt_emitMoonBitExprList(exprs []IRExpr, index int, out string) string {
	return selfhost_compiler_mbt_emitMoonBitExprListFrom(exprs, index, out)
}

func selfhost_compiler_mbt_emitMoonBitExprListFrom(exprs []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(exprs) {
			return out
		}
		return selfhost_compiler_mbt_emitMoonBitExprListFrom(exprs, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+selfhost_compiler_mbt_emitMoonBitExpr(exprs[index]))
	}()
}

func selfhost_compiler_mbt_emitMoonBitMapEntries(entries []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(entries) {
			return out
		}
		return selfhost_compiler_mbt_emitMoonBitMapEntries(entries, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+selfhost_compiler_mbt_emitMoonBitExpr(entries[index].children[0])+": "+selfhost_compiler_mbt_emitMoonBitExpr(entries[index].children[1]))
	}()
}

func selfhost_compiler_mbt_emitMoonBitFields(fields []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(fields) {
			return out
		}
		return selfhost_compiler_mbt_emitMoonBitFields(fields, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+mangleIdent(fields[index].name)+": "+selfhost_compiler_mbt_emitMoonBitExpr(fields[index].children[0]))
	}()
}

func selfhost_compiler_mbt_moonBitUnaryOp(op string) string {
	return func() string {
		switch {
		case op == "!":
			return "!"
		default:
			return op
		}
	}()
}

func selfhost_compiler_mbt_moonBitPostfixOp(op string) string {
	return op
}

func selfhost_compiler_mbt_moonBitBinaryOp(op string) string {
	return func() string {
		switch {
		case op == "&&":
			return "&&"
		case op == "||":
			return "||"
		default:
			return op
		}
	}()
}

func selfhost_compiler_mbt_moonBitType(typeName string) string {
	switch {
	case (typeName == "") || (typeName == "Void"):
		return "Unit"
	case (typeName == "Int") || (typeName == "Int4") || (typeName == "Int8") || (typeName == "Int16"):
		return "Int"
	case (typeName == "UInt") || (typeName == "UInt8") || (typeName == "UInt16"):
		return "UInt"
	case typeName == "Int64":
		return "Int64"
	case typeName == "UInt64":
		return "UInt64"
	case typeName == "Double":
		return "Double"
	case typeName == "Float":
		return "Float"
	case typeName == "BigInt":
		return "BigInt"
	case typeName == "String":
		return "String"
	case typeName == "Char":
		return "Char"
	case typeName == "Bool":
		return "Bool"
	case (typeName == "Dynamic") || (typeName == "Object"):
		return "Unit"
	case (typeName == "Data") || (typeName == "@io.Data"):
		return "Array[Int]"
	default:
		return selfhost_compiler_mbt_moonBitTypeFallback(typeName)
	}
}

func selfhost_compiler_mbt_moonBitTypeFallback(typeName string) string {
	return func() string {
		if strings.HasSuffix(typeName, "?") {
			return selfhost_compiler_mbt_moonBitType(func() string { runes := []rune(typeName); return string(runes[0 : len([]rune(typeName))-1]) }()) + "?"
		}
		return func() string {
			if genericInner(typeName, "Array") != "" {
				return "Array[" + selfhost_compiler_mbt_moonBitType(genericInner(typeName, "Array")) + "]"
			}
			return func() string {
				if genericInner(typeName, "ReadonlyArray") != "" {
					return "Array[" + selfhost_compiler_mbt_moonBitType(genericInner(typeName, "ReadonlyArray")) + "]"
				}
				return func() string {
					if genericInner(typeName, "Map") != "" {
						return "Map[" + selfhost_compiler_mbt_moonBitType(typeArg(genericInner(typeName, "Map"), 0)) + ", " + selfhost_compiler_mbt_moonBitType(typeArg(genericInner(typeName, "Map"), 1)) + "]"
					}
					return func() string {
						if genericInner(typeName, "Set") != "" {
							return "Set[" + selfhost_compiler_mbt_moonBitType(genericInner(typeName, "Set")) + "]"
						}
						return selfhost_compiler_mbt_moonBitNamedType(typeName)
					}()
				}()
			}()
		}()
	}()
}

func selfhost_compiler_mbt_moonBitNamedType(typeName string) string {
	open := strings.Index(typeName, "[")
	return func() string {
		if open < 0 {
			return selfhost_compiler_mbt_moonBitTypeIdent(typeName)
		}
		return selfhost_compiler_mbt_moonBitTypeIdent(func() string { runes := []rune(typeName); return string(runes[0:open]) }()) + "[" + selfhost_compiler_mbt_emitMoonBitTypeArgs(func() string { runes := []rune(typeName); return string(runes[open+1 : len([]rune(typeName))-1]) }()) + "]"
	}()
}

func selfhost_compiler_mbt_emitMoonBitTypeArgs(args string) string {
	return selfhost_compiler_mbt_emitMoonBitTypeArgList(func() []string { parts := strings.Split(args, ","); return parts }(), 0, "")
}

func selfhost_compiler_mbt_emitMoonBitTypeArgList(args []string, index int, out string) string {
	return func() string {
		if index >= len(args) {
			return out
		}
		return selfhost_compiler_mbt_emitMoonBitTypeArgList(args, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_mbt_moonBitType(strings.TrimSpace(args[index])))
	}()
}

func selfhost_compiler_mbt_moonBitZero(typeName string) string {
	return func() string {
		switch {
		case (typeName == "Int") || (typeName == "Int4") || (typeName == "Int8") || (typeName == "Int16") || (typeName == "UInt") || (typeName == "UInt8") || (typeName == "UInt16"):
			return "0"
		case typeName == "Int64":
			return "0L"
		case typeName == "UInt64":
			return "0UL"
		case (typeName == "Double") || (typeName == "Float"):
			return "0.0"
		case typeName == "String":
			return "\"\""
		case typeName == "Bool":
			return "false"
		default:
			return "()"
		}
	}()
}

func selfhost_compiler_mbt_joinMoonBitStrings(values []string, index int, out string) string {
	return func() string {
		if index >= len(values) {
			return out
		}
		return selfhost_compiler_mbt_joinMoonBitStrings(values, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+values[index])
	}()
}

func selfhost_compiler_mbt_joinMoonBitGenericNames(values []string, index int, out string) string {
	return func() string {
		if index >= len(values) {
			return out
		}
		return selfhost_compiler_mbt_joinMoonBitGenericNames(values, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_mbt_moonBitTypeParamIdent(values[index]))
	}()
}

func selfhost_compiler_mbt_moonBitValueIdent(name string) string {
	return func() string {
		if selfhost_compiler_mbt_moonBitLooksLikeTypeName(name) {
			return selfhost_compiler_mbt_moonBitConstructorIdent(name)
		}
		return mangleIdent(name)
	}()
}

func selfhost_compiler_mbt_moonBitTypeIdent(name string) string {
	return func() string {
		if selfhost_compiler_mbt_moonBitLooksGenericParam(name) {
			return selfhost_compiler_mbt_moonBitTypeParamIdent(name)
		}
		return "Rune" + selfhost_compiler_mbt_moonBitSanitizeIdent(name)
	}()
}

func selfhost_compiler_mbt_moonBitConstructorIdent(name string) string {
	return "Rune" + selfhost_compiler_mbt_moonBitSanitizeIdent(name)
}

func selfhost_compiler_mbt_moonBitTypeParamIdent(name string) string {
	return func() string {
		if selfhost_compiler_mbt_moonBitLooksLikeTypeName(name) {
			return name
		}
		return "T"
	}()
}

func selfhost_compiler_mbt_moonBitLooksGenericParam(name string) bool {
	return len([]rune(name)) == 1 && selfhost_compiler_mbt_moonBitLooksLikeTypeName(name)
}

func selfhost_compiler_mbt_moonBitSanitizeIdent(name string) string {
	return strings.ReplaceAll((strings.ReplaceAll((strings.ReplaceAll(name, ".", "_")), "-", "_")), "@", "_")
}

func selfhost_compiler_mbt_moonBitLooksLikeTypeName(name string) bool {
	return func() bool {
		if len([]rune(name)) > 0 {
			return selfhost_compiler_mbt_moonBitIsUpperLetter([]rune(name)[0])
		}
		return false
	}()
}

func selfhost_compiler_mbt_moonBitIsUpperLetter(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}

func generateTypeScript(file IRFile) string {
	resolved := selfhost_compiler_ts_rewriteTSMethodCalls(file)
	out := selfhost_compiler_ts_emitTSImports(resolved)
	if fileUsesUnwrap(resolved) {
		out = out + selfhost_compiler_ts_emitTSUnwrapHelper()
	}
	if fileUsesPathFamily(resolved) {
		out = out + selfhost_compiler_ts_emitTSPathHelpers()
	}
	for _, enumDecl := range resolved.enums {
		_ = enumDecl
		out = out + selfhost_compiler_ts_emitTSEnum(enumDecl) + "\n"
	}
	for _, enumDecl := range resolved.enums {
		_ = enumDecl
		out = out + selfhost_compiler_ts_emitTSEnumMethods(enumDecl)
	}
	for _, typeDecl := range resolved.structs {
		_ = typeDecl
		out = out + selfhost_compiler_ts_emitTSStruct(typeDecl) + "\n"
	}
	for _, typeDecl := range resolved.structs {
		_ = typeDecl
		out = out + selfhost_compiler_ts_emitTSMethods(typeDecl)
	}
	for _, constant := range resolved.constants {
		_ = constant
		out = out + selfhost_compiler_ts_emitTSConst(constant) + "\n"
	}
	for _, fn := range resolved.functions {
		_ = fn
		out = func() string {
			if fn.macro {
				return out
			}
			return out + selfhost_compiler_ts_emitTSFunction(fn) + "\n"
		}()
	}
	return out + selfhost_compiler_ts_emitTSExports(resolved)
}

func selfhost_compiler_ts_emitTSImports(file IRFile) string {
	return selfhost_compiler_ts_emitTSImportList(file.tsImports, 0, "")
}

func selfhost_compiler_ts_rewriteTSMethodCalls(file IRFile) IRFile {
	structs := selfhost_compiler_ts_rewriteTSStructs(file.structs, file.structs, file.enums)
	enums := selfhost_compiler_ts_rewriteTSEnums(file.enums, file.structs, file.enums)
	return IRFile{imports: file.imports, tsImports: file.tsImports, structs: structs, enums: enums, constants: selfhost_compiler_ts_rewriteTSConstants(file.constants, file.structs, file.enums), functions: selfhost_compiler_ts_rewriteTSFunctions(file.functions, file.structs, file.enums), tests: selfhost_compiler_ts_rewriteTSTests(file.tests, file.structs, file.enums), errors: file.errors}
}

func selfhost_compiler_ts_rewriteTSStructs(structs []IRStructType, allStructs []IRStructType, enums []IREnumType) []IRStructType {
	out := []IRStructType{}
	for _, typeDecl := range structs {
		_ = typeDecl
		func() int {
			out = append(out, IRStructType{name: typeDecl.name, private: typeDecl.private, generics: typeDecl.generics, fields: typeDecl.fields, methods: selfhost_compiler_ts_rewriteTSTypeMethods(typeDecl.methods, typeDecl.name, allStructs, enums), sourcePath: typeDecl.sourcePath, line: typeDecl.line, column: typeDecl.column})
			return len(out)
		}()
	}
	return out
}

func selfhost_compiler_ts_rewriteTSEnums(enums []IREnumType, structs []IRStructType, allEnums []IREnumType) []IREnumType {
	out := []IREnumType{}
	for _, typeDecl := range enums {
		_ = typeDecl
		func() int {
			out = append(out, IREnumType{name: typeDecl.name, private: typeDecl.private, generics: typeDecl.generics, members: typeDecl.members, methods: selfhost_compiler_ts_rewriteTSTypeMethods(typeDecl.methods, typeDecl.name, structs, allEnums), sourcePath: typeDecl.sourcePath, line: typeDecl.line, column: typeDecl.column})
			return len(out)
		}()
	}
	return out
}

func selfhost_compiler_ts_rewriteTSConstants(constants []IRConst, structs []IRStructType, enums []IREnumType) []IRConst {
	out := []IRConst{}
	for _, constant := range constants {
		_ = constant
		func() int {
			value := selfhost_compiler_ts_rewriteTSExpr(constant.value, structs, enums, []CompilerTypeBinding{})
			return func() int {
				out = append(out, IRConst{name: constant.name, private: constant.private, typeName: constant.typeName, value: value, line: constant.line, column: constant.column})
				return len(out)
			}()
		}()
	}
	return out
}

func selfhost_compiler_ts_rewriteTSTests(tests []IRTest, structs []IRStructType, enums []IREnumType) []IRTest {
	out := []IRTest{}
	for _, test := range tests {
		_ = test
		func() int {
			bindings := []CompilerTypeBinding{}
			body := selfhost_compiler_ts_rewriteTSExpr(test.body, structs, enums, bindings)
			return func() int {
				out = append(out, IRTest{name: test.name, body: body, line: test.line, column: test.column})
				return len(out)
			}()
		}()
	}
	return out
}

func selfhost_compiler_ts_rewriteTSFunctions(functions []IRFunction, structs []IRStructType, enums []IREnumType) []IRFunction {
	out := []IRFunction{}
	for _, fn := range functions {
		_ = fn
		func() int {
			out = append(out, selfhost_compiler_ts_rewriteTSFunction(fn, structs, enums))
			return len(out)
		}()
	}
	return out
}

func selfhost_compiler_ts_rewriteTSFunction(fn IRFunction, structs []IRStructType, enums []IREnumType) IRFunction {
	return selfhost_compiler_ts_rewriteTSFunctionWithThis(fn, structs, enums, "")
}

func selfhost_compiler_ts_rewriteTSTypeMethods(methods []IRFunction, typeName string, structs []IRStructType, enums []IREnumType) []IRFunction {
	out := []IRFunction{}
	for _, method := range methods {
		_ = method
		func() int {
			out = append(out, selfhost_compiler_ts_rewriteTSFunctionWithThis(method, structs, enums, typeName))
			return len(out)
		}()
	}
	return out
}

func selfhost_compiler_ts_rewriteTSFunctionWithThis(fn IRFunction, structs []IRStructType, enums []IREnumType, thisType string) IRFunction {
	return selfhost_compiler_ts_rewriteTSFunctionBody(fn, structs, enums, thisType)
}

func selfhost_compiler_ts_rewriteTSFunctionBody(fn IRFunction, structs []IRStructType, enums []IREnumType, thisType string) IRFunction {
	bindings := []CompilerTypeBinding{}
	for _, param := range fn.params {
		_ = param
		bindings = selfhost_compiler_ts_tsAddBinding(bindings, param.name, param.typeName)
	}
	bindings = func() []CompilerTypeBinding {
		if thisType == "" {
			return bindings
		}
		return selfhost_compiler_ts_tsAddBinding(bindings, "this", thisType)
	}()
	body := selfhost_compiler_ts_rewriteTSExpr(fn.body, structs, enums, bindings)
	return IRFunction{name: fn.name, private: fn.private, static: fn.static, routine: fn.routine, macro: fn.macro, receiverType: fn.receiverType, generics: fn.generics, params: fn.params, returnType: fn.returnType, body: body, sourcePath: fn.sourcePath, line: fn.line, column: fn.column}
}

func selfhost_compiler_ts_emptyCompilerTypeBinding() CompilerTypeBinding {
	return CompilerTypeBinding{name: "", typeName: ""}
}

func selfhost_compiler_ts_tsAddBinding(bindings []CompilerTypeBinding, name string, typeName string) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		__rune_spread_out := []CompilerTypeBinding{}
		__rune_spread_out = append(__rune_spread_out, bindings...)
		__rune_spread_out = append(__rune_spread_out, selfhost_compiler_ts_tsBinding(name, typeName))
		return __rune_spread_out
	}()
}

func selfhost_compiler_ts_tsBinding(name string, typeName string) CompilerTypeBinding {
	return CompilerTypeBinding{name: name, typeName: typeName}
}

func selfhost_compiler_ts_dropTSBinding(bindings []CompilerTypeBinding, name string, index int, out []CompilerTypeBinding) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		if index >= len(bindings) {
			return out
		}
		return func() []CompilerTypeBinding {
			if bindings[index].name == name {
				return selfhost_compiler_ts_dropTSBinding(bindings, name, index+1, out)
			}
			return selfhost_compiler_ts_dropTSBinding(bindings, name, index+1, func() []CompilerTypeBinding {
				__rune_spread_out := []CompilerTypeBinding{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, bindings[index])
				return __rune_spread_out
			}())
		}()
	}()
}

func selfhost_compiler_ts_tsLookupBinding(bindings []CompilerTypeBinding, name string, index int) CompilerTypeBinding {
	return func() CompilerTypeBinding {
		if index >= len(bindings) {
			return selfhost_compiler_ts_emptyCompilerTypeBinding()
		}
		return func() CompilerTypeBinding {
			if bindings[index].name == name {
				return bindings[index]
			}
			return selfhost_compiler_ts_tsLookupBinding(bindings, name, index+1)
		}()
	}()
}

func selfhost_compiler_ts_tsTypeBase(typeName string) string {
	open := strings.Index(typeName, "[")
	return func() string {
		if open < 0 {
			return typeName
		}
		return func() string { runes := []rune(typeName); return string(runes[0:open]) }()
	}()
}

func selfhost_compiler_ts_rewriteTSExpr(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	return func() IRExpr {
		switch {
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_ts_rewriteTSCall(expr, structs, enums, bindings)
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_ts_rewriteTSBlockExpr(expr, structs, enums, bindings)
		case expr.kind == ExprKind_Let:
			return selfhost_compiler_ts_rewriteTSLet(expr, structs, enums, bindings)
		default:
			return selfhost_compiler_ts_tsRebuildExpr(expr, selfhost_compiler_ts_rewriteTSChildren(expr.children, structs, enums, bindings))
		}
	}()
}

func selfhost_compiler_ts_rewriteTSBlockExpr(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	result := selfhost_compiler_ts_rewriteTSBlockStatements(expr.children, 0, structs, enums, bindings, []IRExpr{})
	return IRExpr{kind: expr.kind, text: expr.text, name: expr.name, value: expr.value, op: expr.op, params: expr.params, children: result, line: expr.line, column: expr.column}
}

func selfhost_compiler_ts_rewriteTSBlockStatements(statements []IRExpr, index int, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding, out []IRExpr) []IRExpr {
	return func() []IRExpr {
		if index >= len(statements) {
			return out
		}
		return selfhost_compiler_ts_rewriteTSBlockStep(statements, index, structs, enums, bindings, out)
	}()
}

func selfhost_compiler_ts_rewriteTSBlockStep(statements []IRExpr, index int, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding, out []IRExpr) []IRExpr {
	statement := selfhost_compiler_ts_rewriteTSStatement(statements[index], structs, enums, bindings)
	nextBindings := selfhost_compiler_ts_tsStatementBindings(statements[index], structs, enums, bindings)
	return selfhost_compiler_ts_rewriteTSBlockStatements(statements, index+1, structs, enums, nextBindings, func() []IRExpr {
		__rune_spread_out := []IRExpr{}
		__rune_spread_out = append(__rune_spread_out, out...)
		__rune_spread_out = append(__rune_spread_out, statement)
		return __rune_spread_out
	}())
}

func selfhost_compiler_ts_rewriteTSStatement(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	return func() IRExpr {
		if expr.kind == ExprKind_Let {
			return selfhost_compiler_ts_rewriteTSLet(expr, structs, enums, bindings)
		}
		return selfhost_compiler_ts_rewriteTSExpr(expr, structs, enums, bindings)
	}()
}

func selfhost_compiler_ts_rewriteTSLet(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	return IRExpr{kind: expr.kind, text: expr.text, name: expr.name, value: expr.value, op: expr.op, params: expr.params, children: selfhost_compiler_ts_tsRewriteChildren(expr.children, structs, enums, bindings), line: expr.line, column: expr.column}
}

func selfhost_compiler_ts_tsLetBindingType(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) string {
	return func() string {
		if expr.value != "" {
			return selfhost_compiler_ts_tsTypeBase(expr.value)
		}
		return func() string {
			if len(expr.children) > 0 {
				return func() string {
					if expr.children[0].kind == ExprKind_Struct {
						return selfhost_compiler_ts_tsTypeBase(expr.children[0].name)
					}
					return selfhost_compiler_ts_tsResolveReceiverType(expr.children[0], structs, enums, bindings)
				}()
			}
			return ""
		}()
	}()
}

func selfhost_compiler_ts_tsStatementBindings(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	kind := expr.kind
	return func() []CompilerTypeBinding {
		if kind == ExprKind_Let {
			return selfhost_compiler_ts_tsAddBinding(bindings, expr.name, selfhost_compiler_ts_tsLetBindingType(expr, structs, enums, bindings))
		}
		return bindings
	}()
}

func selfhost_compiler_ts_rewriteTSCall(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	hasSelectorCallee := len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector
	return func() IRExpr {
		switch {
		case hasSelectorCallee == true:
			return selfhost_compiler_ts_rewriteTSSelectorCall(expr, structs, enums, bindings)
		default:
			return selfhost_compiler_ts_tsRebuildExpr(expr, selfhost_compiler_ts_tsRewriteChildren(expr.children, structs, enums, bindings))
		}
	}()
}

func selfhost_compiler_ts_rewriteTSSelectorCall(expr IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) IRExpr {
	callee := expr.children[0]
	hasReceiver := len(callee.children) > 0
	receiver := func() IRExpr {
		if hasReceiver {
			return selfhost_compiler_ts_rewriteTSExpr(callee.children[0], structs, enums, bindings)
		}
		return emptyIRExpr()
	}()
	receiverType := selfhost_compiler_ts_tsResolveReceiverType(receiver, structs, enums, bindings)
	rewrittenArgs := selfhost_compiler_ts_tsRewriteArgs(expr.children, structs, enums, bindings)
	shouldRewrite := selfhost_compiler_ts_tsFindMethodCall(structs, enums, receiverType, callee.name)
	return func() IRExpr {
		switch {
		case shouldRewrite == true:
			return selfhost_compiler_ts_tsBuildMethodCall(expr, receiverType, callee.name, receiver, rewrittenArgs)
		default:
			return selfhost_compiler_ts_tsRebuildExpr(expr, selfhost_compiler_ts_tsPrependExpr(selfhost_compiler_ts_tsRebuildExpr(callee, func() []IRExpr {
				if hasReceiver {
					return []IRExpr{receiver}
				}
				return []IRExpr{}
			}()), rewrittenArgs))
		}
	}()
}

func selfhost_compiler_ts_tsBuildMethodCall(expr IRExpr, receiverType string, methodName string, receiver IRExpr, args []IRExpr) IRExpr {
	calleeIdent := selfhost_compiler_ts_tsMethodCallee(selfhost_compiler_ts_tsTypeBase(receiverType) + "_" + methodName)
	return IRExpr{kind: expr.kind, text: expr.text, name: expr.name, value: expr.value, op: expr.op, params: expr.params, children: selfhost_compiler_ts_tsCallChildren(calleeIdent, receiver, args), line: expr.line, column: expr.column}
}

func selfhost_compiler_ts_tsMethodCallee(name string) IRExpr {
	return IRExpr{kind: ExprKind_Identifier, text: name, name: name, value: "", op: "", params: []IRParam{}, children: []IRExpr{}, line: 0, column: 0}
}

func selfhost_compiler_ts_tsCallChildren(callee IRExpr, receiver IRExpr, args []IRExpr) []IRExpr {
	return selfhost_compiler_ts_tsPrependExprAll(callee, receiver, args)
}

func selfhost_compiler_ts_tsPrependExprAll(callee IRExpr, receiver IRExpr, args []IRExpr) []IRExpr {
	out := []IRExpr{callee, receiver}
	for _, arg := range args {
		_ = arg
		func() int { out = append(out, arg); return len(out) }()
	}
	return out
}

func selfhost_compiler_ts_tsPrependExpr(head IRExpr, rest []IRExpr) []IRExpr {
	out := []IRExpr{head}
	for _, item := range rest {
		_ = item
		func() int { out = append(out, item); return len(out) }()
	}
	return out
}

func selfhost_compiler_ts_tsRebuildExpr(expr IRExpr, children []IRExpr) IRExpr {
	return IRExpr{kind: expr.kind, text: expr.text, name: expr.name, value: expr.value, op: expr.op, params: expr.params, children: children, line: expr.line, column: expr.column}
}

func selfhost_compiler_ts_tsRewriteChildren(children []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) []IRExpr {
	out := []IRExpr{}
	for _, child := range children {
		_ = child
		func() int {
			out = append(out, selfhost_compiler_ts_rewriteTSExpr(child, structs, enums, bindings))
			return len(out)
		}()
	}
	return out
}

func selfhost_compiler_ts_rewriteTSChildren(children []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) []IRExpr {
	out := []IRExpr{}
	for _, child := range children {
		_ = child
		func() int {
			out = append(out, selfhost_compiler_ts_rewriteTSExpr(child, structs, enums, bindings))
			return len(out)
		}()
	}
	return out
}

func selfhost_compiler_ts_tsRewriteArgs(children []IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) []IRExpr {
	return selfhost_compiler_ts_tsRewriteArgsFrom(children, 1, structs, enums, bindings, []IRExpr{})
}

func selfhost_compiler_ts_tsRewriteArgsFrom(children []IRExpr, index int, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding, out []IRExpr) []IRExpr {
	return func() []IRExpr {
		if index >= len(children) {
			return out
		}
		return selfhost_compiler_ts_tsRewriteArgsFrom(children, index+1, structs, enums, bindings, func() []IRExpr {
			__rune_spread_out := []IRExpr{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, selfhost_compiler_ts_rewriteTSExpr(children[index], structs, enums, bindings))
			return __rune_spread_out
		}())
	}()
}

func selfhost_compiler_ts_tsResolveReceiverType(receiver IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) string {
	return func() string {
		if receiver.text != "" {
			return receiver.text
		}
		return func() string {
			switch {
			case receiver.kind == ExprKind_Identifier:
				return selfhost_compiler_ts_tsLookupBinding(bindings, receiver.name, 0).typeName
			case receiver.kind == ExprKind_This:
				return selfhost_compiler_ts_tsLookupBinding(bindings, "this", 0).typeName
			case receiver.kind == ExprKind_Struct:
				return receiver.name
			case receiver.kind == ExprKind_Selector:
				return selfhost_compiler_ts_tsResolveSelectorFieldType(receiver, structs, enums, bindings)
			default:
				return ""
			}
		}()
	}()
}

func selfhost_compiler_ts_tsResolveSelectorFieldType(receiver IRExpr, structs []IRStructType, enums []IREnumType, bindings []CompilerTypeBinding) string {
	empty := len(receiver.children) == 0
	return func() string {
		switch {
		case empty == true:
			return ""
		default:
			return func() string {
				if receiver.children[0].text != "" {
					return receiver.children[0].text
				}
				return selfhost_compiler_ts_tsStructFieldType(structs, selfhost_compiler_ts_tsTypeBase(selfhost_compiler_ts_tsResolveReceiverType(receiver.children[0], structs, enums, bindings)), receiver.name, 0)
			}()
		}
	}()
}

func selfhost_compiler_ts_tsStructFieldType(structs []IRStructType, typeName string, fieldName string, index int) string {
	return func() string {
		if index >= len(structs) {
			return ""
		}
		return func() string {
			if structs[index].name == typeName {
				return selfhost_compiler_ts_tsStructFieldTypeIn(structs[index].fields, fieldName, 0)
			}
			return selfhost_compiler_ts_tsStructFieldType(structs, typeName, fieldName, index+1)
		}()
	}()
}

func selfhost_compiler_ts_tsStructFieldTypeIn(fields []IRField, fieldName string, index int) string {
	return func() string {
		if index >= len(fields) {
			return ""
		}
		return func() string {
			if fields[index].name == fieldName {
				return fields[index].typeName
			}
			return selfhost_compiler_ts_tsStructFieldTypeIn(fields, fieldName, index+1)
		}()
	}()
}

func selfhost_compiler_ts_tsFindMethodCall(structs []IRStructType, enums []IREnumType, receiverType string, methodName string) bool {
	base := selfhost_compiler_ts_tsTypeBase(receiverType)
	found := selfhost_compiler_ts_tsStructHasInstanceMethod(structs, base, methodName, 0)
	return func() bool {
		switch {
		case found == true:
			return true
		default:
			return selfhost_compiler_ts_tsEnumHasInstanceMethod(enums, base, methodName, 0)
		}
	}()
}

func selfhost_compiler_ts_tsEnumHasInstanceMethod(enums []IREnumType, typeName string, methodName string, index int) bool {
	return func() bool {
		if index >= len(enums) {
			return false
		}
		return func() bool {
			if enums[index].name == typeName {
				return selfhost_compiler_ts_tsStructMethodMatches(enums[index].methods, methodName, 0)
			}
			return selfhost_compiler_ts_tsEnumHasInstanceMethod(enums, typeName, methodName, index+1)
		}()
	}()
}

func selfhost_compiler_ts_tsStructHasInstanceMethod(structs []IRStructType, typeName string, methodName string, index int) bool {
	return func() bool {
		if index >= len(structs) {
			return false
		}
		return func() bool {
			if structs[index].name == typeName {
				return selfhost_compiler_ts_tsStructMethodMatches(structs[index].methods, methodName, 0)
			}
			return selfhost_compiler_ts_tsStructHasInstanceMethod(structs, typeName, methodName, index+1)
		}()
	}()
}

func selfhost_compiler_ts_tsStructMethodMatches(methods []IRFunction, methodName string, index int) bool {
	return func() bool {
		if index >= len(methods) {
			return false
		}
		return func() bool {
			if methods[index].name == methodName && methods[index].static == false {
				return true
			}
			return selfhost_compiler_ts_tsStructMethodMatches(methods, methodName, index+1)
		}()
	}()
}

func selfhost_compiler_ts_emitTSConst(constant IRConst) string {
	return "const " + mangleIdent(constant.name) + ": " + selfhost_compiler_ts_tsType(constant.typeName) + " = " + selfhost_compiler_ts_emitTSExprExpected(constant.value, constant.typeName) + ";"
}

func selfhost_compiler_ts_emitTSImportList(imports []IRTSImport, index int, out string) string {
	return func() string {
		if index >= len(imports) {
			return out
		}
		return selfhost_compiler_ts_emitTSImportList(imports, index+1, out+selfhost_compiler_ts_emitTSImport(imports[index]))
	}()
}

func selfhost_compiler_ts_emitTSImport(importDecl IRTSImport) string {
	names := selfhost_compiler_ts_emitTSImportNames(importDecl.functions, importDecl.values, 0, 0, "")
	return func() string {
		if names == "" {
			return ""
		}
		return "import { " + names + " } from " + selfhost_compiler_ts_tsQuoteString(selfhost_compiler_ts_tsRuntimeSpecifier(importDecl.specifier)) + ";\n"
	}()
}

func selfhost_compiler_ts_emitTSImportNames(functions []IRFunction, values []IRConst, fnIndex int, valueIndex int, out string) string {
	return func() string {
		if fnIndex < len(functions) {
			return selfhost_compiler_ts_emitTSImportNames(functions, values, fnIndex+1, valueIndex, selfhost_compiler_ts_appendTSImportName(out, functions[fnIndex].name))
		}
		return func() string {
			if valueIndex < len(values) {
				return selfhost_compiler_ts_emitTSImportNames(functions, values, fnIndex, valueIndex+1, selfhost_compiler_ts_appendTSImportName(out, values[valueIndex].name))
			}
			return out
		}()
	}()
}

func selfhost_compiler_ts_appendTSImportName(out string, name string) string {
	return out + func() string {
		if out == "" {
			return ""
		}
		return ", "
	}() + name + " as " + mangleIdent(name)
}

func selfhost_compiler_ts_tsQuoteString(value string) string {
	return "\"" + strings.ReplaceAll((strings.ReplaceAll(value, "\\", "\\\\")), "\"", "\\\"") + "\""
}

func selfhost_compiler_ts_tsRuntimeSpecifier(specifier string) string {
	return func() string {
		if specifier == "" || strings.HasPrefix(specifier, "./") || strings.HasPrefix(specifier, "../") || strings.HasPrefix(specifier, "/") || strings.Contains(specifier, "://") {
			return specifier
		}
		return "./" + specifier
	}()
}

func selfhost_compiler_ts_emitTSUnwrapHelper() string {
	return "function __runeUnwrap(value: any): any {\n  if (value && value.__tag === 0) return value.__payload?.[0];\n  if (value && value.__payload && value.__payload.length > 0) throw value.__payload[0];\n  throw new Error(\"Result.Err\");\n}\n\n"
}

func selfhost_compiler_ts_emitTSPathHelpers() string {
	return "function __runePathBasename(path: string): string {\n  const index = path.lastIndexOf(\"/\");\n  if (index < 0) return path;\n  if (index === path.length - 1) return path;\n  return path.slice(index + 1);\n}\n\nfunction __runePathExtname(path: string): string {\n  const base = __runePathBasename(path);\n  const index = base.lastIndexOf(\".\");\n  if (index <= 0) return \"\";\n  return base.slice(index);\n}\n\nfunction __runePathDirname(path: string): string {\n  const index = path.lastIndexOf(\"/\");\n  if (index < 0) return \".\";\n  if (index === 0) return \"/\";\n  return path.slice(0, index);\n}\n\nfunction __runePathJoin(parts: string[]): string {\n  return __runePathNormalize(__runePathJoinParts(parts, 0, \"\"));\n}\n\nfunction __runePathNormalize(path: string): string {\n  const absolute = path.startsWith(\"/\");\n  const out = __runePathNormalizeParts(path.split(\"/\"), 0, absolute, []);\n  const joined = __runePathJoinParts(out, 0, \"\");\n  if (absolute) return \"/\" + joined;\n  return joined === \"\" ? \".\" : joined;\n}\n\nfunction __runePathResolve(parts: string[]): string {\n  if (parts.length === 0) return \".\";\n  return __runePathNormalize(__runePathJoin(parts));\n}\n\nfunction __runePathRelative(from: string, to: string): string {\n  const fromParts = __runePathParts(__runePathResolve([from]));\n  const toParts = __runePathParts(__runePathResolve([to]));\n  let index = 0;\n  while (index < fromParts.length && index < toParts.length && fromParts[index] === toParts[index]) index++;\n  let out = \"\";\n  for (let i = index; i < fromParts.length; i++) out = __runePathAppendPart(out, \"..\");\n  for (let i = index; i < toParts.length; i++) out = __runePathAppendPart(out, toParts[i]);\n  return out === \"\" ? \".\" : out;\n}\n\nfunction __runePathParts(path: string): string[] {\n  return __runePathNormalize(path).split(\"/\").filter((part) => part !== \"\");\n}\n\nfunction __runePathJoinParts(parts: string[], index: number, out: string): string {\n  for (let i = index; i < parts.length; i++) out = __runePathAppendPart(out, parts[i]);\n  return out;\n}\n\nfunction __runePathAppendPart(out: string, part: string): string {\n  if (out === \"\") return part;\n  if (part === \"\") return out;\n  return out + \"/\" + part;\n}\n\nfunction __runePathNormalizeParts(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  while (index < parts.length) {\n    const part = parts[index];\n    if (part === \"\" || part === \".\") {\n      index++;\n      continue;\n    }\n    if (part === \"..\") return __runePathNormalizeParent(parts, index, absolute, out);\n    return __runePathNormalizePush(parts, index, absolute, out, part);\n  }\n  return out;\n}\n\nfunction __runePathNormalizeParent(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  if (out.length > 0) return __runePathNormalizePop(parts, index, absolute, out);\n  if (absolute) return __runePathNormalizeParts(parts, index + 1, absolute, out);\n  return __runePathNormalizePush(parts, index, absolute, out, \"..\");\n}\n\nfunction __runePathNormalizePop(parts: string[], index: number, absolute: boolean, out: string[]): string[] {\n  return __runePathNormalizeParts(parts, index + 1, absolute, out.slice(0, out.length - 1));\n}\n\nfunction __runePathNormalizePush(parts: string[], index: number, absolute: boolean, out: string[], part: string): string[] {\n  return __runePathNormalizeParts(parts, index + 1, absolute, [...out, part]);\n}\n\nfunction __runePathCollectParts(parts: string[], index: number, out: string[]): string[] {\n  for (let i = index; i < parts.length; i++) {\n    if (parts[i] !== \"\") out.push(parts[i]);\n  }\n  return out;\n}\n\nfunction __runePathCollectPart(parts: string[], index: number, out: string[]): string[] {\n  if (index < parts.length) out.push(parts[index]);\n  return __runePathCollectParts(parts, index + 1, out);\n}\n\nfunction __runePathRelativeFromParts(fromParts: string[], toParts: string[], index: number): string {\n  while (index < fromParts.length && index < toParts.length && fromParts[index] === toParts[index]) index++;\n  return __runePathRelativeTail(fromParts, toParts, index, index, \"\");\n}\n\nfunction __runePathRelativeTail(fromParts: string[], toParts: string[], fromIndex: number, toIndex: number, out: string): string {\n  for (let i = fromIndex; i < fromParts.length; i++) out = __runePathAppendPart(out, \"..\");\n  for (let i = toIndex; i < toParts.length; i++) out = __runePathAppendPart(out, toParts[i]);\n  return out === \"\" ? \".\" : out;\n}\n\n"
}

func selfhost_compiler_ts_emitTSEnum(enumDecl IREnumType) string {
	return func() string {
		if selfhost_compiler_ts_tsEnumHasPayload(enumDecl.members, 0) {
			return selfhost_compiler_ts_emitTSPayloadEnum(enumDecl)
		}
		return selfhost_compiler_ts_emitTSSimpleEnum(enumDecl)
	}()
}

func selfhost_compiler_ts_emitTSSimpleEnum(enumDecl IREnumType) string {
	out := "type " + mangleIdent(enumDecl.name) + " = number;\n"
	out = out + "const " + mangleIdent(enumDecl.name) + " = {\n"
	out = out + selfhost_compiler_ts_emitTSEnumMembers(enumDecl.members, 0, "")
	return out + "} as const;\n"
}

func selfhost_compiler_ts_emitTSPayloadEnum(enumDecl IREnumType) string {
	out := "type " + mangleIdent(enumDecl.name) + selfhost_compiler_ts_emitTSGenerics(enumDecl.generics) + " =\n"
	out = out + selfhost_compiler_ts_emitTSPayloadEnumMembers(enumDecl.members, 0, "")
	out = out + ";\n\n"
	return out + selfhost_compiler_ts_emitTSPayloadEnumConstructors(enumDecl.name, enumDecl.generics, enumDecl.members, 0, "")
}

func selfhost_compiler_ts_emitTSPayloadEnumMembers(members []IREnumMember, index int, out string) string {
	return func() string {
		if index >= len(members) {
			return out
		}
		return selfhost_compiler_ts_emitTSPayloadEnumMembers(members, index+1, out+selfhost_compiler_ts_emitTSPayloadEnumMember(members[index], index, index == 0))
	}()
}

func selfhost_compiler_ts_emitTSPayloadEnumMember(member IREnumMember, index int, first bool) string {
	prefix := func() string {
		if first {
			return "  "
		}
		return "| "
	}()
	return prefix + "{ __tag: " + compilerIntToString(index) + "; __payload: " + selfhost_compiler_ts_emitTSPayloadTuple(member.params, 0, "") + " }\n"
}

func selfhost_compiler_ts_emitTSPayloadTuple(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return "[" + out + "]"
		}
		return selfhost_compiler_ts_emitTSPayloadTuple(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_ts_tsType(params[index].typeName))
	}()
}

func selfhost_compiler_ts_emitTSPayloadEnumConstructors(enumName string, generics []string, members []IREnumMember, index int, out string) string {
	return func() string {
		if index >= len(members) {
			return out
		}
		return selfhost_compiler_ts_emitTSPayloadEnumConstructors(enumName, generics, members, index+1, out+selfhost_compiler_ts_emitTSPayloadEnumConstructor(enumName, generics, members[index], index))
	}()
}

func selfhost_compiler_ts_emitTSPayloadEnumConstructor(enumName string, generics []string, member IREnumMember, index int) string {
	typeName := mangleIdent(enumName) + selfhost_compiler_ts_emitTSGenericsUse(generics)
	return func() string {
		if len(member.params) == 0 {
			return "const " + mangleIdent(enumName+"_"+member.name) + ": " + typeName + " = { __tag: " + compilerIntToString(index) + ", __payload: [] };\n"
		}
		return "function " + mangleIdent(member.name) + selfhost_compiler_ts_emitTSGenerics(generics) + "(" + selfhost_compiler_ts_emitTSParams(member.params, 0, "") + "): " + typeName + " {\n" + line(1, "return { __tag: "+compilerIntToString(index)+", __payload: ["+selfhost_compiler_ts_emitTSParamNames(member.params, 0, "")+"] };") + "}\n"
	}()
}

func selfhost_compiler_ts_emitTSParamNames(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_ts_emitTSParamNames(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(params[index].name))
	}()
}

func selfhost_compiler_ts_tsEnumHasPayload(members []IREnumMember, index int) bool {
	return func() bool {
		if index >= len(members) {
			return false
		}
		return func() bool {
			if len(members[index].params) > 0 {
				return true
			}
			return selfhost_compiler_ts_tsEnumHasPayload(members, index+1)
		}()
	}()
}

func selfhost_compiler_ts_emitTSGenericsUse(generics []string) string {
	return func() string {
		if len(generics) == 0 {
			return ""
		}
		return "<" + selfhost_compiler_ts_emitTSGenericNames(generics, 0, "") + ">"
	}()
}

func selfhost_compiler_ts_emitTSEnumMembers(members []IREnumMember, index int, out string) string {
	return func() string {
		if index >= len(members) {
			return out
		}
		return selfhost_compiler_ts_emitTSEnumMembers(members, index+1, out+indent(1)+selfhost_compiler_ts_tsPropertyName(members[index].name)+": "+enumValue(members[index], index)+",\n")
	}()
}

func selfhost_compiler_ts_emitTSStruct(typeDecl IRStructType) string {
	generics := selfhost_compiler_ts_emitTSGenerics(typeDecl.generics)
	out := "type " + mangleIdent(typeDecl.name) + generics + " = {\n"
	for _, field := range typeDecl.fields {
		_ = field
		out = out + indent(1) + selfhost_compiler_ts_tsPropertyName(field.name) + ": " + selfhost_compiler_ts_tsType(field.typeName) + ";\n"
	}
	return out + "};\n"
}

func selfhost_compiler_ts_emitTSMethods(typeDecl IRStructType) string {
	out := ""
	for _, method := range typeDecl.methods {
		_ = method
		out = out + selfhost_compiler_ts_emitTSFunction(selfhost_compiler_ts_methodWithTSReceiver(typeDecl.name, method)) + "\n"
	}
	return out
}

func selfhost_compiler_ts_emitTSEnumMethods(enumDecl IREnumType) string {
	out := ""
	for _, method := range enumDecl.methods {
		_ = method
		out = out + selfhost_compiler_ts_emitTSFunction(selfhost_compiler_ts_methodWithTSReceiver(enumDecl.name, method)) + "\n"
	}
	return out
}

func selfhost_compiler_ts_methodWithTSReceiver(typeName string, method IRFunction) IRFunction {
	return IRFunction{name: typeName + "_" + method.name, private: method.private, static: method.static, routine: method.routine, macro: method.macro, receiverType: method.receiverType, generics: method.generics, params: func() []IRParam {
		switch {
		case method.static == true:
			return method.params
		case method.static == false:
			return selfhost_compiler_ts_prependThisParam(typeName, method.params)
		}
		return nil
	}(), returnType: method.returnType, body: method.body, sourcePath: method.sourcePath, line: method.line, column: method.column}
}

func selfhost_compiler_ts_prependThisParam(typeName string, params []IRParam) []IRParam {
	out := []IRParam{IRParam{name: "this", typeName: typeName, line: 0, column: 0}}
	for _, param := range params {
		_ = param
		func() int { out = append(out, param); return len(out) }()
	}
	return out
}

func selfhost_compiler_ts_emitTSFunction(fn IRFunction) string {
	return func() string {
		if fn.returnType == "WebComponent" {
			return selfhost_compiler_ts_emitTSWebComponentFunction(fn)
		}
		return selfhost_compiler_ts_emitTSPlainFunction(fn)
	}()
}

func selfhost_compiler_ts_emitTSPlainFunction(fn IRFunction) string {
	ret := selfhost_compiler_ts_tsType(fn.returnType)
	out := "function " + mangleIdent(fn.name) + selfhost_compiler_ts_emitTSGenerics(fn.generics) + "(" + selfhost_compiler_ts_emitTSParams(fn.params, 0, "") + "): " + ret + " {\n"
	out = out + selfhost_compiler_ts_emitTSBody(fn.body, returnsValue(fn.returnType), fn.returnType, 1)
	return out + "}\n"
}

func selfhost_compiler_ts_emitTSWebComponentFunction(fn IRFunction) string {
	out := "function " + mangleIdent(fn.name) + selfhost_compiler_ts_emitTSGenerics(fn.generics) + "(" + selfhost_compiler_ts_emitTSParams(fn.params, 0, "") + "): " + selfhost_compiler_ts_tsType(fn.returnType) + " {\n"
	out = out + line(1, "return class extends HTMLElement {")
	out = out + line(2, "connectedCallback(): void {")
	out = out + line(3, "if ((this as any).__runeMounted) return;")
	out = out + line(3, "(this as any).__runeMounted = true;")
	out = out + line(3, "const __root = "+selfhost_compiler_ts_emitTSExprExpected(fn.body, "HTMLElement")+";")
	out = out + line(3, "this.appendChild(__root);")
	out = out + line(2, "}")
	out = out + line(1, "};")
	return out + "}\n"
}

func selfhost_compiler_ts_emitTSBody(expr IRExpr, returns bool, returnType string, level int) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_ts_emitTSBlock(expr.children, 0, returns, returnType, level, "")
		default:
			return line(level, func() string {
				if returns {
					return "return " + selfhost_compiler_ts_emitTSExprExpected(expr, returnType) + ";"
				}
				return selfhost_compiler_ts_emitTSExpr(expr) + ";"
			}())
		}
	}()
}

func selfhost_compiler_ts_emitTSBlock(statements []IRExpr, index int, returns bool, returnType string, level int, out string) string {
	return func() string {
		if index >= len(statements) {
			return func() string {
				if returns && len(statements) == 0 {
					return out + line(level, "return undefined;")
				}
				return out
			}()
		}
		return selfhost_compiler_ts_emitTSBlock(statements, index+1, returns, returnType, level, out+selfhost_compiler_ts_emitTSStatement(statements[index], index == len(statements)-1, returns, returnType, level))
	}()
}

func selfhost_compiler_ts_emitTSStatement(expr IRExpr, last bool, returns bool, returnType string, level int) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Let:
			return selfhost_compiler_ts_emitTSLet(expr, level)
		case expr.kind == ExprKind_ObjectDestructure:
			return selfhost_compiler_ts_emitTSObjectDestructure(expr, level)
		default:
			return func() string {
				if last && returns {
					return line(level, "return "+selfhost_compiler_ts_emitTSExprExpected(expr, returnType)+";")
				}
				return line(level, selfhost_compiler_ts_emitTSExpr(expr)+";")
			}()
		}
	}()
}

func selfhost_compiler_ts_emitTSLet(expr IRExpr, level int) string {
	return line(level, selfhost_compiler_ts_tsLetKeyword(expr.op)+mangleIdent(expr.name)+" = "+selfhost_compiler_ts_emitTSExpr(expr.children[0])+";")
}

func selfhost_compiler_ts_tsLetKeyword(op string) string {
	return func() string {
		switch {
		case op == ":=:":
			return "let "
		default:
			return "const "
		}
	}()
}

func selfhost_compiler_ts_emitTSObjectDestructure(expr IRExpr, level int) string {
	return line(level, "const { "+selfhost_compiler_ts_emitTSObjectDestructureFields(expr.params, 0, "")+" } = "+selfhost_compiler_ts_emitTSExpr(expr.children[0])+";")
}

func selfhost_compiler_ts_emitTSObjectDestructureFields(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_ts_emitTSObjectDestructureFields(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_ts_tsPropertyName(params[index].typeName)+": "+mangleIdent(params[index].name))
	}()
}

func selfhost_compiler_ts_emitTSParams(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_ts_emitTSParams(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(params[index].name)+": "+selfhost_compiler_ts_tsType(params[index].typeName))
	}()
}

func selfhost_compiler_ts_emitTSGenerics(generics []string) string {
	return func() string {
		if len(generics) == 0 {
			return ""
		}
		return "<" + selfhost_compiler_ts_emitTSGenericNames(generics, 0, "") + ">"
	}()
}

func selfhost_compiler_ts_emitTSGenericNames(generics []string, index int, out string) string {
	return func() string {
		if index >= len(generics) {
			return out
		}
		return selfhost_compiler_ts_emitTSGenericNames(generics, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(generics[index]))
	}()
}

func selfhost_compiler_ts_emitTSExports(file IRFile) string {
	exports := selfhost_compiler_ts_emitTSConstExportNames(file.constants, 0, selfhost_compiler_ts_emitTSExportNames(file.functions, 0, ""))
	return func() string {
		if exports == "" {
			return ""
		}
		return "export { " + exports + " };\n"
	}()
}

func selfhost_compiler_ts_emitTSConstExportNames(constants []IRConst, index int, out string) string {
	return func() string {
		if index >= len(constants) {
			return out
		}
		return func() string {
			if constants[index].private {
				return selfhost_compiler_ts_emitTSConstExportNames(constants, index+1, out)
			}
			return selfhost_compiler_ts_emitTSConstExportNames(constants, index+1, selfhost_compiler_ts_appendTSExportName(out, constants[index].name))
		}()
	}()
}

func selfhost_compiler_ts_emitTSExportNames(functions []IRFunction, index int, out string) string {
	return func() string {
		if index >= len(functions) {
			return out
		}
		return func() string {
			if functions[index].macro || functions[index].private {
				return selfhost_compiler_ts_emitTSExportNames(functions, index+1, out)
			}
			return selfhost_compiler_ts_emitTSExportNames(functions, index+1, selfhost_compiler_ts_appendTSExportName(out, functions[index].name))
		}()
	}()
}

func selfhost_compiler_ts_appendTSExportName(out string, name string) string {
	return out + func() string {
		if out == "" {
			return ""
		}
		return ", "
	}() + mangleIdent(name) + " as " + name
}

func selfhost_compiler_ts_emitTSExpr(expr IRExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Identifier:
			return mangleIdent(expr.name)
		case expr.kind == ExprKind_At:
			return "@" + expr.name
		case expr.kind == ExprKind_This:
			return mangleIdent("this")
		case expr.kind == ExprKind_Int:
			return expr.value
		case expr.kind == ExprKind_Double:
			return expr.value
		case expr.kind == ExprKind_BigInt:
			return bigintLiteralDigits(expr.value) + "n"
		case expr.kind == ExprKind_String:
			return expr.value
		case expr.kind == ExprKind_Template:
			return selfhost_compiler_ts_emitTSTemplate(expr)
		case expr.kind == ExprKind_Char:
			return expr.value
		case expr.kind == ExprKind_Regex:
			return expr.value
		case expr.kind == ExprKind_XMLText:
			return selfhost_compiler_ts_tsQuoteString(expr.value)
		case expr.kind == ExprKind_XMLElement:
			return selfhost_compiler_ts_emitTSXMLExpr(expr)
		case expr.kind == ExprKind_Bool:
			return expr.value
		case expr.kind == ExprKind_Null:
			return "null"
		case expr.kind == ExprKind_Unary:
			return expr.op + selfhost_compiler_ts_emitTSExpr(expr.children[0])
		case expr.kind == ExprKind_Postfix:
			return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + expr.op
		case expr.kind == ExprKind_CompileTime:
			return selfhost_compiler_ts_emitTSExpr(expr.children[0])
		case expr.kind == ExprKind_Unwrap:
			return "__runeUnwrap(" + selfhost_compiler_ts_emitTSExpr(expr.children[0]) + ")"
		case expr.kind == ExprKind_Binary:
			return selfhost_compiler_ts_emitTSBinary(expr)
		case expr.kind == ExprKind_Ternary:
			return selfhost_compiler_ts_emitTSTernary(expr)
		case expr.kind == ExprKind_Assign:
			return selfhost_compiler_ts_emitTSAssign(expr)
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_ts_emitTSCall(expr)
		case expr.kind == ExprKind_Lambda:
			return selfhost_compiler_ts_emitTSLambda(expr)
		case expr.kind == ExprKind_Selector:
			return selfhost_compiler_ts_emitTSSelector(expr)
		case expr.kind == ExprKind_Index:
			return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + "[" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + "]"
		case expr.kind == ExprKind_Array:
			return "[" + selfhost_compiler_ts_emitTSExprList(expr.children, 0, "") + "]"
		case expr.kind == ExprKind_Tuple:
			return "[" + selfhost_compiler_ts_emitTSExprList(expr.children, 0, "") + "]"
		case expr.kind == ExprKind_Map:
			return "new Map([" + selfhost_compiler_ts_emitTSMapEntries(expr.children, 0, "") + "])"
		case expr.kind == ExprKind_Spread:
			return "..." + selfhost_compiler_ts_emitTSExpr(expr.children[0])
		case expr.kind == ExprKind_Reactive:
			return selfhost_compiler_ts_emitTSExpr(expr.children[0])
		case expr.kind == ExprKind_Struct:
			return "{" + selfhost_compiler_ts_emitTSFields(expr.children, 0, "") + "}"
		case expr.kind == ExprKind_Object:
			return "{" + selfhost_compiler_ts_emitTSFields(expr.children, 0, "") + "}"
		case expr.kind == ExprKind_Block:
			return "(() => {\n" + selfhost_compiler_ts_emitTSBlock(expr.children, 0, true, "Dynamic", 1, "") + "})()"
		default:
			return "undefined"
		}
	}()
}

func selfhost_compiler_ts_emitTSXMLExpr(expr IRExpr) string {
	return "(() => {\n" + selfhost_compiler_ts_emitTSXMLExprBody(expr, "__el0", 1) + line(1, "return __el0;") + "})()"
}

func selfhost_compiler_ts_emitTSXMLCreateElement(expr IRExpr) string {
	return "document.createElement(" + selfhost_compiler_ts_tsQuoteString(expr.name) + ")"
}

func selfhost_compiler_ts_emitTSXMLExprBody(expr IRExpr, name string, level int) string {
	out := line(level, "const "+name+" = "+selfhost_compiler_ts_emitTSXMLCreateElement(expr)+";")
	out = out + selfhost_compiler_ts_emitTSXMLParts(name, expr.children, 0, level, 1)
	return out
}

func selfhost_compiler_ts_emitTSXMLParts(parent string, children []IRExpr, index int, level int, childIndex int) string {
	return func() string {
		if index >= len(children) {
			return ""
		}
		return selfhost_compiler_ts_emitTSXMLPart(parent, children, index, level, childIndex)
	}()
}

func selfhost_compiler_ts_emitTSXMLPart(parent string, children []IRExpr, index int, level int, childIndex int) string {
	child := children[index]
	nextIndex := func() int {
		if child.kind == ExprKind_XMLElement {
			return childIndex + 1
		}
		return childIndex
	}()
	return selfhost_compiler_ts_emitTSXMLChild(parent, child, level, childIndex) + selfhost_compiler_ts_emitTSXMLParts(parent, children, index+1, level, nextIndex)
}

func selfhost_compiler_ts_emitTSXMLChild(parent string, child IRExpr, level int, childIndex int) string {
	return func() string {
		switch {
		case child.kind == ExprKind_Field:
			return selfhost_compiler_ts_emitTSXMLAttr(parent, child, level)
		case child.kind == ExprKind_XMLText:
			return line(level, parent+".appendChild(document.createTextNode("+selfhost_compiler_ts_tsQuoteString(child.value)+"));")
		case child.kind == ExprKind_XMLElement:
			return selfhost_compiler_ts_emitTSXMLNestedChild(parent, child, level, childIndex)
		default:
			return selfhost_compiler_ts_emitTSXMLExprChild(parent, child, level, childIndex)
		}
	}()
}

func selfhost_compiler_ts_emitTSXMLNestedChild(parent string, child IRExpr, level int, childIndex int) string {
	name := parent + "_" + strconv.Itoa(childIndex)
	return selfhost_compiler_ts_emitTSXMLExprBody(child, name, level) + line(level, parent+".appendChild("+name+");")
}

func selfhost_compiler_ts_emitTSXMLExprChild(parent string, child IRExpr, level int, childIndex int) string {
	childType := child.text
	childName := parent + "_child" + strconv.Itoa(childIndex)
	return func() string {
		if strings.HasPrefix(childType, "Array[") {
			return line(level, "for (const "+childName+" of "+selfhost_compiler_ts_emitTSExpr(child)+") {") + line(level+1, parent+".appendChild("+childName+");") + line(level, "}")
		}
		return func() string {
			if childType == "HTMLElement" {
				return line(level, parent+".appendChild("+selfhost_compiler_ts_emitTSExpr(child)+");")
			}
			return line(level, parent+".appendChild(document.createTextNode(String("+selfhost_compiler_ts_emitTSExpr(child)+")));")
		}()
	}()
}

func selfhost_compiler_ts_emitTSXMLAttr(parent string, attr IRExpr, level int) string {
	return func() string {
		if attr.op == "xmlEvent" {
			return selfhost_compiler_ts_emitTSXMLEventAttr(parent, attr, level)
		}
		return selfhost_compiler_ts_emitTSXMLPlainAttr(parent, attr, level)
	}()
}

func selfhost_compiler_ts_emitTSXMLEventAttr(parent string, attr IRExpr, level int) string {
	return func() string {
		if len(attr.children) == 0 {
			return ""
		}
		return line(level, parent+".addEventListener("+selfhost_compiler_ts_tsQuoteString(attr.name)+", () => { "+selfhost_compiler_ts_emitTSExpr(attr.children[0])+"; });")
	}()
}

func selfhost_compiler_ts_emitTSXMLPlainAttr(parent string, attr IRExpr, level int) string {
	prefix := parent + ".setAttribute(" + selfhost_compiler_ts_tsQuoteString(attr.name)
	return func() string {
		if len(attr.children) == 0 {
			return line(level, prefix+", \"\");")
		}
		return line(level, prefix+", String("+selfhost_compiler_ts_emitTSExpr(attr.children[0])+"));")
	}()
}

func selfhost_compiler_ts_emitTSTemplate(expr IRExpr) string {
	raw := expr.value
	inner := func() string {
		if len([]rune(raw)) >= 2 {
			return func() string { runes := []rune(raw); return string(runes[1 : len([]rune(raw))-1]) }()
		}
		return raw
	}()
	return selfhost_compiler_ts_emitTSTemplateParts(func() []string { parts := strings.Split(inner, "<<<RUNE_TEMPLATE_PART>>>"); return parts }(), expr.children, 0, "`") + "`"
}

func selfhost_compiler_ts_emitTSTemplateParts(segments []string, children []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(segments) {
			return out
		}
		return selfhost_compiler_ts_emitTSTemplateAtom(segments, children, index, out)
	}()
}

func selfhost_compiler_ts_emitTSTemplateAtom(segments []string, children []IRExpr, index int, out string) string {
	textOut := out + selfhost_compiler_ts_tsEscapeTemplateText(segments[index])
	next := func() string {
		if index < len(children) {
			return textOut + ("${" + selfhost_compiler_ts_emitTSExpr(children[index]) + "}")
		}
		return textOut
	}()
	return selfhost_compiler_ts_emitTSTemplateParts(segments, children, index+1, next)
}

func selfhost_compiler_ts_tsEscapeTemplateText(text string) string {
	return selfhost_compiler_ts_tsEscapeTemplateChars(text, 0, "")
}

func selfhost_compiler_ts_tsEscapeTemplateChars(text string, index int, out string) string {
	return func() string {
		if index >= len([]rune(text)) {
			return out
		}
		return selfhost_compiler_ts_tsEscapeTemplateChar(text, index, out)
	}()
}

func selfhost_compiler_ts_tsEscapeTemplateChar(text string, index int, out string) string {
	return func() string {
		switch {
		case []rune(text)[index] == '$' && index+1 < len([]rune(text)) && []rune(text)[index+1] == '{' == true:
			return selfhost_compiler_ts_tsEscapeTemplateChars(text, index+2, out+"\\${")
		default:
			return selfhost_compiler_ts_tsEscapeTemplateChars(text, index+1, out+string(([]rune(text)[index])))
		}
	}()
}

func selfhost_compiler_ts_emitTSExprExpected(expr IRExpr, expected string) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_ts_emitTSCallExpected(expr, expected)
		default:
			return selfhost_compiler_ts_emitTSExpr(expr)
		}
	}()
}

func selfhost_compiler_ts_emitTSCallExpected(expr IRExpr, expected string) string {
	args := genericInner(expected, "Result")
	return func() string {
		if args != "" && selfhost_compiler_ts_isTSResultConstructorCall(expr) {
			return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + "<" + selfhost_compiler_ts_emitTSTypeArgs(args) + ">(" + selfhost_compiler_ts_emitTSExprListFrom(expr.children, 1, "") + ")"
		}
		return selfhost_compiler_ts_emitTSCall(expr)
	}()
}

func selfhost_compiler_ts_isTSResultConstructorCall(expr IRExpr) bool {
	return expr.kind == ExprKind_Call && len(expr.children) > 0 && expr.children[0].kind == ExprKind_Identifier && (expr.children[0].name == "Ok" || expr.children[0].name == "Err")
}

func selfhost_compiler_ts_emitTSBinary(expr IRExpr) string {
	return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + " " + selfhost_compiler_ts_tsBinaryOp(expr.op) + " " + selfhost_compiler_ts_emitTSExpr(expr.children[1])
}

func selfhost_compiler_ts_emitTSTernary(expr IRExpr) string {
	return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + " ? " + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + " : " + func() string {
		if len(expr.children) > 2 {
			return selfhost_compiler_ts_emitTSExpr(expr.children[2])
		}
		return "undefined"
	}()
}

func selfhost_compiler_ts_emitTSAssign(expr IRExpr) string {
	return func() string {
		if len(expr.children) == 2 {
			return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + " = " + selfhost_compiler_ts_emitTSExpr(expr.children[1])
		}
		return mangleIdent(expr.name) + " = " + selfhost_compiler_ts_emitTSExpr(expr.children[0])
	}()
}

func selfhost_compiler_ts_emitTSCall(expr IRExpr) string {
	return func() string {
		switch {
		case moduleCallKey(expr) == "io.println":
			return "console.log(" + selfhost_compiler_ts_emitTSExprListFrom(expr.children, 1, "") + ")"
		case moduleCallKey(expr) == "json.stringify":
			return "JSON.stringify(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "json.parse":
			return "JSON.parse(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "map.new":
			return "new Map()"
		case moduleCallKey(expr) == "set.new":
			return "new Set()"
		case moduleCallKey(expr) == "path.isAbsolute":
			return selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ".startsWith(\"/\")"
		case moduleCallKey(expr) == "path.basename":
			return "__runePathBasename(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.extname":
			return "__runePathExtname(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.dirname":
			return "__runePathDirname(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.join":
			return "__runePathJoin(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.normalize":
			return "__runePathNormalize(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.resolve":
			return "__runePathResolve(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.relative":
			return "__runePathRelative(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ")"
		case moduleCallKey(expr) == "path.joinParts":
			return "__runePathJoinParts(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ")"
		case moduleCallKey(expr) == "path.appendPathPart":
			return "__runePathAppendPart(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ")"
		case moduleCallKey(expr) == "path.normalizeParts":
			return "__runePathNormalizeParts(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[4]) + ")"
		case moduleCallKey(expr) == "path.normalizePart":
			return "__runePathNormalizeParts(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[4]) + ")"
		case moduleCallKey(expr) == "path.normalizeParent":
			return "__runePathNormalizeParent(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[4]) + ")"
		case moduleCallKey(expr) == "path.normalizePop":
			return "__runePathNormalizePop(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[4]) + ")"
		case moduleCallKey(expr) == "path.normalizePush":
			return "__runePathNormalizePush(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[4]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[5]) + ")"
		case moduleCallKey(expr) == "path.pathParts":
			return "__runePathParts(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "path.collectPathParts":
			return "__runePathCollectParts(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ")"
		case moduleCallKey(expr) == "path.collectPathPart":
			return "__runePathCollectPart(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ")"
		case moduleCallKey(expr) == "path.relativeFromParts":
			return "__runePathRelativeFromParts(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ")"
		case moduleCallKey(expr) == "path.relativeTail":
			return "__runePathRelativeTail(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[3]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[4]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[5]) + ")"
		case moduleCallKey(expr) == "process.platform":
			return "\"js\""
		case moduleCallKey(expr) == "process.cwd":
			return "\".\""
		case moduleCallKey(expr) == "process.env":
			return "null"
		case moduleCallKey(expr) == "process.argv":
			return "[]"
		case moduleCallKey(expr) == "int.toString":
			return "String(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int.toDouble":
			return selfhost_compiler_ts_emitTSExpr(expr.children[1])
		case moduleCallKey(expr) == "int.toBigInt":
			return "BigInt(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int4.fromInt":
			return "((__value: number): number => { const __n = __value & 0xf; return __n >= 8 ? __n - 16 : __n; })(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int8.fromInt":
			return "((__value: number): number => (__value << 24) >> 24)(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int16.fromInt":
			return "((__value: number): number => (__value << 16) >> 16)(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "int64.fromInt":
			return "BigInt(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "uint.fromInt":
			return "(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + " >>> 0)"
		case moduleCallKey(expr) == "uint8.fromInt":
			return "(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + " & 0xff)"
		case moduleCallKey(expr) == "uint16.fromInt":
			return "(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + " & 0xffff)"
		case moduleCallKey(expr) == "uint64.fromInt":
			return "BigInt.asUintN(64, BigInt(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + "))"
		case moduleCallKey(expr) == "float.fromDouble":
			return "Math.fround(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case (moduleCallKey(expr) == "int4.toInt") || (moduleCallKey(expr) == "int8.toInt") || (moduleCallKey(expr) == "int16.toInt") || (moduleCallKey(expr) == "uint.toInt") || (moduleCallKey(expr) == "uint8.toInt") || (moduleCallKey(expr) == "uint16.toInt") || (moduleCallKey(expr) == "float.toDouble"):
			return selfhost_compiler_ts_emitTSExpr(expr.children[1])
		case (moduleCallKey(expr) == "int64.toInt") || (moduleCallKey(expr) == "uint64.toInt"):
			return "Number(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "bigint.fromInt":
			return "BigInt(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "bigint.toString":
			return "String(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "bigint.toDouble":
			return "Number(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "double.trunc":
			return "Math.trunc(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "double.floor":
			return "Math.floor(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "double.ceil":
			return "Math.ceil(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		case moduleCallKey(expr) == "double.round":
			return "Math.round(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ")"
		default:
			return selfhost_compiler_ts_emitTSMaybeCoreMethodCall(expr)
		}
	}()
}

func selfhost_compiler_ts_emitTSMaybeCoreMethodCall(expr IRExpr) string {
	return func() string {
		if len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector {
			return selfhost_compiler_ts_emitTSCoreMethodCall(expr, expr.children[0])
		}
		return selfhost_compiler_ts_emitTSDefaultCall(expr)
	}()
}

func selfhost_compiler_ts_emitTSCoreMethodCall(expr IRExpr, selector IRExpr) string {
	return func() string {
		if len(selector.children) > 0 && selector.children[0].kind != ExprKind_At {
			return func() string {
				switch {
				case (selector.name == "length") || (selector.name == "byteLength"):
					return selfhost_compiler_ts_emitTSCoreLength(selector.children[0])
				case selector.name == "isEmpty":
					return "(" + selfhost_compiler_ts_emitTSCoreLength(selector.children[0]) + " === 0)"
				case selector.name == "at":
					return selfhost_compiler_ts_emitTSCoreAt(expr, selector.children[0])
				case selector.name == "slice":
					return selfhost_compiler_ts_emitTSCoreSlice(expr, selector.children[0])
				default:
					return selfhost_compiler_ts_emitTSDefaultCall(expr)
				}
			}()
		}
		return selfhost_compiler_ts_emitTSDefaultCall(expr)
	}()
}

func selfhost_compiler_ts_emitTSCoreLength(receiver IRExpr) string {
	return func() string {
		if receiver.text == "String" {
			return "Array.from(" + selfhost_compiler_ts_emitTSExpr(receiver) + ").length"
		}
		return selfhost_compiler_ts_emitTSExpr(receiver) + ".length"
	}()
}

func selfhost_compiler_ts_emitTSCoreAt(expr IRExpr, receiver IRExpr) string {
	return func() string {
		if receiver.text == "String" {
			return "(Array.from(" + selfhost_compiler_ts_emitTSExpr(receiver) + ")[" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + "] ?? \"\")"
		}
		return selfhost_compiler_ts_emitTSExpr(receiver) + "[" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + "]"
	}()
}

func selfhost_compiler_ts_emitTSCoreSlice(expr IRExpr, receiver IRExpr) string {
	return func() string {
		if receiver.text == "String" {
			return "Array.from(" + selfhost_compiler_ts_emitTSExpr(receiver) + ").slice(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ").join(\"\")"
		}
		return selfhost_compiler_ts_emitTSExpr(receiver) + ".slice(" + selfhost_compiler_ts_emitTSExpr(expr.children[1]) + ", " + selfhost_compiler_ts_emitTSExpr(expr.children[2]) + ")"
	}()
}

func selfhost_compiler_ts_emitTSDefaultCall(expr IRExpr) string {
	return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + "(" + selfhost_compiler_ts_emitTSExprListFrom(expr.children, 1, "") + ")"
}

func selfhost_compiler_ts_emitTSLambda(expr IRExpr) string {
	return "(" + selfhost_compiler_ts_emitTSParams(expr.params, 0, "") + ") => " + selfhost_compiler_ts_emitTSExpr(expr.children[0])
}

func selfhost_compiler_ts_emitTSSelector(expr IRExpr) string {
	return func() string {
		switch {
		case expr.children[0].kind == ExprKind_At:
			return selfhost_compiler_ts_emitTSAtSelector(expr)
		case expr.children[0].kind == ExprKind_Identifier:
			return func() string {
				switch {
				case expr.op == "::":
					return mangleIdent(expr.children[0].name + "_" + expr.name)
				default:
					return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + "." + selfhost_compiler_ts_tsPropertyName(expr.name)
				}
			}()
		default:
			return selfhost_compiler_ts_emitTSExpr(expr.children[0]) + "." + selfhost_compiler_ts_tsPropertyName(expr.name)
		}
	}()
}

func selfhost_compiler_ts_emitTSAtSelector(expr IRExpr) string {
	imported := expr.children[0].value != ""
	return func() string {
		switch {
		case imported == true:
			return mangleIdent(expr.name)
		default:
			return "@" + expr.children[0].name + "." + expr.name
		}
	}()
}

func selfhost_compiler_ts_emitTSExprList(exprs []IRExpr, index int, out string) string {
	return selfhost_compiler_ts_emitTSExprListFrom(exprs, index, out)
}

func selfhost_compiler_ts_emitTSExprListFrom(exprs []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(exprs) {
			return out
		}
		return selfhost_compiler_ts_emitTSExprListFrom(exprs, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+selfhost_compiler_ts_emitTSExpr(exprs[index]))
	}()
}

func selfhost_compiler_ts_emitTSMapEntries(entries []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(entries) {
			return out
		}
		return selfhost_compiler_ts_emitTSMapEntries(entries, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+"["+selfhost_compiler_ts_emitTSExpr(entries[index].children[0])+", "+selfhost_compiler_ts_emitTSExpr(entries[index].children[1])+"]")
	}()
}

func selfhost_compiler_ts_emitTSFields(fields []IRExpr, index int, out string) string {
	return func() string {
		if index >= len(fields) {
			return out
		}
		return selfhost_compiler_ts_emitTSFields(fields, index+1, out+func() string {
			if out == "" {
				return ""
			}
			return ", "
		}()+selfhost_compiler_ts_tsPropertyName(fields[index].name)+": "+selfhost_compiler_ts_emitTSExpr(fields[index].children[0]))
	}()
}

func selfhost_compiler_ts_tsBinaryOp(op string) string {
	return func() string {
		switch {
		case op == "==":
			return "==="
		case op == "!=":
			return "!=="
		default:
			return op
		}
	}()
}

func selfhost_compiler_ts_tsType(typeName string) string {
	switch {
	case (typeName == "") || (typeName == "Void"):
		return "void"
	case (typeName == "Int") || (typeName == "Int4") || (typeName == "Int8") || (typeName == "Int16") || (typeName == "UInt") || (typeName == "UInt8") || (typeName == "UInt16") || (typeName == "Double") || (typeName == "Float"):
		return "number"
	case (typeName == "BigInt") || (typeName == "Int64") || (typeName == "UInt64"):
		return "bigint"
	case (typeName == "String") || (typeName == "Char"):
		return "string"
	case typeName == "Bool":
		return "boolean"
	case typeName == "Dynamic":
		return "any"
	case typeName == "HTMLElement":
		return "HTMLElement"
	case typeName == "WebComponent":
		return "CustomElementConstructor"
	case (typeName == "Data") || (typeName == "@io.Data"):
		return "Uint8Array"
	default:
		return selfhost_compiler_ts_tsTypeFallback(typeName)
	}
}

func selfhost_compiler_ts_tsTypeFallback(typeName string) string {
	return func() string {
		if strings.HasSuffix(typeName, "?") {
			return selfhost_compiler_ts_tsType(func() string { runes := []rune(typeName); return string(runes[0 : len([]rune(typeName))-1]) }()) + " | null"
		}
		return func() string {
			if genericInner(typeName, "Array") != "" {
				return selfhost_compiler_ts_tsType(genericInner(typeName, "Array")) + "[]"
			}
			return func() string {
				if genericInner(typeName, "ReadonlyArray") != "" {
					return "ReadonlyArray<" + selfhost_compiler_ts_tsType(genericInner(typeName, "ReadonlyArray")) + ">"
				}
				return func() string {
					if genericInner(typeName, "Map") != "" {
						return "Map<" + selfhost_compiler_ts_tsType(typeArg(genericInner(typeName, "Map"), 0)) + ", " + selfhost_compiler_ts_tsType(typeArg(genericInner(typeName, "Map"), 1)) + ">"
					}
					return func() string {
						if genericInner(typeName, "Set") != "" {
							return "Set<" + selfhost_compiler_ts_tsType(genericInner(typeName, "Set")) + ">"
						}
						return selfhost_compiler_ts_tsNamedType(typeName)
					}()
				}()
			}()
		}()
	}()
}

func selfhost_compiler_ts_tsNamedType(typeName string) string {
	open := strings.Index(typeName, "[")
	return func() string {
		if open < 0 {
			return mangleIdent(typeName)
		}
		return mangleIdent(func() string { runes := []rune(typeName); return string(runes[0:open]) }()) + "<" + selfhost_compiler_ts_emitTSTypeArgs(func() string { runes := []rune(typeName); return string(runes[open+1 : len([]rune(typeName))-1]) }()) + ">"
	}()
}

func selfhost_compiler_ts_emitTSTypeArgs(args string) string {
	return selfhost_compiler_ts_emitTSTypeArgList(func() []string { parts := strings.Split(args, ","); return parts }(), 0, "")
}

func selfhost_compiler_ts_emitTSTypeArgList(args []string, index int, out string) string {
	return func() string {
		if index >= len(args) {
			return out
		}
		return selfhost_compiler_ts_emitTSTypeArgList(args, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_ts_tsType(strings.TrimSpace(args[index])))
	}()
}

func selfhost_compiler_ts_tsPropertyName(name string) string {
	return name
}

func selfhost_compiler_ts_joinStrings(values []string, index int, out string) string {
	return func() string {
		if index >= len(values) {
			return out
		}
		return selfhost_compiler_ts_joinStrings(values, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+values[index])
	}()
}

func generateDeclarations(file IRFile) string {
	out := selfhost_compiler_dts_dtsPreamble()
	out = out + selfhost_compiler_dts_dtsStructs(file.structs, 0)
	out = out + selfhost_compiler_dts_dtsEnums(file.enums, 0)
	out = out + selfhost_compiler_dts_dtsConsts(file.constants, 0)
	out = out + selfhost_compiler_dts_dtsFunctions(file.functions, 0)
	return out + selfhost_compiler_dts_dtsExports(file)
}

func selfhost_compiler_dts_dtsPreamble() string {
	return "type RuneResult<T, E> = { ok: true; value: T } | { ok: false; error: E };\ntype RuneError = { code: number; message: string; cause: RuneError | null };\ntype RuneIter<T> = { next: () => [T, boolean] };\ntype RuneFileStat = { size: number; isFile: boolean; isDirectory: boolean };\ntype RuneTCPConnection = { socket: unknown };\ntype RuneTCPListener = { server: unknown; address: string };\ndeclare class RuneStringBuffer {}\ndeclare class RuneBuffer {}\ndeclare class RuneReader {}\ndeclare class RuneWriter {}\n\n"
}

func selfhost_compiler_dts_dtsStructs(structs []IRStructType, index int) string {
	return func() string {
		if index >= len(structs) {
			return ""
		}
		return func() string {
			if structs[index].private {
				return selfhost_compiler_dts_dtsStructs(structs, index+1)
			}
			return selfhost_compiler_dts_dtsStruct(structs[index]) + selfhost_compiler_dts_dtsStructs(structs, index+1)
		}()
	}()
}

func selfhost_compiler_dts_dtsStruct(typeDecl IRStructType) string {
	generics := selfhost_compiler_dts_dtsGenerics(typeDecl.generics)
	out := "type " + mangleIdent(typeDecl.name) + generics + " = {\n"
	out = out + selfhost_compiler_dts_dtsStructFields(typeDecl.fields, 0)
	return out + "};\n" + "\n"
}

func selfhost_compiler_dts_dtsStructFields(fields []IRField, index int) string {
	return func() string {
		if index >= len(fields) {
			return ""
		}
		return func() string {
			if fields[index].private {
				return selfhost_compiler_dts_dtsStructFields(fields, index+1)
			}
			return indent(1) + selfhost_compiler_dts_dtsPropertyName(fields[index].name) + ": " + selfhost_compiler_dts_dtsType(fields[index].typeName) + ";\n" + selfhost_compiler_dts_dtsStructFields(fields, index+1)
		}()
	}()
}

func selfhost_compiler_dts_dtsEnums(enums []IREnumType, index int) string {
	return func() string {
		if index >= len(enums) {
			return ""
		}
		return func() string {
			if enums[index].private {
				return selfhost_compiler_dts_dtsEnums(enums, index+1)
			}
			return selfhost_compiler_dts_dtsEnum(enums[index]) + selfhost_compiler_dts_dtsEnums(enums, index+1)
		}()
	}()
}

func selfhost_compiler_dts_dtsEnum(enumDecl IREnumType) string {
	name := mangleIdent(enumDecl.name)
	generics := selfhost_compiler_dts_dtsGenerics(enumDecl.generics)
	shape := func() string {
		switch {
		case selfhost_compiler_dts_dtsEnumHasPayload(enumDecl.members, 0) == true:
			return "{ tag: number; payload: any[] }"
		case selfhost_compiler_dts_dtsEnumHasPayload(enumDecl.members, 0) == false:
			return "number"
		}
		return ""
	}()
	out := "type " + name + generics + " = " + shape + ";\n"
	out = out + "declare const " + name + ": {\n"
	out = out + selfhost_compiler_dts_dtsEnumMembers(enumDecl.members, 0)
	return out + "};\n" + "\n"
}

func selfhost_compiler_dts_dtsEnumHasPayload(members []IREnumMember, index int) bool {
	return func() bool {
		if index >= len(members) {
			return false
		}
		return func() bool {
			if len(members[index].params) > 0 {
				return true
			}
			return selfhost_compiler_dts_dtsEnumHasPayload(members, index+1)
		}()
	}()
}

func selfhost_compiler_dts_dtsEnumMembers(members []IREnumMember, index int) string {
	return func() string {
		if index >= len(members) {
			return ""
		}
		return selfhost_compiler_dts_dtsEnumMembersAt(members, index)
	}()
}

func selfhost_compiler_dts_dtsEnumMembersAt(members []IREnumMember, index int) string {
	return func() string {
		if members[index].private {
			return selfhost_compiler_dts_dtsEnumMembers(members, index+1)
		}
		return indent(1) + "readonly " + selfhost_compiler_dts_dtsPropertyName(members[index].name) + ": " + enumValue(members[index], index) + ";\n" + selfhost_compiler_dts_dtsEnumMembers(members, index+1)
	}()
}

func selfhost_compiler_dts_dtsConsts(constants []IRConst, index int) string {
	return func() string {
		if index >= len(constants) {
			return ""
		}
		return selfhost_compiler_dts_dtsConstsAt(constants, index)
	}()
}

func selfhost_compiler_dts_dtsConstsAt(constants []IRConst, index int) string {
	return func() string {
		if constants[index].private {
			return selfhost_compiler_dts_dtsConsts(constants, index+1)
		}
		return "declare const " + mangleIdent(constants[index].name) + ": " + selfhost_compiler_dts_dtsType(constants[index].typeName) + ";\n" + selfhost_compiler_dts_dtsConsts(constants, index+1)
	}()
}

func selfhost_compiler_dts_dtsFunctions(functions []IRFunction, index int) string {
	return func() string {
		if index >= len(functions) {
			return ""
		}
		return selfhost_compiler_dts_dtsFunctionAt(functions, index) + selfhost_compiler_dts_dtsFunctions(functions, index+1)
	}()
}

func selfhost_compiler_dts_dtsFunctionAt(functions []IRFunction, index int) string {
	return func() string {
		if functions[index].macro || functions[index].private {
			return ""
		}
		return selfhost_compiler_dts_dtsFunction(functions[index])
	}()
}

func selfhost_compiler_dts_dtsFunction(fn IRFunction) string {
	returnType := selfhost_compiler_dts_dtsType(fn.returnType)
	wrapped := func() string {
		switch {
		case fn.routine == true:
			return "Promise<" + returnType + ">"
		case fn.routine == false:
			return returnType
		}
		return ""
	}()
	return "declare function " + mangleIdent(fn.name) + selfhost_compiler_dts_dtsGenerics(fn.generics) + "(" + selfhost_compiler_dts_dtsParams(fn.params, 0, "") + "): " + wrapped + ";\n"
}

func selfhost_compiler_dts_dtsGenerics(generics []string) string {
	return func() string {
		if len(generics) == 0 {
			return ""
		}
		return "<" + selfhost_compiler_dts_dtsGenericNames(generics, 0, "") + ">"
	}()
}

func selfhost_compiler_dts_dtsGenericNames(generics []string, index int, out string) string {
	return func() string {
		if index >= len(generics) {
			return out
		}
		return selfhost_compiler_dts_dtsGenericNames(generics, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(generics[index])+" extends unknown")
	}()
}

func selfhost_compiler_dts_dtsParams(params []IRParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_dts_dtsParams(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+mangleIdent(params[index].name)+": "+selfhost_compiler_dts_dtsType(params[index].typeName))
	}()
}

func selfhost_compiler_dts_dtsType(typeName string) string {
	return selfhost_compiler_dts_dtsTypeRef(selfhost_compiler_dts_parseDtsTypeRef(typeName))
}

func selfhost_compiler_dts_parseDtsTypeRef(typeName string) ParsedTypeRef {
	return selfhost_compiler_dts_dtsTypeRefPostfix(lex(typeName), 0, selfhost_compiler_dts_dtsTypeRoot(typeName))
}

func selfhost_compiler_dts_dtsTypeRoot(typeName string) ParsedTypeRef {
	first := selfhost_compiler_dts_dtsFirstTypeToken(lex(typeName), 0)
	return func() ParsedTypeRef {
		switch {
		case first.kind == TokenKind_BitAnd:
			return namedParsedTypeRef(selfhost_compiler_dts_dtsTypeToken("Dynamic"))
		case first.kind == TokenKind_At:
			return selfhost_compiler_dts_dtsQualifiedTypeRef(lex(typeName), first)
		case first.kind == TokenKind_LParen:
			return selfhost_compiler_dts_dtsParenTypeRef(typeName, first)
		case first.kind == TokenKind_Ident:
			return namedParsedTypeRef(first)
		default:
			return namedParsedTypeRef(selfhost_compiler_dts_dtsTypeToken(typeName))
		}
	}()
}

func selfhost_compiler_dts_dtsFirstTypeToken(tokens []Token, index int) Token {
	return func() Token {
		if index >= len(tokens) {
			return selfhost_compiler_dts_dtsTypeToken("Dynamic")
		}
		return func() Token {
			if tokens[index].kind == TokenKind_EOF {
				return selfhost_compiler_dts_dtsTypeToken("Dynamic")
			}
			return tokens[index]
		}()
	}()
}

func selfhost_compiler_dts_dtsQualifiedTypeRef(tokens []Token, at Token) ParsedTypeRef {
	module := selfhost_compiler_dts_dtsTokenAt(tokens, 1)
	name := selfhost_compiler_dts_dtsTokenAt(tokens, 3)
	return func() ParsedTypeRef {
		if module.kind == TokenKind_Ident && name.kind == TokenKind_Ident {
			return qualifiedParsedTypeRef(module, name)
		}
		return namedParsedTypeRef(selfhost_compiler_dts_dtsTypeToken("Dynamic"))
	}()
}

func selfhost_compiler_dts_dtsParenTypeRef(typeName string, open Token) ParsedTypeRef {
	close := strings.LastIndex(typeName, ")")
	inner := func() string {
		if close > 0 {
			return func() string { runes := []rune(typeName); return string(runes[1:close]) }()
		}
		return ""
	}()
	parts := selfhost_compiler_dts_dtsSplitTopLevelTypeArgs(inner)
	return func() ParsedTypeRef {
		if len(parts) == 1 {
			return groupedTypeRef(selfhost_compiler_dts_parseDtsTypeRef(parts[0]), open)
		}
		return tupleTypeRef(selfhost_compiler_dts_dtsTupleParamsFromStrings(parts, 0, []ParsedTypeParam{}), open)
	}()
}

func selfhost_compiler_dts_dtsTypeToken(typeName string) Token {
	return Token{kind: TokenKind_Ident, lexeme: typeName, offset: 0, line: 0, column: 0}
}

func selfhost_compiler_dts_dtsTypeRefPostfix(tokens []Token, index int, typeRef ParsedTypeRef) ParsedTypeRef {
	return func() ParsedTypeRef {
		if index >= len(tokens) {
			return typeRef
		}
		return selfhost_compiler_dts_dtsTypeRefPostfixStep(tokens, index, typeRef)
	}()
}

func selfhost_compiler_dts_dtsTypeRefPostfixStep(tokens []Token, index int, typeRef ParsedTypeRef) ParsedTypeRef {
	token := tokens[index]
	return func() ParsedTypeRef {
		switch {
		case token.kind == TokenKind_LBracket:
			return selfhost_compiler_dts_dtsTypeRefPostfix(tokens, selfhost_compiler_dts_dtsSkipBracket(tokens, index+1, 1), selfhost_compiler_dts_dtsApplyTypeArgs(typeRef, selfhost_compiler_dts_dtsBracketArgs(tokens, index)))
		case token.kind == TokenKind_Question:
			return selfhost_compiler_dts_dtsTypeRefPostfix(tokens, index+1, nullableTypeRef(typeRef))
		default:
			return selfhost_compiler_dts_dtsTypeRefPostfix(tokens, index+1, typeRef)
		}
	}()
}

func selfhost_compiler_dts_dtsApplyTypeArgs(typeRef ParsedTypeRef, args []ParsedTypeRef) ParsedTypeRef {
	return typeRefWithArgs(typeRef, args)
}

func selfhost_compiler_dts_dtsBracketArgs(tokens []Token, open int) []ParsedTypeRef {
	close := selfhost_compiler_dts_dtsBracketClose(tokens, open+1, 1)
	inner := selfhost_compiler_dts_dtsTokenRangeText(tokens, open+1, close, "")
	parts := selfhost_compiler_dts_dtsSplitTopLevelTypeArgs(inner)
	return selfhost_compiler_dts_dtsTypeRefsFromStrings(parts, 0, []ParsedTypeRef{})
}

func selfhost_compiler_dts_dtsBracketClose(tokens []Token, index int, depth int) int {
	return func() int {
		if index >= len(tokens) {
			return index
		}
		return func() int {
			if tokens[index].kind == TokenKind_LBracket {
				return selfhost_compiler_dts_dtsBracketClose(tokens, index+1, depth+1)
			}
			return func() int {
				if tokens[index].kind == TokenKind_RBracket {
					return func() int {
						if depth == 1 {
							return index
						}
						return selfhost_compiler_dts_dtsBracketClose(tokens, index+1, depth-1)
					}()
				}
				return selfhost_compiler_dts_dtsBracketClose(tokens, index+1, depth)
			}()
		}()
	}()
}

func selfhost_compiler_dts_dtsSkipBracket(tokens []Token, index int, depth int) int {
	return func() int {
		if index >= len(tokens) {
			return index
		}
		return func() int {
			if tokens[index].kind == TokenKind_LBracket {
				return selfhost_compiler_dts_dtsSkipBracket(tokens, index+1, depth+1)
			}
			return func() int {
				if tokens[index].kind == TokenKind_RBracket {
					return func() int {
						if depth == 1 {
							return index + 1
						}
						return selfhost_compiler_dts_dtsSkipBracket(tokens, index+1, depth-1)
					}()
				}
				return selfhost_compiler_dts_dtsSkipBracket(tokens, index+1, depth)
			}()
		}()
	}()
}

func selfhost_compiler_dts_dtsTokenAt(tokens []Token, index int) Token {
	return func() Token {
		if index >= len(tokens) {
			return selfhost_compiler_dts_dtsTypeToken("Dynamic")
		}
		return tokens[index]
	}()
}

func selfhost_compiler_dts_dtsTokenRangeText(tokens []Token, index int, endExclusive int, out string) string {
	return func() string {
		if index >= endExclusive || index >= len(tokens) {
			return out
		}
		return selfhost_compiler_dts_dtsTokenRangeText(tokens, index+1, endExclusive, out+tokens[index].lexeme)
	}()
}

func selfhost_compiler_dts_dtsSplitTopLevelTypeArgs(source string) []string {
	return selfhost_compiler_dts_dtsSplitTopLevelTypeArgsLoop(source, 0, 0, 0, []string{})
}

func selfhost_compiler_dts_dtsSplitTopLevelTypeArgsLoop(source string, index int, start int, depth int, out []string) []string {
	return func() []string {
		if index >= len([]rune(source)) {
			return selfhost_compiler_dts_dtsPushTypeArg(out, strings.TrimSpace((func() string { runes := []rune(source); return string(runes[start:len([]rune(source))]) }())))
		}
		return selfhost_compiler_dts_dtsSplitTopLevelTypeArgsStep(source, index, start, depth, out)
	}()
}

func selfhost_compiler_dts_dtsSplitTopLevelTypeArgsStep(source string, index int, start int, depth int, out []string) []string {
	ch := []rune(source)[index]
	return func() []string {
		switch {
		case (ch == '[') || (ch == '('):
			return selfhost_compiler_dts_dtsSplitTopLevelTypeArgsLoop(source, index+1, start, depth+1, out)
		case (ch == ']') || (ch == ')'):
			return selfhost_compiler_dts_dtsSplitTopLevelTypeArgsLoop(source, index+1, start, depth-1, out)
		case ch == ',':
			return func() []string {
				if depth == 0 {
					return selfhost_compiler_dts_dtsSplitTopLevelTypeArgsLoop(source, index+1, index+1, depth, selfhost_compiler_dts_dtsPushTypeArg(out, strings.TrimSpace((func() string { runes := []rune(source); return string(runes[start:index]) }()))))
				}
				return selfhost_compiler_dts_dtsSplitTopLevelTypeArgsLoop(source, index+1, start, depth, out)
			}()
		default:
			return selfhost_compiler_dts_dtsSplitTopLevelTypeArgsLoop(source, index+1, start, depth, out)
		}
	}()
}

func selfhost_compiler_dts_dtsPushTypeArg(out []string, item string) []string {
	if item != "" {
		func() int { out = append(out, item); return len(out) }()
	}
	return out
}

func selfhost_compiler_dts_dtsTypeRefsFromStrings(parts []string, index int, out []ParsedTypeRef) []ParsedTypeRef {
	return func() []ParsedTypeRef {
		if index >= len(parts) {
			return out
		}
		return selfhost_compiler_dts_dtsTypeRefsFromStringsPush(parts, index, out)
	}()
}

func selfhost_compiler_dts_dtsTypeRefsFromStringsPush(parts []string, index int, out []ParsedTypeRef) []ParsedTypeRef {
	out = append(out, selfhost_compiler_dts_parseDtsTypeRef(parts[index]))
	return selfhost_compiler_dts_dtsTypeRefsFromStrings(parts, index+1, out)
}

func selfhost_compiler_dts_dtsTupleParamsFromStrings(parts []string, index int, out []ParsedTypeParam) []ParsedTypeParam {
	return func() []ParsedTypeParam {
		if index >= len(parts) {
			return out
		}
		return selfhost_compiler_dts_dtsTupleParamsFromStringsPush(parts, index, out)
	}()
}

func selfhost_compiler_dts_dtsTupleParamsFromStringsPush(parts []string, index int, out []ParsedTypeParam) []ParsedTypeParam {
	param := emptyParsedTypeParam()
	param.name = ""
	param.optional = false
	param.typeRef = selfhost_compiler_dts_parseDtsTypeRef(parts[index])
	out = append(out, param)
	return selfhost_compiler_dts_dtsTupleParamsFromStrings(parts, index+1, out)
}

func selfhost_compiler_dts_clearNullableTypeRef(typeRef ParsedTypeRef) ParsedTypeRef {
	return ParsedTypeRef{kind: typeRef.kind, name: typeRef.name, module: typeRef.module, nullable: false, args: typeRef.args, params: typeRef.params, returnTypes: typeRef.returnTypes, line: typeRef.line, column: typeRef.column}
}

func selfhost_compiler_dts_dtsTypeRef(typeRef ParsedTypeRef) string {
	return func() string {
		switch {
		case typeRef.nullable == true:
			return selfhost_compiler_dts_dtsNullableTypeRef(typeRef)
		case typeRef.nullable == false:
			return selfhost_compiler_dts_dtsPlainTypeRef(typeRef)
		}
		return ""
	}()
}

func selfhost_compiler_dts_dtsNullableTypeRef(typeRef ParsedTypeRef) string {
	return selfhost_compiler_dts_dtsTypeRef(selfhost_compiler_dts_clearNullableTypeRef(typeRef)) + " | null"
}

func selfhost_compiler_dts_dtsPlainTypeRef(typeRef ParsedTypeRef) string {
	return func() string {
		switch {
		case typeRef.kind == TypeRefKind_Name:
			return selfhost_compiler_dts_dtsNamedTypeRef(typeRef)
		case typeRef.kind == TypeRefKind_Group:
			return selfhost_compiler_dts_dtsGroupedTypeRef(typeRef)
		case typeRef.kind == TypeRefKind_Tuple:
			return selfhost_compiler_dts_dtsTupleTypeRef(typeRef)
		case typeRef.kind == TypeRefKind_Function:
			return selfhost_compiler_dts_dtsFunctionTypeRef(typeRef)
		default:
			return "any"
		}
	}()
}

func selfhost_compiler_dts_dtsNamedTypeRef(typeRef ParsedTypeRef) string {
	typeName := selfhost_compiler_dts_dtsQualifiedTypeName(typeRef)
	return func() string {
		switch {
		case (typeName == "") || (typeName == "Void"):
			return "void"
		case (typeName == "Int") || (typeName == "Int4") || (typeName == "Int8") || (typeName == "Int16") || (typeName == "UInt") || (typeName == "UInt8") || (typeName == "UInt16") || (typeName == "Byte") || (typeName == "Double") || (typeName == "Float"):
			return "number"
		case (typeName == "BigInt") || (typeName == "Int64") || (typeName == "UInt64"):
			return "bigint"
		case (typeName == "String") || (typeName == "Char"):
			return "string"
		case typeName == "Bool":
			return "boolean"
		case typeName == "Null":
			return "null"
		case typeName == "Object":
			return "object"
		case typeName == "Bytes":
			return "DataView"
		case typeName == "Buffer":
			return "RuneBuffer"
		case typeName == "Reader":
			return "RuneReader"
		case typeName == "Writer":
			return "RuneWriter"
		case typeName == "StringBuffer":
			return "RuneStringBuffer"
		case typeName == "FileStat":
			return "RuneFileStat"
		case typeName == "TCPConnection":
			return "RuneTCPConnection"
		case typeName == "TCPListener":
			return "RuneTCPListener"
		case (typeName == "Data") || (typeName == "@io.Data"):
			return "Uint8Array"
		case typeName == "Error":
			return "RuneError"
		case typeName == "Never":
			return "never"
		case typeName == "Symbol":
			return "symbol"
		case typeName == "Regex":
			return "RegExp"
		case typeName == "HTMLElement":
			return "HTMLElement"
		case typeName == "WebComponent":
			return "CustomElementConstructor"
		case (typeName == "Dynamic") || (typeName == "Unknown"):
			return "any"
		default:
			return selfhost_compiler_dts_dtsStructuredNamedTypeRef(typeRef)
		}
	}()
}

func selfhost_compiler_dts_dtsQualifiedTypeName(typeRef ParsedTypeRef) string {
	return func() string {
		if typeRef.module == "" {
			return typeRef.name
		}
		return "@" + typeRef.module + "." + typeRef.name
	}()
}

func selfhost_compiler_dts_dtsStructuredNamedTypeRef(typeRef ParsedTypeRef) string {
	return func() string {
		switch {
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "Array":
			return selfhost_compiler_dts_dtsArrayTypeRef(typeRef)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "Result":
			return selfhost_compiler_dts_dtsGenericTypeRef("RuneResult", typeRef.args)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "Task":
			return selfhost_compiler_dts_dtsGenericTypeRef("Promise", typeRef.args)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "Iter":
			return selfhost_compiler_dts_dtsGenericTypeRef("RuneIter", typeRef.args)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "ReadonlyArray":
			return selfhost_compiler_dts_dtsGenericTypeRef("ReadonlyArray", typeRef.args)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "Tuple":
			return selfhost_compiler_dts_dtsTupleArgs(typeRef.args, false)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "ReadonlyTuple":
			return selfhost_compiler_dts_dtsTupleArgs(typeRef.args, true)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "Map":
			return selfhost_compiler_dts_dtsGenericTypeRef("Map", typeRef.args)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "Set":
			return selfhost_compiler_dts_dtsGenericTypeRef("Set", typeRef.args)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "WeakMap":
			return selfhost_compiler_dts_dtsGenericTypeRef("WeakMap", typeRef.args)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "WeakSet":
			return selfhost_compiler_dts_dtsGenericTypeRef("WeakSet", typeRef.args)
		case selfhost_compiler_dts_dtsQualifiedTypeName(typeRef) == "Record":
			return selfhost_compiler_dts_dtsGenericTypeRef("Record", typeRef.args)
		default:
			return selfhost_compiler_dts_dtsNamedType(typeRef)
		}
	}()
}

func selfhost_compiler_dts_dtsArrayTypeRef(typeRef ParsedTypeRef) string {
	return func() string {
		if len(typeRef.args) == 0 {
			return "any[]"
		}
		return selfhost_compiler_dts_dtsTypeRef(typeRef.args[0]) + "[]"
	}()
}

func selfhost_compiler_dts_dtsGenericTypeRef(name string, args []ParsedTypeRef) string {
	return name + "<" + selfhost_compiler_dts_dtsTypeRefs(args, 0, "") + ">"
}

func selfhost_compiler_dts_dtsNamedType(typeRef ParsedTypeRef) string {
	prefix := func() string {
		if typeRef.module == "" {
			return ""
		}
		return "@" + typeRef.module + "."
	}()
	return func() string {
		if len(typeRef.args) == 0 {
			return mangleIdent(prefix + typeRef.name)
		}
		return mangleIdent(prefix+typeRef.name) + "<" + selfhost_compiler_dts_dtsTypeRefs(typeRef.args, 0, "") + ">"
	}()
}

func selfhost_compiler_dts_dtsGroupedTypeRef(typeRef ParsedTypeRef) string {
	return func() string {
		if len(typeRef.args) == 0 {
			return "()"
		}
		return "(" + selfhost_compiler_dts_dtsTypeRef(typeRef.args[0]) + ")"
	}()
}

func selfhost_compiler_dts_dtsTupleTypeRef(typeRef ParsedTypeRef) string {
	return selfhost_compiler_dts_dtsTupleParams(typeRef.params, false)
}

func selfhost_compiler_dts_dtsTupleParams(params []ParsedTypeParam, readonly bool) string {
	items := selfhost_compiler_dts_dtsTypeParams(params, 0, "")
	return func() string {
		if readonly {
			return "readonly [" + items + "]"
		}
		return "[" + items + "]"
	}()
}

func selfhost_compiler_dts_dtsTupleArgs(args []ParsedTypeRef, readonly bool) string {
	items := selfhost_compiler_dts_dtsTypeRefs(args, 0, "")
	return func() string {
		if readonly {
			return "readonly [" + items + "]"
		}
		return "[" + items + "]"
	}()
}

func selfhost_compiler_dts_dtsFunctionTypeRef(typeRef ParsedTypeRef) string {
	ret := func() string {
		if len(typeRef.returnTypes) == 0 {
			return "void"
		}
		return selfhost_compiler_dts_dtsTypeRef(typeRef.returnTypes[0])
	}()
	return "(" + selfhost_compiler_dts_dtsTypeParams(typeRef.params, 0, "") + ") => " + ret
}

func selfhost_compiler_dts_dtsTypeRefs(refs []ParsedTypeRef, index int, out string) string {
	return func() string {
		if index >= len(refs) {
			return out
		}
		return selfhost_compiler_dts_dtsTypeRefs(refs, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_dts_dtsTypeRef(refs[index]))
	}()
}

func selfhost_compiler_dts_dtsTypeParams(params []ParsedTypeParam, index int, out string) string {
	return func() string {
		if index >= len(params) {
			return out
		}
		return selfhost_compiler_dts_dtsTypeParams(params, index+1, out+func() string {
			if index == 0 {
				return ""
			}
			return ", "
		}()+selfhost_compiler_dts_dtsTypeParam(params[index]))
	}()
}

func selfhost_compiler_dts_dtsTypeParam(param ParsedTypeParam) string {
	prefix := func() string {
		if param.name == "" {
			return ""
		}
		return mangleIdent(param.name) + func() string {
			if param.optional {
				return "?: "
			}
			return ": "
		}()
	}()
	return prefix + selfhost_compiler_dts_dtsTypeRef(param.typeRef)
}

func selfhost_compiler_dts_dtsExports(file IRFile) string {
	return func() string {
		if len(file.structs)+len(file.enums)+len(file.constants)+len(file.functions) == 0 {
			return ""
		}
		return "\n" + selfhost_compiler_dts_dtsStructExports(file.structs, 0) + selfhost_compiler_dts_dtsEnumExports(file.enums, 0) + selfhost_compiler_dts_dtsConstExports(file.constants, 0) + selfhost_compiler_dts_dtsFunctionExports(file.functions, 0)
	}()
}

func selfhost_compiler_dts_dtsStructExports(structs []IRStructType, index int) string {
	return func() string {
		if index >= len(structs) {
			return ""
		}
		return func() string {
			if structs[index].private {
				return selfhost_compiler_dts_dtsStructExports(structs, index+1)
			}
			return selfhost_compiler_dts_dtsExportTypeAlias(structs[index].name) + selfhost_compiler_dts_dtsStructExports(structs, index+1)
		}()
	}()
}

func selfhost_compiler_dts_dtsEnumExports(enums []IREnumType, index int) string {
	return func() string {
		if index >= len(enums) {
			return ""
		}
		return func() string {
			if enums[index].private {
				return selfhost_compiler_dts_dtsEnumExports(enums, index+1)
			}
			return selfhost_compiler_dts_dtsExportTypeAlias(enums[index].name) + selfhost_compiler_dts_dtsExportValueAlias(enums[index].name) + selfhost_compiler_dts_dtsEnumExports(enums, index+1)
		}()
	}()
}

func selfhost_compiler_dts_dtsConstExports(constants []IRConst, index int) string {
	return func() string {
		if index >= len(constants) {
			return ""
		}
		return func() string {
			if constants[index].private {
				return selfhost_compiler_dts_dtsConstExports(constants, index+1)
			}
			return selfhost_compiler_dts_dtsExportValueAlias(constants[index].name) + selfhost_compiler_dts_dtsConstExports(constants, index+1)
		}()
	}()
}

func selfhost_compiler_dts_dtsFunctionExports(functions []IRFunction, index int) string {
	return func() string {
		if index >= len(functions) {
			return ""
		}
		return func() string {
			if functions[index].macro || functions[index].private {
				return selfhost_compiler_dts_dtsFunctionExports(functions, index+1)
			}
			return selfhost_compiler_dts_dtsExportValueAlias(functions[index].name) + selfhost_compiler_dts_dtsFunctionExports(functions, index+1)
		}()
	}()
}

func selfhost_compiler_dts_dtsExportTypeAlias(name string) string {
	return func() string {
		if selfhost_compiler_dts_dtsCanUseBareProperty(name) {
			return "export type " + name + " = " + mangleIdent(name) + ";\n"
		}
		return "export type { " + mangleIdent(name) + " as " + selfhost_compiler_dts_dtsExportName(name) + " };\n"
	}()
}

func selfhost_compiler_dts_dtsExportValueAlias(name string) string {
	return func() string {
		if selfhost_compiler_dts_dtsCanUseBareProperty(name) {
			return "export declare const " + name + ": typeof " + mangleIdent(name) + ";\n"
		}
		return "export { " + mangleIdent(name) + " as " + selfhost_compiler_dts_dtsExportName(name) + " };\n"
	}()
}

func selfhost_compiler_dts_dtsPropertyName(name string) string {
	return func() string {
		if selfhost_compiler_dts_dtsCanUseBareProperty(name) {
			return name
		}
		return selfhost_compiler_dts_dtsQuoteName(name)
	}()
}

func selfhost_compiler_dts_dtsExportName(name string) string {
	return selfhost_compiler_dts_dtsPropertyName(name)
}

func selfhost_compiler_dts_dtsQuoteName(name string) string {
	return "\"" + selfhost_compiler_dts_dtsEscapeName(name, 0, "") + "\""
}

func selfhost_compiler_dts_dtsEscapeName(name string, index int, out string) string {
	return func() string {
		if index >= len([]rune(name)) {
			return out
		}
		return selfhost_compiler_dts_dtsEscapeName(name, index+1, out+selfhost_compiler_dts_dtsEscapeNameChar([]rune(name)[index]))
	}()
}

func selfhost_compiler_dts_dtsEscapeNameChar(ch rune) string {
	return func() string {
		switch {
		case ch == '\\':
			return "\\\\"
		case ch == '"':
			return "\\\""
		case ch == '\n':
			return "\\n"
		case ch == '\r':
			return "\\r"
		case ch == '\t':
			return "\\t"
		default:
			return string(ch)
		}
	}()
}

func selfhost_compiler_dts_dtsCanUseBareProperty(name string) bool {
	return name != "" && selfhost_compiler_dts_dtsReservedName(name) == false && selfhost_compiler_dts_dtsSafeIdent(name, 0, true)
}

func selfhost_compiler_dts_dtsSafeIdent(name string, index int, first bool) bool {
	return func() bool {
		if index >= len([]rune(name)) {
			return true
		}
		return func() bool {
			if selfhost_compiler_dts_dtsSafeIdentChar([]rune(name)[index], first) {
				return selfhost_compiler_dts_dtsSafeIdent(name, index+1, false)
			}
			return false
		}()
	}()
}

func selfhost_compiler_dts_dtsSafeIdentChar(ch rune, first bool) bool {
	return func() bool {
		switch {
		case (ch == '_') || (ch == '$') || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z'):
			return true
		case (ch >= '0' && ch <= '9'):
			return first == false
		default:
			return false
		}
	}()
}

func selfhost_compiler_dts_dtsReservedName(name string) bool {
	return func() bool {
		switch {
		case (name == "await") || (name == "break") || (name == "case") || (name == "catch") || (name == "class") || (name == "const") || (name == "continue") || (name == "debugger") || (name == "default") || (name == "delete") || (name == "do") || (name == "else") || (name == "enum") || (name == "export") || (name == "extends") || (name == "false") || (name == "finally") || (name == "for") || (name == "function") || (name == "if") || (name == "implements") || (name == "import") || (name == "in") || (name == "instanceof") || (name == "interface") || (name == "let") || (name == "new") || (name == "null") || (name == "package") || (name == "private") || (name == "protected") || (name == "public") || (name == "return") || (name == "static") || (name == "super") || (name == "switch") || (name == "this") || (name == "throw") || (name == "true") || (name == "try") || (name == "typeof") || (name == "var") || (name == "void") || (name == "while") || (name == "with") || (name == "yield"):
			return true
		default:
			return false
		}
	}()
}

func compileTypeScript(source string) CompileResult {
	return compile(source, "ts")
}

func compileGo(source string) CompileResult {
	return compile(source, "go")
}

func compileMoonBit(source string) CompileResult {
	return compile(source, "mbt")
}

func compileDeclarations(source string) CompileResult {
	return compile(source, "dts")
}

func checkSource(source string) CompileResult {
	return checkSourceWithPath(source, "")
}

func checkSourceWithPath(source string, sourcePath string) CompileResult {
	return selfhost_compiler_compiler_checkFile(selfhost_compiler_compiler_lowerCompilerSourceWithPath(source, sourcePath))
}

func compile(source string, target string) CompileResult {
	file := selfhost_compiler_compiler_lowerCompilerSource(source)
	return selfhost_compiler_compiler_compileFile(file, target)
}

func compileTypeScriptFiles(files []SourceFile) CompileResult {
	return compileFiles(files, "ts")
}

func compileGoFiles(files []SourceFile) CompileResult {
	return compileFiles(files, "go")
}

func compileMoonBitFiles(files []SourceFile) CompileResult {
	return compileFiles(files, "mbt")
}

func compileDeclarationsFiles(files []SourceFile) CompileResult {
	return compileFiles(files, "dts")
}

func compileFiles(files []SourceFile, target string) CompileResult {
	file := selfhost_compiler_compiler_lowerFiles(files)
	return selfhost_compiler_compiler_compileFile(file, target)
}

func discoverSources(root string) runeTask[runeResult[[]SourceFile, *runeError]] {
	return runeGo(func() runeResult[[]SourceFile, *runeError] {
		result4 := runeAwait(runeFsReadFileText(root))
		if !result4.ok {
			return runeErr[[]SourceFile, *runeError](result4.err)
		}
		source := result4.value
		return runeOk[[]SourceFile, *runeError](selfhost_compiler_compiler_discoverSourceGraph([]SourceFile{SourceFile{path: root, source: source}}))
	})
}

func hostBridgeSources(root string, files []SourceFile) []SourceFile {
	return selfhost_compiler_compiler_discoverSourceGraph(files)
}

func selfhost_compiler_compiler_discoverSourceGraph(files []SourceFile) []SourceFile {
	return selfhost_compiler_compiler_discoverMergeSourceFiles([]SourceFile{}, files, 0)
}

func getSelfhostSources(root string) runeTask[runeResult[SelfhostSources, *runeError]] {
	return runeGo(func() runeResult[SelfhostSources, *runeError] {
		result5 := runeAwait(discoverSources(root))
		if !result5.ok {
			return runeErr[SelfhostSources, *runeError](result5.err)
		}
		files := result5.value
		return runeOk[SelfhostSources, *runeError](SelfhostSources{files: files})
	})
}

func __discoverSourcesPath(root string) runeTask[runeResult[[]SourceFile, *runeError]] {
	return runeGo(func() runeResult[[]SourceFile, *runeError] {
		result6 := runeAwait(discoverSources(root))
		if !result6.ok {
			return runeErr[[]SourceFile, *runeError](result6.err)
		}
		files := result6.value
		return runeOk[[]SourceFile, *runeError](files)
	})
}

func selfhost_compiler_compiler_discoverSourceImportPaths(file SourceFile) []string {
	parsed := selfhost_compiler_compiler_expandCompilerMacros(parse(file.source))
	return selfhost_compiler_compiler_appendDiscoverImportPaths([]string{}, file.path, parsed.imports, 0)
}

func selfhost_compiler_compiler_appendDiscoverImportPaths(pending []string, basePath string, imports []ParsedImport, index int) []string {
	return func() []string {
		if index >= len(imports) {
			return pending
		}
		return selfhost_compiler_compiler_appendDiscoverImportPath(pending, basePath, imports, index)
	}()
}

func selfhost_compiler_compiler_appendDiscoverImportPath(pending []string, basePath string, imports []ParsedImport, index int) []string {
	importDecl := imports[index]
	skip := importDecl.go_ || importDecl.module
	return func() []string {
		switch {
		case skip == true:
			return selfhost_compiler_compiler_appendDiscoverImportPaths(pending, basePath, imports, index+1)
		default:
			return selfhost_compiler_compiler_appendDiscoverImportPaths(func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, pending...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_resolveCompilerImportPath(basePath, importDecl.path))
				return __rune_spread_out
			}(), basePath, imports, index+1)
		}
	}()
}

func selfhost_compiler_compiler_discoverMergeSourceFiles(left []SourceFile, right []SourceFile, index int) []SourceFile {
	return func() []SourceFile {
		if index >= len(right) {
			return left
		}
		return selfhost_compiler_compiler_discoverMergeSourceFiles(func() []SourceFile {
			if selfhost_compiler_compiler_compilerContainsSourceFile(left, right[index].path, 0) {
				return left
			}
			return func() []SourceFile {
				__rune_spread_out := []SourceFile{}
				__rune_spread_out = append(__rune_spread_out, left...)
				__rune_spread_out = append(__rune_spread_out, right[index])
				return __rune_spread_out
			}()
		}(), right, index+1)
	}()
}

func selfhost_compiler_compiler_compilerContainsSourceFile(files []SourceFile, path string, index int) bool {
	return func() bool {
		if index >= len(files) {
			return false
		}
		return selfhost_compiler_compiler_compilerPathNormalize(files[index].path) == selfhost_compiler_compiler_compilerPathNormalize(path) || selfhost_compiler_compiler_compilerContainsSourceFile(files, path, index+1)
	}()
}

func selfhost_compiler_compiler_compileFile(file IRFile, target string) CompileResult {
	return func() CompileResult {
		if len(file.errors) > 0 {
			return selfhost_compiler_compiler_compileResult(false, "", selfhost_compiler_compiler_parseErrorMessages(file.errors))
		}
		return selfhost_compiler_compiler_compileCheckedFile(file, target)
	}()
}

func selfhost_compiler_compiler_checkFile(file IRFile) CompileResult {
	return func() CompileResult {
		if len(file.errors) > 0 {
			return selfhost_compiler_compiler_compileResult(false, "", selfhost_compiler_compiler_parseErrorMessages(file.errors))
		}
		return selfhost_compiler_compiler_checkCheckedFile(file)
	}()
}

func selfhost_compiler_compiler_checkCheckedFile(file IRFile) CompileResult {
	inferred := inferFile(file)
	errors := selfhost_compiler_compiler_checkFileErrors(inferred)
	return func() CompileResult {
		if len(errors) > 0 {
			return selfhost_compiler_compiler_compileResult(false, "", errors)
		}
		return selfhost_compiler_compiler_compileResult(true, "", []string{})
	}()
}

func selfhost_compiler_compiler_compileCheckedFile(file IRFile, target string) CompileResult {
	inferred := inferFile(file)
	errors := selfhost_compiler_compiler_checkFileErrors(inferred)
	targetErrors := selfhost_compiler_compiler_checkTargetFileErrors(inferred, target)
	allErrors := func() []string {
		__rune_spread_out := []string{}
		__rune_spread_out = append(__rune_spread_out, errors...)
		__rune_spread_out = append(__rune_spread_out, targetErrors...)
		return __rune_spread_out
	}()
	return func() CompileResult {
		if len(allErrors) > 0 {
			return selfhost_compiler_compiler_compileResult(false, "", allErrors)
		}
		return func() CompileResult {
			switch {
			case target == "ts":
				return selfhost_compiler_compiler_compileResult(true, generateTypeScript(inferred), []string{})
			case target == "go":
				return selfhost_compiler_compiler_compileResult(true, generateGo(inferred), []string{})
			case target == "mbt":
				return selfhost_compiler_compiler_compileResult(true, generateMoonBit(inferred), []string{})
			case target == "dts":
				return selfhost_compiler_compiler_compileResult(true, generateDeclarations(inferred), []string{})
			default:
				return selfhost_compiler_compiler_compileResult(false, "", selfhost_compiler_compiler_unsupportedTargetErrors(target))
			}
		}()
	}()
}

func selfhost_compiler_compiler_checkTargetFileErrors(file IRFile, target string) []string {
	return func() []string {
		switch {
		case target == "ts":
			return selfhost_compiler_compiler_checkTypeScriptTargetFileErrors(file)
		case target == "go":
			return selfhost_compiler_compiler_checkGoTargetFileErrors(file)
		case target == "mbt":
			return selfhost_compiler_compiler_checkMoonBitTargetFileErrors(file)
		default:
			return []string{}
		}
	}()
}

func selfhost_compiler_compiler_checkTypeScriptTargetFileErrors(file IRFile) []string {
	hasGoImports := selfhost_compiler_compiler_fileHasGoImports(file)
	return func() []string {
		switch {
		case hasGoImports == true:
			return []string{"TypeScript backend does not support Go package imports"}
		default:
			return func() []string {
				switch {
				case selfhost_compiler_compiler_fileUsesGoFFI(file) == true:
					return []string{"TypeScript backend does not support @go FFI"}
				default:
					return []string{}
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkGoTargetFileErrors(file IRFile) []string {
	hasTypeScriptImports := len(file.tsImports) > 0
	return func() []string {
		switch {
		case hasTypeScriptImports == true:
			return []string{"Go backend does not support TypeScript imports"}
		default:
			return []string{}
		}
	}()
}

func selfhost_compiler_compiler_checkMoonBitTargetFileErrors(file IRFile) []string {
	errors := []string{}
	hasTypeScriptImports := len(file.tsImports) > 0
	errors = selfhost_compiler_compiler_compilerAppendErrorIf(errors, hasTypeScriptImports, "MoonBit backend does not support TypeScript imports")
	hasGoImports := selfhost_compiler_compiler_fileHasGoImports(file)
	errors = selfhost_compiler_compiler_compilerAppendErrorIf(errors, hasGoImports, "MoonBit backend does not support Go package imports")
	hasGoFFI := hasGoImports == false && selfhost_compiler_compiler_fileUsesGoFFI(file)
	errors = selfhost_compiler_compiler_compilerAppendErrorIf(errors, hasGoFFI, "MoonBit backend does not support @go FFI")
	return errors
}

func selfhost_compiler_compiler_fileUsesGoFFI(file IRFile) bool {
	return fileUsesModuleCall(file, "go.stmt") || fileUsesModuleCall(file, "go.expr")
}

func selfhost_compiler_compiler_compilerAppendErrorIf(errors []string, condition bool, message string) []string {
	return func() []string {
		switch {
		case condition == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, message)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_fileHasGoImports(file IRFile) bool {
	return selfhost_compiler_compiler_fileHasGoImportsAt(file.imports, 0)
}

func selfhost_compiler_compiler_fileHasGoImportsAt(imports []IRImport, index int) bool {
	done := index >= len(imports)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return selfhost_compiler_compiler_fileHasGoImportAt(imports, index)
		}
	}()
}

func selfhost_compiler_compiler_fileHasGoImportAt(imports []IRImport, index int) bool {
	return func() bool {
		switch {
		case imports[index].go_ == true:
			return true
		default:
			return selfhost_compiler_compiler_fileHasGoImportsAt(imports, index+1)
		}
	}()
}

func selfhost_compiler_compiler_checkFileErrors(file IRFile) []string {
	callables := selfhost_compiler_compiler_compilerCallables(file)
	bindings := selfhost_compiler_compiler_compilerInitialBindings(file, callables)
	knownTypes := selfhost_compiler_compiler_compilerKnownTypes(file)
	errors := selfhost_compiler_compiler_checkDuplicateDeclarations(file, []string{})
	errors = selfhost_compiler_compiler_checkDeclarationTypes(file, knownTypes, errors)
	for _, constant := range file.constants {
		_ = constant
		errors = selfhost_compiler_compiler_checkConstantErrors(constant, file.structs, callables, errors, bindings)
	}
	for _, fn := range file.functions {
		_ = fn
		errors = selfhost_compiler_compiler_checkTopLevelFunctionErrors(fn, file.structs, callables, errors, bindings)
	}
	for _, typeDecl := range file.structs {
		_ = typeDecl
		func() {
			for _, method := range typeDecl.methods {
				_ = method
				errors = selfhost_compiler_compiler_checkMethodErrors(typeDecl.name, method, file.structs, callables, errors, bindings)
			}
		}()
	}
	for _, typeDecl := range file.enums {
		_ = typeDecl
		func() {
			for _, method := range typeDecl.methods {
				_ = method
				errors = selfhost_compiler_compiler_checkMethodErrors(typeDecl.name, method, file.structs, callables, errors, bindings)
			}
		}()
	}
	for _, testDecl := range file.tests {
		_ = testDecl
		errors = selfhost_compiler_compiler_checkExpr(testDecl.body, file.structs, callables, errors, bindings)
	}
	return errors
}

func selfhost_compiler_compiler_checkConstantErrors(constant IRConst, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExpr(constant.value, structs, callables, errors, bindings)
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(constant.value, structs, callables, bindings)
	mismatch := selfhost_compiler_compiler_compilerShouldCheckArgType(constant.typeName, actual) && selfhost_compiler_compiler_compilerTypesCompatible(constant.typeName, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, checked...)
				__rune_spread_out = append(__rune_spread_out, "constant "+constant.name+" has type "+actual+", expected "+constant.typeName)
				return __rune_spread_out
			}()
		default:
			return checked
		}
	}()
}

func selfhost_compiler_compiler_checkTopLevelFunctionErrors(fn IRFunction, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case fn.macro == true:
			return selfhost_compiler_compiler_checkMacroFunctionReturn(fn, structs, callables, errors, selfhost_compiler_compiler_compilerSourcePathBindings(fn.sourcePath, bindings))
		default:
			return selfhost_compiler_compiler_checkFunctionErrors(fn, structs, callables, errors, selfhost_compiler_compiler_compilerSourcePathBindings(fn.sourcePath, bindings))
		}
	}()
}

func selfhost_compiler_compiler_compilerSourcePathBindings(sourcePath string, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		switch {
		case sourcePath == "":
			return bindings
		default:
			return selfhost_compiler_compiler_addCompilerValueBinding(bindings, "__sourcePath", sourcePath)
		}
	}()
}

func selfhost_compiler_compiler_checkMacroFunctionReturn(fn IRFunction, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	expected := fn.returnType
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(fn.body, structs, callables, selfhost_compiler_compiler_compilerFunctionBindings(fn.params, bindings))
	shouldCheck := expected != "" && expected != "Dynamic" && actual != ""
	return func() []string {
		switch {
		case shouldCheck == true:
			return selfhost_compiler_compiler_checkFunctionReturnType(fn.name, expected, actual, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_compilerKnownTypes(file IRFile) []string {
	names := selfhost_compiler_compiler_compilerBuiltinTypes()
	for _, typeDecl := range file.structs {
		_ = typeDecl
		func() int { names = append(names, typeDecl.name); return len(names) }()
	}
	for _, typeDecl := range file.enums {
		_ = typeDecl
		func() int { names = append(names, typeDecl.name); return len(names) }()
	}
	return names
}

func selfhost_compiler_compiler_compilerBuiltinTypes() []string {
	return []string{"Int", "Int4", "Int8", "Int16", "Int64", "UInt", "UInt8", "UInt16", "UInt64", "Double", "Float", "Bool", "String", "Char", "BigInt", "Byte", "Bytes", "Object", "Dynamic", "Void", "Null", "Error", "Regex", "Symbol", "MacroContext", "HTMLElement", "WebComponent", "Array", "ReadonlyArray", "Tuple", "ReadonlyTuple", "Map", "Set", "Result"}
}

func selfhost_compiler_compiler_checkDuplicateDeclarations(file IRFile, errors []string) []string {
	next := selfhost_compiler_compiler_checkDuplicateStructTypeNames(file.structs, 0, errors)
	next = selfhost_compiler_compiler_checkDuplicateEnumTypeNames(file.enums, file.structs, 0, next)
	next = selfhost_compiler_compiler_checkDuplicateConstantNames(file.constants, 0, next)
	next = selfhost_compiler_compiler_checkDuplicateFunctionNames(file.functions, 0, next)
	for _, typeDecl := range file.structs {
		_ = typeDecl
		next = selfhost_compiler_compiler_checkDuplicateStructMembers(typeDecl, next)
	}
	for _, typeDecl := range file.enums {
		_ = typeDecl
		next = selfhost_compiler_compiler_checkDuplicateEnumMembers(typeDecl, next)
	}
	return next
}

func selfhost_compiler_compiler_checkDuplicateStructTypeNames(structs []IRStructType, index int, errors []string) []string {
	done := index >= len(structs)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateStructTypeName(structs, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateStructTypeName(structs []IRStructType, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerStructNameAppearsAfter(structs, structs[index].name, index+1)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate type \""+structs[index].name+"\"")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateStructTypeNames(structs, index+1, next)
}

func selfhost_compiler_compiler_checkDuplicateEnumTypeNames(enums []IREnumType, structs []IRStructType, index int, errors []string) []string {
	done := index >= len(enums)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateEnumTypeName(enums, structs, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateEnumTypeName(enums []IREnumType, structs []IRStructType, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerEnumNameAppearsAfter(enums, enums[index].name, index+1) || selfhost_compiler_compiler_compilerStructNameAppearsAfter(structs, enums[index].name, 0)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate type \""+enums[index].name+"\"")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateEnumTypeNames(enums, structs, index+1, next)
}

func selfhost_compiler_compiler_checkDuplicateFunctionNames(functions []IRFunction, index int, errors []string) []string {
	done := index >= len(functions)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateFunctionName(functions, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateFunctionName(functions []IRFunction, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerFunctionNameAppearsBefore(functions, functions[index].name, functions[index].macro, index-1)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate function \""+functions[index].name+"\"")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateFunctionNames(functions, index+1, next)
}

func selfhost_compiler_compiler_checkDuplicateConstantNames(constants []IRConst, index int, errors []string) []string {
	done := index >= len(constants)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateConstantName(constants, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateConstantName(constants []IRConst, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerConstNameAppearsBefore(constants, constants[index].name, index-1)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate declaration \""+constants[index].name+"\"")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateConstantNames(constants, index+1, next)
}

func selfhost_compiler_compiler_checkDuplicateStructMembers(typeDecl IRStructType, errors []string) []string {
	next := selfhost_compiler_compiler_checkDuplicateStructFields(typeDecl.fields, 0, errors)
	return selfhost_compiler_compiler_checkDuplicateStructMethods(typeDecl.name, typeDecl.methods, 0, next)
}

func selfhost_compiler_compiler_checkDuplicateStructFields(fields []IRField, index int, errors []string) []string {
	done := index >= len(fields)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateStructField(fields, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateStructField(fields []IRField, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerFieldNameAppearsBefore(fields, fields[index].name, index-1)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate field \""+fields[index].name+"\"")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateStructFields(fields, index+1, next)
}

func selfhost_compiler_compiler_checkDuplicateStructMethods(typeName string, methods []IRFunction, index int, errors []string) []string {
	done := index >= len(methods)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateStructMethod(typeName, methods, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateStructMethod(typeName string, methods []IRFunction, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerMethodNameAppearsBefore(methods, methods[index].name, index-1)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate method "+typeName+"."+methods[index].name)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateStructMethods(typeName, methods, index+1, next)
}

func selfhost_compiler_compiler_checkDuplicateEnumMembers(typeDecl IREnumType, errors []string) []string {
	next := selfhost_compiler_compiler_checkDuplicateEnumConstructors(typeDecl.name, typeDecl.members, 0, errors)
	return selfhost_compiler_compiler_checkDuplicateEnumMethods(typeDecl.name, typeDecl.methods, 0, next)
}

func selfhost_compiler_compiler_checkDuplicateEnumConstructors(enumName string, members []IREnumMember, index int, errors []string) []string {
	done := index >= len(members)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateEnumConstructor(enumName, members, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateEnumConstructor(enumName string, members []IREnumMember, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerEnumMemberNameAppearsBefore(members, members[index].name, index-1)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate enum member "+enumName+"."+members[index].name)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateEnumConstructors(enumName, members, index+1, next)
}

func selfhost_compiler_compiler_checkDuplicateEnumMethods(enumName string, methods []IRFunction, index int, errors []string) []string {
	done := index >= len(methods)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateEnumMethod(enumName, methods, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateEnumMethod(enumName string, methods []IRFunction, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerMethodNameAppearsBefore(methods, methods[index].name, index-1)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate method "+enumName+"."+methods[index].name)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateEnumMethods(enumName, methods, index+1, next)
}

func selfhost_compiler_compiler_checkDeclarationTypes(file IRFile, knownTypes []string, errors []string) []string {
	next := errors
	for _, constant := range file.constants {
		_ = constant
		next = selfhost_compiler_compiler_checkConstDeclarationTypes(constant, knownTypes, next)
	}
	for _, typeDecl := range file.structs {
		_ = typeDecl
		next = selfhost_compiler_compiler_checkStructDeclarationTypes(typeDecl, knownTypes, next)
	}
	for _, typeDecl := range file.enums {
		_ = typeDecl
		next = selfhost_compiler_compiler_checkEnumDeclarationTypes(typeDecl, knownTypes, next)
	}
	for _, fn := range file.functions {
		_ = fn
		next = selfhost_compiler_compiler_checkFunctionDeclarationTypes(fn, knownTypes, next)
	}
	return next
}

func selfhost_compiler_compiler_checkConstDeclarationTypes(constant IRConst, knownTypes []string, errors []string) []string {
	return selfhost_compiler_compiler_checkCompilerTypeName(constant.typeName, knownTypes, []string{}, errors)
}

func selfhost_compiler_compiler_checkStructDeclarationTypes(typeDecl IRStructType, knownTypes []string, errors []string) []string {
	next := errors
	for _, field := range typeDecl.fields {
		_ = field
		next = selfhost_compiler_compiler_checkCompilerTypeName(field.typeName, knownTypes, typeDecl.generics, next)
	}
	for _, method := range typeDecl.methods {
		_ = method
		next = selfhost_compiler_compiler_checkFunctionDeclarationTypesWithGenerics(method, knownTypes, typeDecl.generics, next)
	}
	return next
}

func selfhost_compiler_compiler_checkEnumDeclarationTypes(typeDecl IREnumType, knownTypes []string, errors []string) []string {
	next := errors
	for _, member := range typeDecl.members {
		_ = member
		func() {
			for _, param := range member.params {
				_ = param
				next = selfhost_compiler_compiler_checkCompilerTypeName(param.typeName, knownTypes, typeDecl.generics, next)
			}
		}()
	}
	for _, method := range typeDecl.methods {
		_ = method
		next = selfhost_compiler_compiler_checkFunctionDeclarationTypesWithGenerics(method, knownTypes, typeDecl.generics, next)
	}
	return next
}

func selfhost_compiler_compiler_checkFunctionDeclarationTypes(fn IRFunction, knownTypes []string, errors []string) []string {
	return selfhost_compiler_compiler_checkFunctionDeclarationTypesWithGenerics(fn, knownTypes, []string{}, errors)
}

func selfhost_compiler_compiler_checkFunctionDeclarationTypesWithGenerics(fn IRFunction, knownTypes []string, parentGenerics []string, errors []string) []string {
	generics := selfhost_compiler_compiler_compilerMergeGenerics(parentGenerics, fn.generics)
	next := selfhost_compiler_compiler_checkDuplicateParams(fn.params, 0, errors)
	next = selfhost_compiler_compiler_checkCompilerTypeName(fn.returnType, knownTypes, generics, next)
	for _, param := range fn.params {
		_ = param
		next = selfhost_compiler_compiler_checkCompilerTypeName(param.typeName, knownTypes, generics, next)
	}
	return next
}

func selfhost_compiler_compiler_checkDuplicateParams(params []IRParam, index int, errors []string) []string {
	done := index >= len(params)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkDuplicateParam(params, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkDuplicateParam(params []IRParam, index int, errors []string) []string {
	duplicate := selfhost_compiler_compiler_compilerParamNameAppearsBefore(params, params[index].name, index-1)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate parameter \""+params[index].name+"\"")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkDuplicateParams(params, index+1, next)
}

func selfhost_compiler_compiler_compilerMergeGenerics(parentGenerics []string, functionGenerics []string) []string {
	out := append([]string{}, parentGenerics[0:len(parentGenerics)]...)
	for _, name := range functionGenerics {
		_ = name
		func() int { out = append(out, name); return len(out) }()
	}
	return out
}

func selfhost_compiler_compiler_checkCompilerTypeName(typeName string, knownTypes []string, generics []string, errors []string) []string {
	normalized := selfhost_compiler_compiler_compilerNormalizeTypeName(typeName)
	shouldSkip := selfhost_compiler_compiler_compilerShouldSkipTypeName(normalized)
	return func() []string {
		switch {
		case shouldSkip == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkCompilerNamedType(normalized, knownTypes, generics, errors)
		}
	}()
}

func selfhost_compiler_compiler_compilerNormalizeTypeName(typeName string) string {
	nullable := strings.HasSuffix(typeName, "?")
	return func() string {
		switch {
		case nullable == true:
			return selfhost_compiler_compiler_compilerNormalizeTypeName(func() string { runes := []rune(typeName); return string(runes[0 : len([]rune(typeName))-1]) }())
		default:
			return typeName
		}
	}()
}

func selfhost_compiler_compiler_compilerShouldSkipTypeName(typeName string) bool {
	return typeName == "" || (strings.HasPrefix(typeName, "@") || strings.HasPrefix(typeName, "(") || strings.HasPrefix(typeName, "Syntax"))
}

func selfhost_compiler_compiler_checkCompilerNamedType(typeName string, knownTypes []string, generics []string, errors []string) []string {
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	known := selfhost_compiler_compiler_compilerContains(knownTypes, base) || selfhost_compiler_compiler_compilerContains(generics, base)
	return func() []string {
		switch {
		case known == true:
			return selfhost_compiler_compiler_checkCompilerTypeArgs(typeName, knownTypes, generics, errors)
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "unknown type \""+base+"\"")
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkCompilerTypeArgs(typeName string, knownTypes []string, generics []string, errors []string) []string {
	inner := selfhost_compiler_compiler_compilerGenericInner(typeName)
	empty := inner == ""
	return func() []string {
		switch {
		case empty == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkCompilerTypeArgList(func() []string { parts := strings.Split(inner, ","); return parts }(), knownTypes, generics, errors, 0)
		}
	}()
}

func selfhost_compiler_compiler_checkCompilerTypeArgList(args []string, knownTypes []string, generics []string, errors []string, index int) []string {
	done := index >= len(args)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkCompilerTypeArgList(args, knownTypes, generics, selfhost_compiler_compiler_checkCompilerTypeName(strings.TrimSpace(args[index]), knownTypes, generics, errors), index+1)
		}
	}()
}

func selfhost_compiler_compiler_compilerGenericInner(typeName string) string {
	open := strings.Index(typeName, "[")
	complete := open >= 0 && strings.HasSuffix(typeName, "]")
	return func() string {
		switch {
		case complete == true:
			return func() string { runes := []rune(typeName); return string(runes[open+1 : len([]rune(typeName))-1]) }()
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerCallables(file IRFile) []CompilerCallable {
	callables := []CompilerCallable{}
	for _, fn := range file.functions {
		_ = fn
		func() int {
			if fn.macro {
				return 0
			}
			return func() int {
				callables = append(callables, selfhost_compiler_compiler_compilerFunctionCallable(fn))
				return len(callables)
			}()
		}()
	}
	for _, importDecl := range file.tsImports {
		_ = importDecl
		func() {
			for _, fn := range importDecl.functions {
				_ = fn
				func() int {
					callables = append(callables, selfhost_compiler_compiler_compilerCallable(fn.name, len(fn.params), fn.returnType, selfhost_compiler_compiler_compilerParamTypeNames(fn.params), false, ""))
					return len(callables)
				}()
			}
		}()
	}
	for _, typeDecl := range file.structs {
		_ = typeDecl
		func() {
			for _, method := range typeDecl.methods {
				_ = method
				func() int {
					callables = append(callables, selfhost_compiler_compiler_compilerMethodCallable(typeDecl.name, method))
					return len(callables)
				}()
			}
		}()
	}
	for _, typeDecl := range file.enums {
		_ = typeDecl
		func() {
			for _, method := range typeDecl.methods {
				_ = method
				func() int {
					callables = append(callables, selfhost_compiler_compiler_compilerMethodCallable(typeDecl.name, method))
					return len(callables)
				}()
			}
		}()
	}
	for _, typeDecl := range file.enums {
		_ = typeDecl
		func() {
			for _, member := range typeDecl.members {
				_ = member
				func() int {
					callables = append(callables, selfhost_compiler_compiler_compilerCallable(member.name, len(member.params), typeDecl.name, selfhost_compiler_compiler_compilerParamTypeNames(member.params), false, ""))
					return len(callables)
				}()
			}
		}()
	}
	return callables
}

func selfhost_compiler_compiler_compilerFunctionCallable(fn IRFunction) CompilerCallable {
	return selfhost_compiler_compiler_compilerCallable(fn.name, len(fn.params), fn.returnType, selfhost_compiler_compiler_compilerParamTypeNames(fn.params), fn.private, fn.sourcePath)
}

func selfhost_compiler_compiler_compilerMethodCallable(typeName string, method IRFunction) CompilerCallable {
	return selfhost_compiler_compiler_compilerCallable(selfhost_compiler_compiler_compilerMethodCallableName(typeName, method), len(method.params), method.returnType, selfhost_compiler_compiler_compilerParamTypeNames(method.params), method.private, method.sourcePath)
}

func selfhost_compiler_compiler_compilerMethodCallableName(typeName string, method IRFunction) string {
	return func() string {
		switch {
		case method.static == true:
			return selfhost_compiler_compiler_compilerStaticMethodName(typeName, method.name)
		default:
			return selfhost_compiler_compiler_compilerInstanceMethodName(typeName, method.name)
		}
	}()
}

func selfhost_compiler_compiler_compilerStaticMethodName(typeName string, methodName string) string {
	return typeName + "::" + methodName
}

func selfhost_compiler_compiler_compilerInstanceMethodName(typeName string, methodName string) string {
	return typeName + "." + methodName
}

func selfhost_compiler_compiler_checkFunctionErrors(fn IRFunction, structs []IRStructType, callables []CompilerCallable, errors []string, baseBindings []CompilerTypeBinding) []string {
	bindings := selfhost_compiler_compiler_compilerFunctionBindings(fn.params, baseBindings)
	return selfhost_compiler_compiler_checkFunctionReturn(fn, structs, callables, selfhost_compiler_compiler_checkExpr(fn.body, structs, callables, errors, bindings), bindings)
}

func selfhost_compiler_compiler_checkMethodErrors(typeName string, method IRFunction, structs []IRStructType, callables []CompilerCallable, errors []string, baseBindings []CompilerTypeBinding) []string {
	return selfhost_compiler_compiler_checkFunctionErrors(method, structs, callables, errors, selfhost_compiler_compiler_compilerMethodBaseBindings(typeName, method, baseBindings))
}

func selfhost_compiler_compiler_compilerMethodBaseBindings(typeName string, method IRFunction, baseBindings []CompilerTypeBinding) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		switch {
		case method.static == true:
			return baseBindings
		default:
			return selfhost_compiler_compiler_addCompilerTypeBinding(baseBindings, "this", typeName)
		}
	}()
}

func selfhost_compiler_compiler_checkExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_compiler_checkBlockExpr(expr.children, 0, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Let:
			return selfhost_compiler_compiler_checkLetExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Struct:
			return selfhost_compiler_compiler_checkStructExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Object:
			return selfhost_compiler_compiler_checkObjectExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Ternary:
			return selfhost_compiler_compiler_checkTernaryExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Unary:
			return selfhost_compiler_compiler_checkUnaryExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Postfix:
			return selfhost_compiler_compiler_checkPostfixExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Binary:
			return selfhost_compiler_compiler_checkBinaryExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Array:
			return selfhost_compiler_compiler_checkArrayExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Map:
			return selfhost_compiler_compiler_checkMapExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Index:
			return selfhost_compiler_compiler_checkIndexExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_compiler_checkCallExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Selector:
			return selfhost_compiler_compiler_checkSelectorValueExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Lambda:
			return selfhost_compiler_compiler_checkLambdaExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Assign:
			return selfhost_compiler_compiler_checkAssignExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_PatternBlock:
			return selfhost_compiler_compiler_checkPatternBlockExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Match:
			return selfhost_compiler_compiler_checkMatchExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_ObjectDestructure:
			return selfhost_compiler_compiler_checkObjectDestructureExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Unwrap:
			return selfhost_compiler_compiler_checkUnwrapExpr(expr, structs, callables, errors, bindings)
		case expr.kind == ExprKind_Identifier:
			return selfhost_compiler_compiler_checkIdentifierExpr(expr, callables, bindings, errors)
		case expr.kind == ExprKind_This:
			return selfhost_compiler_compiler_checkThisExpr(bindings, errors)
		default:
			return selfhost_compiler_compiler_checkExprDefault(expr, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkIdentifierExpr(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case selfhost_compiler_compiler_compilerIdentifierDefined(expr, callables, bindings) == true:
			return errors
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "undefined name \""+expr.name+"\"")
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerIdentifierDefined(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) bool {
	binding := selfhost_compiler_compiler_compilerIdentifierBinding(expr, bindings)
	return func() bool {
		if binding.name == "" {
			return selfhost_compiler_compiler_findCompilerCallable(callables, expr.name, 0).name != ""
		}
		return true
	}()
}

func selfhost_compiler_compiler_compilerIdentifierBinding(expr IRExpr, bindings []CompilerTypeBinding) CompilerTypeBinding {
	binding := selfhost_compiler_compiler_findCompilerTypeBinding(bindings, expr.name, 0)
	return func() CompilerTypeBinding {
		if binding.name != "" {
			return binding
		}
		return func() CompilerTypeBinding {
			if strings.HasPrefix(expr.text, "$") {
				return selfhost_compiler_compiler_findCompilerTypeBinding(bindings, expr.text, 0)
			}
			return selfhost_compiler_compiler_emptyCompilerTypeBinding()
		}()
	}()
}

func selfhost_compiler_compiler_checkThisExpr(bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case selfhost_compiler_compiler_findCompilerTypeBinding(bindings, "this", 0).typeName == "":
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "implicit this selector can only be used inside a method")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkCallExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return selfhost_compiler_compiler_checkCallArgExprs(expr.children, 1, structs, callables, selfhost_compiler_compiler_checkExprCall(expr, structs, callables, errors, bindings), bindings)
}

func selfhost_compiler_compiler_checkCallArgExprs(args []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	done := index >= len(args)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkCallArgExprs(args, index+1, structs, callables, selfhost_compiler_compiler_checkExpr(args[index], structs, callables, errors, bindings), bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkSelectorValueExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return selfhost_compiler_compiler_checkSelectorReceiverValueExpr(expr, structs, callables, bindings, selfhost_compiler_compiler_checkSelectorExpr(expr, structs, callables, bindings, errors))
}

func selfhost_compiler_compiler_checkSelectorReceiverValueExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	hasReceiver := len(expr.children) > 0
	return func() []string {
		switch {
		case hasReceiver == true:
			return selfhost_compiler_compiler_checkSelectorReceiverValue(expr, expr.children[0], structs, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkSelectorReceiverValue(expr IRExpr, receiver IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	staticTypeReceiver := selfhost_compiler_compiler_compilerStaticSelectorReceiver(expr, receiver)
	skipReceiver := staticTypeReceiver || receiver.kind == ExprKind_At
	return func() []string {
		switch {
		case skipReceiver == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkExpr(receiver, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_compilerStaticSelectorReceiver(expr IRExpr, receiver IRExpr) bool {
	return func() bool {
		switch {
		case receiver.kind == ExprKind_Identifier:
			return expr.op == "::" || selfhost_compiler_compiler_compilerLooksLikeTypeName(receiver.name)
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_compilerLooksLikeTypeName(name string) bool {
	empty := len([]rune(name)) == 0
	return func() bool {
		switch {
		case empty == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerUppercaseChar([]rune(name)[0])
		}
	}()
}

func selfhost_compiler_compiler_compilerUppercaseChar(ch rune) bool {
	return ch >= 'A' && ch <= 'Z'
}

func selfhost_compiler_compiler_checkLambdaExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	hasBody := len(expr.children) > 0
	return func() []string {
		switch {
		case hasBody == true:
			return selfhost_compiler_compiler_checkExpr(expr.children[0], structs, callables, errors, selfhost_compiler_compiler_compilerFunctionBindings(expr.params, bindings))
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkObjectExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return selfhost_compiler_compiler_checkObjectMembers(expr.children, 0, structs, callables, errors, bindings, selfhost_compiler_compiler_compilerObjectType(expr.children, 0, structs, callables, bindings, "{"))
}

func selfhost_compiler_compiler_compilerObjectType(members []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, out string) string {
	spreadType := selfhost_compiler_compiler_compilerObjectSpreadType(members, index, structs, callables, bindings)
	return func() string {
		if spreadType == "" {
			return selfhost_compiler_compiler_compilerObjectTypeFields(members, index, structs, callables, bindings, out)
		}
		return spreadType
	}()
}

func selfhost_compiler_compiler_compilerObjectSpreadType(members []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		if index >= len(members) {
			return ""
		}
		return func() string {
			if members[index].kind == ExprKind_Spread {
				return selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(members[index].children[0], structs, callables, bindings)
			}
			return selfhost_compiler_compiler_compilerObjectSpreadType(members, index+1, structs, callables, bindings)
		}()
	}()
}

func selfhost_compiler_compiler_compilerObjectTypeFields(members []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, out string) string {
	done := index >= len(members)
	return func() string {
		switch {
		case done == true:
			return out + "}"
		default:
			return selfhost_compiler_compiler_compilerObjectTypeMember(members, index, structs, callables, bindings, out)
		}
	}()
}

func selfhost_compiler_compiler_compilerObjectTypeMember(members []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, out string) string {
	member := members[index]
	return func() string {
		if member.kind == ExprKind_Spread {
			return selfhost_compiler_compiler_compilerObjectTypeFields(members, index+1, structs, callables, bindings, out)
		}
		return selfhost_compiler_compiler_compilerObjectTypeNamedMember(members, index, structs, callables, bindings, out)
	}()
}

func selfhost_compiler_compiler_compilerObjectTypeNamedMember(members []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, out string) string {
	member := members[index]
	separator := func() string {
		if out == "{" {
			return ""
		}
		return ";"
	}()
	typeName := func() string {
		switch {
		case member.kind == ExprKind_Method:
			return member.text
		case member.kind == ExprKind_PrivateMethod:
			return member.text
		default:
			return func() string {
				if len(member.children) > 0 {
					return selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(member.children[0], structs, callables, bindings)
				}
				return ""
			}()
		}
	}()
	return selfhost_compiler_compiler_compilerObjectTypeFields(members, index+1, structs, callables, bindings, out+separator+member.name+":"+typeName)
}

func selfhost_compiler_compiler_checkObjectMembers(members []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding, objectType string) []string {
	done := index >= len(members)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkObjectMember(members, index, structs, callables, errors, bindings, objectType)
		}
	}()
}

func selfhost_compiler_compiler_checkObjectMember(members []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding, objectType string) []string {
	member := members[index]
	next := selfhost_compiler_compiler_checkObjectMemberExpr(member, structs, callables, errors, bindings, objectType)
	return selfhost_compiler_compiler_checkObjectMembers(members, index+1, structs, callables, next, bindings, objectType)
}

func selfhost_compiler_compiler_checkObjectMemberExpr(member IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding, objectType string) []string {
	return func() []string {
		switch {
		case member.kind == ExprKind_Spread:
			return selfhost_compiler_compiler_checkObjectSpreadExpr(member, structs, callables, errors, bindings, objectType)
		case member.kind == ExprKind_Method:
			return selfhost_compiler_compiler_checkObjectMethodExpr(member, structs, callables, errors, bindings, objectType)
		case member.kind == ExprKind_PrivateMethod:
			return selfhost_compiler_compiler_checkObjectMethodExpr(member, structs, callables, errors, bindings, objectType)
		default:
			return selfhost_compiler_compiler_checkExprDefault(member, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkObjectSpreadExpr(member IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding, objectType string) []string {
	value := member.children[0]
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings)
	valid := actual != "" && actual != "Dynamic" && actual == objectType
	checked := selfhost_compiler_compiler_checkExpr(value, structs, callables, errors, bindings)
	return func() []string {
		if valid {
			return checked
		}
		return func() []string {
			__rune_spread_out := []string{}
			__rune_spread_out = append(__rune_spread_out, checked...)
			__rune_spread_out = append(__rune_spread_out, "object spread expects "+objectType+", got "+actual)
			return __rune_spread_out
		}()
	}()
}

func selfhost_compiler_compiler_checkObjectMethodExpr(method IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding, objectType string) []string {
	hasBody := len(method.children) > 0
	return func() []string {
		switch {
		case hasBody == true:
			return selfhost_compiler_compiler_checkExpr(method.children[0], structs, callables, errors, selfhost_compiler_compiler_compilerFunctionBindings(method.params, selfhost_compiler_compiler_addCompilerTypeBinding(bindings, "this", objectType)))
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkAssignExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return selfhost_compiler_compiler_checkAssignExpressionExpr(expr, structs, callables, errors, bindings)
}

func selfhost_compiler_compiler_checkAssignExpressionExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	complete := len(expr.children) >= 2
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkAssignTargetExpr(expr.children[0], expr.children[1], structs, callables, errors, bindings)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkAssignTargetExpr(target IRExpr, value IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case target.kind == ExprKind_Identifier:
			return selfhost_compiler_compiler_checkIdentifierAssignExpr(target, value, structs, callables, errors, bindings)
		case target.kind == ExprKind_Index:
			return selfhost_compiler_compiler_checkIndexAssignExpr(target, value, structs, callables, errors, bindings)
		default:
			return selfhost_compiler_compiler_checkTargetAssignExpr(target, value, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkIdentifierAssignExpr(target IRExpr, value IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	binding := selfhost_compiler_compiler_compilerIdentifierBinding(target, bindings)
	checked := selfhost_compiler_compiler_checkExpr(value, structs, callables, errors, bindings)
	return func() []string {
		switch {
		case binding.name == "":
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, checked...)
				__rune_spread_out = append(__rune_spread_out, "cannot assign undefined name \""+target.name+"\"")
				return __rune_spread_out
			}()
		default:
			return selfhost_compiler_compiler_checkAssignmentType(binding.typeName, selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings), checked)
		}
	}()
}

func selfhost_compiler_compiler_checkIndexAssignExpr(target IRExpr, value IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkIndexExpr(target, structs, callables, selfhost_compiler_compiler_checkExpr(value, structs, callables, errors, bindings), bindings)
	return selfhost_compiler_compiler_checkAssignmentType(selfhost_compiler_compiler_inferCompilerIndexType(target, callables, bindings), selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings), checked)
}

func selfhost_compiler_compiler_checkTargetAssignExpr(target IRExpr, value IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExpr(value, structs, callables, selfhost_compiler_compiler_checkExpr(target, structs, callables, errors, bindings), bindings)
	return selfhost_compiler_compiler_checkAssignmentType(selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(target, structs, callables, bindings), selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings), checked)
}

func selfhost_compiler_compiler_checkAssignmentType(expected string, actual string, errors []string) []string {
	mismatch := selfhost_compiler_compiler_compilerShouldCheckArgType(expected, actual) && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "assignment has type "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkPatternBlockExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return selfhost_compiler_compiler_checkPatternBlockBranches(expr.children, 0, structs, callables, errors, bindings, "")
}

func selfhost_compiler_compiler_checkMatchExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	hasSubject := len(expr.children) > 0
	checked := func() []string {
		switch {
		case hasSubject == true:
			return selfhost_compiler_compiler_checkExpr(expr.children[0], structs, callables, errors, bindings)
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkPatternBlockBranches(expr.children, 1, structs, callables, checked, bindings, "")
}

func selfhost_compiler_compiler_checkPatternBlockBranches(branches []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding, expected string) []string {
	done := index >= len(branches)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkPatternBlockBranch(branches, index, structs, callables, errors, bindings, expected)
		}
	}()
}

func selfhost_compiler_compiler_checkPatternBlockBranch(branches []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding, expected string) []string {
	branch := branches[index]
	complete := len(branch.children) >= 2
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkPatternBlockBranchValue(branches, index, branch.children[1], structs, callables, errors, bindings, expected)
		default:
			return selfhost_compiler_compiler_checkPatternBlockBranches(branches, index+1, structs, callables, errors, bindings, expected)
		}
	}()
}

func selfhost_compiler_compiler_checkPatternBlockBranchValue(branches []IRExpr, index int, value IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding, expected string) []string {
	checked := selfhost_compiler_compiler_checkExpr(value, structs, callables, errors, bindings)
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings)
	nextExpected := selfhost_compiler_compiler_compilerNextPatternBranchType(expected, actual)
	nextErrors := selfhost_compiler_compiler_checkPatternBranchTypeError(expected, actual, checked)
	return selfhost_compiler_compiler_checkPatternBlockBranches(branches, index+1, structs, callables, nextErrors, bindings, nextExpected)
}

func selfhost_compiler_compiler_compilerNextPatternBranchType(expected string, actual string) string {
	return func() string {
		switch {
		case expected == "":
			return actual
		default:
			return expected
		}
	}()
}

func selfhost_compiler_compiler_checkPatternBranchTypeError(expected string, actual string, errors []string) []string {
	mismatch := selfhost_compiler_compiler_compilerShouldCheckArgType(expected, actual) && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "pattern branch returns "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkObjectDestructureExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	hasValue := len(expr.children) > 0
	return func() []string {
		switch {
		case hasValue == true:
			return selfhost_compiler_compiler_checkObjectDestructureValue(expr, structs, callables, errors, bindings)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkObjectDestructureValue(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	value := expr.children[0]
	checked := selfhost_compiler_compiler_checkExpr(value, structs, callables, errors, bindings)
	sourceType := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings)
	return selfhost_compiler_compiler_checkObjectDestructureSourceType(expr.params, sourceType, structs, checked)
}

func selfhost_compiler_compiler_checkObjectDestructureSourceType(params []IRParam, sourceType string, structs []IRStructType, errors []string) []string {
	return func() []string {
		switch {
		case sourceType == "":
			return errors
		default:
			return func() []string {
				typeDecl := selfhost_compiler_compiler_findCompilerStruct(structs, selfhost_compiler_compiler_compilerTypeBase(sourceType), 0)
				found := typeDecl.name != ""
				return func() []string {
					switch {
					case found == true:
						return selfhost_compiler_compiler_checkObjectDestructureFields(params, 0, sourceType, typeDecl.fields, errors)
					default:
						return func() []string {
							__rune_spread_out := []string{}
							__rune_spread_out = append(__rune_spread_out, errors...)
							__rune_spread_out = append(__rune_spread_out, "type "+sourceType+" has no fields")
							return __rune_spread_out
						}()
					}
				}()
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkObjectDestructureFields(params []IRParam, index int, sourceType string, fields []IRField, errors []string) []string {
	done := index >= len(params)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkObjectDestructureField(params, index, sourceType, fields, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkObjectDestructureField(params []IRParam, index int, sourceType string, fields []IRField, errors []string) []string {
	fieldName := params[index].typeName
	duplicate := selfhost_compiler_compiler_compilerDestructureFieldAppearsBefore(params, fieldName, index-1)
	field := selfhost_compiler_compiler_findCompilerStructField(fields, fieldName, 0)
	next := func() []string {
		switch {
		case duplicate == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "duplicate destructured field \""+fieldName+"\"")
				return __rune_spread_out
			}()
		default:
			return func() []string {
				switch {
				case field.name == "":
					return func() []string {
						__rune_spread_out := []string{}
						__rune_spread_out = append(__rune_spread_out, errors...)
						__rune_spread_out = append(__rune_spread_out, "type "+sourceType+" has no field \""+fieldName+"\"")
						return __rune_spread_out
					}()
				default:
					return errors
				}
			}()
		}
	}()
	return selfhost_compiler_compiler_checkObjectDestructureFields(params, index+1, sourceType, fields, next)
}

func selfhost_compiler_compiler_compilerDestructureFieldAppearsBefore(params []IRParam, name string, index int) bool {
	done := index < 0
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerDestructureFieldAppearsBeforeAt(params, name, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerDestructureFieldAppearsBeforeAt(params []IRParam, name string, index int) bool {
	matched := params[index].typeName == name
	return func() bool {
		switch {
		case matched == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerDestructureFieldAppearsBefore(params, name, index-1)
		}
	}()
}

func selfhost_compiler_compiler_checkUnwrapExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	complete := len(expr.children) > 0
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkUnwrapSourceExpr(expr.children[0], structs, callables, errors, bindings)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkUnwrapSourceExpr(source IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExpr(source, structs, callables, errors, bindings)
	sourceType := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(source, structs, callables, bindings)
	return selfhost_compiler_compiler_checkUnwrapSourceType(sourceType, checked)
}

func selfhost_compiler_compiler_checkUnwrapSourceType(sourceType string, errors []string) []string {
	return func() []string {
		switch {
		case sourceType == "":
			return errors
		default:
			return func() []string {
				switch {
				case selfhost_compiler_compiler_compilerResultOkType(sourceType) == "":
					return func() []string {
						__rune_spread_out := []string{}
						__rune_spread_out = append(__rune_spread_out, errors...)
						__rune_spread_out = append(__rune_spread_out, "operator '?' expects Result, got "+sourceType)
						return __rune_spread_out
					}()
				default:
					return errors
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkExprDefault(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	next := selfhost_compiler_compiler_checkExprCall(expr, structs, callables, errors, bindings)
	for _, child := range expr.children {
		_ = child
		next = selfhost_compiler_compiler_checkExpr(child, structs, callables, next, bindings)
	}
	return next
}

func selfhost_compiler_compiler_checkTernaryExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExprDefault(expr, structs, callables, errors, bindings)
	withCondition := selfhost_compiler_compiler_checkTernaryCondition(expr, callables, bindings, checked)
	return selfhost_compiler_compiler_checkTernaryBranches(expr, callables, bindings, withCondition)
}

func selfhost_compiler_compiler_checkTernaryCondition(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	complete := len(expr.children) > 0
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkTernaryConditionType(expr.children[0], callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkTernaryConditionType(condition IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	actual := selfhost_compiler_compiler_inferCompilerExprType(condition, callables, bindings)
	mismatch := actual != "" && actual != "Bool"
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "ternary condition expects Bool, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkTernaryBranches(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	complete := len(expr.children) >= 3
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkTernaryBranchTypes(expr.children[1], expr.children[2], callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkTernaryBranchTypes(consequence IRExpr, alternative IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	left := selfhost_compiler_compiler_inferCompilerExprType(consequence, callables, bindings)
	right := selfhost_compiler_compiler_inferCompilerExprType(alternative, callables, bindings)
	shouldCheck := left != "" && right != ""
	mismatch := shouldCheck && selfhost_compiler_compiler_inferCompilerCommonType(left, right) == ""
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "ternary branches return "+left+" and "+right)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkUnaryExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExprDefault(expr, structs, callables, errors, bindings)
	return func() []string {
		switch {
		case expr.op == "!":
			return selfhost_compiler_compiler_checkBoolOperand("!", expr.children[0], callables, bindings, checked)
		case expr.op == "-":
			return selfhost_compiler_compiler_checkNumericUnaryOperand(expr, callables, bindings, checked)
		case expr.op == "~":
			return selfhost_compiler_compiler_checkBitwiseUnaryOperand(expr, callables, bindings, checked)
		default:
			return checked
		}
	}()
}

func selfhost_compiler_compiler_checkPostfixExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExprDefault(expr, structs, callables, errors, bindings)
	return func() []string {
		switch {
		case expr.op == "++":
			return selfhost_compiler_compiler_checkPostfixNumericOperand(expr, callables, bindings, checked)
		default:
			return checked
		}
	}()
}

func selfhost_compiler_compiler_checkBinaryExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExprDefault(expr, structs, callables, errors, bindings)
	boolOp := func() bool {
		switch {
		case (expr.op == "&&") || (expr.op == "||"):
			return true
		default:
			return false
		}
	}()
	return func() []string {
		switch {
		case boolOp == true:
			return selfhost_compiler_compiler_checkBinaryBoolOperands(expr, callables, bindings, checked)
		default:
			return selfhost_compiler_compiler_checkTypedBinaryExpr(expr, callables, bindings, checked)
		}
	}()
}

func selfhost_compiler_compiler_checkTypedBinaryExpr(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	checked := selfhost_compiler_compiler_checkEqualityComparisonExpr(expr, callables, bindings, errors)
	checked = selfhost_compiler_compiler_checkOrderedComparisonExpr(expr, callables, bindings, checked)
	checked = selfhost_compiler_compiler_checkNumericBinaryExpr(expr, callables, bindings, checked)
	return selfhost_compiler_compiler_checkBitwiseBinaryExpr(expr, callables, bindings, checked)
}

func selfhost_compiler_compiler_checkBinaryBoolOperands(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	complete := len(expr.children) >= 2
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkBoolOperand(expr.op, expr.children[1], callables, bindings, selfhost_compiler_compiler_checkBoolOperand(expr.op, expr.children[0], callables, bindings, errors))
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkBoolOperand(op string, operand IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	actual := selfhost_compiler_compiler_inferCompilerExprType(operand, callables, bindings)
	mismatch := actual != "" && actual != "Bool"
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "operator '"+op+"' expects Bool, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkArrayExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExprDefault(expr, structs, callables, errors, bindings)
	return selfhost_compiler_compiler_checkArrayElementTypes(expr.children, callables, bindings, checked, 0, "")
}

func selfhost_compiler_compiler_checkArrayElementTypes(elements []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string, index int, expected string) []string {
	done := index >= len(elements)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkArrayElementType(elements, callables, bindings, errors, index, expected)
		}
	}()
}

func selfhost_compiler_compiler_checkArrayElementType(elements []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string, index int, expected string) []string {
	elem := elements[index]
	actual := selfhost_compiler_compiler_inferCompilerArrayLiteralElementType(elem, callables, bindings)
	nextExpected := selfhost_compiler_compiler_compilerNextArrayElementType(expected, actual)
	nextErrors := selfhost_compiler_compiler_checkArrayElementTypeError(elem, expected, actual, callables, bindings, errors)
	return selfhost_compiler_compiler_checkArrayElementTypes(elements, callables, bindings, nextErrors, index+1, nextExpected)
}

func selfhost_compiler_compiler_compilerNextArrayElementType(expected string, actual string) string {
	return func() string {
		switch {
		case expected == "":
			return actual
		default:
			return expected
		}
	}()
}

func selfhost_compiler_compiler_checkArrayElementTypeError(elem IRExpr, expected string, actual string, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	spreadErrors := selfhost_compiler_compiler_checkArraySpreadElementType(elem, callables, bindings, errors)
	mismatch := expected != "" && actual != "" && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, spreadErrors...)
				__rune_spread_out = append(__rune_spread_out, "array element has type "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		default:
			return spreadErrors
		}
	}()
}

func selfhost_compiler_compiler_checkArraySpreadElementType(elem IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case elem.kind == ExprKind_Spread:
			return selfhost_compiler_compiler_checkArraySpreadReceiverType(elem, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkArraySpreadReceiverType(elem IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	actual := selfhost_compiler_compiler_inferCompilerExprType(elem.children[0], callables, bindings)
	mismatch := actual != "" && selfhost_compiler_compiler_compilerArrayElementType(actual) == ""
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "spread expects Array, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkMapExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExprDefault(expr, structs, callables, errors, bindings)
	return selfhost_compiler_compiler_checkMapEntryTypes(expr.children, callables, bindings, checked, 0, "", "")
}

func selfhost_compiler_compiler_checkMapEntryTypes(entries []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string, index int, expectedKey string, expectedValue string) []string {
	done := index >= len(entries)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkMapEntryType(entries, callables, bindings, errors, index, expectedKey, expectedValue)
		}
	}()
}

func selfhost_compiler_compiler_checkMapEntryType(entries []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string, index int, expectedKey string, expectedValue string) []string {
	entry := entries[index]
	key := selfhost_compiler_compiler_inferCompilerMapEntryKeyType(entry, callables, bindings)
	value := selfhost_compiler_compiler_inferCompilerMapEntryValueType(entry, callables, bindings)
	nextKey := selfhost_compiler_compiler_compilerNextMapEntryType(expectedKey, key)
	nextValue := selfhost_compiler_compiler_compilerNextMapEntryType(expectedValue, value)
	checked := selfhost_compiler_compiler_checkMapEntryTypeErrors(entry, expectedKey, key, expectedValue, value, errors)
	return selfhost_compiler_compiler_checkMapEntryTypes(entries, callables, bindings, checked, index+1, nextKey, nextValue)
}

func selfhost_compiler_compiler_compilerNextMapEntryType(expected string, actual string) string {
	return func() string {
		switch {
		case expected == "":
			return actual
		default:
			return expected
		}
	}()
}

func selfhost_compiler_compiler_checkMapEntryTypeErrors(entry IRExpr, expectedKey string, key string, expectedValue string, value string, errors []string) []string {
	checked := selfhost_compiler_compiler_checkMapEntryKeyTypeError(expectedKey, key, errors)
	return selfhost_compiler_compiler_checkMapEntryValueTypeError(expectedValue, value, checked)
}

func selfhost_compiler_compiler_checkMapEntryKeyTypeError(expected string, actual string, errors []string) []string {
	checked := selfhost_compiler_compiler_checkMapKeySupported(actual, errors)
	mismatch := expected != "" && actual != "" && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, checked...)
				__rune_spread_out = append(__rune_spread_out, "map key has type "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		default:
			return checked
		}
	}()
}

func selfhost_compiler_compiler_checkMapKeySupported(actual string, errors []string) []string {
	invalid := actual != "" && selfhost_compiler_compiler_compilerSupportedMapKeyType(actual) == false
	return func() []string {
		switch {
		case invalid == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "map key type "+actual+" is not supported")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkMapEntryValueTypeError(expected string, actual string, errors []string) []string {
	mismatch := expected != "" && actual != "" && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "map value has type "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkIndexExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExprDefault(expr, structs, callables, errors, bindings)
	complete := len(expr.children) >= 2
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkIndexExprTypes(expr, callables, bindings, checked)
		default:
			return checked
		}
	}()
}

func selfhost_compiler_compiler_checkIndexExprTypes(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	receiver := selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
	index := selfhost_compiler_compiler_inferCompilerExprType(expr.children[1], callables, bindings)
	return selfhost_compiler_compiler_checkIndexReceiverType(receiver, index, expr.children[1], errors)
}

func selfhost_compiler_compiler_checkIndexReceiverType(receiver string, index string, indexExpr IRExpr, errors []string) []string {
	arrayElem := selfhost_compiler_compiler_compilerArrayElementType(receiver)
	mapKey := selfhost_compiler_compiler_compilerMapKeyType(receiver)
	tuple := selfhost_compiler_compiler_compilerTupleElementTypes(receiver)
	return func() []string {
		switch {
		case arrayElem == "":
			return selfhost_compiler_compiler_checkNonArrayIndexReceiver(receiver, mapKey, tuple, index, indexExpr, errors)
		default:
			return selfhost_compiler_compiler_checkArrayIndexType(index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkNonArrayIndexReceiver(receiver string, mapKey string, tuple []string, index string, indexExpr IRExpr, errors []string) []string {
	return func() []string {
		switch {
		case mapKey == "":
			return selfhost_compiler_compiler_checkTupleOrUnknownIndexReceiver(receiver, tuple, index, indexExpr, errors)
		default:
			return selfhost_compiler_compiler_checkMapIndexType(mapKey, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkTupleOrUnknownIndexReceiver(receiver string, tuple []string, index string, indexExpr IRExpr, errors []string) []string {
	return func() []string {
		switch {
		case len(tuple) > 0 == true:
			return selfhost_compiler_compiler_checkTupleIndexType(tuple, index, indexExpr, errors)
		default:
			return func() []string {
				switch {
				case receiver == "":
					return errors
				default:
					return func() []string {
						__rune_spread_out := []string{}
						__rune_spread_out = append(__rune_spread_out, errors...)
						__rune_spread_out = append(__rune_spread_out, "type "+receiver+" is not indexable")
						return __rune_spread_out
					}()
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkArrayIndexType(index string, errors []string) []string {
	mismatch := index != "" && index != "Int"
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "array index expects Int, got "+index)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkMapIndexType(expected string, actual string, errors []string) []string {
	mismatch := expected != "" && actual != "" && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "map index has type "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkTupleIndexType(tuple []string, index string, indexExpr IRExpr, errors []string) []string {
	mismatch := index != "" && index != "Int"
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "tuple index expects an integer literal")
				return __rune_spread_out
			}()
		default:
			return selfhost_compiler_compiler_checkTupleIndexLiteral(tuple, indexExpr, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkTupleIndexLiteral(tuple []string, indexExpr IRExpr, errors []string) []string {
	return func() []string {
		switch {
		case indexExpr.kind == ExprKind_Int:
			return selfhost_compiler_compiler_checkTupleIndexRange(tuple, selfhost_compiler_compiler_compilerParseIntText(indexExpr.value), errors)
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "tuple index expects an integer literal")
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkTupleIndexRange(tuple []string, index int, errors []string) []string {
	outOfRange := index < 0 || index >= len(tuple)
	return func() []string {
		switch {
		case outOfRange == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "tuple index "+compilerIntToString(index)+" out of range")
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkEqualityComparisonExpr(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case selfhost_compiler_compiler_compilerEqualityComparisonOp(expr.op) == true:
			return selfhost_compiler_compiler_checkEqualityComparisonOperands(expr, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkEqualityComparisonOperands(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	complete := len(expr.children) >= 2
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkEqualityComparisonTypes(expr, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkEqualityComparisonTypes(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	left := selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
	right := selfhost_compiler_compiler_inferCompilerExprType(expr.children[1], callables, bindings)
	mismatch := left != "" && right != "" && selfhost_compiler_compiler_compilerTypesComparable(left, right) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "cannot compare "+left+" and "+right)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkOrderedComparisonExpr(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case selfhost_compiler_compiler_compilerOrderedComparisonOp(expr.op) == true:
			return selfhost_compiler_compiler_checkOrderedComparisonOperands(expr, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkOrderedComparisonOperands(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	complete := len(expr.children) >= 2
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkOrderedComparisonMatch(expr, callables, bindings, selfhost_compiler_compiler_checkOrderedComparisonOperand(expr.children[1], callables, bindings, selfhost_compiler_compiler_checkOrderedComparisonOperand(expr.children[0], callables, bindings, errors)))
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkOrderedComparisonOperand(operand IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	actual := selfhost_compiler_compiler_inferCompilerExprType(operand, callables, bindings)
	mismatch := actual != "" && selfhost_compiler_compiler_compilerOrderedComparisonType(actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "ordered comparison expects a numeric type or String, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkOrderedComparisonMatch(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	left := selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
	right := selfhost_compiler_compiler_inferCompilerExprType(expr.children[1], callables, bindings)
	mismatch := left != "" && right != "" && left != right
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "ordered comparison requires matching types, got "+left+" and "+right)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_compilerOrderedComparisonOp(op string) bool {
	return func() bool {
		switch {
		case (op == "<") || (op == "<=") || (op == ">") || (op == ">="):
			return true
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_compilerEqualityComparisonOp(op string) bool {
	return func() bool {
		switch {
		case (op == "==") || (op == "!="):
			return true
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_compilerOrderedComparisonType(typeName string) bool {
	numeric := selfhost_compiler_compiler_compilerNumericType(typeName)
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	return numeric || (base == "String" || base == "Char")
}

func selfhost_compiler_compiler_checkNumericUnaryOperand(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	actual := selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
	mismatch := actual != "" && selfhost_compiler_compiler_compilerNumericType(actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "operator '-' expects a numeric type, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkPostfixNumericOperand(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	complete := len(expr.children) > 0
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkPostfixNumericOperandType(expr.op, selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings), errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkPostfixNumericOperandType(op string, actual string, errors []string) []string {
	mismatch := actual != "" && selfhost_compiler_compiler_compilerNumericType(actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "operator '"+op+"' expects a numeric type, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkBitwiseUnaryOperand(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	actual := selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
	mismatch := actual != "" && selfhost_compiler_compiler_compilerBitwiseType(actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "operator '~' expects an integer type, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkNumericBinaryExpr(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case selfhost_compiler_compiler_compilerArithmeticOp(expr.op) == true:
			return selfhost_compiler_compiler_checkArithmeticOperands(expr, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkBitwiseBinaryExpr(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case selfhost_compiler_compiler_compilerBitwiseOp(expr.op) == true:
			return selfhost_compiler_compiler_checkBitwiseOperands(expr, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkArithmeticOperands(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	complete := len(expr.children) >= 2
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkArithmeticOperandTypes(expr, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkBitwiseOperands(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	complete := len(expr.children) >= 2
	return func() []string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_checkBitwiseOperandTypes(expr, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkArithmeticOperandTypes(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	left := selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
	right := selfhost_compiler_compiler_inferCompilerExprType(expr.children[1], callables, bindings)
	return func() []string {
		switch {
		case expr.op == "+":
			return selfhost_compiler_compiler_checkPlusOperandTypes(left, right, errors)
		default:
			return selfhost_compiler_compiler_checkNumericOperandTypes(expr.op, left, right, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkBitwiseOperandTypes(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	left := selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
	right := selfhost_compiler_compiler_inferCompilerExprType(expr.children[1], callables, bindings)
	checked := selfhost_compiler_compiler_checkBitwiseOperand(right, selfhost_compiler_compiler_checkBitwiseOperand(left, errors))
	withUnsigned := selfhost_compiler_compiler_checkUnsignedShiftLeftOperand(expr.op, left, checked)
	return selfhost_compiler_compiler_checkBitwiseOperandMatch(left, right, withUnsigned)
}

func selfhost_compiler_compiler_checkPlusOperandTypes(left string, right string, errors []string) []string {
	stringConcat := left == "String" || right == "String"
	return func() []string {
		switch {
		case stringConcat == true:
			return selfhost_compiler_compiler_checkStringConcatOperandTypes(left, right, errors)
		default:
			return selfhost_compiler_compiler_checkNumericOperandTypes("+", left, right, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkStringConcatOperandTypes(left string, right string, errors []string) []string {
	next := selfhost_compiler_compiler_checkStringConcatOperand(left, errors)
	return selfhost_compiler_compiler_checkStringConcatOperand(right, next)
}

func selfhost_compiler_compiler_checkStringConcatOperand(actual string, errors []string) []string {
	mismatch := actual != "" && actual != "String"
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "string concatenation expects String, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkBitwiseOperand(actual string, errors []string) []string {
	mismatch := actual != "" && selfhost_compiler_compiler_compilerBitwiseType(actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "bitwise operator expects integer operands, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkUnsignedShiftLeftOperand(op string, left string, errors []string) []string {
	invalid := op == ">>>" && left != "" && selfhost_compiler_compiler_compilerUnsignedIntegerType(left) == false
	return func() []string {
		switch {
		case invalid == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "operator '>>>' expects an unsigned integer left operand, got "+left)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkBitwiseOperandMatch(left string, right string, errors []string) []string {
	shouldCheck := left != "" && right != "" && (selfhost_compiler_compiler_compilerBitwiseType(left) && selfhost_compiler_compiler_compilerBitwiseType(right))
	mismatch := shouldCheck && left != right
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "bitwise operator requires matching integer types, got "+left+" and "+right)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkNumericOperandTypes(op string, left string, right string, errors []string) []string {
	checked := selfhost_compiler_compiler_checkNumericOperand(right, selfhost_compiler_compiler_checkNumericOperand(left, errors))
	return selfhost_compiler_compiler_checkNumericOperandMatch(op, left, right, checked)
}

func selfhost_compiler_compiler_checkNumericOperand(actual string, errors []string) []string {
	mismatch := actual != "" && selfhost_compiler_compiler_compilerNumericType(actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "arithmetic expects numeric operands, got "+actual)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkNumericOperandMatch(op string, left string, right string, errors []string) []string {
	shouldCheck := left != "" && right != "" && (selfhost_compiler_compiler_compilerNumericType(left) && selfhost_compiler_compiler_compilerNumericType(right))
	mismatch := shouldCheck && left != right
	withMatch := func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "arithmetic requires matching numeric types, got "+left+" and "+right)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
	return selfhost_compiler_compiler_checkModuloOperand(op, left, right, withMatch)
}

func selfhost_compiler_compiler_checkModuloOperand(op string, left string, right string, errors []string) []string {
	invalid := op == "%" && (left == "Double" || left == "Float" || (right == "Double" || right == "Float"))
	return func() []string {
		switch {
		case invalid == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "operator '%' expects integer operands, got "+left)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_compilerArithmeticOp(op string) bool {
	return func() bool {
		switch {
		case (op == "+") || (op == "-") || (op == "*") || (op == "/") || (op == "%"):
			return true
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_compilerBitwiseOp(op string) bool {
	return func() bool {
		switch {
		case (op == "&") || (op == "|") || (op == "^") || (op == "<<") || (op == ">>") || (op == ">>>"):
			return true
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_compilerNumericType(typeName string) bool {
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	return selfhost_compiler_compiler_compilerIntegerType(base) || base == "BigInt" || (base == "Double" || base == "Float")
}

func selfhost_compiler_compiler_compilerBitwiseType(typeName string) bool {
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	return selfhost_compiler_compiler_compilerIntegerType(base) || base == "BigInt"
}

func selfhost_compiler_compiler_compilerIntegerType(base string) bool {
	return selfhost_compiler_compiler_compilerSignedIntegerType(base) || selfhost_compiler_compiler_compilerUnsignedIntegerType(base)
}

func selfhost_compiler_compiler_compilerSignedIntegerType(base string) bool {
	return base == "Int" || base == "Int4" || (base == "Int8" || base == "Int16") || base == "Int64"
}

func selfhost_compiler_compiler_compilerUnsignedIntegerType(base string) bool {
	return base == "UInt" || base == "UInt8" || base == "UInt16" || base == "UInt64"
}

func selfhost_compiler_compiler_compilerSupportedMapKeyType(typeName string) bool {
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	return base == "String" || base == "Char" || (base == "Bool" || selfhost_compiler_compiler_compilerIntegerType(base)) || (base == "Double" || base == "Float")
}

func selfhost_compiler_compiler_compilerParseIntText(text string) int {
	return func() int {
		switch {
		case strings.HasPrefix(text, "-") == true:
			return 0 - selfhost_compiler_compiler_compilerParseUnsignedIntText(text, 1, 0)
		default:
			return selfhost_compiler_compiler_compilerParseUnsignedIntText(text, 0, 0)
		}
	}()
}

func selfhost_compiler_compiler_compilerParseUnsignedIntText(text string, index int, out int) int {
	done := index >= len([]rune(text))
	return func() int {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_compilerParseUnsignedIntText(text, index+1, out*10+selfhost_compiler_compiler_compilerDigitValue([]rune(text)[index]))
		}
	}()
}

func selfhost_compiler_compiler_compilerDigitValue(ch rune) int {
	return func() int {
		switch {
		case ch == '0':
			return 0
		case ch == '1':
			return 1
		case ch == '2':
			return 2
		case ch == '3':
			return 3
		case ch == '4':
			return 4
		case ch == '5':
			return 5
		case ch == '6':
			return 6
		case ch == '7':
			return 7
		case ch == '8':
			return 8
		case ch == '9':
			return 9
		default:
			return 0
		}
	}()
}

func selfhost_compiler_compiler_checkLetExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	hasValue := len(expr.children) > 0
	return func() []string {
		switch {
		case hasValue == true:
			return selfhost_compiler_compiler_checkLetValueExpr(expr, structs, callables, errors, bindings)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkLetValueExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	value := expr.children[0]
	checked := selfhost_compiler_compiler_checkExprExpected(value, expr.value, structs, callables, errors, bindings)
	return selfhost_compiler_compiler_checkLetDeclaredType(expr.name, value, expr.value, structs, callables, bindings, checked)
}

func selfhost_compiler_compiler_checkExprExpected(expr IRExpr, expected string, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkExpr(expr, structs, callables, errors, bindings)
	return selfhost_compiler_compiler_checkExpectedExprType(expr, expected, structs, callables, bindings, checked)
}

func selfhost_compiler_compiler_checkExpectedExprType(expr IRExpr, expected string, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	checked := func() []string {
		switch {
		case moduleCallKey(expr) == "json.parse":
			return selfhost_compiler_compiler_checkJsonParseTarget(expected, structs, errors)
		default:
			return errors
		}
	}()
	return checked
}

func selfhost_compiler_compiler_checkExpectedSelectorExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case expr.kind == ExprKind_Selector:
			return selfhost_compiler_compiler_checkSelectorExpr(expr, structs, callables, bindings, errors)
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_compiler_checkExpectedBlockSelectorExpr(expr.children, structs, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkExpectedBlockSelectorExpr(statements []IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	empty := len(statements) == 0
	return func() []string {
		switch {
		case empty == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkExpectedSelectorExpr(statements[len(statements)-1], structs, callables, bindings, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkSelectorExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	hasReceiver := len(expr.children) > 0
	return func() []string {
		switch {
		case hasReceiver == true:
			return selfhost_compiler_compiler_checkSelectorReceiverExpr(expr, expr.children[0], structs, callables, bindings, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkSelectorReceiverExpr(expr IRExpr, receiver IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case receiver.kind == ExprKind_At:
			return errors
		default:
			return selfhost_compiler_compiler_checkStructSelectorExpr(expr, selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(receiver, structs, callables, bindings), structs, bindings, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkStructSelectorExpr(expr IRExpr, receiverType string, structs []IRStructType, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case receiverType == "":
			return errors
		default:
			return func() []string {
				switch {
				case selfhost_compiler_compiler_compilerStructuralObjectFieldType(receiverType, expr.name) == "":
					return func() []string {
						typeDecl := selfhost_compiler_compiler_findCompilerStruct(structs, selfhost_compiler_compiler_compilerTypeBase(receiverType), 0)
						found := typeDecl.name != ""
						return func() []string {
							switch {
							case found == true:
								return func() []string {
									switch {
									case selfhost_compiler_compiler_compilerCanAccessPrivate(typeDecl.private, typeDecl.sourcePath, bindings) == true:
										return selfhost_compiler_compiler_checkStructSelectorField(expr, receiverType, typeDecl, bindings, errors)
									default:
										return func() []string {
											__rune_spread_out := []string{}
											__rune_spread_out = append(__rune_spread_out, errors...)
											__rune_spread_out = append(__rune_spread_out, "type \""+typeDecl.name+"\" is private")
											return __rune_spread_out
										}()
									}
								}()
							default:
								return func() []string {
									__rune_spread_out := []string{}
									__rune_spread_out = append(__rune_spread_out, errors...)
									__rune_spread_out = append(__rune_spread_out, "type "+receiverType+" has no fields")
									return __rune_spread_out
								}()
							}
						}()
					}()
				default:
					return errors
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerStructuralObjectFieldType(typeName string, fieldName string) string {
	valid := strings.HasPrefix(typeName, "{") && strings.HasSuffix(typeName, "}")
	return func() string {
		switch {
		case valid == true:
			return selfhost_compiler_compiler_compilerStructuralObjectFieldTypeInParts(func() []string {
				parts := strings.Split((func() string { runes := []rune(typeName); return string(runes[1 : len([]rune(typeName))-1]) }()), ";")
				return parts
			}(), fieldName, 0)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerStructuralObjectFieldTypeInParts(parts []string, fieldName string, index int) string {
	return func() string {
		if index >= len(parts) {
			return ""
		}
		return func() string {
			switch {
			case selfhost_compiler_compiler_compilerStructuralObjectFieldTypePart(parts[index], fieldName) == "":
				return selfhost_compiler_compiler_compilerStructuralObjectFieldTypeInParts(parts, fieldName, index+1)
			default:
				return selfhost_compiler_compiler_compilerStructuralObjectFieldTypePart(parts[index], fieldName)
			}
		}()
	}()
}

func selfhost_compiler_compiler_compilerStructuralObjectFieldTypePart(part string, fieldName string) string {
	pieces := func() []string { parts := strings.Split(part, ":"); return parts }()
	return func() string {
		if len(pieces) == 2 && pieces[0] == fieldName {
			return pieces[1]
		}
		return ""
	}()
}

func selfhost_compiler_compiler_checkStructSelectorField(expr IRExpr, receiverType string, typeDecl IRStructType, bindings []CompilerTypeBinding, errors []string) []string {
	field := selfhost_compiler_compiler_findCompilerStructField(typeDecl.fields, expr.name, 0)
	found := field.name != ""
	return func() []string {
		switch {
		case found == true:
			return selfhost_compiler_compiler_checkCompilerPrivateAccess("field", typeDecl.name+"."+field.name, field.private, typeDecl.sourcePath, bindings, errors)
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "type "+receiverType+" has no field \""+expr.name+"\"")
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkLetDeclaredType(name string, value IRExpr, expected string, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings)
	mismatch := expected != "?" && selfhost_compiler_compiler_compilerShouldCheckArgType(expected, actual) && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "binding \""+name+"\" has type "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkJsonParseTarget(expected string, structs []IRStructType, errors []string) []string {
	return func() []string {
		switch {
		case expected == "":
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "@json.parse target type cannot be inferred; add ': Type' to the binding")
				return __rune_spread_out
			}()
		case expected == "Dynamic":
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "@json.parse target type cannot be inferred; add ': Type' to the binding")
				return __rune_spread_out
			}()
		case expected == "Object":
			return errors
		default:
			return selfhost_compiler_compiler_checkJsonParseFromJsonTarget(expected, structs, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkJsonParseFromJsonTarget(expected string, structs []IRStructType, errors []string) []string {
	typeDecl := selfhost_compiler_compiler_findCompilerStruct(structs, selfhost_compiler_compiler_compilerTypeBase(expected), 0)
	ok := typeDecl.name != "" && selfhost_compiler_compiler_compilerStructHasFromJson(typeDecl, expected, 0)
	return func() []string {
		switch {
		case ok == true:
			return errors
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "@json.parse target type "+expected+" does not implement &FromJson")
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerStructHasFromJson(typeDecl IRStructType, expected string, index int) bool {
	done := index >= len(typeDecl.methods)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerStructHasFromJsonAt(typeDecl, expected, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerStructHasFromJsonAt(typeDecl IRStructType, expected string, index int) bool {
	method := typeDecl.methods[index]
	matched := method.name == "fromJson" && method.static && selfhost_compiler_compiler_compilerFromJsonSignatureMatches(method, expected)
	return func() bool {
		switch {
		case matched == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerStructHasFromJson(typeDecl, expected, index+1)
		}
	}()
}

func selfhost_compiler_compiler_compilerFromJsonSignatureMatches(method IRFunction, expected string) bool {
	arityOk := len(method.params) == 1
	return func() bool {
		switch {
		case arityOk == true:
			return method.params[0].typeName == "String" && selfhost_compiler_compiler_compilerTypesCompatible(expected, method.returnType)
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_checkBlockExpr(statements []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	done := index >= len(statements)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkBlockExprStep(statements, index, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkBlockExprStep(statements []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	statement := statements[index]
	nextErrors := selfhost_compiler_compiler_checkExpr(statement, structs, callables, errors, bindings)
	nextBindings := selfhost_compiler_compiler_blockBindingsAfterStatement(statement, structs, callables, bindings)
	return selfhost_compiler_compiler_checkBlockExpr(statements, index+1, structs, callables, nextErrors, nextBindings)
}

func selfhost_compiler_compiler_checkStructExpr(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	typeDecl := selfhost_compiler_compiler_findCompilerStruct(structs, expr.name, 0)
	found := typeDecl.name != ""
	return func() []string {
		switch {
		case found == true:
			return func() []string {
				switch {
				case selfhost_compiler_compiler_compilerCanAccessPrivate(typeDecl.private, typeDecl.sourcePath, bindings) == true:
					return selfhost_compiler_compiler_checkStructExprFields(expr, typeDecl, structs, callables, errors, bindings)
				default:
					return func() []string {
						__rune_spread_out := []string{}
						__rune_spread_out = append(__rune_spread_out, errors...)
						__rune_spread_out = append(__rune_spread_out, "type \""+expr.name+"\" is private")
						return __rune_spread_out
					}()
				}
			}()
		default:
			return selfhost_compiler_compiler_checkStructExprUnknownType(expr, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkStructExprUnknownType(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	next := func() []string {
		__rune_spread_out := []string{}
		__rune_spread_out = append(__rune_spread_out, errors...)
		__rune_spread_out = append(__rune_spread_out, "unknown type \""+expr.name+"\"")
		return __rune_spread_out
	}()
	for _, field := range expr.children {
		_ = field
		next = selfhost_compiler_compiler_checkExpr(field.children[0], structs, callables, next, bindings)
	}
	return next
}

func selfhost_compiler_compiler_checkStructExprFields(expr IRExpr, typeDecl IRStructType, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	checked := selfhost_compiler_compiler_checkStructExprFieldValues(expr.name, expr.children, typeDecl, 0, structs, callables, errors, bindings)
	return selfhost_compiler_compiler_checkStructMissingFields(expr.name, typeDecl.fields, expr.children, 0, checked)
}

func selfhost_compiler_compiler_checkStructExprFieldValues(typeName string, fields []IRExpr, typeDecl IRStructType, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	done := index >= len(fields)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkStructExprFieldValue(typeName, fields, typeDecl, index, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkStructExprFieldValue(typeName string, fields []IRExpr, typeDecl IRStructType, index int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	field := fields[index]
	value := field.children[0]
	expected := selfhost_compiler_compiler_findCompilerStructField(typeDecl.fields, field.name, 0)
	found := expected.name != ""
	next := func() []string {
		switch {
		case found == true:
			return func() []string {
				checkedField := selfhost_compiler_compiler_checkCompilerPrivateAccess("field", typeName+"."+field.name, expected.private, typeDecl.sourcePath, bindings, errors)
				checkedValue := selfhost_compiler_compiler_checkExprExpected(value, expected.typeName, structs, callables, checkedField, bindings)
				return selfhost_compiler_compiler_checkStructFieldType(typeName, field.name, value, expected.typeName, structs, callables, bindings, checkedValue)
			}()
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_checkExpr(value, structs, callables, errors, bindings)...)
				__rune_spread_out = append(__rune_spread_out, "type "+typeName+" has no field \""+field.name+"\"")
				return __rune_spread_out
			}()
		}
	}()
	return selfhost_compiler_compiler_checkStructExprFieldValues(typeName, fields, typeDecl, index+1, structs, callables, next, bindings)
}

func selfhost_compiler_compiler_checkStructFieldType(typeName string, fieldName string, value IRExpr, expected string, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings)
	mismatch := selfhost_compiler_compiler_compilerShouldCheckArgType(expected, actual) && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "field "+typeName+"."+fieldName+" has type "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkStructMissingFields(typeName string, expectedFields []IRField, fields []IRExpr, index int, errors []string) []string {
	done := index >= len(expectedFields)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkStructMissingField(typeName, expectedFields, fields, index, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkStructMissingField(typeName string, expectedFields []IRField, fields []IRExpr, index int, errors []string) []string {
	field := expectedFields[index]
	present := selfhost_compiler_compiler_compilerExprFieldContains(fields, field.name, 0)
	return func() []string {
		switch {
		case present == true:
			return selfhost_compiler_compiler_checkStructMissingFields(typeName, expectedFields, fields, index+1, errors)
		default:
			return selfhost_compiler_compiler_checkStructMissingFields(typeName, expectedFields, fields, index+1, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "missing field "+typeName+"."+field.name)
				return __rune_spread_out
			}())
		}
	}()
}

func selfhost_compiler_compiler_compilerExprFieldContains(fields []IRExpr, name string, index int) bool {
	done := index >= len(fields)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerExprFieldContainsAt(fields, name, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerExprFieldContainsAt(fields []IRExpr, name string, index int) bool {
	matched := fields[index].name == name
	return func() bool {
		switch {
		case matched == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerExprFieldContains(fields, name, index+1)
		}
	}()
}

func selfhost_compiler_compiler_blockBindingsAfterStatement(statement IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		switch {
		case statement.kind == ExprKind_Let:
			return selfhost_compiler_compiler_bindCompilerLet(statement, structs, callables, bindings)
		case statement.kind == ExprKind_ObjectDestructure:
			return selfhost_compiler_compiler_bindCompilerObjectDestructure(statement, structs, callables, bindings)
		default:
			return bindings
		}
	}()
}

func selfhost_compiler_compiler_bindCompilerLet(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	hasValue := len(expr.children) > 0
	return func() []CompilerTypeBinding {
		switch {
		case hasValue == true:
			return selfhost_compiler_compiler_addCompilerValueBinding(bindings, expr.name, selfhost_compiler_compiler_compilerLetBindingType(expr, structs, callables, bindings))
		default:
			return bindings
		}
	}()
}

func selfhost_compiler_compiler_compilerLetBindingType(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case (expr.value == "") || (expr.value == "?"):
			return selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(expr.children[0], structs, callables, bindings)
		default:
			return expr.value
		}
	}()
}

func selfhost_compiler_compiler_addCompilerTypeBinding(bindings []CompilerTypeBinding, name string, typeName string) []CompilerTypeBinding {
	known := typeName != ""
	return func() []CompilerTypeBinding {
		switch {
		case known == true:
			return func() []CompilerTypeBinding {
				__rune_spread_out := []CompilerTypeBinding{}
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_dropCompilerTypeBinding(bindings, name, 0, []CompilerTypeBinding{})...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerTypeBinding(name, typeName))
				return __rune_spread_out
			}()
		default:
			return bindings
		}
	}()
}

func selfhost_compiler_compiler_addCompilerValueBinding(bindings []CompilerTypeBinding, name string, typeName string) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		__rune_spread_out := []CompilerTypeBinding{}
		__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_dropCompilerTypeBinding(bindings, name, 0, []CompilerTypeBinding{})...)
		__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerTypeBinding(name, typeName))
		return __rune_spread_out
	}()
}

func selfhost_compiler_compiler_bindCompilerObjectDestructure(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	hasValue := len(expr.children) > 0
	return func() []CompilerTypeBinding {
		switch {
		case hasValue == true:
			return selfhost_compiler_compiler_bindCompilerObjectDestructureParams(expr.params, 0, selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(expr.children[0], structs, callables, bindings), structs, bindings)
		default:
			return bindings
		}
	}()
}

func selfhost_compiler_compiler_bindCompilerObjectDestructureParams(params []IRParam, index int, sourceType string, structs []IRStructType, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	done := index >= len(params)
	return func() []CompilerTypeBinding {
		switch {
		case done == true:
			return bindings
		default:
			return selfhost_compiler_compiler_bindCompilerObjectDestructureParams(params, index+1, sourceType, structs, selfhost_compiler_compiler_addCompilerValueBinding(bindings, params[index].name, selfhost_compiler_compiler_compilerObjectDestructureFieldType(params[index], sourceType, structs)))
		}
	}()
}

func selfhost_compiler_compiler_compilerObjectDestructureFieldType(param IRParam, sourceType string, structs []IRStructType) string {
	return func() string {
		switch {
		case sourceType == "":
			return ""
		default:
			return func() string {
				typeDecl := selfhost_compiler_compiler_findCompilerStruct(structs, selfhost_compiler_compiler_compilerTypeBase(sourceType), 0)
				field := selfhost_compiler_compiler_findCompilerStructField(typeDecl.fields, param.typeName, 0)
				return field.typeName
			}()
		}
	}()
}

func selfhost_compiler_compiler_dropCompilerTypeBinding(bindings []CompilerTypeBinding, name string, index int, out []CompilerTypeBinding) []CompilerTypeBinding {
	done := index >= len(bindings)
	return func() []CompilerTypeBinding {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_dropCompilerTypeBindingStep(bindings, name, index, out)
		}
	}()
}

func selfhost_compiler_compiler_dropCompilerTypeBindingStep(bindings []CompilerTypeBinding, name string, index int, out []CompilerTypeBinding) []CompilerTypeBinding {
	matched := bindings[index].name == name
	return func() []CompilerTypeBinding {
		switch {
		case matched == true:
			return selfhost_compiler_compiler_dropCompilerTypeBinding(bindings, name, index+1, out)
		default:
			return selfhost_compiler_compiler_dropCompilerTypeBinding(bindings, name, index+1, func() []CompilerTypeBinding {
				__rune_spread_out := []CompilerTypeBinding{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, bindings[index])
				return __rune_spread_out
			}())
		}
	}()
}

func selfhost_compiler_compiler_checkExprCall(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	identifierCall := expr.kind == ExprKind_Call && len(expr.children) > 0 && expr.children[0].kind == ExprKind_Identifier
	return func() []string {
		switch {
		case identifierCall == true:
			return selfhost_compiler_compiler_checkIdentifierCall(expr, expr.children[0].name, len(expr.children)-1, structs, callables, errors, bindings)
		default:
			return selfhost_compiler_compiler_checkSelectorCall(expr, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkSelectorCall(expr IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	selectorCall := expr.kind == ExprKind_Call && len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector
	return func() []string {
		switch {
		case selectorCall == true:
			return selfhost_compiler_compiler_checkSelectorCallExpr(expr, expr.children[0], structs, callables, errors, bindings)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkSelectorCallExpr(expr IRExpr, selector IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	hasReceiver := len(selector.children) > 0
	return func() []string {
		switch {
		case hasReceiver == true:
			return selfhost_compiler_compiler_checkSelectorCallReceiver(expr, selector, selector.children[0], structs, callables, errors, bindings)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkSelectorCallReceiver(expr IRExpr, selector IRExpr, receiver IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case receiver.kind == ExprKind_At:
			return selfhost_compiler_compiler_checkAtSelectorCall(expr, selector, receiver, structs, callables, errors, bindings)
		case receiver.kind == ExprKind_Identifier:
			return selfhost_compiler_compiler_checkIdentifierSelectorCall(expr, selector, receiver, structs, callables, errors, bindings)
		default:
			return selfhost_compiler_compiler_checkInstanceSelectorCall(expr, selector, receiver, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkAtSelectorCall(expr IRExpr, selector IRExpr, receiver IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	importPath := compilerIRAtImportPath(receiver)
	return func() []string {
		switch {
		case importPath == "":
			return selfhost_compiler_compiler_checkModuleSelectorCall(expr, selector, receiver, errors)
		default:
			return func() []string {
				if compilerGoPackageImportPath(importPath) != "" {
					return errors
				}
				return selfhost_compiler_compiler_checkIdentifierCall(expr, selector.name, len(expr.children)-1, structs, callables, errors, bindings)
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkModuleSelectorCall(expr IRExpr, selector IRExpr, receiver IRExpr, errors []string) []string {
	return func() []string {
		switch {
		case receiver.name == "go":
			return selfhost_compiler_compiler_checkGoModuleSelectorCall(expr, selector.name, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkGoModuleSelectorCall(expr IRExpr, name string, errors []string) []string {
	return func() []string {
		switch {
		case name == "expr":
			return selfhost_compiler_compiler_checkGoExprCall(expr, errors)
		case name == "stmt":
			return selfhost_compiler_compiler_checkGoStmtCall(expr, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkGoExprCall(expr IRExpr, errors []string) []string {
	return func() []string {
		match7 := selfhost_compiler_compiler_compilerCallArgCount(expr)
		switch {
		case match7 == 1:
			return selfhost_compiler_compiler_checkGoExprStringLiteral(expr, errors)
		case true:
			count := match7
			_ = count
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "@go.expr expects 1 args, got "+strconv.Itoa(count))
				return __rune_spread_out
			}()
		}
		return nil
	}()
}

func selfhost_compiler_compiler_checkGoStmtCall(expr IRExpr, errors []string) []string {
	return func() []string {
		match8 := selfhost_compiler_compiler_compilerCallArgCount(expr)
		switch {
		case match8 == 1:
			return selfhost_compiler_compiler_checkGoStmtStringLiteral(expr, errors)
		case true:
			count := match8
			_ = count
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "@go.stmt expects 1 args, got "+strconv.Itoa(count))
				return __rune_spread_out
			}()
		}
		return nil
	}()
}

func selfhost_compiler_compiler_checkGoExprStringLiteral(expr IRExpr, errors []string) []string {
	return func() []string {
		switch {
		case expr.children[1].kind == ExprKind_String:
			return errors
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "@go.expr body must be a string literal")
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkGoStmtStringLiteral(expr IRExpr, errors []string) []string {
	return func() []string {
		switch {
		case expr.children[1].kind == ExprKind_String:
			return errors
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "@go.stmt argument must be a string literal")
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerCallArgCount(expr IRExpr) int {
	return len(expr.children) - 1
}

func selfhost_compiler_compiler_checkIdentifierSelectorCall(expr IRExpr, selector IRExpr, receiver IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	return func() []string {
		switch {
		case selector.op == "::":
			return selfhost_compiler_compiler_checkStaticSelectorCall(expr, selector, receiver.name, structs, callables, errors, bindings)
		default:
			return selfhost_compiler_compiler_checkInstanceSelectorCall(expr, selector, receiver, structs, callables, errors, bindings)
		}
	}()
}

func selfhost_compiler_compiler_checkStaticSelectorCall(expr IRExpr, selector IRExpr, receiverType string, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	name := selfhost_compiler_compiler_compilerStaticMethodName(receiverType, selector.name)
	callable := selfhost_compiler_compiler_findCompilerCallable(callables, name, 0)
	found := callable.name != ""
	return func() []string {
		switch {
		case found == true:
			return func() []string {
				switch {
				case selfhost_compiler_compiler_checkCallableVisibility(callable, bindings) == true:
					return selfhost_compiler_compiler_checkCallableCall(callable, expr, len(expr.children)-1, structs, callables, bindings, errors)
				default:
					return func() []string {
						__rune_spread_out := []string{}
						__rune_spread_out = append(__rune_spread_out, errors...)
						__rune_spread_out = append(__rune_spread_out, "static method \""+name+"\" is private")
						return __rune_spread_out
					}()
				}
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkInstanceSelectorCall(expr IRExpr, selector IRExpr, receiver IRExpr, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	receiverType := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(receiver, structs, callables, bindings)
	known := receiverType != ""
	return func() []string {
		switch {
		case known == true:
			return selfhost_compiler_compiler_checkKnownInstanceSelectorCall(expr, selector, selfhost_compiler_compiler_compilerTypeBase(receiverType), structs, callables, errors, bindings)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkKnownInstanceSelectorCall(expr IRExpr, selector IRExpr, receiverType string, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	name := selfhost_compiler_compiler_compilerInstanceMethodName(receiverType, selector.name)
	callable := selfhost_compiler_compiler_findCompilerCallable(callables, name, 0)
	found := callable.name != ""
	return func() []string {
		switch {
		case found == true:
			return func() []string {
				switch {
				case selfhost_compiler_compiler_checkCallableVisibility(callable, bindings) == true:
					return selfhost_compiler_compiler_checkCallableCall(callable, expr, len(expr.children)-1, structs, callables, bindings, errors)
				default:
					return func() []string {
						__rune_spread_out := []string{}
						__rune_spread_out = append(__rune_spread_out, errors...)
						__rune_spread_out = append(__rune_spread_out, "method \""+name+"\" is private")
						return __rune_spread_out
					}()
				}
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_checkIdentifierCall(expr IRExpr, name string, arity int, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	callable := selfhost_compiler_compiler_findCompilerCallable(callables, name, 0)
	found := callable.name != ""
	return func() []string {
		switch {
		case found == true:
			return func() []string {
				switch {
				case selfhost_compiler_compiler_checkCallableVisibility(callable, bindings) == true:
					return selfhost_compiler_compiler_checkCallableCall(callable, expr, arity, structs, callables, bindings, errors)
				default:
					return func() []string {
						__rune_spread_out := []string{}
						__rune_spread_out = append(__rune_spread_out, errors...)
						__rune_spread_out = append(__rune_spread_out, "function \""+name+"\" is private")
						return __rune_spread_out
					}()
				}
			}()
		default:
			return selfhost_compiler_compiler_checkUndefinedIdentifierCall(name, bindings, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkUndefinedIdentifierCall(name string, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case selfhost_compiler_compiler_findCompilerTypeBinding(bindings, name, 0).typeName == "MacroFunction":
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, name+" is a macro and can only be used with '#'")
				return __rune_spread_out
			}()
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "undefined function "+name)
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkCallableVisibility(callable CompilerCallable, bindings []CompilerTypeBinding) bool {
	return selfhost_compiler_compiler_compilerCanAccessPrivate(callable.private, callable.sourcePath, bindings)
}

func selfhost_compiler_compiler_checkCompilerPrivateAccess(kind string, name string, private bool, sourcePath string, bindings []CompilerTypeBinding, errors []string) []string {
	return func() []string {
		switch {
		case selfhost_compiler_compiler_compilerCanAccessPrivate(private, sourcePath, bindings) == true:
			return errors
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, kind+" \""+name+"\" is private")
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerCanAccessPrivate(private bool, sourcePath string, bindings []CompilerTypeBinding) bool {
	return func() bool {
		switch {
		case private == false:
			return true
		default:
			return func() bool {
				current := selfhost_compiler_compiler_compilerCurrentSourcePath(bindings)
				return sourcePath == "" || (current == "" || current == sourcePath)
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerCurrentSourcePath(bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_findCompilerTypeBinding(bindings, "__sourcePath", 0).typeName
}

func selfhost_compiler_compiler_checkCallableCall(callable CompilerCallable, expr IRExpr, arity int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string) []string {
	arityOk := callable.arity == arity
	return func() []string {
		switch {
		case arityOk == true:
			return selfhost_compiler_compiler_checkCallableArgTypes(callable, expr, structs, callables, bindings, errors, 0)
		default:
			return selfhost_compiler_compiler_checkCallableArity(callable, arity, errors)
		}
	}()
}

func selfhost_compiler_compiler_checkCallableArity(callable CompilerCallable, arity int, errors []string) []string {
	ok := callable.arity == arity
	return func() []string {
		switch {
		case ok == true:
			return errors
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "function \""+callable.name+"\" expects "+compilerIntToString(callable.arity)+" args, got "+compilerIntToString(arity))
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_checkCallableArgTypes(callable CompilerCallable, expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string, index int) []string {
	done := index >= len(callable.paramTypes)
	return func() []string {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_checkCallableArgType(callable, expr, structs, callables, bindings, errors, index)
		}
	}()
}

func selfhost_compiler_compiler_checkCallableArgType(callable CompilerCallable, expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, errors []string, index int) []string {
	expected := callable.paramTypes[index]
	arg := expr.children[index+1]
	checked := selfhost_compiler_compiler_checkExpectedExprType(arg, expected, structs, callables, bindings, errors)
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(arg, structs, callables, bindings)
	mismatch := selfhost_compiler_compiler_compilerShouldCheckArgType(expected, actual) && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	next := func() []string {
		switch {
		case mismatch == true:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, checked...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerArgumentTypeError(callable.name, index+1, actual, expected))
				return __rune_spread_out
			}()
		default:
			return checked
		}
	}()
	return selfhost_compiler_compiler_checkCallableArgTypes(callable, expr, structs, callables, bindings, next, index+1)
}

func selfhost_compiler_compiler_compilerArgumentTypeError(name string, index int, actual string, expected string) string {
	return "argument " + compilerIntToString(index) + " to \"" + name + "\" has type " + actual + ", expected " + expected
}

func selfhost_compiler_compiler_compilerShouldCheckArgType(expected string, actual string) bool {
	return expected != "" && (expected != "Dynamic" && actual != "") && selfhost_compiler_compiler_compilerTypeIsGenericPlaceholder(expected) == false
}

func selfhost_compiler_compiler_compilerTypeIsGenericPlaceholder(typeName string) bool {
	return len([]rune(typeName)) == 1 && ([]rune(typeName)[0] >= 'A' && []rune(typeName)[0] <= 'Z')
}

func selfhost_compiler_compiler_checkFunctionReturn(fn IRFunction, structs []IRStructType, callables []CompilerCallable, errors []string, bindings []CompilerTypeBinding) []string {
	expected := fn.returnType
	checked := selfhost_compiler_compiler_checkExpectedExprType(fn.body, expected, structs, callables, bindings, errors)
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(fn.body, structs, callables, bindings)
	shouldCheck := expected != "" && expected != "Dynamic" && actual != ""
	return func() []string {
		switch {
		case shouldCheck == true:
			return func() []string {
				if expected == "WebComponent" && actual == "HTMLElement" && selfhost_compiler_compiler_compilerExprCanBuildWebComponent(fn.body) {
					return checked
				}
				return selfhost_compiler_compiler_checkFunctionReturnType(fn.name, expected, actual, checked)
			}()
		default:
			return checked
		}
	}()
}

func selfhost_compiler_compiler_compilerExprCanBuildWebComponent(expr IRExpr) bool {
	return func() bool {
		switch {
		case expr.kind == ExprKind_XMLElement:
			return true
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_compiler_compilerLastExprCanBuildWebComponent(expr.children)
		case expr.kind == ExprKind_Ternary:
			return len(expr.children) >= 3 && selfhost_compiler_compiler_compilerExprCanBuildWebComponent(expr.children[1]) && selfhost_compiler_compiler_compilerExprCanBuildWebComponent(expr.children[2])
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_compilerLastExprCanBuildWebComponent(exprs []IRExpr) bool {
	return func() bool {
		if len(exprs) == 0 {
			return false
		}
		return selfhost_compiler_compiler_compilerExprCanBuildWebComponent(exprs[len(exprs)-1])
	}()
}

func selfhost_compiler_compiler_checkFunctionReturnType(name string, expected string, actual string, errors []string) []string {
	ok := selfhost_compiler_compiler_compilerTypesCompatible(expected, actual)
	return func() []string {
		switch {
		case ok == true:
			return errors
		default:
			return func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, "function \""+name+"\" returns "+actual+", expected "+expected)
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerTypesCompatible(expected string, actual string) bool {
	nullable := strings.HasSuffix(expected, "?")
	return func() bool {
		switch {
		case nullable == true:
			return actual == "Null" || selfhost_compiler_compiler_compilerTypesCompatible(func() string { runes := []rune(expected); return string(runes[0 : len([]rune(expected))-1]) }(), actual)
		default:
			return selfhost_compiler_compiler_compilerNonNullableTypesCompatible(expected, actual)
		}
	}()
}

func selfhost_compiler_compiler_compilerNonNullableTypesCompatible(expected string, actual string) bool {
	same := expected == actual
	return func() bool {
		switch {
		case same == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerGenericTypesCompatible(expected, actual)
		}
	}()
}

func selfhost_compiler_compiler_compilerGenericTypesCompatible(expected string, actual string) bool {
	expectedBase := selfhost_compiler_compiler_compilerTypeBase(expected)
	actualBase := selfhost_compiler_compiler_compilerTypeBase(actual)
	sameBase := expectedBase == actualBase
	return func() bool {
		switch {
		case sameBase == true:
			return selfhost_compiler_compiler_compilerGenericTypeArgsCompatible(expected, actual)
		default:
			return expectedBase == actual
		}
	}()
}

func selfhost_compiler_compiler_compilerGenericTypeArgsCompatible(expected string, actual string) bool {
	expectedArgs := selfhost_compiler_compiler_compilerGenericArgs(expected)
	actualArgs := selfhost_compiler_compiler_compilerGenericArgs(actual)
	hasExpected := len(expectedArgs) > 0
	hasActual := len(actualArgs) > 0
	return func() bool {
		switch {
		case hasExpected == true:
			return func() bool {
				switch {
				case hasActual == true:
					return selfhost_compiler_compiler_compilerGenericArgListsCompatible(expectedArgs, actualArgs, 0)
				default:
					return true
				}
			}()
		default:
			return true
		}
	}()
}

func selfhost_compiler_compiler_compilerGenericArgListsCompatible(expected []string, actual []string, index int) bool {
	lengthMismatch := len(expected) != len(actual)
	return func() bool {
		switch {
		case lengthMismatch == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerGenericArgListsCompatibleAt(expected, actual, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerGenericArgListsCompatibleAt(expected []string, actual []string, index int) bool {
	done := index >= len(expected)
	return func() bool {
		switch {
		case done == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerTypesCompatible(strings.TrimSpace(expected[index]), strings.TrimSpace(actual[index])) && selfhost_compiler_compiler_compilerGenericArgListsCompatibleAt(expected, actual, index+1)
		}
	}()
}

func selfhost_compiler_compiler_compilerTypesComparable(left string, right string) bool {
	unknown := left == "" || right == ""
	return func() bool {
		switch {
		case unknown == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerKnownTypesComparable(left, right)
		}
	}()
}

func selfhost_compiler_compiler_compilerKnownTypesComparable(left string, right string) bool {
	same := left == right
	return func() bool {
		switch {
		case same == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerNullableTypesComparable(left, right)
		}
	}()
}

func selfhost_compiler_compiler_compilerNullableTypesComparable(left string, right string) bool {
	leftNullable := strings.HasSuffix(left, "?")
	return func() bool {
		switch {
		case leftNullable == true:
			return selfhost_compiler_compiler_compilerLeftNullableComparable(left, right)
		default:
			return selfhost_compiler_compiler_compilerRightNullableComparable(left, right)
		}
	}()
}

func selfhost_compiler_compiler_compilerLeftNullableComparable(left string, right string) bool {
	rightNull := right == "Null"
	return func() bool {
		switch {
		case rightNull == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerTypesComparable(func() string { runes := []rune(left); return string(runes[0 : len([]rune(left))-1]) }(), right)
		}
	}()
}

func selfhost_compiler_compiler_compilerRightNullableComparable(left string, right string) bool {
	rightNullable := strings.HasSuffix(right, "?")
	return func() bool {
		switch {
		case rightNullable == true:
			return selfhost_compiler_compiler_compilerRightNullableInnerComparable(left, right)
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_compilerRightNullableInnerComparable(left string, right string) bool {
	leftNull := left == "Null"
	return func() bool {
		switch {
		case leftNull == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerTypesComparable(left, func() string { runes := []rune(right); return string(runes[0 : len([]rune(right))-1]) }())
		}
	}()
}

func selfhost_compiler_compiler_compilerTypeBase(typeName string) string {
	signal := strings.HasPrefix(typeName, "$")
	base := func() string {
		if signal {
			return func() string { runes := []rune(typeName); return string(runes[1:len([]rune(typeName))]) }()
		}
		return typeName
	}()
	generic := strings.Index(base, "[")
	return func() string {
		if generic < 0 {
			return base
		}
		return func() string { runes := []rune(base); return string(runes[0:generic]) }()
	}()
}

func selfhost_compiler_compiler_compilerGenericArgs(typeName string) []string {
	inner := selfhost_compiler_compiler_compilerGenericInner(typeName)
	return func() []string {
		switch {
		case inner == "":
			return []string{}
		default:
			return func() []string { parts := strings.Split(inner, ","); return parts }()
		}
	}()
}

func selfhost_compiler_compiler_compilerArrayType(elem string) string {
	return func() string {
		switch {
		case elem == "":
			return "Array"
		default:
			return "Array[" + elem + "]"
		}
	}()
}

func selfhost_compiler_compiler_compilerMapType(key string, value string) string {
	complete := key != "" && value != ""
	return func() string {
		switch {
		case complete == true:
			return "Map[" + key + ", " + value + "]"
		default:
			return "Map"
		}
	}()
}

func selfhost_compiler_compiler_compilerTupleLiteralType(elements []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return "Tuple[" + selfhost_compiler_compiler_compilerTupleLiteralTypes(elements, callables, bindings, 0, "") + "]"
}

func selfhost_compiler_compiler_compilerTupleLiteralTypes(elements []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, index int, out string) string {
	done := index >= len(elements)
	return func() string {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_compilerTupleLiteralTypes(elements, callables, bindings, index+1, out+selfhost_compiler_compiler_compilerTupleTypeSeparator(index)+selfhost_compiler_compiler_inferCompilerExprType(elements[index], callables, bindings))
		}
	}()
}

func selfhost_compiler_compiler_compilerTupleTypeSeparator(index int) string {
	return func() string {
		switch {
		case index == 0:
			return ""
		default:
			return ","
		}
	}()
}

func selfhost_compiler_compiler_compilerNullableType(inner string) string {
	return func() string {
		switch {
		case inner == "":
			return ""
		case inner == "Null":
			return "Null"
		default:
			return func() string {
				switch {
				case strings.HasSuffix(inner, "?") == true:
					return inner
				default:
					return inner + "?"
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerArrayElementType(typeName string) string {
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	args := selfhost_compiler_compiler_compilerGenericArgs(typeName)
	valid := base == "Array" && len(args) == 1
	return func() string {
		switch {
		case valid == true:
			return strings.TrimSpace(args[0])
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerMapKeyType(typeName string) string {
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	args := selfhost_compiler_compiler_compilerGenericArgs(typeName)
	valid := base == "Map" && len(args) == 2
	return func() string {
		switch {
		case valid == true:
			return strings.TrimSpace(args[0])
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerMapValueType(typeName string) string {
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	args := selfhost_compiler_compiler_compilerGenericArgs(typeName)
	valid := base == "Map" && len(args) == 2
	return func() string {
		switch {
		case valid == true:
			return strings.TrimSpace(args[1])
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerTupleElementTypes(typeName string) []string {
	base := selfhost_compiler_compiler_compilerTypeBase(typeName)
	args := selfhost_compiler_compiler_compilerGenericArgs(typeName)
	valid := (base == "Tuple" || base == "ReadonlyTuple") && len(args) > 0
	return func() []string {
		switch {
		case valid == true:
			return selfhost_compiler_compiler_compilerTrimTypes(args, 0, []string{})
		default:
			return []string{}
		}
	}()
}

func selfhost_compiler_compiler_compilerTrimTypes(types []string, index int, out []string) []string {
	done := index >= len(types)
	return func() []string {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_compilerTrimTypes(types, index+1, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, strings.TrimSpace(types[index]))
				return __rune_spread_out
			}())
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerExprType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Identifier:
			return selfhost_compiler_compiler_inferCompilerIdentifierType(expr, bindings)
		case expr.kind == ExprKind_This:
			return selfhost_compiler_compiler_findCompilerTypeBinding(bindings, "this", 0).typeName
		case expr.kind == ExprKind_Selector:
			return selfhost_compiler_compiler_inferCompilerSelectorType(expr, callables, bindings)
		case expr.kind == ExprKind_String:
			return "String"
		case expr.kind == ExprKind_Template:
			return "String"
		case expr.kind == ExprKind_XMLText:
			return "String"
		case expr.kind == ExprKind_XMLElement:
			return "HTMLElement"
		case expr.kind == ExprKind_Int:
			return "Int"
		case expr.kind == ExprKind_Double:
			return "Double"
		case expr.kind == ExprKind_BigInt:
			return "BigInt"
		case expr.kind == ExprKind_Char:
			return "Char"
		case expr.kind == ExprKind_Bool:
			return "Bool"
		case expr.kind == ExprKind_Null:
			return "Null"
		case expr.kind == ExprKind_Struct:
			return expr.name
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_compiler_inferCompilerCallType(expr, callables, bindings)
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_compiler_inferCompilerBlockType(expr, callables, bindings)
		case expr.kind == ExprKind_Ternary:
			return selfhost_compiler_compiler_inferCompilerTernaryType(expr, callables, bindings)
		case expr.kind == ExprKind_Unary:
			return selfhost_compiler_compiler_inferCompilerUnaryType(expr, callables, bindings)
		case expr.kind == ExprKind_Postfix:
			return selfhost_compiler_compiler_inferCompilerPostfixType(expr, callables, bindings)
		case expr.kind == ExprKind_Binary:
			return selfhost_compiler_compiler_inferCompilerBinaryType(expr, callables, bindings)
		case expr.kind == ExprKind_Array:
			return selfhost_compiler_compiler_inferCompilerArrayType(expr, callables, bindings)
		case expr.kind == ExprKind_Tuple:
			return selfhost_compiler_compiler_inferCompilerTupleType(expr, callables, bindings)
		case expr.kind == ExprKind_Map:
			return selfhost_compiler_compiler_inferCompilerMapType(expr, callables, bindings)
		case expr.kind == ExprKind_Spread:
			return selfhost_compiler_compiler_inferCompilerSpreadType(expr, callables, bindings)
		case expr.kind == ExprKind_Index:
			return selfhost_compiler_compiler_inferCompilerIndexType(expr, callables, bindings)
		case expr.kind == ExprKind_PatternBlock:
			return selfhost_compiler_compiler_inferCompilerPatternBlockType(expr, callables, bindings)
		case expr.kind == ExprKind_Match:
			return selfhost_compiler_compiler_inferCompilerMatchType(expr, callables, bindings)
		case expr.kind == ExprKind_Unwrap:
			return selfhost_compiler_compiler_inferCompilerUnwrapType(expr, callables, bindings)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerIdentifierType(expr IRExpr, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_compilerIdentifierBinding(expr, bindings).typeName
}

func selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Selector:
			return selfhost_compiler_compiler_inferCompilerSelectorTypeWithStructs(expr, structs, callables, bindings)
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_compiler_inferCompilerBlockTypeWithStructs(expr, structs, callables, bindings)
		case expr.kind == ExprKind_PatternBlock:
			return selfhost_compiler_compiler_inferCompilerPatternBlockTypeWithStructs(expr, structs, callables, bindings)
		case expr.kind == ExprKind_Match:
			return selfhost_compiler_compiler_inferCompilerMatchTypeWithStructs(expr, structs, callables, bindings)
		case expr.kind == ExprKind_Unwrap:
			return selfhost_compiler_compiler_inferCompilerUnwrapTypeWithStructs(expr, structs, callables, bindings)
		default:
			return selfhost_compiler_compiler_inferCompilerExprType(expr, callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerSelectorTypeWithStructs(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	hasReceiver := len(expr.children) > 0
	return func() string {
		switch {
		case hasReceiver == true:
			return selfhost_compiler_compiler_inferCompilerSelectorTypeFromReceiverWithStructs(expr, expr.children[0], structs, callables, bindings)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerSelectorTypeFromReceiverWithStructs(expr IRExpr, receiver IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case receiver.kind == ExprKind_At:
			return selfhost_compiler_compiler_inferCompilerSelectorTypeFromReceiver(expr, receiver, callables, bindings)
		default:
			return selfhost_compiler_compiler_inferCompilerStructSelectorType(expr, selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(receiver, structs, callables, bindings), structs)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerStructSelectorType(expr IRExpr, receiverType string, structs []IRStructType) string {
	return func() string {
		switch {
		case receiverType == "":
			return ""
		default:
			return func() string {
				typeDecl := selfhost_compiler_compiler_findCompilerStruct(structs, selfhost_compiler_compiler_compilerTypeBase(receiverType), 0)
				field := selfhost_compiler_compiler_findCompilerStructField(typeDecl.fields, expr.name, 0)
				return field.typeName
			}()
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerArrayType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_compilerArrayType(selfhost_compiler_compiler_inferCompilerArrayElementType(expr.children, callables, bindings, 0, ""))
}

func selfhost_compiler_compiler_inferCompilerArrayElementType(elements []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, index int, expected string) string {
	done := index >= len(elements)
	return func() string {
		switch {
		case done == true:
			return expected
		default:
			return selfhost_compiler_compiler_inferCompilerArrayElementTypeAt(elements, callables, bindings, index, expected)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerArrayElementTypeAt(elements []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, index int, expected string) string {
	actual := selfhost_compiler_compiler_inferCompilerArrayLiteralElementType(elements[index], callables, bindings)
	next := selfhost_compiler_compiler_compilerNextArrayElementType(expected, actual)
	return selfhost_compiler_compiler_inferCompilerArrayElementType(elements, callables, bindings, index+1, next)
}

func selfhost_compiler_compiler_inferCompilerArrayLiteralElementType(elem IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case elem.kind == ExprKind_Spread:
			return selfhost_compiler_compiler_compilerArrayElementType(selfhost_compiler_compiler_inferCompilerExprType(elem.children[0], callables, bindings))
		default:
			return selfhost_compiler_compiler_inferCompilerExprType(elem, callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerTupleType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_compilerTupleLiteralType(expr.children, callables, bindings)
}

func selfhost_compiler_compiler_inferCompilerPostfixType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	complete := len(expr.children) > 0
	return func() string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerMapType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_compilerMapType(selfhost_compiler_compiler_inferCompilerMapKeyType(expr.children, callables, bindings, 0, ""), selfhost_compiler_compiler_inferCompilerMapValueType(expr.children, callables, bindings, 0, ""))
}

func selfhost_compiler_compiler_inferCompilerMapKeyType(entries []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, index int, expected string) string {
	done := index >= len(entries)
	return func() string {
		switch {
		case done == true:
			return expected
		default:
			return selfhost_compiler_compiler_inferCompilerMapKeyType(entries, callables, bindings, index+1, selfhost_compiler_compiler_compilerNextMapEntryType(expected, selfhost_compiler_compiler_inferCompilerMapEntryKeyType(entries[index], callables, bindings)))
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerMapValueType(entries []IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding, index int, expected string) string {
	done := index >= len(entries)
	return func() string {
		switch {
		case done == true:
			return expected
		default:
			return selfhost_compiler_compiler_inferCompilerMapValueType(entries, callables, bindings, index+1, selfhost_compiler_compiler_compilerNextMapEntryType(expected, selfhost_compiler_compiler_inferCompilerMapEntryValueType(entries[index], callables, bindings)))
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerMapEntryKeyType(entry IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	complete := len(entry.children) > 0
	return func() string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_inferCompilerExprType(entry.children[0], callables, bindings)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerMapEntryValueType(entry IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	complete := len(entry.children) > 1
	return func() string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_inferCompilerExprType(entry.children[1], callables, bindings)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerSpreadType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	complete := len(expr.children) > 0
	return func() string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerIndexType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	complete := len(expr.children) >= 2
	return func() string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_inferCompilerIndexReceiverType(selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings), expr.children[1])
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerPatternBlockType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_inferCompilerPatternBlockTypeWithStructs(expr, []IRStructType{}, callables, bindings)
}

func selfhost_compiler_compiler_inferCompilerMatchType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_inferCompilerMatchTypeWithStructs(expr, []IRStructType{}, callables, bindings)
}

func selfhost_compiler_compiler_inferCompilerMatchTypeWithStructs(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_inferCompilerPatternBranchTypesWithStructs(expr.children, 1, structs, callables, bindings, "")
}

func selfhost_compiler_compiler_inferCompilerUnwrapType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_inferCompilerUnwrapTypeWithStructs(expr, []IRStructType{}, callables, bindings)
}

func selfhost_compiler_compiler_inferCompilerUnwrapTypeWithStructs(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	complete := len(expr.children) > 0
	return func() string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_compilerResultOkType(selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(expr.children[0], structs, callables, bindings))
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerResultOkType(sourceType string) string {
	base := selfhost_compiler_compiler_compilerTypeBase(sourceType)
	args := selfhost_compiler_compiler_compilerGenericArgs(sourceType)
	valid := base == "Result" && len(args) == 2
	return func() string {
		switch {
		case valid == true:
			return strings.TrimSpace(args[0])
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerPatternBlockTypeWithStructs(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_inferCompilerPatternBranchTypesWithStructs(expr.children, 0, structs, callables, bindings, "")
}

func selfhost_compiler_compiler_inferCompilerPatternBranchTypesWithStructs(branches []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, expected string) string {
	done := index >= len(branches)
	return func() string {
		switch {
		case done == true:
			return expected
		default:
			return selfhost_compiler_compiler_inferCompilerPatternBranchTypeWithStructs(branches, index, structs, callables, bindings, expected)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerPatternBranchTypeWithStructs(branches []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, expected string) string {
	branch := branches[index]
	complete := len(branch.children) >= 2
	return func() string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_inferCompilerPatternBranchValueTypeWithStructs(branches, index, branch.children[1], structs, callables, bindings, expected)
		default:
			return selfhost_compiler_compiler_inferCompilerPatternBranchTypesWithStructs(branches, index+1, structs, callables, bindings, expected)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerPatternBranchValueTypeWithStructs(branches []IRExpr, index int, value IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, expected string) string {
	actual := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(value, structs, callables, bindings)
	next := selfhost_compiler_compiler_inferCompilerPatternCommonType(expected, actual)
	return selfhost_compiler_compiler_inferCompilerPatternBranchTypesWithStructs(branches, index+1, structs, callables, bindings, next)
}

func selfhost_compiler_compiler_inferCompilerPatternCommonType(expected string, actual string) string {
	return func() string {
		switch {
		case expected == "":
			return actual
		default:
			return func() string {
				switch {
				case actual == "":
					return expected
				default:
					return selfhost_compiler_compiler_inferCompilerCommonType(expected, actual)
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerIndexReceiverType(receiverType string, indexExpr IRExpr) string {
	arrayElem := selfhost_compiler_compiler_compilerArrayElementType(receiverType)
	mapValue := selfhost_compiler_compiler_compilerMapValueType(receiverType)
	tuple := selfhost_compiler_compiler_compilerTupleElementTypes(receiverType)
	return func() string {
		switch {
		case arrayElem == "":
			return selfhost_compiler_compiler_inferCompilerNonArrayIndexReceiverType(mapValue, tuple, indexExpr)
		default:
			return arrayElem
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerNonArrayIndexReceiverType(mapValue string, tuple []string, indexExpr IRExpr) string {
	return func() string {
		switch {
		case mapValue == "":
			return selfhost_compiler_compiler_inferCompilerTupleIndexReceiverType(tuple, indexExpr)
		default:
			return selfhost_compiler_compiler_compilerNullableType(mapValue)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerTupleIndexReceiverType(tuple []string, indexExpr IRExpr) string {
	return func() string {
		switch {
		case len(tuple) > 0 == true:
			return selfhost_compiler_compiler_inferCompilerTupleIndexLiteralType(tuple, indexExpr)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerTupleIndexLiteralType(tuple []string, indexExpr IRExpr) string {
	return func() string {
		switch {
		case indexExpr.kind == ExprKind_Int:
			return selfhost_compiler_compiler_inferCompilerTupleIndexElementType(tuple, selfhost_compiler_compiler_compilerParseIntText(indexExpr.value))
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerTupleIndexElementType(tuple []string, index int) string {
	outOfRange := index < 0 || index >= len(tuple)
	return func() string {
		switch {
		case outOfRange == true:
			return ""
		default:
			return tuple[index]
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerCallType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	identifierCall := len(expr.children) > 0 && expr.children[0].kind == ExprKind_Identifier
	return func() string {
		switch {
		case identifierCall == true:
			return selfhost_compiler_compiler_compilerCallableReturnOrText(selfhost_compiler_compiler_findCompilerCallable(callables, expr.children[0].name, 0), expr.text)
		default:
			return selfhost_compiler_compiler_inferCompilerSelectorCallType(expr, callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_compilerCallableReturnOrText(callable CompilerCallable, fallback string) string {
	found := callable.name != ""
	return func() string {
		switch {
		case found == true:
			return callable.returnType
		default:
			return fallback
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerSelectorCallType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	selectorCall := len(expr.children) > 0 && expr.children[0].kind == ExprKind_Selector
	return func() string {
		switch {
		case selectorCall == true:
			return selfhost_compiler_compiler_inferCompilerSelectorCallTypeFromSelector(expr, expr.children[0], callables, bindings)
		default:
			return expr.text
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerSelectorCallTypeFromSelector(expr IRExpr, selector IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	hasReceiver := len(selector.children) > 0
	return func() string {
		switch {
		case hasReceiver == true:
			return selfhost_compiler_compiler_inferCompilerSelectorCallTypeFromReceiver(expr, selector, selector.children[0], callables, bindings)
		default:
			return expr.text
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerSelectorCallTypeFromReceiver(expr IRExpr, selector IRExpr, receiver IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case receiver.kind == ExprKind_At:
			return selfhost_compiler_compiler_inferCompilerAtSelectorCallType(expr, selector, receiver, callables)
		case receiver.kind == ExprKind_Identifier:
			return selfhost_compiler_compiler_inferCompilerIdentifierSelectorCallType(expr, selector, receiver, callables, bindings)
		default:
			return selfhost_compiler_compiler_inferCompilerInstanceSelectorCallType(expr, selector, receiver, callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerAtSelectorCallType(expr IRExpr, selector IRExpr, receiver IRExpr, callables []CompilerCallable) string {
	importPath := compilerIRAtImportPath(receiver)
	return func() string {
		switch {
		case importPath == "" == true:
			return expr.text
		default:
			return func() string {
				if compilerGoPackageImportPath(importPath) != "" {
					return expr.text
				}
				return selfhost_compiler_compiler_compilerCallableReturnOrText(selfhost_compiler_compiler_findCompilerCallable(callables, selector.name, 0), expr.text)
			}()
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerSelectorType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	hasReceiver := len(expr.children) > 0
	return func() string {
		switch {
		case hasReceiver == true:
			return selfhost_compiler_compiler_inferCompilerSelectorTypeFromReceiver(expr, expr.children[0], callables, bindings)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerSelectorTypeFromReceiver(expr IRExpr, receiver IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case receiver.kind == ExprKind_At:
			return func() string {
				switch {
				case compilerIRAtImportPath(receiver) == "":
					return ""
				default:
					return selfhost_compiler_compiler_findCompilerTypeBinding(bindings, expr.name, 0).typeName
				}
			}()
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerIdentifierSelectorCallType(expr IRExpr, selector IRExpr, receiver IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case selector.op == "::":
			return selfhost_compiler_compiler_compilerCallableReturnOrText(selfhost_compiler_compiler_findCompilerCallable(callables, selfhost_compiler_compiler_compilerStaticMethodName(receiver.name, selector.name), 0), expr.text)
		default:
			return selfhost_compiler_compiler_inferCompilerInstanceSelectorCallType(expr, selector, receiver, callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerInstanceSelectorCallType(expr IRExpr, selector IRExpr, receiver IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	receiverType := selfhost_compiler_compiler_inferCompilerExprType(receiver, callables, bindings)
	known := receiverType != ""
	return func() string {
		switch {
		case known == true:
			return selfhost_compiler_compiler_compilerCallableReturnOrText(selfhost_compiler_compiler_findCompilerCallable(callables, selfhost_compiler_compiler_compilerInstanceMethodName(selfhost_compiler_compiler_compilerTypeBase(receiverType), selector.name), 0), expr.text)
		default:
			return expr.text
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerBlockType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_inferCompilerBlockTypeWithStructs(expr, []IRStructType{}, callables, bindings)
}

func selfhost_compiler_compiler_inferCompilerBlockTypeWithStructs(expr IRExpr, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_inferCompilerBlockTypeAtWithStructs(expr.children, 0, structs, callables, bindings, "Void")
}

func selfhost_compiler_compiler_inferCompilerBlockTypeAtWithStructs(statements []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, lastType string) string {
	done := index >= len(statements)
	return func() string {
		switch {
		case done == true:
			return lastType
		default:
			return selfhost_compiler_compiler_inferCompilerBlockTypeStepWithStructs(statements, index, structs, callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerBlockTypeStepWithStructs(statements []IRExpr, index int, structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	statement := statements[index]
	statementType := selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(statement, structs, callables, bindings)
	nextBindings := selfhost_compiler_compiler_blockBindingsAfterStatement(statement, structs, callables, bindings)
	return selfhost_compiler_compiler_inferCompilerBlockTypeAtWithStructs(statements, index+1, structs, callables, nextBindings, statementType)
}

func selfhost_compiler_compiler_inferCompilerTernaryType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	complete := len(expr.children) >= 3
	return func() string {
		switch {
		case complete == true:
			return selfhost_compiler_compiler_inferCompilerCommonType(selfhost_compiler_compiler_inferCompilerExprType(expr.children[1], callables, bindings), selfhost_compiler_compiler_inferCompilerExprType(expr.children[2], callables, bindings))
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerUnaryType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case expr.op == "!":
			return "Bool"
		default:
			return selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerBinaryType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	comparison := func() bool {
		switch {
		case (expr.op == "==") || (expr.op == "!=") || (expr.op == "<") || (expr.op == "<=") || (expr.op == ">") || (expr.op == ">="):
			return true
		default:
			return false
		}
	}()
	boolOp := func() bool {
		switch {
		case (expr.op == "&&") || (expr.op == "||"):
			return true
		default:
			return false
		}
	}()
	knownBool := comparison || boolOp
	return func() string {
		switch {
		case knownBool == true:
			return "Bool"
		default:
			return selfhost_compiler_compiler_inferCompilerBinaryValueType(expr, callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerBinaryValueType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case selfhost_compiler_compiler_compilerArithmeticOp(expr.op) == true:
			return selfhost_compiler_compiler_inferCompilerNumericBinaryType(expr, callables, bindings)
		default:
			return selfhost_compiler_compiler_inferCompilerBitwiseBinaryType(expr, callables, bindings)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerNumericBinaryType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return selfhost_compiler_compiler_inferCompilerNumericPairType(selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings), selfhost_compiler_compiler_inferCompilerExprType(expr.children[1], callables, bindings))
}

func selfhost_compiler_compiler_inferCompilerBitwiseBinaryType(expr IRExpr, callables []CompilerCallable, bindings []CompilerTypeBinding) string {
	return func() string {
		switch {
		case selfhost_compiler_compiler_compilerBitwiseOp(expr.op) == true:
			return selfhost_compiler_compiler_inferCompilerBitwisePairType(selfhost_compiler_compiler_inferCompilerExprType(expr.children[0], callables, bindings), selfhost_compiler_compiler_inferCompilerExprType(expr.children[1], callables, bindings))
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerBitwisePairType(left string, right string) string {
	missing := left == "" || right == ""
	return func() string {
		switch {
		case missing == true:
			return ""
		default:
			return selfhost_compiler_compiler_inferCompilerMatchingPairType(left, right)
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerMatchingPairType(left string, right string) string {
	matched := left == right
	return func() string {
		switch {
		case matched == true:
			return left
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_inferCompilerNumericPairType(left string, right string) string {
	return func() string {
		if left == "" || right == "" {
			return ""
		}
		return func() string {
			if left == "Double" || right == "Double" {
				return "Double"
			}
			return selfhost_compiler_compiler_inferCompilerCommonType(left, right)
		}()
	}()
}

func selfhost_compiler_compiler_inferCompilerCommonType(left string, right string) string {
	return func() string {
		if left == right {
			return left
		}
		return ""
	}()
}

func selfhost_compiler_compiler_findCompilerCallable(callables []CompilerCallable, name string, index int) CompilerCallable {
	done := index >= len(callables)
	return func() CompilerCallable {
		switch {
		case done == true:
			return selfhost_compiler_compiler_emptyCompilerCallable()
		default:
			return selfhost_compiler_compiler_findCompilerCallableAt(callables, name, index)
		}
	}()
}

func selfhost_compiler_compiler_findCompilerCallableAt(callables []CompilerCallable, name string, index int) CompilerCallable {
	matched := callables[index].name == name
	return func() CompilerCallable {
		switch {
		case matched == true:
			return callables[index]
		default:
			return selfhost_compiler_compiler_findCompilerCallable(callables, name, index+1)
		}
	}()
}

func selfhost_compiler_compiler_compilerCallable(name string, arity int, returnType string, paramTypes []string, private bool, sourcePath string) CompilerCallable {
	return CompilerCallable{name: name, arity: arity, returnType: returnType, paramTypes: paramTypes, private: private, sourcePath: sourcePath}
}

func selfhost_compiler_compiler_emptyCompilerCallable() CompilerCallable {
	return selfhost_compiler_compiler_compilerCallable("", 0, "", []string{}, false, "")
}

func selfhost_compiler_compiler_compilerParamTypeNames(params []IRParam) []string {
	names := []string{}
	for _, param := range params {
		_ = param
		func() int { names = append(names, param.typeName); return len(names) }()
	}
	return names
}

func selfhost_compiler_compiler_compilerParamBindings(params []IRParam) []CompilerTypeBinding {
	return selfhost_compiler_compiler_compilerFunctionBindings(params, []CompilerTypeBinding{})
}

func selfhost_compiler_compiler_compilerInitialBindings(file IRFile, callables []CompilerCallable) []CompilerTypeBinding {
	bindings := []CompilerTypeBinding{}
	bindings = selfhost_compiler_compiler_compilerMacroFunctionBindings(file.functions, bindings)
	for _, constant := range file.constants {
		_ = constant
		bindings = selfhost_compiler_compiler_compilerConstBinding(file.structs, callables, bindings, constant)
	}
	for _, importDecl := range file.tsImports {
		_ = importDecl
		bindings = selfhost_compiler_compiler_compilerImportValueBindings(importDecl.values, bindings)
	}
	return bindings
}

func selfhost_compiler_compiler_compilerMacroFunctionBindings(functions []IRFunction, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	out := bindings
	for _, fn := range functions {
		_ = fn
		out = selfhost_compiler_compiler_compilerMacroFunctionBinding(fn, out)
	}
	return out
}

func selfhost_compiler_compiler_compilerMacroFunctionBinding(fn IRFunction, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	return func() []CompilerTypeBinding {
		switch {
		case fn.macro == true:
			return selfhost_compiler_compiler_addCompilerTypeBinding(bindings, fn.name, "MacroFunction")
		default:
			return bindings
		}
	}()
}

func selfhost_compiler_compiler_compilerConstBinding(structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, constant IRConst) []CompilerTypeBinding {
	return selfhost_compiler_compiler_addCompilerTypeBinding(bindings, constant.name, selfhost_compiler_compiler_compilerConstBindingType(structs, callables, bindings, constant))
}

func selfhost_compiler_compiler_compilerConstBindingType(structs []IRStructType, callables []CompilerCallable, bindings []CompilerTypeBinding, constant IRConst) string {
	return func() string {
		switch {
		case constant.typeName == "":
			return selfhost_compiler_compiler_inferCompilerExprTypeWithStructs(constant.value, structs, callables, bindings)
		default:
			return constant.typeName
		}
	}()
}

func selfhost_compiler_compiler_compilerImportValueBindings(values []IRConst, bindings []CompilerTypeBinding) []CompilerTypeBinding {
	out := bindings
	for _, value := range values {
		_ = value
		out = selfhost_compiler_compiler_addCompilerTypeBinding(out, value.name, value.typeName)
	}
	return out
}

func selfhost_compiler_compiler_compilerFunctionBindings(params []IRParam, baseBindings []CompilerTypeBinding) []CompilerTypeBinding {
	bindings := baseBindings
	for _, param := range params {
		_ = param
		bindings = selfhost_compiler_compiler_addCompilerValueBinding(bindings, param.name, param.typeName)
	}
	return bindings
}

func selfhost_compiler_compiler_compilerTypeBinding(name string, typeName string) CompilerTypeBinding {
	return CompilerTypeBinding{name: name, typeName: typeName}
}

func selfhost_compiler_compiler_emptyCompilerTypeBinding() CompilerTypeBinding {
	return selfhost_compiler_compiler_compilerTypeBinding("", "")
}

func selfhost_compiler_compiler_findCompilerTypeBinding(bindings []CompilerTypeBinding, name string, index int) CompilerTypeBinding {
	done := index >= len(bindings)
	return func() CompilerTypeBinding {
		switch {
		case done == true:
			return selfhost_compiler_compiler_emptyCompilerTypeBinding()
		default:
			return selfhost_compiler_compiler_findCompilerTypeBindingAt(bindings, name, index)
		}
	}()
}

func selfhost_compiler_compiler_findCompilerTypeBindingAt(bindings []CompilerTypeBinding, name string, index int) CompilerTypeBinding {
	matched := bindings[index].name == name
	return func() CompilerTypeBinding {
		switch {
		case matched == true:
			return bindings[index]
		default:
			return selfhost_compiler_compiler_findCompilerTypeBinding(bindings, name, index+1)
		}
	}()
}

func selfhost_compiler_compiler_findCompilerStruct(structs []IRStructType, name string, index int) IRStructType {
	done := index >= len(structs)
	return func() IRStructType {
		switch {
		case done == true:
			return selfhost_compiler_compiler_emptyCompilerStruct()
		default:
			return selfhost_compiler_compiler_findCompilerStructAt(structs, name, index)
		}
	}()
}

func selfhost_compiler_compiler_findCompilerStructAt(structs []IRStructType, name string, index int) IRStructType {
	matched := structs[index].name == name
	return func() IRStructType {
		switch {
		case matched == true:
			return structs[index]
		default:
			return selfhost_compiler_compiler_findCompilerStruct(structs, name, index+1)
		}
	}()
}

func selfhost_compiler_compiler_emptyCompilerStruct() IRStructType {
	return IRStructType{name: "", private: false, generics: []string{}, fields: []IRField{}, methods: []IRFunction{}, sourcePath: "", line: 0, column: 0}
}

func selfhost_compiler_compiler_findCompilerStructField(fields []IRField, name string, index int) IRField {
	done := index >= len(fields)
	return func() IRField {
		switch {
		case done == true:
			return selfhost_compiler_compiler_emptyCompilerField()
		default:
			return selfhost_compiler_compiler_findCompilerStructFieldAt(fields, name, index)
		}
	}()
}

func selfhost_compiler_compiler_findCompilerStructFieldAt(fields []IRField, name string, index int) IRField {
	matched := fields[index].name == name
	return func() IRField {
		switch {
		case matched == true:
			return fields[index]
		default:
			return selfhost_compiler_compiler_findCompilerStructField(fields, name, index+1)
		}
	}()
}

func selfhost_compiler_compiler_emptyCompilerField() IRField {
	return IRField{name: "", private: false, typeName: "", jsonName: "", jsonIgnore: false, line: 0, column: 0}
}

func selfhost_compiler_compiler_compilerContains(values []string, value string) bool {
	return selfhost_compiler_compiler_compilerContainsAt(values, value, 0)
}

func selfhost_compiler_compiler_compilerContainsAt(values []string, value string, index int) bool {
	return func() bool {
		if index >= len(values) {
			return false
		}
		return func() bool {
			if values[index] == value {
				return true
			}
			return selfhost_compiler_compiler_compilerContainsAt(values, value, index+1)
		}()
	}()
}

func selfhost_compiler_compiler_compilerStructNameAppearsAfter(structs []IRStructType, name string, index int) bool {
	done := index >= len(structs)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return func() bool {
				switch {
				case structs[index].name == name == true:
					return true
				default:
					return selfhost_compiler_compiler_compilerStructNameAppearsAfter(structs, name, index+1)
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerEnumNameAppearsAfter(enums []IREnumType, name string, index int) bool {
	done := index >= len(enums)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return func() bool {
				switch {
				case enums[index].name == name == true:
					return true
				default:
					return selfhost_compiler_compiler_compilerEnumNameAppearsAfter(enums, name, index+1)
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerFunctionNameAppearsBefore(functions []IRFunction, name string, macro bool, index int) bool {
	done := index < 0
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerFunctionNameAppearsBeforeAt(functions, name, macro, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerFunctionNameAppearsBeforeAt(functions []IRFunction, name string, macro bool, index int) bool {
	matched := functions[index].name == name && functions[index].macro == macro
	return func() bool {
		switch {
		case matched == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerFunctionNameAppearsBefore(functions, name, macro, index-1)
		}
	}()
}

func selfhost_compiler_compiler_compilerConstNameAppearsBefore(constants []IRConst, name string, index int) bool {
	done := index < 0
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerConstNameAppearsBeforeAt(constants, name, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerConstNameAppearsBeforeAt(constants []IRConst, name string, index int) bool {
	matched := constants[index].name == name
	return func() bool {
		switch {
		case matched == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerConstNameAppearsBefore(constants, name, index-1)
		}
	}()
}

func selfhost_compiler_compiler_compilerFieldNameAppearsBefore(fields []IRField, name string, index int) bool {
	done := index < 0
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return func() bool {
				switch {
				case fields[index].name == name == true:
					return true
				default:
					return selfhost_compiler_compiler_compilerFieldNameAppearsBefore(fields, name, index-1)
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerMethodNameAppearsBefore(methods []IRFunction, name string, index int) bool {
	done := index < 0
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return func() bool {
				switch {
				case methods[index].name == name == true:
					return true
				default:
					return selfhost_compiler_compiler_compilerMethodNameAppearsBefore(methods, name, index-1)
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerEnumMemberNameAppearsBefore(members []IREnumMember, name string, index int) bool {
	done := index < 0
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return func() bool {
				switch {
				case members[index].name == name == true:
					return true
				default:
					return selfhost_compiler_compiler_compilerEnumMemberNameAppearsBefore(members, name, index-1)
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerParamNameAppearsBefore(params []IRParam, name string, index int) bool {
	done := index < 0
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return func() bool {
				switch {
				case params[index].name == name == true:
					return true
				default:
					return selfhost_compiler_compiler_compilerParamNameAppearsBefore(params, name, index-1)
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_lowerFiles(files []SourceFile) IRFile {
	return func() IRFile {
		if len(files) == 0 {
			return selfhost_compiler_compiler_emptyCompilerIRFile()
		}
		return selfhost_compiler_compiler_lowerReachableRuneFiles(files, []string{files[0].path}, []string{}, selfhost_compiler_compiler_emptyCompilerIRFile())
	}()
}

func selfhost_compiler_compiler_lowerCompilerSource(source string) IRFile {
	return lowerParsed(selfhost_compiler_compiler_expandCompilerMacros(parse(source)))
}

func selfhost_compiler_compiler_lowerCompilerSourceWithPath(source string, sourcePath string) IRFile {
	return withIRFileSourcePath(selfhost_compiler_compiler_lowerCompilerSource(source), sourcePath)
}

func selfhost_compiler_compiler_expandCompilerMacros(file ParsedFile) ParsedFile {
	imports := selfhost_compiler_compiler_compilerMergeParsedImports(file.imports, selfhost_compiler_compiler_compilerImportExpressions(file), 0)
	errors := selfhost_compiler_compiler_compilerMacroErrors(file, file.errors)
	out := ParsedFile{imports: imports, constants: file.constants, types: []ParsedType{}, functions: []ParsedFunction{}, tests: []ParsedTest{}, errors: errors}
	for _, typeDecl := range file.types {
		_ = typeDecl
		func() int {
			out.types = append(out.types, selfhost_compiler_compiler_expandCompilerTypeMacros(typeDecl))
			return len(out.types)
		}()
	}
	for _, fn := range file.functions {
		_ = fn
		func() int {
			out.functions = append(out.functions, selfhost_compiler_compiler_expandCompilerFunctionMacros(fn))
			return len(out.functions)
		}()
	}
	for _, testDecl := range file.tests {
		_ = testDecl
		func() int {
			out.tests = append(out.tests, selfhost_compiler_compiler_expandCompilerTestMacros(testDecl))
			return len(out.tests)
		}()
	}
	return out
}

func selfhost_compiler_compiler_expandCompilerTypeMacros(typeDecl ParsedType) ParsedType {
	typeName := selfhost_compiler_compiler_compilerRenameDeclarationName(typeDecl.annotations, typeDecl.name)
	fields := []ParsedField{}
	methods := []ParsedFunction{}
	members := []ParsedEnumMember{}
	for _, field := range typeDecl.fields {
		_ = field
		func() int {
			fields = append(fields, selfhost_compiler_compiler_expandCompilerFieldMacros(field))
			return len(fields)
		}()
	}
	for _, method := range typeDecl.methods {
		_ = method
		func() int {
			methods = append(methods, selfhost_compiler_compiler_expandCompilerFunctionMacros(method))
			return len(methods)
		}()
	}
	for _, member := range typeDecl.members {
		_ = member
		func() int {
			members = append(members, selfhost_compiler_compiler_expandCompilerEnumMemberMacros(member))
			return len(members)
		}()
	}
	return ParsedType{name: typeName, private: typeDecl.private, enum: typeDecl.enum, annotations: typeDecl.annotations, generics: typeDecl.generics, fields: fields, methods: selfhost_compiler_compiler_expandCompilerTypeMacroMethods(typeDecl, typeName, methods), members: members, line: typeDecl.line, column: typeDecl.column}
}

func selfhost_compiler_compiler_expandCompilerTypeMacroMethods(typeDecl ParsedType, typeName string, methods []ParsedFunction) []ParsedFunction {
	return func() []ParsedFunction {
		switch {
		case typeDecl.enum == true:
			return methods
		default:
			return selfhost_compiler_compiler_expandCompilerStructMacroMethods(typeDecl, typeName, methods)
		}
	}()
}

func selfhost_compiler_compiler_expandCompilerStructMacroMethods(typeDecl ParsedType, typeName string, methods []ParsedFunction) []ParsedFunction {
	shouldAddFromJson := selfhost_compiler_compiler_compilerHasAnnotation(typeDecl.annotations, "#", "json", "object") && selfhost_compiler_compiler_compilerHasFromJsonMethod(methods, 0) == false
	return func() []ParsedFunction {
		switch {
		case shouldAddFromJson == true:
			return func() []ParsedFunction {
				__rune_spread_out := []ParsedFunction{}
				__rune_spread_out = append(__rune_spread_out, methods...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerJsonFromJsonMethod(typeDecl, typeName))
				return __rune_spread_out
			}()
		default:
			return methods
		}
	}()
}

func selfhost_compiler_compiler_compilerJsonFromJsonMethod(typeDecl ParsedType, typeName string) ParsedFunction {
	return ParsedFunction{name: "fromJson", private: false, static: true, routine: false, macro: false, annotations: selfhost_compiler_compiler_compilerEmptyAnnotations(), receiverType: typeName, generics: []string{}, params: []ParsedParam{ParsedParam{name: "text", typeRef: selfhost_compiler_compiler_compilerTypeRef("String", typeDecl.line, typeDecl.column), line: typeDecl.line, column: typeDecl.column}}, returnType: selfhost_compiler_compiler_compilerGenericTypeRef(typeName, typeDecl.generics, typeDecl.line, typeDecl.column), body: selfhost_compiler_compiler_compilerJsonParseExpr(typeDecl.line, typeDecl.column), line: typeDecl.line, column: typeDecl.column}
}

func selfhost_compiler_compiler_compilerJsonParseExpr(line int, column int) ParsedExpr {
	jsonModule := selfhost_compiler_compiler_compilerParsedExpr(ExprKind_At, "@", "json", "", "", []ParsedParam{}, []ParsedExpr{}, line, column)
	parseSelector := selfhost_compiler_compiler_compilerParsedExpr(ExprKind_Selector, "parse", "parse", "", ".", []ParsedParam{}, []ParsedExpr{jsonModule}, line, column)
	textArg := selfhost_compiler_compiler_compilerParsedExpr(ExprKind_Identifier, "text", "text", "", "", []ParsedParam{}, []ParsedExpr{}, line, column)
	return selfhost_compiler_compiler_compilerParsedExpr(ExprKind_Call, "parse", "", "", "", []ParsedParam{}, []ParsedExpr{parseSelector, textArg}, line, column)
}

func selfhost_compiler_compiler_compilerEmptyAnnotations() []ParsedAnnotation {
	return []ParsedAnnotation{ParsedAnnotation{marker: "", module: "", name: "", args: []ParsedExpr{}, line: 0, column: 0}}
}

func selfhost_compiler_compiler_compilerParsedExpr(kind ExprKind, text string, name string, value string, op string, params []ParsedParam, children []ParsedExpr, line int, column int) ParsedExpr {
	return ParsedExpr{kind: kind, text: text, name: name, value: value, op: op, params: params, children: children, line: line, column: column}
}

func selfhost_compiler_compiler_compilerGenericTypeRef(name string, generics []string, line int, column int) ParsedTypeRef {
	args := selfhost_compiler_compiler_compilerGenericTypeRefArgs(generics, 0, line, column, []ParsedTypeRef{})
	return selfhost_compiler_compiler_compilerTypeRefWithArgs(name, args, line, column)
}

func selfhost_compiler_compiler_compilerGenericTypeRefArgs(generics []string, index int, line int, column int, out []ParsedTypeRef) []ParsedTypeRef {
	done := index >= len(generics)
	return func() []ParsedTypeRef {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_compilerGenericTypeRefArgsStep(generics, index, line, column, out)
		}
	}()
}

func selfhost_compiler_compiler_compilerGenericTypeRefArgsStep(generics []string, index int, line int, column int, out []ParsedTypeRef) []ParsedTypeRef {
	out = append(out, selfhost_compiler_compiler_compilerTypeRef(generics[index], line, column))
	return selfhost_compiler_compiler_compilerGenericTypeRefArgs(generics, index+1, line, column, out)
}

func selfhost_compiler_compiler_compilerTypeRef(name string, line int, column int) ParsedTypeRef {
	return selfhost_compiler_compiler_compilerTypeRefWithArgs(name, []ParsedTypeRef{}, line, column)
}

func selfhost_compiler_compiler_compilerTypeRefWithArgs(name string, args []ParsedTypeRef, line int, column int) ParsedTypeRef {
	return ParsedTypeRef{kind: TypeRefKind_Name, name: name, module: "", nullable: false, args: args, params: []ParsedTypeParam{}, returnTypes: []ParsedTypeRef{}, line: line, column: column}
}

func selfhost_compiler_compiler_compilerHasFromJsonMethod(methods []ParsedFunction, index int) bool {
	done := index >= len(methods)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerHasFromJsonMethodAt(methods, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerHasFromJsonMethodAt(methods []ParsedFunction, index int) bool {
	matched := methods[index].name == "fromJson"
	return func() bool {
		switch {
		case matched == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerHasFromJsonMethod(methods, index+1)
		}
	}()
}

func selfhost_compiler_compiler_compilerHasAnnotation(annotations []ParsedAnnotation, marker string, module string, name string) bool {
	return selfhost_compiler_compiler_compilerHasAnnotationAt(annotations, marker, module, name, 0)
}

func selfhost_compiler_compiler_compilerHasAnnotationAt(annotations []ParsedAnnotation, marker string, module string, name string, index int) bool {
	done := index >= len(annotations)
	return func() bool {
		switch {
		case done == true:
			return false
		default:
			return selfhost_compiler_compiler_compilerHasAnnotationStep(annotations, marker, module, name, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerHasAnnotationStep(annotations []ParsedAnnotation, marker string, module string, name string, index int) bool {
	annotation := annotations[index]
	matched := annotation.marker == marker && annotation.module == module && annotation.name == name
	return func() bool {
		switch {
		case matched == true:
			return true
		default:
			return selfhost_compiler_compiler_compilerHasAnnotationAt(annotations, marker, module, name, index+1)
		}
	}()
}

func selfhost_compiler_compiler_expandCompilerFieldMacros(field ParsedField) ParsedField {
	return ParsedField{name: selfhost_compiler_compiler_compilerRenameDeclarationName(field.annotations, field.name), private: field.private, annotations: field.annotations, typeRef: field.typeRef, line: field.line, column: field.column}
}

func selfhost_compiler_compiler_expandCompilerEnumMemberMacros(member ParsedEnumMember) ParsedEnumMember {
	return ParsedEnumMember{name: selfhost_compiler_compiler_compilerRenameDeclarationName(member.annotations, member.name), private: member.private, annotations: member.annotations, value: member.value, params: member.params, line: member.line, column: member.column}
}

func selfhost_compiler_compiler_expandCompilerFunctionMacros(fn ParsedFunction) ParsedFunction {
	return ParsedFunction{name: selfhost_compiler_compiler_compilerRenameDeclarationName(fn.annotations, fn.name), private: fn.private, static: fn.static, routine: fn.routine, macro: fn.macro, annotations: fn.annotations, receiverType: fn.receiverType, generics: fn.generics, params: fn.params, returnType: fn.returnType, body: selfhost_compiler_compiler_expandCompilerNamespaceAliases(fn.body, []CompilerNamespaceAlias{}), line: fn.line, column: fn.column}
}

func selfhost_compiler_compiler_compilerMacroErrors(file ParsedFile, errors []ParseError) []ParseError {
	functionErrors := selfhost_compiler_compiler_compilerMacroFunctionErrors(file.functions, 0, errors)
	methodErrors := selfhost_compiler_compiler_compilerMacroMethodErrors(file.types, 0, functionErrors)
	return selfhost_compiler_compiler_compilerAnnotationErrors(file, methodErrors)
}

func selfhost_compiler_compiler_compilerMacroMethodErrors(types []ParsedType, index int, errors []ParseError) []ParseError {
	done := index >= len(types)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerMacroMethodErrors(types, index+1, selfhost_compiler_compiler_compilerMacroTypeMethodErrors(types[index].methods, 0, errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerMacroTypeMethodErrors(methods []ParsedFunction, index int, errors []ParseError) []ParseError {
	done := index >= len(methods)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerMacroTypeMethodErrors(methods, index+1, selfhost_compiler_compiler_compilerMacroTypeMethodError(methods[index], errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerMacroTypeMethodError(method ParsedFunction, errors []ParseError) []ParseError {
	return func() []ParseError {
		switch {
		case method.macro == true:
			return func() []ParseError {
				__rune_spread_out := []ParseError{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError("macro declarations must be top-level functions", method.line, method.column))
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_compilerMacroFunctionErrors(functions []ParsedFunction, index int, errors []ParseError) []ParseError {
	done := index >= len(functions)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerMacroFunctionErrors(functions, index+1, selfhost_compiler_compiler_compilerMacroFunctionError(functions[index], errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerMacroFunctionError(fn ParsedFunction, errors []ParseError) []ParseError {
	return func() []ParseError {
		switch {
		case fn.macro == true:
			return func() []ParseError {
				next := func() []ParseError {
					switch {
					case selfhost_compiler_compiler_compilerSyntaxMacroSignatureOk(fn) == true:
						return errors
					default:
						return func() []ParseError {
							__rune_spread_out := []ParseError{}
							__rune_spread_out = append(__rune_spread_out, errors...)
							__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError("macro "+fn.name+" must accept SyntaxFile and MacroContext first and return SyntaxFile", fn.line, fn.column))
							return __rune_spread_out
						}()
					}
				}()
				return selfhost_compiler_compiler_compilerMacroPurityError(fn, next)
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_compilerMacroPurityError(fn ParsedFunction, errors []ParseError) []ParseError {
	return func() []ParseError {
		match9 := selfhost_compiler_compiler_compilerParsedMacroPurityMessage(fn.body)
		switch {
		case match9 == "":
			return errors
		case true:
			message := match9
			_ = message
			return func() []ParseError {
				__rune_spread_out := []ParseError{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError("macro "+fn.name+" is not pure: "+message, fn.line, fn.column))
				return __rune_spread_out
			}()
		}
		return nil
	}()
}

func selfhost_compiler_compiler_compilerParsedMacroPurityMessage(expr ParsedExpr) string {
	current := selfhost_compiler_compiler_compilerParsedMacroCallPurityMessage(expr)
	return func() string {
		switch {
		case current == "":
			return selfhost_compiler_compiler_compilerParsedMacroChildPurityMessage(expr.children, 0)
		default:
			return current
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedMacroChildPurityMessage(children []ParsedExpr, index int) string {
	done := index >= len(children)
	return func() string {
		switch {
		case done == true:
			return ""
		default:
			return selfhost_compiler_compiler_compilerParsedMacroChildPurityMessageAt(children, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedMacroChildPurityMessageAt(children []ParsedExpr, index int) string {
	message := selfhost_compiler_compiler_compilerParsedMacroPurityMessage(children[index])
	return func() string {
		switch {
		case message == "":
			return selfhost_compiler_compiler_compilerParsedMacroChildPurityMessage(children, index+1)
		default:
			return message
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedMacroCallPurityMessage(expr ParsedExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_Call:
			return selfhost_compiler_compiler_compilerParsedCallPurityMessage(expr)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedCallPurityMessage(expr ParsedExpr) string {
	return func() string {
		switch {
		case len(expr.children) == 0:
			return "cannot prove dynamic call is pure"
		default:
			return selfhost_compiler_compiler_compilerParsedCalleePurityMessage(expr.children[0])
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedCalleePurityMessage(callee ParsedExpr) string {
	return func() string {
		switch {
		case callee.kind == ExprKind_Selector:
			return selfhost_compiler_compiler_compilerParsedSelectorPurityMessage(callee)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedSelectorPurityMessage(selector ParsedExpr) string {
	return func() string {
		switch {
		case len(selector.children) == 0:
			return ""
		default:
			return selfhost_compiler_compiler_compilerParsedAtSelectorPurityMessage(selector, selector.children[0])
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedAtSelectorPurityMessage(selector ParsedExpr, receiver ParsedExpr) string {
	return func() string {
		switch {
		case receiver.kind == ExprKind_At:
			return selfhost_compiler_compiler_compilerParsedModuleSelectorPurityMessage(receiver.name, selector.name)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedModuleSelectorPurityMessage(module string, name string) string {
	return func() string {
		switch {
		case module == "io":
			return selfhost_compiler_compiler_compilerParsedIOSelectorPurityMessage(name)
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedIOSelectorPurityMessage(name string) string {
	return func() string {
		switch {
		case name == "println":
			return "calls impure function @io.println"
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_compilerSyntaxMacroSignatureOk(fn ParsedFunction) bool {
	returnOk := typeRefToString(fn.returnType) == "SyntaxFile"
	paramsOk := len(fn.params) >= 2 && (typeRefToString(fn.params[0].typeRef) == "SyntaxFile" && typeRefToString(fn.params[1].typeRef) == "MacroContext")
	return returnOk && paramsOk
}

func selfhost_compiler_compiler_compilerAnnotationErrors(file ParsedFile, errors []ParseError) []ParseError {
	next := selfhost_compiler_compiler_compilerTypeAnnotationErrors(file.types, file.functions, 0, errors)
	return selfhost_compiler_compiler_compilerFunctionAnnotationErrors(file.functions, file.functions, 0, next)
}

func selfhost_compiler_compiler_compilerTypeAnnotationErrors(types []ParsedType, functions []ParsedFunction, index int, errors []ParseError) []ParseError {
	done := index >= len(types)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerTypeAnnotationErrors(types, functions, index+1, selfhost_compiler_compiler_compilerTypeAnnotationError(types[index], functions, errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerTypeAnnotationError(typeDecl ParsedType, functions []ParsedFunction, errors []ParseError) []ParseError {
	next := selfhost_compiler_compiler_compilerAnnotationListErrors(typeDecl.annotations, functions, 0, errors)
	next = selfhost_compiler_compiler_compilerFieldAnnotationErrors(typeDecl.fields, functions, 0, next)
	next = selfhost_compiler_compiler_compilerEnumMemberAnnotationErrors(typeDecl.members, functions, 0, next)
	return selfhost_compiler_compiler_compilerFunctionAnnotationErrors(typeDecl.methods, functions, 0, next)
}

func selfhost_compiler_compiler_compilerFieldAnnotationErrors(fields []ParsedField, functions []ParsedFunction, index int, errors []ParseError) []ParseError {
	done := index >= len(fields)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerFieldAnnotationErrors(fields, functions, index+1, selfhost_compiler_compiler_compilerAnnotationListErrors(fields[index].annotations, functions, 0, errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerEnumMemberAnnotationErrors(members []ParsedEnumMember, functions []ParsedFunction, index int, errors []ParseError) []ParseError {
	done := index >= len(members)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerEnumMemberAnnotationErrors(members, functions, index+1, selfhost_compiler_compiler_compilerAnnotationListErrors(members[index].annotations, functions, 0, errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerFunctionAnnotationErrors(functions []ParsedFunction, topLevelFunctions []ParsedFunction, index int, errors []ParseError) []ParseError {
	done := index >= len(functions)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerFunctionAnnotationErrors(functions, topLevelFunctions, index+1, selfhost_compiler_compiler_compilerAnnotationListErrors(functions[index].annotations, topLevelFunctions, 0, errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerAnnotationListErrors(annotations []ParsedAnnotation, functions []ParsedFunction, index int, errors []ParseError) []ParseError {
	done := index >= len(annotations)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerAnnotationListErrors(annotations, functions, index+1, selfhost_compiler_compiler_compilerAnnotationError(annotations[index], functions, errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerAnnotationError(annotation ParsedAnnotation, functions []ParsedFunction, errors []ParseError) []ParseError {
	return func() []ParseError {
		switch {
		case annotation.marker == "#":
			return selfhost_compiler_compiler_compilerHashAnnotationError(annotation, functions, errors)
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_compilerHashAnnotationError(annotation ParsedAnnotation, functions []ParsedFunction, errors []ParseError) []ParseError {
	return func() []ParseError {
		switch {
		case annotation.module == "":
			return selfhost_compiler_compiler_compilerLocalAnnotationError(annotation, functions, errors)
		default:
			return selfhost_compiler_compiler_compilerModuleAnnotationError(annotation, errors)
		}
	}()
}

func selfhost_compiler_compiler_compilerLocalAnnotationError(annotation ParsedAnnotation, functions []ParsedFunction, errors []ParseError) []ParseError {
	return func() []ParseError {
		switch {
		case annotation.name == "":
			return errors
		case annotation.name == "alias":
			return errors
		default:
			return selfhost_compiler_compiler_compilerResolvedLocalAnnotationError(annotation, selfhost_compiler_compiler_findCompilerMacroBinding(functions, annotation.name, 0), errors)
		}
	}()
}

func selfhost_compiler_compiler_compilerResolvedLocalAnnotationError(annotation ParsedAnnotation, binding CompilerMacroBinding, errors []ParseError) []ParseError {
	return func() []ParseError {
		switch {
		case binding.name == "":
			return func() []ParseError {
				__rune_spread_out := []ParseError{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError("unknown macro #"+annotation.name, annotation.line, annotation.column))
				return __rune_spread_out
			}()
		default:
			return func() []ParseError {
				switch {
				case binding.macro == true:
					return selfhost_compiler_compiler_compilerAnnotationArgErrors(annotation, binding, errors)
				default:
					return func() []ParseError {
						__rune_spread_out := []ParseError{}
						__rune_spread_out = append(__rune_spread_out, errors...)
						__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError("#"+annotation.name+" refers to a function that is not a macro", annotation.line, annotation.column))
						return __rune_spread_out
					}()
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerModuleAnnotationError(annotation ParsedAnnotation, errors []ParseError) []ParseError {
	binding := selfhost_compiler_compiler_compilerBuiltinMacroBinding(annotation.module, annotation.name)
	return func() []ParseError {
		switch {
		case binding.name == "":
			return selfhost_compiler_compiler_compilerUnknownModuleMacroError(annotation, errors)
		default:
			return selfhost_compiler_compiler_compilerAnnotationArgErrors(annotation, binding, errors)
		}
	}()
}

func selfhost_compiler_compiler_compilerUnknownModuleMacroError(annotation ParsedAnnotation, errors []ParseError) []ParseError {
	return func() []ParseError {
		switch {
		case selfhost_compiler_compiler_compilerKnownOrdinaryAnnotationFunction(annotation.module, annotation.name) == true:
			return func() []ParseError {
				__rune_spread_out := []ParseError{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError("#"+annotation.module+"."+annotation.name+" refers to a function that is not a macro", annotation.line, annotation.column))
				return __rune_spread_out
			}()
		default:
			return func() []ParseError {
				__rune_spread_out := []ParseError{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError("unknown macro #"+annotation.module+"."+annotation.name, annotation.line, annotation.column))
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerAnnotationArgErrors(annotation ParsedAnnotation, binding CompilerMacroBinding, errors []ParseError) []ParseError {
	arityOk := len(annotation.args) == len(binding.paramTypes)
	return func() []ParseError {
		switch {
		case arityOk == true:
			return selfhost_compiler_compiler_compilerAnnotationArgTypeErrors(annotation, binding, 0, errors)
		default:
			return func() []ParseError {
				__rune_spread_out := []ParseError{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError("function \""+binding.name+"\" expects "+compilerIntToString(len(binding.paramTypes))+" args, got "+compilerIntToString(len(annotation.args)), annotation.line, annotation.column))
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerAnnotationArgTypeErrors(annotation ParsedAnnotation, binding CompilerMacroBinding, index int, errors []ParseError) []ParseError {
	done := index >= len(binding.paramTypes)
	return func() []ParseError {
		switch {
		case done == true:
			return errors
		default:
			return selfhost_compiler_compiler_compilerAnnotationArgTypeErrors(annotation, binding, index+1, selfhost_compiler_compiler_compilerAnnotationArgTypeError(annotation, binding, index, errors))
		}
	}()
}

func selfhost_compiler_compiler_compilerAnnotationArgTypeError(annotation ParsedAnnotation, binding CompilerMacroBinding, index int, errors []ParseError) []ParseError {
	expected := binding.paramTypes[index]
	actual := selfhost_compiler_compiler_compilerParsedAnnotationArgType(annotation.args[index])
	shouldCheck := selfhost_compiler_compiler_compilerShouldCheckArgType(expected, actual)
	mismatch := shouldCheck && selfhost_compiler_compiler_compilerTypesCompatible(expected, actual) == false
	return func() []ParseError {
		switch {
		case mismatch == true:
			return func() []ParseError {
				__rune_spread_out := []ParseError{}
				__rune_spread_out = append(__rune_spread_out, errors...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_compilerParseError(selfhost_compiler_compiler_compilerArgumentTypeError(binding.name, index+1, actual, expected), annotation.line, annotation.column))
				return __rune_spread_out
			}()
		default:
			return errors
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedAnnotationArgType(expr ParsedExpr) string {
	return func() string {
		switch {
		case expr.kind == ExprKind_String:
			return "String"
		case expr.kind == ExprKind_Template:
			return "String"
		case expr.kind == ExprKind_XMLText:
			return "String"
		case expr.kind == ExprKind_XMLElement:
			return "HTMLElement"
		case expr.kind == ExprKind_Int:
			return "Int"
		case expr.kind == ExprKind_Double:
			return "Double"
		case expr.kind == ExprKind_BigInt:
			return "BigInt"
		case expr.kind == ExprKind_Char:
			return "Char"
		case expr.kind == ExprKind_Bool:
			return "Bool"
		case expr.kind == ExprKind_Null:
			return "Null"
		default:
			return ""
		}
	}()
}

func selfhost_compiler_compiler_findCompilerMacroBinding(functions []ParsedFunction, name string, index int) CompilerMacroBinding {
	done := index >= len(functions)
	return func() CompilerMacroBinding {
		switch {
		case done == true:
			return selfhost_compiler_compiler_emptyCompilerMacroBinding()
		default:
			return selfhost_compiler_compiler_findCompilerMacroBindingAt(functions, name, index)
		}
	}()
}

func selfhost_compiler_compiler_findCompilerMacroBindingAt(functions []ParsedFunction, name string, index int) CompilerMacroBinding {
	matched := functions[index].name == name
	return func() CompilerMacroBinding {
		switch {
		case matched == true:
			return selfhost_compiler_compiler_compilerMacroBindingFromFunction(functions[index])
		default:
			return selfhost_compiler_compiler_findCompilerMacroBinding(functions, name, index+1)
		}
	}()
}

func selfhost_compiler_compiler_compilerMacroBindingFromFunction(fn ParsedFunction) CompilerMacroBinding {
	return CompilerMacroBinding{name: fn.name, macro: fn.macro, paramTypes: func() []string {
		switch {
		case fn.macro == true:
			return selfhost_compiler_compiler_compilerVisibleMacroParamTypes(fn.params, 2, []string{})
		default:
			return selfhost_compiler_compiler_compilerParsedParamTypeNames(fn.params, 0, []string{})
		}
	}()}
}

func selfhost_compiler_compiler_compilerVisibleMacroParamTypes(params []ParsedParam, index int, out []string) []string {
	done := index >= len(params)
	return func() []string {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_compilerVisibleMacroParamTypes(params, index+1, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, typeRefToString(params[index].typeRef))
				return __rune_spread_out
			}())
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedParamTypeNames(params []ParsedParam, index int, out []string) []string {
	done := index >= len(params)
	return func() []string {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_compilerParsedParamTypeNames(params, index+1, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, typeRefToString(params[index].typeRef))
				return __rune_spread_out
			}())
		}
	}()
}

func selfhost_compiler_compiler_compilerBuiltinMacroBinding(module string, name string) CompilerMacroBinding {
	return func() CompilerMacroBinding {
		switch {
		case module == "macro":
			return selfhost_compiler_compiler_compilerMacroModuleBinding(name)
		case module == "json":
			return selfhost_compiler_compiler_compilerJsonModuleMacroBinding(name)
		case module == "cli":
			return selfhost_compiler_compiler_compilerCliModuleMacroBinding(name)
		default:
			return selfhost_compiler_compiler_emptyCompilerMacroBinding()
		}
	}()
}

func selfhost_compiler_compiler_compilerMacroModuleBinding(name string) CompilerMacroBinding {
	return func() CompilerMacroBinding {
		switch {
		case name == "renameDeclaration":
			return selfhost_compiler_compiler_compilerMacroBinding("renameDeclaration", true, []string{"String"})
		default:
			return selfhost_compiler_compiler_emptyCompilerMacroBinding()
		}
	}()
}

func selfhost_compiler_compiler_compilerJsonModuleMacroBinding(name string) CompilerMacroBinding {
	return func() CompilerMacroBinding {
		switch {
		case name == "object":
			return selfhost_compiler_compiler_compilerMacroBinding("object", true, []string{})
		case name == "name":
			return selfhost_compiler_compiler_compilerMacroBinding("name", true, []string{"String"})
		case name == "ignore":
			return selfhost_compiler_compiler_compilerMacroBinding("ignore", true, []string{})
		default:
			return selfhost_compiler_compiler_emptyCompilerMacroBinding()
		}
	}()
}

func selfhost_compiler_compiler_compilerCliModuleMacroBinding(name string) CompilerMacroBinding {
	return func() CompilerMacroBinding {
		switch {
		case name == "command":
			return selfhost_compiler_compiler_compilerMacroBinding("command", true, []string{"String", "String", "String"})
		case name == "flag":
			return selfhost_compiler_compiler_compilerMacroBinding("flag", true, []string{"String", "String"})
		case name == "option":
			return selfhost_compiler_compiler_compilerMacroBinding("option", true, []string{"String", "String", "String", "String"})
		case name == "arg":
			return selfhost_compiler_compiler_compilerMacroBinding("arg", true, []string{"String"})
		case name == "parser":
			return selfhost_compiler_compiler_compilerMacroBinding("parser", true, []string{"String", "String", "String", "String", "String"})
		case name == "main":
			return selfhost_compiler_compiler_compilerMacroBinding("main", true, []string{})
		default:
			return selfhost_compiler_compiler_emptyCompilerMacroBinding()
		}
	}()
}

func selfhost_compiler_compiler_compilerKnownOrdinaryAnnotationFunction(module string, name string) bool {
	return func() bool {
		switch {
		case module == "json":
			return name == "parse" || name == "stringify"
		case module == "go":
			return name == "import" || name == "stmt" || name == "expr"
		default:
			return false
		}
	}()
}

func selfhost_compiler_compiler_compilerMacroBinding(name string, macro bool, paramTypes []string) CompilerMacroBinding {
	return CompilerMacroBinding{name: name, macro: macro, paramTypes: paramTypes}
}

func selfhost_compiler_compiler_emptyCompilerMacroBinding() CompilerMacroBinding {
	return selfhost_compiler_compiler_compilerMacroBinding("", false, []string{})
}

func selfhost_compiler_compiler_compilerParseError(message string, line int, column int) ParseError {
	return ParseError{message: message, line: line, column: column}
}

func selfhost_compiler_compiler_expandCompilerTestMacros(testDecl ParsedTest) ParsedTest {
	return ParsedTest{name: testDecl.name, body: selfhost_compiler_compiler_expandCompilerNamespaceAliases(testDecl.body, []CompilerNamespaceAlias{}), line: testDecl.line, column: testDecl.column}
}

func selfhost_compiler_compiler_compilerImportExpressions(file ParsedFile) []ParsedImport {
	imports := []ParsedImport{}
	for _, typeDecl := range file.types {
		_ = typeDecl
		imports = selfhost_compiler_compiler_compilerCollectTypeImportExprs(typeDecl, imports)
	}
	for _, fn := range file.functions {
		_ = fn
		imports = selfhost_compiler_compiler_compilerCollectFunctionImportExprs(fn, imports)
	}
	for _, testDecl := range file.tests {
		_ = testDecl
		imports = selfhost_compiler_compiler_compilerCollectImportExprs(testDecl.body, imports)
	}
	return imports
}

func selfhost_compiler_compiler_compilerCollectTypeImportExprs(typeDecl ParsedType, imports []ParsedImport) []ParsedImport {
	out := imports
	for _, method := range typeDecl.methods {
		_ = method
		out = selfhost_compiler_compiler_compilerCollectFunctionImportExprs(method, out)
	}
	return out
}

func selfhost_compiler_compiler_compilerCollectFunctionImportExprs(fn ParsedFunction, imports []ParsedImport) []ParsedImport {
	return selfhost_compiler_compiler_compilerCollectImportExprs(fn.body, imports)
}

func selfhost_compiler_compiler_compilerCollectImportExprs(expr ParsedExpr, imports []ParsedImport) []ParsedImport {
	next := func() []ParsedImport {
		switch {
		case expr.kind == ExprKind_At:
			return selfhost_compiler_compiler_compilerAppendImportExpr(imports, expr)
		default:
			return imports
		}
	}()
	for _, child := range expr.children {
		_ = child
		next = selfhost_compiler_compiler_compilerCollectImportExprs(child, next)
	}
	return next
}

func selfhost_compiler_compiler_compilerAppendImportExpr(imports []ParsedImport, expr ParsedExpr) []ParsedImport {
	path := selfhost_compiler_compiler_compilerAtImportPath(expr)
	goPath := compilerGoPackageImportPath(path)
	return func() []ParsedImport {
		if path == "" {
			return imports
		}
		return selfhost_compiler_compiler_compilerAppendParsedImportIfMissing(imports, ParsedImport{path: func() string {
			if goPath != "" {
				return goPath
			}
			return path
		}(), go_: goPath != "", module: false, line: expr.line, column: expr.column})
	}()
}

func selfhost_compiler_compiler_compilerMergeParsedImports(imports []ParsedImport, extra []ParsedImport, index int) []ParsedImport {
	return func() []ParsedImport {
		if index >= len(extra) {
			return imports
		}
		return selfhost_compiler_compiler_compilerMergeParsedImports(selfhost_compiler_compiler_compilerAppendParsedImportIfMissing(imports, extra[index]), extra, index+1)
	}()
}

func selfhost_compiler_compiler_compilerAppendParsedImportIfMissing(imports []ParsedImport, importDecl ParsedImport) []ParsedImport {
	return func() []ParsedImport {
		switch {
		case selfhost_compiler_compiler_compilerParsedImportContains(imports, importDecl.path, importDecl.go_, importDecl.module, 0) == true:
			return imports
		default:
			return func() []ParsedImport {
				__rune_spread_out := []ParsedImport{}
				__rune_spread_out = append(__rune_spread_out, imports...)
				__rune_spread_out = append(__rune_spread_out, importDecl)
				return __rune_spread_out
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerParsedImportContains(imports []ParsedImport, path string, go_ bool, module bool, index int) bool {
	return func() bool {
		if index >= len(imports) {
			return false
		}
		return func() bool {
			if imports[index].path == path && imports[index].go_ == go_ && imports[index].module == module {
				return true
			}
			return selfhost_compiler_compiler_compilerParsedImportContains(imports, path, go_, module, index+1)
		}()
	}()
}

func selfhost_compiler_compiler_compilerEmptyParsedImport() ParsedImport {
	return ParsedImport{path: "", go_: false, module: false, line: 0, column: 0}
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliases(expr ParsedExpr, aliases []CompilerNamespaceAlias) ParsedExpr {
	return func() ParsedExpr {
		switch {
		case expr.kind == ExprKind_Block:
			return selfhost_compiler_compiler_expandCompilerNamespaceAliasBlock(expr, aliases)
		case expr.kind == ExprKind_Lambda:
			return selfhost_compiler_compiler_expandCompilerNamespaceAliasLambda(expr, aliases)
		case expr.kind == ExprKind_Selector:
			return selfhost_compiler_compiler_expandCompilerNamespaceAliasSelector(expr, aliases)
		default:
			return selfhost_compiler_compiler_expandCompilerNamespaceAliasChildren(expr, aliases)
		}
	}()
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasBlock(expr ParsedExpr, aliases []CompilerNamespaceAlias) ParsedExpr {
	return selfhost_compiler_compiler_compilerWithChildren(expr, selfhost_compiler_compiler_expandCompilerNamespaceAliasBlockChildren(expr.children, 0, aliases, []ParsedExpr{}))
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasBlockChildren(statements []ParsedExpr, index int, aliases []CompilerNamespaceAlias, out []ParsedExpr) []ParsedExpr {
	done := index >= len(statements)
	return func() []ParsedExpr {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_expandCompilerNamespaceAliasBlockStep(statements, index, aliases, out)
		}
	}()
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasBlockStep(statements []ParsedExpr, index int, aliases []CompilerNamespaceAlias, out []ParsedExpr) []ParsedExpr {
	statement := statements[index]
	alias := selfhost_compiler_compiler_compilerNamespaceAliasFromLet(statement)
	isAlias := alias.name != ""
	return func() []ParsedExpr {
		switch {
		case isAlias == true:
			return selfhost_compiler_compiler_expandCompilerNamespaceAliasBlockChildren(statements, index+1, selfhost_compiler_compiler_addCompilerNamespaceAlias(aliases, alias), out)
		default:
			return func() []ParsedExpr {
				expanded := selfhost_compiler_compiler_expandCompilerNamespaceAliases(statement, aliases)
				nextAliases := selfhost_compiler_compiler_compilerNamespaceAliasesAfterBinding(expanded, aliases)
				return selfhost_compiler_compiler_expandCompilerNamespaceAliasBlockChildren(statements, index+1, nextAliases, func() []ParsedExpr {
					__rune_spread_out := []ParsedExpr{}
					__rune_spread_out = append(__rune_spread_out, out...)
					__rune_spread_out = append(__rune_spread_out, expanded)
					return __rune_spread_out
				}())
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerNamespaceAliasesAfterBinding(statement ParsedExpr, aliases []CompilerNamespaceAlias) []CompilerNamespaceAlias {
	return func() []CompilerNamespaceAlias {
		switch {
		case statement.kind == ExprKind_Let:
			return selfhost_compiler_compiler_dropCompilerNamespaceAlias(aliases, statement.name, 0, []CompilerNamespaceAlias{})
		case statement.kind == ExprKind_ObjectDestructure:
			return selfhost_compiler_compiler_dropCompilerNamespaceAliasParams(aliases, statement.params, 0)
		default:
			return aliases
		}
	}()
}

func selfhost_compiler_compiler_dropCompilerNamespaceAliasParams(aliases []CompilerNamespaceAlias, params []ParsedParam, index int) []CompilerNamespaceAlias {
	return func() []CompilerNamespaceAlias {
		if index >= len(params) {
			return aliases
		}
		return selfhost_compiler_compiler_dropCompilerNamespaceAliasParams(selfhost_compiler_compiler_dropCompilerNamespaceAlias(aliases, params[index].name, 0, []CompilerNamespaceAlias{}), params, index+1)
	}()
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasLambda(expr ParsedExpr, aliases []CompilerNamespaceAlias) ParsedExpr {
	return selfhost_compiler_compiler_compilerWithChildren(expr, selfhost_compiler_compiler_expandCompilerNamespaceAliasChildrenList(expr.children, 0, selfhost_compiler_compiler_dropCompilerNamespaceAliasParams(aliases, expr.params, 0), []ParsedExpr{}))
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasSelector(expr ParsedExpr, aliases []CompilerNamespaceAlias) ParsedExpr {
	noReceiver := len(expr.children) == 0
	return func() ParsedExpr {
		switch {
		case noReceiver == true:
			return expr
		default:
			return selfhost_compiler_compiler_expandCompilerNamespaceAliasSelectorReceiver(expr, selfhost_compiler_compiler_expandCompilerNamespaceAliases(expr.children[0], aliases), aliases)
		}
	}()
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasSelectorReceiver(expr ParsedExpr, receiver ParsedExpr, aliases []CompilerNamespaceAlias) ParsedExpr {
	canAlias := receiver.kind == ExprKind_Identifier
	alias := func() CompilerNamespaceAlias {
		if canAlias {
			return selfhost_compiler_compiler_findCompilerNamespaceAlias(aliases, receiver.name, 0)
		}
		return selfhost_compiler_compiler_emptyCompilerNamespaceAlias()
	}()
	found := alias.name != ""
	return func() ParsedExpr {
		switch {
		case found == true:
			return selfhost_compiler_compiler_expandCompilerNamespaceAliasSelectorFound(expr, receiver, alias)
		default:
			return selfhost_compiler_compiler_compilerWithChildren(expr, []ParsedExpr{receiver})
		}
	}()
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasSelectorFound(expr ParsedExpr, receiver ParsedExpr, alias CompilerNamespaceAlias) ParsedExpr {
	moduleAlias := alias.module != ""
	return func() ParsedExpr {
		switch {
		case alias.go_ == true:
			return selfhost_compiler_compiler_compilerWithChildren(expr, []ParsedExpr{selfhost_compiler_compiler_compilerImportAtExpr("go:"+alias.importPath, receiver.line, receiver.column)})
		default:
			return func() ParsedExpr {
				switch {
				case moduleAlias == true:
					return selfhost_compiler_compiler_compilerWithChildren(expr, []ParsedExpr{selfhost_compiler_compiler_compilerModuleAtExpr(alias.module, receiver.line, receiver.column)})
				default:
					return selfhost_compiler_compiler_compilerParsedExpr(ExprKind_Identifier, expr.name, expr.name, "", "", []ParsedParam{}, []ParsedExpr{}, expr.line, expr.column)
				}
			}()
		}
	}()
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasChildren(expr ParsedExpr, aliases []CompilerNamespaceAlias) ParsedExpr {
	return selfhost_compiler_compiler_compilerWithChildren(expr, selfhost_compiler_compiler_expandCompilerNamespaceAliasChildrenList(expr.children, 0, aliases, []ParsedExpr{}))
}

func selfhost_compiler_compiler_expandCompilerNamespaceAliasChildrenList(children []ParsedExpr, index int, aliases []CompilerNamespaceAlias, out []ParsedExpr) []ParsedExpr {
	return func() []ParsedExpr {
		if index >= len(children) {
			return out
		}
		return selfhost_compiler_compiler_expandCompilerNamespaceAliasChildrenList(children, index+1, aliases, func() []ParsedExpr {
			__rune_spread_out := []ParsedExpr{}
			__rune_spread_out = append(__rune_spread_out, out...)
			__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_expandCompilerNamespaceAliases(children[index], aliases))
			return __rune_spread_out
		}())
	}()
}

func selfhost_compiler_compiler_compilerNamespaceAliasFromLet(expr ParsedExpr) CompilerNamespaceAlias {
	valid := expr.kind == ExprKind_Let && len(expr.children) > 0
	return func() CompilerNamespaceAlias {
		switch {
		case valid == true:
			return selfhost_compiler_compiler_compilerNamespaceAliasFromLetValue(expr.name, expr.children[0])
		default:
			return selfhost_compiler_compiler_emptyCompilerNamespaceAlias()
		}
	}()
}

func selfhost_compiler_compiler_compilerNamespaceAliasFromLetValue(name string, value ParsedExpr) CompilerNamespaceAlias {
	valid := value.kind == ExprKind_At
	return func() CompilerNamespaceAlias {
		switch {
		case valid == true:
			return selfhost_compiler_compiler_compilerNamespaceAliasFromAt(name, value)
		default:
			return selfhost_compiler_compiler_emptyCompilerNamespaceAlias()
		}
	}()
}

func selfhost_compiler_compiler_compilerNamespaceAliasFromAt(name string, expr ParsedExpr) CompilerNamespaceAlias {
	importPath := selfhost_compiler_compiler_compilerAtImportPath(expr)
	goPath := compilerGoPackageImportPath(importPath)
	return func() CompilerNamespaceAlias {
		if importPath != "" {
			return func() CompilerNamespaceAlias {
				if goPath != "" {
					return selfhost_compiler_compiler_compilerGoNamespaceAlias(name, goPath)
				}
				return selfhost_compiler_compiler_compilerNamespaceAlias(name, "")
			}()
		}
		return func() CompilerNamespaceAlias {
			if expr.name != "" {
				return selfhost_compiler_compiler_compilerNamespaceAlias(name, expr.name)
			}
			return selfhost_compiler_compiler_emptyCompilerNamespaceAlias()
		}()
	}()
}

func selfhost_compiler_compiler_compilerAtImportPath(expr ParsedExpr) string {
	return func() string {
		if expr.kind == ExprKind_At && expr.value != "" {
			return selfhost_compiler_compiler_compilerUnquoteString(expr.value)
		}
		return ""
	}()
}

func selfhost_compiler_compiler_compilerImportAtExpr(path string, line int, column int) ParsedExpr {
	return selfhost_compiler_compiler_compilerParsedExpr(ExprKind_At, "@", "", "\""+path+"\"", "", []ParsedParam{}, []ParsedExpr{}, line, column)
}

func selfhost_compiler_compiler_compilerModuleAtExpr(module string, line int, column int) ParsedExpr {
	return selfhost_compiler_compiler_compilerParsedExpr(ExprKind_At, "@", module, "", "", []ParsedParam{}, []ParsedExpr{}, line, column)
}

func selfhost_compiler_compiler_compilerWithChildren(expr ParsedExpr, children []ParsedExpr) ParsedExpr {
	return selfhost_compiler_compiler_compilerParsedExpr(expr.kind, expr.text, expr.name, expr.value, expr.op, expr.params, children, expr.line, expr.column)
}

func selfhost_compiler_compiler_compilerNamespaceAlias(name string, module string) CompilerNamespaceAlias {
	return CompilerNamespaceAlias{name: name, module: module, importPath: "", go_: false}
}

func selfhost_compiler_compiler_compilerGoNamespaceAlias(name string, importPath string) CompilerNamespaceAlias {
	return CompilerNamespaceAlias{name: name, module: "", importPath: importPath, go_: true}
}

func selfhost_compiler_compiler_emptyCompilerNamespaceAlias() CompilerNamespaceAlias {
	return selfhost_compiler_compiler_compilerNamespaceAlias("", "")
}

func selfhost_compiler_compiler_addCompilerNamespaceAlias(aliases []CompilerNamespaceAlias, alias CompilerNamespaceAlias) []CompilerNamespaceAlias {
	return func() []CompilerNamespaceAlias {
		__rune_spread_out := []CompilerNamespaceAlias{}
		__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_dropCompilerNamespaceAlias(aliases, alias.name, 0, []CompilerNamespaceAlias{})...)
		__rune_spread_out = append(__rune_spread_out, alias)
		return __rune_spread_out
	}()
}

func selfhost_compiler_compiler_dropCompilerNamespaceAlias(aliases []CompilerNamespaceAlias, name string, index int, out []CompilerNamespaceAlias) []CompilerNamespaceAlias {
	return func() []CompilerNamespaceAlias {
		if index >= len(aliases) {
			return out
		}
		return func() []CompilerNamespaceAlias {
			if aliases[index].name == name {
				return selfhost_compiler_compiler_dropCompilerNamespaceAlias(aliases, name, index+1, out)
			}
			return selfhost_compiler_compiler_dropCompilerNamespaceAlias(aliases, name, index+1, func() []CompilerNamespaceAlias {
				__rune_spread_out := []CompilerNamespaceAlias{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, aliases[index])
				return __rune_spread_out
			}())
		}()
	}()
}

func selfhost_compiler_compiler_findCompilerNamespaceAlias(aliases []CompilerNamespaceAlias, name string, index int) CompilerNamespaceAlias {
	return func() CompilerNamespaceAlias {
		if index >= len(aliases) {
			return selfhost_compiler_compiler_emptyCompilerNamespaceAlias()
		}
		return func() CompilerNamespaceAlias {
			if aliases[index].name == name {
				return aliases[index]
			}
			return selfhost_compiler_compiler_findCompilerNamespaceAlias(aliases, name, index+1)
		}()
	}()
}

func selfhost_compiler_compiler_compilerRenameDeclarationName(annotations []ParsedAnnotation, fallback string) string {
	return selfhost_compiler_compiler_compilerRenameDeclarationNameAt(annotations, fallback, 0)
}

func selfhost_compiler_compiler_compilerRenameDeclarationNameAt(annotations []ParsedAnnotation, fallback string, index int) string {
	done := index >= len(annotations)
	return func() string {
		switch {
		case done == true:
			return fallback
		default:
			return selfhost_compiler_compiler_compilerRenameDeclarationNameStep(annotations, fallback, index)
		}
	}()
}

func selfhost_compiler_compiler_compilerRenameDeclarationNameStep(annotations []ParsedAnnotation, fallback string, index int) string {
	annotation := annotations[index]
	matched := annotation.marker == "#" && annotation.module == "macro" && annotation.name == "renameDeclaration" && len(annotation.args) > 0
	return func() string {
		switch {
		case matched == true:
			return selfhost_compiler_compiler_compilerAnnotationStringArg(annotation, 0, fallback)
		default:
			return selfhost_compiler_compiler_compilerRenameDeclarationNameAt(annotations, fallback, index+1)
		}
	}()
}

func selfhost_compiler_compiler_compilerAnnotationStringArg(annotation ParsedAnnotation, index int, fallback string) string {
	valid := index < len(annotation.args) && annotation.args[index].kind == ExprKind_String
	return func() string {
		switch {
		case valid == true:
			return selfhost_compiler_compiler_compilerUnquoteString(annotation.args[index].value)
		default:
			return fallback
		}
	}()
}

func selfhost_compiler_compiler_compilerUnquoteString(raw string) string {
	return func() string {
		if len([]rune(raw)) >= 2 {
			return func() string { runes := []rune(raw); return string(runes[1 : len([]rune(raw))-1]) }()
		}
		return raw
	}()
}

func selfhost_compiler_compiler_lowerReachableRuneFiles(files []SourceFile, pending []string, seen []string, out IRFile) IRFile {
	empty := len(pending) == 0
	return func() IRFile {
		switch {
		case empty == true:
			return out
		default:
			return selfhost_compiler_compiler_lowerReachableRuneFile(files, pending[0], append([]string{}, pending[1:len(pending)]...), seen, out)
		}
	}()
}

func selfhost_compiler_compiler_lowerReachableRuneFile(files []SourceFile, path string, rest []string, seen []string, out IRFile) IRFile {
	alreadySeen := selfhost_compiler_compiler_compilerContains(seen, path)
	return func() IRFile {
		switch {
		case alreadySeen == true:
			return selfhost_compiler_compiler_lowerReachableRuneFiles(files, rest, seen, out)
		default:
			return selfhost_compiler_compiler_lowerUnseenRuneFile(files, path, rest, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, seen...)
				__rune_spread_out = append(__rune_spread_out, path)
				return __rune_spread_out
			}(), out)
		}
	}()
}

func selfhost_compiler_compiler_lowerUnseenRuneFile(files []SourceFile, path string, rest []string, seen []string, out IRFile) IRFile {
	file := selfhost_compiler_compiler_findSourceFile(files, path, 0)
	missing := len(file.path) == 0
	return func() IRFile {
		switch {
		case missing == true:
			return selfhost_compiler_compiler_lowerReachableRuneFiles(files, rest, seen, selfhost_compiler_compiler_mergeCompilerIRFile(out, selfhost_compiler_compiler_missingImportFile(path)))
		default:
			return selfhost_compiler_compiler_lowerFoundSourceFile(files, file, rest, seen, out)
		}
	}()
}

func selfhost_compiler_compiler_lowerFoundSourceFile(files []SourceFile, file SourceFile, rest []string, seen []string, out IRFile) IRFile {
	typeScript := selfhost_compiler_compiler_sourceFileIsTypeScript(file)
	return func() IRFile {
		switch {
		case typeScript == true:
			return selfhost_compiler_compiler_lowerReachableRuneFiles(files, rest, seen, out)
		default:
			return selfhost_compiler_compiler_lowerFoundRuneSourceFile(files, file, rest, seen, out)
		}
	}()
}

func selfhost_compiler_compiler_lowerFoundRuneSourceFile(files []SourceFile, file SourceFile, rest []string, seen []string, out IRFile) IRFile {
	lowered := selfhost_compiler_compiler_lowerCompilerSourceWithPath(file.source, file.path)
	merged := selfhost_compiler_compiler_mergeCompilerIRFile(out, lowered)
	withTypeScript := selfhost_compiler_compiler_lowerTypeScriptImportsForRuneFile(files, file.path, lowered.imports, 0, merged)
	return selfhost_compiler_compiler_lowerReachableRuneFiles(files, selfhost_compiler_compiler_appendImportPaths(rest, file.path, lowered.imports, 0), seen, withTypeScript)
}

func selfhost_compiler_compiler_lowerTypeScriptImportsForRuneFile(files []SourceFile, basePath string, imports []IRImport, index int, out IRFile) IRFile {
	done := index >= len(imports)
	return func() IRFile {
		switch {
		case done == true:
			return out
		default:
			return selfhost_compiler_compiler_lowerTypeScriptImportForRuneFile(files, basePath, imports, index, out)
		}
	}()
}

func selfhost_compiler_compiler_lowerTypeScriptImportForRuneFile(files []SourceFile, basePath string, imports []IRImport, index int, out IRFile) IRFile {
	importDecl := imports[index]
	shouldLoad := importDecl.go_ == false && importDecl.module == false && strings.HasSuffix(importDecl.path, ".ts")
	return func() IRFile {
		switch {
		case shouldLoad == true:
			return selfhost_compiler_compiler_lowerTypeScriptImportPathForRuneFile(files, basePath, imports, index, out)
		default:
			return selfhost_compiler_compiler_lowerTypeScriptImportsForRuneFile(files, basePath, imports, index+1, out)
		}
	}()
}

func selfhost_compiler_compiler_lowerTypeScriptImportPathForRuneFile(files []SourceFile, basePath string, imports []IRImport, index int, out IRFile) IRFile {
	importDecl := imports[index]
	resolved := selfhost_compiler_compiler_resolveCompilerImportPath(basePath, importDecl.path)
	file := selfhost_compiler_compiler_findSourceFile(files, resolved, 0)
	missing := len(file.path) == 0
	next := func() IRFile {
		switch {
		case missing == true:
			return selfhost_compiler_compiler_mergeCompilerIRFile(out, selfhost_compiler_compiler_missingImportFile(resolved))
		default:
			return selfhost_compiler_compiler_mergeCompilerIRFile(out, selfhost_compiler_compiler_lowerTypeScriptSourceFile(file, importDecl.path))
		}
	}()
	return selfhost_compiler_compiler_lowerTypeScriptImportsForRuneFile(files, basePath, imports, index+1, next)
}

func selfhost_compiler_compiler_appendImportPaths(pending []string, basePath string, imports []IRImport, index int) []string {
	done := index >= len(imports)
	return func() []string {
		switch {
		case done == true:
			return pending
		default:
			return selfhost_compiler_compiler_appendImportPath(pending, basePath, imports, index)
		}
	}()
}

func selfhost_compiler_compiler_appendImportPath(pending []string, basePath string, imports []IRImport, index int) []string {
	skip := imports[index].go_ || imports[index].module
	return func() []string {
		switch {
		case skip == true:
			return selfhost_compiler_compiler_appendImportPaths(pending, basePath, imports, index+1)
		default:
			return selfhost_compiler_compiler_appendImportPathIfRune(pending, basePath, imports, index)
		}
	}()
}

func selfhost_compiler_compiler_appendImportPathIfRune(pending []string, basePath string, imports []IRImport, index int) []string {
	typeScript := imports[index].module || strings.HasSuffix(imports[index].path, ".ts")
	return func() []string {
		switch {
		case typeScript == true:
			return selfhost_compiler_compiler_appendImportPaths(pending, basePath, imports, index+1)
		default:
			return selfhost_compiler_compiler_appendImportPaths(func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, pending...)
				__rune_spread_out = append(__rune_spread_out, selfhost_compiler_compiler_resolveCompilerImportPath(basePath, imports[index].path))
				return __rune_spread_out
			}(), basePath, imports, index+1)
		}
	}()
}

func selfhost_compiler_compiler_findSourceFile(files []SourceFile, path string, index int) SourceFile {
	exact := selfhost_compiler_compiler_findSourceFileExact(files, selfhost_compiler_compiler_compilerPathNormalize(path), index)
	found := len(exact.path) == 0 == false
	return func() SourceFile {
		switch {
		case found == true:
			return exact
		default:
			return selfhost_compiler_compiler_findSourceFileBasename(files, path, index)
		}
	}()
}

func selfhost_compiler_compiler_findSourceFileExact(files []SourceFile, path string, index int) SourceFile {
	done := index >= len(files)
	return func() SourceFile {
		switch {
		case done == true:
			return selfhost_compiler_compiler_emptySourceFile()
		default:
			return selfhost_compiler_compiler_findSourceFileExactAt(files, path, index)
		}
	}()
}

func selfhost_compiler_compiler_findSourceFileExactAt(files []SourceFile, path string, index int) SourceFile {
	matched := selfhost_compiler_compiler_compilerPathNormalize(files[index].path) == path
	return func() SourceFile {
		switch {
		case matched == true:
			return files[index]
		default:
			return selfhost_compiler_compiler_findSourceFileExact(files, path, index+1)
		}
	}()
}

func selfhost_compiler_compiler_findSourceFileBasename(files []SourceFile, path string, index int) SourceFile {
	done := index >= len(files)
	return func() SourceFile {
		switch {
		case done == true:
			return selfhost_compiler_compiler_emptySourceFile()
		default:
			return selfhost_compiler_compiler_findSourceFileBasenameAt(files, path, index)
		}
	}()
}

func selfhost_compiler_compiler_findSourceFileBasenameAt(files []SourceFile, path string, index int) SourceFile {
	matched := selfhost_compiler_compiler_compilerPathBasename(files[index].path) == selfhost_compiler_compiler_compilerPathBasename(path)
	return func() SourceFile {
		switch {
		case matched == true:
			return files[index]
		default:
			return selfhost_compiler_compiler_findSourceFileBasename(files, path, index+1)
		}
	}()
}

func selfhost_compiler_compiler_emptySourceFile() SourceFile {
	return SourceFile{path: "", source: ""}
}

func selfhost_compiler_compiler_missingImportFile(path string) IRFile {
	out := selfhost_compiler_compiler_emptyCompilerIRFile()
	out.errors = append(out.errors, ParseError{message: "missing import " + path, line: 0, column: 0})
	return out
}

func selfhost_compiler_compiler_sourceFileIsTypeScript(file SourceFile) bool {
	return strings.HasSuffix(file.path, ".ts")
}

func selfhost_compiler_compiler_resolveCompilerImportPath(basePath string, importPath string) string {
	absolute := strings.HasPrefix(importPath, "/")
	return func() string {
		switch {
		case absolute == true:
			return selfhost_compiler_compiler_compilerPathNormalize(importPath)
		default:
			return selfhost_compiler_compiler_compilerPathNormalize(selfhost_compiler_compiler_compilerPathJoin(selfhost_compiler_compiler_compilerPathDirname(basePath), importPath))
		}
	}()
}

func selfhost_compiler_compiler_compilerPathDirname(path string) string {
	slash := strings.LastIndex(path, "/")
	return func() string {
		if slash < 0 {
			return ""
		}
		return func() string {
			if slash == 0 {
				return "/"
			}
			return func() string { runes := []rune(path); return string(runes[0:slash]) }()
		}()
	}()
}

func selfhost_compiler_compiler_compilerPathJoin(base string, child string) string {
	return func() string {
		if len(base) == 0 {
			return child
		}
		return func() string {
			if base == "/" {
				return "/" + child
			}
			return base + "/" + child
		}()
	}()
}

func selfhost_compiler_compiler_compilerPathNormalize(path string) string {
	absolute := strings.HasPrefix(path, "/")
	parts := func() []string { parts := strings.Split(path, "/"); return parts }()
	normalized := selfhost_compiler_compiler_compilerPathNormalizeParts(parts, 0, absolute, []string{})
	joined := selfhost_compiler_compiler_compilerPathJoinParts(normalized, 0, "")
	return func() string {
		switch {
		case absolute == true:
			return "/" + joined
		default:
			return func() string {
				if len(joined) == 0 {
					return "."
				}
				return joined
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerPathNormalizeParts(parts []string, index int, absolute bool, out []string) []string {
	return func() []string {
		if index >= len(parts) {
			return out
		}
		return selfhost_compiler_compiler_compilerPathNormalizePart(parts, index, absolute, out)
	}()
}

func selfhost_compiler_compiler_compilerPathNormalizePart(parts []string, index int, absolute bool, out []string) []string {
	part := parts[index]
	return func() []string {
		if len(part) == 0 || part == "." {
			return selfhost_compiler_compiler_compilerPathNormalizeParts(parts, index+1, absolute, out)
		}
		return func() []string {
			if part == ".." {
				return selfhost_compiler_compiler_compilerPathNormalizeParent(parts, index, absolute, out)
			}
			return selfhost_compiler_compiler_compilerPathNormalizeParts(parts, index+1, absolute, func() []string {
				__rune_spread_out := []string{}
				__rune_spread_out = append(__rune_spread_out, out...)
				__rune_spread_out = append(__rune_spread_out, part)
				return __rune_spread_out
			}())
		}()
	}()
}

func selfhost_compiler_compiler_compilerPathNormalizeParent(parts []string, index int, absolute bool, out []string) []string {
	canPop := len(out) > 0 && out[len(out)-1] != ".."
	return func() []string {
		switch {
		case canPop == true:
			return selfhost_compiler_compiler_compilerPathNormalizeParts(parts, index+1, absolute, append([]string{}, out[0:len(out)-1]...))
		default:
			return func() []string {
				if absolute {
					return selfhost_compiler_compiler_compilerPathNormalizeParts(parts, index+1, absolute, out)
				}
				return selfhost_compiler_compiler_compilerPathNormalizeParts(parts, index+1, absolute, func() []string {
					__rune_spread_out := []string{}
					__rune_spread_out = append(__rune_spread_out, out...)
					__rune_spread_out = append(__rune_spread_out, "..")
					return __rune_spread_out
				}())
			}()
		}
	}()
}

func selfhost_compiler_compiler_compilerPathJoinParts(parts []string, index int, out string) string {
	return func() string {
		if index >= len(parts) {
			return out
		}
		return selfhost_compiler_compiler_compilerPathJoinParts(parts, index+1, selfhost_compiler_compiler_compilerPathJoin(out, parts[index]))
	}()
}

func selfhost_compiler_compiler_lowerTypeScriptSourceFile(file SourceFile, specifier string) IRFile {
	out := selfhost_compiler_compiler_emptyCompilerIRFile()
	if specifier != "" {
		func() int {
			out.tsImports = append(out.tsImports, selfhost_compiler_compiler_parseTypeScriptImport(file.path, specifier, file.source))
			return len(out.tsImports)
		}()
	}
	return out
}

func selfhost_compiler_compiler_typeScriptImportSpecifier(path string, imports []IRImport, index int) string {
	return func() string {
		if index >= len(imports) {
			return ""
		}
		return func() string {
			if strings.HasSuffix(imports[index].path, ".ts") && selfhost_compiler_compiler_compilerPathBasename(path) == selfhost_compiler_compiler_compilerPathBasename(imports[index].path) {
				return imports[index].path
			}
			return selfhost_compiler_compiler_typeScriptImportSpecifier(path, imports, index+1)
		}()
	}()
}

func selfhost_compiler_compiler_compilerPathBasename(path string) string {
	slash := strings.LastIndex(path, "/")
	return func() string {
		if slash < 0 {
			return path
		}
		return func() string { runes := []rune(path); return string(runes[slash+1 : len([]rune(path))]) }()
	}()
}

func selfhost_compiler_compiler_parseTypeScriptImport(path string, specifier string, source string) IRTSImport {
	imports := IRTSImport{path: path, specifier: specifier, functions: []IRFunction{}, values: []IRConst{}, line: 0, column: 0}
	for _, line := range func() []string { parts := strings.Split(source, "\n"); return parts }() {
		_ = line
		imports = selfhost_compiler_compiler_parseTypeScriptExportLine(imports, strings.TrimSpace(line))
	}
	return imports
}

func selfhost_compiler_compiler_parseTypeScriptExportLine(imports IRTSImport, line string) IRTSImport {
	return func() IRTSImport {
		if strings.HasPrefix(line, "export async function ") {
			return selfhost_compiler_compiler_pushTypeScriptFunction(imports, func() string {
				runes := []rune(line)
				return string(runes[len([]rune("export async function ")):len([]rune(line))])
			}(), true)
		}
		return func() IRTSImport {
			if strings.HasPrefix(line, "export function ") {
				return selfhost_compiler_compiler_pushTypeScriptFunction(imports, func() string {
					runes := []rune(line)
					return string(runes[len([]rune("export function ")):len([]rune(line))])
				}(), false)
			}
			return func() IRTSImport {
				if strings.HasPrefix(line, "export const ") {
					return selfhost_compiler_compiler_pushTypeScriptValue(imports, func() string {
						runes := []rune(line)
						return string(runes[len([]rune("export const ")):len([]rune(line))])
					}())
				}
				return func() IRTSImport {
					if strings.HasPrefix(line, "export let ") {
						return selfhost_compiler_compiler_pushTypeScriptValue(imports, func() string {
							runes := []rune(line)
							return string(runes[len([]rune("export let ")):len([]rune(line))])
						}())
					}
					return func() IRTSImport {
						if strings.HasPrefix(line, "export var ") {
							return selfhost_compiler_compiler_pushTypeScriptValue(imports, func() string {
								runes := []rune(line)
								return string(runes[len([]rune("export var ")):len([]rune(line))])
							}())
						}
						return imports
					}()
				}()
			}()
		}()
	}()
}

func selfhost_compiler_compiler_pushTypeScriptFunction(imports IRTSImport, text string, routine bool) IRTSImport {
	open := strings.Index(text, "(")
	close := strings.Index(text, ")")
	name := func() string {
		if open < 0 {
			return ""
		}
		return strings.TrimSpace((func() string { runes := []rune(text); return string(runes[0:open]) }()))
	}()
	returnType := selfhost_compiler_compiler_typeScriptReturnTypeName(text)
	if name != "" {
		func() int {
			imports.functions = append(imports.functions, IRFunction{name: name, private: false, static: false, routine: routine, macro: false, receiverType: "", generics: []string{}, params: func() []IRParam {
				if open >= 0 && close > open {
					return selfhost_compiler_compiler_parseTypeScriptParams(func() string { runes := []rune(text); return string(runes[open+1 : close]) }())
				}
				return []IRParam{}
			}(), returnType: returnType, body: emptyIRExpr(), sourcePath: "", line: 0, column: 0})
			return len(imports.functions)
		}()
	}
	return imports
}

func selfhost_compiler_compiler_pushTypeScriptValue(imports IRTSImport, text string) IRTSImport {
	end := selfhost_compiler_compiler_typeScriptNameEnd(text)
	name := strings.TrimSpace((func() string { runes := []rune(text); return string(runes[0:end]) }()))
	typeName := selfhost_compiler_compiler_typeScriptValueTypeName(text)
	if name != "" {
		func() int {
			imports.values = append(imports.values, IRConst{name: name, private: false, typeName: typeName, value: emptyIRExpr(), line: 0, column: 0})
			return len(imports.values)
		}()
	}
	return imports
}

func selfhost_compiler_compiler_typeScriptNameEnd(text string) int {
	colon := strings.Index(text, ":")
	equal := strings.Index(text, "=")
	return func() int {
		if colon >= 0 && (equal < 0 || colon < equal) {
			return colon
		}
		return func() int {
			if equal >= 0 {
				return equal
			}
			return len([]rune(text))
		}()
	}()
}

func selfhost_compiler_compiler_typeScriptReturnTypeName(text string) string {
	close := strings.Index(text, ")")
	colon := func() int {
		if close >= 0 {
			return strings.Index((func() string { runes := []rune(text); return string(runes[close+1 : len([]rune(text))]) }()), ":")
		}
		return -1
	}()
	return func() string {
		if close >= 0 && colon >= 0 {
			return selfhost_compiler_compiler_typeScriptTextType(strings.TrimSpace((func() string {
				runes := []rune(text)
				return string(runes[close+1+colon+1 : selfhost_compiler_compiler_typeScriptReturnTypeEnd(text)])
			}())))
		}
		return "Dynamic"
	}()
}

func selfhost_compiler_compiler_typeScriptReturnTypeEnd(text string) int {
	brace := strings.Index(text, "{")
	semi := strings.Index(text, ";")
	return func() int {
		if brace >= 0 && (semi < 0 || brace < semi) {
			return brace
		}
		return func() int {
			if semi >= 0 {
				return semi
			}
			return len([]rune(text))
		}()
	}()
}

func selfhost_compiler_compiler_typeScriptValueTypeName(text string) string {
	colon := strings.Index(text, ":")
	return func() string {
		if colon < 0 {
			return "Dynamic"
		}
		return selfhost_compiler_compiler_typeScriptTextType(strings.TrimSpace((func() string {
			runes := []rune(text)
			return string(runes[colon+1 : selfhost_compiler_compiler_typeScriptNameEnd(func() string { runes := []rune(text); return string(runes[colon+1 : len([]rune(text))]) }())+colon+1])
		}())))
	}()
}

func selfhost_compiler_compiler_parseTypeScriptParams(text string) []IRParam {
	params := []IRParam{}
	for _, param := range func() []string { parts := strings.Split(text, ","); return parts }() {
		_ = param
		func() {
			if strings.TrimSpace(param) != "" {
				func() int {
					params = append(params, selfhost_compiler_compiler_parseTypeScriptParam(strings.TrimSpace(param)))
					return len(params)
				}()
				return
			}
		}()
	}
	return params
}

func selfhost_compiler_compiler_emptyIRParam() IRParam {
	return IRParam{name: "", typeName: "Dynamic", line: 0, column: 0}
}

func selfhost_compiler_compiler_emptyIRConst() IRConst {
	return IRConst{name: "", private: false, typeName: "Dynamic", value: emptyIRExpr(), line: 0, column: 0}
}

func selfhost_compiler_compiler_parseTypeScriptParam(text string) IRParam {
	colon := strings.Index(text, ":")
	rawName := func() string {
		if colon < 0 {
			return text
		}
		return func() string { runes := []rune(text); return string(runes[0:colon]) }()
	}()
	name := strings.TrimSpace((strings.ReplaceAll((strings.ReplaceAll(rawName, "...", "")), "?", "")))
	return IRParam{name: name, typeName: func() string {
		if colon < 0 {
			return "Dynamic"
		}
		return selfhost_compiler_compiler_typeScriptTextType(strings.TrimSpace((func() string { runes := []rune(text); return string(runes[colon+1 : len([]rune(text))]) }())))
	}(), line: 0, column: 0}
}

func selfhost_compiler_compiler_typeScriptTextType(text string) string {
	return func() string {
		switch {
		case text == "string":
			return "String"
		case text == "boolean":
			return "Bool"
		case text == "bigint":
			return "BigInt"
		case text == "number":
			return "Double"
		case (text == "void") || (text == "undefined"):
			return "Void"
		default:
			return "Dynamic"
		}
	}()
}

func selfhost_compiler_compiler_emptyCompilerIRFile() IRFile {
	return IRFile{imports: []IRImport{}, tsImports: []IRTSImport{}, structs: []IRStructType{}, enums: []IREnumType{}, constants: []IRConst{}, functions: []IRFunction{}, tests: []IRTest{}, errors: []ParseError{}}
}

func selfhost_compiler_compiler_mergeCompilerIRFile(out IRFile, file IRFile) IRFile {
	for _, importDecl := range file.imports {
		_ = importDecl
		func() int { out.imports = append(out.imports, importDecl); return len(out.imports) }()
	}
	for _, importDecl := range file.tsImports {
		_ = importDecl
		func() int { out.tsImports = append(out.tsImports, importDecl); return len(out.tsImports) }()
	}
	for _, typeDecl := range file.structs {
		_ = typeDecl
		func() int { out.structs = append(out.structs, typeDecl); return len(out.structs) }()
	}
	for _, typeDecl := range file.enums {
		_ = typeDecl
		func() int { out.enums = append(out.enums, typeDecl); return len(out.enums) }()
	}
	for _, constant := range file.constants {
		_ = constant
		func() int { out.constants = append(out.constants, constant); return len(out.constants) }()
	}
	for _, fn := range file.functions {
		_ = fn
		func() int { out.functions = append(out.functions, fn); return len(out.functions) }()
	}
	for _, testDecl := range file.tests {
		_ = testDecl
		func() int { out.tests = append(out.tests, testDecl); return len(out.tests) }()
	}
	for _, error_ := range file.errors {
		_ = error_
		func() int { out.errors = append(out.errors, error_); return len(out.errors) }()
	}
	return out
}

func selfhost_compiler_compiler_compileResult(ok bool, output string, errors []string) CompileResult {
	return CompileResult{ok: ok, output: output, errors: errors}
}

func selfhost_compiler_compiler_unsupportedTargetErrors(target string) []string {
	errors := []string{}
	errors = append(errors, "unsupported target "+target)
	return errors
}

func selfhost_compiler_compiler_parseErrorMessages(errors []ParseError) []string {
	out := []string{}
	for _, error_ := range errors {
		_ = error_
		func() int {
			out = append(out, "line "+compilerIntToString(error_.line)+":"+compilerIntToString(error_.column)+": "+error_.message)
			return len(out)
		}()
	}
	return out
}
