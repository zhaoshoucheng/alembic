package streaming

import (
	"encoding/json"
	"strings"
	"unicode/utf8"
)

type JsonStreamParser struct {
	buf     strings.Builder
	current map[string]any
	raw     string
}

func NewJsonStreamParser() *JsonStreamParser {
	return &JsonStreamParser{}
}

type ParseResult struct {
	Value   map[string]any
	Raw     string
	Changed bool
}

func (p *JsonStreamParser) Parse(chunk string) ParseResult {
	p.buf.WriteString(chunk)
	p.raw = p.buf.String()

	value, ok := parsePartialJSON(p.raw)
	if !ok {
		return ParseResult{Value: p.current, Raw: p.raw, Changed: false}
	}

	changed := !jsonEqual(p.current, value)
	if changed {
		p.current = value
	}
	return ParseResult{Value: value, Raw: p.raw, Changed: changed}
}

func (p *JsonStreamParser) CurrentValue() map[string]any {
	return p.current
}

func (p *JsonStreamParser) Reset() {
	p.buf.Reset()
	p.current = nil
	p.raw = ""
}

func parsePartialJSON(raw string) (map[string]any, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, false
	}

	var result map[string]any
	if err := json.Unmarshal([]byte(raw), &result); err == nil {
		return result, true
	}

	repaired := repairJSON(raw)
	if err := json.Unmarshal([]byte(repaired), &result); err == nil {
		return result, true
	}
	return nil, false
}

func repairJSON(raw string) string {
	var b strings.Builder
	b.Grow(len(raw) + 16)

	inString := false
	escaped := false
	var stack []rune

	for i := 0; i < len(raw); {
		r, size := utf8.DecodeRuneInString(raw[i:])
		if r == utf8.RuneError && size <= 1 {
			break
		}

		if inString {
			if escaped {
				b.WriteRune(r)
				escaped = false
			} else if r == '\\' {
				b.WriteRune(r)
				escaped = true
			} else if r == '"' {
				b.WriteRune(r)
				inString = false
			} else {
				b.WriteRune(r)
			}
		} else {
			switch r {
			case '"':
				b.WriteRune(r)
				inString = true
			case '{':
				b.WriteRune(r)
				stack = append(stack, '}')
			case '[':
				b.WriteRune(r)
				stack = append(stack, ']')
			case '}', ']':
				b.WriteRune(r)
				if len(stack) > 0 {
					stack = stack[:len(stack)-1]
				}
			default:
				b.WriteRune(r)
			}
		}
		i += size
	}

	if inString {
		b.WriteByte('"')
	}

	// 关闭 trailing comma before closing brackets
	result := b.String()
	result = strings.TrimRight(result, " \t\n\r")
	result = strings.TrimRight(result, ",")

	for i := len(stack) - 1; i >= 0; i-- {
		result += string(stack[i])
	}

	return result
}

func jsonEqual(a, b map[string]any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	aj, err1 := json.Marshal(a)
	bj, err2 := json.Marshal(b)
	if err1 != nil || err2 != nil {
		return false
	}
	return string(aj) == string(bj)
}
