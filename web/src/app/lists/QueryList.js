import React, { useMemo } from 'react'
import { useSelector } from 'react-redux'
import p from 'prop-types'
import { useQuery } from '@apollo/react-hooks'
import { Grid, makeStyles } from '@material-ui/core'
import { once } from 'lodash-es'
import { PaginatedList } from './PaginatedList'
import { ITEMS_PER_PAGE, POLL_INTERVAL } from '../config'
import { searchSelector, urlKeySelector } from '../selectors'
import { fieldAlias } from '../util/graphql'
import Search from '../util/Search'
import { GraphQLClientWithErrors } from '../apollo'

const useStyles = makeStyles({
  searchGridItem: {
    display: 'flex',
    justifyContent: 'flex-end',
  },
})

const buildFetchMore = (fetchMore, after, stopPolling) => {
  return once(num => {
    stopPolling()
    return fetchMore({
      variables: {
        input: {
          first: num,
          after,
        },
      },
      updateQuery: (prev, { fetchMoreResult }) => {
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

export default function QueryList(props) {
  const { noSearch, query, filter, searchAdornment, ...listProps } = props
  const { input, ...vars } = props.variables

  const classes = useStyles()
  const searchParam = useSelector(searchSelector)
  const urlKey = useSelector(urlKeySelector)
  const aliasedQuery = useMemo(() => fieldAlias(query, 'data'), [query])

  const variables = {
    ...vars,
    input: {
      first: ITEMS_PER_PAGE,
      search: searchParam,
      ...input,
    },
  }

  if (noSearch) {
    delete variables.input.search
  }

  const { data, loading, fetchMore, stopPolling } = useQuery(aliasedQuery, {
    client: GraphQLClientWithErrors,
    variables,
    fetchPolicy: 'network-only',
    pollInterval: POLL_INTERVAL,
  })

  let items = []
  let loadMore

  if (data && data.data && data.data.nodes) {
    items = data.data.nodes.map(props.mapDataNode)
    if (data.data.pageInfo.hasNextPage) {
      loadMore = buildFetchMore(
        fetchMore,
        data.data.pageInfo.endCursor,
        stopPolling,
      )
    }
  }

  return (
    <Grid container spacing={2}>
      {/* Such that filtering/searching isn't re-rendered with the page content */}
      <Grid item xs={12} className={classes.searchGridItem}>
        {filter}
        <Search endAdornment={searchAdornment} />
      </Grid>

      <Grid item xs={12}>
        <PaginatedList
          {...listProps}
          key={urlKey}
          items={items}
          itemsPerPage={variables.input.first}
          loadMore={loadMore}
          isLoading={loading}
        />
      </Grid>
    </Grid>
  )
}

QueryList.propTypes = {
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
  query: p.object.isRequired,

  // mapDataNode should map the struct from each node in `nodes` to the struct required by a PaginatedList item.
  mapDataNode: p.func,

  // variables will be added to the initial query. Useful for things like `favoritesFirst` or alert filters.
  // Note: The `input.search` and `input.first` parameters are included by default, but can be overridden.
  variables: p.object,

  // If set, the search string param is ignored.
  noSearch: p.bool,

  // filters unrelated to search, but still modify results, rendered to
  // the left of the search text field
  filter: p.node,

  // filters that enhance the search string, set within the search text field
  searchAdornment: p.node,

  /**
   * All other props are passed to PaginatedList
   */
}

QueryList.defaultProps = {
  mapDataNode: n => ({ title: n.name, url: n.id, subText: n.description }),
  variables: {},
}
