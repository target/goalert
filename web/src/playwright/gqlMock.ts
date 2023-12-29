import { Page, Request, Route } from 'playwright/test'
import { ConfigID, ConfigType, Mutation, Query } from '../schema'

type DeepPartial<T> = T extends object
  ? {
      [P in keyof T]?: DeepPartial<T[P]>
    }
  : T

type Error = FieldError | MultiFieldError | GenericError

interface GenericError {
  message: string
}

interface FieldError extends GenericError {
  message: string
  path: string[]
  extensions: {
    fieldName: string
    isFieldError: true
  }
}

interface MultiFieldError extends GenericError {
  message: string
  extensions: {
    isMultiFieldError: true
    errors: {
      message: string
      fieldName: string
    }[]
  }
}

interface GraphQLResponse {
  data?: DeepPartial<Query | Mutation>
  errors?: Error[]
}

export type OperationHandler = (
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  variables: Record<string, any>,
  query: string,
) => GraphQLResponse

export const defaultMocks: Record<string, OperationHandler> = {
  useExpFlag: () => ({
    data: {
      experimentalFlags: [],
    },
  }),
  RequireConfigData: () => ({
    data: {
      user: {
        id: '00000000-0000-0000-0000-000000000000',
        name: 'Test User',
        role: 'admin',
      },
      config: [
        {
          id: 'General.ApplicationName',
          type: 'string',
          value: 'ComponentTest',
        },
      ] as { id: ConfigID; type: ConfigType; value: string }[],
      integrationKeyTypes: [
        {
          id: 'example-enabled',
          name: 'Enabled Example',
          label: 'Enabled Example Value',
          enabled: true,
        },
        {
          id: 'example-disabled',
          name: 'Disabled Example',
          label: 'Disabled Example Value',
          enabled: false,
        },
      ],
    },
  }),
}

type RawHandler = (route: Route, request: Request) => unknown

/** GQLMock is a helper class for mocking GraphQL requests.
 *
 * Note: `init()` must be called before any other methods.
 *
 * Example:
 * ```
 *   const mock = new GQLMock(page)
 *   await mock.init()
 *   mock.setGQL('MyQuery', (vars, query) => {
 *     expect(vars).toMatchObject({ number: '123' })
 *     expect(query).toMatch(/query MyQuery/)
 *     return {
 *       data: {
 *         myQuery: {
 *           id: '123',
 *           name: 'Test',
 *         },
 *       },
 *     }
 *   })
 * ```
 *  */
export class GQLMock {
  private page: Page
  private mocks: Record<string, OperationHandler> = { ...defaultMocks }
  private raw: Record<string, RawHandler> = {}
  private _init = false

  constructor(page: Page) {
    this.page = page
  }

  /** init must be called before any other methods.
   *
   * This will register the route handler for /api/graphql.
   */
  async init(): Promise<void> {
    await this.page.route('/api/graphql', (route, req) => {
      const body: {
        operationName: string
        query: string
        variables: { number: string }
      } = route.request().postDataJSON()

      if (this.raw[body.operationName]) {
        return this.raw[body.operationName](route, req)
      }

      if (this.mocks[body.operationName]) {
        const resp = this.mocks[body.operationName](body.variables, body.query)
        return route.fulfill({
          status: 200,
          contentType: 'application/json',
          body: JSON.stringify(resp),
        })
      }

      throw new Error(`no mock for ${body.operationName}`)
    })
    this._init = true
  }

  /** setRaw allows direct access to the Route and Request objects.
   *
   * This can be used to similate a server 500 error, or other
   * non-200 response.
   *
   * For example, to simulate a server error:
   * ```
   * mock.setRaw('MyQuery', (route, request) => {
   *  route.fulfill({
   *   status: 500,
   *   contentType: 'application/json',
   *   body: JSON.stringify({ error: 'server error' }),
   * })
   * ```
   */
  setRaw(operationName: string, handler: RawHandler): void {
    if (!this._init) throw new Error('must call init() first')
    delete this.mocks[operationName]
    this.raw[operationName] = handler
  }

  /** setGQL allows mocking a GraphQL response.
   *
   * This can be used to simulate the response for a single named
   * GraphQL operation.
   *
   *
   * For example, if the application registers the following query:
   * ```
   * const query = gql
   *   query GetUserName($id: ID!) {
   *     user(id: $id) {
   *       id
   *       name
   *     }
   *   }`
   * ```
   *
   * Then the following code will mock the response:
   * ```
   * mock.setGQL('GetUserName', (vars) => {
   *   expect(vars).toMatchObject({ id: '123' })
   *   return {
   *     data: {
   *       user: {
   *         id: '123',
   *         name: 'Test',
   *       },
   *     },
   *  }
   * ```
   */
  setGQL(operationName: string, handler: OperationHandler): void {
    if (!this._init) throw new Error('must call init() first')
    delete this.raw[operationName]
    this.mocks[operationName] = handler
  }
}
