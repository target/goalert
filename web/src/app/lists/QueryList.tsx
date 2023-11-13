import React, { useMemo, useState, useEffect } from 'react'
import {
  useQuery,
  OperationVariables,
  QueryResult,
  DocumentNode,
} from '@apollo/client'
import { Card, Grid } from '@mui/material'
import { once } from 'lodash'
import { useURLKey, useURLParam } from '../actions/hooks'
import { PaginatedList, PaginatedListItemProps } from './PaginatedList'
import { ITEMS_PER_PAGE, POLL_INTERVAL } from '../config'
import { fieldAlias } from '../util/graphql'
import { GraphQLClientWithErrors } from '../apollo'
import ControlledPaginatedList, {
  ControlledPaginatedListProps,
} from './ControlledPaginatedList'
import { PageControls } from './PageControls'
import { ListHeader } from './ListHeader'
import CreateFAB from './CreateFAB'
import { useIsWidthDown } from '../util/useWidth'

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

export interface _QueryListProps extends ControlledPaginatedListProps {
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

  query: DocumentNode

  // mapDataNode should map the struct from each node in `nodes` to the struct required by a PaginatedList item
  mapDataNode?: (n: ObjectMap) => PaginatedListItemProps

  // variables will be added to the initial query. Useful for things like `favoritesFirst` or alert filters
  // note: The `input.search` and `input.first` parameters are included by default, but can be overridden
  variables?: OperationVariables

  // mapVariables transforms query variables just before submission
  mapVariables?: (vars: OperationVariables) => OperationVariables

  renderCreateDialog?: (onClose: () => void) => React.ReactNode | undefined

  createLabel?: string
  hideCreate?: boolean
}

export type QueryListProps = Omit<_QueryListProps, 'items'>

export default function QueryList(props: QueryListProps): React.ReactNode {
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
    renderCreateDialog,
    createLabel,
    hideCreate,
    ...listProps
  } = props
  const { input, ...vars } = variables
  const [page, setPage] = useState(0)
  const [showCreate, setShowCreate] = useState(false)
  const isMobile = useIsWidthDown('md')

  const [searchParam] = useURLParam('search', '')
  const urlKey = useURLKey()
  const aliasedQuery = useMemo(() => fieldAlias(query, 'data'), [query])

  // reset pageNumber on page reload
  useEffect(() => {
    setPage(0)
  }, [urlKey])

  const queryVariables = {
    ...vars,
    input: {
      first: ITEMS_PER_PAGE,
      ...input,
    },
  }

  if (searchParam) {
    queryVariables.input.search = searchParam
  }

  const { data, loading, fetchMore, stopPolling } = useQuery(aliasedQuery, {
    client: GraphQLClientWithErrors,
    variables: mapVariables(queryVariables),
    fetchPolicy: 'network-only',
    pollInterval: POLL_INTERVAL,
  })

  const nodes = data?.data?.nodes ?? []
  const items = nodes.map(mapDataNode)
  const itemCount = items.length
  let loadMore: ((numberToLoad?: number) => void) | undefined

  // isLoading returns true if the parent says we are, or
  // we are currently on an incomplete page and `loadMore` is available.
  const isLoading = (() => {
    if (!data && loading) return true

    // We are on a future/incomplete page and loadMore is true
    if ((page + 1) * ITEMS_PER_PAGE > itemCount && loadMore) return true

    return false
  })()

  const pageCount = Math.ceil(items.length / ITEMS_PER_PAGE)

  if (itemCount < ITEMS_PER_PAGE && page > 0) setPage(0)

  if (data?.data?.pageInfo?.hasNextPage) {
    loadMore = buildFetchMore(
      fetchMore,
      data.data.pageInfo.endCursor,
      stopPolling,
      queryVariables.input.first,
    )
  }

  function renderList(): React.ReactNode {
    if (
      props.checkboxActions?.length ||
      props.secondaryActions ||
      !props.noSearch
    ) {
      return (
        <ControlledPaginatedList
          {...listProps}
          listHeader={
            <ListHeader
              cardHeader={props.cardHeader}
              headerNote={props.headerNote}
              headerAction={props.headerAction}
            />
          }
          items={items}
          itemsPerPage={queryVariables.input.first}
          page={page}
          isLoading={isLoading}
          loadMore={loadMore}
          noSearch={noSearch}
          renderCreateDialog={renderCreateDialog}
          createLabel={createLabel}
          hideCreate={hideCreate}
        />
      )
    }

    return (
      <Grid item xs={12}>
        <Card>
          <ListHeader
            cardHeader={props.cardHeader}
            headerNote={props.headerNote}
            headerAction={props.headerAction}
          />
          <PaginatedList
            {...listProps}
            key={urlKey}
            items={items}
            page={page}
            isLoading={isLoading}
            itemsPerPage={queryVariables.input.first}
            loadMore={loadMore}
          />
        </Card>
      </Grid>
    )
  }

  return (
    <Grid container spacing={2}>
      {renderList()}
      {!props.infiniteScroll && (
        <PageControls
          pageCount={pageCount}
          loadMore={loadMore}
          page={page}
          setPage={setPage}
          isLoading={isLoading}
        />
      )}
      {!hideCreate && isMobile && renderCreateDialog && createLabel && (
        <React.Fragment>
          <CreateFAB
            onClick={() => setShowCreate(true)}
            title={`Create ${createLabel}`}
          />
          {showCreate && renderCreateDialog(() => setShowCreate(false))}
        </React.Fragment>
      )}
    </Grid>
  )
}
