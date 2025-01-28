import React from 'react'
import Typography from '@mui/material/Typography'
import {
  SentimentDissatisfied,
  SentimentVeryDissatisfied,
} from '@mui/icons-material'
import { useTheme } from '@mui/material'

interface ErrorsProps {
  error?: string
  type?: string
}

export function PageNotFound(): React.JSX.Element {
  const theme = useTheme()
  return (
    <div style={{ textAlign: 'center', color: theme.palette.text.primary }}>
      <SentimentDissatisfied style={{ height: '33vh', width: '33vw' }} />
      <Typography variant='h5'>
        Sorry, the page you were trying to reach could not be found.
      </Typography>
    </div>
  )
}

export function ObjectNotFound(props: ErrorsProps): React.JSX.Element {
  const theme = useTheme()
  return (
    <div style={{ textAlign: 'center', color: theme.palette.text.primary }}>
      <SentimentDissatisfied style={{ height: '33vh', width: '33vw' }} />
      <Typography variant='h5'>
        Sorry, the {props.type || 'thing'} you were looking for could not be
        found.
      </Typography>
      <Typography variant='caption'>
        Someone may have deleted it, or it never existed.
      </Typography>
    </div>
  )
}

export function GenericError(props: ErrorsProps): React.JSX.Element {
  const theme = useTheme()
  let errorText
  if (props.error) {
    errorText = <Typography variant='caption'>{props.error}</Typography>
  }
  return (
    <div style={{ textAlign: 'center', color: theme.palette.text.primary }}>
      <SentimentVeryDissatisfied style={{ height: '33vh', width: '33vw' }} />
      <Typography variant='h5'>Sorry, an error occurred.</Typography>
      {errorText}
    </div>
  )
}
