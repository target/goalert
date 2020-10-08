import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import { Filter as LabelFilterIcon } from 'mdi-material-ui'
import FilterContainer from '../util/FilterContainer'
import TelTextField from '../util/TelTextField'
import { searchSelector } from '../selectors'
import { setURLParam } from '../actions'
import { useDispatch, useSelector } from 'react-redux'

export default function UserPhoneNumberFilterContainer(props) {
  const searchParam = useSelector(searchSelector) // current total search string on page load
  const dispatch = useDispatch()
  const setSearchParam = (value) => dispatch(setURLParam('search', value))

  return (
    <FilterContainer
      icon={<LabelFilterIcon />}
      title='Search by Phone Number'
      iconButtonProps={{
        'data-cy': 'services-filter-button',
        color: 'default',
        edge: 'end',
        size: 'small',
      }}
      onReset={() => setSearchParam()}
      anchorRef={props.anchorRef}
    >
      <Grid data-cy='phone-number-container' item xs={12}>
        <TelTextField
          onChange={(e) => setSearchParam('phone=' + e.target.value)}
          value={searchParam.replace(/^phone=/, '')}
          fullWidth
          label='Search by Phone Number'
          helperText='Please provide your country code e.g. +1 (USA)'
          type='tel'
        />
      </Grid>
    </FilterContainer>
  )
}

UserPhoneNumberFilterContainer.propTypes = {
  onReset: p.func,

  // optionally anchors the popover to a specified element's ref
  anchorRef: p.object,
}
