package output

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/colinbeales/gh-ulb/internal/api"
)

func sampleBudgets() []api.Budget {
	return []api.Budget{
		{
			ID:                  "b1",
			BudgetAmount:        50.0,
			BudgetScope:         "user",
			BudgetProductSku:    "premium_requests",
			BudgetType:          "BundlePricing",
			PreventFurtherUsage: true,
			User:                "octocat",
		},
		{
			ID:                  "b2",
			BudgetAmount:        25.0,
			BudgetScope:         "user",
			BudgetProductSku:    "premium_requests",
			BudgetType:          "BundlePricing",
			PreventFurtherUsage: false,
			User:                "monalisa",
		},
	}
}

func TestPrintBudgets_Table(t *testing.T) {
	var buf bytes.Buffer
	err := PrintBudgets(&buf, sampleBudgets(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ID") || !strings.Contains(out, "SCOPE") || !strings.Contains(out, "USER") ||
		!strings.Contains(out, "AMOUNT") || !strings.Contains(out, "PREVENT OVERAGE") {
		t.Errorf("output missing expected headers, got:\n%s", out)
	}
	if !strings.Contains(out, "b1") {
		t.Errorf("output missing budget ID b1, got:\n%s", out)
	}
	if !strings.Contains(out, "octocat") {
		t.Errorf("output missing user octocat, got:\n%s", out)
	}
	if !strings.Contains(out, "$50.00") {
		t.Errorf("output missing amount $50.00, got:\n%s", out)
	}
	if !strings.Contains(out, "b2") {
		t.Errorf("output missing budget ID b2, got:\n%s", out)
	}
	if !strings.Contains(out, "monalisa") {
		t.Errorf("output missing user monalisa, got:\n%s", out)
	}
}

func TestPrintBudgets_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := PrintBudgets(&buf, sampleBudgets(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var result []api.Budget
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("output is not valid JSON: %v\n%s", err, buf.String())
	}
	if len(result) != 2 {
		t.Errorf("expected 2 budgets in JSON, got %d", len(result))
	}
	if result[0].ID != "b1" {
		t.Errorf("expected first ID b1, got %s", result[0].ID)
	}
	if result[0].BudgetAmount != 50.0 {
		t.Errorf("expected first amount 50.0, got %f", result[0].BudgetAmount)
	}
	if result[1].User != "monalisa" {
		t.Errorf("expected second user monalisa, got %s", result[1].User)
	}
}

func TestPrintBudgets_Empty(t *testing.T) {
	var buf bytes.Buffer
	err := PrintBudgets(&buf, []api.Budget{}, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	// Should still contain the header line
	if !strings.Contains(out, "ID") {
		t.Errorf("empty table output missing header, got:\n%s", out)
	}
	// Should not contain any data rows
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 header line for empty table, got %d lines:\n%s", len(lines), out)
	}
}

func TestPrintBudget_Single(t *testing.T) {
	b := &api.Budget{
		ID:                  "b3",
		BudgetAmount:        100.0,
		BudgetScope:         "user",
		BudgetProductSku:    "premium_requests",
		BudgetType:          "BundlePricing",
		PreventFurtherUsage: true,
		User:                "hubot",
	}

	var buf bytes.Buffer
	err := PrintBudget(&buf, b, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "b3") {
		t.Errorf("output missing budget ID b3, got:\n%s", out)
	}
	if !strings.Contains(out, "hubot") {
		t.Errorf("output missing user hubot, got:\n%s", out)
	}
	if !strings.Contains(out, "$100.00") {
		t.Errorf("output missing amount $100.00, got:\n%s", out)
	}
}
