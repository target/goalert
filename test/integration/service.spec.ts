import { test, expect } from '@playwright/test'
import { userSessionFile } from './lib'
import Chance from 'chance'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: userSessionFile })

// test create, verify, and delete of an EMAIL contact method
test('Service', async ({ page, isMobile }) => {
  let name = 'pw-service ' + c.name()
  const description = c.sentence()

  await page.goto('./services')
  await page.getByRole('button', { name: 'Create Service' }).click()

  await page.fill('input[name=name]', name)
  await page.fill('textarea[name=description]', description)

  await page.click('[role=dialog] button[type=submit]')

  // We should be on the details page, so let's try editing it after validating the data on the page.

  // We should have a heading with the service name
  await expect(page.getByRole('heading', { name, level: 1 })).toBeVisible()

  // and the breadcrumb link
  await expect(page.getByRole('link', { name, exact: true })).toBeVisible()

  // We should also find the description on the page
  await expect(page.getByText(description)).toBeVisible()

  // Lastly ensure there is a link to a policy named "<name> Policy"
  await expect(page.getByRole('link', { name: name + ' Policy' })).toBeVisible()

  // Now let's edit the service name
  await page.getByRole('button', { name: 'Edit' }).click()

  name = 'pw-service ' + c.name()
  await page.fill('input[name=name]', name)
  await page.click('[role=dialog] button[type=submit]')

  await expect(page.getByRole('heading', { name, level: 1 })).toBeVisible()

  // Create a label and value
  const key = `${c.word({ length: 4 })}/${c.word({ length: 3 })}`
  const value = c.word({ length: 8 })
  await page
    .getByRole('link', { name: 'Labels Group together services' })
    .click()
  if (isMobile) {
    await page.getByRole('button', { name: 'Add' }).click()
  } else {
    await page.getByTestId('create-label').click()
  }

  await page.getByPlaceholder('Start typing...').fill(key)
  await page.getByText('Create "' + key + '"').click()
  await page.getByLabel('Value', { exact: true }).fill(value)
  await page.getByRole('button', { name: 'Submit' }).click()

  // return to service
  if (isMobile) {
    await page.getByRole('button', { name: 'Back' }).click()
  } else {
    await page.getByRole('link', { name: name, exact: true }).click()
  }

  // make integration key
  const intKey = c.word({ length: 5 }) + ' Key'
  await page
    .getByRole('link', {
      name: 'Integration Keys Manage keys used to create alerts',
    })
    .click()
  if (isMobile) {
    await page.getByRole('button', { name: 'Create Integration Key' }).click()
  } else {
    await page.getByTestId('create-key').click()
  }
  await page.getByLabel('Name').click()
  await page.getByLabel('Name').fill(intKey)
  await page.getByRole('button', { name: 'Submit' }).click()

  await page.goto('./services')

  // Check that filter content doesn't exist yet
  await expect(
    page.getByRole('button', { name: 'filter-done' }),
  ).not.toBeVisible()

  // Open filter
  if (isMobile) {
    await page.getByRole('button', { name: 'Search' }).click()
  }
  await page.getByRole('button', { name: 'Search Services by Filters' }).click()

  // check exists? or is that redundant

  // check that can't filter by value without a key
  await expect(page.getByLabel('Select Label Value')).toBeDisabled()

  // filter by label key
  await page
    .locator('div')
    .filter({ hasText: /^Select Label Key$/ })
    .getByRole('button', { name: 'Open' })
    .click()
  await page.getByRole('option', { name: key }).getByRole('listitem').click()
  await page.getByRole('button', { name: 'Done' }).click()

  // check if filtered?
  await expect(
    page.getByRole('link', { name: name + ' ' + description }),
  ).toBeVisible()

  // filter with key and value
  await page.getByRole('button', { name: 'Search Services by Filters' }).click()

  await page
    .locator('div')
    .filter({ hasText: /^Select Label Value$/ })
    .getByRole('button', { name: 'Open' })
    .click()
  await page.getByRole('option', { name: value }).getByRole('listitem').click()
  await page.getByRole('button', { name: 'Done' }).click()

  await expect(
    page.getByRole('link', { name: name + ' ' + description }),
  ).toBeVisible()

  // reset filters?
  await page.getByRole('button', { name: 'Search Services by Filters' }).click()
  await page.getByRole('button', { name: 'Reset' }).click()

  // filter by integration key
  // await page.getByRole('button', { name: 'Search Services by Filters' }).click()
  await page
    .locator('div')
    .filter({ hasText: /^Select Integration Key$/ })
    .getByRole('button', { name: 'Open' })
    .click()
  await page.getByRole('option', { name: intKey }).getByRole('listitem').click()
  await page.getByRole('button', { name: 'Done' }).click()
  // remove done?
  await expect(
    page.getByRole('link', { name: name + ' ' + description }),
  ).toBeVisible()

  // reset filters?
  await page.getByRole('button', { name: 'Search Services by Filters' }).click()
  await page.getByRole('button', { name: 'Reset' }).click()

  // load in filters from URL
  await page.goto('./services?search=' + key + '=*')
  await expect(
    page.getByRole('link', { name: name + ' ' + description }),
  ).toBeVisible()

  // close the filter
  // check its gone?

  // We should be on the services list page, so let's try searching for the service we just created. We add a space to the beginning of the name to ensure we are searching for the full name and not a substring.
  await page.fill('input[name=search]', ' ' + name + ' ')

  // We should find the service in the list, lets go to it
  await page.getByRole('link', { name }).click()

  // Maintenance mode
  await page.getByRole('button', { name: 'Maintenance Mode' }).click()

  // We should be in the Set Maintenance Mode dialog
  await expect(
    page.getByRole('heading', { name: 'Set Maintenance Mode' }),
  ).toBeVisible()

  // Submit
  await page.click('[role=dialog] button[type=submit]')

  // We should be back on the details page, but with an alert at the top saing In Maintenance Mode
  await expect(page.getByRole('alert')).toContainText('In Maintenance Mode')
  // Hit the cancel button in the banner
  await page.getByRole('button', { name: 'Cancel' }).click()

  // We should be back on the details page, but with no more banner
  await expect(page.getByRole('alert')).not.toBeVisible()

  // Finally, let's delete the service
  await page.getByRole('button', { name: 'Delete' }).click()
  // and confirm
  await page.click('[role=dialog] button[type=submit]')

  // we should be back on the services list page, so let's search for the service we just deleted
  if (isMobile) {
    await page.getByRole('button', { name: 'Search' }).click()
  }
  await page.fill('input[name=search]', ' ' + name + ' ')

  // We should see "No results" on the page
  await expect(page.getByText('No results')).toBeVisible()
})
