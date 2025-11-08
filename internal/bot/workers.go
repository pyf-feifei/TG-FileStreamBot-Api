package bot

import (
	"EverythingSuckz/fsb/config"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/celestix/gotgproto"
	"github.com/celestix/gotgproto/sessionMaker"
	"github.com/glebarez/sqlite"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type Worker struct {
	ID     int
	Client *gotgproto.Client
	Self   *tg.User
	log    *zap.Logger
}

func (w *Worker) String() string {
	return fmt.Sprintf("{Worker (%d|@%s)}", w.ID, w.Self.Username)
}

type BotWorkers struct {
	Bots     []*Worker
	starting int
	index    int
	mut      sync.Mutex
	log      *zap.Logger
}

var Workers *BotWorkers = &BotWorkers{
	log:  nil,
	Bots: make([]*Worker, 0),
}

// 上传专用worker管理器
type UploadWorkerManager struct {
	workers       []*Worker
	lastUse       map[int]time.Time // workerID -> 最后使用时间
	mutex         sync.Mutex
	cooldown      time.Duration       // API调用冷却时间
	logger        *zap.Logger
	currentIndex  int
}

var uploadManager *UploadWorkerManager

// 初始化上传worker管理器
func InitUploadWorkerManager(log *zap.Logger, cooldownSeconds int) {
	uploadManager = &UploadWorkerManager{
		workers:   Workers.Bots,
		lastUse:   make(map[int]time.Time),
		mutex:     sync.Mutex{},
		cooldown:   time.Duration(cooldownSeconds) * time.Second,
		logger:    log.Named("UploadWorkerManager"),
	}
}

// 获取下一个可用的上传worker
func GetNextUploadWorker() *Worker {
	if uploadManager == nil {
		return GetNextWorker() // 回退到普通选择
	}

	uploadManager.mutex.Lock()
	defer uploadManager.mutex.Unlock()

	now := time.Now()

	// 查找不在冷却期的worker
	for i := 0; i < len(uploadManager.workers); i++ {
		workerIndex := (uploadManager.currentIndex + i) % len(uploadManager.workers)
		worker := uploadManager.workers[workerIndex]

		// 检查worker是否可用
		if lastUse, exists := uploadManager.lastUse[worker.ID]; !exists ||
			now.Sub(lastUse) > uploadManager.cooldown {
			uploadManager.currentIndex = workerIndex
			uploadManager.lastUse[worker.ID] = now
			uploadManager.logger.Debug("选择上传worker",
				zap.Int("workerID", worker.ID),
				zap.String("username", worker.Self.Username),
				zap.Duration("cooldownWait", now.Sub(lastUse)))
			return worker
		}
	}

	// 所有worker都在冷却，选择等待时间最短的
	shortestWait := time.Hour
	var selectedWorker *Worker

	for _, worker := range uploadManager.workers {
		if lastUse, exists := uploadManager.lastUse[worker.ID]; exists {
			waitTime := uploadManager.cooldown - now.Sub(lastUse)
			if waitTime < shortestWait {
				shortestWait = waitTime
				selectedWorker = worker
			}
		} else {
			shortestWait = 0
			selectedWorker = worker
			break
		}
	}

	if selectedWorker != nil {
		uploadManager.currentIndex = (uploadManager.currentIndex + 1) % len(uploadManager.workers)
		uploadManager.lastUse[selectedWorker.ID] = now

		if shortestWait > 0 {
			uploadManager.logger.Warn("所有worker都在冷却期，选择等待时间最短的",
				zap.Int("workerID", selectedWorker.ID),
				zap.String("username", selectedWorker.Self.Username),
				zap.Duration("waitTime", shortestWait))
		} else {
			uploadManager.logger.Debug("选择可用上传worker",
				zap.Int("workerID", selectedWorker.ID),
				zap.String("username", selectedWorker.Self.Username))
		}

		return selectedWorker
	}

	return GetNextWorker() // 如果没有worker可用，回退到普通选择
}

// 获取worker统计信息
func GetUploadWorkerStats() map[string]interface{} {
	if uploadManager == nil {
		return map[string]interface{}{
			"totalWorkers": len(Workers.Bots),
			"availableWorkers": len(Workers.Bots),
			"uploadManagerEnabled": false,
		}
	}

	uploadManager.mutex.Lock()
	defer uploadManager.mutex.Unlock()

	now := time.Now()
	availableCount := 0
	cooldownCount := 0

	for _, worker := range uploadManager.workers {
		if lastUse, exists := uploadManager.lastUse[worker.ID]; exists {
			if now.Sub(lastUse) > uploadManager.cooldown {
				availableCount++
			} else {
				cooldownCount++
			}
		} else {
			availableCount++
		}
	}

	return map[string]interface{}{
		"totalWorkers":      len(uploadManager.workers),
		"availableWorkers":   availableCount,
		"cooldownWorkers":    cooldownCount,
		"cooldownDuration":   uploadManager.cooldown.Seconds(),
		"uploadManagerEnabled": true,
	}
}

func (w *BotWorkers) Init(log *zap.Logger) {
	w.log = log.Named("Workers")
}

func (w *BotWorkers) AddDefaultClient(client *gotgproto.Client, self *tg.User) {
	if w.Bots == nil {
		w.Bots = make([]*Worker, 0)
	}
	w.incStarting()
	w.Bots = append(w.Bots, &Worker{
		Client: client,
		ID:     w.starting,
		Self:   self,
		log:    w.log,
	})
	w.log.Sugar().Info("Default bot loaded")
}

func (w *BotWorkers) incStarting() {
	w.mut.Lock()
	defer w.mut.Unlock()
	w.starting++
}

func (w *BotWorkers) Add(token string) (err error) {
	w.incStarting()
	var botID int = w.starting
	client, err := startWorker(w.log, token, botID)
	if err != nil {
		return err
	}
	w.log.Sugar().Infof("Bot @%s loaded with ID %d", client.Self.Username, botID)
	w.Bots = append(w.Bots, &Worker{
		Client: client,
		ID:     botID,
		Self:   client.Self,
		log:    w.log,
	})
	return nil
}

func GetNextWorker() *Worker {
	Workers.mut.Lock()
	defer Workers.mut.Unlock()
	index := (Workers.index + 1) % len(Workers.Bots)
	Workers.index = index
	worker := Workers.Bots[index]
	Workers.log.Sugar().Debugf("Using worker %d", worker.ID)
	return worker
}

func StartWorkers(log *zap.Logger) (*BotWorkers, error) {
	Workers.Init(log)

	if len(config.ValueOf.MultiTokens) == 0 {
		Workers.log.Sugar().Info("No worker bot tokens provided, skipping worker initialization")
		return Workers, nil
	}
	Workers.log.Sugar().Info("Starting")
	if config.ValueOf.UseSessionFile {
		Workers.log.Sugar().Info("Using session file for workers")
		newpath := filepath.Join(".", "sessions")
		if err := os.MkdirAll(newpath, os.ModePerm); err != nil {
			Workers.log.Error("Failed to create sessions directory", zap.Error(err))
			return nil, err
		}
	}

	var wg sync.WaitGroup
	var successfulStarts int32
	totalBots := len(config.ValueOf.MultiTokens)

	for i := 0; i < totalBots; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			done := make(chan error, 1)
			go func() {
				err := Workers.Add(config.ValueOf.MultiTokens[i])
				done <- err
			}()

			select {
			case err := <-done:
				if err != nil {
					Workers.log.Error("Failed to start worker", zap.Int("index", i), zap.Error(err))
				} else {
					atomic.AddInt32(&successfulStarts, 1)
				}
			case <-ctx.Done():
				Workers.log.Error("Timed out starting worker", zap.Int("index", i))
			}
		}(i)
	}

	wg.Wait() // Wait for all goroutines to finish
	Workers.log.Sugar().Infof("Successfully started %d/%d bots", successfulStarts, totalBots)
	return Workers, nil
}

func startWorker(l *zap.Logger, botToken string, index int) (*gotgproto.Client, error) {
	log := l.Named("Worker").Sugar()
	log.Infof("Starting worker with index - %d", index)
	var sessionType sessionMaker.SessionConstructor
	if config.ValueOf.UseSessionFile {
		sessionType = sessionMaker.SqlSession(sqlite.Open(fmt.Sprintf("sessions/worker-%d.session", index)))
	} else {
		sessionType = sessionMaker.SimpleSession()
	}
	client, err := gotgproto.NewClient(
		int(config.ValueOf.ApiID),
		config.ValueOf.ApiHash,
		gotgproto.ClientTypeBot(botToken),
		&gotgproto.ClientOpts{
			Session:          sessionType,
			DisableCopyright: true,
			Middlewares:      GetFloodMiddleware(log.Desugar()),
		},
	)
	if err != nil {
		return nil, err
	}
	return client, nil
}
