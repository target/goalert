import React, { useState } from 'react'
import { FormContainer } from '../../forms'
import { Button, Grid, Typography } from '@mui/material'
import { ActionInput } from '../../../schema'
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

interface DefaultActionFormProps {
  value: ActionInput[]
  onChange: (val: ActionInput[]) => void
  default?: boolean
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

export default function DefaultActionForm(
  props: DefaultActionFormProps,
): JSX.Element {
  const types = useDynamicActionTypes()

  const [currentAction, setCurrentAction] = useState<ActionValue>(
    defaults(types[0]),
  )
  const [addActionError, setAddActionError] = useState<CombinedError>()

  const validationClient = useClient()

  function handleDelete(a: ActionInput): void {
    if (!props.onChange) return
    props.onChange(props.value.filter((b) => a !== b))
  }

  return (
    <FormContainer
      value={{ value: props.value }}
      onChange={props.onChange}
      errors={fieldErrors(addActionError)}
    >
      <Grid container spacing={2}>
        <Grid item xs={12}>
          <Typography variant='h6' color='textPrimary'>
            Actions
          </Typography>
        </Grid>
        <Grid item xs={12} container spacing={1} sx={{ p: 1 }}>
          {props.value.map((a) => (
            <Grid item key={JSON.stringify(a.dest)}>
              <DestinationInputChip
                value={{ type: a.dest.type, values: [] }}
                onDelete={() => handleDelete(a)}
              />
            </Grid>
          ))}
          {props.value.length === 0 && (
            <Grid item xs={12}>
              <Typography variant='body2' color='textSecondary'>
                No actions
              </Typography>
            </Grid>
          )}
        </Grid>

        <Grid item xs={12} container spacing={2}>
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
                    console.info(currentAction)
                    // clear the current action
                    setCurrentAction(
                      defaults(
                        types.find((t) => t.type === currentAction.destType) ||
                          types[0],
                      ),
                    )

                    props.onChange(props.value.concat(act))
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
