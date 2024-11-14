import http from 'k6/http'
import { Rotation } from './rotation.ts'
import { User } from './user.ts'
import { login } from './login.ts'

type Node = {
  id: string
}

interface CreateableClass<T, P> {
  new (gql: GraphQL, id: string): T
  randParams(gqp: GraphQL): P
  name: string
}

// GraphQL is a helper class to interact
// with the GraphQL API.
export class GraphQL {
  constructor(
    public token: string,
    public host: string = 'http://localhost:3030',
  ) {}

  // We use any here since we don't have an automated way to type GraphQL queries yet. Types are introduced on the method level.
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  query(query: string, variables: Record<string, unknown> = {}): any {
    const res = http.post(
      this.host + '/api/graphql',
      JSON.stringify({ query, variables }),
      {
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${this.token}`,
        },
      },
    )

    if (res.status !== 200) {
      console.log('Sent:', JSON.stringify({ query, variables }))
      throw new Error(`Unexpected status code: ${res.status}\n` + res.body)
    }

    const body = JSON.parse(res.body as string) as unknown as {
      errors: unknown[]
      data: object
    }

    if (body.errors) {
      throw new Error(`GraphQL errors: ${JSON.stringify(body.errors)}`)
    }

    return body.data
  }

  // userID returns the ID of the currently logged in user.
  get userID(): string {
    return this.query(
      `{
        user {
          id
        }
      }`,
    ).user.id
  }

  protected create<T>(ClassType: CreateableClass<T, unknown>): T {
    return this.createWith(ClassType, ClassType.randParams(this))
  }

  private createWith<T, P>(ClassType: CreateableClass<T, P>, params: P): T {
    const id = this.query(
      `mutation Create${ClassType.name}($input: Create${ClassType.name}Input!){create${ClassType.name}(input: $input){id}}`,
      {
        input: params,
      },
    )[`create${ClassType.name}`].id
    return new ClassType(this, id)
  }

  protected list<T>(ClassType: CreateableClass<T, unknown>): T[] {
    return this.query(`{${ClassType.name.toLowerCase()}s{nodes{id}}}`)[
      `${ClassType.name}s`
    ].nodes.map((n: Node) => new ClassType(this, n.id))
  }

  get rotations(): Rotation[] {
    return this.list(Rotation)
  }

  rotation(id: string): Rotation {
    return new Rotation(this, id)
  }

  genRotation(): Rotation {
    return this.create(Rotation)
  }

  get users(): User[] {
    return this.query(`{users{nodes{id}}}`).users.nodes.map((n: Node) =>
      this.user(n.id),
    )
  }

  user(id: string): User {
    return new User(this, id)
  }
  genUser(): User {
    return this.create(User)
  }

  // newAdmin will generate a user and then log in as that user, returning the token.
  //
  // The returned User object is the admin user, linked to the _original_ token (which makes it safe to call delete() on it).
  newAdmin(): [GraphQL, User] {
    const params = User.randParams(this)
    params.role = 'admin'
    const user = this.createWith(User, params)
    return [new GraphQL(login(params.username, params.password)), user]
  }
}
