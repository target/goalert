import React from 'react'
import makeStyles from '@mui/styles/makeStyles'
import { theme } from '../../mui'
import { ISODateTimePicker } from '../../util/ISOPickers'
import { useURLParam } from '../../actions'
import { Button } from '@mui/material'
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
  }
})

export default function OutgoingLogsFilter(): JSX.Element {
  const classes = useStyles()

  const [start, setStart] = useURLParam<string>('start', '')
  const [end, setEnd] = useURLParam<string>('end', '')

  // todo: make reset button reset ISODateTimePicker

  const resetFilters = (): void => {
    setStart('')
    setEnd('')
  }

  return (
    <div className={classes.filterContainer}>
      <ISODateTimePicker
        placeholder='Start'
        name='startDate'
        value={start}
        onChange={(newVal) => setStart(newVal as string)}
        className={classes.textField}
        margin='dense'
      />
      <div className={classes.spacer} />
      <ISODateTimePicker
        placeholder='End'
        name='endDate'
        value={end}
        onChange={(newVal) => setEnd(newVal as string)}
        className={classes.textField}
        margin='dense'
      />
      <Button type='button' onClick={resetFilters}>
        <RestartAltIcon />
      </Button>
    </div>
  )
}
