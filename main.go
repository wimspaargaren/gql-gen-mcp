// Package main provides the main entry point for the application.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"gopkg.in/yaml.v3"

	"github.com/wimspaargaren/gql-gen-mcp/internal/gen"
)

// Root represents the root of the YAML configuration file.
type Root struct {
	Schemas []Schema `yaml:"schemas"`
}

// Schema represents a schema configuration in the YAML file.
type Schema struct {
	Name   string `yaml:"name"`
	Dir    string `yaml:"dir"`
	Output string `yaml:"output"`
}

func main() {
	schemaConfigurations, err := parseYamlFile()
	if err != nil {
		log.Fatalf("error parsing yaml file: %s", err)
	}
	for _, schemaConf := range schemaConfigurations {
		generator := gen.NewGenerator(schemaConf.Schema,
			gen.WithOutputDir(schemaConf.OutputDirectory),
		)
		err := generator.Generate()
		if err != nil {
			log.Fatalf("unable to generate mcp server code: %s", err)
		}
	}
}

// SchemaConfiguration represents the configuration for a schema.
type SchemaConfiguration struct {
	Schema          *ast.Schema
	OutputDirectory string
}

func parseYamlFile() ([]*SchemaConfiguration, error) {
	res := []*SchemaConfiguration{}
	data, err := os.ReadFile(".gql-gen-mcp.yaml")
	if err != nil {
		return nil, fmt.Errorf("no .gql-gen-mcp.yaml file found: %w", err)
	}
	t := Root{}
	err = yaml.Unmarshal(data, &t)
	if err != nil {
		return nil, fmt.Errorf("incorrect yaml file: %w", err)
	}

	for i := 0; i < len(t.Schemas); i++ {
		schema := t.Schemas[i]
		schemaConf, err := readSchema(schema)
		if err != nil {
			return nil, fmt.Errorf("error reading schema: %s %w", schema.Name, err)
		}

		res = append(res, schemaConf)
	}
	return res, nil
}

func readSchema(schema Schema) (*SchemaConfiguration, error) {
	schemaText, err := schemaText(schema)
	if err != nil {
		return nil, fmt.Errorf("error reading schema: %s %w", schema.Name, err)
	}
	// ignore federated annotations, as they cannot be parsed by gqlparser.
	schemaText = strings.ReplaceAll(schemaText, "@key", "")
	schemaText = strings.ReplaceAll(schemaText, "@external", "")
	gqlSchema, err := gqlparser.LoadSchema(&ast.Source{
		Input: schemaText,
	})
	if err != nil {
		return nil, fmt.Errorf("error parsing schema graphql: %w", err)
	}
	return &SchemaConfiguration{
		Schema:          gqlSchema,
		OutputDirectory: schema.Output,
	}, nil
}

func schemaText(schema Schema) (string, error) {
	dirEntry, err := os.ReadDir(schema.Dir)
	if err != nil {
		return "", fmt.Errorf("error while reading directory: %s %w", schema.Dir, err)
	}
	totalSchema := ""
	for _, entry := range dirEntry {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(entry.Name(), ".graphql") &&
			!strings.HasSuffix(entry.Name(), ".graphqls") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(schema.Dir, entry.Name()))
		if err != nil {
			return "", fmt.Errorf("error reading file: %s %w", entry.Name(), err)
		}
		totalSchema += string(b) + "\n"
	}
	return totalSchema, nil
}
