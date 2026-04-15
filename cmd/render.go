package cmd

import (
	"strings"

	"github.com/1broseidon/cymbal/index"
)

func renderJSONOrFrontmatter(jsonOut bool, data any, meta []kv, content string) error {
	if jsonOut {
		return writeJSON(data)
	}
	frontmatter(meta, content)
	return nil
}

func formatImporterResults(results []index.ImporterResult) string {
	var content strings.Builder
	for _, r := range results {
		content.WriteString(r.RelPath)
		content.WriteByte(':')
		content.WriteString(r.Import)
		content.WriteByte('\n')
	}
	return content.String()
}
