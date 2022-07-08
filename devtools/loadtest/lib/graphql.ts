import http from 'k6/http'
import Chance from 'chance'
import { genTZ } from './util'
import {
  CreateUserOverrideInput,
  ScheduleTarget,
  TargetInput,
} from '../../../web/src/schema'

// Instantiate Chance so it can be used
var gen = new Chance()

class IDFetchType {
  constructor(c: Client, id: string, queryName: string) {
    this.c = c
    this.id = id
    this.queryName = queryName
  }

  public c: Client
  public id: string
  private queryName: string

  simpleField(fieldName: string, suffix?: string) {
    return this.c.query(
      `query($id: ID!){${this.queryName}(id: $id){${fieldName}${
        suffix || ''
      }}}`,
      { id: this.id },
    ).data[this.queryName][fieldName]
  }
  simpleFieldMap<T>(fieldName: string, Type: T): Array<T> {
    return this.simpleField(fieldName, '{id}').map(
      (obj) => new Type(this.c, obj.id),
    )
  }
  simpleUpdateField(fieldName: string, typeName: string, value: any) {
    const name = this.queryName[0].toUpperCase() + this.queryName.slice(1)
    return this.c.query(
      `mutation($id: ID!, $value: ${typeName}){
        update${name}(input:{id: $id, ${fieldName}: $value})
      }`,
      { id: this.id, value },
    )
  }

  delete() {
    return this.c.query(
      `mutation($id: ID!){
        deleteAll(input:[{id: $id, type: ${this.queryName.replace(
          'userContactMethod',
          'contactMethod',
        )}}])
      }`,
      { id: this.id },
    )
  }
}

class UserContactMethod extends IDFetchType {
  constructor(c: Client, id: string) {
    super(c, id, 'userContactMethod')
  }

  get name(): string {
    return this.simpleField('name')
  }
  set name(newName: string) {
    this.simpleUpdateField('name', 'String!', newName)
  }

  get type() {
    return this.simpleField('type')
  }
  get value() {
    return this.simpleField('value')
  }
  get formattedValue() {
    return this.simpleField('formattedValue')
  }
  get disabled() {
    return this.simpleField('disabled')
  }
  get lastTestVerifyAt() {
    return this.simpleField('lastTestVerifyAt')
  }
}

class Schedule extends IDFetchType {
  constructor(c: Client, id: string) {
    super(c, id, 'schedule')
  }

  get name() {
    return this.simpleField('name')
  }
  set name(newName) {
    this.simpleUpdateField('name', 'String!', newName)
  }

  get description() {
    return this.simpleField('description')
  }
  set description(newDescription) {
    this.simpleUpdateField('description', 'String', newDescription)
  }

  get timeZone() {
    return this.simpleField('timeZone')
  }
  set timeZone(newTimeZone) {
    this.simpleUpdateField('timeZone', 'String!', newTimeZone)
  }

  get targets(): Array<TargetInput> {
    return this.simpleField('targets', '{target{id, type}}').map(
      (t: ScheduleTarget) => t.target,
    )
  }

  setTarget(target: TargetInput) {
    this.c.query(
      `mutation ($id: ID!, $type: TargetType!, $scheduleID: ID!) {
        updateScheduleTarget(
          input: {
            target: {id: $id, type: $type}, 
            scheduleID: $scheduleID, 
            rules: [{weekdayFilter:[true,true,true,true,true,true,true]}]
          }
        )
      }`,
      {
        id: target.id,
        type: target.type,
        scheduleID: this.id,
      },
    )
  }

  clearTarget(target: TargetInput) {
    this.c.query(
      `mutation ($id: ID!, $type: TargetType!, $scheduleID: ID!) {
        updateScheduleTarget(
          input: {
            target: {id: $id, type: $type}, 
            scheduleID: $scheduleID, 
            rules: []
          }
        )
      }`,
      {
        id: target.id,
        type: target.type,
        scheduleID: this.id,
      },
    )
  }
}

class Service extends IDFetchType {
  constructor(c: Client, id: string) {
    super(c, id, 'service')
  }
  get name() {
    return this.simpleField('name')
  }
  set name(newName) {
    this.simpleUpdateField('name', 'String', newName)
  }
}

class EP extends IDFetchType {
  constructor(c: Client, id: string) {
    super(c, id, 'escalationPolicy')
  }

  get name() {
    return this.simpleField('name')
  }
  set name(newName) {
    this.simpleUpdateField('name', 'String', newName)
  }

  get description() {
    return this.simpleField('description')
  }
  set description(value) {
    this.simpleUpdateField('description', 'String', value)
  }
}

