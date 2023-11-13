import React, { useState, useContext } from 'react'
import { useMutation, gql } from 'urql'
import Button from '@mui/material/Button'
import Checkbox from '@mui/material/Checkbox'
import Dialog from '@mui/material/Dialog'
import DialogContent from '@mui/material/DialogContent'
import DialogTitle from '@mui/material/DialogTitle'
import FormGroup from '@mui/material/FormGroup'
import FormControlLabel from '@mui/material/FormControlLabel'
import TextField from '@mui/material/TextField'
import { options } from './AlertFeedback'
import { DialogActions, Typography } from '@mui/material'
import { NotificationContext } from '../../main/SnackbarNotification'

const updateMutation = gql`
  mutation UpdateAlertsMutation($input: UpdateAlertsInput!) {
    updateAlerts(input: $input) {
      status
      id
    }
  }
`

interface AlertFeedbackDialogProps {
  open: boolean
  onClose: () => void
  alertIDs: Array<string>
}

export default function AlertFeedbackDialog(
  props: AlertFeedbackDialogProps,
): React.ReactNode {
  const { alertIDs, open, onClose } = props

  const [noiseReasons, setNoiseReasons] = useState<Array<string>>([])
  const [other, setOther] = useState('')
  const [otherChecked, setOtherChecked] = useState(false)
  const [mutationStatus, commit] = useMutation(updateMutation)
  const { error } = mutationStatus

  const { setNotification } = useContext(NotificationContext)

  function handleCheck(
    e: React.ChangeEvent<HTMLInputElement>,
    noiseReason: string,
  ): void {
    if (e.target.checked) {
      setNoiseReasons([...noiseReasons, noiseReason])
    } else {
      setNoiseReasons(noiseReasons.filter((n) => n !== noiseReason))
    }
  }

  function handleSubmit(): void {
    let n = noiseReasons.slice()
    if (other !== '' && otherChecked) n = [...n, other]
    commit({
      input: {
        alertIDs,
        noiseReason: n.join('|'),
      },
    }).then((result) => {
      const numUpdated = result.data.updateAlerts.length
      const count = alertIDs.length

      const msg = `${numUpdated} of ${count} alert${
        count === 1 ? '' : 's'
      } updated`

      setNotification({
        message: msg,
        severity: 'info',
      })

      onClose()
    })
  }

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle data-cy='dialog-title'>
        Mark the Following Alerts as Noise
      </DialogTitle>
      <DialogContent>
        <Typography sx={{ pb: 1 }}>Alerts: {alertIDs.join(', ')}</Typography>
        <FormGroup>
          {options.map((option) => (
            <FormControlLabel
              key={option}
              label={option}
              control={
                <Checkbox
                  data-cy={option}
                  checked={noiseReasons.includes(option)}
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
      </DialogContent>
      <DialogActions>
        <Button
          aria-label='Submit noise reasons'
          variant='contained'
          onClick={handleSubmit}
          disabled={!noiseReasons.length && !other}
        >
          Submit
        </Button>
      </DialogActions>
    </Dialog>
  )
}
