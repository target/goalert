declare namespace Cypress {
  interface Chainable {
    login: typeof login
    adminLogin: typeof adminLogin
  }
}
