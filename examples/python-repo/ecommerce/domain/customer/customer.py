class Customer:
    def __init__(self, customer_id, name, email):
        if customer_id <= 0:
            raise ValueError("customer id must be positive")
        if not name or not name.strip():
            raise ValueError("customer name must not be blank")
        if not email or "@" not in email:
            raise ValueError("invalid email")

        self._id = customer_id
        self._name = name
        self._email = email
        self._shipping_address = None

    @property
    def id(self):
        return self._id

    @property
    def name(self):
        return self._name

    @property
    def email(self):
        return self._email

    @property
    def shipping_address(self):
        return self._shipping_address

    def update_name(self, name):
        if not name or not name.strip():
            raise ValueError("name must not be blank")
        # DELIBERATE: unused variable (Ruff F841)
        old_name = self._name
        self._name = name

    def update_email(self, email):
        if not email or "@" not in email:
            raise ValueError("invalid email")
        self._email = email

    def set_shipping_address(self, address):
        self._shipping_address = address
