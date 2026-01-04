package middleware

import (
	"context"

	"github.com/afex/hystrix-go/hystrix"
	"github.com/micro/go-micro/v2/server"
	"user-service/config"
)

// MicroRateLimit 微服务限流拦截器（基于Hystrix）
func MicroRateLimit(maxConcurrent int) server.HandlerInterceptor {
	return func(ctx context.Context, req server.Request, rsp interface{}, next server.HandlerFunc) error {
		// 配置Hystrix命令
		commandName := req.Service() + "." + req.Endpoint()
		hystrix.ConfigureCommand(commandName, hystrix.CommandConfig{
			MaxConcurrentRequests: maxConcurrent,
			Timeout:               3000,
		})

		// 执行Hystrix命令
		var err error
		hystrix.Do(commandName, func() error {
			err = next(ctx, req, rsp)
			return err
		}, func(e error) error {
			// 降级逻辑
			return e
		})
		return err
	}
}