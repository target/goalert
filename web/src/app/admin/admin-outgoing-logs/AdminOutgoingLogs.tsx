import React from 'react'
import Grid from '@material-ui/core/Grid'
import { Typography, Box } from '@material-ui/core'
import { makeStyles } from '@material-ui/core/styles'
import { GenericError } from '../../error-pages'
import Spinner from '../../loading/components/Spinner'
import { ISOTimestamp } from '../../../schema'
import OutgoingLogCard from './OutgoingLogCard'

export interface DebugMessage {
  // will come from graphql
  id: string
  createdAt: ISOTimestamp
  updatedAt: ISOTimestamp
  type: string
  status: string
  userID?: string
  userName?: string
  source?: string
  destination: string
  serviceID?: string
  serviceName?: string
  alertID?: number
  providerID?: string
}

const mockDebugMessages: DebugMessage[] = [
  {
    id: 'f97ed4ba-4e3c-444c-a1cb-94d1faf6b2ee',
    createdAt: '2021-11-29T22:03:09.970749Z',
    updatedAt: '2021-11-29T22:04:16.754981Z',
    type: 'Alert',
    status: 'Delivered: delivered',
    userID: '00000000-0000-0000-0000-000000000001',
    userName: 'Admin McAdminFace',
    source: '+1 763-220-6186 (SMS)',
    destination: '+1 651-242-1695 (SMS)',
    serviceID: '942e18a6-3d65-4f91-baea-dc32bfd614c1',
    serviceName: 'Spencertest',
    alertID: undefined,
    providerID: 'SM1fe5a96351a84fe1b83cb787af36c5da',
  },
  {
    id: '386dfdd1-a2be-4cde-94f6-942207b247d9',
    createdAt: '2021-11-29T22:00:50.004685Z',
    updatedAt: '2021-11-29T22:00:51.595646Z',
    type: 'Alert',
    status: 'Failed (permanent): contact method disabled',
    userID: '00000000-0000-0000-0000-000000000001',
    userName: 'Admin McAdminFace',
    source: undefined,
    destination: '+1 763-302-9175 (Voice)',
    serviceID: '942e18a6-3d65-4f91-baea-dc32bfd614c1',
    serviceName: 'Spencertest',
    alertID: undefined,
    providerID: undefined,
  },
  {
    id: '2324dd1b-0387-4f1d-b50f-e7fe392bcfac',
    createdAt: '2021-11-29T21:59:49.997589Z',
    updatedAt: '2021-11-29T22:00:56.666274Z',
    type: 'Alert',
    status: 'Delivered: delivered',
    userID: '00000000-0000-0000-0000-000000000001',
    userName: 'Admin McAdminFace',
    source: '+1 763-220-6186 (SMS)',
    destination: '+1 651-242-1695 (SMS)',
    serviceID: '942e18a6-3d65-4f91-baea-dc32bfd614c1',
    serviceName: 'Spencertest',
    alertID: undefined,
    providerID: 'SMf8e76c3b36f84ad6a34355b7f36ed96b',
  },
  {
    id: 'd00b0f80-4334-4bf9-ac46-caf531973d54',
    createdAt: '2021-08-31T20:58:00.207714Z',
    updatedAt: '2021-08-31T20:58:01.445524Z',
    type: 'Alert',
    status: 'Failed (permanent): contact method disabled',
    userID: '00000000-0000-0000-0000-000000000001',
    userName: 'Admin McAdminFace',
    source: undefined,
    destination: '+1 763-302-9175 (Voice)',
    serviceID: '942e18a6-3d65-4f91-baea-dc32bfd614c1',
    serviceName: 'Spencertest',
    alertID: undefined,
    providerID: undefined,
  },
]

// const debugMessageLogsQuery = gql`
//   query ($number: String!) {
//     phoneNumberInfo(number: $number) {
//       id
//       valid
//       regionCode
//       countryCode
//       formatted
//       error
//     }
//   }
// `

const useStyles = makeStyles((theme) => ({
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

  // const { data, loading, error } = useQuery(debugMessageLogsQuery)
  const { data, loading, error } = {
    // mock data
    data: {
      debugMessages: mockDebugMessages,
    },
    loading: false,
    error: undefined as any,
  }

  if (error) {
    return <GenericError error={error.message} />
  }

  if (loading && !data) {
    return <Spinner />
  }

  const handleCardClick = (id: string): void => {
    console.log('clicked', id)
  }

  return (
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
            flexDirection='column'
            alignItems='stretch'
            width='full'
          >
            {data.debugMessages.map((debugMessage) => (
              <OutgoingLogCard
                key={debugMessage.id}
                debugMessage={debugMessage}
                onClick={() => handleCardClick(debugMessage.id)}
              />
            ))}
          </Box>
        </Grid>
      </Grid>
    </Grid>
  )
}
