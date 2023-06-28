package app

// Trigger will start a processing cycle (normally ever ~5s)
func (app *App) Trigger() {
	_ = app.mgr.WaitForStartup(app.LogBackgroundContext())

	if app.Engine != nil {
		app.Engine.Trigger()
	}
}
