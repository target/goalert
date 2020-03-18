import { Chance } from 'chance'
import { DateTime } from 'luxon'

const c = new Chance()

declare global {
  namespace Cypress {
    interface Chainable {
      createAlert: typeof createAlert
      createManyAlerts: typeof createManyAlerts
      closeAlert: typeof closeAlert
      createAlertLogs: typeof createAlertLogs
    }
  }

  interface Alert {
    number: number
    id: number
    summary: string
    details: string
    serviceID: string
    service: Service
  }

  interface AlertOptions {
    summary?: string
    details?: string
    serviceID?: string

    service?: ServiceOptions
  }

  interface AlertLogOptions {
    count?: number
    alertID?: number
    alert?: AlertOptions
  }

  interface AlertLogs {
    alert: Alert
    logs: Array<AlertLog>
  }
  interface AlertLog {
    timestamp: string
    message: string
  }
}

function createAlertLogs(opts?: AlertLogOptions): Cypress.Chainable<AlertLogs> {
  if (!opts) return createAlertLogs({})
  if (!opts.count) opts.count = c.integer({ min: 1, max: 50 })
  if (!opts.alertID) {
    return cy
      .createAlert(opts.alert)
      .then(alert => createAlertLogs({ ...opts, alertID: alert.number }))
  }

  const genMeta = () =>
    JSON.stringify({
      NewStepIndex: c.integer({ min: 0, max: 5 }),
      Repeat: false,
      Forced: false,
      Deleted: false,
      OldDelayMinutes: c.integer({ min: 1, max: 60 }),
    })

  let query = `INSERT INTO alert_logs (alert_id, timestamp, event, meta, message) values\n`
  const n = DateTime.utc()
  const vals = []
  for (let i = 0; i < opts.count; i++) {
    vals.push(
      `(${opts.alertID},'${n
        .plus({
          milliseconds: i,
        })
        .toISO()}', 'escalated', '${genMeta()}', '')`,
    )
  }
  query += vals.join(',') + ';'

  return cy
    .sql(query)
    .then(() => getAlert(opts.alertID as number))
    .then(alert => {
      return getAlertLogs(opts.alertID as number).then(logs => {
        return {
          alert,
          logs,
        }
      })
    })
}

function getAlertLogs(id: number): Cypress.Chainable<Array<AlertLog>> {
  const query = `query GetLogs($id: Int!, $after: String!) {
    alert(id: $id) {
      recentEvents(input:{limit:149, after: $after}) {
        nodes {
          timestamp
          message
        }
        pageInfo {
          hasNextPage
          endCursor
        }
      }
    }
  }`

  const next = (logs: Array<AlertLog>, after: string): any =>
    cy.graphql2(query, { id, after }).then(res => {
      if (res.alert.recentEvents.pageInfo.hasNextPage) {
        return next(
          logs.concat(res.alert.recentEvents.nodes as Array<AlertLog>),
          res.alert.recentEvents.pageInfo.endCursor,
        )
      }

      return logs.concat(res.alert.recentEvents.nodes)
    })

  return next([], '')
}

function getAlert(id: number): Cypress.Chainable<Alert> {
  const query = `query GetAlert($id: Int!) {
    alert(id: $id) {
      number: _id, id, summary, details, serviceID: service_id,
      service {
        id, name, description
        epID: escalation_policy_id,
        ep: escalation_policy {
            id
            name
            description
            repeat
        }
      }
    }
  }`

  return cy.graphql(query, { id }).then(res => res.alert)
}

function createAlert(a?: AlertOptions): Cypress.Chainable<Alert> {
  if (!a) a = {}
  const query = `mutation CreateAlert($input: CreateAlertInput){
      createAlert(input: $input) {
        number: _id, id, summary, details, serviceID: service_id,
        service {
          id, name, description
          epID: escalation_policy_id,
          ep: escalation_policy {
              id
              name
              description
              repeat
          }
        }
      }
    }`

  if (!a.serviceID) {
    return cy
      .createService(a.service)
      .then(svc => createAlert({ ...a, serviceID: svc.id }))
  }

  return cy
    .graphql(query, {
      input: {
        service_id: a.serviceID,
        summary: a.summary || c.sentence({ words: 3 }),
        details: a.details || c.sentence({ words: 5 }),
      },
    })
    .then(res => res.createAlert)
}

// global scope if createManyAlerts is called more than once in a given test suite
let dedupIdx = 0
function createManyAlerts(
  count: number,
  alertOptions?: AlertOptions,
): Cypress.Chainable {
  if (!alertOptions?.serviceID) {
    return cy
      .createService(alertOptions?.service)
      .then(res =>
        createManyAlerts(count, { ...alertOptions, serviceID: res.id }),
      )
  }

  // build query
  let query =
    'insert into alerts (service_id, summary, details, dedup_key) values '
  const rows: Array<string> = []
  for (let i = 0; i < count; i++) {
    const summary = alertOptions.summary || c.word()
    const details = alertOptions.details || c.sentence()
    const dedupKey = 'manual:1:createManyAlerts_' + dedupIdx

    rows.push(
      `('${alertOptions?.serviceID}', '${summary}', '${details}', '${dedupKey}')`,
    )

    dedupIdx++
  }
  query = query + rows.join(',') + ';'

  return cy.sql(query)
}

function closeAlert(id: number): Cypress.Chainable<Alert> {
  const query = `
    mutation {
      updateAlertStatus(input: $input) { id }
    }
  `

  return cy.graphql(query, { input: { id } })
}

Cypress.Commands.add('createAlert', createAlert)
Cypress.Commands.add('createManyAlerts', createManyAlerts)
Cypress.Commands.add('createAlertLogs', createAlertLogs)
Cypress.Commands.add('closeAlert', closeAlert)
