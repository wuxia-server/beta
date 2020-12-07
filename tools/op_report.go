package tools

import (
	"fmt"
	"github.com/labstack/echo"
	"github.com/liangdas/mqant/log"
	beegolog "github.com/liangdas/mqant/log/beego"
	"github.com/liangdas/mqant/module"
	"os"
	"strings"
	"time"
	"beta/echo/utils"
)

const TIME_LAYOUT = "2006-01-02 15:04:05"

var op *beegolog.BeeLogger

var platformDomain, platformName string

func InitOpReport(app module.App, PlatformName, PlatformDomain string) error {
	platformDomain = PlatformDomain
	platformName = PlatformName
	defaultBIPath := fmt.Sprintf("%s/op", app.WorkDir())
	_, err := os.Open(defaultBIPath)
	if err != nil {
		//文件不存在
		err := os.Mkdir(defaultBIPath, os.ModePerm) //
		if err != nil {
			fmt.Println(err)
		}
	}
	op = log.NewBeegoLogger(false, app.GetProcessID(), defaultBIPath, app.GetSettings().OP)
	return nil
}

type OpEvent struct {
	PlatformDomain string            `json:"platform_domain"` // 平台域名或标识
	PlatformName   string            `json:"platform_name"`   // 区分不同的后台，比如大棋牌gm_dqp,安徽gm_anhui,um
	PlatformZone   string            `json:"platform_zone"`   // 时区
	ActionTime     string            `json:"action_time"`     //  操作时间  2019-01-01 01:01:01
	ActionUser     string            `json:"action_user"`     // 操作姓名
	ActionAccount  string            `json:"action_account"`  // 操作账号
	ActionDevice   string            `json:"action_device"`   // 操作设备、浏览器信息、系统信息（安卓  ios  windows）
	SourceIp       string            `json:"source_ip"`       // 来源ip,本地IP(网卡配置的IP)
	ActionType     string            `json:"action_type"`     // 操作类型，比如update delete select
	ActionDetails  string            `json:"action_details"`  // 操作详细内容，比如查询的某一项内容或者页面
	ExtraParams    map[string]string `json:"extra_info"`
	/**
	{"param01":"v1","param02":"v2"}
	extra_info字段为必需字段，字段值为json格式，
	主要是针对不同类型的后台重要补充信息的扩展，param01 param02字段名可以自定义，其值尽可能是字符串。
	字段数量可自行添加，如param03、param04。比如扩展字段可以包括url。字段中的value都需要是string类型。
	这样是为了避免 当不同后台扩展字段中出现同样的key时，value的数据类型不冲突。
	*/
}

func OpReport(ctx echo.Context, event OpEvent) {
	event.PlatformDomain = platformDomain
	event.PlatformName = platformName
	t := time.Now()
	_, offset := t.Zone()
	tStr := t.Format(TIME_LAYOUT)
	event.ActionTime = tStr
	if offset > 0 {
		event.PlatformZone = fmt.Sprintf("+%d", offset/3600)
	} else {
		event.PlatformZone = fmt.Sprintf("%d", offset/3600)
	}
	if ctx != nil {
		event.SourceIp = strings.Join([]string{ctx.RealIP(), ctx.Request().Header.Get("X-Local-IP")}, ",")
		event.ActionDevice = ctx.Request().UserAgent()
	}
	op.BiReport(utils.Struct2String(event))
}
