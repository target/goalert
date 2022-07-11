import React, { Fragment } from 'react'
import { gql, useQuery, useMutation } from 'urql'
import { Button } from '@mui/material'
import { DateTime } from 'luxon'
import Notices, { Notice } from '../details/Notices'

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

// todo: add error handling
export default function ServiceMaintenanceNotice({
  serviceID,
  extraNotices = [],
}: ServiceMaintenanceNoticeProps): JSX.Element {
  const [{ fetching, data }] = useQuery({
    query,
    variables: { serviceID },
    pause: !serviceID,
  })
  const [updateServiceStatus, updateService] = useMutation(mutation)
  if (
    (!data && fetching) ||
    (!updateServiceStatus.data && updateServiceStatus.fetching)
  ) {
    return <Fragment />
  }

  const exp = DateTime.fromISO(data.service.maintenanceExpiresAt)
  const isMaintMode = exp.isValid && exp > DateTime.local()
  if (!isMaintMode) return <Fragment />

  return (
    <Notices
      notices={[
        {
          type: 'WARNING',
          message: 'In Maintenance Mode',
          details: `Ends at ${exp.toFormat('FFF')}`,
          action: (
            <Button
              onClick={() => {
                updateService({
                  input: {
                    id: serviceID,
                    maintenanceExpiresAt: DateTime.local()
                      .minus({
                        years: 1,
                      })
                      .toISO(),
                  },
                })
              }}
            >
              Cancel
            </Button>
          ),
        },
        ...extraNotices,
      ]}
    />
  )
}
