package formula

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// TokenType represents the type of a lexical token
type TokenType int

const (
	TokenNumber TokenType = iota
	TokenString
	TokenIdent
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
	TokenLParen
	TokenRParen
	TokenComma
	TokenEQ
	TokenNEQ
	TokenLT
	TokenGT
	TokenLTE
	TokenGTE
	TokenEOF
)

type Token struct {
	Type  TokenType
	Value string
}

// Tokenize breaks an expression string into tokens
func Tokenize(expr string) []Token {
	var tokens []Token
	i := 0
	for i < len(expr) {
		ch := expr[i]
		if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' {
			i++
			continue
		}
		if ch == '+' {
			tokens = append(tokens, Token{TokenPlus, "+"})
			i++
		} else if ch == '-' {
			tokens = append(tokens, Token{TokenMinus, "-"})
			i++
		} else if ch == '*' {
			tokens = append(tokens, Token{TokenStar, "*"})
			i++
		} else if ch == '/' {
			tokens = append(tokens, Token{TokenSlash, "/"})
			i++
		} else if ch == '(' {
			tokens = append(tokens, Token{TokenLParen, "("})
			i++
		} else if ch == ')' {
			tokens = append(tokens, Token{TokenRParen, ")"})
			i++
		} else if ch == ',' {
			tokens = append(tokens, Token{TokenComma, ","})
			i++
		} else if ch == '=' {
			tokens = append(tokens, Token{TokenEQ, "="})
			i++
		} else if ch == '!' && i+1 < len(expr) && expr[i+1] == '=' {
			tokens = append(tokens, Token{TokenNEQ, "!="})
			i += 2
		} else if ch == '<' {
			if i+1 < len(expr) && expr[i+1] == '=' {
				tokens = append(tokens, Token{TokenLTE, "<="})
				i += 2
			} else {
				tokens = append(tokens, Token{TokenLT, "<"})
				i++
			}
		} else if ch == '>' {
			if i+1 < len(expr) && expr[i+1] == '=' {
				tokens = append(tokens, Token{TokenGTE, ">="})
				i += 2
			} else {
				tokens = append(tokens, Token{TokenGT, ">"})
				i++
			}
		} else if ch == '"' || ch == '\'' {
			quote := ch
			j := i + 1
			for j < len(expr) && expr[j] != quote {
				j++
			}
			tokens = append(tokens, Token{TokenString, expr[i+1 : j]})
			if j < len(expr) {
				j++
			}
			i = j
		} else if ch >= '0' && ch <= '9' || ch == '.' {
			j := i
			for j < len(expr) && (expr[j] >= '0' && expr[j] <= '9' || expr[j] == '.') {
				j++
			}
			tokens = append(tokens, Token{TokenNumber, expr[i:j]})
			i = j
		} else if unicode.IsLetter(rune(ch)) || ch == '_' {
			j := i
			for j < len(expr) && (unicode.IsLetter(rune(expr[j])) || unicode.IsDigit(rune(expr[j])) || expr[j] == '_') {
				j++
			}
			tokens = append(tokens, Token{TokenIdent, expr[i:j]})
			i = j
		} else {
			i++
		}
	}
	tokens = append(tokens, Token{TokenEOF, ""})
	return tokens
}

// AST node types
type Node interface {
	nodeType() string
}

type NumberNode struct{ Value float64 }
type StringNode struct{ Value string }
type IdentNode struct{ Name string }
type BinaryNode struct {
	Op    string
	Left  Node
	Right Node
}
type FuncCallNode struct {
	Name string
	Args []Node
}
type IfNode struct {
	Cond Node
	Then Node
	Else Node
}

func (NumberNode) nodeType() string   { return "number" }
func (StringNode) nodeType() string   { return "string" }
func (IdentNode) nodeType() string    { return "ident" }
func (BinaryNode) nodeType() string   { return "binary" }
func (FuncCallNode) nodeType() string { return "func" }
func (IfNode) nodeType() string       { return "if" }

// Parser
type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
}

func (p *Parser) peek() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{TokenEOF, ""}
}

func (p *Parser) advance() Token {
	t := p.peek()
	p.pos++
	return t
}

func (p *Parser) expect(tt TokenType) Token {
	t := p.advance()
	if t.Type != tt {
		panic(fmt.Sprintf("expected token type %d, got %d (%s)", tt, t.Type, t.Value))
	}
	return t
}

func Parse(expr string) (Node, error) {
	tokens := Tokenize(expr)
	p := NewParser(tokens)
	defer func() {
		if r := recover(); r != nil {
			// handled below
		}
	}()
	node := p.parseExpr()
	return node, nil
}

func (p *Parser) parseExpr() Node {
	return p.parseComparison()
}

