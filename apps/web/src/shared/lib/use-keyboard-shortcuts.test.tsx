import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { useState } from "react";
import { useListKeyboardNav } from "./use-keyboard-shortcuts";

function TestList({ items }: { items: string[] }) {
  const [activeIndex, setActiveIndex] = useState(-1);
  const { containerRef } = useListKeyboardNav({
    itemCount: items.length,
    activeIndex,
    onActiveChange: setActiveIndex,
    onSelect: () => {},
  });

  return (
    <div ref={containerRef} tabIndex={0} data-testid="list">
      {items.map((item, i) => (
        <div key={item} data-active={i === activeIndex ? "true" : undefined}>
          {item}
        </div>
      ))}
      <span data-testid="active">{activeIndex}</span>
    </div>
  );
}

describe("useListKeyboardNav", () => {
  it("moves down on j key", async () => {
    const user = userEvent.setup();
    render(<TestList items={["A", "B", "C"]} />);
    const list = screen.getByTestId("list");
    list.focus();
    await user.keyboard("j");
    expect(screen.getByTestId("active").textContent).toBe("0");
    await user.keyboard("j");
    expect(screen.getByTestId("active").textContent).toBe("1");
  });

  it("moves up on k key", async () => {
    const user = userEvent.setup();
    render(<TestList items={["A", "B", "C"]} />);
    const list = screen.getByTestId("list");
    list.focus();
    await user.keyboard("j");
    await user.keyboard("j");
    await user.keyboard("k");
    expect(screen.getByTestId("active").textContent).toBe("0");
  });

  it("does not go below last item", async () => {
    const user = userEvent.setup();
    render(<TestList items={["A", "B"]} />);
    const list = screen.getByTestId("list");
    list.focus();
    await user.keyboard("j");
    await user.keyboard("j");
    await user.keyboard("j");
    expect(screen.getByTestId("active").textContent).toBe("1");
  });

  it("does not go above first item", async () => {
    const user = userEvent.setup();
    render(<TestList items={["A", "B"]} />);
    const list = screen.getByTestId("list");
    list.focus();
    await user.keyboard("k");
    expect(screen.getByTestId("active").textContent).toBe("-1");
  });
});
