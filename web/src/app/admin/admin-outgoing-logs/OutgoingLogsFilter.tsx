import React from 'react'
import makeStyles from '@mui/styles/makeStyles'
import { theme } from '../../mui'
import { ISODateTimePicker } from '../../util/ISOPickers'

const useStyles = makeStyles<typeof theme>((theme) => {
  return {
    filterContainer: {
      display: 'flex',
      flexDirection: 'row',
    },
    spacer: {
      width: '9px',
    },
    textField: {
      backgroundColor: 'white',
      borderRadius: '4px',
      [theme.breakpoints.down('sm')]: {
        flex: 1,
      },
      [theme.breakpoints.up('md')]: {
        minWidth: 250,
        '& input:not(:placeholder-shown)': {
          minWidth: 275,
        },
        '& input': {
          minWidth: 180,
          transitionProperty: 'min-width',
          transitionDuration: theme.transitions.duration.standard,
          transitionTimingFunction: theme.transitions.easing.easeInOut,
        },
      },
    },
  }
})

export interface FilterValues {
  startDate?: string
  endDate?: string
}

interface OutgoingLogsFilterProps {
  value: FilterValues
  onChange: (filter: FilterValues) => void
}

export default function OutgoingLogsFilter({
  value,
  onChange,
}: OutgoingLogsFilterProps): JSX.Element {
  const classes = useStyles()

  return (
    <div className={classes.filterContainer}>
      <ISODateTimePicker
        placeholder='Start'
        name='startDate'
        // timeZone={zone} // todo?
        onChange={(newVal: any) => onChange({ ...value, startDate: newVal })}
        value={value.startDate}
        className={classes.textField}
        margin='dense'
      />
      <div className={classes.spacer} />
      {/*  end */}
      <ISODateTimePicker
        placeholder='End'
        name='endDate'
        // timeZone={zone} // todo?
        onChange={(newVal: any) => onChange({ ...value, startDate: newVal })}
        value={value.endDate}
        className={classes.textField}
        margin='dense'
      />
    </div>
  )
}
