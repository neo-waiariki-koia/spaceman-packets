#!/bin/bash
/server &
P1=$!
/client &
P2=$!
wait $P1 $P2