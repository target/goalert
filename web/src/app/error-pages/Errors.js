import React from 'react'
import Typography from '@mui/material/Typography'
import {
  SentimentDissatisfied,
  SentimentVeryDissatisfied,
} from '@mui/icons-material'
import { useTheme } from '@mui/material'
import p from 'prop-types'

export function PageNotFound() {
  const theme = useTheme()
  return (
    <div style={{ textAlign: 'center' }}>
      <SentimentDissatisfied
        sx={{
          height: '33vh',
          width: '33vw',
          color: theme.palette.text.primary,
        }}
      />
      <Typography color='textPrimary' variant='h5'>
        Sorry, the page you were trying to reach could not be found.
      </Typography>
    </div>
  )
}

export function ObjectNotFound(props) {
  const theme = useTheme()
  console.log(theme.palette.text.primary)
  return (
    <div style={{ textAlign: 'center' }}>
      <SentimentDissatisfied
        sx={{
          height: '33vh',
          width: '33vw',
          color: theme.palette.text.primary,
        }}
      />
      <Typography color='textPrimary' variant='h5'>
        Sorry, the {props.type || 'thing'} you were looking for could not be
        found.
      </Typography>
      <Typography color='textPrimary' variant='caption'>
        Someone may have deleted it, or it never existed.
      </Typography>
    </div>
  )
}
ObjectNotFound.propTypes = {
  type: p.string,
}

export function GenericError(props) {
  let errorText
  if (props.error) {
    errorText = <Typography variant='caption'>{props.error}</Typography>
  }
  return (
    <div style={{ textAlign: 'center' }}>
      <SentimentVeryDissatisfied style={{ height: '33vh', width: '33vw' }} />
      <Typography variant='h5'>Sorry, an error occurred.</Typography>
      {errorText}
    </div>
  )
}

GenericError.propTypes = {
  error: p.string,
}
