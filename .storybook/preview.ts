import type { Preview } from '@storybook/react'
import DefaultDecorator from '../web/src/app/storybook/decorators'
import { initialize, mswLoader } from 'msw-storybook-addon'
import {
  handleDefaultConfig,
  handleExpFlags,
} from '../web/src/app/storybook/graphql'

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
    msw: {
      handlers: [handleDefaultConfig, handleExpFlags()],
    },
  },
  decorators: [DefaultDecorator],
  loaders: [mswLoader],
}

export default preview
