package chapter02

import (
	"context"
	"crypto"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"redis-practice/common"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	Client *common.Client
}

func NewCacheClient(conn *common.Client) *Cache {
	return &Cache{Client: conn}
}

// CheckToken 检查该 token 是否被授权，返回相应的 user id
func (c *Cache) CheckToken(ctx context.Context, token string) string {
	return c.Client.HGet(ctx, common.LoginHash, token).Val()
}

// UpdateTokenBehavior 新的请求到来时，存储/更新用户的 token，所对应的最后访问时间，所浏览的商品
func (c *Cache) UpdateTokenBehavior(ctx context.Context, token, user, item string) {
	// 存储/更新用户 token
	c.Client.HSet(ctx, common.LoginHash, token, user)
	now := time.Now().Unix()
	// 更新最后访问时间
	c.Client.ZAdd(ctx, common.Recent, redis.Z{
		Score:  float64(now),
		Member: token,
	})
	// 更新浏览商品
	if item != "" {
		c.Client.ZAdd(ctx, common.ViewedPre+token, redis.Z{
			Score:  float64(now),
			Member: item,
		})
		// 删除排名在 倒数第一 到 正数 26 之间的所有成员
		c.Client.ZRemRangeByRank(ctx, common.ViewedPre+token, 0, -26)
		// 某个商品被浏览，将将其的分值-1，使得被浏览次数最多的商品在最前面
		c.Client.ZIncrBy(ctx, common.Viewed, -1, item)
	}
}

// RescaleViewed 定期调整商品总浏览数集合
func (c *Cache) RescaleViewed(ctx context.Context) {
	for !common.QUIT {
		c.Client.ZRemRangeByRank(ctx, common.Viewed, 0, -20001)
		c.Client.ZInterStore(ctx, common.Viewed, &redis.ZStore{
			Keys:      []string{common.Viewed},
			Weights:   []float64{0.5},
			Aggregate: "",
		})
		time.Sleep(300 * time.Second)
	}
}

// CleanSession 清理会话
func (c *Cache) CleanSession(ctx context.Context) {
	// 如果退出了就不再执行
	for !common.QUIT {
		// 检查是否到达限制，如果未达限制则休眠再重新检查
		size := c.Client.ZCard(ctx, common.Recent).Val()
		if size <= common.LIMIT {
			time.Sleep(1 * time.Second)
			continue
		}
		// 一次最多删除一百条
		endIndex := common.LIMIT
		if size < endIndex {
			endIndex = size
		}
		// 在最近访问中找出最早的一部分 token 记录
		tokens := c.Client.ZRange(ctx, common.Recent, 0, endIndex-1).Val()

		sessionKeys := make([]string, 0, len(tokens))
		for _, token := range tokens {
			sessionKeys = append(sessionKeys, common.ViewedPre+token)
		}

		c.Client.HDel(ctx, common.LoginHash, tokens...)
		c.Client.ZRem(ctx, common.Recent, tokens)
		c.Client.Del(ctx, sessionKeys...)
	}

	defer atomic.AddInt32(&common.FLAG, -1)

}

// AddToCart 增加商品到购物车
func (c *Cache) AddToCart(ctx context.Context, session, item string, count int) {
	switch {
	case count <= 0:
		c.Client.Del(ctx, common.CartPre+session)

	case count > 0:
		c.Client.HSet(ctx, common.CartPre+session, item, count)
	}
}

// CleanFullSession 清理包括购物车在内的会话
func (c *Cache) CleanFullSession(ctx context.Context) {
	// 如果退出了就不再执行
	for !common.QUIT {
		// 检查是否到达限制，如果未达限制则休眠再重新检查
		size := c.Client.ZCard(ctx, common.Recent).Val()
		if size <= common.LIMIT {
			time.Sleep(1 * time.Second)
			continue
		}
		// 一次最多删除一百条
		endIndex := common.LIMIT
		if size < endIndex {
			endIndex = size
		}
		// 在最近访问中找出最早的一部分 token 记录
		tokens := c.Client.ZRange(ctx, common.Recent, 0, endIndex-1).Val()

		sessionKeys := make([]string, 0, len(tokens))
		for _, token := range tokens {
			sessionKeys = append(sessionKeys, common.ViewedPre+token)
			sessionKeys = append(sessionKeys, common.CartPre+token)

		}

		c.Client.HDel(ctx, common.LoginHash, tokens...)
		c.Client.ZRem(ctx, common.Recent, tokens)
		c.Client.Del(ctx, sessionKeys...)

	}

	defer atomic.AddInt32(&common.FLAG, -1)
}

