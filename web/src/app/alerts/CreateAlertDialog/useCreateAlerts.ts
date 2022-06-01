import {
  DocumentNode,
  gql,
  MutationFunction,
  MutationResult,
  useMutation,
} from '@apollo/client'
import { fieldAlias, mergeFields, mapInputVars } from '../../util/graphql'
import { GraphQLClientWithErrors } from '../../apollo'

interface Variable {
  summary: string
  details: string
  serviceID: string
}

const baseMutation = gql`
  mutation CreateAlertMutation($input: CreateAlertInput!) {
    createAlert(input: $input) {
      id
    }
  }
`

const getAliasedMutation = (
  mutation: DocumentNode,
  index: string | number,
): DocumentNode =>
  mapInputVars(fieldAlias(mutation, 'alias' + index), {
    input: 'input' + index,
  })

// useCreateAlerts will return mutation, status and a function for mapping
// field/paths from the response to the respective service ID.
export const useCreateAlerts = (value: {
  serviceIDs: string[]
  summary: string
  details: string
}): [MutationFunction, MutationResult, (alias: string | number) => string] => {
  // 1. build mutation
  let m = getAliasedMutation(baseMutation, 0)
  for (let i = 1; i < value.serviceIDs.length; i++) {
    m = mergeFields(m, getAliasedMutation(baseMutation, i))
  }

  // 2. build variables, alias -> service ID map
  const variables: { [key: string]: Variable } = {}
  const aliasIDMap: { [key: string]: string } = {}
  value.serviceIDs.forEach((svcID, i) => {
    aliasIDMap['alias' + i] = svcID
    variables[`input${i}`] = {
      summary: value.summary.trim(),
      details: value.details.trim(),
      serviceID: svcID,
    }
  })

  // 3. build mutation with variables
  const [mutate, status] = useMutation(m, {
    variables,
    client: GraphQLClientWithErrors,
  })

  return [mutate, status, (alias: string | number) => aliasIDMap[alias]]
}
