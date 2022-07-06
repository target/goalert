import React, { useState, useRef } from 'react'
import { gql, useMutation } from '@apollo/client'
import {
  ButtonGroup,
  Button,
  Popper,
  Grow,
  Paper,
  ClickAwayListener,
  MenuList,
  MenuItem,
  Tooltip,
  Typography,
} from '@mui/material'
import { ArrowDropDown, Info } from '@mui/icons-material'
import { DateTime } from 'luxon'

interface Props {
  serviceID: string
  expiresAt?: string
}

const options = ['1 hour', '2 hours', '4 hours']

const mutation = gql`
  mutation updateService($input: UpdateServiceInput!) {
    updateService(input: $input)
  }
`

export default function ServiceMaintenanceMode(p: Props): JSX.Element {
  const [open, setOpen] = useState(false)
  const [selectedIndex, setSelectedIndex] = useState(0)

  const anchorRef = useRef<HTMLDivElement>(null)
  const [updateService] = useMutation(mutation)

  function handleStartMaintenance(): void {
    let expiresAt
    switch (selectedIndex) {
      case 0:
        expiresAt = DateTime.now().plus({ hours: 1 })
        break
      case 1:
        expiresAt = DateTime.now().plus({ hours: 2 })
        break
      case 2:
        expiresAt = DateTime.now().plus({ hours: 4 })
    }

    updateService({
      variables: {
        input: {
          id: p.serviceID,
          maintenanceExpiresAt: expiresAt,
        },
      },
    })
  }

  function handleMenuItemClick(
    event: React.MouseEvent<HTMLLIElement, MouseEvent>,
    index: number,
  ): void {
    setSelectedIndex(index)
    setOpen(false)
  }

  function handleToggle(): void {
    setOpen((prevOpen) => !prevOpen)
  }

  function handleClose(): void {
    setOpen(false)
  }

  return (
    <div
      style={{
        display: 'flex',
        alignItems: 'center',
      }}
    >
      <ButtonGroup
        variant='contained'
        ref={anchorRef}
        aria-label='split button'
      >
        <Button onClick={handleStartMaintenance}>Start Maintenance Mode</Button>
        <Button
          size='small'
          aria-controls={open ? 'split-button-menu' : undefined}
          aria-expanded={open ? 'true' : undefined}
          aria-label='select merge strategy'
          aria-haspopup='menu'
          onClick={handleToggle}
        >
          <ArrowDropDown />
        </Button>
      </ButtonGroup>
      <Tooltip
        title={
          <Typography variant='body2'>
            Pause all outgoing notifications and escalations for{' '}
            {options[selectedIndex]}. Alerts may still be created and will
            continue as normal after maintenance mode expires.
          </Typography>
        }
        sx={{ pl: 1 }}
      >
        <Info color='secondary' />
      </Tooltip>
      <Popper
        open={open}
        anchorEl={anchorRef.current}
        role={undefined}
        transition
        disablePortal
        placement='bottom'
      >
        {({ TransitionProps }) => (
          <Grow
            {...TransitionProps}
            style={{
              transformOrigin: 'center top',
            }}
          >
            <Paper>
              <ClickAwayListener onClickAway={handleClose}>
                <MenuList id='split-button-menu' autoFocusItem>
                  {options.map((option, index) => (
                    <MenuItem
                      key={option}
                      selected={index === selectedIndex}
                      onClick={(event) => handleMenuItemClick(event, index)}
                    >
                      {option}
                    </MenuItem>
                  ))}
                </MenuList>
              </ClickAwayListener>
            </Paper>
          </Grow>
        )}
      </Popper>
    </div>
  )
}
