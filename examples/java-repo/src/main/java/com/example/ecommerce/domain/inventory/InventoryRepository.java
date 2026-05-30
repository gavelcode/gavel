package com.example.ecommerce.domain.inventory;

import java.util.Optional;

public interface InventoryRepository {
    void save(Stock stock);
    Optional<Stock> findByProductId(long productId);
}
