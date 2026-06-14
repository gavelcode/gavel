import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { Tabs } from "./tabs";

const items = ["Overview", "Issues", "Settings"];

describe("Tabs", () => {
  it("should render a tablist with tab roles", () => {
    render(<Tabs items={items} active={0} />);
    expect(screen.getByRole("tablist")).toBeInTheDocument();
    expect(screen.getAllByRole("tab")).toHaveLength(3);
  });

  it("should mark active tab with aria-selected", () => {
    render(<Tabs items={items} active={1} />);
    const tabs = screen.getAllByRole("tab");
    expect(tabs[0]).toHaveAttribute("aria-selected", "false");
    expect(tabs[1]).toHaveAttribute("aria-selected", "true");
    expect(tabs[2]).toHaveAttribute("aria-selected", "false");
  });

  it("should call onChange when a tab is clicked", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Tabs items={items} active={0} onChange={onChange} />);

    await user.click(screen.getAllByRole("tab")[2]);
    expect(onChange).toHaveBeenCalledWith(2);
  });

  it("should render underline variant by default", () => {
    const { container } = render(<Tabs items={items} active={0} />);
    expect(container.querySelector("[role='tablist']")?.className).toContain("border-b");
  });

  it("should render pill variant without border-bottom", () => {
    const { container } = render(<Tabs items={items} active={0} variant="pill" />);
    expect(container.querySelector("[role='tablist']")?.className).not.toContain("border-b");
  });

  it("should apply rounded background on active pill tab", () => {
    render(<Tabs items={items} active={1} variant="pill" />);
    const activeTab = screen.getAllByRole("tab")[1];
    expect(activeTab.className).toContain("bg-muted");
    expect(activeTab.className).toContain("rounded-md");
  });

  it("should have focus-visible ring on tab buttons", () => {
    render(<Tabs items={items} active={0} />);
    const tabs = screen.getAllByRole("tab");
    tabs.forEach((tab) => {
      expect(tab.className).toContain("focus-visible:ring-1");
    });
  });

  it("should move to next tab on ArrowRight key", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Tabs items={items} active={0} onChange={onChange} />);
    const tabs = screen.getAllByRole("tab");
    tabs[0].focus();
    await user.keyboard("{ArrowRight}");
    expect(onChange).toHaveBeenCalledWith(1);
  });

  it("should move to previous tab on ArrowLeft key", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Tabs items={items} active={2} onChange={onChange} />);
    const tabs = screen.getAllByRole("tab");
    tabs[2].focus();
    await user.keyboard("{ArrowLeft}");
    expect(onChange).toHaveBeenCalledWith(1);
  });

  it("should wrap to last tab on ArrowLeft from first", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Tabs items={items} active={0} onChange={onChange} />);
    const tabs = screen.getAllByRole("tab");
    tabs[0].focus();
    await user.keyboard("{ArrowLeft}");
    expect(onChange).toHaveBeenCalledWith(2);
  });

  it("should wrap to first tab on ArrowRight from last", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Tabs items={items} active={2} onChange={onChange} />);
    const tabs = screen.getAllByRole("tab");
    tabs[2].focus();
    await user.keyboard("{ArrowRight}");
    expect(onChange).toHaveBeenCalledWith(0);
  });

  it("should move to first tab on Home key", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Tabs items={items} active={2} onChange={onChange} />);
    const tabs = screen.getAllByRole("tab");
    tabs[2].focus();
    await user.keyboard("{Home}");
    expect(onChange).toHaveBeenCalledWith(0);
  });

  it("should move to last tab on End key", async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<Tabs items={items} active={0} onChange={onChange} />);
    const tabs = screen.getAllByRole("tab");
    tabs[0].focus();
    await user.keyboard("{End}");
    expect(onChange).toHaveBeenCalledWith(2);
  });

  it("should set tabIndex=-1 on inactive tabs and tabIndex=0 on active", () => {
    render(<Tabs items={items} active={1} />);
    const tabs = screen.getAllByRole("tab");
    expect(tabs[0]).toHaveAttribute("tabindex", "-1");
    expect(tabs[1]).toHaveAttribute("tabindex", "0");
    expect(tabs[2]).toHaveAttribute("tabindex", "-1");
  });
});
