package gen

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/go-openapi/spec"
	"github.com/venosm/swag"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"sigs.k8s.io/yaml"
)

var open = os.Open

// DefaultOverridesFile is the location swagger will look for type overrides.
const DefaultOverridesFile = ".swag"

type genTypeWriter func(*Config, *spec.Swagger) error

// Gen presents a generate tool for swag.
type Gen struct {
	json          func(data interface{}) ([]byte, error)
	jsonIndent    func(data interface{}) ([]byte, error)
	jsonToYAML    func(data []byte) ([]byte, error)
	outputTypeMap map[string]genTypeWriter
	debug         Debugger
}

// Debugger is the interface that wraps the basic Printf method.
type Debugger interface {
	Printf(format string, v ...interface{})
}

// New creates a new Gen.
func New() *Gen {
	gen := Gen{
		json: json.Marshal,
		jsonIndent: func(data interface{}) ([]byte, error) {
			return json.MarshalIndent(data, "", "    ")
		},
		jsonToYAML: yaml.JSONToYAML,
		debug:      log.New(os.Stdout, "", log.LstdFlags),
	}

	gen.outputTypeMap = map[string]genTypeWriter{
		"go":   gen.writeDocSwagger,
		"json": gen.writeJSONSwagger,
		"yaml": gen.writeYAMLSwagger,
		"yml":  gen.writeYAMLSwagger,
	}

	return &gen
}

// Config presents Gen configurations.
type Config struct {
	Debugger swag.Debugger

	// SearchDir the swag would parse,comma separated if multiple
	SearchDir string

	// excludes dirs and files in SearchDir,comma separated
	Excludes string

	// outputs only specific extension
	ParseExtension string

	// OutputDir represents the output directory for all the generated files
	OutputDir string

	// OutputTypes define types of files which should be generated
	OutputTypes []string

	// MainAPIFile the Go file path in which 'swagger general API Info' is written
	MainAPIFile string

	// PropNamingStrategy represents property naming strategy like snake case,camel case,pascal case
	PropNamingStrategy string

	// MarkdownFilesDir used to find markdown files, which can be used for tag descriptions
	MarkdownFilesDir string

	// CodeExampleFilesDir used to find code example files, which can be used for x-codeSamples
	CodeExampleFilesDir string

	// InstanceName is used to get distinct names for different swagger documents in the
	// same project. The default value is "swagger".
	InstanceName string

	// ParseDepth dependency parse depth
	ParseDepth int

	// ParseVendor whether swag should be parse vendor folder
	ParseVendor bool

	// ParseDependencies whether swag should be parse outside dependency folder: 0 none, 1 models, 2 operations, 3 all
	ParseDependency int

	// UseStructNames stick to the struct name instead of those ugly full-path names
	UseStructNames bool

	// ParseInternal whether swag should parse internal packages
	ParseInternal bool

	// Strict whether swag should error or warn when it detects cases which are most likely user errors
	Strict bool

	// GeneratedTime whether swag should generate the timestamp at the top of docs.go
	GeneratedTime bool

	// RequiredByDefault set validation required for all fields by default
	RequiredByDefault bool

	// OverridesFile defines global type overrides.
	OverridesFile string

	// ParseGoList whether swag use go list to parse dependency
	ParseGoList bool

	// include only tags mentioned when searching, comma separated
	Tags string

	// LeftTemplateDelim defines the left delimiter for the template generation
	LeftTemplateDelim string

	// RightTemplateDelim defines the right delimiter for the template generation
	RightTemplateDelim string

	// PackageName defines package name of generated `docs.go`
	PackageName string

	// CollectionFormat set default collection format
	CollectionFormat string

	// Parse only packages whose import path match the given prefix, comma separated
	PackagePrefix string

	// State set host state
	State string

	// ParseFuncBody whether swag should parse api info inside of funcs
	ParseFuncBody bool
}

