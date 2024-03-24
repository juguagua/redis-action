package chapter05

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"redis-practice/common"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Cache struct {
	*common.Client
}

func (c *Cache) LogRecent(ctx context.Context, name, message, severity string, pipeline redis.Pipeliner) {
	if severity == "" {
		severity = "INFO"
	}
	// 1. 组合日志名和严重等级作为列表名
	logListKey := fmt.Sprintf("%s%s:%s", common.RecentLogListPre, name, severity)
	// 2. 格式化日志消息体，格式为`时间_message`
	msg := fmt.Sprintf("%s-%s", time.Now().Local().String(), message)
	// 3. 流水线执行日志的入队，并维持日志队列大小为 100
	if pipeline == nil {
		pipeline = c.Client.Pipeline()
	}
	pipeline.LPush(ctx, logListKey, msg)
	pipeline.LRange(ctx, logListKey, 0, 99)

	if _, err := pipeline.Exec(ctx); err != nil {
		logrus.Error("log recent list add failed", err)
	}
}

func (c *Cache) LogCommon(ctx context.Context, name, message, severity string, timeout int64) {
	if severity == "" {
		severity = "INFO"
	}
	// 1. 组合日志名和严重等级作为有序集合名
	logZSetKey := fmt.Sprintf("%s%s:%s", common.FrequencyLogZSetPre, name, severity)
	// 2. 组合集合名和 start 作为记录当前小时数的键
	startKey := fmt.Sprintf("%s:start", logZSetKey)
	// 3. 设置循环的超时时间
	end := time.Now().Add(time.Duration(timeout) * time.Millisecond)
	// 4. 重试窗口时间
	for time.Now().Before(end) {
		// 监视 当前小时数的键，如果其发生变化，日志则应当存储到新的有序集合中
		err := c.Client.Watch(ctx, func(tx *redis.Tx) error {
			// 得到当前时间
			hourStart := time.Now().Local().Hour()
			// 查看当前小时数记录的值
			existing, _ := strconv.Atoi(tx.Get(ctx, startKey).Val())

			_, err := tx.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
				// 如果记录的小时数已经过期，即来到了新的小时区间，则应该将旧的日志集合封存，将日志写入新的日志集合中
				if existing != 0 && existing < hourStart {
					pipeline.Rename(ctx, logZSetKey, fmt.Sprintf("%s:last", logZSetKey))
					pipeline.Set(ctx, startKey, hourStart, 0)
				} else if existing == 0 {
					pipeline.Set(ctx, startKey, hourStart, 0)
				}
				pipeline.ZIncrBy(ctx, logZSetKey, 1, message)
				c.LogRecent(ctx, name, message, severity, pipeline)

				return nil
			})
			if err != nil {
				logrus.Error("log common pipeline failed", err)
				return err
			}

			return nil
		}, startKey)

		if err != nil {
			logrus.Error("watch failed, err: ", err)
			continue
		}
	}
}
