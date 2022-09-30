#!/bin/bash
lines=$1
echo "Going to write $lines lines for each Origin AS sorted by their type."
shift
for F in "$@"
do
	echo "Going to handle file $F"
	if [[ "$lines" == "all"  ]]; then
		lines2=$(cat "$F" | wc -l )
	else
		lines2="$(($lines+1))"
	fi
	(head -n 1 "$F" && tail -n +2 "$F" | sort -n -r -k 3,3 -t ",") |head -$lines2 > "victims-${F}";
        (head -n 1 "$F" && tail -n +2 "$F" | sort -n -r -k 4,4 -t ",") |head -$lines2 > "same-${F}";
        (head -n 1 "$F" && tail -n +2 "$F" | sort -n -r -k 5,5 -t ",") |head -$lines2 > "attackers-${F}";
        (head -n 1 "$F" && tail -n +2 "$F" | awk -F "," '{print $2-$6","$0}' | sort -n -r -k 1,1 -t ",") |cut -d',' -f2- |head -$lines2 > "leastlegit-${F}";
done


