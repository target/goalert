import { Page } from '@playwright/test'

export async function dropdownSelect(
  page: Page,
  selector: string,
  label: string,
): Promise<void> {
  await page.click(selector, { force: true }) // force click so the wrapper element is clicked, as the input is hidden
  await page.locator('[role=option]', { hasText: label }).click()
}
