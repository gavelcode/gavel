import { test, expect } from "./fixtures";

test.describe("API Tokens", () => {
  test("creates a token and shows it in the list", async ({
    authedPage: page,
  }) => {
    await page.goto("/tokens");

    await page.getByRole("button", { name: "Create token" }).click();

    await page.getByLabel("Name").fill("e2e-test-token");
    await page.getByLabel("ingest").check();
    await page.getByRole("button", { name: "Create", exact: true }).click();

    await expect(page.getByText("gav_").first()).toBeVisible({ timeout: 10_000 });

    await expect(page.getByText("e2e-test-token")).toBeVisible();
  });

  test("token list shows existing tokens", async ({ authedPage: page }) => {
    await page.request.post(`/api/v1/me/tokens`, {
      data: { name: "pre-seeded-token", scopes: ["ingest"] },
    });

    await page.goto("/tokens");

    await expect(page.getByText("pre-seeded-token")).toBeVisible({
      timeout: 10_000,
    });
  });
});
