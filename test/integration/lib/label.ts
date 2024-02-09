import { Page, expect } from '@playwright/test'

export async function setLabel(
    page: Page,
    key: string,
    value: number,
    isMobile: boolean,
): Promise<void> {
  // Create a label for the service
  await page.getByRole('link', { name: 'Labels' }).click()

  if (isMobile) {
    await page.getByRole('button', { name: 'Add' }).click()
  } else {
    await page.getByTestId('create-label').click()
  }

  await page.getByLabel('Key', { exact: true }).fill(key)
  await page.getByText('Create "' + key + '"').click()
  await page.getByLabel('Value', { exact: true }).fill(value)
  await page.click('[role=dialog] button[type=submit]')

  await expect(page.getByText(key)).toBeVisible()
  await expect(page.getByText(value)).toBeVisible()
}
