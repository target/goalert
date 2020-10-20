import { gql, useQuery, useMutation } from '@apollo/client'
import React, { useState } from 'react'

import p from 'prop-types'

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
  const { data, ...dataStatus } = useQuery(query, {
    variables: { id: serviceID },
  })
  const [save, saveStatus] = useMutation(mutation, {
    variables: { input: { ...value, id: serviceID } },
    onCompleted: onClose,
  })

  const defaults = {
    // default value is the service name & description with the ep.id
    ..._.chain(data).get('service').pick(['name', 'description']).value(),
    escalationPolicyID: _.get(data, 'service.ep.id'),
  }

  const fieldErrs = fieldErrors(saveStatus.error)

  return (
    <FormDialog
      title='Edit Service'
      loading={saveStatus.loading || (!data && dataStatus.loading)}
      errors={nonFieldErrors(saveStatus.error).concat(
        nonFieldErrors(dataStatus.error),
      )}
      onClose={onClose}
      onSubmit={() => save()}
      form={
        <ServiceForm
          epRequired
          errors={fieldErrs}
          disabled={Boolean(
            saveStatus.loading ||
              (!data && dataStatus.loading) ||
              dataStatus.error,
          )}
          value={value || defaults}
          onChange={(value) => setValue(value)}
        />
      }
    />
  )
}
ServiceEditDialog.propTypes = {
  serviceID: p.string.isRequired,
  onClose: p.func,
}
