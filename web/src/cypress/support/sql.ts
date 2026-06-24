declare global {
  namespace Cypress {
    interface Chainable {
      /** Executes a query directly against the test DB (no results). */
      sql: typeof sql

      /** Fast-forwards the test DB clock by the specified duration and triggers a data refetch after. */
      fastForward: typeof fastForward

      /** Triggers the engine to run and triggers a data refetch after. */
      engineTrigger: typeof engineTrigger

      /** Refetches all data from the backend. */
      refetchAll: typeof refetchAll

      /** Stops the passage of time. */
      stopTime: typeof stopTime

      /** Resumes the passage of time. */
      resumeTime: typeof startTime
    }
  }
}

function refetchAll(): Cypress.Chainable {
  return cy.window().invoke('refetchAll')
}

let timeIsStopped = false
function fastForward(duration: string): Cypress.Chainable {
  if (!timeIsStopped)
    throw new Error('Time is not stopped, cannot fast forward')
  return cy.task('db:fastforward', duration)
}

function sql(query: string): Cypress.Chainable {
  const dbURL =
    Cypress.env('DB_URL') || 'postgres://goalert@localhost:5432?sslmode=disable'

  return cy.exec(`go tool psql-lite -tx -d "$DB" -c "$QUERY"`, {
    env: {
      DB: dbURL,
      QUERY: query,
    },
  })
}

function engineTrigger(): Cypress.Chainable {
  if (!timeIsStopped) return cy.task('engine:trigger').refetchAll()

  // Since time is stopped, we resume it (which causes a trigger) and then stop it again. We want to avoid running the engine when we're doing anything with time, but sometimes we're trying to insert things at specific times in the DB and need the engine to record those actions at the "current" time.
  return cy
    .task('engine:stop')
    .task('engine:setapionly', false)
    .task('engine:start')
    .task('engine:trigger')
    .task('engine:stop')
    .task('engine:setapionly', true)
    .task('engine:start')
    .reload()
}

function stopTime(): Cypress.Chainable {
  return cy
    .task('engine:stop')
    .task('engine:setapionly', true)
    .task('db:setTimeSpeed', 0)
    .task('engine:start')
    .then(() => {
      timeIsStopped = true
    })
}

function startTime(): Cypress.Chainable {
  return cy
    .task('engine:stop')
    .task('engine:setapionly', false)
    .task('db:setTimeSpeed', 1)
    .task('engine:start')
    .then(() => {
      timeIsStopped = false
    })
    .task('engine:trigger')
}

Cypress.Commands.add('sql', sql)
Cypress.Commands.add('fastForward', fastForward)
Cypress.Commands.add('engineTrigger', engineTrigger)
Cypress.Commands.add('refetchAll', refetchAll)
Cypress.Commands.add('stopTime', stopTime)
Cypress.Commands.add('resumeTime', startTime)

export {}
