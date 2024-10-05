#!/bin/bash

ME=$(basename "$0")

function usage() {
   local MSG=$1

   cat<<-EOT
$MSG
This script will rclone the intelchain db to datadir/archive directory.

Usage: $ME [options] datadir shard

datadir:    the root directory of the intelchain db (default: /home/intelchain)
shard:      the shard number to sync (valid value: 0,1,2,3)

Options:
   -h       print this help message
   -c       clean up backup db after rclone
   -a       sync archival db, instead of regular db

EOT
   exit 1
}

CLEAN=false
FOLDER=mainnet.min
CONFIG=/etc/intelchain/rclone.conf

while getopts ":hca" opt; do
   case $opt in
      c) CLEAN=true ;;
      a) FOLDER=mainnet.archival ;;
      *) usage ;;
   esac
done

shift $((OPTIND - 1))

if [ $# != 2 ]; then
   usage
fi

DATADIR="$1"
SHARD="$2"

if [ ! -d "$DATADIR" ]; then
   usage "ERROR: no datadir directory: $DATADIR"
fi

case "$SHARD" in
   0|1|2|3) ;;
   *) usage "ERROR: invalid shard number: $SHARD" ;;
esac

mkdir -p "${DATADIR}/archive"

rclone --config "${CONFIG}" sync -vvv "itc:pub.intelchain.org/${FOLDER}/intelchain_db_${SHARD}" "${DATADIR}/archive/intelchain_db_${SHARD}" > "${DATADIR}/archive/archive-${SHARD}.log" 2>&1

[ -d "${DATADIR}/intelchain_db_${SHARD}" ] && mv -f "${DATADIR}/intelchain_db_${SHARD}" "${DATADIR}/archive/intelchain_db_${SHARD}.bak"
mv -f "${DATADIR}/archive/intelchain_db_${SHARD}" "${DATADIR}/intelchain_db_${SHARD}"

if $CLEAN; then
   rm -rf "${DATADIR}/archive/intelchain_db_${SHARD}.bak"
fi
