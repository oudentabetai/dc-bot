package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/oudentabetai/pterodactyl-go/pterodactyl"
	"github.com/oudentabetai/pterodactyl-go/storage"
)

const panelBaseURL = "https://web.ofton.dev"

type PteroResponse[T any] struct {
	Data []T `json:"data"`
}

type Server struct {
	Attributes struct {
		ID         int    `json:"id"`
		UUID       string `json:"uuid"`
		Identifier string `json:"identifier"`
		Name       string `json:"name"`
		Status     string `json:"status"`
	} `json:"attributes"`
}

type User struct {
	Attributes struct {
		ID     int    `json:"id"`
		Name   string `json:"username"`
		Email  string `json:"email"`
		IsRoot bool   `json:"is_root"`
	} `json:"attributes"`
}

type Status string

const (
	StatusRunning  Status = "running"
	StatusOffline  Status = "offline"
	StatusStarting Status = "starting"
	StatusStopping Status = "stopping"
)

func (s Status) ToJapanese() string {
	switch s {
	case StatusRunning:
		return "起動中"
	case StatusOffline:
		return "停止"
	case StatusStarting:
		return "起動処理中"
	case StatusStopping:
		return "停止処理中"
	default:
		return string(s)
	}
}

func GetRoles(roleIDs []string) []string {
	var serverIDs []string
	for _, roleID := range roleIDs {
		serverIDs = append(serverIDs, storage.ConfigMgr.GetServerID(roleID)...)
	}
	return serverIDs
}

func DecodeServerResponse(resp *http.Response) *PteroResponse[Server] {
	var serverres PteroResponse[Server]
	if err := json.NewDecoder(resp.Body).Decode(&serverres); err != nil {
		log.Printf("Decode error: %v", err)
		return nil
	}
	return &serverres
}

func ListServers(serverres []Server) *discordgo.MessageEmbed {
	if serverres == nil {
		return &discordgo.MessageEmbed{
			Title:       "サーバーリスト",
			Description: "サーバー情報の取得に失敗しました。",
			Color:       0xff0000,
		}
	}

	embed := &discordgo.MessageEmbed{
		Title:       "サーバーリスト",
		Description: "各サーバーの状態とパネルリンクをまとめています。",
		Color:       0x00ff00,
	}

	for _, sv := range serverres {
		serverURL := panelBaseURL + "/server/" + url.PathEscape(sv.Attributes.Identifier)
		status := Status(pterodactyl.GetServerStatus(sv.Attributes.Identifier))

		fieldValue := fmt.Sprintf("📊 **Status**: %s %s\n🔑 **Identifier**: `%s`\n🔗 [パネルを開く](%s)",
			statusEmoji(status),
			status.ToJapanese(),
			sv.Attributes.Identifier,
			serverURL,
		)

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%s", sv.Attributes.Name),
			Value:  fieldValue,
			Inline: false,
		})
	}
	return embed
}

func statusEmoji(status Status) string {
	switch status {
	case StatusRunning:
		return "🟢"
	case StatusOffline:
		return "🔴"
	case StatusStarting:
		return "🟡"
	case StatusStopping:
		return "🟠"
	default:
		return "⚪"
	}
}

func ServerDetailEmbed(server Server) *discordgo.MessageEmbed {
	serverURL := panelBaseURL + "/server/" + url.PathEscape(server.Attributes.Identifier)
	status := Status(pterodactyl.GetServerStatus(server.Attributes.Identifier))

	statusEmoji := "⚪"
	switch status {
	case StatusRunning:
		statusEmoji = "🟢"
	case StatusOffline:
		statusEmoji = "🔴"
	case StatusStarting:
		statusEmoji = "🟡"
	case StatusStopping:
		statusEmoji = "🟠"
	}

	return &discordgo.MessageEmbed{
		Title:       "📦 " + server.Attributes.Name,
		URL:         serverURL,
		Description: "サーバー詳細",
		Color:       0x5865f2,
		Fields: []*discordgo.MessageEmbedField{
			{
				Name:   "識別子",
				Value:  fmt.Sprintf("`%s`", server.Attributes.Identifier),
				Inline: true,
			},
			{
				Name:   "状態",
				Value:  fmt.Sprintf("%s %s", statusEmoji, status.ToJapanese()),
				Inline: true,
			},
			{
				Name:   "パネル",
				Value:  fmt.Sprintf("[Web パネルを開く](%s)", serverURL),
				Inline: false,
			},
		},
		Footer: &discordgo.MessageEmbedFooter{
			Text: strings.TrimPrefix(serverURL, panelBaseURL),
		},
	}
}

func ListUsers(userres *PteroResponse[User]) string {
	if userres == nil {
		return "ユーザーの情報を取得できませんでした"
	}
	result := "ユーザー情報:\n"
	for _, user := range userres.Data {
		result += "ID: " + string(rune(user.Attributes.ID)) + "\n"
	}
	return result
}

func GetAccessibleServers(m *discordgo.Member) []Server {
	if m == nil || m.User == nil {
		return []Server{}
	}

	serverHasAccess := make([]Server, 0)
	servers := DecodeServerResponse(pterodactyl.GetServers(&discordgo.Session{})).Data
	if m.User.ID == storage.Envs.OWNER_ID {
		if servers == nil {
			return []Server{}
		}
		serverHasAccess = servers
		return serverHasAccess
	}
	var serverIDs []string
	for _, role := range m.Roles {
		serverIDs = append(serverIDs, storage.ConfigMgr.GetServerID(role)...)
		log.Print("Role ID: ", role, " Server IDs: ", serverIDs)
	}

	for _, server := range servers {
		for _, id := range serverIDs {
			attributes := server.Attributes.Identifier
			if attributes == id {
				log.Printf("User has access to server: %s (ID: %s)", server.Attributes.Name, server.Attributes.Identifier)
				serverHasAccess = append(serverHasAccess, server)
			}

		}
	}
	return serverHasAccess
}