// Build builds swagger json file  for given searchDir and mainAPIFile. Returns json.
func (g *Gen) Build(config *Config) error {
	if config.Debugger != nil {
		g.debug = config.Debugger
	}
	if config.InstanceName == "" {
		config.InstanceName = swag.Name
	}

	searchDirs := strings.Split(config.SearchDir, ",")
	for _, searchDir := range searchDirs {
		if _, err := os.Stat(searchDir); os.IsNotExist(err) {
			return fmt.Errorf("dir: %s does not exist", searchDir)
		}
	}

	if config.LeftTemplateDelim == "" {
		config.LeftTemplateDelim = "{{"
	}

	if config.RightTemplateDelim == "" {
		config.RightTemplateDelim = "}}"
	}

	var overrides map[string]string

	if config.OverridesFile != "" {
		overridesFile, err := open(config.OverridesFile)
		if err != nil {
			// Don't bother reporting if the default file is missing; assume there are no overrides
			if !(config.OverridesFile == DefaultOverridesFile && os.IsNotExist(err)) {
				return fmt.Errorf("could not open overrides file: %w", err)
			}
		} else {
			g.debug.Printf("Using overrides from %s", config.OverridesFile)

			overrides, err = parseOverrides(overridesFile)
			if err != nil {
				return err
			}
		}
	}

	g.debug.Printf("Generate swagger docs....")

	p := swag.New(
		swag.SetParseDependency(config.ParseDependency),
		swag.SetUseStructName(config.UseStructNames),
		swag.SetMarkdownFileDirectory(config.MarkdownFilesDir),
		swag.SetDebugger(config.Debugger),
		swag.SetExcludedDirsAndFiles(config.Excludes),
		swag.SetParseExtension(config.ParseExtension),
		swag.SetCodeExamplesDirectory(config.CodeExampleFilesDir),
		swag.SetStrict(config.Strict),
		swag.SetOverrides(overrides),
		swag.ParseUsingGoList(config.ParseGoList),
		swag.SetTags(config.Tags),
		swag.SetCollectionFormat(config.CollectionFormat),
		swag.SetPackagePrefix(config.PackagePrefix),
	)

	p.PropNamingStrategy = config.PropNamingStrategy
	p.ParseVendor = config.ParseVendor
	p.ParseInternal = config.ParseInternal
	p.RequiredByDefault = config.RequiredByDefault
	p.HostState = config.State
	p.ParseFuncBody = config.ParseFuncBody

	if err := p.ParseAPIMultiSearchDir(searchDirs, config.MainAPIFile, config.ParseDepth); err != nil {
		return err
	}

	swagger := p.GetSwagger()

	if err := os.MkdirAll(config.OutputDir, os.ModePerm); err != nil {
		return err
	}

	for _, outputType := range config.OutputTypes {
		outputType = strings.ToLower(strings.TrimSpace(outputType))
		if typeWriter, ok := g.outputTypeMap[outputType]; ok {
			if err := typeWriter(config, swagger); err != nil {
				return err
			}
		} else {
			log.Printf("output type '%s' not supported", outputType)
		}
	}

	return nil
}

func (g *Gen) writeDocSwagger(config *Config, swagger *spec.Swagger) error {
	var filename = "docs.go"

	if config.State != "" {
		filename = config.State + "_" + filename
	}

	if config.InstanceName != swag.Name {
		filename = config.InstanceName + "_" + filename
	}

	docFileName := path.Join(config.OutputDir, filename)

	absOutputDir, err := filepath.Abs(config.OutputDir)
	if err != nil {
		return err
	}

	var packageName string
	if len(config.PackageName) > 0 {
		packageName = config.PackageName
	} else {
		packageName = filepath.Base(absOutputDir)
		packageName = strings.ReplaceAll(packageName, "-", "_")
	}

	docs, err := os.Create(docFileName)
	if err != nil {
		return err
	}
	defer docs.Close()

	// Write doc
	err = g.writeGoDoc(packageName, docs, swagger, config)
	if err != nil {
		return err
	}

	g.debug.Printf("create docs.go at %+v", docFileName)

	return nil
}

