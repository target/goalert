import { test, expect } from '@playwright/test'
import Chance from 'chance'
const c = new Chance()

test('update system limit values', async ({ page }) => {
  const cmLimit = c.integer({ min: 15, max: 1000 }).toString()
  const epActionsLimit = c.integer({ min: 15, max: 1000 }).toString()

  console.log('in test')

  await page.goto('/admin/limits')
  page.waitForSelector('nav ul div text=System Limits')

  await page.fill('input[name="ContactMethodsPerUser"]', cmLimit)
  await page.fill('input[name="EPActionsPerStep"]', epActionsLimit)
  await page.getByText('Save').click()

  const dialog = page.getByRole('dialog')
  expect(dialog).toContainText('Apply Configuration Change?')
  expect(dialog).toContainText('+' + cmLimit)
  expect(dialog).toContainText('+' + epActionsLimit)
  await dialog.getByText('Confirm').click()

  await expect(page.locator('input[name="ContactMethodsPerUser"]')).toHaveValue(
    cmLimit,
  )
  await expect(page.locator('input[name="EPActionsPerStep"]')).toHaveValue(
    epActionsLimit,
  )
})

// test('reset pending system limit value changes', async ({ page, browser }) => {
//   const ContactMethodsPerUser = limits.get(
//     'ContactMethodsPerUser',
//   ) as SystemLimits
//   const EPActionsPerStep = limits.get('EPActionsPerStep') as SystemLimits
//   cy.form({
//     ContactMethodsPerUser: c.integer({ min: 0, max: 1000 }).toString(),
//     EPActionsPerStep: c.integer({ min: 0, max: 1000 }).toString(),
//   })
//   cy.get('button[data-cy="reset"]').click()
//   cy.get('input[name="ContactMethodsPerUser"]').should(
//     'have.value',
//     ContactMethodsPerUser.value.toString(),
//   )
//   cy.get('input[name="EPActionsPerStep"]').should(
//     'have.value',
//     EPActionsPerStep.value.toString(),
//   )
// })
