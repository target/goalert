import React, { useMemo } from 'react'
import { useSelector } from 'react-redux'
import { useQuery } from '@apollo/react-hooks'
import { Grid } from '@material-ui/core'
import { once } from 'lodash-es'
import { PaginatedList, PaginatedListItemProps } from './PaginatedList'
import { ITEMS_PER_PAGE, POLL_INTERVAL } from '../config'
import { searchSelector, urlKeySelector } from '../selectors'
import { fieldAlias } from '../util/graphql'
import { GraphQLClientWithErrors } from '../apollo'
import ControlledPaginatedList, {
  ControlledPaginatedListProps,
} from './ControlledPaginatedList'
import { QueryResult } from '@apollo/react-common'

// any && object type map
// used for objects with unknown key/values from parent
interface ObjectMap {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  [key: string]: any
}

const buildFetchMore = (
  fetchMore: QueryResult['fetchMore'],
  after: string,
  stopPolling: QueryResult['stopPolling'],
  itemsPerPage: number,
): ((numberToLoad?: number) => void) | undefined => {
  return once((newLimit?: number) => {
    stopPolling()
    return fetchMore({
      variables: {
        input: {
          first: newLimit || itemsPerPage,
          after,
        },
      },
      updateQuery: (prev: ObjectMap, { fetchMoreResult }: ObjectMap) => {
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

export interface QueryListProps extends ControlledPaginatedListProps {
  /*
   * query must provide a single field that returns nodes
   *
   * For example:
   *   ```graphql
   *   query Services {
   *     services {
   *       nodes {
   *         id
   *         name
   *         description
   *       }
   *     }
   *   }
   *  ```
   */
  query: object

  // mapDataNode should map the struct from each node in `nodes` to the struct required by a PaginatedList item
  mapDataNode?: (n: ObjectMap) => PaginatedListItemProps

  // variables will be added to the initial query. Useful for things like `favoritesFirst` or alert filters
  // note: The `input.search` and `input.first` parameters are included by default, but can be overridden
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  variables?: any
  mapVariables?: (vars: any) => any
}

export default function QueryList(props: QueryListProps): JSX.Element {
  const {
    mapDataNode = (n: ObjectMap) => ({
      id: n.id,
      title: n.name,
      url: n.id,
      subText: n.description,
    }),
    query,
    variables = {},
    noSearch,
    mapVariables = (v) => v,
    ...listProps
  } = props
  const { input, ...vars } = variables

  const searchParam = useSelector(searchSelector)
  const urlKey = useSelector(urlKeySelector)
  const aliasedQuery = useMemo(() => fieldAlias(query, 'data'), [query])

  const queryVariables = {
    ...vars,
    input: {
      first: ITEMS_PER_PAGE,
      search: searchParam,
      ...input,
    },
  }

  if (noSearch) {
    delete queryVariables.input.search
  }

  const { data, loading, fetchMore, stopPolling } = useQuery(aliasedQuery, {
    client: GraphQLClientWithErrors,
    variables: mapVariables(queryVariables),
    fetchPolicy: 'network-only',
    pollInterval: POLL_INTERVAL,
  })

  const nodes = data?.data?.nodes ?? []
  const items = nodes.map(mapDataNode)
  let loadMore: ((numberToLoad?: number) => void) | undefined

  if (data?.data?.pageInfo?.hasNextPage) {
    loadMore = buildFetchMore(
      fetchMore,
      data.data.pageInfo.endCursor,
      stopPolling,
      queryVariables.input.first,
    )
  }

  if (Boolean(props.checkboxActions) || !props.noSearch) {
    return (
      <Grid container spacing={2}>
        <ControlledPaginatedList
          {...listProps}
          items={items}
          itemsPerPage={queryVariables.input.first}
          loadMore={loadMore}
          isLoading={!data && loading}
        />
      </Grid>
    )
  }

  return (
    <Grid container spacing={2}>
      <Grid item xs={12}>
        <PaginatedList
          {...listProps}
          key={urlKey}
          items={items}
          itemsPerPage={queryVariables.input.first}
          loadMore={loadMore}
          isLoading={!data && loading}
        />
      </Grid>
    </Grid>
  )
}
