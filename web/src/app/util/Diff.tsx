import React from 'react'
import { diffChars, diffWords, Change } from 'diff'
import Typography from '@material-ui/core/Typography'
import { makeStyles } from '@material-ui/core'

const useStyles = makeStyles({
  // bg color for the whole first line
  oldLine: {
    backgroundColor: '#FFEEF0',
  },
  // removed chars darker red
  removed: {
    backgroundColor: '#FCB8C0',
  },
  // bg color for the whole second line
  newLine: {
    backgroundColor: '#E6FFED',
  },
  // added chars darker green
  added: {
    backgroundColor: '#AEF1BE',
  },
  // spacing on +/- symbols in diff
  symbol: {
    minWidth: '2em',
    display: 'inline-block',
  },
})

interface DiffProps {
  oldValue: string
  newValue: string
  type: 'chars' | 'words'
}

/*
 * Diff displays a difference in characters
 * or words from an old and new value
 */
export default function Diff(props: DiffProps) {
  const { oldValue, newValue, type } = props

  const classes = useStyles()
  const { oldLine, removed, newLine, added } = classes

  let diff: Change[] = []
  if (type === 'chars') diff = diffChars(oldValue, newValue)
  if (type === 'words') diff = diffWords(oldValue, newValue)

  const oldWithRemoved = diff.map((part, idx) => {
    if (part.added) return

    return (
      <span key={idx} className={part.removed ? removed : undefined}>
        {part.value}
      </span>
    )
  })

  const newWithAdded = diff.map((part, idx) => {
    if (part.removed) return

    return (
      <span key={idx} className={part.added ? added : undefined}>
        {part.value}
      </span>
    )
  })

  const hideRemoved = diff.length === 1 && diff[0].added // net new string
  const hideAdded = diff.length === 1 && diff[0].removed // deleted whole string

  return (
    <React.Fragment>
      {!hideRemoved && (
        <Typography className={oldLine} data-cy='old'>
          <span className={classes.symbol}>&nbsp;-</span>
          {oldWithRemoved}
        </Typography>
      )}
      {!hideAdded && (
        <Typography className={newLine} data-cy='new'>
          <span className={classes.symbol}>&nbsp;+</span>
          {newWithAdded}
        </Typography>
      )}
    </React.Fragment>
  )
}
