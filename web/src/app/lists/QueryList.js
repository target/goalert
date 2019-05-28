import React from 'react'
import p from 'prop-types'

import { PaginatedList } from './PaginatedList'
import { ITEMS_PER_PAGE } from '../config'
import { once } from 'lodash-es'
import { connect } from 'react-redux'
import { searchSelector } from '../selectors'
import Query from '../util/Query'
import { fieldAlias } from '../util/graphql'

const mapStateToProps = state => ({
  search: searchSelector(state),
})

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
        key={this.props.search}
        items={items}
        loadMore={loadMore}
        isLoading={loading}
      />
    )
  }

  render() {
    const { input, ...vars } = this.props.variables

    const variables = {
      ...vars,
      input: {
        first: ITEMS_PER_PAGE,
        search: this.props.search,
        ...input,
      },
    }
    if (this.props.noSearch) {
      delete variables.input.search
    }
    return (
      <Query
        query={fieldAlias(this.props.query, 'data')}
        variables={variables}
        noPoll
        notifyOnNetworkStatusChange
        render={this.renderContent}
      />
    )
  }
}
