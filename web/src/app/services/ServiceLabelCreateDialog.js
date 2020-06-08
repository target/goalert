import React, { useState } from 'react'
import p from 'prop-types'
import { Redirect } from 'react-router-dom'

import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
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
      cache.writeData({
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

  const { loading, data, error } = createLabelStatus
  if (data && data.createLabel) {
    return <Redirect push to={`/services/${data.createLabel.id}`} />
  }

  return (
    <FormDialog
      title='Set Label Value'
      loading={loading}
      errors={nonFieldErrors(error)}
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
          disabled={loading}
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
