#!/bin/bash

DEP_MAP_FILE="/tmp/mao_deps_map.csv"

allDepsDup=""
for gogo in `find | grep "\.go"`
do
    moduleName=`cat ${gogo} | grep -m1 "MODULE_NAME =" | awk -F '"' '{print $2}'`
    for deps in `cat ${gogo} | grep "MaoCommon" | grep "Service" | grep -v "//" | awk -F 'Get' '{print $2}' | sed 's/()//' | grep -vE ^$`
    do
        thisDep="$moduleName=>$deps"
        allDepsDup="$allDepsDup\n$thisDep"
    done
done

echo "=================================================="
echo -e $allDepsDup | grep -vE ^$ | wc -l
echo -e $allDepsDup | grep -vE ^$ | sort -u | wc -l
echo "**************************************************"

allDepsUnique=`echo -e $allDepsDup | grep -vE ^$ | sort -u`
allDepsDup=`echo -e $allDepsDup | grep -vE ^$`

rm $DEP_MAP_FILE 2> /dev/null

for uni in $allDepsUnique
do
    count=0
    for dup in $allDepsDup
    do
        if [ $uni = $dup ]
        then
            let count++
            # count=$((count+1))
        fi
    done
    echo "$uni=>$count"
    echo "$uni=>$count" | sed 's/=>/,/g' >> $DEP_MAP_FILE
done

echo "=================================================="
