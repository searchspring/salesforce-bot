package boost

// enums for the verbs/actions/commands for `/boost` in Slack
const (
	Status = iota
	Restart
	Exclusions
	Pause
)

var SlackCommands = map[int]string{
	Status:     "status",
	Restart:    "restart",
	Exclusions: "exclusions",
	Pause:      "pause",
}

var SlackOptions = []string{
	"`/boost " + SlackCommands[Status] + " <siteId>` - gets current status of a site",
	"`/boost " + SlackCommands[Restart] + " <siteId>` - to restart a site",
	"`/boost " + SlackCommands[Exclusions] + " <siteId>` - list exclusion stats for a site",
	"`/boost " + SlackCommands[Pause] + " <siteId>` - pause updates for a site",
}

const boostAdminUrl = "https://boostadmin.azurewebsites.net"
