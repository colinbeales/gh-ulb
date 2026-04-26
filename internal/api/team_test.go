package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestListOrgTeamMembers_SinglePage(t *testing.T) {
	members := []TeamMember{{Login: "octocat", ID: 1}, {Login: "monalisa", ID: 2}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		if page == "1" {
			json.NewEncoder(w).Encode(members)
		} else {
			json.NewEncoder(w).Encode([]TeamMember{})
		}
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	result, err := ListOrgTeamMembers(client, "my-org", "my-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 members, got %d", len(result))
	}
	if result[0].Login != "octocat" {
		t.Errorf("expected first login octocat, got %s", result[0].Login)
	}
}

func TestListOrgTeamMembers_Pagination(t *testing.T) {
	page1 := make([]TeamMember, 3)
	for i := range page1 {
		page1[i] = TeamMember{Login: "user" + strconv.Itoa(i+1), ID: i + 1}
	}
	page2 := []TeamMember{{Login: "user4", ID: 4}}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1":
			json.NewEncoder(w).Encode(page1)
		case "2":
			json.NewEncoder(w).Encode(page2)
		default:
			json.NewEncoder(w).Encode([]TeamMember{})
		}
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	result, err := ListOrgTeamMembers(client, "my-org", "my-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 4 {
		t.Errorf("expected 4 members, got %d", len(result))
	}
	if result[3].Login != "user4" {
		t.Errorf("expected last login user4, got %s", result[3].Login)
	}
}

func TestListEnterpriseTeamMembers_SinglePage(t *testing.T) {
	members := []TeamMember{{Login: "octocat", ID: 1}, {Login: "monalisa", ID: 2}}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		if page == "1" {
			json.NewEncoder(w).Encode(members)
		} else {
			json.NewEncoder(w).Encode([]TeamMember{})
		}
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	result, err := ListEnterpriseTeamMembers(client, "my-enterprise", "my-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 members, got %d", len(result))
	}
	if result[0].Login != "octocat" {
		t.Errorf("expected first login octocat, got %s", result[0].Login)
	}
}

func TestListEnterpriseTeamMembers_Pagination(t *testing.T) {
	page1 := []TeamMember{{Login: "user1", ID: 1}, {Login: "user2", ID: 2}}
	page2 := []TeamMember{{Login: "user3", ID: 3}}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := r.URL.Query().Get("page")
		w.Header().Set("Content-Type", "application/json")
		switch page {
		case "1":
			json.NewEncoder(w).Encode(page1)
		case "2":
			json.NewEncoder(w).Encode(page2)
		default:
			json.NewEncoder(w).Encode([]TeamMember{})
		}
	}))
	defer ts.Close()

	client := newTestClient(t, ts)
	result, err := ListEnterpriseTeamMembers(client, "my-enterprise", "my-team")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Errorf("expected 3 members, got %d", len(result))
	}
	if result[2].Login != "user3" {
		t.Errorf("expected last login user3, got %s", result[2].Login)
	}
}
