import React, { useState } from 'react'
import { gql, useQuery, useMutation } from 'urql'

import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceForm from './ServiceForm'
import Spinner from '../loading/components/Spinner'

interface Value {
  name: string
  description: string
  escalationPolicyID?: string
}

const query = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
      description
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

export default function ServiceEditDialog(props: {
  serviceID: string
  onClose: () => void
}): React.ReactNode {
  const [value, setValue] = useState<Value | null>(null)
  const [{ data, fetching: dataFetching, error: dataError }] = useQuery({
    query,
    variables: { id: props.serviceID },
  })

  const [saveStatus, save] = useMutation(mutation)

  if (dataFetching && !data) {
    return <Spinner />
  }

  const defaultValue = {
    name: data?.service?.name,
    description: data?.service?.description,
    escalationPolicyID: data?.service?.ep?.id,
  }

  const fieldErrs = fieldErrors(saveStatus.error)

  return (
    <FormDialog
      title='Edit Service'
      loading={saveStatus.fetching || (!data && dataFetching)}
      errors={nonFieldErrors(saveStatus.error).concat(
        nonFieldErrors(dataError),
      )}
      onClose={props.onClose}
      onSubmit={() => {
        save(
          {
            input: {
              ...value,
              id: props.serviceID,
            },
          },
          {
            additionalTypenames: ['Service'],
          },
        ).then((res) => {
          if (res.error) return
          props.onClose()
        })
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
