import React, { useState } from 'react'
import { gql, useQuery } from 'urql'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import Box from '@mui/material/Box'
import CardContent from '@mui/material/CardContent'
import CreateFAB from '../../lists/CreateFAB'
import FlatList, { FlatListListItem } from '../../lists/FlatList'
import ServiceRuleCreateDialog from './ServiceRuleCreateDialog'
import { useIsWidthDown } from '../../util/useWidth'
import { Add } from '@mui/icons-material'
import makeStyles from '@mui/styles/makeStyles'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import {
  Content,
  IntegrationKey,
  ServiceRule,
  ServiceRuleAction,
  ServiceRuleFilter,
} from '../../../schema'
import { sortItems } from '../IntegrationKeyList'
import { Typography, Chip, Switch } from '@mui/material'
import OtherActions from '../../util/OtherActions'
import ServiceRuleEditDialog, { getCustomFields } from './ServiceRuleEditDialog'
import ServiceRuleDeleteDialog from './ServiceRuleDeleteDialog'
import { destType } from './ServiceRuleForm'
import toTitleCase from '../../util/toTitleCase'

const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      rules {
        id
        name
        serviceID
        actions {
          destType
          destID
          destValue
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

function ActionDetails(props: { action: ServiceRuleAction }): JSX.Element {
  const { action } = props
  const classes = useStyles()

  return (
    <Box className={classes.actionItems}>
      <Typography variant='body1'>{action.destType}</Typography>
      <Box className={classes.items}>
        {action.contents.map((content: Content, idx: number) => {
          return (
            <React.Fragment key={idx}>
              <Typography variant='body1'>
                {toTitleCase(content.prop.replace(/_/g, ' '))}
              </Typography>
              <Typography className={classes.items} variant='body1'>
                {content.value}
              </Typography>
            </React.Fragment>
          )
        })}
      </Box>
    </Box>
  )
}

function RuleDetails(props: { rule: ServiceRule }): JSX.Element {
  const { rule } = props
  const classes = useStyles()
  const customFields = getCustomFields(rule)

  return (
    <Grid container spacing={1}>
      <Grid item style={{ flexGrow: 1 }} xs={12}>
        <Typography variant='h4'>{rule.name}</Typography>
      </Grid>
      <Grid item style={{ flexGrow: 1 }} xs={12}>
        <Typography variant='body1'>Integration Keys</Typography>
        <Box className={classes.items}>
          {rule.integrationKeys.map((key: IntegrationKey) => (
            <Chip key={key.id} label={key.name} className={classes.chip} />
          ))}
        </Box>
      </Grid>
      {rule.filters.length > 0 && (
        <Grid item style={{ flexGrow: 1 }} xs={12}>
          <Typography variant='body1'>Filters</Typography>
          <Box className={classes.items}>
            {rule.filters.map((f: ServiceRuleFilter, idx: number) => (
              <Chip
                key={idx}
                label={`${f.field} ${f.operator} ${f.value}`}
                className={classes.chip}
              />
            ))}
          </Box>
        </Grid>
      )}
      <Grid item style={{ flexGrow: 1 }} xs={12}>
        <Typography variant='body1'>Create Alert</Typography>
        <Switch checked={rule.sendAlert} disabled />
      </Grid>
      {customFields && (
        <Grid item style={{ flexGrow: 1 }} xs={12}>
          <Typography variant='body1'>Alert Custom Fields</Typography>
          <Box className={classes.items}>
            <Typography variant='body1' display='inline'>
              Summary
            </Typography>
            <Typography variant='body1' display='inline'>
              : {`${customFields?.summary}`}
            </Typography>
          </Box>
          <Box className={classes.items}>
            <Typography variant='body1' display='inline'>
              Details
            </Typography>
            <Typography variant='body1' display='inline'>
              : {`${customFields?.details}`}
            </Typography>
          </Box>
        </Grid>
      )}
      <Grid item style={{ flexGrow: 1 }} xs={12}>
        <Typography variant='body1'>Destinations</Typography>
        {rule.actions.map((action, idx) => {
          if (action.destType !== destType.ALERT) {
            return <ActionDetails key={idx} action={action} />
          }
        })}
      </Grid>
    </Grid>
  )
}

export default function ServiceRulesList(props: {
  serviceID: string
}): JSX.Element {
  const classes = useStyles()
  const isMobile = useIsWidthDown('md')
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
        subText: <RuleDetails key={rule.id} rule={rule} />,
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
      }),
    )

  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.spacing}>
        <Card>
          <CardContent>
            <FlatList
              data-cy='int-keys'
              headerNote='Rules are used to determine the action taken when a signal is received.'
              emptyMessage='No rules exist for this service.'
              items={items}
              headerAction={
                isMobile ? undefined : (
                  <Button
                    variant='contained'
                    onClick={(): void => setCreate(true)}
                    startIcon={<Add />}
                    data-testid='create-key'
                  >
                    Create New Rule
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
