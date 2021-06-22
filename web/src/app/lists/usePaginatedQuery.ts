/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable @typescript-eslint/explicit-function-return-type */
/* eslint-disable @typescript-eslint/explicit-module-boundary-types */
import {
  BaseQueryOptions,
  DocumentNode,
  QueryResult,
  useQuery,
} from '@apollo/client'
import _ from 'lodash'
import { useMemo } from 'react'
import { useURLParam } from '../actions'
import { GraphQLClientWithErrors } from '../apollo'
import { ITEMS_PER_PAGE, POLL_INTERVAL } from '../config'
import { fieldAlias } from '../util/graphql'

const makeLoadMore = (
  fetchMore: QueryResult['fetchMore'],
  after: string,
  stopPolling: QueryResult['stopPolling'],
  itemsPerPage: number,
): ((numberToLoad?: number) => void) | undefined => {
  return _.once((newLimit?: number) => {
    stopPolling()
    return fetchMore({
      variables: {
        input: {
          first: newLimit || itemsPerPage,
          after,
        },
      },
      updateQuery: (
        prev: Record<string, any>,
        { fetchMoreResult }: Record<string, any>,
      ) => {
        if (!fetchMoreResult) return prev

        return {
          ...fetchMoreResult,
          data: {
            ...fetchMoreResult.data,
            nodes: prev.data.nodes.concat(fetchMoreResult.data.nodes),
          },
        }
      },
    })
  })
}

export function usePaginatedQuery(
  query: DocumentNode,
  variables: BaseQueryOptions['variables'] = {},
  searchable = true,
) {
  const [search] = useURLParam('search', '')
  const { input, ...vars } = variables

  const qvars = {
    ...vars,
    input: {
      first: ITEMS_PER_PAGE,
      search, // uses URL param by default, override by passing in variables.input.search
      ...input,
    },
  }

  if (!searchable) {
    delete qvars.input.search
  }

  const aliasedQuery = useMemo(() => fieldAlias(query, 'data'), [query])
  const q = useQuery(aliasedQuery, {
    client: GraphQLClientWithErrors,
    variables: qvars,
    fetchPolicy: 'network-only',
    pollInterval: POLL_INTERVAL,
  })

  let loadMore
  if (q.data?.data?.pageInfo?.hasNextPage) {
    loadMore = makeLoadMore(
      q.fetchMore,
      q.data?.data?.pageInfo?.endCursor,
      q.stopPolling,
      qvars.input.first,
    )
  }

  return { q, loadMore }
}
