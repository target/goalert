import { Client } from './lib/graphql.js'
import Chance from 'https://chancejs.com/chance.min.js'

// Instantiate Chance so it can be used
var gen = new Chance()

function updateContactMethodName(c) {
  gen.pickone(c.user().contactMethods).name = gen.name()
}

function createNewRotation(c) {
  c.newRotation()
}

function updateRotation(c) {
  const r = c.randRotation()
  r.userIDs = gen.pickset(c.users(), 3).map((u) => u.id)
  let tz = gen.timezone()
  if (!tz.utc) {
    tz = 'Etc/UTC'
  } else {
    tz = tz.utc[0]
  }
  r.timeZone = tz
  r.activeUserIndex = 2
}

function deleteRotation(c) {
  c.randRotation().delete()
}
function createService(c) {
  c.newService()
}
function createEP(c) {
  c.newEP()
}
function deleteService(c) {
  c.randService().delete()
}

export default function LoginLogout() {
  const c = new Client('http://localhost:3030')

  // perform 15 random actions then logout
  for (let i = 0; i < 15; i++) {
    const action = gen.pickone([
      updateContactMethodName,
      createNewRotation,
      updateRotation,
      deleteRotation,
      createService,
      createEP,
      deleteService,
    ])
    action(c)
  }

  c.logout()
}
