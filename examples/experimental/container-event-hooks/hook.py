#! /usr/bin/env python

# Example container event lifecycle hook script.
#
# The Kubelet feeds this program on standard input a JSON-encoded event that
# looks like:
#
# {
#   "name": "PRE_START",
#   "eventContext": {
#     "pod": { ... },
#     "container": { ... },
#     "config": { ... }
#   }
# }
#
# The hook script must emit the (potentially modified) eventContext part, also
# JSON-encoded, to standard output.
#
# {
#   "pod": { ... },
#   "container": { ... },
#   "config": { ... }
# }
#
# Bytes output to standard error will be included in the Kubelet logs.

import json
import sys

# Log program name.
print >> sys.stderr, sys.argv[0]

event = json.loads(''.join(sys.stdin.readlines()))

# Log receipt of event to standard error.
print >> sys.stderr, "Received container lifecycle event: [%s]" % event['name']

# Graffiti the docker container labels to show we were here.
event['config']['Config']['Labels']['this-label-written-by'] = 'hook'

# Emit the event context JSON to standard output.
print json.dumps({
    'pod': event['pod'],
    'container': event['container'],
    'config': event['config']
})

sys.exit(0)
