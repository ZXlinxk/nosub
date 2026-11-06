package schema

import "strings"

func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_'
}

// splitStatements splits SQL text into individual statements delimited by semicolons.
func splitStatements(sql string) []string {
	var stmts []string
	var b strings.Builder
	n := len(sql)
	i := 0
	depth := 0
	lineStart := true

	for i < n {
		c := sql[i]

		if lineStart && (c == ' ' || c == '\t' || c == '\r') {
			b.WriteByte(c)
			i++
			continue
		}

		if lineStart && c == '#' {
			for i < n && sql[i] != '\n' {
				i++
			}
			if i < n && sql[i] == '\n' {
				b.WriteByte('\n')
				i++
				lineStart = true
			}
			continue
		}

		if i+1 < n && c == '-' && sql[i+1] == '-' {
			for i < n && sql[i] != '\n' {
				i++
			}
			if i < n && sql[i] == '\n' {
				b.WriteByte('\n')
				i++
				lineStart = true
			}
			continue
		}

		if i+1 < n && c == '/' && sql[i+1] == '*' {
			i += 2
			for i+1 < n && !(sql[i] == '*' && sql[i+1] == '/') {
				if sql[i] == '\n' {
					b.WriteByte('\n')
				}
				i++
			}
			if i+1 < n {
				i += 2
			} else {
				i = n
			}
			b.WriteByte(' ')
			continue
		}

		if c == '\'' {
			b.WriteByte(c)
			i++
			for i < n {
				ch := sql[i]
				if ch == '\\' && i+1 < n {
					b.WriteByte(ch)
					b.WriteByte(sql[i+1])
					i += 2
					continue
				}
				if ch == '\'' {
					b.WriteByte(ch)
					i++
					if i < n && sql[i] == '\'' {
						b.WriteByte(sql[i])
						i++
						continue
					}
					break
				}
				b.WriteByte(ch)
				i++
			}
			lineStart = false
			continue
		}

		if c == '"' || c == '`' || c == '[' {
			closeCh := c
			if c == '[' {
				closeCh = ']'
			}
			b.WriteByte(c)
			i++
			for i < n {
				ch := sql[i]
				b.WriteByte(ch)
				i++
				if ch == closeCh {
					if closeCh != ']' && i < n && sql[i] == closeCh {
						b.WriteByte(sql[i])
						i++
						continue
					}
					break
				}
			}
			lineStart = false
			continue
		}

		if c == '$' {
			j := i + 1
			for j < n && isIdentChar(sql[j]) {
				j++
			}
			if j < n && sql[j] == '$' {
				tag := sql[i : j+1]
				b.WriteString(tag)
				i = j + 1
				closeIdx := strings.Index(sql[i:], tag)
				if closeIdx != -1 {
					b.WriteString(sql[i : i+closeIdx+len(tag)])
					i += closeIdx + len(tag)
				} else {
					b.WriteString(sql[i:])
					i = n
				}
				lineStart = false
				continue
			}
		}

		if c == '(' {
			depth++
			b.WriteByte(c)
			i++
			lineStart = false
			continue
		}

		if c == ')' {
			if depth > 0 {
				depth--
			}
			b.WriteByte(c)
			i++
			lineStart = false
			continue
		}

		if c == ';' && depth == 0 {
			stmt := strings.TrimSpace(b.String())
			if stmt != "" {
				stmts = append(stmts, stmt)
			}
			b.Reset()
			i++
			lineStart = false
			continue
		}

		if c == '\n' {
			lineStart = true
		} else {
			lineStart = false
		}
		b.WriteByte(c)
		i++
	}

	stmt := strings.TrimSpace(b.String())
	if stmt != "" {
		stmts = append(stmts, stmt)
	}
	return stmts
}

type tokenType int

const (
	tokWord tokenType = iota
	tokString
	tokParenOpen
	tokParenClose
	tokComma
	tokEqual
	tokDot
	tokColonColon
)

type token struct {
	typ  tokenType
	text string
	raw  string
}

func tokenize(s string) []token {
	var tokens []token
	n := len(s)
	i := 0

	for i < n {
		c := s[i]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			i++
			continue
		}

		if c == '\'' {
			start := i
			i++
			var val strings.Builder
			for i < n {
				ch := s[i]
				if ch == '\\' && i+1 < n {
					val.WriteByte(s[i+1])
					i += 2
					continue
				}
				if ch == '\'' {
					i++
					if i < n && s[i] == '\'' {
						val.WriteByte('\'')
						i++
						continue
					}
					break
				}
				val.WriteByte(ch)
				i++
			}
			tokens = append(tokens, token{
				typ:  tokString,
				text: s[start:i],
				raw:  val.String(),
			})
			continue
		}

		if c == '"' || c == '`' || c == '[' {
			closeCh := c
			if c == '[' {
				closeCh = ']'
			}
			start := i
			i++
			var val strings.Builder
			for i < n {
				ch := s[i]
				if ch == closeCh {
					i++
					if closeCh != ']' && i < n && s[i] == closeCh {
						val.WriteByte(ch)
						i++
						continue
					}
					break
				}
				val.WriteByte(ch)
				i++
			}
			tokens = append(tokens, token{
				typ:  tokWord,
				text: s[start:i],
				raw:  val.String(),
			})
			continue
		}

		if c == '(' {
			tokens = append(tokens, token{typ: tokParenOpen, text: "("})
			i++
			continue
		}
		if c == ')' {
			tokens = append(tokens, token{typ: tokParenClose, text: ")"})
			i++
			continue
		}
		if c == ',' {
			tokens = append(tokens, token{typ: tokComma, text: ","})
			i++
			continue
		}
		if c == '=' {
			tokens = append(tokens, token{typ: tokEqual, text: "="})
			i++
			continue
		}
		if c == '.' {
			tokens = append(tokens, token{typ: tokDot, text: "."})
			i++
			continue
		}
		if i+1 < n && c == ':' && s[i+1] == ':' {
			tokens = append(tokens, token{typ: tokColonColon, text: "::"})
			i += 2
			continue
		}

		start := i
		for i < n {
			ch := s[i]
			if ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r' ||
				ch == '(' || ch == ')' || ch == ',' || ch == '=' || ch == '.' ||
				ch == '\'' || ch == '"' || ch == '`' || ch == '[' || ch == ']' ||
				ch == ';' {
				break
			}
			if i+1 < n && ch == ':' && s[i+1] == ':' {
				break
			}
			i++
		}
		if i == start {
			i++
			continue
		}
		w := s[start:i]
		tokens = append(tokens, token{
			typ:  tokWord,
			text: w,
			raw:  w,
		})
	}
	return tokens
}

func consumeQualifiedName(tokens []token, pos *int) string {
	if *pos >= len(tokens) {
		return ""
	}
	var parts []string
	parts = append(parts, tokens[*pos].raw)
	*pos++
	for *pos+1 < len(tokens) && tokens[*pos].typ == tokDot {
		*pos++
		parts = append(parts, tokens[*pos].raw)
		*pos++
	}
	return cleanName(strings.Join(parts, "."))
}
