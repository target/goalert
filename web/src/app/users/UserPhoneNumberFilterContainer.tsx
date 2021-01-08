import React, { Ref } from 'react'
import Grid from '@material-ui/core/Grid'
import { Filter as LabelFilterIcon } from 'mdi-material-ui'
import FilterContainer from '../util/FilterContainer'
import TelTextField from '../util/TelTextField'
import { useURLParam } from '../actions'

interface UserPhoneNumberFilterContainerProps {
  anchorRef?: Ref<HTMLElement>
}

export default function UserPhoneNumberFilterContainer(
  props: UserPhoneNumberFilterContainerProps,
): JSX.Element {
  const [searchParam, setSearchParam] = useURLParam('search', '' as string)

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
      onReset={() => setSearchParam('')}
      anchorRef={props.anchorRef}
    >
      <Grid data-cy='phone-number-container' item xs={12}>
        <TelTextField
          onChange={(e) =>
            setSearchParam(e.target.value ? 'phone=' + e.target.value : '')
          }
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
