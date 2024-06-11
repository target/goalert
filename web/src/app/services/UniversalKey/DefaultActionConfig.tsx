import { Button, Card } from '@mui/material'
import React, { Suspense, useState } from 'react'
import FlatList, { FlatListListItem } from '../../lists/FlatList'
import { Edit } from '@mui/icons-material'
import DefaultActionEditDialog from './DefaultActionEditDialog'

interface UniversalKeyDefaultActionProps {
  serviceID: string
  keyID: string
}

export default function UniversalKeyDefaultAction(
  props: UniversalKeyDefaultActionProps,
): React.ReactNode {
  const [edit, setEdit] = useState(false)

  const items: FlatListListItem[] = [
    {
      title: 'default action',
      subText: 'condition',
    },
  ]

  return (
    <React.Fragment>
      <Card>
        <FlatList
          emptyMessage='No rules exist for this integration key.'
          headerAction={
            <Button
              variant='contained'
              startIcon={<Edit />}
              onClick={() => setEdit(true)}
            >
              Edit Default Action
            </Button>
          }
          headerNote='Default Actions are taken when no other rules match.'
          items={items}
        />
      </Card>
      <Suspense>
        {edit && (
          <DefaultActionEditDialog
            onClose={() => setEdit(false)}
            keyID={props.keyID}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
