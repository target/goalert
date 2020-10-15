import React from 'react'
import IconButton from '@material-ui/core/IconButton'
import FavoriteFilledIcon from '@material-ui/icons/Star'
import FavoriteBorderIcon from '@material-ui/icons/StarBorder'
import Tooltip from '@material-ui/core/Tooltip'
import Spinner from '../loading/components/Spinner'

interface SetFavoriteButtonProps {
  typeName: 'rotation' | 'service' | 'schedule'
  isFavorite?: boolean
  loading: boolean
  onClick: () => void
}

export function SetFavoriteButton({
  typeName,
  isFavorite,
  loading,
  onClick,
}: SetFavoriteButtonProps): JSX.Element {
  let icon = isFavorite ? <FavoriteFilledIcon /> : <FavoriteBorderIcon />
  if (loading) {
    icon = <Spinner />
  }

  const content = (
    <form
      onSubmit={(e) => {
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

  switch (typeName) {
    case 'service':
      return (
        <Tooltip
          title={
            isFavorite
              ? 'Unfavorite this service to stop seeing its alerts on your homepage'
              : 'Favorite this service to always see its alerts on your homepage'
          }
        >
          {content}
        </Tooltip>
      )
    default:
      return content
  }
}
