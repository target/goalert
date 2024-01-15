import type { Preview } from '@storybook/react'
import DefaultDecorator from '../web/src/app/storybook/decorators'
import { initialize, mswLoader } from 'msw-storybook-addon'
import { handleDefaultConfig } from '../web/src/app/storybook/graphql'

initialize({
  onUnhandledRequest: 'bypass',
})

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: '^on[A-Z].*' },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
    msw: {
      handlers: [handleDefaultConfig],
    },
  },
  decorators: [DefaultDecorator],
  loaders: [mswLoader],
}

export default preview
