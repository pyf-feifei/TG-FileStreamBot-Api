package config

import (
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var ValueOf = &config{}

type allowedUsers []int64

func (au *allowedUsers) Decode(value string) error {
	if value == "" {
		return nil
	}
	ids := strings.Split(string(value), ",")
	for _, id := range ids {
		idInt, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			return err
		}
		*au = append(*au, idInt)
	}
	return nil
}

type config struct {
	ApiID          int32        `envconfig:"API_ID" required:"true"`
	ApiHash        string       `envconfig:"API_HASH" required:"true"`
	BotToken       string       `envconfig:"BOT_TOKEN" required:"true"`
	LogChannelID   int64        `envconfig:"LOG_CHANNEL" required:"true"`
	Dev            bool         `envconfig:"DEV" default:"false"`
	Port           int          `envconfig:"PORT" default:"8080"`
	Host           string       `envconfig:"HOST" default:""`
	HashLength     int          `envconfig:"HASH_LENGTH" default:"6"`
	UseSessionFile bool         `envconfig:"USE_SESSION_FILE" default:"true"`
	UserSession    string       `envconfig:"USER_SESSION"`
	UsePublicIP    bool         `envconfig:"USE_PUBLIC_IP" default:"false"`
	AllowedUsers   allowedUsers `envconfig:"ALLOWED_USERS"`
	MultiTokens    []string

	// 上传功能配置
	EnableUploadAPI     bool     `envconfig:"ENABLE_UPLOAD_API" default:"false"`
	UploadAuthToken     string   `envconfig:"UPLOAD_AUTH_TOKEN"`
	MaxFileSize        int64    `envconfig:"MAX_FILE_SIZE" default:"2147483648"` // 2GB
	UserQuota          int64    `envconfig:"USER_QUOTA" default:"0"`  // 0 = 不限制配额
	AllowedMimeTypes   string   `envconfig:"ALLOWED_MIME_TYPES" default:"image/jpeg,image/png,image/gif,video/mp4,video/avi,application/pdf,text/plain,application/zip"`
	AllowedExtensions  string   `envconfig:"ALLOWED_EXTENSIONS" default:".jpg,.jpeg,.png,.gif,.mp4,.avi,.pdf,.txt,.zip"`
	UploadsPerMinute   int      `envconfig:"UPLOADS_PER_MINUTE" default:"5"`
	UploadsPerHour     int      `envconfig:"UPLOADS_PER_HOUR" default:"50"`
	ConcurrentUploads   int      `envconfig:"CONCURRENT_UPLOADS_PER_USER" default:"3"`
	APICooldownSeconds int      `envconfig:"API_COOLDOWN_SECONDS" default:"1"`
	EnableProtection   bool     `envconfig:"ENABLE_PROTECTION_MODE" default:"true"`
	EnableDeepScan     bool     `envconfig:"ENABLE_DEEP_SCAN" default:"false"`
	
	// 代理配置
	TelegramProxy      string   `envconfig:"TELEGRAM_PROXY" default:""` // socks5://127.0.0.1:1080
}

var botTokenRegex = regexp.MustCompile(`MULTI\_TOKEN\d+=(.*)`)

func (c *config) loadFromEnvFile(log *zap.Logger) {
	envPath := filepath.Clean("fsb.env")
	log.Sugar().Infof("Trying to load ENV vars from %s", envPath)
	err := godotenv.Load(envPath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Sugar().Errorf("ENV file not found: %s", envPath)
			log.Sugar().Info("Please create fsb.env file")
			log.Sugar().Info("For more info, refer: https://github.com/EverythingSuckz/TG-FileStreamBot/tree/golang#setting-up-things")
			log.Sugar().Info("Please ignore this message if you are hosting it in a service like Heroku or other alternatives.")
		} else {
			log.Fatal("Unknown error while parsing env file.", zap.Error(err))
		}
	}
}

