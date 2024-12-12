import React, { Suspense, useState } from 'react'
import { Button, Card, Grid, Typography } from '@mui/material'
import { Add } from '@mui/icons-material'
import FlatList, { FlatListListItem } from '../../lists/FlatList'
import UniversalKeyRuleDialog from './UniversalKeyRuleDialog'
import UniversalKeyRuleRemoveDialog from './UniversalKeyRuleRemoveDialog'
import OtherActions from '../../util/OtherActions'
import { gql, useQuery } from 'urql'
import { IntegrationKey, Service } from '../../../schema'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import UniversalKeyActionsList from './UniversalKeyActionsList'
import { UniversalKeyActionDialog } from './UniversalKeyActionDialog'

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
          continueAfterMatch
          actions {
            dest {
              type
              args
            }
            params
          }
        }
      }
    }
  }
`

/* truncateCond truncates the condition expression to 50 characters and a single line. */
function truncateCond(cond: string): string {
  const singleLineCond = cond.replace(/\s+/g, ' ')
  if (singleLineCond.length > 50) {
    return singleLineCond.slice(0, 50) + '...'
  }
  return singleLineCond
}

export default function UniversalKeyRuleList(
  props: UniversalKeyRuleListProps,
): JSX.Element {
  const [create, setCreate] = useState(false)
  const [edit, setEdit] = useState('')
  const [remove, setRemove] = useState('')
  const [editAction, setEditAction] = useState<null | {
    ruleID: string
    actionIndex: number
  }>(null)
  const [addAction, setAddAction] = useState<null | { ruleID: string }>(null)

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
  ).map((rule, idx) => ({
    title: (
      <Typography component='h4' variant='subtitle1' sx={{ pb: 1 }}>
        <b>{rule.name}:</b>
        &nbsp;
        {rule.description}
      </Typography>
    ),
    subText: (
      <Grid container>
        <Grid item xs={12}>
          <b>If</b>
          <code style={{ paddingLeft: '1em' }}>
            {truncateCond(rule.conditionExpr)}
          </code>
        </Grid>
        <Grid item xs={12}>
          <b>Then</b>
        </Grid>
        <Grid item xs={12} sx={{ paddingLeft: '1em' }}>
          <UniversalKeyActionsList
            actions={rule.actions}
            onEdit={(index) =>
              setEditAction({ ruleID: rule.id, actionIndex: index })
            }
          />
        </Grid>
        <Grid item xs={12}>
          <b>Finally</b> {rule.continueAfterMatch ? 'continue' : 'stop'}
        </Grid>
      </Grid>
    ),
    secondaryAction: (
      <OtherActions
        actions={[
          {
            label: 'Add Action',
            onClick: () => setAddAction({ ruleID: rule.id }),
          },
          {
            label: 'Edit Rule',
            onClick: () => setEdit(rule.id),
          },
          {
            label: 'Delete Rule',
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
          onReorder={() => {}}
        />
      </Card>

      <Suspense>
        {(create || edit) && (
          <UniversalKeyRuleDialog
            onClose={() => {
              setCreate(false)
              setEdit('')
            }}
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
        {editAction && (
          <UniversalKeyActionDialog
            onClose={() => setEditAction(null)}
            keyID={props.keyID}
            ruleID={editAction.ruleID}
            actionIndex={editAction.actionIndex}
          />
        )}
        {addAction && (
          <UniversalKeyActionDialog
            onClose={() => setAddAction(null)}
            keyID={props.keyID}
            ruleID={addAction.ruleID}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
