import React, { Suspense, useState } from 'react'
import { Button, Card, CardContent } from '@mui/material'
import { Add } from '@mui/icons-material'
import FlatList, { FlatListListItem } from '../../lists/FlatList'
import { useIsWidthDown } from '../../util/useWidth'
import UniversalKeyRuleCreateDialog from './UniversalKeyRuleCreateDialog'
import UniversalKeyRuleEditDialog from './UniversalKeyRuleEditDialog'
import UniversalKeyRuleRemoveDialog from './UniversalKeyRuleRemoveDialog'
import OtherActions from '../../util/OtherActions'

export default function UniversalKeyRuleList(): JSX.Element {
  const [create, setCreate] = useState(false)
  const [edit, setEdit] = useState(false)
  const [remove, setRemove] = useState(false)
  const isMobile = useIsWidthDown('md')

  const items: FlatListListItem[] = [
    {
      title: 'testRule1',
      subText: 'Example Filter, Example Dest',
      secondaryAction: (
        <OtherActions
          actions={[
            {
              label: 'Edit',
              onClick: () => setEdit(true),
            },
            {
              label: 'Delete',
              onClick: () => setRemove(true),
            },
          ]}
        />
      ),
    },
  ]

  return (
    <React.Fragment>
      <Card>
        <FlatList
          emptyMessage='No rules exist for this integration key.'
          headerAction={
            isMobile ? undefined : (
              <Button
                variant='contained'
                startIcon={<Add />}
                onClick={() => setCreate(true)}
              >
                Create Rule
              </Button>
            )
          }
          headerNote='Rules are a set of filters that allow notifications to be sent to a specific destination. '
          items={items}
        />
      </Card>

      <Suspense>
        {create && (
          <UniversalKeyRuleCreateDialog onClose={() => setCreate(false)} />
        )}
        {edit && <UniversalKeyRuleEditDialog onClose={() => setEdit(false)} />}
        {remove && (
          <UniversalKeyRuleRemoveDialog onClose={() => setRemove(false)} />
        )}
      </Suspense>
    </React.Fragment>
  )
}
