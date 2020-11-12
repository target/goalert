import { gql, useMutation } from '@apollo/client'
import React, { useState } from 'react'

import p from 'prop-types'

import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceLabelForm from './ServiceLabelForm'

const mutation = gql`
  mutation($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`
const query = gql`
  query($serviceID: ID!) {
    service(id: $serviceID) {
      id
      labels {
        key
        value
      }
    }
  }
`

export default function ServiceLabelCreateDialog(props) {
  const [value, setValue] = useState({ key: '', value: '' })

  const [createLabel, createLabelStatus] = useMutation(mutation, {
    onCompleted: props.onClose,
    update: (cache) => {
      const { service } = cache.readQuery({
        query,
        variables: { serviceID: props.serviceID },
      })
      const labels = (service.labels || []).filter((l) => l.key !== value.key)
      if (value.value) {
        labels.push({ ...value, __typename: 'Label' })
      }
      cache.writeQuery({
        query,
        variables: { serviceID: props.serviceID },
        data: {
          service: {
            ...service,
            labels,
          },
        },
      })
    },
  })

  return (
    <FormDialog
      title='Set Label Value'
      loading={createLabelStatus.loading}
      errors={nonFieldErrors(createLabelStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        return createLabel({
          variables: {
            input: {
              ...value,
              target: { type: 'service', id: props.serviceID },
            },
          },
        })
      }}
      form={
        <ServiceLabelForm
          errors={fieldErrors}
          disabled={createLabelStatus.loading}
          value={value}
          onChange={(val) => setValue(val)}
        />
      }
    />
  )
}

ServiceLabelCreateDialog.propTypes = {
  serviceID: p.string.isRequired,
  onClose: p.func,
}
