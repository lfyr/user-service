package config

import (
	"time"

	"github.com/afex/hystrix-go/hystrix"
)

// 全局配置结构体（所有配置集中管理，后续可改为从 yaml/viper 读取）
type GlobalConfig struct {
	Etcd    EtcdConfig    `json:"etcd"`
	Hystrix HystrixConfig `json:"hystrix"`
	Gin     GinConfig     `json:"gin"`
	Service ServiceConfig `json:"service"`
}

// Etcd 配置
type EtcdConfig struct {
	Adders []string `json:"addrs"` // Etcd 地址列表（集群支持）
}

// Hystrix 熔断降级配置（针对用户服务）
type HystrixConfig struct {
	CommandName            string `json:"command_name"`
	Timeout                int    `json:"timeout"`                  // 超时时间（毫秒）
	MaxConcurrentRequests  int    `json:"max_concurrent_requests"`  // 最大并发
	ErrorPercentThreshold  int    `json:"error_percent_threshold"`  // 错误率阈值
	SleepWindow            int    `json:"sleep_window"`             // 熔断休眠窗口（毫秒）
	RequestVolumeThreshold int    `json:"request_volume_threshold"` // 最小触发请求数
}

// Gin Web 服务配置
type GinConfig struct {
	Port string `json:"port"` // 监听端口
	Mode string `json:"mode"` // 运行模式（debug/release）
}

// 微服务配置
type ServiceConfig struct {
	UserName string        `json:"user_name"` // 用户服务名称
	Version  string        `json:"version"`   // 服务版本
	Timeout  time.Duration `json:"timeout"`   // 服务端处理超时
}

// 初始化全局默认配置（后续可替换为配置文件读取）
func InitGlobalConfig() *GlobalConfig {
	return &GlobalConfig{
		Etcd: EtcdConfig{
			Adders: []string{"127.0.0.1:2379"},
		},
		Hystrix: HystrixConfig{
			CommandName:            "user.service.Register",
			Timeout:                3000,
			MaxConcurrentRequests:  100,
			ErrorPercentThreshold:  50,
			SleepWindow:            5000,
			RequestVolumeThreshold: 10,
		},
		Gin: GinConfig{
			Port: ":8080",
			Mode: "debug",
		},
		Service: ServiceConfig{
			UserName: "user.service",
			Version:  "v1.0.0",
			Timeout:  3 * time.Second,
		},
	}
}

// 初始化 Hystrix 配置
func InitHystrixConfig(hc HystrixConfig) {
	hystrix.ConfigureCommand(hc.CommandName, hystrix.CommandConfig{
		Timeout:                hc.Timeout,
		MaxConcurrentRequests:  hc.MaxConcurrentRequests,
		ErrorPercentThreshold:  hc.ErrorPercentThreshold,
		SleepWindow:            hc.SleepWindow,
		RequestVolumeThreshold: hc.RequestVolumeThreshold,
	})
}
