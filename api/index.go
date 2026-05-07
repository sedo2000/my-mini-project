package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// هيكل البيانات المستلمة من التطبيق المصغر
type WebAppSignal struct {
	UserID   int64  `json:"user_id"`
	UserName string `json:"user_name"`
	Action   string `json:"action"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// إعدادات CORS للسماح بالاتصال من المتصفح
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

	// 1. معالجة الإشارات القادمة من الـ Mini App (الترحيب)
	var signal WebAppSignal
	if err := json.Unmarshal(rawData, &signal); err == nil && signal.Action == "welcome_trigger" {
		msg := tgbotapi.NewMessage(signal.UserID, fmt.Sprintf("أهلاً بك يا %s! ✨ تم الدخول بنجاح، بالتوفيق في الاختبار.", signal.UserName))
		bot.Send(msg)
		w.WriteHeader(http.StatusOK)
		return
	}

	// 2. معالجة تحديثات تيليجرام (الرسائل والأزرار)
	var update tgbotapi.Update
	if err := json.Unmarshal(rawData, &update); err == nil {
		handleUpdates(bot, update, r)
	}

	w.WriteHeader(http.StatusOK)
}

func handleUpdates(bot *tgbotapi.BotAPI, update tgbotapi.Update, r *http.Request) {
	const channelUsername = "@boxtoolls" // اسم قناتك
	// رابط التطبيق (تلقائي حسب دومين المشروع)
	webAppURL := "https://" + r.Host + "/indexq.html"

	// إذا ضغط المستخدم على زر "تحقق من الاشتراك"
	if update.CallbackQuery != nil && update.CallbackQuery.Data == "verify_sub" {
		userID := update.CallbackQuery.From.ID
		if checkSubscription(bot, channelUsername, userID) {
			// حذف رسالة التحقق وإرسال زر الدخول
			bot.Send(tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID))
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, "✅ تم التحقق! يمكنك الآن الدخول:")
			msg.ReplyMarkup = createMainKeyboard(webAppURL)
			bot.Send(msg)
		} else {
			// إظهار تنبيه للمستخدم
			callbackConfig := tgbotapi.NewCallbackWithAlert(update.CallbackQuery.ID, "❌ لم تشترك في القناة بعد!")
			bot.Request(callbackConfig)
		}
		return
	}

	// إذا أرسل المستخدم رسالة (مثل /start)
	if update.Message != nil {
		userID := update.Message.From.ID
		if checkSubscription(bot, channelUsername, userID) {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "أهلاً بك مجدداً! اضغط بالأسفل للبدء:")
			msg.ReplyMarkup = createMainKeyboard(webAppURL)
			bot.Send(msg)
		} else {
			// رسالة الاشتراك الإجباري
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "⚠️ عذراً، يجب عليك الاشتراك في قناة البوت أولاً لتتمكن من استخدامه.")
			btnSub := tgbotapi.NewInlineKeyboardButtonURL("📢 اشترك هنا", "https://t.me/boxtoolls")
			btnVerify := tgbotapi.NewInlineKeyboardButtonData("✅ تحقق من الاشتراك", "verify_sub")
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btnSub), tgbotapi.NewInlineKeyboardRow(btnVerify))
			bot.Send(msg)
		}
	}
}

// دالة التحقق من الاشتراك
func checkSubscription(bot *tgbotapi.BotAPI, channel string, userID int64) bool {
	member, err := bot.GetChatMember(tgbotapi.GetChatMemberConfig{
		ChatConfigWithUser: tgbotapi.ChatConfigWithUser{SuperGroupUsername: channel, UserID: userID},
	})
	if err != nil {
		return false
	}
	return member.Status == "member" || member.Status == "administrator" || member.Status == "creator"
}

// دالة إنشاء زر فتح التطبيق المصغر
func createMainKeyboard(url string) tgbotapi.InlineKeyboardMarkup {
	webApp := tgbotapi.WebAppInfo{URL: url}
	btn := tgbotapi.NewInlineKeyboardButtonWebApp("🔗 دخول الاختبار", webApp)
	return tgbotapi.NewInlineKeyboardMarkup(tgbotapi.NewInlineKeyboardRow(btn))
}
