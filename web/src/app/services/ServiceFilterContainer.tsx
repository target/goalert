import React, { Ref } from 'react'
import Grid from '@mui/material/Grid'
import Typography from '@mui/material/Typography'
import { Filter as LabelFilterIcon } from 'mdi-material-ui'

import { LabelKeySelect } from '../selection/LabelKeySelect'
import { LabelValueSelect } from '../selection/LabelValueSelect'
import { IntegrationKeySelect } from '../selection/IntegrationKeySelect'
import FilterContainer from '../util/FilterContainer'

interface Value {
  labelKey: string
  labelValue: string
  integrationKey?: string
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
): React.JSX.Element {
  const { labelKey, labelValue, integrationKey } = props.value
  return (
    <FilterContainer
      icon={<LabelFilterIcon />}
      title='Search Services by Filters'
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
          <i>Search by Integration Key</i>
        </Typography>
      </Grid>
      <Grid data-cy='integration-key-container' item xs={12}>
        <IntegrationKeySelect
          name='integration-key'
          label='Select Integration Key'
          value={integrationKey}
          formatInputOnChange={(input: string): string => {
            if (input.indexOf('token=') > -1) {
              input = input.substring(input.indexOf('token=') + 6)
            }
            return input
          }}
          onChange={(integrationKey: string) =>
            props.onChange({ ...props.value, integrationKey })
          }
        />
      </Grid>
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
          onChange={(labelKey: string) =>
            props.onChange({ ...props.value, labelKey })
          }
        />
      </Grid>
      <Grid data-cy='label-value-container' item xs={12}>
        <LabelValueSelect
          name='label-value'
          label='Select Label Value'
          labelKey={labelKey}
          value={labelValue}
          onChange={(v: string) =>
            props.onChange({ ...props.value, labelValue: v || '' })
          }
          disabled={!labelKey}
        />
      </Grid>
    </FilterContainer>
  )
}
