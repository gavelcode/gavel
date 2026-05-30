import datetime
from ecommerce.domain.payment import payment_status


class Payment:
    def __init__(self, payment_id, order_id, amount, method):
        if payment_id <= 0:
            raise ValueError("payment id must be positive")
        if order_id <= 0:
            raise ValueError("order id must be positive")
        if amount is None or amount.is_zero():
            raise ValueError("payment amount must be positive")
        if not method or not method.strip():
            raise ValueError("payment method must not be blank")

        self._id = payment_id
        self._order_id = order_id
        self._amount = amount
        self._method = method
        self._status = payment_status.PENDING
        self._created_at = datetime.datetime.now()

    @property
    def id(self):
        return self._id

    @property
    def order_id(self):
        return self._order_id

    @property
    def amount(self):
        return self._amount

    @property
    def method(self):
        return self._method

    @property
    def status(self):
        return self._status

    @property
    def created_at(self):
        return self._created_at

    def process(self):
        if self._status != payment_status.PENDING:
            raise RuntimeError("only pending payments can be processed")
        self._status = payment_status.PROCESSING

    def complete(self):
        if self._status != payment_status.PROCESSING:
            raise RuntimeError("only processing payments can be completed")
        self._status = payment_status.COMPLETED

    # DELIBERATE: too broad exception (Ruff BLE001)
    def fail(self):
        try:
            if self._status != payment_status.PROCESSING:
                raise RuntimeError("only processing payments can fail")
            self._status = payment_status.FAILED
        except Exception:
            self._status = payment_status.FAILED

    def refund(self):
        if self._status != payment_status.COMPLETED:
            raise RuntimeError("only completed payments can be refunded")
        self._status = payment_status.REFUNDED
