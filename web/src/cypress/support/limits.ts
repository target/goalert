declare namespace Cypress {
  interface Chainable {
    getLimits: typeof getLimits
    updateLimits: typeof updateLimits
  }
}

interface Limits {
  id: ID
  description: string
  value: number
}
interface SystemLimitInput {
  id: ID
  value: number
}
enum ID {
  NotificationRulesPerUser = 'notification_rules_per_user',
  ContactMethodsPerUser = 'contact_methods_per_user',
  EPStepsPerPolicy = 'ep_steps_per_policy',
  EPActionsPerStep = 'ep_actions_per_step',
  ParticipantsPerRotation = 'participants_per_rotation',
  RulesPerSchedule = 'rules_per_schedule',
  IntegrationKeysPerService = 'integration_keys_per_service',
  UnackedAlertsPerService = 'unacked_alerts_per_service',
  TargetsPerSchedule = 'targets_per_schedule',
  HeartbeatMonitorsPerService = 'heartbeat_monitors_per_service',
  UserOverridesPerSchedule = 'user_overrides_per_schedule',
}

function getLimits(): Cypress.Chainable<Limits> {
  const query = `query getLimits() {
    systemLimits {
      id
      description
      value
    }
  }`

  return cy.graphql2(query).then(res => res.systemLimits)
}

function updateLimits(input: SystemLimitInput[]): Cypress.Chainable<Boolean> {
  const query = `mutation updateLimits($input: [SystemLimitInput!]!){
    setSystemLimits(input: $input)
  }`

  return cy.graphql2(query, { input: input })
}

Cypress.Commands.add('getLimits', getLimits)
Cypress.Commands.add('updateLimits', updateLimits)
