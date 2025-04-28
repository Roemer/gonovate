package datasources

import "github.com/google/go-github/v71/github"

func getGitHubClient(ds *datasourceBase) *github.Client {
	// Create a blank client
	client := github.NewClient(nil)
	// Get a host rule if any was defined
	relevantHostRule := ds.getHostRuleForHost("api.github.com")
	// Add the token to the client
	if relevantHostRule != nil {
		token := relevantHostRule.TokendExpanded()
		// But only if the token is set
		if len(token) > 0 {
			client = client.WithAuthToken(token)
		}
	}
	// Return the client
	return client
}
