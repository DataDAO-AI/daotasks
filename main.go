package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
)

const (
	dataDaoGuildId = "1202300593214070864"
)

var (
	s *discordgo.Session
	g *github.Client
)

func main() {
	// Load env variables and make sure the required ones are set
	_, err := os.Stat(".env")
	if !os.IsNotExist(err) {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Could not load .env file: %s\n", err)
		}
	}

	for _, v := range []string{"GITHUB_TOKEN", "DISCORD_TOKEN"} {
		if d := os.Getenv(v); d == "" {
			log.Fatalf("Could not find '%s' environment variable\n", v)
		}
	}

	// Set up GitHub client
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")})
	tc := oauth2.NewClient(ctx, ts)
	g = github.NewClient(tc)
	cancel()

	// Set up Discord session
	if s, err = discordgo.New("Bot " + os.Getenv("DISCORD_TOKEN")); err != nil {
		log.Fatalf("Error creating Discord session: %s\n", err)
	}
	s.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Printf("Logged in as: %s#%s", s.State.User.Username, s.State.User.Discriminator)
	})

	// Handle slash commands
	s.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if i.Interaction.Type != discordgo.InteractionApplicationCommand {
			return
		}

		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		} else {
			log.Printf("Unknown command: %s\n", i.ApplicationCommandData().Name)
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "Unknown command.",
				},
			})
		}
	})

	s.Identify.Intents = discordgo.IntentsAll
	s.ShouldReconnectOnError = true
	s.ShouldRetryOnRateLimit = true

	err = s.Open()
	if err != nil {
		log.Fatalf("Error opening connection to Discord: %s\n", err)
	}
	defer s.Close()

	// Register slash commands
	registeredCommands, err := s.ApplicationCommandBulkOverwrite(s.State.User.ID, dataDaoGuildId, commands)
	if err != nil {
		log.Fatalf("Error registering commands: %v\n", err)
	}

	// Defer command deletion
	defer func() {
		var wg sync.WaitGroup
		for _, v := range registeredCommands {
			wg.Add(1)
			go func() {
				err := s.ApplicationCommandDelete(s.State.User.ID, v.GuildID, v.ID)
				if err != nil {
					log.Printf("Error deleting command '%s': %v\n", v.Name, err)
				}
				wg.Done()
			}()
		}
		wg.Wait()
	}()

	log.Println("Bot is now running. Press CTRL+C to exit.")

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-ch
}
