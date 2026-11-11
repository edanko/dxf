package query

import (
	"fmt"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"unicode"

	"github.com/edanko/dxf/entity"
)

// Token types for the lexer
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenNumber
	TokenString
	TokenLParen
	TokenRParen
	TokenLBracket
	TokenRBracket
	TokenComma
	TokenAnd
	TokenOr
	TokenNot
	TokenEquals
	TokenNotEquals
	TokenGreaterThan
	TokenGreaterEquals
	TokenLessThan
	TokenLessEquals
	TokenContains
	TokenMatches
	TokenIn
	TokenMinus
)

type Token struct {
	Type  TokenType
	Value string
	Pos   int
}

func (t Token) String() string {
	if t.Type == TokenEOF {
		return "EOF"
	}
	return fmt.Sprintf("%s(%q)", tokenTypeName(t.Type), t.Value)
}

func tokenTypeName(t TokenType) string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdentifier:
		return "IDENT"
	case TokenNumber:
		return "NUMBER"
	case TokenString:
		return "STRING"
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenLBracket:
		return "["
	case TokenRBracket:
		return "]"
	case TokenComma:
		return ","
	case TokenAnd:
		return "AND"
	case TokenOr:
		return "OR"
	case TokenNot:
		return "NOT"
	case TokenEquals:
		return "=="
	case TokenNotEquals:
		return "!="
	case TokenGreaterThan:
		return ">"
	case TokenGreaterEquals:
		return ">="
	case TokenLessThan:
		return "<"
	case TokenLessEquals:
		return "<="
	case TokenContains:
		return "CONTAINS"
	case TokenMatches:
		return "MATCHES"
	case TokenIn:
		return "IN"
	case TokenMinus:
		return "-"
	default:
		return fmt.Sprintf("Token(%d)", t)
	}
}

// Lexer converts query string into tokens
type Lexer struct {
	input    string
	position int
	tokens   []Token
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: input}
}

func (l *Lexer) Lex() ([]Token, error) {
	l.tokens = make([]Token, 0)

	for l.position < len(l.input) {
		if err := l.skipWhitespace(); err != nil {
			return nil, err
		}

		if l.position >= len(l.input) {
			break
		}

		ch := l.input[l.position]

		switch {
		case unicode.IsLetter(rune(ch)):
			token := l.readIdentifier()
			if keyword, ok := isKeyword(token.Value); ok {
				l.tokens = append(l.tokens, Token{Type: keyword, Value: token.Value, Pos: token.Pos})
			} else {
				l.tokens = append(l.tokens, Token{Type: TokenIdentifier, Value: token.Value, Pos: token.Pos})
			}

		case unicode.IsDigit(rune(ch)) || ch == '.':
			token := l.readNumber()
			l.tokens = append(l.tokens, Token{Type: TokenNumber, Value: token.Value, Pos: token.Pos})

		case ch == '\'' || ch == '"':
			token, err := l.readString()
			if err != nil {
				return nil, err
			}
			l.tokens = append(l.tokens, Token{Type: TokenString, Value: token.Value, Pos: token.Pos})

		case ch == '(':
			l.tokens = append(l.tokens, Token{Type: TokenLParen, Value: "(", Pos: l.position})
			l.position++

		case ch == ')':
			l.tokens = append(l.tokens, Token{Type: TokenRParen, Value: ")", Pos: l.position})
			l.position++

		case ch == '[':
			l.tokens = append(l.tokens, Token{Type: TokenLBracket, Value: "[", Pos: l.position})
			l.position++

		case ch == ']':
			l.tokens = append(l.tokens, Token{Type: TokenRBracket, Value: "]", Pos: l.position})
			l.position++

		case ch == ',':
			l.tokens = append(l.tokens, Token{Type: TokenComma, Value: ",", Pos: l.position})
			l.position++

		case ch == '-':
			l.tokens = append(l.tokens, Token{Type: TokenMinus, Value: "-", Pos: l.position})
			l.position++

		default:
			remaining := l.input[l.position:]
			if strings.HasPrefix(remaining, "!=") {
				l.tokens = append(l.tokens, Token{Type: TokenNotEquals, Value: "!=", Pos: l.position})
				l.position += 2
			} else if strings.HasPrefix(remaining, ">=") {
				l.tokens = append(l.tokens, Token{Type: TokenGreaterEquals, Value: ">=", Pos: l.position})
				l.position += 2
			} else if strings.HasPrefix(remaining, "<=") {
				l.tokens = append(l.tokens, Token{Type: TokenLessEquals, Value: "<=", Pos: l.position})
				l.position += 2
			} else if strings.HasPrefix(remaining, "==") {
				l.tokens = append(l.tokens, Token{Type: TokenEquals, Value: "==", Pos: l.position})
				l.position += 2
			} else if remaining[0] == '=' {
				l.tokens = append(l.tokens, Token{Type: TokenEquals, Value: "=", Pos: l.position})
				l.position++
			} else if remaining[0] == '>' {
				l.tokens = append(l.tokens, Token{Type: TokenGreaterThan, Value: ">", Pos: l.position})
				l.position++
			} else if remaining[0] == '<' {
				l.tokens = append(l.tokens, Token{Type: TokenLessThan, Value: "<", Pos: l.position})
				l.position++
			} else {
				return nil, fmt.Errorf("unexpected character at position %d: %c", l.position, ch)
			}
		}
	}

	l.tokens = append(l.tokens, Token{Type: TokenEOF, Value: "", Pos: l.position})
	return l.tokens, nil
}

