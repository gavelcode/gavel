import { test, expect } from "./fixtures";

test.describe("Projects", () => {
  test("overview shows project and navigates to detail", async ({
    authedPage: page,
  }) => {
    await page.goto("/gavelspaces/gavel");

    await expect(page.getByText("Core Library")).toBeVisible({
      timeout: 10_000,
    });

    await page.getByText("Core Library").click();

    await expect(page).toHaveURL(/\/projects\//);
  });

  test("project detail page loads project key in URL", async ({
    authedPage: page,
  }) => {
    await page.goto("/gavelspaces/gavel");
    await page.getByText("Core Library").click();

    await expect(page).toHaveURL(/\/projects\/core/, { timeout: 10_000 });
  });
});
