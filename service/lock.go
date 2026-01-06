package main

import (
	"context"
	"errors"
	"fmt"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"log"
	"sync"
	"time"
)

// EtcdDistLock 基于官方concurrency包封装的分布式锁结构体
type EtcdDistLock struct {
	etcdClient *clientv3.Client     // ETCD客户端实例
	session    *concurrency.Session // ETCD会话（管理租约续期）
	mutex      *concurrency.Mutex   // ETCD互斥锁
	lockPrefix string               // 锁前缀（区分不同资源的锁，必须唯一）
}

// NewEtcdDistLock 初始化分布式锁实例
// endpoints: ETCD集群地址（单机版为[]string{"127.0.0.1:2379"}）
// lockPrefix: 锁前缀（如"/dist/lock/stock/"，保护库存资源）
func NewEtcdDistLock(endpoints []string, lockPrefix string) (*EtcdDistLock, error) {
	// 1. 配置ETCD客户端
	config := clientv3.Config{
		Endpoints:   endpoints,       // ETCD地址
		DialTimeout: 5 * time.Second, // 连接超时时间
		Username:    "",              // 无认证则留空
		Password:    "",              // 无认证则留空
	}

	// 2. 创建ETCD客户端
	cli, err := clientv3.New(config)
	if err != nil {
		return nil, fmt.Errorf("创建ETCD客户端失败：%w", err)
	}

	// 3. 初始化分布式锁实例
	return &EtcdDistLock{
		etcdClient: cli,
		lockPrefix: lockPrefix,
	}, nil
}

// Lock 获取分布式锁（阻塞直到获取成功或上下文超时）
// ctx: 上下文（用于控制超时、取消）
// ttl: 租约过期时间（秒），会话会自动续期，建议设置10-30秒
func (l *EtcdDistLock) Lock(ctx context.Context, ttl int) error {
	if l.etcdClient == nil {
		return errors.New("ETCD客户端未初始化")
	}

	// 1. 创建会话（自动处理租约续期，ttl为租约有效期）
	// WithTTL：设置租约TTL
	// 会话内部会启动协程，每隔ttl/3时间向ETCD发送续期请求，保证租约不失效
	session, err := concurrency.NewSession(l.etcdClient, concurrency.WithTTL(ttl))
	if err != nil {
		return fmt.Errorf("创建ETCD会话失败：%w", err)
	}

	// 2. 基于会话和锁前缀创建互斥锁
	// 锁的完整键格式：lockPrefix + 会话ID + 有序序号（由concurrency包自动生成）
	mutex := concurrency.NewMutex(session, l.lockPrefix)

	// 3. 阻塞获取锁（直到获取成功或ctx超时/取消）
	log.Println("尝试获取分布式锁...")
	if err := mutex.Lock(ctx); err != nil {
		session.Close() // 获取锁失败，关闭会话释放资源
		return fmt.Errorf("获取分布式锁失败：%w", err)
	}

	// 4. 保存会话和锁实例
	l.session = session
	l.mutex = mutex
	log.Println("成功获取分布式锁")
	return nil
}

// Unlock 释放分布式锁
// ctx: 上下文（用于控制超时）
func (l *EtcdDistLock) Unlock(ctx context.Context) error {
	if l.mutex == nil || l.session == nil {
		return errors.New("未获取到分布式锁，无需释放")
	}

	// 1. 主动释放锁（删除ETCD中的锁键）
	log.Println("尝试释放分布式锁...")
	if err := l.mutex.Unlock(ctx); err != nil {
		return fmt.Errorf("释放分布式锁失败：%w", err)
	}

	// 2. 关闭会话（销毁租约，停止续期协程）
	l.session.Close()

	// 3. 重置锁实例状态
	l.session = nil
	l.mutex = nil
	log.Println("成功释放分布式锁")
	return nil
}

// Close 关闭ETCD客户端，释放资源
func (l *EtcdDistLock) Close() error {
	if l.etcdClient != nil {
		return l.etcdClient.Close()
	}
	return nil
}

// ---------------------- 测试代码：多客户端竞争同一把锁 ----------------------
func main() {
	// 1. 配置参数
	etcdEndpoints := []string{"127.0.0.1:2379"} // ETCD单机版地址
	lockPrefix := "/dist/lock/test_resource/"   // 同一资源的锁前缀（保证唯一）
	clientCount := 5                            // 模拟5个客户端竞争锁
	leaseTTL := 10                              // 租约TTL 10秒
	var wg sync.WaitGroup                       // 等待所有客户端执行完成

	// 2. 启动多个客户端协程
	for clientID := 0; clientID < clientCount; clientID++ {
		wg.Add(1)
		go func(cid int) {
			defer wg.Done()

			// ① 初始化分布式锁
			lock, err := NewEtcdDistLock(etcdEndpoints, lockPrefix)
			if err != nil {
				log.Printf("客户端%d：初始化锁失败 - %v", cid, err)
				return
			}
			defer lock.Close() // 最后关闭ETCD客户端

			// ② 定义上下文（设置30秒超时，避免无限阻塞）
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel() // 上下文用完取消

			// ③ 获取分布式锁
			if err := lock.Lock(ctx, leaseTTL); err != nil {
				log.Printf("客户端%d：获取锁失败 - %v", cid, err)
				return
			}

			// ④ 执行业务逻辑（模拟耗时操作，如操作共享资源）
			log.Printf("客户端%d：开始执行业务逻辑（耗时2秒）", cid)
			time.Sleep(2 * time.Second) // 模拟业务耗时
			log.Printf("客户端%d：业务逻辑执行完成", cid)

			// ⑤ 释放分布式锁
			if err := lock.Unlock(ctx); err != nil {
				log.Printf("客户端%d：释放锁失败 - %v", cid, err)
				return
			}
		}(clientID)
	}

	// 3. 等待所有客户端执行完成
	wg.Wait()
	log.Println("所有客户端执行完毕")
}
