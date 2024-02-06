import React from 'react'
import type { Meta, StoryObj, } from '@storybook/react'
import UserContactMethodCreateDialogDest from './UserContactMethodCreateDialogDest'
import { expect } from '@storybook/jest'
import { within, screen, userEvent } from '@storybook/testing-library'
import { handleDefaultConfig, defaultConfig } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'util/UserContactMethodCreateDialogDest',
  component: UserContactMethodCreateDialogDest,
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [handleDefaultConfig,
        graphql.query('UserConflictCheck', () => {
    return HttpResponse.json({
      data: { users: { nodes: [{name: defaultConfig.user.name ,id: defaultConfig.user.id}] }},
    })
  }),
  graphql.query('useExpFlag', () => {
    return HttpResponse.json({
      data: { users: { nodes: [{name: defaultConfig.user.name ,id: defaultConfig.user.id}] }},
    })
  }),
    graphql.mutation('CreateUserContactMethodInput', ({ variables: vars }) => {
        return HttpResponse.json({
        data: {
            createUserContactMethod: { id: '00000000-0000-0000-0000-000000000000'}
        },
        })
    }),
    graphql.query('ValidateDestination', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValidate:
                vars.input.value === 'https://test.com' ||
                vars.input.value === '+12225558989' ||
                vars.input.value === 'valid@email.com',
            },
          })
        })
    ],
    },
  },
  render: function Component(args) {
    const [, setArgs] = useArgs()
    const onClose = (contactMethodID: string | undefined): void => {
      if (args.onClose) args.onClose(contactMethodID)
      setArgs({ value: contactMethodID })
    }
    return <UserContactMethodCreateDialogDest {...args} onClose={onClose} />
  },
} satisfies Meta<typeof UserContactMethodCreateDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const SingleField: Story = {
  args: {
    userID: defaultConfig.user.id,
    title: 'Create New Contact Method',
    subtitle: 'Create New Contact Method Subtitle',
  },
  play: async ({ canvasElement }) => {
    // const canvas = within(canvasElement)

    await expect(
      screen.getByPlaceholderText('11235550123'),
    ).toBeVisible()

    await expect(screen.getByLabelText('Phone Number')).toBeVisible()
    // ensure check icon for valid number
    await userEvent.clear(screen.getByLabelText('Phone Number'))
    await userEvent.type(screen.getByLabelText('Phone Number'), '12225558989')
    await expect(await screen.findByTestId('CheckIcon')).toBeVisible()
    await expect(
      screen.getByText(
        'Include country code e.g. +1 (USA), +91 (India), +44 (UK)',
      )
    ).toBeVisible()
    await expect(screen.getByText('+')).toBeVisible()

    const submitButton = await screen.getByRole('button', { name: /SUBMIT/i });
    await userEvent.click(submitButton)

    // await expect(screen.getByText('Please fill out this field.')).toBeVisible()
    // await expect(within(document.body).findByText('Please fill out this field.').toBeVisible())

  },
}

export const MultiField: Story = {
  args: {
    userID: defaultConfig.user.id,
    title: 'Create New Contact Method',
    subtitle: 'Create New Contact Method Subtitle',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure information for phone number renders correctly
    await expect(canvas.getByLabelText('First Item')).toBeVisible()
    await expect(canvas.getByText('Some hint text')).toBeVisible()
    await expect(canvas.getByText('+')).toBeVisible()
    await expect(canvas.getByPlaceholderText('11235550123')).toBeVisible()

    // ensure information for email renders correctly
    await expect(
      canvas.getByPlaceholderText('foobar@example.com'),
    ).toBeVisible()
    await expect(canvas.getByLabelText('Second Item')).toBeVisible()

    // ensure information for slack renders correctly
    await expect(canvas.getByPlaceholderText('slack user ID')).toBeVisible()
    await expect(canvas.getByLabelText('Third Item')).toBeVisible()
  },
}

export const DisabledField: Story = {
  args: {
    userID: defaultConfig.user.id,
    title: 'Create New Contact Method',
    subtitle: 'Create New Contact Method Subtitle',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    // ensure information renders correctly
    await expect(
      canvas.getByPlaceholderText('This field is disabled.'),
    ).toBeVisible()
  },
}

// export const FieldError: Story = {
//   args: {
//     destType: 'triple-field',
//     value: [
//       {
//         fieldID: 'first-field',
//         value: '',
//       },
//       {
//         fieldID: 'second-field',
//         value: 'test@example.com',
//       },
//       {
//         fieldID: 'third-field',
//         value: '',
//       },
//     ],
//     disabled: false,
//     destFieldErrors: [
//       {
//         fieldID: 'third-field',
//         message: 'This is an error message (third)',
//       },
//       {
//         fieldID: 'first-field',
//         message: 'This is an error message (first)',
//       },
//     ],
//   },
// }
