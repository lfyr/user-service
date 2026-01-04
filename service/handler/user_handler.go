package main

import (
	"context"
	"fmt"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"github.com/go-micro/plugins/v4/server/grpc"
	"go-micro.dev/v4"
	"go-micro.dev/v4/logger"
	"go-micro.dev/v4/registry"
	"math/rand"
	"time"
	"user-service/proto/user"
)

// 1. 定义服务端结构体（用于绑定接口方法）
type UserServiceImpl struct{}

// 2. 模拟内存数据库（存储已注册用户，避免重复注册）
var userDB = make(map[string]bool) // key：用户名，value：是否已注册

// 3. 实现 Protobuf 定义的 Register 接口方法
// 入参：context.Context（微服务必备，传递元数据、超时等）、*user.RegisterRequest（注册请求）
// 返回：*user.RegisterResponse（注册响应）、error（错误信息）
func (u *UserServiceImpl) Register(ctx context.Context, req *user.RegisterRequest, rsp *user.RegisterResponse) error {
	// 步骤 1：参数校验
	if req.Username == "" || req.Password == "" {
		rsp.Code = 1
		rsp.Message = "用户名和密码不能为空"
		return nil
	}
	logger.Infof("用户注册请求：用户名=%s，密码=%s，邮箱=%s", req.Username, req.Password, req.Email)
	// 步骤 2：检查用户名是否已注册
	if _, exists := userDB[req.Username]; exists {
		rsp.Code = 1
		rsp.Message = "用户名已存在，请更换用户名"
		return nil
	}

	// 步骤 3：生成用户ID（模拟唯一ID，生产环境可用雪花算法）
	rand.Seed(time.Now().UnixNano())
	userId := fmt.Sprintf("USER_%d%06d", time.Now().Unix(), rand.Intn(999999))

	// 步骤 4：存入内存数据库
	userDB[req.Username] = true

	// 步骤 5：构造响应结果
	rsp.Code = 0
	rsp.Message = "注册成功"
	rsp.UserId = userId

	logger.Infof("用户注册成功：用户名=%s，用户ID=%s", req.Username, userId)
	return nil
}

func main() {
	// 步骤 1：创建微服务实例
	etcdReg := etcd.NewRegistry(
		registry.Addrs(fmt.Sprintf("%s:%s", "127.0.0.1", "2379")),
	)
	service := micro.NewService(
		micro.Server(grpc.NewServer()),
		micro.Name("user.service"), // 服务名称（客户端通过该名称调用）
		micro.Version("v1.0.0"),    // 服务版本（可选）
		micro.Registry(etcdReg),    // etcd 注册中心
	)

	// 步骤 2：初始化微服务（加载配置、服务发现等）
	service.Init()

	// 步骤 3：注册 UserService 接口到微服务
	err := user.RegisterUserServiceHandler(service.Server(), &UserServiceImpl{})
	if err != nil {
		logger.Fatalf("注册服务失败：%v", err)
	}

	// 步骤 4：启动微服务
	if err = service.Run(); err != nil {
		logger.Fatalf("启动服务失败：%v", err)
	}
}
