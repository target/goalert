import React, { ReactElement, useEffect, useMemo } from 'react'
import { useSelector } from 'react-redux'
import { useQuery } from '@apollo/react-hooks'
import { Grid, makeStyles } from '@material-ui/core'
import { once } from 'lodash-es'
import { PaginatedList, PaginatedListItemProps } from './PaginatedList'
import { ITEMS_PER_PAGE, POLL_INTERVAL } from '../config'
import { searchSelector, urlKeySelector } from '../selectors'
import { fieldAlias } from '../util/graphql'
import Search from '../util/Search'
import { GraphQLClientWithErrors } from '../apollo'

const useStyles = makeStyles({
  search: {
    paddingLeft: '0.5em',
  },
})

// any && object type map
// used for objects with unknown key/values from parent
interface ObjectMap {
  [key: string]: any
}

const buildFetchMore = (
  fetchMore: Function,
  after: string,
  stopPolling: Function,
  itemsPerPage: number,
) => {
  return once(() => {
    stopPolling()
    return fetchMore({
      variables: {
        input: {
          first: itemsPerPage,
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

export default function QueryList(props: {
  // query must provide a single field that returns nodes
  //
  // For example:
  // ```graphql
  // query Services {
  //   services {
  //     nodes {
  //       id
  //       name
  //       description
  //     }
  //   }
  // }
  // ```
  query: object

  // mapDataNode should map the struct from each node in `nodes` to the struct required by a PaginatedList item
  mapDataNode?: (n: ObjectMap) => PaginatedListItemProps

  // variables will be added to the initial query. Useful for things like `favoritesFirst` or alert filters
  // note: The `input.search` and `input.first` parameters are included by default, but can be overridden
  variables?: any

  // if set, the search string param is ignored
  noSearch?: boolean

  // controls unrelated to search, but still modify results, rendered to
  // the left of the search text field
  controls: ReactElement

  // filters additional to search, set in the search text field
  searchAdornment?: ReactElement

  // invoked when the amount of items queried changes
  onDataChange: (nodes: Array<ObjectMap>) => void
}) {
  const { query, mapDataNode = (n: ObjectMap) => ({
    title: n.name,
    url: n.id,
    subText: n.description,
  }), noSearch, controls, searchAdornment, onDataChange, variables = {}, ...listProps } = props
  const { input, ...vars } = variables

  const classes = useStyles()
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
    variables: queryVariables,
    fetchPolicy: 'network-only',
    pollInterval: POLL_INTERVAL,
  })

  const nodes = data?.data?.nodes ?? []
  const items = nodes.map(mapDataNode)
  let loadMore

  if (data?.data?.pageInfo?.hasNextPage) {
    loadMore = buildFetchMore(
      fetchMore,
      data.data.pageInfo.endCursor,
      stopPolling,
      queryVariables.input.first,
    )
  }

  useEffect(() => {
    if (onDataChange) {
      onDataChange(nodes)
    }
  }, [nodes.length])

  return (
    <Grid container spacing={2}>
      {/* Such that filtering/searching isn't re-rendered with the page content */}
      <Grid container item xs={12} justify='flex-end' alignItems='center'>
        {controls}
        <Grid item className={classes.search}>
          <Search endAdornment={searchAdornment} />
        </Grid>
      </Grid>

      <Grid item xs={12}>
        <PaginatedList
          {...listProps}
          key={urlKey}
          items={items}
          itemsPerPage={queryVariables.input.first}
          loadMore={loadMore}
          isLoading={loading}
        />
      </Grid>
    </Grid>
  )
}
