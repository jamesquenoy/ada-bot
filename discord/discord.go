package discord

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/adayoung/ada-bot/ire"
)

var BotID string
var dg *discordgo.Session

func init() {
	rand.Seed(time.Now().Unix())
}

func InitDiscordSession(token string, q_length int, wait_ms string) error {
	// Create a new Discord session using the provided login information.
	var err error
	if dg, err = discordgo.New(fmt.Sprintf("Bot %s", token)); err == nil {
		if u, err := dg.User("@me"); err == nil {
			BotID = u.ID

			dg.AddHandler(ready)
			// Add handlers for messages received
			dg.AddHandler(messageCreate)
			if err := dg.Open(); err == nil {
				fmt.Println("Successfully launched a new Discord session.")
			} else {
				return err // Error at opening Discord Session
			}
		} else {
			return err // Error at obtaining account details
		}
	} else {
		return err // Error at creating a new Discord session
	}

	if _wait_ms, err := time.ParseDuration(wait_ms); err == nil {
		messageQueue = make(chan message, q_length)
		rateLimit = time.NewTicker(_wait_ms)
		go dispatchMessages()
	} else {
		return err
	}
	return nil
}

func PostMessage(c string, m string) {
	mq := message{ChannelID: c, Message: m}
	messageQueue <- mq
}

func CloseDiscordSession() {
	dg.Close()
}

func ready(s *discordgo.Session, r *discordgo.Ready) {
	if guilds, err := s.UserGuilds(); err != nil {
		fmt.Println("ERROR: We couldn't get UserGuilds")
		log.Fatalf("error: %v", err)
	} else {
		for index, guild := range guilds {
			fmt.Printf("[%d] ------------------------------\n", index)
			fmt.Println("Guild ID: ", guild.ID)
			fmt.Println("Guild Name: ", guild.Name)
			fmt.Println("Guild Permissions: ", guild.Permissions)

		}
		fmt.Println("----------------------------------")
	}
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.ID == BotID { // ignore the bot's own messages from processing
		return
	}

	var GuildID string
	if c, err := s.State.Channel(m.ChannelID); err != nil {
		fmt.Println("Oops, error at getting session.State.Channel,", err)
		return // Not a fatal error
	} else {
		if c.GuildID == "" {
			fmt.Printf("Message received from %s: %s\n", m.Author.Username, m.Content)
			if strings.ToLower(m.Content) == "ping" {
				PostMessage(m.ChannelID, "Pong!")
				return
			}
		} else {
			GuildID = c.GuildID
		}
	}

	if strings.ToLower(m.Content) == "!ping" {
		PostMessage(m.ChannelID, "Pong!")
	}

	if strings.ToLower(m.Content) == "!pink" {
		PostMessage(m.ChannelID, "I love pink!")
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "!decide") {
		choices := strings.Split(m.Content[8:], " or ")
		the_answer := choices[rand.Intn(len(choices))]
		PostMessage(m.ChannelID, fmt.Sprintf("The correct answer is **%s**", the_answer))
	}

	if strings.HasPrefix(strings.ToLower(m.Content), "!whois") {
		r_player := strings.ToLower(strings.TrimSpace(m.Content[7:]))
		if strings.HasPrefix(r_player, "<@") { // It's a @mention and requires fetching a 'Nick'
			filter_exp := regexp.MustCompile("[^0-9]+")          // A userID is numbers only, we shall filter!
			r_player = filter_exp.ReplaceAllString(r_player, "") // Get rid of anything but numbers
			if member, err := s.State.Member(GuildID, r_player); err == nil {
				if member != nil {
					r_player = member.Nick
					if r_player == "" {
						r_player = member.User.Username
					}
				}
			} else {
				log.Printf("error: %v", err) // Not a fatal error, r_player is left unmodified
			}
		}

		if g_player, err := ire.GetPlayer(r_player); err == nil {
			if g_player != nil {
				PostMessage(m.ChannelID, fmt.Sprintf("```%s```", g_player))
			} else {
				PostMessage(m.ChannelID, fmt.Sprintf("Oops, I couldn't find %s :frowning:", r_player))
			}
		} else {
			log.Printf("error: %v", err) // Not a fatal error
		}
	}

	if strings.ToLower(m.Content) == "!help" {
		help_text := `
I have the following commands available:
!ping                  - Pong!
!whois <name>          - Lookup <name> in game and report findings.
!decide thing or thang - Let the bot decide between two or more things for you!`
		PostMessage(m.ChannelID, fmt.Sprintf("```%s```", help_text))
	}
}
