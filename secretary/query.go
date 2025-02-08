package secretary

import (
	"strings"
)

type SQLStatementType int

const (
	Select SQLStatementType = iota
	Insert
	Update
	Delete
	Unknown
)

type SQLStatement struct {
	Type   SQLStatementType
	Table  string
	Fields []string
	Values []string
	Where  string
	Set    map[string]string
}

func ParseSQL(query string) SQLStatement {
	query = strings.TrimSpace(query)
	query = strings.ToUpper(query)

	stmt := SQLStatement{}

	switch {
	case strings.HasPrefix(query, "SELECT"):
		stmt.Type = Select
		parseSelect(&stmt, query)
	case strings.HasPrefix(query, "INSERT"):
		stmt.Type = Insert
		parseInsert(&stmt, query)
	case strings.HasPrefix(query, "UPDATE"):
		stmt.Type = Update
		parseUpdate(&stmt, query)
	case strings.HasPrefix(query, "DELETE"):
		stmt.Type = Delete
		parseDelete(&stmt, query)
	default:
		stmt.Type = Unknown
	}

	return stmt
}

func parseSelect(stmt *SQLStatement, query string) {
	// Extract fields
	fromIndex := strings.Index(query, "FROM")
	fields := query[len("SELECT"):fromIndex]
	stmt.Fields = parseFieldList(fields)

	// Extract table
	whereIndex := strings.Index(query, "WHERE")
	if whereIndex == -1 {
		stmt.Table = strings.TrimSpace(query[fromIndex+len("FROM"):])
	} else {
		stmt.Table = strings.TrimSpace(query[fromIndex+len("FROM") : whereIndex])
	}

	// Extract WHERE clause
	if whereIndex != -1 {
		stmt.Where = strings.TrimSpace(query[whereIndex+len("WHERE"):])
	}
}

func parseInsert(stmt *SQLStatement, query string) {
	// Extract table name
	intoIndex := strings.Index(query, "INTO")
	parenIndex := strings.Index(query, "(")
	stmt.Table = strings.TrimSpace(query[intoIndex+len("INTO") : parenIndex])

	// Extract fields
	fieldsEnd := strings.Index(query, ")")
	fields := query[parenIndex+1 : fieldsEnd]
	stmt.Fields = parseFieldList(fields)

	// Extract values
	valuesIndex := strings.Index(query, "VALUES")
	values := query[valuesIndex+len("VALUES"):]
	values = strings.Trim(values, "()")
	stmt.Values = parseValueList(values)
}

func parseUpdate(stmt *SQLStatement, query string) {
	// Extract table name
	setIndex := strings.Index(query, "SET")
	stmt.Table = strings.TrimSpace(query[len("UPDATE"):setIndex])

	// Extract SET clause
	whereIndex := strings.Index(query, "WHERE")
	setClause := query[setIndex+len("SET"):]
	if whereIndex != -1 {
		setClause = query[setIndex+len("SET") : whereIndex]
	}

	stmt.Set = make(map[string]string)
	pairs := strings.Split(setClause, ",")
	for _, pair := range pairs {
		parts := strings.Split(pair, "=")
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			stmt.Set[key] = value
		}
	}

	// Extract WHERE clause
	if whereIndex != -1 {
		stmt.Where = strings.TrimSpace(query[whereIndex+len("WHERE"):])
	}
}

func parseDelete(stmt *SQLStatement, query string) {
	// Extract table name
	fromIndex := strings.Index(query, "FROM")
	whereIndex := strings.Index(query, "WHERE")
	if whereIndex == -1 {
		stmt.Table = strings.TrimSpace(query[fromIndex+len("FROM"):])
	} else {
		stmt.Table = strings.TrimSpace(query[fromIndex+len("FROM") : whereIndex])
	}

	// Extract WHERE clause
	if whereIndex != -1 {
		stmt.Where = strings.TrimSpace(query[whereIndex+len("WHERE"):])
	}
}

func parseFieldList(fields string) []string {
	fields = strings.TrimSpace(fields)
	if fields == "*" {
		return []string{"*"}
	}
	return splitByComma(fields)
}

func parseValueList(values string) []string {
	return splitByComma(values)
}

func splitByComma(s string) []string {
	var result []string
	var buffer strings.Builder
	inQuotes := false

	for _, r := range s {
		switch {
		case r == '\'':
			inQuotes = !inQuotes
			buffer.WriteRune(r)
		case r == ',' && !inQuotes:
			result = append(result, strings.TrimSpace(buffer.String()))
			buffer.Reset()
		default:
			buffer.WriteRune(r)
		}
	}

	if buffer.Len() > 0 {
		result = append(result, strings.TrimSpace(buffer.String()))
	}

	return result
}
