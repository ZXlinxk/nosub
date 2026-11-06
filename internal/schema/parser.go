package schema

import (
	"cmp"
	"slices"
	"strings"
)

func findTable(s *Schema, name string) *Table {
	for i := range s.Tables {
		if s.Tables[i].Name == name {
			return &s.Tables[i]
		}
	}
	return nil
}

func findCol(tb *Table, name string) *Column {
	for i := range tb.Columns {
		if tb.Columns[i].Name == name {
			return &tb.Columns[i]
		}
	}
	return nil
}

func isModifierStarter(tokens []token, i int) bool {
	if i >= len(tokens) {
		return false
	}
	w := strings.ToUpper(tokens[i].text)
	switch w {
	case "NOT":
		return i+1 < len(tokens) && strings.EqualFold(tokens[i+1].text, "NULL")
	case "NULL", "DEFAULT", "UNIQUE", "AUTO_INCREMENT", "AUTOINCREMENT", "COMMENT", "COLLATE", "CHARSET":
		return true
	case "PRIMARY":
		return i+1 < len(tokens) && strings.EqualFold(tokens[i+1].text, "KEY")
	case "GENERATED":
		return true
	case "REFERENCES":
		return true
	case "CHARACTER":
		return i+1 < len(tokens) && strings.EqualFold(tokens[i+1].text, "SET")
	case "CHECK", "CONSTRAINT":
		return true
	}
	return false
}

func formatType(tokens []token) string {
	var b strings.Builder
	depth := 0
	for i, tok := range tokens {
		switch {
		case tok.typ == tokParenOpen:
			depth++
			b.WriteString("(")
		case tok.typ == tokParenClose:
			depth--
			b.WriteString(")")
		case tok.typ == tokComma:
			b.WriteString(",")
			if depth == 0 {
				b.WriteString(" ")
			}
		case tok.typ == tokDot:
			b.WriteString(".")
		default:
			if i > 0 && !strings.HasPrefix(tok.text, "[") {
				prev := tokens[i-1].typ
				if prev != tokParenOpen && prev != tokDot && prev != tokComma {
					b.WriteString(" ")
				}
			}
			b.WriteString(tok.text)
		}
	}
	return strings.TrimPrefix(b.String(), "public.")
}

func formatDefault(tokens []token) string {
	if len(tokens) == 1 && strings.EqualFold(tokens[0].text, "NULL") {
		return ""
	}
	return joinExpr(tokens)
}

func joinExpr(tokens []token) string {
	var b strings.Builder
	for i, tok := range tokens {
		if i > 0 {
			prev := tokens[i-1].typ
			tight := prev == tokColonColon || tok.typ == tokColonColon ||
				prev == tokDot || tok.typ == tokDot ||
				prev == tokParenOpen || tok.typ == tokParenClose ||
				(tok.typ == tokParenOpen && prev == tokWord) ||
				strings.HasPrefix(tok.text, "[")
			if !tight {
				b.WriteString(" ")
			}
		}
		b.WriteString(tok.text)
	}
	return b.String()
}

// closingParen returns the index of the parenthesis matching the one at open, or -1.
func closingParen(tokens []token, open int) int {
	depth := 0
	for j := open; j < len(tokens); j++ {
		switch tokens[j].typ {
		case tokParenOpen:
			depth++
		case tokParenClose:
			depth--
			if depth == 0 {
				return j
			}
		}
	}
	return -1
}

func splitOnCommas(tokens []token) [][]token {
	var out [][]token
	var cur []token
	depth := 0
	for _, tok := range tokens {
		switch tok.typ {
		case tokParenOpen:
			depth++
		case tokParenClose:
			depth--
		}
		if tok.typ == tokComma && depth == 0 {
			out = append(out, cur)
			cur = nil
			continue
		}
		cur = append(cur, tok)
	}
	if len(cur) > 0 {
		out = append(out, cur)
	}
	return out
}

func wordAt(tokens []token, i int, words ...string) bool {
	if i+len(words) > len(tokens) {
		return false
	}
	for j, w := range words {
		if !strings.EqualFold(tokens[i+j].text, w) {
			return false
		}
	}
	return true
}

