import React, { useState, useEffect } from 'react'
import { gql, useMutation } from 'urql'
import { RadioGroup, Radio, Typography } from '@mui/material'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormControl from '@mui/material/FormControl'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'
import { DateTime } from 'luxon'
import { Time } from '../util/Time'
import { ISODateTimePicker } from '../util/ISOPickers'

interface Props {
  serviceID: string
  expiresAt?: string
  onClose: () => void
}

function label(hours: number): JSX.Element {
  return (
    <span>
      <Time duration={{ hours }} /> (
      <Time prefix='ends ' time={DateTime.local().plus({ hours }).toISO()} />)
    </span>
  )
}

function ServiceMaintenanceForm(props: {
  value: DateTime
  onChange: (val: DateTime) => void
}): JSX.Element {
  const [selectedOption, setSelectedOption] = useState(1)

  function handleOption(value: number): void {
    setSelectedOption(value)
    if (value === -1) return
    props.onChange(DateTime.local().plus({ hours: value }))
  }

  return (
    <FormControl>
      <RadioGroup
        value={selectedOption}
        onChange={(e) => handleOption(parseInt(e.target.value, 10))}
      >
        <FormControlLabel value={1} control={<Radio />} label={label(1)} />
        <FormControlLabel value={2} control={<Radio />} label={label(2)} />
        <FormControlLabel value={4} control={<Radio />} label={label(4)} />
        <FormControlLabel value={-1} control={<Radio />} label='Specify' />
      </RadioGroup>
      <ISODateTimePicker
        value={props.value.toISO()}
        disabled={selectedOption !== -1}
        onChange={(iso) => props.onChange(DateTime.fromISO(iso))}
        min={DateTime.local().plus({ hours: 1 }).toISO()}
        max={DateTime.local().plus({ hours: 24 }).toISO()}
        sx={{
          marginLeft: (theme) => theme.spacing(3.75),
          marginTop: (theme) => theme.spacing(1),
        }}
      />
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
  const [endTime, setEndTime] = useState(DateTime.local().plus({ hours: 1 }))
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
        <Typography>
          Pause all outgoing notifications and escalations for{' '}
          <Time
            duration={{
              hours: Math.ceil(endTime.diffNow('hours').hours),
            }}
          />
          . Incoming alerts will still be created and will continue as normal
          after maintenance mode ends.
        </Typography>
      }
      loading={updateServiceStatus.fetching}
      errors={nonFieldErrors(updateServiceStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        updateService(
          {
            input: {
              id: props.serviceID,
              maintenanceExpiresAt: endTime.toISO(),
            },
          },
          {
            additionalTypenames: ['Service'],
          },
        )
      }
      form={
        <ServiceMaintenanceForm
          onChange={(value) => setEndTime(value)}
          value={endTime}
        />
      }
    />
  )
}
