import React, { useState } from 'react'
import { gql } from '@apollo/client'

import p from 'prop-types'
import { Mutation } from '@apollo/client/react/components'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Query from '../util/Query'

import FormDialog from '../dialogs/FormDialog'
import ServiceLabelForm from './ServiceLabelForm'

const mutation = gql`
  mutation ($input: SetLabelInput!) {
    setLabel(input: $input)
  }
`
const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      id
      labels {
        key
        value
      }
    }
  }
`

export default function ServiceLabelEditDialog(props) {
  const { onClose, labelKey, serviceID } = props
  const [value, setValue] = useState(null)

  function renderDialog(data, commit, status) {
    const { loading, error } = status
    return (
      <FormDialog
        title='Update Label Value'
        loading={loading}
        errors={nonFieldErrors(error)}
        onClose={onClose}
        onSubmit={() => {
          if (!value) {
            return onClose()
          }
          return commit({
            variables: {
              input: {
                ...value,
                target: { type: 'service', id: serviceID },
              },
            },
          })
        }}
        form={
          <ServiceLabelForm
            errors={fieldErrors(error)}
            editValueOnly
            disabled={loading}
            value={value || { key: data.key, value: data.value }}
            onChange={(value) => setValue(value)}
          />
        }
      />
    )
  }

  function renderMutation(data) {
    return (
      <Mutation
        mutation={mutation}
        onCompleted={onClose}
        update={(cache) => {
          const { service } = cache.readQuery({
            query,
            variables: { serviceID },
          })
          const labels = (service.labels || []).filter(
            (l) => l.key !== value.key,
          )
          if (value.value) {
            labels.push({ ...value, __typename: 'Label' })
          }
          cache.writeQuery({
            query,
            variables: { serviceID },
            data: {
              service: {
                ...service,
                labels,
              },
            },
          })
        }}
      >
        {(commit, status) => renderDialog(data, commit, status)}
      </Mutation>
    )
  }

  function renderQuery() {
    return (
      <Query
        noPoll
        query={query}
        variables={{ serviceID }}
        render={({ data }) =>
          renderMutation(data.service.labels.find((l) => l.key === labelKey))
        }
      />
    )
  }

  return renderQuery()
}

ServiceLabelEditDialog.propTypes = {
  serviceID: p.string.isRequired,
  labelKey: p.string.isRequired,
  onClose: p.func,
}
