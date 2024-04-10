import { Box, Button, Typography } from '@mui/material'
import React from 'react'
import { ActionInput, DestinationTypeInfo } from '../../schema'
import { useDynamicActionTypes } from '../util/RequireConfig'
import MoreHorizIcon from '@mui/icons-material/MoreHoriz'
import RuleEditorActionDialog from './RuleEditorActionDialog'
import DestinationInputChip from '../util/DestinationInputChip'

export const makeDefaultAction = (t: DestinationTypeInfo): ActionInput => ({
  dest: {
    type: t.type,
    values: [{ fieldID: 'email-address', value: 'foo@example.com' }],
  },
  params: (t.dynamicParams || []).map((p) => ({
    paramID: p.paramID,
    expr: 'body.' + p.paramID,
  })),
})

export type RuleEditorActionsManagerProps = {
  value: ActionInput[]
  onChange: (value: ActionInput[]) => void
}

export default function RuleEditorActionsManager(
  props: RuleEditorActionsManagerProps,
): React.ReactNode {
  const actTypes = useDynamicActionTypes()
  const [editActionIndex, setEditActionIndex] = React.useState<number | null>(
    null,
  )
  const actionLabel = (a: ActionInput): string =>
    actTypes.find((t) => t.type === a.dest.type)?.name || a.dest.type

  return (
    <React.Fragment>
      {editActionIndex !== null && (
        <RuleEditorActionDialog
          action={props.value[editActionIndex]}
          onClose={(newAction) => {
            setEditActionIndex(null)
            if (newAction === null) return

            props.onChange([
              ...props.value.slice(0, editActionIndex),
              newAction,
              ...props.value.slice(editActionIndex + 1),
            ])
          }}
        />
      )}
      <Box
        sx={{
          borderRadius: 1,
          bgcolor: 'secondary.dark',
          padding: '16px',
        }}
      >
        <Typography variant='h6' component='div'>
          Actions{' '}
          <Button
            onClick={() => {
              const newActionIndex = props.value.length
              props.onChange([...props.value, makeDefaultAction(actTypes[0])])
              setEditActionIndex(newActionIndex)
            }}
          >
            Add Action
          </Button>
        </Typography>
        {props.value.length === 0 && (
          <Typography color='textSecondary'>
            <Box
              sx={{
                borderRadius: 1,
                padding: '0px 8px',
                justifyContent: 'space-between',
                display: 'flex',
                alignItems: 'center',
              }}
            >
              <Typography>-- No Action/Drop Request --</Typography>
            </Box>
          </Typography>
        )}
        {props.value.map((a, i) => (
          <Typography key={i} color='textSecondary'>
            <Box
              sx={{
                borderRadius: 1,
                padding: '0px 8px',
                justifyContent: 'space-between',
                display: 'flex',
                alignItems: 'center',
              }}
            >
              <DestinationInputChip value={a.dest} />
              <Button
                onClick={() => setEditActionIndex(i)}
                endIcon={<MoreHorizIcon />}
              />
              <Button
                onClick={() =>
                  props.onChange(props.value.filter((_, j) => j !== i))
                }
              >
                Delete Action
              </Button>
            </Box>
          </Typography>
        ))}
      </Box>
    </React.Fragment>
  )
}
