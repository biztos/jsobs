# jsobs TODO list

(more or less prioritized)

## Logging! Want some kind of debug logger we can easily integrate.

Best idea right now is have an interface that just happens to match the
logger of choice, and a default implementation that just uses logging to
keep it simple.  Then pass the nicer logger in at construction in other code.
At the client level.

## Better testing of pgclient purge wait

Actually have a simple solution:

- create one rec to purge
- lock it for update
- go unlock it after sleep N
- go purge
- wait for not purging, will be > N

## Continuous purge for pgclient via timer... MAYBE!

On the one hand, you should not worry about purging at the end of every
process because you could end up with a massive purge.

On that same hand, it's interesting to do this via a timer and thus have a
continuous background purge process running every N seconds, which should keep
your database shiny clean.

On the other hand, for my own immediate needs PurgeOnShutdown is sufficient.