func SetFlagsFromConfig(cmd *cobra.Command) {
	cmd.Flags().Int32("api-id", ValueOf.ApiID, "Telegram API ID")
	cmd.Flags().String("api-hash", ValueOf.ApiHash, "Telegram API Hash")
	cmd.Flags().String("bot-token", ValueOf.BotToken, "Telegram Bot Token")
	cmd.Flags().Int64("log-channel", ValueOf.LogChannelID, "Telegram Log Channel ID")
	cmd.Flags().Bool("dev", ValueOf.Dev, "Enable development mode")
	cmd.Flags().IntP("port", "p", ValueOf.Port, "Server port")
	cmd.Flags().String("host", ValueOf.Host, "Server host that will be included in links")
	cmd.Flags().Int("hash-length", ValueOf.HashLength, "Hash length in links")
	cmd.Flags().Bool("use-session-file", ValueOf.UseSessionFile, "Use session files")
	cmd.Flags().String("user-session", ValueOf.UserSession, "Pyrogram user session")
	cmd.Flags().Bool("use-public-ip", ValueOf.UsePublicIP, "Use public IP instead of local IP")
	cmd.Flags().String("multi-token-txt-file", "", "Multi token txt file (Not implemented)")

	// 上传API相关命令行参数
	cmd.Flags().Bool("enable-upload-api", ValueOf.EnableUploadAPI, "Enable upload API")
	cmd.Flags().String("upload-auth-token", ValueOf.UploadAuthToken, "Upload API authentication token")
	cmd.Flags().Int64("max-file-size", ValueOf.MaxFileSize, "Maximum file size for upload (bytes)")
	cmd.Flags().Int64("user-quota", ValueOf.UserQuota, "User storage quota (bytes)")
	cmd.Flags().String("allowed-mime-types", ValueOf.AllowedMimeTypes, "Allowed MIME types for upload")
	cmd.Flags().String("allowed-extensions", ValueOf.AllowedExtensions, "Allowed file extensions for upload")
	cmd.Flags().Int("uploads-per-minute", ValueOf.UploadsPerMinute, "Uploads allowed per minute per user")
	cmd.Flags().Int("uploads-per-hour", ValueOf.UploadsPerHour, "Uploads allowed per hour per user")
	cmd.Flags().Int("concurrent-uploads", ValueOf.ConcurrentUploads, "Concurrent uploads per user")
	cmd.Flags().Int("api-cooldown-seconds", ValueOf.APICooldownSeconds, "API cooldown seconds")
	cmd.Flags().Bool("enable-protection", ValueOf.EnableProtection, "Enable protection mode")
	cmd.Flags().Bool("enable-deep-scan", ValueOf.EnableDeepScan, "Enable deep file scanning")
}

func (c *config) loadConfigFromArgs(log *zap.Logger, cmd *cobra.Command) {
	apiID, _ := cmd.Flags().GetInt32("api-id")
	if apiID != 0 {
		os.Setenv("API_ID", strconv.Itoa(int(apiID)))
	}
	apiHash, _ := cmd.Flags().GetString("api-hash")
	if apiHash != "" {
		os.Setenv("API_HASH", apiHash)
	}
	botToken, _ := cmd.Flags().GetString("bot-token")
	if botToken != "" {
		os.Setenv("BOT_TOKEN", botToken)
	}
	logChannelID, _ := cmd.Flags().GetString("log-channel")
	if logChannelID != "" {
		os.Setenv("LOG_CHANNEL", logChannelID)
	}
	dev, _ := cmd.Flags().GetBool("dev")
	if dev {
		os.Setenv("DEV", strconv.FormatBool(dev))
	}
	port, _ := cmd.Flags().GetInt("port")
	if port != 0 {
		os.Setenv("PORT", strconv.Itoa(port))
	}
	host, _ := cmd.Flags().GetString("host")
	if host != "" {
		os.Setenv("HOST", host)
	}
	hashLength, _ := cmd.Flags().GetInt("hash-length")
	if hashLength != 0 {
		os.Setenv("HASH_LENGTH", strconv.Itoa(hashLength))
	}
	useSessionFile, _ := cmd.Flags().GetBool("use-session-file")
	if useSessionFile {
		os.Setenv("USE_SESSION_FILE", strconv.FormatBool(useSessionFile))
	}
	userSession, _ := cmd.Flags().GetString("user-session")
	if userSession != "" {
		os.Setenv("USER_SESSION", userSession)
	}
	usePublicIP, _ := cmd.Flags().GetBool("use-public-ip")
	if usePublicIP {
		os.Setenv("USE_PUBLIC_IP", strconv.FormatBool(usePublicIP))
	}
	multiTokens, _ := cmd.Flags().GetString("multi-token-txt-file")
	if multiTokens != "" {
		os.Setenv("MULTI_TOKEN_TXT_FILE", multiTokens)
		// TODO: Add support for importing tokens from a separate file
	}

	// 上传API配置处理
	enableUploadAPI, _ := cmd.Flags().GetBool("enable-upload-api")
	if enableUploadAPI {
		os.Setenv("ENABLE_UPLOAD_API", strconv.FormatBool(enableUploadAPI))
	}
	uploadAuthToken, _ := cmd.Flags().GetString("upload-auth-token")
	if uploadAuthToken != "" {
		os.Setenv("UPLOAD_AUTH_TOKEN", uploadAuthToken)
	}
	maxFileSize, _ := cmd.Flags().GetInt64("max-file-size")
	if maxFileSize != 0 {
		os.Setenv("MAX_FILE_SIZE", strconv.FormatInt(maxFileSize, 10))
	}
	userQuota, _ := cmd.Flags().GetInt64("user-quota")
	if userQuota != 0 {
		os.Setenv("USER_QUOTA", strconv.FormatInt(userQuota, 10))
	}
	allowedMimeTypes, _ := cmd.Flags().GetString("allowed-mime-types")
	if allowedMimeTypes != "" {
		os.Setenv("ALLOWED_MIME_TYPES", allowedMimeTypes)
	}
	allowedExtensions, _ := cmd.Flags().GetString("allowed-extensions")
	if allowedExtensions != "" {
		os.Setenv("ALLOWED_EXTENSIONS", allowedExtensions)
	}
	uploadsPerMinute, _ := cmd.Flags().GetInt("uploads-per-minute")
	if uploadsPerMinute != 0 {
		os.Setenv("UPLOADS_PER_MINUTE", strconv.Itoa(uploadsPerMinute))
	}
	uploadsPerHour, _ := cmd.Flags().GetInt("uploads-per-hour")
	if uploadsPerHour != 0 {
		os.Setenv("UPLOADS_PER_HOUR", strconv.Itoa(uploadsPerHour))
	}
	concurrentUploads, _ := cmd.Flags().GetInt("concurrent-uploads")
	if concurrentUploads != 0 {
		os.Setenv("CONCURRENT_UPLOADS_PER_USER", strconv.Itoa(concurrentUploads))
	}
	apiCooldownSeconds, _ := cmd.Flags().GetInt("api-cooldown-seconds")
	if apiCooldownSeconds != 0 {
		os.Setenv("API_COOLDOWN_SECONDS", strconv.Itoa(apiCooldownSeconds))
	}
	enableProtection, _ := cmd.Flags().GetBool("enable-protection")
	if enableProtection {
		os.Setenv("ENABLE_PROTECTION_MODE", strconv.FormatBool(enableProtection))
	}
	enableDeepScan, _ := cmd.Flags().GetBool("enable-deep-scan")
	if enableDeepScan {
		os.Setenv("ENABLE_DEEP_SCAN", strconv.FormatBool(enableDeepScan))
	}
}

