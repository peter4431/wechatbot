package bootstrap

import (
	"fmt"
	"github.com/eatmoreapple/openwechat"
	"github.com/qingconglaixueit/wechatbot/config"
	"github.com/qingconglaixueit/wechatbot/handlers"
	"github.com/qingconglaixueit/wechatbot/pkg/logger"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	idBotMap = sync.Map{} // string:*Bot
)

func Run() {
	//bot := openwechat.DefaultBot()
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式，上面登录不上的可以尝试切换这种模式

	// 注册消息处理函数
	handler, err := handlers.NewHandler()
	if err != nil {
		logger.Danger("register error: %v", err)
		return
	}
	bot.MessageHandler = handler

	// 注册登陆二维码回调
	bot.UUIDCallback = handlers.QrCodeCallBack

	// 创建热存储容器对象
	reloadStorage := openwechat.NewFileHotReloadStorage("data/storage.json")

	// 执行热登录
	err = bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption())
	if err != nil {
		logger.Warning(fmt.Sprintf("login error: %v ", err))
		return
	}
	// 阻塞主goroutine, 直到发生异常或者用户主动退出
	bot.Block()
}

func MultiRun() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")

	// 检查 id 并查看登录状态，根据如果没有登录，则可以重新登录
	r.GET("/info/:id", func(c *gin.Context) {
		key := c.Query("key")
		conf := config.LoadConfig()
		if key != conf.ManageKey {
			c.String(http.StatusForbidden, "需要管理密码")
			return
		}
		id := c.Param("id")
		if id == "" {
			c.String(http.StatusNotFound, "需要用户 ID")
			return
		}
		bot := getBot(id)
		isLogin := bot.Alive()
		loginUrl := ""
		if !isLogin {
			loginUrl = botTryLogin(bot, id)
		}

		c.HTML(http.StatusOK, "info.html", gin.H{
			"id":       id,
			"isLogin":  isLogin,
			"loginUrl": loginUrl,
		})
	})

	if err := r.Run(":5000"); err != nil {
		fmt.Printf("gin err:%v", err)
	}
}

func getBot(id string) *openwechat.Bot {
	var ret *openwechat.Bot
	if data, ok := idBotMap.Load(id); ok {
		ret, _ = data.(*openwechat.Bot)
	}

	if ret == nil {
		ret = newBot()
		idBotMap.Store(id, ret)
	}
	return ret
}

func newBot() *openwechat.Bot {
	bot := openwechat.DefaultBot(openwechat.Desktop) // 桌面模式，上面登录不上的可以尝试切换这种模式
	return bot
}

func botTryLogin(bot *openwechat.Bot, id string) string {
	if bot.UUID() != "" {
		return openwechat.GetQrcodeUrl(bot.UUID())
	}

	var url string
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		// 注册消息处理函数
		handler, err := handlers.NewHandler()
		if err != nil {
			logger.Danger("register error: %v", err)
		}
		bot.MessageHandler = handler

		// 注册登陆二维码回调
		bot.UUIDCallback = func(uuid string) {
			handlers.QrCodeCallBack(uuid)
			url = openwechat.GetQrcodeUrl(uuid)
			wg.Done()
		}

		// 创建热存储容器对象
		reloadStorage := openwechat.NewFileHotReloadStorage(fmt.Sprintf("data/%s.json", id))

		// 执行热登录
		err = bot.HotLogin(reloadStorage, openwechat.NewRetryLoginOption())
		if err != nil {
			logger.Warning(fmt.Sprintf("login error: %v ", err))
			return
		}
		// 阻塞主goroutine, 直到发生异常或者用户主动退出
		bot.Block()

		// 退出后需要重新生成 bot
		idBotMap.Delete(id)
	}()
	wg.Wait()
	return url
}
