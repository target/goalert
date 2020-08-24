import React from 'react'
import Lottie from 'react-lottie'
import * as animationData from './check-mark-success.json'

interface SuccessAnimationProps {
  isStopped: boolean
}

export default function SuccessAnimation(
  props: SuccessAnimationProps,
): JSX.Element {
  const defaultOptions = {
    loop: false,
    autoplay: false,
    animationData: animationData,
    rendererSettings: {
      preserveAspectRatio: 'xMidYMid slice',
    },
  }

  return (
    <Lottie
      options={defaultOptions}
      isStopped={props.isStopped}
      height={175}
      width={175}
    />
  )
}
