package handlers

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"github.com/qingconglaixueit/wechatbot/services"
)

var userMessageHandler = NewUserMessageHandler()

// UserMessageHandler 私聊消息处理
type UserMessageHandler struct {
	userCache services.UserCacheInterface
	msgCache  services.MsgCacheInterface
}

func UserMessageContextHandler() func(ctx *openwechat.MessageContext) {
	return func(ctx *openwechat.MessageContext) {
		msg := ctx.Message
		err := userMessageHandler.handle(msg)
		if err != nil {
			logger.Warning(fmt.Sprintf("handle user message error: %s", err))
		}
	}
}

// NewUserMessageHandler 创建私聊处理器
func NewUserMessageHandler() MessageHandlerInterface {
	return &UserMessageHandler{
		userCache: services.GetUserCache(),
		msgCache:  services.GetMsgCache(),
	}
}

// handle 处理消息
func (h *UserMessageHandler) handle(msg *openwechat.Message) error {
	if !msg.IsText() {
		return nil
	}

	content := msg.Content
	msgId := msg.MsgId
	sender, _ := msg.Sender()
	openId := sender.ID()

	logger.Info(fmt.Sprintf("Received User %v Text Msg : %v", sender.NickName, content))

	if h.msgCache.IfProcessed(msgId) {
		fmt.Println("msgId", msgId, "processed")
		return nil
	}
	h.msgCache.TagProcessed(msgId)
	qParsed := parseContent(content)
	if len(qParsed) == 0 {
		fmt.Println("msgId", msgId, "message.text is empty")
		if _, err := msg.ReplyText("🤖️：你想知道什么呢~"); err != nil {
			fmt.Printf("send err%v\n", err)
		}
		return nil
	}

	if qParsed == "/clear" || qParsed == "清除" {
		h.userCache.Clear(openId)
		_, _ = msg.ReplyText("🤖️：AI机器人已清除记忆")
		return nil
	}

	prompt := h.userCache.Get(openId)
	prompt = fmt.Sprintf("%s\nQ:%s\nA:", prompt, qParsed)
	completions, err := services.Completions(prompt)
	ok := true
	if err != nil {
		_, _ = msg.ReplyText(fmt.Sprintf("🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err))
		return nil
	}
	if len(completions) == 0 {
		ok = false
	}
	if ok {
		h.userCache.Set(openId, qParsed, completions)
		_, err = msg.ReplyText(completions)
		if err != nil {
			_, _ = msg.ReplyText(fmt.Sprintf("🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err))
			return nil
		}
	}
	return nil
}
