import React, { useState, useEffect } from 'react'
import { useQuery, useMutation, gql } from 'urql'
import Button from '@mui/material/Button'
import Checkbox from '@mui/material/Checkbox'
import FormGroup from '@mui/material/FormGroup'
import FormControlLabel from '@mui/material/FormControlLabel'
import TextField from '@mui/material/TextField'
import FormDialog from '../../dialogs/FormDialog'
import { nonFieldErrors } from '../../util/errutil'
import { Card, CardContent, CardHeader, Typography } from '@mui/material'
import ErrorBoundary from '../../main/ErrorBoundary'
import { DEBOUNCE_DELAY } from '../../config'
import { AlertStatus } from '../../../schema'
import CardActions from '../../details/CardActions'

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

  const [{ data }] = useQuery({
    query,
    variables: {
      id: alertID,
    },
  })

  const options = [
    'False positive',
    "Wasn't Actionable",
    'Resolved itself',
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
  const { error } = mutationStatus

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
      <Card>
        <CardHeader title='Is this alert noise?' />
        <CardContent sx={{ pt: 0 }}>
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
                    if (!e.target.checked) {
                      setOther('')
                    }
                  }}
                />
              }
              disableTypography
            />
          </FormGroup>
          {error?.message && (
            <Typography color='error' sx={{ pt: 2 }}>
              {error?.message}
            </Typography>
          )}
        </CardContent>
        <CardActions
          primaryActions={[
            <Button key='submit' variant='contained' onClick={handleSubmit}>
              Submit
            </Button>,
          ]}
        />
      </Card>
    </React.Fragment>
  )
}
