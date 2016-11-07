## `go-rsyslog-pstats`

[![License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)](http://opensource.org/licenses/MIT)
[![Build Status](https://img.shields.io/travis/slackhq/go-rsyslog-pstats.svg?style=flat-square)](https://travis-ci.org/slackhq/go-rsyslog-pstats)

Parses and forwards [rsyslog process stats](http://www.rsyslog.com/doc/master/configuration/modules/impstats.html)
to a local [statsite](https://github.com/armon/statsite), [statsd](https://github.com/etsy/statsd), or wire
protocol compatible service

## Getting set up

`go-rsyslog-pstats` is meant to be executed by `rsyslog` under
[omprog](http://www.rsyslog.com/doc/master/configuration/modules/omprog.html)

Once you have `go-rsyslog-pstats` installed you can add the following lines to your `rsyslog` config
and restart `rsyslog`

    module(load="omprog")
    ruleset(name="pstats"){
        action(type="omprog" binary="/path/to/go-rsyslog-pstats --port 8125")
    }
    module(load="impstats" interval="10" severity="7" format="json" ruleset="pstats")

## Output

All stats ingested will generally be put into the following format

    rsyslog.<ORIGIN>.<NAME>.<STAT>
    
- `origin` describes the type of module emitting the metrics ie. input modules, actions, etc.
- `name` is the name assigned to the module either automatically or through configuration.
- `stat` is the actual metric emitted by the module.

Each of the above parts are sanitized and will not contain `.`. This was primarily done to avoid
further nesting of values and to simplify graphing everything under a given module.

### Examples

Actions generate the following:

    rsyslog.core_action.custom_relp.failed:0|g
    rsyslog.core_action.custom_relp.suspended:0|g
    rsyslog.core_action.custom_relp.suspended_duration:0|g
    rsyslog.core_action.custom_relp.resumed:0|g
    rsyslog.core_action.custom_relp.processed:1223|g

Queues will generate the following (note this example is of a disk assisted queue, hence the `_da`):

    rsyslog.core_queue.action_1_queue_da.size:0|g
    rsyslog.core_queue.action_1_queue_da.enqueued:0|g
    rsyslog.core_queue.action_1_queue_da.full:0|g
    rsyslog.core_queue.action_1_queue_da.discarded_full:0|g
    rsyslog.core_queue.action_1_queue_da.discarded_nf:0|g
    rsyslog.core_queue.action_1_queue_da.maxqsize:0|g

Rsyslog runtime stats will look like:

    rsyslog.resource_usage.nivcsw:6|g
    rsyslog.resource_usage.utime:0|g
    rsyslog.resource_usage.maxrss:4676|g
    rsyslog.resource_usage.minflt:420|g
    rsyslog.resource_usage.majflt:9|g
    rsyslog.resource_usage.stime:8000|g
    rsyslog.resource_usage.inblock:152|g
    rsyslog.resource_usage.oublock:40|g
    rsyslog.resource_usage.nvcsw:44|g
    
Inputs and [dynstats](http://www.rsyslog.com/doc/master/configuration/dyn_stats.html) are also supported but the
stats they emit are specific to each input module or configuration.
