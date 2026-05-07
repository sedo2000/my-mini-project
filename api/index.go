package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type WebAppSignal struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	Action   string `json:"action"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return
	}

	var rawData json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawData); err != nil {
		w.WriteHeader(http.StatusOK)
		return
	}

	var signal WebAppSignal
	if err := json.Unmarshal(rawData, &signal); err == nil && signal.Action == "welcome_trigger" {
		msg := tgbotapi.NewMessage(signal.UserID, fmt.Sprintf("أهلاً بك يا %s! ✨ تم الدخول بنجاح.", signal.UserName))
		bot.Send(msg)
		w.WriteHeader(http.StatusOK)
		return
	}

	var update tgbotapi.Update
	if err := json.Unmarshal(rawData, &update); err == nil {
		handleUpdates(bot, update, r)
	}

	w.WriteHeader(http.StatusOK)
}

func handleUpdates(bot *tgbotapi.BotAPI, update tgbotapi.Update, r *http.Request) {
	const channelUsername = "@boxtoolls"
	webAppURL := "https://" + r.Host + "/indexq.html"

	if update.CallbackQuery != nil && update.CallbackQuery.Data == "verify_sub" {
		userID := update.CallbackQuery.From.ID
		if checkSubscription(bot, channelUsername, userID) {
			bot.Send(tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "✅ تم التحقق! يمكنك الآن الدخول:")
			msg.ReplyMarkup = createMainKeyboard(webAppURL)
			bot.Send(msg)
		} else {
			bot.Request(tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "❌ لم تشترك بعد!"))
		}
		return
	}

	if update.Message != nil {
		userID := update.Message.From.ID
		if checkSubscription(bot, channelUsername, userID) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "اضغط بالأسفل لفتح التطبيق:")
			msg.ReplyMarkup = createMainKeyboard(webAppURL)
			bot.Send(msg)
		} else {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ اشترك في القناة أولاً.")
			btnSub := tgbotapi.NewInlineKeyboardButtonURL("📢 اشترك هنا", "https://t.me/boxtoolls")
			btnVerify := tgbotapi.NewInlineKeyboardButtonData("✅ تحقق", "verify_sub")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
				tgbotapi.NewInlineKeyboardRow(btnSub),
				tgbotapi.NewInlineKeyboardRow(btnVerify),
			)
			bot.Send(msg)
		}
	}
}

func checkSubscription(bot *tgbotapi.BotAPI, channel string, userID int64) bool {
	member, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{SuperGroupUsername: channel, UserID: userID},
	})
	return err == nil && (member.Status == "member" || member.Status == "administrator" || member.Status == "creator")
}

// تعديل الدالة لتجنب أخطاء الـ Build تماماً
func createMainKeyboard(url string) tgbotapi.InlineKeyboardMarkup {
	// تعريف الزر بشكل يدوي مباشر
	button := tgbotapi.InlineKeyboardButton{
		Text: "🔗 دخول الاختبار",
		WebApp: &tgbotapi.WebAppInfo{URL: url},
	}
	return tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(button),
	)
}
