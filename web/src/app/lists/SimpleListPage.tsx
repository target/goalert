import React, { ComponentType, useState } from 'react'
import QueryList, { QueryListProps } from './QueryList'
import CreateFAB from './CreateFAB'

interface SimpleListPageProps extends QueryListProps {
  createDialogComponent: ComponentType<{ onClose: () => void }>
  createLabel: string
}

export default function SimpleListPage(
  props: SimpleListPageProps,
): JSX.Element {
  const { createDialogComponent, createLabel, ...rest } = props
  const [create, setCreate] = useState(false)
  const DialogComponent = createDialogComponent

  return (
    <React.Fragment>
      <QueryList {...rest} />
      <CreateFAB
        onClick={() => setCreate(true)}
        title={`Create ${createLabel}`}
      />
      {create && <DialogComponent onClose={() => setCreate(false)} />}
    </React.Fragment>
  )
}