func parseReference(tokens []token, i *int, cols []string) ForeignKey {
	fk := ForeignKey{Columns: cols, RefTable: consumeQualifiedName(tokens, i)}
	fk.RefColumns = parseParenCols(tokens, i)
	for *i < len(tokens) {
		switch {
		case wordAt(tokens, *i, "MATCH"):
			*i += 2
		case wordAt(tokens, *i, "ON", "DELETE"):
			*i += 2
			fk.OnDelete = refAction(tokens, i)
		case wordAt(tokens, *i, "ON", "UPDATE"):
			*i += 2
			fk.OnUpdate = refAction(tokens, i)
		default:
			return fk
		}
	}
	return fk
}

func refAction(tokens []token, i *int) string {
	if *i >= len(tokens) {
		return ""
	}
	action := strings.ToUpper(tokens[*i].text)
	*i++
	if (action == "SET" || action == "NO") && *i < len(tokens) {
		action += " " + strings.ToUpper(tokens[*i].text)
		*i++
	}
	return action
}

func parseParenCols(tokens []token, pos *int) []string {
	if *pos >= len(tokens) || tokens[*pos].typ != tokParenOpen {
		return nil
	}
	end := closingParen(tokens, *pos)
	if end < 0 {
		*pos = len(tokens)
		return nil
	}
	var cols []string
	for _, item := range splitOnCommas(tokens[*pos+1 : end]) {
		// Anything after the name is a sort order or a MySQL prefix length.
		if len(item) > 0 {
			cols = append(cols, unquote(item[0].raw))
		}
	}
	*pos = end + 1
	return cols
}

func parseColumn(tokens []token, tb *Table) {
	colName := unquote(tokens[0].raw)
	col := Column{Name: colName, Nullable: true}

	modIdx := len(tokens)
	depth := 0
	for i := 1; i < len(tokens); i++ {
		if tokens[i].typ == tokParenOpen {
			depth++
		} else if tokens[i].typ == tokParenClose {
			if depth > 0 {
				depth--
			}
		} else if depth == 0 && isModifierStarter(tokens, i) {
			modIdx = i
			break
		}
	}

	col.Type = formatType(tokens[1:modIdx])
	if strings.Contains(strings.ToUpper(col.Type), "SERIAL") {
		col.AutoIncrement = true
	}

	i := modIdx
	for i < len(tokens) {
		w := strings.ToUpper(tokens[i].text)
		switch w {
		case "NOT":
			if i+1 < len(tokens) && strings.EqualFold(tokens[i+1].text, "NULL") {
				col.Nullable = false
				i += 2
				continue
			}
			i++
		case "NULL":
			col.Nullable = true
			i++
		case "PRIMARY":
			if i+1 < len(tokens) && strings.EqualFold(tokens[i+1].text, "KEY") {
				tb.PrimaryKey = append(tb.PrimaryKey, col.Name)
				col.Nullable = false
				i += 2
				continue
			}
			i++
		case "UNIQUE":
			col.Unique = true
			i++
			if i < len(tokens) && strings.EqualFold(tokens[i].text, "KEY") {
				i++
			}
		case "AUTO_INCREMENT", "AUTOINCREMENT":
			col.AutoIncrement = true
			i++
		case "GENERATED":
			for i < len(tokens) && !(i > modIdx && isModifierStarter(tokens, i)) {
				col.AutoIncrement = col.AutoIncrement || wordAt(tokens, i, "IDENTITY")
				i++
			}
		case "COMMENT":
			i++
			if i < len(tokens) && tokens[i].typ == tokEqual {
				i++
			}
			if i < len(tokens) && tokens[i].typ == tokString {
				col.Comment = tokens[i].raw
				i++
			}
		case "COLLATE":
			i += 2
		case "CHARSET":
			i += 2
		case "CHARACTER":
			if i+2 < len(tokens) && strings.EqualFold(tokens[i+1].text, "SET") {
				i += 3
			} else {
				i++
			}
		case "CHECK":
			i++
			if i < len(tokens) && tokens[i].typ == tokParenOpen {
				d := 1
				i++
				for i < len(tokens) && d > 0 {
					if tokens[i].typ == tokParenOpen {
						d++
					} else if tokens[i].typ == tokParenClose {
						d--
					}
					i++
				}
			}
		case "REFERENCES":
			i++
			tb.ForeignKeys = append(tb.ForeignKeys, parseReference(tokens, &i, []string{col.Name}))
		case "DEFAULT":
			i++
			defStart := i
			d := 0
			for i < len(tokens) {
				if tokens[i].typ == tokParenOpen {
					d++
				} else if tokens[i].typ == tokParenClose {
					if d > 0 {
						d--
					}
				} else if d == 0 && isModifierStarter(tokens, i) {
					break
				}
				i++
			}
			col.Default = formatDefault(tokens[defStart:i])
		default:
			i++
		}
	}

	tb.Columns = append(tb.Columns, col)
}

