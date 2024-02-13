import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodCreateDialogDest from './UserContactMethodCreateDialogDest'
import { expect, userEvent, waitFor, within } from '@storybook/test'
import {
  handleDefaultConfig,
  defaultConfig,
  handleExpFlags,
} from '../storybook/graphql'
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
        handleExpFlags('dest-types'),
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
        graphql.mutation(
          'CreateUserContactMethodInput',
          ({ variables: vars }) => {
            if (vars.input.name === 'error-test') {
              return HttpResponse.json({
                data: null,
                errors: [
                  {
                    message: 'This is a field-error',
                    path: [
                      'createUserContactMethod',
                      'input',
                      'dest',
                      'values',
                      'phone-number',
                    ],
                    extensions: {
                      code: 'INVALID_DESTINATION_FIELD_VALUE',
                    },
                  },
                  {
                    message: 'This indicates an invalid destination type',
                    path: ['createUserContactMethod', 'input', 'dest', 'type'],
                    extensions: {
                      code: 'INVALID_DESTINATION_TYPE',
                    },
                  },
                  {
                    message: 'This is a generic error',
                  },
                ],
              })
            }
            return HttpResponse.json({
              data: {
                createUserContactMethod: {
                  id: '00000000-0000-0000-0000-000000000000',
                },
              },
            })
          },
        ),
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
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    await userEvent.clear(await canvas.findByPlaceholderText('11235550123'))

    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByPlaceholderText('11235550123'),
        '12225558989',
      )
    })

    const submitButton = await canvas.findByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)

    await userEvent.clear(await canvas.findByLabelText('Name'))
    await userEvent.type(await canvas.findByLabelText('Name'), 'TEST')

    const retryButton = await canvas.findByRole('button', { name: /RETRY/i })
    await userEvent.click(retryButton)
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
    // Select the next Dest Type
    await userEvent.click(await canvas.findByLabelText('Dest Type'))
    await userEvent.click(
      await canvas.findByText('Multi Field Destination Type'),
    )

    // ensure information for phone number renders correctly
    await userEvent.clear(await canvas.findByLabelText('First Item'))
    await waitFor(async () => {
      await userEvent.type(
        await canvas.findByLabelText('First Item'),
        '12225558989',
      )
    })

    await waitFor(async () => {
      await expect(await canvas.findByTestId('CheckIcon')).toBeVisible()
    })

    // ensure information for email renders correctly
    await expect(
      await canvas.findByPlaceholderText('foobar@example.com'),
    ).toBeVisible()
    await userEvent.clear(
      await canvas.findByPlaceholderText('foobar@example.com'),
    )
    await userEvent.type(
      await await canvas.findByPlaceholderText('foobar@example.com'),
      'valid@email.com',
    )

    // ensure information for slack renders correctly
    await expect(
      await canvas.findByPlaceholderText('slack user ID'),
    ).toBeVisible()
    await expect(await canvas.findByLabelText('Third Item')).toBeVisible()
    await userEvent.clear(await canvas.findByLabelText('Third Item'))
    await userEvent.type(await canvas.findByLabelText('Third Item'), '@slack')

    // Try to submit without all feilds complete
    const submitButton = await canvas.findByRole('button', { name: /SUBMIT/i })
    await userEvent.click(submitButton)

    // Name field
    await userEvent.clear(await canvas.findByLabelText('Name'))
    await userEvent.type(await canvas.findByLabelText('Name'), 'TEST')

    const retryButton = await canvas.findByRole('button', { name: /RETRY/i })
    await userEvent.click(retryButton)
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
    // Open option select
    await userEvent.click(await canvas.findByLabelText('Dest Type'))

    // Ensure disabled
    await expect(
      await canvas.findByLabelText(
        'Send alert status updates (not supported for this type)',
      ),
    ).toBeDisabled()
  },
}

export const ErrorField: Story = {
  args: {
    userID: defaultConfig.user.id,
    title: 'Create New Contact Method',
    subtitle: 'Create New Contact Method Subtitle',
  },

  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)

    await userEvent.type(await canvas.findByLabelText('Name'), 'error-test')
    await userEvent.type(
      await canvas.findByPlaceholderText('11235550123'),
      '123',
    )

    const submitButton = await canvas.findByText('Submit')
    await userEvent.click(submitButton)

    await waitFor(async () => {
      await expect(
        await canvas.findByText('This is a field-error'),
      ).toBeVisible()
      await expect(
        await canvas.findByText('This indicates an invalid destination type'),
      ).toBeVisible()
      await expect(
        await canvas.findByText('This is a generic error'),
      ).toBeVisible()
    })
  },
}
