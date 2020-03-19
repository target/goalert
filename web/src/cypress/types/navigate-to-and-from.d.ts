// eslint-disable-next-line @typescript-eslint/no-unused-vars
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
