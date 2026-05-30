from abc import ABC, abstractmethod


class InventoryRepository(ABC):
    @abstractmethod
    def save(self, stock):
        pass

    @abstractmethod
    def find_by_product_id(self, product_id):
        pass
