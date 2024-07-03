import React, { useState } from 'react'
import {
  SpeedDial,
  SpeedDialIcon,
  SpeedDialAction,
  SpeedDialActionProps,
  SpeedDialProps,
} from '@mui/material'
import { makeStyles } from '@mui/styles'

const useStyles = makeStyles({
  speedDial: {
    position: 'fixed',
    bottom: '16px',
    right: '16px',
    zIndex: 9001,
  },
  staticTooltipLabel: {
    whiteSpace: 'nowrap',
  },
  disabledStaticTooltipLabel: {
    whiteSpace: 'nowrap',
    color: 'rgb(138,138,138)',
    backgroundColor: 'rgb(185,185,185)',
  },
})

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
  const classes = useStyles()

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
      className={classes.speedDial}
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
            classes={{
              staticTooltipLabel: action.disabled
                ? classes.disabledStaticTooltipLabel
                : classes.staticTooltipLabel,
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
