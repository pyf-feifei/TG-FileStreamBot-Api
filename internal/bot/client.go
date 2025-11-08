package bot

import (
	"EverythingSuckz/fsb/config"
	"EverythingSuckz/fsb/internal/commands"
	"EverythingSuckz/fsb/internal/utils"
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"
	"github.com/gotd/td/telegram/dcs"
)

var Bot *gotgproto.Client

func StartClient(log *zap.Logger) (*gotgproto.Client, error) {
	// 验证代理配置
	if config.ValueOf.TelegramProxy != "" {
		if err := utils.ValidateProxyURL(config.ValueOf.TelegramProxy); err != nil {
			log.Error("Invalid Telegram proxy configuration", zap.Error(err))
			return nil, err
		}
		log.Info("Using Telegram proxy", zap.String("proxy", config.ValueOf.TelegramProxy))
	}

	// 创建自定义Resolver（支持代理）
	var resolver dcs.Resolver
	if config.ValueOf.TelegramProxy != "" {
		dialFunc, err := utils.CreateSOCKS5Dialer(config.ValueOf.TelegramProxy)
		if err != nil {
			log.Error("Failed to create SOCKS5 dialer", zap.Error(err))
			return nil, err
		}
		
		// 使用自定义Dialer创建Plain resolver
		resolver = dcs.Plain(dcs.PlainOptions{
			Dial: dcs.DialFunc(dialFunc),
		})
		log.Info("✓ SOCKS5代理已配置", zap.String("proxy", config.ValueOf.TelegramProxy))
	} else {
		resolver = dcs.DefaultResolver()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	resultChan := make(chan struct {
		client *gotgproto.Client
		err    error
	})
	go func(ctx context.Context) {
		client, err := gotgproto.NewClient(
			int(config.ValueOf.ApiID),
			config.ValueOf.ApiHash,
			gotgproto.ClientTypeBot(config.ValueOf.BotToken),
			&gotgproto.ClientOpts{
				Session: sessionMaker.SqlSession(
					sqlite.Open("fsb.session"),
				),
				DisableCopyright: true,
				Resolver:         resolver, // 使用自定义Resolver
			},
		)
		resultChan <- struct {
			client *gotgproto.Client
			err    error
		}{client, err}
	}(ctx)

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultChan:
		if result.err != nil {
			return nil, result.err
		}
		commands.Load(log, result.client.Dispatcher)
		log.Info("Client started", zap.String("username", result.client.Self.Username))
		Bot = result.client
		return result.client, nil
	}
}
