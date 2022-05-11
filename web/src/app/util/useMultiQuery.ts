import { DocumentNode } from 'graphql'
import { useQuery, UseQueryArgs, UseQueryResponse } from 'urql'
import { mergeFields, prefixQuery } from './graphql'
import { print } from 'graphql/language/printer'
import _ from 'lodash'

interface MultiQueryHookOptions extends UseQueryArgs {
  query: DocumentNode
  variables: Record<string, unknown>[]
}
interface MultiQueryResult extends UseQueryResponse {
  // matching type
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  data?: any[]
}

const queryCache: Record<string, DocumentNode> = {}

export default function useMultiQuery(
  options: MultiQueryHookOptions,
): MultiQueryResult {
  let variables: Record<string, unknown> = {}
  let multiQuery: DocumentNode = null as unknown as DocumentNode

  // TODO: for cache-first, try cache-only query before joining

  options.variables.forEach((vars, i) => {
    variables = {
      ...variables,
      ..._.mapKeys(vars, (val, key) => `q${i}_${key}`),
    }
    multiQuery = mergeFields(multiQuery, prefixQuery(options.query, `q${i}_`))
  })

  let pause = options.pause || false
  if (multiQuery) {
    const queryKey = JSON.stringify(multiQuery)
    if (queryCache[queryKey]) multiQuery = queryCache[queryKey]
    else queryCache[queryKey] = multiQuery
  } else {
    // no variables passed, nothing to do
    multiQuery = options.query
    pause = true
  }

  const [{ data, ...resp }, refetch] = useQuery({
    ...options,
    query: print(multiQuery),
    variables,
    pause,
  })

  if (data) {
    const newData = options.variables.map((vars, i) => {
      const prefix = `q${i}_`

      return _.mapKeys(
        _.pickBy(data, (val, key) => key.startsWith(prefix)),
        (val, key) => key.substr(prefix.length),
      )
    })

    return [
      {
        ...resp,
        data: newData,
      },
      refetch,
    ]
  }

  return [{ ...resp, data: undefined }, refetch]
}
