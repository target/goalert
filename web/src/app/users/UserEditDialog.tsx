import React from 'react'
import { gql, useQuery, useMutation } from '@apollo/client'
import Spinner from '../loading/components/Spinner'
import FormDialog from '../dialogs/FormDialog'
import { useSessionInfo } from '../util/RequireConfig'
import { FormContainer, FormField } from '../forms'
import Grid from '@material-ui/core/Grid'
import { Checkbox, Table, TableHead, TableRow, TableCell, TableBody, Hidden } from '@material-ui/core'
import _ from 'lodash'

const roles = [
    'admin',
    'user',
]

// const roles_array = [true, false]

const query = gql`
  query($id: ID!) {
    user(id: $id) {
      name
      role
    }
  }
`
const mutation = gql`
  mutation updateUser($input: UpdateUserInput!) {
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
  const [editUser, { loading: mLoading, error }] = useMutation(mutation, {
    variables: {
      input: [
        {
          id: props.userID,
          type: 'user',
        },
      ],
    },
  })

  if (!isSessionReady || (!data && qLoading)) return <Spinner />

  return (
    <FormDialog
      title='Edit User Role'
      confirm
      subTitle={`This will edit this user's role: ${data?.user?.name}`}
      loading={mLoading}
      errors={error ? [error] : []}
      onClose={props.onClose}
      onSubmit={() => editUser()}
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
             name={role} // name expects a string  
             // mapValue={() => roles_array[rIdx]}       
                mapValue={(value: boolean, changed: boolean) => {
                console.log('Changed :',changed)
                if (changed === true){ return value} // if value has come from mapOnChangeValue
                if (role === data?.user?.role) { value = true }
                else { value = false }
                return value
             }}
             mapOnChangeValue={(value: boolean, changed: boolean) => {
                console.log(value)
                changed = true 
                return !value
             }}    
                    
             /*mapOnChangeValue={(value: string, formValue: Value) => {
                 if (formValue.role != data?.user?.role){
                      return formValue.role
                  }
             }}*/
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