import { test, expect } from '@playwright/test'
import { userSessionFile } from './lib'
import Chance from 'chance'
import { createService } from './lib/service'
const c = new Chance()

test.describe.configure({ mode: 'parallel' })
test.use({ storageState: userSessionFile })

test('Service Information', async ({ page }) => {
  let name = 'pw-service ' + c.name()
  const description = c.sentence()

  await createService(page, name, description)

  //We should be on the details page, so let's try editing it after validating the data on the page.

  //We should have a heading with the service name
  await expect(page.getByRole('heading', { name, level: 1 })).toBeVisible()

  //and the breadcrumb link
  await expect(page.getByRole('link', { name, exact: true }))toBeVisible()

  // Lastly ensure there is a link to a policy named "<name> Policy"
  await expect(page.getByRole('link', { name: name + ' Policy' })).toBeVisible()
})

test('Service Editing', async ({ page }) => {
  await page.getByRole('button', { name: 'Edit' }).click()

  name = 'pw-service ' + c.name()
  await page.fill('input[name=name]', name)
  await page.click('[role=dialog] button[type=submit]')

  await expect(page.getByRole('heading', { name, level: 1})).toBeVisible()
})

test('Heartbeat Monitors', async ({ page, isMobile }) => {
  // Navigate to the heartbeat monitors
  await page.getByRole('link', { name: 'Heartbeat Monitors' }).click()

  // Cancel out of create
  if (isMobile) {
    await page.getByRole('button', { name: 'Create Heartbeat Monitor' }).click()
  } else {
    await page.getByTestId('create-monitor').click()
  }
  await page.getByRole('button', { name: 'Cancel' }).click()

  // Create a heartbeat monitor using invalid name
  let timeoutMinutes = (Math.trunc(Math.random() * 10) + 5).toString()
  const invalidHMName = 'a'
  if (isMobile) {
    await page.getByRole('button', { name: 'Create Heartbeat Monitor' }).click()
  } else {
    await page.getByTestId('create-monitor').click()
  }
  await page.getByLabel('Name').fill(invalidHMName)
  await page.getByLabel('Timeout').fill(timeoutMinutes)
  await page.getByRole('button', { name: 'Submit' }).click()

  // Should see error message
  await expect(page.getByText('Must be at least 2 characters')).toBeVisible()

  // Use valid name instead
  let hmName = c.word({ length: 5 }) + ' Monitor'
  await page.getByLabel('Name').fill(hmName)
  await page.getByRole('button', { name: 'Retry' }).click()

  // Should see the heartbeat monitor created
  await expect(page.getByText(hmName)).toBeVisible()
  await expect(page.getByText(timeoutMinutes)).toBeVisible()

  // Cancel out of edit
  await page.getByRole('button', { name: 'Other Actions' }).click()
  await page.getByRole('menuitem', { name: 'Edit' }).click()
  await page.getByRole('button', { name: 'Cancel' }).click()

  // Edit the heartbeat monitor
  hmName = c.word({ length: 5 })
  timeoutMinutes = (Math.trunc(Math.random() * 10) + 5).toString()
  await page.getByRole('button', { name: 'Other Actions' }).click()
  await page.getByRole('menuitem', { name: 'Edit' }).click()
  await page.getByLabel('Name').fill(hmName)
  await page.getByLabel('Timeout').fill(timeoutMinutes)
  await page.getByRole('button', { name: 'Submit' }).click()

  // Should see the edited heartbeat monitor
  await expect(page.getByText(hmName)).toBeVisible()
  await expect(page.getByText(timeoutMinutes)).toBeVisible()

  // Cancel out of delete
  await page.getByRole('button', { name: 'Other Actions' }).click()
  await page.getByRole('menuitem', { name: 'Delete' }).click()
  await page.getByRole('button', { name: 'Cancel' }).click()

  // Delete the heartbeat monitor
  await page.getByRole('button', { name: 'Other Actions' }).click()
  await page.getByRole('menuitem', { name: 'Delete' }).click()
  await page.getByRole('button', { name: 'Confirm' }).click()
  await page.getByText('No heartbeat monitors exist for this service.').click()
})

