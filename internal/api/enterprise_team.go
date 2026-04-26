package api

import "fmt"

func ListEnterpriseTeamMembers(client *Client, enterprise, teamSlug string) ([]TeamMember, error) {
	var all []TeamMember
	for page := 1; ; page++ {
		path := fmt.Sprintf("enterprises/%s/teams/%s/memberships?per_page=100&page=%d", enterprise, teamSlug, page)
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
