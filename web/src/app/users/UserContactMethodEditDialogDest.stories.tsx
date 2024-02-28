import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodEditDialogDest from './UserContactMethodEditDialogDest'
import { expect, userEvent, waitFor, within, screen } from '@storybook/test'
import {
  handleDefaultConfig,
  // defaultConfig,
  handleExpFlags,
} from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'
import { DestFieldValueError, InputFieldError } from '../util/errtypes'

const meta = {
  title: 'users/UserContactMethodEditDialogDest',
  component: UserContactMethodEditDialogDest,
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
        graphql.query('userCm', () => {
          return HttpResponse.json({
            data: {
              userContactMethod: {
                id: '00000000-0000-0000-0000-000000000000',
                name: 'single-field contact method',
                dest: {
                  type: 'single-field',
                  values: [
                    {
                      fieldID: 'phone-number',
                      value: '+15555555555',
                      label: '+1 555-555-5555',
                    },
                  ],
                  value: 'http://localhost:8080',
                },
                disabled: false,
                pending: false,
              },
            },
          })
        }),
        graphql.mutation('updateUserContactMethod', ({ variables: vars }) => {
          if (vars.input.name === 'error-test') {
            return HttpResponse.json({
              data: null,
              errors: [
                {
                  message: 'This is a dest field-error',
                  path: ['updateUserContactMethod', 'input', 'dest'],
                  extensions: {
                    code: 'INVALID_DEST_FIELD_VALUE',
                    fieldID: 'phone-number',
                  },
                } satisfies DestFieldValueError,
                {
                  message: 'This indicates an invalid destination type',
                  path: ['updateUserContactMethod', 'input', 'dest', 'type'],
                  extensions: {
                    code: 'INVALID_INPUT_VALUE',
                  },
                } satisfies InputFieldError,
                {
                  message: 'Name error',
                  path: ['updateUserContactMethod', 'input', 'name'],
                  extensions: {
                    code: 'INVALID_INPUT_VALUE',
                  },
                } satisfies InputFieldError,
                {
                  message: 'This is a generic error',
                },
              ],
            })
          }
          return HttpResponse.json({
            data: {
              updateUserContactMethod: {
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
    return <UserContactMethodEditDialogDest {...args} onClose={onClose} />
  },
} satisfies Meta<typeof UserContactMethodEditDialogDest>

export default meta
type Story = StoryObj<typeof meta>

export const SingleField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },
  play: async () => {
    await userEvent.click(await screen.findByLabelText('Destination Type'))

    // incorrectly believes that the following fields are not visible
    expect(
      await screen.findByRole('option', { hidden: true, name: 'Single Field' }),
    ).toBeInTheDocument()
    expect(
      await screen.findByRole('option', { hidden: true, name: 'Multi Field' }),
    ).toBeInTheDocument()
    expect(
      await screen.findByText('This is disabled'), // does not register as an option
    ).toBeInTheDocument()
    expect(
      await screen.findByRole('option', {
        hidden: true,
        name: 'Single With Status',
      }),
    ).toBeInTheDocument()
    expect(
      await screen.findByRole('option', {
        hidden: true,
        name: 'Single With Required Status',
      }),
    ).toBeInTheDocument()
  },
}

export const MultiField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },
  play: async () => {
    // Select the multi-field Dest Type
    await userEvent.click(await screen.findByLabelText('Destination Type'))
    await userEvent.click(
      await screen.findByRole('option', { hidden: true, name: 'Multi Field' }),
    )

    await expect(await screen.findByLabelText('Name')).toBeVisible()
    await expect(await screen.findByLabelText('Destination Type')).toBeVisible()
    await expect(await screen.findByLabelText('First Item')).toBeVisible()
    await expect(await screen.findByLabelText('Second Item')).toBeVisible()
    await expect(await screen.findByLabelText('Third Item')).toBeVisible()
  },
}

export const StatusUpdates: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },
  play: async () => {
    // Open option select
    await userEvent.click(await screen.findByLabelText('Destination Type'))
    await userEvent.click(
      await screen.findByRole('option', { hidden: true, name: 'Single Field' }),
    )
    await expect(
      await screen.findByLabelText(
        'Send alert status updates (not supported for this type)',
      ),
    ).toBeDisabled()

    await userEvent.click(await screen.findByLabelText('Destination Type'))
    await userEvent.click(
      await screen.findByRole('option', {
        hidden: true,
        name: 'Single With Status',
      }),
    )
    await expect(
      await screen.findByLabelText('Send alert status updates'),
    ).not.toBeDisabled()

    await userEvent.click(await screen.findByLabelText('Destination Type'))
    await userEvent.click(
      await screen.findByRole('option', {
        hidden: true,
        name: 'Single With Required Status',
      }),
    )
    await expect(
      await screen.findByLabelText(
        'Send alert status updates (cannot be disabled for this type)',
      ),
    ).toBeDisabled()
  },
}

export const ErrorField: Story = {
  args: {
    contactMethodID: '00000000-0000-0000-0000-000000000000',
  },

  play: async () => {
    await userEvent.clear(await screen.findByLabelText('Name'))
    await userEvent.type(await screen.findByLabelText('Name'), 'error-test')
    await userEvent.type(
      await screen.findByPlaceholderText('11235550123'),
      '123',
    )

    const submitButton = await screen.findByText('Submit')
    await userEvent.click(submitButton)

    // response should set error on all fields plus the generic error
    await waitFor(
      async () => {
        await expect(await screen.findByLabelText('Name')).toBeInvalid()

        await expect(await screen.findByText('Name error')).toBeVisible()

        await expect(
          // mui puts aria-invalid on the input, but not the combobox (which the label points to)
          await screen.findByText('This indicates an invalid destination type'),
        ).toBeVisible()
        await expect(await screen.findByLabelText('Phone Number')).toBeInvalid()
        await expect(
          await screen.findByText('This is a dest field-error'),
        ).toBeVisible()

        await expect(
          await screen.findByText('This is a generic error'),
        ).toBeVisible()
      },
      { timeout: 5000 },
    )
  },
}
