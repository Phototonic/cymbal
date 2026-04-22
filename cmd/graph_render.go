package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/1broseidon/cymbal/index"
)

func renderGraph(format index.GraphFormat, graph *index.GraphResult) error {
	switch format {
	case index.GraphFormatDot:
		_, err := fmt.Fprint(os.Stdout, renderGraphDOT(graph))
		return err
	case index.GraphFormatMermaid:
		_, err := fmt.Fprint(os.Stdout, renderGraphMermaid(graph))
		return err
	default:
		return writeJSON(graph)
	}
}

func renderGraphMermaid(graph *index.GraphResult) string {
	if len(graph.Nodes) == 0 && len(graph.Edges) == 0 {
		return "flowchart LR\n%% no edges\n"
	}
	var b strings.Builder
	b.WriteString("flowchart LR\n")
	for _, node := range graph.Nodes {
		fmt.Fprintf(&b, "  %s[%q]\n", node.ID, node.Label)
	}
	for _, edge := range graph.Edges {
		arrow := "-->"
		if !edge.Resolved {
			arrow = "-.->"
		}
		fmt.Fprintf(&b, "  %s %s %s\n", edge.From, arrow, edge.To)
	}
	return b.String()
}

func renderGraphDOT(graph *index.GraphResult) string {
	if len(graph.Nodes) == 0 && len(graph.Edges) == 0 {
		return "digraph cymbal { /* no edges */ }\n"
	}
	var b strings.Builder
	b.WriteString("digraph cymbal {\n")
	for _, node := range graph.Nodes {
		fmt.Fprintf(&b, "  %s [label=%q];\n", node.ID, node.Label)
	}
	for _, edge := range graph.Edges {
		attrs := ""
		if !edge.Resolved {
			attrs = " [style=dashed]"
		}
		fmt.Fprintf(&b, "  %s -> %s%s;\n", edge.From, edge.To, attrs)
	}
	b.WriteString("}\n")
	return b.String()
}