func (g *Gen) writeJSONSwagger(config *Config, swagger *spec.Swagger) error {
	var filename = "swagger.json"

	if config.State != "" {
		filename = config.State + "_" + filename
	}

	if config.InstanceName != swag.Name {
		filename = config.InstanceName + "_" + filename
	}

	jsonFileName := path.Join(config.OutputDir, filename)

	b, err := g.jsonIndent(swagger)
	if err != nil {
		return err
	}

	// Convert to OpenAPI 3.0 format if needed
	if swagger.Swagger == "3.0.0" {
		b, err = g.convertToOpenAPI3(b)
		if err != nil {
			return err
		}
	}

	err = g.writeFile(b, jsonFileName)
	if err != nil {
		return err
	}

	g.debug.Printf("create swagger.json at %+v", jsonFileName)

	return nil
}

func (g *Gen) writeYAMLSwagger(config *Config, swagger *spec.Swagger) error {
	var filename = "swagger.yaml"

	if config.State != "" {
		filename = config.State + "_" + filename
	}

	if config.InstanceName != swag.Name {
		filename = config.InstanceName + "_" + filename
	}

	yamlFileName := path.Join(config.OutputDir, filename)

	b, err := g.json(swagger)
	if err != nil {
		return err
	}

	// Convert to OpenAPI 3.0 format if needed
	if swagger.Swagger == "3.0.0" {
		b, err = g.convertToOpenAPI3(b)
		if err != nil {
			return err
		}
	}

	y, err := g.jsonToYAML(b)
	if err != nil {
		return fmt.Errorf("cannot covert json to yaml error: %s", err)
	}

	// Post-process YAML to use literal block scalars for multiline descriptions
	y = g.formatMultilineDescriptions(y)

	err = g.writeFile(y, yamlFileName)
	if err != nil {
		return err
	}

	g.debug.Printf("create swagger.yaml at %+v", yamlFileName)

	return nil
}

func (g *Gen) writeFile(b []byte, file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write(b)

	return err
}

func (g *Gen) formatMultilineDescriptions(yamlContent []byte) []byte {
	content := string(yamlContent)
	lines := strings.Split(content, "\n")
	var result []string

	for _, line := range lines {
		// Look for description lines with quoted multiline content
		if strings.Contains(line, "description:") && strings.Contains(line, "\"") {
			// Extract the indentation
			indent := ""
			for _, char := range line {
				if char == ' ' {
					indent += " "
				} else {
					break
				}
			}

			// Check if the quoted string contains HTML or multiple lines
			quotedStart := strings.Index(line, "\"")
			if quotedStart == -1 {
				result = append(result, line)
				continue
			}

			quotedContent := line[quotedStart:]

			// If it's a simple quoted string, unquote it and check for HTML or newlines
			if strings.HasPrefix(quotedContent, "\"") && strings.HasSuffix(strings.TrimSpace(quotedContent), "\"") {
				unquoted := quotedContent[1 : len(quotedContent)-1]
				unquoted = strings.ReplaceAll(unquoted, "\\n", "\n")
				unquoted = strings.ReplaceAll(unquoted, "\\\"", "\"")

				// If contains HTML tags or multiple lines, use literal block scalar
				if strings.Contains(unquoted, "<") && strings.Contains(unquoted, ">") || strings.Contains(unquoted, "\n") {
					result = append(result, indent+"description: |")

					// Split content into lines and add proper indentation
					contentLines := strings.Split(unquoted, "\n")
					for _, contentLine := range contentLines {
						result = append(result, indent+"  "+contentLine)
					}
					continue
				}
			}
		}

		result = append(result, line)
	}

	return []byte(strings.Join(result, "\n"))
}

func (g *Gen) formatSource(src []byte) []byte {
	code, err := format.Source(src)
	if err != nil {
		code = src // Formatter failed, return original code.
	}

	return code
}

