declare global {
  namespace Cypress {
    interface Chainable {
      /** Test the accessibility of the current state of the page */
      validateA11y: typeof validateA11y
    }
  }
}

// no selector provided will result in the entire page being checked
// inject should be true once per page, after the cy.visit call
function validateA11y(selector = 'main[id="content"]', inject = true): void {
  if (inject) cy.injectAxe()
  cy.checkA11y(selector, {
    includedImpacts: ['critical'], // only report and assert for critical impact items
    runOnly: [
      'wcag2a',
      'wcag2aa',
      'wcag2aaa',
      'wcag21a',
      'wcag21aa',
      'wcag21aaa',
      'best-practice',
      'wcag***',
    ],
  })
}

Cypress.Commands.add('validateA11y', validateA11y)

export {}
