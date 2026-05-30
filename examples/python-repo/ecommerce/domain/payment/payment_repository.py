from abc import ABC, abstractmethod


class PaymentRepository(ABC):
    @abstractmethod
    def save(self, payment):
        pass

    @abstractmethod
    def find_by_id(self, payment_id):
        pass

    @abstractmethod
    def find_by_order_id(self, order_id):
        pass

    @abstractmethod
    def next_id(self):
        pass
