package main

import (
	"encoding/json"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		log.Fatal("Failed to find bot token in TELEGRAM_BOT_TOKEN environment variable")
	}

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = true

	log.Printf("Authorized as %s", bot.Self.UserName)

	hook := os.Getenv("APP_WEBHOOK")
	_, err = bot.SetWebhook(tgbotapi.NewWebhook(hook + token))
	if err != nil {
		log.Fatal(err)
	}

	// Handler for the webhook
	http.HandleFunc("/hook/"+token, func(w http.ResponseWriter, r *http.Request) {
		var update tgbotapi.Update
		if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
			log.Println("Failed to decode update:", err)
			return
		}

		if update.Message.IsCommand() {
			switch update.Message.Command() {
			case "check":
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please, send image:")
				msg.ReplyMarkup = tgbotapi.ForceReply{ForceReply: true}

				if _, err := bot.Send(msg); err != nil {
					log.Println("Error sending:", err)
				}
			}
		}

		if update.Message.Photo != nil {
			photoProcess(update, bot, token)
		}

		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		}
	})

	// Start the HTTP server
	go http.ListenAndServe("0.0.0.0:8181", nil)

	select {}
}

func photoProcess(update tgbotapi.Update, bot *tgbotapi.BotAPI, token string) {
	fileID := (*update.Message.Photo)[len(*update.Message.Photo)-1].FileID

	fileConfig := tgbotapi.FileConfig{FileID: fileID}
	file, err := bot.GetFile(fileConfig)
	if err != nil {
		log.Println("Failed to get file:", err)
		return
	}

	fileURL := "https://api.telegram.org/file/bot" + token + "/" + file.FilePath
	log.Println("File URL:", fileURL)

	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "File URL:"+fileURL)
	if _, err := bot.Send(msg); err != nil {
		log.Println("Error sending:", err)
		return
	}
}
