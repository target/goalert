import axe from 'axe-core/axe.js'

declare global {
  namespace Cypress {
    interface Chainable {
      /** Test the accessibility of the current state of the page */
      validateA11y: typeof validateA11y
    }
  }
}

// no selector provided will result in the entire page being validated
function validateA11y(selector = 'main[id="content"]'): void {
  cy.window().then((win: Cypress.AUTWindow & { _axeInjected?: boolean }) => {
    if (win._axeInjected) return
    win._axeInjected = true

    // cy.injectAxe() exists, but injecting source manually for CI
    // https://github.com/component-driven/cypress-axe/issues/118
    return cy.window({ log: false }).then((window) => {
      const script = window.document.createElement('script')
      script.innerHTML = axe.source
      window.document.head.appendChild(script)
    })
  })

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
    ],
  })
}

Cypress.Commands.add('validateA11y', validateA11y)

export {}
