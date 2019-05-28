import { Chance } from 'chance'

const c = new Chance()

declare global {
  namespace Cypress {
    interface Chainable {
      /** Creates a new user profile. */
      createUser: typeof createUser

      /** Creates multiple new user profiles. */
      createManyUsers: typeof createManyUsers

      /**
       * Resets the test user profile, including any existing contact methods.
       */
      resetProfile: typeof resetProfile

      /** Adds a contact method. If userID is missing, the test user's will be used. */
      addContactMethod: typeof addContactMethod

      /** Adds a notification rule. If userID is missing, the test user's will be used. */
      addNotificationRule: typeof addNotificationRule
    }
  }

  type UserRole = 'user' | 'admin'
  interface Profile {
    id: string
    name: string
    email: string
    role: UserRole
  }
  interface UserOptions {
    name?: string
    email?: string
    role?: UserRole
  }

  type ContactMethodType = 'SMS' | 'VOICE'
  interface ContactMethod {
    id: string
    userID: string
    name: string
    type: ContactMethodType
    value: string
  }
  interface ContactMethodOptions {
    userID?: string
    name?: string
    type?: ContactMethodType
    value?: string
  }
  interface NotificationRule {
    id: string
    userID: string
    cmID: string
    cm: ContactMethod
    delay: number
  }
  interface NotificationRuleOptions {
    userID?: string
    delay?: number
    cmID?: string
    cm?: ContactMethodOptions
  }
}

function createManyUsers(
  users: Array<UserOptions>,
): Cypress.Chainable<Array<Profile>> {
  const profiles: Array<Profile> = users.map(user => ({
    id: c.guid(),
    name: user.name || c.word({ length: 12 }),
    email: user.email || c.email(),
    role: user.role || 'user',
  }))

  const dbURL =
    Cypress.env('DB_URL') || 'postgres://goalert@localhost:5432?sslmode=disable'

  const dbQuery =
    `insert into users (id, name, email, role) values` +
    profiles
      .map(p => `('${p.id}', '${p.name}', '${p.email}', '${p.role}')`)
      .join(',') +
    `;`

  return cy.exec(`psql -d '${dbURL}' -c "${dbQuery}"`).then(() => profiles)
}

function createUser(user?: UserOptions): Cypress.Chainable<Profile> {
  if (!user) user = {}
  return createManyUsers([user]).then(p => p[0])
}

function addContactMethod(
  cm?: ContactMethodOptions,
): Cypress.Chainable<ContactMethod> {
  if (!cm) cm = {}
  if (!cm.userID) {
    return cy
      .fixture('profile')
      .then(prof => addContactMethod({ ...cm, userID: prof.id }))
  }

  const query = `mutation addCM($input: CreateContactMethodInput!){
        createContactMethod(input: $input) {
            id
            name
            type
            value
        }
    }`

  const newPhone = '+1763' + c.integer({ min: 3000000, max: 3999999 })
  return cy
    .graphql(query, {
      input: {
        user_id: cm.userID,
        name: cm.name || 'SM CM ' + c.word({ length: 8 }),
        type: cm.type || c.pickone(['SMS', 'VOICE']),
        value: cm.value || newPhone,
      },
    })
    .then(newCM => {
      newCM = newCM.createContactMethod
      newCM.userID = cm && cm.userID
      return newCM
    })
}
function addNotificationRule(
  nr?: NotificationRuleOptions,
): Cypress.Chainable<NotificationRule> {
  if (!nr) nr = {}
  if (!nr.userID) {
    return cy
      .fixture('profile')
      .then(prof => addNotificationRule({ ...nr, userID: prof.id }))
  }
  if (!nr.cmID) {
    return cy
      .addContactMethod({ ...nr.cm, userID: nr.userID })
      .then(cm => addNotificationRule({ ...nr, cmID: cm.id }))
  }

  const query = `mutation addNR($input: CreateNotificationRuleInput!){
        createNotificationRule(input: $input) {
            id
            delay: delay_minutes
            cmID: contact_method_id
            cm: contact_method {
                id
                name
                type
                value
            }
        }
    }`

  return cy
    .graphql(query, {
      input: {
        user_id: nr.userID,
        delay_minutes: nr.delay || c.integer({ min: 0, max: 15 }),
        contact_method_id: nr.cmID,
      },
    })
    .then(newNR => {
      newNR = newNR.createNotificationRule
      const userID = nr && nr.userID
      newNR.userID = userID
      newNR.cm.userID = userID
      return newNR
    })
}
function clearContactMethods(id: string): Cypress.Chainable {
  const list = `{
        user(id: "${id}") {
            contact_methods { id }
        }
    }`
  return cy.graphql(list).then(res => {
    if (!res.user.contact_methods.length) return
    res.user.contact_methods.forEach((cm: any) => {
      cy.graphql(`
                mutation{
                    deleteContactMethod(input:{id:"${cm.id}"}) {deleted_id}
                }
            `)
    })
  })
}
function resetProfile(prof?: Profile): Cypress.Chainable<Profile> {
  if (!prof) {
    return cy.fixture('profile').then(resetProfile)
  }

  const query = `mutation updateUser($input: UpdateUserInput!){
          updateUser(input: $input) {
              id
              name
              email
              role
          }
      }`

  return clearContactMethods(prof.id)
    .graphql(query, {
      input: {
        id: prof.id,
        name: prof.name,
        email: prof.email,
        role: prof.role,
      },
    })
    .then(res => res.updateUser)
}

Cypress.Commands.add('createUser', createUser)
Cypress.Commands.add('createManyUsers', createManyUsers)
Cypress.Commands.add('resetProfile', resetProfile)
Cypress.Commands.add('addContactMethod', addContactMethod)
Cypress.Commands.add('addNotificationRule', addNotificationRule)
