import React from 'react'
import List from '@mui/material/List'
import ListItem from '@mui/material/ListItem'
import ListItemText from '@mui/material/ListItemText'
import Typography from '@mui/material/Typography'
import Box from '@mui/material/Box'
import { FormContainer } from '../forms'
import _, { defaultTo } from 'lodash'
import {
  StringInput,
  StringListInput,
  IntegerInput,
  BoolInput,
} from './AdminFieldComponents'
import { ConfigValue } from '../../schema'
import { Alert } from '@mui/material'

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

export default function AdminSection(
  props: AdminSectionProps,
): React.JSX.Element {
  // TODO: add 'reset to default' buttons
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
              sx={(theme) => ({
                minHeight: '71px',
                padding: '1em',
                ...(_.has(value, f.id)
                  ? { bgcolor: 'action.selected' }
                  : {}),
              })}
              divider={idx !== fields.length - 1}
            >
              <ListItemText
                sx={(theme) => ({
                  maxWidth: '50%',
                  [theme.breakpoints.up('md')]: {
                    maxWidth: '65%',
                  },
                })}
                primary={(f.deprecated ? '* ' : '') + f.label}
                secondary={
                  f.deprecated ? (
                    <React.Fragment>
                      <Alert
                        severity='warning'
                        sx={{ pb: 1 }}
                        style={{ margin: '1em' }}
                      >
                        Deprecated: {f.deprecated}
                      </Alert>
                      {f.description}
                    </React.Fragment>
                  ) : (
                    f.description
                  )
                }
              />
              <Box
                sx={(theme) => ({
                  width: '50%',
                  [theme.breakpoints.up('md')]: {
                    width: '35%',
                  },
                  display: 'flex',
                  justifyContent: 'flex-end',
                })}
              >
                <Field
                  name={f.id}
                  value={defaultTo(value[f.id], f.value)}
                  password={f.password}
                  onChange={(val) =>
                    props.onChange(f.id, val === f.value ? null : val)
                  }
                />
              </Box>
            </ListItem>
          )
        })}
      </List>
    </FormContainer>
  )
}
