package plate

import (
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/rail44/plate/query"
	"golang.org/x/tools/go/packages"
)

// Schema defines the database schema for generating query builders
type Schema struct {
	Tables    []TableConfig
	Junctions []JunctionConfig
}

// GenerateOptions holds options for code generation
type GenerateOptions struct {
	OutputDir string
	Clean     bool
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

// Generator is responsible for generating query builder code
type Generator struct {
	schema    Schema
	outputDir string
}

// NewGenerator creates a new Generator instance
func NewGenerator() *Generator {
	return &Generator{}
}

// GenerateFiles generates query builder code and returns the files without writing them
func (g *Generator) GenerateFiles(schema Schema, outputDir string) (GeneratedFiles, error) {
	g.schema = schema
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
	for _, tc := range g.schema.Tables {
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

// Generate generates and writes query builder code to the output directory
func (g *Generator) Generate(schema Schema, opts GenerateOptions) error {
	files, err := g.GenerateFiles(schema, opts.OutputDir)
	if err != nil {
		return err
	}

	if opts.Clean {
		return files.WriteToDirectoryClean(opts.OutputDir)
	}
	return files.WriteToDirectory(opts.OutputDir)
}

// validateConfig validates the generator configuration
func (g *Generator) validateConfig() error {
	// Check for duplicate table names
	seen := make(map[string]bool)

	for _, tc := range g.schema.Tables {
		name := g.getTypeName(tc.Schema)
		if seen[name] {
			return fmt.Errorf("duplicate table: %s", name)
		}
		seen[name] = true
	}

	// Validate junction tables have exactly 2 relations
	for _, jc := range g.schema.Junctions {
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

	for _, tc := range g.schema.Tables {
		name := g.getTypeName(tc.Schema)
		m[name] = tc.Schema
	}

	for _, jc := range g.schema.Junctions {
		name := g.getTypeName(jc.Schema)
		m[name] = jc.Schema
	}

	return m
}

// buildRelationMap builds a map of relations including derived ones
func (g *Generator) buildRelationMap() map[string][]generatedRelation {
	relations := make(map[string][]generatedRelation)

	// 1. Process BelongsTo relations from regular tables
	for _, tc := range g.schema.Tables {
		typeName := g.getTypeName(tc.Schema)

		for _, rel := range tc.Relations {
			// Add BelongsTo relation
			relations[typeName] = append(relations[typeName], generatedRelation{
				Name:   rel.Name,
				Type:   "belongs_to",
				Target: rel.Target,
				Keys:   query.KeyPair{From: rel.From, To: rel.To},
			})

			// Generate reverse HasMany relation if ReverseName is specified
			if rel.ReverseName != "" {
				relations[rel.Target] = append(relations[rel.Target], generatedRelation{
					Name:   rel.ReverseName,
					Type:   "has_many",
					Target: typeName,
					Keys:   query.KeyPair{From: rel.To, To: rel.From},
				})
			}
		}
	}

	// 2. Process junction tables to generate ManyToMany relations
	for _, jc := range g.schema.Junctions {
		if len(jc.Relations) != 2 {
			continue // Skip invalid junction tables
		}

		junctionName := g.getTypeName(jc.Schema)
		rel1 := jc.Relations[0]
		rel2 := jc.Relations[1]

		// Generate ManyToMany from first table to second
		if rel1.ReverseName != "" {
			relations[rel1.Target] = append(relations[rel1.Target], generatedRelation{
				Name:          rel1.ReverseName,
				Type:          "many_to_many",
				Target:        rel2.Target,
				Keys:          query.KeyPair{From: rel1.To, To: rel1.From},
				JunctionTable: junctionName,
				JunctionKeys:  query.KeyPair{From: rel2.From, To: rel2.To},
			})
		}

		// Generate ManyToMany from second table to first
		if rel2.ReverseName != "" {
			relations[rel2.Target] = append(relations[rel2.Target], generatedRelation{
				Name:          rel2.ReverseName,
				Type:          "many_to_many",
				Target:        rel1.Target,
				Keys:          query.KeyPair{From: rel2.To, To: rel2.From},
				JunctionTable: junctionName,
				JunctionKeys:  query.KeyPair{From: rel1.From, To: rel1.To},
			})
		}
	}

	return relations
}

// generatedRelation represents a relation that will be generated
type generatedRelation struct {
	Name          string
	Type          string // "belongs_to", "has_many", "many_to_many"
	Target        string
	Keys          query.KeyPair
	JunctionTable string        // For many_to_many
	JunctionKeys  query.KeyPair // For many_to_many
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
		for _, tc := range g.schema.Tables {
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
func (g *Generator) generateQueryBuilder(tc TableConfig, tableMap map[string]TableSchema, relationMap map[string][]generatedRelation) (string, error) {
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
	tablesImportPath, err := g.getTablesImportPath()
	if err != nil {
		return "", fmt.Errorf("failed to get tables import path: %v", err)
	}

	plateImportPath, err := g.getPlateImportPath()
	if err != nil {
		return "", fmt.Errorf("failed to get plate import path: %v", err)
	}

	imports := []string{
		"github.com/cloudspannerecosystem/memefish/ast",
		plateImportPath + "/query",
		tablesImportPath,
		plateImportPath + "/types",
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

// getTablesImportPath returns the import path for the generated tables package
func (g *Generator) getTablesImportPath() (string, error) {
	// Convert output directory to absolute path
	absOutputDir, err := filepath.Abs(g.outputDir)
	if err != nil {
		return "", fmt.Errorf("failed to get absolute path: %v", err)
	}

	// Load package information from the parent directory of output
	// This ensures we're in a valid package context
	parentDir := filepath.Dir(absOutputDir)
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedModule,
		Dir:  parentDir,
	}

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return "", fmt.Errorf("failed to load package information: %v", err)
	}

	if len(pkgs) == 0 {
		return "", fmt.Errorf("no package found at %s", parentDir)
	}

	pkg := pkgs[0]
	if pkg.Module == nil {
		return "", fmt.Errorf("no module information found (is there a go.mod in the project?)")
	}

	// Build import path from package path + relative path to output + /tables
	outputBase := filepath.Base(absOutputDir)
	return pkg.PkgPath + "/" + outputBase + "/tables", nil
}

// getPlateImportPath returns the import path for the plate module
func (g *Generator) getPlateImportPath() (string, error) {
	// Load the plate package to get its import path
	cfg := &packages.Config{
		Mode: packages.NeedName,
	}

	// Try to load github.com/rail44/plate first
	pkgs, err := packages.Load(cfg, "github.com/rail44/plate")
	if err == nil && len(pkgs) > 0 && pkgs[0].PkgPath != "" {
		return pkgs[0].PkgPath, nil
	}

	// If that fails, the plate module might be in development or vendored
	// In this case, we need to detect it from the runtime
	// Get the import path of the Generator type itself
	t := reflect.TypeOf(g)
	pkgPath := t.PkgPath()

	// pkgPath should be something like "github.com/rail44/plate" or a local path
	if pkgPath != "" {
		return pkgPath, nil
	}

	return "", fmt.Errorf("could not determine plate module import path")
}
