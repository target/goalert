import type { Preview } from '@storybook/react'
import DefaultDecorator from '../web/src/app/storybook/decorators'
import { initialize, mswLoader } from 'msw-storybook-addon'
import { defaultConfig } from '../web/src/app/storybook/graphql'
import { graphql, HttpResponse } from 'msw'

initialize({
  onUnhandledRequest: 'bypass',
})

export type GraphQLRequestHandler = (variables: unknown) => object
export type GraphQLParams = Record<string, GraphQLRequestHandler>

const componentConfig: Record<string, GraphQLParams> = {}

export const mswHandler = graphql.operation((params) => {
  const url = new URL(params.request.url)
  const gql = componentConfig[url.pathname]
  const handler = gql[params.operationName]
  if (!handler) {
    return HttpResponse.json({
      errors: [
        { message: `No mock defined for operation '${params.operationName}'.` },
      ],
    })
  }

  if (typeof handler === 'function') {
    return HttpResponse.json(handler(params.variables))
  }

  return HttpResponse.json(handler)
})

interface LoaderArg {
  id: string
  parameters: {
    graphql?: GraphQLParams
  }
}

/**
 * graphqlLoader is a loader function that sets up the GraphQL mocks for the
 * component. It stores them in a global object, componentConfig, which is used
 * by the mswHandler to resolve the mocks.
 *
 * We need to do this because the browser will render all components at once,
 * and so we need to handle all GraphQL mocks at once.
 *
 * The way this works is that each component get's a unique URL for GraphQL
 * (see decorators.tsx). This URL is used to store the GraphQL mocks for that
 * component in componentConfig.
 */
export function graphQLLoader(arg: LoaderArg): void {
  const path = '/' + encodeURIComponent(arg.id) + '/api/graphql'
  componentConfig[path] = arg.parameters.graphql || {}
}

const preview: Preview = {
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    graphql: {
      // default GraphQL handlers
      RequireConfig: { data: defaultConfig },
      useExpFlag: { data: { experimentalFlags: [] } },
    },
    msw: { handlers: [mswHandler] },
  },
  decorators: [DefaultDecorator],
  loaders: [graphQLLoader, mswLoader],
}

export default preview