// Read and parse the overrides file.
func parseOverrides(r io.Reader) (map[string]string, error) {
	overrides := make(map[string]string)
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := scanner.Text()

		// Skip comments
		if len(line) > 1 && line[0:2] == "//" {
			continue
		}

		parts := strings.Fields(line)

		switch len(parts) {
		case 0:
			// only whitespace
			continue
		case 2:
			// either a skip or malformed
			if parts[0] != "skip" {
				return nil, fmt.Errorf("could not parse override: '%s'", line)
			}

			overrides[parts[1]] = ""
		case 3:
			// either a replace or malformed
			if parts[0] != "replace" {
				return nil, fmt.Errorf("could not parse override: '%s'", line)
			}

			overrides[parts[1]] = parts[2]
		default:
			return nil, fmt.Errorf("could not parse override: '%s'", line)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading overrides file: %w", err)
	}

	return overrides, nil
}

func (g *Gen) writeGoDoc(packageName string, output io.Writer, swagger *spec.Swagger, config *Config) error {
	generator, err := template.New("swagger_info").Funcs(template.FuncMap{
		"printDoc": func(v string) string {
			// For OpenAPI 3.0, schemes are part of servers array, so don't add them separately
			if swagger.Swagger == "3.0.0" {
				// Just sanitize backticks for OpenAPI 3.0 - schemes are already handled in servers
				return strings.Replace(v, "`", "`+\"`\"+`", -1)
			} else {
				// Add schemes for Swagger 2.0
				v = "{\n    \"schemes\": " + config.LeftTemplateDelim + " marshal .Schemes " + config.RightTemplateDelim + "," + v[1:]
				// Sanitize backticks
				return strings.Replace(v, "`", "`+\"`\"+`", -1)
			}
		},
	}).Parse(packageTemplate)
	if err != nil {
		return err
	}

	swaggerSpec := &spec.Swagger{
		VendorExtensible: swagger.VendorExtensible,
		SwaggerProps: spec.SwaggerProps{
			ID:       swagger.ID,
			Consumes: swagger.Consumes,
			Produces: swagger.Produces,
			Swagger:  swagger.Swagger,
			Info: &spec.Info{
				VendorExtensible: swagger.Info.VendorExtensible,
				InfoProps: spec.InfoProps{
					Description:    config.LeftTemplateDelim + "escape .Description" + config.RightTemplateDelim,
					Title:          config.LeftTemplateDelim + ".Title" + config.RightTemplateDelim,
					TermsOfService: swagger.Info.TermsOfService,
					Contact:        swagger.Info.Contact,
					License:        swagger.Info.License,
					Version:        config.LeftTemplateDelim + ".Version" + config.RightTemplateDelim,
				},
			},
			Host:                config.LeftTemplateDelim + ".Host" + config.RightTemplateDelim,
			BasePath:            config.LeftTemplateDelim + ".BasePath" + config.RightTemplateDelim,
			Paths:               swagger.Paths,
			Definitions:         swagger.Definitions,
			Parameters:          swagger.Parameters,
			Responses:           swagger.Responses,
			SecurityDefinitions: swagger.SecurityDefinitions,
			Security:            swagger.Security,
			Tags:                swagger.Tags,
			ExternalDocs:        swagger.ExternalDocs,
		},
	}

	// crafted docs.json
	buf, err := g.jsonIndent(swaggerSpec)
	if err != nil {
		return err
	}

	// Convert to OpenAPI 3.0 format if needed (for template)
	if swagger.Swagger == "3.0.0" {
		buf, err = g.convertToOpenAPI3(buf)
		if err != nil {
			return err
		}
	}

	state := ""
	if len(config.State) > 0 {
		state = cases.Title(language.English).String(strings.ToLower(config.State))
	}

	buffer := &bytes.Buffer{}

	err = generator.Execute(buffer, struct {
		Timestamp          time.Time
		Doc                string
		Host               string
		PackageName        string
		BasePath           string
		Title              string
		Description        string
		Version            string
		State              string
		InstanceName       string
		Schemes            []string
		GeneratedTime      bool
		LeftTemplateDelim  string
		RightTemplateDelim string
	}{
		Timestamp:          time.Now(),
		GeneratedTime:      config.GeneratedTime,
		Doc:                string(buf),
		Host:               swagger.Host,
		PackageName:        packageName,
		BasePath:           swagger.BasePath,
		Schemes:            swagger.Schemes,
		Title:              swagger.Info.Title,
		Description:        swagger.Info.Description,
		Version:            swagger.Info.Version,
		State:              state,
		InstanceName:       config.InstanceName,
		LeftTemplateDelim:  config.LeftTemplateDelim,
		RightTemplateDelim: config.RightTemplateDelim,
	})
	if err != nil {
		return err
	}

	code := g.formatSource(buffer.Bytes())

	// write
	_, err = output.Write(code)

	return err
}

