import React from 'react'
import {
  FindAndReplaceTuple,
  findAndReplace,
} from 'mdast-util-find-and-replace'
import { Time } from './Time'
import CopyText from './CopyText'

type Node = {
  type: 'element'
  value: React.ReactNode
}

// 2006-01-02T15:04:05.999999999Z07:00
const isoTimestampRegex =
  /\d{4}-[01]\d-[0-3]\dT[0-2]\d:[0-5]\d:[0-5]\d(\.\d+)?([+-][0-2]\d:[0-5]\d|Z)/g
function fromISO(iso: string): Node {
  return {
    type: 'element',
    value: <CopyText noTypography title={<Time time={iso} />} value={iso} />,
  }
}

// Mon Jan _2 15:04:05 MST 2006
const unixStampRegex = /\w+\s+\w+\s+\d+\s+\d{2}:\d{2}:\d{2}\s\w+\s\d{4}/g
function fromUnixStamp(unix: string): Node {
  return fromISO(new Date(unix).toISOString())
}

// mdast types are wrong, so we have to cast to unknown and then to the correct
// type.
const isoTuple = [isoTimestampRegex, fromISO] as unknown as FindAndReplaceTuple
const unixTuple = [
  unixStampRegex,
  fromUnixStamp,
] as unknown as FindAndReplaceTuple

export default function timestampSupport() {
  return function (tree: Parameters<typeof findAndReplace>[0]) {
    findAndReplace(tree, [isoTuple, unixTuple])
  }
}
