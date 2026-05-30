import unittest

from ecommerce.domain.order.order import Order
from ecommerce.domain.order.order_line import OrderLine
from ecommerce.domain.order.money import Money
from ecommerce.domain.order import order_status


class TestOrder(unittest.TestCase):
    def test_create_pending_order(self):
        order = Order(1, 100)
        self.assertEqual(order.id, 1)
        self.assertEqual(order.customer_id, 100)
        self.assertEqual(order.status, order_status.PENDING)
        self.assertEqual(len(order.lines), 0)

    def test_reject_zero_id(self):
        with self.assertRaises(ValueError):
            Order(0, 100)

    def test_reject_zero_customer_id(self):
        with self.assertRaises(ValueError):
            Order(1, 0)

    def test_add_line_and_calculate_total(self):
        order = Order(1, 100)
        line1 = OrderLine(1, "Laptop", 2, Money.of(999.99, "USD"))
        line2 = OrderLine(2, "Mouse", 3, Money.of(29.99, "USD"))
        order.add_line(line1)
        order.add_line(line2)
        self.assertEqual(len(order.lines), 2)
        self.assertFalse(order.total().is_zero())

    def test_confirm_pending_order(self):
        order = Order(1, 100)
        order.add_line(OrderLine(1, "Laptop", 1, Money.of(999.99, "USD")))
        order.confirm()
        self.assertEqual(order.status, order_status.CONFIRMED)

    def test_cannot_confirm_non_pending(self):
        order = Order(1, 100)
        order.add_line(OrderLine(1, "Laptop", 1, Money.of(999.99, "USD")))
        order.confirm()
        with self.assertRaises(RuntimeError):
            order.confirm()

    def test_mark_paid(self):
        order = Order(1, 100)
        order.add_line(OrderLine(1, "Laptop", 1, Money.of(999.99, "USD")))
        order.confirm()
        order.mark_paid()
        self.assertEqual(order.status, order_status.PAID)

    def test_cancel_pending(self):
        order = Order(1, 100)
        order.cancel()
        self.assertEqual(order.status, order_status.CANCELLED)

    def test_empty_order_total_is_zero(self):
        order = Order(1, 100)
        self.assertTrue(order.total().is_zero())


if __name__ == "__main__":
    unittest.main()
