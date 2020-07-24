import React from 'react'
import List from '@material-ui/core/List'
import ListItem from '@material-ui/core/ListItem'
import ListItemText from '@material-ui/core/ListItemText'
import Typography from '@material-ui/core/Typography'
import { FormContainer } from '../forms'
import { defaultTo } from 'lodash-es'
import { makeStyles, Theme } from '@material-ui/core/styles'
import {
  StringInput,
  StringListInput,
  IntegerInput,
  BoolInput,
} from './AdminFieldComponents'
import { ConfigValue } from '../../schema'
import { yellow } from '@material-ui/core/colors'

const components = {
  string: StringInput,
  stringList: StringListInput,
  integer: IntegerInput,
  boolean: BoolInput,
}

interface FieldProps extends ConfigValue {
  label: string
}

interface AdminSectionProps {
  headerNote?: string
  value: { [id: string]: string }
  fields: FieldProps[]
  onChange: (id: string, value: null | string) => void
}

const useStyles = makeStyles((theme: Theme) => ({
  activeItem: {
    backgroundColor: yellow[300],
  },
  listItem: {
    // leaves some room around fields without descriptions
    // 71px is the height of the checkbox field without w/o a desc
    minHeight: '71px',
    padding: '1em',
  },
  listItemText: {
    maxWidth: '50%',
    [theme.breakpoints.up('md')]: {
      maxWidth: '65%',
    },
  },
  listItemAction: {
    width: '50%',
    [theme.breakpoints.up('md')]: {
      width: '35%',
    },
    display: 'flex',
    justifyContent: 'flex-end',
  },
}))

export default function AdminSection(props: AdminSectionProps): JSX.Element {
  // TODO: add 'reset to default' buttons
  const classes = useStyles()
  const { fields, value, headerNote } = props

  return (
    <FormContainer>
      <List>
        {headerNote && (
          <ListItem>
            <ListItemText
              disableTypography
              secondary={
                <Typography color='textSecondary'>{headerNote}</Typography>
              }
              style={{ fontStyle: 'italic' }}
            />
          </ListItem>
        )}
        {fields.map((f: FieldProps, idx: number) => {
          const Field = components[f.type]
          return (
            <ListItem
              key={f.id}
              className={
                classes.listItem +
                ' ' +
                (_.has(value, f.id) ? classes.activeItem : '')
              }
              divider={idx !== fields.length - 1}
            >
              <ListItemText
                className={classes.listItemText}
                primary={f.label}
                secondary={f.description}
              />
              <div className={classes.listItemAction}>
                <Field
                  type={f.id === 'Twilio.FromNumber' ? 'tel' : 'text'}
                  name={f.id}
                  value={defaultTo(value[f.id], f.value)}
                  password={f.password}
                  onChange={(val) =>
                    props.onChange(f.id, val === f.value ? null : val)
                  }
                />
              </div>
            </ListItem>
          )
        })}
      </List>
    </FormContainer>
  )
}