func (l *Lexer) skipWhitespace() error {
	for l.position < len(l.input) && unicode.IsSpace(rune(l.input[l.position])) {
		l.position++
	}
	return nil
}

func (l *Lexer) readIdentifier() Token {
	start := l.position
	for l.position < len(l.input) && (unicode.IsLetter(rune(l.input[l.position])) || unicode.IsDigit(rune(l.input[l.position])) || l.input[l.position] == '_' || l.input[l.position] == '$') {
		l.position++
	}
	return Token{Value: l.input[start:l.position], Pos: start}
}

func (l *Lexer) readNumber() Token {
	start := l.position

	for l.position < len(l.input) && (unicode.IsDigit(rune(l.input[l.position])) || l.input[l.position] == '.') {
		l.position++
	}

	return Token{Value: l.input[start:l.position], Pos: start}
}

func (l *Lexer) readString() (Token, error) {
	start := l.position
	quote := l.input[l.position]
	l.position++

	var value strings.Builder
	for l.position < len(l.input) {
		ch := l.input[l.position]
		if ch == quote {
			l.position++
			return Token{Type: TokenString, Value: value.String(), Pos: start}, nil
		}
		if ch == '\\' && l.position+1 < len(l.input) {
			l.position++
			next := l.input[l.position]
			switch next {
			case 'n':
				value.WriteByte('\n')
			case 't':
				value.WriteByte('\t')
			case '"':
				value.WriteByte('"')
			case '\'':
				value.WriteByte('\'')
			case '\\':
				value.WriteByte('\\')
			default:
				value.WriteByte(next)
			}
		} else {
			value.WriteByte(ch)
		}
		l.position++
	}

	return Token{}, fmt.Errorf("unterminated string at position %d", start)
}

func isKeyword(ident string) (TokenType, bool) {
	switch strings.ToUpper(ident) {
	case "AND":
		return TokenAnd, true
	case "OR":
		return TokenOr, true
	case "NOT":
		return TokenNot, true
	case "IN":
		return TokenIn, true
	case "CONTAINS":
		return TokenContains, true
	case "MATCHES":
		return TokenMatches, true
	default:
		return TokenEOF, false
	}
}

// Expression types
type Expression interface {
	Match(entity.Entity) bool
	String() string
}

type OrExpression struct {
	Left  Expression
	Right Expression
}

func (e *OrExpression) Match(ent entity.Entity) bool {
	return e.Left.Match(ent) || e.Right.Match(ent)
}

func (e *OrExpression) String() string {
	return fmt.Sprintf("(%s OR %s)", e.Left.String(), e.Right.String())
}

type AndExpression struct {
	Left  Expression
	Right Expression
}

func (e *AndExpression) Match(ent entity.Entity) bool {
	return e.Left.Match(ent) && e.Right.Match(ent)
}

func (e *AndExpression) String() string {
	return fmt.Sprintf("(%s AND %s)", e.Left.String(), e.Right.String())
}

type NotExpression struct {
	Expression Expression
}

func (e *NotExpression) Match(ent entity.Entity) bool {
	return !e.Expression.Match(ent)
}

func (e *NotExpression) String() string {
	return fmt.Sprintf("NOT (%s)", e.Expression.String())
}

type ComparisonExpression struct {
	Property string
	Operator Operator
	Value    any
	Pos      int
}

// CallExpression is a bare function-call predicate such as
// bbox_intersects(0, 0, 10, 10) or has_xdata('MYAPP').
type CallExpression struct {
	Fn   string
	Args []any
	Pos  int
}

