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
        onDelete={() => onChange(value.filter((f) => f !== v))}
      />
    )
  })

  return (
    <Grid container>
      <Grid item>{props.value.length ? selected : 'No webhooks selected'}</Grid>
      <Grid container item>
        <Grid item>
          {/* input field for adding new webhook*/}
          <TextField
            variant='outlined'
            size='small'
            value={newURL}
            onChange={(e) => setNewURL(e.target.value)}
            error={!isValidURL(newURL) && newURL.length > 0}
            placeholder='https://example.com/...'
            InputProps={{
              endAdornment: (
                <Chip
                  color='primary' // for white text
                  component='button'
                  label='Add'
                  size='small'
                  icon={<Add fontSize='small' />}
                  onClick={() => {
                    if (!newURL) return
                    if (!isValidURL(newURL)) return

                    onChange([...value, newURL])
                    setNewURL('')
                  }}
                />
              ),
            }}
          />
        </Grid>
      </Grid>
    </Grid>
  )
}
