import React, { useState } from 'react'
import { useMutation, gql, useQuery } from 'urql'

import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import {
  IntegrationKey,
  ServiceRuleActionInput,
  ServiceRuleFilterInput,
  ServiceRuleFilterValueType,
  SlackChannel,
} from '../../../schema'
import ServiceRuleForm, { ServiceRuleValue, destType } from './ServiceRuleForm'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'

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

const query = gql`
  query ($input: SlackChannelSearchOptions) {
    slackChannels(input: $input) {
      nodes {
        id
        name
      }
    }
  }
`

// getValidActions returns a list of valid actions
export const getValidActions = (
  v: ServiceRuleValue,
  slackChannels: SlackChannel[],
): ServiceRuleActionInput[] => {
  const validActions: ServiceRuleActionInput[] = []
  if (v.actions.length === 0) return []

  v.actions.forEach((action: ServiceRuleActionInput) => {
    if (action.destType === destType.SLACK) {
      validActions.push({
        destID: action.destID,
        destType: action.destType,
        destValue: action.destValue,
        contents: [
          { prop: 'message', value: action.contents[0].value },
          {
            prop: 'channel',
            value:
              slackChannels.find((chan) => chan.id === action.destID)?.name ||
              '',
          },
        ],
      })
    } else if (action.destType && action.destType !== destType.ALERT) {
      validActions.push(action)
    }
  })

  // add Alert information
  if (v.sendAlert) {
    if (v.customFields) {
      validActions.push({
        destType: destType.ALERT,
        destID: '',
        destValue: '',
        contents: [
          { prop: 'summary', value: v.customFields.summary },
          { prop: 'details', value: v.customFields.details },
        ],
      })
    } else {
      validActions.push({
        destType: destType.ALERT,
        destID: '',
        destValue: '',
        contents: [],
      })
    }
  }
  return validActions
}

export const getFilterValueType = (v: string): ServiceRuleFilterValueType => {
  let valueType: ServiceRuleFilterValueType = 'UNKNOWN'
  v = v.trim()
  if (v.toLowerCase() === 'true' || v.toLowerCase() === 'false') {
    valueType = 'BOOL'
  } else if (!isNaN(Number(v)) && !isNaN(parseFloat(v))) {
    valueType = 'NUMBER'
  } else {
    valueType = 'STRING'
  }

  return valueType
}

export const getFiltersWithValueTypes = (
  v: ServiceRuleValue,
): ServiceRuleFilterInput[] => {
  const filters: ServiceRuleFilterInput[] = []
  v.filters.forEach((filter: ServiceRuleFilterInput) => {
    filter.valueType = getFilterValueType(filter.value)
    filters.push(filter)
  })
  return filters
}

// getIntegrationKeyValues returns a list of integration key IDs from key objects
export const getIntegrationKeyValues = (v: ServiceRuleValue): string[] => {
  const integrationKeys: string[] = []
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  v.integrationKeys.forEach((key: any): void => {
    integrationKeys.push(key.value)
  })
  return integrationKeys
}

export default function ServiceRuleCreateDialog(props: {
  serviceID: string
  onClose: () => void
  integrationKeys: IntegrationKey[]
}): JSX.Element {
  const { serviceID, onClose, integrationKeys } = props
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

  const [{ fetching, error, data }] = useQuery({
    query,
    variables: {},
  })
  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <FormDialog
      maxWidth='sm'
      title='Create New Service Rule'
      loading={createRuleStatus.fetching}
      errors={nonFieldErrors(createRuleStatus.error)}
      onClose={onClose}
      onSubmit={() => {
        const validActions = getValidActions(value, data.slackChannels.nodes)
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
              filters: getFiltersWithValueTypes(value),
              integrationKeys: getIntegrationKeyValues(value),
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
          integrationKeys={integrationKeys}
        />
      }
    />
  )
}
