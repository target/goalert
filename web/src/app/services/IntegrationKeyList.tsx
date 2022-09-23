import React, { ReactNode, useState, ReactElement } from 'react'
import { gql, useQuery } from 'urql'
import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import IconButton from '@mui/material/IconButton'
import { Trash } from '../icons'
import IntegrationKeyCreateDialog from './IntegrationKeyCreateDialog'
import IntegrationKeyDeleteDialog from './IntegrationKeyDeleteDialog'
import RequireConfig from '../util/RequireConfig'
import CopyText from '../util/CopyText'
import AppLink from '../util/AppLink'

import makeStyles from '@mui/styles/makeStyles'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { IntegrationKey } from '../../schema'

interface Item {
  title: string
  subText: ReactElement
  secondaryAction: ReactElement
}

const query = gql`
  query ($serviceID: ID!) {
    service(id: $serviceID) {
      id # need to tie the result to the correct record
      integrationKeys {
        id
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

const sortItems = (a: IntegrationKey, b: IntegrationKey): number => {
  if (a.name.toLowerCase() < b.name.toLowerCase()) return -1
  if (a.name.toLowerCase() > b.name.toLowerCase()) return 1
  if (a.name < b.name) return -1
  if (a.name > b.name) return 1
  return 0
}

export function IntegrationKeyDetails(props: {
  href: string
  label: string
  type: string
}): JSX.Element {
  let copyText: ReactNode = (
    <CopyText title={'Copy ' + props.label} value={props.href} asURL />
  )

  // if link is not properly present, do not display to copy
  if (props.type === 'email' && !props.href.startsWith('mailto:')) {
    copyText = null
  }

  return (
    <React.Fragment>
      {copyText}
      {props.type === 'email' && (
        <RequireConfig
          configID='Mailgun.Enable'
          else={
            <React.Fragment>
              Email integration keys are currently disabled.
            </React.Fragment>
          }
        />
      )}
    </React.Fragment>
  )
}

export default function IntegrationKeyList(props: {
  serviceID: string
}): JSX.Element {
  const classes = useStyles()

  const [create, setCreate] = useState<boolean>(false)
  const [deleteDialog, setDeleteDialog] = useState<string | null>(null)

  const [{ fetching, error, data }] = useQuery({
    query,
    variables: { serviceID: props.serviceID },
  })

  const typeLabels = {
    generic: 'Generic API Key',
    grafana: 'Grafana Webhook URL',
    site24x7: 'Site24x7 Webhook URL',
    email: 'Email Address',
    prometheusAlertmanager: 'Alertmanager Webhook URL',
  }
  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const items = (data.service.integrationKeys || [])
    .slice()
    .sort(sortItems)
    .map(
      (key: IntegrationKey): Item => ({
        title: key.name,
        subText: (
          <IntegrationKeyDetails
            key={key.id}
            href={key.href}
            label={typeLabels[key.type]}
            type={key.type}
          />
        ),
        secondaryAction: (
          <IconButton
            onClick={(): void => setDeleteDialog(key.id)}
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
              headerNote={
                <React.Fragment>
                  API Documentation is available{' '}
                  <AppLink to='/docs'>here</AppLink>.
                </React.Fragment>
              }
              emptyMessage='No integration keys exist for this service.'
              items={items}
            />
          </CardContent>
        </Card>
      </Grid>
      <CreateFAB
        onClick={(): void => setCreate(true)}
        title='Create Integration Key'
      />
      {create && (
        <IntegrationKeyCreateDialog
          serviceID={props.serviceID}
          onClose={(): void => setCreate(false)}
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
