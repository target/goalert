import http from 'k6/http'
import { Rotation } from './rotation.ts'

export class GraphQL {
  constructor(
    public token: string,
    public host: string = 'http://localhost:3030',
  ) {}

  query(query: string, variables: Record<string, unknown> = {}): any {
    let res = http.post(
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

  get userID(): string {
    return this.query(
      `{
        user {
          id
        }
      }`,
    ).user.id
  }

  get rotations(): Rotation[] {
    return this.query(`{rotations{nodes{id}}}`).rotations.nodes.map(
      (n) => new Rotation(this, n.id),
    )
  }
}
