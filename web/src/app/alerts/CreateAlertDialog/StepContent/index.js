import React from 'react'
import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'
import _ from 'lodash-es'
import Step0 from './Step0'
import Step1 from './Step1'
import Step2 from './Step2'
import Step3 from './Step3'

const query = gql`
  query($input: ServiceSearchOptions) {
    services(input: $input) {
      nodes {
        id
        name
        isFavorite
      }
    }
  }
`

export default props => {
  const { formFields, mutationStatus, onChange } = props

  // TODO loading error handles
  const { data } = useQuery(query, {
    variables: {
      input: {
        search: formFields.searchQuery,
        favoritesFirst: true,
        omit: formFields.selectedServices,
        favoritesOnly: formFields.searchQuery.length === 0,
      },
    },
  })

  const queriedServices = _.get(data, 'services.nodes', [])

  switch (props.activeStep) {
    case 0:
      return <Step0 />
    case 1:
      return (
        <Step1
          formFields={formFields}
          onChange={onChange}
          queriedServices={queriedServices}
        />
      )
    case 2:
      return <Step2 formFields={formFields} />
    case 3:
      return <Step3 formFields={formFields} mutationStatus={mutationStatus} />
    default:
      return 'Unknown step'
  }
}
