package app

import "context"

// Trigger will start a processing cycle (normally ever ~5s)
func (app *App) Trigger() {
	app.mgr.WaitForStartup(context.Background())

	if app.engine != nil {
		app.engine.Trigger()
	}
}
