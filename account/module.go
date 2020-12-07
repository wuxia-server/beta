/**
一定要记得在confin.json配置这个模块的参数,否则无法使用
*/
package account

import (
	"errors"
	"fmt"
	"github.com/liangdas/mqant/conf"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/log"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/module/base"
	"google.golang.org/protobuf/proto"
	pb "github.com/wuxia-server/protobuf/golang/message"
)

var Module = func() module.Module {
	user := new(Account)
	return user
}

type Account struct {
	basemodule.BaseModule
}

func (acc *Account) GetType() string {
	//很关键,需要与配置文件中的Module配置对应
	return "account"
}
func (acc *Account) Version() string {
	//可以在监控时了解代码版本
	return "1.0.0"
}
func (acc *Account) OnInit(app module.App, settings *conf.ModuleSettings) {
	acc.BaseModule.OnInit(acc, app, settings)
	acc.GetServer().RegisterGO(pb.C2SLoginDEBUG, acc.LoginWithUserId)
}

func (acc *Account) Run(closeSig chan bool) {
	<-closeSig
}

func (acc *Account) OnDestroy() {
	//一定别忘了关闭RPC
	acc.GetServer().OnDestroy()
}


func (acc *Account) LoginWithUserId(session gate.Session, req *pb.C2S_Login_DEBUG) (*pb.S2C_Login, error) {
	if session != nil {
		errstr := session.Bind(fmt.Sprintf("%v", req.UserId))
		if errstr != "" {
			return nil, errors.New(errstr)
		}
	}
	log.Info("玩家登陆了")
	return &pb.S2C_Login{
		UserId:       *proto.Int64(req.UserId),
		Nick:         "",
		Avatar:       *proto.String(""),
		Lang:         *proto.String(""),
		AvatarSource: *proto.Uint32(2),
		SystemAvatar: *proto.String(""),
	}, nil
}