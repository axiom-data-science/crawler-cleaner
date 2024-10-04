# crawler-cleaner 🕷️🧹✨

Processes json web log input from stdin (one json log object per line),
removing any user agents determined to be crawlers/bot/scrapers.

The project uses the
[crawler-user-agents](https://github.com/monperrus/crawler-user-agents)
project as the user agent database.

By default crawler-cleaner looks for the user agent in the top level
`http_user_agent` field in each json log. This may be configured using the
`-user-agent-key` flag (but must be top level).

Detected crawler logs can be discarded (default) or written to a separate
file/stream. JSON parse errors (error messages between json logs etc)
can also be written to a separate file. The following strings have
special meaning for the output files;

* `0`, `/dev/null`, `null` - discard output
* `-`, `/dev/stdout`, `stdout` - write output to stdout
- `+`, `/dev/stderr`, `stderr` - write output to stderr

## Example usage

```
$ ./crawler-cleaner -help
Usage of ./crawler-cleaner:
  -crawler-output string
        File to write crawler output to (default "/dev/null")
  -error-output string
        File to write unparsable json iput to (default "/dev/null")
  -extra-crawler-agents-file string
        File containing additional crawler user agent patterns, one per line
  -non-crawler-output string
        File to write non-crawler output to (default "/dev/stdout")
  -user-agent-key string
        Json key for user agent (default "http_user_agent")
```

```
$ cat web.log | ./crawler-cleaner -crawler-output ./crawlers.log \
  -non-crawler-output ./legit.log -error-output errors.log
```

or

```
$ <web.log ./crawler-cleaner -non-crawler-output stderr > ./legit.log 2> ./crawlers.log
```

## Reviewing results

After running, it's useful to examine the user agents in both non-crawler
and crawler outputs to identify any adjustments needed. Example command
to view counts of user agents using [`jq`](https://jqlang.github.io/jq/):

```
<non-crawler.log jq '.http_user_agent' | sort | uniq -c | sort -n | less
```

## Adding extra crawler agents

To add crawler agents to the set obtained from
[crawler-user-agents](https://github.com/monperrus/crawler-user-agents/blob/master/crawler-user-agents.json),
you may create a text file (default `./extra-user-agents.txt` and add user agent patterns,
one per line. Note that these patterns can be regular expressions, but forward slashes `/` do
not need to be escaped as they are in the `crawler-user-agents.json` file
(i.e. use `meta-externalagent/` and not `meta-externalagent\\/`.
