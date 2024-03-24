package homework14

import (
	"context"
	"github.com/nacos-group/nacos-sdk-go/v2/clients"
	"github.com/nacos-group/nacos-sdk-go/v2/clients/naming_client"
	"github.com/nacos-group/nacos-sdk-go/v2/common/constant"
	"github.com/nacos-group/nacos-sdk-go/v2/vo"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
	"testing"
)

type NacosTestSuite struct {
	suite.Suite
	client naming_client.INamingClient
}

func (s *NacosTestSuite) SetupSuite() {
	clientConfig := constant.ClientConfig{
		NamespaceId:         "e525eafa-f7d7-4029-83d9-008937f9d468",
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
	}
	serverConfigs := []constant.ServerConfig{
		{
			IpAddr:      "localhost",
			ContextPath: "/nacos",
			Port:        8888,
			Scheme:      "http",
		},
	}
	namingClient, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ClientConfig:  &clientConfig,
			ServerConfigs: serverConfigs,
		},
	)
	require.NoError(s.T(), err)
	s.client = namingClient
}

func (s *NacosTestSuite) TestNacosServer() {
	l, err := net.Listen("tcp", ":8888")
	require.NoError(s.T(), err)
	srv := grpc.NewServer()
	RegisterUserServiceServer(srv, &Server{})
	ok, err := s.client.RegisterInstance(vo.RegisterInstanceParam{
		Ip:          GetOutboundIP(),
		Port:        8888,
		ServiceName: "user",
		Weight:      10,
		Enable:      true,
		Healthy:     true,
	})
	require.NoError(s.T(), err)
	require.True(s.T(), ok)
	err = srv.Serve(l)
	require.NoError(s.T(), err)
}

func (s *NacosTestSuite) TestClient() {
	nacosResolver := &nacosResolverBuilder{
		client: s.client,
	}
	cc, err := grpc.Dial("nacos:///user",
		grpc.WithResolvers(nacosResolver),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	require.NoError(s.T(), err)
	client := NewUserServiceClient(cc)
	resp, err := client.GetById(context.Background(), &GetByIdReq{
		Id: 123,
	})
	require.NoError(s.T(), err)
	s.T().Log(resp.User)
}

func TestNacos(t *testing.T) {
	suite.Run(t, new(NacosTestSuite))
}
