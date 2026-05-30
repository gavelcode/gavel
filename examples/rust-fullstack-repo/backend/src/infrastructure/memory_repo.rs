use std::collections::HashMap;

use crate::domain::order::Order;

pub struct InMemoryOrderRepository {
    orders: HashMap<String, Order>,
}

impl InMemoryOrderRepository {
    pub fn new() -> Self {
        Self {
            orders: HashMap::new(),
        }
    }

    pub fn save(&mut self, order: Order) {
        self.orders.insert(order.id.clone(), order);
    }

    pub fn find_by_id(&self, id: &str) -> Option<&Order> {
        self.orders.get(id)
    }

    pub fn find_all(&self) -> Vec<&Order> {
        self.orders.values().collect()
    }

    // clippy::unnecessary_mut_passed — self doesn't need to be &mut
    pub fn count(&mut self) -> usize {
        self.orders.len()
    }

    pub fn remove(&mut self, id: &str) -> Option<Order> {
        self.orders.remove(id)
    }
}
