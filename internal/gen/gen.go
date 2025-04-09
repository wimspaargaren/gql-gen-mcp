// Package gen provides a code generator for generating MCP tools and server code based on a GraphQL schema.
package gen

import (
	"bytes"
	_ "embed"
	"fmt"
	"go/format"
	"log"
	"os"
	"strings"
	"text/template"

	"github.com/vektah/gqlparser/v2/ast"

	"github.com/wimspaargaren/gql-gen-mcp/internal/tools"
)

// Options contains options for the code generator.
type Options struct {
	// OutputDir is the directory where the generated files will be saved.
	OutputDir string
}

func defaultGenOpts() *Options {
	return &Options{
		OutputDir: "gql-gen-mcp",
	}
}

// Option is a function that modifies the GenOpts.
type Option func(*Options)

// WithOutputDir sets the output directory for the generated files.
func WithOutputDir(dir string) Option {
	return func(opts *Options) {
		opts.OutputDir = dir
	}
}

//go:embed templates/tool-template.tmpl
var toolTemplateContent string

//go:embed templates/server-template.tmpl
var serverTemplateContent string

// Generator is responsible for generating code based on the provided schema.
type Generator struct {
	tools   []tools.Tool
	options *Options
}

// NewGenerator creates a new Generator instance with the provided schema.
func NewGenerator(schema *ast.Schema, options ...Option) *Generator {
	genOpts := defaultGenOpts()
	for _, opt := range options {
		opt(genOpts)
	}

	schemaTools := tools.GetToolsForSchema(schema)
	return &Generator{
		tools:   schemaTools,
		options: genOpts,
	}
}

// Generate generates the code based on the provided schema and options.
func (g *Generator) Generate() error {
	data := TemplateData{
		Tools: g.tools,
	}

	err := g.generateTools(data)
	if err != nil {
		return fmt.Errorf("error generating tools: %w", err)
	}
	err = g.generateServer()
	if err != nil {
		return fmt.Errorf("error generating server: %w", err)
	}
	return nil
}

// TemplateData represents the data structure used in the template.
type TemplateData struct {
	Tools []tools.Tool
}

func (g *Generator) generateTools(data TemplateData) error {
	funcMap := template.FuncMap{
		"capitalise": func(s string) string {
			if len(s) == 0 {
				return s
			}
			return strings.ToUpper(string(s[0])) + s[1:]
		},
	}
	tpl, err := template.New("mcp-tool-gql").Funcs(funcMap).Parse(toolTemplateContent)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	var buf bytes.Buffer
	err = tpl.Execute(&buf, data)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return g.writeFile(buf, "tools")
}

func (g *Generator) generateServer() error {
	fileExists := fileExists(fmt.Sprintf("%s/main.go", g.options.OutputDir))
	if fileExists {
		return nil
	}
	tpl, err := template.New("mcp-server-gql").Parse(serverTemplateContent)
	if err != nil {
		return fmt.Errorf("error parsing template: %w", err)
	}

	// Create a buffer to hold the template output
	var buf bytes.Buffer
	err = tpl.Execute(&buf, g.tools)
	if err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	return g.writeFile(buf, "main")
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func (g *Generator) writeFile(buffer bytes.Buffer, fileName string) error {
	formattedCode, err := format.Source(buffer.Bytes())
	if err != nil {
		return fmt.Errorf("error formatting generated code: %w \n%s", err, buffer.String())
	}

	err = os.MkdirAll(g.options.OutputDir, 0o750)
	if err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	outputFile, err := os.Create(fmt.Sprintf("%s/%s.go", g.options.OutputDir, fileName))
	if err != nil {
		return fmt.Errorf("error creating output file: %w", err)
	}
	defer func() {
		err = outputFile.Close()
		if err != nil {
			log.Default().Println("close output file failed", err)
		}
	}()

	_, err = outputFile.Write(formattedCode)
	if err != nil {
		return fmt.Errorf("error writing to output file: %w", err)
	}

	return nil
}
