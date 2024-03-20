package service

import (
	"context"
	"fmt"

	"redis-practice/chapter01"
	"redis-practice/chapter01/dto"
)

type ArticleService struct {
	repo *chapter01.ArticleRepo
}

// NewArticleService creates a new ArticleService.
func NewArticleService(repo *chapter01.ArticleRepo) *ArticleService {
	return &ArticleService{repo: repo}
}

// PostArticle handles the business logic for posting a new article.
func (s *ArticleService) PostArticle(ctx context.Context, user, title, link string) (string, error) {
	// 这里可以添加业务逻辑，比如验证输入数据的有效性

	// 调用DAO层发布文章
	articleID := s.repo.PostArticle(ctx, user, title, link)
	if articleID == "" {
		return "", fmt.Errorf("failed to post article")
	}
	return articleID, nil
}

// VoteArticle handles the logic for voting on an article.
func (s *ArticleService) VoteArticle(ctx context.Context, articleID, userID string) error {
	// 可以在这里添加业务逻辑，比如检查用户是否已经投票

	// 调用DAO层进行投票
	s.repo.ArticleUpVote(ctx, "article:"+articleID, userID)
	return nil
}

// GetArticles handles fetching articles with pagination and sorting.
func (s *ArticleService) GetArticles(ctx context.Context) ([]dto.Article, error) {
	// 调用DAO层获取文章列表
	articlesData, err := s.repo.GetArticles(ctx)
	if err != nil {
		return nil, err
	}
	// 转换数据到DTO对象，这里简单模拟转换过程
	var articles []dto.Article
	for _, data := range articlesData {
		articles = append(articles, dto.Article{
			Title: data["title"],
			Link:  data["link"],
			// 根据需要填充其他字段
		})
	}

	return articles, nil
}
