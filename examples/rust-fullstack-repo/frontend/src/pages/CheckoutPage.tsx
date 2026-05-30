import { useState } from "react";
import { OrderForm } from "../components/OrderForm";
import { useOrders } from "../hooks/useOrders";

export function CheckoutPage() {
  const { orders, fetchOrders, addOrder } = useOrders();
  const [status, setStatus] = useState("idle");

  const handleSubmit = (productId: string, quantity: number) => {
    // deliberate eqeqeq: == instead of ===
    if (status == "processing") {
      return;
    }
    setStatus("processing");

    addOrder({
      id: `ORD-${Date.now()}`,
      total: quantity * 9.99,
      status: "pending",
    });

    setStatus("idle");
  };

  return (
    <div>
      <h1>Checkout</h1>
      <OrderForm onSubmit={handleSubmit} isLoggedIn={true} />
      <button onClick={fetchOrders}>Refresh Orders</button>
      <ul>
        {orders.map((order) => (
          <li key={order.id}>
            {order.id} - ${order.total} - {order.status}
          </li>
        ))}
      </ul>
    </div>
  );
}
