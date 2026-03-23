package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/1broseidon/cymbal/internal/index"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context <symbol>",
	Short: "Bundled context: source, type references, callers, and imports",
	Long: `Show bundled context for a symbol: source code, referenced types,
callers, and imports of the defining file.

Examples:
  cymbal context OpenStore
  cymbal context ParseFile --callers 10`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		dbPath := getDBPath(cmd)
		jsonOut := getJSONFlag(cmd)
		callers, _ := cmd.Flags().GetInt("callers")

		result, err := index.SymbolContext(dbPath, name, callers)
		if err != nil {
			var ambig *index.AmbiguousError
			if errors.As(err, &ambig) {
				if jsonOut {
					return writeJSON(map[string]any{
						"ambiguous": true,
						"matches":   ambig.Matches,
					})
				}
				fmt.Fprintf(os.Stderr, "Multiple matches for '%s' — be more specific:\n", name)
				for _, r := range ambig.Matches {
					fmt.Printf("  %-12s %-40s %s:%d-%d\n", r.Kind, r.Name, r.RelPath, r.StartLine, r.EndLine)
				}
				os.Exit(1)
			}
			return err
		}

		if jsonOut {
			return writeJSON(result)
		}

		sym := result.Symbol

		// Header.
		fmt.Printf("=== Symbol: %s ===\n", sym.Name)
		fmt.Printf("Kind:     %s\n", sym.Kind)
		fmt.Printf("File:     %s:%d-%d\n", sym.RelPath, sym.StartLine, sym.EndLine)
		fmt.Printf("Language: %s\n", sym.Language)

		// Source.
		fmt.Println()
		fmt.Println("=== Source ===")
		lines := strings.Split(strings.TrimRight(result.Source, "\n"), "\n")
		for i, line := range lines {
			fmt.Printf("%4d  %s\n", sym.StartLine+i, line)
		}

		// Type references.
		if len(result.TypeRefs) > 0 {
			fmt.Println()
			fmt.Printf("=== Type References ===\n")
			for _, r := range result.TypeRefs {
				fmt.Printf("  %-12s %-20s %s:%d\n", r.Kind, r.Name, r.RelPath, r.StartLine)
			}
		}

		// Callers.
		if len(result.Callers) > 0 {
			fmt.Println()
			fmt.Printf("=== Callers (%d) ===\n", len(result.Callers))
			for _, r := range result.Callers {
				fmt.Printf("  %s:%d\n", r.RelPath, r.Line)
			}
		}

		// File imports.
		if len(result.FileImports) > 0 {
			fmt.Println()
			fmt.Println("=== File Imports ===")
			for _, imp := range result.FileImports {
				fmt.Printf("  %s\n", imp)
			}
		}

		return nil
	},
}

func init() {
	contextCmd.Flags().IntP("callers", "n", 20, "max callers to show")
	rootCmd.AddCommand(contextCmd)
}
