import React, { useMemo, useState } from 'react'
import p from 'prop-types'
import { FormContainer, FormField } from '../forms'
import { TimeZoneSelect } from '../selection'
import {
  TextField,
  Grid,
  MenuItem,
  Typography,
  makeStyles,
  Switch,
  FormControlLabel,
} from '@material-ui/core'
import { startCase } from 'lodash'
import { ISODateTimePicker } from '../util/ISOPickers'
import { getNextHandoffs } from './util'
import NumberField from '../util/NumberField'
import { DateTime } from 'luxon'

const rotationTypes = ['hourly', 'daily', 'weekly']

const useStyles = makeStyles({
  listNone: {
    listStyle: 'none',
  },
  bolder: {
    fontWeight: 'bolder',
  },
})
export default function RotationForm(props) {
  const { value } = props
  const classes = useStyles()
  const [configInZone, setConfigInZone] = useState(false)
  const localZone = useMemo(() => DateTime.local().zone.name, [])
  const configZone = configInZone ? value.timeZone : 'local'

  const [minStart, maxStart] = useMemo(
    () => [
      DateTime.fromObject({ year: 2000 }),
      DateTime.fromObject({ year: 2500 }),
    ],
    [],
  )

  const isValidStart = useMemo(
    () =>
      DateTime.fromISO(value.start) >= minStart &&
      DateTime.fromISO(value.start) <= maxStart,
    [value.start],
  )

  // NOTE memoize to prevent calculation on each poll request
  const nextHandoffs = useMemo(() => {
    console.log('running')
    return isValidStart
      ? getNextHandoffs(
          3,
          value.start,
          value.type,
          value.shiftLength,
          configZone,
        )
      : []
  }, [value.start, value.type, value.shiftLength, value.timeZone, configInZone])

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
        <Grid item xs={localZone === value.timeZone ? 12 : 6}>
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
        {localZone !== value.timeZone && (
          <Grid item xs={6}>
            <FormControlLabel
              control={
                <Switch
                  checked={configInZone}
                  onChange={() => setConfigInZone(!configInZone)}
                  value={value.timeZone}
                />
              }
              label={`Configure in ${value.timeZone}`}
            />
          </Grid>
        )}

        <Grid item xs={12}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            timeZone={configZone}
            label='Initial Handoff Time'
            name='start'
            min={minStart.toISO()}
            max={maxStart.toISO()}
            required
          />
        </Grid>

        <Grid item xs={12}>
          <Typography variant='body2' className={classes.bolder}>
            Upcoming Handoff times:
          </Typography>
          {nextHandoffs.map((text, i) => (
            <Typography
              key={i}
              component='li'
              className={classes.listNone}
              variant='body2'
            >
              {text}
            </Typography>
          ))}
        </Grid>
      </Grid>
    </FormContainer>
  )
}

RotationForm.propTypes = {
  value: p.shape({
    name: p.string.isRequired,
    description: p.string.isRequired,
    timeZone: p.string.isRequired,
    type: p.oneOf(rotationTypes).isRequired,
    shiftLength: p.number.isRequired,
    start: p.string.isRequired,
  }).isRequired,

  errors: p.arrayOf(
    p.shape({
      field: p.oneOf([
        'name',
        'description',
        'timeZone',
        'type',
        'start',
        'shiftLength',
      ]).isRequired,
      message: p.string.isRequired,
    }),
  ),

  onChange: p.func.isRequired,
}
