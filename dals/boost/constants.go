package boost

// verbs/actions/commands for `/boost` in Slack
const (
	SiteStatus     = "status"
	SiteRestart    = "restart"
	SiteExclusions = "exclusions"
)

var SlackOptions = []string{
	"`/boost " + SiteStatus + " <siteId>` - gets current status of a site",
	"`/boost " + SiteRestart + " <siteId>` - to restart a site",
	"`/boost " + SiteExclusions + " <siteId>` - list exclusion stats for a site",
}
