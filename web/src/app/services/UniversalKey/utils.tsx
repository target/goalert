import React from 'react'
import { Notice } from '../../details/Notices'
import { FormControlLabel, Checkbox } from '@mui/material'

export function getNotice(
  showNotice: boolean,
  hasConfirmed: boolean,
  setHasConfirmed: (b: boolean) => void,
): Notice[] {
  if (!showNotice) return []

  return [
    {
      type: 'WARNING',
      message: 'No actions',
      details:
        'If you submit with no actions created, nothing will happen on this step',
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
