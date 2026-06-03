package pterodactyl

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/oudentabetai/pterodactyl-go/storage"
)

var BASE_URL = "https://web.ofton.dev/api/"

type ResourceResponse struct {
	Attributes struct {
		CurrentState string `json:"current_state"`
	} `json:"attributes"`
}

type userListResponse struct {
	Data []struct {
		Attributes struct {
			ID int `json:"id"`
		} `json:"attributes"`
	} `json:"data"`
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

func Fetch(URL string, apiKey string) *http.Response {
	log.Print("Fetching :" + URL)
	req, err := http.NewRequest(http.MethodGet, URL, nil)

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/vnd.pterodactyl.v1+json")

	if err != nil {
		log.Printf("Request error: %v", err)
		return nil
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Do error: %v", err)
	}

	return resp
}

func Post(URL string, apiKey string, body string) *http.Response {
	log.Print("Posting :" + URL)
	req, err := http.NewRequest(http.MethodPost, URL, strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Accept", "application/vnd.pterodactyl.v1+json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Printf("Request creation error: %v", err)
		return nil
	}
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Do error: %v", err)
		return nil
	}
	defer resp.Body.Close()
	return resp
}

func GetUser() string {
	if storage.Envs.PANEL_API_TOKEN == "" {
		log.Print("APPLICATION_API_KEY is empty")
		return "APIキーが未設定です"
	}

	resp := Fetch(BASE_URL+"application/users", storage.Envs.PANEL_API_TOKEN)
	if resp == nil {
		return "API呼び出しに失敗しました"
	}
	defer resp.Body.Close()

	var userres userListResponse
	if err := json.NewDecoder(resp.Body).Decode(&userres); err != nil {
		log.Printf("Decode error: %v", err)
		return "ユーザーの情報を取得できませんでした"
	}

	result := "ユーザー情報:\n"
	for _, user := range userres.Data {
		result += fmt.Sprintf("ID: %d\n", user.Attributes.ID)
	}
	return result
}

func GetServers(s *discordgo.Session) *http.Response {
	if storage.Envs.PANEL_API_TOKEN == "" {
		log.Print("API keys are empty")
		return nil
	}

	resp := Fetch(BASE_URL+"application/servers", storage.Envs.PANEL_API_TOKEN)
	if resp == nil {
		log.Print("Failed to fetch server list")
		return nil
	}
	return resp

}

func PowerServer(serverIdentifier string, signal string) string {
	log.Printf("PowerServer called with identifier: %s, signal: %s", serverIdentifier, signal)
	if storage.Envs.PANEL_CLIENT_TOKEN == "" {
		log.Print("API keys are empty")
		return "APIキーが未設定です"
	}
	resp := Post(BASE_URL+"client/servers/"+serverIdentifier+"/power", storage.Envs.PANEL_CLIENT_TOKEN, `{"signal":"`+signal+`"}`)

	if resp.StatusCode == http.StatusNoContent {
		return "サーバーに`" + signal + "`シグナルを送信しました"
	} else {
		log.Printf("Unexpected status code: %d", resp.StatusCode)
		return "サーバーへのシグナル送信に失敗しました"
	}
}

func GetServerStatus(serverIdentifier string) Status {
	if storage.Envs.PANEL_CLIENT_TOKEN == "" {
		log.Print("API keys are empty")
		return ""
	}
	resp := Fetch(BASE_URL+"client/servers/"+serverIdentifier+"/resources", storage.Envs.PANEL_CLIENT_TOKEN)
	if resp == nil {
		log.Print("Failed to fetch server resources")
		return ""
	}
	defer resp.Body.Close()
	var resourceRes ResourceResponse
	if err := json.NewDecoder(resp.Body).Decode(&resourceRes); err != nil {
		log.Printf("Decode error: %v", err)
		return ""
	}
	return Status(resourceRes.Attributes.CurrentState)
}
