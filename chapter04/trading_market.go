package chapter04

import (
	"context"
	"errors"
	"fmt"
	"time"

	"redis-practice/common"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Cache struct {
	Client *common.Client
}

// Sell 用户将商品放入买卖市场
func (c *Cache) Sell(ctx context.Context, user string, goods string, price float64) bool {
	bag := common.UserBagPre + user           // 卖家背包
	item := fmt.Sprintf("%s:%s", user, goods) // 要卖出放入市场的商品
	end := time.Now().Unix() + 5

	// 循环重试
	for time.Now().Unix() < end {
		// 监视背包库存
		err := c.Client.Watch(ctx, func(tx *redis.Tx) error {
			// 如果商品在背包中才放入到市场上
			if tx.SIsMember(ctx, bag, goods).Val() {
				_, err := tx.Pipelined(ctx, func(pipe redis.Pipeliner) error {
					// 将商品添加到市场有序集合中
					pipe.ZAdd(ctx, common.Market, redis.Z{
						Score:  price,
						Member: item,
					})
					pipe.SRem(ctx, bag, goods)
					return nil
				})
				return err
			}
			return errors.New("goods not exist")
		}, bag)

		if err == nil {
			return true
		}

		logrus.Infof("User %s attempt to sell goods %s to market failed, error: %v", user, goods, err)
	}

	return false
}

func (c *Cache) Purchase(ctx context.Context, seller, buyer string, goods string) bool {
	item := fmt.Sprintf("%s:%s", seller, goods) // 交易市场上的商品

	end := time.Now().Unix() + 5

	for time.Now().Unix() < end {
		err := c.Client.Watch(ctx, func(tx *redis.Tx) error {
			if _, err := tx.Pipelined(ctx, func(pipeline redis.Pipeliner) error {
				price := tx.ZScore(ctx, common.Market, item).Val()
				if funds, _ := tx.HGet(ctx, common.UserPre+buyer, "funds").Float64(); funds > price {
					pipeline.HIncrBy(ctx, common.UserPre+seller, "funds", int64(price))
					pipeline.HIncrBy(ctx, common.UserPre+buyer, "funds", -int64(price))
					pipeline.SAdd(ctx, common.UserBagPre+buyer, goods)
					pipeline.ZRem(ctx, common.Market, item)
					return nil
				}
				return errors.New("can't afford this item")
			}); err != nil {
				return err
			}

			return nil
		}, common.Market, common.UserPre+buyer)

		if err != nil {
			logrus.Info(err)
			return false
		}
	}

	return true
}
