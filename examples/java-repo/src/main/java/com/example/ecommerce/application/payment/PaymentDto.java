package com.example.ecommerce.application.payment;

import com.example.ecommerce.domain.payment.Payment;

public class PaymentDto {

    public final long id;
    public final long orderId;
    public final String amount;
    public final String method;
    public final String status;

    public PaymentDto(long id, long orderId, String amount, String method, String status) {
        this.id = id;
        this.orderId = orderId;
        this.amount = amount;
        this.method = method;
        this.status = status;
    }

    public static PaymentDto fromDomain(Payment payment) {
        return new PaymentDto(
            payment.getId(),
            payment.getOrderId(),
            payment.getAmount().toString(),
            payment.getMethod(),
            payment.getStatus().name());
    }
}
