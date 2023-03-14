import { Add } from '@mui/icons-material'
import { Chip, Grid, TextField } from '@mui/material'
import React from 'react'
import { WebhookChip } from '../util/Chips'

type ChanWebhookSelectProps = {
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

export const ChanWebhookSelect: React.FC<ChanWebhookSelectProps> = (props) => {
  const [newURL, setNewURL] = React.useState<string>('')
  const { value, onChange = () => {} } = props

  const selected = props.value.map((v) => {
    return (
      <Grid item key={v}>
        <WebhookChip
          id={v}
          onDelete={() => onChange(value.filter((f) => f !== v))}
        />
      </Grid>
    )
  })

  return (
    <Grid container spacing={1}>
      <Grid container spacing={1}>
        {props.value.length ? selected : ''}
      </Grid>
      <Grid container item>
        <TextField
          variant='outlined'
          fullWidth
          value={newURL}
          onChange={(e) => {
            setNewURL(e.target.value.trim())
          }}
          error={
            (!isValidURL(newURL) && newURL.length > 0) ||
            value.indexOf(newURL) > -1
          }
          helperText={
            !isValidURL(newURL) && newURL.length > 0
              ? 'Please provide a valid URL.'
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
