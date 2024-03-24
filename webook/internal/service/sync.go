package service

import (
	"context"
	"webook/webook/internal/domain"
	"webook/webook/internal/repository"
)

type SyncService interface {
	InputArticle(ctx context.Context, article domain.Article) error
	InputUser(ctx context.Context, user domain.User) error
	InputAny(ctx context.Context, index, docID, data string) error
}

type syncService struct {
	userRepo    repository.UserRepository
	articleRepo repository.ArticleRepository
	likeRepo    repository.LikeRepository
	collectRepo repository.CollectRepository
	anyRepo     repository.AnyRepository
}

func (s *syncService) InputAny(ctx context.Context, index, docID, data string) error {
	return s.anyRepo.Input(ctx, index, docID, data)
}

func (s *syncService) InputArticle(ctx context.Context, article domain.Article) error {
	return s.articleRepo.InputArticle(ctx, article)
}

func (s *syncService) InputUser(ctx context.Context, user domain.User) error {
	return s.userRepo.InputUser(ctx, user)
}

func (s *syncService) InputLike(ctx context.Context, like domain.Like) error {
	return s.likeRepo.InputLike(ctx, like)
}

func (s *syncService) InputCollect(ctx context.Context, collect domain.Collect) error {
	return s.collectRepo.InputCollect(ctx, collect)
}

func NewSyncService(
	anyRepo repository.AnyRepository,
	userRepo repository.UserRepository,
	articleRepo repository.ArticleRepository,
	likeRepo repository.LikeRepository,
	collectRepo repository.CollectRepository) SyncService {
	return &syncService{
		userRepo:    userRepo,
		articleRepo: articleRepo,
		anyRepo:     anyRepo,
		likeRepo:    likeRepo,
		collectRepo: collectRepo,
	}
}
