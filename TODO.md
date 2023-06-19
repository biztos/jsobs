# jsobs SPECULATIVE TODO list

## MAYBE logging!

Want some kind of debug logger we can easily integrate.

Best idea right now is have an interface that just happens to match the
logger of choice, and a default implementation that just uses logging to
keep it simple.  Then pass the nicer logger in at construction in other code.
At the `jsobs.Client` level.

**HOWEVER** it might be just as good to keep this API simpler and let the
caller do the logging if they want.  Or split the difference and have the
backend log verbosely at debug? Meh in that case I want "debug" versus
"superdebug" with the latter showing SQL...

OK no logging for now.

## Continuous purge for pgclient via timer... MAYBE!

On the one hand, you should not worry about purging at the end of every
process because you could end up with a massive purge.

On that same hand, it's interesting to do this via a timer and thus have a
continuous background purge process running every N seconds, which should keep
your database shiny clean.

On the other hand, for my own immediate needs PurgeOnShutdown is sufficient.

