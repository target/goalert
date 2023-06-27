import React, { useState, useEffect } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import Button from '@mui/material/Button'
import Radio from '@mui/material/Radio'
import RadioGroup from '@mui/material/RadioGroup'
import FormControlLabel from '@mui/material/FormControlLabel'
import TextField from '@mui/material/TextField'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'

const query = gql`
  query AlertFeedbackQuery($id: Int!) {
    alert(id: $id) {
      id
      feedback {
        note
      }
    }
  }
`

const mutation = gql`
  mutation UpdateFeedbackMutation($input: UpdateAlertFeedbackInput!) {
    updateAlertFeedback(input: $input)
  }
`

interface AlertFeedbackProps {
  alertID: number
}

export default function AlertFeedback(props: AlertFeedbackProps): JSX.Element {
  const { alertID } = props
  const [showDialog, setShowDialog] = useState(false)

  const [{ data }] = useQuery({
    query,
    variables: {
      id: alertID,
    },
  })
  const [note, setNote] = useState(data?.alert?.feedback?.note ?? '')
  const [other, setOther] = useState('')
  const [mutationStatus, commit] = useMutation(mutation)

  const options = [
    'False positive',
    'Resolved itself',
    "Wasn't actionable",
    'Poor details',
  ]

  const dataNote = data?.alert?.feedback?.note ?? ''
  useEffect(() => {
    if (options.includes(dataNote)) {
      setNote(dataNote)
    } else {
      setNote('Other')
      setOther(dataNote)
    }
  }, [dataNote])

  return (
    <React.Fragment>
      <Button variant='contained' onClick={() => setShowDialog(true)}>
        Problem?
      </Button>
      {showDialog && (
        <FormDialog
          title='Problem with this alert?'
          loading={mutationStatus.fetching}
          errors={nonFieldErrors(mutationStatus.error)}
          onClose={() => setShowDialog(false)}
          onSubmit={() =>
            commit({
              input: {
                alertID,
                note: note === 'Other' ? other : note,
              },
            }).then((result) => {
              if (!result.error) setShowDialog(false)
            })
          }
          form={
            <RadioGroup
              name='controlled-radio-buttons-group'
              value={note}
              onChange={(e) => setNote(e.target.value)}
            >
              {options.map((o) => (
                <FormControlLabel
                  key={o}
                  value={o}
                  label={o}
                  control={<Radio />}
                />
              ))}
              <FormControlLabel
                value='Other'
                label={
                  <TextField
                    fullWidth
                    size='small'
                    onFocus={() => setNote('Other')}
                    value={other}
                    placeholder='Other (please specify)'
                    onChange={(e) => setOther(e.target.value)}
                  />
                }
                control={<Radio />}
                disableTypography
              />
            </RadioGroup>
          }
        />
      )}
    </React.Fragment>
  )
}
