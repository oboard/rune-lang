package lexer

import "fmt"

type Kind int

const (
	EOF Kind = iota
	Illegal
	Newline

	Ident
	Int
	String

	At
	Dot
	Comma
	Colon
	LParen
	RParen
	LBracket
	RBracket
	LBrace
	RBrace

	FatArrow
	Assign
	Declare
	MutDeclare
	SignalDeclare
	Arrow

	Plus
	Minus
	Star
	Slash
	Percent
	Bang

	EqualEqual
	BangEqual
	Less
	LessEqual
	Greater
	GreaterEqual

	Underscore
)

type Position struct {
	Offset int
	Line   int
	Column int
}

func (p Position) String() string {
	if p.Line <= 0 {
		return "<unknown>"
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

type Token struct {
	Kind   Kind
	Lexeme string
	Pos    Position
}

func (t Token) String() string {
	if t.Lexeme == "" {
		return t.Kind.String()
	}
	return fmt.Sprintf("%s(%q)", t.Kind, t.Lexeme)
}

func (k Kind) String() string {
	switch k {
	case EOF:
		return "EOF"
	case Illegal:
		return "Illegal"
	case Newline:
		return "Newline"
	case Ident:
		return "Ident"
	case Int:
		return "Int"
	case String:
		return "String"
	case At:
		return "@"
	case Dot:
		return "."
	case Comma:
		return ","
	case Colon:
		return ":"
	case LParen:
		return "("
	case RParen:
		return ")"
	case LBracket:
		return "["
	case RBracket:
		return "]"
	case LBrace:
		return "{"
	case RBrace:
		return "}"
	case FatArrow:
		return "=>"
	case Assign:
		return "="
	case Declare:
		return ":="
	case MutDeclare:
		return "~="
	case SignalDeclare:
		return "$="
	case Arrow:
		return "->"
	case Plus:
		return "+"
	case Minus:
		return "-"
	case Star:
		return "*"
	case Slash:
		return "/"
	case Percent:
		return "%"
	case Bang:
		return "!"
	case EqualEqual:
		return "=="
	case BangEqual:
		return "!="
	case Less:
		return "<"
	case LessEqual:
		return "<="
	case Greater:
		return ">"
	case GreaterEqual:
		return ">="
	case Underscore:
		return "_"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}
