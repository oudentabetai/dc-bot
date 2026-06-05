package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/oudentabetai/dc-bot/discord"
	"github.com/oudentabetai/dc-bot/storage"
)

var (
	GuildID string
	dgs     *discordgo.Session
	version = "dev"
)

func main() {
	sessionManager := &discord.DiscordSessionManager{}
	if err := storage.ConfigMgr.Load(); err != nil {
		log.Fatalf("設定ファイルの読み込みに失敗: %v", err)
	}
	dgs = sessionManager.InitializeSession(storage.Envs.DISCORD_BOT_TOKEN)
	dgs.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentsGuilds | discordgo.IntentsGuildMembers | discordgo.IntentsAll | discordgo.PermissionSendMessages
	if err := dgs.Open(); err != nil {
		log.Fatalf("Discordセッションのオープンに失敗: %v", err)
	}
	dgs.AddHandler(discord.OnMessageCreate)
	dgs.AddHandler(discord.OnInteractionCreate)
	defer dgs.Close()
	sendStartupVersionLog(dgs)
	log.Println("ボットが起動しました。Ctrl+Cで終了します。")

	//deleteAllGlobalCommands(dgs, os.Getenv("APPLICATION_ID"))
	discord.SyncCommands(dgs, "", storage.Envs.APPLICATION_ID)
	waitForExitSignal()
}

func waitForExitSignal() {
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc
}

func sendStartupVersionLog(s *discordgo.Session) {
	if storage.Envs.LOG_CHANNEL_ID == "" {
		return
	}

	msg := "起動しました。リリースバージョン: " + version
	if _, err := s.ChannelMessageSend(storage.Envs.LOG_CHANNEL_ID, msg); err != nil {
		log.Printf("起動ログの送信に失敗: %v", err)
	}
}

func deleteAllGlobalCommands(s *discordgo.Session, appID string) {
	_, err := s.ApplicationCommandBulkOverwrite(appID, "", []*discordgo.ApplicationCommand{})

	if err != nil {
		log.Printf("グローバルコマンドの削除に失敗しました: %v", err)
		return
	}
	log.Println("すべてのグローバルコマンドを削除しました。")
}
