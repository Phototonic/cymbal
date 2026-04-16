package cmd

import "testing"

func TestFeatureParseVisibilityFilter(t *testing.T) {
	tests := []struct {
		name       string
		visibility string
		exported   bool
		want       string
		wantErr    bool
	}{
		{name: "empty", visibility: "", exported: false, want: "", wantErr: false},
		{name: "normalized public", visibility: "PUBLIC", exported: false, want: "public", wantErr: false},
		{name: "normalized private", visibility: " Private ", exported: false, want: "private", wantErr: false},
		{name: "exported shorthand", visibility: "", exported: true, want: "public", wantErr: false},
		{name: "conflict exported with private", visibility: "private", exported: true, want: "", wantErr: true},
		{name: "invalid token", visibility: "friend", exported: false, want: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseVisibilityFilter(tt.visibility, tt.exported)
			if (err != nil) != tt.wantErr {
				t.Fatalf("unexpected error state: err=%v wantErr=%v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Fatalf("expected visibility %q, got %q", tt.want, got)
			}
		})
	}
}
