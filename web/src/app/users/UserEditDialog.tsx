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
import { UserRole } from '../../schema'
import { Props } from 'react-infinite-scroll-component'

const roles = [
    'admin',
    'user',
]

const query = gql`
  query($id: ID!) {
    user(id: $id) {
      name
      role
    }
  }
`
const mutation = gql`
  mutation($input: UpdateUserInput!) {
    updateUser(input: $input)
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
  const [editUser, editUserStatus] = useMutation(mutation, {
  onCompleted: props.onComplete,
  })
    
  var uRole = data?.user?.role

 if (!isSessionReady || (!data && qLoading)) return <Spinner />
 
 var userRoles = [false, true]
 if (data?.user?.role === 'admin') {
    userRoles = [true,false]
 }

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
            // TODO  
            role: uRole,
          },
        },
        })
      }     
      form={
        <FormContainer>
        <Grid container spacing={2}>
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
            //name={`userRoles[${rIdx}]`.toString()} // name expects a string
             name={role }
             mapValue={(value: boolean) => {
             console.log('In mapValue: ', value)
             if (!value) {
                return userRoles[rIdx]
             }
             return value
             }}
             // mapValue={() => roles_array[rIdx]}       
             /*mapValue={(value: boolean) => {
                //if (changed === true){ return value} // if value has come from mapOnChangeValue
                if (role === data?.user?.role) { value = true }
                else { value = false }
                return value
             }}*/
             mapOnChangeValue={(value: boolean) => {
                 userRoles[rIdx] = !value
                 if (!value == true) {
                     uRole = role
                 }
                return !value
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
      }   
    />     
  )
}

export default UserEditDialog