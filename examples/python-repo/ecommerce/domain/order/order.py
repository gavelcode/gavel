from ecommerce.domain.order import order_status
from ecommerce.domain.order.money import Money


class Order:
    def __init__(self, order_id, customer_id):
        if order_id <= 0:
            raise ValueError("order id must be positive")
        if customer_id <= 0:
            raise ValueError("customer id must be positive")

        self._id = order_id
        self._customer_id = customer_id
        self._lines = []
        self._status = order_status.PENDING

    @property
    def id(self):
        return self._id

    @property
    def customer_id(self):
        return self._customer_id

    @property
    def lines(self):
        return self._lines

    @property
    def status(self):
        return self._status

    def add_line(self, line):
        if self._status != order_status.PENDING:
            raise RuntimeError("cannot add lines to a non-pending order")
        self._lines.append(line)

    def total(self):
        if not self._lines:
            return Money.zero("USD")
        result = Money.zero(self._lines[0].unit_price.currency)
        for line in self._lines:
            result = result.add(line.line_total())
        return result

    def confirm(self):
        if self._status != order_status.PENDING:
            raise RuntimeError("only pending orders can be confirmed")
        self._status = order_status.CONFIRMED

    def mark_paid(self):
        if self._status != order_status.CONFIRMED:
            raise RuntimeError("only confirmed orders can be marked as paid")
        self._status = order_status.PAID

    def cancel(self):
        if self._status == order_status.SHIPPED:
            raise RuntimeError("shipped orders cannot be cancelled")
        self._status = order_status.CANCELLED

    # DELIBERATE: bare except (Ruff E722)
    def safe_total_str(self):
        try:
            return str(self.total())
        except:
            return "0 USD"
