import React, { useState } from 'react'
import { Button, Card, Grid } from '@mui/material'
import ResetIcon from '@mui/icons-material/Replay'
import FilterIcon from '@mui/icons-material/FilterAlt'
import { ISODateTimePicker } from '../../util/ISOPickers'
import Search from '../../util/Search'
import { useURLParams } from '../../actions'
import FilterContainer from '../../util/FilterContainer'
import { ServiceSelect, UserSelect } from '../../selection'

interface Props {
  resetCount: () => void
}

export default function DebugMessagesControls(props: Props): JSX.Element {
  const [params, setParams] = useURLParams({
    search: '',
    start: '',
    end: '',
  })

  const [filterByUser, setFilterByUser] = useState('')
  const [filterByService, setFilterByService] = useState('')

  return (
    <Card>
      <Grid container spacing={1} sx={{ padding: 2 }}>
        <Grid item sx={{ flex: 1 }}>
          <Search
            transition={false}
            fullWidth
            endAdornment={
              <FilterContainer icon={<FilterIcon />}>
                <Grid item xs={12}>
                  <UserSelect
                    label='Select a user...'
                    value={filterByUser}
                    onChange={(val) => {
                      setFilterByUser(val)
                      setParams({ ...params, search: val })
                    }}
                  />
                </Grid>
                <Grid item xs={12}>
                  <ServiceSelect
                    label='Select a service...'
                    value={filterByService}
                    onChange={(val) => {
                      setFilterByService(val)
                      setParams({ ...params, search: val })
                    }}
                  />
                </Grid>
              </FilterContainer>
            }
          />
        </Grid>
        <Grid item>
          <ISODateTimePicker
            placeholder='Start'
            name='startDate'
            value={params.start}
            onChange={(newStart) => {
              setParams({ ...params, start: newStart as string })
              props.resetCount()
            }}
            label='Created After'
            size='small'
            variant='outlined'
          />
        </Grid>
        <Grid item>
          <ISODateTimePicker
            placeholder='End'
            name='endDate'
            value={params.end}
            label='Created Before'
            onChange={(newEnd) => {
              setParams({ ...params, end: newEnd as string })
              props.resetCount()
            }}
            size='small'
            variant='outlined'
          />
        </Grid>
        <Grid item>
          <Button
            aria-label='Reset Filters'
            variant='outlined'
            onClick={() => {
              setParams({
                search: '',
                start: '',
                end: '',
              })
              props.resetCount()
            }}
            endIcon={<ResetIcon />}
            sx={{ height: '100%' }}
          >
            Reset
          </Button>
        </Grid>
      </Grid>
    </Card>
  )
}
