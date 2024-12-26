import React, { useState } from 'react'
import { gql, useMutation, CombinedError } from 'urql'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import FormDialog from '../dialogs/FormDialog'
import ServiceForm, { Value } from './ServiceForm'
import { Redirect } from 'wouter'
import { Label } from '../../schema'

interface InputVar {
  name: string
  description: string
  escalationPolicyID?: string
  favorite: boolean
  labels: Label[]
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
  { name, description, escalationPolicyID, labels }: Value,
  attempt = 0,
): InputVar {
  const vars: InputVar = {
    name,
    description,
    escalationPolicyID,
    favorite: true,
    labels,
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
}): React.JSX.Element {
  const [value, setValue] = useState<Value>({
    name: '',
    description: '',
    escalationPolicyID: '',
    labels: [],
  })

  const [createKeyStatus, commit] = useMutation(createMutation)

  const { data, error } = createKeyStatus
  if (data && data.createService) {
    return <Redirect to={`/services/${data.createService.id}`} />
  }

  const fieldErrs = fieldErrors(error).filter(
    (e) => !e.field.startsWith('newEscalationPolicy.'),
  )

  return (
    <FormDialog
      title='Create New Service'
      errors={nonFieldErrors(error)}
      onClose={props.onClose}
      onSubmit={() => {
        let n = 1
        const onErr = (err: CombinedError): Awaited<Promise<unknown>> => {
          // retry if it's a policy name conflict
          if (
            err.graphQLErrors &&
            err.graphQLErrors[0].extensions &&
            err.graphQLErrors[0].extensions.isFieldError &&
            err.graphQLErrors[0].extensions.fieldName ===
              'newEscalationPolicy.Name'
          ) {
            n++
            return commit({
              variables: {
                input: inputVars(value, n),
              },
            }).then(null, onErr)
          }
        }

        return commit(
          {
            input: inputVars(value),
          },
          {
            additionalTypenames: ['ServiceConnection'],
          },
        ).then(null, onErr)
      }}
      form={
        <ServiceForm
          errors={fieldErrs}
          value={value}
          onChange={(val: Value) => setValue(val)}
        />
      }
    />
  )
}
