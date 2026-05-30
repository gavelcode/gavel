import unittest
from decimal import Decimal

from ecommerce.domain.order.money import Money


class TestMoney(unittest.TestCase):
    def test_create_money(self):
        money = Money.of(10.50, "USD")
        self.assertEqual(money.amount, Decimal("10.5"))
        self.assertEqual(money.currency, "USD")

    def test_zero_money(self):
        zero = Money.zero("EUR")
        self.assertTrue(zero.is_zero())
        self.assertFalse(zero.is_negative())

    def test_add_same_currency(self):
        a = Money.of(10.00, "USD")
        b = Money.of(5.50, "USD")
        result = a.add(b)
        self.assertEqual(result.amount, Decimal("15.5"))

    def test_subtract_same_currency(self):
        a = Money.of(10.00, "USD")
        b = Money.of(3.00, "USD")
        result = a.subtract(b)
        self.assertEqual(result.amount, Decimal("7.0"))

    def test_multiply(self):
        price = Money.of(29.99, "USD")
        total = price.multiply(3)
        self.assertEqual(total.amount, Decimal("89.97"))

    def test_reject_different_currencies(self):
        usd = Money.of(10.00, "USD")
        eur = Money.of(5.00, "EUR")
        with self.assertRaises(ValueError):
            usd.add(eur)

    def test_negative_detection(self):
        a = Money.of(5.00, "USD")
        b = Money.of(10.00, "USD")
        diff = a.subtract(b)
        self.assertTrue(diff.is_negative())


if __name__ == "__main__":
    unittest.main()
