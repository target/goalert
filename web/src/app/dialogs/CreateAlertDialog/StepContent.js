import React from 'react'
import { Grid, TextField, InputAdornment } from '@material-ui/core'

import { FormContainer, FormField } from '../../forms'
import ServiceLabelFilterContainer from '../../services/ServiceLabelFilterContainer'
import { Search as SearchIcon } from '@material-ui/icons'

export default props => {
  switch (props.activeStep) {
    case 0:
      return (
        <Grid container spacing={2}>
          <FormContainer onChange={props.onChange}>
            <Grid item xs={12}>
              <FormField
                fullWidth
                label='Alert Summary'
                name='summary'
                required
                component={TextField}
              />
            </Grid>
            <Grid item xs={12}>
              <FormField
                fullWidth
                label='Alert Details'
                name='details'
                required
                component={TextField}
              />
            </Grid>
          </FormContainer>
        </Grid>
      )
    case 1:
      return (
        <Grid item xs={12}>
          <FormContainer onChange={props.onChange}>
            {/* <Grid item xs={12}>
              <SelectedServices />
            </Grid> */}
            <FormField
              fullWidth
              label='Search Query'
              name='searchQuery'
              fieldName='searchQuery'
              required
              component={TextField}
              InputProps={{
                // ref: fieldRef,
                startAdornment: (
                  <InputAdornment position='start'>
                    <SearchIcon color='action' />
                  </InputAdornment>
                ),
                endAdornment: (
                  <ServiceLabelFilterContainer
                    // anchorRef={fieldRef}
                    labelKey={'key'}
                    labelValue={'() => console.log(value)'}
                    onKeyChange={() => console.log('onKeyChange')}
                    onValueChange={() => console.log('onValueChange')}
                    // onReset={() => setSearchParam()}
                  />
                ),
              }}
            />
          </FormContainer>
        </Grid>
      )

    case 2:
      return 'plz confirm ur info'
    default:
      return 'Unknown step'
  }
}
