OLD_CONTAINERS=$(for i in $(cat containers/versions.json | jq 'keys[]' -r); do echo projectunik/$i:$(cat containers/versions.json | jq '.[$key]' --arg key $i -r); done)
for i in $(docker images | grep projectunik | awk '{print $1":"$2}'); do if [[ $OLD_CONTAINERS != *$i* ]]; then echo $i; fi; done
