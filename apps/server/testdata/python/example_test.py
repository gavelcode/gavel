import unittest

from example import discount_for


class DiscountTest(unittest.TestCase):
    def test_large_total_gets_larger_discount(self):
        self.assertEqual(discount_for(120), 10)

    def test_medium_total_gets_discount(self):
        self.assertEqual(discount_for(75), 5)


if __name__ == "__main__":
    unittest.main()
