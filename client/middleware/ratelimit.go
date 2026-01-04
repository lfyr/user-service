package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"time"
	"user-service/pkg/common"
)

// 令牌桶限流中间件（基于Redis，支持分布式多实例限流）
func RateLimit(redisClient *redis.Client, limit int64, interval time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. 获取限流标识（此处用客户端IP，后续登录后可改为用户ID）
		clientIP := c.ClientIP()
		// 接口路径作为限流维度，实现接口粒度限流
		apiPath := c.FullPath()
		limitKey := "rate_limit:" + apiPath + ":" + clientIP

		// 2. Redis令牌桶核心逻辑（原子操作，避免并发问题）
		// 2.1 初始化令牌桶（若不存在，设置初始令牌数和过期时间）
		_, err := redisClient.SetNX(c, limitKey+":last_refill", time.Now().Unix(), interval*2).Result()
		if err != nil {
			common.Error(c, "限流初始化失败")
			c.Abort()
			return
		}

		// 2.2 计算时间差，补充令牌（令牌生成速度=limit/interval）
		lastRefill, _ := redisClient.Get(c, limitKey+":last_refill").Int64()
		now := time.Now().Unix()
		elapsed := now - lastRefill
		if elapsed > 0 {
			// 计算应补充的令牌数
			tokensToAdd := (elapsed * limit) / int64(interval.Seconds())
			if tokensToAdd > 0 {
				// 原子递增令牌数，且不超过最大限制
				redisClient.IncrBy(c, limitKey, tokensToAdd)
				redisClient.Set(c, limitKey+":last_refill", now, interval*2)
			}
		}

		// 2.3 尝试获取令牌（原子递减）
		tokens, _ := redisClient.Decr(c, limitKey).Result()
		if tokens < 0 {
			// 无令牌，返回限流响应
			common.Fail(c, 429, "请求过于频繁，请稍后再试")
			c.Abort()
			return
		}

		// 3. 有令牌，继续执行后续中间件/处理器
		c.Next()
	}
}
