import React, { ComponentType, useState } from 'react'
import QueryList, { QueryListProps } from './QueryList'
import CreateFAB from './CreateFAB'

interface SimpleListPageProps extends QueryListProps {
  createLabel?: string
  createDialogComponent?: ComponentType<{ onClose: () => void }>
}

export default function SimpleListPage(
  props: SimpleListPageProps,
): JSX.Element {
  const { createDialogComponent: DialogComponent, createLabel, ...rest } = props
  const [create, setCreate] = useState(false)

  return (
    <React.Fragment>
      <QueryList {...rest} />

      {createLabel && (
        <CreateFAB
          onClick={() => setCreate(true)}
          title={`Create ${createLabel}`}
        />
      )}
      <CreateFAB
        onClick={() => setCreate(true)}
        title={`Create ${createLabel}`}
      />
      {create && DialogComponent && (
        <DialogComponent onClose={() => setCreate(false)} />
      )}
    </React.Fragment>
  )
}
