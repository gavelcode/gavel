import { useState } from "react";

interface OrderSummary {
  id: string;
  total: number;
  status: string;
}

// deliberate no-unused-vars: unusedHelper is never used
function unusedHelper(value: string): string {
  return value.toUpperCase();
}

export function useOrders() {
  const [orders, setOrders] = useState<OrderSummary[]>([]);
  const [loading, setLoading] = useState(false);

  const fetchOrders = async () => {
    setLoading(true);
    const response = await fetch("/api/orders");
    const data = await response.json();
    setOrders(data);
    setLoading(false);
  };

  const addOrder = (order: OrderSummary) => {
    setOrders((prev) => [...prev, order]);
  };

  return { orders, loading, fetchOrders, addOrder };
}
