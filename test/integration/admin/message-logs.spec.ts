import { test, expect } from '@playwright/test'
import { userSessionFile } from '../lib'
import Chance from 'chance'
const c = new Chance()

test('view the logs list with one log', async ({ page, browser }) => {
  //   cy.get('[data-cy="paginated-list"]').as('list')
  //   cy.get('@list').should('have.length', 1)
  //   cy.get('@list')
  //     .eq(0)
  //     .should(
  //       'contain.text',
  //       DateTime.fromISO(debugMessage.createdAt).toFormat('fff'),
  //     )
  //   cy.get('@list')
  //     .eq(0)
  //     .should('contain.text', debugMessage.type + ' Notification')
  //   // todo: destination not supported: phone number value is pre-formatted
  //   // likely need to create a support function cy.getPhoneNumberInfo thru
  //   // gql to verify this info.
  //   cy.get('@list').eq(0).should('contain.text', debugMessage.serviceName)
  //   cy.get('@list').eq(0).should('contain.text', debugMessage.userName)
  //   cy.get('@list').eq(0).should('include.text', debugMessage.status) // "Failed" or "Failed (Permanent)" can exist
})

test('select and view a logs details', async ({ page, browser }) => {
  //   cy.get('[data-cy="paginated-list"]').eq(0).click()
  //   cy.get('[data-cy="debug-message-details"').as('details').should('exist')
  //   // todo: not asserting updatedAt, destination, or providerID
  //   cy.get('@details').should('contain.text', 'ID')
  //   cy.get('@details').should('contain.text', debugMessage.id)
  //   cy.get('@details').should('contain.text', 'Created At')
  //   cy.get('@details').should(
  //     'contain.text',
  //     DateTime.fromISO(debugMessage.createdAt).toFormat('fff'),
  //   )
  //   cy.get('@details').should('contain.text', 'Notification Type')
  //   cy.get('@details').should('contain.text', debugMessage.type)
  //   cy.get('@details').should('contain.text', 'Current Status')
  //   cy.get('@details').should('include.text', debugMessage.status)
  //   cy.get('@details').should('contain.text', 'User')
  //   cy.get('@details').should('contain.text', debugMessage.userName)
  //   cy.get('@details').should('contain.text', 'Service')
  //   cy.get('@details').should('contain.text', debugMessage.serviceName)
  //   cy.get('@details').should('contain.text', 'Alert')
  //   cy.get('@details').should('contain.text', debugMessage.alertID)
})

test('verify user link from a logs details', async ({ page, browser }) => {
  //   cy.get('[data-cy="paginated-list"]').eq(0).click()
  //   cy.get('[data-cy="debug-message-details"')
  //     .find('a')
  //     .contains(debugMessage?.userName ?? '')
  //     .should('have.attr', 'href', pathPrefix() + '/users/' + debugMessage.userID)
  //     .should('have.attr', 'target', '_blank')
  //     .should('have.attr', 'rel', 'noopener noreferrer')
})

test('verify service link from a logs details', async ({ page, browser }) => {
  //   cy.get('[data-cy="paginated-list"]').eq(0).click()
  //   cy.get('[data-cy="debug-message-details"')
  //     .find('a')
  //     .contains(debugMessage?.serviceName ?? '')
  //     .should(
  //       'have.attr',
  //       'href',
  //       pathPrefix() + '/services/' + debugMessage.serviceID,
  //     )
  //     .should('have.attr', 'target', '_blank')
  //     .should('have.attr', 'rel', 'noopener noreferrer')
})
