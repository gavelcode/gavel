#[derive(Debug, Clone)]
pub struct Product {
    pub id: String,
    pub name: String,
    pub price: f64,
    pub category: ProductCategory,
    pub stock: u32,
}

#[derive(Debug, Clone, PartialEq)]
pub enum ProductCategory {
    Electronics,
    Clothing,
    Food,
    Books,
}

impl Product {
    pub fn new(id: String, name: String, price: f64, category: ProductCategory) -> Self {
        Self {
            id,
            name,
            price,
            category,
            stock: 0,
        }
    }

    pub fn restock(&mut self, amount: u32) {
        self.stock += amount;
    }

    pub fn reserve(&mut self, amount: u32) -> Result<(), String> {
        if self.stock < amount {
            return Err("insufficient stock".to_string());
        }
        self.stock -= amount;
        Ok(())
    }

    // clippy::single_match — should use if let
    pub fn discount_rate(&self) -> f64 {
        match self.category {
            ProductCategory::Books => 0.1,
            _ => 0.0,
        }
    }

    pub fn discounted_price(&self) -> f64 {
        self.price * (1.0 - self.discount_rate())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn new_product_starts_with_zero_stock() {
        let p = Product::new("P1".into(), "Widget".into(), 9.99, ProductCategory::Electronics);
        assert_eq!(p.stock, 0);
    }

    #[test]
    fn restock_increases_stock() {
        let mut p = Product::new("P1".into(), "Widget".into(), 9.99, ProductCategory::Electronics);
        p.restock(10);
        assert_eq!(p.stock, 10);
        p.restock(5);
        assert_eq!(p.stock, 15);
    }

    #[test]
    fn reserve_decreases_stock() {
        let mut p = Product::new("P1".into(), "Widget".into(), 9.99, ProductCategory::Electronics);
        p.restock(10);
        assert!(p.reserve(3).is_ok());
        assert_eq!(p.stock, 7);
    }

    #[test]
    fn reserve_insufficient_stock_fails() {
        let mut p = Product::new("P1".into(), "Widget".into(), 9.99, ProductCategory::Electronics);
        p.restock(2);
        assert!(p.reserve(5).is_err());
    }

    #[test]
    fn discount_rate_books_is_ten_percent() {
        let p = Product::new("P1".into(), "Rust Book".into(), 40.0, ProductCategory::Books);
        assert!((p.discount_rate() - 0.1).abs() < f64::EPSILON);
    }

    #[test]
    fn discount_rate_non_books_is_zero() {
        let p = Product::new("P1".into(), "TV".into(), 500.0, ProductCategory::Electronics);
        assert!((p.discount_rate() - 0.0).abs() < f64::EPSILON);
    }

    #[test]
    fn discounted_price_applies_rate() {
        let p = Product::new("P1".into(), "Rust Book".into(), 40.0, ProductCategory::Books);
        assert!((p.discounted_price() - 36.0).abs() < f64::EPSILON);
    }
}
