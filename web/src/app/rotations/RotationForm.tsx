import React, { Suspense } from 'react'
import { TextField, Grid, MenuItem, FormHelperText } from '@mui/material'
import { startCase } from 'lodash'
import { DateTime } from 'luxon'
import { FormContainer, FormField } from '../forms'
import { TimeZoneSelect } from '../selection'
import { ISODateTimePicker } from '../util/ISOPickers'
import NumberField from '../util/NumberField'
import { FieldError } from '../util/errutil'
import { CreateRotationInput } from '../../schema'
import { Time } from '../util/Time'
import RotationFormHandoffTimes from './RotationFormHandoffTimes'
import Spinner from '../loading/components/Spinner'

interface RotationFormProps {
  value: CreateRotationInput
  errors: FieldError[]
  onChange: (value: CreateRotationInput) => void
  disabled?: boolean
}

const rotationTypes = ['hourly', 'daily', 'weekly', 'monthly']

const sameAsLocal = (t: string, z: string): boolean => {
  const inZone = DateTime.fromISO(t, { zone: z })
  const inLocal = DateTime.fromISO(t, { zone: 'local' })
  return inZone.toFormat('Z') === inLocal.toFormat('Z')
}

export default function RotationForm(props: RotationFormProps): React.JSX.Element {
  const { value } = props

  const handoffWarning =
    DateTime.fromISO(value.start).day > 28 && value.type === 'monthly'

  return (
    <FormContainer optionalLabels {...props}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            name='name'
            label='Name'
            required
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TextField}
            multiline
            name='description'
            label='Description'
            required
          />
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={TextField}
            select
            required
            label='Rotation Type'
            name='type'
          >
            {rotationTypes.map((type) => (
              <MenuItem value={type} key={type}>
                {startCase(type)}
              </MenuItem>
            ))}
          </FormField>
        </Grid>
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={NumberField}
            required
            type='number'
            name='shiftLength'
            label='Shift Length'
            min={1}
            max={9000}
          />
        </Grid>
        <Grid item xs={12}>
          <FormField
            fullWidth
            component={TimeZoneSelect}
            multiline
            name='timeZone'
            fieldName='timeZone'
            label='Time Zone'
            required
          />
        </Grid>

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            timeZone={value.timeZone}
            label={`Handoff Time (${value.timeZone})`}
            name='start'
            required
            hint={
              <React.Fragment>
                {sameAsLocal(value.start, value.timeZone) ? undefined : (
                  <Time time={value.start} suffix=' local time' />
                )}
                {handoffWarning && (
                  <FormHelperText
                    sx={{
                      color: (theme) => theme.palette.warning.main,
                      marginLeft: 0,
                      marginRight: 0,
                    }}
                    data-cy='handoff-warning'
                  >
                    Unintended handoff behavior may occur when date starts after
                    the 28th
                  </FormHelperText>
                )}
              </React.Fragment>
            }
          />
        </Grid>
        <Suspense
          fallback={
            <Grid item xs={12} sx={{ height: '7rem' }}>
              <Spinner text='Calculating...' />
            </Grid>
          }
        >
          <RotationFormHandoffTimes value={value} />
        </Suspense>
      </Grid>
    </FormContainer>
  )
}
