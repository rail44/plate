package plate

import (
	"fmt"
	"reflect"
)

// columnInfo represents extracted column information
type columnInfo struct {
	Name        string // Field name (e.g., "ID", "UserID")
	GoType      string // Go type string (e.g., "string", "int64")
	SpannerType string // Spanner type from tag (e.g., "STRING", "INT64")
	ColumnName  string // Database column name from spanner tag
}

// extractColumns extracts column information from a model using reflection
func extractColumns(model interface{}) []columnInfo {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns []columnInfo

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Extract spanner tag
		spannerTag := field.Tag.Get("spanner")
		if spannerTag == "" {
			// Skip fields without spanner tag
			continue
		}

		// Extract spannerType tag
		spannerType := field.Tag.Get("spannerType")
		if spannerType == "" {
			// Try to infer from Go type if spannerType tag is missing
			spannerType = inferSpannerType(field.Type)
		}

		columns = append(columns, columnInfo{
			Name:        field.Name,
			GoType:      getGoTypeString(field.Type),
			SpannerType: spannerType,
			ColumnName:  spannerTag,
		})
	}

	return columns
}

// getGoTypeString converts a reflect.Type to its string representation
func getGoTypeString(t reflect.Type) string {
	// Handle common types
	switch t.Kind() {
	case reflect.Slice:
		return "[]" + getGoTypeString(t.Elem())
	case reflect.Array:
		return fmt.Sprintf("[%d]%s", t.Len(), getGoTypeString(t.Elem()))
	case reflect.Ptr:
		return "*" + getGoTypeString(t.Elem())
	}

	// For types with package path (e.g., time.Time)
	if t.PkgPath() != "" {
		return t.String()
	}

	// For built-in types
	return t.Name()
}

// inferSpannerType attempts to infer Spanner type from Go type
// This is a fallback when spannerType tag is not provided
func inferSpannerType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "STRING"
	case reflect.Int, reflect.Int64:
		return "INT64"
	case reflect.Float64:
		return "FLOAT64"
	case reflect.Bool:
		return "BOOL"
	case reflect.Slice:
		if t.Elem().Kind() == reflect.Uint8 {
			return "BYTES"
		}
		// Handle arrays
		elemType := inferSpannerType(t.Elem())
		if elemType != "" {
			return fmt.Sprintf("ARRAY<%s>", elemType)
		}
	}

	// Check for known types
	typeName := t.String()
	switch typeName {
	case "time.Time":
		return "TIMESTAMP"
	case "civil.Date":
		return "DATE"
	}

	// Default to empty string if we can't infer
	return ""
}
