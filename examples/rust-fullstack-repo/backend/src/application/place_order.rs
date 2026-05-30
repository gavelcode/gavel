use crate::domain::customer::Customer;
use crate::domain::order::Order;
use crate::domain::payment::{Payment, PaymentMethod};

// deliberate archtest: application imports infrastructure (violates application-imports-domain-only rule)
#[allow(unused_imports)]
use crate::infrastructure::memory_repo::InMemoryOrderRepository;

pub struct PlaceOrderCommand {
    pub customer_id: String,
    pub items: Vec<(String, u32, f64)>,
    pub payment_method: PaymentMethod,
}

pub struct PlaceOrderResult {
    pub order_id: String,
    pub payment_id: String,
    pub total: f64,
}

pub fn place_order(
    command: PlaceOrderCommand,
    customer: &Customer,
) -> Result<PlaceOrderResult, String> {
    let order_id = format!("ORD-{}", uuid_stub());
    let mut order = Order::new(order_id.clone(), customer.id.clone());

    for (product_id, quantity, price) in &command.items {
        order.add_line(product_id.clone(), *quantity, *price);
    }

    // clippy::bool_comparison — should use !order.has_lines()
    if order.has_lines() == false {
        return Err("order must have at least one line".to_string());
    }

    order.confirm()?;

    let payment_id = format!("PAY-{}", uuid_stub());
    let mut payment = Payment::new(
        payment_id.clone(),
        order_id.clone(),
        order.total(),
        command.payment_method,
    );
    payment.authorize()?;
    payment.capture()?;

    // clippy::useless_format — format! with no args beyond the string
    let _log_msg = format!("order placed successfully");

    Ok(PlaceOrderResult {
        order_id,
        payment_id,
        total: order.total(),
    })
}

fn uuid_stub() -> String {
    "00000000".to_string()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn place_order_happy_path() {
        let customer = Customer::new("C1".into(), "Alice".into(), "a@test.com".into());
        let cmd = PlaceOrderCommand {
            customer_id: "C1".into(),
            items: vec![("P1".into(), 2, 10.0), ("P2".into(), 1, 5.0)],
            payment_method: PaymentMethod::CreditCard,
        };
        let result = place_order(cmd, &customer).unwrap();
        assert!((result.total - 25.0).abs() < f64::EPSILON);
        assert!(result.order_id.starts_with("ORD-"));
        assert!(result.payment_id.starts_with("PAY-"));
    }

    #[test]
    fn place_order_no_items_fails() {
        let customer = Customer::new("C1".into(), "Alice".into(), "a@test.com".into());
        let cmd = PlaceOrderCommand {
            customer_id: "C1".into(),
            items: vec![],
            payment_method: PaymentMethod::Wallet,
        };
        assert!(place_order(cmd, &customer).is_err());
    }
}
