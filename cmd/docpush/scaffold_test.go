package main

import "testing"

func TestTitleToSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Data Hub & Storage", "data-hub-storage"},
		{"Benefit Enhancements", "benefit-enhancements"},
		{"Change Requests", "change-requests"},
		{"Task Management", "task-management"},
		{"Mass Communications / Email Campaigns", "mass-communications-email-campaigns"},
		{"Virus / Malware Scanning", "virus-malware-scanning"},
		{"Remote Identity Proofing (RIDP)", "remote-identity-proofing-ridp"},
		{"CLI Tools", "cli-tools"},
		{"  Leading & Trailing  ", "leading-trailing"},
		{"Already-Kebab-Case", "already-kebab-case"},
		{"Multiple   Spaces", "multiple-spaces"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := titleToSlug(tt.input)
			if got != tt.want {
				t.Errorf("titleToSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGenerateStub(t *testing.T) {
	stub := generateStub("Data Hub & Storage", "confluence")
	if stub == "" {
		t.Fatal("generateStub returned empty string")
	}
	if !contains(stub, `title: "Data Hub & Storage"`) {
		t.Error("stub missing title in frontmatter")
	}
	if !contains(stub, "# Data Hub & Storage") {
		t.Error("stub missing H1 heading")
	}
	if !contains(stub, "TODO") {
		t.Error("stub missing TODO marker")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
