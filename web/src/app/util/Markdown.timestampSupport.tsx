import React from 'react'
import { findAndReplace } from 'mdast-util-find-and-replace'
import { Time } from './Time'
import CopyText from './CopyText'

const isoTimestampRegex =
  /\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d\.\d+([+-][0-2]\d:[0-5]\d|Z)/g

export default function timestampSupport() {
  return function (tree: Parameters<typeof findAndReplace>[0]) {
    findAndReplace(tree, [
      isoTimestampRegex,

      // @ts-expect-error mdast types are wrong
      (value) => {
        return {
          type: 'element',
          value: (
            <CopyText
              noTypography
              title={<Time time={value} />}
              value={value}
            />
          ),
        }
      },
    ])
  }
}