// convertToOpenAPI3 converts Swagger 2.0 JSON to OpenAPI 3.0 format
func (g *Gen) convertToOpenAPI3(input []byte) ([]byte, error) {
	var doc map[string]interface{}
	if err := json.Unmarshal(input, &doc); err != nil {
		return nil, err
	}

	// Convert swagger field to openapi
	if _, ok := doc["swagger"]; ok {
		doc["openapi"] = doc["swagger"]
		delete(doc, "swagger")
	}

	// Convert host + basePath to servers
	if host, hasHost := doc["host"]; hasHost {
		if basePath, hasBasePath := doc["basePath"]; hasBasePath {
			servers := []map[string]interface{}{
				{
					"url": fmt.Sprintf("http://%s%s", host, basePath),
				},
			}
			doc["servers"] = servers
		} else {
			servers := []map[string]interface{}{
				{
					"url": fmt.Sprintf("http://%s", host),
				},
			}
			doc["servers"] = servers
		}
		delete(doc, "host")
		delete(doc, "basePath")
	}

	// Convert definitions to components.schemas
	if definitions, ok := doc["definitions"]; ok {
		components := map[string]interface{}{
			"schemas": definitions,
		}
		doc["components"] = components
		delete(doc, "definitions")
	}

	// Convert paths structure
	if paths, ok := doc["paths"].(map[string]interface{}); ok {
		g.convertPathsToOpenAPI3(paths)
	}

	// Update all $ref references throughout the document
	g.updateReferences(doc)

	// Remove global consumes/produces as they're handled per operation in OpenAPI 3.0
	delete(doc, "consumes")
	delete(doc, "produces")

	return json.MarshalIndent(doc, "", "    ")
}

func (g *Gen) convertPathsToOpenAPI3(paths map[string]interface{}) {
	for _, pathValue := range paths {
		if pathObj, ok := pathValue.(map[string]interface{}); ok {
			for _, methodValue := range pathObj {
				if methodObj, ok := methodValue.(map[string]interface{}); ok {
					g.convertOperationToOpenAPI3(methodObj)
				}
			}
		}
	}
}

func (g *Gen) convertOperationToOpenAPI3(operation map[string]interface{}) {
	// Convert consumes to requestBody
	if consumes, ok := operation["consumes"].([]interface{}); ok {
		if parameters, hasParams := operation["parameters"].([]interface{}); hasParams {
			for _, param := range parameters {
				if paramObj, ok := param.(map[string]interface{}); ok {
					if paramObj["in"] == "body" {
						requestBody := map[string]interface{}{
							"required": paramObj["required"],
							"content":  map[string]interface{}{},
						}

						content := requestBody["content"].(map[string]interface{})
						for _, consume := range consumes {
							if consumeStr, ok := consume.(string); ok {
								content[consumeStr] = map[string]interface{}{
									"schema": paramObj["schema"],
								}
							}
						}

						operation["requestBody"] = requestBody

						// Remove body parameter from parameters array
						if params, ok := operation["parameters"].([]interface{}); ok {
							newParams := []interface{}{}
							for _, p := range params {
								if pObj, ok := p.(map[string]interface{}); ok {
									if pObj["in"] != "body" {
										// Convert parameter structure for OpenAPI 3.0
										g.convertParameterToOpenAPI3(pObj)
										newParams = append(newParams, pObj)
									}
								}
							}
							if len(newParams) > 0 {
								operation["parameters"] = newParams
							} else {
								delete(operation, "parameters")
							}
						}
						break
					}
				}
			}
		}
		delete(operation, "consumes")
	}

	// Convert produces and responses
	if produces, ok := operation["produces"].([]interface{}); ok {
		if responses, ok := operation["responses"].(map[string]interface{}); ok {
			for _, response := range responses {
				if responseObj, ok := response.(map[string]interface{}); ok {
					if schema, hasSchema := responseObj["schema"]; hasSchema {
						content := map[string]interface{}{}
						for _, produce := range produces {
							if produceStr, ok := produce.(string); ok {
								content[produceStr] = map[string]interface{}{
									"schema": schema,
								}
							}
						}
						responseObj["content"] = content
						delete(responseObj, "schema")
					}
				}
			}
		}
		delete(operation, "produces")
	}

	// Convert parameters to OpenAPI 3.0 format
	if parameters, ok := operation["parameters"].([]interface{}); ok {
		for _, param := range parameters {
			if paramObj, ok := param.(map[string]interface{}); ok {
				g.convertParameterToOpenAPI3(paramObj)
			}
		}
	}
}

