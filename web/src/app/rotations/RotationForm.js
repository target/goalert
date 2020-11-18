import React, { useMemo } from 'react'
import p from 'prop-types'
import { FormContainer, FormField } from '../forms'
import { TimeZoneSelect } from '../selection'
import {
  TextField,
  Grid,
  MenuItem,
  List,
  ListItem,
  ListItemText,
  makeStyles,
} from '@material-ui/core'
import { startCase } from 'lodash'
import { ISODateTimePicker } from '../util/ISOPickers'
import { getNextHandoffs } from './util'
import NumberField from '../util/NumberField'

const rotationTypes = ['hourly', 'daily', 'weekly']

const useStyles = makeStyles({
  noVerticalSpace: {
    marginTop: 0,
    marginBottom: 0,
    paddingTop: 0,
    paddingBottom: 0,
  },
  semiBold: {
    fontWeight: 600,
  },
})
export default function RotationForm(props) {
  const { value } = props
  const classes = useStyles()

  // NOTE memoize to prevent calculation on each poll request
  const nextHandoffs = useMemo(
    () =>
      getNextHandoffs(
        3,
        value.start,
        value.type,
        value.shiftLength,
        value.timeZone,
      ),
    [value.start, value.type, value.shiftLength, value.timeZone],
  )

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
        <Grid item xs={6}>
          <FormField
            fullWidth
            component={ISODateTimePicker}
            label='Initial Handoff Time'
            name='start'
            required
          />
        </Grid>
        <Grid item xs={6}>
          <List
            dense
            disablePadding
            subheader={
              <ListItem disableGutters className={classes.noVerticalSpace}>
                <ListItemText
                  primary='Upcoming Handoff Times:'
                  primaryTypographyProps={{ className: classes.semiBold }}
                />
              </ListItem>
            }
          >
            {nextHandoffs.map((text, i) => {
              return (
                <ListItem
                  key={i}
                  className={classes.noVerticalSpace}
                  disableGutters
                >
                  <ListItemText
                    primary={text}
                    className={classes.noVerticalSpace}
                  />
                </ListItem>
              )
            })}
          </List>
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
