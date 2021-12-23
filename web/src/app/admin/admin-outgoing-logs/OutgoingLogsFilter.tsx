import React, { useState } from 'react'
import makeStyles from '@mui/styles/makeStyles'
import { theme } from '../../mui'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { useResetURLParams, useURLParam } from '../../actions'
import { IconButton } from '@mui/material'
import RestartAltIcon from '@mui/icons-material/RestartAlt'

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
    resetButton: {
      height: 'min-content',
      alignSelf: 'center',
      marginLeft: theme.spacing(1),
    },
  }
})

export default function OutgoingLogsFilter(): JSX.Element {
  const classes = useStyles()

  const [start, setStart] = useURLParam<string>('start', '')
  const [end, setEnd] = useURLParam<string>('end', '')
  const [key, setKey] = useState(0)
  const resetDateRange = useResetURLParams('start', 'end')

  const resetFilters = (): void => {
    resetDateRange()
    setKey(key + 1)
  }

  return (
    <div className={classes.filterContainer} key={key}>
      <div>
        <ISODateTimePicker
          placeholder='Start'
          name='startDate'
          value={start}
          onChange={(newVal) => setStart(newVal as string)}
          className={classes.textField}
          label='Created after'
          margin='dense'
          size='small'
          variant='filled'
        />
        <div className={classes.spacer} />
        <ISODateTimePicker
          placeholder='End'
          name='endDate'
          value={end}
          label='Created before'
          onChange={(newVal) => setEnd(newVal as string)}
          className={classes.textField}
          margin='dense'
          size='small'
          variant='filled'
        />
      </div>

      <IconButton
        className={classes.resetButton}
        type='button'
        onClick={resetFilters}
      >
        <RestartAltIcon />
      </IconButton>
    </div>
  )
}
