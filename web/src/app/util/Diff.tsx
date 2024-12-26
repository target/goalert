import React from 'react'
import { diffChars, diffWords, Change } from 'diff'
import Typography from '@mui/material/Typography'
import { green, red, lightGreen } from '@mui/material/colors'
import { useTheme } from '@mui/material'

interface DiffProps {
  oldValue: string
  newValue: string
  type: 'chars' | 'words'
}

/*
 * Diff displays a difference in characters
 * or words from an old and new value
 */
export default function Diff(props: DiffProps): React.JSX.Element {
  const { oldValue, newValue, type } = props
  const theme = useTheme()

  const [oldLine, newLine, removed, added] =
    theme.palette.mode === 'dark'
      ? [red[900] + '50', green[900] + '50', red[600] + '90', green[600] + '90']
      : [red[100], lightGreen[100], red.A100, lightGreen.A400]

  const symbol = { minWidth: '2em', display: 'inline-block' }

  let diff: Change[] = []
  if (type === 'chars') diff = diffChars(oldValue, newValue)
  if (type === 'words') diff = diffWords(oldValue, newValue)

  const oldWithRemoved = diff.map((part, idx) => {
    if (part.added) return

    return (
      <Typography
        key={idx}
        sx={{ bgcolor: part.removed ? removed : undefined }}
        component='span'
      >
        {part.value}
      </Typography>
    )
  })

  const newWithAdded = diff.map((part, idx) => {
    if (part.removed) return

    return (
      <Typography
        key={idx}
        sx={{ bgcolor: part.added ? added : undefined }}
        component='span'
      >
        {part.value}
      </Typography>
    )
  })

  const hideRemoved = diff.length === 1 && diff[0].added // net new string
  const hideAdded = diff.length === 1 && diff[0].removed // deleted whole string

  return (
    <React.Fragment>
      {!hideRemoved && (
        <Typography sx={{ bgcolor: oldLine }} data-cy='old'>
          <Typography sx={symbol} component='span'>
            &nbsp;-
          </Typography>
          {oldWithRemoved}
        </Typography>
      )}
      {!hideAdded && (
        <Typography sx={{ bgcolor: newLine }} data-cy='new'>
          <Typography sx={symbol} component='span'>
            &nbsp;+
          </Typography>
          {newWithAdded}
        </Typography>
      )}
    </React.Fragment>
  )
}
