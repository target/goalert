import { test, expect, Page } from '@playwright/test'
import { baseURLFromFlags, userSessionFile } from './lib'

test.use({ storageState: userSessionFile })

async function getFlags(page: Page, base: string = ''): Promise<string[]> {
  await page.goto(base + '/api/graphql/explore')

  // We need to wait for the explorer to be "ready"
  // before we can run a query.
  //
  // For that we'll wait for the schema to be loaded.
  await page.click('button[aria-label="Show Documentation Explorer"]')
  await expect(
    page.locator('.graphiql-doc-explorer-section-content'),
  ).toContainText('Mutation')

  // By default, the explorer will open a bunch of comments
  // that we don't need. We'll just open a new tab and
  // type the query we want to run.
  await page.click('button[aria-label="Add tab"]')
  await page.click('.graphiql-editor')
  await page.keyboard.type(`{experimentalFlags`) // trailing curly brace will be added by the autocomplete
  await page.click('button.graphiql-execute-button')

  await expect(page.locator('.result-window')).toContainText(
    'experimentalFlags',
  )

  const res = (await page.innerText('.result-window')).replace(/\s/g, '')

  return JSON.parse(res).data.experimentalFlags
}

test.describe(() => {
  // test a query for the current experimental flags (when example is set)
  test('example experimental flag set', async ({ page, isMobile }) => {
    test.skip(!!isMobile, 'mobile not supported for GraphQL explorer')

    const flags = await getFlags(page, baseURLFromFlags(['example']))
    expect(flags).toContain('example')
  })
})

// test a query for the current experimental flags (when none are set)
test('no experimental flags set', async ({ page, isMobile }) => {
  test.skip(!!isMobile, 'mobile not supported for GraphQL explorer')

  const flags = await getFlags(page)
  expect(flags).toEqual([])
})