// isIndexClause tells "KEY idx (a)" from a column that happens to be called key,
// like "key varchar(10)": only an index has column names inside the parentheses.
func isIndexClause(tokens []token, i int) bool {
	for j := i + 1; j <= i+2 && j+1 < len(tokens); j++ {
		if tokens[j].typ == tokParenOpen {
			next := tokens[j+1]
			return next.typ == tokWord && !(next.text[0] >= '0' && next.text[0] <= '9')
		}
	}
	return false
}

func addUnique(tb *Table, name string, cols []string) {
	if len(cols) == 1 {
		if c := findCol(tb, cols[0]); c != nil {
			c.Unique = true
		}
		return
	}
	if len(cols) > 1 {
		tb.Indexes = append(tb.Indexes, Index{Name: name, Columns: cols, Unique: true})
	}
}

func parseTableElement(tokens []token, tb *Table) {
	if len(tokens) == 0 {
		return
	}
	i := 0
	var name string
	if wordAt(tokens, 0, "CONSTRAINT") {
		if len(tokens) < 3 {
			return
		}
		name = unquote(tokens[1].raw)
		i = 2
	}
	indexName := func() string {
		if i < len(tokens) && tokens[i].typ == tokWord {
			i++
			return unquote(tokens[i-1].raw)
		}
		return name
	}

	switch {
	case wordAt(tokens, i, "PRIMARY", "KEY"):
		i += 2
		tb.PrimaryKey = append(tb.PrimaryKey, parseParenCols(tokens, &i)...)
		for _, pk := range tb.PrimaryKey {
			if c := findCol(tb, pk); c != nil {
				c.Nullable = false
			}
		}
	case wordAt(tokens, i, "FOREIGN", "KEY"):
		i += 2
		indexName()
		cols := parseParenCols(tokens, &i)
		if wordAt(tokens, i, "REFERENCES") {
			i++
			tb.ForeignKeys = append(tb.ForeignKeys, parseReference(tokens, &i, cols))
		}
	case wordAt(tokens, i, "UNIQUE") && (name != "" || i+1 < len(tokens) && (tokens[i+1].typ == tokParenOpen || isIndexClause(tokens, i) || wordAt(tokens, i+1, "KEY") || wordAt(tokens, i+1, "INDEX"))):
		i++
		if wordAt(tokens, i, "KEY") || wordAt(tokens, i, "INDEX") {
			i++
		}
		idx := indexName()
		addUnique(tb, idx, parseParenCols(tokens, &i))
	case (wordAt(tokens, i, "KEY") || wordAt(tokens, i, "INDEX")) && isIndexClause(tokens, i):
		i++
		idx := indexName()
		tb.Indexes = append(tb.Indexes, Index{Name: idx, Columns: parseParenCols(tokens, &i)})
	case wordAt(tokens, i, "FULLTEXT", "KEY"), wordAt(tokens, i, "FULLTEXT", "INDEX"), wordAt(tokens, i, "SPATIAL", "KEY"), wordAt(tokens, i, "SPATIAL", "INDEX"):
		i += 2
		idx := indexName()
		tb.Indexes = append(tb.Indexes, Index{Name: idx, Columns: parseParenCols(tokens, &i)})
	case wordAt(tokens, i, "CHECK") && i+1 < len(tokens) && tokens[i+1].typ == tokParenOpen:
	default:
		if name == "" {
			parseColumn(tokens, tb)
		}
	}
}

