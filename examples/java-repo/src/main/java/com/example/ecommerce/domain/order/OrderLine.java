package com.example.ecommerce.domain.order;

public class OrderLine {

    private final long productId;
    private final String productName;
    private final int quantity;
    private final Money unitPrice;

    public OrderLine(long productId, String productName, int quantity, Money unitPrice) {
        if (productName == null || productName.isEmpty()) {
            throw new IllegalArgumentException("product name must not be empty");
        }
        if (quantity <= 0) {
            throw new IllegalArgumentException("quantity must be positive");
        }
        if (unitPrice == null) {
            throw new IllegalArgumentException("unit price must not be null");
        }
        this.productId = productId;
        this.productName = productName;
        this.quantity = quantity;
        this.unitPrice = unitPrice;
    }

    public long getProductId() {
        return productId;
    }

    public String getProductName() {
        return productName;
    }

    public int getQuantity() {
        return quantity;
    }

    public Money getUnitPrice() {
        return unitPrice;
    }

    public Money lineTotal() {
        return unitPrice.multiply(quantity);
    }
}
