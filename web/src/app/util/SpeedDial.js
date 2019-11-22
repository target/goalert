import React, { useState } from 'react'
import p from 'prop-types'
import { SpeedDial, SpeedDialIcon, SpeedDialAction } from '@material-ui/lab'
import { makeStyles } from '@material-ui/styles'

const useStyles = makeStyles(theme => ({
  speedDial: {
    position: 'fixed',
    bottom: '2em',
    right: '2em',
    zIndex: 9001,
  },
  staticTooltipLabel: {
    whiteSpace: 'nowrap',
  },
}))

export default function CustomSpeedDial(props) {
  const [open, setOpen] = useState(false)
  const classes = useStyles()

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
      className={classes.speedDial}
    >
      {props.actions.map((action, idx) => (
        <SpeedDialAction
          key={idx}
          icon={action.icon}
          tooltipTitle={action.label}
          tooltipOpen
          classes={{ staticTooltipLabel: classes.staticTooltipLabel }}
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
