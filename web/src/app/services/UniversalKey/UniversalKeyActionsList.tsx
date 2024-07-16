import React from 'react'
import { ActionInput } from '../../../schema'
import DestinationInputChip from '../../util/DestinationInputChip'
import { Grid, Typography } from '@mui/material'

export type UniversalKeyActionsListProps = {
  actions: ReadonlyArray<ActionInput>

  noEdit?: boolean // disables onDelete and onChipClick
  onDelete?: (action: ActionInput) => void
  onChipClick?: (action: ActionInput) => void
}

export default function UniversalKeyActionsList(
  props: UniversalKeyActionsListProps,
): React.ReactNode {
  return (
    <React.Fragment>
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
                props.onDelete && !props.noEdit
                  ? () => props.onDelete && props.onDelete(a)
                  : undefined
              }
              onChipClick={
                props.onChipClick && !props.noEdit
                  ? () => props.onChipClick && props.onChipClick(a)
                  : undefined
              }
            />
          </Grid>
        ))}
      </Grid>
      {props.actions.length === 0 && (
        <Grid item xs={12} data-testid='actions-list'>
          <Typography
            variant='body2'
            color='textSecondary'
            data-testid='no-actions'
          >
            No actions
          </Typography>
        </Grid>
      )}
    </React.Fragment>
  )
}
