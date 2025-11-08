package types

import (
	"crypto/md5"
	"encoding/hex"
	"reflect"
	"strconv"
	"time"

	"github.com/gotd/td/tg"
)

type File struct {
	Location tg.InputFileLocationClass
	FileSize int64
	FileName string
	MimeType string
	ID       int64
}

type HashableFileStruct struct {
	FileName string
	FileSize int64
	MimeType string
	FileID   int64
}

// 上传结果
type UploadResult struct {
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	MimeType    string    `json:"mimeType"`
	MessageID   int       `json:"messageId"`
	StreamURL   string    `json:"streamUrl"`
	DownloadURL string    `json:"downloadUrl"`
	Hash        string    `json:"hash"`
	UploadTime  time.Time `json:"uploadTime"`
}

// 用户配额信息
type UserQuotaInfo struct {
	UserID      string  `json:"userId"`
	UsedQuota   int64   `json:"usedQuota"`
	MaxQuota    int64   `json:"maxQuota"`
	QuotaPercent float64 `json:"quotaPercent"`
	Remaining    int64   `json:"remaining"`
}

// 上传状态响应
type UploadStatusResponse struct {
	Success     bool                   `json:"success"`
	Message     string                 `json:"message"`
	TotalFiles  int                    `json:"totalFiles"`
	Results     map[string]interface{}    `json:"results"`
	Timestamp   int64                  `json:"timestamp"`
}

func (f *HashableFileStruct) Pack() string {
	hasher := md5.New()
	val := reflect.ValueOf(*f)
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)

		var fieldValue []byte
		switch field.Kind() {
		case reflect.String:
			fieldValue = []byte(field.String())
		case reflect.Int64:
			fieldValue = []byte(strconv.FormatInt(field.Int(), 10))
		}

		hasher.Write(fieldValue)
	}
	return hex.EncodeToString(hasher.Sum(nil))
}
