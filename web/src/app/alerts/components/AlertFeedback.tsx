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

  const options = [
    'False positive',
    'Resolved itself',
    "Wasn't actionable",
    'Poor details',
  ]

  const dataNotes = data?.alert?.feedback?.note ?? ''

  const getDefaults = (): [Array<string>, string] => {
    const vals = dataNotes !== '' ? dataNotes.split('|') : []
    let defaultValue: Array<string> = []
    let defaultOther = ''
    vals.forEach((val: string) => {
      if (!options.includes(val)) {
        defaultOther = val
      } else {
        defaultValue = [...defaultValue, val]
      }
    })

    return [defaultValue, defaultOther]
  }

  const defaults = getDefaults()
  const [notes, setNotes] = useState<Array<string>>(defaults[0])
  const [other, setOther] = useState(defaults[1])
  const [otherChecked, setOtherChecked] = useState(Boolean(defaults[1]))
  const [mutationStatus, commit] = useMutation(mutation)

  useEffect(() => {
    const v = getDefaults()
    setNotes(v[0])
    setOther(v[1])
    setOtherChecked(Boolean(v[1]))
  }, [dataNotes])

  function handleSubmit(): void {
    let n = notes.slice()
    if (other !== '' && otherChecked) n = [...n, other]
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
              {options.map((option) => (
                <FormControlLabel
                  key={option}
                  label={option}
                  control={
                    <Checkbox
                      checked={notes.includes(option)}
                      onChange={(e) => handleCheck(e, option)}
                    />
                  }
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
