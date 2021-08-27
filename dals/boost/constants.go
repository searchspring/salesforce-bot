package boost

// enums for the verbs/actions/commands for `/boost` in Slack
const (
	Status = iota
	Exclusions
	ResumeUpdates
	PauseUpdates
	Update
	Cancel
)

var SlackCommands = map[int]string{
	Status:        "status",
	Exclusions:    "exclusions",
	Update:        "update",
	ResumeUpdates: "resumeUpdates",
	PauseUpdates:  "pauseUpdates",
	Cancel:        "cancel",
}

var SlackOptions = []string{
	"`/boost " + SlackCommands[Status] + " <siteId>` - gets current status of a site",
	"`/boost " + SlackCommands[Exclusions] + " <siteId>` - list exclusion stats for a site",
	"`/boost " + SlackCommands[Update] + " <siteId>` - update a site",
	"`/boost " + SlackCommands[ResumeUpdates] + " <siteId>` - resume updates for a site",
	"`/boost " + SlackCommands[PauseUpdates] + " <siteId>` - pause updates for a site",
	"`/boost " + SlackCommands[Cancel] + " <siteId>` - cancel current update for a site",
}

const boostAdminUrl = "https://boostadmin.azurewebsites.net"
const mainBoostDispatchQueue = "admin"
