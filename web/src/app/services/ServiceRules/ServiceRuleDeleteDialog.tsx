import React from 'react'
import { gql, useQuery, useMutation } from 'urql'

import { nonFieldErrors } from '../../util/errutil'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import FormDialog from '../../dialogs/FormDialog'

const query = gql`
  query ($id: ID!) {
    serviceRule(id: $id) {
      id
      name
      filters {
        field
        operator
        value
        valueType
      }
      sendAlert
      serviceID
    }
  }
`

const mutation = gql`
  mutation ($id: ID!) {
    deleteServiceRule(id: $id)
  }
`

export default function ServiceRuleDeleteDialog(props: {
  ruleID: string | null
  onClose: () => void
}): JSX.Element {
  const [{ fetching, error, data }] = useQuery({
    query,
    variables: { id: props.ruleID },
  })

  const [deleteRuleStatus, deleteRule] = useMutation(mutation)

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  if (
    !fetching &&
    !deleteRuleStatus.fetching &&
    data?.integrationKey === null
  ) {
    return (
      <FormDialog
        alert
        title='No longer exists'
        onClose={() => props.onClose()}
        subTitle='That service rule does not exist or is already deleted.'
      />
    )
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the service rule: ${data?.serviceRule?.name}`}
      loading={deleteRuleStatus.fetching}
      errors={nonFieldErrors(deleteRuleStatus.error)}
      onClose={props.onClose}
      onSubmit={() => {
        return deleteRule(
          { id: props.ruleID },
          { additionalTypenames: ['ServiceRule'] },
        ).then((res) => {
          if (res.error) return
          props.onClose()
        })
      }}
    />
  )
}
