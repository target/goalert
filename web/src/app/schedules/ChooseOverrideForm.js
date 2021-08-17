import React from 'react'
import p from 'prop-types'
import { FormContainer } from '../forms'
import {
  Grid,
  RadioGroup,
  FormControlLabel,
  Radio,
  FormHelperText,
} from '@material-ui/core'
import { variantDetails } from './ScheduleCalendarOverrideDialog'

export default function ChooseOverrideForm(props) {
  const { value, errors = [], removeUserReadOnly, ...formProps } = props

  const handleVariantChange = (e) => {
    if (e.target.value) {
      props.onChange({ ...value, variant: e.target.value })
    }
  }

  return (
    <FormContainer optionalLabels errors={errors} value={value} {...formProps}>
      <Grid item xs={12}>
        <RadioGroup
          required
          aria-label='Choose an override action'
          name='variant'
          onChange={handleVariantChange}
          value={value.variant}
        >
          {props.variantOptions.map((variant) => (
            <FormControlLabel
              key={variant}
              data-cy={`variant.${variant}`}
              value={variant}
              control={<Radio />}
              label={
                <React.Fragment>
                  {variantDetails[variant].name}
                  <FormHelperText>
                    {variantDetails[variant].helperText}
                  </FormHelperText>
                </React.Fragment>
              }
            />
          ))}
        </RadioGroup>
      </Grid>
    </FormContainer>
  )
}

ChooseOverrideForm.propTypes = {
  value: p.shape({
    addUserID: p.string.isRequired,
    removeUserID: p.string.isRequired,
    start: p.string.isRequired,
    end: p.string.isRequired,
  }).isRequired,

  disabled: p.bool.isRequired,

  onChange: p.func.isRequired,
  removeUserReadOnly: p.bool,
  variantOptions: p.arrayOf(p.string),
}
