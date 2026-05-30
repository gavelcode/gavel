package com.example.ecommerce.domain.product;

import com.example.ecommerce.domain.order.Money;

public class Product {

    private final long id;
    private final String name;
    private final String description;
    private final Money price;

    public Product(long id, String name, String description, Money price) {
        if (id <= 0) {
            throw new IllegalArgumentException("product id must be positive");
        }
        if (name == null || name.isBlank()) {
            throw new IllegalArgumentException("product name must not be blank");
        }
        if (price == null) {
            throw new IllegalArgumentException("price must not be null");
        }
        this.id = id;
        this.name = name;
        this.description = description != null ? description : "";
        this.price = price;
    }

    public long getId() {
        return id;
    }

    public String getName() {
        return name;
    }

    public String getDescription() {
        return description;
    }

    public Money getPrice() {
        return price;
    }
}
