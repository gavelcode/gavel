from ecommerce.application.payment.payment_dto import PaymentDto
from ecommerce.domain.payment.payment import Payment


class ProcessPaymentHandler:
    def __init__(self, payment_repo, order_repo):
        self._payment_repo = payment_repo
        self._order_repo = order_repo

    def execute(self, order_id, payment_method):
        order = self._order_repo.find_by_id(order_id)
        if order is None:
            raise ValueError(f"order not found: {order_id}")

        if not payment_method or not payment_method.strip():
            raise ValueError("payment method must not be blank")

        existing = self._payment_repo.find_by_order_id(order_id)
        if existing is not None:
            raise RuntimeError(f"payment already exists for order: {order_id}")

        payment_id = self._payment_repo.next_id()
        payment = Payment(payment_id, order.id, order.total(), payment_method)

        payment.process()

        # DELIBERATE: deeply nested ifs (complexity)
        if payment.amount is not None:
            if not payment.amount.is_negative():
                if payment.method == "credit_card":
                    if payment.amount.amount > 0:
                        success = True
                    else:
                        success = False
                else:
                    success = True
            else:
                success = False
        else:
            success = False

        if success:
            payment.complete()
            order.mark_paid()
            self._order_repo.save(order)
        else:
            payment.fail()

        self._payment_repo.save(payment)
        return PaymentDto.from_domain(payment)
