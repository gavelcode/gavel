package com.example.ecommerce.domain.order;

import static org.junit.Assert.*;

import java.math.BigDecimal;
import org.junit.Test;

public class MoneyTest {

    @Test
    public void shouldCreateMoney() {
        Money money = Money.of(10.50, "USD");
        assertEquals(new BigDecimal("10.5"), money.getAmount());
        assertEquals("USD", money.getCurrency().getCurrencyCode());
    }

    @Test
    public void shouldCreateZeroMoney() {
        Money zero = Money.zero("EUR");
        assertTrue(zero.isZero());
        assertFalse(zero.isNegative());
    }

    @Test
    public void shouldAddSameCurrency() {
        Money a = Money.of(10.00, "USD");
        Money b = Money.of(5.50, "USD");
        Money sum = a.add(b);
        assertEquals(new BigDecimal("15.5"), sum.getAmount());
    }

    @Test
    public void shouldSubtractSameCurrency() {
        Money a = Money.of(10.00, "USD");
        Money b = Money.of(3.00, "USD");
        Money diff = a.subtract(b);
        assertEquals(new BigDecimal("7.0"), diff.getAmount());
    }

    @Test
    public void shouldMultiplyByQuantity() {
        Money price = Money.of(29.99, "USD");
        Money total = price.multiply(3);
        assertEquals(new BigDecimal("89.97"), total.getAmount());
    }

    @Test(expected = IllegalArgumentException.class)
    public void shouldRejectDifferentCurrencies() {
        Money usd = Money.of(10.00, "USD");
        Money eur = Money.of(5.00, "EUR");
        usd.add(eur);
    }

    @Test
    public void shouldDetectNegativeAmount() {
        Money a = Money.of(5.00, "USD");
        Money b = Money.of(10.00, "USD");
        Money diff = a.subtract(b);
        assertTrue(diff.isNegative());
    }

    @Test
    public void shouldFormatToString() {
        Money money = Money.of(42.00, "USD");
        assertTrue(money.toString().startsWith("42"));
    }
}
