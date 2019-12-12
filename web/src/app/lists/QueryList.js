import React from 'react'
import { useSelector } from 'react-redux'
import p from 'prop-types'
import { useQuery } from '@apollo/react-hooks'
import { Grid, makeStyles } from '@material-ui/core'
import { once } from 'lodash-es'
import { PaginatedList } from './PaginatedList'
import { ITEMS_PER_PAGE } from '../config'
import { searchSelector, urlKeySelector } from '../selectors'
// import { fieldAlias } from '../util/graphql'
import Search from '../util/Search'
import { GraphQLClientWithErrors } from '../apollo'

const useStyles = makeStyles({
  flexGrow: {
    flexGrow: 1,
  },
})

const buildFetchMore = (fetchMore, after) => {
  return once(num =>
    fetchMore({
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
    }),
  )
}

export default function QueryList(props) {
  const { noSearch, query, searchAdornment, ...listProps } = props
  const { input, ...vars } = props.variables

  const classes = useStyles()
  const searchParam = useSelector(searchSelector)
  const urlKey = useSelector(urlKeySelector)

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

  // const aliasedQuery = fieldAlias(query, 'data')
  // TODO use aliasedQuery
  const { data, loading, fetchMore } = useQuery(query, {
    client: GraphQLClientWithErrors,
    variables,
  })

  let items = []
  let loadMore

  if (data && data.data && data.data.nodes) {
    items = data.data.nodes.map(props.mapDataNode)
    if (data.data.pageInfo.hasNextPage) {
      loadMore = buildFetchMore(fetchMore, data.data.pageInfo.endCursor)
    }
  }

  return (
    <Grid container spacing={2}>
      <Grid item className={classes.flexGrow} />

      {/* Such that filtering/searching isn't re-rendered with the page content */}
      <Grid item>
        <Search endAdornment={searchAdornment} />
      </Grid>

      <PaginatedList
        {...listProps}
        key={urlKey}
        items={items}
        loadMore={loadMore}
        isLoading={loading}
      />
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

  // filters additional to search, set in the search text field.
  searchAdornment: p.node,
}

QueryList.defaultProps = {
  mapDataNode: n => ({ title: n.name, url: n.id, subText: n.description }),
  variables: {},
}
