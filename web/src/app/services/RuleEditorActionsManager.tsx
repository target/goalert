import {
  Box,
  Button,
  Fade,
  IconButton,
  Menu,
  MenuItem,
  Typography,
} from '@mui/material'
import React from 'react'
import { ActionInput, DestinationTypeInfo } from '../../schema'
import { useDynamicActionTypes } from '../util/RequireConfig'
import MoreHorizIcon from '@mui/icons-material/MoreHoriz'
import RuleEditorActionDialog from './RuleEditorActionDialog'
import DestinationInputChip from '../util/DestinationInputChip'
import { Warning } from '../icons'
import OtherActions from '../util/OtherActions'

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
  // default is true if this is the default action
  default?: boolean
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
          // bgcolor: 'secondary.dark',
          padding: '16px',
          outline: 'solid',
          outlineWidth: '1px',
        }}
      >
        <Box display='flex' justifyContent='space-between'>
          <Typography variant={props.default ? 'h5' : 'h6'} component='div'>
            {props.default ? 'Default Actions ' : 'Actions '}
          </Typography>
          <Button
            onClick={() => {
              const newActionIndex = props.value.length
              props.onChange([...props.value, makeDefaultAction(actTypes[0])])
              setEditActionIndex(newActionIndex)
            }}
            variant='contained'
            size='small'
          >
            Add Action
          </Button>
        </Box>

        {props.default && (
          <Typography
            sx={{ paddingLeft: '1em', fontStyle: 'italic', pr: 2 }}
            color='textSecondary'
          >
            Action(s) taken when no other rule matches.
          </Typography>
        )}

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
              <Typography>
                <Warning message='The request will be ignored/dropped.' /> No
                Action
              </Typography>
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
              <OtherActions
                IconComponent={MoreHorizIcon}
                actions={[
                  {
                    label: 'Edit',
                    onClick: () => setEditActionIndex(i),
                  },
                  {
                    label: 'Delete',
                    onClick: () =>
                      props.onChange(props.value.filter((_, j) => j !== i)),
                  },
                ]}
                placement='right'
              />
            </Box>
          </Typography>
        ))}
      </Box>
    </React.Fragment>
  )
}
