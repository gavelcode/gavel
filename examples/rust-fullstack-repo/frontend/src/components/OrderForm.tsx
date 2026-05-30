import { useState } from "react";

interface OrderFormProps {
  onSubmit: (productId: string, quantity: number) => void;
  isLoggedIn: boolean;
}

export function OrderForm({ onSubmit, isLoggedIn }: OrderFormProps) {
  const [productId, setProductId] = useState("");
  const [quantity, setQuantity] = useState(1);

  // deliberate react-hooks/rules-of-hooks: conditional hook
  if (isLoggedIn) {
    const [, setAutoFilled] = useState(false);
    setAutoFilled(true);
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    onSubmit(productId, quantity);
  };

  return (
    <form onSubmit={handleSubmit}>
      <input
        type="text"
        value={productId}
        onChange={(e) => setProductId(e.target.value)}
        placeholder="Product ID"
      />
      <input
        type="number"
        value={quantity}
        onChange={(e) => setQuantity(Number(e.target.value))}
        min={1}
      />
      <button type="submit">Add to Order</button>
    </form>
  );
}
