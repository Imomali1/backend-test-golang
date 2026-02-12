package services

import (
	"backend-test-golang/internal/repository"
	"backend-test-golang/pkg/cache"
	"backend-test-golang/pkg/skinport"
	"time"
)

type Service struct {
	defaultCacheTTL time.Duration
	cache           *cache.MemCache
	skinportClient  *skinport.Client
	repo            *repository.Repository
}

func New(cacheTTL int, cache *cache.MemCache, skinportClient *skinport.Client, repo *repository.Repository) *Service {
	return &Service{
		defaultCacheTTL: time.Duration(cacheTTL) * time.Second,
		cache:           cache,
		skinportClient:  skinportClient,
		repo:            repo,
	}
}
