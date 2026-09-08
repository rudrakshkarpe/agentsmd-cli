# Timeout-units task

This held-out task models a production configuration bug: operators provide
`RETRY_DELAY_MS` as an integer, while the implementation incorrectly delegates
to Go's duration-string parser. The visible fixture contains the executable path
and a compatibility-only decoy. Tests are copied into the workspace only after
the agent exits.

The learned condition contains two independently removable rules. The suite
runs the full learned file and one condition per omitted rule, so a result can
distinguish routing value from sandbox-cache value instead of attributing an
improvement to the bundle as a whole.
