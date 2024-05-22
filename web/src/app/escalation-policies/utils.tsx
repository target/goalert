import React from 'react'
import { Notice } from '../details/Notices'
import { FormControlLabel, Checkbox } from '@mui/material'

export function getNotice(
  hasSubmitted: boolean,
  hasConfirmed: boolean,
  setHasConfirmed: (b: boolean) => void,
): Notice[] {
  if (!hasSubmitted) return []

  return [
    {
      type: 'WARNING',
      message: 'No actions',
      details:
        'If you submit with no destinations, nothing will happen on this step',
      action: (
        <FormControlLabel
          control={
            <Checkbox
              checked={hasConfirmed}
              onChange={() => setHasConfirmed(!hasConfirmed)}
            />
          }
          label='I acknowledge the impact of this'
        />
      ),
    },
  ]
}
