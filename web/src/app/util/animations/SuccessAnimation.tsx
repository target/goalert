import React from 'react'
import Lottie from 'react-lottie'
import * as animationData from './check-mark-success.json'

interface SuccessAnimationProps {
  autoplay?: boolean
  isStopped?: boolean
  loop?: boolean
}

export default function SuccessAnimation(
  props: SuccessAnimationProps,
): JSX.Element {
  const defaultOptions = {
    loop: props.loop || false,
    autoplay: props.autoplay || false,
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
