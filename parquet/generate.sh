#!/usr/bin/env bash
set -euo pipefail

message_exit() {
    echo $1
    exit 1
}

type curl &> /dev/null || message_exit "curl is required and is not available"
type thrift &> /dev/null || message_exit "thrift is required and is not available"

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"
PKG=${DIR##*/}
curl https://raw.githubusercontent.com/apache/parquet-format/master/src/main/thrift/parquet.thrift > ${DIR}/parquet.thrift
thrift --out .. --gen go:package=${PKG},read_write_private parquet.thrift