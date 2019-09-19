import React, { useState } from 'react'
import p from 'prop-types'

import gql from 'graphql-tag'
import { useQuery, useMutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'

import FormDialog from '../dialogs/FormDialog'
import ServiceForm from './ServiceForm'
import _ from 'lodash-es'

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

export default function ServiceEditDialog({ serviceID, onClose }) {
  const [value, setValue] = useState(null)
  const { data, loading: dataLoading, error: dataError } = useQuery(query, {
    variables: { id: serviceID },
  })
  const [save, { loading, error }] = useMutation(mutation, {
    variables: { input: { ...value, id: serviceID } },
    onCompleted: onClose,
  })
  const defaults =
    data && data.service && data.service.id
      ? {
          ..._.pick(data.service, ['name', 'description']),
          escalationPolicyID: data.service.ep.id,
        }
      : {}

  const fieldErrs = fieldErrors(error)

  return (
    <FormDialog
      title='Edit Service'
      loading={loading}
      errors={nonFieldErrors(error).concat(nonFieldErrors(dataError))}
      onClose={onClose}
      onSubmit={() => save()}
      form={
        <ServiceForm
          epRequired
          errors={fieldErrs}
          disabled={loading || dataLoading || dataError}
          value={value || defaults}
          onChange={value => setValue(value)}
        />
      }
    />
  )
}
ServiceEditDialog.propTypes = {
  serviceID: p.string.isRequired,
  onClose: p.func,
}
