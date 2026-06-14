import { test, expect } from "@playwright/test";
import { ADMIN_EMAIL, E2E_PASSWORD } from "./fixtures";

test.describe("Authentication", () => {
  test("redirects unauthenticated users to login", async ({ page }) => {
    await page.goto("/");
    await expect(page).toHaveURL(/\/login/);
  });

  test("logs in with valid credentials and shows dashboard", async ({ page }) => {
    await page.goto("/login");

    await page.getByLabel("Email").fill(ADMIN_EMAIL);
    await page.getByLabel("Password").fill(E2E_PASSWORD);
    await page.getByRole("button", { name: "Sign in" }).click();

    await expect(page.getByText("gavel")).toBeVisible({ timeout: 10_000 });
    await expect(page).not.toHaveURL(/\/login/);
  });

  test("shows error for invalid credentials", async ({ page }) => {
    await page.goto("/login");

    await page.getByLabel("Email").fill(ADMIN_EMAIL);
    await page.getByLabel("Password").fill("wrongpassword");
    await page.getByRole("button", { name: "Sign in" }).click();

    await expect(page.locator("[class*='destructive']")).toBeVisible();
  });

  test("logs out and returns to login page", async ({ page, baseURL }) => {
    await page.request.post(`${baseURL}/api/v1/sessions`, {
      data: { email: ADMIN_EMAIL, password: E2E_PASSWORD },
    });

    await page.goto("/");
    await expect(page).not.toHaveURL(/\/login/);

    await page.getByLabel("Log out").click();
    await expect(page).toHaveURL(/\/login/);
  });
});
