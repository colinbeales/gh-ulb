package api

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

var ErrUniversalBudgetNotFound = errors.New("universal budget not found")

const (
	BudgetProductSkuAICredits      = "ai_credits"
	BudgetProductSkuPremiumRequest = "premium_requests"
	EnvUsePremiumRequests          = "GH_ULB_USE_PREMIUM_REQUESTS"
)

// ResolveBudgetProductSku returns the default SKU based on environment policy.
// ai_credits is always the default; premium_requests is only enabled via env opt-in.
func ResolveBudgetProductSku() (string, error) {
	usePremium, err := usePremiumRequestsFromEnv()
	if err != nil {
		return "", err
	}
	if usePremium {
		return BudgetProductSkuPremiumRequest, nil
	}
	return BudgetProductSkuAICredits, nil
}

// ValidateBudgetProductSkuEnv validates GH_ULB_USE_PREMIUM_REQUESTS and returns an error on invalid values.
func ValidateBudgetProductSkuEnv() error {
	_, err := usePremiumRequestsFromEnv()
	return err
}

func usePremiumRequestsFromEnv() (bool, error) {
	v := strings.TrimSpace(strings.ToLower(os.Getenv(EnvUsePremiumRequests)))
	if v == "" {
		return false, nil
	}
	switch v {
	case "1", "true", "yes":
		return true, nil
	case "0", "false", "no":
		return false, nil
	default:
		return false, fmt.Errorf("invalid %s value %q (allowed: true/false, 1/0, yes/no)", EnvUsePremiumRequests, os.Getenv(EnvUsePremiumRequests))
	}
}

type BudgetAlerting struct {
	AlertRecipients []string `json:"alert_recipients"`
	WillAlert       bool     `json:"will_alert"`
}

type Budget struct {
	BudgetAlerting      BudgetAlerting `json:"budget_alerting"`
	ID                  string         `json:"id"`
	BudgetScope         string         `json:"budget_scope"`
	BudgetProductSku    string         `json:"budget_product_sku"`
	BudgetType          string         `json:"budget_type"`
	User                string         `json:"user,omitempty"`
	BudgetAmount        float64        `json:"budget_amount"`
	PreventFurtherUsage bool           `json:"prevent_further_usage"`
}

type CreateBudgetParams struct {
	BudgetAlerting      BudgetAlerting `json:"budget_alerting"`
	BudgetScope         string         `json:"budget_scope"`
	BudgetProductSku    string         `json:"budget_product_sku"`
	BudgetType          string         `json:"budget_type"`
	User                string         `json:"user,omitempty"`
	BudgetAmount        float64        `json:"budget_amount"`
	PreventFurtherUsage bool           `json:"prevent_further_usage"`
}

type ListBudgetsOptions struct {
	User         string
	BudgetTarget string
}

type listBudgetsResponse struct {
	Budgets     []Budget `json:"budgets"`
	HasNextPage bool     `json:"has_next_page"`
}

func CreateBudget(client *Client, enterprise string, params CreateBudgetParams) (*Budget, error) {
	var result Budget
	path := fmt.Sprintf("enterprises/%s/settings/billing/budgets", enterprise)
	if err := client.Post(path, params, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func ListBudgets(client *Client, enterprise string, opts ListBudgetsOptions) ([]Budget, error) {
	var all []Budget
	for page := 1; ; page++ {
		path := fmt.Sprintf("enterprises/%s/settings/billing/budgets?per_page=10&page=%d", enterprise, page)
		if opts.User != "" {
			path += "&user=" + opts.User
		}
		if opts.BudgetTarget != "" {
			path += "&budgetTarget=" + opts.BudgetTarget
		}
		var result listBudgetsResponse
		if err := client.Get(path, &result); err != nil {
			return nil, err
		}
		all = append(all, result.Budgets...)
		if !result.HasNextPage {
			break
		}
	}
	return all, nil
}

func UpdateBudget(client *Client, enterprise string, budgetID string, amount float64) (*Budget, error) {
	var result Budget
	path := fmt.Sprintf("enterprises/%s/settings/billing/budgets/%s", enterprise, budgetID)
	body := map[string]interface{}{"budget_amount": amount, "prevent_further_usage": true}
	if err := client.Patch(path, body, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func DeleteBudget(client *Client, enterprise string, budgetID string) error {
	path := fmt.Sprintf("enterprises/%s/settings/billing/budgets/%s", enterprise, budgetID)
	return client.Delete(path, nil)
}

func DeleteUniversalBudget(client *Client, enterprise string) (string, error) {
	budgets, err := ListBudgets(client, enterprise, ListBudgetsOptions{})
	if err != nil {
		return "", err
	}

	for _, budget := range budgets {
		if budget.BudgetScope == "multi_user_customer" {
			if err := DeleteBudget(client, enterprise, budget.ID); err != nil {
				return "", err
			}
			return budget.ID, nil
		}
	}

	return "", ErrUniversalBudgetNotFound
}
