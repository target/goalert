declare global {
  namespace Cypress {
    interface Chainable {
      /** Executes a query directly against the test DB (no results). */
      sql: typeof sql

      /** Fast-forwards the test DB clock by the specified duration and triggers a data refetch after. */
      fastForward: typeof fastForward

      /** Alters the passage of time. */
      setTimeSpeed: typeof setTimeSpeed

      /** Triggers the engine to run and triggers a data refetch after. */
      engineTrigger: typeof engineTrigger

      /** Refetches all data from the backend. */
      refetchAll: typeof refetchAll
    }
  }
}

function refetchAll(): Cypress.Chainable {
  return cy.window().invoke('refetchAll')
}

function fastForward(duration: string): Cypress.Chainable {
  return cy.task('db:fastforward', duration)
}

function sql(query: string): Cypress.Chainable {
  const dbURL =
    Cypress.env('DB_URL') || 'postgres://goalert@localhost:5432?sslmode=disable'

  return cy.exec(`psql-lite -tx -d "$DB" -c "$QUERY"`, {
    env: {
      DB: dbURL,
      QUERY: query,
    },
  })
}

function engineTrigger(): Cypress.Chainable {
  return cy.task('engine:trigger').refetchAll()
}

function setTimeSpeed(speed: number): Cypress.Chainable {
  return cy.task('db:setTimeSpeed', speed)
}

Cypress.Commands.add('sql', sql)
Cypress.Commands.add('fastForward', fastForward)
Cypress.Commands.add('engineTrigger', engineTrigger)
Cypress.Commands.add('refetchAll', refetchAll)
Cypress.Commands.add('setTimeSpeed', setTimeSpeed)

export {}