func (e *CallExpression) Match(ent entity.Entity) bool {
	r, ok := ContextOf(ent).Call(e.Fn, e.Args).(bool)
	return ok && r
}

func (e *CallExpression) String() string {
	return fmt.Sprintf("%s(%v)", e.Fn, e.Args)
}

// FuncComparison compares the value of a query function such as
// attr('SIZE') == '10'.
type FuncComparison struct {
	Fn       string
	Args     []any
	Operator Operator
	Value    any
	Pos      int
}

func (e *FuncComparison) Match(ent entity.Entity) bool {
	return compareValues(ContextOf(ent).Call(e.Fn, e.Args), e.Value, e.Operator)
}

func (e *FuncComparison) String() string {
	return fmt.Sprintf("%s(%v) %s %v", e.Fn, e.Args, operatorString(e.Operator), e.Value)
}

type Operator int

const (
	OpEquals Operator = iota
	OpNotEquals
	OpGreaterThan
	OpGreaterEquals
	OpLessThan
	OpLessEquals
	OpContains
	OpMatches
	OpIn
)

func (e *ComparisonExpression) Match(ent entity.Entity) bool {
	return compareValues(ContextOf(ent).Property(e.Property), e.Value, e.Operator)
}

func (e *ComparisonExpression) String() string {
	return fmt.Sprintf("%s %s %v", e.Property, operatorString(e.Operator), e.Value)
}

func operatorString(op Operator) string {
	switch op {
	case OpEquals:
		return "=="
	case OpNotEquals:
		return "!="
	case OpGreaterThan:
		return ">"
	case OpGreaterEquals:
		return ">="
	case OpLessThan:
		return "<"
	case OpLessEquals:
		return "<="
	case OpContains:
		return "CONTAINS"
	case OpMatches:
		return "MATCHES"
	case OpIn:
		return "IN"
	default:
		return "?"
	}
}

