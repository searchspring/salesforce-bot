package boost

// verbs/actions/commands for the `/boost` in Slack
const (
	SiteStatus   = "status"
	SiteRestart  = "restart"
	SiteExclusionStats = "exclusionStats"
)

var SlackOptions = []string{
	"`/boost " + SiteStatus + " <siteId>` - gets current status of a site",
	"`/boost " + SiteRestart + " <siteId>` - to restart a site",
	"`/boost " + SiteExclusionStats + " <siteId>` - list exclusion stats for a site",
}
