import React from 'react'
import { ActionInput } from '../../../schema'
import DestinationInputChip from '../../util/DestinationInputChip'
import { Chip, Grid } from '@mui/material'
import { Warning } from '../../icons'

export type UniversalKeyActionsListProps = {
  actions: ReadonlyArray<ActionInput>

  noEdit?: boolean // disables onDelete and onChipClick
  onDelete?: (action: ActionInput) => void
  onChipClick?: (action: ActionInput) => void
  onEdit?: (index: number) => void
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
        {props.actions.map((a, idx) => (
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
              onEdit={
                props.onEdit && !props.noEdit
                  ? () => props.onEdit && props.onEdit(idx)
                  : undefined
              }
            />
          </Grid>
        ))}
      </Grid>
      {props.actions.length === 0 && (
        <Grid item xs={12} data-testid='actions-list'>
          <Chip
            label='No actions'
            icon={
              <div style={{ padding: '4px' }}>
                <Warning
                  placement='bottom'
                  message='With no actions configured, nothing will happen when this rule matches'
                />
              </div>
            }
          />
        </Grid>
      )}
    </React.Fragment>
  )
}
