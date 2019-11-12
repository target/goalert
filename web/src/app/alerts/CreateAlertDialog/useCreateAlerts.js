import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
import { fieldAlias, mergeFields, mapInputVars } from '../../util/graphql'
import { GraphQLClientWithErrors } from '../../apollo'

const baseMutation = gql`
  mutation CreateAlertMutation($input: CreateAlertInput!) {
    createAlert(input: $input) {
      id
    }
  }
`

const getAliasedMutation = (mutation, index) =>
  mapInputVars(fieldAlias(mutation, 'alias' + index), {
    input: 'input' + index,
  })

// useCreateAlerts will return mutation, status and a function for mapping
// field/paths from the response to the respecitve service ID.
export const useCreateAlerts = value => {
  // 1. build mutation
  let m = getAliasedMutation(baseMutation, 0)
  for (let i = 1; i < value.serviceIDs.length; i++) {
    m = mergeFields(m, getAliasedMutation(baseMutation, i))
  }

  // 2. build variables
  let variables = {}
  const aliasIDMap = {}
  value.serviceIDs.forEach((ss, i) => {
    aliasIDMap['alias' + i] = ss
    variables[`input${i}`] = {
      summary: value.summary.trim(),
      details: value.details.trim(),
      serviceID: ss,
    }
  })

  // 3. build mutation with variables
  const [mutate, status] = useMutation(m, {
    variables,
    client: GraphQLClientWithErrors,
  })

  return [mutate, status, alias => aliasIDMap[alias]]
}
