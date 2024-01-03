import type { Preview } from '@storybook/react'
import DefaultDecorator from '../web/src/app/storybook/decorators'
import { initialize, mswLoader } from 'msw-storybook-addon'

initialize()

const preview: Preview = {
  parameters: {
    actions: { argTypesRegex: '^on[A-Z].*' },
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
  decorators: [DefaultDecorator],
  loaders: [mswLoader],
}

export default preview