test('Alerts', async ({ page }) => {
  // Return to the service
  if (isMobile) {
    await page.getByRole('button', { name: 'Back' }).click()
  } else {
    await page.getByRole('link', { name, exact: true }).click()
  }

  // Go to the alerts page
  await page
    .getByRole('link', {
      name: 'Alerts Manage alerts specific to this service',
    })
    .click()

  // Create an alert
  const summary = c.sentence({ words: 3 })
  const details = c.word({ length: 10 })
  await page.getByRole('button', { name: 'Create Alert' }).click()
  await page.getByLabel('Alert Summary').fill(summary)
  await page.getByLabel('Details (optional)').fill(details)
  await page.getByRole('button', { name: 'Next' }).click()
  if (isMobile) {
    await expect(page.getByText('Selected Services (1)' + name)).toBeVisible()
  } else {
    await expect(
      page.getByRole('dialog', { name: 'Create New Alert' }).getByText(name),
    ).toBeVisible()
  }
  await page.getByRole('button', { name: 'Submit' }).click()
  await page.getByRole('button', { name: 'Done' }).click()

  // Alert should be unacknowledged
  await expect(
    page.getByRole('link', { name: ' UNACKNOWLEDGED ' + summary }),
  ).toBeVisible()

  // Acknowledge the alert
  await page.getByRole('button', { name: 'Acknowledge All' }).click()
  await page.getByRole('button', { name: 'Confirm' }).click()
  await expect(
    page.getByRole('link', { name: ' ACKNOWLEDGED ' + summary }),
  ).toBeVisible()
  await expect(
    page.getByRole('link', { name: ' UNACKNOWLEDGED ' + summary }),
  ).toBeHidden()

  // Close the alert
  await page.getByRole('button', { name: 'Close All' }).click()
  await page.getByRole('button', { name: 'Confirm' }).click()
  await expect(page.getByText('No results')).toBeVisible()
})

test('Metric', async ({ page, isMobile}) => {
  // Return to the service
  if (isMobile) {
    await page.getByRole('button', { name: 'Back' }).click()
  } else {
    await page.getByRole('link', { name, exact: true }).click()
  }

  // Navigate to the metrics
  await page.getByRole('link', { name: 'Metrics' }).click()

  // Return to the service
  if (isMobile) {
    await page.getByRole('button', { name: 'Back' }).click()
  } else {
    await page.getByRole('link', { name, exact: true }).click()
  }
})

test('Integration Key', async ({ page, isMobile }) => {
  // Return to the service
  if (isMobile) {
    await page.getByRole('button', { name: 'Back' }).click()
  } else {
    await page.getByRole('link', { name, exact: true }).click()
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

  await expect(page.getByText(intKey)).toBeVisible()
  await expect(page.getByText('Generic Webhook URL')).toBeVisible()

  // Create a second integration key with a different type
  const grafanaKey = c.word({ length: 5 }) + ' Key'
  if (isMobile) {
    await page.getByRole('button', { name: 'Create Integration Key' }).click()
  } else {
    await page.getByTestId('create-key').click()
  }
  await page.getByLabel('Name').fill(grafanaKey)
  await page.getByRole('combobox', { name: 'Generic API' }).click()
  await page.getByRole('option', { name: 'Grafana' }).click()
  await page.getByRole('button', { name: 'Submit' }).click()

  await expect(page.getByText(grafanaKey)).toBeVisible()
  await expect(page.getByText('Grafana')).toBeVisible()

  // Delete the second integration key, confirm it is no longer visible
  await page
    .getByRole('listitem')
    .filter({ hasText: grafanaKey })
    .getByRole('button')
    .click()
  await page.getByRole('button', { name: 'Confirm' }).click()

  await expect(page.getByText(intKey, { exact: true })).toBeVisible()
  await expect(page.getByText(grafanaKey, { exact: true })).toBeHidden()
})

test('Service Creation with Existing Label', async ({ page, isMobile }) => {
  // Make another service
  const diffName = 'pw-service ' + c.name()
  const diffDescription = c.sentence()

  await createService(page, diffName, diffDescription)

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
})

test('Label Filtering', async ({ page, isMobile }) => {
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
  // await page.getByLabel('Select Label Key').click()
  await page.getByRole('combobox', { name: 'Select Label Key' }).fill(key)
  await page.getByText(key).click()
  // await page.getByRole('option', { name: key }).getByRole('listitem').click()
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
})

test('Service Search', async ({ page }) => {
  // We should be on the services list page, so let's search for service by label
  await page.fill('input[name=search]', key + '=' + value)
  await page.getByPlaceholder('Search').press('Enter')

  // We should see the service on the page
  await expect(page.getByText(name, { exact: true })).toBeVisible()

  // Search for service without a label
  await page.fill('input[name=search]', key + '!=' + value)

  // We should not see the service on the page
  await expect(page.getByText(name, { exact: true })).toBeHidden()

  // Try searching for the service by its name. We add a space to the beginning of the name to ensure we are searching for the full name and not a substring.
  await page.fill('input[name=search]', ' ' + name + ' ')

  // We should find the service in the list, lets go to it
  await page.getByRole('link', { name }).click()
})

test('Maintenance Mode', async ({ page }) => {
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
})

test('Service Deletion', async ({ page, isMobile }) => {
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
