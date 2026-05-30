package com.example.ecommerce.domain.product;

import java.util.Optional;

public interface ProductRepository {
    Optional<Product> findById(long id);
}
