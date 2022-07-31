package examples

import (
	"fmt"
	"strings"

	"k8s.io/kubectl/pkg/util/templates"
)

// Example defines fields required to construct
// a CLI command example
type Example struct {
	Descriptions  []string
	RootCommand   string
	CommandString string
}

// String renders a formatted example.
func (e *Example) String() string {
	description := new(strings.Builder)
	for _, d := range e.Descriptions {
		description.WriteString("# ")
		description.WriteString(d)
		description.WriteString("\n")
	}
	return fmt.Sprintf("%s %s %s", description, e.RootCommand, e.CommandString)
}

// FormatExamples formats one or more Example instances.
func FormatExamples(examples ...Example) string {
	allExamples := new(strings.Builder)
	for _, e := range examples {
		allExamples.WriteString(e.String())
		allExamples.WriteString("\n\n")
	}
	return templates.Examples(allExamples.String())
}
