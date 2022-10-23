#!/bin/bash

set -e

BENCHTIME=${BENCHTIME:-"1m"}

go test -bench=.  -cpu 1,2,3 -benchmem -benchtime "$BENCHTIME" -v github.com/acknode/ackstream/app | tee ./tests/benchplots/out.data

awk '/Benchmark/{count ++; gsub(/BenchmarkTest/,""); printf("%d,%s,%s,%s\n",count,$1,$2,$3)}' ./tests/benchplots/out.data > ./tests/benchplots/temp.data
sed '1d' ./tests/benchplots/temp.data > ./tests/benchplots/final.data

gnuplot -e "file_path='./tests/benchplots/final.data'" -e "graphic_file_name='./tests/benchplots/operations.png'" -e "y_label='number of operations'" -e "y_range_min='000000''" -e "y_range_max='500000'" -e "column_1=1" -e "column_2=3" ./tests/benchplots/performance.gp
gnuplot -e "file_path='./tests/benchplots/final.data'" -e "graphic_file_name='./tests/benchplots/time_operations.png'" -e "y_label='each operation in nanoseconds'" -e "y_range_min='0000000''" -e "y_range_max='5000000'" -e "column_1=1" -e "column_2=4" ./tests/benchplots/performance.gp

cat ./tests/benchplots/final.data
rm -f ./tests/benchplots/out.data ./tests/benchplots/temp.data ./tests/benchplots/final.data