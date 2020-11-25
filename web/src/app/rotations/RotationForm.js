import React, { useEffect, useMemo, useState } from 'react'
import p from 'prop-types'
import { FormContainer, FormField } from '../forms'
import { TimeZoneSelect } from '../selection'
import {
  TextField,
  Grid,
  MenuItem,
  Typography,
  makeStyles,
} from '@material-ui/core'
import { startCase } from 'lodash'
import { ISODateTimePicker } from '../util/ISOPickers'
import { getNextHandoffs } from './util'
import NumberField from '../util/NumberField'
import TimeZoneSwitch from '../util/TimeZoneSwitch'
import { useResetURLParams, useURLParam } from '../actions'
import { DateTime } from 'luxon'
import ClickableText from '../util/ClickableText'
import KeyboardArrowDown from '@material-ui/icons/KeyboardArrowDown'
import KeyboardArrowUp from '@material-ui/icons/KeyboardArrowUp'

const rotationTypes = ['hourly', 'daily', 'weekly']

const useStyles = makeStyles({
  listNone: {
    listStyle: 'none',
  },
  padding0: {
    padding: 0,
  },
})
export default function RotationForm(props) {
  const { value } = props
  const classes = useStyles()
  const [zone] = useURLParam('tz', 'local')
  const resetTZ = useResetURLParams('tz')
  const localZone = useMemo(() => DateTime.local().zone.name, [])
  const [showHandoffs, setShowHandoffs] = useState(false)
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

  useEffect(() => {
    if (!isValidStart) setShowHandoffs(false)
  }, [isValidStart])

  useEffect(() => resetTZ(), [value.timeZone])

  // NOTE memoize to prevent calculation on each poll request
  const nextHandoffs = useMemo(() => {
    return isValidStart
      ? getNextHandoffs(3, value.start, value.type, value.shiftLength, zone)
      : []
  }, [value.start, value.type, value.shiftLength, zone])

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
        <Grid item xs={12}>
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
        {localZone !== props.value.timeZone && (
          <Grid item xs={12}>
            <TimeZoneSwitch option={props.value.timeZone} />
          </Grid>
        )}

        <Grid item xs={12}>
          <Grid container spacing={0}>
            <Grid item xs={12}>
              <FormField
                fullWidth
                component={ISODateTimePicker}
                label='Initial Handoff Time'
                name='start'
                min={minStart.toISO()}
                max={maxStart.toISO()}
                required
                hint={
                  <ClickableText
                    disabled={!isValidStart}
                    onClick={() => setShowHandoffs(!showHandoffs)}
                    endIcon={
                      showHandoffs ? <KeyboardArrowUp /> : <KeyboardArrowDown />
                    }
                  >
                    {`${showHandoffs ? 'Hide' : 'Preview'} upcoming handoffs`}
                  </ClickableText>
                }
              />
            </Grid>
            {showHandoffs && (
              <Grid item xs={12} component='ul' className={classes.padding0}>
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
            )}
          </Grid>
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
