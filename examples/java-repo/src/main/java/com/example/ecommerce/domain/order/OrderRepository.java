package com.example.ecommerce.domain.order;

import java.util.Optional;

public interface OrderRepository {
    void save(Order order);
    Optional<Order> findById(long id);
    long nextId();
}
