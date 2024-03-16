package chapter01

import (
	"context"
	"os"
	"testing"

	"redis-practice"
	"redis-practice/common"

	"github.com/stretchr/testify/assert"
)

var client *common.Client
var articleRepo *ArticleRepo
var ctx context.Context

func TestMain(m *testing.M) {
	ctx = context.Background()

	conn := common.ConnectRedis(ctx, &common.RedisConf{
		Addr:     redis_practice.Addr,
		Password: redis_practice.Password,
		DB:       redis_practice.DB,
	})

	defer conn.Close()

	client = common.NewClient(conn)
	articleRepo = &ArticleRepo{client: client}

	code := m.Run()

	os.Exit(code)
}

func TestChapter01(t *testing.T) {
	articleID := articleRepo.PostArticle(ctx, "hualulu", "hualubang up up", "http://www.hualubang.com/1")
	t.Log("posted a new article with id :", articleID)
	assert.Equal(t, "1", articleID, "should be `1`")

	articleID = articleRepo.PostArticle(ctx, "guadandan", "i love hualubang", "http://www.hualubang.com/2")
	t.Log("posted a new article with id :", articleID)
	assert.Equal(t, "2", articleID, "should be `2`")

	articles, _ := articleRepo.GetArticles(ctx)
	assert.EqualValues(t, 2, len(articles), "articles number should be 2")

	articleRepo.ArticleUpVote(ctx, "1", "lurenjia")
	articleRepo.ArticleUpVote(ctx, "1", "lurenyi")

	articles, _ = articleRepo.GetArticles(ctx)
	t.Log(articles)

	assert.Equal(t, "2", articles[0]["votes"], "votes should be 2 of `hualubang up up`")

	defer articleRepo.Reset(ctx)
}
