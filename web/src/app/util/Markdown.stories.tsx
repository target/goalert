import type { Meta, StoryObj } from '@storybook/react'
import { expect, within } from '@storybook/test'
import Markdown from './Markdown'
import { DateTime } from 'luxon'

const meta = {
  title: 'util/Markdown',
  component: Markdown,
  argTypes: {
    value: {
      control: 'textarea',
    },
  },
  decorators: [],
  tags: ['autodocs'],
} satisfies Meta<typeof Markdown>

export default meta
type Story = StoryObj<typeof meta>

export const BasicMarkdown: Story = {
  args: {
    value: '# Hello, world!',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure the markdown is rendered as an h1
    await expect(
      await canvas.findByRole('heading', { level: 1 }),
    ).toHaveTextContent('Hello, world!')
  },
}

export const Timestamps: Story = {
  args: {
    value: 'Timestamps: 2022-01-01T00:00:00Z',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await canvas.findByText('Timestamps:') // wait for render

    const el = canvasElement.querySelector('time')

    await expect(
      DateTime.fromISO(el?.getAttribute('datetime') || '')
        .toUTC() // convert to UTC, in case the browser's timezone is different
        .toISO(),
    ).toBe('2022-01-01T00:00:00.000Z')
  },
}

const validLinkDoc = `
Valid:
- [example.com/foo](https://example.com/foo)
- [example.com](http://example.com)
- [non-domain-name](https://non-domain.example.com)
- [queryparam.example.com](http://queryparam.example.com?test=1)
- [http://exact.example.com/foo?test=1&bar=2](http://exact.example.com/foo?test=1&bar=2)
- [http://exact2.example.com/foo?test=1&bar=2](http://exact2.example.com/foo?test=1&amp;bar=2)
- [http://exact3.example.com/foo?test=1&amp;bar=2](http://exact3.example.com/foo?test=1&bar=2)
- https://plain.example.com
- <http://bad-slack.example.com | sometext>
- https://escapable.example.com/foo?test=a|b&bar=2
`

export const ValidLinks: Story = {
  args: {
    value: validLinkDoc,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await canvas.findByText('Valid:') // wait for render

    await expect(await canvas.findByText('example.com/foo')).toHaveAttribute(
      'href',
      'https://example.com/foo',
    )
    await expect(await canvas.findByText('example.com')).toHaveAttribute(
      'href',
      'http://example.com',
    )
    await expect(await canvas.findByText('non-domain-name')).toHaveAttribute(
      'href',
      'https://non-domain.example.com',
    )
    await expect(
      await canvas.findByText('queryparam.example.com'),
    ).toHaveAttribute('href', 'http://queryparam.example.com?test=1')
    await expect(
      await canvas.findByText('http://exact.example.com/foo?test=1&bar=2'),
    ).toHaveAttribute('href', 'http://exact.example.com/foo?test=1&bar=2')
    await expect(
      await canvas.findByText('http://exact2.example.com/foo?test=1&bar=2'),
    ).toHaveAttribute('href', 'http://exact2.example.com/foo?test=1&bar=2')
    await expect(
      await canvas.findByText('http://exact3.example.com/foo?test=1&bar=2'),
    ).toHaveAttribute('href', 'http://exact3.example.com/foo?test=1&bar=2')
    await expect(
      await canvas.findByText('https://plain.example.com'),
    ).toHaveAttribute('href', 'https://plain.example.com')
    await expect(
      await canvas.findByText('http://bad-slack.example.com'),
    ).toHaveAttribute('href', 'http://bad-slack.example.com')

    // ensure the pipe character is escaped
    await expect(
      await canvas.findByText(
        'https://escapable.example.com/foo?test=a|b&bar=2',
      ),
    ).toHaveAttribute(
      'href',
      'https://escapable.example.com/foo?test=a%7Cb&bar=2',
    )
  },
}

const invalidLinkDoc = `
Invalid:
- [example.com/wrongpath](https://example.com/foo)
- [wrongdomain.example.com](http://example.com)
- [http://wrongQuery.example.com/foo?test=1&bar=3](http://wrongQuery.example.com/foo?test=1&bar=2)
`

export const InvalidLinks: Story = {
  args: {
    value: invalidLinkDoc,
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await canvas.findByText('Invalid:') // wait for render

    expect(canvasElement.querySelectorAll('a')).toHaveLength(0) // no links should be rendered
  },
}