func (p *Parser) parseComparison() Node {
	left := p.parseAddSub()
	for {
		t := p.peek()
		if t.Type == TokenEQ || t.Type == TokenNEQ || t.Type == TokenLT ||
			t.Type == TokenGT || t.Type == TokenLTE || t.Type == TokenGTE {
			p.advance()
			right := p.parseAddSub()
			left = BinaryNode{Op: t.Value, Left: left, Right: right}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseAddSub() Node {
	left := p.parseMulDiv()
	for {
		t := p.peek()
		if t.Type == TokenPlus || t.Type == TokenMinus {
			p.advance()
			right := p.parseMulDiv()
			left = BinaryNode{Op: t.Value, Left: left, Right: right}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parseMulDiv() Node {
	left := p.parsePrimary()
	for {
		t := p.peek()
		if t.Type == TokenStar || t.Type == TokenSlash {
			p.advance()
			right := p.parsePrimary()
			left = BinaryNode{Op: t.Value, Left: left, Right: right}
		} else {
			break
		}
	}
	return left
}

func (p *Parser) parsePrimary() Node {
	t := p.peek()

	switch t.Type {
	case TokenNumber:
		p.advance()
		v, _ := strconv.ParseFloat(t.Value, 64)
		return NumberNode{Value: v}

	case TokenString:
		p.advance()
		return StringNode{Value: t.Value}

	case TokenLParen:
		p.advance()
		node := p.parseExpr()
		p.expect(TokenRParen)
		return node

	case TokenMinus:
		p.advance()
		operand := p.parsePrimary()
		return BinaryNode{Op: "*", Left: NumberNode{Value: -1}, Right: operand}

	case TokenIdent:
		p.advance()
		name := strings.ToUpper(t.Value)

		// IF/THEN/ELSE
		if name == "IF" {
			p.expect(TokenLParen)
			cond := p.parseExpr()
			p.expect(TokenComma)
			then := p.parseExpr()
			p.expect(TokenComma)
			els := p.parseExpr()
			p.expect(TokenRParen)
			return IfNode{Cond: cond, Then: then, Else: els}
		}

		// Function call
		if p.peek().Type == TokenLParen {
			p.advance()
			var args []Node
			if p.peek().Type != TokenRParen {
				args = append(args, p.parseExpr())
				for p.peek().Type == TokenComma {
					p.advance()
					args = append(args, p.parseExpr())
				}
			}
			p.expect(TokenRParen)
			return FuncCallNode{Name: name, Args: args}
		}

		// Column reference (use original case)
		return IdentNode{Name: t.Value}
	}

	p.advance()
	return NumberNode{Value: 0}
}

// Value represents an evaluation result
type Value struct {
	Number float64
	Str    string
	Bool   bool
	IsStr  bool
	IsBool bool
}

func NumVal(f float64) Value  { return Value{Number: f} }
func StrVal(s string) Value   { return Value{Str: s, IsStr: true} }
func BoolVal(b bool) Value    { return Value{Bool: b, IsBool: true} }

func (v Value) AsFloat() float64 {
	if v.IsBool {
		if v.Bool {
			return 1
		}
		return 0
	}
	if v.IsStr {
		f, _ := strconv.ParseFloat(v.Str, 64)
		return f
	}
	return v.Number
}

func (v Value) AsString() string {
	if v.IsStr {
		return v.Str
	}
	if v.IsBool {
		if v.Bool {
			return "true"
		}
		return "false"
	}
	if v.Number == math.Floor(v.Number) {
		return strconv.FormatInt(int64(v.Number), 10)
	}
	return strconv.FormatFloat(v.Number, 'f', 2, 64)
}

func (v Value) Truthy() bool {
	if v.IsBool {
		return v.Bool
	}
	if v.IsStr {
		return v.Str != ""
	}
	return v.Number != 0
}

// Evaluate evaluates a formula against an object set.
// Returns a single aggregated value.
func Evaluate(node Node, objects []map[string]interface{}) Value {
	switch n := node.(type) {
	case NumberNode:
		return NumVal(n.Value)
	case StringNode:
		return StrVal(n.Value)
	case IdentNode:
		// Column reference: return value from first object (for scalar formulas)
		if len(objects) > 0 {
			return toValue(objects[0][n.Name])
		}
		return NumVal(0)
	case BinaryNode:
		left := Evaluate(n.Left, objects)
		right := Evaluate(n.Right, objects)
		return evalBinary(n.Op, left, right)
	case IfNode:
		cond := Evaluate(n.Cond, objects)
		if cond.Truthy() {
			return Evaluate(n.Then, objects)
		}
		return Evaluate(n.Else, objects)
	case FuncCallNode:
		return evalFunc(n.Name, n.Args, objects)
	}
	return NumVal(0)
}

// EvaluatePerRow evaluates a formula for each object, returning per-row results
func EvaluatePerRow(node Node, objects []map[string]interface{}) []Value {
	results := make([]Value, len(objects))
	for i, obj := range objects {
		results[i] = Evaluate(node, []map[string]interface{}{obj})
	}
	return results
}

func evalBinary(op string, left, right Value) Value {
	switch op {
	case "+":
		return NumVal(left.AsFloat() + right.AsFloat())
	case "-":
		return NumVal(left.AsFloat() - right.AsFloat())
	case "*":
		return NumVal(left.AsFloat() * right.AsFloat())
	case "/":
		r := right.AsFloat()
		if r == 0 {
			return NumVal(0)
		}
		return NumVal(left.AsFloat() / r)
	case "=":
		if left.IsStr || right.IsStr {
			return BoolVal(left.AsString() == right.AsString())
		}
		return BoolVal(left.AsFloat() == right.AsFloat())
	case "!=":
		if left.IsStr || right.IsStr {
			return BoolVal(left.AsString() != right.AsString())
		}
		return BoolVal(left.AsFloat() != right.AsFloat())
	case "<":
		return BoolVal(left.AsFloat() < right.AsFloat())
	case ">":
		return BoolVal(left.AsFloat() > right.AsFloat())
	case "<=":
		return BoolVal(left.AsFloat() <= right.AsFloat())
	case ">=":
		return BoolVal(left.AsFloat() >= right.AsFloat())
	}
	return NumVal(0)
}

func evalFunc(name string, args []Node, objects []map[string]interface{}) Value {
	switch name {
	case "SUM":
		if len(args) == 0 {
			return NumVal(0)
		}
		field := getFieldName(args[0])
		var s float64
		for _, o := range objects {
			s += toValue(o[field]).AsFloat()
		}
		return NumVal(s)

	case "COUNT":
		return NumVal(float64(len(objects)))

	case "AVG":
		if len(args) == 0 || len(objects) == 0 {
			return NumVal(0)
		}
		field := getFieldName(args[0])
		var s float64
		for _, o := range objects {
			s += toValue(o[field]).AsFloat()
		}
		return NumVal(s / float64(len(objects)))

	case "MIN":
		if len(args) == 0 || len(objects) == 0 {
			return NumVal(0)
		}
		field := getFieldName(args[0])
		m := toValue(objects[0][field]).AsFloat()
		for _, o := range objects[1:] {
			v := toValue(o[field]).AsFloat()
			if v < m {
				m = v
			}
		}
		return NumVal(m)

	case "MAX":
		if len(args) == 0 || len(objects) == 0 {
			return NumVal(0)
		}
		field := getFieldName(args[0])
		m := toValue(objects[0][field]).AsFloat()
		for _, o := range objects[1:] {
			v := toValue(o[field]).AsFloat()
			if v > m {
				m = v
			}
		}
		return NumVal(m)

	case "DAYS_BETWEEN":
		if len(args) < 2 || len(objects) == 0 {
			return NumVal(0)
		}
		f1 := getFieldName(args[0])
		f2 := getFieldName(args[1])
		t1 := parseTime(fmt.Sprintf("%v", objects[0][f1]))
		t2 := parseTime(fmt.Sprintf("%v", objects[0][f2]))
		return NumVal(math.Abs(t2.Sub(t1).Hours() / 24))

	case "MONTH":
		if len(args) == 0 || len(objects) == 0 {
			return NumVal(0)
		}
		field := getFieldName(args[0])
		t := parseTime(fmt.Sprintf("%v", objects[0][field]))
		return NumVal(float64(t.Month()))

	case "YEAR":
		if len(args) == 0 || len(objects) == 0 {
			return NumVal(0)
		}
		field := getFieldName(args[0])
		t := parseTime(fmt.Sprintf("%v", objects[0][field]))
		return NumVal(float64(t.Year()))

	case "ABS":
		if len(args) > 0 {
			return NumVal(math.Abs(Evaluate(args[0], objects).AsFloat()))
		}
		return NumVal(0)

	case "ROUND":
		if len(args) > 0 {
			return NumVal(math.Round(Evaluate(args[0], objects).AsFloat()))
		}
		return NumVal(0)
	}
	return NumVal(0)
}

func getFieldName(n Node) string {
	switch v := n.(type) {
	case IdentNode:
		return v.Name
	case StringNode:
		return v.Value
	}
	return ""
}

func toValue(v interface{}) Value {
	if v == nil {
		return NumVal(0)
	}
	switch n := v.(type) {
	case float64:
		return NumVal(n)
	case int:
		return NumVal(float64(n))
	case int64:
		return NumVal(float64(n))
	case string:
		if f, err := strconv.ParseFloat(n, 64); err == nil {
			return NumVal(f)
		}
		return StrVal(n)
	case bool:
		return BoolVal(n)
	}
	return StrVal(fmt.Sprintf("%v", v))
}

func parseTime(s string) time.Time {
	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
		"01/02/2006",
	} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
