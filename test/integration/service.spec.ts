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

  // Create a label and value for the service
  const key = `${c.word({ length: 4 })}/${c.word({ length: 3 })}`
  const value = c.word({ length: 8 })
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

  // Return to the service
  if (isMobile) {
    await page.getByRole('button', { name: 'Back' }).click()
  } else {
    await page.getByRole('link', { name: name, exact: true }).click()
  }

  // Make an integration key
  const intKey = c.word({ length: 5 }) + ' Key'
  await page.getByRole('link', { name: 'Integration Keys' }).click()
  if (isMobile) {
    await page.getByRole('button', { name: 'Create Integration Key' }).click()
  } else {
    await page.getByTestId('create-key').click()
  }
  await page.getByLabel('Name').fill(intKey)
  await page.getByRole('button', { name: 'Submit' }).click()

  // Make another service
  const diffName = 'pw-service ' + c.name()
  const diffDescription = c.sentence()

  await page.goto('./services')
  await page.getByRole('button', { name: 'Create Service' }).click()

  await page.fill('input[name=name]', diffName)
  await page.fill('textarea[name=description]', diffDescription)

  await page.click('[role=dialog] button[type=submit]')

  // Set the label with the existing key and a new value
  const diffValue = c.word({ length: 8 })
  await page.getByRole('link', { name: 'Labels' }).click()
  if (isMobile) {
    await page.getByRole('button', { name: 'Add' }).click()
  } else {
    await page.getByTestId('create-label').click()
  }

  await page.getByLabel('Key', { exact: true }).fill(key)
  await page.getByRole('option', { name: key }).getByRole('listitem').click()
  await page.getByLabel('Value', { exact: true }).fill(diffValue)
  await page.click('[role=dialog] button[type=submit]')

  await expect(page.getByText(key)).toBeVisible()
  await expect(page.getByText(diffValue)).toBeVisible()

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

  // Should not allow filtering by label value without a key selected
  await expect(page.getByLabel('Select Label Value')).toBeDisabled()

  // Filter by label key
  await page.getByLabel('Select Label Key').click()
  await page.getByRole('option', { name: key }).getByRole('listitem').click()
  await page.getByRole('button', { name: 'Done' }).click()

  // Check if filtered, should have found both services
  await expect(
    page.getByRole('link', { name: name + ' ' + description }),
  ).toBeVisible()
  await expect(
    page.getByRole('link', { name: diffName + ' ' + diffDescription }),
  ).toBeVisible()

  // Filter by key and the first service's value
  await page.getByRole('button', { name: 'Search Services by Filters' }).click()
  await page.getByLabel('Select Label Value').click()
  await page.getByRole('option', { name: value }).getByRole('listitem').click()
  await page.getByRole('button', { name: 'Done' }).click()

  // Check if filtered, should have found only the first service
  await expect(
    page.getByRole('link', { name: name + ' ' + description }),
  ).toBeVisible()
  await expect(
    page.getByRole('link', { name: diffName + ' ' + diffDescription }),
  ).not.toBeVisible()

  // Reset filters
  await page.getByRole('button', { name: 'Search Services by Filters' }).click()
  await page.getByRole('button', { name: 'Reset' }).click()

  // Filter by integration key, should find the service
  await page.getByLabel('Select Integration Key').click()
  await page.getByRole('option', { name: intKey }).getByRole('listitem').click()
  await page.getByRole('button', { name: 'Done' }).click()
  await expect(
    page.getByRole('link', { name: name + ' ' + description }),
  ).toBeVisible()

  // Load in filters from URL, should find the service
  await page.goto('./services?search=' + key + '=*')
  await expect(
    page.getByRole('link', { name: name + ' ' + description }),
  ).toBeVisible()

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
