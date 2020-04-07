declare namespace Cypress {
  interface Chainable {
    /** Enter a page-level search (from the top bar). Works in mobile and widescreen. */
    pageSearch: typeof pageSearch
  }
}