func compareValues(propValue, cmpValue any, op Operator) bool {
	switch op {
	case OpEquals:
		return propValue == cmpValue
	case OpNotEquals:
		return propValue != cmpValue
	case OpGreaterThan:
		return toFloat(propValue) > toFloat(cmpValue)
	case OpGreaterEquals:
		return toFloat(propValue) >= toFloat(cmpValue)
	case OpLessThan:
		return toFloat(propValue) < toFloat(cmpValue)
	case OpLessEquals:
		return toFloat(propValue) <= toFloat(cmpValue)
	case OpContains:
		return strings.Contains(toString(propValue), toString(cmpValue))
	case OpMatches:
		re, err := regexp.Compile(toString(cmpValue))
		if err != nil {
			return false
		}
		return re.MatchString(toString(propValue))
	case OpIn:
		if list, ok := cmpValue.([]any); ok {
			if slices.Contains(list, propValue) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func toFloat(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case string:
		if f, err := strconv.ParseFloat(x, 64); err == nil {
			return f
		}
	}
	return 0
}

func toString(v any) string {
	switch x := v.(type) {
	case string:
		return x
	default:
		return fmt.Sprintf("%v", x)
	}
}

// Parser converts tokens into expressions
type Parser struct {
	lexer  *Lexer
	tokens []Token
	pos    int
}

func NewParser(input string) (*Parser, error) {
	lexer := NewLexer(input)
	tokens, err := lexer.Lex()
	if err != nil {
		return nil, err
	}
	return &Parser{tokens: tokens}, nil
}

func (p *Parser) Parse() (Expression, error) {
	expr, err := p.parseOr()
	if err != nil {
		return nil, err
	}
	if !p.consume(TokenEOF) {
		return nil, p.unexpected("EOF")
	}
	return expr, nil
}

func (p *Parser) parseOr() (Expression, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.consume(TokenOr) {
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &OrExpression{Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseAnd() (Expression, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.consume(TokenAnd) {
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &AndExpression{Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseNot() (Expression, error) {
	if p.consume(TokenNot) {
		expr, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &NotExpression{Expression: expr}, nil
	}
	if p.consume(TokenLParen) {
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if !p.consume(TokenRParen) {
			return nil, p.unexpected(")")
		}
		return expr, nil
	}
	return p.parseComparison()
}

func (p *Parser) parseComparison() (Expression, error) {
	property, propPos, err := p.parseProperty()
	if err != nil {
		return nil, err
	}

	// property(...) -> function call predicate or function comparison
	if p.check(TokenLParen) {
		p.pos++
		args, err := p.parseArgs()
		if err != nil {
			return nil, err
		}
		if !p.consume(TokenRParen) {
			return nil, p.unexpected(")")
		}
		if p.checkOperator() {
			op, err := p.parseOperator()
			if err != nil {
				return nil, err
			}
			value, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			return &FuncComparison{Fn: property, Args: args, Operator: op, Value: value, Pos: propPos}, nil
		}
		return &CallExpression{Fn: property, Args: args, Pos: propPos}, nil
	}

	op, err := p.parseOperator()
	if err != nil {
		return nil, err
	}

	value, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	return &ComparisonExpression{
		Property: property,
		Operator: op,
		Value:    value,
		Pos:      propPos,
	}, nil
}

func (p *Parser) parseProperty() (string, int, error) {
	token := p.current()
	if token.Type != TokenIdentifier {
		return "", 0, p.unexpected("property name")
	}
	p.pos++
	return token.Value, token.Pos, nil
}

// parseArgs parses a comma-separated argument list (numbers, strings,
// identifiers used as bare values).
func (p *Parser) parseArgs() ([]any, error) {
	var args []any
	if p.check(TokenRParen) {
		return args, nil
	}
	for {
		token := p.current()
		switch token.Type {
		case TokenNumber:
			p.pos++
			if strings.Contains(token.Value, ".") {
				f, err := strconv.ParseFloat(token.Value, 64)
				if err != nil {
					return nil, p.unexpected("number")
				}
				args = append(args, f)
			} else {
				i, err := strconv.ParseInt(token.Value, 10, 64)
				if err != nil {
					return nil, p.unexpected("number")
				}
				args = append(args, int(i))
			}
		case TokenString:
			p.pos++
			args = append(args, strings.Trim(token.Value, "\"'"))
		case TokenMinus:
			p.pos++
			num := p.current()
			if num.Type != TokenNumber {
				return nil, p.unexpected("number after '-'")
			}
			p.pos++
			f, err := strconv.ParseFloat("-"+num.Value, 64)
			if err != nil {
				return nil, p.unexpected("number after '-'")
			}
			args = append(args, f)
		default:
			return nil, p.unexpected("argument")
		}
		if p.consume(TokenComma) {
			continue
		}
		return args, nil
	}
}

func (p *Parser) checkOperator() bool {
	switch p.current().Type {
	case TokenEquals, TokenNotEquals, TokenGreaterThan, TokenGreaterEquals,
		TokenLessThan, TokenLessEquals, TokenContains, TokenMatches, TokenIn:
		return true
	}
	return false
}

func (p *Parser) parseOperator() (Operator, error) {
	token := p.current()
	p.pos++

	switch token.Type {
	case TokenEquals:
		return OpEquals, nil
	case TokenNotEquals:
		return OpNotEquals, nil
	case TokenGreaterThan:
		return OpGreaterThan, nil
	case TokenGreaterEquals:
		return OpGreaterEquals, nil
	case TokenLessThan:
		return OpLessThan, nil
	case TokenLessEquals:
		return OpLessEquals, nil
	case TokenContains:
		return OpContains, nil
	case TokenMatches:
		return OpMatches, nil
	case TokenIn:
		return OpIn, nil
	default:
		return OpEquals, p.unexpected("operator")
	}
}

func (p *Parser) parseValue() (any, error) {
	token := p.current()

	switch token.Type {
	case TokenString:
		p.pos++
		return strings.Trim(token.Value, "\"'"), nil

	case TokenNumber:
		p.pos++
		if strings.Contains(token.Value, ".") {
			return strconv.ParseFloat(token.Value, 64)
		}
		i, err := strconv.ParseInt(token.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		return int(i), nil

	case TokenLParen:
		p.pos++
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if !p.consume(TokenRParen) {
			return nil, p.unexpected(")")
		}
		return expr, nil

	case TokenLBracket:
		p.pos++
		values := make([]any, 0)
		for !p.check(TokenRBracket) {
			value, err := p.parseValue()
			if err != nil {
				return nil, err
			}
			values = append(values, value)
			if p.consume(TokenComma) {
				continue
			}
		}
		if !p.consume(TokenRBracket) {
			return nil, p.unexpected("]")
		}
		return values, nil

	default:
		return nil, p.unexpected("value")
	}
}

func (p *Parser) current() Token {
	return p.tokens[p.pos]
}

func (p *Parser) consume(expected TokenType) bool {
	if p.current().Type == expected {
		p.pos++
		return true
	}
	return false
}

func (p *Parser) check(t TokenType) bool {
	return p.current().Type == t
}

func (p *Parser) unexpected(expected string) error {
	token := p.current()
	return fmt.Errorf("expected %s at position %d, got %s", expected, token.Pos, token)
}

