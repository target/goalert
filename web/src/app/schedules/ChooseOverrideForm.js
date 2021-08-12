import React from 'react'
import p from 'prop-types'
import { FormContainer } from '../forms'
import {
  Grid,
  RadioGroup,
  FormControlLabel,
  FormLabel,
  Radio,
} from '@material-ui/core'

export default function ChooseOverrideForm(props) {
  const { scheduleID, value, removeUserReadOnly, ...formProps } = props

  const handleVariantChange = (e) => {
    if (e.target.value) {
      props.onChange({ ...value, variant: e.target.value })
    }
  }

  return (
    <FormContainer optionalLabels value={value} {...formProps}>
      <Grid item xs={12}>
        <RadioGroup
          isRequired
          aria-label='Choose an override action'
          name='variant'
          onChange={handleVariantChange}
        >
          <FormControlLabel
            data-cy='variant.replace'
            value='replace'
            control={<Radio />}
            label='Replace'
          />
          <FormLabel>This will replace a user from the schedule</FormLabel>

          <FormControlLabel
            data-cy='variant.remove'
            value='remove'
            control={<Radio />}
            label='Remove'
          />
          <FormLabel>This will remove a user from the schedule</FormLabel>
        </RadioGroup>
      </Grid>
    </FormContainer>
  )
}

ChooseOverrideForm.propTypes = {
  scheduleID: p.string.isRequired,

  value: p.shape({
    addUserID: p.string.isRequired,
    removeUserID: p.string.isRequired,
    start: p.string.isRequired,
    end: p.string.isRequired,
  }).isRequired,

  disabled: p.bool.isRequired,

  onChange: p.func.isRequired,
  removeUserReadOnly: p.bool,
}