class Rotation extends IDFetchType {
  constructor(c: Client, id: string) {
    super(c, id, 'rotation')
  }
  get name() {
    return this.simpleField('name')
  }
  get activeUserIndex() {
    return this.simpleField('activeUserIndex')
  }
  set activeUserIndex(idx) {
    this.simpleUpdateField('activeUserIndex', 'Int!', idx)
  }

  get users() {
    return this.userIDs.map((id) => new User(this.c, id))
  }
  get userIDs(): Array<string> {
    return this.simpleField('userIDs')
  }
  set userIDs(ids) {
    this.simpleUpdateField('userIDs', '[ID!]', ids)
  }

  get timeZone() {
    return this.simpleField('timeZone')
  }
  set timeZone(name) {
    this.simpleUpdateField('timeZone', 'String', name)
  }
}

class User extends IDFetchType {
  constructor(c: Client, id: string) {
    super(c, id, 'user')
  }

  get name() {
    return this.simpleField('name')
  }
  get email() {
    return this.simpleField('email')
  }
  get role() {
    return this.simpleField('role')
  }
  get statusUpdateContactMethodID() {
    return this.simpleField('statusUpdateContactMethodID')
  }
  get isFavorite() {
    return this.simpleField('isFavorite')
  }
  get contactMethods(): Array<UserContactMethod> {
    return this.simpleFieldMap('contactMethods', UserContactMethod)
  }

  newContactMethod() {
    const q = this.c.query(
      `mutation($input: CreateUserContactMethodInput!){createUserContactMethod(input:$input){id}}`,
      {
        input: {
          userID: this.id,
          name: 'K6 ' + gen.string({ alpha: true, length: 20 }),
          type: 'SMS',
          value: '+1763555' + gen.string({ numeric: true, length: 4 }),
        },
      },
    )
    const id = q.data.createUserContactMethod.id
    return new UserContactMethod(this.c, id)
  }
}

interface IDNode {
  id: string
}

class UserOverride extends IDFetchType {
  constructor(c: Client, id: string) {
    super(c, id, 'userOverride')
  }

  get addUserID() {
    return this.simpleField('addUserID')
  }
  set addUserID(id: string) {
    this.simpleUpdateField('addUserID', 'ID!', id)
  }

  get removeUserID() {
    return this.simpleField('removeUserID')
  }
  set removeUserID(id: string) {
    this.simpleUpdateField('removeUserID', 'ID!', id)
  }

  get start() {
    return this.simpleField('start')
  }
  set start(date: string) {
    this.simpleUpdateField('start', 'ISOTimestamp!', date)
  }

  get end() {
    return this.simpleField('end')
  }

  set end(date: string) {
    this.simpleUpdateField('end', 'ISOTimestamp!', date)
  }
}

export class Client {
  constructor(baseURL: string) {
    this.baseURL = baseURL
    this.login()
  }

  private baseURL: string

  login(user = 'admin', pass = 'admin123') {
    let resp = http.get(this.baseURL + '/api/v2/identity/providers')
    let providers = JSON.parse(resp.body)
    let loginURL = providers.find((p: { ID: string }) => p.ID === 'basic').URL

    http.post(
      this.baseURL + loginURL,
      {
        username: user,
        password: pass,
      },
      {
        headers: {
          referer: this.baseURL,
        },
      },
    )
  }

  logout() {
    http.get(this.baseURL + '/api/v2/identity/logout')
  }

  userOverride(id: string): UserOverride {
    return new UserOverride(this, id)
  }
  userOverrides(): Array<UserOverride> {
    return this.query(
      `query{userOverrides{nodes{id}}}`,
    ).data.userOverrides.nodes.map((obj: IDNode) => this.userOverride(obj.id))
  }
  randUserOverride(): UserOverride {
    return gen.pickone(this.userOverrides())
  }
  newUserOverride(scheduleID?: string): UserOverride {
    if (!scheduleID) {
      return this.newUserOverride(this.randSchedule().id)
    }

    const startUnix = Date.now() + gen.integer({ min: 30000, max: 1000000 })
    const endUnix = startUnix + gen.integer({ min: 60000, max: 1000000 })

    let addUser, removeUser
    switch (gen.integer({ min: 0, max: 2 })) {
      case 0: // both add and remove
        addUser = this.randUser().id
        removeUser = this.randUser().id
        break
      case 1: // add
        addUser = this.randUser().id
        removeUser = null
        break
      case 2: // remove
        addUser = null
        removeUser = this.randUser().id
        break
    }

    const q = this.query(
      `mutation($input: CreateUserOverrideInput!){createUserOverride(input:$input){id}}`,
      {
        input: {
          scheduleID,
          start: new Date(startUnix).toISOString(),
          end: new Date(endUnix).toISOString(),
          addUserID: addUser,
          removeUserID: removeUser,
        } as CreateUserOverrideInput,
      },
    )

    return this.userOverride(q.data.createUserOverride.id)
  }

