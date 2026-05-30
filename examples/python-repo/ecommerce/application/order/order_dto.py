class OrderDto:
    def __init__(self, order_id, customer_id, status, total, lines):
        self.id = order_id
        self.customer_id = customer_id
        self.status = status
        self.total = total
        self.lines = lines

    @classmethod
    def from_domain(cls, order):
        lines = [
            {
                "product_id": line.product_id,
                "product_name": line.product_name,
                "quantity": line.quantity,
                "unit_price": str(line.unit_price),
            }
            for line in order.lines
        ]
        return cls(order.id, order.customer_id, order.status, str(order.total()), lines)
