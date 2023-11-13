import React, { useState } from 'react'
import { ApolloError, gql, useMutation } from '@apollo/client'

import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceForm, { Value } from './ServiceForm'
import { Redirect } from 'wouter'

interface InputVar {
  name: string
  description: string
  escalationPolicyID?: string
  favorite: boolean
  newEscalationPolicy?: {
    name: string
    description: string
    favorite: boolean
    steps: { delayMinutes: number; targets: { type: string; id: string }[] }[]
  }
}

const createMutation = gql`
  mutation createService($input: CreateServiceInput!) {
    createService(input: $input) {
      id
      name
      description
      escalationPolicyID
    }
  }
`

function inputVars(
  { name, description, escalationPolicyID }: Value,
  attempt = 0,
): InputVar {
  const vars: InputVar = {
    name,
    description,
    escalationPolicyID,
    favorite: true,
  }
  if (!vars.escalationPolicyID) {
    vars.newEscalationPolicy = {
      name: attempt ? `${name} Policy ${attempt}` : name + ' Policy',
      description: 'Auto-generated policy for ' + name,
      favorite: true,
      steps: [
        {
          delayMinutes: 5,
          targets: [
            {
              type: 'user',
              id: '__current_user',
            },
          ],
        },
      ],
    }
  }

  return vars
}

export default function ServiceCreateDialog(props: {
  onClose: () => void
}): React.ReactNode {
  const [value, setValue] = useState<Value>({
    name: '',
    description: '',
    escalationPolicyID: '',
  })

  const [createKey, createKeyStatus] = useMutation(createMutation)

  const { loading, data, error } = createKeyStatus
  if (data && data.createService) {
    return <Redirect to={`/services/${data.createService.id}`} />
  }

  const fieldErrs = fieldErrors(error).filter(
    (e) => !e.field.startsWith('newEscalationPolicy.'),
  )

  return (
    <FormDialog
      title='Create New Service'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => {
        let n = 1
        const onErr = (err: ApolloError): Awaited<Promise<unknown>> => {
          // retry if it's a policy name conflict
          if (
            err.graphQLErrors &&
            err.graphQLErrors[0].extensions &&
            err.graphQLErrors[0].extensions.isFieldError &&
            err.graphQLErrors[0].extensions.fieldName ===
              'newEscalationPolicy.Name'
          ) {
            n++
            return createKey({
              variables: {
                input: inputVars(value, n),
              },
            }).then(null, onErr)
          }
        }

        return createKey({
          variables: {
            input: inputVars(value),
          },
        }).then(null, onErr)
      }}
      form={
        <ServiceForm
          errors={fieldErrs}
          disabled={loading}
          value={value}
          onChange={(val: Value) => setValue(val)}
        />
      }
    />
  )
}
