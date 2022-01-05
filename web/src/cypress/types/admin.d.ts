declare namespace Cypress {
  interface Chainable<Subject> {
    // Creates one outgoing log based on the provided options.
    createOutgoingMessage: createOutgoingMessage
  }
}

interface OutgoingMessageOptions {
  id?: string
  serviceID?: string
  serviceName?: string
  epID?: string
  alertID?: string
  alertLogID?: string
  userID?: string
  userName?: string
  contactMethodID?: string
  messageType?: string
  createdAt?: string
  sentAt?: string
  status?: string
}

type createOutgoingMessageFn = (
  message: OutgoingMessageOptions,
) => Cypress.Chainable
