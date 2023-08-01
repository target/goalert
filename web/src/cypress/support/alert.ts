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
      ackAlert: typeof ackAlert
      escalateAlert: typeof escalateAlert
      setAlertNoise: typeof setAlertNoise
    }
  }

  interface Alert {
    id: number
    alertID: number
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
    id: number
    timestamp: string
    message: string
  }
}

function getAlertLogs(id: number): Cypress.Chainable<Array<AlertLog>> {
  const query = `
    query GetLogs($id: Int!, $after: String!) {
      alert(id: $id) {
        recentEvents(input: { limit: 149, after: $after }) {
          nodes {
            id
            timestamp
            message
          }
          pageInfo {
            hasNextPage
            endCursor
          }
        }
      }
    }
  `

  // NOTE next recursively builds logs to ultimately yield an AlertLog[]
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  const next = (logs: AlertLog[], after: string): Cypress.Chainable<any> =>
    cy.graphql(query, { id, after }).then((res: GraphQLResponse) => {
      const hasNextPage: boolean = res.alert.recentEvents.pageInfo.hasNextPage
      const endCursor: string = res.alert.recentEvents.pageInfo.endCursor
      const nodes: AlertLog[] = res.alert.recentEvents.nodes

      if (hasNextPage) {
        return next(logs.concat(nodes), endCursor)
      }

      return logs.concat(nodes)
    })

  return next([], '')
}

function getAlert(id: number): Cypress.Chainable<Alert> {
  const query = `
    query GetAlert($id: Int!) {
      alert(id: $id) {
        id
        summary
        details
        serviceID
        service {
          id
          name
          description
          epID: escalationPolicyID,
          ep: escalationPolicy {
              id
              name
              description
              repeat
          }
        }
      }
    }
  `

  return cy.graphql(query, { id }).then((res: GraphQLResponse) => res.alert)
}

function createAlertLogs(opts?: AlertLogOptions): Cypress.Chainable<AlertLogs> {
  if (!opts) return createAlertLogs({})
  if (!opts.count) opts.count = c.integer({ min: 1, max: 50 })
  if (!opts.alertID) {
    return cy
      .createAlert(opts.alert)
      .then((alert: Alert) => createAlertLogs({ ...opts, alertID: alert.id }))
  }

  const genMeta = (): string =>
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
    .then((alert: Alert) => {
      return getAlertLogs(opts.alertID as number).then((logs) => {
        return {
          alert,
          logs,
        }
      })
    })
}

function createAlert(a?: AlertOptions): Cypress.Chainable<Alert> {
  if (!a) a = {}
  const query = `
    mutation CreateAlert($input: CreateAlertInput!){
      createAlert(input: $input) {
        id
        summary
        details
        serviceID
        service {
          id
          name
          description
          epID: escalationPolicyID
          ep: escalationPolicy {
            id
            name
            description
            repeat
          }
        }
      }
    }
  `

  if (!a.serviceID) {
    return cy
      .createService(a.service)
      .then((svc: Service) => createAlert({ ...a, serviceID: svc.id }))
  }

  return cy
    .graphql(query, {
      input: {
        serviceID: a.serviceID,
        summary: a.summary || c.sentence({ words: 3 }),
        details: a.details || c.sentence({ words: 5 }),
      },
    })
    .then((res: GraphQLResponse) => res.createAlert)
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
      .then((svc: Service) =>
        createManyAlerts(count, { ...alertOptions, serviceID: svc.id }),
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

function ackAlert(id: number): Cypress.Chainable<void> {
  const query = `
    mutation ($id: Int!) {
      updateAlerts(input: {
        newStatus: StatusAcknowledged
        alertIDs: [$id]
      }) { id }
    }
  `

  return cy.graphqlVoid(query, { id })
}

function closeAlert(id: number): Cypress.Chainable<void> {
  const query = `
    mutation ($id: Int!) {
      updateAlerts(input: {
        newStatus: StatusClosed
        alertIDs: [$id]
      }) { id }
    }
  `

  return cy.graphqlVoid(query, { id })
}

function escalateAlert(id: number): Cypress.Chainable<Alert> {
  const query = `
    mutation ($id: Int!) {
      escalateAlerts(input: [$id]) { id }
    }
  `

  return cy
    .graphql(query, { id })
    .then((res: GraphQLResponse) => res.escalateAlerts)
}

function setAlertNoise(
  id: number,
  noiseReason: string,
): Cypress.Chainable<boolean> {
  const query = `
    mutation($input: SetAlertNoiseReasonInput!) {
      setAlertNoiseReason(input: $input)
    }
  `

  return cy
    .graphql(query, {
      input: {
        alertID: id,
        noiseReason,
      },
    })
    .then((res: GraphQLResponse) => Boolean(res))
}

Cypress.Commands.add('createAlert', createAlert)
Cypress.Commands.add('createManyAlerts', createManyAlerts)
Cypress.Commands.add('createAlertLogs', createAlertLogs)
Cypress.Commands.add('closeAlert', closeAlert)
Cypress.Commands.add('ackAlert', ackAlert)
Cypress.Commands.add('escalateAlert', escalateAlert)
Cypress.Commands.add('setAlertNoise', setAlertNoise)
