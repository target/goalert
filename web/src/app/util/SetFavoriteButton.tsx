import React from 'react'
import IconButton from '@mui/material/IconButton'
import MUIFavoriteIcon from '@mui/icons-material/Favorite'
import NotFavoriteIcon from '@mui/icons-material/FavoriteBorder'
import Tooltip from '@mui/material/Tooltip'
import makeStyles from '@mui/styles/makeStyles'
import _ from 'lodash'
import { Theme } from '@mui/material'

interface SetFavoriteButtonProps {
  typeName: 'rotation' | 'service' | 'schedule' | 'escalationPolicy' | 'user'
  onClick: (event: React.FormEvent<HTMLFormElement>) => void
  isFavorite?: boolean
  loading?: boolean
}

const useStyles = makeStyles((theme: Theme) => ({
  favorited: {
    color:
      theme.palette.mode === 'dark'
        ? theme.palette.primary.main
        : 'rgb(205, 24, 49)',
  },
  notFavorited: {
    color: 'inherit',
  },
}))

export function FavoriteIcon(): React.ReactNode {
  const classes = useStyles()
  return <MUIFavoriteIcon data-cy='fav-icon' className={classes.favorited} />
}

export function SetFavoriteButton({
  typeName,
  onClick,
  isFavorite,
  loading,
}: SetFavoriteButtonProps): React.ReactNode | null {
  const classes = useStyles()

  if (loading) {
    return null
  }

  const icon = isFavorite ? (
    <FavoriteIcon />
  ) : (
    <NotFavoriteIcon className={classes.notFavorited} />
  )

  const content = (
    <form
      onSubmit={(e) => {
        e.preventDefault()
        onClick(e)
      }}
    >
      <IconButton
        aria-label={
          isFavorite
            ? `Unset as a Favorite ${_.startCase(typeName).toLowerCase()}`
            : `Set as a Favorite ${_.startCase(typeName).toLowerCase()}`
        }
        type='submit'
        data-cy='set-fav'
        size='large'
      >
        {icon}
      </IconButton>
    </form>
  )

  return (
    <Tooltip title={isFavorite ? 'Unfavorite' : 'Favorite'} placement='top'>
      {content}
    </Tooltip>
  )
}
