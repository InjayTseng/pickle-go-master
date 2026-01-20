package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-gonic/gin"
)

// SentryConfig Sentry 配置
type SentryConfig struct {
	DSN         string
	Environment string
	Release     string
	Debug       bool
}

// InitSentry 初始化 Sentry SDK
// 初始化 Sentry 錯誤監控
func InitSentry(cfg SentryConfig) error {
	if cfg.DSN == "" {
		// 如果沒有設定 DSN，跳過初始化
		return nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.DSN,
		Environment: cfg.Environment,
		Release:     cfg.Release,
		Debug:       cfg.Debug,

		// 設定取樣率
		TracesSampleRate: 0.1, // 10% 的請求會被追蹤

		// 在傳送前處理事件
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			// 可以在這裡過濾或修改事件
			return event
		},
	})

	return err
}

// Sentry 返回 Sentry 錯誤處理中間件
// 用於捕捉 panic 和記錄錯誤到 Sentry
func Sentry() gin.HandlerFunc {
	return sentrygin.New(sentrygin.Options{
		Repanic: true, // 重新觸發 panic，讓 Recovery 中間件也能處理
	})
}

// SentryRecovery 返回帶有 Sentry 報告的 Recovery 中間件
// 捕捉 panic 並報告到 Sentry
func SentryRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// 取得 stack trace
				stack := string(debug.Stack())

				// 取得 Sentry hub
				if hub := sentrygin.GetHubFromContext(c); hub != nil {
					hub.WithScope(func(scope *sentry.Scope) {
						// 設定額外資訊
						scope.SetRequest(c.Request)
						scope.SetTag("panic", "true")
						scope.SetExtra("stack_trace", stack)

						// 設定使用者資訊（如果有）
						if userID, exists := c.Get("userID"); exists {
							scope.SetUser(sentry.User{ID: fmt.Sprintf("%v", userID)})
						}

						// 捕捉錯誤
						hub.CaptureException(fmt.Errorf("panic recovered: %v", r))
					})

					// 等待事件傳送完成
					hub.Flush(2000)
				}

				// 回傳錯誤回應
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Internal Server Error",
				})
			}
		}()
		c.Next()
	}
}

// CaptureError 手動捕捉錯誤並報告到 Sentry
// 用於在程式中手動報告錯誤
func CaptureError(c *gin.Context, err error, tags map[string]string) {
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			// 設定請求資訊
			scope.SetRequest(c.Request)

			// 設定標籤
			for key, value := range tags {
				scope.SetTag(key, value)
			}

			// 設定使用者資訊
			if userID, exists := c.Get("userID"); exists {
				scope.SetUser(sentry.User{ID: fmt.Sprintf("%v", userID)})
			}

			// 捕捉錯誤
			hub.CaptureException(err)
		})
	}
}

// CaptureMessage 手動捕捉訊息並報告到 Sentry
// 用於記錄重要事件或警告
func CaptureMessage(c *gin.Context, message string, level sentry.Level) {
	if hub := sentrygin.GetHubFromContext(c); hub != nil {
		hub.WithScope(func(scope *sentry.Scope) {
			scope.SetLevel(level)
			scope.SetRequest(c.Request)

			if userID, exists := c.Get("userID"); exists {
				scope.SetUser(sentry.User{ID: fmt.Sprintf("%v", userID)})
			}

			hub.CaptureMessage(message)
		})
	}
}
