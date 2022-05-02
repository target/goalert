import React, { useContext, useState } from 'react'
import Collapse from '@mui/material/Collapse'
import FormControlLabel from '@mui/material/FormControlLabel'
import Grid from '@mui/material/Grid'
import Switch from '@mui/material/Switch'
import ToggleButton from '@mui/material/ToggleButton'
import ToggleButtonGroup from '@mui/material/ToggleButtonGroup'
import Typography from '@mui/material/Typography'
import DarkModeIcon from '@mui/icons-material/DarkMode'
import SystemIcon from '@mui/icons-material/SettingsBrightness'
import LightModeIcon from '@mui/icons-material/LightMode'
import ArrowDropDownIcon from '@mui/icons-material/ArrowDropDown'
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
        <Typography variant='body1'>Appearance</Typography>
      </Grid>
      <Grid item>
        <ToggleButtonGroup color='primary' value={themeMode}>
          <ToggleButton
            value='light'
            onClick={() => setThemeMode('light')}
            sx={{
              borderBottomLeftRadius: showMore ? 0 : 'inherit',
            }}
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
            value='showmore'
            onClick={() => setShowMore(!showMore)}
            sx={{
              borderBottomRightRadius: showMore ? 0 : 'inherit',
            }}
          >
            <ArrowDropDownIcon />
          </ToggleButton>
        </ToggleButtonGroup>
        <Collapse in={showMore}>
          <Grid container direction='column' spacing={2}>
            <Grid item>
              <HexColorPicker
                color={sourceColor}
                onChange={setSourceColor}
                style={{ width: '100%', height: 'fit-content' }}
              />
            </Grid>
            <Grid item>
              <FormControlLabel
                control={
                  <Switch
                    checked={highContrast}
                    onChange={() => setHighContrast(!highContrast)}
                  />
                }
                label='High Contrast'
              />
            </Grid>
          </Grid>
        </Collapse>
      </Grid>
    </Grid>
  )
}
