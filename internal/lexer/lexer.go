package lexer

type mode int

const (
	modeCode mode = iota
	modeXMLTag
	modeXMLText
	modeXMLExpr
)

type Lexer struct {
	input []rune
	start int
	curr  int

	line   int
	column int

	mode          mode
	xmlDepth      int
	xmlClosing    bool
	xmlSelfClosed bool
	xmlExprMode   mode
	xmlExprDepth  int
}

func Lex(src string) []Token {
	l := New(src)
	var tokens []Token
	for {
		tok := l.Next()
		tokens = append(tokens, tok)
		if tok.Kind == EOF {
			return tokens
		}
	}
}

func New(src string) *Lexer {
	return &Lexer{
		input:  []rune(src),
		line:   1,
		column: 1,
	}
}