func parseCreateTable(tokens []token, s *Schema) {
	i := 1
	for wordAt(tokens, i, "TEMP") || wordAt(tokens, i, "TEMPORARY") || wordAt(tokens, i, "UNLOGGED") {
		i++
	}
	if !wordAt(tokens, i, "TABLE") {
		return
	}
	i++
	if wordAt(tokens, i, "IF", "NOT", "EXISTS") {
		i += 3
	}
	name := consumeQualifiedName(tokens, &i)
	if name == "" || i >= len(tokens) || tokens[i].typ != tokParenOpen {
		return
	}
	end := closingParen(tokens, i)
	if end < 0 {
		return
	}

	tb := Table{Name: name}
	for j := end + 1; j < len(tokens); j++ {
		if wordAt(tokens, j, "COMMENT") {
			j++
			if j < len(tokens) && tokens[j].typ == tokEqual {
				j++
			}
			if j < len(tokens) && tokens[j].typ == tokString {
				tb.Comment = tokens[j].raw
			}
		}
	}
	for _, elem := range splitOnCommas(tokens[i+1 : end]) {
		parseTableElement(elem, &tb)
	}
	for _, pk := range tb.PrimaryKey {
		if c := findCol(&tb, pk); c != nil {
			c.Nullable = false
		}
	}
	s.Tables = append(s.Tables, tb)
}

func parseCreateIndex(tokens []token, s *Schema) {
	i := 1
	unique := wordAt(tokens, i, "UNIQUE")
	if unique {
		i++
	}
	if !wordAt(tokens, i, "INDEX") {
		return
	}
	i++
	if wordAt(tokens, i, "CONCURRENTLY") {
		i++
	}
	if wordAt(tokens, i, "IF", "NOT", "EXISTS") {
		i += 3
	}
	if i+1 >= len(tokens) {
		return
	}
	name := unquote(tokens[i].raw)
	i++
	if !wordAt(tokens, i, "ON") {
		return
	}
	i++
	if wordAt(tokens, i, "ONLY") {
		i++
	}
	tb := findTable(s, consumeQualifiedName(tokens, &i))
	if tb == nil {
		return
	}
	if wordAt(tokens, i, "USING") {
		i += 2
	}
	if i >= len(tokens) || tokens[i].typ != tokParenOpen {
		return
	}
	end := closingParen(tokens, i)
	if end < 0 {
		return
	}

	var cols []string
	expr := false
	for _, item := range splitOnCommas(tokens[i+1 : end]) {
		if slices.ContainsFunc(item, func(t token) bool { return t.typ == tokParenOpen }) {
			expr = true
			cols = append(cols, joinExpr(item))
		} else {
			cols = append(cols, unquote(item[0].raw))
		}
	}
	if unique && !expr {
		addUnique(tb, name, cols)
		return
	}
	tb.Indexes = append(tb.Indexes, Index{Name: name, Columns: cols, Unique: unique})
}

func parseAlterTable(tokens []token, s *Schema) {
	i := 2
	if wordAt(tokens, i, "IF", "EXISTS") {
		i += 2
	}
	if wordAt(tokens, i, "ONLY") {
		i++
	}
	tb := findTable(s, consumeQualifiedName(tokens, &i))
	if tb == nil || i >= len(tokens) {
		return
	}
	for _, action := range splitOnCommas(tokens[i:]) {
		switch {
		case wordAt(action, 0, "ADD", "COLUMN"):
			parseTableElement(action[2:], tb)
		case wordAt(action, 0, "ADD"):
			parseTableElement(action[1:], tb)
		case wordAt(action, 0, "MODIFY", "COLUMN"):
			modifyColumn(action[2:], tb)
		case wordAt(action, 0, "MODIFY"):
			modifyColumn(action[1:], tb)
		case wordAt(action, 0, "ALTER", "COLUMN"):
			alterColumn(action[2:], tb)
		}
	}
}

// modifyColumn handles MySQL dumps that declare AUTO_INCREMENT in a later ALTER.
func modifyColumn(tokens []token, tb *Table) {
	if len(tokens) == 0 {
		return
	}
	var parsed Table
	parseColumn(tokens, &parsed)
	if c := findCol(tb, parsed.Columns[0].Name); c != nil {
		n := *c
		*c = parsed.Columns[0]
		c.Unique, c.Comment = n.Unique, cmp.Or(c.Comment, n.Comment)
		c.Nullable = c.Nullable && !slices.Contains(tb.PrimaryKey, c.Name)
	}
}

// alterColumn handles how pg_dump attaches sequences and identities to serial columns.
func alterColumn(tokens []token, tb *Table) {
	if len(tokens) < 2 {
		return
	}
	c := findCol(tb, unquote(tokens[0].raw))
	if c == nil {
		return
	}
	rest := tokens[1:]
	if wordAt(rest, 0, "SET", "DEFAULT") {
		if wordAt(rest, 2, "nextval") {
			c.AutoIncrement = true
			return
		}
		c.Default = formatDefault(rest[2:])
	}
	if slices.ContainsFunc(rest, func(t token) bool { return strings.EqualFold(t.text, "IDENTITY") }) {
		c.AutoIncrement = true
	}
}

