package logs

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// 全局日志变量
var LG *zap.Logger

// 用于配置日志分发lumberjack
type LogConfig struct {
	DebugFileName string `json:"debugFileName"`
	InfoFileName  string `json:"infoFileName"`
	WarnFileName  string `json:"warnFileName"`
	MaxSize       int    `json:"maxsize"`
	MaxAge        int    `json:"max_age"`
	MaxBackups    int    `json:"max_backups"`
}

func InitLogger(cfg *LogConfig) (err error) {
	//初始化lumberjack
	writeSyncerDebug := getLogWriter(cfg.DebugFileName, cfg.MaxSize, cfg.MaxAge, cfg.MaxBackups)
	writeSyncerInfo := getLogWriter(cfg.InfoFileName, cfg.MaxSize, cfg.MaxAge, cfg.MaxBackups)
	writeSyncerWarn := getLogWriter(cfg.WarnFileName, cfg.MaxSize, cfg.MaxAge, cfg.MaxBackups)

	//初始化zap

	//文件输出日志设置
	encoder := getEncoder()
	debugCore := zapcore.NewCore(encoder, writeSyncerDebug, zapcore.DebugLevel)
	infoCore := zapcore.NewCore(encoder, writeSyncerInfo, zapcore.InfoLevel)
	warnCore := zapcore.NewCore(encoder, writeSyncerWarn, zapcore.WarnLevel)

	//标准输出日志设置
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
	std := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), zapcore.DebugLevel)

	// 将 debugCore, infoCore, warnCore, std 打包
	core := zapcore.NewTee(debugCore, infoCore, warnCore, std)
	LG = zap.New(core, zap.AddCaller())
	zap.ReplaceGlobals(LG) // 替换zap包中全局的logger实例，后续在其他包中只需使用zap.L()调用即可
	//L 返回全局 Logger，可以使用 ReplaceGlobals 重新配置。并发使用是安全的。
	return

}

// （不同级别）初始化lumberjack
func getLogWriter(Filename string, MaxSize int, MAxAge int, MaxBackups int) zapcore.WriteSyncer {
	lumberJackLogger := &lumberjack.Logger{
		Filename:   Filename,
		MaxSize:    MaxSize,
		MaxAge:     MAxAge,
		MaxBackups: MaxBackups,
	}
	return zapcore.AddSync(lumberJackLogger)
}

// 配置基础Zap Encoder的输出标准设置
func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.TimeKey = "time"
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderConfig.EncodeDuration = zapcore.SecondsDurationEncoder
	encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
	return zapcore.NewJSONEncoder(encoderConfig)
}

// GinLogger 接收gin框架默认的日志
func GinLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		cost := time.Since(start)
		LG.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
		defer LG.Sync() //关闭前刷新buffer
	}
}

// GinRecovery recover掉项目可能出现的panic，并使用zap记录相关日志
func GinRecovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					LG.Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
					// If the connection is dead, we can't write a status to it.
					c.Error(err.(error)) // nolint: errcheck
					c.Abort()
					return
				}

				if stack {
					LG.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
						zap.String("stack", string(debug.Stack())),
					)
				} else {
					LG.Error("[Recovery from panic]",
						zap.Any("error", err),
						zap.String("request", string(httpRequest)),
					)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