  schedule(id: string): Schedule {
    return new Schedule(this, id)
  }
  schedules(): Array<Schedule> {
    return this.query(`query{schedules{nodes{id}}}`).data.schedules.nodes.map(
      (obj: IDNode) => this.schedule(obj.id),
    )
  }
  randSchedule(): Schedule {
    return gen.pickone(this.schedules())
  }
  newSchedule(): Schedule {
    const q = this.query(
      `mutation($input: CreateScheduleInput!){createSchedule(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.string({ alpha: true, length: 20 }),
          description: gen.sentence(),
          timeZone: genTZ(),
        },
      },
    )
    const id = q.data.createSchedule.id
    return this.schedule(id)
  }

  service(id: string): Service {
    return new Service(this, id)
  }
  services(): Array<Service> {
    return this.query(`query{services{nodes{id}}}`).data.services.nodes.map(
      (u: IDNode) => this.service(u.id),
    )
  }
  randService(): Service {
    return gen.pickone(this.services())
  }

  escalationPolicy(id: string): EP {
    return new EP(this, id)
  }
  escalationPolicies(): Array<EP> {
    return this.query(
      `query{escalationPolicies{nodes{id}}}`,
    ).data.escalationPolicies.nodes.map((u: IDNode) =>
      this.escalationPolicy(u.id),
    )
  }
  randEP(): EP {
    return gen.pickone(this.escalationPolicies())
  }

  newEP() {
    const q = this.query(
      `mutation($input: CreateEscalationPolicyInput!){createEscalationPolicy(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.string({ alpha: true, length: 20 }),
          description: gen.sentence(),
          repeat: gen.integer({ min: 1, max: 5 }),
        },
      },
    )
    const id = q.data.createEscalationPolicy.id
    return new EP(this, id)
  }

  newUser() {
    const q = this.query(
      `mutation($input: CreateUserInput!){createUser(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.name(),
          email: gen.email(),
          role: gen.pickone(['user', 'admin']),
          username: gen.string({ alpha: true, length: 20, casing: 'lower' }),
          password: gen.string({ alpha: true, length: 20 }),
        },
      },
    )
    const id = q.data.createUser.id
    return new User(this, id)
  }

  newService(epID?: string): Service {
    if (!epID) {
      epID = this.randEP().id
    }

    const q = this.query(
      `mutation($input: CreateServiceInput!){createService(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.string({ alpha: true, length: 20 }),
          description: gen.sentence(),
          escalationPolicyID: epID,
        },
      },
    )
    const id = q.data.createService.id
    return new Service(this, id)
  }

  newRotation() {
    const q = this.query(
      `mutation($input: CreateRotationInput!){createRotation(input:$input){id}}`,
      {
        input: {
          name: 'K6 ' + gen.string({ alpha: true, length: 20 }),
          description: gen.sentence(),
          timeZone: genTZ(),
          start: gen.date().toISOString(),
          type: gen.pickone(['hourly', 'weekly']),
          shiftLength: gen.integer({ min: 1, max: 20 }),
        },
      },
    )
    const id = q.data.createRotation.id
    return new Rotation(this, id)
  }

  user(id?: string) {
    if (!id) {
      id = this.query(`query{user{id}}`).data.user.id as string
    }
    return new User(this, id)
  }
  rotation(id: string) {
    return new Rotation(this, id)
  }

  randUser() {
    // don't return current user
    const users = this.users()
    const thisUser = this.user()
    while (true) {
      const u = gen.pickone(users)
      if (u.id === thisUser.id) {
        continue
      }

      return new User(this, u.id)
    }
  }
  users(): Array<User> {
    return this.query(`query{users{nodes{id}}}`).data.users.nodes.map(
      (u: IDNode) => this.user(u.id),
    )
  }

  randRotation() {
    return gen.pickone(this.rotations())
  }
  rotations(): Array<Rotation> {
    return this.query(`query{rotations{nodes{id}}}`).data.rotations.nodes.map(
      (u: IDNode) => this.rotation(u.id),
    )
  }

  query(query: string, variables = {}) {
    const resp = http.post(
      this.baseURL + '/api/graphql',
      JSON.stringify({
        query,
        variables,
      }),
      {
        headers: {
          'Content-Type': 'application/json',
        },
      },
    )

    const res = JSON.parse(resp.body)

    if (res.errors) {
      throw new Error(res.errors[0].message)
    }

    return res
  }
}
