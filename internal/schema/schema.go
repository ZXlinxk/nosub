package schema

import "strings"

// Schema represents a parsed database schema.
type Schema struct {
	Tables []Table
	Enums  []Enum
	Views  []string
}

// Table represents a database table.
type Table struct {
	Name        string
	Comment     string
	Columns     []Column
	PrimaryKey  []string
	ForeignKeys []ForeignKey
	Indexes     []Index
}

// Column represents a table column.
type Column struct {
	Name          string
	Type          string
	Default       string
	Comment       string
	Nullable      bool
	Unique        bool
	AutoIncrement bool
}

// ForeignKey represents a foreign key relationship.
type ForeignKey struct {
	Columns    []string
	RefTable   string
	RefColumns []string
	OnDelete   string
	OnUpdate   string
}

// Index represents a table index.
type Index struct {
	Name    string
	Columns []string
	Unique  bool
}

// Enum represents an enumerated type.
type Enum struct {
	Name   string
	Values []string
}

// MarkdownOptions controls schema markdown formatting.
type MarkdownOptions struct {
	Compact bool
	Only    []string
	Exclude []string
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') ||
			(s[0] == '`' && s[len(s)-1] == '`') {
			return s[1 : len(s)-1]
		}
		if s[0] == '[' && s[len(s)-1] == ']' {
			return s[1 : len(s)-1]
		}
	}
	return s
}

func cleanName(s string) string {
	s = strings.TrimSpace(s)
	parts := strings.Split(s, ".")
	for i := range parts {
		parts[i] = unquote(parts[i])
	}
	if len(parts) == 1 {
		return parts[0]
	}
	if len(parts) == 2 {
		schema := strings.ToLower(parts[0])
		if schema == "public" || schema == "dbo" || schema == "main" {
			return parts[1]
		}
		return parts[0] + "." + parts[1]
	}
	return strings.Join(parts, ".")
}