func (c *config) setupEnvVars(log *zap.Logger, cmd *cobra.Command) {
	c.loadFromEnvFile(log)
	c.loadConfigFromArgs(log, cmd)
	err := envconfig.Process("", c)
	if err != nil {
		log.Fatal("Error while parsing env variables", zap.Error(err))
	}
	var ipBlocked bool
	ip, err := getIP(c.UsePublicIP)
	if err != nil {
		log.Error("Error while getting IP", zap.Error(err))
		ipBlocked = true
	}
	if c.Host == "" {
		c.Host = "http://" + ip + ":" + strconv.Itoa(c.Port)
		if c.UsePublicIP {
			if ipBlocked {
				log.Sugar().Warn("Can't get public IP, using local IP")
			} else {
				log.Sugar().Warn("You are using a public IP, please be aware of the security risks while exposing your IP to the internet.")
				log.Sugar().Warn("Use 'HOST' variable to set a domain name")
			}
		}
		log.Sugar().Info("HOST not set, automatically set to " + c.Host)
	}
	val := reflect.ValueOf(c).Elem()
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "MULTI_TOKEN") {
			c.MultiTokens = append(c.MultiTokens, botTokenRegex.FindStringSubmatch(env)[1])
		}
	}
	val.FieldByName("MultiTokens").Set(reflect.ValueOf(c.MultiTokens))
}

func Load(log *zap.Logger, cmd *cobra.Command) {
	log = log.Named("Config")
	defer log.Info("Loaded config")
	ValueOf.setupEnvVars(log, cmd)
	ValueOf.LogChannelID = int64(stripInt(log, int(ValueOf.LogChannelID)))
	if ValueOf.HashLength == 0 {
		log.Sugar().Info("HASH_LENGTH can't be 0, defaulting to 6")
		ValueOf.HashLength = 6
	}
	if ValueOf.HashLength > 32 {
		log.Sugar().Info("HASH_LENGTH can't be more than 32, changing to 32")
		ValueOf.HashLength = 32
	}
	if ValueOf.HashLength < 5 {
		log.Sugar().Info("HASH_LENGTH can't be less than 5, defaulting to 6")
		ValueOf.HashLength = 6
	}
}

func getIP(public bool) (string, error) {
	var ip string
	var err error
	if public {
		ip, err = GetPublicIP()
	} else {
		ip, err = getInternalIP()
	}
	if ip == "" {
		ip = "localhost"
	}
	if err != nil {
		return "localhost", err
	}
	return ip, nil
}

// https://stackoverflow.com/a/23558495/15807350
func getInternalIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", errors.New("no internet connection")
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func GetPublicIP() (string, error) {
	resp, err := http.Get("https://api.ipify.org?format=text")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	ip, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if !checkIfIpAccessible(string(ip)) {
		return string(ip), errors.New("PORT is blocked by firewall")
	}
	return string(ip), nil
}

func checkIfIpAccessible(ip string) bool {
	conn, err := net.Dial("tcp", ip+":80")
	if err != nil {
		return false
	}
	defer conn.Close()
	return true
}

func stripInt(log *zap.Logger, a int) int {
	strA := strconv.Itoa(abs(a))
	lastDigits := strings.Replace(strA, "100", "", 1)
	result, err := strconv.Atoi(lastDigits)
	if err != nil {
		log.Sugar().Fatalln(err)
		return 0
	}
	return result
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
