package api

import "fmt"

type TeamMember struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
}

func ListOrgTeamMembers(client *Client, org, teamSlug string) ([]TeamMember, error) {
	var all []TeamMember
	for page := 1; ; page++ {
		path := fmt.Sprintf("orgs/%s/teams/%s/members?per_page=100&page=%d", org, teamSlug, page)
		var members []TeamMember
		if err := client.Get(path, &members); err != nil {
			return nil, err
		}
		if len(members) == 0 {
			break
		}
		all = append(all, members...)
	}
	return all, nil
}
