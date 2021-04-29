import React from 'react'
import Grid from '@material-ui/core/Grid'
import TextField from '@material-ui/core/TextField'
import { Checkbox, Table, TableHead, TableRow, TableCell, TableBody, Hidden } from '@material-ui/core'
import { FormContainer, FormField } from '../forms'

const roles = [
    'admin',
    'user',
]

interface Value {
  name: string
  role: string
 // isAdmin: boolean
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
   // checkedUser: !props.value.isAdmin,
   // checkedAdmin: props.value.isAdmin,
    });
    const { ...containerProps } = props

    /*const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setState({ ...state, [event.target.name]: event.target.checked });
    };*/

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
        <Table data-cy='user-roles'>
        <TableHead>
        <TableRow>             
        {roles.map((r) => (
        <TableCell key={r} padding='checkbox'>
            {r}
        </TableCell>
        ))}    
        </TableRow>
        </TableHead>
        <TableBody>
        <Hidden smDown>
        {roles.map((role, rIdx) => (
        <TableCell key={rIdx} padding='checkbox'>
            <FormField
             noError
             component={Checkbox}
             checkbox
             fieldName={role}
             name={role}
             // mapValue={() => roles.filter(r => r === props.value.role)}
              mapValue={() => {
              if (role === props.value.role) return true
              return false
            }}       
            />
        </TableCell>    
        ))}                      
        </Hidden>
        </TableBody>              
        </Table>              
        </Grid>
      </Grid>
    </FormContainer>
  )
}