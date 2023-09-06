import React, { useState } from 'react'
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

interface AlertFeedbackDialogProps {
  open: boolean
  onClose: () => void
  alertIDs: Array<string>
}

export default function AlertFeedbackDialog(
  props: AlertFeedbackDialogProps,
): JSX.Element {
  const { alertIDs, open, onClose } = props

  const [noiseReasons, setNoiseReasons] = useState<Array<string>>([])
  const [other, setOther] = useState('')
  const [otherChecked, setOtherChecked] = useState(false)

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

  return (
    <Dialog open={open} onClose={onClose}>
      <DialogTitle>Mark Selected Alerts as Noise</DialogTitle>
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
      </DialogContent>
      <DialogActions>
        <Button>Submit</Button>
      </DialogActions>
    </Dialog>
  )
}
