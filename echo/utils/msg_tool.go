package utils

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/liangdas/mqant/log"
	"strings"
)

type RetMsg struct {
	Code   int           `json:"code"`
	Result interface{}   `json:"result"`
	Error  string        `json:"error"`
	Trace  log.TraceSpan `json:"trace"`
	ctx    echo.Context  `json:"-"`
}

func NewRetMsg(ctx echo.Context) *RetMsg {
	var trace log.TraceSpan = nil
	if ctx != nil {
		t := ctx.Get("trace")
		if t != nil {
			trace = ctx.Get("trace").(*log.TraceSpanImp)
		}
	}
	return &RetMsg{
		ctx:    ctx,
		Result: map[string]string{},
		Trace:  trace,
	}
}

func (ret *RetMsg) GetTrace() log.TraceSpan {
	return ret.Trace
}

func getRequestID(ctx echo.Context) string {
	if ctx == nil {
		return ""
	}
	return ctx.Get("trace").(log.TraceSpan).TraceId()
}

func (ret *RetMsg) PackError(code int, msg ...interface{}) {
	ret.Code = code
	if len(msg) == 0 {
		ret.Error = getErrorMsg(code)
	} else {
		f := strings.Repeat("%+v ", len(msg))
		ret.Error = fmt.Sprintf(f, msg...)
	}
}

func (ret *RetMsg) Write(result []byte) {
	ret.ctx.Response().Write(result)
}

func (ret *RetMsg) PackResult(result interface{}) {
	ret.Result = result
}
