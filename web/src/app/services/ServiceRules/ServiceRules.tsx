import React, { useState } from 'react'
import { gql, useQuery } from 'urql'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CreateFAB from '../../lists/CreateFAB'
import FlatList, { FlatListListItem } from '../../lists/FlatList'
import IconButton from '@mui/material/IconButton'
import { Trash } from '../../icons'
import ServiceRuleCreateDialog from './ServiceRuleCreateDialog'
import IntegrationKeyDeleteDialog from '../IntegrationKeyDeleteDialog'
import { useIsWidthDown } from '../../util/useWidth'
import { Add } from '@mui/icons-material'
import makeStyles from '@mui/styles/makeStyles'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'
import { IntegrationKey, ServiceRule } from '../../../schema'
import { sortItems } from '../IntegrationKeyList'
import { Typography, Chip } from '@mui/material'

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
})

export function ServiceRuleDetails(props: { rule: ServiceRule }): JSX.Element {
  const { rule } = props
  return (
    <Grid container>
      <Grid item style={{ flexGrow: 1 }} xs={12}>
        <Typography variant='subtitle1'>Integration Keys</Typography>
        {!rule.integrationKeys ? (
          <Typography variant='body1'>No Integrations Keys Set</Typography>
        ) : (
          rule.integrationKeys.map((key: IntegrationKey) => (
            <Chip key={key.id} label={key.name} />
          ))
        )}
      </Grid>
    </Grid>
  )
}

export default function SignalRules(props: { serviceID: string }): JSX.Element {
  const classes = useStyles()
  const isMobile = useIsWidthDown('md')
  const [create, setCreate] = useState<boolean>(false)
  const [deleteDialog, setDeleteDialog] = useState<string | null>(null)

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
        title: rule.name,
        subText: <ServiceRuleDetails key={rule.id} rule={rule} />,
        secondaryAction: (
          <IconButton
            onClick={(): void => setDeleteDialog(rule.id)}
            size='large'
          >
            <Trash />
          </IconButton>
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
      {deleteDialog && (
        <IntegrationKeyDeleteDialog
          integrationKeyID={deleteDialog}
          onClose={(): void => setDeleteDialog(null)}
        />
      )}
    </React.Fragment>
  )
}
