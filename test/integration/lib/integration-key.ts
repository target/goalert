import { Page, expect } from '@playwright/test'

export async function createIntegrationKey(
  page: Page,
  intKey: number,
  isMobile: boolean,
): Promise<void> {
  await page.getByRole('link', { name: 'Integration Keys' }).click()

  if (isMobile) {
    await page.getByRole('button', { name: 'Create Integration Key' }).click()
  } else {
    await page.getByTestId('create-key').click()
  }
  await page.getByLabel('Name').fill(intKey)
  await page.getByRole('button', { name: 'Submit' }).click()

  await expect(page.getByText(intKey)).toBeVisible()
  await expect(page.getByText('Generic Webhook URL')).toBeVisible()
}
