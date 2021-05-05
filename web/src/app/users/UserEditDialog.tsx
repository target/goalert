import React from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { FormContainer, FormField } from '../forms'
import Grid from '@material-ui/core/Grid'
import { Checkbox, Table, TableHead, TableRow, TableCell, TableBody, Hidden } from '@material-ui/core'
import _ from 'lodash'
import { nonFieldErrors } from '../util/errutil'
import FormControlLabel from '@material-ui/core/FormControlLabel'


const query = gql`
  query($id: ID!) {
    user(id: $id) {
      name
      role
    }
  }
`
const mutation = gql`
  mutation($input: SetUserRoleInput!) {
    setUserRole(input: $input)
  }
`

interface UserEditDialogProps {
  userID: string
  onClose: () => void  
}

function UserEditDialog(props: UserEditDialogProps): JSX.Element {
  const { ready: isSessionReady } = useSessionInfo()
    
  const { data, loading: qLoading } = useQuery(query, {
    variables: { id: props.userID },
  })
    
  const [state, setState] = React.useState({
    isAdmin: data?.user?.role === 'admin'?true:false,
   });
  
  const [editUser, editUserStatus] = useMutation(mutation, {
   onCompleted: props.onClose,
  })
    
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setState({ ...state, [event.target.name]: event.target.checked });
  };

 if (!isSessionReady || (!data && qLoading)) return <Spinner />
 // state.isAdmin = data?.user?.role === 'admin'?true:false
 /*var admin = false
 if (data?.user?.role === 'admin') {
    admin = true       
 }*/

  return ( 
    <FormDialog
      title='Edit User Role'
      confirm
      subTitle={`This will edit this user's role: ${data?.user?.name}`}
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
        state.isAdmin === false
          ? [
              {
                type: 'WARNING',
                message: 'The user role is USER',
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
