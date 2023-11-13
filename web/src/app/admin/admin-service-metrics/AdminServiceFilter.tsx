import React, { useState } from 'react'
import FilterList from '@mui/icons-material/FilterList'
import Grid from '@mui/material/Grid'
import { useResetURLParams, useURLParams } from '../../actions'
import {
  Autocomplete,
  Button,
  Chip,
  ClickAwayListener,
  Divider,
  Drawer,
  IconButton,
  List,
  ListItem,
  ListItemText,
  Stack,
  TextField,
  Toolbar,
  Typography,
} from '@mui/material'
import { LabelKeySelect } from '../../selection'
import { LabelValueSelect } from '../../selection/LabelValueSelect'

function AdminServiceFilter(): React.ReactNode {
  const [open, setOpen] = useState<boolean>(false)

  const [params, setParams] = useURLParams({
    epStepTgts: [] as string[],
    intKeyTgts: [] as string[],
    labelKey: '',
    labelValue: '',
  })

  const resetAll = useResetURLParams(
    'epStepTgts',
    'intKeyTgts',
    'labelKey',
    'labelValue',
  )

  const removeFilter = (filterName: string): void => {
    if (filterName === 'labelKey') {
      setParams({ ...params, labelKey: '', labelValue: '' })
      return
    }
    if (filterName === 'labelValue') {
      setParams({ ...params, labelValue: '' })
      return
    }
    setParams({ ...params, [filterName]: [] })
  }

  function renderFilterChips(): React.ReactNode {
    return (
      <Stack direction='row' spacing={1} sx={{ marginTop: '10px' }}>
        {!!params.epStepTgts.length && (
          <Chip
            label={'ep step targets=' + params.epStepTgts.join(',')}
            onClick={() => setOpen(true)}
            onDelete={() => removeFilter('epStepTgts')}
          />
        )}
        {!!params.intKeyTgts.length && (
          <Chip
            label={'integration key targets=' + params.intKeyTgts.join(',')}
            onClick={() => setOpen(true)}
            onDelete={() => removeFilter('intKeyTgts')}
          />
        )}
        {!!params.labelKey.length && (
          <Chip
            label={'key=' + params.labelKey}
            onClick={() => setOpen(true)}
            onDelete={() => {
              removeFilter('labelKey')
            }}
          />
        )}
        {!!params.labelValue.length && (
          <Chip
            label={'value=' + params.labelValue}
            onClick={() => setOpen(true)}
            onDelete={() => removeFilter('labelValue')}
          />
        )}
      </Stack>
    )
  }

  function renderFilterDrawer(): React.ReactNode {
    return (
      <ClickAwayListener
        onClickAway={() => setOpen(false)}
        mouseEvent='onMouseUp'
      >
        <Drawer
          anchor='right'
          open={open}
          variant='persistent'
          data-cy='admin-service-filter'
        >
          <Toolbar />
          <Grid style={{ width: '30vw' }}>
            <Typography variant='h6' style={{ margin: '16px' }}>
              Service Filters
            </Typography>
            <Divider />
            <List>
              <ListItem>
                <ListItemText primary='EP Step Targets' />
              </ListItem>
              <ListItem>
                <Autocomplete
                  multiple
                  fullWidth
                  id='ep-step-targets'
                  options={['slack', 'webhook', 'schedule', 'user', 'rotation']}
                  value={params.epStepTgts}
                  onChange={(_, value) =>
                    setParams({ ...params, epStepTgts: value })
                  }
                  renderInput={(params) => (
                    <TextField {...params} label='Select Channels' />
                  )}
                />
              </ListItem>
              <Divider sx={{ padding: '10px' }} />
              <ListItem>
                <ListItemText primary='Integration Key Targets' />
              </ListItem>
              <ListItem>
                <Autocomplete
                  multiple
                  fullWidth
                  id='int-key-targets'
                  options={[
                    'generic',
                    'grafana',
                    'site24x7',
                    'prometheusAlertmanager',
                    'email',
                  ]}
                  value={params.intKeyTgts}
                  onChange={(_, value) =>
                    setParams({ ...params, intKeyTgts: value })
                  }
                  renderInput={(params) => (
                    <TextField {...params} label='Select Targets' />
                  )}
                />
              </ListItem>
              <Divider sx={{ padding: '10px' }} />
              <ListItem>
                <ListItemText primary='Labels' />
              </ListItem>
              <ListItem>
                <LabelKeySelect
                  name='label-key-select'
                  label='Select Label Key'
                  fullWidth
                  value={params.labelKey}
                  onChange={(value: string) =>
                    setParams({
                      ...params,
                      labelKey: value,
                      labelValue: '',
                    })
                  }
                />
              </ListItem>
              <ListItem>
                <LabelValueSelect
                  name='label-value'
                  fullWidth
                  label='Select Label Value'
                  labelKey={params.labelKey}
                  value={params.labelValue}
                  onChange={(value: string) =>
                    setParams({ ...params, labelValue: value })
                  }
                  disabled={!params.labelKey}
                />
              </ListItem>
              <ListItem>
                <Button
                  variant='outlined'
                  sx={{ marginRight: '10px' }}
                  onClick={resetAll}
                >
                  Clear All
                </Button>
              </ListItem>
            </List>
          </Grid>
        </Drawer>
      </ClickAwayListener>
    )
  }

  return (
    <React.Fragment>
      <Grid container>
        <Grid item xs>
          {renderFilterChips()}
        </Grid>
        <Grid item xs>
          <IconButton
            aria-label='Filter Alerts'
            onClick={() => setOpen(true)}
            size='large'
            sx={{ float: 'right' }}
          >
            <FilterList />
          </IconButton>
        </Grid>
      </Grid>
      {renderFilterDrawer()}
    </React.Fragment>
  )
}

export default AdminServiceFilter
