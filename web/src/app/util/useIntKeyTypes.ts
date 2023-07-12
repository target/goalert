import { gql, useQuery } from 'urql'
import { IntegrationKeyTypeInfo } from '../../schema'

const query = gql`
  query getIntKeyTypes {
    integrationKeyTypes {
      id
      name
      label
      enabled
    }
  }
`

export function useIntKeyTypes(): Array<IntegrationKeyTypeInfo> {
  const [{ data, fetching, error }] = useQuery({
    query,
    requestPolicy: 'cache-first',
  })
  if (fetching) return []
  if (error) return []

  return data.integrationKeyTypes || []
}
