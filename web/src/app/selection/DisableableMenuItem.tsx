import { ListItemText, MenuItem, Tooltip, TooltipProps } from '@mui/material'
import React from 'react'

export interface DisableableMenuItemFields {
  value: string
  label?: string
  disabled: boolean
  disabledMessage: string
}

export function sortDisableableMenuItems(
  a: DisableableMenuItemFields,
  b: DisableableMenuItemFields,
): number {
  if (a.disabled === b.disabled) {
    return 0
  }
  return a.disabled ? 1 : -1
}

// the parent Form element is looking for the value prop, which doesn't exist on
// the base Tooltip.
const MenuTooltip = (props: TooltipProps & { value: string }): React.JSX.Element => {
  return <Tooltip {...props}>{props.children}</Tooltip>
}

export const renderMenuItem = (
  fields: DisableableMenuItemFields,
): React.JSX.Element => {
  const { value, label, disabled, disabledMessage } = fields
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
    <MenuItem key={value} value={value}>
      <ListItemText sx={{ margin: 0 }}>{label ?? value}</ListItemText>
    </MenuItem>
  )
}
