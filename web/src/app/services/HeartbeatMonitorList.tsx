import React, { useState, ReactElement } from 'react'
import { useQuery, gql } from 'urql'
import Button from '@mui/material/Button'
import Grid from '@mui/material/Grid'
import Card from '@mui/material/Card'
import CardContent from '@mui/material/CardContent'
import CreateFAB from '../lists/CreateFAB'
import FlatList from '../lists/FlatList'
import HeartbeatMonitorCreateDialog from './HeartbeatMonitorCreateDialog'
import makeStyles from '@mui/styles/makeStyles'
import HeartbeatMonitorEditDialog from './HeartbeatMonitorEditDialog'
import HeartbeatMonitorDeleteDialog from './HeartbeatMonitorDeleteDialog'
import OtherActions from '../util/OtherActions'
import HeartbeatMonitorStatus from './HeartbeatMonitorStatus'
import CopyText from '../util/CopyText'
import Spinner from '../loading/components/Spinner'
import { GenericError } from '../error-pages'
import { HeartbeatMonitor } from '../../schema'
import { useIsWidthDown } from '../util/useWidth'
import { Add } from '@mui/icons-material'
import { Time } from '../util/Time'

// generates a single alert if a POST is not received before the timeout
const HEARTBEAT_MONITOR_DESCRIPTION =
  'Heartbeat monitors create an alert if no heartbeat is received (a POST request) before the configured timeout.'

const query = gql`
  query monitorQuery($serviceID: ID!) {
    service(id: $serviceID) {
      id # need to tie the result to the correct record
      heartbeatMonitors {
        id
        name
        timeoutMinutes
        lastState
        lastHeartbeat
        href
      }
    }
  }
`

const useStyles = makeStyles(() => ({
  text: {
    display: 'flex',
    alignItems: 'center',
    width: 'fit-content',
  },
  spacing: {
    marginBottom: 96,
  },
}))

const sortItems = (a: HeartbeatMonitor, b: HeartbeatMonitor): number => {
  if (a.name.toLowerCase() < b.name.toLowerCase()) return -1
  if (a.name.toLowerCase() > b.name.toLowerCase()) return 1
  if (a.name < b.name) return -1
  if (a.name > b.name) return 1
  return 0
}

export default function HeartbeatMonitorList(props: {
  serviceID: string
}): JSX.Element {
  const classes = useStyles()
  const isMobile = useIsWidthDown('md')
  const [showCreateDialog, setShowCreateDialog] = useState(false)
  const [showEditDialogByID, setShowEditDialogByID] = useState<string | null>(
    null,
  )
  const [showDeleteDialogByID, setShowDeleteDialogByID] = useState<
    string | null
  >(null)

  const [{ data, fetching, error }] = useQuery({
    query,
    variables: { serviceID: props.serviceID },
  })

  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  function renderList(monitors: HeartbeatMonitor[]): ReactElement {
    const items = (monitors || [])
      .slice()
      .sort(sortItems)
      .map((monitor) => ({
        icon: (
          <HeartbeatMonitorStatus
            lastState={monitor.lastState}
            lastHeartbeat={monitor.lastHeartbeat}
          />
        ),
        title: monitor.name,
        subText: (
          <React.Fragment>
            <Time
              prefix='Timeout: '
              duration={{ minutes: monitor.timeoutMinutes }}
              precise
              units={['weeks', 'days', 'hours', 'minutes']}
            />
            <br />
            <CopyText title='Copy URL' value={monitor.href} asURL />
          </React.Fragment>
        ),
        secondaryAction: (
          <OtherActions
            actions={[
              {
                label: 'Edit',
                onClick: () => setShowEditDialogByID(monitor.id),
              },
              {
                label: 'Delete',
                onClick: () => setShowDeleteDialogByID(monitor.id),
              },
            ]}
          />
        ),
      }))

    return (
      <FlatList
        data-cy='monitors'
        emptyMessage='No heartbeat monitors exist for this service.'
        headerNote={HEARTBEAT_MONITOR_DESCRIPTION}
        items={items}
        headerAction={
          isMobile ? undefined : (
            <Button
              variant='contained'
              onClick={() => setShowCreateDialog(true)}
              startIcon={<Add />}
              data-testid='create-monitor'
            >
              Create Heartbeat Monitor
            </Button>
          )
        }
      />
    )
  }

  return (
    <React.Fragment>
      <Grid item xs={12} className={classes.spacing}>
        <Card>
          <CardContent>
            {renderList(data.service.heartbeatMonitors)}
          </CardContent>
        </Card>
      </Grid>
      {isMobile && (
        <CreateFAB
          onClick={() => setShowCreateDialog(true)}
          title='Create Heartbeat Monitor'
        />
      )}
      {showCreateDialog && (
        <HeartbeatMonitorCreateDialog
          serviceID={props.serviceID}
          onClose={() => setShowCreateDialog(false)}
        />
      )}
      {showEditDialogByID && (
        <HeartbeatMonitorEditDialog
          monitorID={showEditDialogByID}
          onClose={() => setShowEditDialogByID(null)}
        />
      )}
      {showDeleteDialogByID && (
        <HeartbeatMonitorDeleteDialog
          monitorID={showDeleteDialogByID}
          onClose={() => setShowDeleteDialogByID(null)}
        />
      )}
    </React.Fragment>
  )
}
