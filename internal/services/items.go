package services

import (
	"backend-test-golang/internal/models"
	errs "backend-test-golang/pkg/errors"
	"backend-test-golang/pkg/skinport"
	"context"
	"errors"
	"fmt"
	"log"
)

const (
	skinportItemsCacheKey = "skinport:items"
)

func (s *Service) GetItems(ctx context.Context) ([]*models.ItemResponse, error) {
	cachedItems, err := s.getCachedItems()
	if err == nil {
		return cachedItems, nil
	}

	if !errors.Is(err, errs.ErrNotFound) {
		log.Printf("[WARN] GetItems: failed to get cached items: %v\n", err)
	}

	tradableItems, err := s.skinportClient.GetItems(ctx, true)
	if err != nil {
		log.Printf("[ERROR] GetItems: failed to get tradable items: %v\n", err)
		return nil, err
	}

	nonTradableItems, err := s.skinportClient.GetItems(ctx, false)
	if err != nil {
		log.Printf("[ERROR] GetItems: failed to get non-tradable items: %v\n", err)
		return nil, err
	}

	items := mergeItems(tradableItems, nonTradableItems)

	s.cacheItems(items)

	return items, nil
}

func mergeItems(tradableItems, nonTradableItems []skinport.Item) []*models.ItemResponse {
	itemResponseMap := make(map[string]*models.ItemResponse)

	for _, item := range tradableItems {
		itemResponseMap[item.MarketHashName] = &models.ItemResponse{
			MarketHashName:   item.MarketHashName,
			Currency:         item.Currency,
			MinPriceTradable: item.MinPrice,
		}
	}

	for _, item := range nonTradableItems {
		_, exists := itemResponseMap[item.MarketHashName]
		if exists {
			itemResponseMap[item.MarketHashName].MinPriceNonTradable = item.MinPrice
			continue
		}

		itemResponseMap[item.MarketHashName] = &models.ItemResponse{
			MarketHashName:      item.MarketHashName,
			Currency:            item.Currency,
			MinPriceNonTradable: item.MinPrice,
		}
	}

	itemResponses := make([]*models.ItemResponse, 0, len(itemResponseMap))
	for _, item := range itemResponseMap {
		itemResponses = append(itemResponses, item)
	}

	return itemResponses
}

func (s *Service) getCachedItems() ([]*models.ItemResponse, error) {
	entry, err := s.cache.Get(skinportItemsCacheKey)
	if err != nil {
		return nil, err
	}

	result, ok := entry.([]*models.ItemResponse)
	if !ok {
		return nil, fmt.Errorf("failed to get items: %w", errs.ErrInvalidCacheEntry)
	}

	return result, nil
}

func (s *Service) cacheItems(items []*models.ItemResponse) {
	s.cache.Set(skinportItemsCacheKey, items, s.defaultCacheTTL)
}
