import React, { useState, ReactElement } from 'react'
import QueryList, { QueryListProps } from './QueryList'
import CreateFAB from './CreateFAB'

interface SimpleListPageProps extends QueryListProps {
  createForm: ReactElement
  createLabel: string
}

export default function SimpleListPage(
  props: SimpleListPageProps,
): JSX.Element {
  const { createForm, createLabel, ...rest } = props
  const [create, setCreate] = useState(false)

  return (
    <React.Fragment>
      <QueryList {...rest} />

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
