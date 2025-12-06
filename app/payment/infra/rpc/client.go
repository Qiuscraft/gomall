package rpc

import (
	"sync"

	"github.com/cloudwego/biz-demo/gomall/app/payment/conf"
	"github.com/cloudwego/biz-demo/gomall/common/clientsuite"
	"github.com/cloudwego/biz-demo/gomall/rpc_gen/kitex_gen/order/orderservice"
	"github.com/cloudwego/kitex/client"
)

var (
	OrderClient orderservice.Client
	once        sync.Once
	err         error
)

// InitClient 初始化 order 的 RPC 客户端。
func InitClient() {
	once.Do(func() {
		registryAddr := conf.GetConf().Registry.RegistryAddress[0]
		serviceName := conf.GetConf().Kitex.Service
		opts := []client.Option{
			client.WithSuite(clientsuite.CommonGrpcClientSuite{
				RegistryAddr:       registryAddr,
				CurrentServiceName: serviceName,
			}),
		}
		OrderClient, err = orderservice.NewClient("order", opts...)
		if err != nil {
			panic(err)
		}
	})
}
