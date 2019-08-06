package app

func (app *App) listen() error { return app.listenNoUpgrade() }
