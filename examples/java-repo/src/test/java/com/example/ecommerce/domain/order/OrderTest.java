package com.example.ecommerce.domain.order;

import static org.junit.Assert.*;

import org.junit.Test;

public class OrderTest {

    @Test
    public void shouldCreatePendingOrder() {
        Order order = new Order(1, 100);
        assertEquals(1, order.getId());
        assertEquals(100, order.getCustomerId());
        assertEquals(OrderStatus.PENDING, order.getStatus());
        assertTrue(order.getLines().isEmpty());
    }

    @Test(expected = IllegalArgumentException.class)
    public void shouldRejectZeroId() {
        new Order(0, 100);
    }

    @Test(expected = IllegalArgumentException.class)
    public void shouldRejectZeroCustomerId() {
        new Order(1, 0);
    }

    @Test
    public void shouldAddLineAndCalculateTotal() {
        Order order = new Order(1, 100);
        OrderLine line1 = new OrderLine(1, "Laptop", 2, Money.of(999.99, "USD"));
        OrderLine line2 = new OrderLine(2, "Mouse", 3, Money.of(29.99, "USD"));

        order.addLine(line1);
        order.addLine(line2);

        assertEquals(2, order.getLines().size());
        Money total = order.total();
        assertNotNull(total);
    }

    @Test
    public void shouldConfirmPendingOrder() {
        Order order = new Order(1, 100);
        order.addLine(new OrderLine(1, "Laptop", 1, Money.of(999.99, "USD")));
        order.confirm();
        assertEquals(OrderStatus.CONFIRMED, order.getStatus());
    }

    @Test(expected = IllegalStateException.class)
    public void shouldNotConfirmNonPendingOrder() {
        Order order = new Order(1, 100);
        order.addLine(new OrderLine(1, "Laptop", 1, Money.of(999.99, "USD")));
        order.confirm();
        order.confirm();
    }

    @Test
    public void shouldMarkConfirmedOrderAsPaid() {
        Order order = new Order(1, 100);
        order.addLine(new OrderLine(1, "Laptop", 1, Money.of(999.99, "USD")));
        order.confirm();
        order.markPaid();
        assertEquals(OrderStatus.PAID, order.getStatus());
    }

    @Test
    public void shouldCancelPendingOrder() {
        Order order = new Order(1, 100);
        order.cancel();
        assertEquals(OrderStatus.CANCELLED, order.getStatus());
    }

    @Test
    public void shouldReturnZeroTotalForEmptyOrder() {
        Order order = new Order(1, 100);
        Money total = order.total();
        assertTrue(total.isZero());
    }
}
