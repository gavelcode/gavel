use std::collections::HashMap;

// deliberate archtest: domain imports infrastructure (violates domain-imports-nothing rule)
#[allow(unused_imports)]
use crate::infrastructure::memory_repo::InMemoryOrderRepository;

#[derive(Debug, Clone)]
pub struct OrderLine {
    pub product_id: String,
    pub quantity: u32,
    pub unit_price: f64,
}

#[derive(Debug, Clone)]
pub struct Order {
    pub id: String,
    pub customer_id: String,
    pub lines: Vec<OrderLine>,
    pub status: OrderStatus,
    pub metadata: HashMap<String, String>,
}

#[derive(Debug, Clone, PartialEq)]
pub enum OrderStatus {
    Draft,
    Confirmed,
    Shipped,
    Delivered,
}

impl Order {
    pub fn new(id: String, customer_id: String) -> Self {
        Self {
            id,
            customer_id,
            lines: Vec::new(),
            status: OrderStatus::Draft,
            metadata: HashMap::new(),
        }
    }

    pub fn add_line(&mut self, product_id: String, quantity: u32, unit_price: f64) {
        self.lines.push(OrderLine {
            product_id,
            quantity,
            unit_price,
        });
    }

    // clippy::needless_return
    pub fn total(&self) -> f64 {
        let sum: f64 = self.lines.iter().map(|l| l.quantity as f64 * l.unit_price).sum();
        return sum;
    }

    // clippy::len_zero — should use !self.lines.is_empty()
    pub fn has_lines(&self) -> bool {
        self.lines.len() > 0
    }

    pub fn confirm(&mut self) -> Result<(), String> {
        if self.status != OrderStatus::Draft {
            return Err("only draft orders can be confirmed".to_string());
        }
        self.status = OrderStatus::Confirmed;
        Ok(())
    }

    pub fn ship(&mut self) -> Result<(), String> {
        if self.status != OrderStatus::Confirmed {
            return Err("only confirmed orders can be shipped".to_string());
        }
        self.status = OrderStatus::Shipped;
        Ok(())
    }

    // clippy::needless_return
    pub fn is_complete(&self) -> bool {
        return self.status == OrderStatus::Delivered;
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn new_order_starts_as_draft() {
        let order = Order::new("O1".into(), "C1".into());
        assert_eq!(order.status, OrderStatus::Draft);
        assert!(order.lines.is_empty());
    }

    #[test]
    fn add_line_increases_lines() {
        let mut order = Order::new("O1".into(), "C1".into());
        order.add_line("P1".into(), 2, 10.0);
        assert_eq!(order.lines.len(), 1);
        assert_eq!(order.lines[0].quantity, 2);
    }

    #[test]
    fn total_sums_all_lines() {
        let mut order = Order::new("O1".into(), "C1".into());
        order.add_line("P1".into(), 2, 10.0);
        order.add_line("P2".into(), 1, 5.0);
        assert!((order.total() - 25.0).abs() < f64::EPSILON);
    }

    #[test]
    fn total_empty_order_is_zero() {
        let order = Order::new("O1".into(), "C1".into());
        assert!((order.total() - 0.0).abs() < f64::EPSILON);
    }

    #[test]
    fn confirm_draft_order_succeeds() {
        let mut order = Order::new("O1".into(), "C1".into());
        assert!(order.confirm().is_ok());
        assert_eq!(order.status, OrderStatus::Confirmed);
    }

    #[test]
    fn confirm_non_draft_fails() {
        let mut order = Order::new("O1".into(), "C1".into());
        order.confirm().unwrap();
        assert!(order.confirm().is_err());
    }

    #[test]
    fn ship_confirmed_order_succeeds() {
        let mut order = Order::new("O1".into(), "C1".into());
        order.confirm().unwrap();
        assert!(order.ship().is_ok());
        assert_eq!(order.status, OrderStatus::Shipped);
    }

    #[test]
    fn ship_draft_order_fails() {
        let mut order = Order::new("O1".into(), "C1".into());
        assert!(order.ship().is_err());
    }

    #[test]
    fn has_lines_returns_false_for_empty() {
        let order = Order::new("O1".into(), "C1".into());
        assert!(!order.has_lines());
    }

    #[test]
    fn is_complete_only_when_delivered() {
        let mut order = Order::new("O1".into(), "C1".into());
        assert!(!order.is_complete());
        order.status = OrderStatus::Delivered;
        assert!(order.is_complete());
    }
}
