import React, { useState } from 'react'
import { gql, useQuery } from 'urql'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CreateFAB from '../../lists/CreateFAB'
import FlatList, { FlatListListItem } from '../../lists/FlatList'
import ServiceRuleCreateDialog from './ServiceRuleCreateDialog'
import { useIsWidthDown } from '../../util/useWidth'
import { Add } from '@mui/icons-material'
import makeStyles from '@mui/styles/makeStyles'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import { ServiceRule } from '../../../schema'
import { sortItems } from '../IntegrationKeyList'
import OtherActions from '../../util/OtherActions'
import ServiceRuleEditDialog from './ServiceRuleEditDialog'
import ServiceRuleDeleteDialog from './ServiceRuleDeleteDialog'
import ServiceRulesDrawer from './ServiceRulesDrawer'

const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      rules {
        id
        name
        serviceID
        actions {
          destType
          contents {
            prop
            value
          }
        }
        filters {
          field
          operator
          value
        }
        sendAlert
        integrationKeys {
          id
          name
          type
          name
          href
        }
      }
      integrationKeys {
        id
        name
        type
        name
        href
      }
    }
  }
`

const useStyles = makeStyles({
  copyIcon: {
    paddingRight: '0.25em',
    color: 'black',
  },
  keyLink: {
    display: 'flex',
    alignItems: 'center',
    width: 'fit-content',
  },
  spacing: {
    marginBottom: 96,
  },
  chip: {
    marginRight: '0.5em',
  },
  items: {
    paddingLeft: '1em',
  },
  actionItems: {
    padding: '0.5em',
    borderRadius: 2,
    boxShadow: '0px 0px 0px 1px rgba(0, 0, 0, 0.23)',
    marginBottom: '1em',
  },
})

export default function ServiceRulesList(props: {
  serviceID: string
}): JSX.Element {
  const classes = useStyles()
  const isMobile = useIsWidthDown('md')
  const [selectedRule, setSelectedRule] = useState<ServiceRule | null>(null)
  const [create, setCreate] = useState<boolean>(false)
  const [editRule, setEditRule] = useState<ServiceRule | null>(null)
  const [deleteRule, setDeleteRule] = useState<string | null>(null)

  const [{ fetching, error, data }] = useQuery({
    query,
    variables: { serviceID: props.serviceID },
  })
  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const items = (data.service.rules || [])
    .slice()
    .sort(sortItems)
    .map(
      (rule: ServiceRule): FlatListListItem => ({
        selected: rule.id === selectedRule?.id,
        highlight: rule.id === selectedRule?.id,
        title: rule.name,
        secondaryAction: (
          <OtherActions
            actions={[
              {
                label: 'Edit',
                onClick: () => setEditRule(rule),
              },
              {
                label: 'Delete',
                onClick: () => setDeleteRule(rule.id),
              },
            ]}
          />
        ),
        onClick: () => setSelectedRule(rule),
      }),
    )

  return (
    <React.Fragment>
      <ServiceRulesDrawer
        onClose={() => {
          setSelectedRule(null)
        }}
        rule={selectedRule}
        integrationKeys={data.service.integrationKeys}
      />
      <Grid item xs={12} className={classes.spacing}>
        <Card>
          <CardContent>
            <FlatList
              data-cy='signal-rules'
              headerNote='Rules are used to determine the action taken when a signal is received.'
              emptyMessage='No rules exist for this service.'
              items={items}
              headerAction={
                isMobile ? undefined : (
                  <Button
                    variant='contained'
                    onClick={(): void => setCreate(true)}
                    startIcon={<Add />}
                    data-testid='create-signal-rule'
                  >
                    Create Signal Rule
                  </Button>
                )
              }
            />
          </CardContent>
        </Card>
      </Grid>
      {isMobile && (
        <CreateFAB
          onClick={(): void => setCreate(true)}
          title='Create Signal Rule'
        />
      )}

      {create && (
        <ServiceRuleCreateDialog
          serviceID={props.serviceID}
          onClose={(): void => setCreate(false)}
          integrationKeys={data.service.integrationKeys}
        />
      )}
      {editRule && (
        <ServiceRuleEditDialog
          serviceID={props.serviceID}
          rule={editRule}
          onClose={(): void => setEditRule(null)}
          integrationKeys={data.service.integrationKeys}
        />
      )}
      {deleteRule && (
        <ServiceRuleDeleteDialog
          ruleID={deleteRule}
          onClose={(): void => setDeleteRule(null)}
        />
      )}
    </React.Fragment>
  )
}
