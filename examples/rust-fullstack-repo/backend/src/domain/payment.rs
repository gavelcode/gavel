#[derive(Debug, Clone)]
pub struct Payment {
    pub id: String,
    pub order_id: String,
    pub amount: f64,
    pub status: PaymentStatus,
    pub method: PaymentMethod,
}

#[derive(Debug, Clone, PartialEq)]
pub enum PaymentStatus {
    Pending,
    Authorized,
    Captured,
    Refunded,
}

#[derive(Debug, Clone, PartialEq)]
pub enum PaymentMethod {
    CreditCard,
    BankTransfer,
    Wallet,
}

impl Payment {
    pub fn new(id: String, order_id: String, amount: f64, method: PaymentMethod) -> Self {
        Self {
            id,
            order_id,
            amount,
            status: PaymentStatus::Pending,
            method,
        }
    }

    pub fn authorize(&mut self) -> Result<(), String> {
        if self.status != PaymentStatus::Pending {
            return Err("payment already processed".to_string());
        }
        self.status = PaymentStatus::Authorized;
        Ok(())
    }

    pub fn capture(&mut self) -> Result<(), String> {
        if self.status != PaymentStatus::Authorized {
            return Err("payment must be authorized first".to_string());
        }
        self.status = PaymentStatus::Captured;
        Ok(())
    }

    // clippy::redundant_clone — order_id is not used after this
    pub fn receipt_reference(&self) -> String {
        let reference = self.order_id.clone();
        format!("PAY-{}-{}", self.id, reference.clone())
    }

    pub fn is_captured(&self) -> bool {
        self.status == PaymentStatus::Captured
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn new_payment_is_pending() {
        let p = Payment::new("P1".into(), "O1".into(), 100.0, PaymentMethod::CreditCard);
        assert_eq!(p.status, PaymentStatus::Pending);
    }

    #[test]
    fn authorize_pending_succeeds() {
        let mut p = Payment::new("P1".into(), "O1".into(), 100.0, PaymentMethod::CreditCard);
        assert!(p.authorize().is_ok());
        assert_eq!(p.status, PaymentStatus::Authorized);
    }

    #[test]
    fn authorize_non_pending_fails() {
        let mut p = Payment::new("P1".into(), "O1".into(), 100.0, PaymentMethod::CreditCard);
        p.authorize().unwrap();
        assert!(p.authorize().is_err());
    }

    #[test]
    fn capture_authorized_succeeds() {
        let mut p = Payment::new("P1".into(), "O1".into(), 100.0, PaymentMethod::CreditCard);
        p.authorize().unwrap();
        assert!(p.capture().is_ok());
        assert!(p.is_captured());
    }

    #[test]
    fn capture_pending_fails() {
        let mut p = Payment::new("P1".into(), "O1".into(), 100.0, PaymentMethod::CreditCard);
        assert!(p.capture().is_err());
    }

    #[test]
    fn receipt_reference_format() {
        let p = Payment::new("P1".into(), "O1".into(), 100.0, PaymentMethod::BankTransfer);
        assert_eq!(p.receipt_reference(), "PAY-P1-O1");
    }
}
