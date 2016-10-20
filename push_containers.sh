#!/bin/bash
for i in $(printf "%s\n" $(cat containers/versions.json | jq -r 'keys[]')); do echo "pushing projectunik/$i:$(cat containers/versions.json  | jq .['$arg'] -r --arg arg $i)" && docker push projectunik/$i:$(cat containers/versions.json  | jq .['$arg'] -r --arg arg $i); done
