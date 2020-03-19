// eslint-disable-next-line @typescript-eslint/no-unused-vars
namespace Cypress {
  interface Chainable {
    login: typeof login
    adminLogin: typeof adminLogin
  }
}
