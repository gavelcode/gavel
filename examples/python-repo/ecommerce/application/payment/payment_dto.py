class PaymentDto:
    def __init__(self, payment_id, order_id, amount, method, status):
        self.id = payment_id
        self.order_id = order_id
        self.amount = amount
        self.method = method
        self.status = status

    @classmethod
    def from_domain(cls, payment):
        return cls(
            payment.id,
            payment.order_id,
            str(payment.amount),
            payment.method,
            payment.status,
        )
