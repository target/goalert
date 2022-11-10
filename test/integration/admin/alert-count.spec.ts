import { test, expect } from '@playwright/test'
import { userSessionFile } from '../lib'
import Chance from 'chance'
const c = new Chance()

test('should display alert counts', async ({ page, browser }) => {
  //   const now = DateTime.local().minus({ hours: 22 }).toLocaleString({
  //     month: 'short',
  //     day: 'numeric',
  //     hour: 'numeric',
  //     minute: 'numeric',
  //   })
  //   cy.get(`[data-cy="${svc1.name}-${now}"]`).trigger('mouseover', 0, 0, {
  //     force: true,
  //   })
  //   cy.get('[data-cy=alert-count-graph]')
  //     .should('contain', now)
  //     .should('contain', `${svc1.name}: 1`)
  //     .should('contain', `${svc2.name}: 2`)
  //   cy.get('[data-cy=alert-count-table]')
  //     .should('contain', svc1.name)
  //     .should('contain', svc2.name)
})
