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
  } catch (err) {
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
    await page.context().tracing.start({ screenshots: true, snapshots: true })
    try {
      await page.goto('./profile')
      await login(page, c.user, c.pass)
      await expect(page.locator('h1')).toContainText(c.name)
      await page.context().storageState({ path })
    } finally {
      await page.context().tracing.stop({ path: `trace-${c.user}.zip` })
      await page.close()
    }
  }

  await Promise.all([
    createSession(adminSessionFile, adminUserCreds),
    createSession(userSessionFile, normalUserCreds),
  ])

  await browser.close()
}
