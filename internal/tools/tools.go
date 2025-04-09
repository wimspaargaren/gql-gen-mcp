// Package tools converts a GraphQL AST to a set of MCP tools that can be used to interact with a GraphQL API.
package tools

import (
	"fmt"
	"strings"

	"github.com/vektah/gqlparser/v2/ast"
)

// ResolverType represents the type of resolver.
type ResolverType string

// Resolver types.
const (
	// QueryResolver is a query resolver.
	QueryResolver ResolverType = "query"
	// MutationResolver is a mutation resolver.
	MutationResolver ResolverType = "mutation"
	// SubscriptionResolver is a subscription resolver.
	SubscriptionResolver ResolverType = "subscription"
)

// Tool represents an MCP tool that can be used to interact with a GraphQL API.
type Tool struct {
	Name        string
	Description string
	Args        []*ToolArg
	Query       string
}

// Type represents the type of tool.
type Type string

const (
	// TypeString is a string type.
	TypeString Type = "String"
	// TypeNumber is a number type.
	TypeNumber Type = "Number"
	// TypeBoolean is a boolean type.
	TypeBoolean Type = "Boolean"
	// TypeArray is an array type.
	TypeArray Type = "Array"
	// TypeObject is an object type.
	TypeObject Type = "Object"
	// TypeEnum is an enum type.
	TypeEnum Type = "Enum"
)

// String returns the string representation of the type.
func (t Type) String() string {
	switch t {
	case TypeString, TypeEnum:
		return "String"
	case TypeNumber:
		return "Number"
	case TypeBoolean:
		return "Boolean"
	case TypeArray:
		return "Array"
	case TypeObject:
		return "Object"
	default:
		return string(t)
	}
}

// PropertyDefinitionString returns the property definition string for the type.
func (t Type) PropertyDefinitionString() string {
	switch t {
	case TypeString, TypeEnum:
		return "string"
	case TypeNumber:
		return "number"
	case TypeBoolean:
		return "boolean"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	default:
		return string(t)
	}
}

// ToolArg represents an argument for a tool.
type ToolArg struct {
	Name        string
	Description string
	Type        Type
	Required    bool
	Enum        []string
	Properties  string
	Items       string
}

// Schema represents the schema to be used for generating tools.
type Schema struct {
	astSchema   *ast.Schema
	cyclicTypes map[string]bool
}

func GetToolsForSchema(astSchema *ast.Schema) []Tool { //nolint:revive
	res := []Tool{}
	cyclicTypes := map[string]bool{}
	for k, v := range astSchema.Types {
		if strings.HasPrefix(k, "__") || k == "Query" || k == "Mutation" {
			continue
		}
		for _, f := range v.Fields {
			hasCycle := hasCycle(f, map[string]struct{}{}, astSchema)
			if hasCycle {
				cyclicTypes[k] = true
				continue
			}
		}
	}
	schema := &Schema{
		astSchema:   astSchema,
		cyclicTypes: cyclicTypes,
	}
	if schema.astSchema.Query != nil {
		for _, v := range schema.astSchema.Query.Fields {
			if v.Name == "__schema" ||
				v.Name == "__type" {
				continue
			}
			res = append(res, toolFromFieldDefinition(v, schema, QueryResolver))
		}
	}
	if schema.astSchema.Mutation != nil {
		for _, v := range schema.astSchema.Mutation.Fields {
			if v.Name == "__schema" ||
				v.Name == "__type" {
				continue
			}
			res = append(res, toolFromFieldDefinition(v, schema, MutationResolver))
		}
	}
	return res
}

func hasCycle(v *ast.FieldDefinition, cycleTracker map[string]struct{}, astSchema *ast.Schema) bool {
	if _, ok := cycleTracker[v.Name]; ok {
		return true
	}
	cycleTracker[v.Name] = struct{}{}

	subType, ok := astSchema.Types[v.Type.Name()]
	if !ok {
		return false
	}
	for _, f := range subType.Fields {
		if hasCycle(f, cycleTracker, astSchema) {
			return true
		}
	}
	return false
}

