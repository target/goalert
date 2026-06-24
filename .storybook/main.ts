import type { StorybookConfig } from '@storybook/react-vite'

const config: StorybookConfig = {
  staticDirs: ['./static'],
  stories: ['../web/src/**/*.stories.@(js|jsx|mjs|ts|tsx)'],
  addons: ['@storybook/addon-links'],

  framework: {
    name: '@storybook/react-vite',
    options: {},
  },
}
export default config
