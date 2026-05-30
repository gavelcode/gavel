package com.example.ecommerce.domain.customer;

// DELIBERATE: unused import (PMD UnusedImports)
import java.util.Collections;

public class Customer {

    private final long id;
    private String name;
    private String email;
    private Address shippingAddress;

    public Customer(long id, String name, String email) {
        if (id <= 0) {
            throw new IllegalArgumentException("customer id must be positive");
        }
        if (name == null || name.isBlank()) {
            throw new IllegalArgumentException("customer name must not be blank");
        }
        if (email == null || !email.contains("@")) {
            throw new IllegalArgumentException("invalid email");
        }
        this.id = id;
        this.name = name;
        this.email = email;
    }

    public long getId() {
        return id;
    }

    public String getName() {
        return name;
    }

    public String getEmail() {
        return email;
    }

    public Address getShippingAddress() {
        return shippingAddress;
    }

    public void updateName(String name) {
        if (name == null || name.isBlank()) {
            throw new IllegalArgumentException("name must not be blank");
        }
        this.name = name;
    }

    public void updateEmail(String email) {
        if (email == null || !email.contains("@")) {
            throw new IllegalArgumentException("invalid email");
        }
        this.email = email;
    }

    public void setShippingAddress(Address address) {
        this.shippingAddress = address;
    }
}
