import React, { useContext, useState } from 'react'
import Collapse from '@mui/material/Collapse'
import FormLabel from '@mui/material/FormLabel'
import Grid from '@mui/material/Grid'
import List from '@mui/material/List'
import ListItemButton from '@mui/material/ListItemButton'
import ListItemIcon from '@mui/material/ListItemIcon'
import ListItemText from '@mui/material/ListItemText'
import ToggleButton from '@mui/material/ToggleButton'
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup'
import { blueGrey } from '@mui/material/colors'

import DarkModeIcon from '@mui/icons-material/DarkMode'
import SystemIcon from '@mui/icons-material/SettingsBrightness'
import LightModeIcon from '@mui/icons-material/LightMode'
import ExpandMoreIcon from '@mui/icons-material/ExpandLess'
import ExpandLessIcon from '@mui/icons-material/ExpandMore'
import DefaultColorIcon from '@mui/icons-material/Circle'
import PaletteIcon from '@mui/icons-material/Palette'
import { ThemeContext } from './themeConfig'

export default function ThemePicker(): React.JSX.Element {
  const { themeMode, setThemeMode, sourceColor, setSourceColor } =
    useContext(ThemeContext)
  const [showMore, setShowMore] = useState(false)

  const defaultSelected = sourceColor === blueGrey[500]
  const customSelected = sourceColor !== blueGrey[500]

  return (
    <Grid container direction='column' spacing={2}>
      <Grid item>
        <FormLabel>Appearance</FormLabel>
      </Grid>
      <Grid item>
        <ToggleButtonGroup color='primary' value={themeMode}>
          <ToggleButton
            value='light'
            onClick={() => setThemeMode('light')}
            sx={{ borderBottomLeftRadius: showMore ? 0 : 'inherit' }}
          >
            <LightModeIcon sx={{ paddingRight: 1 }} />
            Light
          </ToggleButton>
          <ToggleButton value='system' onClick={() => setThemeMode('system')}>
            <SystemIcon sx={{ paddingRight: 1 }} />
            System
          </ToggleButton>
          <ToggleButton value='dark' onClick={() => setThemeMode('dark')}>
            <DarkModeIcon sx={{ paddingRight: 1 }} />
            Dark
          </ToggleButton>
          <ToggleButton
            value='reset'
            aria-label='More Options'
            onClick={() => setShowMore(!showMore)}
            sx={{ borderBottomRightRadius: showMore ? 0 : 'inherit' }}
          >
            {showMore ? <ExpandMoreIcon /> : <ExpandLessIcon />}
          </ToggleButton>
        </ToggleButtonGroup>
        <Collapse in={showMore}>
          <List
            disablePadding
            sx={{
              border: (theme) => '1px solid ' + theme.palette.divider,
              borderRadius: '0 0 8px 8px',
              borderTop: 0,
            }}
          >
            <ListItemButton
              selected={defaultSelected}
              onClick={() => {
                setSourceColor(blueGrey[500])
              }}
            >
              <ListItemIcon>
                <DefaultColorIcon
                  color={defaultSelected ? 'primary' : 'inherit'}
                />
              </ListItemIcon>
              <ListItemText
                primary='Default'
                sx={{
                  color: (t) =>
                    defaultSelected
                      ? t.palette.primary.main
                      : t.palette.text.secondary,
                }}
              />
            </ListItemButton>
            <ListItemButton
              selected={customSelected}
              onClick={() => {
                document.getElementById('custom-color-picker')?.click()
              }}
            >
              <input
                id='custom-color-picker'
                onChange={(e) => setSourceColor(e.target.value)}
                // onInput required for Cypress test to execute a `trigger:input` event
                onInput={(e) => setSourceColor(e.currentTarget.value)}
                type='color'
                value={sourceColor}
                style={{
                  position: 'absolute',
                  opacity: 0,
                  border: 'none',
                  padding: 0,
                }}
              />
              <ListItemIcon>
                <PaletteIcon color={customSelected ? 'primary' : 'inherit'} />
              </ListItemIcon>
              <ListItemText
                primary='Choose Color...'
                sx={{
                  color: (t) =>
                    customSelected
                      ? t.palette.primary.main
                      : t.palette.text.secondary,
                }}
              />
            </ListItemButton>
          </List>
        </Collapse>
      </Grid>
    </Grid>
  )
}
