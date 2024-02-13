import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodCreateDialogDest from './UserContactMethodCreateDialogDest'
import { expect, userEvent, waitFor, screen } from '@storybook/test'
import { handleDefaultConfig, defaultConfig } from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'

const meta = {
  title: 'users/UserContactMethodCreateDialogDest',
  component: UserContactMethodCreateDialogDest,
  tags: ['autodocs'],
  parameters: {
    docs: {
      story: {
        inline: false,
        iframeHeight: 500,
      },
    },
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
    return (
      <UserContactMethodCreateDialogDest
        {...args}
        disablePortal
        onClose={onClose}
      />
    )
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
    await userEvent.clear(await screen.findByPlaceholderText('11235550123'))

    await waitFor(async () => {
      await userEvent.type(
        await screen.findByPlaceholderText('11235550123'),
        '12225558989',
      )
    })

    const submitButton = await screen.findByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)

    await userEvent.clear(await screen.findByLabelText('Name'))
    await userEvent.type(await screen.findByLabelText('Name'), 'TEST')

    const retryButton = await screen.findByRole('button', { name: /RETRY/i })
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
    await userEvent.click(await screen.findByLabelText('Dest Type'))
    await userEvent.click(
      await screen.findByText('Multi Field Destination Type'),
    )

    // ensure information for phone number renders correctly
    await userEvent.clear(await screen.findByLabelText('First Item'))
    await waitFor(async () => {
      await userEvent.type(
        await screen.findByLabelText('First Item'),
        '12225558989',
      )
    })

    await waitFor(async () => {
      await expect(await screen.findByTestId('CheckIcon')).toBeVisible()
    })

    // ensure information for email renders correctly
    await expect(
      await screen.findByPlaceholderText('foobar@example.com'),
    ).toBeVisible()
    await userEvent.clear(
      await screen.findByPlaceholderText('foobar@example.com'),
    )
    await userEvent.type(
      await await screen.findByPlaceholderText('foobar@example.com'),
      'valid@email.com',
    )

    // ensure information for slack renders correctly
    await expect(
      await screen.findByPlaceholderText('slack user ID'),
    ).toBeVisible()
    await expect(await screen.findByLabelText('Third Item')).toBeVisible()
    await userEvent.clear(await screen.findByLabelText('Third Item'))
    await userEvent.type(await screen.findByLabelText('Third Item'), '@slack')

    // Try to submit without all feilds complete
    const submitButton = await screen.findByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)

    // Name field
    await userEvent.clear(await screen.findByLabelText('Name'))
    await userEvent.type(await screen.findByLabelText('Name'), 'TEST')

    const retryButton = await screen.findByRole('button', { name: /RETRY/i })
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
    await userEvent.click(await screen.findByLabelText('Dest Type'))

    // Ensure disabled
    await expect(
      await screen.findByLabelText(
        'Send alert status updates (not supported for this type)',
      ),
    ).toBeDisabled()
  },
}
