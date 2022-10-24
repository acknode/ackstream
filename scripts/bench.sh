#!/bin/bash

set -e

MODULE_NAME="github.com/acknode/ackstream"
BENCH_DIR="./tests/benchplots"
BENCH_PROFILE="$BENCH_DIR/performance.gp"
BENCH_TIME=${BENCH_TIME:-"15m"}
BENCH_PARALLEL=${BENCH_PARALLEL:-1}

BENCH_PACKAGES=${BENCH_PACKAGES:-"app services/events"}
for pkg in $BENCH_PACKAGES
do
  echo "PWD=${PWD}, PKG=${pkg}"
  PKG_NAME="${MODULE_NAME}/${pkg}"

  BENCH_NAME="${pkg//\//_}"
  OUT="$BENCH_DIR/$BENCH_NAME.out.data"
  FINAL="$BENCH_DIR/$BENCH_NAME.final.data"
  go test -bench=. -cpu 2,3 -benchmem -benchtime "$BENCH_TIME" -timeout 20m -v "$PKG_NAME" > "$OUT"

  # analyze reports
  awk '/Benchmark/{count++; gsub(/BenchmarkTest/,""); { if ($2 == "") count-=1; else printf("%d,%s,%s,%s\n",count,$1,$2,$3);}} ' "$OUT" > "$FINAL"
  cat "$FINAL"

  # generate report by count
  REPORT_OPERATIONS_COUNT="$BENCH_DIR/$BENCH_NAME.operations.count.png"
  REPORT_OPERATIONS_COUNT_PLOT_MAX=$((BENCH_PARALLEL * 500))
  echo $REPORT_OPERATIONS_COUNT_PLOT_MAX
  gnuplot -e "file_path='$FINAL'" -e "graphic_file_name='$REPORT_OPERATIONS_COUNT'" -e "y_label='number of operations'" -e "y_range_min='0''" -e "y_range_max='$REPORT_OPERATIONS_COUNT_PLOT_MAX'" -e "column_1=1" -e "column_2=3" $BENCH_PROFILE
  # generate report by time
  REPORT_OPERATIONS_TIME="$BENCH_DIR/$BENCH_NAME.operations.time.png"
  REPORT_OPERATIONS_TIME_PLOT_MAX=$((BENCH_PARALLEL * 500000))
  echo $REPORT_OPERATIONS_TIME_PLOT_MAX
  gnuplot -e "file_path='$FINAL'" -e "graphic_file_name='$REPORT_OPERATIONS_TIME'" -e "y_label='each operation in nanoseconds'" -e "y_range_min='0''" -e "y_range_max='$REPORT_OPERATIONS_TIME_PLOT_MAX'" -e "column_1=1" -e "column_2=4" $BENCH_PROFILE

  rm -rf "$OUT" "$FINAL"
done
