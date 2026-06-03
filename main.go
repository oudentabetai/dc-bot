package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/oudentabetai/pterodactyl-go/discord"
	"github.com/oudentabetai/pterodactyl-go/storage"
)

var (
	GuildID string
	dgs     *discordgo.Session
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

func deleteAllGlobalCommands(s *discordgo.Session, appID string) {
	_, err := s.ApplicationCommandBulkOverwrite(appID, "", []*discordgo.ApplicationCommand{})

	if err != nil {
		log.Printf("グローバルコマンドの削除に失敗しました: %v", err)
		return
	}
	log.Println("すべてのグローバルコマンドを削除しました。")
}
