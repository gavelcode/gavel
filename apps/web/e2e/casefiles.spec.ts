import { test, expect } from "./fixtures";

test.describe("Case Files", () => {
  test("lists case files and navigates to detail", async ({
    authedPage: page,
  }) => {
    await page.goto("/gavelspaces/gavel/case-files");

    await expect(page.getByText("e2e1234a")).toBeVisible({ timeout: 10_000 });

    await page.getByRole("link").filter({ hasText: /^[0-9a-f]{8}$/ }).click();

    await expect(page).toHaveURL(/\/casefiles\//);
  });

  test("casefile detail shows findings", async ({ authedPage: page }) => {
    await page.goto("/gavelspaces/gavel/case-files");

    await page.getByRole("link").filter({ hasText: /^[0-9a-f]{8}$/ }).click();

    await expect(
      page.getByText("Error return value not checked").first(),
    ).toBeVisible({ timeout: 10_000 });
  });

  test("findings tab in gavelspace shows findings", async ({
    authedPage: page,
  }) => {
    await page.goto("/gavelspaces/gavel/findings");

    await expect(page.getByText("golangci-lint").first()).toBeVisible({
      timeout: 10_000,
    });
  });
});
