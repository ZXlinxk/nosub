package schema

import (
	"fmt"
	"path"
	"strings"
)

// Markdown formats a Schema into a Markdown representation.
func Markdown(s *Schema, o MarkdownOptions) string {
	var tables []Table
	for _, tb := range s.Tables {
		if len(o.Only) > 0 {
			matched := false
			for _, pat := range o.Only {
				if m, _ := path.Match(pat, tb.Name); m {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}
		if len(o.Exclude) > 0 {
			excluded := false
			for _, pat := range o.Exclude {
				if m, _ := path.Match(pat, tb.Name); m {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		tables = append(tables, tb)
	}

	var b strings.Builder
	b.WriteString("# Database schema\n\n")

	if o.Compact {
		for _, tb := range tables {
			singleFK := make(map[string]ForeignKey)
			for _, fk := range tb.ForeignKeys {
				if len(fk.Columns) == 1 && len(fk.RefColumns) == 1 {
					singleFK[fk.Columns[0]] = fk
				}
			}

			var colStrs []string
			for _, col := range tb.Columns {
				var parts []string
				parts = append(parts, col.Name)
				if len(tb.PrimaryKey) == 1 && tb.PrimaryKey[0] == col.Name {
					parts = append(parts, "PK")
				}
				if fk, ok := singleFK[col.Name]; ok {
					parts = append(parts, "-> "+fk.RefTable+"."+fk.RefColumns[0])
				}
				colStrs = append(colStrs, strings.Join(parts, " "))
			}
			b.WriteString(tb.Name + "(" + strings.Join(colStrs, ", ") + ")\n")
		}
		return b.String()
	}

	n := len(tables)
	word := "tables"
	if n == 1 {
		word = "table"
	}
	b.WriteString(fmt.Sprintf("%d %s\n\n", n, word))

	for i, tb := range tables {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("## " + tb.Name + "\n")
		if tb.Comment != "" {
			b.WriteString(tb.Comment + "\n")
		}

		singleFK := make(map[string]ForeignKey)
		for _, fk := range tb.ForeignKeys {
			if len(fk.Columns) == 1 && len(fk.RefColumns) == 1 {
				singleFK[fk.Columns[0]] = fk
			}
		}

		for _, col := range tb.Columns {
			isSimplePK := len(tb.PrimaryKey) == 1 && tb.PrimaryKey[0] == col.Name

			line := "- " + col.Name + " " + col.Type
			if isSimplePK {
				line += " PK"
			}
			if fk, ok := singleFK[col.Name]; ok {
				line += " -> " + fk.RefTable + "." + fk.RefColumns[0]
				if fk.OnDelete != "" {
					line += " ON DELETE " + fk.OnDelete
				}
				if fk.OnUpdate != "" {
					line += " ON UPDATE " + fk.OnUpdate
				}
			}
			if !col.Nullable && !isSimplePK {
				line += " NOT NULL"
			}
			if col.Unique {
				line += " UNIQUE"
			}
			if col.Default != "" {
				line += " DEFAULT " + col.Default
			}
			if col.Comment != "" {
				line += " -- " + col.Comment
			}
			b.WriteString(line + "\n")
		}

		if len(tb.PrimaryKey) > 1 {
			b.WriteString("Primary key: " + strings.Join(tb.PrimaryKey, ", ") + "\n")
		}

		for _, fk := range tb.ForeignKeys {
			if len(fk.Columns) > 1 {
				fkLine := fmt.Sprintf("Foreign key: (%s) -> %s(%s)",
					strings.Join(fk.Columns, ", "),
					fk.RefTable,
					strings.Join(fk.RefColumns, ", "))
				if fk.OnDelete != "" {
					fkLine += " ON DELETE " + fk.OnDelete
				}
				if fk.OnUpdate != "" {
					fkLine += " ON UPDATE " + fk.OnUpdate
				}
				b.WriteString(fkLine + "\n")
			}
		}

		if len(tb.Indexes) > 0 {
			var idxStrs []string
			for _, idx := range tb.Indexes {
				var is string
				if idx.Unique {
					is = "UNIQUE "
				}
				is += idx.Name + " (" + strings.Join(idx.Columns, ", ") + ")"
				idxStrs = append(idxStrs, is)
			}
			b.WriteString("Indexes: " + strings.Join(idxStrs, "; ") + "\n")
		}
	}

	if len(s.Enums) > 0 {
		b.WriteString("\nEnums:\n")
		for _, e := range s.Enums {
			b.WriteString(fmt.Sprintf("- %s: %s\n", e.Name, strings.Join(e.Values, ", ")))
		}
	}
	if len(s.Views) > 0 {
		b.WriteString("\nViews: " + strings.Join(s.Views, ", ") + "\n")
	}

	return b.String()
}
