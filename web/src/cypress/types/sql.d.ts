declare namespace Cypress {
  interface Chainable {
    /** Executes a query directly against the test DB (no results). */
    sql: typeof sql
  }
}
