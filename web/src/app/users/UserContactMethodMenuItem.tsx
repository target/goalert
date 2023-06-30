import { ListItemText, MenuItem, Tooltip, TooltipProps } from '@mui/material'
import React from 'react'

export interface ContactMethodOption {
  value: string
  label?: string
  disabled: boolean
  disabledMessage: string
}

// the parent Form element is looking for the value prop, which doesn't exist on
// the base Tooltip.
const MenuTooltip = (props: TooltipProps & { value: string }): JSX.Element => {
  return <Tooltip {...props}>{props.children}</Tooltip>
}

export const renderContactMethod = (
  notificationType: ContactMethodOption,
): JSX.Element => {
  const { value, label, disabled, disabledMessage } = notificationType
  return disabled ? (
    // tooltips don't work on disabled elements so the MenuItem must be wrapped in a <span/>
    <MenuTooltip
      key={value}
      value={value}
      title={disabledMessage}
      placement='left'
    >
      <span>
        <MenuItem key={value} value={value} disabled>
          <ListItemText>{label ?? value}</ListItemText>
        </MenuItem>
      </span>
    </MenuTooltip>
  ) : (
    <MenuItem
      key={value}
      value={value}
      sx={{ paddingTop: 2.5, paddingBottom: 2.5 }}
    >
      <ListItemText sx={{ position: 'absolute', margin: 0 }}>
        {label ?? value}
      </ListItemText>
    </MenuItem>
  )
}
