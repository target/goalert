import { DocumentNode } from 'graphql'
import {
  OperationVariables,
  QueryHookOptions,
  QueryResult,
  useQuery,
} from 'react-apollo'
import { mergeFields, prefixQuery } from './graphql'
import _ from 'lodash-es'

interface MultiQueryHookOptions extends QueryHookOptions {
  variables: OperationVariables[]
}
interface MultiQueryResult extends QueryResult {
  data: any[] | undefined
}

export default function useMultiQuery(
  query: DocumentNode,
  options: MultiQueryHookOptions,
): MultiQueryResult {
  let variables: OperationVariables = {}
  let multiQuery: DocumentNode = (null as unknown) as DocumentNode

  options.variables.forEach((vars, i) => {
    variables = { ...variables, ..._.mapKeys(vars, (key) => `q${i}_${key}`) }
    multiQuery = mergeFields(multiQuery, prefixQuery(query, `q${i}_`))
  })

  const { data, ...resp } = useQuery(multiQuery, { ...options, variables })

  if (data) {
    const newData = options.variables.map((vars, i) =>
      _.pickBy(data, (val, key) => key.startsWith(`q${i}_`)),
    )
    return {
      ...resp,
      data: newData,
    }
  }

  return { ...resp, data: undefined }
}
