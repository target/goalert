import React from 'react'
import { gql, useMutation } from '@apollo/client'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { FormContainer } from '../forms'
import Grid from '@material-ui/core/Grid'
import { Checkbox, Table, TableHead, TableRow, TableCell, TableBody, Hidden } from '@material-ui/core'
import _ from 'lodash'
import { nonFieldErrors } from '../util/errutil'


const mutation = gql`
  mutation($input: SetUserRoleInput!) {
    setUserRole(input: $input)
  }
`

interface UserEditDialogProps {
  userID: string
  role: string
  onClose: () => void  
}

function UserEditDialog(props: UserEditDialogProps): JSX.Element {
  const { ready: isSessionReady } = useSessionInfo()
    
  const [state, setState] = React.useState({
    isAdmin: props.role === 'admin'?true:false,
   });

  const [editUser, editUserStatus] = useMutation(mutation, {
   onCompleted: props.onClose,
  })
    
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setState({ ...state, [event.target.name]: event.target.checked });
  };

  if (!isSessionReady) return <Spinner />

  return ( 
    <FormDialog
      title='Edit User Role'
      confirm
      errors={nonFieldErrors(editUserStatus.error)}
      onClose={props.onClose}
      onSubmit={() =>
        editUser({
          variables: {
          input: {
            id: props.userID,
            role: state.isAdmin? 'admin':'user',
          },
        },
        })
      }
     notices={
        props.role === 'admin' && state.isAdmin === false
          ? [
              {
                type: 'WARNING',
                message: 'Updating role to User',
                details: 'This user is currently an Admin, changing the role will set the role to User instead.'
              },
            ]
          : []
     }    
      form={
        <FormContainer>
        <Grid container spacing={2}>
        <Grid item xs={12}>
        <Table data-cy='user-roles'>
        <TableHead>
        <TableRow>
        <TableCell padding='checkbox'>
            admin
        </TableCell>                          
        </TableRow>
        </TableHead>
        <TableBody>
        <Hidden smDown>
        <TableCell padding='checkbox'>   
         <Checkbox
            checked={state.isAdmin}
            onChange={ handleChange }
            name='isAdmin'
          />                                                       
        </TableCell>                       
        </Hidden>
        </TableBody>              
        </Table>              
        </Grid>
      </Grid>
    </FormContainer>
      }   
    />     
  )
}

export default UserEditDialog
