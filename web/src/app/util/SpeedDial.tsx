import React, { useState } from 'react'
import {
  SpeedDial,
  SpeedDialIcon,
  SpeedDialAction,
  SpeedDialActionProps,
  SpeedDialProps,
} from '@mui/material'

interface CustomSpeedDialProps {
  label: SpeedDialProps['ariaLabel']
  actions: {
    icon: SpeedDialActionProps['icon']
    onClick: NonNullable<SpeedDialActionProps['onClick']>
    label: SpeedDialActionProps['tooltipTitle'] &
      SpeedDialActionProps['aria-label']
    disabled?: boolean
  }[]
}

export default function CustomSpeedDial(
  props: CustomSpeedDialProps,
): JSX.Element {
  const [open, setOpen] = useState(false)

  return (
    <SpeedDial
      ariaLabel={props.label}
      FabProps={
        {
          'data-cy': 'page-fab',
        } as SpeedDialProps['FabProps']
      }
      icon={<SpeedDialIcon />}
      onClick={() => setOpen(!open)}
      onClose={() => setOpen(false)}
      onMouseEnter={() => setOpen(true)}
      onMouseLeave={() => setOpen(false)}
      open={open}
      sx={{
        position: 'fixed',
        bottom: '16px',
        right: '16px',
        zIndex: 9001,
      }}
      TransitionProps={{
        unmountOnExit: true,
      }}
    >
      {props.actions
        .slice()
        .reverse()
        .map((action, idx) => (
          <SpeedDialAction
            key={idx}
            icon={action.icon}
            tooltipTitle={action.label}
            tooltipOpen
            sx={{
              '& .MuiSpeedDialAction-staticTooltipLabel': action.disabled
                ? {
                    whiteSpace: 'nowrap',
                    color: 'rgb(138,138,138)',
                    backgroundColor: 'rgb(185,185,185)',
                  }
                : { whiteSpace: 'nowrap' },
            }}
            aria-label={action.label}
            FabProps={{ disabled: action.disabled }}
            onClick={(e) => {
              if (action.disabled) {
                e.stopPropagation()
                return
              }
              action.onClick(e)
              setOpen(false)
            }}
          />
        ))}
    </SpeedDial>
  )
}
