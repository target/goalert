import React, { useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'

import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceForm from './ServiceForm'
import Spinner from '../loading/components/Spinner'
import { Label } from '../../schema'

interface Value {
  name: string
  description: string
  escalationPolicyID?: string
  labels: Label[]
}

const query = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
      description
      labels {
        key
        value
      }
      ep: escalationPolicy {
        id
        name
      }
    }
  }
`
const mutation = gql`
  mutation updateService($input: UpdateServiceInput!) {
    updateService(input: $input)
  }
`
const setLabel = gql`
  mutation setLabel($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`

export default function ServiceEditDialog(props: {
  serviceID: string
  onClose: () => void
}): JSX.Element {
  const [value, setValue] = useState<Value | null>(null)
  const [{ data, fetching: dataFetching, error: dataError }] = useQuery({
    query,
    variables: { id: props.serviceID },
  })

  const [saveStatus, save] = useMutation(mutation)
  const [saveLabelStatus, saveLabel] = useMutation(setLabel)

  if (dataFetching && !data) {
    return <Spinner />
  }

  const defaultValue = {
    name: data?.service?.name,
    description: data?.service?.description,
    escalationPolicyID: data?.service?.ep?.id,
    labels: data?.service?.labels || [],
  }

  const fieldErrs = fieldErrors(saveStatus.error).concat(
    fieldErrors(saveLabelStatus.error),
  )

  return (
    <FormDialog
      title='Edit Service'
      loading={saveStatus.fetching || (!data && dataFetching)}
      errors={nonFieldErrors(saveStatus.error).concat(
        nonFieldErrors(dataError),
        nonFieldErrors(saveLabelStatus.error),
      )}
      onClose={props.onClose}
      onSubmit={async () => {
        const saveRes = await save(
          {
            input: {
              id: props.serviceID,
              name: value?.name || '',
              description: value?.description || '',
              escalationPolicyID: value?.escalationPolicyID || '',
            },
          },
          {
            additionalTypenames: ['Service'],
          },
        )
        if (saveRes.error) return

        for (const label of value?.labels || []) {
          const res = await saveLabel({
            input: {
              target: {
                type: 'service',
                id: props.serviceID,
              },
              key: label.key,
              value: label.value,
            },
          })
          if (res.error) return
        }

        props.onClose()
      }}
      form={
        <ServiceForm
          epRequired
          errors={fieldErrs}
          disabled={Boolean(
            saveStatus.fetching || (!data && dataFetching) || dataError,
          )}
          value={value || defaultValue}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
