package lexer

type Lexer struct {
	input  []rune
	start  int
	curr   int
	line   int
	column int
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
