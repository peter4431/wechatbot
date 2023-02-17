package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"github.com/qingconglaixueit/wechatbot/services"
)

var groupMessageHandler = NewGroupMessageHandler()

// GroupMessageHandler 群消息处理
type GroupMessageHandler struct {
	userCache services.UserCacheInterface
	msgCache  services.MsgCacheInterface
}

func GroupMessageContextHandler() func(ctx *openwechat.MessageContext) {
	return func(ctx *openwechat.MessageContext) {
		msg := ctx.Message
		// 处理用户消息
		err := groupMessageHandler.handle(msg)
		if err != nil {
			logger.Warning(fmt.Sprintf("handle group message error: %s", err))
		}
	}
}

// NewGroupMessageHandler 创建群消息处理器
func NewGroupMessageHandler() MessageHandlerInterface {
	return &GroupMessageHandler{
		userCache: services.GetUserCache(),
		msgCache:  services.GetMsgCache(),
	}
}

func toJson(data interface{}) string {
	ret, _ := json.Marshal(data)
	if ret != nil {
		return string(ret)
	}
	return ""
}

// handle 处理消息
func (g *GroupMessageHandler) handle(msg *openwechat.Message) error {
	if !msg.IsText() {
		return nil
	}
	if !msg.IsAt() {
		return nil
	}

	sender, err := msg.Sender()
	if err != nil {
		return err
	}
	group := &openwechat.Group{User: sender}
	//groupSender, err := msg.SenderInGroup()

	logger.Info(fmt.Sprintf("Received Group %v Text Msg : %v", group.NickName, msg.Content))

	ifMention := judgeIfMentionMe(msg)
	if !ifMention {
		return nil
	}
	content := msg.Content
	msgId := msg.MsgId
	openId := sender.ID()

	if g.msgCache.IfProcessed(msgId) {
		fmt.Println("msgId", msgId, "processed")
		return nil
	}
	g.msgCache.TagProcessed(msgId)
	qParsed := parseContent(content)
	if len(qParsed) == 0 {
		_, _ = msg.ReplyText("🤖️：你想知道什么呢~")
		fmt.Println("msgId", msgId, "message.text is empty")
		return nil
	}

	if qParsed == "/clear" || qParsed == "清除" {
		g.userCache.Clear(openId)
		_, _ = msg.ReplyText("🤖️：AI机器人已清除记忆")
		return nil
	}

	prompt := g.userCache.Get(openId)
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
		g.userCache.Set(openId, qParsed, completions)
		_, err = msg.ReplyText(completions)
		if err != nil {
			_, _ = msg.ReplyText(fmt.Sprintf("🤖️：消息机器人摆烂了，请稍后再试～\n错误信息: %v", err))
			return nil
		}
	}
	return nil
}

func judgeIfMentionMe(event *openwechat.Message) bool {
	return true
}
