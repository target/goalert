import React, { Ref, useEffect, useState } from 'react'
import Grid from '@mui/material/Grid'
import { Filter as LabelFilterIcon } from 'mdi-material-ui'
import FilterContainer from '../util/FilterContainer'
import TelTextField from '../util/TelTextField'
import { useURLParam } from '../actions'
import { DEBOUNCE_DELAY } from '../config'

interface UserPhoneNumberFilterContainerProps {
  anchorRef?: Ref<HTMLElement>
}

export default function UserPhoneNumberFilterContainer(
  props: UserPhoneNumberFilterContainerProps,
): React.ReactNode {
  const [searchParam, setSearchParam] = useURLParam('search', '' as string)
  const [search, setSearch] = useState(searchParam)

  // If the page search param changes, we update state directly.
  useEffect(() => {
    setSearch(searchParam)
  }, [searchParam])

  // When typing, we setup a debounce before updating the URL.
  useEffect(() => {
    const t = setTimeout(() => {
      setSearchParam(search)
    }, DEBOUNCE_DELAY)

    return () => clearTimeout(t)
  }, [search])

  return (
    <FilterContainer
      icon={<LabelFilterIcon />}
      title='Search by Phone Number'
      iconButtonProps={{
        'data-cy': 'users-filter-button',
        color: 'default',
        edge: 'end',
        size: 'small',
      }}
      onReset={() => setSearch('')}
      anchorRef={props.anchorRef}
    >
      <Grid data-cy='phone-number-container' item xs={12}>
        <TelTextField
          onChange={(e) =>
            setSearch(e.target.value ? 'phone=' + e.target.value : '')
          }
          value={search.replace(/^phone=/, '')}
          fullWidth
          name='user-phone-search'
          label='Search by Phone Number'
        />
      </Grid>
    </FilterContainer>
  )
}
