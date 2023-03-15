import React from 'react'
import { gql, useQuery, useMutation } from 'urql'
import { Button, Grid } from '@mui/material'
import { DateTime } from 'luxon'
import Notices, { Notice } from '../details/Notices'
import { Time } from '../util/Time'

const query = gql`
  query serviceMaintenanceQuery($serviceID: ID!) {
    service(id: $serviceID) {
      maintenanceExpiresAt
    }
  }
`

const mutation = gql`
  mutation updateService($input: UpdateServiceInput!) {
    updateService(input: $input)
  }
`

interface ServiceMaintenanceNoticeProps {
  serviceID: string
  extraNotices?: Array<Notice>
}

// assumed that this is rendered within a Grid container
export default function ServiceMaintenanceNotice({
  serviceID,
  extraNotices = [],
}: ServiceMaintenanceNoticeProps): JSX.Element | null {
  const [, updateService] = useMutation(mutation)
  const [{ fetching, data }] = useQuery({
    query,
    variables: { serviceID },
    pause: !serviceID,
  })

  const maintMode = data?.service?.maintenanceExpiresAt
  if ((!data && fetching) || !maintMode) {
    return null
  }

  return (
    <Grid item sx={{ width: '100%' }}>
      <Notices
        notices={[
          {
            type: 'WARNING',
            message: 'In Maintenance Mode',
            details: (
              <React.Fragment>
                Ends <Time format='relative' time={maintMode} precise />
              </React.Fragment>
            ),
            action: (
              <Button
                aria-label='Cancel Maintenance Mode'
                onClick={() => {
                  updateService(
                    {
                      input: {
                        id: serviceID,
                        maintenanceExpiresAt: DateTime.local()
                          .minus({
                            years: 1,
                          })
                          .toISO(),
                      },
                    },
                    { additionalTypenames: ['Service'] },
                  )
                }}
              >
                Cancel
              </Button>
            ),
          },
          ...extraNotices,
        ]}
      />
    </Grid>
  )
}