func (g *Gen) convertParameterToOpenAPI3(param map[string]interface{}) {
	// For non-body parameters, wrap type info in schema
	if param["in"] != "body" && param["in"] != "formData" {
		schema := map[string]interface{}{}

		// Move type, format, enum, etc. to schema
		typeFields := []string{"type", "format", "enum", "minimum", "maximum", "minLength", "maxLength", "pattern", "items", "default", "example"}
		for _, field := range typeFields {
			if value, exists := param[field]; exists {
				schema[field] = value
				delete(param, field)
			}
		}

		if len(schema) > 0 {
			param["schema"] = schema
		}

		// Ensure required is boolean for path parameters
		if param["in"] == "path" {
			param["required"] = true
		}
	}

	// Update $ref references
	if schema, ok := param["schema"].(map[string]interface{}); ok {
		g.updateReferences(schema)
	}
}

func (g *Gen) updateReferences(obj map[string]interface{}) {
	for key, value := range obj {
		if key == "$ref" {
			if ref, ok := value.(string); ok {
				if strings.HasPrefix(ref, "#/definitions/") {
					obj[key] = strings.Replace(ref, "#/definitions/", "#/components/schemas/", 1)
				}
			}
		} else if subObj, ok := value.(map[string]interface{}); ok {
			g.updateReferences(subObj)
		} else if arr, ok := value.([]interface{}); ok {
			for _, item := range arr {
				if itemObj, ok := item.(map[string]interface{}); ok {
					g.updateReferences(itemObj)
				}
			}
		}
	}
}

var packageTemplate = `// Package {{.PackageName}} Code generated by swag/swag{{ if .GeneratedTime }} at {{ .Timestamp }}{{ end }}. DO NOT EDIT
package {{.PackageName}}

import "github.com/venosm/swag"

const docTemplate{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }}{{ .State }} = ` + "`{{ printDoc .Doc}}`" + `

// Swagger{{ .State }}Info{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }} holds exported Swagger Info so clients can modify it
var Swagger{{ .State }}Info{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }} = &swag.Spec{
	Version:     {{ printf "%q" .Version}},
	Host:        {{ printf "%q" .Host}},
	BasePath:    {{ printf "%q" .BasePath}},
	Schemes:     []string{ {{ range $index, $schema := .Schemes}}{{if gt $index 0}},{{end}}{{printf "%q" $schema}}{{end}} },
	Title:       {{ printf "%q" .Title}},
	Description: {{ printf "%q" .Description}},
	InfoInstanceName: {{ printf "%q" .InstanceName }},
	SwaggerTemplate: docTemplate{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }}{{ .State }},
	LeftDelim:        {{ printf "%q" .LeftTemplateDelim}},
	RightDelim:       {{ printf "%q" .RightTemplateDelim}},
}

func init() {
	swag.Register(Swagger{{ .State }}Info{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }}.InstanceName(), Swagger{{ .State }}Info{{ if ne .InstanceName "swagger" }}{{ .InstanceName }} {{- end }})
}
`
