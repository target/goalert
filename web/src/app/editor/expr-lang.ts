import { parser } from './expr-parser'
import {
  LRLanguage,
  LanguageSupport,
  indentNodeProp,
  foldNodeProp,
  foldInside,
  delimitedIndent,
} from '@codemirror/language'
import { styleTags, tags as t } from '@lezer/highlight'

export const exprHighlighting = styleTags({
  'Number! RangeExpression!': t.number,
  'Number!': t.number,
  Identifier: t.variableName,
  'Property/Identifier': t.propertyName,
  'True False': t.bool,
  Nil: t.null,
  'StringLiteral!': t.string,
  Operator: t.operator,
  LineComment: t.lineComment,
  '( )': t.paren,
  '[ ]': t.squareBracket,
  '{ }': t.brace,
  ',': t.separator,
  ':': t.punctuation,
  'Hash Env': t.keyword,
  PredicateFunction: t.function(t.keyword),
  'FunctionCall/Identifier': t.function(t.variableName),
})

export const ExprLanguage = LRLanguage.define({
  parser: parser.configure({
    props: [
      indentNodeProp.add({
        Application: delimitedIndent({ closing: ')', align: false }),
      }),
      foldNodeProp.add({
        Application: foldInside,
      }),
      exprHighlighting,
    ],
  }),
  languageData: {
    commentTokens: { line: ';' },
  },
})

export function Expr() {
  return new LanguageSupport(ExprLanguage)
}
