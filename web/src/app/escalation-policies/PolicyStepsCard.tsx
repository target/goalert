import React, { Suspense, useEffect, useState } from 'react'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import Dialog from '@mui/material/Dialog'
import Typography from '@mui/material/Typography'
import { Add } from '@mui/icons-material'
import { gql, useMutation } from 'urql'
import FlatList from '../lists/FlatList'
import CreateFAB from '../lists/CreateFAB'
import PolicyStepCreateDialogDest from './PolicyStepCreateDialogDest'
import { useResetURLParams, useURLParam } from '../actions'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import { useIsWidthDown } from '../util/useWidth'
import { reorderList } from '../rotations/util'
import PolicyStepDeleteDialog from './PolicyStepDeleteDialog'
import PolicyStepEditDialogDest from './PolicyStepEditDialogDest'
import OtherActions from '../util/OtherActions'
import { renderChipsDest, renderDelayMessage } from './stepUtil'
import { Destination } from '../../schema'

const mutation = gql`
  mutation UpdateEscalationPolicyMutation(
    $input: UpdateEscalationPolicyInput!
  ) {
    updateEscalationPolicy(input: $input)
  }
`

type StepInfo = {
  id: string
  delayMinutes: number
  stepNumber: number
  actions: Destination[]
}

export type PolicyStepsCardProps = {
  escalationPolicyID: string
  repeat: number
  steps: Array<StepInfo>
}

export default function PolicyStepsCard(
  props: PolicyStepsCardProps,
): React.ReactNode {
  const isMobile = useIsWidthDown('md')
  const stepNumParam = 'createStep'
  const [createStep, setCreateStep] = useURLParam<boolean>(stepNumParam, false)
  const resetCreateStep = useResetURLParams(stepNumParam)

  const [stepIDs, setStepIDs] = useState<string[]>(props.steps.map((s) => s.id))

  useEffect(() => {
    setStepIDs(props.steps.map((s) => s.id))
  }, [props.steps.map((s) => s.id).join(',')]) // update steps when order changes

  const orderedSteps = stepIDs
    .map((id) => props.steps.find((s) => s.id === id))
    .filter((s) => s) as StepInfo[]

  const [editStepID, setEditStepID] = useURLParam<string>('editStep', '')
  const editStep = props.steps.find((step) => step.id === editStepID)
  const resetEditStep = useResetURLParams('editStep')
  const [deleteStep, setDeleteStep] = useState('')

  const [updateError, setUpdateError] = useState<Error | null>(null)
  const [status, commit] = useMutation(mutation)

  useEffect(() => {
    if (status.error) {
      setUpdateError(status.error)
      setStepIDs(props.steps.map((s) => s.id))
    }
  }, [status.error])

  async function onReorder(
    oldIndex: number,
    newIndex: number,
  ): Promise<unknown> {
    const newStepIDs = reorderList(stepIDs, oldIndex, newIndex)
    setStepIDs(newStepIDs)

    return commit(
      {
        input: {
          id: props.escalationPolicyID,
          stepIDs: newStepIDs,
        },
      },
      { additionalTypenames: ['EscalationPolicy'] },
    )
  }

  function renderRepeatText(): React.ReactNode {
    if (!stepIDs.length) {
      return null
    }

    let text = ''
    if (props.repeat === 0) text = 'Do not repeat'
    else if (props.repeat === 1) text = 'Repeat once'
    else text = `Repeat ${props.repeat} times`

    return (
      <Typography variant='subtitle1' component='p' sx={{ pl: 2, pb: 2 }}>
        {text}
      </Typography>
    )
  }

  return (
    <React.Fragment>
      {isMobile && (
        <CreateFAB onClick={() => setCreateStep(true)} title='Create Step' />
      )}
      {createStep && (
        <PolicyStepCreateDialogDest
          escalationPolicyID={props.escalationPolicyID}
          onClose={resetCreateStep}
        />
      )}
      <Card>
        <CardHeader
          title='Escalation Steps'
          component='h3'
          sx={{ paddingBottom: 0, margin: 0 }}
          action={
            !isMobile && (
              <Button
                variant='contained'
                onClick={() => setCreateStep(true)}
                startIcon={<Add />}
              >
                Create Step
              </Button>
            )
          }
        />
        <FlatList
          data-cy='steps-list'
          emptyMessage='No steps currently on this Escalation Policy'
          headerNote='Notify the following:'
          items={orderedSteps.map((step, idx) => ({
            id: step.id,
            disableTypography: true,
            title: (
              <Typography component='h4' variant='subtitle1' sx={{ pb: 1 }}>
                <b>Step #{idx + 1}:</b>
              </Typography>
            ) as unknown as string, // needed to work around MUI incorrect types
            subText: (
              <React.Fragment>
                {renderChipsDest(step.actions)}
                {renderDelayMessage(
                  step,
                  idx,
                  props.repeat,
                  idx === orderedSteps.length - 1,
                )}
              </React.Fragment>
            ),
            secondaryAction: (
              <OtherActions
                actions={[
                  {
                    label: 'Edit',
                    onClick: () => setEditStepID(step.id),
                  },
                  {
                    label: 'Delete',
                    onClick: () => setDeleteStep(step.id),
                  },
                ]}
              />
            ),
          }))}
          onReorder={onReorder}
        />
        {renderRepeatText()}
      </Card>
      <Dialog open={Boolean(updateError)} onClose={() => setUpdateError(null)}>
        <DialogTitleWrapper
          fullScreen={useIsWidthDown('md')}
          title='An error occurred'
        />
        <DialogContentError error={updateError?.message} />
      </Dialog>
      <Suspense>
        {editStep && (
          <React.Fragment>
            <PolicyStepEditDialogDest
              escalationPolicyID={props.escalationPolicyID}
              onClose={resetEditStep}
              stepID={editStep.id}
            />
          </React.Fragment>
        )}
        {deleteStep && (
          <PolicyStepDeleteDialog
            escalationPolicyID={props.escalationPolicyID}
            onClose={() => setDeleteStep('')}
            stepID={deleteStep}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
