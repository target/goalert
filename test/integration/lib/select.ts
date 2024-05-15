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

export async function pageAction(
  page: Page,
  mobileAction: string,
  wideAction: string = mobileAction,
): Promise<void> {
  const vp = page.viewportSize()
  const mobile = vp && vp.width < 400
  if (mobile) {
    const hasPopup = await page
      .locator('button[data-cy="page-fab"]')
      .getAttribute('aria-haspopup')
    if (hasPopup) {
      await page.hover('button[data-cy="page-fab"]')
    }

    await page.getByLabel(mobileAction).locator('button').click()
    return
  }

  await page.getByRole('button', { name: wideAction }).click()
}
