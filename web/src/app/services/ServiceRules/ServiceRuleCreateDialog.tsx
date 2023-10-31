import React, { useState } from 'react'
import { useMutation, gql } from 'urql'

import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import {
  CreateServiceRuleInput,
  IntegrationKey,
  ServiceRuleActionInput,
  ServiceRuleFilterInput,
  ServiceRuleFilterValueType,
  UpdateServiceRuleInput,
} from '../../../schema'
import ServiceRuleForm from './ServiceRuleForm'

const mutation = gql`
  mutation ($input: CreateServiceRuleInput!) {
    createServiceRule(input: $input) {
      name
      serviceID
      actions {
        destType
        destID
        destValue
        contents {
          prop
          value
        }
      }
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
  integrationKeys: IntegrationKey[]
}): JSX.Element {
  const { serviceID, onClose } = props
  const [value, setValue] = useState<
    CreateServiceRuleInput | UpdateServiceRuleInput
  >({
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
  const getValidActions = (): ServiceRuleActionInput[] => {
    const validActions: ServiceRuleActionInput[] = []
    if (value.actions.length === 0) return []

    value.actions.forEach((action: ServiceRuleActionInput) => {
      if (action.destType) validActions.push(action)
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

  // getIntegrationKeyValues returns a list of integration key IDs from key objects
  const getIntegrationKeyValues = (): string[] => {
    const integrationKeys: string[] = []
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    value.integrationKeys.forEach((key: any): void => {
      integrationKeys.push(key.value)
    })
    return integrationKeys
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
              actions: validActions,
              filters: getFiltersWithValueTypes(),
              integrationKeys: getIntegrationKeyValues(),
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
          integrationKeys={props.integrationKeys}
        />
      }
    />
  )
}
