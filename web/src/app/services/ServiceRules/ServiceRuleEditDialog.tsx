import React, { useState } from 'react'
import { useMutation, useQuery, gql } from 'urql'

import { fieldErrors, nonFieldErrors } from '../../util/errutil'
import FormDialog from '../../dialogs/FormDialog'
import { IntegrationKey, ServiceRule } from '../../../schema'
import ServiceRuleForm, {
  CustomFields,
  ServiceRuleValue,
  destType,
} from './ServiceRuleForm'
import {
  getFiltersWithValueTypes,
  getIntegrationKeyValues,
  getValidActions,
} from './ServiceRuleCreateDialog'
import Spinner from '../../loading/components/Spinner'
import { GenericError } from '../../error-pages'

const mutation = gql`
  mutation ($input: UpdateServiceRuleInput!) {
    updateServiceRule(input: $input) {
      id
      name
      serviceID
      actions {
        destType
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

export const getCustomFields = (
  r: ServiceRule | null,
): CustomFields | undefined => {
  const customFields: CustomFields = {
    summary: '',
    details: '',
  }
  r?.actions.map((action) => {
    if (action.destType === destType.ALERT) {
      action.contents.map((content) => {
        if (content.prop === 'summary') {
          customFields.summary = content.value
        } else if (content.prop === 'details') {
          customFields.details = content.value
        }
      })
    }
  })

  return customFields.summary && customFields.details ? customFields : undefined
}

export default function ServiceRuleEditDialog(props: {
  rule: ServiceRule
  onClose: () => void
  integrationKeys: IntegrationKey[]
  serviceID: string
}): JSX.Element {
  const { rule, onClose, serviceID, integrationKeys } = props

  const [value, setValue] = useState<ServiceRuleValue>({
    name: rule.name,
    filters: rule.filters,
    sendAlert: rule.sendAlert,
    actions: rule.actions,
    integrationKeys: rule.integrationKeys.map((key) => {
      return {
        label: key.name,
        value: key.id,
      }
    }),
    customFields: getCustomFields(rule),
  })
  const [actionsError, setActionsError] = useState<boolean>(false)
  const [editRuleStatus, commit] = useMutation(mutation)

  const [{ fetching, error, data }] = useQuery({
    query,
    variables: {},
  })
  if (fetching && !data) return <Spinner />
  if (error) return <GenericError error={error.message} />

  return (
    <FormDialog
      maxWidth='sm'
      title='Edit New Service Rule'
      loading={editRuleStatus.fetching}
      errors={nonFieldErrors(editRuleStatus.error)}
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
              id: rule.id,
              name: value.name,
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
          errors={fieldErrors(editRuleStatus.error)}
          disabled={editRuleStatus.fetching}
          value={value}
          onChange={(value: ServiceRuleValue): void => {
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
