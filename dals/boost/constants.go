package boost

// enums for the verbs/actions/commands for `/boost` in Slack
const (
	Status = iota
	Restart
	Exclusions
)

var SlackCommands = map[int]string{
	Status:     "status",
	Restart:    "restart",
	Exclusions: "exclusions",
}

var SlackOptions = []string{
	"`/boost " + SlackCommands[Status] + " <siteId>` - gets current status of a site",
	"`/boost " + SlackCommands[Restart] + " <siteId>` - to restart a site",
	"`/boost " + SlackCommands[Exclusions] + " <siteId>` - list exclusion stats for a site",
}
