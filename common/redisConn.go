package common

import (
	"context"
	"errors"
	"fmt"
	"regexp"

	goversion "github.com/hashicorp/go-version"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	REDIS_MIN_VERSION = "5.0.0"
)

type RedisConf struct {
	Addr     string
	Password string
	DB       int
}

type Client struct {
	*redis.Client
}

func NewClient(c *redis.Client) *Client {
	return &Client{c}
}

func ConnectRedis(ctx context.Context, conf *RedisConf) *redis.Client {
	conn := redis.NewClient(&redis.Options{
		Addr:     conf.Addr,
		Password: conf.Password,
		DB:       conf.DB,
	})

	if conn == nil {
		panic("redis connect failed")
	}

	if _, err := conn.Ping(ctx).Result(); err != nil {
		logrus.Error("ping failed，error: ", err)
		return nil
	}

	if _, err := checkServerVersion(ctx, conn); err != nil {
		logrus.Error("check server version not passed : ", err)
		return nil
	}

	return conn
}

func checkServerVersion(ctx context.Context, conn *redis.Client) (string, error) {
	cmd := conn.Info(ctx, "server")

	serverInfo := cmd.Val()
	if len(serverInfo) == 0 {
		return "", errors.New(fmt.Sprintf("Get server info failed, err : %v", cmd.Err()))
	}

	logrus.Info("server info : ", serverInfo)

	r := regexp.MustCompile(`redis_version:((\d+\.)+\d+)`)
	matchSlice := r.FindStringSubmatch(serverInfo)
	if len(matchSlice) < 2 {
		return "", errors.New("Regexp not match redis_version")
	}
	ver := matchSlice[1]

	return checkVersion(ver, REDIS_MIN_VERSION)
}

func checkVersion(serverVer, minVer string) (string, error) {
	serverV, err := goversion.NewVersion(serverVer)
	if err != nil {
		return "", err
	}

	minV, err := goversion.NewVersion(minVer)
	if err != nil {
		return "", err
	}

	if serverV.LessThan(minV) {
		return "", errors.New(fmt.Sprintf("The redis server version %v is too early, need greater than %v", serverVer, minVer))
	}

	return serverVer, nil
}
