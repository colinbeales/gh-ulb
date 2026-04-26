package csv

import (
	stdcsv "encoding/csv"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type UserBudgetEntry struct {
	Username  string
	Amount    float64
	HasAmount bool
}

func Parse(r io.Reader) ([]UserBudgetEntry, error) {
	reader := stdcsv.NewReader(r)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("reading CSV: %w", err)
	}
	if len(records) == 0 {
		return nil, nil
	}

	headers := records[0]
	usernameCol := -1
	amountCol := -1
	for i, h := range headers {
		switch strings.ToLower(strings.TrimSpace(h)) {
		case "username", "user", "login":
			usernameCol = i
		case "amount":
			amountCol = i
		}
	}
	if usernameCol == -1 {
		return nil, fmt.Errorf("CSV must have a 'username', 'user', or 'login' column")
	}

	var entries []UserBudgetEntry
	for rowIdx, record := range records[1:] {
		username := strings.TrimSpace(record[usernameCol])
		if username == "" {
			continue
		}
		entry := UserBudgetEntry{Username: username}
		if amountCol != -1 && amountCol < len(record) {
			raw := strings.TrimSpace(record[amountCol])
			if raw != "" {
				v, parseErr := strconv.ParseFloat(raw, 64)
				if parseErr != nil {
					return nil, fmt.Errorf("row %d: invalid amount %q: %w", rowIdx+2, raw, parseErr)
				}
				entry.Amount = v
				entry.HasAmount = true
			}
		}
		entries = append(entries, entry)
	}
	return entries, nil
}
