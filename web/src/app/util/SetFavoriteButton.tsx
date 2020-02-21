import React from 'react'
import IconButton from '@material-ui/core/IconButton'
import FavoriteFilledIcon from '@material-ui/icons/Star'
import FavoriteBorderIcon from '@material-ui/icons/StarBorder'
import Spinner from '../loading/components/Spinner'

interface SetFavoriteButtonProps {
  typeName: 'rotation' | 'service' | 'schedule'
  isFavorite?: boolean
  loading: boolean
  onClick: Function
}

export function SetFavoriteButton({
  typeName,
  isFavorite,
  loading,
  onClick,
}: SetFavoriteButtonProps) {
  let icon = isFavorite ? <FavoriteFilledIcon /> : <FavoriteBorderIcon />
  if (loading) {
    icon = <Spinner />
  }
  return (
    <form
      onSubmit={e => {
        e.preventDefault()
        onClick()
      }}
    >
      <IconButton
        aria-label={
          isFavorite
            ? `Unset as a Favorite ${typeName}`
            : `Set as a Favorite ${typeName}`
        }
        type='submit'
        color='inherit'
        data-cy='set-fav'
      >
        {icon}
      </IconButton>
    </form>
  )
}
