import { test, expect } from "./fixtures";

test.describe("Gavelspaces", () => {
  test("home page shows gavelspace and navigates to detail", async ({
    authedPage: page,
  }) => {
    await page.goto("/");

    await expect(page).toHaveURL(/\/gavelspaces\/gavel/, { timeout: 10_000 });
    await expect(page.getByText("1 project")).toBeVisible();
  });

  test("gavelspace detail shows tabs and project count", async ({
    authedPage: page,
  }) => {
    await page.goto("/gavelspaces/gavel");

    await expect(page.getByText("1 project")).toBeVisible({ timeout: 10_000 });

    await expect(page.getByRole("link", { name: "Overview" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Projects" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Findings" })).toBeVisible();
    await expect(page.getByRole("link", { name: "Case Files" })).toBeVisible();
  });

  test("gavelspace overview shows project strip", async ({
    authedPage: page,
  }) => {
    await page.goto("/gavelspaces/gavel");

    await expect(page.getByText("Core Library")).toBeVisible({
      timeout: 10_000,
    });
  });
});
