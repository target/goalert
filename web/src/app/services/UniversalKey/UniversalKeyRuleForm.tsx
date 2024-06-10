import React, { useState } from 'react'
import { FormContainer, FormField } from '../../forms'
import {
  Button,
  Divider,
  FormControl,
  FormControlLabel,
  FormLabel,
  Grid,
  Radio,
  RadioGroup,
  TextField,
  Typography,
} from '@mui/material'
import { ActionInput, KeyRuleInput } from '../../../schema'
import { useDynamicActionTypes } from '../../util/RequireConfig'
import DestinationInputChip from '../../util/DestinationInputChip'
import { CombinedError, gql, useClient } from 'urql'
import DynamicActionField, {
  Value as ActionValue,
  defaults,
  valueToActionInput,
} from '../../selection/DynamicActionField'
import { Add } from '@mui/icons-material'
import { fieldErrors } from '../../util/errutil'

interface UniversalKeyRuleFormProps {
  value: KeyRuleInput
  onChange: (val: KeyRuleInput) => void
}

const query = gql`
  query DestDisplayInfo($input: DestinationInput!) {
    destinationDisplayInfo(input: $input) {
      text
      iconURL
      iconAltText
      linkURL
    }
  }
`

export default function UniversalKeyRuleForm(
  props: UniversalKeyRuleFormProps,
): JSX.Element {
  const types = useDynamicActionTypes()

  const [currentAction, setCurrentAction] = useState<ActionValue>(
    defaults(types[0]),
  )
  const [addActionError, setAddActionError] = useState<CombinedError>()
  const [stopOrContinue, setStopOrContinue] = useState<'stop' | 'continue'>(
    'continue',
  )

  const validationClient = useClient()

  function handleDelete(a: ActionInput): void {
    if (!props.onChange) return
    props.onChange({
      ...props.value,
      actions: props.value.actions.filter((b) => a !== b),
    })
  }

  return (
    <FormContainer
      value={props.value}
      onChange={props.onChange}
      errors={fieldErrors(addActionError)}
    >
      <Grid container justifyContent='space-between' spacing={2}>
        <Grid
          item
          xs={12}
          md={5.8}
          container
          spacing={2}
          alignContent='flex-start'
        >
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              label='Name'
              name='name'
              required
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              label='Description'
              name='description'
            />
          </Grid>
          <Grid item xs={12}>
            <FormField
              fullWidth
              component={TextField}
              label='Condition (Expr syntax)'
              name='conditionExpr'
              required
              multiline
              rows={3}
            />
          </Grid>
          <Grid item xs={12} sx={{ mb: -1 }}>
            <Typography variant='h6' color='textPrimary'>
              Actions
            </Typography>
          </Grid>
          <Grid item xs={12} container spacing={1} sx={{ p: 1 }}>
            {props.value.actions.map((a) => (
              <Grid item key={JSON.stringify(a.dest)}>
                <DestinationInputChip
                  value={a.dest}
                  onDelete={() => handleDelete(a)}
                />
              </Grid>
            ))}
            {props.value.actions.length === 0 && (
              <Grid item xs={12}>
                <Typography variant='body2' color='textSecondary'>
                  No actions
                </Typography>
              </Grid>
            )}
          </Grid>
          <Grid item xs={12}>
            <FormControl>
              <FormLabel id='demo-row-radio-buttons-group-label'>
                After actions complete:
              </FormLabel>
              <RadioGroup row name='stop-or-continue' value={stopOrContinue}>
                <FormControlLabel
                  value={'continue'}
                  onChange={() => setStopOrContinue('continue')}
                  control={<Radio />}
                  label='Continue processing rules'
                />
                <FormControlLabel
                  value='stop'
                  onChange={() => setStopOrContinue('stop')}
                  control={<Radio />}
                  label='Stop at this rule'
                />
              </RadioGroup>
            </FormControl>
          </Grid>
        </Grid>

        <Grid item sx={{ width: 'fit-content' }}>
          <Divider orientation='vertical' />
        </Grid>

        <Grid item xs={12} md={5.8} container spacing={2}>
          <DynamicActionField
            value={currentAction}
            onChange={setCurrentAction}
          />

          <Grid
            item
            xs={12}
            sx={{
              display: 'flex',
              // justifyContent: 'flex-end',
              alignItems: 'flex-end',
            }}
          >
            <Button
              fullWidth
              startIcon={<Add />}
              variant='contained'
              color='secondary'
              sx={{ height: 'fit-content' }}
              onClick={() => {
                const act = valueToActionInput(currentAction)
                validationClient
                  .query(query, {
                    input: act.dest,
                  })
                  .toPromise()
                  .then((res) => {
                    if (res.error) {
                      setAddActionError(res.error) // todo: not showing in dialog?
                      console.log(res)
                      return
                    }

                    // clear the current action
                    setCurrentAction(
                      defaults(
                        types.find((t) => t.type === currentAction.destType) ||
                          types[0],
                      ),
                    )

                    props.onChange({
                      ...props.value,
                      actions: props.value.actions.concat(act),
                    })
                  })
              }}
            >
              Add Action
            </Button>
          </Grid>
        </Grid>
      </Grid>
    </FormContainer>
  )
}
