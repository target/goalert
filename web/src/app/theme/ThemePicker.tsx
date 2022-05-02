import React, { useContext, useState } from 'react'
import Collapse from '@mui/material/Collapse'
import FormControlLabel from '@mui/material/FormControlLabel'
import FormGroup from '@mui/material/FormGroup'
import FormLabel from '@mui/material/FormLabel'
import Grid from '@mui/material/Grid'
import IconButton from '@mui/material/IconButton'
import Switch from '@mui/material/Switch'
import ToggleButton from '@mui/material/ToggleButton'
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import SystemIcon from '@mui/icons-material/SettingsBrightness'
import LightModeIcon from '@mui/icons-material/LightMode'
import ResetIcon from '@mui/icons-material/Refresh'
import ExpandMoreIcon from '@mui/icons-material/ExpandLess'
import ExpandLessIcon from '@mui/icons-material/ExpandMore'
import { blueGrey } from '@mui/material/colors'
import { HexColorPicker } from 'react-colorful'
import { ThemeContext } from './themeConfig'

export default function ThemePicker(): JSX.Element {
  const {
    themeMode,
    setThemeMode,
    sourceColor,
    setSourceColor,
    highContrast,
    setHighContrast,
  } = useContext(ThemeContext)
  const [showMore, setShowMore] = useState(false)

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
          <FormGroup
            sx={{
              p: 1,
              border: (theme) => '1px solid ' + theme.palette.divider,
              borderTop: 0,
            }}
          >
            <FormControlLabel
              control={
                <IconButton onClick={() => setSourceColor(blueGrey[500])}>
                  <ResetIcon />
                </IconButton>
              }
              label='Reset to Default'
              labelPlacement='start'
              sx={{ ml: 0, justifyContent: 'space-between' }}
            />
            <FormControlLabel
              control={
                <Switch
                  checked={highContrast}
                  onChange={() => setHighContrast(!highContrast)}
                />
              }
              label='Use High Contrast'
              labelPlacement='start'
              sx={{ ml: 0, justifyContent: 'space-between' }}
            />
          </FormGroup>
          <HexColorPicker
            color={sourceColor}
            onChange={setSourceColor}
            style={{ width: '100%', height: 'fit-content' }}
          />
        </Collapse>
      </Grid>
    </Grid>
  )
}
