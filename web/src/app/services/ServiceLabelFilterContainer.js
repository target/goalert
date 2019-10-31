import React from 'react'
import p from 'prop-types'
import Grid from '@material-ui/core/Grid'
import Typography from '@material-ui/core/Typography'
import { Filter as LabelFilterIcon } from 'mdi-material-ui'

import { LabelKeySelect } from '../selection/LabelKeySelect'
import { LabelValueSelect } from '../selection/LabelValueSelect'
import FilterContainer from '../util/FilterContainer'

export default function ServiceLabelFilterContainer(props) {
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
          value={props.labelKey}
          onChange={props.onKeyChange}
        />
      </Grid>
      <Grid data-cy='label-value-container' item xs={12}>
        <LabelValueSelect
          name='label-value'
          label='Select Label Value'
          keyValue={props.labelKey}
          value={props.labelValue}
          onChange={props.onValueChange}
          disabled={!props.labelKey}
        />
      </Grid>
    </FilterContainer>
  )
}

ServiceLabelFilterContainer.propTypes = {
  labelKey: p.string,
  labelValue: p.string,
  onKeyChange: p.func.isRequired,
  onValueChange: p.func.isRequired,
  onReset: p.func,

  // optionally anchors the popover to a specified element's ref
  anchorRef: p.object,
}
