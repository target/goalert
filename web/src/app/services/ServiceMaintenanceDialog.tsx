import React, { useState, useEffect } from 'react'
import { gql, useMutation } from 'urql'
import { RadioGroup, Radio } from '@mui/material'

import FormControlLabel from '@mui/material/FormControlLabel'
import FormControl from '@mui/material/FormControl'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'
import { DateTime } from 'luxon'
import { Time } from '../util/Time'

interface Props {
  serviceID: string
  expiresAt?: string
  onClose: () => void
}

function label(hours: number): JSX.Element {
  return (
    <span>
      For <Time duration={{ hours }} /> (
      <Time prefix='ends ' time={DateTime.local().plus({ hours }).toISO()} />)
    </span>
  )
}

function ServiceMaintenanceForm(props: {
  onChange: (val: number) => void
  selectedIndex: number
}): JSX.Element {
  return (
    <FormControl>
      <RadioGroup
        value={props.selectedIndex}
        onChange={(e) => props.onChange(parseInt(e.target.value, 10))}
      >
        <FormControlLabel value={1} control={<Radio />} label={label(1)} />
        <FormControlLabel value={2} control={<Radio />} label={label(2)} />
        <FormControlLabel value={4} control={<Radio />} label={label(4)} />
      </RadioGroup>
    </FormControl>
  )
}

const mutation = gql`
  mutation updateService($input: UpdateServiceInput!) {
    updateService(input: $input)
  }
`

export default function ServiceMaintenanceModeDialog(
  props: Props,
): JSX.Element {
  const [selectedHours, setSelectedHours] = useState(1)
  const [updateServiceStatus, updateService] = useMutation(mutation)

  useEffect(() => {
    if (!updateServiceStatus.data) return
    props.onClose()
  }, [updateServiceStatus.data])

  return (
    <FormDialog
      maxWidth='sm'
      title='Set Maintenance Mode'
      subTitle={
        <React.Fragment>
          Pause all outgoing notifications and escalations for{' '}
          <Time duration={{ hours: selectedHours }} />. Incoming alerts will
          still be created and will continue as normal after maintenance mode
          ends.
        </React.Fragment>
      }
      loading={updateServiceStatus.fetching}
      errors={nonFieldErrors(updateServiceStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        updateService(
          {
            input: {
              id: props.serviceID,
              maintenanceExpiresAt: DateTime.local()
                .plus({ hours: selectedHours })
                .toISO(),
            },
          },
          {
            additionalTypenames: ['Service'],
          },
        )
      }
      form={
        <ServiceMaintenanceForm
          onChange={(value) => setSelectedHours(value)}
          selectedIndex={selectedHours}
        />
      }
    />
  )
}
