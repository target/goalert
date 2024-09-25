import { chromium, FullConfig, expect } from '@playwright/test'
import fs from 'fs'
import {
  adminSessionFile,
  userSessionFile,
  login,
  adminUserCreds,
  normalUserCreds,
  Creds,
} from '../lib'

async function canRead(file: string): Promise<boolean> {
  try {
    await fs.promises.access(file, fs.constants.R_OK)
    return true
  } catch {
    return false
  }
}

export default async function globalSetup(config: FullConfig): Promise<void> {
  // return if both files are readable
  const [adminReadable, userReadable] = await Promise.all([
    canRead(adminSessionFile),
    canRead(userSessionFile),
  ])
  if (adminReadable && userReadable) {
    return
  }

  const browser = await chromium.launch()

  async function createSession(path: string, c: Creds): Promise<void> {
    const page = await browser.newPage({
      baseURL: config.projects[0].use.baseURL,
    })
    try {
      await page.context().tracing.start({ screenshots: true, snapshots: true })
      await page.goto('./profile')
      await login(page, c.user, c.pass)
      await expect(page.getByRole('link', { name: c.name })).toBeVisible({
        timeout: 30000,
      })
      await page.context().storageState({ path })
      await page.context().tracing.stop()
      await page.close()
    } catch (error) {
      await page.context().tracing.stop({
        path: `test-results/failed-setup-${c.user}-trace.zip`,
      })
      await page.close()
      throw error
    }
  }

  await Promise.all([
    createSession(adminSessionFile, adminUserCreds),
    createSession(userSessionFile, normalUserCreds),
  ])

  await browser.close()
}
