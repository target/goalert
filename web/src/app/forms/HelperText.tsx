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

/** HelperText is a component that displays a hint or error message.
 *
 * It is intended to be used as the `helperText` prop of a TextField (or other MUI form components).
 */
export function HelperText(props: HelperTextProps): React.ReactNode {
  let content
  if (props.error) {
    const isMultiLine = props.error.includes('\n')
    let msg: React.ReactNode = props.error.replace(/^./, (str) =>
      str.toUpperCase(),
    )
    if (isMultiLine) {
      msg = (
        <span style={{ whiteSpace: 'pre-wrap', fontFamily: 'monospace' }}>
          {msg}
        </span>
      )
    }
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
