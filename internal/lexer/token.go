package lexer

import "fmt"

type Kind int

const (
	EOF Kind = iota
	Illegal
	Newline

	Ident
	Int
	Double
	BigInt
	String
	Regex
	XMLText

	At
	Dollar
	Dot
	DotDotDot
	Comma
	Colon
	LParen
	RParen
	LBracket
	RBracket
	LBrace
	RBrace
	Question

	FatArrow
	Assign
	Declare
	MutDeclare
	SignalDeclare
	Arrow

	Plus
	PlusPlus
	Minus
	Star
	Slash
	Percent
	Bang
	Tilde
	BitAnd
	BitOr
	BitXor
	ShiftLeft
	ShiftRight
	UnsignedShiftRight
	AndAnd
	OrOr

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
	case Double:
		return "Double"
	case BigInt:
		return "BigInt"
	case String:
		return "String"
	case Regex:
		return "Regex"
	case XMLText:
		return "XMLText"
	case At:
		return "@"
	case Dollar:
		return "$"
	case Dot:
		return "."
	case DotDotDot:
		return "..."
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
	case Question:
		return "?"
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
	case PlusPlus:
		return "++"
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
	case Tilde:
		return "~"
	case BitAnd:
		return "&"
	case BitOr:
		return "|"
	case BitXor:
		return "^"
	case ShiftLeft:
		return "<<"
	case ShiftRight:
		return ">>"
	case UnsignedShiftRight:
		return ">>>"
	case AndAnd:
		return "&&"
	case OrOr:
		return "||"
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
