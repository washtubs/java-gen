package main

import (
	"flag"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
)

const getter = `public {{ .Type }} {{if .IsBoolean }}is{{ else }}get{{ end }}{{ .IdentC }}() {
    return this.{{ .Ident }};
}
`

const setter = `public void set{{ .IdentC }}({{ .Type }} {{ .Ident }}) {
    this.{{ .Ident }} = {{ .Ident }};
}
`

const wither = `public Builder with{{ .IdentC }}({{ .Type }} {{ .Ident }}) {
    this.{{ .Ident }} = {{ .Ident }};
    return this;
}
`

const getters = `{{ range .Fields -}}
` + getter + `
{{ end }}`

const setters = `{{ range .Fields -}}
` + setter + `
{{ end }}`

const withers = `{{ range .Fields -}}
` + wither + `
{{ end }}`

const getterSetters = `{{ range .Fields -}}
` + getter + "\n" + setter + `
{{ end }}`

const constructor = `public {{.Class}}(
{{- range $index, $element := .Fields -}}{{if $index}},{{end}}
    {{ .Type }} {{ .Ident }}
{{- end }}
) {
{{- range .Fields }}
    this.{{ .Ident }} = {{ .Ident }};
{{- end }}
}
`

const constructorNonNull = `public {{.Class}}(
{{- range $index, $element := .Fields -}}{{if $index}},{{end}}
    {{ .Type }} {{ .Ident }}
{{- end }}
) {
{{- range .Fields }}
    this.{{ .Ident }} = requireNonNull({{ .Ident }}, "{{.Ident}} was null");
{{- end }}
}
`

const constructorBuilder = `public {{.Class}}(Builder b) {
{{- range .Fields }}
    this.{{ .Ident }} = requireNonNull(b.{{ .Ident }}, "{{.Ident}} was null");
{{- end }}
}
`

type Field struct {
	Parent    string
	Type      string
	Ident     string
	IdentC    string
	Line      int
	IsBoolean bool
}

type Top struct {
	Fields []Field
	Class  string
}

func extractFields(file string) []Field {
	cmd := exec.Command("_java-gen-helper.js", file)
	out, err := cmd.CombinedOutput()
	if err != nil {
		panic(err.Error())
	}
	lines := strings.Split(string(out), "\n")

	result := make([]Field, 0)

	parent := ""
	for _, line := range lines {
		fields := strings.Split(line, "|")
		if len(fields) == 1 {
			parent = line
			continue
		}
		if parent == "" {
			log.Fatal("Parent class not set")
		}
		if len(fields) != 3 {
			log.Fatal("Fields does not equal 3")
		}

		line, err := strconv.Atoi(strings.TrimSpace(fields[2]))
		if err != nil {
			panic(err.Error())
		}

		t := strings.TrimSpace(fields[0])

		ident := strings.TrimSpace(fields[1])
		identC := strings.ToUpper(ident[0:1]) + ident[1:]

		result = append(result, Field{
			Parent:    parent,
			Type:      t,
			Ident:     ident,
			IdentC:    identC,
			Line:      line,
			IsBoolean: t == "boolean" || t == "Boolean",
		})
	}
	return result
}

func main() {
	t := flag.String("type", "get", "get | set | with | constructor")
	startLine := flag.Int("start", 0, "The line to start reading fields from")
	endLine := flag.Int("end", 0, "The line to stop reading fields from (inclusive)")
	file := flag.String("file", "", "The source file")
	flag.Parse()
	if *file == "" {
		log.Fatal("Must supply a file")
	}
	if *startLine == 0 {
		log.Fatal("Must supply a start line")
	}
	if *endLine == 0 {
		log.Fatal("Must supply a end line")
	}

	fields := extractFields(*file)
	filtered := make([]Field, 0, *endLine-*startLine+1)
	for _, field := range fields {
		if field.Line >= *startLine && field.Line <= *endLine {
			filtered = append(filtered, field)
		}
	}

	if len(filtered) == 0 {
		log.Fatal("Nothing to do")
	}

	tplText := getter
	switch *t {
	case "get":
		tplText = getters
	case "set":
		tplText = setters
	case "getset":
		tplText = getterSetters
	case "with":
		tplText = withers
	case "constructor":
		tplText = constructor
	case "constructorNonNull":
		tplText = constructorNonNull
	case "constructorBuilder":
		tplText = constructorBuilder
	default:
		log.Fatal("Invalid type " + *t)
	}

	tpl, err := template.New("java").Parse(tplText)
	if err != nil {
		log.Fatalf("Failed to parse template: %s", err.Error())
	}

	top := Top{}
	top.Fields = filtered
	top.Class = fields[0].Parent
	err = tpl.Execute(os.Stdout, top)
	if err != nil {
		log.Fatal("Failed to execute template:", err)
	}

}
