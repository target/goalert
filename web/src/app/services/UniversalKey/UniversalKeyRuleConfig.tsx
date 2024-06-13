import React, { Suspense, useState } from 'react'
import { Button, Card } from '@mui/material'
import { Add } from '@mui/icons-material'
import FlatList, { FlatListListItem } from '../../lists/FlatList'
import UniversalKeyRuleCreateDialog from './UniversalKeyRuleCreateDialog'
import UniversalKeyRuleEditDialog from './UniversalKeyRuleEditDialog'
import UniversalKeyRuleRemoveDialog from './UniversalKeyRuleRemoveDialog'
import OtherActions from '../../util/OtherActions'
import { gql, useQuery } from 'urql'
import { IntegrationKey, Service } from '../../../schema'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'

interface UniversalKeyRuleListProps {
  serviceID: string
  keyID: string
}

const query = gql`
  query UniversalKeyPage($keyID: ID!) {
    integrationKey(id: $keyID) {
      id
      config {
        rules {
          id
          name
          description
          conditionExpr
          actions {
            params {
              paramID
            }
          }
        }
      }
    }
  }
`

export default function UniversalKeyRuleList(
  props: UniversalKeyRuleListProps,
): JSX.Element {
  const [create, setCreate] = useState(false)
  const [edit, setEdit] = useState('')
  const [remove, setRemove] = useState('')

  const [{ data, fetching, error }] = useQuery<{
    integrationKey: IntegrationKey
    service: Service
  }>({
    query,
    variables: {
      keyID: props.keyID,
      serviceID: props.serviceID,
    },
  })

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const items: FlatListListItem[] = (
    data?.integrationKey.config.rules ?? []
  ).map((rule) => ({
    title: rule.name,
    subText: rule.description,
    secondaryAction: (
      <OtherActions
        actions={[
          {
            label: 'Edit',
            onClick: () => setEdit(rule.id),
          },
          {
            label: 'Delete',
            onClick: () => setRemove(rule.id),
          },
        ]}
      />
    ),
  }))

  return (
    <React.Fragment>
      <Card>
        <FlatList
          emptyMessage='No rules exist for this integration key.'
          headerAction={
            <Button
              variant='contained'
              startIcon={<Add />}
              onClick={() => setCreate(true)}
            >
              Create Rule
            </Button>
          }
          headerNote='Rules are a set of filters that allow notifications to be sent to a specific destination. '
          items={items}
        />
      </Card>

      <Suspense>
        {create && (
          <UniversalKeyRuleCreateDialog
            onClose={() => setCreate(false)}
            keyID={props.keyID}
          />
        )}
        {edit && (
          <UniversalKeyRuleEditDialog
            onClose={() => setEdit('')}
            keyID={props.keyID}
            ruleID={edit}
          />
        )}
        {remove && (
          <UniversalKeyRuleRemoveDialog
            onClose={() => setRemove('')}
            keyID={props.keyID}
            ruleID={remove}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
