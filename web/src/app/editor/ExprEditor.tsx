import React from 'react'
import CodeMirror from '@uiw/react-codemirror'
import { Expr } from './expr-lang'
import { useTheme } from '../theme/useTheme'
import { bracketMatching } from '@codemirror/language'
import { Grid, IconButton } from '@mui/material'
import { AutoFixHigh } from '@mui/icons-material'

export type ExprEditorProps = {
  value: string
  onChange: (value: string) => void
  minHeight?: string
  maxHeight?: string
}

export default function ExprEditor(props: ExprEditorProps): React.ReactNode {
  const theme = useTheme()

  return (
    <Grid container>
      <Grid item flexGrow={1}>
        <CodeMirror
          value={props.value}
          theme={theme}
          onChange={props.onChange}
          extensions={[bracketMatching(), Expr()]}
          minHeight={props.minHeight}
          maxHeight={props.maxHeight}
        />
      </Grid>
    </Grid>
  )
}
