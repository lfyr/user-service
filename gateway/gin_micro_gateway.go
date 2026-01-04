package main

import (
	"context"
	"github.com/gin-gonic/gin"
	"github.com/go-micro/plugins/v4/client/grpc"
	"github.com/go-micro/plugins/v4/registry/etcd"
	"go-micro.dev/v4"
	"go-micro.dev/v4/registry"
	"net/http"
	"time"
	"user-service/proto/user" // 导入你的 Protobuf 生成代码
)

// 全局变量：创建 user.service 的 RPC 客户端代理（复用 go-micro 客户端）
var userRpcClient user.UserService

// 初始化 go-micro + Etcd + RPC 客户端
func initMicro() {
	// 1. 初始化 Etcd 注册中心（与原有 user.service 配置一致）
	etcdReg := etcd.NewRegistry(
		registry.Addrs("127.0.0.1:2379"),
	)

	// 2. 初始化 go-micro 服务（仅用于创建 RPC 客户端，无需启动 Server）
	microService := micro.NewService(
		micro.Name("user.gin.gateway"), // 网关服务名称（标识作用）
		micro.Registry(etcdReg),        // 绑定 Etcd 注册中心
		micro.Client(grpc.NewClient()), // 绑定 gRPC 客户端（与 user.service 协议匹配）
	)
	microService.Init()

	// 3. 创建 user.service 的 RPC 客户端代理（通过服务名发现）
	userRpcClient = user.NewUserService("user.service", microService.Client())
}

func main() {
	// 1. 初始化 go-micro + RPC 客户端
	initMicro()

	// 2. 初始化 Gin 引擎（默认模式，生产环境可改为 gin.ReleaseMode）
	r := gin.Default()

	// 3. 配置全局中间件（可选，提升接口健壮性）
	r.Use()

	// 4. 注册 HTTP 路由（RESTful 风格，对应原有 RPC 接口）
	apiGroup := r.Group("/api") // 接口前缀分组
	{
		// POST /client/register：用户注册（对应 RPC Register 接口）
		apiGroup.POST("/register", registerHandler)

		// GET /client/users：查询用户列表（对应 RPC List 接口）
		// apiGroup.GET("/users", listUsersHandler)
	}

	// 5. 启动 Gin 服务（监听 8080 端口，前端可访问 http://localhost:8080/api/xxx）
	if err := r.Run(":8080"); err != nil {
		panic("Gin 服务启动失败：" + err.Error())
	}
}

// 跨域中间件（解决前端跨域请求问题）
func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// 处理用户注册 HTTP 请求（转发到 RPC Register 接口）
func registerHandler(c *gin.Context) {
	// 1. 定义 HTTP 请求参数结构体（与前端传入的 JSON 格式对应）
	var reqDTO struct {
		Username string `json:"username" binding:"required"` // binding:"required" 必传校验
		Password string `json:"password" binding:"required"`
		Email    string `json:"email"`
	}

	// 2. 解析 JSON 请求参数（Gin 内置绑定，自动校验必传项）
	if err := c.ShouldBindJSON(&reqDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    1,
			"message": "参数解析失败：" + err.Error(),
			"data":    nil,
		})
		return
	}

	// 3. 构造 RPC 请求参数（转换为 Protobuf 定义的 RegisterRequest）
	rpcReq := &user.RegisterRequest{
		Username: reqDTO.Username,
		Password: reqDTO.Password,
		Email:    reqDTO.Email,
	}

	// 4. 调用 RPC 接口（添加超时控制，避免长时间阻塞）
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel() // 超时后释放资源
	rpcRsp, err := userRpcClient.Register(ctx, rpcReq)

	// 5. 处理 RPC 调用错误
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    1,
			"message": "注册失败：" + err.Error(),
			"data":    nil,
		})
		return
	}

	// 6. 转换 RPC 响应为 HTTP 响应（统一返回格式）
	c.JSON(http.StatusOK, gin.H{
		"code":    rpcRsp.Code,
		"message": rpcRsp.Message,
		"data": gin.H{
			"userId": rpcRsp.UserId,
		},
	})
}

// 处理查询用户列表 HTTP 请求（转发到 RPC List 接口）
//func listUsersHandler(c *gin.Context) {
//	// 1. 构造 RPC 请求参数（ListRequest 暂无需参数）
//	rpcReq := &user.ListRequest{}
//
//	// 2. 调用 RPC 接口（添加超时控制）
//	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//	defer cancel()
//	rpcRsp, err := userRpcClient.List(ctx, rpcReq)
//
//	// 3. 处理 RPC 调用错误
//	if err != nil {
//		c.JSON(http.StatusInternalServerError, gin.H{
//			"code":    1,
//			"message": "查询用户列表失败：" + err.Error(),
//			"data":    nil,
//		})
//		return
//	}
//
//	// 4. 转换 RPC 响应为 HTTP 响应（统一返回格式）
//	c.JSON(http.StatusOK, gin.H{
//		"code":    rpcRsp.Code,
//		"message": rpcRsp.Message,
//		"data": gin.H{
//			"total":     len(rpcRsp.Usernames),
//			"usernames": rpcRsp.Usernames,
//		},
//	})
//}
