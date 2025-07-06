package plate

import (
	"fmt"
	"path/filepath"
	"reflect"
	
	"golang.org/x/tools/go/packages"
)

// GeneratorConfig holds the configuration for generating query builders
type GeneratorConfig struct {
	Tables    []TableConfig
	Junctions []JunctionConfig
}

// TableConfig represents a table that needs a query builder
type TableConfig struct {
	Schema    TableSchema
	Relations []Relation
}

// JunctionConfig represents a junction table for many-to-many relationships
type JunctionConfig struct {
	Schema    TableSchema
	Relations []Relation // Must have exactly 2 relations
}

// TableSchema contains the basic information about a table
type TableSchema struct {
	TableName string      // Database table name (e.g., "user")
	Model     interface{} // Model instance (e.g., models.User{})
}

// Relation represents a belongs_to relationship
type Relation struct {
	Name        string // Relationship name (e.g., "Author")
	Target      string // Target table name (e.g., "User")
	From        string // Source column (e.g., "UserID")
	To          string // Target column (e.g., "ID")
	ReverseName string // Optional: Name for the reverse HasMany relation (e.g., "Posts")
}

// KeyPair represents a pair of keys for relationships
type KeyPair struct {
	From string // Key from the source table
	To   string // Key in the target table
}

// Generator is responsible for generating query builder code
type Generator struct {
	config GeneratorConfig
	outputDir string
}

// NewGenerator creates a new Generator instance
func NewGenerator() *Generator {
	return &Generator{}
}

// Generate generates query builder code based on the provided configuration
func (g *Generator) Generate(config GeneratorConfig, outputDir string) (GeneratedFiles, error) {
	g.config = config
	g.outputDir = outputDir

	// Validate configuration
	if err := g.validateConfig(); err != nil {
		return GeneratedFiles{}, fmt.Errorf("invalid configuration: %w", err)
	}

	// Build internal data structures
	tableMap := g.buildTableMap()
	relationMap := g.buildRelationMap()

	// Generate files
	files := make(map[string]string)

	// Generate tables package
	tablesCode, err := g.generateTablesPackage(tableMap)
	if err != nil {
		return GeneratedFiles{}, fmt.Errorf("failed to generate tables package: %w", err)
	}
	files["tables/tables.go"] = tablesCode

	// Generate query builder packages
	for _, tc := range g.config.Tables {
		typeName := g.getTypeName(tc.Schema)
		packageName := g.toPackageName(typeName)
		
		code, err := g.generateQueryBuilder(tc, tableMap, relationMap)
		if err != nil {
			return GeneratedFiles{}, fmt.Errorf("failed to generate query builder for %s: %w", typeName, err)
		}
		files[fmt.Sprintf("%s/%s.go", packageName, packageName)] = code
	}

	return GeneratedFiles{Files: files}, nil
}

// validateConfig validates the generator configuration
func (g *Generator) validateConfig() error {
	// Check for duplicate table names
	seen := make(map[string]bool)
	
	for _, tc := range g.config.Tables {
		name := g.getTypeName(tc.Schema)
		if seen[name] {
			return fmt.Errorf("duplicate table: %s", name)
		}
		seen[name] = true
	}

	// Validate junction tables have exactly 2 relations
	for _, jc := range g.config.Junctions {
		if len(jc.Relations) != 2 {
			name := g.getTypeName(jc.Schema)
			return fmt.Errorf("junction table %s must have exactly 2 relations, got %d", name, len(jc.Relations))
		}
	}

	return nil
}

// getTypeName extracts the type name from a model
func (g *Generator) getTypeName(schema TableSchema) string {
	t := reflect.TypeOf(schema.Model)
	return t.Name()
}

// toPackageName converts a type name to a package name
func (g *Generator) toPackageName(typeName string) string {
	// Simple conversion: "UserProfile" -> "user_profile"
	// This is a placeholder - implement proper snake_case conversion
	return toSnakeCase(typeName)
}

// buildTableMap creates a map of type names to table schemas
func (g *Generator) buildTableMap() map[string]TableSchema {
	m := make(map[string]TableSchema)
	
	for _, tc := range g.config.Tables {
		name := g.getTypeName(tc.Schema)
		m[name] = tc.Schema
	}
	
	for _, jc := range g.config.Junctions {
		name := g.getTypeName(jc.Schema)
		m[name] = jc.Schema
	}
	
	return m
}

