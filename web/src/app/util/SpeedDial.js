import React, { useState } from 'react'
import p from 'prop-types'
import { SpeedDial, SpeedDialIcon, SpeedDialAction } from '@material-ui/lab'

export default function CustomSpeedDial(props) {
  const [open, setOpen] = useState(false)

  const doToggle = () => setOpen(!open)
  const doOpen = () => setOpen(true)
  const doClose = () => setOpen(false)

  return (
    <SpeedDial
      ariaLabel={props.label}
      FabProps={{
        'data-cy': 'page-fab',
      }}
      icon={<SpeedDialIcon />}
      onClick={doToggle}
      onClose={doClose}
      onMouseEnter={doOpen}
      onMouseLeave={doClose}
      open={open}
      style={{
        position: 'fixed',
        bottom: '2em',
        right: '2em',
        zIndex: 9001,
      }}
    >
      {props.actions
        .slice() // TODO why this?
        .reverse() // TODO why this? If we cut this line, we have to reorder the action arrays everywhere else (3 instances)
        .map((action, idx) => (
          <SpeedDialAction
            key={idx}
            icon={action.icon}
            tooltipTitle={action.label}
            // tooltipOpen // TODO set to persist speed dial on mobile. Clunky default styles need to be overridden.
            // TooltipClasses={{ tooltip: { whiteSpace: 'nowrap' } }}
            aria-label={action.label}
            onClick={() => {
              setOpen(false)
              action.onClick()
            }}
          />
        ))}
    </SpeedDial>
  )
}

CustomSpeedDial.propsTypes = {
  label: p.string.isRequired,
  actions: p.arrayOf(
    p.shape({
      icon: p.element.isRequired,
      onClick: p.func.isRequired,
      label: p.string.isRequired,
    }),
  ).isRequired,
}
