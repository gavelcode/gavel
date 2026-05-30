from abc import ABC, abstractmethod


class ProductRepository(ABC):
    @abstractmethod
    def find_by_id(self, product_id):
        pass
