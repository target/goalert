import { Chance } from 'chance'
import { DateTime } from 'luxon'

const c = new Chance()

const messageTypes = [
  'alert_notification',
  'verification_message',
  'test_notification',
  'alert_status_update',
  'alert_notification_bundle',
  'alert_status_update_bundle',
  'schedule_on_call_notification',
]

const statuses = ['delivered', 'failed']

// createOutgoingMessage inserts mock outgoing message data to the db based on some provided options
// returns true/false depending on success of db insert
// params not inserted: contact method dest, provider info, and status details
function createOutgoingMessage(msg: OutgoingMessageOptions): Cypress.Chainable {
  const createdAt =
    msg?.createdAt ??
    DateTime.fromJSDate(
      c.date({
        year: new Date().getFullYear(),
        string: false,
      }) as Date,
    ).toISO()

  const sentAt = DateTime.fromISO(createdAt).plus(1000).toISO()

  // craft message for insert
  const m = {
    id: c.guid(),
    userID: msg.userID,
    serviceID: msg.serviceID,
    alertID: msg.alertID,
    createdAt,
    sentAt,
    messageType: msg?.messageType || c.pickone(messageTypes),
    status: msg?.status || c.pickone(statuses),
  }

  // create a contact method from user
  const cm = {
    id: c.guid(),
    userID: msg.userID,
    name: c.word(),
    type: c.pickone(['SMS', 'VOICE']),
    value: c.phone(),
    disabled: true,
  }

  console.log('userID: ', msg.userID)

  return cy
    .sql(
      `insert into user_contact_methods (id, user_id, name, type, value, disabled) values` +
        `('${cm.id}', '${cm.userID}', '${cm.name}', '${cm.type}', '${cm.value}', '${cm.disabled}');`,
    )
    .then(() => {
      const dbQuery =
        `insert into outgoing_messages (id, user_id, contact_method_id, service_id, alert_id, created_at, sent_at, message_type, last_status) values` +
        `('${m.id}', '${m.userID}', '${cm.id}', '${m.serviceID}', '${m.alertID}', '${m.createdAt}', '${m.sentAt}', '${m.messageType}', '${m.status}');`

      return cy.sql(dbQuery)
    })
}

Cypress.Commands.add('createOutgoingMessage', createOutgoingMessage)
