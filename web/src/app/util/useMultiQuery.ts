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

const queryCache: Record<string, DocumentNode> = {}

export default function useMultiQuery(
  query: DocumentNode,
  options: MultiQueryHookOptions,
): MultiQueryResult {
  let variables: OperationVariables = {}
  let multiQuery: DocumentNode = (null as unknown) as DocumentNode

  // TODO: for cache-first, try cache-only query before joining

  options.variables.forEach((vars, i) => {
    variables = {
      ...variables,
      ..._.mapKeys(vars, (val, key) => `q${i}_${key}`),
    }
    multiQuery = mergeFields(multiQuery, prefixQuery(query, `q${i}_`))
  })

  if (multiQuery) {
    const queryKey = JSON.stringify(multiQuery)
    if (queryCache[queryKey]) multiQuery = queryCache[queryKey]
    else queryCache[queryKey] = multiQuery
  } else {
    // no variables passed, nothing to do
    multiQuery = query
    variables.skip = true
  }

  const { data, ...resp } = useQuery(multiQuery, { ...options, variables })

  if (data) {
    const newData = options.variables.map((vars, i) => {
      const prefix = `q${i}_`

      return _.mapKeys(
        _.pickBy(data, (val, key) => key.startsWith(prefix)),
        (val, key) => key.substr(prefix.length),
      )
    })

    return {
      ...resp,
      data: newData,
    }
  }

  return { ...resp, data: undefined }
}
