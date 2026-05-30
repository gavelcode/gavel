package com.example.ecommerce.domain.payment;

import static org.junit.Assert.*;

import com.example.ecommerce.domain.order.Money;
import org.junit.Test;

public class PaymentTest {

    @Test
    public void shouldCreatePendingPayment() {
        Payment payment = new Payment(1, 100, Money.of(50.00, "USD"), "credit_card");
        assertEquals(1, payment.getId());
        assertEquals(100, payment.getOrderId());
        assertEquals(PaymentStatus.PENDING, payment.getStatus());
        assertEquals("credit_card", payment.getMethod());
        assertNotNull(payment.getCreatedAt());
    }

    @Test(expected = IllegalArgumentException.class)
    public void shouldRejectZeroAmount() {
        new Payment(1, 100, Money.zero("USD"), "credit_card");
    }

    @Test(expected = IllegalArgumentException.class)
    public void shouldRejectBlankMethod() {
        new Payment(1, 100, Money.of(50.00, "USD"), "");
    }

    @Test
    public void shouldProcessAndComplete() {
        Payment payment = new Payment(1, 100, Money.of(50.00, "USD"), "credit_card");
        payment.process();
        assertEquals(PaymentStatus.PROCESSING, payment.getStatus());
        payment.complete();
        assertEquals(PaymentStatus.COMPLETED, payment.getStatus());
    }

    @Test
    public void shouldProcessAndFail() {
        Payment payment = new Payment(1, 100, Money.of(50.00, "USD"), "credit_card");
        payment.process();
        payment.fail();
        assertEquals(PaymentStatus.FAILED, payment.getStatus());
    }

    @Test
    public void shouldRefundCompletedPayment() {
        Payment payment = new Payment(1, 100, Money.of(50.00, "USD"), "credit_card");
        payment.process();
        payment.complete();
        payment.refund();
        assertEquals(PaymentStatus.REFUNDED, payment.getStatus());
    }

    @Test(expected = IllegalStateException.class)
    public void shouldNotProcessNonPendingPayment() {
        Payment payment = new Payment(1, 100, Money.of(50.00, "USD"), "credit_card");
        payment.process();
        payment.process();
    }

    @Test(expected = IllegalStateException.class)
    public void shouldNotRefundNonCompletedPayment() {
        Payment payment = new Payment(1, 100, Money.of(50.00, "USD"), "credit_card");
        payment.refund();
    }
}
