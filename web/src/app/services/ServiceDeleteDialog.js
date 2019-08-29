import React, { useState } from 'react'
import p from 'prop-types'
import { useDispatch } from 'react-redux'
import gql from 'graphql-tag'
import { useQuery, useMutation } from 'react-apollo'
import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Checkbox from '@material-ui/core/Checkbox'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import { push } from 'connected-react-router'

function DeleteForm({ epName, error, value, onChange }) {
  return (
    <FormControl error={Boolean(error)} style={{ width: '100%' }}>
      <FormControlLabel
        control={
          <Checkbox
            checked={value}
            onChange={e => onChange(e.target.checked)}
            value='delete-escalation-policy'
          />
        }
        label={`Also delete escalation policy: ${epName}`}
      />
      <FormHelperText>{error}</FormHelperText>
    </FormControl>
  )
}
DeleteForm.propTypes = {
  epName: p.string.isRequired,
  error: p.string,
  value: p.bool,
  onChange: p.func.isRequired,
}

const query = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
      escalationPolicyID
      escalationPolicy {
        id
        name
      }
    }
  }
`
const mutation = gql`
  mutation delete($input: [TargetInput!]!) {
    deleteAll(input: $input)
  }
`

export default function ServiceDeleteDialog({ serviceID, onClose }) {
  const [deleteEP, setDeleteEP] = useState(true)
  const { data, loading: dataLoading } = useQuery(query, {
    variables: { id: serviceID },
  })
  const input = [{ type: 'service', id: serviceID }]
  const dispatch = useDispatch()
  const refetch = ['servicesQuery']
  const [deleteService, { loading, error }] = useMutation(mutation, {
    variables: { input },
    awaitRefetchQueries: true,
    refetchQueries: refetch,
    onCompleted: () => dispatch(push('/services')),
  })

  if (dataLoading) return <Spinner />

  if (deleteEP) {
    refetch.push('epsQuery')
    input.push({
      type: 'escalationPolicy',
      id: data.service.escalationPolicyID,
    })
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={`This will delete the service: ${data.service.name}`}
      caption='Deleting a service will also delete all associated integration keys and alerts.'
      loading={loading}
      errors={nonFieldErrors(error)}
      onClose={onClose}
      onSubmit={() => deleteService()}
      form={
        <DeleteForm
          epName={data.service.escalationPolicy.name}
          error={
            fieldErrors(error).find(f => f.field === 'escalationPolicyID') &&
            'Escalation policy is currently in use.'
          }
          onChange={deleteEP => setDeleteEP(deleteEP)}
          value={deleteEP}
        />
      }
    />
  )
}
ServiceDeleteDialog.propTypes = {
  serviceID: p.string.isRequired,
  onClose: p.func,
}
