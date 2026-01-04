package main

import (
	"context"
	"fmt"
	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/registry"
	"user-service/proto/user"
)

func main() {
	// 步骤 1：创建微服务客户端实例
	etcdReg := etcd.NewRegistry(
		registry.Addrs("127.0.0.1:2379"), // etcd 服务地址
	)
	service := micro.NewService(
		micro.Client(grpc.NewClient()),
		micro.Name("user.client"), // 客户端服务名称（仅用于标识）
		micro.Registry(etcdReg),
	)
	service.Init()

	// 步骤 2：创建 UserService 客户端代理（通过服务名称 "user.service" 发现服务端）
	userClient := user.NewUserService("user.service", service.Client())

	// 步骤 3：构造注册请求参数
	req := &user.RegisterRequest{
		Username: "test_user_001",    // 用户名
		Password: "123456Abc",        // 密码
		Email:    "test@example.com", // 邮箱
	}

	// 步骤 4：调用服务端的 Register 方法
	rsp, err := userClient.Register(context.Background(), req)
	if err != nil {
		logger.Fatalf("调用注册接口失败：%v", err)
	}

	// 步骤 5：打印响应结果
	fmt.Println("===== 注册响应结果 =====")
	fmt.Printf("状态码：%d\n", rsp.Code)
	fmt.Printf("提示信息：%s\n", rsp.Message)
	if rsp.UserId != "" {
		fmt.Printf("用户ID：%s\n", rsp.UserId)
	}
}
