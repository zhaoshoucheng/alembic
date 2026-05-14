package builtin

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"github.com/shoucheng/my-first-agent/internal/tools"
)

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (c *Calculator) Name() string { return "calculator" }

func (c *Calculator) Description() string {
	return tools.NormalizeDescription(
		"Evaluate a deterministic arithmetic expression.",
		"Supports +, -, *, /, parentheses, decimals, and unary plus/minus.",
	)
}

func (c *Calculator) Parameters() tools.JSONSchema {
	return tools.ObjectSchema(map[string]any{
		"expression": tools.StringProperty("Arithmetic expression to evaluate, for example: (25 * 4) + 10"),
	}, "expression")
}

func (c *Calculator) Execute(_ context.Context, args map[string]any) (string, error) {
	expression, _ := args["expression"].(string)
	value, err := evalExpression(expression)
	if err != nil {
		return "", err
	}
	return strconv.FormatFloat(value, 'f', -1, 64), nil
}

type expressionParser struct {
	input []rune
	pos   int
}

func evalExpression(input string) (float64, error) {
	p := &expressionParser{input: []rune(input)}
	value, err := p.parseExpression()
	if err != nil {
		return 0, err
	}
	p.skipSpaces()
	if !p.eof() {
		return 0, fmt.Errorf("unexpected token %q at position %d", string(p.peek()), p.pos)
	}
	return value, nil
}

func (p *expressionParser) parseExpression() (float64, error) {
	left, err := p.parseTerm()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		switch p.peek() {
		case '+':
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left += right
		case '-':
			p.pos++
			right, err := p.parseTerm()
			if err != nil {
				return 0, err
			}
			left -= right
		default:
			return left, nil
		}
	}
}

func (p *expressionParser) parseTerm() (float64, error) {
	left, err := p.parseFactor()
	if err != nil {
		return 0, err
	}
	for {
		p.skipSpaces()
		switch p.peek() {
		case '*':
			p.pos++
			right, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			left *= right
		case '/':
			p.pos++
			right, err := p.parseFactor()
			if err != nil {
				return 0, err
			}
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			left /= right
		default:
			return left, nil
		}
	}
}

func (p *expressionParser) parseFactor() (float64, error) {
	p.skipSpaces()
	switch p.peek() {
	case '+':
		p.pos++
		return p.parseFactor()
	case '-':
		p.pos++
		value, err := p.parseFactor()
		return -value, err
	case '(':
		p.pos++
		value, err := p.parseExpression()
		if err != nil {
			return 0, err
		}
		p.skipSpaces()
		if p.peek() != ')' {
			return 0, fmt.Errorf("missing closing parenthesis")
		}
		p.pos++
		return value, nil
	default:
		return p.parseNumber()
	}
}

func (p *expressionParser) parseNumber() (float64, error) {
	p.skipSpaces()
	start := p.pos
	for !p.eof() {
		r := p.peek()
		if !unicode.IsDigit(r) && r != '.' {
			break
		}
		p.pos++
	}
	if start == p.pos {
		if p.eof() {
			return 0, fmt.Errorf("unexpected end of expression")
		}
		return 0, fmt.Errorf("expected number at position %d, got %q", p.pos, string(p.peek()))
	}
	raw := strings.TrimSpace(string(p.input[start:p.pos]))
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid number %q: %w", raw, err)
	}
	return value, nil
}

func (p *expressionParser) skipSpaces() {
	for !p.eof() && unicode.IsSpace(p.peek()) {
		p.pos++
	}
}

func (p *expressionParser) eof() bool {
	return p.pos >= len(p.input)
}

func (p *expressionParser) peek() rune {
	if p.eof() {
		return 0
	}
	return p.input[p.pos]
}
