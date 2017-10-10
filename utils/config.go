package utils

type AppConfig struct {
	Port            int    `long:"port" short:"p" description:"the port to bind with, chose a random free port if not provided"`
	Verbose         []bool `long:"verbose" short:"v" description:"show verbose debug information"`
	TeamName        string `long:"team-name" short:"t" description:"my team name, generated randomly if not provided"`
	BaseMediatorURL string `long:"base-mediator-url" short:"m" default:"http://localhost:8080" description:"the game mediator contact point"`
	ManualExit      []bool `long:"manual-exit" description:"if set, do not automatically exit on mediator game end signal"`
}
