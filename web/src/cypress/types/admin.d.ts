declare namespace Cypress {
  interface Chainable<Subject> {
    // Creates one outgoing log based on the provided options.
    createOutgoingMessage: createOutgoingMessage
  }
}

interface OutgoingMessageOptions {
  serviceID: string
  alertID: string
  userID: string

  id?: string
  messageType?: string
  createdAt?: string
  status?: string
}

type createOutgoingMessageFn = (
  log: OutgoingMessageOptions,
) => Cypress.Chainable
