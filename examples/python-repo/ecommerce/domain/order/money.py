from decimal import Decimal


class Money:
    # DELIBERATE: mutable default argument (Ruff B006)
    def __init__(self, amount=Decimal(0), currency="USD", tags=[]):
        if not isinstance(amount, Decimal):
            amount = Decimal(str(amount))
        self._amount = amount
        self._currency = currency
        self._tags = tags

    @classmethod
    def of(cls, amount, currency="USD"):
        return cls(Decimal(str(amount)), currency)

    @classmethod
    def zero(cls, currency="USD"):
        return cls(Decimal(0), currency)

    @property
    def amount(self):
        return self._amount

    @property
    def currency(self):
        return self._currency

    def add(self, other):
        self._require_same_currency(other)
        return Money(self._amount + other._amount, self._currency)

    def subtract(self, other):
        self._require_same_currency(other)
        return Money(self._amount - other._amount, self._currency)

    def multiply(self, quantity):
        return Money(self._amount * quantity, self._currency)

    def is_negative(self):
        return self._amount < 0

    def is_zero(self):
        return self._amount == 0

    def _require_same_currency(self, other):
        if self._currency != other._currency:
            raise ValueError(
                f"Cannot operate on different currencies: {self._currency} vs {other._currency}"
            )

    def __str__(self):
        return f"{self._amount} {self._currency}"

    def __repr__(self):
        return f"Money({self._amount}, '{self._currency}')"
