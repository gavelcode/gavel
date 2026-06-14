import { screen, render } from "@testing-library/react";
import "@/test/msw-server";
import { Providers } from "./providers";

describe("Providers", () => {
  it("renders children within the provider tree", async () => {
    render(
      <Providers>
        <div>test child</div>
      </Providers>,
    );
    expect(await screen.findByText("test child")).toBeInTheDocument();
  });
});
