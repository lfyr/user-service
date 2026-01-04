package common

import (
	"fmt"
	"math/rand"
	"time"
)

// 生成唯一用户ID（后续可替换为雪花算法）
func GenerateUserId() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("USER_%d%06d", time.Now().Unix(), rand.Intn(999999))
}