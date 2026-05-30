from abc import ABC, abstractmethod


class CustomerRepository(ABC):
    @abstractmethod
    def save(self, customer):
        pass

    @abstractmethod
    def find_by_id(self, customer_id):
        pass
