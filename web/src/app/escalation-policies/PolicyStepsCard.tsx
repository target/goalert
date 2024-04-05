import React, { Suspense, useRef, useState } from 'react'
import Button from '@mui/material/Button'
import Card from '@mui/material/Card'
import CardHeader from '@mui/material/CardHeader'
import Dialog from '@mui/material/Dialog'
import Typography from '@mui/material/Typography'
import { Add } from '@mui/icons-material'
import { gql, useMutation } from '@apollo/client'
import FlatList from '../lists/FlatList'
import CreateFAB from '../lists/CreateFAB'
import PolicyStepCreateDialog from './PolicyStepCreateDialog'
import PolicyStepCreateDialogDest from './PolicyStepCreateDialogDest'
import { useResetURLParams, useURLParam } from '../actions'
import DialogTitleWrapper from '../dialogs/components/DialogTitleWrapper'
import DialogContentError from '../dialogs/components/DialogContentError'
import { policyStepsQuery } from './PolicyStepsQuery'
import { useIsWidthDown } from '../util/useWidth'
import { reorderList } from '../rotations/util'
import PolicyStepEditDialog from './PolicyStepEditDialog'
import PolicyStepDeleteDialog from './PolicyStepDeleteDialog'
import PolicyStepEditDialogDest from './PolicyStepEditDialogDest'
import OtherActions from '../util/OtherActions'
import {
  getStepNumber,
  renderChips,
  renderChipsDest,
  renderDelayMessage,
} from './stepUtil'
import { useExpFlag } from '../util/useExpFlag'
import { Destination, EscalationPolicy, Target } from '../../schema'

const mutation = gql`
  mutation UpdateEscalationPolicyMutation(
    $input: UpdateEscalationPolicyInput!
  ) {
    updateEscalationPolicy(input: $input)
  }
`

export type PolicyStepsCardProps = {
  escalationPolicyID: string
  repeat: number
  steps: Array<{
    id: string
    delayMinutes: number
    stepNumber: number
    actions?: Destination[]
    targets: Target[]
  }>
}

export default function PolicyStepsCard(
  props: PolicyStepsCardProps,
): React.ReactNode {
  const hasDestTypesFlag = useExpFlag('dest-types')

  const { escalationPolicyID, repeat, steps = [] } = props

  const isMobile = useIsWidthDown('md')
  const stepNumParam = 'createStep'
  const [createStep, setCreateStep] = useURLParam<boolean>(stepNumParam, false)
  const resetCreateStep = useResetURLParams(stepNumParam)

  const oldID = useRef(null)
  const oldIdx = useRef(null)
  const newIdx = useRef(null)

  type Swap = { oldIndex: number; newIndex: number }
  const [lastSwap, setLastSwap] = useState<Array<Swap>>([])

  const [error, setError] = useState<Error | null>(null)

  const [editStepID, setEditStepID] = useURLParam<string>('editStep', '')
  const editStep = steps.find((step) => step.id === editStepID)
  const resetEditStep = useResetURLParams('editStep')
  const [deleteStep, setDeleteStep] = useState('')

  const [updateEscalationPolicy] = useMutation(mutation, {
    onCompleted: () => {
      oldID.current = null
      oldIdx.current = null
      newIdx.current = null
    },
    onError: (err) => setError(err),
  })

  async function onReorder(
    oldIndex: number,
    newIndex: number,
  ): Promise<unknown> {
    setLastSwap(lastSwap.concat({ oldIndex, newIndex }))

    const updatedStepIDs = reorderList(
      steps.map((step) => step.id),
      oldIndex,
      newIndex,
    )

    return updateEscalationPolicy({
      variables: {
        input: {
          id: escalationPolicyID,
          stepIDs: updatedStepIDs,
        },
      },
      update: (cache, { data }) => {
        // mutation returns true on a success
        if (!data.updateEscalationPolicy) {
          return
        }

        // get the current state of the steps in the cache
        const cacheData = cache.readQuery<{
          escalationPolicy: EscalationPolicy
        }>({
          query: policyStepsQuery,
          variables: { id: escalationPolicyID },
        })
        if (!cacheData) throw new Error('Cache data not found')
        const escalationPolicy = cacheData.escalationPolicy
        const steps = escalationPolicy.steps.slice()

        if (steps.length > 0) {
          const newSteps = reorderList(steps, oldIndex, newIndex)

          // write new steps order to cache
          cache.writeQuery({
            query: policyStepsQuery,
            variables: { id: escalationPolicyID },
            data: {
              escalationPolicy: {
                ...escalationPolicy,
                steps: newSteps,
              },
            },
          })
        }
      },
      optimisticResponse: {
        __typename: 'Mutation',
        updateEscalationPolicy: true,
      },
    })
  }

  function renderRepeatText(): React.ReactNode {
    if (!steps.length) {
      return null
    }

    let text = ''
    if (repeat === 0) text = 'Do not repeat'
    else if (repeat === 1) text = 'Repeat once'
    else text = `Repeat ${repeat} times`

    return (
      <Typography variant='subtitle1' component='p' sx={{ pl: 2, pb: 2 }}>
        {text}
      </Typography>
    )
  }

  const { message: errMsg } = error || {}

  return (
    <React.Fragment>
      {isMobile && (
        <CreateFAB onClick={() => setCreateStep(true)} title='Create Step' />
      )}
      {createStep && (
        <React.Fragment>
          {hasDestTypesFlag ? (
            <PolicyStepCreateDialogDest
              escalationPolicyID={escalationPolicyID}
              onClose={resetCreateStep}
            />
          ) : (
            <PolicyStepCreateDialog
              escalationPolicyID={escalationPolicyID}
              onClose={resetCreateStep}
            />
          )}
        </React.Fragment>
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
          items={steps.map((step) => ({
            id: step.id,
            disableTypography: true,
            title: (
              <Typography component='h4' variant='subtitle1' sx={{ pb: 1 }}>
                <b>Step #{getStepNumber(step.id, steps)}:</b>
              </Typography>
            ) as unknown as string, // needed to work around MUI incorrect types
            subText: (
              <React.Fragment>
                {step.actions
                  ? renderChipsDest(step.actions)
                  : renderChips(step)}
                {renderDelayMessage(steps, step, repeat)}
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
      <Dialog open={Boolean(error)} onClose={() => setError(null)}>
        <DialogTitleWrapper
          fullScreen={useIsWidthDown('md')}
          title='An error occurred'
        />
        <DialogContentError error={errMsg} />
      </Dialog>
      <Suspense>
        {editStep && (
          <React.Fragment>
            {hasDestTypesFlag ? (
              <PolicyStepEditDialogDest
                escalationPolicyID={escalationPolicyID}
                onClose={resetEditStep}
                stepID={editStep.id}
              />
            ) : (
              <PolicyStepEditDialog
                escalationPolicyID={escalationPolicyID}
                onClose={resetEditStep}
                step={editStep}
              />
            )}
          </React.Fragment>
        )}
        {deleteStep && (
          <PolicyStepDeleteDialog
            escalationPolicyID={escalationPolicyID}
            onClose={() => setDeleteStep('')}
            stepID={deleteStep}
          />
        )}
      </Suspense>
    </React.Fragment>
  )
}
