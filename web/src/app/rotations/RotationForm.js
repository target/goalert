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
  Switch,
  FormControlLabel,
} from '@material-ui/core'
import { startCase } from 'lodash'
import { ISODateTimePicker } from '../util/ISOPickers'
import NumberField from '../util/NumberField'
import { DateTime } from 'luxon'
import { useQuery, gql } from '@apollo/client'
import { getHours } from './util'

const query = gql`
  query upcomingHandoffTimes($input: UpcomingHandoffTimesInput) {
    upcomingHandoffTimes(input: $input)
  }
`

const rotationTypes = ['hourly', 'daily', 'weekly']

const useStyles = makeStyles({
  listNone: {
    listStyle: 'none',
  },
  bolder: {
    fontWeight: 'bolder',
  },
  flex: {
    display: 'flex',
  },
})

export default function RotationForm(props) {
  const { value } = props
  const classes = useStyles()
  const localZone = useMemo(() => DateTime.local().zone.name, [])
  const [configInZone, setConfigInZone] = useState(false)
  const configZone = configInZone ? value.timeZone : 'local'
  const [upcomingHandoffs, setUpcomingHandoffs] = useState([])

  const [minStart, maxStart] = useMemo(
    () => [
      DateTime.fromObject({ year: 2000 }),
      DateTime.fromObject({ year: 2500 }),
    ],
    [],
  )

  const handoffsToShow = 3
  const { data } = useQuery(query, {
    variables: {
      input: {
        start: value.start,
        timeZone: value.timeZone,
        hours: getHours(value.shiftLength, value.type),
        count: handoffsToShow,
      },
    },
  })

  useEffect(() => {
    if (data?.upcomingHandoffTimes) {
      const upcomingHandoffTimes = data.upcomingHandoffTimes.map((iso) =>
        DateTime.fromISO(iso)
          .setZone(configZone)
          .toLocaleString(DateTime.DATETIME_FULL),
      )
      setUpcomingHandoffs(upcomingHandoffTimes)
    }
  }, [configInZone, data])

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
          <Grid item xs={6} className={classes.flex}>
            <Grid container justify='center'>
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
          {upcomingHandoffs.map((text, i) => (
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
