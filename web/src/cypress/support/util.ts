declare global {
  type ScreenFormat = 'mobile' | 'tablet' | 'widescreen'
}

export function screen(): ScreenFormat {
  const width = Cypress.config().viewportWidth
  if (width < 600) return 'mobile'
  if (width < 960) return 'tablet'

  return 'widescreen'
}

export function screenName(): string {
  switch (screen()) {
    case 'mobile':
      return 'Mobile'
    case 'tablet':
      return 'Tablet'
  }

  return 'Wide'
}

export function testScreen(
  label: string,
  fn: (screen: ScreenFormat) => void,
  skipLogin = false,
  adminLogin = false,
) {
  describe(label, () => {
    before(() => {
      Cypress.Cookies.debug(true)
    })
    if (!skipLogin) {
      before(() => cy.resetConfig()[adminLogin ? 'adminLogin' : 'login']())
      it(adminLogin ? 'admin login' : 'login', () => {}) // required due to mocha skip bug
      beforeEach(() => Cypress.Cookies.preserveOnce('goalert_session.2'))
    }
    describe(screenName(), () => fn(screen()))
  })
}
