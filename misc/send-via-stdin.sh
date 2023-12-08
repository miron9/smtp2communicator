#!/usr/bin/env bash

cat ./test_data/cron-message.txt | sendmail --verbosity debug

