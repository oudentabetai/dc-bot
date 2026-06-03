package discord

import (
	"log"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/oudentabetai/pterodactyl-go/pterodactyl"
	"github.com/oudentabetai/pterodactyl-go/storage"
)

var (
	suffix   string = "!!"
	OWNER_ID string = "967088187405107220"
)

type SessionManager interface {
	InitializeSession(token string) *discordgo.Session
}

type DiscordSessionManager struct{}

func (d *DiscordSessionManager) InitializeSession(token string) *discordgo.Session {
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Discordセッションの作成に失敗: %v", err)
	}
	return dg
}

func OnMessageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic recovered in OnMessageCreate: %v", r)
		}
	}()

	if m == nil || m.Author == nil {
		return
	}

	linkfixer.LinkFixer(s, i)

	//TextCommand(s, m)
}


func TextCommand(s *discordgo.Session, m *discordgo.MessageCreate) {
	u := m.Author
	if !u.Bot {
		if strings.HasPrefix(m.Content, suffix) {
			log.Printf("Received command: %s from user: %s", m.Content, u.Username)
		}
		if strings.HasPrefix(m.Content, suffix+"user") {
			s.ChannelMessageSend(m.ChannelID, pterodactyl.GetUser())
		}
		if strings.HasPrefix(m.Content, suffix+"servers") {

		}
		if strings.HasPrefix(m.Content, suffix+"setrole") {
			if m.Author.ID != OWNER_ID {
				log.Print("User does not have permission to set role: " + OWNER_ID + " vs " + m.Author.ID)
				s.ChannelMessageSend(m.ChannelID, "このコマンドを使用する権限がありません。")
				return
			}
			fields := strings.Fields(m.Content)
			if len(fields) == 3 {
				s.ChannelMessageSend(m.ChannelID, storage.ConfigMgr.SetRole(fields[1], fields[2]))
			} else {
				s.ChannelMessageSend(m.ChannelID, "コマンドの形式が正しくありません。例: !!setrole <roleID> <serverID>")
			}
		}
		if strings.HasPrefix(m.Content, suffix+"server") {
			fields := strings.Fields(m.Content)
			if len(fields) == 3 {
				s.ChannelMessageSend(m.ChannelID, pterodactyl.PowerServer(fields[1], fields[2]))
			} else {
				s.ChannelMessageSend(m.ChannelID, "コマンドの形式が正しくありません。例: !!server <action> <serverIdentifier>")
			}
		}
		if strings.HasPrefix(m.Content, suffix+"test") {
		}
	}
}