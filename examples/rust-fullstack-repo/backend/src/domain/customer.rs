#[derive(Debug, Clone)]
pub struct Customer {
    pub id: String,
    pub name: String,
    pub email: String,
    pub tier: CustomerTier,
}

#[derive(Debug, Clone, PartialEq)]
pub enum CustomerTier {
    Standard,
    Premium,
    Enterprise,
}

impl Customer {
    pub fn new(id: String, name: String, email: String) -> Self {
        Self {
            id,
            name,
            email,
            tier: CustomerTier::Standard,
        }
    }

    // clippy::ptr_arg — should take &str instead of &String
    pub fn update_email(&mut self, new_email: &String) {
        self.email = new_email.clone();
    }

    // clippy::ptr_arg — should take &str instead of &String
    pub fn matches_name(&self, query: &String) -> bool {
        self.name.to_lowercase().contains(&query.to_lowercase())
    }

    pub fn upgrade_to_premium(&mut self) {
        self.tier = CustomerTier::Premium;
    }

    pub fn is_premium(&self) -> bool {
        self.tier == CustomerTier::Premium || self.tier == CustomerTier::Enterprise
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn new_customer_is_standard() {
        let c = Customer::new("C1".into(), "Alice".into(), "alice@test.com".into());
        assert_eq!(c.tier, CustomerTier::Standard);
    }

    #[test]
    fn update_email_changes_email() {
        let mut c = Customer::new("C1".into(), "Alice".into(), "old@test.com".into());
        let new_email = "new@test.com".to_string();
        c.update_email(&new_email);
        assert_eq!(c.email, "new@test.com");
    }

    #[test]
    fn matches_name_case_insensitive() {
        let c = Customer::new("C1".into(), "Alice Smith".into(), "a@test.com".into());
        let query = "alice".to_string();
        assert!(c.matches_name(&query));
    }

    #[test]
    fn matches_name_no_match() {
        let c = Customer::new("C1".into(), "Alice".into(), "a@test.com".into());
        let query = "Bob".to_string();
        assert!(!c.matches_name(&query));
    }

    #[test]
    fn upgrade_to_premium() {
        let mut c = Customer::new("C1".into(), "Alice".into(), "a@test.com".into());
        assert!(!c.is_premium());
        c.upgrade_to_premium();
        assert!(c.is_premium());
    }

    #[test]
    fn enterprise_is_premium() {
        let mut c = Customer::new("C1".into(), "Alice".into(), "a@test.com".into());
        c.tier = CustomerTier::Enterprise;
        assert!(c.is_premium());
    }
}
