import { test, expect } from '@playwright/test'
import { adminSessionFile } from './lib'
import Chance from 'chance'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: adminSessionFile })

// 1. Create a new API key (using query)
// 2. Verify the API key is listed in the table
// 3. Duplicate the API key
// 4. Verify the duplicate API key is listed in the table
// 5. Use the duplicate to delete the original
// 6. Verify the original is no longer listed in the table
// 7. Verify deleting the duplicate using the original fails (key deleted)
// 8. Delete the duplicate via the UI

const query = `
query ListAPIKeys {
    gqlAPIKeys {
        id
        name
    }
}

mutation DeleteAPIKey($id: ID!) {
    deleteGQLAPIKey(id: $id)
}

query ServiceInfo($firstID: ID!) {
  service(id: $firstID) {
    id
  }
}

query ServiceInfo2($secondID: ID!) {
  service(id: $secondID) {
    id
  }
}
`

test('GQL API keys', async ({ page, request, isMobile, baseURL }) => {
  // skip this test if we're running on mobile
  if (isMobile) return

  if (!baseURL) throw new Error('baseURL is required')

  const baseName =
    'apikeytest ' +
    c.string({ length: 12, casing: 'lower', symbols: false, alpha: true })

  // add 3 as a suffix so that the duplicate code will increment it to 4, and we can distinguish it from the original.
  const originalName = baseName + ' 3'
  const duplicateName = baseName + ' 4'

  const descrtiption = c.sentence({ words: 5 })

  await page.goto('./')

  // click on Admin, then API Keys
  await page.click('text=Admin')

  await page.locator('nav').locator('text=API Keys').click()
  await page.click('text=Create API Key')

  await page.fill('[name="name"]', originalName)
  await page.fill('[name="description"]', descrtiption)
  await page.click('[aria-haspopup="listbox"]')
  // click the li with the text "Admin"
  await page.click('li:text("Admin")')

  const editor = page.locator('.cm-editor')
  await editor.click()
  await page.keyboard.insertText(query)

  await page.click('text=Submit')

  // get the token from `[aria-label="Copy"]`
  const originalToken = await page.textContent('[aria-label="Copy"]')

  await page.click('text=Okay')

  // expect we have a <p> tag with the name as the text
  await expect(page.locator('p', { hasText: originalName })).toBeVisible()

  // click on it to open the drawer
  await page.locator('li', { hasText: originalName }).click()

  await page.click('text=Duplicate')
  await page.click('text=Submit')

  const duplicateToken = await page.textContent('[aria-label="Copy"]')

  await page.click('text=Okay')

  await expect(page.locator('li', { hasText: duplicateName })).toBeVisible()

  const gqlURL = baseURL.replace(/\/$/, '') + '/api/graphql'
  let resp = await request.post(gqlURL, {
    headers: {
      Authorization: `Bearer ${duplicateToken}`,
    },
    data: { query, operationName: 'ListAPIKeys' },
  })

  expect(resp.status()).toBe(200)
  const data = await resp.json()
  expect(data).toHaveProperty('data')
  expect(data).not.toHaveProperty('errors')
  expect(data.data).toHaveProperty('gqlAPIKeys')

  // Reproduce issue #3662
  resp = await request.post(gqlURL, {
    headers: {
      Authorization: `Bearer ${duplicateToken}`,
    },
    data: {
      variables: {
        firstID: '00000000-0000-0000-0000-000000000000',
      },
      operationName: 'ServiceInfo',
    },
  })
  expect(resp.status()).toBe(200)
  expect(await resp.json()).not.toHaveProperty('errors')

  resp = await request.post(gqlURL, {
    headers: {
      Authorization: `Bearer ${duplicateToken}`,
    },
    data: { query: '{wrongQuery}' },
  })

  expect(resp.status()).toBe(200)
  const badResp = await resp.json()

  expect(badResp).toHaveProperty('errors')

  type Key = {
    id: string
    name: string
  }

  const originalID = data.data.gqlAPIKeys.find(
    (key: Key) => key.name === originalName,
  ).id
  const duplicateID = data.data.gqlAPIKeys.find(
    (key: Key) => key.name === duplicateName,
  ).id
  expect(originalID).toBeDefined()
  expect(duplicateID).toBeDefined()

  // Delete the original using the duplicate via fetch call
  resp = await request.post(gqlURL, {
    headers: {
      Authorization: `Bearer ${duplicateToken}`,
    },
    data: {
      // Note: `query` is omitted to validate that it is not required for gql API keys.
      operationName: 'DeleteAPIKey',
      variables: {
        id: originalID,
      },
    },
  })

  expect(resp.status()).toBe(200)
  expect(await resp.json()).not.toHaveProperty('errors')

  await page.reload()

  await expect(page.locator('li', { hasText: originalName })).not.toBeVisible()

  // Attempt to delete the duplicate using the original via fetch call
  resp = await request.post(gqlURL, {
    headers: {
      Authorization: `Bearer ${originalToken}`,
    },
    data: {
      query,
      operationName: 'DeleteAPIKey',
      variables: {
        id: duplicateID,
      },
    },
  })

  // expect the delete to fail, since the original was already deleted, and can no longer be used
  expect(resp.status()).toBe(401)

  // Delete the duplicate via the UI menu
  // find a div with a <p> tag with the text duplicateName, then click [aria-label="Other Actions"]
  await page
    .locator('li', { hasText: duplicateName })
    .locator('[aria-label="Other Actions"]')
    .click()
  await page.locator('[role=menuitem]', { hasText: 'Delete' }).click()
  await page.click('text=Confirm')

  // expect the duplicate to be gone
  await expect(page.locator('li', { hasText: duplicateName })).not.toBeVisible()
})
