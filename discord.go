package main

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
)

// Send a basic ephemeral response to an interaction.
func ephemeralResponse(i *discordgo.InteractionCreate, msg string) {
	if err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}); err != nil {
		log.Printf("Error sending message '%s' to interaction: %v\n", msg, err)
	}
}

// Split string by line into <2000 character messages to stay below Discord's limit.
// Do this by line to avoid splitting tags and titles.
func chunkText(text string) []string {

	charLimit := 2000 // Discord character limit is 2000 characters.
	var chunks []string
	lines := strings.Split(text, "\n")
	chunk := ""
	for _, line := range lines {
		// If the line is too long to fit in a single message, split it by charLimit.
		for len(line) > charLimit {
			splitAt := charLimit - len(chunk)
			chunk += line[:splitAt]
			chunks = append(chunks, chunk)
			chunk = ""
			line = line[splitAt:]
		}

		if len(chunk)+len(line) > charLimit {
			chunks = append(chunks, chunk)
			chunk = ""
		}
		chunk += line + "\n"
	}

	// Add any leftover characters to the last chunk.
	if chunk != "" {
		chunks = append(chunks, chunk)
	}

	return chunks
}
