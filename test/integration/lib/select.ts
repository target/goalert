import { Page } from '@playwright/test'

export async function dropdownSelect(
  page: Page,
  fieldLabel: string,
  optionLabel: string,
): Promise<void> {
  await page
    .locator('div', { has: page.locator('label', { hasText: fieldLabel }) })
    .locator('[role=combobox]')
    .click()

  await page.locator('[role=option]', { hasText: optionLabel }).click()
}
