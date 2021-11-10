package app

// Trigger will start a processing cycle (normally ever ~5s)
func (app *App) Trigger() {
	app.mgr.WaitForStartup(app.LogContext())

	if app.Engine != nil {
		app.Engine.Trigger()
	}
}
