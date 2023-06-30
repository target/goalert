import { ListItemText, MenuItem, Tooltip, TooltipProps } from '@mui/material'
import React from 'react'

export interface NotificationTypeOption {
  label: string
  value: string
  disabled: boolean
  disabledMessage: string
}

// the parent Form element is looking for the value prop, which doesn't exist on
// the base Tooltip.
const MenuTooltip = (props: TooltipProps & { value: string }): JSX.Element => {
  return <Tooltip {...props}>{props.children}</Tooltip>
}

export const renderNotificationType = (
  notificationType: NotificationTypeOption,
): JSX.Element => {
  const { label, value, disabled, disabledMessage } = notificationType
  return disabled ? (
    // tooltips don't work on disabled elements so the MenuItem must be wrapped in a <span/>
    <MenuTooltip key={value} value={value} title={disabledMessage}>
      <span>
        <MenuItem key={value} value={value} disabled>
          <ListItemText>{label}</ListItemText>
        </MenuItem>
      </span>
    </MenuTooltip>
  ) : (
    <MenuItem key={value} value={value} disabled={disabled}>
      <ListItemText>{label}</ListItemText>
    </MenuItem>
  )
}
