const CY: any = Cypress
// Fail-fast-all-files
before(function() {
  cy.getCookie('has-failed-test').then(cookie => {
    if (cookie && typeof cookie === 'object' && cookie.value === 'true') {
      CY.runner.stop()
    }
  })
})

// Fail-fast-single-file
afterEach(function() {
  if (this.currentTest && this.currentTest.state === 'failed') {
    cy.setCookie('has-failed-test', 'true')
    CY.runner.stop()
  }
})

export {}
