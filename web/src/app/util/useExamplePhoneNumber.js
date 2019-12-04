import gql from 'graphql-tag'
import { useQuery } from 'react-apollo'

const query = gql`
  query($input: String!) {
    examplePhoneNumber(input: $input)
  }
`

export default function useExamplePhoneNumber(countryCode) {
  const { data } = useQuery(query, {
    variables: {
      input: countryCode,
    },
  })

  return data && data.examplePhoneNumber
}
