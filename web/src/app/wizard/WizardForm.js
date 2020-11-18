import React from 'react'
import p from 'prop-types'
import StepIcon from '@material-ui/core/StepIcon'
import FormControl from '@material-ui/core/FormControl'
import FormControlLabel from '@material-ui/core/FormControlLabel'
import FormLabel from '@material-ui/core/FormLabel'
import Grid from '@material-ui/core/Grid'
import Radio from '@material-ui/core/Radio'
import RadioGroup from '@material-ui/core/RadioGroup'
import TextField from '@material-ui/core/TextField'
import Typography from '@material-ui/core/Typography'
import { FormContainer, FormField } from '../forms'
import WizardScheduleForm from './WizardScheduleForm'
import { value as valuePropType } from './propTypes'
import withStyles from '@material-ui/core/styles/withStyles'
import withWidth, { isWidthDown } from '@material-ui/core/withWidth'
import MaterialSelect from '../selection/MaterialSelect'
import { set } from 'lodash'

const styles = {
  fieldItem: {
    marginLeft: '2.5em',
  },
  stepItem: {
    display: 'flex',
    alignItems: 'center',
  },
}

@withWidth()
@withStyles(styles)
export default class WizardForm extends React.PureComponent {
  static propTypes = {
    onChange: p.func.isRequired,
    value: valuePropType,
    errors: p.arrayOf(
      p.shape({
        message: p.string.isRequired,
      }),
    ),
  }

  handleSecondaryScheduleToggle = (e) => {
    const { onChange, value } = this.props
    onChange(set(value, ['secondarySchedule', 'enable'], e.target.value))
    this.forceUpdate()
  }

  sectionHeading = (text) => {
    return (
      <Typography component='h2' variant='h6'>
        {text}
      </Typography>
    )
  }

  render() {
    const { classes, onChange, value, width } = this.props

    return (
      <FormContainer optionalLabels {...this.props}>
        <Grid container spacing={2}>
          <Grid item className={classes.stepItem}>
            <StepIcon icon='1' />
          </Grid>
          <Grid item xs={10}>
            {this.sectionHeading('Team Details')}
          </Grid>
          <Grid item xs={12} className={classes.fieldItem}>
            <FormField
              component={TextField}
              name='teamName'
              errorName='newEscalationPolicy.name'
              label={`What is your team's name?`}
              formLabel
              fullWidth={isWidthDown('md', width)}
              required
            />
          </Grid>
          <Grid item className={classes.stepItem}>
            <StepIcon icon='2' />
          </Grid>
          <Grid item xs={10}>
            {this.sectionHeading('Primary Schedule')}
          </Grid>
          <WizardScheduleForm onChange={onChange} value={value} />
          <Grid item className={classes.stepItem}>
            <StepIcon icon='3' />
          </Grid>
          <Grid item xs={10}>
            {this.sectionHeading('Secondary Schedule')}
          </Grid>
          <Grid item xs={12} className={classes.fieldItem}>
            <FormControl>
              <FormLabel>
                Will your team need a <b>secondary</b> schedule to escalate to?
              </FormLabel>
              <RadioGroup
                aria-label='Create Secondary Schedule'
                name='secondary'
                row
                value={value.secondarySchedule.enable}
                onChange={this.handleSecondaryScheduleToggle}
              >
                <FormControlLabel
                  data-cy='secondary.yes'
                  value='yes'
                  control={<Radio />}
                  label='Yes'
                />
                <FormControlLabel
                  data-cy='secondary.no'
                  value='no'
                  control={<Radio />}
                  label='No'
                />
              </RadioGroup>
            </FormControl>
          </Grid>
          {value.secondarySchedule.enable === 'yes' && (
            <WizardScheduleForm onChange={onChange} value={value} secondary />
          )}
          <Grid item className={classes.stepItem}>
            <StepIcon icon='4' />
          </Grid>
          <Grid item xs={10}>
            {this.sectionHeading('Escalation Policy')}
          </Grid>
          <Grid item xs={12} className={classes.fieldItem}>
            <FormField
              component={TextField}
              name='delayMinutes'
              errorName='newEscalationPolicy.steps0.delayMinutes'
              label={this.getDelayLabel()}
              formLabel
              fullWidth={isWidthDown('md', width)}
              required
              type='number'
              placeholder='15'
              min={1}
              max={9000}
            />
          </Grid>
          <Grid item xs={12} className={classes.fieldItem}>
            <FormField
              component={TextField}
              name='repeat'
              errorName='newEscalationPolicy.repeat'
              label='How many times would you like to repeat alerting your team?'
              formLabel
              fullWidth={isWidthDown('md', width)}
              required
              type='number'
              placeholder='3'
              mapOnChangeValue={(value) => value.toString()}
              min={0}
              max={5}
            />
          </Grid>
          <Grid item className={classes.stepItem}>
            <StepIcon icon='5' />
          </Grid>
          <Grid item xs={10}>
            {this.sectionHeading('Service')}
          </Grid>
          <Grid item xs={12} className={classes.fieldItem}>
            <FormField
              component={MaterialSelect}
              name='key'
              label='How would you like to connect your application with GoAlert?'
              formLabel
              fullWidth={isWidthDown('md', width)}
              required
              options={[
                {
                  label: 'Generic API',
                  value: 'generic',
                },
                {
                  label: 'Grafana Webhook URL',
                  value: 'grafana',
                },
                {
                  label: 'Site24x7 Webhook URL',
                  value: 'site24x7',
                },
                {
                  label: 'Prometheus Alertmanager Webhook URL',
                  value: 'prometheusAlertmanager',
                },
                {
                  label: 'Email',
                  value: 'email',
                },
              ]}
            />
          </Grid>
        </Grid>
      </FormContainer>
    )
  }

  getDelayLabel = () => {
    if (this.props.value.secondarySchedule.enable === 'yes') {
      return 'How long would you like to wait until escalating to your secondary schedule (in minutes)?'
    }
    return 'How long would you like to wait until alerting your primary schedule again (in minutes)?'
  }
}
