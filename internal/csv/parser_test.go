package csv

import (
	"os"
	"strings"
	"testing"
)

func TestParse_ValidCSV(t *testing.T) {
	f, err := os.Open("../../testdata/valid.csv")
	if err != nil {
		t.Fatalf("opening testdata: %v", err)
	}
	defer f.Close()

	entries, err := Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}

	want := []struct {
		username string
		amount   float64
	}{
		{"octocat", 50.0},
		{"monalisa", 25.0},
		{"hubot", 10.0},
	}
	for i, w := range want {
		if entries[i].Username != w.username {
			t.Errorf("entry %d: expected username %q, got %q", i, w.username, entries[i].Username)
		}
		if entries[i].Amount != w.amount {
			t.Errorf("entry %d: expected amount %f, got %f", i, w.amount, entries[i].Amount)
		}
		if !entries[i].HasAmount {
			t.Errorf("entry %d: expected HasAmount=true", i)
		}
	}
}

func TestParse_UsernameOnly(t *testing.T) {
	f, err := os.Open("../../testdata/username_only.csv")
	if err != nil {
		t.Fatalf("opening testdata: %v", err)
	}
	defer f.Close()

	entries, err := Parse(f)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for i, e := range entries {
		if e.HasAmount {
			t.Errorf("entry %d: expected HasAmount=false, got true", i)
		}
		if e.Amount != 0 {
			t.Errorf("entry %d: expected Amount=0, got %f", i, e.Amount)
		}
	}
}

func TestParse_InvalidCSV(t *testing.T) {
	f, err := os.Open("../../testdata/invalid.csv")
	if err != nil {
		t.Fatalf("opening testdata: %v", err)
	}
	defer f.Close()

	_, err = Parse(f)
	if err == nil {
		t.Fatal("expected error for missing username column, got nil")
	}
}

func TestParse_EmptyFile(t *testing.T) {
	entries, err := Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("expected 0 entries, got %d", len(entries))
	}
}

func TestParse_MalformedAmount(t *testing.T) {
	input := "username,amount\noctocat,notanumber\n"
	_, err := Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("expected parse error for bad amount, got nil")
	}
}

func TestParse_SkipsEmptyUsernames(t *testing.T) {
	input := "username,amount\noctocat,10.0\n,20.0\nmonalisa,30.0\n"
	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries (empty username skipped), got %d", len(entries))
	}
	if entries[0].Username != "octocat" {
		t.Errorf("expected first username octocat, got %s", entries[0].Username)
	}
	if entries[1].Username != "monalisa" {
		t.Errorf("expected second username monalisa, got %s", entries[1].Username)
	}
}

func TestParse_CaseInsensitiveHeaders(t *testing.T) {
	input := "LOGIN,AMOUNT\noctocat,15.0\n"
	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Username != "octocat" {
		t.Errorf("expected username octocat, got %s", entries[0].Username)
	}
	if entries[0].Amount != 15.0 {
		t.Errorf("expected amount 15.0, got %f", entries[0].Amount)
	}
}

func TestParse_UserColumnAlias(t *testing.T) {
	input := "user,amount\nhubot,5.0\n"
	entries, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Username != "hubot" {
		t.Errorf("expected username hubot, got %s", entries[0].Username)
	}
}
