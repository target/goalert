import React from 'react'
import CodeMirror, { EditorView } from '@uiw/react-codemirror'
import { Expr } from './expr-lang'
import { useTheme } from '../theme/useTheme'
import { bracketMatching } from '@codemirror/language'
import { Grid } from '@mui/material'
import { nonce } from '../env'

export type ExprEditorProps = {
  value: string
  onChange: (value: string) => void
  minHeight?: string
  maxHeight?: string
  onFocus?: () => void
  onBlur?: () => void
}

export default function ExprEditor(props: ExprEditorProps): React.ReactNode {
  const theme = useTheme()

  return (
    <Grid container>
      <Grid item flexGrow={1}>
        <CodeMirror
          value={props.value}
          theme={theme}
          onFocus={props.onFocus}
          onBlur={props.onBlur}
          onChange={props.onChange}
          extensions={[
            EditorView.cspNonce.of(nonce),
            bracketMatching(),
            Expr(),
            EditorView.lineWrapping,
          ]}
          minHeight={props.minHeight}
          maxHeight={props.maxHeight}
        />
      </Grid>
    </Grid>
  )
}
