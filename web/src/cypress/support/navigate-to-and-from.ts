declare global {
  namespace Cypress {
    interface Chainable {
      /**
       * Navigate to an extended details page
       * and verify navigating back to main
       * details page
       */
      navigateToAndFrom: typeof navigateToAndFrom
    }
  }
}

function startCase(str: string): string {
  return str
    .split(' ')
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(' ')
}

/*
 * screen: screen size
 * pageName: name of title when on main details page
 * targetName: name of schedule, service, etc when on an information card route
 * linkName: name of route that is being viewed
 * route: actual route to verify
 */
function navigateToAndFrom(
  screen: string,
  _pageName: string, // details page title
  _targetName: string, // item name/title
  _linkName: string, // sub page title
  route: string,
): void {
  const pageName = startCase(_pageName)
  const targetName = _targetName // target name formatting should be preserved
  const linkName = startCase(_linkName).replace('On Call', 'On-Call')

  // navigate to extended details view
  cy.get(`[data-cy=breadcrumb-0]`).should('include.text', pageName)
  cy.get(`[data-cy=breadcrumb-1]`).should('include.text', targetName)
  cy.get('ul[data-cy="route-links"] li').contains(linkName).click()
  cy.get(`[data-cy=breadcrumb-2]`).should('include.text', linkName)

  // verify url
  cy.url().should('include', route)

  if (screen === 'widescreen') {
    cy.get(`[data-cy=breadcrumb-1]`)
      // navigate back to details page
      .click()

    // verify back on details page
    cy.get(`[data-cy=breadcrumb-2]`).should('not.exist')
  } else if (screen === 'mobile' || screen === 'tablet') {
    cy.get(`[data-cy=breadcrumb-1]`).should('not.exist')

    // navigate back to details page
    cy.get('button[data-cy=nav-back-icon]').click()
    cy.get(`[data-cy=breadcrumb-1]`).should('be.visible')
  }
}

Cypress.Commands.add('navigateToAndFrom', navigateToAndFrom)

export {}
