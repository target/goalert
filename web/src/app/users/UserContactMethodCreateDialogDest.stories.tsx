import React from 'react'
import type { Meta, StoryObj } from '@storybook/react'
import UserContactMethodCreateDialogDest from './UserContactMethodCreateDialogDest'
import { expect, fn, userEvent, waitFor, within } from '@storybook/test'
import {
  handleDefaultConfig,
  defaultConfig,
  handleExpFlags,
} from '../storybook/graphql'
import { useArgs } from '@storybook/preview-api'
import { HttpResponse, graphql } from 'msw'
import { DestFieldValueError, InputFieldError } from '../util/errtypes'

const meta = {
  title: 'users/UserContactMethodCreateDialogDest',
  component: UserContactMethodCreateDialogDest,
  tags: ['autodocs'],
  args: {
    onClose: fn(),
  },
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
        graphql.mutation(
          'CreateUserContactMethodInput',
          ({ variables: vars }) => {
            if (vars.input.name === 'error-test') {
              return HttpResponse.json({
                data: null,
                errors: [
                  {
                    message: 'This is a dest field-error',
                    path: ['createUserContactMethod', 'input', 'dest'],
                    extensions: {
                      code: 'INVALID_DEST_FIELD_VALUE',
                      fieldID: 'phone-number',
                    },
                  } satisfies DestFieldValueError,
                  {
                    message: 'This indicates an invalid destination type',
                    path: ['createUserContactMethod', 'input', 'dest', 'type'],
                    extensions: {
                      code: 'INVALID_INPUT_VALUE',
                    },
                  } satisfies InputFieldError,
                  {
                    message: 'Name error',
                    path: ['createUserContactMethod', 'input', 'name'],
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
    await userEvent.click(await canvas.findByLabelText('Destination Type'))

    // incorrectly believes that the following fields are not visible
    expect(
      await canvas.findByRole('option', { hidden: true, name: 'Single Field' }),
    ).toBeInTheDocument()
    expect(
      await canvas.findByRole('option', { hidden: true, name: 'Multi Field' }),
    ).toBeInTheDocument()
    expect(
      await canvas.findByText('This is disabled'), // does not register as an option
    ).toBeInTheDocument()
    expect(
      await canvas.findByRole('option', {
        hidden: true,
        name: 'Single With Status',
      }),
    ).toBeInTheDocument()
    expect(
      await canvas.findByRole('option', {
        hidden: true,
        name: 'Single With Required Status',
      }),
    ).toBeInTheDocument()
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
    // Select the multi-field Dest Type
    await userEvent.click(await canvas.findByLabelText('Destination Type'))
    await userEvent.click(
      await canvas.findByRole('option', { hidden: true, name: 'Multi Field' }),
    )

    await waitFor(async function Labels() {
      await expect(await canvas.findByLabelText('Name')).toBeVisible()
      await expect(
        await canvas.findByLabelText('Destination Type'),
      ).toBeVisible()
      await expect(await canvas.findByLabelText('First Item')).toBeVisible()
      await expect(await canvas.findByLabelText('Second Item')).toBeVisible()
      await expect(await canvas.findByLabelText('Third Item')).toBeVisible()
    })
  },
}

export const StatusUpdates: Story = {
  args: {
    userID: defaultConfig.user.id,
    title: 'Create New Contact Method',
    subtitle: 'Create New Contact Method Subtitle',
  },
  play: async ({ canvasElement }) => {
    const canvas = within(canvasElement)
    // Open option select
    await userEvent.click(await canvas.findByLabelText('Destination Type'))
    await userEvent.click(
      await canvas.findByRole('option', { hidden: true, name: 'Single Field' }),
    )
    await expect(
      await canvas.findByLabelText(
        'Send alert status updates (not supported for this type)',
      ),
    ).toBeDisabled()

    await userEvent.click(await canvas.findByLabelText('Destination Type'))
    await userEvent.click(
      await canvas.findByRole('option', {
        hidden: true,
        name: 'Single With Status',
      }),
    )
    await expect(
      await canvas.findByLabelText('Send alert status updates'),
    ).not.toBeDisabled()

    await userEvent.click(await canvas.findByLabelText('Destination Type'))
    await userEvent.click(
      await canvas.findByRole('option', {
        hidden: true,
        name: 'Single With Required Status',
      }),
    )
    await expect(
      await canvas.findByLabelText(
        'Send alert status updates (cannot be disabled for this type)',
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

    // response should set error on all fields plus the generic error
    await waitFor(async () => {
      await expect(await canvas.findByLabelText('Name')).toBeInvalid()

      await expect(await canvas.findByText('Name error')).toBeVisible()

      await expect(
        // mui puts aria-invalid on the input, but not the combobox (which the label points to)
        canvasElement.querySelector('input[name="dest.type"]'),
      ).toBeInvalid()
      await expect(await canvas.findByLabelText('Phone Number')).toBeInvalid()
      await expect(
        await canvas.findByText('This is a dest field-error'),
      ).toBeVisible()

      await expect(
        await canvas.findByText('This is a generic error'),
      ).toBeVisible()
    })
  },
}
