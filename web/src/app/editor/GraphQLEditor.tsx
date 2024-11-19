import React from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { graphql as graphqlLang } from 'cm6-graphql'
import { buildClientSchema, getIntrospectionQuery, parse, print } from 'graphql'
import { gql, useQuery } from 'urql'
import { useTheme } from '../theme/useTheme'
import { bracketMatching } from '@codemirror/language'
import { Grid, IconButton } from '@mui/material'
import { AutoFixHigh } from '@mui/icons-material'

const query = gql(getIntrospectionQuery())

export type GraphQLEditorProps = {
  value: string
  onChange: (value: string) => void
  minHeight?: string
  maxHeight?: string
  readOnly?: boolean
}

export default function GraphQLEditor(
  props: GraphQLEditorProps,
): React.ReactNode {
  const [q] = useQuery({ query })
  if (q.error) throw q.error
  const schema = buildClientSchema(q.data)
  const theme = useTheme()

  return (
    <Grid container>
      <Grid item flexGrow={1}>
        <CodeMirror
          value={props.value}
          theme={theme}
          readOnly={props.readOnly}
          onChange={props.onChange}
          extensions={[bracketMatching(), graphqlLang(schema)]}
          minHeight={props.minHeight}
          maxHeight={props.maxHeight}
        />
      </Grid>
      {!props.readOnly && (
        <Grid item>
          <IconButton
            onClick={() => {
              props.onChange(print(parse(props.value)) + '\n')
            }}
            title='Format query'
          >
            <AutoFixHigh />
          </IconButton>
        </Grid>
      )}
    </Grid>
  )
}
