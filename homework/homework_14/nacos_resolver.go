package homework14

import (
	"fmt"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/model"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"google.golang.org/grpc/resolver"
)

type nacosResolverBuilder struct {
	client naming_client.INamingClient
}

func (n *nacosResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	res := &nacosResolver{client: n.client, target: target, cc: cc}
	return res, res.subscribe()
}

func (n nacosResolverBuilder) Scheme() string {
	return "nacos"
}

type nacosResolver struct {
	client naming_client.INamingClient
	target resolver.Target
	cc     resolver.ClientConn
}

func (n nacosResolver) ResolveNow(options resolver.ResolveNowOptions) {
	//TODO implement me
	panic("implement me")
}

func (n nacosResolver) subscribe() error {
	err := n.client.Subscribe(&vo.SubscribeParam{
		ServiceName: n.target.Endpoint(),
		SubscribeCallback: func(services []model.Instance, err error) {
			if err != nil {
				return
			}
			err = n.reportAddrs(services)
			if err != nil {
				// 更新节点失败，记录日志
			}
		},
	})
	return err
}

func (n nacosResolver) Close() {
}

func (n nacosResolver) reportAddrs(services []model.Instance) error {
	addrs := make([]resolver.Address, 0, len(services))
	for _, svc := range services {
		addrs = append(addrs, resolver.Address{
			Addr: fmt.Sprintf("%s:%d", svc.Ip, svc.Port)})
	}
	return n.cc.UpdateState(resolver.State{Addresses: addrs})
}
