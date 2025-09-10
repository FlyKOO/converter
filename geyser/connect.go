package geyserAdapter

// 与 Geyser 节点交互的辅助工具

import (
	"crypto/x509"
	"fmt"
	"net/url"
	"time"

	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// GeyserUtils 提供与 Geyser gRPC 服务交互的常用方法
type GeyserUtils struct{}

// NewGeyserAdapter 创建 Geyser 工具实例
func NewGeyserAdapter() GeyserUtils { return GeyserUtils{} }

// CreateSubscriptionRequest 构造订阅请求，只订阅指定账户的交易
func (x GeyserUtils) CreateSubscriptionRequest(account string) *pb.SubscribeRequest {
	f := false // 只接收成功交易
	return &pb.SubscribeRequest{
		Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
			"": {
				Failed:         &f,
				AccountInclude: []string{account},
			},
		},
	}
}

// CreateGRPCConnection 根据 endpoint 创建 gRPC 连接
func (x GeyserUtils) CreateGRPCConnection(endpoint string) (*grpc.ClientConn, error) {
	// 解析地址
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, fmt.Errorf("endpoint 地址无效: %w", err)
	}

	// 默认端口
	port := u.Port()
	if port == "" {
		port = "80"
	}

	// 配置 keepalive 参数
	opts := []grpc.DialOption{
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                1000 * time.Second,
			Timeout:             time.Second,
			PermitWithoutStream: true,
		}),
	}

	// 根据协议选择 TLS 或明文连接
	if u.Scheme == "https" {
		pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("获取系统证书池失败: %w", err)
		}
		opts = append(opts, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	return grpc.Dial(fmt.Sprintf("%s:%s", u.Hostname(), port), opts...)
}
