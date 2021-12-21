import React, { useState } from 'react'
import { useQuery, gql } from '@apollo/client'
import { Box, Grid, Typography } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import OutgoingLogsList from './OutgoingLogsList'
import Search from '../../util/Search'
import OutgoingLogsFilter, { FilterValues } from './OutgoingLogsFilter'
import OutgoingLogDetails from './OutgoingLogDetails'
import { theme } from '../../mui'
import { DebugMessage } from '../../../schema'
import { useURLParam } from '../../actions'

const debugMessageLogsQuery = gql`
  query debugMessageLogsQuery {
    debugMessages {
      id
      createdAt
      updatedAt
      type
      status
      userID
      userName
      source
      destination
      serviceID
      serviceName
      alertID
      providerID
    }
  }
`

const useStyles = makeStyles<typeof theme>((theme) => ({
  gridContainer: {
    [theme.breakpoints.up('md')]: {
      justifyContent: 'center',
    },
  },
  groupTitle: {
    fontSize: '1.1rem',
  },
  saveDisabled: {
    color: 'rgba(255, 255, 255, 0.5)',
  },
  card: {
    margin: theme.spacing(1),
    cursor: 'pointer',
  },
}))

export default function AdminOutgoingLogs(): JSX.Element {
  const classes = useStyles()
  const [selectedLog, setSelectedLog] = useState<DebugMessage | null>(null)
  const [filter, setFilter] = useState<FilterValues>({})

  const { data, loading, error } = useQuery(debugMessageLogsQuery)
  const [searchParam] = useURLParam('search', '')

  if (error) return <GenericError error={error.message} />
  if (loading && !data) return <Spinner />

  return (
    <React.Fragment>
      <OutgoingLogDetails
        open={Boolean(selectedLog)}
        onClose={() => setSelectedLog(null)}
        log={selectedLog}
      />
      <Grid container spacing={2} className={classes.gridContainer}>
        <Grid container item xs={12}>
          <Grid item xs={12}>
            <Typography
              component='h2'
              variant='subtitle1'
              color='textSecondary'
              classes={{ subtitle1: classes.groupTitle }}
            >
              Outgoing Messages
            </Typography>
          </Grid>
          <Grid item xs={12}>
            <Box
              display='flex'
              flexDirection='row'
              alignItems='center'
              justifyContent='space-between'
            >
              <div>
                <OutgoingLogsFilter value={filter} onChange={setFilter} />
              </div>
              <div>
                <Search />
              </div>
            </Box>
          </Grid>
          <Grid item xs={12}>
            {data.debugMessages ? (
              <OutgoingLogsList
                filter={filter}
                searchTerm={searchParam}
                debugMessages={data.debugMessages}
                onSelect={setSelectedLog}
              />
            ) : null}
          </Grid>
        </Grid>
      </Grid>
    </React.Fragment>
  )
}
