import { gql, useQuery, useMutation } from '@apollo/client'
import React, { useState } from 'react'
import p from 'prop-types'

import { fieldErrors, nonFieldErrors } from '../util/errutil'
import Checkbox from '@material-ui/core/Checkbox'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import FormControl from '@material-ui/core/FormControl'
import FormHelperText from '@material-ui/core/FormHelperText'
import Typography from '@material-ui/core/Typography'
import FormDialog from '../dialogs/FormDialog'
import Spinner from '../loading/components/Spinner'
import _ from 'lodash-es'

function DeleteForm({ epName, error, value, onChange }) {
  return (
    <FormControl error={Boolean(error)} style={{ width: '100%' }}>
      <FormControlLabel
        control={
          <Checkbox
            checked={value}
            onChange={(e) => onChange(e.target.checked)}
            value='delete-escalation-policy'
          />
        }
        label={
          <React.Fragment>
            Also delete escalation policy: {epName}
          </React.Fragment>
        }
      />
      <FormHelperText>{error}</FormHelperText>
    </FormControl>
  )
}
DeleteForm.propTypes = {
  epName: p.node.isRequired,
  error: p.string,
  value: p.bool,
  onChange: p.func.isRequired,
}

const query = gql`
  query service($id: ID!) {
    service(id: $id) {
      id
      name
      ep: escalationPolicy {
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
  const { data, ...dataStatus } = useQuery(query, {
    variables: { id: serviceID },
  })
  const input = [{ type: 'service', id: serviceID }]
  const [deleteService, deleteServiceStatus] = useMutation(mutation, {
    variables: { input },
  })

  const epID = _.get(data, 'service.ep.id')
  const epName = _.get(
    data,
    'service.ep.name',
    <Spinner text='fetching policy...' />,
  )

  if (deleteEP) {
    input.push({
      type: 'escalationPolicy',
      id: epID,
    })
  }

  return (
    <FormDialog
      title='Are you sure?'
      confirm
      subTitle={
        <Typography>
          This will delete the service:{' '}
          {_.get(data, 'service.name', <Spinner text='loading...' />)}
        </Typography>
      }
      caption='Deleting a service will also delete all associated integration keys and alerts.'
      loading={deleteServiceStatus.loading || (!data && dataStatus.loading)}
      errors={nonFieldErrors(deleteServiceStatus.error)}
      onClose={onClose}
      onSubmit={() => deleteService()}
      form={
        <DeleteForm
          epName={epName}
          error={
            fieldErrors(deleteServiceStatus.error).find(
              (f) => f.field === 'escalationPolicyID',
            ) && 'Escalation policy is currently in use.'
          }
          onChange={(deleteEP) => setDeleteEP(deleteEP)}
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