func toolFromFieldDefinition(v *ast.FieldDefinition, schema *Schema, resolverType ResolverType) Tool {
	tool := Tool{
		Name:        v.Name,
		Description: strings.ReplaceAll(v.Description, "\n", " "),
	}
	for _, a := range v.Arguments {
		tool.Args = append(tool.Args, parseArgs(a, schema))
	}
	tool.Query = getQuery(v, schema, resolverType)
	return tool
}

func getQuery(v *ast.FieldDefinition, schema *Schema, resolverType ResolverType) string {
	args := []string{}
	queryInput := []string{}

	for _, f := range v.Arguments {
		args = append(args, fmt.Sprintf("$%s: %s", f.Name, f.Type.String()))
		queryInput = append(queryInput, fmt.Sprintf("%s: $%s", f.Name, f.Name))
	}
	indent := 4
	responseQuery := getResponseQuery(v, schema, indent, map[string]bool{})
	argumentsList := ""
	if len(v.Arguments) > 0 {
		argumentsList = fmt.Sprintf("(%s)", strings.Join(args, ", "))
	}
	queryInputList := ""
	if len(queryInput) > 0 {
		queryInputList = fmt.Sprintf("(%s)", strings.Join(queryInput, ", "))
	}

	query := fmt.Sprintf(`
		%s %s %s {
			%s%s %s
		}
	`, resolverType, v.Name, argumentsList, v.Name, queryInputList, responseQuery)

	return query
}

func getResponseQuery(v *ast.FieldDefinition, schema *Schema, indent int, visited map[string]bool) string { //nolint:revive
	if visited[v.Type.Name()] && schema.cyclicTypes[v.Type.Name()] {
		return ""
	}
	visited[v.Type.Name()] = true
	object, ok := schema.astSchema.Types[v.Type.Name()]
	if !ok {
		return ""
	}

	if object.Kind == ast.Union {
		resp := "{\n"
		resp += strings.Repeat("\t", indent) + "__typename\n"
		for _, f := range object.Types {
			object, ok := schema.astSchema.Types[f]
			if !ok {
				panic(fmt.Sprintf("Type %s not found in schema", f))
			}
			resp += strings.Repeat("\t", indent) + "... on " + f + " {\n"
			indent++
			resp += getResponseQueryForFields(object.Fields, schema, indent, visited)

			indent--
			resp += strings.Repeat("\t", indent) + "}\n"
		}

		resp += strings.Repeat("\t", indent-1) + "}\n"
		return resp
	}
	if len(object.Fields) == 0 {
		return ""
	}

	resp := "{\n"
	resp += getResponseQueryForFields(object.Fields, schema, indent, visited)
	resp += strings.Repeat("\t", indent-1) + "}\n"
	return resp
}

func getResponseQueryForFields(fields ast.FieldList, schema *Schema, indent int, visited map[string]bool) string {
	result := ""
	for _, f := range fields {
		if visited[f.Type.Name()] && schema.cyclicTypes[f.Type.Name()] {
			continue
		}
		result += strings.Repeat("\t", indent) + fmt.Sprintf("%s ", f.Name)
		sub := getResponseQuery(f, schema, indent+1, visited)
		if sub != "" {
			result += sub
		} else {
			result += "\n"
		}
	}
	return result
}

func parseArgs(a *ast.ArgumentDefinition, schema *Schema) *ToolArg {
	res := ToolArg{
		Name:        a.Name,
		Description: strings.ReplaceAll(a.Description, "\n", " "),
		Required:    a.Type.NonNull,
	}

	return resolveToolArgType(&res, a.Type, schema)
}

