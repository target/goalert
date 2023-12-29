import { Page, Request, Route } from 'playwright/test'
import { ConfigID, ConfigType, Mutation, Query } from '../schema'

type RecursivePartial<T> = {
  [P in keyof T]?: T[P] extends (infer U)[]
    ? RecursivePartial<U>[]
    : T[P] extends object | undefined
      ? RecursivePartial<T[P]>
      : T[P]
}

type GraphQLResponse = {
  data?: RecursivePartial<Query | Mutation>
  errors?: { message: string }[] // TODO
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

export class GQLMock {
  page: Page
  mocks: Record<string, OperationHandler> = { ...defaultMocks }
  raw: Record<string, RawHandler> = {}

  constructor(page: Page) {
    this.page = page
  }

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
  }

  setRaw(name: string, handler: RawHandler): void {
    delete this.mocks[name]
    this.raw[name] = handler
  }

  setGQL(name: string, handler: OperationHandler): void {
    delete this.raw[name]
    this.mocks[name] = handler
  }
}
