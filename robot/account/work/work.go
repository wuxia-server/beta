// Copyright 2014 hey Author. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package work

import (
	"fmt"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/liangdas/armyant/task"
	"github.com/liangdas/armyant/work"
	"google.golang.org/protobuf/proto"
	"math/rand"
	"net/url"
	"time"
	pb "github.com/wuxia-server/protobuf/golang/message"
)

func NewWork(manager *Manager) *Work {
	wo := new(Work)
	//wo.rs = &rsync.LRsync{
	//	BlockSize: 16,
	//}
	wo.manager = manager
	//rand.NewSource(time.Now().UnixNano())
	//wo.name = fmt.Sprintf("%v", rand.Intn(100))
	////opts := wo.GetDefaultOptions("ws://ludots.touch4.me:3653")
	//opts := wo.GetDefaultOptions("ws://ludowss.qijihdhk.com")
	//opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
	//	fmt.Println("ConnectionLost", err.Error())
	//})
	//opts.SetOnConnectHandler(func(client MQTT.Client) {
	//	fmt.Println("OnConnectHandler")
	//})
	//// load root ca
	//// 需要一个证书，这里使用的这个网站提供的证书https://curl.haxx.se/docs/caextract.html
	//err := wo.Connect(opts)
	//if err != nil {
	//	fmt.Println(err.Error())
	//}
	//
	//wo.On("/prop/s2c_propchange/", func(client MQTT.Client, msg MQTT.Message) {
	//	fmt.Println("me is", wo.name, msg.Topic(), "=》", string(msg.Payload()))
	//})
	//
	//wo.On("/item/s2c_itemchange/", func(client MQTT.Client, msg MQTT.Message) {
	//	fmt.Println("me is", wo.name, msg.Topic(), "=》", string(msg.Payload()))
	//})
	return wo
}

/**
Work 代表一个协程内具体执行任务工作者
*/
type Work struct {
	work.MqttWork
	manager *Manager
	name    string
	data    []byte
}

/**
每一次请求都会调用该函数,在该函数内实现具体请求操作

task:=task.Task{
		N:1000,	//一共请求次数，会被平均分配给每一个并发协程
		C:100,		//并发数
		//QPS:10,		//每一个并发平均每秒请求次数(限流) 不填代表不限流
}

N/C 可计算出每一个Work(协程) RunWorker将要调用的次数
*/
func (wo *Work) RunWorker(t task.Task) {
	rand.NewSource(time.Now().UnixNano())
	wo.name = fmt.Sprintf("%v", rand.Intn(100))
	opts := wo.GetDefaultOptions("ws://127.0.0.1:3653")
	//opts := wo.GetDefaultOptions("ws://ludots.touch4.me:3653")
	//opts := wo.GetDefaultOptions("ws://ludowss.qijihdhk.com")
	opts.SetConnectionLostHandler(func(client MQTT.Client, err error) {
		fmt.Println("ConnectionLost", err.Error())
	})
	opts.SetOnConnectHandler(func(client MQTT.Client) {
		//fmt.Println("OnConnectHandler")
	})
	// load root ca
	// 需要一个证书，这里使用的这个网站提供的证书https://curl.haxx.se/docs/caextract.html
	err := wo.Connect(opts)
	if err != nil {
		fmt.Println(err.Error())
	}


	//fmt.Println("-----开始登陆-----")
	var userId int64 = 825176369
	c2s_login := &pb.C2S_Login_DEBUG{
		UserId:   *proto.Int64(userId),
		ClientId: *proto.String("Android_5.06_tyGuest%2CtyAccount.alipay.0-hall28.kuaishoulm.fish3d19"),
	}
	c2s_b, err := proto.Marshal(c2s_login)
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	u, _ := url.Parse("account://modulus/account/c2s_login_debug/")
	msg, err := wo.RequestURI(u, c2s_b)
	if err != nil {
		fmt.Println(msg.Topic(), err.Error())
		return
	}
	//fmt.Println("-----登陆完成-----")
	res := &pb.S2C_Response{}
	proto.Unmarshal(msg.Payload(), res)
	if err != nil {
		fmt.Println(msg.Topic(), err.Error())
		return
	}

	if res.Error != "" {
		fmt.Println(msg.Topic(), res.Error)
		return
	}
	fmt.Println("-----登陆成功-----")
	s2c_login := &pb.S2C_Login{}
	proto.Unmarshal(res.Result, s2c_login)
	if err != nil {
		fmt.Println(msg.Topic(), err.Error())
		return
	}
	fmt.Println(msg.Topic(), s2c_login.UserId)


}
func (wo *Work) Init(t task.Task) {

}
func (wo *Work) Close(t task.Task) {
	wo.GetClient().Disconnect(0)
}

