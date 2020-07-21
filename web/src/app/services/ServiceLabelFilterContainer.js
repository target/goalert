import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { Filter as LabelFilterIcon } from 'mdi-material-ui'

import { LabelKeySelect } from '../selection/LabelKeySelect'
import { LabelValueSelect } from '../selection/LabelValueSelect'
import FilterContainer from '../util/FilterContainer'

export default function ServiceLabelFilterContainer(props) {
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

ServiceLabelFilterContainer.propTypes = {
  value: p.shape({ labelKey: p.string, labelValue: p.string }),
  onChange: p.func,
  onReset: p.func,

  // optionally anchors the popover to a specified element's ref
  anchorRef: p.object,
}
