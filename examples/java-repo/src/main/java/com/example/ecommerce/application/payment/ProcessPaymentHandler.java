package com.example.ecommerce.application.payment;

import com.example.ecommerce.domain.order.Order;
import com.example.ecommerce.domain.order.OrderRepository;
import com.example.ecommerce.domain.payment.Payment;
import com.example.ecommerce.domain.payment.PaymentRepository;
import java.util.Optional;

public class ProcessPaymentHandler {

    private final PaymentRepository paymentRepo;
    private final OrderRepository orderRepo;

    public ProcessPaymentHandler(PaymentRepository paymentRepo, OrderRepository orderRepo) {
        this.paymentRepo = paymentRepo;
        this.orderRepo = orderRepo;
    }

    // DELIBERATE: duplicated validation block with PlaceOrderHandler (CPD)
    public PaymentDto execute(long orderId, String paymentMethod) {
        Optional<Order> orderOpt = orderRepo.findById(orderId);
        if (!orderOpt.isPresent()) {
            throw new IllegalArgumentException("order not found: " + orderId);
        }
        Order order = orderOpt.get();

        if (paymentMethod == null || paymentMethod.isBlank()) {
            throw new IllegalArgumentException("payment method must not be blank");
        }

        Optional<Payment> existingPayment = paymentRepo.findByOrderId(orderId);
        if (existingPayment.isPresent()) {
            throw new IllegalStateException("payment already exists for order: " + orderId);
        }

        long paymentId = paymentRepo.nextId();
        Payment payment = new Payment(paymentId, order.getId(), order.total(), paymentMethod);

        payment.process();

        boolean success = processWithGateway(payment);
        if (success) {
            payment.complete();
            order.markPaid();
            orderRepo.save(order);
        } else {
            payment.fail();
        }

        paymentRepo.save(payment);
        return PaymentDto.fromDomain(payment);
    }

    private boolean processWithGateway(Payment payment) {
        return !payment.getAmount().isNegative();
    }
}
