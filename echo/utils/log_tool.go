package utils

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/liangdas/mqant/log"
)

// Logger 日志对象
type Logger struct {
	trace      log.TraceSpan
	userAgent  string
	ip         string
	path       string
	method     string
	proto      string
	status     int
	accessId   string
	signType   string
	secretType string
	random     string
	timestamp  string
	signature  string
}
type ContextParam struct {
	IP         string `json:"ip"`
	UserAgent  string `json:"user_agent"`
	Path       string `json:"path"`
	Method     string `json:"method"`
	Proto      string `json:"proto"`
	Status     int    `json:"status"`
	AccessId   string `json:"access_id"`
	SignType   string `json:"sign_type"`
	SecretType string `json:"secret_type"`
	Random     string `json:"random"`
	Timestamp  string `json:"timestamp"`
	Signature  string `json:"signature"`
}

// Debugf 调试格式化打印
func (logger *Logger) Debugf(format string, a ...interface{}) {
	log.TDebug(logger.trace, format, a...)
}

// Errorf 错误格式化打印
func (logger *Logger) Errorf(format string, a ...interface{}) {
	log.TError(logger.trace, format, a...)
}

// Warnf 警告格式化打印
func (logger *Logger) Warnf(format string, a ...interface{}) {
	log.TWarning(logger.trace, format, a...)
}

// Infof 重要信息格式化打印
func (logger *Logger) Infof(format string, a ...interface{}) {
	log.TInfo(logger.trace, format, a...)
}

// GetLogger 获取 Logger 对象
func GetLogger(ctx echo.Context) *Logger {
	return &Logger{
		trace:      ctx.Get("trace").(log.TraceSpan),
		userAgent:  ctx.Request().Header.Get("User-Agent"),
		ip:         ctx.RealIP(),
		path:       ctx.Request().URL.Path,
		method:     ctx.Request().Method,
		proto:      ctx.Request().Proto,
		status:     ctx.Response().Status,
		accessId:   ctx.Request().Header.Get("Access-Id"),
		signType:   ctx.Request().Header.Get("Sign-Type"),
		secretType: ctx.Request().Header.Get("Secret-Type"),
		random:     ctx.Request().Header.Get("Random"),
		timestamp:  ctx.Request().Header.Get("Timestamp"),
		signature:  ctx.Request().Header.Get("Signature"),
	}
}

func GetRootLogger() *Logger {
	return &Logger{
		trace: log.CreateRootTrace(),
	}
}

func CreateLogger(trace, span string) *Logger {
	return &Logger{
		trace: log.CreateTrace(trace, span),
	}
}

type LogMsg struct {
	EventId     string
	SubEventId  string
	ErrorMsg    string
	ErrorReport string
	EventParams interface{}
	Rparam      interface{}
}

// Debugf 调试格式化打印
func (logger *Logger) JDebug(msg LogMsg) {
	log.TDebug(logger.trace, fmt.Sprintf("@%v", msg.EventId), msg.SubEventId, msg.ErrorMsg, msg.ErrorReport, msg.EventParams, msg.Rparam, logger.getContextParam())
}

// Errorf 错误格式化打印
func (logger *Logger) JError(msg LogMsg) {
	log.TError(logger.trace, fmt.Sprintf("@%v", msg.EventId), msg.SubEventId, msg.ErrorMsg, msg.ErrorReport, msg.EventParams, msg.Rparam, logger.getContextParam())
}

// Warnf 警告格式化打印
func (logger *Logger) JWarn(msg LogMsg) {
	log.TWarning(logger.trace, fmt.Sprintf("@%v", msg.EventId), msg.SubEventId, msg.ErrorMsg, msg.ErrorReport, msg.EventParams, msg.Rparam, logger.getContextParam())
}

// Infof 重要信息格式化打印
func (logger *Logger) JInfo(msg LogMsg) {
	log.TInfo(logger.trace, fmt.Sprintf("@%v", msg.EventId), msg.SubEventId, msg.ErrorMsg, msg.ErrorReport, msg.EventParams, msg.Rparam, logger.getContextParam())
}

func (logger *Logger) getContextParam() ContextParam {
	cp := &ContextParam{}
	cp.IP = logger.ip
	cp.UserAgent = logger.userAgent
	cp.Path = logger.path
	cp.Method = logger.method
	cp.Proto = logger.proto
	cp.Status = logger.status
	cp.AccessId = logger.accessId
	cp.SignType = logger.signType
	cp.SecretType = logger.secretType
	cp.Random = logger.random
	cp.Timestamp = logger.timestamp
	cp.Signature = logger.signature
	return *cp
}
