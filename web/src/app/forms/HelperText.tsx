import React from 'react'
import FormHelperText from '@mui/material/FormHelperText'
import AppLink from '../util/AppLink'
import { Grid } from '@mui/material'

export type HelperTextProps = {
  hint?: string
  hintURL?: string
  error?: string
  errorURL?: string
  maxLength?: number
  value?: string
}

export function HelperText(props: HelperTextProps) {
  let content
  if (props.error) {
    const msg = props.error.replace(/^./, (str) => str.toUpperCase())
    content = props.errorURL ? (
      <AppLink to={props.errorURL} newTab data-cy='error-help-link'>
        {msg}
      </AppLink>
    ) : (
      msg
    )
  } else {
    content = props.hintURL ? (
      <AppLink to={props.hintURL} newTab data-cy='hint-help-link'>
        {props.hint}
      </AppLink>
    ) : (
      props.hint
    )
  }

  if (props.maxLength) {
    return (
      <Grid container spacing={2}>
        <Grid item xs={10}>
          <FormHelperText component='span'>{content}</FormHelperText>
        </Grid>
        <Grid item xs={2}>
          <FormHelperText style={{ textAlign: 'right' }}>
            {props.value?.length || 0}/{props.maxLength}
          </FormHelperText>
        </Grid>
      </Grid>
    )
  }

  return <FormHelperText component='span'>{content}</FormHelperText>
}