// CacheRequest 带缓存的请求处理
func (c *Cache) CacheRequest(ctx context.Context, req string, callback func(string) string) string {
	// 1. 判断该请求是否可以缓存，如果不行直接调用相应的处理函数
	if !c.CanCache(ctx, req) {
		return callback(req)
	}
	// 2. 如果可以缓存就返回缓存结果，如果缓存没有就处理后加入到缓存中
	pageKey := common.ReqCachePre + req
	content := c.Client.Get(ctx, pageKey).Val()

	if content == "" {
		res := callback(req)
		c.Client.Set(ctx, pageKey, res, 300*time.Second)
	}
	return content
}

// CanCache 判断请求是否能够缓存
func (c *Cache) CanCache(ctx context.Context, req string) bool {
	// 1. 提取请求的 item id
	itemID := extractItemId(req)
	// 2. 如果 item id 为空或者请求是动态的（内容根据用户不同而变化）则不能缓存
	if itemID == "" || isDynamic(req) {
		return false
	}

	// 3. 判断该 item 是否在总的浏览过的商品中且是总浏览量前 10000
	rank := c.Client.ZRank(ctx, common.Viewed, itemID).Val()

	return rank != 0 && rank < 10000
}

func extractItemId(request string) string {
	parsed, _ := url.Parse(request)
	queryValue, _ := url.ParseQuery(parsed.RawQuery)
	query := queryValue.Get("item")
	return query
}

func isDynamic(req string) bool {
	// 1. 将请求解析为结构体
	parsed, _ := url.Parse(req)
	// 2. 解析查询参数
	queryValue, _ := url.ParseQuery(parsed.RawQuery)
	// 3. 如果查询参数中有特定标记 _ 说明是动态的
	for _, v := range queryValue {
		for _, j := range v {
			if strings.Contains(j, "_") {
				return true
			}
		}
	}
	return false
}

func hashRequest(req string) string {
	hash := crypto.MD5.New()
	hash.Write([]byte(req))
	res := hash.Sum(nil)
	return hex.EncodeToString(res)
}

// 调度有序集合和延迟哈希集合缓存
func (c *Cache) ScheduleRowCache(ctx context.Context, rowID string, delay int64) {
	c.Client.ZAdd(ctx, common.Schedule, redis.Z{
		Score:  float64(time.Now().Unix() + delay),
		Member: rowID,
	})
	c.Client.HSet(ctx, common.Schedule, rowID, delay)
}

func (c *Cache) CacheRows(ctx context.Context) {
	for !common.QUIT {
		// 得到下一次的更新缓存时间
		next := c.Client.ZRangeWithScores(ctx, common.Schedule, 0, 0).Val()
		now := time.Now().Unix()
		if len(next) == 0 || next[0].Score > float64(now) {
			time.Sleep(50 * time.Millisecond)
			continue
		}

		rowID := next[0].Member.(string)
		delayStr := c.Client.HGet(ctx, common.Delay, rowID).Val()
		delay, _ := strconv.ParseFloat(delayStr, 10)

		if delay <= 0 {
			c.Client.ZRem(ctx, common.Schedule, rowID)
			c.Client.HDel(ctx, common.Delay, rowID)
			c.Client.Del(ctx, common.RowCachePre+rowID)
			continue
		}

		row := c.GetDataFromDB(rowID)

		c.Client.ZAdd(ctx, common.Schedule, redis.Z{Member: rowID, Score: float64(now) + delay})
		jsonRow, err := json.Marshal(row)
		if err != nil {
			log.Fatalf("marshal json failed, data is: %v, err is: %v\n", row, err)
		}
		c.Client.Set(ctx, common.RowCachePre+rowID, jsonRow, 0)

	}
	defer atomic.AddInt32(&common.FLAG, -1)
}

func (c *Cache) GetDataFromDB(rowID string) string {
	return "1111"
}
