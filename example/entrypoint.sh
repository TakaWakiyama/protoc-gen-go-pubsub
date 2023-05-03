#!/bin/bash

set -em

export PUBSUB_PROJECT_ID=$PROJECT_ID
export PUBSUB_EMULATOR_HOST=0.0.0.0:8085

gcloud beta emulators pubsub start --project=$PUBSUB_PROJECT_ID --host-port=$PUBSUB_EMULATOR_HOST --quiet &

while ! nc -z localhost 8085; do
  sleep 0.1
done

. env/bin/activate
python3 publisher.py $PUBSUB_PROJECT_ID create $TOPIC_ID
python3 subscriber.py $PUBSUB_PROJECT_ID create-push $TOPIC_ID $SUBSCRIPTION_ID $PUSH_ENDPOINT

fg %1