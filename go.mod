module beta

go 1.13

require (
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/eclipse/paho.mqtt.golang v1.2.0
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.0
	github.com/google/btree v1.0.0 // indirect
	github.com/hashicorp/consul/api v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/json-iterator/go v1.1.10
	github.com/labstack/echo v3.3.10+incompatible
	github.com/labstack/gommon v0.3.0 // indirect
	github.com/liangdas/armyant v0.0.4
	github.com/liangdas/mqant v1.3.99
	github.com/nats-io/nats-server/v2 v2.1.9 // indirect
	github.com/nats-io/nats.go v1.10.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/common v0.6.0
	github.com/shirou/gopsutil v3.20.11+incompatible
	github.com/stretchr/testify v1.4.0
	github.com/valyala/fasttemplate v1.2.1 // indirect
	github.com/wuxia-server/protobuf v0.0.1
	go.uber.org/zap v1.10.0
	golang.org/x/net v0.0.0-20200226121028-0de0cce0169b // indirect
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/sys v0.0.0-20200615200032-f1bc736245b1 // indirect
	google.golang.org/protobuf v1.25.0
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	gopkg.in/yaml.v2 v2.2.8 // indirect

)

//replace github.com/wuxia-server/protobuf => /work/go/protobuf

replace google.golang.org/grpc => google.golang.org/grpc v1.26.0