func parseComment(tokens []token, s *Schema) {
	if len(tokens) < 4 || !strings.EqualFold(tokens[1].text, "ON") {
		return
	}
	targetType := strings.ToUpper(tokens[2].text)
	i := 3

	if targetType == "TABLE" {
		tableName := consumeQualifiedName(tokens, &i)
		if i < len(tokens) && strings.EqualFold(tokens[i].text, "IS") {
			i++
			if i < len(tokens) && tokens[i].typ == tokString {
				if tb := findTable(s, tableName); tb != nil {
					tb.Comment = tokens[i].raw
				}
			}
		}
	} else if targetType == "COLUMN" {
		var parts []string
		for i < len(tokens) && !strings.EqualFold(tokens[i].text, "IS") {
			if tokens[i].typ != tokDot {
				parts = append(parts, tokens[i].raw)
			}
			i++
		}
		if len(parts) >= 2 {
			colName := unquote(parts[len(parts)-1])
			tableName := cleanName(strings.Join(parts[:len(parts)-1], "."))
			if i < len(tokens) && strings.EqualFold(tokens[i].text, "IS") {
				i++
				if i < len(tokens) && tokens[i].typ == tokString {
					if tb := findTable(s, tableName); tb != nil {
						if c := findCol(tb, colName); c != nil {
							c.Comment = tokens[i].raw
						}
					}
				}
			}
		}
	}
}

func parseCreateType(tokens []token, s *Schema) {
	i := 2
	enumName := consumeQualifiedName(tokens, &i)
	if enumName == "" {
		return
	}
	if i+1 >= len(tokens) || !strings.EqualFold(tokens[i].text, "AS") || !strings.EqualFold(tokens[i+1].text, "ENUM") {
		return
	}
	i += 2

	if i >= len(tokens) || tokens[i].typ != tokParenOpen {
		return
	}
	i++

	var vals []string
	for i < len(tokens) && tokens[i].typ != tokParenClose {
		if tokens[i].typ == tokString {
			vals = append(vals, tokens[i].raw)
		}
		i++
	}
	s.Enums = append(s.Enums, Enum{
		Name:   enumName,
		Values: vals,
	})
}

func parseCreateView(tokens []token, s *Schema) {
	i := 1
	if i < len(tokens) && strings.EqualFold(tokens[i].text, "OR") &&
		i+1 < len(tokens) && strings.EqualFold(tokens[i+1].text, "REPLACE") {
		i += 2
	}
	if i < len(tokens) && strings.EqualFold(tokens[i].text, "MATERIALIZED") {
		i++
	}
	if i >= len(tokens) || !strings.EqualFold(tokens[i].text, "VIEW") {
		return
	}
	i++

	viewName := consumeQualifiedName(tokens, &i)
	if viewName != "" {
		s.Views = append(s.Views, viewName)
	}
}

// ParseDDL parses SQL DDL statements and returns a Schema.
// ParseDDL parses SQL DDL statements and returns a Schema.
func ParseDDL(sql string) (*Schema, error) {
	var stmts [][]token
	for _, stmt := range splitStatements(sql) {
		if t := tokenize(stmt); len(t) > 1 {
			stmts = append(stmts, t)
		}
	}

	s := &Schema{}
	// Tables come first because a dump may alter or comment on them before they are defined.
	for _, t := range stmts {
		if !wordAt(t, 0, "CREATE") {
			continue
		}
		switch strings.ToUpper(t[1].text) {
		case "TABLE", "TEMP", "TEMPORARY", "UNLOGGED":
			parseCreateTable(t, s)
		case "TYPE":
			parseCreateType(t, s)
		case "VIEW", "OR", "MATERIALIZED":
			parseCreateView(t, s)
		}
	}
	for _, t := range stmts {
		switch {
		case wordAt(t, 0, "CREATE", "INDEX"), wordAt(t, 0, "CREATE", "UNIQUE"):
			parseCreateIndex(t, s)
		case wordAt(t, 0, "ALTER", "TABLE"):
			parseAlterTable(t, s)
		case wordAt(t, 0, "COMMENT"):
			parseComment(t, s)
		}
	}
	return s, nil
}
