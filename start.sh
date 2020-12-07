#!/bin/bash
base_path=/home/ludo/ludo
exec_path=${base_path}/ludo-server
conf_path=${base_path}/bin/conf/server.json
log_path=${base_path}/logs
bi_path=${base_path}/bi
stop_path=${base_path}
sh ${stop_path}/stop.sh
nohup ${exec_path} -conf ${conf_path} -log ${log_path} -bi ${bi_path} -wd ${base_path} > ${stop_path}/nohup.log 2>&1 &