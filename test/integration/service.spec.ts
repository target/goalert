import { test, expect } from '@playwright/test'
import { userSessionFile } from './lib'
import Chance from 'chance'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: userSessionFile })

// test create, verify, and delete of an EMAIL contact method
test('Service', async ({ page, browser, isMobile }) => {
  let name = 'pw-service ' + c.name()
  const description = c.sentence()

  await page.goto('./services')
  await page.getByRole('button', { name: 'Create Service' }).click()

  await page.fill('input[name=name]', name)
  await page.fill('textarea[name=description]', description)

  await page.click('[role=dialog] button[type=submit]')

  // We should be on the details page, so let's try editing it after validating the data on the page.

  // We should have a heading with the service name
  await expect(
    page.getByRole('heading', { name: name, level: 1 }),
  ).toBeVisible()

  // and the breadcrumb link
  await expect(
    page.getByRole('link', { name: name, exact: true }),
  ).toBeVisible()

  // We should also find the description on the page
  await expect(page.getByText(description)).toBeVisible()

  // Lastly ensure there is a link to a policy named "<name> Policy"
  await expect(page.getByRole('link', { name: name + ' Policy' })).toBeVisible()

  // Now let's edit the service name
  await page.getByRole('button', { name: 'Edit' }).click()

  name = 'pw-service ' + c.name()
  await page.fill('input[name=name]', name)
  await page.click('[role=dialog] button[type=submit]')

  await expect(
    page.getByRole('heading', { name: name, level: 1 }),
  ).toBeVisible()

  await page.goto('./services')

  // We should be on the services list page, so let's try searching for the service we just created. We add a space to the beginning of the name to ensure we are searching for the full name and not a substring.
  await page.fill('input[name=search]', ' ' + name + ' ')

  // We should find the service in the list, lets go to it
  await page.getByRole('link', { name: name }).click()

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
  await page.fill('input[name=search]', ' ' + name + ' ')

  // We should see "No results" on the page
  await expect(page.getByText('No results')).toBeVisible()
})
