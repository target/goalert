import React, { ReactElement, useMemo, useState } from 'react'
import { useSelector } from 'react-redux'
import { useQuery } from '@apollo/react-hooks'
import { Grid } from '@material-ui/core'
import { once } from 'lodash-es'
import { PaginatedList, PaginatedListItemProps } from './PaginatedList'
import { ITEMS_PER_PAGE, POLL_INTERVAL } from '../config'
import { searchSelector, urlKeySelector } from '../selectors'
import { fieldAlias } from '../util/graphql'
import { GraphQLClientWithErrors } from '../apollo'
import ListControls, { CheckboxActions } from './ListControls'

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

  // TODO: Clean up pass through props to ListControls
  // filter results, rendered to
  // the left of the search text field
  filter?: ReactElement

  // disables rendering list controls component with URL controlled search
  noSearch?: boolean

  // filters additional to search, set in the search text field
  searchAdornment?: ReactElement

  // renders list controls component with checkbox actions
  // NOTE: this will replace any icons set on each item with a checkbox
  actions?: CheckboxActions[]
}) {
  const {
    mapDataNode = (n: ObjectMap) => ({
      title: n.name,
      url: n.id,
      subText: n.description,
    }),
    query,
    filter,
    searchAdornment,
    variables = {},
    actions,
    noSearch,
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

  return (
    <Grid container spacing={2}>
      {/* Such that filtering/searching isn't re-rendered with the page content */}
      {(actions || !noSearch) && (
        <ListControls
          actions={actions}
          filter={filter}
          itemIDs={items.map((i: any) => i.id)}
          withSearch={!noSearch}
        />
      )}

      <Grid item xs={12}>
        <PaginatedList
          {...listProps}
          key={urlKey}
          items={items}
          itemsPerPage={queryVariables.input.first}
          loadMore={loadMore}
          isLoading={loading}
          withCheckboxes={Boolean(actions)}
        />
      </Grid>
    </Grid>
  )
}
