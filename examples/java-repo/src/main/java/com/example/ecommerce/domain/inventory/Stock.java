package com.example.ecommerce.domain.inventory;

public class Stock {

    private final long productId;
    private int quantity;

    public Stock(long productId, int quantity) {
        if (productId <= 0) {
            throw new IllegalArgumentException("product id must be positive");
        }
        if (quantity < 0) {
            throw new IllegalArgumentException("stock quantity cannot be negative");
        }
        this.productId = productId;
        this.quantity = quantity;
    }

    public long getProductId() {
        return productId;
    }

    public int getQuantity() {
        return quantity;
    }

    public boolean hasEnough(int requested) {
        return quantity >= requested;
    }

    public void reserve(int amount) {
        if (amount <= 0) {
            throw new IllegalArgumentException("reserve amount must be positive");
        }
        if (!hasEnough(amount)) {
            throw new IllegalStateException(
                "insufficient stock for product " + productId + ": available=" + quantity + ", requested=" + amount);
        }
        this.quantity -= amount;
    }

    public void restock(int amount) {
        if (amount <= 0) {
            throw new IllegalArgumentException("restock amount must be positive");
        }
        this.quantity += amount;
    }
}
