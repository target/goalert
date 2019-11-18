import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import { PaginatedList } from './PaginatedList'
import { ITEMS_PER_PAGE } from '../config'
import { once } from 'lodash-es'
import { connect } from 'react-redux'
import { searchSelector } from '../selectors'
import Query from '../util/Query'
import { fieldAlias } from '../util/graphql'
import Search from '../util/Search'
import withStyles from '@material-ui/core/styles/withStyles'

const styles = {
  flexGrow: {
    flexGrow: 1,
  },
}

const mapStateToProps = state => ({
  search: searchSelector(state),
  routeKey: state.router.location.key,
})

@withStyles(styles)
@connect(mapStateToProps)
export default class QueryList extends React.PureComponent {
  static propTypes = {
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

    // provided by redux
    search: p.string,
    routeKey: p.string,

    // filters additional to search, set in the search text field.
    searchAdornment: p.node,
  }

  static defaultProps = {
    mapDataNode: n => ({ title: n.name, url: n.id, subText: n.description }),
    variables: {},
  }

  buildFetchMore = (fetchMore, after) => {
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

  renderContent = ({ data, loading, fetchMore }) => {
    let items = []
    let loadMore
    const {
      classes,
      query,
      mapDataNode,
      variables,
      noSearch,
      search,
      ...listProps
    } = this.props

    if (data && data.data && data.data.nodes) {
      items = data.data.nodes.map(this.props.mapDataNode)
      if (data.data.pageInfo.hasNextPage) {
        loadMore = this.buildFetchMore(fetchMore, data.data.pageInfo.endCursor)
      }
    }

    return (
      <PaginatedList
        {...listProps}
        key={this.props.routeKey}
        items={items}
        loadMore={loadMore}
        isLoading={loading}
      />
    )
  }

  render() {
    const { classes, noSearch, query, search, searchAdornment } = this.props
    const { input, ...vars } = this.props.variables

    const variables = {
      ...vars,
      input: {
        first: ITEMS_PER_PAGE,
        search,
        ...input,
      },
    }

    if (noSearch) {
      delete variables.input.search
    }

    return (
      <Grid container spacing={2}>
        <Grid item className={classes.flexGrow} />

        {/* Such that filtering/searching isn't re-rendered with the page content */}
        <Grid item>
          <Search endAdornment={searchAdornment} />
        </Grid>

        <Grid item xs={12}>
          <Query
            query={fieldAlias(query, 'data')}
            variables={variables}
            noPoll
            noSpin
            notifyOnNetworkStatusChange
            render={this.renderContent}
          />
        </Grid>
      </Grid>
    )
  }
}
