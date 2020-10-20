import { gql, useQuery } from '@apollo/client'
import React, { useState } from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Card from '@material-ui/core/Card'
import CardContent from '@material-ui/core/CardContent'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import IconButton from '@material-ui/core/IconButton'
import { Trash } from '../icons'
import IntegrationKeyCreateDialog from './IntegrationKeyCreateDialog'
import IntegrationKeyDeleteDialog from './IntegrationKeyDeleteDialog'
import RequireConfig from '../util/RequireConfig'
import CopyText from '../util/CopyText'
import { AppLink } from '../util/AppLink'

import { makeStyles } from '@material-ui/core'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'

const query = gql`
  query($serviceID: ID!) {
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

const sortItems = (a, b) => {
  if (a.name.toLowerCase() < b.name.toLowerCase()) return -1
  if (a.name.toLowerCase() > b.name.toLowerCase()) return 1
  if (a.name < b.name) return -1
  if (a.name > b.name) return 1
  return 0
}

export function IntegrationKeyDetails(props) {
  let copyText = <CopyText title={'Copy ' + props.label} value={props.href} />

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
          else='Email integration keys are currently disabled.'
        />
      )}
    </React.Fragment>
  )
}

IntegrationKeyDetails.propTypes = {
  href: p.string.isRequired,
  label: p.string.isRequired,
  type: p.string.isRequired,
}

export default function IntegrationKeyList(props) {
  const classes = useStyles()

  const [create, setCreate] = useState(false)
  const [deleteDialog, setDeleteDialog] = useState(null)

  const { loading, error, data } = useQuery(query, {
    variables: { serviceID: props.serviceID },
  })

  const typeLabels = {
    generic: 'Generic API Key',
    grafana: 'Grafana Webhook URL',
    site24x7: 'Site24x7 Webhook URL',
    email: 'Email Address',
    prometheusAlertmanager: 'Alertmanager Webhook URL',
  }
  if (loading && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  const items = (data.service.integrationKeys || [])
    .slice()
    .sort(sortItems)
    .map((key) => ({
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
        <IconButton onClick={() => setDeleteDialog(key.id)}>
          <Trash />
        </IconButton>
      ),
    }))

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
        onClick={() => setCreate(true)}
        title='Create Integration Key'
      />
      {create && (
        <IntegrationKeyCreateDialog
          serviceID={props.serviceID}
          onClose={() => setCreate(false)}
        />
      )}
      {deleteDialog && (
        <IntegrationKeyDeleteDialog
          integrationKeyID={deleteDialog}
          onClose={() => setDeleteDialog(null)}
        />
      )}
    </React.Fragment>
  )
}

IntegrationKeyList.propTypes = {
  serviceID: p.string.isRequired,
}
