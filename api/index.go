package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil { return }

	var update tgbotapi.Update
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		return
	}

	if update.Message != nil {
		userID := update.Message.From.ID
		chatID := update.Message.Chat.ID
		
		// فحص الاشتراك في القناة
		member, _ := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
			ChatConfigWithUser: tgbotapi.ChatConfigWithUser{SuperGroupUsername: "@boxtoolls", UserID: userID},
		})

		if member.Status == "left" || member.Status == "" {
			msg := tgbotapi.NewMessage(chatID, "⚠️ اشترك أولاً في @boxtoolls ثم أرسل /start")
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(chatID, "✅ تم التحقق! اضغط للدخول:")
			// بناء الزر يدوياً لتجنب خطأ undefined
			button := tgbotapi.InlineKeyboardButton{
				Text: "🚀 فتح التطبيق",
				WebApp: &tgbotapi.WebAppInfo{URL: "https://" + r.Host + "/index.html"},
			}
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(button))
			bot.Send(msg)
		}
	}
	w.WriteHeader(http.StatusOK)
}
