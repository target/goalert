import { type Page, test, expect } from '@playwright/test'
import libCoverage from 'istanbul-lib-coverage'
import libReport from 'istanbul-lib-report'
import reports from 'istanbul-reports'
import v8toIstanbul from 'v8-to-istanbul'
import { promises as fs } from 'fs'

let page: Page

test.beforeAll(async ({ browser }) => {
  page = await browser.newPage()
  await page.goto('')
  await page.coverage.startJSCoverage()
})

export async function saveV8Coverage(page: Page): Promise<void> {
  const coverage = await page.coverage.stopJSCoverage()
  const map = libCoverage.createCoverageMap()

  for (const entry of coverage) {
    if (entry.url === '') {
      continue
    }

    const scriptPath = `test${new URL(entry.url).pathname}`
    const converter = v8toIstanbul(
      scriptPath,
      0,
      { source: entry?.source ?? '' },
      (filepath) => {
        const normalized = filepath.replace(/\\/g, '/')
        const ret = normalized.includes('node_modules/')
        return ret
      },
    )

    await converter.load()
    converter.applyCoverage(entry.functions)

    const data = converter.toIstanbul()
    map.merge(data)
  }

  await fs.rm('coverage', { force: true, recursive: true })
  const context = libReport.createContext({ coverageMap: map })
  reports.create('html').execute(context)
}

test.afterAll(async () => {
  await saveV8Coverage(page)
  await page.close()
})

export { test, expect }
