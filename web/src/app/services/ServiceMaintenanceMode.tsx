import React, { useState, useRef } from 'react'
import { gql, useMutation } from '@apollo/client'
import {
  ButtonGroup,
  Button,
  Popover,
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

  function calcExp(index: number = selectedIndex): string {
    switch (index) {
      case 0:
        return DateTime.now().plus({ hours: 1 }).toISO()
      case 1:
        return DateTime.now().plus({ hours: 2 }).toISO()
      case 2:
        return DateTime.now().plus({ hours: 4 }).toISO()
      default:
        return ''
    }
  }

  function handleStartMaintenance(): void {
    updateService({
      variables: {
        input: {
          id: p.serviceID,
          maintenanceExpiresAt: calcExp(selectedIndex),
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

  function handleOpen(): void {
    setOpen(true)
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
          onClick={handleOpen}
        >
          <ArrowDropDown />
        </Button>
      </ButtonGroup>
      <Tooltip
        title={
          <Typography variant='body2'>
            Pause all outgoing notifications and escalations for{' '}
            {options[selectedIndex]}. Alerts may still be created and will
            continue as normal after maintenance mode ends.
          </Typography>
        }
        sx={{ pl: 1 }}
      >
        <Info color='secondary' />
      </Tooltip>
      <Popover
        anchorEl={anchorRef.current}
        open={open}
        onClose={handleClose}
        anchorOrigin={{
          vertical: 'bottom',
          horizontal: 'center',
        }}
        transformOrigin={{
          vertical: 'top',
          horizontal: 'center',
        }}
      >
        <MenuList
          id='split-button-menu'
          autoFocusItem
          sx={{ width: anchorRef.current?.offsetWidth }}
        >
          {options.map((option, index) => (
            <MenuItem
              key={option}
              selected={index === selectedIndex}
              onClick={(event) => handleMenuItemClick(event, index)}
              sx={{ display: 'block', lineHeight: 1 }}
            >
              <Typography variant='body2'>Ends in {option}</Typography>
              <Typography variant='caption' color='textSecondary'>
                At {DateTime.fromISO(calcExp(index)).toFormat('t ZZZZ')}
              </Typography>
            </MenuItem>
          ))}
        </MenuList>
      </Popover>
    </div>
  )
}
