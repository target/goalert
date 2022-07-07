import { Client } from './lib/graphql'
import Chance from 'chance'
import { genTZ } from './lib/util'
import {
  ScheduleTarget,
  ScheduleTargetInput,
  TargetInput,
} from '../../web/src/schema'

// Instantiate Chance so it can be used
var gen = new Chance()

function updateContactMethodName(c: Client) {
  gen.pickone(c.user().contactMethods).name = gen.name()
}

function createNewRotation(c: Client) {
  c.newRotation()
}

function updateRotation(c: Client) {
  const r = c.randRotation()
  const n = gen.integer({ min: 3, max: 20 })
  r.userIDs = gen.pickset(c.users(), n).map((u) => u.id)
  r.timeZone = genTZ()
  r.activeUserIndex = gen.integer({ min: 0, max: n - 1 })
}

function deleteRotation(c: Client) {
  c.randRotation().delete()
}
function createService(c: Client) {
  c.newService()
}
function createEP(c: Client) {
  c.newEP()
}
function deleteService(c: Client) {
  c.randService().delete()
}
function createUser(c: Client) {
  c.newUser()
}
function deleteUser(c: Client) {
  c.randUser().delete()
}
function createSchedule(c: Client) {
  c.newSchedule()
}
function deleteSchedule(c: Client) {
  c.randSchedule().delete()
}
function updateSchedule(c: Client) {
  const s = c.randSchedule()
  s.name = gen.name()
  s.description = gen.sentence()
  s.timeZone = genTZ()
}
function createUserOverride(c: Client) {
  c.newUserOverride()
}
function deleteUserOverride(c: Client) {
  c.randUserOverride().delete()
}
function updateUserOverride(c: Client) {
  const u = c.randUserOverride()
  if (u.addUserID) {
    u.addUserID = c.randUser().id
  }
  if (u.removeUserID) {
    u.removeUserID = c.randUser().id
  }
}

function addScheduleTarget(c: Client) {
  const s = c.randSchedule()

  c.randSchedule().setTarget(
    gen.bool()
      ? { type: 'user', id: c.randUser().id }
      : { type: 'rotation', id: c.randRotation().id },
  )
}

function deleteScheduleTarget(c: Client) {
  const s = c.randSchedule()

  if (s.targets.length === 0) return

  s.clearTarget(gen.pickone(s.targets))
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
      createUser,
      deleteUser,
      createSchedule,
      deleteSchedule,
      updateSchedule,
      addScheduleTarget,
      deleteScheduleTarget,
      createUserOverride,
      updateUserOverride,
      deleteUserOverride,
    ])
    action(c)
  }

  c.logout()
}
