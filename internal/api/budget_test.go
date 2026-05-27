package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func newTestClient(t *testing.T, ts *httptest.Server) *Client {
	t.Helper()
	c, err := NewTestClient(ts.URL)
	if err != nil {
		t.Fatalf("newTestClient: %v", err)
	}
	return c
}

func TestCreateBudget_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(Budget{ID: "b1", BudgetAmount: 10.0, BudgetScope: "user"})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	params := CreateBudgetParams{
		BudgetAmount:     10.0,
		BudgetScope:      "user",
		BudgetProductSku: "premium_requests",
		BudgetType:       "BundlePricing",
		User:             "octocat",
	}
	b, err := CreateBudget(client, "my-enterprise", params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.ID != "b1" {
		t.Errorf("expected ID b1, got %s", b.ID)
	}
	if b.BudgetAmount != 10.0 {
		t.Errorf("expected amount 10.0, got %f", b.BudgetAmount)
	}
	if b.BudgetScope != "user" {
		t.Errorf("expected scope user, got %s", b.BudgetScope)
	}
}

func TestCreateBudget_HTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"message": "validation failed"}`))
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	_, err := CreateBudget(client, "my-enterprise", CreateBudgetParams{BudgetAmount: 10.0})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListBudgets_NoFilter(t *testing.T) {
	budgets := []Budget{
		{ID: "b1", BudgetAmount: 10.0, BudgetScope: "user", User: "octocat"},
		{ID: "b2", BudgetAmount: 20.0, BudgetScope: "user", User: "monalisa"},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"budgets": budgets})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	result, err := ListBudgets(client, "my-enterprise", ListBudgetsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 budgets, got %d", len(result))
	}
	if result[0].ID != "b1" {
		t.Errorf("expected first ID b1, got %s", result[0].ID)
	}
	if result[1].User != "monalisa" {
		t.Errorf("expected second user monalisa, got %s", result[1].User)
	}
}

func TestListBudgets_WithUserFilter(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"budgets": []Budget{}, "has_next_page": false})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	_, err := ListBudgets(client, "my-enterprise", ListBudgetsOptions{User: "octocat"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	q, _ := url.ParseQuery(gotQuery)
	if q.Get("user") != "octocat" {
		t.Errorf("expected user=octocat in query, got %s", gotQuery)
	}
}

func TestListBudgets_Pagination(t *testing.T) {
	page1 := []Budget{
		{ID: "b1", BudgetAmount: 10.0, BudgetScope: "user", User: "user1"},
	}
	page2 := []Budget{
		{ID: "b2", BudgetAmount: 20.0, BudgetScope: "user", User: "user2"},
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("page") {
		case "1":
			json.NewEncoder(w).Encode(map[string]interface{}{"budgets": page1, "has_next_page": true})
		default:
			json.NewEncoder(w).Encode(map[string]interface{}{"budgets": page2, "has_next_page": false})
		}
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	result, err := ListBudgets(client, "my-enterprise", ListBudgetsOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 budgets across pages, got %d", len(result))
	}
	if result[0].ID != "b1" || result[1].ID != "b2" {
		t.Errorf("unexpected budget IDs: got %s, %s", result[0].ID, result[1].ID)
	}
}

func TestUpdateBudget_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Errorf("expected PATCH, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Budget{ID: "b1", BudgetAmount: 99.0, BudgetScope: "user", User: "octocat"})
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	b, err := UpdateBudget(client, "my-enterprise", "b1", 99.0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if b.BudgetAmount != 99.0 {
		t.Errorf("expected amount 99.0, got %f", b.BudgetAmount)
	}
	if b.ID != "b1" {
		t.Errorf("expected ID b1, got %s", b.ID)
	}
}

func TestDeleteBudget_Success(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	err := DeleteBudget(client, "my-enterprise", "b1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteUniversalBudget_Success(t *testing.T) {
	var deletedID string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/budgets") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"budgets": []Budget{
					{ID: "u1", BudgetScope: "multi_user_customer"},
					{ID: "b1", BudgetScope: "user", User: "octocat"},
				},
				"has_next_page": false,
			})
			return
		}

		if r.Method == http.MethodDelete {
			parts := strings.Split(r.URL.Path, "/")
			deletedID = parts[len(parts)-1]
			w.WriteHeader(http.StatusNoContent)
			return
		}

		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	id, err := DeleteUniversalBudget(client, "my-enterprise")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "u1" {
		t.Fatalf("expected deleted ID u1, got %s", id)
	}
	if deletedID != "u1" {
		t.Fatalf("expected DELETE on u1, got %s", deletedID)
	}
}

func TestDeleteUniversalBudget_NotFound(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/budgets") && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"budgets": []Budget{
					{ID: "b1", BudgetScope: "user", User: "octocat"},
				},
				"has_next_page": false,
			})
			return
		}

		if r.Method == http.MethodDelete {
			t.Fatal("unexpected delete request when universal budget does not exist")
		}

		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	_, err := DeleteUniversalBudget(client, "my-enterprise")
	if !errors.Is(err, ErrUniversalBudgetNotFound) {
		t.Fatalf("expected ErrUniversalBudgetNotFound, got %v", err)
	}
}
