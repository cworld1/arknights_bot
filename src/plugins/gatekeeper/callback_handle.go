package gatekeeper

import (
	bot "arknights_bot/config"
	"arknights_bot/utils"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strconv"
	"strings"
)

func CallBackData(callBack tgbotapi.Update) error {
	callbackQuery := callBack.CallbackQuery
	data := callBack.CallbackData()
	d := strings.Split(data, ",")

	if len(d) < 4 {
		return nil
	}

	userId, _ := strconv.ParseInt(d[1], 10, 64)
	chatId := callbackQuery.Message.Chat.ID
	var joinMessageId int

	if d[2] == "PASS" || d[2] == "BAN" {
		joinMessageId, _ = strconv.Atoi(d[3])

		if !utils.IsAdmin(chatId, callbackQuery.From.ID) {
			answer := tgbotapi.NewCallbackWithAlert(callbackQuery.ID, "无使用权限！")
			bot.Arknights.Send(answer)
			return nil
		}

		if d[2] == "PASS" {
			pass(chatId, userId, callbackQuery, true)
		}

		if d[2] == "BAN" {
			chatMember := tgbotapi.ChatMemberConfig{ChatID: chatId, UserID: userId}
			ban(chatId, userId, callbackQuery, chatMember, joinMessageId)
		}

		return nil
	}

	joinMessageId, _ = strconv.Atoi(d[4])

	if userId != callbackQuery.From.ID {
		answer := tgbotapi.NewCallbackWithAlert(callbackQuery.ID, "这不是你的验证！")
		bot.Arknights.Send(answer)
		return nil
	}

	if d[2] != d[3] {
		answer := tgbotapi.NewCallbackWithAlert(callbackQuery.ID, "验证未通过，请一分钟后再试！")
		bot.Arknights.Send(answer)
		chatMember := tgbotapi.ChatMemberConfig{ChatID: chatId, UserID: userId}
		ban(chatId, userId, callbackQuery, chatMember, joinMessageId)
		go unban(chatMember)
		return nil
	}

	answer := tgbotapi.NewCallbackWithAlert(callbackQuery.ID, "验证通过！")
	bot.Arknights.Send(answer)
	pass(chatId, userId, callbackQuery, false)

	return nil
}

func pass(chatId int64, userId int64, callbackQuery *tgbotapi.CallbackQuery, adminPass bool) string {
	bot.Arknights.Send(tgbotapi.RestrictChatMemberConfig{
		Permissions: &tgbotapi.ChatPermissions{
			CanSendMessages:       true,
			CanSendMediaMessages:  true,
			CanSendPolls:          true,
			CanSendOtherMessages:  true,
			CanAddWebPagePreviews: true,
			CanInviteUsers:        true,
			CanChangeInfo:         true,
			CanPinMessages:        true,
		},
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatId,
			UserID: userId,
		},
	})
	val := fmt.Sprintf("verify%d%d", chatId, userId)
	utils.RedisDelSetItem("verify", val)
	delMsg := tgbotapi.NewDeleteMessage(chatId, callbackQuery.Message.MessageID)
	bot.Arknights.Send(delMsg)

	if !adminPass {
		// 新人发送box提醒
		sendMessage := tgbotapi.NewMessage(chatId, fmt.Sprintf("欢迎[%s](tg://user?id=%d)，请向群内发送自己的干员列表截图（或其他截图证明您是真正的玩家），否则可能会被移出群聊。\n建议阅读群公约：[点击阅读](https://t.me/%s/14713)", utils.EscapesMarkdownV2(utils.GetFullName(callbackQuery.From)), callbackQuery.From.ID, callbackQuery.Message.Chat.UserName))
		sendMessage.ParseMode = tgbotapi.ModeMarkdownV2
		bot.Arknights.Send(sendMessage)
	}
	return val
}

func ban(chatId int64, userId int64, callbackQuery *tgbotapi.CallbackQuery, chatMember tgbotapi.ChatMemberConfig, joinMessageId int) {
	banChatMemberConfig := tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: chatMember,
		RevokeMessages:   true,
	}
	bot.Arknights.Send(banChatMemberConfig)
	delMsg := tgbotapi.NewDeleteMessage(chatId, callbackQuery.Message.MessageID)
	bot.Arknights.Send(delMsg)
	val := fmt.Sprintf("verify%d%d", chatId, userId)
	utils.RedisDelSetItem("verify", val)
	delJoinMessage := tgbotapi.NewDeleteMessage(chatId, joinMessageId)
	bot.Arknights.Send(delJoinMessage)
}
