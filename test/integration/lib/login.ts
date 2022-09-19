import { Page } from '@playwright/test'
import path from 'path'

export const adminSessionFile = path.join(
  __dirname,
  '../../../admin.session.json',
)
export const userSessionFile = path.join(
  __dirname,
  '../../../user.session.json',
)

export const adminUserCreds = {
  name: 'Admin McIntegrationFace',
  user: 'admin',
  pass: 'admin123',
}
export const normalUserCreds = {
  name: 'User McIntegrationFace',
  user: 'user',
  pass: 'user1234',
}

export type Creds = typeof adminUserCreds | typeof normalUserCreds

export async function login(
  page: Page,
  user: string,
  pass: string,
): Promise<void> {
  await page.fill('input[name=username]', user)
  await page.fill('input[name=password]', pass)
  await page.click('button[type=submit] >> "Login"')
}

export async function logout(page: Page): Promise<void> {
  // click logout from manage profile
  await page.locator('[aria-label="Manage Profile"]').click()
  await page.locator('button >> "Logout"').click()
}
