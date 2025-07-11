package plate

import (
	"fmt"
	"reflect"
	"strings"
)

// ColumnInfo represents extracted column information
type ColumnInfo struct {
	Name        string // Field name (e.g., "ID", "UserID")
	GoType      string // Go type string (e.g., "string", "int64")
	SpannerType string // Spanner type from tag (e.g., "STRING", "INT64")
	ColumnName  string // Database column name from spanner tag
}

// extractColumns extracts column information from a model using reflection
func extractColumns(model interface{}) []ColumnInfo {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	var columns []ColumnInfo

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

		columns = append(columns, ColumnInfo{
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

// hasOrderableType checks if a Spanner type supports ordering operations (Lt, Gt, Le, Ge)
func hasOrderableType(spannerType string) bool {
	switch spannerType {
	case "INT64", "FLOAT64", "TIMESTAMP", "DATE":
		return true
	default:
		return false
	}
}

// hasStringOperations checks if a Spanner type supports string operations (Like)
func hasStringOperations(spannerType string) bool {
	return spannerType == "STRING"
}

// hasArrayOperations checks if a Spanner type supports array operations
func hasArrayOperations(spannerType string) bool {
	return strings.HasPrefix(spannerType, "ARRAY<")
}

// getTypeName extracts the type name from a model instance
func getTypeName(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// getPackagePath extracts the package path from a model instance
func getPackagePath(model interface{}) string {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath()
}
