import { Page } from '@playwright/test'

export async function createService(
  page: Page,
  name: string,
  description: string,
): Promise<void> {
  await page.goto('./services')
  await page.getByRole('button', { name: 'Create Service' }).click()

  await page.fill('input[name=name]', name)
  await page.fill('textarea[name=description]', description)

  await page.click('[role=dialog] button[type=submit]')
  await page.waitForURL(/services\/[0-9a-f]+/)
}
