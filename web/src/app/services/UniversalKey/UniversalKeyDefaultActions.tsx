import { Button, Card } from '@mui/material'
import React, { Suspense, useState } from 'react'
import UniversalKeyActionsList from './UniversalKeyActionsList'
import { gql, useQuery } from 'urql'
import { IntegrationKey } from '../../../schema'
import { Add } from '../../icons'
import { UniversalKeyActionDialog } from './UniversalKeyActionDialog'
import CompList from '../../lists/CompList'
import { CompListItemText } from '../../lists/CompListItems'

interface UniversalKeyDefaultActionProps {
  serviceID: string
  keyID: string
}

const query = gql`
  query UniversalKeyPage($keyID: ID!) {
    integrationKey(id: $keyID) {
      id
      config {
        defaultActions {
          dest {
            type
            args
          }
          params
        }
      }
    }
  }
`

export default function UniversalKeyDefaultActions(
  props: UniversalKeyDefaultActionProps,
): React.ReactNode {
  const [editActionIndex, setEditActionIndex] = useState(-1)
  const [addAction, setAddAction] = useState(false)
  const [q] = useQuery<{ integrationKey: IntegrationKey }>({
    query,
    variables: { keyID: props.keyID },
  })

  return (
    <React.Fragment>
      <Card>
        <CompList
          emptyMessage='No default action'
          action={
            <Button
              variant='contained'
              startIcon={<Add />}
              onClick={() => setAddAction(true)}
            >
              Add Action
            </Button>
          }
          note='Default actions are performed when zero rules match.'
        >
          <CompListItemText
            title={
              <UniversalKeyActionsList
                actions={q.data?.integrationKey.config.defaultActions ?? []}
                onEdit={(index) => setEditActionIndex(index)}
              />
            }
          />
        </CompList>
      </Card>
      <Suspense>
        {editActionIndex > -1 && (
          <UniversalKeyActionDialog
            onClose={() => setEditActionIndex(-1)}
            keyID={props.keyID}
            actionIndex={editActionIndex}
          />
        )}
        {addAction && (
          <UniversalKeyActionDialog
            onClose={() => setAddAction(false)}
            keyID={props.keyID}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
