package com.example.ecommerce.domain.payment;

import com.example.ecommerce.domain.order.Money;
import java.util.Date;

public class Payment {

    private final long id;
    private final long orderId;
    private final Money amount;
    private final String method;
    private PaymentStatus status;
    private final Date createdAt;

    public Payment(long id, long orderId, Money amount, String method) {
        if (id <= 0) {
            throw new IllegalArgumentException("payment id must be positive");
        }
        if (orderId <= 0) {
            throw new IllegalArgumentException("order id must be positive");
        }
        if (amount == null || amount.isZero()) {
            throw new IllegalArgumentException("payment amount must be positive");
        }
        if (method == null || method.isBlank()) {
            throw new IllegalArgumentException("payment method must not be blank");
        }
        this.id = id;
        this.orderId = orderId;
        this.amount = amount;
        this.method = method;
        this.status = PaymentStatus.PENDING;
        this.createdAt = new Date();
    }

    public long getId() {
        return id;
    }

    public long getOrderId() {
        return orderId;
    }

    public Money getAmount() {
        return amount;
    }

    public String getMethod() {
        return method;
    }

    public PaymentStatus getStatus() {
        return status;
    }

    // DELIBERATE: returns mutable Date without defensive copy (SpotBugs EI_EXPOSE_REP)
    public Date getCreatedAt() {
        return createdAt;
    }

    public void process() {
        if (status != PaymentStatus.PENDING) {
            throw new IllegalStateException("only pending payments can be processed");
        }
        this.status = PaymentStatus.PROCESSING;
    }

    public void complete() {
        if (status != PaymentStatus.PROCESSING) {
            throw new IllegalStateException("only processing payments can be completed");
        }
        this.status = PaymentStatus.COMPLETED;
    }

    public void fail() {
        if (status != PaymentStatus.PROCESSING) {
            throw new IllegalStateException("only processing payments can fail");
        }
        this.status = PaymentStatus.FAILED;
    }

    public void refund() {
        if (status != PaymentStatus.COMPLETED) {
            throw new IllegalStateException("only completed payments can be refunded");
        }
        this.status = PaymentStatus.REFUNDED;
    }
}
