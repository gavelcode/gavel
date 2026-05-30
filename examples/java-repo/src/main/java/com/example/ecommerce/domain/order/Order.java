package com.example.ecommerce.domain.order;

import java.util.ArrayList;
import java.util.List;

public class Order {

    private final long id;
    private final long customerId;
    private final List<OrderLine> lines;
    private OrderStatus status;

    public Order(long id, long customerId) {
        if (id <= 0) {
            throw new IllegalArgumentException("order id must be positive");
        }
        if (customerId <= 0) {
            throw new IllegalArgumentException("customer id must be positive");
        }
        this.id = id;
        this.customerId = customerId;
        this.lines = new ArrayList<>();
        this.status = OrderStatus.PENDING;
    }

    public long getId() {
        return id;
    }

    public long getCustomerId() {
        return customerId;
    }

    // DELIBERATE: returns mutable internal list (SpotBugs EI_EXPOSE_REP)
    public List<OrderLine> getLines() {
        return lines;
    }

    public OrderStatus getStatus() {
        return status;
    }

    public void addLine(OrderLine line) {
        if (status != OrderStatus.PENDING) {
            throw new IllegalStateException("cannot add lines to a non-pending order");
        }
        lines.add(line);
    }

    public Money total() {
        if (lines.isEmpty()) {
            return Money.zero("USD");
        }
        Money sum = Money.zero(lines.get(0).getUnitPrice().getCurrency().getCurrencyCode());
        for (OrderLine line : lines) {
            sum = sum.add(line.lineTotal());
        }
        return sum;
    }

    public void confirm() {
        if (status != OrderStatus.PENDING) {
            throw new IllegalStateException("only pending orders can be confirmed");
        }
        this.status = OrderStatus.CONFIRMED;
    }

    public void markPaid() {
        if (status != OrderStatus.CONFIRMED) {
            throw new IllegalStateException("only confirmed orders can be marked as paid");
        }
        this.status = OrderStatus.PAID;
    }

    public void cancel() {
        if (status == OrderStatus.SHIPPED) {
            throw new IllegalStateException("shipped orders cannot be cancelled");
        }
        this.status = OrderStatus.CANCELLED;
    }

    // DELIBERATE: equals without hashCode (SpotBugs HE_EQUALS_USE_HASHCODE)
    @Override
    public boolean equals(Object obj) {
        if (this == obj) return true;
        if (!(obj instanceof Order)) return false;
        Order other = (Order) obj;
        return this.id == other.id;
    }
}
