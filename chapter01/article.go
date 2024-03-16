package chapter01

import (
	"context"
	"strconv"
	"time"

	"redis-practice/common"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

type Article interface {
	PostArticle(context.Context, string, string, string) string
	ArticleUpVote(context.Context, string, string)
	ArticleDownVote(context.Context, string, string)
	GetArticles(context.Context, int64, string) []map[string]string
	AddRemoveGroups(context.Context, string, []string, []string)
	GetGroupArticle(context.Context, string, string, int64) []map[string]string
	Reset()
}

type ArticleRepo struct {
	client *common.Client
}

func NewArticleRepo(conn *common.Client) *ArticleRepo {
	return &ArticleRepo{conn}
}

// PostArticle post article，return new article id
func (r ArticleRepo) PostArticle(ctx context.Context, author string, title string, link string) string {
	articleID := strconv.Itoa(int(r.client.Incr(ctx, common.ArticleID).Val()))

	// voted set of the article
	votedSetKey := common.VotedSetPre + articleID
	r.client.SAdd(ctx, votedSetKey, author)
	r.client.Expire(ctx, votedSetKey, common.OneWeekInSeconds*time.Second)

	now := time.Now().Unix()

	// attributes hash set of article
	r.client.HSet(ctx, common.ArticleHashSetPre+articleID, map[string]any{
		"id":     articleID,
		"title":  title,
		"link":   link,
		"poster": author,
		"time":   now,
		"votes":  0,
	})

	// score sorted set
	r.client.ZAdd(ctx, common.Score, redis.Z{
		Score:  float64(now),
		Member: articleID,
	})
	// time sorted set
	r.client.ZAdd(ctx, common.Time, redis.Z{
		Score:  float64(now),
		Member: articleID,
	})

	return articleID
}

func (r ArticleRepo) ArticleUpVote(ctx context.Context, articleID, user string) {
	// voting deadline
	cutoff := r.client.ZScore(ctx, common.Time, articleID).Val() + common.OneWeekInSeconds
	if time.Now().Unix() > int64(cutoff) {
		logrus.Info("Voting has closed...")
		return
	}

	if r.client.SAdd(ctx, common.VotedSetPre+articleID, user).Val() != 0 {
		r.client.ZIncrBy(ctx, common.Score, common.VotedScore, articleID)
		r.client.HIncrBy(ctx, common.ArticleHashSetPre+articleID, "votes", 1)
	}

}

func (r ArticleRepo) ArticleDownVote(ctx context.Context, articleID, user string) {
	// voting deadline
	cutoff := r.client.ZScore(ctx, common.Time, articleID).Val() + common.OneWeekInSeconds
	if time.Now().Unix() > int64(cutoff) {
		logrus.Info("Voting has closed...")
		return
	}

	if r.client.SAdd(ctx, common.VotedSetPre+articleID, user).Val() != 0 {
		r.client.ZIncrBy(ctx, common.Score, -common.VotedScore, articleID)
		r.client.HIncrBy(ctx, common.ArticleHashSetPre+articleID, "votes", -1)
	}
}

func (r ArticleRepo) GetArticles(ctx context.Context) ([]map[string]string, error) {
	members, err := r.client.ZRevRange(ctx, common.Score, 0, -1).Result()
	if err != nil {
		return nil, err
	}

	articles := make([]map[string]string, 0, len(members))
	for _, articleID := range members {
		articles = append(articles, r.client.HGetAll(ctx, common.ArticleHashSetPre+articleID).Val())
	}

	return articles, nil
}

func (r ArticleRepo) AddRemoveGroups(s string, strings []string, strings2 []string) {
	//TODO implement me
	panic("implement me")
}

func (r ArticleRepo) GetGroupArticle(s string, s2 string, i int64) []map[string]string {
	//TODO implement me
	panic("implement me")
}

func (r ArticleRepo) Reset(ctx context.Context) {
	r.client.FlushDB(ctx)
}
