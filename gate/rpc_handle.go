package mgate

import (
	"fmt"
	"github.com/liangdas/mqant/gate"
	"github.com/liangdas/mqant/module"
	"github.com/liangdas/mqant/registry"
	"github.com/liangdas/mqant/selector"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type WeightNodeSlice []*WeightNode

func (s WeightNodeSlice) Len() int { return len(s) }

func (s WeightNodeSlice) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s WeightNodeSlice) Less(i, j int) bool {

	return strings.EqualFold(s[i].Node.Metadata["hostname"], s[j].Node.Metadata["hostname"])
}

type WeightNode struct {
	Node   *registry.Node
	Weight int
}

func (this *Gate) Selector(session gate.Session, topic string, u *url.URL) (s module.ServerSession, err error) {
	moduleType := u.Scheme
	nodeId := u.Hostname()

	//使用自己的
	if nodeId == "modulus" {
		//取模
	} else if nodeId == "cache" {
		//缓存
	} else if nodeId == "random" {
		//随机
	}else {
		//
		//指定节点规则就是 module://[user:pass@]nodeId/path
		//方式1
		//moduleType=fmt.Sprintf("%v@%v",moduleType,u.Hostname())
		//方式2
		serverID := fmt.Sprintf("%v@%v", moduleType, nodeId)
		return this.GetRouteServer(moduleType,selector.WithStrategy(func(services []*registry.Service) selector.Next {
			var nodes []WeightNode
			// Filter the nodes for datacenter
			for _, service := range services {
				for _, node := range service.Nodes {
					if node.Id == serverID {
						weight := 100
						if w, ok := node.Metadata["weight"]; ok {
							wint, err := strconv.Atoi(w)
							if err == nil {
								weight = wint
							}
						}
						if state, ok := node.Metadata["state"]; ok {
							if state != "forbidden" {
								nodes = append(nodes, WeightNode{
									Node:   node,
									Weight: weight,
								})
							}
						} else {
							nodes = append(nodes, WeightNode{
								Node:   node,
								Weight: weight,
							})
						}
					}
				}
			}
			//log.Info("services[0] $v",services[0].Nodes[0])
			return func() (*registry.Node, error) {
				if len(nodes) == 0 {
					return nil, fmt.Errorf("no node")
				}
				rand.Seed(time.Now().UnixNano())
				//按权重选
				total := 0
				for _, n := range nodes {
					total += n.Weight
				}
				if total > 0 {
					weight := rand.Intn(total)
					togo := 0
					for _, a := range nodes {
						if (togo <= weight) && (weight < (togo + a.Weight)) {
							return a.Node, nil
						} else {
							togo += a.Weight
						}
					}
				}
				//降级为随机
				index := rand.Intn(int(len(nodes)))
				return nodes[index].Node, nil
			}
		}));
	}
	return this.GetRouteServer(moduleType)
}
