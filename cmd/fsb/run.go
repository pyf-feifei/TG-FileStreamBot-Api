package main

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/bot"
	"EverythingSuckz/fsb/internal/cache"
	"EverythingSuckz/fsb/internal/routes"
	"EverythingSuckz/fsb/internal/types"
	"EverythingSuckz/fsb/internal/utils"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

var runCmd = &cobra.Command{
	Use:                "run",
	Short:              "Run the bot with the given configuration.",
	DisableSuggestions: false,
	Run:                runApp,
}

var startTime time.Time = time.Now()

func runApp(cmd *cobra.Command, args []string) {
	utils.InitLogger(config.ValueOf.Dev)
	log := utils.Logger
	mainLogger := log.Named("Main")
	mainLogger.Info("Starting server")
	config.Load(log, cmd)
	
	// 设置全局 HTTP 客户端的代理
	if config.ValueOf.Dev {
		mainLogger.Info("开发模式已启用，正在配置代理设置")
		// 为默认的 HTTP 客户端配置代理
		utils.SetupProxy(http.DefaultClient)
	}
	
	router := getRouter(log)

	mainBot, err := bot.StartClient(log)
	if err != nil {
		mainLogger.Error("Failed to start main bot", zap.Error(err))
		mainLogger.Warn("⚠️  Telegram客户端启动失败，上传功能将不可用")
		mainLogger.Warn("⚠️  但HTTP服务器将继续启动，您可以测试其他API功能")
	} else {
		cache.InitCache(log)
		workers, err := bot.StartWorkers(log)
		if err != nil {
			mainLogger.Error("Failed to start workers", zap.Error(err))
			mainLogger.Warn("⚠️  Worker启动失败，将只使用主bot")
		} else {
			workers.AddDefaultClient(mainBot, mainBot.Self)
		}

		// 初始化上传worker管理器
		bot.InitUploadWorkerManager(log, config.ValueOf.APICooldownSeconds)

		bot.StartUserBot(log)
		mainLogger.Info("✅ Telegram客户端已连接")
	}

	mainLogger.Info("Server started", zap.Int("port", config.ValueOf.Port))
	mainLogger.Info("File Stream Bot", zap.String("version", versionString))
	mainLogger.Sugar().Infof("Server is running at %s", config.ValueOf.Host)
	err = router.Run(fmt.Sprintf(":%d", config.ValueOf.Port))
	if err != nil {
		mainLogger.Sugar().Fatalln(err)
	}
}

func getRouter(log *zap.Logger) *gin.Engine {
	if config.ValueOf.Dev {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()
	router.Use(gin.ErrorLogger())
	router.GET("/", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, types.RootResponse{
			Message: "Server is running.",
			Ok:      true,
			Uptime:  utils.TimeFormat(uint64(time.Since(startTime).Seconds())),
			Version: versionString,
		})
	})
	routes.Load(log, router)
	return router
}
