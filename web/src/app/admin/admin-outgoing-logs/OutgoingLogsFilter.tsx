import React, { useState, useEffect } from 'react'
import { TextField } from '@mui/material'
import makeStyles from '@mui/styles/makeStyles'
import { theme } from '../../mui'

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

interface FilterValues {
  startDate?: string
  endDate?: string
}

interface OutgoingLogsFilterProps {
  onChange?: (filter: FilterValues) => void
}

export default function OutgoingLogsFilter({
  onChange,
}: OutgoingLogsFilterProps): JSX.Element {
  const classes = useStyles()

  const [value, setValue] = useState<FilterValues>({})

  useEffect(() => {
    if (onChange) onChange(value)
  }, [value])

  return (
    <div className={classes.filterContainer}>
      {/*  start */}
      <TextField
        placeholder='Start'
        name='startDate'
        onChange={(e) => setValue({ ...value, startDate: e.target.value })}
        value={value.startDate}
        className={classes.textField}
        margin='dense'
      />
      <div className={classes.spacer} />
      {/*  end */}
      <TextField
        placeholder='End'
        name='endDate'
        onChange={(e) => setValue({ ...value, endDate: e.target.value })}
        value={value.endDate}
        className={classes.textField}
        margin='dense'
      />
    </div>
  )
}
