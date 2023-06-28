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

  const [{ data }] = useQuery({
    query,
    variables: {
      id: alertID,
    },
  })

  const dataNotes = data?.alert?.feedback?.note ?? ''
  const defaultValue = dataNotes !== '' ? dataNotes.split('|') : []

  // TODO: set "other" from default value

  const [notes, setNotes] = useState<Array<string>>(defaultValue)
  const [other, setOther] = useState('')
  const [otherChecked, setOtherChecked] = useState(false)
  const [mutationStatus, commit] = useMutation(mutation)

  useEffect(() => {
    setNotes(defaultValue)
    // if (options.includes(dataNote)) {
    //   setNote(dataNote)
    // } else {
    //   setNote(['Other'])
    //   setOther(dataNote)
    // }
  }, [dataNotes])

  function handleSubmit(): void {
    let n = notes.slice()
    if (other !== '') n = [...n, other]
    commit({
      input: {
        alertID,
        note: n.join('|'),
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
              <FormControlLabel
                label='False positive'
                control={
                  <Checkbox
                    checked={notes.includes('False positive')}
                    onChange={(e) => handleCheck(e, 'False positive')}
                  />
                }
              />
              <FormControlLabel
                label='Resolved itself'
                control={
                  <Checkbox
                    checked={notes.includes('Resolved itself')}
                    onChange={(e) => handleCheck(e, 'Resolved itself')}
                  />
                }
              />
              <FormControlLabel
                label="Wasn't actionable"
                control={
                  <Checkbox
                    checked={notes.includes("Wasn't actionable")}
                    onChange={(e) => handleCheck(e, "Wasn't actionable")}
                  />
                }
              />
              <FormControlLabel
                label='Poor details'
                control={
                  <Checkbox
                    checked={notes.includes('Poor details')}
                    onChange={(e) => handleCheck(e, 'Poor details')}
                  />
                }
              />
              <FormControlLabel
                value='Other'
                label={
                  <TextField
                    fullWidth
                    size='small'
                    value={other}
                    placeholder='Other (please specify)'
                    onFocus={() => setOtherChecked(true)}
                    onChange={(e) => {
                      setOther(e.target.value)
                    }}
                  />
                }
                control={
                  <Checkbox
                    checked={otherChecked}
                    onChange={(e) => {
                      setOtherChecked(e.target.checked)
                      if (!e) {
                        setOther('')
                      }
                    }}
                  />
                }
                disableTypography
              />
            </FormGroup>
          }
        />
      )}
    </React.Fragment>
  )
}
