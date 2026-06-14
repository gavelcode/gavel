import { test as base, expect, type Page } from "@playwright/test";

const ADMIN_EMAIL = "admin@local.dev";
const E2E_PASSWORD = "E2eTestPass1!";

async function loginViaAPI(page: Page, baseURL: string): Promise<void> {
  const response = await page.request.post(`${baseURL}/api/v1/sessions`, {
    data: { email: ADMIN_EMAIL, password: E2E_PASSWORD },
  });
  expect(response.ok()).toBe(true);
}

export const test = base.extend<{ authedPage: Page }>({
  authedPage: async ({ page, baseURL }, use) => {
    await loginViaAPI(page, baseURL!);
    await use(page);
  },
});

export { expect };
export { ADMIN_EMAIL, E2E_PASSWORD };
