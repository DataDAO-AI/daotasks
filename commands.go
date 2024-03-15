package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
)

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "help",
		Description: "How to use this bot.",
	},
	{
		Name:        "identify",
		Description: "Link your GitHub username with your Discord account",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "github-username",
				Description: "Your GitHub username",
				Required:    true,
			},
		},
	},
	{
		Name:        "all",
		Description: "List all open tasks",
	},
	{
		Name:        "mine",
		Description: "List all open tasks assigned to you",
	},
	// TODO
	/*{
		Name:        "subscribe-all",
		Description: "Receive notifications for all new tasks",
	},
	{
		Name:        "subscribe-mine",
		Description: "Receive notifications for new tasks assigned to you",
	},*/
}

// A mapping of Discord user IDs to GitHub logins.
var githubLogins = make(map[string]string)

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
	"help": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		ephemeralResponse(i, "This bot lets you read from and subscribe to the DataDAO todo list."+
			"\n\nUse the `/identify` command to link your GitHub username with your Discord account, then use `/all` and `/mine`"+
			" to see all open tasks and open tasks assigned to you, respectively."+
			"\n\nYou can be notified of all new tasks with `/subscribe-all`, or new tasks assigned to you with `/subscribe-mine`.")
	},
	"identify": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Member == nil {
			ephemeralResponse(i, "You must be in a server to use this command.")
			return
		}

		options := i.ApplicationCommandData().Options
		if len(options) != 1 || options[0].Type != discordgo.ApplicationCommandOptionString {
			ephemeralResponse(i, "Invalid options provided. Please provide a single GitHub username.")
			return
		}

		githubUsername := options[0].StringValue()
		if githubUsername == "" {
			ephemeralResponse(i, "You must provide a non-empty GitHub username.")
			return
		}

		githubLogins[i.Member.User.ID] = githubUsername
		ephemeralResponse(i, "Your Discord account has been linked to `"+githubUsername+"` on GitHub.")
	},
	"all": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		issues, err := issues("open")
		if err != nil {
			log.Printf("Error getting open issues: %v\n", err)
			ephemeralResponse(i, "Encountered an error while fetching issues. Please contact an administrator.")
			return
		}

		// Respond in 2,000 character chunks
		msg := issueDescriptions(issues)
		for _, text := range chunkText(msg) {
			ephemeralResponse(i, text)
		}
	},
	"mine": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		issues, err := issues("open")
		if err != nil {
			log.Printf("Error getting open issues: %v\n", err)
			ephemeralResponse(i, "Encountered an error while fetching issues. Please contact an administrator.")
			return
		}

		var myIssues []*github.Issue
		for _, issue := range issues {
			for _, assignee := range issue.Assignees {
				if githubLogins[i.Member.User.ID] == assignee.GetLogin() {
					myIssues = append(myIssues, issue)
					break
				}
			}
		}

		// Respond in 2,000 character chunks
		msg := issueDescriptions(myIssues)
		for _, text := range chunkText(msg) {
			ephemeralResponse(i, text)
		}
	},
}
