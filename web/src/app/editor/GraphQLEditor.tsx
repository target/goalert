import React from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { graphql as graphqlLang } from 'cm6-graphql'
import { buildClientSchema, getIntrospectionQuery } from 'graphql'
import { gql, useQuery } from 'urql'
import { useTheme } from '../theme/useTheme'
import { bracketMatching } from '@codemirror/language'

const query = gql(getIntrospectionQuery())

export type GraphQLEditorProps = {
  value: string
  onChange: (value: string) => void
  minHeight?: string
}

export default function GraphQLEditor(
  props: GraphQLEditorProps,
): React.ReactNode {
  const [q] = useQuery({ query })
  if (q.error) throw q.error
  const schema = buildClientSchema(q.data)
  const theme = useTheme()

  return (
    <CodeMirror
      value={props.value}
      theme={theme}
      onChange={props.onChange}
      extensions={[bracketMatching(), graphqlLang(schema)]}
      minHeight={props.minHeight}
    />
  )
}
