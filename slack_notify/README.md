# Slack Notifier

This small helper will execute a given command and send a configured message to
slack that has

* status success on return value 0 of the command, and
* status error on any other return value.

For example

	slack_notify "testing slack_notify" do_something

would call the `do_something` command (that of course must be available in the
callers path) and send a message with the text "testing slack_notify" to the
slack organization and channel configured in the `${HOME}/.slack.conf` file.

The configuration file is in JSON format and has the following form:

	{
		"webhook_url": "https://hooks.slack.com/services/...",
		"channel": "#slack_notify_channel",
		"username": "Mr. Slack Notify",
		"emoji": ":ghost:"
	}

If `channel`, `username` or `emoji` are not specified the values of the given
webhook URL are used (see your slack integration configuration for that
respective webhook).
