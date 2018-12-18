#!/bin/bash

sleep 30
echo Start $HOSTNAME $WORKER_ID

while : ; do
  echo Update Mecab
  cd /work/ntrust-worker
  ./update_mecab.sh

  cd /work/ntrust-worker
  ./basic_measure_worker.py
  retcode=$?

  if [ $retcode -ne 99 ]; then
      break
  fi
done