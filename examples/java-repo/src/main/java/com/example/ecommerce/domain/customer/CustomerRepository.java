package com.example.ecommerce.domain.customer;

import java.util.Optional;

public interface CustomerRepository {
    void save(Customer customer);
    Optional<Customer> findById(long id);
}
