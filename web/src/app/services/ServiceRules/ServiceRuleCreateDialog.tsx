import React, { useState } from 'react'
import { useMutation, gql } from 'urql'

import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import ServiceRuleForm, {
  ServiceRuleAction,
  ServiceRuleValue,
} from './ServiceRuleForm'
import {
  ServiceRuleFilterInput,
  ServiceRuleFilterValueType,
} from '../../../schema'

const mutation = gql`
  mutation ($input: CreateServiceRuleInput!) {
    createServiceRule(input: $input) {
      name
      serviceID
      sendAlert
      filters {
        field
        operator
        value
        valueType
      }
      integrationKeys {
        id
        name
      }
    }
  }
`

export default function ServiceRuleCreateDialog(props: {
  serviceID: string
  onClose: () => void
}): JSX.Element {
  const { serviceID, onClose } = props
  const [value, setValue] = useState<ServiceRuleValue>({
    name: '',
    serviceID,
    filters: [],
    sendAlert: false,
    actions: [],
    integrationKeys: [],
  })
  const [actionsError, setActionsError] = useState<boolean>(false)

  const [createRuleStatus, commit] = useMutation(mutation)

  // getValidActions filters out any actions that have an empty destination
  const getValidActions = (): ServiceRuleAction[] => {
    const validActions: ServiceRuleAction[] = []
    if (value.actions.length === 0) return []

    value.actions.forEach((action: ServiceRuleAction) => {
      if (action.destination) validActions.push(action)
    })

    return validActions
  }

  const getFilterValueType = (value: string): ServiceRuleFilterValueType => {
    let valueType: ServiceRuleFilterValueType = 'UNKNOWN'
    value = value.trim()
    if (value.toLowerCase() === 'true' || value.toLowerCase() === 'false') {
      valueType = 'BOOL'
    } else if (!isNaN(Number(value)) && !isNaN(parseFloat(value))) {
      valueType = 'NUMBER'
    } else {
      valueType = 'STRING'
    }

    return valueType
  }

  const getFiltersWithValueTypes = (): ServiceRuleFilterInput[] => {
    const filters: ServiceRuleFilterInput[] = []
    value.filters.forEach((filter: ServiceRuleFilterInput) => {
      filter.valueType = getFilterValueType(filter.value)
      filters.push(filter)
    })
    return filters
  }

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Service Rule'
      loading={createRuleStatus.fetching}
      errors={nonFieldErrors(createRuleStatus.error)}
      onClose={onClose}
      onSubmit={() => {
        const validActions = getValidActions()
        if (validActions.length === 0) {
          setActionsError(true)
          return
        }
        setActionsError(false)

        commit(
          {
            input: {
              name: value.name,
              serviceID,
              sendAlert: value.sendAlert,
              actions: JSON.stringify(validActions),
              filters: getFiltersWithValueTypes(),
              integrationKeys: value.integrationKeys,
            },
          },
          { additionalTypenames: ['ServiceRule'] },
        ).then(onClose)
      }}
      form={
        <ServiceRuleForm
          errors={fieldErrors(createRuleStatus.error)}
          disabled={createRuleStatus.fetching}
          value={value}
          onChange={(value): void => {
            setValue(value)
          }}
          serviceID={serviceID}
          actionsError={actionsError}
        />
      }
    />
  )
}
