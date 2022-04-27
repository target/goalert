declare global {
  namespace Cypress {
    interface Chainable {
      /** Executes a query directly against the test DB (no results). */
      sql: typeof sql

      /** Fast-forwards the test DB clock by the specified duration. */
      fastForward: typeof fastForward
    }
  }
}

function fastForward(duration: string): Cypress.Chainable {
  return cy.task('db:fastforward', duration).task('engine:trigger')
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

Cypress.Commands.add('sql', sql)
Cypress.Commands.add('fastForward', fastForward)

export {}
