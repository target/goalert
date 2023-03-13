import { Add } from '@mui/icons-material'
import { Chip, Grid, TextField } from '@mui/material'
import React from 'react'
import { WebhookChip } from '../util/Chips'

type WebhookSelectProps = {
  value: Array<string>
  onChange: (value: Array<string>) => void
}

function isValidURL(str: string): boolean {
  try {
    // eslint-disable-next-line no-new
    new URL(str)
    return true
  } catch {
    return false
  }
}

export const WebhookSelect: React.FC<WebhookSelectProps> = (props) => {
  const [newURL, setNewURL] = React.useState<string>('')
  const { value, onChange = () => {} } = props

  const selected = props.value.map((v) => {
    return (
      <WebhookChip
        id={v}
        key={v}
        onDelete={() => onChange(value.filter((f) => f !== v))}
      />
    )
  })

  return (
    <Grid container spacing={1}>
      <Grid container>{props.value.length ? selected : ''}</Grid>
      <Grid container item>
        <TextField
          variant='outlined'
          fullWidth
          value={newURL}
          onChange={(e) => {
            setNewURL(e.target.value)
          }}
          error={
            (!isValidURL(newURL) && newURL.length > 0) ||
            value.indexOf(newURL) > -1
          }
          helperText={
            !isValidURL(newURL) && newURL.length > 0
              ? 'Invalid URL.'
              : value.indexOf(newURL) > -1
              ? 'Duplicate URL.'
              : ''
          }
          placeholder='https://example.com/...'
          InputProps={{
            endAdornment: (
              <Chip
                color='primary' // for white text
                component='button'
                label='Add'
                size='medium'
                icon={<Add fontSize='small' />}
                onClick={() => {
                  if (
                    !newURL ||
                    !isValidURL(newURL) ||
                    value.indexOf(newURL) > -1
                  )
                    return

                  onChange([...value, newURL])
                  setNewURL('')
                }}
              />
            ),
          }}
        />
      </Grid>
    </Grid>
  )
}
