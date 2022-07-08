import React, { useState, useEffect } from 'react'
import { gql, useMutation } from 'urql'
import { FormLabel, RadioGroup, Radio } from '@mui/material'

import FormControlLabel from '@mui/material/FormControlLabel'
import FormControl from '@mui/material/FormControl'
import FormDialog from '../dialogs/FormDialog'
import { nonFieldErrors } from '../util/errutil'
import { DateTime } from 'luxon'

interface Props {
  serviceID: string
  expiresAt?: string
  onClose: () => void
}

const options = ['1 hour', '2 hours', '4 hours']

function calcExp(index: number): string {
  switch (index) {
    case 0:
      return DateTime.now().plus({ hours: 1 }).toISO()
    case 1:
      return DateTime.now().plus({ hours: 2 }).toISO()
    case 2:
      return DateTime.now().plus({ hours: 4 }).toISO()
    default:
      return ''
  }
}

function label(index: number): string {
  return `Until ${DateTime.fromISO(calcExp(index)).toFormat('t ZZZZ')}`
}

function ServiceMaintenanceForm(props: {
  onChange: (val: number) => void
  selectedIndex: number
}): JSX.Element {
  return (
    <FormControl>
      <FormLabel>
        Pause all outgoing notifications and escalations for{' '}
        {options[props.selectedIndex]}. Alerts may still be created and will
        continue as normal after maintenance mode ends.
      </FormLabel>
      <RadioGroup onChange={(e) => props.onChange(parseInt(e.target.value))}>
        <FormControlLabel value={0} control={<Radio />} label={label(0)} />
        <FormControlLabel value={1} control={<Radio />} label={label(1)} />
        <FormControlLabel value={2} control={<Radio />} label={label(2)} />
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
  const [selectedIndex, setSelectedIndex] = useState(0)
  const [updateServiceStatus, updateService] = useMutation(mutation)

  useEffect(() => {
    if (!updateServiceStatus.data) return
    props.onClose()
  }, [updateServiceStatus.data])

  return (
    <FormDialog
      maxWidth='sm'
      title='Set Maintenance Mode'
      loading={updateServiceStatus.fetching}
      errors={nonFieldErrors(updateServiceStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        updateService({
          input: {
            id: props.serviceID,
            maintenanceExpiresAt: calcExp(selectedIndex),
          },
        })
      }
      form={
        <ServiceMaintenanceForm
          onChange={(value) => setSelectedIndex(value)}
          selectedIndex={selectedIndex}
        />
      }
    />
  )
}
