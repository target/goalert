declare namespace Cypress {
  interface Chainable {
    /** Executes a query directly against the test DB (no results). */
    sql: typeof sql
  }
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
