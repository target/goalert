import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodCreateDialogDest from './UserContactMethodCreateDialogDest'
import { expect } from '@storybook/jest'
import { screen, userEvent, waitFor } from '@storybook/testing-library'
import { handleDefaultConfig, defaultConfig } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'users/UserContactMethodCreateDialogDest',
  component: UserContactMethodCreateDialogDest,
  tags: ['autodocs'],
  parameters: {
    msw: {
      handlers: [
        handleDefaultConfig,
        graphql.query('UserConflictCheck', () => {
          return HttpResponse.json({
            data: {
              users: {
                nodes: [
                  { name: defaultConfig.user.name, id: defaultConfig.user.id },
                ],
              },
            },
          })
        }),
        graphql.query('useExpFlag', () => {
          return HttpResponse.json({
            data: {
              users: {
                nodes: [
                  { name: defaultConfig.user.name, id: defaultConfig.user.id },
                ],
              },
            },
          })
        }),
        graphql.mutation('CreateUserContactMethodInput', () => {
          return HttpResponse.json({
            data: {
              createUserContactMethod: {
                id: '00000000-0000-0000-0000-000000000000',
              },
            },
          })
        }),
        graphql.query('ValidateDestination', ({ variables: vars }) => {
          return HttpResponse.json({
            data: {
              destinationFieldValidate:
                vars.input.value === '@slack' ||
                vars.input.value === '+12225558989' ||
                vars.input.value === 'valid@email.com',
            },
          })
        }),
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
  play: async () => {
    await userEvent.clear(screen.getByLabelText('Phone Number'))

    await waitFor(async () => {
      await userEvent.type(screen.getByLabelText('Phone Number'), '12225558989')
    })
    await expect(await screen.findByTestId('CheckIcon')).toBeVisible()

    const submitButton = await screen.getByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)

    await userEvent.clear(screen.getByLabelText('Name'))
    await userEvent.type(screen.getByLabelText('Name'), 'TEST')

    const retryButton = await screen.getByRole('button', { name: /RETRY/i })
    await userEvent.click(retryButton)
  },
}

export const MultiField: Story = {
  args: {
    userID: defaultConfig.user.id,
    title: 'Create New Contact Method',
    subtitle: 'Create New Contact Method Subtitle',
  },
  play: async () => {
    // Select the next Dest Type
    await userEvent.click(await screen.getByLabelText('Dest Type'))
    await userEvent.click(
      await screen.getByText('Multi Field Destination Type'),
    )

    // ensure information for phone number renders correctly
    await userEvent.clear(screen.getByLabelText('First Item'))
    await waitFor(async () => {
      await userEvent.type(screen.getByLabelText('First Item'), '12225558989')
    })
    await expect(await screen.findByTestId('CheckIcon')).toBeVisible()

    // ensure information for email renders correctly
    await expect(
      screen.getByPlaceholderText('foobar@example.com'),
    ).toBeVisible()
    await userEvent.clear(screen.getByLabelText('Second Item'))
    await userEvent.type(
      screen.getByLabelText('Second Item'),
      'valid@email.com',
    )

    // ensure information for slack renders correctly
    await expect(screen.getByPlaceholderText('slack user ID')).toBeVisible()
    await expect(screen.getByLabelText('Third Item')).toBeVisible()
    await userEvent.clear(screen.getByLabelText('Third Item'))
    await userEvent.type(screen.getByLabelText('Third Item'), '@slack')

    // Try to submit without all feilds complete
    const submitButton = await screen.getByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)

    // Name field
    await userEvent.clear(screen.getByLabelText('Name'))
    await userEvent.type(screen.getByLabelText('Name'), 'TEST')

    const retryButton = await screen.getByRole('button', { name: /RETRY/i })
    await userEvent.click(retryButton)
  },
}

export const DisabledField: Story = {
  args: {
    userID: defaultConfig.user.id,
    title: 'Create New Contact Method',
    subtitle: 'Create New Contact Method Subtitle',
  },
  play: async () => {
    // Open option select
    await userEvent.click(await screen.getByLabelText('Dest Type'))

    // Attempt to click the disabled option
    const disabledOption = await screen.getByText('This is disabled')
    // Ensure no clicked occurred
    userEvent.click(disabledOption, {
      pointerEventsCheck: 0,
    })
  },
}
