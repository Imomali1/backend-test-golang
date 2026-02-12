package services

import (
	"testing"

	"backend-test-golang/internal/models"
	"backend-test-golang/pkg/skinport"
)

func TestMergeItems(t *testing.T) {
	t.Run("merge items with both tradable and non-tradable", func(t *testing.T) {
		tradablePrice := 25.99
		nonTradablePrice := 23.50

		tradableItems := []skinport.Item{
			{
				MarketHashName: "AK-47 | Redline (Field-Tested)",
				Currency:       "EUR",
				MinPrice:       &tradablePrice,
			},
		}

		nonTradableItems := []skinport.Item{
			{
				MarketHashName: "AK-47 | Redline (Field-Tested)",
				Currency:       "EUR",
				MinPrice:       &nonTradablePrice,
			},
		}

		result := mergeItems(tradableItems, nonTradableItems)

		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}

		item := result[0]
		if item.MarketHashName != "AK-47 | Redline (Field-Tested)" {
			t.Errorf("expected name 'AK-47 | Redline (Field-Tested)', got '%s'", item.MarketHashName)
		}
		if item.MinPriceTradable == nil || *item.MinPriceTradable != tradablePrice {
			t.Errorf("expected tradable price %.2f, got %v", tradablePrice, item.MinPriceTradable)
		}
		if item.MinPriceNonTradable == nil || *item.MinPriceNonTradable != nonTradablePrice {
			t.Errorf("expected non-tradable price %.2f, got %v", nonTradablePrice, item.MinPriceNonTradable)
		}
	})

	t.Run("item exists only in tradable list", func(t *testing.T) {
		tradablePrice := 85.00

		tradableItems := []skinport.Item{
			{
				MarketHashName: "AWP | Asiimov (Field-Tested)",
				Currency:       "EUR",
				MinPrice:       &tradablePrice,
			},
		}

		var nonTradableItems []skinport.Item

		result := mergeItems(tradableItems, nonTradableItems)

		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}

		item := result[0]
		if item.MinPriceTradable == nil || *item.MinPriceTradable != tradablePrice {
			t.Errorf("expected tradable price %.2f, got %v", tradablePrice, item.MinPriceTradable)
		}
		if item.MinPriceNonTradable != nil {
			t.Errorf("expected non-tradable price to be nil, got %v", *item.MinPriceNonTradable)
		}
	})

	t.Run("item exists only in non-tradable list", func(t *testing.T) {
		nonTradablePrice := 45.00

		var tradableItems []skinport.Item

		nonTradableItems := []skinport.Item{
			{
				MarketHashName: "M4A4 | Howl (Factory New)",
				Currency:       "EUR",
				MinPrice:       &nonTradablePrice,
			},
		}

		result := mergeItems(tradableItems, nonTradableItems)

		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}

		item := result[0]
		if item.MinPriceTradable != nil {
			t.Errorf("expected tradable price to be nil, got %v", *item.MinPriceTradable)
		}
		if item.MinPriceNonTradable == nil || *item.MinPriceNonTradable != nonTradablePrice {
			t.Errorf("expected non-tradable price %.2f, got %v", nonTradablePrice, item.MinPriceNonTradable)
		}
	})

	t.Run("merge multiple different items", func(t *testing.T) {
		price1T := 25.99
		price2T := 85.00
		price2NT := 82.00
		price3NT := 45.00

		tradableItems := []skinport.Item{
			{MarketHashName: "Item 1", Currency: "EUR", MinPrice: &price1T},
			{MarketHashName: "Item 2", Currency: "EUR", MinPrice: &price2T},
		}

		nonTradableItems := []skinport.Item{
			{MarketHashName: "Item 2", Currency: "EUR", MinPrice: &price2NT},
			{MarketHashName: "Item 3", Currency: "EUR", MinPrice: &price3NT},
		}

		result := mergeItems(tradableItems, nonTradableItems)

		if len(result) != 3 {
			t.Fatalf("expected 3 items, got %d", len(result))
		}

		itemMap := make(map[string]*models.ItemResponse)
		for _, item := range result {
			itemMap[item.MarketHashName] = item
		}

		if _, exists := itemMap["Item 1"]; !exists {
			t.Error("Item 1 not found in result")
		}
		if _, exists := itemMap["Item 2"]; !exists {
			t.Error("Item 2 not found in result")
		}
		if _, exists := itemMap["Item 3"]; !exists {
			t.Error("Item 3 not found in result")
		}

		item2 := itemMap["Item 2"]
		if item2.MinPriceTradable == nil || item2.MinPriceNonTradable == nil {
			t.Error("Item 2 should have both tradable and non-tradable prices")
		}
	})

	t.Run("empty lists", func(t *testing.T) {
		result := mergeItems([]skinport.Item{}, []skinport.Item{})

		if len(result) != 0 {
			t.Errorf("expected 0 items for empty lists, got %d", len(result))
		}
	})

	t.Run("nil prices are handled", func(t *testing.T) {
		tradableItems := []skinport.Item{
			{
				MarketHashName: "Test Item",
				Currency:       "EUR",
				MinPrice:       nil,
			},
		}

		var nonTradableItems []skinport.Item

		result := mergeItems(tradableItems, nonTradableItems)

		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}

		item := result[0]
		if item.MinPriceTradable != nil {
			t.Errorf("expected nil tradable price, got %v", *item.MinPriceTradable)
		}
	})
}

func TestMergeItems_Consistency(t *testing.T) {
	t.Run("currency and name are preserved", func(t *testing.T) {
		price := 100.00
		tradableItems := []skinport.Item{
			{
				MarketHashName: "Special Item",
				Currency:       "USD",
				MinPrice:       &price,
			},
		}

		result := mergeItems(tradableItems, []skinport.Item{})

		if len(result) != 1 {
			t.Fatalf("expected 1 item, got %d", len(result))
		}

		if result[0].MarketHashName != "Special Item" {
			t.Errorf("expected name 'Special Item', got '%s'", result[0].MarketHashName)
		}
		if result[0].Currency != "USD" {
			t.Errorf("expected currency 'USD', got '%s'", result[0].Currency)
		}
	})
}
