import { Checkbox, Tooltip, Typography } from '@mui/material'
import FormControlLabel from '@mui/material/FormControlLabel'
import Grid from '@mui/material/Grid'
import InfoIcon from '@mui/icons-material/Info'
import React, { useState } from 'react'
import { FormField } from '../../../forms'
import { SlackChannelSelect, SlackUserGroupSelect } from '../../../selection'
import { useExpFlag } from '../../../util/useExpFlag'
import { SlackFields } from '../util'

export type SlackFieldsFormProps = {
  slackFields: SlackFields
  onChange: (val: SlackFields) => void
}

export default function SlackFieldsForm(
  props: SlackFieldsFormProps,
): JSX.Element {
  const { slackFields, onChange } = props
  const slackUGEnabled = useExpFlag('slack-ug')
  const [slackUGChecked, setSlackUGChecked] = useState(
    !!slackFields.slackUserGroup,
  )

  return (
    <React.Fragment>
      <Grid item>
        <FormField
          component={SlackChannelSelect}
          fullWidth
          required
          label='Slack Channel'
          name='channelFields.slackChannelID'
        />
      </Grid>

      {slackUGEnabled && (
        <Grid item>
          <FormControlLabel
            sx={{ pb: 2 }}
            control={
              <Checkbox
                checked={slackUGChecked}
                onChange={() => {
                  const newVal = !slackUGChecked
                  setSlackUGChecked(newVal)
                  if (!newVal) {
                    onChange({
                      slackChannelID: slackFields.slackChannelID,
                    })
                  }
                }}
              />
            }
            label={
              <Typography sx={{ display: 'flex' }}>
                Also set the members of a Slack user group?
                <Tooltip
                  data-cy='fts-tooltip'
                  disableFocusListener
                  placement='right'
                  title='This will edit your user group in Slack to ensure that only the members in the selected group are also on-call'
                >
                  <InfoIcon color='primary' sx={{ pl: 0.5 }} />
                </Tooltip>
              </Typography>
            }
          />
          {slackUGChecked && (
            <FormField
              component={SlackUserGroupSelect}
              fullWidth
              label='Slack User Group'
              name='channelFields.slackUserGroup'
              mapOnChangeValue={(v) => {
                //   setSlackType('usergroup')
                return v
              }}
            />
          )}
        </Grid>
      )}
    </React.Fragment>
  )
}
