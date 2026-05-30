import unittest

from ecommerce.domain.payment.payment import Payment
from ecommerce.domain.payment import payment_status
from ecommerce.domain.order.money import Money


class TestPayment(unittest.TestCase):
    def test_create_pending_payment(self):
        payment = Payment(1, 100, Money.of(50.00, "USD"), "credit_card")
        self.assertEqual(payment.id, 1)
        self.assertEqual(payment.order_id, 100)
        self.assertEqual(payment.status, payment_status.PENDING)
        self.assertEqual(payment.method, "credit_card")
        self.assertIsNotNone(payment.created_at)

    def test_reject_zero_amount(self):
        with self.assertRaises(ValueError):
            Payment(1, 100, Money.zero("USD"), "credit_card")

    def test_reject_blank_method(self):
        with self.assertRaises(ValueError):
            Payment(1, 100, Money.of(50.00, "USD"), "")

    def test_process_and_complete(self):
        payment = Payment(1, 100, Money.of(50.00, "USD"), "credit_card")
        payment.process()
        self.assertEqual(payment.status, payment_status.PROCESSING)
        payment.complete()
        self.assertEqual(payment.status, payment_status.COMPLETED)

    def test_process_and_fail(self):
        payment = Payment(1, 100, Money.of(50.00, "USD"), "credit_card")
        payment.process()
        payment.fail()
        self.assertEqual(payment.status, payment_status.FAILED)

    def test_refund_completed(self):
        payment = Payment(1, 100, Money.of(50.00, "USD"), "credit_card")
        payment.process()
        payment.complete()
        payment.refund()
        self.assertEqual(payment.status, payment_status.REFUNDED)

    def test_cannot_process_non_pending(self):
        payment = Payment(1, 100, Money.of(50.00, "USD"), "credit_card")
        payment.process()
        with self.assertRaises(RuntimeError):
            payment.process()

    def test_cannot_refund_non_completed(self):
        payment = Payment(1, 100, Money.of(50.00, "USD"), "credit_card")
        with self.assertRaises(RuntimeError):
            payment.refund()


if __name__ == "__main__":
    unittest.main()