func resolveToolArgType(res *ToolArg, t *ast.Type, schema *Schema) *ToolArg { //nolint:revive
	toolType := graphQLTypeToToolType(t, schema)
	res.Type = toolType
	switch toolType {
	case TypeString, TypeNumber, TypeBoolean:
		return res
	case TypeObject:
		subType, ok := schema.astSchema.Types[t.Name()]
		if !ok {
			panic(fmt.Sprintf("Type %s not found in schema", t.Name()))
		}
		res.Properties = resolveObjectProperties(subType.Fields, schema)
		return res
	case TypeArray:
		res.Items = resolveObjectProperties(arrayTypeToFieldList(t), schema)
		return res
	case TypeEnum:
		subType, ok := schema.astSchema.Types[t.Name()]
		if !ok {
			panic(fmt.Sprintf("Type %s not found in schema", t.Name()))
		}
		enumValues := subType.EnumValues
		enums := []string{}
		for _, enum := range enumValues {
			enums = append(enums, enum.Name)
		}
		res.Enum = enums
		return res
	default:
		panic(fmt.Sprintf("Type %s not found in schema", t.Name()))
	}
}

func isArray(t *ast.Type) bool {
	return t.NamedType == ""
}

// https://spec.graphql.org/draft/#sec-Type-System
func graphQLTypeToToolType(t *ast.Type, schema *Schema) Type { //nolint:revive
	if isArray(t) {
		return TypeArray
	}
	switch t.Name() {
	case "String", "ID", "DateTime":
		return TypeString
	case "Int", "Float":
		return TypeNumber
	case "Boolean":
		return TypeBoolean
	default:
		subType, ok := schema.astSchema.Types[t.Name()]
		if !ok {
			panic(fmt.Sprintf("Type %s not found in schema", t.Name()))
		}
		switch subType.Kind {
		case ast.Enum:
			return TypeEnum
		case ast.InputObject, ast.Object:
			return TypeObject
		case ast.Scalar, ast.Interface, ast.Union:
			panic(fmt.Sprintf("Type %s not found in schema", t.Name()))
		default:
			panic(fmt.Sprintf("Type %s not found in schema", t.Name()))
		}
	}
}

func resolveObjectProperties(fields ast.FieldList, schema *Schema) string { //nolint:revive
	res := "map[string]any{"
	subFields := []string{}
	for _, f := range fields {
		toolType := graphQLTypeToToolType(f.Type, schema)
		subField := ""
		if f.Name != "" {
			subField += `"` + f.Name + `": map[string]any{`
		}
		keyVals := []string{}
		keyVals = append(keyVals, `"type": "`+toolType.PropertyDefinitionString()+`"`)
		if f.Description != "" {
			keyVals = append(keyVals, `"description": "`+strings.ReplaceAll(f.Description, "\n", " ")+`"`)
		}

		if toolType == TypeObject {
			subType, ok := schema.astSchema.Types[f.Type.Name()]
			if !ok {
				panic(fmt.Sprintf("Type %s not found in schema", f.Type.Name()))
			}
			keyVals = append(keyVals, `"properties": `+resolveObjectProperties(subType.Fields, schema))
		}
		if toolType == TypeEnum {
			subType, ok := schema.astSchema.Types[f.Type.Name()]
			if !ok {
				panic(fmt.Sprintf("Type %s not found in schema", f.Type.Name()))
			}
			enumValues := subType.EnumValues
			enums := []string{}
			for _, enum := range enumValues {
				enums = append(enums, fmt.Sprintf(`"%s"`, enum.Name))
			}
			keyVals = append(keyVals, `"enum": `+fmt.Sprintf("[]string{%s}", strings.Join(enums, ", ")))
		}
		if toolType == TypeArray {
			keyVals = append(keyVals, `"items": `+resolveObjectProperties(arrayTypeToFieldList(f.Type), schema))
		}
		subField += strings.Join(keyVals, ", ")
		if f.Name != "" {
			subField += `}`
		}
		subFields = append(subFields, subField)
	}
	res += strings.Join(subFields, ", ")
	res += "}"
	return res
}

func arrayTypeToFieldList(t *ast.Type) ast.FieldList {
	arrayType := &ast.Type{
		Elem:      t,
		NamedType: t.Elem.Name(),
		Position:  t.Position,
	}
	return ast.FieldList{
		&ast.FieldDefinition{
			Type: arrayType,
		},
	}
}
