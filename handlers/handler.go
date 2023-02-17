package handlers

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/patrickmn/go-cache"
	"github.com/qingconglaixueit/wechatbot/config"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"runtime"
	"strings"
	"time"
)

var c = cache.New(config.LoadConfig().SessionTimeout, time.Minute*5)

// MessageHandlerInterface 消息处理接口
type MessageHandlerInterface interface {
	handle(msg *openwechat.Message) error
}

// QrCodeCallBack 登录扫码回调，
func QrCodeCallBack(uuid string) {
	if runtime.GOOS == "windows" {
		// 运行在Windows系统上
		openwechat.PrintlnQrcodeUrl(uuid)
	} else {
		println("访问下面网址扫描二维码登录")
		qrcodeUrl := openwechat.GetQrcodeUrl(uuid)
		println(qrcodeUrl)
	}
}

func NewHandler() (msgFunc func(msg *openwechat.Message), err error) {
	dispatcher := openwechat.NewMessageMatchDispatcher()

	// 处理群消息
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return message.IsSendByGroup()
	}, GroupMessageContextHandler())

	// 好友申请
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return message.IsFriendAdd()
	}, func(ctx *openwechat.MessageContext) {
		msg := ctx.Message
		if config.LoadConfig().AutoPass {
			_, err := msg.Agree("")
			if err != nil {
				logger.Warning(fmt.Sprintf("add friend agree error : %v", err))
				return
			}
		}
	})

	// 私聊
	// 获取用户消息处理器
	dispatcher.RegisterHandler(func(message *openwechat.Message) bool {
		return !(strings.Contains(message.Content, config.LoadConfig().SessionClearToken) || message.IsSendByGroup() || message.IsFriendAdd())
	}, UserMessageContextHandler())
	return dispatcher.AsMessageHandler(), nil
}
