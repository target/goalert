import {
  graphQLLoader,
  mswHandler,
} from './../web/src/app/storybook/graphql-loader'
import type { Preview } from '@storybook/react'
import DefaultDecorator from '../web/src/app/storybook/decorators'
import { initialize, mswLoader } from 'msw-storybook-addon'
import { defaultConfig } from '../web/src/app/storybook/graphql'

initialize({
  onUnhandledRequest: 'bypass',
})

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
      RequireConfig: defaultConfig,
      useExpFlag: { data: { experimentalFlags: [] } },
    },
    msw: { handlers: [mswHandler] },
  },
  decorators: [DefaultDecorator],
  loaders: [graphQLLoader, mswLoader],
}

export default preview
