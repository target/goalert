import React from 'react'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { Checkbox, Hidden, Table, TableCell, TableHead, TableRow, TableBody, MenuItem } from '@material-ui/core'
import { FormContainer, FormField } from '../forms'
import FormGroup from '@material-ui/core/FormGroup';
import FormControlLabel from '@material-ui/core/FormControlLabel';

const roles = [
    'user',
    'admin',
    'unknown'
]

interface Value {
  name: string
  role: string
  isAdmin: boolean
}

interface UserFormProps {
  value: Value

  errors: {
    field: 'name' | 'role'
    message: string
  }[]

  onChange: (val: Value) => void

  disabled: boolean
}



export default function UserForm(props: UserFormProps): JSX.Element {
    const [state, setState] = React.useState({
    checkedUser: !props.value.isAdmin,
    checkedAdmin: props.value.isAdmin,
    });
    const { ...containerProps } = props

    const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setState({ ...state, [event.target.name]: event.target.checked });
    };
    
    /*if (props.value.role == 'admin') {
        //setState({ ...state, checkedAdmin: true })
        state.checkedAdmin = true
        //setState({ checkedAdmin: true, checkedUser: false })
    }
    if (props.value.role == 'user') {
        //setState({ checkedAdmin: false, checkedUser: true})
         state.checkedUser = true
    }*/

  return (
    <FormContainer {...containerProps}>
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <FormField
            fullWidth
            label='Name'
            name='name'
            component={TextField}
            disabled={true}
          />
        </Grid>
        <Grid item xs={12}>
        <FormControlLabel
            control={
            <Checkbox
            checked={state.checkedAdmin}
            onChange={handleChange}
            //onChange={(newValue) => setState(newValue)}
            // onChange={(value) => this.setState({ value })}
            name="checkedAdmin"
            color="primary"
            />
            }
            label="Admin"
        /> 
        <FormControlLabel
            control={
            <Checkbox
            checked={state.checkedUser}
            onChange={handleChange}
            name="checkedUser"
            color="primary"
          />
        }
        label="User"
        />
        </Grid>
      </Grid>
    </FormContainer>
  )  
}