import { Button, Card } from '@mui/material'
import React, { Suspense, useState } from 'react'
import FlatList from '../../lists/FlatList'
import DefaultActionEditDialog from './DefaultActionEditDialog'
import UniversalKeyActionsList from './UniversalKeyActionsList'
import { gql, useQuery } from 'urql'
import { IntegrationKey } from '../../../schema'
import { Add } from '../../icons'

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
  const [edit, setEdit] = useState(false)
  const [q] = useQuery<{ integrationKey: IntegrationKey }>({
    query,
    variables: { keyID: props.keyID },
  })

  return (
    <React.Fragment>
      <Card>
        <FlatList
          emptyMessage='No default action'
          headerAction={
            <Button
              variant='contained'
              startIcon={<Add />}
              onClick={() => setEdit(true)}
            >
              Add Action
            </Button>
          }
          headerNote='Default actions are performed when zero rules match.'
          items={[
            {
              title: (
                <UniversalKeyActionsList
                  actions={q.data?.integrationKey.config.defaultActions ?? []}
                />
              ),
            },
          ]}
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
