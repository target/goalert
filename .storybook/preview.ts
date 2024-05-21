import {
  defaultConfig,
  mockConfig,
  mockExpFlags,
} from './../web/src/app/storybook/graphql'
import type { Preview } from '@storybook/react'
import DefaultDecorator from '../web/src/app/storybook/decorators'

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    fetchMock: {
      debug: true,
      useFetchMock: (fetchMock) => {
        fetchMock.config.overwriteRoutes = false
      },
      catchAllMocks: [
        mockConfig(defaultConfig),
        mockExpFlags(),
        {
          matcher: {
            name: 'GraphQL Catch All',
            url: 'path:/api/graphql',
            response: (name, req) => {
              const body = JSON.parse(req.body)
              return {
                body: {
                  errors: [
                    {
                      message: `No mocks defined for operation '${body.operationName}'.`,
                    },
                  ],
                },
              }
            },
          },
        },
      ],
    },
  },
  decorators: [DefaultDecorator],
}

export default preview
