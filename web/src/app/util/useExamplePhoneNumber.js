import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'

const query = gql`
  query($input: String!) {
    examplePhoneNumber(input: $input)
  }
`

export default function useExamplePhoneNumber(countryCode) {
  return useQuery(query, {
    variables: {
      input: countryCode,
    },
  })
}