// buildRelationMap builds a map of relations including derived ones
func (g *Generator) buildRelationMap() map[string][]GeneratedRelation {
	relations := make(map[string][]GeneratedRelation)
	
	// 1. Process BelongsTo relations from regular tables
	for _, tc := range g.config.Tables {
		typeName := g.getTypeName(tc.Schema)
		
		for _, rel := range tc.Relations {
			// Add BelongsTo relation
			relations[typeName] = append(relations[typeName], GeneratedRelation{
				Name:   rel.Name,
				Type:   "belongs_to",
				Target: rel.Target,
				Keys:   KeyPair{From: rel.From, To: rel.To},
			})
			
			// Generate reverse HasMany relation if ReverseName is specified
			if rel.ReverseName != "" {
				relations[rel.Target] = append(relations[rel.Target], GeneratedRelation{
					Name:   rel.ReverseName,
					Type:   "has_many",
					Target: typeName,
					Keys:   KeyPair{From: rel.To, To: rel.From},
				})
			}
		}
	}
	
	// 2. Process junction tables to generate ManyToMany relations
	for _, jc := range g.config.Junctions {
		if len(jc.Relations) != 2 {
			continue // Skip invalid junction tables
		}
		
		junctionName := g.getTypeName(jc.Schema)
		rel1 := jc.Relations[0]
		rel2 := jc.Relations[1]
		
		// Generate ManyToMany from first table to second
		if rel1.ReverseName != "" {
			relations[rel1.Target] = append(relations[rel1.Target], GeneratedRelation{
				Name:          rel1.ReverseName,
				Type:          "many_to_many",
				Target:        rel2.Target,
				Keys:          KeyPair{From: rel1.To, To: rel1.From},
				JunctionTable: junctionName,
				JunctionKeys:  KeyPair{From: rel2.From, To: rel2.To},
			})
		}
		
		// Generate ManyToMany from second table to first
		if rel2.ReverseName != "" {
			relations[rel2.Target] = append(relations[rel2.Target], GeneratedRelation{
				Name:          rel2.ReverseName,
				Type:          "many_to_many",
				Target:        rel1.Target,
				Keys:          KeyPair{From: rel2.To, To: rel2.From},
				JunctionTable: junctionName,
				JunctionKeys:  KeyPair{From: rel1.From, To: rel1.To},
			})
		}
	}
	
	return relations
}


// GeneratedRelation represents a relation that will be generated
type GeneratedRelation struct {
	Name          string
	Type          string // "belongs_to", "has_many", "many_to_many"
	Target        string
	Keys          KeyPair
	JunctionTable string // For many_to_many
	JunctionKeys  KeyPair // For many_to_many
}

// generateTablesPackage generates the tables package containing all table definitions
func (g *Generator) generateTablesPackage(tableMap map[string]TableSchema) (string, error) {
	tmpl, err := getTemplates()
	if err != nil {
		return "", err
	}
	
	// Prepare table data
	var tables []tableTemplateData
	for typeName, schema := range tableMap {
		// Only include tables that have query builders
		for _, tc := range g.config.Tables {
			if g.getTypeName(tc.Schema) == typeName {
				tables = append(tables, tableTemplateData{
					TypeName:  typeName,
					TableName: schema.TableName,
				})
				break
			}
		}
	}
	
	data := templateData{
		Tables: tables,
	}
	
	return renderTemplate(tmpl, "tables", data)
}

// generateQueryBuilder generates a query builder package for a specific table
func (g *Generator) generateQueryBuilder(tc TableConfig, tableMap map[string]TableSchema, relationMap map[string][]GeneratedRelation) (string, error) {
	tmpl, err := getTemplates()
	if err != nil {
		return "", err
	}
	
	typeName := g.getTypeName(tc.Schema)
	packageName := g.toPackageName(typeName)
	
	// Extract columns
	columns := extractColumns(tc.Schema.Model)
	
	// Get relations for this table
	relations := relationMap[typeName]
	
	// Determine imports
	imports := []string{
		"github.com/cloudspannerecosystem/memefish/ast",
		g.getBaseImportPath() + "/query",
		g.getTablesImportPath(),
		g.getBaseImportPath() + "/types",
	}
	
	// Add time import if any column uses time.Time
	needsTime := false
	for _, col := range columns {
		if col.GoType == "time.Time" {
			needsTime = true
			break
		}
	}
	if needsTime {
		imports = append([]string{"time"}, imports...)
	}
	
	data := templateData{
		PackageName: packageName,
		TypeName:    typeName,
		TableName:   tc.Schema.TableName,
		Columns:     columns,
		Relations:   relations,
		Imports:     imports,
	}
	
	return renderTemplate(tmpl, "queryBuilder", data)
}

// getBaseImportPath returns the base import path for generated packages
func (g *Generator) getBaseImportPath() string {
	// Get the import path of the plate package itself
	t := reflect.TypeOf(Generator{})
	pkgPath := t.PkgPath()
	
	// pkgPath will be something like "github.com/rail44/plate"
	// or "github.com/user/project/vendor/github.com/rail44/plate"
	
	// Find the last occurrence of "plate" to handle vendored packages
	if idx := lastIndex(pkgPath, "/plate"); idx >= 0 {
		return pkgPath[:idx+6] // Include "/plate"
	}
	
	// Fallback to the package path as-is
	return pkgPath
}

// getTablesImportPath returns the import path for the generated tables package
func (g *Generator) getTablesImportPath() string {
	// Use packages.Load to get package information
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
		Dir:  g.outputDir,
	}
	
	pkgs, err := packages.Load(cfg, ".")
	if err != nil || len(pkgs) == 0 || pkgs[0].Module == nil {
		// Fallback to default if package information not found
		return g.getBaseImportPath() + "/tables"
	}
	
	pkg := pkgs[0]
	module := pkg.Module
	
	// Calculate import path based on module and directory
	absOutputDir, _ := filepath.Abs(g.outputDir)
	relPath, err := filepath.Rel(module.Dir, absOutputDir)
	if err != nil {
		return g.getBaseImportPath() + "/tables"
	}
	
	// Build import path
	if relPath == "." {
		return module.Path + "/tables"
	}
	return module.Path + "/" + filepath.ToSlash(relPath) + "/tables"
}

// lastIndex finds the last occurrence of substr in s
func lastIndex(s, substr string) int {
	last := -1
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			last = i
		}
	}
	return last
}

