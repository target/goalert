import { Chance } from 'chance'
import { DateTime } from 'luxon'
import toTitleCase from '../../app/util/toTitleCase'
const c = new Chance()

declare global {
  namespace Cypress {
    interface Chainable {
      // Creates one outgoing log based on the provided options.
      createOutgoingMessage: typeof createOutgoingMessage
    }
  }

  interface OutgoingMessageOptions {
    id?: string
    serviceID?: string
    serviceName?: string
    epID?: string
    alertID?: number
    alertLogID?: number
    userID?: string
    userName?: string
    contactMethodID?: string
    messageType?: string
    createdAt?: string
    sentAt?: string
    status?: string
  }
}

const messageTypes = [
  'alert_notification',
  'alert_status_update',
  'test_notification',
  'alert_notification_bundle',
  'schedule_on_call_notification',
]

function msgTypeToDebugMsg(type: string): string {
  switch (type) {
    case 'alert_notification':
      return 'Alert'
    case 'alert_status_update':
      return 'AlertStatus'
    case 'test_notification':
      return 'Test'
    case 'alert_notification_bundle':
      return 'AlertBundle'
    case 'schedule_on_call_notification':
      return 'ScheduleOnCallUsers'
    default:
      console.warn('could not process unknown type for DebugMessage type')
      return ''
  }
}

const statuses = ['delivered', 'failed']

// createOutgoingMessage inserts mock outgoing message data to the db based on some provided options
// does not support setting the notification channel, updatedAt, source, providerID,
// or the 'verification_message' and 'alert_status_update_bundle' message types.
function createOutgoingMessage(
  msg: OutgoingMessageOptions = {},
): Cypress.Chainable {
  // create all unused optional params before attempting db insert

  // user and contact method
  if (!msg.userID) {
    return cy
      .createUser({ name: msg.userName })
      .then((u: Profile) =>
        createOutgoingMessage({ ...msg, userID: u.id, userName: u.name }),
      )
  }
  // needed for db constraint + destination
  if (!msg.contactMethodID) {
    return cy
      .addContactMethod({ userID: msg.userID })
      .then((cm: ContactMethod) =>
        createOutgoingMessage({ ...msg, contactMethodID: cm.id }),
      )
  }

  // escalation policy with user on a step
  if (!msg.epID) {
    return cy.createEP().then((ep: EP) =>
      cy
        .createEPStep({
          targets: [
            {
              id: msg.userID as string, // guaranteed above
              type: 'user',
            },
          ],
        })
        .then(() => {
          return createOutgoingMessage({
            ...msg,
            epID: ep.id,
          })
        }),
    )
  }

  // service
  if (!msg.serviceID) {
    return cy.createService({ name: msg.serviceName }).then((s: Service) =>
      createOutgoingMessage({
        ...msg,
        serviceID: s.id,
        serviceName: s.name,
      }),
    )
  }

  // alert and alert log
  if (!msg.alertID) {
    return cy.createAlert({ serviceID: msg.serviceID }).then((a: Alert) =>
      createOutgoingMessage({
        ...msg,
        alertID: a.id,
      }),
    )
  }
  // needed for db constraint
  if (!msg.alertLogID) {
    return cy
      .createAlertLogs({ alertID: msg.alertID })
      .then((alertLogObj: AlertLogs) =>
        createOutgoingMessage({
          ...msg,
          alertLogID: alertLogObj.logs[0].id,
        }),
      )
  }

  const someDate = DateTime.fromJSDate(
    c.date({
      year: new Date().getFullYear(),
      string: false,
    }) as Date,
  )
  const createdAt = msg?.createdAt ?? someDate.toISO()
  const sentAt = msg?.sentAt ?? DateTime.fromISO(createdAt).plus(1000).toISO()

  // craft helper message obj for insert
  const m = {
    id: msg?.id ?? c.guid(),
    serviceID: msg.serviceID,
    epID: msg.epID,
    alertID: msg.alertID,
    alertLogID: msg.alertLogID,
    userID: msg.userID,
    contactMethodID: msg.contactMethodID,
    createdAt,
    sentAt,
    messageType: msg?.messageType ?? c.pickone(messageTypes),
    status: msg?.status ?? c.pickone(statuses),
  }

  return cy
    .sql(
      `insert into outgoing_messages ` +
        `(id, user_id, contact_method_id, service_id, escalation_policy_id, alert_id, ` +
        `alert_log_id, created_at, sent_at, message_type, last_status) values` +
        `('${m.id}', '${m.userID}', '${m.contactMethodID}', '${m.serviceID}', '${m.epID}', '` +
        `${m.alertID}', '${m.alertLogID}', '${m.createdAt}', '${m.sentAt}', '${m.messageType}', '${m.status}');`,
    )
    .then(() => ({
      id: m.id,
      createdAt: m.createdAt,
      updatedAt: '',
      type: msgTypeToDebugMsg(m.messageType),
      status: toTitleCase(m.status),
      userID: m.userID,
      userName: msg.userName,
      destination: m.contactMethodID,
      serviceID: m.serviceID,
      serviceName: msg.serviceName,
      alertID: m.alertID,
    }))
}

Cypress.Commands.add('createOutgoingMessage', createOutgoingMessage)
