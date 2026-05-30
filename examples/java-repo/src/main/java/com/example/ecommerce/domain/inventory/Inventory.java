package com.example.ecommerce.domain.inventory;

import java.util.HashMap;
import java.util.Map;

public class Inventory {

    private final Map<Long, Stock> stocks;

    public Inventory() {
        this.stocks = new HashMap<>();
    }

    public void addStock(Stock stock) {
        stocks.put(stock.getProductId(), stock);
    }

    public Stock getStock(long productId) {
        Stock stock = stocks.get(productId);
        if (stock == null) {
            throw new IllegalArgumentException("no stock record for product " + productId);
        }
        return stock;
    }

    public boolean isAvailable(long productId, int quantity) {
        Stock stock = stocks.get(productId);
        return stock != null && stock.hasEnough(quantity);
    }

    public void reserve(long productId, int quantity) {
        getStock(productId).reserve(quantity);
    }

    public void restock(long productId, int quantity) {
        getStock(productId).restock(quantity);
    }
}
