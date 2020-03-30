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

/*
 * screen: screen size
 * pageName: name of title when on main details page
 * targetName: name of schedule, service, etc when on an information card route
 * detailsName: name of route that is being viewed
 * route: actual route to verify
 */
function navigateToAndFrom(
  screen: string,
  pageName: string, // details page title
  targetName: string, // item name/title
  detailsName: string, // sub page title
  route: string,
): void {
  // navigate to extended details view
  cy.get('[data-cy=app-bar]').should('contain', pageName)
  cy.get('ul[data-cy="route-links"] li')
    .contains(detailsName)
    .click()

  // verify url
  cy.url().should('include', route)

  if (screen === 'widescreen') {
    // verify on new view
    cy.get('[data-cy=app-bar]')
      .should('contain', targetName)
      .should('contain', detailsName)

      // navigate back to details page
      .contains(targetName)
      .click()

    // verify back on details page
    cy.get('[data-cy=app-bar]').should('contain', pageName)
  } else if (screen === 'mobile' || screen === 'tablet') {
    // verify on new view
    cy.get('[data-cy=app-bar]')
      .should('contain', detailsName)
      .should('not.contain', targetName)

    // navigate back to details page
    cy.get('button[data-cy=nav-back-icon]').click()

    // verify back on details page
    cy.get('[data-cy=app-bar]').should('contain', pageName)

    if (!route.includes('profile')) {
      cy.get('[data-cy=app-bar]').should('not.contain', targetName)
    }
  }
}

Cypress.Commands.add('navigateToAndFrom', navigateToAndFrom)

export {}
