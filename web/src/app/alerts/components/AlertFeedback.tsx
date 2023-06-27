import React, { useState, useEffect } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import Button from '@mui/material/Button'
import Checkbox from '@mui/material/Checkbox'
import FormGroup from '@mui/material/FormGroup'
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

export const mutation = gql`
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

  const options = [
    'False positive',
    'Resolved itself',
    "Wasn't actionable",
    'Poor details',
  ]

  // const [{ data }] = useQuery({
  //   query,
  //   variables: {
  //     id: alertID,
  //   },
  // })
  const [notes, setNotes] = useState<Array<string>>([])
  const [other, setOther] = useState('')
  const [mutationStatus, commit] = useMutation(mutation)

  // const dataNote = data?.alert?.feedback?.note ?? ''
  // useEffect(() => {
  //   if (options.includes(dataNote)) {
  //     setNote(dataNote)
  //   } else {
  //     setNote(['Other'])
  //     setOther(dataNote)
  //   }
  // }, [dataNote])

  function handleSubmit(): void {
    commit({
      input: {
        alertID,
        note: notes.join('|'),
      },
    }).then((result) => {
      if (!result.error) setShowDialog(false)
    })
  }

  function handleCheck(
    e: React.ChangeEvent<HTMLInputElement>,
    note: string,
  ): void {
    if (e.target.checked) {
      setNotes([...notes, note])
    } else {
      setNotes(notes.filter((n) => n !== note))
    }
  }

  return (
    <React.Fragment>
      <Button variant='contained' onClick={() => setShowDialog(true)}>
        Problem?
      </Button>
      {showDialog && (
        <FormDialog
          title='Having a problem with this alert?'
          loading={mutationStatus.fetching}
          errors={nonFieldErrors(mutationStatus.error)}
          onClose={() => setShowDialog(false)}
          onSubmit={handleSubmit}
          form={
            <FormGroup>
              {options.map((o) => (
                <FormControlLabel
                  key={o}
                  label={o}
                  control={<Checkbox onChange={(e) => handleCheck(e, o)} />}
                />
              ))}
              <FormControlLabel
                value='Other'
                label={
                  <TextField
                    fullWidth
                    size='small'
                    value={other}
                    placeholder='Other (please specify)'
                    onChange={(e) => setOther(e.target.value)}
                  />
                }
                control={<Checkbox onChange={(e) => handleCheck(e, other)} />}
                disableTypography
              />
            </FormGroup>
          }
        />
      )}
    </React.Fragment>
  )
}
