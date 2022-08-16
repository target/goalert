import React, { Ref } from 'react'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import { Filter as LabelFilterIcon } from 'mdi-material-ui'

import { LabelKeySelect } from '../selection/LabelKeySelect'
import { LabelValueSelect } from '../selection/LabelValueSelect'
import FilterContainer from '../util/FilterContainer'

interface Value {
  labelKey: string
  labelValue: string
}

interface ServiceFilterContainerProps {
  value: Value
  onChange: (val: Value) => void
  onReset: () => void

  // optionally anchors the popover to a specified element's ref
  anchorRef?: Ref<HTMLElement>
}

export default function ServiceFilterContainer(
  props: ServiceFilterContainerProps,
): JSX.Element {
  const { labelKey, labelValue } = props.value
  return (
    <FilterContainer
      icon={<LabelFilterIcon />}
      title='Search by Labels'
      iconButtonProps={{
        'data-cy': 'services-filter-button',
        color: 'default',
        edge: 'end',
        size: 'small',
      }}
      onReset={props.onReset}
      anchorRef={props.anchorRef}
    >
      <Grid item xs={12}>
        <Typography color='textSecondary'>
          <i>Search by Label</i>
        </Typography>
      </Grid>
      <Grid data-cy='label-key-container' item xs={12}>
        <LabelKeySelect
          name='label-key'
          label='Select Label Key'
          value={labelKey}
          onChange={(labelKey) => props.onChange({ ...props.value, labelKey })}
        />
      </Grid>
      <Grid data-cy='label-value-container' item xs={12}>
        <LabelValueSelect
          name='label-value'
          label='Select Label Value'
          labelKey={labelKey}
          value={labelValue}
          onChange={(v) =>
            props.onChange({ ...props.value, labelValue: v || '' })
          }
          disabled={!labelKey}
        />
      </Grid>
    </FilterContainer>
  )
}
