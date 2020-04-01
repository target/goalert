import { Chance } from 'chance'

const c = new Chance()

function createManyUsers(
  users: Array<UserOptions>,
): Cypress.Chainable<Array<Profile>> {
  const profiles: Array<Profile> = users.map(user => ({
    id: c.guid(),
    name: user.name || c.word({ length: 12 }),
    email: user.email || c.email(),
    role: user.role || 'user',
  }))

  const dbQuery =
    `insert into users (id, name, email, role) values` +
    profiles
      .map(p => `('${p.id}', '${p.name}', '${p.email}', '${p.role}')`)
      .join(',') +
    `;`

  return cy.sql(dbQuery).then(() => profiles)
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

  const mutation = `
    mutation ($input: CreateUserContactMethodInput!) {
      createUserContactMethod(input: $input) {
        id
        name
        type
        value
      }
    }
  `

  const newPhone = '+1763' + c.integer({ min: 3000000, max: 3999999 })
  return cy
    .graphql2(mutation, {
      input: {
        userID: cm.userID,
        name: cm.name || 'SM CM ' + c.word({ length: 8 }),
        type: cm.type || c.pickone(['SMS', 'VOICE']),
        value: cm.value || newPhone,
      },
    })
    .then((res: GraphQLResponse) => {
      res = res.createUserContactMethod
      res.userID = cm && cm.userID
      return res
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

  if (!nr.contactMethodID) {
    return cy
      .addContactMethod({ ...nr.contactMethod, userID: nr.userID })
      .then((cm: ContactMethod) =>
        addNotificationRule({ ...nr, contactMethodID: cm.id }),
      )
  }

  const mutation = `
    mutation ($input: CreateUserNotificationRuleInput!) {
      createUserNotificationRule(input: $input) {
        id
        delayMinutes
        contactMethodID
        contactMethod {
          id
          name
          type
          value
        }
      }
    }
  `

  return cy
    .graphql2(mutation, {
      input: {
        userID: nr.userID,
        contactMethodID: nr.contactMethodID,
        delayMinutes: nr.delayMinutes || c.integer({ min: 0, max: 15 }),
      },
    })
    .then((res: GraphQLResponse) => {
      res = res.createUserNotificationRule

      const userID = nr && nr.userID
      res.userID = userID
      res.contactMethod.userID = userID

      return res
    })
}

function clearContactMethods(id: string): Cypress.Chainable {
  const query = `
    query($id: ID!) {
      user(id: $id) {
        contactMethods {
          id
        }
      }
    }
  `

  const mutation = `
    mutation($input: [TargetInput!]!) {
      deleteAll(input: $input)
    }
  `

  return cy.graphql2(query, { id }).then((res: GraphQLResponse) => {
    if (!res.user.contactMethods.length) return

    res.user.contactMethods.forEach((cm: ContactMethod) => {
      cy.graphql2(mutation, {
        input: [
          {
            type: 'contactMethod',
            id: cm.id,
          },
        ],
      })
    })
  })
}

function resetProfile(prof?: Profile): Cypress.Chainable {
  if (!prof) {
    return cy.fixture('profile').then(resetProfile)
  }

  const mutation = `
    mutation updateUser($input: UpdateUserInput!) {
      updateUser(input: $input)
    }
  `

  return clearContactMethods(prof.id).graphql2(mutation, {
    input: {
      id: prof.id,
      name: prof.name,
      email: prof.email,
      role: prof.role,
    },
  })
}

Cypress.Commands.add('createUser', createUser)
Cypress.Commands.add('createManyUsers', createManyUsers)
Cypress.Commands.add('resetProfile', resetProfile)
Cypress.Commands.add('addContactMethod', addContactMethod)
Cypress.Commands.add('addNotificationRule', addNotificationRule)
