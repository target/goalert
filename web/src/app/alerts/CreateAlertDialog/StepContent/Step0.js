import React from 'react'
import { Grid, TextField } from '@material-ui/core'
import { FormField } from '../../../forms'

export default props => (
  <Grid container spacing={2}>
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
        multiline
        rows={17}
        variant='outlined'
        label='Alert Details'
        name='details'
        required
        component={TextField}
      />
    </Grid>
  </Grid>
)

// MuiGrid-root MuiGrid-container MuiGrid-spacing-xs-2 MuiGrid-direction-xs-column
// height: 100%

// MuiGrid-root MuiGrid-item MuiGrid-grid-xs-12
// flex: 0

// MuiGrid-root MuiGrid-item MuiGrid-grid-xs-12
// flex: 1

// MuiFormControl-root MuiFormControl-fullWidth
// height: 100%

// MuiFormControl-root MuiTextField-root MuiFormControl-fullWidth
// flex: 1

// MuiInputBase-root MuiOutlinedInput-root MuiInputBase-fullWidth MuiInputBase-formControl MuiInputBase-multiline MuiOutlinedInput-multiline
// flex: 1;
// flex-direction: column;

// textarea
// flex: 1;
