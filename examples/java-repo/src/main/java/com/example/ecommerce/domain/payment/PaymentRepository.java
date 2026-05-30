package com.example.ecommerce.domain.payment;

import java.util.Optional;

public interface PaymentRepository {
    void save(Payment payment);
    Optional<Payment> findById(long id);
    Optional<Payment> findByOrderId(long orderId);
    long nextId();
}
