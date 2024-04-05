import React, { useState } from 'react'

import FormDialog from '../dialogs/FormDialog'
import { DestinationInput, FieldValueInput } from '../../schema'
import { FormContainer } from '../forms'
import { Grid, InputAdornment, TextField, Typography } from '@mui/material'
import { renderMenuItem } from '../selection/DisableableMenuItem'
import DestinationField from '../selection/DestinationField'

export type Action = {
  dest: DestinationInput

  params: FieldValueInput[]
}

export default function RuleEditorActionDialog(props: {
  onClose: (expr: string | null) => void
}): JSX.Element {
  const [value, setValue] = useState<string>(props.expr)
  const [actionType, setActionType] = useState('create-alert')
  const [destValues, setDestValues] = useState<FieldValueInput[]>([])
  const [slackParam, setSlackParam] = useState('body.message')

  return (
    <FormDialog
      maxWidth='sm'
      title='Edit Action'
      onClose={() => props.onClose(null)}
      onSubmit={() => props.onClose(value)}
      form={
        <FormContainer>
          <Grid container spacing={2}>
            <Grid item xs={12}>
              <TextField
                fullWidth
                select
                label='Action Type'
                value={actionType}
                onChange={(e) => setActionType(e.target.value)}
              >
                {renderMenuItem({
                  value: 'create-alert',
                  label: 'Create/Update Alert',
                  disabled: false,
                  disabledMessage: '',
                })}
                {renderMenuItem({
                  value: 'send-slack',
                  label: 'Send Slack Message',
                  disabled: false,
                  disabledMessage: '',
                })}
                {renderMenuItem({
                  value: 'drop',
                  label: 'Drop/Ignore Request',
                  disabled: false,
                  disabledMessage: '',
                })}
              </TextField>
            </Grid>

            {actionType === 'drop' && (
              <Grid item xs={12}>
                <Typography>The request will be dropped.</Typography>
              </Grid>
            )}
            {actionType === 'send-slack' && (
              <React.Fragment>
                <Grid item xs={12}>
                  <DestinationField
                    value={destValues}
                    destType='builtin-slack-channel'
                    onChange={(v) => setDestValues(v)}
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    multiline
                    InputProps={{
                      startAdornment: (
                        <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                          Message:
                        </InputAdornment>
                      ),
                      endAdornment: (
                        <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                          string
                        </InputAdornment>
                      ),
                    }}
                    value={slackParam}
                    onChange={(e) => setSlackParam(e.target.value)}
                  />
                </Grid>
              </React.Fragment>
            )}

            {actionType === 'create-alert' && (
              <React.Fragment>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    multiline
                    InputProps={{
                      startAdornment: (
                        <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                          Summary:
                        </InputAdornment>
                      ),
                      endAdornment: (
                        <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                          string
                        </InputAdornment>
                      ),
                    }}
                    value='body.title'
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    multiline
                    InputProps={{
                      startAdornment: (
                        <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                          Details:
                        </InputAdornment>
                      ),
                      endAdornment: (
                        <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                          string
                        </InputAdornment>
                      ),
                    }}
                    value='body.details'
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    multiline
                    InputProps={{
                      startAdornment: (
                        <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                          Dedup:
                        </InputAdornment>
                      ),
                      endAdornment: (
                        <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                          string
                        </InputAdornment>
                      ),
                    }}
                    value='body.title + body.details'
                  />
                </Grid>
                <Grid item xs={12}>
                  <TextField
                    fullWidth
                    multiline
                    InputProps={{
                      startAdornment: (
                        <InputAdornment position='start' sx={{ mb: '0.1em' }}>
                          Close Alert?:
                        </InputAdornment>
                      ),
                      endAdornment: (
                        <InputAdornment position='end' sx={{ mb: '0.1em' }}>
                          bool
                        </InputAdornment>
                      ),
                    }}
                    value='body.state == "firing"'
                  />
                </Grid>
              </React.Fragment>
            )}
          </Grid>
        </FormContainer>
      }
    />
  )
}
