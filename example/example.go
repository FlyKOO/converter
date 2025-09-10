package main

// 示例程序：连接 Geyser 节点并解析交易

import (
	"context"
	"fmt"
	"log"

	geyserAdapter "github.com/FlyKOO/converter/geyser"
	"github.com/FlyKOO/converter/shared"
	"github.com/FlyKOO/converter/utils"

	"github.com/mr-tron/base58"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
)

const (
	targetAccount = "JUP6LkbZbjS1jKKwapdHNy74zcZ3tLUZoi5QNyVTaV4" // 需要监控的目标账户
	endpointURL   = "http://yourgrpc:10000"                       // gRPC 服务地址
)

func main() {
	// 初始化 Geyser 适配器
	grpcAdapter := geyserAdapter.NewGeyserAdapter()

	// 创建与 Geyser 服务的 gRPC 连接
	conn, err := grpcAdapter.CreateGRPCConnection(endpointURL)
	if err != nil {
		log.Fatalf("连接失败: %s", err)
	}
	defer conn.Close()

	// 创建客户端并订阅交易
	client := pb.NewGeyserClient(conn)
	stream, err := client.Subscribe(context.Background())
	if err != nil {
		log.Fatalf("创建流失败: %s", err)
	}

	if err := stream.Send(grpcAdapter.CreateSubscriptionRequest(targetAccount)); err != nil {
		log.Fatalf("订阅失败: %s", err)
	}

	fmt.Printf("🔭 正在监控账户 %s 的交易\n", targetAccount)

	// 持续接收交易更新
	for {
		update, err := stream.Recv()
		if err != nil {
			log.Fatalf("流错误: %s", err)
		}
		go processUpdate(update)
	}
}

func processUpdate(update *pb.SubscribeUpdate) *shared.TransactionDetails {
	// ping 消息不包含交易内容
	if update.GetTransaction() == nil {
		return nil
	}

	tx := update.GetTransaction()
	signature := tx.GetTransaction().GetSignature()

	// 处理没有签名的交易
	if signature == nil {
		return nil
	}

	sigStr := base58.Encode(signature)
	txDetails, err := utils.ProcessTransactionToStruct(tx, sigStr)
	if err != nil {
		fmt.Printf("处理交易详情出错: %s", err)
	}
	fmt.Println("已处理交易", txDetails.Signature)
	return txDetails
}
