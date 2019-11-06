import gql from 'graphql-tag'
import { useMutation } from 'react-apollo'
import { fieldAlias, mergeFields, mapInputVars } from '../../util/graphql'

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

const useCreateAlerts = formFields => {
  // 1. build mutation
  let m = getAliasedMutation(baseMutation, 0)
  for (let i = 1; i < formFields.selectedServices.length; i++) {
    m = mergeFields(m, getAliasedMutation(baseMutation, i))
  }

  // 2. build variables
  let variables = {}
  formFields.selectedServices.forEach((ss, i) => {
    variables[`input${i}`] = {
      summary: formFields.Summary,
      details: formFields.Details,
      serviceID: ss,
    }
  })

  // 3. execute mutation with variables
  return useMutation(m, {
    variables,
    skip: formFields.selectedServices.length === 0,
  })
}

export { useCreateAlerts as default }
