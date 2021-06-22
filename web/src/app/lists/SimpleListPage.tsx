/* eslint-disable @typescript-eslint/no-explicit-any */
import React, { useState, ReactElement } from 'react'
import CreateFAB from './CreateFAB'
import ControlledPaginatedList, {
  ControlledPaginatedListProps,
} from './ControlledPaginatedList'
import { DocumentNode } from '@apollo/client'
import { usePaginatedQuery } from './usePaginatedQuery'

interface SimpleListPageProps extends ControlledPaginatedListProps {
  query: DocumentNode
  variables: Record<string, any>
  mapDataNode: (
    node: Record<string, any>,
  ) => ControlledPaginatedListProps['items']
  createForm: ReactElement
  createLabel: string
}

export default function SimpleListPage(
  props: SimpleListPageProps,
): JSX.Element {
  const {
    query,
    variables,
    mapDataNode,
    createForm,
    createLabel,
    ...listProps
  } = props
  const [create, setCreate] = useState(false)

  const { q, loadMore } = usePaginatedQuery(query, variables)
  const items = (q.data?.data?.nodes ?? []).map(mapDataNode)

  return (
    <React.Fragment>
      <ControlledPaginatedList
        {...listProps}
        items={items}
        loadMore={loadMore}
        isLoading={!q.data || q.loading}
      />

      {createForm && (
        <CreateFAB
          onClick={() => setCreate(true)}
          title={`Create ${createLabel}`}
        />
      )}

      {create &&
        React.cloneElement(createForm, {
          onClose: () => setCreate(false),
        })}
    </React.Fragment>
  )
}
