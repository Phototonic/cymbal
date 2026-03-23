// Package summarize generates AI summaries for code symbols using oneagent.
// No API keys required — uses whatever agent CLIs (claude, codex, etc.) the user
// already has installed and authenticated.
package summarize

import (
	"fmt"
	"os"
	"strings"

	"github.com/1broseidon/oneagent"
)

const batchSize = 10

// SymbolInput is the minimal info needed to summarize a symbol.
type SymbolInput struct {
	Name     string
	Kind     string
	Language string
	Source   string
}

// Summarizer generates one-line descriptions for code symbols.
type Summarizer struct {
	backends map[string]oneagent.Backend
	backend  string
	model    string
}

// New creates a Summarizer. Backend and model are optional.
func New(backend, model string) (*Summarizer, error) {
	backends, err := oneagent.LoadBackends("")
	if err != nil {
		return nil, fmt.Errorf("loading backends: %w", err)
	}

	if backend == "" {
		backend = pickAvailableBackend(backends)
	}

	if _, ok := backends[backend]; !ok {
		return nil, fmt.Errorf("backend %q not found — available: %s", backend, availableBackends(backends))
	}

	return &Summarizer{
		backends: backends,
		backend:  backend,
		model:    model,
	}, nil
}

// SummarizeBatch generates summaries for a batch of symbols in one prompt.
// Returns a map of index -> summary for successful results.
func (s *Summarizer) SummarizeBatch(batch []SymbolInput) map[int]string {
	if len(batch) == 0 {
		return nil
	}

	// Single symbol — simpler prompt.
	if len(batch) == 1 {
		result := s.summarizeSingle(batch[0])
		if result == "" {
			return nil
		}
		return map[int]string{0: result}
	}

	// Build batch prompt.
	var b strings.Builder
	b.WriteString("For each numbered code symbol below, write exactly ONE short summary sentence (max 15 words). ")
	b.WriteString("Output format: one line per symbol, numbered to match. No preamble, no quotes, no extra text.\n\n")

	for i, sym := range batch {
		fmt.Fprintf(&b, "--- %d. %s %s (%s) ---\n%s\n\n", i+1, sym.Kind, sym.Name, sym.Language, sym.Source)
	}

	resp := oneagent.Run(s.backends, oneagent.RunOpts{
		Backend: s.backend,
		Model:   s.model,
		Prompt:  b.String(),
	})

	if resp.Error != "" {
		return nil
	}

	return parseBatchResponse(resp.Result, len(batch))
}

func (s *Summarizer) summarizeSingle(sym SymbolInput) string {
	prompt := fmt.Sprintf(
		"Summarize this %s %s in exactly one sentence (max 15 words). No preamble, no quotes, just the summary:\n\n```%s\n%s\n```",
		sym.Language, sym.Kind, sym.Language, sym.Source,
	)

	resp := oneagent.Run(s.backends, oneagent.RunOpts{
		Backend: s.backend,
		Model:   s.model,
		Prompt:  prompt,
	})

	if resp.Error != "" {
		return ""
	}
	return cleanSummary(resp.Result)
}

// parseBatchResponse extracts numbered summaries from the model's response.
func parseBatchResponse(response string, count int) map[int]string {
	results := make(map[int]string)
	lines := strings.Split(strings.TrimSpace(response), "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to parse "1. summary" or "1: summary" or just numbered lines.
		idx := -1
		rest := line
		for i, ch := range line {
			if ch >= '0' && ch <= '9' {
				continue
			}
			if (ch == '.' || ch == ':' || ch == ')') && i > 0 {
				n := 0
				for _, d := range line[:i] {
					n = n*10 + int(d-'0')
				}
				if n >= 1 && n <= count {
					idx = n - 1
					rest = strings.TrimSpace(line[i+1:])
				}
			}
			break
		}

		if idx >= 0 && rest != "" {
			results[idx] = cleanSummary(rest)
		}
	}
	return results
}

func cleanSummary(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "\"'`")
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		s = s[:idx]
	}
	// Cap length.
	if len(s) > 200 {
		s = s[:200]
	}
	return s
}

// BatchSize returns the recommended batch size.
func BatchSize() int {
	return batchSize
}

// Backend returns the backend being used.
func (s *Summarizer) Backend() string {
	return s.backend
}

// pickAvailableBackend tries common backends in preference order.
func pickAvailableBackend(backends map[string]oneagent.Backend) string {
	for _, name := range []string{"claude", "codex", "opencode", "pi", "gemini"} {
		if b, ok := backends[name]; ok {
			if _, found := oneagent.ResolveBackendProgram(b); found {
				return name
			}
		}
	}
	for name, b := range backends {
		if _, found := oneagent.ResolveBackendProgram(b); found {
			return name
		}
	}
	return "claude"
}

func availableBackends(backends map[string]oneagent.Backend) string {
	var names []string
	for name, b := range backends {
		if _, found := oneagent.ResolveBackendProgram(b); found {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "(none installed)"
	}
	return strings.Join(names, ", ")
}

// Available reports whether at least one backend is usable.
func Available() bool {
	backends, err := oneagent.LoadBackends("")
	if err != nil {
		return false
	}
	for _, b := range backends {
		if _, found := oneagent.ResolveBackendProgram(b); found {
			return true
		}
	}
	return false
}

// PrintAvailable prints which backends are ready to use.
func PrintAvailable() {
	backends, err := oneagent.LoadBackends("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "No backends available: %v\n", err)
		return
	}
	for name, b := range backends {
		_, found := oneagent.ResolveBackendProgram(b)
		status := "not found"
		if found {
			status = "ready"
		}
		fmt.Fprintf(os.Stderr, "  %-12s %s\n", name, status)
	}
}
