package remotemonitor

/*

	Remote monitor allows monitoring external functionality and communication of GoAlert instances.

	## GoAlert Configuration

	GoAlert instances being monitored must have the following configured:
	- A "main" service for heartbeats and errors
	- "main" service should have a heartbeat defined for each remote monitor
	- "main" service EP should immediately notify someone
	- A "monitor" service for each remote monitor (used for creating test alerts)
	- A "monitor" user for each remote monitor with an SMS contact method, and immediate notification rule
	- "monitor" service EP steps should point to the corresponding user
	- "monitor" service EP should wait at least 1 minute before escalating to someone

*/
