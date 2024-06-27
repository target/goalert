import React from 'react'
import { ActionInput } from '../../../schema'
import DestinationInputChip from '../../util/DestinationInputChip'
import { Grid, Typography } from '@mui/material'

export type UniversalKeyActionsListProps = {
  actions: ReadonlyArray<ActionInput>

  noHeader?: boolean
  onDelete?: (action: ActionInput) => void
  onChipClick?: (action: ActionInput) => void
}

export default function UniversalKeyActionsList(
  props: UniversalKeyActionsListProps,
): React.ReactNode {
  return (
    <React.Fragment>
      {!props.noHeader && (
        <Grid item xs={12}>
          <Typography variant='h6' color='textPrimary'>
            Actions
          </Typography>
        </Grid>
      )}
      <Grid
        item
        xs={12}
        container
        spacing={1}
        sx={{ p: 1 }}
        data-testid='actions-list'
      >
        {props.actions.map((a) => (
          <Grid item key={JSON.stringify(a.dest)}>
            <DestinationInputChip
              value={a.dest}
              onDelete={
                props.onDelete
                  ? () => props.onDelete && props.onDelete(a)
                  : undefined
              }
              onChipClick={
                props.onChipClick
                  ? () => props.onChipClick && props.onChipClick(a)
                  : undefined
              }
            />
          </Grid>
        ))}
        {props.actions.length === 0 && (
          <Grid item xs={12}>
            <Typography
              variant='body2'
              color='textSecondary'
              data-testid='no-actions'
            >
              No actions
            </Typography>
          </Grid>
        )}
      </Grid>
    </React.Fragment>
  )
}
